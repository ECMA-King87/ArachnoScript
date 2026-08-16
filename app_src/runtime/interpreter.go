// The core runtime for the language.
// It implements the interpreter.
package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
)

type Node = *parser.Node

var PDB = lib.NewPDB("src")

func (ctx *EvalContext) evalFnStmt(node Node) *Function {
	decl := node.Data.(parser.FnDecl)
	name := decl.Name.Data.(string)
	fn := NewFunction(name, node.Children, node.Loc, decl, ctx, NewObject())
	ctx.init(NewString(fn.name), fn, ConstDecl, node.Loc)
	return fn
}

func (ctx *EvalContext) evalVarDecl(node Node) *Undefined {
	stmt := node.Data.(parser.VarDecl)
	for _, decl := range stmt.Decls {
		var value Value = undefined
		if decl.Rhs != nil {
			value = ctx.EvalExpr(decl.Rhs)
		}
		ctx.declareName(decl.Lhs, value, decl.Lhs.Loc, DeclType(stmt.Kind))
	}
	return undefined
}

func (ctx *EvalContext) evalIfStmt(node Node) *Undefined {
	stmt := node.Data.(parser.IfStmt)
	condNode := stmt.Condition
	blockCtx := NewContext(NewScope(ctx.Scope, ctx.object, BlockScope, 0))
	if condNode.Decl != nil {
		decl := condNode.Decl.Data.(parser.VarDecl)
		d := decl.Decls[0]
		blockCtx.declareName(d.Lhs, ctx.EvalExpr(d.Rhs), d.Rhs.Loc, DeclType(decl.Kind))
	}
	condition := valueIsTruthy(blockCtx.EvalExpr(condNode.Node))
	ifBlock := node.Children
	if condition {
		blockCtx.ExecBlock(ifBlock)
	} else if len(stmt.ElseBlock) > 0 {
		blockCtx.ExecBlock(stmt.ElseBlock)
	}
	blockCtx.invalidate()
	return undefined
}

func (ctx *EvalContext) evalWhileStmt(node Node) *Undefined {
	stmt := node.Data.(parser.WhileLoop)
	condNode := stmt.Condition
	body := stmt.Body
	for {
		blockCtx := NewContext(NewScope(ctx.Scope, ctx.object, LoopScope, 0))
		if stmt.Do {
			_, _, br := blockCtx.ExecBlock(body)
			if br {
				break
			}
			if condNode.Decl != nil {
				decl := condNode.Decl.Data.(parser.VarDecl)
				d := decl.Decls[0]
				blockCtx.declareName(d.Lhs, ctx.EvalExpr(d.Rhs), d.Rhs.Loc, DeclType(decl.Kind))
			}
			condition := valueIsTruthy(blockCtx.EvalExpr(condNode.Node))
			if !condition {
				break
			}
		} else {
			if condNode.Decl != nil {
				decl := condNode.Decl.Data.(parser.VarDecl)
				d := decl.Decls[0]
				blockCtx.declareName(d.Lhs, ctx.EvalExpr(d.Rhs), d.Rhs.Loc, DeclType(decl.Kind))
			}
			condition := valueIsTruthy(blockCtx.EvalExpr(condNode.Node))
			if condition {
				_, _, br := blockCtx.ExecBlock(body)
				if br || !blockCtx.valid {
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
	ctx._continue = true
	return undefined
}

func (ctx *EvalContext) evalReturnStmt(node Node) *Undefined {
	if ctx.scopeOf(FunctionScope) == nil {
		ctx.errorWithSource(SyntaxError, 0, node.Loc, "A 'return' statement can only be used within a function body.")
	}
	expr := node.Children[0]
	if expr != nil {
		ctx.stack.Push(ctx.EvalExpr(expr))
	} else {
		ctx.stack.Push(undefined)
	}
	ctx.returned = true
	ctx.terminate = true
	return undefined
}

func (ctx *EvalContext) evalThrowStmt(node Node) *Undefined {
	expr := node.Children[0]
	ctx.error(Error, 0, lib.DumbyLoc, ctx.EvalExpr(expr))
	return undefined
}

func (ctx *EvalContext) evalBlockStmt(node Node) Value {
	blockCtx := NewContext(NewScope(ctx.Scope, undefined, BlockScope, 0))
	blockCtx.ExecBlock(node.Children)
	blockCtx.invalidate()
	return undefined
}

func (ctx *EvalContext) evalSwitchStmt(node Node) Value {
	stmt := node.Data.(parser.SwitchStmt)
	cond := stmt.Condition
	blockCtx := NewContext(NewScope(ctx.Scope, undefined, SwitchScope, 0))
	// Use blockCtx and EvalStmt for the case of var decls.
	var condValue Value
	if cond.Tag == parser.T_VARDECL {
		vardecl := cond.Data.(parser.VarDecl)
		decl := vardecl.Decls[0]
		condValue = blockCtx.EvalStmt(decl.Rhs)
		ctx.declareName(decl.Lhs, condValue, decl.Rhs.Loc, DeclType(vardecl.Kind))
	} else {
		condValue = blockCtx.EvalStmt(cond)
	}
	noMatch := true
switch_:
	for i, matches := range stmt.Matches {
		for _, match := range matches {
			if compareVals(condValue, blockCtx.EvalExpr(match), CMP_EQUALS) {
				noMatch = false
				t, c, b := blockCtx.ExecBlock(stmt.Cases[i])
				blockCtx.invalidate()
				if t && c && b {
					// fallthrough
					if i+1 >= len(stmt.Matches) {
						ctx.errorWithSource(SyntaxError, 0, node.Loc, "Cannot fallthrough final case in switch.")
					}
					continue
				}
				break switch_
			}
		}
	}
	if noMatch && len(stmt.Default) > 0 {
		blockCtx.ExecBlock(stmt.Default)
		blockCtx.invalidate()
	}
	return undefined
}

func (ctx *EvalContext) evalClassDecl(node Node) Value {
	decl := node.Data.(parser.ClassDecl)
	name := decl.Name
	if decl.Anonymous {
		name = parser.AnonymousName
	}
	cl := NewClass(name, decl, ctx, node.Loc)
	if decl.Anonymous {
		return cl
	}
	return ctx.init(NewString(name), cl, ConstDecl, node.Loc)
}

func (ctx *EvalContext) evalTryCatchFinally(node Node) Value {
	stmt := node.Data.(parser.TryCatch)
	tryBlock := stmt.Try
	tryCtx := NewContext(NewScope(ctx.Scope, ctx.object, TryScope, 0))
	tryCtx.ExecBlock(tryBlock)
	catchBlock := stmt.Catch
	if !tryCtx.valid && catchBlock != nil && !tryCtx.returned {
		capture := stmt.Capture
		catchCtx := NewContext(NewScope(ctx.Scope, ctx.object, BlockScope, 0))
		arg := tryCtx.stack.Pop()
		catchCtx.declareParam(capture, arg, 0, Args{arg})
		catchCtx.ExecBlock(catchBlock)
	}
	finally := stmt.Finally
	if finally != nil {
		finallyCtx := NewContext(NewScope(ctx.Scope, ctx.object, BlockScope, 0))
		finallyCtx.ExecBlock(finally)
	}
	return undefined
}

func (ctx *EvalContext) evalForLoop(node Node) Value {
	stmt := node.Data.(parser.ForLoop)
	left := stmt.LHS
	right := stmt.RHS
	op := stmt.Op
	val := ctx.EvalExpr(right)
	enum := []Value{}
	switch op {
	case parser.In_Op:
		switch v := val.(type) {
		case *Array:
			for i := range v.elements.Len() {
				enum = append(enum, Number(i))
			}
		default:
			cast := valToObjectCast(v)
			if v == undefined || v == null || cast == null {
				ctx.errorWithSource(TypeError, 0, right.Loc, "Cannot iterate over value, it is not iterable.")
			}
			cast.own.ForEach(func(key *String, _ PropertyDescriptor, _ bool) {
				enum = append(enum, key)
			})
		}
	case parser.Of_Op:
		switch v := val.(type) {
		case *Array:
			v.elements.ForEach(func(_ int, v Value) {
				enum = append(enum, v)
			})
		default:
			cast := valToObjectCast(v)
			if v == undefined || v == null || cast == null {
				ctx.errorWithSource(TypeError, 0, right.Loc, "Cannot iterate over value, it is not iterable.")
			}
			cast.own.ForEach(func(k *String, _ PropertyDescriptor, _ bool) {
				enum = append(enum, ctx.getMember(cast, k))
			})
		}
	}
	for _, e := range enum {
		body := node.Children
		evalCtx := NewContext(NewScope(ctx.Scope, ctx.object, LoopScope, 0))
		evalCtx.declareName(left, e, right.Loc, DeclType(stmt.DeclKind))
		_, _, b := evalCtx.ExecBlock(body)
		evalCtx.invalidate()
		if b {
			break
		}
	}
	return undefined
}

func (ctx *EvalContext) declareName(left Node, val Value, rightLoc parser.Loc, varType DeclType) Value {
	switch left.Tag {
	case parser.T_IDENT:
		name := NewString(string(left.Data.(parser.Identifier)))
		ctx.init(name, val, varType, left.Loc)
	case parser.T_ARRAY_LIT:
		names := left.Children
		if val.typeof() != ArrayType {
			ctx.errorWithSource(TypeError, ctx.path, rightLoc, "Cannot array destructure type ", val.typeof().string, ".")
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
				collected := NewArray()
				for j := i; j < arrLen; j++ {
					collected.push(arr.elements.At(j))
				}
				ident := NewString(string(operand.Data.(parser.Identifier)))
				ctx.init(ident, collected, varType, name.Loc)
				break
			} else {
				if arrLen > i {
					el = arr.elements.At(i)
				}
				ident := NewString(string(name.Data.(parser.Identifier)))
				ctx.init(ident, el, varType, name.Loc)
			}
		}
	case parser.T_OBJECT_LIT:
		if val.typeof() != ObjectType {
			ctx.errorWithSource(TypeError, ctx.path, rightLoc, "Cannot object destructure type ", val.typeof().string, ".")
		}
		dest := left.Data.(parser.ObjectDest)
		props := dest.Props
		for idx, key := range dest.Keys {
			prop := props[idx]
			keyVal, name := ctx.evalDestKey(prop, key)
			var value Value
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

func (ctx *EvalContext) evalMemberKey(isComputed bool, key *parser.Node) (propKey *String) {
	if isComputed {
		keyValue := ctx.EvalExpr(key)
		propKey = NewString(keyValue.toString().string)
	} else {
		propKey = NewString(string(key.Data.(parser.Identifier)))
	}
	return
}

func (ctx *EvalContext) evalDestKey(prop parser.ObjectDestProp, key *parser.Node) (keyVal *String, name *String) {
	name = NewString(string(prop.Data.(parser.Identifier)))
	if prop.Computed {
		keyVal = ctx.EvalExpr(key).toString()
	} else {
		keyVal = name
	}
	return
}

func lookupProp(obj *Object, keyVal *String, seen map[uintptr]bool) (owner *Object, desc PropertyDescriptor, found bool) {
	if seen == nil {
		seen = make(map[uintptr]bool, 8)
	}
	h := getValueHash(obj)
	if seen[h] {
		return null, DefaultPropDesc(undefined), false // cycle detected
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

func (ctx *EvalContext) getMember(v Value, keyVal *String) Value {
	var object *Object
	switch o := v.(type) {
	case *Instance:
		object = o.Object
	case *Class:
		object = o.Object
	case *Function:
		object = o.Object
	case *Object:
		object = o
	case *ScopeObject:
		object = scopeToObject(o)
	default:
		// Ignore other types
		return undefined
	}
	_, desc, found := lookupProp(object, keyVal, nil)
	if !found {
		return undefined
	}
	if !desc.public && ctx.findScopeWith(v) == nil {
		return undefined
	}
	if desc.getter != nil {
		return ctx.execFrame(desc.getter, Args{}, desc.getter.Loc)
	} else if desc.setter != nil {
		// set-only value
		return undefined
	}
	return desc.value
}

// v is the object we're considering
// v != obj when obj == v.(Instance|Class|Funtion).Object
func (ctx *EvalContext) setMember(obj *Object, v Value, keyVal *String, value Value, loc parser.Loc) Value {
	owner, desc, found := lookupProp(obj, keyVal, nil)
	if !found {
		obj.own.Set(keyVal, DefaultPropDesc(value))
		return value
	}
	canSet := ctx.findScopeWith(v) != nil
	if !desc.public && !canSet {
		ctx.errorWithSource(SyntaxError, ctx.path, loc, lib.Sprintf("Cannot assign to '%s' because it is not an exposed property.", keyVal.toString().string))
	}
	if desc.setter != nil {
		ctx.execFrame(desc.setter, Args{value}, loc)
		return value
	}
	if desc.getter != nil || !desc.writable {
		ctx.errorWithSource(SyntaxError, ctx.path, loc, lib.Sprintf("Cannot assign to '%s' because it is a read-only property.", keyVal.toString().string))
	}
	owner.own.Set(keyVal, PropDesc(value, desc.configurable, desc.enumerable, desc.writable, desc.public, false, desc.getter, desc.setter))
	return value
}

// Bypass descriptor rules
func lookupMember(obj *Object, keyVal *String, seen map[uintptr]bool) Value {
	if seen == nil {
		seen = make(map[uintptr]bool, 8)
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
	return NewString(lib.Unquote(lib.Sprintf("\"%s\"", str)))
}

func (ctx *EvalContext) evalUnaryExpr(node Node) Number {
	expr := node.Data.(parser.OpOpExpr)
	operand := expr.Operand
	v := ctx.EvalExpr(operand)
	if expr.Op == parser.PLUS {
		if v.typeof() != StringType {
			ctx.errorWithSource(TypeError, ctx.path, node.Loc, "The operand of '+' operator must be of type string.")
		}
		return Number(lib.ParseNumber(v.(*String).string))
	}
	if v.typeof() != NumberType {
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
	typeofA := a.typeof()
	typeofB := b.typeof()
	if typeofA == StringType || typeofB == StringType {
		return NewString(a.toString().string + b.toString().string)
	}
	const err = "The operands of an arithmetic operator must be of type number."
	if typeofA != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, left.Loc, err)
	}
	if typeofB != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, right.Loc, err)
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

func (ctx *EvalContext) evalBitwiseBinaryExpr(node Node) Value {
	expr := node.Data.(parser.LROpExpr)
	var a, b = ctx.EvalExpr(expr.Lhs), ctx.EvalExpr(expr.Rhs)
	const err = "The operands of an arithmetic operator must be of type number."
	if a.typeof() != NumberType {
		ctx.errorWithSource(TypeError, ctx.path, expr.Lhs.Loc, err)
	}
	if b.typeof() != NumberType {
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
		ctx.raise(BuildError, "Unhandled bitwise op: ", tag.Name())
	}
	return null
}

func (ctx *EvalContext) evalBitwiseNotExpr(node Node) Value {
	expr := node.Data.(parser.OperandExpr).Operand
	a := ctx.EvalExpr(expr)
	const err = "The operand of the bitwise not operator (~) must be of type number."
	if a.typeof() != NumberType {
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

func (ctx *EvalContext) evalCompareList(node Node) Value {
	ctx.errorWithSource(SyntaxError, ctx.path, node.Loc, "This expression can only appear in comparison expressions at the left-hand-side of comparison operators.")
	return undefined
}

func (ctx *EvalContext) evalIdent(node Node) Value {
	ident := node.Data.(parser.Identifier)
	return ctx.getValue(NewString(string(ident)), &node.Loc)
}

func (ctx *EvalContext) evalTypeof(node Node) *String {
	return ctx.EvalExpr(node.Data.(parser.OperandExpr).Operand).typeof()
}

func (ctx *EvalContext) evalNullishExpr(node Node) Value {
	expr := node.Data.(parser.LROpExpr)
	a := ctx.EvalExpr(expr.Lhs)
	if valueIsNullish(a) {
		return ctx.EvalExpr(expr.Rhs)
	}
	return a
}

func (ctx *EvalContext) evalArrayLit(node Node) *Array {
	arr := NewArray()
	for _, el := range node.Children {
		arr.push(ctx.EvalExpr(el))
	}
	return arr
}

func (ctx *EvalContext) evalObjectLit(node Node) *Object {
	object := NewObject()
	expr := node.Data.(parser.ObjectLiteral)
	for idx, key := range expr.Keys {
		prop := expr.Props[idx]
		propKey := ctx.evalMemberKey(prop.Computed, key)
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
					Name:        propKey.toString().string,
					Anonymous:   decl.Anonymous,
					DefProp:     decl.DefProp,
					Methods:     decl.Methods,
					Props:       decl.Props,
					Constructor: decl.Constructor,
					Extends:     decl.Extends,
				}
			}
		}
		if prop.Accessor {
			decl := prop.Node.Data.(parser.FnDecl)
			fn := ctx.EvalExpr(&parser.Node{
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
			object.own.Set(propKey, PropDesc(undefined, true, true, false, true, false, getter, setter))
		} else {
			value := ctx.EvalExpr(prop.Node)
			object.own.Set(propKey, DefaultPropDesc(value))
		}
	}
	return object
}

func (ctx *EvalContext) evalGroupingExpr(node Node) Value {
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

func (ctx *EvalContext) evalFnExpr(node Node) *Function {
	decl := node.Data.(parser.FnDecl)
	name := ""
	if decl.Name.Tag == parser.T_STRING {
		name = decl.Name.Data.(string)
	}
	if decl.Name.Tag == parser.T_IDENT {
		name = string(decl.Name.Data.(parser.Identifier))
	}
	var this Value
	if decl.Arrow {
		this = ctx.object
	} else {
		this = NewObject()
	}
	fn := NewFunction(name, node.Children, node.Loc, decl, ctx, this)
	return fn
}

func (ctx *EvalContext) evalCallExpr(node Node) Value {
	expr := node.Data.(parser.CallExpr)
	caller := expr.Caller
	args := ctx.evalArgs(expr.Args)
	callerVal := ctx.EvalExpr(caller)
	return ctx.callWithThis(callerVal, args, &node.Loc)
}

func (ctx *EvalContext) evalArgs(args []Node) Args {
	argsVals := Args{}
	for _, arg := range args {
		if arg.Tag == parser.T_RESTORSPREAD {
			operand := ctx.EvalExpr(arg.Data.(Node))
			if operand.typeof() != ArrayType {
				ctx.errorWithSource(TypeError, ctx.path, arg.Loc, "Cannot array spread type ", operand.typeof().string, ".")
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

func (ctx *EvalContext) callWithThis(fun Value, args Args, loc *parser.Loc) Value {
	callEnv := ctx.Scope
	switch fn := fun.(type) {
	case *Function:
		if fn.async {
			callEnv.errorWithSource(BuildError, callEnv.path, *loc, "Async functions are not functional yet.")
		}
		return ctx.execFrame(fn, args, *loc)
	case *Macro:
		return fn.call(args, ctx, loc)
	default:
		callEnv.errorWithSource(TypeError, callEnv.path, *loc, lib.Sprintf("This expression is not callable. Type '%s' is not a function.", fn.typeof().string))
	}
	return null
}

func (ctx *EvalContext) evalMemberExpr(node Node) Value {
	_, member := ctx.evalMember(node)
	return member
}

func (ctx *EvalContext) evalMember(node Node) (object Value, member Value) {
	expr := node.Data.(parser.MemberExpr)
	objNode := expr.Object
	object = ctx.EvalExpr(objNode)
	memberNode := expr.Member
	memberKey := ctx.memberKey(expr)
	member = undefined
	switch obj := object.(type) {
	default:
		ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Cannot read properties of %s (reading '%s')", obj.typeof().string, lib.Green(memberKey.toString().string)))
	case *Object:
		member = ctx.getMember(obj, toValidPropKey(memberKey))
		return
	case *Array:
		if memberKey.typeof() != NumberType {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Type '%s' cannot be used to index type array.", lib.Green(memberKey.typeof().string)))
		}
		idx := int(lib.TruncFloat(float64(memberKey.(Number))))
		if idx < obj.elements.Len() {
			member = obj.elements.At(idx)
			return
		} // undefined
	case *String:
		if memberKey.typeof() != NumberType {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, lib.Sprintf("Type '%s' cannot be used to index type string.", lib.Green(memberKey.typeof().string)))
		}
		idx := int(lib.TruncFloat(float64(memberKey.(Number))))
		if idx < 0 {
			idx += len(obj.string)
		}
		if idx < len(obj.string) {
			member = NewString(string(rune(obj.string[idx])))
			return
		} // undefined
	case *Instance:
		member = ctx.getMember(obj, toValidPropKey(memberKey))
		return
	case *Function:
		member = ctx.getMember(obj, toValidPropKey(memberKey))
		return
	case *Class:
		member = ctx.getMember(obj, toValidPropKey(memberKey))
		return
	case *ScopeObject:
		if expr.Computed {
			ctx.errorWithSource(TypeError, ctx.path, memberNode.Loc, "Member expression on scope object cannot be computed.")
		}
		member = ctx.getMember(obj, toValidPropKey(memberKey))
		return
	}
	return
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
	fn *Function
	// Essentially for stack message
	// callEnv *EvalContext
	callEnvPath int
	callEnvLoc  lib.Loc
	// scope       *EvalContext
}

const MAX_CALLSTACK_SIZE = 1_000_000

// ctx is the EvalContext of the callsite.
// fnCtx is the context the function body will execute.
// loc is the Loc of the callsite.
func (ctx *EvalContext) execFrame(fn *Function, args Args, loc lib.Loc) Value {
	fnScope := NewFunctionScope(fn, fn.this)
	// Use the callstack of the caller.
	fnWorker := fnScope.worker
	fnScope.worker = ctx.Worker
	evalCtx := NewContext(fnScope)
	evalCtx.declareParams(fn, args)
	defer func() {
		fnScope.worker = fnWorker
		ctx.callstack.Pop()
	}()
	if ctx.callstack.Len() >= MAX_CALLSTACK_SIZE {
		ctx.stackOverFlow()
	}
	ctx.callstack.Push(&CallFrame{
		fn: fn,
		// scope:       evalCtx,
		callEnvPath: ctx.path,
		callEnvLoc:  loc,
	})
	evalCtx.ExecBlock(fn.body)
	if evalCtx.returned {
		evalCtx.returned = false
		evalCtx.terminate = false
	} else {
		evalCtx.stack.Push(undefined)
	}
	rtv := evalCtx.stack.Pop()
	switch rtv.(type) {
	// Closures
	case *Function:
	default:
		evalCtx.invalidate()
	}
	return rtv
}

func (ctx *EvalContext) stackOverFlow() {
	ctx.raise(RangeError, "Maximum stack size exceeded.", lib.EOL, newStackTrace(10, ctx).string)
}

func (fnCtx *EvalContext) declareParams(fn *Function, args Args) {
	paramsLen := len(fn.params)
	argsLen := len(args)
	for i, p := range fn.params {
		var value Value = undefined
		if i < argsLen {
			value = args[i]
		}
		if fnCtx.declareParam(p, value, i, args) {
			// rest parameter
			if i != paramsLen-1 {
				fnCtx.errorWithSource(SyntaxError, fnCtx.path, p.Loc, "A rest parameter must be last in a parameter list.")
			}
			break
		}
	}
}

func (ctx *EvalContext) declareParam(p Node, argument Value, i int, args Args) (shouldReturn bool) {
	switch p.Tag {
	case parser.T_IDENT:
		ctx.init(NewString(string(p.Data.(parser.Identifier))), argument, MutableDecl, p.Loc)
	case parser.T_RESTORSPREAD:
		node := p.Data.(Node)
		arr := NewArray()
		argsLen := len(args)
		if i < argsLen {
			for j := i; j < argsLen; j++ {
				arr.push(args[j])
			}
		}
		ctx.declareParam(node, arr, i, args)
		return true
	case parser.T_OBJECT_LIT:
		// Tries to cast to object type.
		cast := valToObjectCast(argument)
		if cast == null {
			ctx.errorWithSource(TypeError, 0, p.Loc, "Cannot cast type ", lib.Green(argument.typeof().string), " to object for destructuring.")
		}
		dest := p.Data.(parser.ObjectDest)
		keys := dest.Keys
		props := dest.Props
		for i, k := range keys {
			prop := props[i]
			key, name := ctx.evalDestKey(prop, k)
			// If not found, desc.value holds undefined
			_, desc, _ := lookupProp(cast, key, nil)
			var value = desc.value
			if valueIsNullish(value) && prop.Default != nil {
				value = ctx.EvalExpr(prop.Default)
			}
			ctx.init(name, value, MutableDecl, p.Loc)
		}
	default:
		ctx.errorWithSource(BuildError, ctx.path, p.Loc, "Unhandled function parameter node.")
	}
	return false
}

func (ctx *EvalContext) evalAssignmentExpr(node Node) Value {
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

func (ctx *EvalContext) handleAssign(op parser.TokenTag, left Value, right Value, loc parser.Loc, lhs, rhs Node) Value {
	validatedForArithmeticAssign := false
	goto skip
errorOnArithmeticAssign:
	if left.typeof() != NumberType {
		ctx.errorWithSource(SyntaxError, ctx.path, lhs.Loc, "The left-hand side of an arithmetic operation must be of type 'number'.")
	}
	if right.typeof() != NumberType {
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
	case parser.PLUS_EQUALS:
		if left.typeof() == NumberType && right.typeof() == NumberType {
			right = ctx.evalBinaryOp(left.(Number), right.(Number), parser.PLUS, loc)
			// validatedForArithmeticAssign = true
		} else if left.typeof() == StringType || right.typeof() == StringType {
			right = NewString(left.toString().string + right.toString().string)
		} else if !validatedForArithmeticAssign {
			validatedForArithmeticAssign = true
			goto errorOnArithmeticAssign
		}
	case
		parser.MINUS_EQUALS, parser.DIVIDE_EQUALS,
		parser.TIMES_EQUALS, parser.MODULO_EQUALS,
		parser.STAR2_EQUALS:
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
		name := NewString(string(lhs.Data.(parser.Identifier)))
		ctx.assignName(name, right, &lhs.Loc)
	case parser.T_MEMBER:
		expr := lhs.Data.(parser.MemberExpr)
		// _ = expr.Computed
		// memberNode := expr.Member
		objectNode := expr.Object
		objectVal := ctx.EvalExpr(objectNode)
		memberVal := ctx.memberKey(expr)
		switch obj := objectVal.(type) {
		case *Array:
			// Array index is already validated when evaluating left.
			idx := float64(memberVal.(Number))
			if lib.Modulo(idx, 1) != 0 || idx < 0 {
				ctx.errorWithSource(TypeError, ctx.path, lhs.Loc, "Array index must be a positive integer, received negative or floating point number ", memberVal.toString().string, ".")
			}
			return obj.elements.Set(uintptr(idx), right)
		case *Object:
			return ctx.setMember(obj, obj, toValidPropKey(memberVal), right, loc)
		case *Instance:
			return ctx.setMember(obj.Object, obj, toValidPropKey(memberVal), right, loc)
		case *ScopeObject:
			name := memberVal.toString()
			if _, ok := obj.names.Get(name); ok {
				// If the name exists in scope
				obj.assignName(name, right, &lhs.Loc)
			} else {
				obj.init(name, right, MutableDecl, lhs.Loc)
			}
		case *Function:
			return ctx.setMember(obj.Object, obj, toValidPropKey(memberVal), right, loc)
		default:
			ctx.errorWithSource(BuildError, ctx.path, loc, "Cannot assign to type ", obj.typeof().string, ".")
		}
	case parser.T_ARRAY_LIT:
		if right.typeof() != ArrayType {
			ctx.errorWithSource(TypeError, ctx.path, rhs.Loc, "Cannot array destructure type ", right.typeof().string, ".")
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

func toValidPropKey(prop Value) *String {
	switch prop.typeof() {
	case StringType:
		return prop.(*String)
	case SymbolType:
		return NewString(prop.toString().string)
	}
	return NewString(prop.toString().string)
}

func (ctx *EvalContext) evalNotExpr(node Node) Value {
	operand := node.Data.(parser.OperandExpr).Operand
	val := ctx.EvalExpr(operand)
	return !valueIsTruthy(val)
}

func (ctx *EvalContext) evalIncreExpr(node Node) Value {
	expr := node.Data.(parser.IncreExpr)
	operand, pre, op := expr.Operand, expr.Pre, expr.Op
	val := ctx.EvalExpr(operand)
	if val.typeof() != NumberType {
		ctx.errorWithSource(TypeError, 0, node.Loc, "Operand of an arithmetic operation must be a number.")
	}
	n := val.(Number)
	if pre {
		switch op {
		case parser.MINUS2:
			return ctx.handleAssign(parser.PLUS_EQUALS, n, Number(-1), node.Loc, operand, operand)
		case parser.PLUS2:
			return ctx.handleAssign(parser.PLUS_EQUALS, n, Number(1), node.Loc, operand, operand)
		}
	} else {
		switch op {
		case parser.MINUS2:
			ctx.handleAssign(parser.PLUS_EQUALS, n, Number(-1), node.Loc, operand, operand)
		case parser.PLUS2:
			ctx.handleAssign(parser.PLUS_EQUALS, n, Number(1), node.Loc, operand, operand)
		}
	}
	return n
}

func (ctx *EvalContext) evalLogicalExpr(node Node) Value {
	expr := node.Data.(parser.LROpExpr)
	left, op := expr.Lhs, expr.Op
	val := ctx.EvalExpr(left)
	cond := bool(valueIsTruthy(val))
	switch {
	case op == parser.OR && cond:
		return val
	case op == parser.AND && !cond:
		return val
	}
	return ctx.EvalExpr(expr.Rhs)
}

func (ctx *EvalContext) evalNewExpr(node Node) Value {
	expr := node.Data.(parser.OperandExpr)
	operand := expr.Operand
	var classVal Value
	var args = Args{}
	switch operand.Tag {
	case parser.T_CALLEXPR:
		e := operand.Data.(parser.CallExpr)
		args = ctx.evalArgs(e.Args)
		classVal = ctx.EvalExpr(e.Caller)
	default:
		classVal = ctx.EvalExpr(operand)
	}
	var class *Class
	switch classVal := classVal.(type) {
	case *Class:
		class = classVal
	default:
		ctx.errorWithSource(TypeError, 0, node.Loc, "Cannot instantiate type ", classVal.typeof().string, ", it is not a constructor.")
	}
	extends := class.extends
	ctorFn := class.ctor
	this := NewInstance(class)
	classBody := NewContext(ctorFn.declScope)
	classBody.object = this
	ptr := classBody.names.GetS(ThisStr)
	classBody.assign(ptr, this)
	ctx.instantiate(class, classBody, this)
	if extends != nil {
		super := NewInstance(extends)
		superBody := NewContext(extends.ctor.declScope)
		superBody.object = super
		ptr := superBody.names.GetS(ThisStr)
		superBody.assign(ptr, super)
		ctx.instantiate(extends, superBody, super)
		this.proto.proto = super.Object
	}
	if ctorFn != nil {
		ctx.execCtor(ctorFn, this, args, node.Loc)
	}
	return this
}

func (*EvalContext) instantiate(class *Class, classBody *EvalContext, this *Instance) {
	defProp := class.defProp
	methods := class.methods
	props := class.props
	modifiers := class.modifiers
	if defProp.Node != nil {
		decl := defProp.Data.(parser.Decl)
		name := classBody.evalMemberKey(defProp.Computed, decl.Lhs)
		private := lib.InSlice(defProp.Modifiers, parser.K_private)
		static := lib.InSlice(defProp.Modifiers, parser.K_static)
		var rhs Value
		if decl.Rhs == nil {
			rhs = undefined
		} else {
			rhs = classBody.EvalExpr(decl.Rhs)
		}
		// NOTE: Use the struct directly to allow explanations.
		this.proto.own.Set(name, PropertyDescriptor{
			value: rhs,
			// Cannot reconfigure descriptors.
			configurable: false,
			enumerable:   true,
			writable:     true,
			public:       !private,
			getter:       nil,
			setter:       nil,
			static:       static,
		})
		this.defaultVal = rhs
	}
	for i, p := range props {
		decl := p.Data.(parser.Decl)
		name := classBody.evalMemberKey(p.Computed, decl.Lhs)
		private := lib.InSlice(modifiers[i], parser.K_private)
		static := lib.InSlice(modifiers[i], parser.K_static)
		var rhs Value
		if decl.Rhs == nil {
			rhs = undefined
		} else {
			rhs = classBody.EvalExpr(decl.Rhs)
		}
		// NOTE: Use the struct directly to allow explanations.
		this.proto.own.Set(name, PropertyDescriptor{
			value: rhs,
			// Cannot reconfigure descriptors.
			configurable: false,
			enumerable:   true,
			writable:     true,
			public:       !private,
			getter:       nil,
			setter:       nil,
			static:       static,
		})
	}
	for i, m := range methods {
		private := lib.InSlice(modifiers[i], parser.K_private)
		static := lib.InSlice(modifiers[i], parser.K_static)
		switch m.Tag {
		case parser.T_ACCESSOR:
			acc := m.Data.(parser.Accessor)
			fn := acc.Node
			accProp := classBody.evalFnExpr(fn)
			var getter, setter *Function
			if acc.Getter {
				getter = accProp
			} else {
				setter = accProp
			}
			this.proto.own.Set(
				classBody.evalMemberKey(m.Computed, fn.Data.(parser.FnDecl).Name),
				PropertyDescriptor{
					value:        undefined,
					configurable: false,
					enumerable:   true,
					// Methods and accessors are constants.
					writable: false,
					public:   !private,
					getter:   getter,
					setter:   setter,
					static:   static,
				},
			)
			// Method
		default:
			fn := classBody.evalFnExpr(m.Node)
			nameNode := m.Data.(parser.FnDecl).Name
			this.proto.own.Set(classBody.evalMemberKey(m.Computed, nameNode), PropertyDescriptor{
				value:        fn,
				configurable: false,
				enumerable:   true,
				writable:     false,
				public:       !private,
				getter:       nil,
				setter:       nil,
				static:       static,
			})
		}
	}
}

func (ctx *EvalContext) execCtor(ctorFn *Function, this Value, args Args, loc parser.Loc) {
	ctorFn.this = this
	ctx.execFrame(ctorFn, args, loc)
}

func (ctx *EvalContext) evalSuperExpr(node Node) *Undefined {
	class, this := ctx.locateClassForSuper(ctx.Scope, node.Loc)
	args := ctx.evalArgs(node.Children)
	ctx.execCtor(class.ctor, this, args, node.Loc)
	return undefined
}

func (ctx *EvalContext) locateClassForSuper(scope *Scope, loc parser.Loc) (*Class, *Instance) {
	var class *Class
	if scope == nil || scope.of == FunctionScope {
		ctx.errorWithSource(SyntaxError, 0, loc, "Super calls are not permitted outside constructors or in nested functions inside constructors.")
	}
	if scope.of != ClassScope {
		return ctx.locateClassForSuper(scope.parent, loc)
	}
	this := scope.object.(*Instance)
	class = this.class
	if class.extends == nil {
		ctx.errorWithSource(SyntaxError, 0, loc, "'super' can only be referenced in a derived class.")
	}
	return class, this
}

func (ctx *EvalContext) evalInExpr(node Node) Boolean {
	expr := node.Data.(parser.LROpExpr)
	lNode := expr.Lhs
	rNode := expr.Rhs
	lhs := ctx.EvalExpr(lNode)
	rhs := ctx.EvalExpr(rNode)
	key := toValidPropKey(lhs)
	var object *Object
	switch v := rhs.(type) {
	case *Array, Boolean, *Macro, Number, *RAW, *String, *Symbol, *Undefined:
		ctx.errorWithSource(TypeError, 0, node.Loc, lib.Sprintf("Cannot use 'in' operator to search for '%s' in %s", key.string, rhs.toString().string))
	case *Class:
		object = v.Object
	case *Function:
		object = v.Object
	case *Instance:
		object = v.Object
	case *Object:
		object = v
	case *ScopeObject:
		return Boolean(v.names.Has(key))
	}
	_, _, f := lookupProp(object, key, nil)
	return Boolean(f)
}
