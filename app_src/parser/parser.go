package parser

import "aspire/are/main/lib"

// var debugMode = false

// type Cache struct {
// 	ParserVersion string
// 	Hash          string
// 	SourceModTime int
// 	AST           Node
// }

type Loc = lib.Loc

type root struct {
	Imports []string
	Program *Program
}

type Parser struct {
	path       int
	lexer      *Lexer
	tokens     []Token
	tokenIndex uint64
	OwnModule  *Module
	builtin    bool
	Invalid    bool

	*root
}

func newRoot() *root {
	return &root{
		Program: &Program{
			Main:    Module{},
			Modules: []*Module{},
		},
		Imports: []string{},
	}
}

func init() {
	defStmt(_const, parseVarDeclStmt)
	defStmt(_let, parseVarDeclStmt)
	defStmt(_var, parseVarDeclStmt)
	defStmt(_if, parseIfStmt)
	defStmt(O_BRACE, parseBlockStmt)
	defStmt(LABEL, parseLabel)
	defStmt(_throw, parseThrowStmt)
	defStmt(_fn, parseFunDecl)
	defStmt(_async, parseFunDecl)
	defStmt(_return, parseReturnStmt)
	defStmt(_while, parseWhileStmt)
	defStmt(_do, parseDoWhileStmt)
	defStmt(_for, parseForStmt)
	defStmt(_switch, parseSwitchStmt)
	defStmt(_yield, parse_yield_stmt)
	defStmt(_import, parse_import_stmt)
	defStmt(_export, parse_export_stmt)
	defStmt(_class, parse_class_stmt)
	defStmt(_try, parse_try_catch_finally)
	defStmt(_fallthrough, parse_fallthrough)
	defStmt(_break, parse_break_stmt)
	defStmt(_continue, parse_continue_stmt)

	defNud(IDENTIFIER, parse_ident)
	defNud(STRING, parse_string)
	defNud(NUMBER, parse_number)
	defNud(NOT, parse_not_expr)
	defNud(O_PAREN, parse_open_paren)

	defNud(_new, parse_new_expr)
	defNud(_void, parse_void_expr)
	defNud(_async, parse_async_fn_expr)
	defNud(_typeof, parse_typeof_expr)
	defNud(_match, parse_match_expr)
	defNud(_from, parse_from_expr)
	defNud(_class, parse_class_expr)
	defNud(_gt, parse_globalthis)
	defNud(_super, parse_super_expr)
	defNud(_await, parse_await_expr)

	defNud(O_BRACE, parse_object_lit)
	defNud(O_BRACKET, parse_array_lit)

	defNud(PLUS2, parse_plus2_prefix)
	defNud(MINUS2, parse_minus2_prefix)

	defNud(BITNOT, parse_bitnot_expr)
	defNud(MINUS, parse_unary_expr)

	defLed(PLUS, parse_binary_expr, ADDITIVE_BP)
	defLed(MINUS, parse_binary_expr, ADDITIVE_BP)
	defLed(TIMES, parse_binary_expr, MULTIPLICATIVE_BP)
	defLed(STAR2, parse_binary_expr, MULTIPLICATIVE_BP)
	defLed(MODULO, parse_binary_expr, MULTIPLICATIVE_BP)
	defLed(DIVIDE, parse_binary_expr, MULTIPLICATIVE_BP)

	defLed(PLUS2, parse_plus2_postfix, POSTFIX_BP)
	defLed(MINUS2, parse_minus2_postfix, POSTFIX_BP)

	defLed(EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(PLUS_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(MINUS_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(TIMES_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(STAR2_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(MODULO_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(DIVIDE_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)
	defLed(NULLISH_EQUALS, parse_assignment_expr, ASSIGNMENT_BP)

	defLed(OR, parse_logical_expr, LOGICAL_OR_BP)
	defLed(AND, parse_logical_expr, LOGICAL_AND_BP)

	defLed(GT, parse_comparison_expr, COMPARISON_BP)
	defLed(LT, parse_comparison_expr, COMPARISON_BP)
	defLed(GT_EQUALS, parse_comparison_expr, COMPARISON_BP)
	defLed(LT_EQUALS, parse_comparison_expr, COMPARISON_BP)
	defLed(EQUALS2, parse_comparison_expr, EQUALITY_BP)
	defLed(EQUALS3, parse_comparison_expr, EQUALITY_BP)
	defLed(NOT_EQUALS, parse_comparison_expr, EQUALITY_BP)
	defLed(NOT_EQUALS2, parse_comparison_expr, EQUALITY_BP)

	defLed(O_PAREN, parse_call_expr, CALL_BP)
	defLed(DOT, parse_member_expr, MEMBER_BP)
	defLed(O_BRACKET, parse_member_expr, MEMBER_BP)

	defLed(QUESTION, parse_ternary_expr, TERNARY_BP)

	defLed(BITOR, parse_bitwise_expr, BITOR_BP)
	defLed(BITAND, parse_bitwise_expr, BITAND_BP)
	defLed(SHBITL, parse_bitwise_expr, SHIFT_BP)
	defLed(SHBITR, parse_bitwise_expr, SHIFT_BP)
	defLed(XOR, parse_bitwise_expr, BITXOR_BP)
	defLed(BITCLEAR, parse_bitwise_expr, BITCLEAR_BP)
	defLed(BITOR_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)
	defLed(BITAND_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)
	defLed(SHBITL_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)
	defLed(SHBITR_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)
	defLed(XOR_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)
	defLed(BITCLEAR_EQUALS, parse_bitwise_expr, ASSIGNMENT_BP)

	defLed(_instanceof, parse_instanceof_expr, COMPARISON_BP)

	defLed(NULLISH, parse_nullish_expr, NULLISH_COALLESCE_BP)
}

func NewBuiltinsParser(name string) *Parser {
	return newBuiltinsParser(name, newRoot())
}

func newBuiltinsParser(name string, root *root) *Parser {
	path := "stdlib/" + name
	var bytes []byte
	var key int
	if lib.DEBUG_MODE {
		// const PATH = "C:\\Users\\ecmak\\ArachnoScript\\ARE\\v0.2.5\\app_src\\lib\\stdlib\\"
		// b, k, err := lib.ReadFile(PATH + name)
		b, k, err := lib.ReadFile(
			lib.JoinPaths(lib.DirOf(lib.ExecPath()), "..", "app_src", "lib", "stdlib", name),
		)
		if err != nil {
			lib.Panic(err)
		}
		bytes = b
		key = k
	} else {
		// path := lib.JoinPaths("stdlib", name)
		b, err := lib.EMBEDFS.ReadFile(path)
		if err != nil {
			lib.Panic(err)
		}
		bytes = b
		key = lib.WriteCacheFile(bytes, path)
	}
	if !lib.DEBUG_MODE {
		path = lib.AnonymousPath
	}
	lexer := NewLexer(bytes, key)
	tokens := lexer.Tokenize()
	module := &Module{
		Path: key,
		Body: []Node{},
	}
	root.Program.Modules = append(root.Program.Modules, module)
	return &Parser{
		path:       key,
		lexer:      lexer,
		tokens:     tokens,
		tokenIndex: 0,
		OwnModule:  module,
		root:       root,
		builtin:    true,
		Invalid:    false,
	}
}

var AnonymousCount = 0

// path: represents the path to the module to parse.
// if the path does not exist, the parser will attempt to parse 'path'.
// handle this if it is unwanted behavior.
func NewParser(path string, isRepl bool) *Parser {
	return newParser(path, isRepl, newRoot())
}

func newParser(path string, isRepl bool, root *root) *Parser {
	defer func() { recover() }()
	abs := path
	if !lib.IsAbs(path) {
		abs = lib.Abs(path)
	}
	var pk int
	var buffer []byte
	if lib.PathExists(abs) {
		b, k, err := lib.ReadFile(abs)
		if err != nil {
			(&Parser{}).parseError(PathError, err.Error())
		}
		buffer = b
		pk = k
	} else {
		AnonymousCount--
		pk = AnonymousCount
		buffer = []byte(path)
		lib.WriteAnonymousFile(buffer, pk)
	}
	lexer := NewLexer(buffer, pk)
	tokens := lexer.Tokenize()
	p := &Parser{
		path:       pk,
		lexer:      lexer,
		tokens:     tokens,
		tokenIndex: 0,
		root:       root,
		OwnModule:  &Module{},
		builtin:    false,
		Invalid:    false,
	}
	if isRepl {
		root.Program.Modules = append(root.Program.Modules, p.OwnModule)
	} else {
		p.OwnModule = &root.Program.Main
		p.OwnModule.Path = pk
	}
	return p
}

func (p *Parser) Parse() {
	defer func() { p.Invalid = recover() != nil }()
	for p.not_eof() {
		if p.eatSemiColons() {
			continue
		}
		p.OwnModule.Body = append(p.OwnModule.Body, p.parseStmt())
	}
}

func (p *Parser) expect(t TokenTag) Token {
	if p.notAt(t) {
		p.parseError(SyntaxError, lib.Sprintf("%s\x1b[36m'%s'\x1b[0m.", "Expected a token of type ", t.Lexeme()))
	}
	return p.eat()
}

func (p *Parser) eatSemiColons() bool {
	if p.isAt(SEMICOLON) {
		for p.isAt(SEMICOLON) {
			p.next()
		}
		return true
	}
	return false
}

func (p *Parser) eatSemiColon() bool {
	if p.isAt(SEMICOLON) {
		p.next()
		return true
	}
	return false
}

func (p *Parser) eatComma() bool {
	if p.isAt(COMMA) {
		p.eat()
		return true
	}
	return false
}

// @##############################################@
// @##############################################@
// @################# Statements #################@
// @##############################################@
// @##############################################@
func (p *Parser) parseStmt() Node {
	if handler, ok := stmtLU[p.atType(0)]; ok {
		defer p.eatSemiColons()
		return handler(p)
	}
	return p.parseExprStmt()
}

func parseFunDecl(p *Parser) Node {
	var isAsync = false
	if p.isAt(_async) {
		p.next()
		isAsync = true
	}
	loc := p.expect(_fn).loc // function
	nameTk := p.expect(IDENTIFIER)
	return parseFnFromParams(p, StringNode(p.src(nameTk), nameTk.loc), isAsync, false, loc)
}

const AnonymousName = "(anonymous)"

func parseFunExpr(p *Parser) Node {
	isAsync := false
	start := p.currLoc()
	if p.isAt(_async) {
		p.next()
		isAsync = true
	}
	p.expect(_fn)
	// Function expression is Anonymous
	return parseFnFromParams(p, StringNode(AnonymousName, lib.DumbyLoc), isAsync, true, start)
}

func parseFunMethod(p *Parser) Node {
	isAsync := false
	idx := p.tokenIndex
	if p.isAt(_async) {
		p.next()
		isAsync = true
	}
	if p.isAt(ACCESSOR) {
		tk := p.eat()
		getter := tk.tag == _get
		if isAsync {
			p.tokenIndex = idx
			p.parseError(SyntaxError, "'async' modifier cannot be used here.")
		}
		return Node{
			Tag:      T_ACCESSOR,
			Children: nil,
			Data: Accessor{
				Getter: getter,
				Node:   parseMethod(p, isAsync),
			},
			Loc: tk.loc,
		}
	}
	return parseMethod(p, isAsync)
}

func parseMethod(p *Parser, isAsync bool) Node {
	var NameNode Node
	switch p.atType(0) {
	case O_BRACKET:
		p.next() // [
		NameNode = p.parseExpr(DEFAULT_BP)
		p.expect(C_BRACKET)
	case IDENTIFIER, STRING, NUMBER:
		NameNode = parse_ident(p)
	default:
		p.expect(IDENTIFIER)
	}
	// Method cannot be Anonymous
	return parseFnFromParams(p, NameNode, isAsync, false, NameNode.Loc)
}

func parseFnFromParams(p *Parser, NameNode Node, isAsync, isAnony bool, start Loc) Node {
	params := p.parse_params()
	body := p.parseBlock()
	return Node{
		Tag:      T_FNDECL,
		Children: body,
		Data: FnDecl{
			Params:    params,
			Name:      NameNode,
			Async:     isAsync,
			Arrow:     false,
			Anonymous: isAnony,
		},
		Loc: start,
	}
}

func parseReturnStmt(p *Parser) Node {
	end := p.expect(_return).loc // return
	var value Node
	if !p.eatSemiColons() && p.notAt(C_BRACE) {
		value = p.parseExpr(DEFAULT_BP)
	}
	return Node{
		Tag:      T_RETURN,
		Children: []Node{value},
		Data:     nil,
		Loc:      end,
	}
}

func parseThrowStmt(p *Parser) Node {
	start := p.expect(_throw).loc // throw
	value := p.parseExpr(DEFAULT_BP)
	return Node{
		Tag:      T_THROW,
		Children: []Node{value},
		Data:     nil,
		Loc:      p.loc(start, value.Loc),
	}
}

func parseWhileStmt(p *Parser) Node {
	start := p.expect(_while).loc // while
	condition := p.parseCondition()
	end := condition.Loc
	var body []Node
	if !p.eatSemiColon() {
		body = p.parseBlockOrStmt()
	}
	return Node{
		Tag:      T_WHILE,
		Children: nil,
		Data: WhileLoop{
			Body:      body,
			Condition: condition,
			Do:        false,
		},
		Loc: p.loc(start, end),
	}
}

func parseForStmt(p *Parser) Node {
	start := p.expect(_for).loc // 'for'
	p.expect(O_PAREN)
	tag := T_TFORLOOP
	var before_lhs Node
	var condition Node
	var after_rhs Node
	op := 0
	var data any = TForLoop{}
	parseTForLoop := func() {
		if p.isAt(VARDECL) {
			before_lhs = parseVarDecl(p)
		} else {
			before_lhs = p.parseExpr(DEFAULT_BP)
		}
		p.expect(SEMICOLON)
		condition = p.parseExpr(DEFAULT_BP)
		p.expect(SEMICOLON)
		after_rhs = p.parseExpr(DEFAULT_BP)
	}
	if p.isAt(VARDECL) {
		initIdx := p.tokenIndex
		declKind := handleDeclKeyword(p)
		p.next()
		lhs := p.parseVarDeclLhs()
		if p.isAt(_in) || p.isAt(_of) {
			if p.eat().tag == _in {
				op = In_Op
			} else {
				op = Of_Op
			}
			after_rhs = p.parseExpr(DEFAULT_BP)
			data = ForLoop{
				LHS:      lhs,
				RHS:      after_rhs,
				Op:       op,
				DeclKind: declKind,
			}
			tag = T_FORLOOP
		} else {
			p.tokenIndex = initIdx
			parseTForLoop()
			data = TForLoop{
				Before:    before_lhs,
				Condition: condition,
				AfterExec: after_rhs,
			}
		}
	} else {
		parseTForLoop()
		data = TForLoop{
			Before:    before_lhs,
			Condition: condition,
			AfterExec: after_rhs,
		}
	}
	end := p.expect(C_PAREN).loc
	var body []Node
	if !p.eatSemiColon() {
		body = p.parseBlockOrStmt()
	}
	return Node{
		Tag:      tag,
		Children: body,
		Data:     data,
		Loc:      p.loc(start, end),
	}
}

func parseDoWhileStmt(p *Parser) Node {
	start := p.expect(_do).loc // do
	var body = p.parseBlockOrStmt()
	p.expect(_while)
	condition := p.parseCondition()
	return Node{
		Tag:      T_WHILE,
		Children: nil,
		Data: WhileLoop{
			Do:        true,
			Body:      body,
			Condition: condition,
		},
		Loc: p.loc(start, condition.Loc),
	}
}

func parseVarDecl(p *Parser) Node {
	declType := handleDeclKeyword(p)
	startLoc := p.expect(VARDECL).loc
	lhs := p.parseVarDeclLhs()
	const ConstMustBeInitialized = "A 'const' declaration must be initialized!"
	rhs := p.parseDeclRhs(declType, ConstMustBeInitialized)
	decls := []Decl{{
		Lhs: lhs,
		Rhs: rhs,
	}}
	for p.eatComma() {
		lhs := p.parseVarDeclLhs()
		rhs := p.parseDeclRhs(declType, ConstMustBeInitialized)
		decls = append(decls, Decl{
			Lhs: lhs,
			Rhs: rhs,
		})
	}
	endLoc := p.currLoc()
	return Node{
		Data: VarDecl{
			Kind:  declType,
			Decls: decls,
		},
		Loc: p.loc(startLoc, endLoc),
		Tag: T_VARDECL,
	}
}

func handleDeclKeyword(p *Parser) DeclKind {
	declType := MutableDecl
	switch p.atType(0) {
	case _const:
		declType = ConstantDecl
	case _var:
		declType = HoistedDecl
	case _let:
		break
	default:
		p.parseError(SyntaxError, "Expected a declaration keyword.")
	}
	return declType
}

func parseVarDeclStmt(p *Parser) Node {
	defer p.eatSemiColon()
	return parseVarDecl(p)
}

func (p *Parser) parseDeclRhs(declType DeclKind, errMsg string) Node {
	var rhs Node
	if declType == ConstantDecl && p.notAt(EQUALS) {
		p.parseError(SyntaxError, errMsg)
	} else if p.isAt(EQUALS) {
		p.next()
		rhs = p.parseExpr(DEFAULT_BP)
	}
	return rhs
}

func (p *Parser) parseVarDeclLhs() Node {
	switch p.atType(0) {
	case IDENTIFIER:
		return parse_ident(p)
	case O_BRACE:
		return p.parse_object_destructuring()
	case O_BRACKET:
		return p.parse_array_destructuring()
	default:
		p.parseError(SyntaxError, "Expected identifier on Left-Hand-Side of Variable Declaration.")
		return Node{}
	}
}

func (p *Parser) parse_array_destructuring() Node {
	start := p.expect(O_BRACKET).loc
	elements := []Node{}
	for p.not_eof() && p.notAt(C_BRACKET) {
		if p.isAt(DOT3) {
			loc := p.eat().loc
			elements = append(elements, Node{
				Tag:      T_RESTORSPREAD,
				Children: nil,
				Data:     parse_ident(p),
				Loc:      loc,
			})
		} else {
			elements = append(elements, parse_ident(p))
		}
		if !p.eatComma() && p.notAt(C_BRACKET) {
			p.parseError(SyntaxError, "',' expected.")
		}
	}
	end := p.expect(C_BRACKET).loc
	return Node{
		Tag:      T_ARRAY_LIT,
		Children: elements,
		Data:     nil,
		Loc:      p.loc(start, end),
	}
}

func (p *Parser) parse_object_destructuring() Node {
	start := p.expect(O_BRACE).loc
	props := map[NodeIndex]ObjectDestProp{}
	keys := []Node{}
	for p.not_eof() && p.notAt(C_BRACE) {
		var value, def Node
		var computed bool = false
		idx := len(keys)
		switch p.atType(0) {
		case IDENTIFIER:
			ident := parse_ident(p)
			keys = append(keys, ident)
			if p.notAt(COLON) {
				value = ident
			} else {
				p.next() // :
				value = parse_ident(p)
			}
		case STRING, NUMBER:
			key := p.parseExpr(PRIMARY_BP)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				value = parseFnFromParams(p, key, false, false, key.Loc)
			} else {
				p.expect(COLON)
				value = parse_ident(p)
			}
		case O_BRACKET:
			p.next() // [
			key := p.parseExpr(DEFAULT_BP)
			p.expect(C_BRACKET)
			keys = append(keys, key)
			p.expect(COLON)
			value = parse_ident(p)
			computed = true
		default:
			p.parseError(SyntaxError, "Identifier expected.")
		}
		if p.isAt(EQUALS) {
			p.next()
			def = p.parseExpr(DEFAULT_BP)
		}
		props[idx] = ObjectDestProp{
			Computed: computed,
			Node:     value,
			Default:  def,
		}
		if !p.eatComma() && p.notAt(C_BRACE) {
			p.parseError(SyntaxError, "Expected a ',' in object literal.")
		}
	}
	end := p.expect(C_BRACE).loc
	return Node{
		Tag:      T_OBJECT_LIT,
		Children: nil,
		Data: ObjectDest{
			Props: props,
			Keys:  keys,
		},
		Loc: p.loc(start, end),
	}
}

func parseIfStmt(p *Parser) Node {
	loc := p.expect(_if).loc
	condition := p.parseCondition()
	body := p.parseBlockOrStmt()
	var elseBlock []Node
	if p.isAt(_else) {
		p.next()
		elseBlock = p.parseBlockOrStmt()
	}
	return Node{
		Tag:      T_IFSTMT,
		Children: body,
		Data: IfStmt{
			Condition: condition,
			ElseBlock: elseBlock,
		},
		Loc: loc,
	}
}

func (p *Parser) parseCondition() Node {
	p.expect(O_PAREN)
	defer p.expect(C_PAREN)
	if p.isAt(VARDECL) {
		return parseVarDecl(p)
	}
	return p.parseExpr(DEFAULT_BP)
}

func parseBlockStmt(p *Parser) Node {
	loc := p.currLoc()
	return Node{
		Tag:      T_BLOCK,
		Children: p.parseBlock(),
		Data:     nil,
		Loc:      loc,
	}
}

func (p *Parser) parseBlock() []Node {
	block := []Node{}
	p.expect(O_BRACE)
	for p.not_eof() && p.notAt(C_BRACE) {
		block = append(block, p.parseStmt())
	}
	p.expect(C_BRACE)
	return block
}

func (p *Parser) parseBlockOrStmt() []Node {
	if p.isAt(O_BRACE) {
		return p.parseBlock()
	}
	return []Node{p.parseStmt()}
}

func parseLabel(p *Parser) Node {
	loc := p.currLoc()
	// remove the preceding '@'
	l := p.src(p.expect(LABEL))[1:] // label
	return Node{
		Tag:      T_LABEL,
		Children: nil,
		Data:     l,
		Loc:      loc,
	}
}

func parseSwitchStmt(p *Parser) Node {
	loc := p.expect(_switch).loc
	condition := p.parseCondition()
	p.expect(O_BRACE)
	cases := map[NodeIndex][]Node{}
	matches := []Node{}
	var defaultCase []Node
	for p.not_eof() && p.notAt(C_BRACE) {
		idx := len(matches)
		if p.isAt(_default) {
			p.next()
			p.expect(COLON)
			body := p.parseBlockOrStmt()
			defaultCase = body
		} else {
			p.expect(_case)
			match := p.parseExpr(DEFAULT_BP)
			p.expect(COLON)
			body := p.parseBlockOrStmt()
			matches = append(matches, match)
			cases[idx] = body
		}
	}
	p.expect(C_BRACE)
	return Node{
		Tag:      T_SWITCHSTMT,
		Children: nil,
		Data: SwitchStmt{
			Condition: condition,
			Cases:     cases,
			Matches:   matches,
			Default:   defaultCase,
		},
		Loc: loc,
	}
}

func parse_yield_stmt(p *Parser) Node {
	loc := p.expect(_yield).loc
	return Node{
		Tag:      T_YIELDSTMT,
		Children: nil,
		Data:     OperandExpr{p.parseExpr(DEFAULT_BP)},
		Loc:      p.loc(loc, p.currLoc()),
	}
}

func parse_import_stmt(p *Parser) Node {
	loc := p.expect(_import).loc
	imp_path := "invalid-path.as"
	ns := ""
	ucc := false
	var named Node
	switch p.atType(0) {
	case STRING:
		ucc = true
		imp_path = p.parseScriptConcurrent(parse_string(p).Data.(string))
	case IDENTIFIER:
		ns = string(parse_ident(p).Data.(Identifier))
		if p.isAt(COMMA) {
			p.next()
			named = p.parse_object_destructuring()
		}
		imp_path = parse_from_expr(p).Data.(FromExpr).Path
	case O_BRACE:
		named = p.parse_object_destructuring()
		imp_path = parse_from_expr(p).Data.(FromExpr).Path
	default:
		p.parseError(SyntaxError, "'from' keyword expected.")
	}
	return Node{
		Tag:      T_IMPORT,
		Children: nil,
		Data: ImportStmt{
			Namespace:         ns,
			Named:             named,
			From:              imp_path,
			UseCurrentContext: ucc,
		},
		Loc: loc,
	}
}

func parse_export_stmt(p *Parser) Node {
	loc := p.expect(_export).loc
	var export Node
	if p.isAt(DECL) {
		export = p.parseStmt()
	} else if p.isAt(O_BRACE) {
		export = p.parse_object_destructuring()
	} else {
		p.parseError(SyntaxError, "Declaration expected.")
	}
	return Node{
		Tag:      T_EXPORT,
		Children: nil,
		Data:     ExportStmt{export},
		Loc:      loc,
	}
}

func parse_class_stmt(p *Parser) Node {
	loc := p.expect(_class).loc
	name := string(parse_ident(p).Data.(Identifier))
	return parse_class_decl(p, name, false, loc)
}

func parse_class_decl(p *Parser, name string, isAnony bool, loc lib.Loc) Node {
	methods := []Node{}
	props := []Node{}
	var defaultProp, constructor, extends Node
	if p.isAt(_extends) {
		p.next()
		extends = p.parseExpr(ASSIGNMENT_BP)
	}
	p.expect(O_BRACE)
	for p.not_eof() && p.notAt(C_BRACE) {
		start := p.currLoc()
		switch p.atType(0) {
		case _ctor:
			ctor := p.eat()
			// Constructor's name is the class name.
			constructor = parseFnFromParams(p, StringNode(p.src(ctor), ctor.loc), false, false, start)
		default:
			// TODO: add class prop and method nodes to node.go to add modifiers.
			modifiers := p.parse_modifiers()
			if lib.InSlice(modifiers, _default) {
				defaultProp = p.parse_class_prop()
			} else if p.isAt(_async) {
				methods = append(methods, parseFunMethod(p))
			} else {
				var NameNode Node
				switch p.atType(0) {
				case O_BRACKET:
					p.next() // [
					NameNode = p.parseExpr(DEFAULT_BP)
					p.expect(C_BRACKET)
				case IDENTIFIER, STRING, NUMBER:
					NameNode = parse_ident(p)
				default:
					p.expect(IDENTIFIER)
				}
				if p.isAt(O_PAREN) {
					methods = append(methods, parseFnFromParams(p, NameNode, false, false, start))
				} else {
					if p.isAt(EQUALS) {
						p.next()
						rhs := p.parseExpr(DEFAULT_BP)
						props = append(props, Node{
							Tag:      T_VARDECL,
							Children: nil,
							Data: Decl{
								Lhs: NameNode,
								Rhs: rhs,
							},
							Loc: p.loc(start, rhs.Loc),
						})
					}
				}
			}
		}
		p.eatSemiColons()
	}
	p.expect(C_BRACE)
	return Node{
		Tag:      T_CLASSDECL,
		Children: nil,
		Data: ClassDecl{
			DefaultProp: defaultProp,
			Methods:     methods,
			Props:       props,
			Constructor: constructor,
			Extends:     extends,
			Name:        name,
			Anonymous:   isAnony,
		},
		Loc: loc,
	}
}

func (p *Parser) parse_class_prop() Node {
	var lhs, rhs Node
	start := p.currLoc()
	switch p.atType(0) {
	case STRING, NUMBER, IDENTIFIER:
		lhs = p.parseExpr(PRIMARY_BP)
	case O_BRACKET:
		p.next() // [
		lhs = p.parseExpr(DEFAULT_BP)
		p.expect(C_BRACKET)
	default:
		p.parseError(SyntaxError, "Unexpected token. A constructor, method, accessor, or property was expected.")
	}
	if p.isAt(EQUALS) {
		p.next()
		rhs = p.parseExpr(DEFAULT_BP)
	}
	return Node{
		Tag:      T_VARDECL,
		Children: nil,
		Data: Decl{
			Lhs: lhs,
			Rhs: rhs,
		},
		Loc: p.loc(start, rhs.Loc),
	}
}

func (p *Parser) parse_modifiers() (modifiers []TokenTag) {
	if p.isAt(_private) || p.isAt(_public) {
		modifiers = append(modifiers, p.eat().tag)
	}
	if p.isAt(_static) {
		modifiers = append(modifiers, p.eat().tag)
	}
	if p.isAt(_default) {
		modifiers = append(modifiers, p.eat().tag)
	}
	return
}

func parse_try_catch_finally(p *Parser) Node {
	loc := p.expect(_try).loc
	try := p.parseBlock()
	var catch, finally []Node
	var capture Node
	if p.isAt(_catch) {
		p.next()
		if p.isAt(O_PAREN) {
			p.next()
			capture = p.parse_param()
			p.expect(C_PAREN)
		}
		catch = p.parseBlock()
	}
	if p.isAt(_finally) {
		p.next()
		finally = p.parseBlock()
	}
	return Node{
		Tag:      T_TRYCATCH,
		Children: nil,
		Data: TryCatch{
			try:     try,
			catch:   catch,
			capture: capture,
			finally: finally,
		},
		Loc: loc,
	}
}

func parse_fallthrough(p *Parser) Node {
	return Node{
		Tag:      T_FALLTHROUGH,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(_fallthrough).loc,
	}
}

func parse_break_stmt(p *Parser) Node {
	return Node{
		Tag:      T_BREAKSTMT,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(_break).loc,
	}
}

func parse_continue_stmt(p *Parser) Node {
	return Node{
		Tag:      T_CONTINUESTMT,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(_continue).loc,
	}
}

// @#############################################@
// @#############################################@
// @################ Expressions ################@
// @#############################################@
// @#############################################@

func (p *Parser) parseExprStmt() Node {
	defer p.eatSemiColon()
	return p.parseExpr(DEFAULT_BP)
}

func (p *Parser) parseExpr(bp BindingPower) Node {
	// find a nud handler for the current token
	nud, ok := nudLU[p.atType(0)]
	if !ok {
		// raise an error if not found
		p.parseError(UnexpectedToken)
	}
	// parse token(s)...
	left := nud(p)

	for p.not_eof() && bpLU[p.atType(0)] > bp {
		led := ledLU[p.atType(0)]
		left = led(p, left, bpLU[p.atType(0)])
	}

	return left
}

func (p *Parser) parse_exprs(valid_tags ...NodeTag) Node {
	tokenIndex := p.tokenIndex
	expr := p.parseExpr(DEFAULT_BP)
	if lib.InSlice(valid_tags, expr.Tag) {
		return expr
	}
	p.tokenIndex = tokenIndex
	p.parseError(SyntaxError, lib.Sprintf("Unexpected expression: %s", expr.Tag.Name()))
	return Node{}
}

func (p *Parser) parse_param() Node {
	switch p.atType(0) {
	case DOT3:
		start := p.eat().loc
		param := p.parse_exprs(T_IDENT, T_ARRAY_LIT)
		return Node{
			Tag:      T_RESTORSPREAD,
			Children: nil,
			Data:     param,
			Loc:      p.loc(start, param.Loc),
		}
	case TIMES:
		lib.Panic("language feature unimplemented.")
		return Node{}
	case O_BRACE:
		return p.parse_object_destructuring()
	case O_BRACKET:
		return p.parse_array_destructuring()
	default:
		return parse_ident(p)
	}
}

func (p *Parser) parse_params() []Node {
	p.expect(O_PAREN)
	args := []Node{}
	for p.not_eof() && p.notAt(C_PAREN) {
		args = append(args, p.parse_param())
		if !p.eatComma() && p.notAt(C_PAREN) {
			p.expect(COMMA)
		}
	}
	p.expect(C_PAREN)
	return args
}

// parse_args, parse_grouping_expr
func (p *Parser) parse_expr_list() []Node {
	p.expect(O_PAREN)
	exprs := []Node{}
	for p.not_eof() && p.notAt(C_PAREN) {
		exprs = append(exprs, p.parseExpr(DEFAULT_BP))
		if !p.eatComma() && p.notAt(C_PAREN) {
			p.expect(COMMA)
		}
	}
	p.expect(C_PAREN)
	return exprs
}

func parse_globalthis(p *Parser) Node {
	loc := p.expect(_gt).loc
	return Node{
		Tag:      T_GT,
		Children: nil,
		Data:     nil,
		Loc:      loc,
	}
}

func parse_await_expr(p *Parser) Node {
	start := p.expect(_await).loc
	operand := p.parseExpr(ASSIGNMENT_BP)
	return Node{
		Tag:      T_AWAIT,
		Children: nil,
		Data:     OperandExpr{operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_super_expr(p *Parser) Node {
	start := p.expect(_super).loc
	params := p.parse_expr_list()
	return Node{
		Tag:      T_SUPER,
		Children: params,
		Data:     nil,
		Loc:      p.loc(start, p.currLoc()),
	}
}

func parse_class_expr(p *Parser) Node {
	loc := p.expect(_class).loc
	return parse_class_decl(p, "", true, loc)
}

func parse_nullish_expr(p *Parser, operand Node, bp BindingPower) Node {
	p.next() // ??
	right := p.parseExpr(bp)
	return Node{
		Tag:      T_NULLISH,
		Children: nil,
		Data: LROpExpr{
			Lhs: operand,
			Rhs: right,
		},
		Loc: p.loc(operand.Loc, right.Loc),
	}
}

func parse_instanceof_expr(p *Parser, operand Node, bp BindingPower) Node {
	p.next()
	right := p.parseExpr(bp)
	return Node{
		Tag:      T_INSTANCEOF,
		Children: nil,
		Data: LROpExpr{
			Lhs: operand,
			Rhs: right,
			Op:  0,
		},
		Loc: p.loc(operand.Loc, right.Loc),
	}
}

func parse_bitwise_expr(p *Parser, operand Node, bp BindingPower) Node {
	var nodeTag NodeTag
	tkTag := p.eat().tag
	switch tkTag {
	case XOR:
		nodeTag = T_XOR
	case BITAND:
		nodeTag = T_BITAND
	case BITOR:
		nodeTag = T_BITOR
	case SHBITL:
		nodeTag = T_SHBITL
	case SHBITR:
		nodeTag = T_SHBITR
	case BITCLEAR:
		nodeTag = T_BITCLEAR
	case XOR_EQUALS,
		BITAND_EQUALS,
		BITOR_EQUALS,
		SHBITL_EQUALS,
		SHBITR_EQUALS,
		BITCLEAR_EQUALS:
		nodeTag = T_ASSIGN
	default:
		lib.Panic("unhandled tag.")
	}
	right := p.parseExpr(bp)
	return Node{
		Tag:      nodeTag,
		Children: nil,
		Data: LROpExpr{
			Lhs: operand,
			Rhs: right,
			Op:  tkTag,
		},
		Loc: p.loc(operand.Loc, right.Loc),
	}
}

func parse_unary_expr(p *Parser) Node {
	return parse_operand_expr(p, MINUS, UNARY_BP, T_UNARY)
}

func parse_bitnot_expr(p *Parser) Node {
	return parse_operand_expr(p, BITNOT, UNARY_BP, T_BITNOT)
}

func parse_plus2_prefix(p *Parser) Node {
	start := p.eat().loc // ++
	operand := p.parseExpr(UNARY_BP)
	return Node{
		Tag:      T_INCREEXPR,
		Children: nil,
		Data: IncreExpr{
			Operand: operand,
			Pre:     true,
			Op:      PLUS2,
		},
		Loc: p.loc(start, operand.Loc),
	}
}

func parse_minus2_prefix(p *Parser) Node {
	start := p.eat().loc // --
	operand := p.parseExpr(UNARY_BP)
	return Node{
		Tag:      T_INCREEXPR,
		Children: nil,
		Data: IncreExpr{
			Operand: operand,
			Pre:     true,
			Op:      MINUS2,
		},
		Loc: p.loc(start, operand.Loc),
	}
}

func parse_plus2_postfix(p *Parser, operand Node, bp BindingPower) Node {
	end := p.eat().loc // ++
	return Node{
		Tag:      T_INCREEXPR,
		Children: nil,
		Data: IncreExpr{
			Operand: operand,
			Pre:     false,
			Op:      PLUS2,
		},
		Loc: p.loc(operand.Loc, end),
	}
}

func parse_minus2_postfix(p *Parser, operand Node, bp BindingPower) Node {
	end := p.eat().loc // --
	return Node{
		Tag:      T_INCREEXPR,
		Children: nil,
		Data: IncreExpr{
			Operand: operand,
			Pre:     false,
			Op:      MINUS2,
		},
		Loc: p.loc(operand.Loc, end),
	}
}

func parse_object_lit(p *Parser) Node {
	props := map[NodeIndex]ObjectProp{}
	keys := []Node{}
	start := p.expect(O_BRACE).loc
	for p.not_eof() && p.notAt(C_BRACE) {
		idx := len(keys)
		switch p.atType(0) {
		case IDENTIFIER:
			if p.atType(1) == O_PAREN {
				node := parseFunMethod(p)
				keys = append(keys, node.Data.(FnDecl).Name)
				props[idx] = ObjectProp{Node: node, Computed: false}
			} else {
				ident := parse_ident(p)
				keys = append(keys, ident)
				var value Node
				if p.notAt(COLON) {
					value = ident
				} else {
					p.next() // :
					value = p.parseExpr(DEFAULT_BP)
				}
				props[idx] = ObjectProp{Node: value, Computed: false, Accessor: false}
			}
		case STRING, NUMBER:
			key := p.parseExpr(PRIMARY_BP)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				props[idx] = ObjectProp{Node: parseFnFromParams(p, key, false, false, key.Loc), Computed: false, Accessor: false}
			} else {
				p.expect(COLON)
				value := p.parseExpr(DEFAULT_BP)
				props[idx] = ObjectProp{Node: value, Computed: false, Accessor: false}
			}
		case O_BRACKET:
			p.next() // [
			key := p.parseExpr(DEFAULT_BP)
			p.expect(C_BRACKET)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				props[idx] = ObjectProp{Node: parseFnFromParams(p, key, false, false, key.Loc), Computed: true, Accessor: false}
			} else {
				p.expect(COLON)
				value := p.parseExpr(DEFAULT_BP)
				props[idx] = ObjectProp{Node: value, Computed: true, Accessor: false}
			}
		case _async:
			node := parseFunMethod(p)
			keys = append(keys, node.Data.(FnDecl).Name)
			props[idx] = ObjectProp{Node: node, Computed: false, Accessor: false}
		case DOT3:
			p.next()
			keys = append(keys, Node{})
			operand := p.parse_exprs(T_OBJECT_LIT, T_IDENT, T_ARRAY_LIT)
			props[idx] = ObjectProp{Node: Node{
				Tag:      T_RESTORSPREAD,
				Children: nil,
				// T_RESTORSPREAD always returns direct Node.
				Data: operand,
				Loc:  p.loc(start, operand.Loc),
			}, Computed: false, Accessor: false}
		case _get, _set:
			fn := parseFunMethod(p)
			keys = append(keys, fn.Data.(Accessor).Data.(FnDecl).Name)
			props[idx] = ObjectProp{Node: fn, Computed: false, Accessor: true}
		default:
			p.parseError(SyntaxError, "Property key expected (identifier).")
		}
		// if shouldCheckComma && !p.eatComma() && p.notAt(C_BRACE) {
		// 	p.expect(COMMA)
		// }
		p.eatComma()
	}
	end := p.expect(C_BRACE).loc
	return Node{
		Tag:      T_OBJECT_LIT,
		Children: nil,
		Data: ObjectLiteral{
			Props: props,
			Keys:  keys,
		},
		Loc: p.loc(start, end),
	}
}

func parse_array_lit(p *Parser) Node {
	elements := []Node{}
	start := p.expect(O_BRACKET).loc
	for p.not_eof() && p.notAt(C_BRACKET) {
		if p.isAt(DOT3) {
			startLoc := p.eat().loc
			operand := p.parseExpr(ASSIGNMENT_BP)
			elements = append(elements, Node{
				Tag:      T_RESTORSPREAD,
				Children: nil,
				Data:     operand,
				Loc:      p.loc(startLoc, operand.Loc),
			})
		} else {
			elements = append(elements, p.parseExpr(DEFAULT_BP))
		}
		if !p.eatComma() && p.notAt(C_BRACKET) {
			p.expect(COMMA)
		}
	}
	end := p.expect(C_BRACKET).loc
	return Node{
		Tag:      T_ARRAY_LIT,
		Children: elements,
		Data:     nil,
		Loc:      p.loc(start, end),
	}
}

func parse_ternary_expr(p *Parser, condition Node, bp BindingPower) Node {
	p.next() // ?
	Then := p.parseExpr(LOGICAL_OR_BP)
	p.expect(COLON)
	Else := p.parseExpr(bp)
	return Node{
		Tag:      T_TERNARY,
		Children: nil,
		Data:     TernaryExpr{Condition: condition, Then: Then, Else: Else},
		Loc:      p.loc(condition.Loc, Else.Loc),
	}
}

func parse_match_expr(p *Parser) Node {
	loc := p.expect(_match).loc
	operand := p.parseCondition()
	matches := []Node{}
	cases := map[NodeIndex][]Node{}
	var elseCase []Node
	p.expect(O_BRACE)
	for p.not_eof() && p.notAt(C_BRACE) {
		if p.isAt(_else) {
			p.next()
			p.expect(ARROW)
			// The match expression either returns the expression
			// by the match, or returns the yielded value in the block
			body := p.parseBlockOrStmt()
			elseCase = body
		} else {
			idx := len(matches)
			key := p.parseExpr(DEFAULT_BP)
			p.expect(ARROW)
			body := p.parseBlockOrStmt()
			matches = append(matches, key)
			cases[idx] = body
		}
		p.eatComma()
	}
	p.expect(C_BRACE)
	return Node{
		Tag:      T_MATCHEXPR,
		Children: nil,
		Data: MatchExpr{
			Operand: operand,
			Cases:   cases,
			Matches: matches,
			Else:    elseCase,
		},
		Loc: loc,
	}
}

func parse_from_expr(p *Parser) Node {
	loc := p.expect(_from).loc
	path := parse_string(p).Data.(string)
	return Node{
		Tag:      T_FROMEXPR,
		Children: nil,
		Data:     FromExpr{p.parseScriptConcurrent(path)},
		Loc:      loc,
	}
}

func (p *Parser) parseScriptConcurrent(path string) (resolvedPath string) {
	if p.builtin {
		resolvedPath = lib.JoinPaths(lib.DirOf(lib.PathFromKey(p.path)), path)
		lib.Go(func() {
			newp := newBuiltinsParser(path, p.root)
			newp.Parse()
			p.root.Program.Modules = append(p.root.Program.Modules, newp.OwnModule)
		})
	} else {
		if lib.IsAbs(path) {
			resolvedPath = path
		} else {
			resolvedPath = lib.JoinPaths(lib.DirOf(lib.PathFromKey(p.path)), path)
		}
		if !lib.InSlice(p.root.Imports, resolvedPath) {
			lib.Go(func() {
				newp := newParser(resolvedPath, false, p.root)
				newp.Parse()
				p.root.Program.Modules = append(p.root.Program.Modules, newp.OwnModule)
				p.root.Imports = append(p.root.Imports, resolvedPath)
			})
		}
	}
	return
}

func parse_void_expr(p *Parser) Node {
	return parse_operand_expr(p, _void, UNARY_BP, T_VOID)
}

func parse_typeof_expr(p *Parser) Node {
	return parse_operand_expr(p, _typeof, UNARY_BP, T_TYPEOFEXPR)
}

func parse_new_expr(p *Parser) Node {
	return parse_operand_expr(p, _new, CALL_BP, T_NEWEXPR)
}

func parse_async_fn_expr(p *Parser) Node {
	if p.atType(1) == _fn {
		return parseFunExpr(p)
	}
	start := p.expect(_async).loc // async
	params := p.parse_params()
	end := p.expect(ARROW).loc
	var body []Node
	if p.isAt(O_BRACE) {
		body = p.parseBlock()
	} else {
		body = []Node{{
			Tag:      T_RETURN,
			Children: []Node{p.parseExpr(DEFAULT_BP)},
			Data:     nil,
			Loc:      end,
		}}
	}
	return Node{
		Tag:      T_FNDECL,
		Children: body,
		Data: FnDecl{
			Async:     true,
			Params:    params,
			Name:      StringNode(AnonymousName, lib.DumbyLoc),
			Arrow:     true,
			Anonymous: true,
		},
		Loc: p.loc(start, end),
	}
}

func parse_operand_expr(p *Parser, tktag TokenTag, bp BindingPower, nodeTag NodeTag) Node {
	start := p.expect(tktag).loc
	operand := p.parseExpr(bp)
	return Node{
		Tag:      nodeTag,
		Children: nil,
		Data:     OperandExpr{Operand: operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_member_expr(p *Parser, object Node, bp BindingPower) Node {
	start := object.Loc
	computed := false
	member := Node{}
	if p.isAt(O_BRACKET) {
		p.next()
		computed = true
		member = p.parseExpr(DEFAULT_BP)
		p.expect(C_BRACKET)
	} else {
		p.next() // .
		member = parse_ident(p)
	}
	return Node{
		Tag:      T_MEMBER,
		Children: nil,
		Data: MemberExpr{
			Object:   object,
			Member:   member,
			Computed: computed,
		},
		Loc: p.loc(start, member.Loc),
	}
}

func parse_call_expr(p *Parser, caller Node, bp BindingPower) Node {
	loc := caller.Loc
	arguments := p.parse_expr_list()
	return Node{
		Tag:      T_CALLEXPR,
		Children: nil,
		Data:     CallExpr{Caller: caller, Args: arguments},
		Loc:      loc,
	}
}

func parse_comparison_expr(p *Parser, left Node, bp BindingPower) Node {
	return parseLRExpr(p, bp, left, T_COMPARE)
}

func parse_logical_expr(p *Parser, left Node, bp BindingPower) Node {
	return parseLRExpr(p, bp, left, T_LOGICAL)
}

func parse_not_expr(p *Parser) Node {
	start := p.eat().loc
	operand := p.parseExpr(DEFAULT_BP)
	return Node{
		Tag:      T_NOT,
		Children: nil,
		Data:     OperandExpr{Operand: operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_assignment_expr(p *Parser, left Node, bp BindingPower) Node {
	return parseLRExpr(p, bp, left, T_ASSIGN)
}

func parseLRExpr(p *Parser, bp BindingPower, left Node, tag NodeTag) Node {
	op := p.eat().tag
	right := p.parseExpr(bp)
	return Node{
		Tag:      tag,
		Children: nil,
		Data: LROpExpr{
			Lhs: left, Rhs: right, Op: op,
		},
		Loc: p.loc(left.Loc, right.Loc),
	}
}

func parse_binary_expr(p *Parser, left Node, bp BindingPower) Node {
	return parseLRExpr(p, bp, left, T_BINARY_EXP)
}

func parse_open_paren(p *Parser) Node {
	start := p.eat().loc
	can_parse_fn := true
	if p.isAt(C_PAREN) {
		if p.atType(1) != ARROW {
			p.parseError(UnexpectedToken)
		}
		p.next()           // )
		end := p.eat().loc // =>
		var body []Node
		if p.isAt(O_BRACE) {
			body = p.parseBlock()
		} else {
			body = []Node{{
				Tag:      T_RETURN,
				Children: []Node{p.parseExpr(DEFAULT_BP)},
				Data:     nil,
				Loc:      end,
			}}
		}
		return Node{
			Tag:      T_FNDECL,
			Children: body,
			Data: FnDecl{
				Async:     false,
				Params:    []Node{},
				Name:      StringNode(AnonymousName, lib.DumbyLoc),
				Arrow:     true,
				Anonymous: true,
			},
			Loc: p.loc(start, end),
		}
	}
	nodes := []Node{}
	// TODO: add Object and Array destructuring
	valid_params := []NodeTag{T_IDENT}
	for p.notAt(C_PAREN) && p.not_eof() {
		var expr Node
		if len(nodes) > 0 {
			expr = p.parseExpr(DEFAULT_BP)
		} else {
			expr = p.parseExpr(BITOR_BP)
		}
		nodes = append(nodes, expr)
		can_parse_fn = lib.InSlice(valid_params, expr.Tag)
		if p.isAt(BITOR) && len(nodes) == 1 {
			return p.parseCompareList(nodes)
		}
		if p.notAt(C_PAREN) {
			p.expect(COMMA)
		}
	}
	end := p.expect(C_PAREN).loc
	if can_parse_fn && p.isAt(ARROW) {
		end := p.eat().loc
		var body []Node
		if p.isAt(O_BRACE) {
			body = p.parseBlock()
		} else {
			body = []Node{{
				Tag:      T_RETURN,
				Children: []Node{p.parseExpr(DEFAULT_BP)},
				Data:     nil,
				Loc:      end,
			}}
		}
		return Node{
			Tag:      T_FNDECL,
			Children: body,
			Data: FnDecl{
				Async:     false,
				Params:    nodes,
				Name:      StringNode(AnonymousName, lib.DumbyLoc),
				Arrow:     true,
				Anonymous: true,
			},
			Loc: p.loc(start, end),
		}
	}
	if len(nodes) == 1 {
		return nodes[0]
	}
	return Node{
		Tag:      T_GROUPING,
		Data:     nil,
		Children: nodes,
		Loc:      p.loc(start, end),
	}
}

func (p *Parser) parseCompareList(list []Node) Node {
	for p.isAt(BITOR) && p.not_eof() {
		p.next()
		list = append(list, p.parseExpr(BITOR_BP))
	}
	p.expect(C_PAREN)
	return Node{
		Tag:      T_COMPARE_LIST,
		Children: list,
		Data:     nil,
		Loc:      p.loc(list[0].Loc, list[len(list)-1].Loc),
	}
}

func parse_string(p *Parser) Node {
	loc := p.currLoc()
	tk := p.expect(STRING)
	s := p.src(tk)
	return Node{
		Tag:  T_STRING,
		Data: s[1 : len(s)-1],
		Loc:  loc,
	}
}

func StringNode(str string, loc Loc) Node {
	return Node{
		Tag:  T_STRING,
		Data: str,
		Loc:  loc,
	}
}

func parse_number(p *Parser) Node {
	loc := p.currLoc()
	tk := p.expect(NUMBER)
	return Node{
		Tag:  T_NUMBER,
		Data: lib.ParseNumber(p.src(tk)),
		Loc:  loc,
	}
}

func NumberNode(num float64, loc Loc) Node {
	return Node{
		Tag:  T_NUMBER,
		Data: num,
		Loc:  loc,
	}
}

func parse_ident(p *Parser) Node {
	loc := p.currLoc()
	tk := p.expect(IDENTIFIER)
	return Node{
		Tag:  T_IDENT,
		Data: Identifier(p.src(tk)),
		Loc:  loc,
	}
}

func (p *Parser) parseError(errname ErrorName, additionals ...string) {
	name := "SyntaxError"
	message := ""
	// Use a string builder for efficient string concatenation and to
	// make it safe for concurrent parser errors
	b := lib.NewStringBuilder()
	b.WriteString("ParseError: ")
	switch errname {
	case UnexpectedToken:
		message = lib.Sprintf("Unexpected token reached: %s%s%s", p.at(0).tag.Lexeme(), lib.EOL, lib.SourceLog(p.path, p.currLoc()))
	case SyntaxError:
		message = lib.Sprintf("%s%s%s", additionals[0], lib.EOL, lib.SourceLog(p.path, p.currLoc()))
		additionals = additionals[1:]
	case PathError:
		name = "PathError"
		message = additionals[0]
		additionals = additionals[1:]
	default:
		lib.Panic("unhandled error name")
	}
	b.WriteString(lib.Red(string(name)))
	b.WriteString(lib.EOL)
	b.WriteString(message)
	b.WriteString(lib.EOL)
	for i, m := range additionals {
		b.WriteString(m)
		if i == len(additionals)-1 && additionals[i][len(additionals[i])-1] != '\n' {
			b.WriteString(lib.EOL)
		}
	}
	lib.Print(b.String())
	// lib.ExitWith1()
	// exit thread instead of crashing immediately
	// lib.Goexit()
	panic("Parse Error...")
}

func (p *Parser) not_eof() bool {
	return p.atType(0) != EOF
}

func (p *Parser) loc(startLoc, end Loc) Loc {
	return Loc{
		Line:  startLoc.Line,
		Col:   startLoc.Col,
		Start: startLoc.Start,
		End:   end.End,
	}
}

func (p *Parser) currLoc() Loc {
	return p.at(0).loc
}

func (p *Parser) src(tk Token) string {
	return p.lexer.src(tk)
}

func (p *Parser) locSrc(loc Loc) string {
	return p.lexer.src(token(0, loc))
}

func (p *Parser) at(idx int) Token {
	l := len(p.tokens)
	idx = int(p.tokenIndex) + idx
	if idx >= l || idx < 0 {
		lib.Panic("unexpected behaviour")
	}
	return p.tokens[idx]
}

func (p *Parser) atType(idx int) TokenTag {
	return p.at(idx).tag
}

func (p *Parser) next() {
	p.tokenIndex++
}

func (p *Parser) eat() Token {
	defer p.next()
	return p.at(0)
}

func (p *Parser) isAt(t TokenTag) bool {
	currT := p.atType(0)
	switch t {
	case DECL:
		return currT > decl_keyword_start && currT < decl_keyword_end
	case VARDECL:
		return currT > var_decl_keyword_start && currT < var_decl_keyword_end
	case LITERAL:
		return currT > literal_start && currT < literal_end
	case BINARYOP:
		return currT > binary_op_start && currT < binary_op_end
	case ACCESSOR:
		return currT > accessor_start && currT < accessor_end
	}
	return currT == t
}

func (p *Parser) notAt(t TokenTag) bool {
	return !p.isAt(t)
}

func (p *Parser) debug(n Node) {
	lib.Printf("Node \x1b[32m%s\x1b[0m [%d:%d, %d:%d]%s", n.Tag.Name(), n.Loc.Line, n.Loc.Col, n.Loc.Start, n.Loc.End, lib.EOL)
}

func (p *Parser) PrintNodes(count uint) {
	lib.Println(lib.Yellow("----------- AST Nodes ----------"))
	for _, n := range p.Program.Main.Body {
		p.debug(n)
	}
	for i := uint(0); i < count && i < uint(len(p.Program.Modules)); i++ {
		for _, n := range p.Program.Modules[i].Body {
			p.debug(n)
		}
	}
}

// func (p *Parser)
