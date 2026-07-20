// The core runtime for the language.
// It implements the interpreter.
package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
)

type Node = parser.Node

var PDB = lib.NewPDB("src")

func (ctx *EvalContext) evalFnStmt(node parser.Node) *Function {
	fn := ctx.evalFnExpr(node)
	ctx.init(fn.name, fn, ConstDecl, node.Loc)
	return fn
}

func (ctx *EvalContext) evalVarDecl(node Node) *Undefined {
	stmt := node.Data.(parser.VarDecl)
	for _, decl := range stmt.Decls {
		ctx.declareName(decl, DeclType(stmt.Kind))
	}
	return undefined
}

func (ctx *EvalContext) evalIfStmt(node Node) *Undefined {
	stmt := node.Data.(parser.IfStmt)
	condNode := stmt.Condition
	var condition Boolean = false
	blockCtx := NewContext(NewScope(ctx.Scope, ctx.objectHash, BlockScope, -1))
	switch condNode.Tag {
	case parser.T_VARDECL:
		decl := condNode.Data.(parser.VarDecl)
		d := decl.Decls[0]
		condition = valueIsTruthy(ctx.declareName(d, DeclType(decl.Kind)))
	default:
		condition = valueIsTruthy(ctx.EvalExpr(condNode))
	}
	ifBlock := node.Children
	if condition {
		blockCtx.ExecBlock(ifBlock)
	} else {
		blockCtx.ExecBlock(stmt.ElseBlock)
	}
	blockCtx.invalidate()
	return undefined
}

func (ctx *EvalContext) evalWhileStmt(node Node) *Undefined {
	stmt := node.Data.(parser.WhileLoop)
	condNode := stmt.Condition
	body := stmt.Body
	var condition Boolean = false
	var d parser.Decl

	for {
		blockCtx := NewContext(NewScope(ctx.Scope, ctx.objectHash, LoopScope, -1))
		if stmt.Do {
			_, _, br := blockCtx.ExecBlock(body)
			if br {
				break
			}
			switch condNode.Tag {
			case parser.T_VARDECL:
				decl := condNode.Data.(parser.VarDecl)
				d = decl.Decls[0]
				condition = valueIsTruthy(blockCtx.declareName(d, DeclType(decl.Kind)))
				condNode = d.Rhs
			default:
				condition = valueIsTruthy(ctx.EvalExpr(condNode))
			}
			if !condition {
				break
			}
		} else {
			switch condNode.Tag {
			case parser.T_VARDECL:
				decl := condNode.Data.(parser.VarDecl)
				d = decl.Decls[0]
				condition = valueIsTruthy(blockCtx.declareName(d, DeclType(decl.Kind)))
				condNode = d.Rhs
			default:
				condition = valueIsTruthy(ctx.EvalExpr(condNode))
			}
			if condition {
				blockCtx.ExecBlock(body)
				if blockCtx.invalid {
					break
				}
			} else {
				break
			}
		}
		blockCtx.invalidate()
	}
	return undefined
}

func (ctx *EvalContext) evalBreakStmt(node Node) *Undefined {
	if ctx.scopeOf(SwitchScope) == nil && ctx.scopeOf(LoopScope) == nil {
		ctx.errorWithSource(SyntaxError, ctx.path, node.Loc, "A 'break' statement can only be used within an enclosing iteration or switch statement.")
	}
	ctx.terminate = true
	ctx._break = true
	return undefined
}

func (ctx *EvalContext) evalContinueStmt(node Node) *Undefined {
	if ctx.scopeOf(SwitchScope) == nil && ctx.scopeOf(LoopScope) == nil {
		ctx.errorWithSource(SyntaxError, ctx.path, node.Loc, "A 'continue' statement can only be used within an enclosing iteration or switch statement.")
	}
	ctx.terminate = true
	// ctx._continue = true
	return undefined
}

func (ctx *EvalContext) evalReturnStmt(node Node) *Undefined {
	expr := node.Children[0]
	if expr.Tag != parser.T_INVALID {
		ctx.stack.Push(ctx.EvalExpr(expr))
	} else {
		ctx.stack.Push(undefined)
	}
	ctx.returned = true
	ctx.terminate = true
	return undefined
}

func (*EvalContext) evalBlockStmt(scope *Scope, node Node) Value {
	blockCtx := NewContext(NewScope(scope, UndefinedHash, BlockScope, -1))
	blockCtx.ExecBlock(node.Children)
	blockCtx.invalidate()
	return undefined
}

func (ctx *EvalContext) declareName(decl parser.Decl, varType DeclType) Value {
	left := decl.Lhs
	right := decl.Rhs
	val := ctx.EvalExpr(right)
	switch left.Tag {
	case parser.T_IDENT:
		name := string(left.Data.(parser.Identifier))
		ctx.init(name, val, varType, left.Loc)
	case parser.T_ARRAY_LIT:
		names := left.Children
		if ctx.typeof(val) != ArrayType {
			ctx.errorWithSource(TypeError, ctx.path, right.Loc, "Cannot array destructure type ", ctx.typeof(val), ".")
		}
		arr := val.(*Array)
		for i, name := range names {
			var el Value = undefined
			arrLen := arr.elements.Len()
			if name.Tag == parser.T_RESTORSPREAD {
				if i < len(names)-1 {
					ctx.errorWithSource(SyntaxError, ctx.path, name.Loc, "A rest element must be last in a destructuring pattern.")
				}
				operand := name.Data.(Node)
				ident := string(operand.Data.(parser.Identifier))
				collected := NewArray()
				for j := i; j < arrLen; j++ {
					collected.push(arr.elements.At(j))
				}
				ctx.init(ident, collected, varType, name.Loc)
				break
			} else {
				if arrLen > i {
					el = arr.elements.At(i)
				}
				ident := string(name.Data.(parser.Identifier))
				ctx.init(ident, el, varType, name.Loc)
			}
		}
	case parser.T_OBJECT_LIT:
		if ctx.typeof(val) != ObjectType {
			ctx.errorWithSource(TypeError, ctx.path, right.Loc, "Cannot object destructure type ", ctx.typeof(val), ".")
		}
		dest := left.Data.(parser.ObjectDest)
		props := dest.Props
		for idx, key := range dest.Keys {
			prop := props[idx]
			var keyVal *String
			var value Value
			name := ""
			name = string(prop.Data.(parser.Identifier))
			if prop.Computed {
				keyVal = NewString(ctx.EvalExpr(key).toString())
			} else {
				keyVal = NewString(name)
			}
			obj := val.(*Object)
			value = lookupMember(obj, keyVal, nil)
			def := prop.Default
			if def.Tag != parser.T_INVALID && valueIsNullish(value) {
				value = ctx.EvalExpr(def)
			}
			ctx.init(name, value, varType, key.Loc)
		}
	default:
		ctx.errorWithSource(BuildError, ctx.path, left.Loc, "Unhandled LHS in variable declaration.")
	}
	return val
}

func lookupProp(obj *Object, keyVal *String, seen map[uint64]bool) (owner *Object, desc PropertyDescriptor, found bool) {
	if seen == nil {
		seen = make(map[uint64]bool, 8)
	}
	h := getValueHash(obj)
	if seen[h] {
		return nil, DefaultPropDesc(undefined), false // cycle detected
	}
	seen[h] = true
	defer delete(seen, h) // backtrack
	desc, found = obj.own.Get(keyVal)
	if found {
		return obj, desc, true
	}
	if obj.proto != nil {
		return lookupProp(obj.proto, keyVal, seen)
	}
	return null, DefaultPropDesc(undefined), false
}

func (ctx *EvalContext) getMember(obj *Object, keyVal *String) Value {
	this, desc, found := lookupProp(obj, keyVal, nil)
	if !found {
		return undefined
	}
	if !desc.public && ctx.findScopeWith(getValueHash(obj)) == nil {
		return undefined
	}
	if desc.getter != nil {
		fnCtx := NewContext(NewFunctionScope(desc.getter, this))
		ctx.declareParams(desc.getter, Args{}, fnCtx)
		fnCtx.init("this", this, ConstDecl, desc.getter.Loc)
		return ctx.execFrame(desc.getter, fnCtx)
	} else if desc.setter != nil {
		// set-only value
		return undefined
	}
	return desc.value
}

func (ctx *EvalContext) setMember(obj *Object, keyVal *String, value Value, scope *Scope, loc parser.Loc) Value {
	this, desc, found := lookupProp(obj, keyVal, nil)
	if !found {
		obj.own.Set(keyVal, DefaultPropDesc(value))
		return value
	}
	if desc.setter != nil {
		fnCtx := NewContext(NewFunctionScope(desc.setter, this))
		ctx.declareParams(desc.setter, Args{value}, fnCtx)
		fnCtx.init("this", this, ConstDecl, desc.setter.Loc)
		ctx.execFrame(desc.setter, fnCtx)
	}
	if desc.getter != nil || !desc.writeable {
		scope.errorWithSource(SyntaxError, scope.path, loc, "Cannot assign to '%s' because it is a read-only property.", keyVal.toString())
	}
	if desc.public || scope.findScopeWith(getValueHash(obj)) != nil {
		obj.own.Set(keyVal, PropDesc(value, desc.configurable, desc.enumerable, desc.writeable, desc.public, desc.getter, desc.setter))
	} else {
		scope.errorWithSource(SyntaxError, scope.path, loc, "Cannot assign to '%s' because it is not an exposed property.", keyVal.toString())
	}
	return value
}

// Bypass descriptor rules
func lookupMember(obj *Object, keyVal *String, seen map[uint64]bool) Value {
	if seen == nil {
		seen = make(map[uint64]bool, 8)
	}
	h := getValueHash(obj)
	if seen[h] {
		return undefined
	}
	seen[h] = true
	defer delete(seen, h)
	member, exists := obj.own.Get(keyVal)
	if exists {
		return member.value
	}
	if obj.proto != nil {
		return lookupMember(obj.proto, keyVal, seen)
	}
	return undefined
}

// ----------------------------------------------------------------------

func (*EvalContext) evalNumber(n float64) Number {
	return Number(n)
}

func (*EvalContext) evalString(str string) *String {
	return NewString(lib.Unquote(str))
}

func (ctx *EvalContext) evalUnaryExpr(node Node) Number {
	operand := node.Data.(parser.OperandExpr).Operand
	v := ctx.EvalExpr(operand)
	if ctx.typeof(v) != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, node.Loc, "The operand of '-' operator must be of type number.")
	}
	return -v.(Number)
}

func (ctx *EvalContext) evalBinaryExpr(node Node) Value {
	expr := node.Data.(parser.LROpExpr)
	left := expr.Lhs
	right := expr.Rhs
	a := ctx.EvalExpr(left)
	b := ctx.EvalExpr(right)
	typeofA := ctx.typeof(a)
	typeofB := ctx.typeof(b)
	if typeofA == StringType || typeofB == StringType {
		return NewString(a.toString() + b.toString())
	}
	const err = "The operands of an arithmetic operator must be of type number."
	if typeofA != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, left.Loc, err)
	}
	if typeofB != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, left.Loc, err)
	}
	return ctx.evalBinaryOp(a.(Number), b.(Number), expr.Op, node.Loc)
}

func (ctx *EvalContext) evalBinaryOp(a, b Number, op parser.TokenTag, loc parser.Loc) Number {
	switch op {
	case parser.PLUS:
		return a + b
	case parser.MINUS:
		return a - b
	case parser.TIMES:
		return a * b
	case parser.STAR2:
		return Number(lib.Pow(float64(a), float64(b)))
	case parser.DIVIDE:
		if b == 0 {
			if lib.ENV.StrictMode {
				ctx.errorWithSource(MathError, ctx.path, loc, "Division by zero.")
			}
			return Infinity
		}
		return a / b
	case parser.MODULO:
		return Number(lib.Modulo(float64(a), float64(b)))
	default:
		ctx.errorWithSource(BuildError, ctx.path, loc, "Unhandled arithmetic op.")
	}
	return 0
}

func (ctx *EvalContext) evalBitwiseBinaryExpr(node parser.Node) Value {
	expr := node.Data.(parser.LROpExpr)
	var a, b = ctx.EvalExpr(expr.Lhs), ctx.EvalExpr(expr.Rhs)
	const err = "The operands of an arithmetic operator must be of type number."
	if ctx.typeof(a) != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, expr.Lhs.Loc, err)
	}
	if ctx.typeof(b) != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, expr.Rhs.Loc, err)
	}
	return ctx.evalBitwiseOp(node.Tag, a.(Number), b.(Number))
}

func (ctx *EvalContext) evalBitwiseOp(tag parser.NodeTag, a, b Number) Value {
	a = Number(lib.TruncFloat(float64(a)))
	b = Number(lib.TruncFloat(float64(b)))
	switch tag {
	case parser.T_BITAND:
		return Number(float64(int64(a) & int64(b)))
	case parser.T_BITCLEAR:
		return Number(float64(int64(a) &^ int64(b)))
	case parser.T_BITOR:
		return Number(float64(int64(a) | int64(b)))
	case parser.T_SHBITL:
		return Number(float64(int64(a) << int64(b)))
	case parser.T_SHBITR:
		return Number(float64(int64(a) >> int64(b)))
	case parser.T_XOR:
		return Number(float64(int64(a) ^ int64(b)))
	default:
		ctx.error(BuildError, "Unhandled bitwise op: ", tag.Name())
	}
	return null
}

func (ctx *EvalContext) evalBitwiseNotExpr(node parser.Node) Value {
	expr := node.Data.(parser.OperandExpr).Operand
	a := ctx.EvalExpr(expr)
	const err = "The operand of the bitwise not operator (~) must be of type number."
	if ctx.typeof(a) != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, expr.Loc, err)
	}
	return ctx.evalBitwiseNot(a.(Number))
}

func (ctx *EvalContext) evalBitwiseNot(a Number) Value {
	return Number(float64(^int64(lib.TruncFloat(float64(a)))))
}

func (ctx *EvalContext) evalComparison(node Node) Value {
	expr := node.Data.(parser.LROpExpr)
	left := expr.Lhs
	right := expr.Rhs
	a := ctx.EvalExpr(left)
	op := expr.Op
	if right.Tag == parser.T_COMPARE_LIST {
		exprs := right.Children
		for _, exp := range exprs {
			b := ctx.EvalExpr(exp)
			if ctx.compareWith(op, a, b, ctx.Scope, node) {
				return True
			}
		}
		return False
	}
	b := ctx.EvalExpr(right)
	return ctx.compareWith(op, a, b, ctx.Scope, node)
}

func (*EvalContext) compareWith(op parser.TokenTag, a Value, b Value, scope *Scope, node Node) Boolean {
	switch op {
	case parser.GT_EQUALS:
		if compareVals(a, b, CMP_EQUALS) {
			return True
		}
		fallthrough
	case parser.GT:
		return Boolean(compareVals(a, b, CMP_GREATERTHAN))
	case parser.LT_EQUALS:
		if compareVals(a, b, CMP_EQUALS) {
			return True
		}
		fallthrough
	case parser.LT:
		return Boolean(compareVals(a, b, CMP_LESSTHAN))
	// no difference between the two for now
	case parser.EQUALS2, parser.EQUALS3:
		return Boolean(compareVals(a, b, CMP_EQUALS))
	case parser.NOT_EQUALS2, parser.NOT_EQUALS:
		return Boolean(!compareVals(a, b, CMP_EQUALS))
	default:
		scope.errorWithSource(BuildError, scope.path, node.Loc, "Unhandled comparison op.")
	}
	return False
}

func (ctx *EvalContext) evalCompareList(node parser.Node) Value {
	ctx.errorWithSource(SyntaxError, ctx.path, node.Loc, "This expression can only appear in comparison expressions at the left-hand-side of comparison operators.")
	return undefined
}

func (ctx *EvalContext) evalIdent(node parser.Node) Value {
	ident := node.Data.(parser.Identifier)
	return ctx.getValue(string(ident), &node.Loc)
}

func (ctx *EvalContext) evalTypeof(node parser.Node) *String {
	return NewString(ctx.typeof(ctx.EvalExpr(node.Data.(parser.OperandExpr).Operand)))
}

func (ctx *EvalContext) evalNullishExpr(node parser.Node) Value {
	expr := node.Data.(parser.LROpExpr)
	a := ctx.EvalExpr(expr.Lhs)
	if valueIsNullish(a) {
		return ctx.EvalExpr(expr.Rhs)
	}
	return a
}

func (ctx *EvalContext) evalArrayLit(node parser.Node) *Array {
	arr := NewArray()
	for _, el := range node.Children {
		arr.push(ctx.EvalExpr(el))
	}
	return arr
}

func (ctx *EvalContext) evalObjectLit(node parser.Node) *Object {
	object := NewObject()
	expr := node.Data.(parser.ObjectLiteral)
	for idx, key := range expr.Keys {
		var propKey *String
		prop := expr.Props[idx]
		if key.Tag == parser.T_IDENT && !prop.Computed {
			propKey = NewString(string(key.Data.(parser.Identifier)))
		} else {
			keyValue := ctx.EvalExpr(key)
			propKey = NewString(keyValue.toString())
		}
		switch prop.Tag {
		case parser.T_FNDECL:
			decl := prop.Data.(parser.FnDecl)
			if decl.Anonymous {
				prop.Data = parser.FnDecl{
					Async:     decl.Async,
					Params:    decl.Params,
					Name:      key,
					Arrow:     decl.Arrow,
					Anonymous: decl.Anonymous,
				}
			}
		case parser.T_CLASSDECL:
			decl := prop.Data.(parser.ClassDecl)
			if decl.Anonymous {
				prop.Data = parser.ClassDecl{
					Name:        propKey.toString(),
					Anonymous:   decl.Anonymous,
					DefaultProp: decl.DefaultProp,
					Methods:     decl.Methods,
					Props:       decl.Props,
					Constructor: decl.Constructor,
					Extends:     decl.Extends,
				}
			}
		}
		if prop.Accessor {
			decl := prop.Node.Data.(parser.FnDecl)
			fn := ctx.EvalExpr(Node{
				Tag:      parser.T_FNDECL,
				Children: prop.Children,
				Data:     decl,
				Loc:      prop.Loc,
			}).(*Function)
			var getter, setter *Function
			if prop.Node.Data.(parser.Accessor).Getter {
				getter = fn
			} else {
				setter = fn
			}
			object.own.Set(propKey, PropDesc(undefined, true, true, false, true, getter, setter))
		} else {
			value := ctx.EvalExpr(prop.Node)
			object.own.Set(propKey, DefaultPropDesc(value))
		}
	}
	return object
}

func (ctx *EvalContext) evalGroupingExpr(node parser.Node) Value {
	var value Value
	exprs := node.Children
	exprsLen := len(exprs)
	for i, exp := range exprs {
		v := ctx.EvalExpr(exp)
		if i == exprsLen-1 {
			value = v
		}
	}
	return value
}

func (ctx *EvalContext) evalFnExpr(node parser.Node) *Function {
	decl := node.Data.(parser.FnDecl)
	name := decl.Name.Data.(string)
	obj := NewObject()
	obj.proto = NewObject()
	obj.proto.own.Set(NewString("name"), PropDesc(NewString(name), false, false, false, true, nil, nil))
	fn := &Function{
		name:      name,
		body:      node.Children,
		async:     decl.Async,
		arrow:     decl.Arrow,
		Object:    obj,
		declScope: ctx.Scope,
		params:    decl.Params,
		Loc:       node.Loc,
	}
	return fn
}

func (ctx *EvalContext) evalCallExpr(node parser.Node) Value {
	expr := node.Data.(parser.CallExpr)
	caller := expr.Caller
	callerVal := ctx.EvalExpr(caller)
	args := ctx.evalArgs(expr.Args)
	return ctx.callWithThis(callerVal, callerVal, args, &node.Loc)
}

func (ctx *EvalContext) evalArgs(args []parser.Node) Args {
	argsVals := Args{}
	for _, arg := range args {
		if arg.Tag == parser.T_RESTORSPREAD {
			operand := ctx.EvalExpr(arg.Data.(Node))
			if ctx.typeof(operand) != ArrayType {
				ctx.errorWithSource(TypeError, ctx.path, arg.Loc, "Cannot array spread type ", ctx.typeof(operand), ".")
			}
			arr := operand.(*Array)
			arr.elements.ForEach(func(_ int, v Value) {
				argsVals = append(argsVals, v)
			})
		} else {
			argsVals = append(argsVals, ctx.EvalExpr(arg))
		}
	}
	return argsVals
}

func (ctx *EvalContext) callWithThis(fun, thisVal Value, args Args, loc *parser.Loc) Value {
	callEnv := ctx.Scope
start:
	switch fn := fun.(type) {
	case *Pointer:
		fun = fn.value()
		goto start
	case *Function:
		fnCtx := NewContext(NewScope(fn.declScope, getValueHash(thisVal), FunctionScope, -1))
		if fn.async {
			callEnv.errorWithSource(BuildError, callEnv.path, *loc, "Async functions are not functional yet.")
		}
		if fn.arrow {
			fnCtx.init("this", thisVal, ConstDecl, fn.Loc)
		} else {
			fnCtx.init("this", NewObject(), ConstDecl, fn.Loc)
		}
		ctx.declareParams(fn, args, fnCtx)
		return ctx.execFrame(fn, fnCtx)
	case *Macro:
		return fn.call(args, ctx, loc)
	default:
		callEnv.errorWithSource(TypeError, callEnv.path, *loc, lib.Sprintf("This expression is not callable. Type '%s' is not a function.", ctx.typeof(fn)))
	}
	return null
}

func (ctx *EvalContext) evalMemberExpr(node parser.Node) Value {
	expr := node.Data.(parser.MemberExpr)
	objNode := expr.Object
	object := ctx.EvalExpr(objNode)
	memberNode := expr.Member
	memberKey := ctx.memberKey(expr)
start:
	switch obj := object.(type) {
	// case Number, *Undefined:
	case *Pointer:
		object = obj.value()
		goto start
	default:
		ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Cannot read properties of %s (reading '%s')", lib.Green(ctx.typeof(obj)), memberKey.toString()))
	case *Object:
		return ctx.getMember(obj, memberKey.(*String))
	case *Array:
		if ctx.typeof(memberKey) != NumberType {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Type '%s' cannot be used to index type array.", lib.Green(ctx.typeof(memberKey))))
		}
		idx := int(lib.TruncFloat(float64(memberKey.(Number))))
		if idx < obj.elements.Len() {
			return obj.elements.At(idx)
		} // undefined
	case *String:
		if ctx.typeof(memberKey) != NumberType {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Type '%s' cannot be used to index type string.", lib.Green(ctx.typeof(memberKey))))
		}
		idx := int(lib.TruncFloat(float64(memberKey.(Number))))
		if idx < len(obj.string) {
			return NewString(string(rune(obj.string[idx])))
		} // undefined
	case *Instance:
		return ctx.getMember(obj.Object, memberKey.(*String))
	case *Function:
		return ctx.getMember(obj.Object, memberKey.(*String))
	case *Class:
		return ctx.getMember(obj.Object, memberKey.(*String))
	case *ScopeObject:
		if expr.Computed {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, "Member expression on scope object cannot be computed.")
		}
		return ctx.getMember(scopeToObject(obj), memberKey.(*String))
	}
	return undefined
}

func (ctx *EvalContext) memberKey(expr parser.MemberExpr) Value {
	var memberKey Value
	if expr.Computed {
		memberKey = ctx.EvalExpr(expr.Member)
	} else {
		memberKey = NewString(string(expr.Member.Data.(parser.Identifier)))
	}
	return memberKey
}

type CallFrame struct {
	fn    *Function
	scope *EvalContext
}

const MAX_CALLSTACK_SIZE = 1_000_000

// ctx is the EvalContext of the callsite.
func (ctx *EvalContext) execFrame(fn *Function, fnCtx *EvalContext) Value {
	if ctx.callstack.Len() >= MAX_CALLSTACK_SIZE {
		ctx.stackOverFlow()
	}
	ctx.callstack.Push(CallFrame{
		fn:    fn,
		scope: fnCtx,
	})
	fnCtx.ExecBlock(fn.body)
	if ctx.returned {
		ctx.returned = false
		ctx.terminate = false
	} else {
		ctx.stack.Push(undefined)
	}
	rtv := ctx.stack.Pop()
	switch rtv.(type) {
	case *Function:
	default:
		fnCtx.invalidate()
	}
	return rtv
}

func (ctx *EvalContext) stackOverFlow() {
	panic("unimplemented")
}

func (ctx *EvalContext) declareParams(fn *Function, args Args, fnCtx *EvalContext) {
	paramsLen := len(fn.params)
	argsLen := len(args)
	for i, p := range fn.params {
		var value Value = undefined
		if i < argsLen {
			value = args[i]
		}
		if fnCtx.declareParam(p, p, value, i, args, ctx.Scope) {
			// rest parameter
			if i != paramsLen-1 {
				fnCtx.errorWithSource(SyntaxError, fnCtx.path, p.Loc, "A rest parameter must be last in a parameter list.")
			}
			break
		}
	}
}

func (ctx *EvalContext) declareParam(p, argNode parser.Node, argument Value, i int, args Args, callEnv *Scope) (shouldReturn bool) {
	switch p.Tag {
	case parser.T_IDENT:
		switch argument.(type) {
		case *String, Number, *Pointer:
			break
		default:
			// Only identifiers for noctx.
			if argNode.Tag == parser.T_IDENT {
				name := string(argNode.Data.(parser.Identifier))
				s := callEnv.resolve(name, callEnv, &argNode.Loc)
				ptr, _ := s.names.Get(name)
				argument = &Pointer{
					ptr:   ptr,
					Scope: callEnv,
				}
			}
		}
		ctx.init(string(p.Data.(parser.Identifier)), argument, MutableDecl, p.Loc)
	case parser.T_RESTORSPREAD:
		node := p.Data.(Node)
		arr := NewArray()
		argsLen := len(args)
		if i < argsLen {
			for j := i; j < argsLen; j++ {
				arr.push(args[j])
			}
		}
		ctx.declareParam(node, p, arr, i, args, callEnv)
		return true
	default:
		ctx.errorWithSource(BuildError, ctx.path, p.Loc, "Unhandled function parameter node.")
	}
	return false
}

func (ctx *EvalContext) evalAssignmentExpr(node parser.Node) Value {
	expr := node.Data.(parser.LROpExpr)
	lhs := expr.Lhs
	rhs := expr.Rhs
	op := expr.Op
	var left Value
	if lhs.Tag == parser.T_ARRAY_LIT {
		// prevent re-eval in array destructuring pattern
		left = NewArray()
	} else {
		left = ctx.EvalExpr(lhs)
	}
	right := ctx.EvalExpr(rhs)

	return ctx.handleAssign(op, left, right, node.Loc, lhs, rhs)
}

func (ctx *EvalContext) handleAssign(op parser.TokenTag, left Value, right Value, loc parser.Loc, lhs, rhs parser.Node) Value {
	// isEqualsAssign := false
	validatedForArithmeticAssign := false
	goto skip
errorOnArithmeticAssign:
	if ctx.typeof(left) != NumberType {
		ctx.errorWithSource(SyntaxError, ctx.path, lhs.Loc, "The left-hand side of an arithmetic operation must be of type 'number'.")
	}
	if ctx.typeof(right) != NumberType {
		ctx.errorWithSource(SyntaxError, ctx.path, rhs.Loc, "The right-hand side of an arithmetic operation must be of type 'number'.")
	}
skip:

	switch op {
	case parser.NULLISH_EQUALS:
		if valueIsNullish(left) {
			// return right (default value / fallback)
			return ctx.handleAssign(parser.EQUALS, left, right, loc, lhs, rhs)
		}
		return left
	case parser.EQUALS:
	case
		parser.MINUS_EQUALS, parser.PLUS_EQUALS,
		parser.DIVIDE_EQUALS, parser.TIMES_EQUALS,
		parser.MODULO_EQUALS, parser.STAR2_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		var aOp parser.TokenTag
		switch op {
		case parser.MINUS_EQUALS:
			aOp = parser.MINUS
		case parser.PLUS_EQUALS:
			aOp = parser.PLUS
		case parser.DIVIDE_EQUALS:
			aOp = parser.DIVIDE
		case parser.TIMES_EQUALS:
			aOp = parser.TIMES
		case parser.MODULO_EQUALS:
			aOp = parser.MODULO
		case parser.STAR2_EQUALS:
			aOp = parser.STAR2
		}
		right = ctx.evalBinaryOp(left.(Number), right.(Number), aOp, loc)
	case parser.BITAND_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_BITAND, left.(Number), right.(Number))
	case parser.BITOR_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_BITOR, left.(Number), right.(Number))
	case parser.BITCLEAR_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_BITCLEAR, left.(Number), right.(Number))
	case parser.SHBITL_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_SHBITL, left.(Number), right.(Number))
	case parser.SHBITR_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_SHBITR, left.(Number), right.(Number))
	case parser.XOR_EQUALS:
		if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
		right = ctx.evalBitwiseOp(parser.T_XOR, left.(Number), right.(Number))
	default:
		ctx.errorWithSource(BuildError, ctx.path, loc, "Unsupported assignment operator: ", op.Lexeme())
	}

	switch lhs.Tag {
	default:
		ctx.errorWithSource(SyntaxError, ctx.path, lhs.Loc, "The left-hand side of an assignment expression must be a variable or a property access.")
	case parser.T_IDENT:
		name := string(lhs.Data.(parser.Identifier))
		ctx.assignName(name, right, &lhs.Loc)
	case parser.T_MEMBER:
		expr := lhs.Data.(parser.MemberExpr)
		// _ = expr.Computed
		// memberNode := expr.Member
		objectNode := expr.Object
		objectVal := ctx.EvalExpr(objectNode)
		memberVal := ctx.memberKey(expr)
	start:
		switch obj := objectVal.(type) {
		case *Pointer:
			objectVal = obj.value()
			goto start
		case *Array:
			// Array index is already validated when evaluating left.
			idx := float64(memberVal.(Number))
			if lib.Modulo(idx, 1) != 0 || idx < 0 {
				ctx.errorWithSource(TypeError, ctx.path, lhs.Loc, "Array index must be a positive integer, received negative or floating point number ", memberVal.toString(), ".")
			}
			return obj.elements.Set(uintptr(idx), right)
		case *Object:
			return ctx.setMember(obj, memberVal.(*String), right, ctx.Scope, loc)
		case *Instance:
			return ctx.setMember(obj.Object, memberVal.(*String), right, ctx.Scope, loc)
		case *ScopeObject:
			name := memberVal.toString()
			if _, ok := obj.names.Get(name); ok {
				// If the value exists in scope
				obj.assignName(name, right, &lhs.Loc)
			} else {
				obj.init(name, right, MutableDecl, lhs.Loc)
			}
		default:
			ctx.errorWithSource(BuildError, ctx.path, loc, "Cannot assign to type ", ctx.typeof(objectVal), ".")
		}
	case parser.T_ARRAY_LIT:
		if ctx.typeof(right) != ArrayType {
			ctx.errorWithSource(TypeError, ctx.path, rhs.Loc, "Cannot array destructure type ", ctx.typeof(right), ".")
		}
		els := lhs.Children
		arr := right.(*Array)
		for i, el := range els {
			arrLen := arr.elements.Len()
			if el.Tag == parser.T_RESTORSPREAD {
				node := el.Data.(Node)
				if i < arrLen-1 {
					ctx.errorWithSource(SyntaxError, ctx.path, node.Loc, "A rest element must be last in a destructuring pattern.")
				}
				collected := NewArray()
				for j := i; j < arrLen; j++ {
					collected.push(arr.elements.At(j))
				}
				ctx.handleAssign(parser.EQUALS, ctx.EvalExpr(node), collected, loc, node, rhs)
				break
			}
			var value Value = undefined
			if i < arrLen {
				value = arr.elements.At(i)
			}
			ctx.handleAssign(op, ctx.EvalExpr(el), value, loc, el, rhs)
		}
	}
	return right
}

func (ctx *EvalContext) evalNotExpr(node parser.Node) Value {
	operand := node.Data.(parser.OperandExpr).Operand
	val := ctx.EvalExpr(operand)
	return !valueIsTruthy(val)
}

// A safe alternative to Value.typeof()
func (ctx *EvalContext) typeof(v Value) string {
	if v == nil {
		// null
		return ObjectType
	}
	return v.typeof()
}
