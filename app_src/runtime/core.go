package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
)

type (
	TaskQueue struct {
		queue []Value
	}
	MicroTaskQueue struct {
		queue []Value
	}
	Runtime struct{}
	Worker  struct {
		// workerPool chan Node
		runtime   *Runtime
		callstack *lib.Array[*CallFrame]
		stack     *lib.Array[Value]
		// *parser.Program
		// *parser.ModuleParser
		// modulePath string
		TaskQueue
		MicroTaskQueue
		// Used in the REPL
		progCtx   *EvalContext
		isRepl    bool
		_break    bool
		returned  bool
		_continue bool
		terminate bool
	}
)

func SetupARE() {
	// TODO: this implementation is not cross platform.
	if !lib.DEBUG_MODE {
		value, err := lib.ReadRegistryValue("Environment", "Path")
		if err == nil {
			execPath := lib.DirOf(lib.ExecPath())
			if value == "" {
				lib.WriteRegistryValue("Environment", "Path", execPath)
			} else {
				found := false
				for _, path := range lib.SplitStr(value, ";") {
					if lib.EqualFold(lib.CleanPath(path), execPath) {
						found = true
						break
					}
				}
				if !found {
					lib.WriteRegistryValue("Environment", "Path", lib.Sprint(value, ";", execPath))
				}
			}
		}
	}
}

func init() {
	initMacros()
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

type EvalContext struct {
	*Worker
	*Scope
}

func NewContext(scope *Scope) *EvalContext {
	return &EvalContext{Worker: scope.worker, Scope: scope}
}

func (r *Runtime) Worker(isRepl bool) *Worker {
	return newWorker(r, isRepl)
}

func (*Runtime) ParseProgram(path string, builtin, isMain, isRepl bool) *parser.ModuleParser {
	return parser.GlobalParser.ParseProgram(path, builtin, isMain, isRepl)
}

func newWorker(r *Runtime, isRepl bool) *Worker {
	return &Worker{
		TaskQueue:      TaskQueue{queue: []Value{}},
		MicroTaskQueue: MicroTaskQueue{queue: []Value{}},
		runtime:        r,
		callstack:      lib.NewArray[*CallFrame](10),
		stack:          lib.NewArray[Value](1),
		// workerPool:     make(chan Node, lib.NumCPU),
		progCtx:   nil,
		isRepl:    isRepl,
		returned:  false,
		_break:    false,
		_continue: false,
		terminate: false,
	}
}

// // Creates a coroutine
// func (w *Worker) Co(n Node) {
// 	w.workerPool <- n
// }

// func (w *Worker) StartTasks() {
// 	for task := range w.workerPool {
// 		w.progCtx.EvalStmt(task)
// 	}
// }

func (w *Worker) crash() {
	// if lib.DEBUG_MODE {
	// 	lib.Println("Worker crashed...")
	// }
	panic("Worker crashed...")
	// lib.Goexit()
}

var globalThis = func() *ScopeObject {
	scopeObject := &ScopeObject{NewScope(nil, undefined, GlobalScope, 0), null}
	scopeObject.init(NullStr, null, ConstDecl, lib.DumbyLoc)
	scopeObject.init(UndefinedStr, undefined, ConstDecl, lib.DumbyLoc)
	scopeObject.init(TrueStr, True, ConstDecl, lib.DumbyLoc)
	scopeObject.init(FalseStr, False, ConstDecl, lib.DumbyLoc)
	scopeObject.init(InfinityStr, Infinity, ConstDecl, lib.DumbyLoc)
	scopeObject.init(NaNStr, NaN, ConstDecl, lib.DumbyLoc)
	scopeObject.init(NewString("#_stdin"), MK_RAW(lib.Stdin), ConstDecl, lib.DumbyLoc)
	scopeObject.init(NewString("#_stdout"), MK_RAW(lib.Stdout), ConstDecl, lib.DumbyLoc)
	return scopeObject
}()

// exports == nil when prog.Invalid == true
func (w *Worker) ExecModule(prog *parser.Module) (exports *Object) {
	if prog.Invalid {
		return
	}
	exports = NewObject()
	defer func() {
		if !lib.DEBUG_MODE {
			err := recover()
			if err != nil {
				lib.Println(err)
			}
		}
	}()
	if w.progCtx == nil {
		progScope := NewScope(globalThis.Scope, undefined, ModuleScope, 0)
		progScope.path = prog.Path
		progScope.worker = w
		w.progCtx = NewContext(progScope)
	}
	ctx := w.progCtx
	progLen := len(prog.Body)
	for i := 0; i < progLen && ctx.valid; i++ {
		node := prog.Body[i]
		var val Value = undefined
		switch node.Tag {
		case parser.T_IMPORT:
			stmt := node.Data.(parser.ImportStmt)
			key := parser.GlobalParser.Imports.GetS(stmt.From)
			if prog.Path == key {
				ctx.errorWithSource(Error, 0, node.Loc, "This module imports itself.")
			}
			lib.Assert(key != 0)
			module := parser.GlobalParser.Program.Modules[key]
			if stmt.UseCurrentContext {
				worker := newWorker(w.runtime, false)
				exps := worker.ExecModule(module)
				if exps == nil {
					exports = nil
					prog.Invalid = true
					return
				}
				// w.progCtx.names.Copy(worker.progCtx.names)
				worker.progCtx.names.ForEach(func(name *String, ptr int, _ bool) {
					varType, _ := worker.progCtx.varTypes.Get(name)
					if name == ThisStr {
						return
					}
					w.progCtx.init(name, worker.progCtx.memory[ptr], varType, node.Loc)
				})
				exports.own.Copy(exps.own)
			} else {
				worker := newWorker(w.runtime, false)
				exps := worker.ExecModule(module)
				if stmt.Named.Tag != parser.T_INVALID {
					names := stmt.Named.Data.(parser.ObjectLiteral)
					for i, K := range names.Keys {
						alias := names.Props[i]
						var name *String
						var key *String
						switch K.Tag {
						case parser.T_IDENT:
							if alias.Tag == parser.T_INVALID {
								name = NewString(string(K.Data.(parser.Identifier)))
								break
							}
							fallthrough
						default:
							name = NewString(string(alias.Data.(parser.Identifier)))
							key = name
						}
						var value Value = undefined
						if pd, ok := exports.own.Get(key); ok {
							value = pd.value
						}
						ctx.init(name, value, ConstDecl, node.Loc)
					}
				}
				if stmt.Namespace != "" {
					ctx.init(NewString(stmt.Namespace), exps, ConstDecl, node.Loc)
				}
			}
		case parser.T_EXPORT:
			ctx.errorWithSource(BuildError, 0, node.Loc, "Export statements unimplemented.")
		case parser.T_RETURN:
			ctx.errorWithSource(SyntaxError, 0, node.Loc, "A 'return' statement can only be used within a function body.")
		case parser.T_LABEL:
			ctx.EvalLabel(&i, node, prog.Body)
		default:
			val = ctx.EvalStmt(node)
		}
		if w.isRepl {
			lib.Stdout.WriteString(w.Inspect(val) + lib.EOL)
		}
	}
	return
}

func (ctx *EvalContext) ExecBlock(block []Node) (terminate, cont, br bool) {
	for i := 0; i < len(block) && ctx.valid; i++ {
		node := block[i]
		var rtv Value = undefined
		switch node.Tag {
		case parser.T_LABEL:
			ctx.EvalLabel(&i, node, block)
		default:
			rtv = ctx.EvalStmt(node)
		}
		if ctx.terminate {
			terminate = true
			if (ctx.of == SwitchScope || ctx.of == LoopScope) && (ctx._break || ctx._continue) {
				br = ctx._break
				cont = ctx._continue
				ctx._break = false
				ctx._continue = false
				ctx.terminate = false
			}
			break
		}
		if ctx.isRepl {
			lib.Stdout.WriteString(ctx.Inspect(rtv) + lib.EOL)
		}
	}
	return
}

var benchmarkCount int

func (ctx *EvalContext) EvalLabel(i *int, node Node, block []Node) *Undefined {
	(*i)++
	label := node.Data.(string)
	if *i >= len(block) {
		ctx.errorWithSource(SyntaxError, 0, node.Loc, "A label must precede a statement.", lib.EOL,
			"@", label, " does not have a statement to label.")
	}
	labeled := block[*i]
	switch label {
	case "debug":
		if labeled.Tag == parser.T_LABEL {
			ctx.errorWithSource(SyntaxError, 0, node.Loc, "Label ", "@", label, " cannot take another label as a statement.")
		} else {
			lib.Stdout.WriteString(
				ctx.Inspect(ctx.EvalStmt(labeled)) + lib.EOL)
		}
	case "coroutine":
		// to carry out a benchmark on a coroutine, do a `@coroutine` on `@benchmark`
		// to make it `@coroutine @benchmark ...`
		lib.Go(func() {
			if labeled.Tag == parser.T_LABEL {
				ctx.EvalLabel(i, labeled, block)
			} else {
				ctx.EvalStmt(labeled)
			}
		})
	case "benchmark":
		if labeled.Tag == parser.T_LABEL && labeled.Data == "coroutine" {
			ctx.errorWithSource(SyntaxError, 0, node.Loc, "Label ", "@", label, " cannot take label @coroutine as a statement.",
				lib.EOL, "Do `@coroutine @benchmark ...` instead.")
		} else {
			benchmarkCount++
			id := benchmarkCount
			t := lib.TimeNow()
			ctx.EvalStmt(labeled)
			d := lib.TimeSince(t)
			lib.Stdout.WriteString("Benchmark " + lib.Sprint(id) + "took ")
			lib.Stdout.WriteString(d.String() + "." + lib.EOL)
		}
	default:
		ctx.errorWithSource(SyntaxError, 0, node.Loc, "Unknown label: ", label)
	}
	return undefined
}

func (ctx *EvalContext) EvalStmt(node Node) Value {
	var (
		scope *Scope = ctx.Scope
	)
	switch node.Tag {
	case parser.T_IMPORT:
		scope.errorWithSource(SyntaxError, 0, node.Loc, "An import declaration can only be used at the top level of a module.")
	case parser.T_EXPORT:
		scope.errorWithSource(SyntaxError, 0, node.Loc, "An export declaration can only be used at the top level of a module.")
	case parser.T_BLOCK:
		return ctx.evalBlockStmt(node)
	case parser.T_VARDECL:
		return ctx.evalVarDecl(node)
	case parser.T_FNDECL:
		return ctx.evalFnStmt(node)
	case parser.T_IFSTMT:
		return ctx.evalIfStmt(node)
	case parser.T_WHILE:
		return ctx.evalWhileStmt(node)
	case parser.T_BREAKSTMT:
		return ctx.evalBreakStmt(node)
	case parser.T_CONTINUESTMT:
		return ctx.evalContinueStmt(node)
	case parser.T_RETURN:
		return ctx.evalReturnStmt(node)
	case parser.T_SWITCHSTMT:
		return ctx.evalSwitchStmt(node)
	case parser.T_THROW:
		return ctx.evalThrowStmt(node)
	case parser.T_CLASSDECL:
		return ctx.evalClassDecl(node)
	case parser.T_TRYCATCH:
		return ctx.evalTryCatchFinally(node)
	case parser.T_FORLOOP:
		return ctx.evalForLoop(node)
	default:
		return ctx.EvalExpr(node)
	}
	return null
}

// func (ctx *EvalContext) EvalPrimitive(node Node) Value {
// 	switch value := ctx.EvalExpr(node).(type) {
// 	case *Instance, *Object:
// 		this, desc, found := lookupProp(valToObjectCast(value), ToPrimitiveTagKey, nil)
// 		if !found {
// 			return value
// 		}
// 		if member, ok := desc.value.(*Function); ok {
// 			return ctx.callWithThis(member, this, Args{}, &node.Loc)
// 		}
// 		return value
// 	default:
// 		return value
// 	}
// }

func (ctx *EvalContext) EvalExpr(node Node) Value {
	switch node.Tag {
	case parser.T_NUMBER:
		return ctx.evalNumber(node.Data.(float64))
	case parser.T_STRING:
		return ctx.evalString(node.Data.(string))
	case parser.T_COMPARE:
		return ctx.evalComparison(node)
	case parser.T_BINARY_EXP:
		return ctx.evalBinaryExpr(node)
	case parser.T_UNARY:
		return ctx.evalUnaryExpr(node)
	case parser.T_IDENT:
		return ctx.evalIdent(node)
	case parser.T_TYPEOFEXPR:
		return ctx.evalTypeof(node)
	case parser.T_NULLISH:
		return ctx.evalNullishExpr(node)
	case parser.T_ARRAY_LIT:
		return ctx.evalArrayLit(node)
	case parser.T_OBJECT_LIT:
		return ctx.evalObjectLit(node)
	case parser.T_BITAND, parser.T_BITCLEAR, parser.T_BITOR, parser.T_SHBITR, parser.T_SHBITL, parser.T_XOR:
		return ctx.evalBitwiseBinaryExpr(node)
	case parser.T_BITNOT:
		return ctx.evalBitwiseNotExpr(node)
	case parser.T_COMPARE_LIST:
		return ctx.evalCompareList(node)
	case parser.T_RESTORSPREAD:
		ctx.errorWithSource(SyntaxError, 0, node.Loc, "'...' is unexpected here.")
	case parser.T_GROUPING:
		return ctx.evalGroupingExpr(node)
	case parser.T_FNDECL:
		return ctx.evalFnExpr(node)
	case parser.T_GT:
		return globalThis
	case parser.T_CALLEXPR:
		return ctx.evalCallExpr(node)
	case parser.T_MEMBER:
		return ctx.evalMemberExpr(node)
	case parser.T_ASSIGN:
		return ctx.evalAssignmentExpr(node)
	case parser.T_NOT:
		return ctx.evalNotExpr(node)
	case parser.T_INCREEXPR:
		return ctx.evalIncreExpr(node)
	case parser.T_LOGICAL:
		return ctx.evalLogicalExpr(node)
	case parser.T_CLASSDECL:
		return ctx.evalClassDecl(node)
	case parser.T_NEWEXPR:
		return ctx.evalNewExpr(node)
	case parser.T_SUPER:
		return ctx.evalSuperExpr(node)
	case parser.T_INEXPR:
		return ctx.evalInExpr(node)
	default:
		ctx.errorWithSource(BuildError, 0, node.Loc, "Unhandled AST node.")
	}
	return null
}

type (
	Visited map[uintptr]struct{}

	ValueInspector struct {
		stack  Visited          // current recursion stack
		ids    map[uintptr]int  // permanent ids for refs
		cyclic map[uintptr]bool // objects that participate in cycles
		nextID int
		worker *Worker
	}
)

func NewValueInspector(worker *Worker) ValueInspector {
	return ValueInspector{
		stack:  make(Visited),
		ids:    make(map[uintptr]int),
		cyclic: make(map[uintptr]bool),
		nextID: 0,
		worker: worker,
	}
}

func (w *Worker) Inspect(obj Value) string {
	inspector := NewValueInspector(w)
	inspector.walk(obj)
	return obj.inspect(0, inspector)
}

func (i *ValueInspector) walk(v Value) {
	if v == nil {
		return
	}
	switch v := v.(type) {
	case *Object:
		i.walkObject(v)
	case *Array:
		i.walkArray(v)
	case *Instance:
		i.walkObject(v.Object)
	case *ScopeObject:
		i.walkObject(scopeToObject(v))
	}
}

func (i *ValueInspector) walkArray(arr *Array) {
	if arr == nil {
		return
	}

	ptr := lib.PointerOf(arr)

	if _, ok := i.stack[ptr]; ok {
		if _, exists := i.ids[ptr]; !exists {
			i.nextID++
			i.ids[ptr] = i.nextID
		}
		i.cyclic[ptr] = true
		return
	}

	i.stack[ptr] = struct{}{}
	defer delete(i.stack, ptr)

	arr.elements.ForEach(func(_ int, v Value) {
		i.walk(v)
	})
}

func (i *ValueInspector) walkObject(obj *Object) {
	if obj == nil {
		return
	}
	ptr := lib.PointerOf(obj)
	if _, ok := i.stack[ptr]; ok {
		if _, exists := i.ids[ptr]; !exists {
			i.nextID++
			i.ids[ptr] = i.nextID
		}
		i.cyclic[ptr] = true
		return
	}
	i.stack[ptr] = struct{}{}
	defer delete(i.stack, ptr)
	obj.own.ForEach(func(_ *String, pd PropertyDescriptor, _ bool) {
		i.walk(pd.value)
	})
}

func newStackTrace(n int, ctx *EvalContext) *String {
	stackLen := ctx.callstack.Len()
	if n == 0 || stackLen == 0 {
		return &emptyString
	}
	b := lib.NewStringBuilder()
	fn := ctx.callstack.At(-1)
	b.WriteString(lib.SourceWithinRange(fn.callEnvPath, fn.callEnvLoc))
	sliced := ctx.callstack.Slice(max(0, stackLen-n), n)
	arrLen := sliced.Len()
	for idx := -1; idx+arrLen >= 0; idx-- {
		cf := sliced.At(idx)
		path := lib.PathFromKey(cf.callEnvPath)
		dirPath := lib.DirOf(path)
		dir, err := lib.LocalizePath(dirPath)
		if err != nil {
			dir = dirPath
		}
		b.WriteString(lib.Sprintf("\tat %s (%s%s:\x1b[33m%d\x1b[0m:\x1b[33m%d\x1b[0m)%s",
			lib.Bold(lib.Italic(cf.fn.name)),
			lib.Dull(dir+string(lib.Separator)),
			lib.Cyan(lib.BaseName(path)),
			cf.callEnvLoc.Line, cf.callEnvLoc.Col, lib.EOL))
	}
	return NewString(b.String())
}

type CMP_OP lib.Enum

const (
	CMP_EQUALS CMP_OP = iota
	// CMP_EQUALS2

	CMP_GREATERTHAN
	CMP_LESSTHAN
)

func compareVals(a Value, b Value, op CMP_OP) bool {
	switch op {
	case CMP_EQUALS:
		// null
		if a == nil && b == nil {
			return true
		}
		if a.typeof() == b.typeof() {
			switch a.typeof() {
			// comparable types
			case NumberType, BooleanType:
				return a == b
			case StringType:
				// No duplicate strings can exist, so we can use pointer comparison.
				return a.(*String) == b.(*String)
			case SymbolType:
				return a.(*Symbol).description == b.(*Symbol).description
			case UndefinedType:
				return true
			case ClassType, FunctionType, MacroType, ScopeType:
				// Compare Identity (pointers)
				return a == b
			case ObjectType, ArrayType, InstanceType:
				if getValueHash(a) != getValueHash(b) {
					return false
				}
				return lib.DeepEqual(a, b)
			case RawType:
				return lib.DeepEqual(a, b)
			}
		}
		return false
	case CMP_GREATERTHAN:
		return toInt(a) > toInt(b)
	case CMP_LESSTHAN:
		return toInt(a) < toInt(b)
	default:
		lib.Panic("Unhandled operation.")
		return false
	}
}

func getValueHash(a Value) uintptr {
	switch v := a.(type) {
	case *Object:
		if v == nil {
			return 0
		}
		if v.hash == 0 {
			return hashValue(v)
		}
		return v.hash
	case *Array:
		if v.hash == 0 {
			return hashValue(v)
		}
		return v.hash
	case *Instance:
		if v.hash == 0 {
			return hashValue(v)
		}
		return v.hash
	default:
		return hashValue(v)
	}
}

const HashSeed = 3

func hashValue(a Value) uintptr {
	hash := uintptr(0)
	if a == nil {
		return 0
	}
	switch v := a.(type) {
	case Boolean:
		if v {
			return ^hash - 1
		}
		return ^hash - 2
	case *Undefined:
		return ^hash
	case *Array:
		if v.hash == 0 {
			v.elements.ForEach(func(i int, el Value) {
				v.hash += hashValue(Number(i))
				if el == v {
					v.hash += uintptr(lib.Pointer(v))
				} else {
					v.hash += getValueHash(el)
					v.hash *= HashSeed
				}
			})
		}
		hash = v.hash
	case *Object:
		if v == nil {
			return 0
		}
		if v.hash == 0 {
			v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
				v.hash += getValueHash(key)
				v.hash += getValueHash(pd.value)
				v.hash *= HashSeed
			})
		}
		hash = v.hash
	// does not check prototype chain
	case *Instance:
		if v.hash == 0 {
			v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
				v.hash += getValueHash(key)
				v.hash += getValueHash(pd.value)
				v.hash *= HashSeed
			})
			v.hash += (lib.PointerOf(v.proto))
		}
		hash = v.hash
	case *String:
		return uintptr(lib.Pointer(v))
	case Number:
		hash = uintptr(*(*uint64)(lib.Pointer(&v)))
	case *Class:
		return uintptr(lib.Pointer(v))
	case *Function:
		return uintptr(lib.Pointer(v))
	case *Macro:
		return uintptr(lib.Pointer(v))
	}
	return hash
}

func valueIsTruthy(v Value) Boolean {
	switch v := v.(type) {
	case Number:
		return v != 0
	case *String:
		return len(v.string) > 0
	case *Object:
		return v != nil
	case *Function, *Class, *Macro, *Symbol:
		return true
	case Boolean:
		return v
	case *Array:
		return v.elements.Len() > 0
	}
	return false
}

func valueIsNullish(v Value) Boolean {
	switch v := v.(type) {
	case *Undefined:
		return true
	case *Object:
		return v == nil // null
	}
	return false
}

// func Error()
