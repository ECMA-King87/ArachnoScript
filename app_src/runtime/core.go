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
	Runtime struct {
		isRepl bool
	}
	Worker struct {
		runtime   *Runtime
		callstack *lib.Array[CallFrame]
		stack     *lib.Array[Value]
		// *parser.Program
		*parser.Parser
		TaskQueue
		MicroTaskQueue
		_break    bool
		invalid   bool
		returned  bool
		_continue bool
		terminate bool
	}
)

func NewRuntime(isRepl bool) *Runtime {
	initMacros()
	return &Runtime{isRepl}
}

type EvalContext struct {
	*Worker
	*Scope
}

func NewContext(scope *Scope) *EvalContext {
	return &EvalContext{Worker: scope.worker, Scope: scope}
}

func (r *Runtime) Worker(path string, builtin bool) *Worker {
	var p *parser.Parser
	if builtin {
		p = parser.NewBuiltinsParser(path)
	} else {
		p = parser.NewParser(path, r.isRepl)
	}
	p.Parse()
	return &Worker{
		Parser:         p,
		TaskQueue:      TaskQueue{},
		MicroTaskQueue: MicroTaskQueue{},
		runtime:        r,
		callstack:      lib.NewArray[CallFrame](10),
		stack:          lib.NewArray[Value](1),
		invalid:        p.Invalid,
		returned:       false,
	}
}

func (w *Worker) crash() {
	// if lib.DEBUG_MODE {
	// 	lib.Println("Worker crashed...")
	// }
	panic("Worker crashed...")
	// lib.Goexit()
}

var UndefinedHash = hashValue(undefined)
var globalThis = func() *ScopeObject {
	scopeObject := &ScopeObject{NewScope(nil, UndefinedHash, GlobalScope, -1), null}
	scopeObject.init("null", null, ConstDecl, lib.DumbyLoc)
	scopeObject.init("undefined", undefined, ConstDecl, lib.DumbyLoc)
	scopeObject.init("true", Boolean(true), ConstDecl, lib.DumbyLoc)
	scopeObject.init("false", Boolean(false), ConstDecl, lib.DumbyLoc)
	scopeObject.init("Infinity", Infinity, ConstDecl, lib.DumbyLoc)
	scopeObject.init("NaN", NaN, ConstDecl, lib.DumbyLoc)
	scopeObject.init("#_stdin", MK_RAW(lib.Stdin), ConstDecl, lib.DumbyLoc)
	scopeObject.init("#_stdout", MK_RAW(lib.Stdout), ConstDecl, lib.DumbyLoc)
	return scopeObject
}()

func (w *Worker) Run() {
	defer func() {
		err := recover()
		if err != nil {
			lib.Println(err)
		}
	}()
	if !w.invalid {
		w.ExecModule(w.Parser.OwnModule)
	}
}

func (w *Worker) ExecModule(prog *parser.Module) {
	progScope := NewScope(globalThis.Scope, UndefinedHash, ModuleScope, -1)
	progScope.path = prog.Path
	progScope.worker = w
	ctx := NewContext(progScope)
	// exports :=
	progLen := len(prog.Body)
	for i := 0; i < progLen && progScope.valid; i++ {
		node := prog.Body[i]
		var rtv Value = undefined
		switch node.Tag {
		case parser.T_IMPORT:
			progScope.error(BuildError, "Import statements uimplemented.", lib.EOL)
		case parser.T_EXPORT:
			progScope.error(BuildError, "Export statements uimplemented.", lib.EOL)
		case parser.T_LABEL:
			ctx.EvalLabel(&i, node, prog.Body)
		default:
			if i == progLen-1 {
				rtv = ctx.EvalStmt(node)
			} else {
				ctx.EvalStmt(node)
			}
		}
		if ctx.runtime.isRepl {
			lib.Stdout.WriteString(ctx.Inspect(rtv) + lib.EOL)
		}
	}
}

func (ctx *EvalContext) ExecBlock(block []parser.Node) (terminate, cont, br bool) {
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
			terminate = ctx.terminate
			if (ctx.of == SwitchScope || ctx.of == LoopScope) && (ctx._break || ctx._continue) {
				br = ctx._break
				cont = ctx._continue
				ctx._break = false
				ctx._continue = false
				ctx.terminate = false
			}
			break
		}
		if ctx.runtime.isRepl {
			lib.Stdout.WriteString(ctx.Inspect(rtv) + lib.EOL)
		}
	}
	return
}

func (ctx *EvalContext) EvalLabel(i *int, node parser.Node, block []parser.Node) *Undefined {
	*i++
	label := node.Data.(string)
	if *i >= len(block) {
		ctx.error(SyntaxError, "A label must precede a statement.", lib.EOL,
			"@", label, " does not have a statement to label.", lib.EOL, lib.SourceLog(ctx.path, node.Loc))
	}
	labeled := block[*i]
	switch label {
	case "debug":
		lib.Stdout.WriteString(
			ctx.Inspect(ctx.EvalStmt(labeled)) + lib.EOL)
	case "coroutine":
		// to carry out a benchmark on a coroutine, do a `@coroutine` on `@benchmark`
		lib.Go(func() {
			ctx.EvalStmt(labeled)
		})
	case "benchmark":
		t := lib.TimeNow()
		ctx.EvalStmt(labeled)
		d := lib.TimeSince(t)
		lib.Stdout.WriteString(d.String() + lib.EOL)
	default:
		ctx.error(SyntaxError, "Unknown label: ", label, lib.EOL, lib.SourceWithinRange(ctx.path, node.Loc))
	}
	return undefined
}

func (ctx *EvalContext) EvalStmt(node Node) Value {
	var (
		scope *Scope = ctx.Scope
	)
	switch node.Tag {
	case parser.T_IMPORT:
		scope.error(SyntaxError, "An import declaration can only be used at the top level of a module.", lib.EOL, lib.SourceLog(scope.path, node.Loc))
	case parser.T_EXPORT:
		scope.error(SyntaxError, "An export declaration can only be used at the top level of a module.", lib.EOL, lib.SourceLog(scope.path, node.Loc))
	case parser.T_BLOCK:
		return ctx.evalBlockStmt(scope, node)
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
	default:
		return ctx.EvalExpr(node)
	}
	return null
}

func (ctx *EvalContext) EvalExpr(node Node) Value {
	var scope *Scope = ctx.Scope
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
		scope.error(SyntaxError, "'...' is unexpected here.", lib.EOL, lib.SourceLog(scope.path, node.Loc))
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
	default:
		scope.error(BuildError, "Unhandled AST node.", lib.EOL, lib.SourceLog(scope.path, node.Loc))
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
	}
)

func NewValueInspector() ValueInspector {
	return ValueInspector{
		stack:  make(Visited),
		ids:    make(map[uintptr]int),
		cyclic: make(map[uintptr]bool),
	}
}

func (w *Worker) Inspect(obj Value) string {
	inspector := NewValueInspector()
	inspector.Analyze(obj)
	return obj.inspect(0, inspector)
}

func (i *ValueInspector) Analyze(v Value) {
	i.walk(v)
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

const (
	CMP_EQUALS lib.Enum = iota
	// CMP_EQUALS2

	CMP_GREATERTHAN
	CMP_LESSTHAN
)

func compareVals(a Value, b Value, op lib.Enum) bool {
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
			case ObjectType, ArrayType, ClassType, FunctionType, InstanceType, MacroType:
				return getValueHash(a) == getValueHash(b)
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

func getValueHash(a Value) uint64 {
	switch v := a.(type) {
	case *Object:
		if v.hash == 0 {
			hashValue(v)
		}
		return v.hash
	case *Array:
		if v.hash == 0 {
			hashValue(v)
		}
		return v.hash
	case *Instance:
		if v.hash == 0 {
			hashValue(v)
		}
		return v.hash
	default:
		return hashValue(v)
	}
}

func hashValue(a Value) uint64 {
	hash := uint64(0)
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
		v.elements.ForEach(func(i int, el Value) {
			v.hash += uint64(i) + getValueHash(el)
		})
		hash = v.hash
	case *Object:
		if v == nil {
			return 0
		}
		v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
			v.hash += getValueHash(key)
			v.hash += getValueHash(pd.value)
		})
		hash = v.hash
	// does not check prototype chain
	case *Instance:
		if v.hash == 0 {
			v.own.ForEach(func(key *String, pd PropertyDescriptor, _ bool) {
				v.hash += getValueHash(key)
				v.hash += getValueHash(pd.value)
			})
			v.hash += uint64(lib.PointerOf(v.proto))
		}
		hash = v.hash
	case *String:
		return uint64(uintptr(lib.Pointer(v)))
	case Number:
		hash = *(*uint64)(lib.Pointer(&v))
	case *Class:
		return uint64(uintptr(lib.Pointer(v)))
	case *Function:
		return uint64(uintptr(lib.Pointer(v)))
	case *Macro:
		return uint64(uintptr(lib.Pointer(v)))
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
