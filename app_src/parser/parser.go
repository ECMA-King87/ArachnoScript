package parser

import (
	"aspire/are/main/lib"
)

// type Cache struct {
// 	ParserVersion string
// 	Hash          string
// 	SourceModTime int
// 	AST          * Node
// }

type Loc = lib.Loc

type Parser struct {
	Imports *lib.Map[string, int]
	Program *Program
	errors  *lib.Array[string]
	lib.WaitGroup
	mu lib.RWMutex
}

func (p *Parser) ParseProgram(path string, builtin, isMain, isRepl bool) (m *ModuleParser) {
	if builtin {
		m = NewBuiltinsParser(path)
	} else {
		m = NewModuleParser(path, isMain, isRepl)
	}
	m.Parse()
	p.Wait()
	if p.errors.Len() != 0 {
		p.errors.ForEach(func(_ int, s string) {
			lib.Println(s)
		})
	}
	return
}

var GlobalParser = NewParser()

type ModuleParser struct {
	path       int
	Path       string
	lexer      *Lexer
	tokens     []Token
	tokenIndex uint64
	builtin    bool
	OwnModule  *Module
}

func init() {
	defStmt(K_const, parseVarDeclStmt)
	defStmt(K_let, parseVarDeclStmt)
	defStmt(K_var, parseVarDeclStmt)
	defStmt(K_if, parseIfStmt)
	defStmt(O_BRACE, parseBlockStmt)
	defStmt(LABEL, parseLabel)
	defStmt(K_throw, parseThrowStmt)
	defStmt(K_fn, parseFunDecl)
	defStmt(K_async, parseFunDecl)
	defStmt(K_return, parseReturnStmt)
	defStmt(K_while, parseWhileStmt)
	defStmt(K_do, parseDoWhileStmt)
	defStmt(K_for, parseForStmt)
	defStmt(K_switch, parseSwitchStmt)
	defStmt(K_yield, parse_yield_stmt)
	defStmt(K_import, parse_import_stmt)
	defStmt(K_export, parse_export_stmt)
	defStmt(K_class, parse_class_stmt)
	defStmt(K_try, parse_try_catch_finally)
	defStmt(K_fallthrough, parse_fallthrough)
	defStmt(K_break, parse_break_stmt)
	defStmt(K_continue, parse_continue_stmt)

	defNud(IDENTIFIER, parse_ident)
	defNud(STRING, parse_string)
	defNud(NUMBER, parse_number)
	defNud(NOT, parse_not_expr)
	defNud(O_PAREN, parse_open_paren)

	defNud(K_new, parse_new_expr)
	defNud(K_void, parse_void_expr)
	defNud(K_async, parse_async_fn_expr)
	defNud(K_typeof, parse_typeof_expr)
	defNud(K_match, parse_match_expr)
	defNud(K_from, parse_from_expr)
	defNud(K_class, parse_class_expr)
	defNud(K_gt, parse_globalthis)
	defNud(K_super, parse_super_expr)
	defNud(K_await, parse_await_expr)
	defNud(K_fn, parseFunExpr)

	defNud(O_BRACE, parse_object_lit)
	defNud(O_BRACKET, parse_array_lit)

	defNud(PLUS2, parse_plus2_prefix)
	defNud(MINUS2, parse_minus2_prefix)

	defNud(BITNOT, parse_bitnot_expr)
	defNud(MINUS, parse_unary_expr)
	defNud(PLUS, parse_unary_expr)

	defLed(PLUS, parse_binary_expr, ADDITIVE_BP)
	defLed(MINUS, parse_binary_expr, ADDITIVE_BP)
	defLed(TIMES, parse_binary_expr, MULTIPLICATIVE_BP)
	defLed(STAR2, parse_binary_expr, EXPONENT_BP)
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

	defLed(K_instanceof, parse_instanceof_expr, COMPARISON_BP)
	defLed(K_in, parse_in_expr, IN_BP)

	defLed(NULLISH, parse_nullish_expr, NULLISH_COALLESCE_BP)
}

func NewParser() *Parser {
	return &Parser{
		Imports: lib.NewMap[string, int](),
		Program: &Program{
			Main:    &Module{},
			Modules: map[int]*Module{},
		},
		errors: lib.NewArray[string](0),
	}
}

func NewBuiltinsParser(name string) *ModuleParser {
	path := "stdlib/" + name
	var bytes []byte
	var key int
	if lib.DEBUG_MODE {
		b, k, err := lib.ReadFile(
			lib.JoinPaths(lib.DirOf(lib.ExecPath()), "..", "app_src", "lib", "stdlib", name),
		)
		if err != nil {
			(&ModuleParser{}).parseError(PathError, err.Error())
		}
		bytes = b
		key = k
	} else {
		b, err := lib.EMBEDFS.ReadFile(path)
		if err != nil {
			(&ModuleParser{}).parseError(PathError, err.Error())
		}
		bytes = b
		key = lib.WriteCacheFile(bytes, path)
	}
	// GlobalParser.Imports = append(GlobalParser.Imports, path)
	if !lib.DEBUG_MODE {
		path = lib.AnonymousPath
	}
	lexer := NewLexer(bytes, key)
	tokens := lexer.Tokenize()
	module := &Module{
		Path:    key,
		Body:    []*Node{},
		Invalid: false,
	}
	GlobalParser.mu.Lock()
	GlobalParser.Program.Modules[key] = module
	GlobalParser.mu.Unlock()
	return &ModuleParser{
		path:       key,
		lexer:      lexer,
		tokens:     tokens,
		tokenIndex: 0,
		OwnModule:  module,
		builtin:    true,
		Path:       path,
	}
}

// path: represents the path to the module to parse.
// if the path does not exist, the parser will attempt to parse 'path'.
// handle this if it is unwanted behavior.
func NewModuleParser(path string, isMain, isRepl bool) (mp *ModuleParser) {
	defer func() { recover() }()
	var pk int
	var buffer []byte
	if isRepl {
		buffer = []byte(path)
		pk = lib.WriteAnonymousFile(buffer, 0)
		path = lib.AnonymousPath
	} else {
		abs := path
		if !lib.IsAbs(path) {
			abs = lib.Abs(path)
		}
		b, k, err := lib.ReadFile(abs)
		if err != nil {
			(&ModuleParser{}).parseError(PathError, err.Error())
		}
		buffer = b
		pk = k
	}
	lexer := NewLexer(buffer, pk)
	tokens := lexer.Tokenize()
	mp = &ModuleParser{
		path:       pk,
		lexer:      lexer,
		tokens:     tokens,
		tokenIndex: 0,
		OwnModule: &Module{
			Path:    pk,
			Body:    []*Node{},
			Invalid: false,
		},
		builtin: false,
		Path:    path,
	}
	GlobalParser.Imports.Set(path, pk)
	GlobalParser.mu.Lock()
	if isMain {
		GlobalParser.Program.Main = mp.OwnModule
	} else {
		GlobalParser.Program.Modules[pk] = mp.OwnModule
	}
	GlobalParser.mu.Unlock()
	return
}

func (p *ModuleParser) Parse() {
	if p == nil {
		return
	}
	defer func() {
		err := recover()
		p.OwnModule.Invalid = err != nil
		if p.OwnModule.Invalid {
			GlobalParser.errors.Push(err.(string))
		}
	}()
	for p.not_eof() {
		if p.eatSemiColons() {
			continue
		}
		p.OwnModule.Body = append(p.OwnModule.Body, p.parseStmt())
	}
}

func (p *ModuleParser) expect(t TokenTag) Token {
	if p.notAt(t) {
		p.parseError(SyntaxError, lib.Sprintf("%s\x1b[36m'%s'\x1b[0m.", "Expected a token of type ", t.Lexeme()))
	}
	return p.eat()
}

func (p *ModuleParser) eatSemiColons() bool {
	if p.isAt(SEMICOLON) {
		for p.isAt(SEMICOLON) {
			p.next()
		}
		return true
	}
	return false
}

func (p *ModuleParser) eatSemiColon() bool {
	if p.isAt(SEMICOLON) {
		p.next()
		return true
	}
	return false
}

func (p *ModuleParser) eatComma() bool {
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
func (p *ModuleParser) parseStmt() *Node {
	if handler, ok := stmtLU[p.atType(0)]; ok {
		defer p.eatSemiColons()
		return handler(p)
	}
	return p.parseExprStmt()
}

func parseFunDecl(p *ModuleParser) *Node {
	var isAsync = false
	if p.isAt(K_async) {
		p.next()
		isAsync = true
	}
	loc := p.expect(K_fn).loc // function
	nameTk := p.expect(IDENTIFIER)
	return parseFnFromParams(p, StringNode(p.src(nameTk), nameTk.loc), isAsync, false, false, loc)
}

const AnonymousName = "(anonymous)"

func parseFunExpr(p *ModuleParser) *Node {
	isAsync := false
	start := p.currLoc()
	if p.isAt(K_async) {
		p.next()
		isAsync = true
	}
	p.expect(K_fn)
	// Function expression is Anonymous
	return parseFnFromParams(p, StringNode(AnonymousName, lib.DumbyLoc), isAsync, true, false, start)
}

func parseFunMethod(p *ModuleParser) Member {
	isAsync := false
	idx := p.tokenIndex
	if p.isAt(K_async) {
		p.next()
		isAsync = true
	}
	if p.isAt(ACCESSOR) {
		tk := p.eat()
		getter := tk.tag == K_get
		if isAsync {
			p.tokenIndex = idx
			p.parseError(SyntaxError, "'async' modifier cannot be used here.")
		}
		method := parseMethod(p, isAsync)
		return Member{method.Computed, &Node{
			Tag:      T_ACCESSOR,
			Children: nil,
			Data: Accessor{
				Getter: getter,
				Node:   method.Node,
			},
			Loc: tk.loc,
		}}
	}
	return parseMethod(p, isAsync)
}

func parseMethod(p *ModuleParser, isAsync bool) Member {
	var NameNode *Node
	computed := true
	switch p.atType(0) {
	case O_BRACKET:
		p.next() // [
		NameNode = p.parseExpr(DEFAULT_BP)
		p.expect(C_BRACKET)
	case IDENTIFIER:
		computed = false
		NameNode = parse_ident(p)
	case STRING:
		NameNode = parse_string(p)
	case NUMBER:
		NameNode = parse_number(p)
	default:
		computed = false
		NameNode = parse_prop_name(p)
	}
	// Method cannot be Anonymous
	return Member{Node: parseFnFromParams(p, NameNode, isAsync, false, true, NameNode.Loc), Computed: computed}
}

func parseFnFromParams(p *ModuleParser, NameNode *Node, isAsync, isAnony, isMethod bool, start Loc) *Node {
	params := p.parse_params()
	body := p.parseBlock()
	return &Node{
		Tag:      T_FNDECL,
		Children: body,
		Data: FnDecl{
			Params:    params,
			Name:      NameNode,
			Async:     isAsync,
			Arrow:     isMethod,
			Anonymous: isAnony,
		},
		Loc: start,
	}
}

func parseReturnStmt(p *ModuleParser) *Node {
	end := p.expect(K_return).loc // return
	var value *Node
	if !p.eatSemiColons() && p.notAt(C_BRACE) {
		value = p.parseExpr(DEFAULT_BP)
	}
	return &Node{
		Tag:      T_RETURN,
		Children: []*Node{value},
		Data:     nil,
		Loc:      end,
	}
}

func parseThrowStmt(p *ModuleParser) *Node {
	start := p.expect(K_throw).loc // throw
	value := p.parseExpr(DEFAULT_BP)
	return &Node{
		Tag:      T_THROW,
		Children: []*Node{value},
		Data:     nil,
		Loc:      p.loc(start, value.Loc),
	}
}

func parseWhileStmt(p *ModuleParser) *Node {
	start := p.expect(K_while).loc // while
	condition := p.parseCondition()
	end := condition.Node.Loc
	var body []*Node
	if !p.eatSemiColon() {
		body = p.parseBlockOrStmt()
	}
	return &Node{
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

func parseForStmt(p *ModuleParser) *Node {
	start := p.expect(K_for).loc // 'for'
	p.expect(O_PAREN)
	tag := T_TFORLOOP
	var before_lhs *Node
	var condition *Node
	var after_rhs *Node
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
		if p.isAt(K_in) || p.isAt(K_of) {
			if p.eat().tag == K_in {
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
	var body []*Node
	if !p.eatSemiColon() {
		body = p.parseBlockOrStmt()
	}
	return &Node{
		Tag:      tag,
		Children: body,
		Data:     data,
		Loc:      p.loc(start, end),
	}
}

func parseDoWhileStmt(p *ModuleParser) *Node {
	start := p.expect(K_do).loc // do
	var body = p.parseBlockOrStmt()
	p.expect(K_while)
	condition := p.parseCondition()
	return &Node{
		Tag:      T_WHILE,
		Children: nil,
		Data: WhileLoop{
			Do:        true,
			Body:      body,
			Condition: condition,
		},
		Loc: p.loc(start, condition.Node.Loc),
	}
}

func parseVarDecl(p *ModuleParser) *Node {
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
	return &Node{
		Data: VarDecl{
			Kind:  declType,
			Decls: decls,
		},
		Loc: p.loc(startLoc, endLoc),
		Tag: T_VARDECL,
	}
}

func handleDeclKeyword(p *ModuleParser) DeclKind {
	declType := MutableDecl
	switch p.atType(0) {
	case K_const:
		declType = ConstantDecl
	case K_var:
		declType = HoistedDecl
	case K_let:
		break
	default:
		p.parseError(SyntaxError, "Expected a declaration keyword.")
	}
	return declType
}

func parseVarDeclStmt(p *ModuleParser) *Node {
	defer p.eatSemiColon()
	return parseVarDecl(p)
}

func (p *ModuleParser) parseDeclRhs(declType DeclKind, errMsg string) *Node {
	var rhs *Node
	if declType == ConstantDecl && p.notAt(EQUALS) {
		p.parseError(SyntaxError, errMsg)
	} else if p.isAt(EQUALS) {
		p.next()
		rhs = p.parseExpr(DEFAULT_BP)
	}
	return rhs
}

func (p *ModuleParser) parseVarDeclLhs() *Node {
	switch p.atType(0) {
	case IDENTIFIER:
		return parse_ident(p)
	case O_BRACE:
		return p.parse_object_destructuring()
	case O_BRACKET:
		return p.parse_array_destructuring()
	default:
		p.parseError(SyntaxError, "Expected identifier on Left-Hand-Side of Variable Declaration.")
		return &Node{}
	}
}

// report == true if the expr is valid destructuring syntax
func (p *ModuleParser) parse_array_or_destructuring() (node *Node, report bool) {
	idx := p.tokenIndex
	defer func() {
		if recover() != nil {
			// report = false
			p.tokenIndex = idx
			p.OwnModule.Invalid = false
			node = parse_array_lit(p)
		}
	}()
	node = p.parse_array_destructuring()
	report = true
	return
}

func (p *ModuleParser) parse_array_destructuring() *Node {
	start := p.expect(O_BRACKET).loc
	elements := []*Node{}
	for p.not_eof() && p.notAt(C_BRACKET) {
		if p.isAt(DOT3) {
			loc := p.eat().loc
			elements = append(elements, &Node{
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
	return &Node{
		Tag:      T_ARRAY_LIT,
		Children: elements,
		Data:     nil,
		Loc:      p.loc(start, end),
	}
}

// report == true if the expr is valid destructuring syntax
func (p *ModuleParser) parse_object_or_destructuring() (node *Node, report bool) {
	idx := p.tokenIndex
	defer func() {
		if recover() != nil {
			// report = false
			p.tokenIndex = idx
			p.OwnModule.Invalid = false
			node = parse_object_lit(p)
		}
	}()
	node = p.parse_object_destructuring()
	report = true
	return
}

func (p *ModuleParser) parse_object_destructuring() *Node {
	start := p.expect(O_BRACE).loc
	props := map[NodeIndex]ObjectDestProp{}
	keys := []*Node{}
	for p.not_eof() && p.notAt(C_BRACE) {
		var value, def *Node
		var computed bool = false
		idx := len(keys)
		at := p.atType(0)
		switch {
		case at == IDENTIFIER || p.isAt(KEYWORD):
			ident := parse_prop_name(p)
			keys = append(keys, ident)
			if p.notAt(COLON) {
				value = ident
			} else {
				p.next() // :
				value = parse_ident(p)
			}
		case at == STRING || at == NUMBER:
			key := p.parseExpr(PRIMARY_BP)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				value = parseFnFromParams(p, key, false, false, true, key.Loc)
			} else {
				p.expect(COLON)
				value = parse_ident(p)
			}
		case at == O_BRACKET:
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
	return &Node{
		Tag:      T_OBJECT_LIT,
		Children: nil,
		Data: ObjectDest{
			Props: props,
			Keys:  keys,
		},
		Loc: p.loc(start, end),
	}
}

func parseIfStmt(p *ModuleParser) *Node {
	loc := p.expect(K_if).loc
	condition := p.parseCondition()
	body := p.parseBlockOrStmt()
	var elseBlock []*Node
	if p.isAt(K_else) {
		p.next()
		elseBlock = p.parseBlockOrStmt()
	}
	return &Node{
		Tag:      T_IFSTMT,
		Children: body,
		Data: IfStmt{
			Condition: condition,
			ElseBlock: elseBlock,
		},
		Loc: loc,
	}
}

func (p *ModuleParser) parseCondition() (cond Condition) {
	p.expect(O_PAREN)
	defer p.expect(C_PAREN)
	if p.isAt(VARDECL) {
		cond.Decl = parseVarDecl(p)
		p.expect(SEMICOLON)
	}
	cond.Node = p.parseExpr(DEFAULT_BP)
	return
}

func parseBlockStmt(p *ModuleParser) *Node {
	loc := p.currLoc()
	return &Node{
		Tag:      T_BLOCK,
		Children: p.parseBlock(),
		Data:     nil,
		Loc:      loc,
	}
}

func (p *ModuleParser) parseBlock() []*Node {
	block := []*Node{}
	p.expect(O_BRACE)
	for p.not_eof() && p.notAt(C_BRACE) {
		block = append(block, p.parseStmt())
	}
	p.expect(C_BRACE)
	return block
}

func (p *ModuleParser) parseBlockOrStmt() []*Node {
	if p.isAt(O_BRACE) {
		return p.parseBlock()
	}
	return []*Node{p.parseStmt()}
}

func parseLabel(p *ModuleParser) *Node {
	loc := p.currLoc()
	// remove the preceding '@'
	l := p.src(p.expect(LABEL))[1:] // label
	return &Node{
		Tag:      T_LABEL,
		Children: nil,
		Data:     l,
		Loc:      loc,
	}
}

func parseSwitchStmt(p *ModuleParser) *Node {
	loc := p.expect(K_switch).loc

	p.expect(O_PAREN)
	var condition *Node
	if p.isAt(VARDECL) {
		condition = parseVarDecl(p)
	} else {
		condition = p.parseExpr(DEFAULT_BP)
	}
	p.expect(C_PAREN)

	p.expect(O_BRACE)
	cases := map[NodeIndex][]*Node{}
	matches := [][]*Node{}
	var defaultCase []*Node
	for p.not_eof() && p.notAt(C_BRACE) {
		idx := len(matches)
		if p.isAt(K_default) {
			p.next()
			p.expect(COLON)
			body := p.parseBlockOrStmt()
			defaultCase = body
		} else {
			p.expect(K_case)
			match := []*Node{p.parseExpr(DEFAULT_BP)}
			for p.isAt(COMMA) {
				p.next()
				match = append(match, p.parseExpr(DEFAULT_BP))
			}
			p.expect(COLON)
			body := p.parseBlockOrStmt()
			matches = append(matches, match)
			cases[idx] = body
		}
	}
	p.expect(C_BRACE)
	return &Node{
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

func parse_yield_stmt(p *ModuleParser) *Node {
	loc := p.expect(K_yield).loc
	return &Node{
		Tag:      T_YIELDSTMT,
		Children: nil,
		Data:     OperandExpr{p.parseExpr(DEFAULT_BP)},
		Loc:      p.loc(loc, p.currLoc()),
	}
}

func parse_import_stmt(p *ModuleParser) *Node {
	loc := p.expect(K_import).loc
	imp_path := "invalid-path.as"
	ns := ""
	ucc := false
	var named *Node
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
	return &Node{
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

func parse_export_stmt(p *ModuleParser) *Node {
	loc := p.expect(K_export).loc
	var export *Node
	if p.isAt(DECL) {
		export = p.parseStmt()
	} else if p.isAt(O_BRACE) {
		export = p.parse_object_destructuring()
	} else {
		p.parseError(SyntaxError, "Declaration expected.")
	}
	return &Node{
		Tag:      T_EXPORT,
		Children: nil,
		Data:     ExportStmt{export},
		Loc:      loc,
	}
}

func parse_class_stmt(p *ModuleParser) *Node {
	loc := p.expect(K_class).loc
	name := string(parse_ident(p).Data.(Identifier))
	return parse_class_decl(p, name, false, loc)
}

func parse_class_decl(p *ModuleParser, name string, isAnony bool, loc lib.Loc) *Node {
	methods := map[int]Member{}
	props := map[int]Member{}
	memberModifiers := [][]TokenTag{}
	var defaultProp DefaultProp
	var constructor, extends *Node
	if p.isAt(K_extends) {
		p.next()
		extends = p.parseExpr(ASSIGNMENT_BP)
	}
	p.expect(O_BRACE)
	// Increment after use
	memberCount := 0
	hasCtor := false
	for p.not_eof() && p.notAt(C_BRACE) {
		start := p.currLoc()
		switch p.atType(0) {
		case K_ctor:
			if hasCtor {
				p.parseError(SyntaxError, "Multiple constructor implementations are not allowed.", lib.EOL, "First implementation is ", lib.SourceAtPosition(p.Path, constructor.Loc))
			}
			hasCtor = true
			ctor := p.eat()
			// Constructor's name is the class name.
			constructor = parseFnFromParams(p, StringNode(name, ctor.loc), false, false, true, start)
		default:
			computed := true
			modifiers := p.parse_modifiers()
			memberModifiers = append(memberModifiers, modifiers)
			if lib.InSlice(modifiers, K_default) {
				defaultProp = DefaultProp{modifiers, p.parse_class_prop()}
			} else if p.isAt(K_async) || p.isAt(K_get) || p.isAt(K_set) {
				methods[memberCount] = parseFunMethod(p)
			} else {
				var NameNode *Node
				at := p.atType(0)
				switch {
				case at == O_BRACKET:
					p.next() // [
					NameNode = p.parseExpr(DEFAULT_BP)
					p.expect(C_BRACKET)
				case at == IDENTIFIER:
					computed = false
					NameNode = parse_ident(p)
				case p.isAt(KEYWORD):
					computed = false
					NameNode = parse_prop_name(p)
				case at == STRING:
					NameNode = parse_string(p)
				case at == NUMBER:
					NameNode = parse_number(p)
				default:
					p.expect(IDENTIFIER)
				}
				if p.isAt(O_PAREN) {
					methods[memberCount] = Member{Node: parseFnFromParams(p, NameNode, false, false, true, start), Computed: computed}
				} else {
					if p.isAt(EQUALS) {
						p.next()
						rhs := p.parseExpr(DEFAULT_BP)
						props[memberCount] = Member{Computed: computed, Node: &Node{
							Tag:      T_VARDECL,
							Children: nil,
							Data: Decl{
								Lhs: NameNode,
								Rhs: rhs,
							},
							Loc: p.loc(start, rhs.Loc),
						}}
					} else {
						props[memberCount] = Member{Computed: computed, Node: &Node{
							Tag:      T_VARDECL,
							Children: nil,
							Data: Decl{
								Lhs: NameNode,
								Rhs: nil,
							},
							Loc: start,
						}}
					}
				}
			}
			memberCount++
		}
		p.eatSemiColons()
	}
	p.expect(C_BRACE)
	return &Node{
		Tag:      T_CLASSDECL,
		Children: nil,
		Data: ClassDecl{
			DefProp:         defaultProp,
			Methods:         methods,
			Props:           props,
			Constructor:     constructor,
			Extends:         extends,
			Name:            name,
			Anonymous:       isAnony,
			MemberModifiers: memberModifiers,
		},
		Loc: loc,
	}
}

func (p *ModuleParser) parse_class_prop() Member {
	var lhs, rhs *Node
	start := p.currLoc()
	at := p.atType(0)
	computed := true
	switch {
	case p.isAt(KEYWORD) || at == IDENTIFIER:
		computed = false
		lhs = parse_prop_name(p)
	case at == STRING || at == NUMBER:
		lhs = p.parseExpr(PRIMARY_BP)
	case at == O_BRACKET:
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
	return Member{Node: &Node{
		Tag:      T_VARDECL,
		Children: nil,
		Data: Decl{
			Lhs: lhs,
			Rhs: rhs,
		},
		Loc: p.loc(start, rhs.Loc),
	}, Computed: computed}
}

func (p *ModuleParser) parse_modifiers() (modifiers []TokenTag) {
	if p.isAt(K_private) || p.isAt(K_public) {
		modifiers = append(modifiers, p.eat().tag)
	}
	if p.isAt(K_static) {
		modifiers = append(modifiers, p.eat().tag)
	}
	if p.isAt(K_default) {
		modifiers = append(modifiers, p.eat().tag)
	}
	return
}

func parse_try_catch_finally(p *ModuleParser) *Node {
	loc := p.expect(K_try).loc
	try := p.parseBlock()
	var catch, finally []*Node
	var capture *Node
	if p.isAt(K_catch) {
		p.next()
		if p.isAt(O_PAREN) {
			p.next()
			capture = p.parse_param()
			p.expect(C_PAREN)
		}
		catch = p.parseBlock()
	} else if p.notAt(K_finally) {
		p.expect(K_finally)
	}
	if p.isAt(K_finally) {
		p.next()
		finally = p.parseBlock()
	}
	return &Node{
		Tag:      T_TRYCATCH,
		Children: nil,
		Data: TryCatch{
			Try:     try,
			Catch:   catch,
			Capture: capture,
			Finally: finally,
		},
		Loc: loc,
	}
}

func parse_fallthrough(p *ModuleParser) *Node {
	return &Node{
		Tag:      T_FALLTHROUGH,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(K_fallthrough).loc,
	}
}

func parse_break_stmt(p *ModuleParser) *Node {
	return &Node{
		Tag:      T_BREAKSTMT,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(K_break).loc,
	}
}

func parse_continue_stmt(p *ModuleParser) *Node {
	return &Node{
		Tag:      T_CONTINUESTMT,
		Children: nil,
		Data:     nil,
		Loc:      p.expect(K_continue).loc,
	}
}

// @#############################################@
// @#############################################@
// @################ Expressions ################@
// @#############################################@
// @#############################################@

func (p *ModuleParser) parseExprStmt() *Node {
	defer p.eatSemiColon()
	return p.parseExpr(DEFAULT_BP)
}

func (p *ModuleParser) parseExpr(bp BindingPower) *Node {
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

func (p *ModuleParser) parse_exprs(valid_tags ...NodeTag) *Node {
	tokenIndex := p.tokenIndex
	expr := p.parseExpr(DEFAULT_BP)
	if lib.InSlice(valid_tags, expr.Tag) {
		return expr
	}
	p.tokenIndex = tokenIndex
	p.parseError(SyntaxError, lib.Sprintf("Unexpected expression: %s", expr.Tag.Name()))
	return &Node{}
}

func (p *ModuleParser) parse_param() *Node {
	switch p.atType(0) {
	case DOT3:
		start := p.eat().loc
		param := p.parse_exprs(T_IDENT, T_ARRAY_LIT)
		return &Node{
			Tag:      T_RESTORSPREAD,
			Children: nil,
			Data:     param,
			Loc:      p.loc(start, param.Loc),
		}
	case TIMES:
		lib.Panic("language feature unimplemented.")
		return &Node{}
	case O_BRACE:
		return p.parse_object_destructuring()
	case O_BRACKET:
		return p.parse_array_destructuring()
	default:
		return parse_ident(p)
	}
}

func (p *ModuleParser) parse_params() []*Node {
	p.expect(O_PAREN)
	args := []*Node{}
	for p.not_eof() && p.notAt(C_PAREN) {
		args = append(args, p.parse_param())
		if !p.eatComma() && p.notAt(C_PAREN) {
			p.expect(COMMA)
		}
	}
	p.expect(C_PAREN)
	return args
}

func (p *ModuleParser) parse_args() []*Node {
	p.expect(O_PAREN)
	exprs := []*Node{}
	for p.not_eof() && p.notAt(C_PAREN) {
		if p.isAt(DOT3) {
			p.next()
			expr := p.parseExpr(DEFAULT_BP)
			exprs = append(exprs, &Node{
				Tag:      T_RESTORSPREAD,
				Children: nil,
				Data:     expr,
				Loc:      expr.Loc,
			})
		} else {
			exprs = append(exprs, p.parseExpr(DEFAULT_BP))
		}
		if !p.eatComma() && p.notAt(C_PAREN) {
			p.expect(COMMA)
		}
	}
	p.expect(C_PAREN)
	return exprs
}

// func (p *ModuleParser) parse_expr_list() []*Node {
// 	p.expect(O_PAREN)
// 	exprs := []*Node{}
// 	for p.not_eof() && p.notAt(C_PAREN) {
// 		exprs = append(exprs, p.parseExpr(DEFAULT_BP))
// 		if !p.eatComma() && p.notAt(C_PAREN) {
// 			p.expect(COMMA)
// 		}
// 	}
// 	p.expect(C_PAREN)
// 	return exprs
// }

func parse_globalthis(p *ModuleParser) *Node {
	loc := p.expect(K_gt).loc
	return &Node{
		Tag:      T_GT,
		Children: nil,
		Data:     nil,
		Loc:      loc,
	}
}

func parse_await_expr(p *ModuleParser) *Node {
	start := p.expect(K_await).loc
	operand := p.parseExpr(ASSIGNMENT_BP)
	return &Node{
		Tag:      T_AWAIT,
		Children: nil,
		Data:     OperandExpr{operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_super_expr(p *ModuleParser) *Node {
	start := p.expect(K_super).loc
	params := p.parse_args()
	return &Node{
		Tag:      T_SUPER,
		Children: params,
		Data:     nil,
		Loc:      p.loc(start, p.currLoc()),
	}
}

func parse_class_expr(p *ModuleParser) *Node {
	loc := p.expect(K_class).loc
	return parse_class_decl(p, AnonymousName, true, loc)
}

func parse_nullish_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	// ??
	return parse_lrop_expr(p, bp, left, T_NULLISH)
}

func parse_lrop_expr(p *ModuleParser, bp BindingPower, left *Node, nodeTag NodeTag) *Node {
	p.next()
	right := p.parseExpr(bp)
	return &Node{
		Tag:      nodeTag,
		Children: nil,
		Data: LROpExpr{
			Lhs: left,
			Rhs: right,
			Op:  0,
		},
		Loc: p.loc(left.Loc, right.Loc),
	}
}

func parse_instanceof_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parse_lrop_expr(p, bp, left, T_INSTANCEOF)
}

func parse_in_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parse_lrop_expr(p, bp, left, T_INEXPR)
}

func parse_bitwise_expr(p *ModuleParser, operand *Node, bp BindingPower) *Node {
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
	return &Node{
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

func parse_unary_expr(p *ModuleParser) *Node {
	op := p.eat() // + | -
	operand := p.parseExpr(UNARY_BP)
	return &Node{
		Tag:      T_UNARY,
		Children: nil,
		Data: OpOpExpr{
			Operand: operand,
			Op:      op.tag,
		},
		Loc: p.loc(op.loc, operand.Loc),
	}
}

func parse_bitnot_expr(p *ModuleParser) *Node {
	return parse_operand_expr(p, BITNOT, UNARY_BP, T_BITNOT)
}

func parse_plus2_prefix(p *ModuleParser) *Node {
	start := p.eat().loc // ++
	operand := p.parseExpr(UNARY_BP)
	return &Node{
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

func parse_minus2_prefix(p *ModuleParser) *Node {
	start := p.eat().loc // --
	operand := p.parseExpr(UNARY_BP)
	return &Node{
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

func parse_plus2_postfix(p *ModuleParser, operand *Node, bp BindingPower) *Node {
	end := p.eat().loc // ++
	return &Node{
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

func parse_minus2_postfix(p *ModuleParser, operand *Node, bp BindingPower) *Node {
	end := p.eat().loc // --
	return &Node{
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

func parse_object_lit(p *ModuleParser) *Node {
	props := map[NodeIndex]ObjectProp{}
	keys := []*Node{}
	start := p.expect(O_BRACE).loc
	for p.not_eof() && p.notAt(C_BRACE) {
		idx := len(keys)
		at := p.atType(0)
		switch {
		case at == IDENTIFIER || (p.isAt(KEYWORD) && at != K_async && at != K_get && at != K_set):
			if p.atType(1) == O_PAREN {
				node := parseFunMethod(p)
				keys = append(keys, node.Data.(FnDecl).Name)
				props[idx] = ObjectProp{Node: node.Node, Computed: node.Computed}
			} else {
				ident := parse_prop_name(p)
				keys = append(keys, ident)
				var value *Node
				if p.notAt(COLON) {
					value = ident
				} else {
					p.next() // :
					value = p.parseExpr(DEFAULT_BP)
				}
				props[idx] = ObjectProp{Node: value, Computed: false, Accessor: false}
			}
		case at == STRING || at == NUMBER:
			key := p.parseExpr(PRIMARY_BP)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				props[idx] = ObjectProp{Node: parseFnFromParams(p, key, false, false, true, key.Loc), Computed: true, Accessor: false}
			} else {
				p.expect(COLON)
				value := p.parseExpr(DEFAULT_BP)
				props[idx] = ObjectProp{Node: value, Computed: true, Accessor: false}
			}
		case at == O_BRACKET:
			p.next() // [
			key := p.parseExpr(DEFAULT_BP)
			p.expect(C_BRACKET)
			keys = append(keys, key)
			if p.isAt(O_PAREN) {
				props[idx] = ObjectProp{Node: parseFnFromParams(p, key, false, false, true, key.Loc), Computed: true, Accessor: false}
			} else {
				p.expect(COLON)
				value := p.parseExpr(DEFAULT_BP)
				props[idx] = ObjectProp{Node: value, Computed: true, Accessor: false}
			}
		case at == K_async:
			node := parseFunMethod(p)
			keys = append(keys, node.Data.(FnDecl).Name)
			props[idx] = ObjectProp{Node: node.Node, Computed: node.Computed, Accessor: false}
		case at == DOT3:
			p.next()
			keys = append(keys, &Node{})
			operand := p.parse_exprs(T_OBJECT_LIT, T_IDENT, T_ARRAY_LIT)
			props[idx] = ObjectProp{Node: &Node{
				Tag:      T_RESTORSPREAD,
				Children: nil,
				// T_RESTORSPREAD always returns direct* Node.
				Data: operand,
				Loc:  p.loc(start, operand.Loc),
			}, Computed: false, Accessor: false}
		case at == K_get || at == K_set:
			fn := parseFunMethod(p)
			keys = append(keys, fn.Data.(Accessor).Data.(FnDecl).Name)
			props[idx] = ObjectProp{Node: fn.Node, Computed: fn.Computed, Accessor: true}
		default:
			p.parseError(SyntaxError, "Property key expected (identifier).")
		}
		// if shouldCheckComma && !p.eatComma() && p.notAt(C_BRACE) {
		// 	p.expect(COMMA)
		// }
		p.eatComma()
	}
	end := p.expect(C_BRACE).loc
	return &Node{
		Tag:      T_OBJECT_LIT,
		Children: nil,
		Data: ObjectLiteral{
			Props: props,
			Keys:  keys,
		},
		Loc: p.loc(start, end),
	}
}

func parse_array_lit(p *ModuleParser) *Node {
	elements := []*Node{}
	start := p.expect(O_BRACKET).loc
	for p.not_eof() && p.notAt(C_BRACKET) {
		if p.isAt(DOT3) {
			startLoc := p.eat().loc
			operand := p.parseExpr(ASSIGNMENT_BP)
			elements = append(elements, &Node{
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
	return &Node{
		Tag:      T_ARRAY_LIT,
		Children: elements,
		Data:     nil,
		Loc:      p.loc(start, end),
	}
}

func parse_ternary_expr(p *ModuleParser, condition *Node, bp BindingPower) *Node {
	p.next() // ?
	Then := p.parseExpr(LOGICAL_OR_BP)
	p.expect(COLON)
	Else := p.parseExpr(bp)
	return &Node{
		Tag:      T_TERNARY,
		Children: nil,
		Data:     TernaryExpr{Condition: condition, Then: Then, Else: Else},
		Loc:      p.loc(condition.Loc, Else.Loc),
	}
}

func parse_match_expr(p *ModuleParser) *Node {
	loc := p.expect(K_match).loc
	operand := p.parseCondition()
	matches := []*Node{}
	cases := map[NodeIndex][]*Node{}
	var elseCase []*Node
	p.expect(O_BRACE)
	for p.not_eof() && p.notAt(C_BRACE) {
		if p.isAt(K_else) {
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
	return &Node{
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

func parse_from_expr(p *ModuleParser) *Node {
	loc := p.expect(K_from).loc
	path := parse_string(p).Data.(string)
	return &Node{
		Tag:      T_FROMEXPR,
		Children: nil,
		Data:     FromExpr{p.parseScriptConcurrent(path)},
		Loc:      loc,
	}
}

func (p *ModuleParser) parseScriptConcurrent(path string) (resolvedPath string) {
	if p.builtin {
		resolvedPath = lib.ToLowerCase(lib.JoinPaths(
			// Directory of current library module.
			lib.DirOf(lib.PathFromKey(p.path)),
			// Other module to parse.
			path))
		// Only parse if module has not been parsed before.
		needs := GlobalParser.Imports.Len() == 0 || !GlobalParser.Imports.Has(resolvedPath)
		if needs {
			GlobalParser.Go(func() {
				newp := NewBuiltinsParser(path)
				newp.Parse()
				GlobalParser.Imports.Set(resolvedPath, newp.path)
				GlobalParser.mu.Lock()
				GlobalParser.Program.Modules[newp.path] = newp.OwnModule
				GlobalParser.mu.Unlock()
			})
		}
	} else {
		if lib.IsAbs(path) {
			resolvedPath = lib.ToLowerCase(path)
		} else {
			resolvedPath = lib.ToLowerCase(lib.JoinPaths(lib.DirOf(lib.PathFromKey(p.path)), path))
		}
		// Only parse if module has not been parsed before.
		shouldParse := GlobalParser.Imports.Len() == 0 || !GlobalParser.Imports.Has(resolvedPath)
		if shouldParse {
			GlobalParser.Go(func() {
				newp := NewModuleParser(resolvedPath, false, false)
				newp.Parse()
				GlobalParser.Imports.Set(resolvedPath, newp.path)
				GlobalParser.mu.Lock()
				GlobalParser.Program.Modules[newp.path] = newp.OwnModule
				GlobalParser.mu.Unlock()
			})
		}
	}
	return
}

func parse_void_expr(p *ModuleParser) *Node {
	return parse_operand_expr(p, K_void, UNARY_BP, T_VOID)
}

func parse_typeof_expr(p *ModuleParser) *Node {
	return parse_operand_expr(p, K_typeof, UNARY_BP, T_TYPEOFEXPR)
}

func parse_new_expr(p *ModuleParser) *Node {
	return parse_operand_expr(p, K_new, CALL_BP-1, T_NEWEXPR)
}

func parse_async_fn_expr(p *ModuleParser) *Node {
	if p.atType(1) == K_fn {
		return parseFunExpr(p)
	}
	start := p.expect(K_async).loc // async
	params := p.parse_params()
	end := p.expect(ARROW).loc
	var body []*Node
	if p.isAt(O_BRACE) {
		body = p.parseBlock()
	} else {
		body = []*Node{{
			Tag:      T_RETURN,
			Children: []*Node{p.parseExpr(DEFAULT_BP)},
			Data:     nil,
			Loc:      end,
		}}
	}
	return &Node{
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

func parse_operand_expr(p *ModuleParser, tktag TokenTag, bp BindingPower, nodeTag NodeTag) *Node {
	start := p.expect(tktag).loc
	operand := p.parseExpr(bp)
	return &Node{
		Tag:      nodeTag,
		Children: nil,
		Data:     OperandExpr{Operand: operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_member_expr(p *ModuleParser, object *Node, bp BindingPower) *Node {
	start := object.Loc
	computed := false
	member := &Node{}
	if p.isAt(O_BRACKET) {
		p.next()
		computed = true
		member = p.parseExpr(DEFAULT_BP)
		p.expect(C_BRACKET)
	} else {
		p.next() // .
		member = parse_prop_name(p)
	}
	return &Node{
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

func parse_call_expr(p *ModuleParser, caller *Node, bp BindingPower) *Node {
	loc := caller.Loc
	arguments := p.parse_args()
	return &Node{
		Tag:      T_CALLEXPR,
		Children: nil,
		Data:     CallExpr{Caller: caller, Args: arguments},
		Loc:      loc,
	}
}

func parse_comparison_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parseLRExpr(p, bp, left, T_COMPARE)
}

func parse_logical_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parseLRExpr(p, bp, left, T_LOGICAL)
}

func parse_not_expr(p *ModuleParser) *Node {
	start := p.eat().loc
	operand := p.parseExpr(DEFAULT_BP)
	return &Node{
		Tag:      T_NOT,
		Children: nil,
		Data:     OperandExpr{Operand: operand},
		Loc:      p.loc(start, operand.Loc),
	}
}

func parse_assignment_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parseLRExpr(p, bp, left, T_ASSIGN)
}

func parseLRExpr(p *ModuleParser, bp BindingPower, left *Node, tag NodeTag) *Node {
	op := p.eat().tag
	right := p.parseExpr(bp)
	return &Node{
		Tag:      tag,
		Children: nil,
		Data: LROpExpr{
			Lhs: left, Rhs: right, Op: op,
		},
		Loc: p.loc(left.Loc, right.Loc),
	}
}

func parse_binary_expr(p *ModuleParser, left *Node, bp BindingPower) *Node {
	return parseLRExpr(p, bp, left, T_BINARY_EXP)
}

func parse_open_paren(p *ModuleParser) *Node {
	start := p.eat().loc
	can_parse_fn := true
	if p.isAt(C_PAREN) {
		if p.atType(1) != ARROW {
			p.parseError(UnexpectedToken)
		}
		p.next()           // )
		end := p.eat().loc // =>
		var body []*Node
		if p.isAt(O_BRACE) {
			body = p.parseBlock()
		} else {
			body = []*Node{{
				Tag:      T_RETURN,
				Children: []*Node{p.parseExpr(DEFAULT_BP)},
				Data:     nil,
				Loc:      end,
			}}
		}
		return &Node{
			Tag:      T_FNDECL,
			Children: body,
			Data: FnDecl{
				Async:     false,
				Params:    []*Node{},
				Name:      StringNode(AnonymousName, lib.DumbyLoc),
				Arrow:     true,
				Anonymous: true,
			},
			Loc: p.loc(start, end),
		}
	}
	nodes := []*Node{}
	// TODO: enable default params
	valid_params := []NodeTag{T_IDENT, T_ARRAY_LIT, T_OBJECT_LIT}
	for p.notAt(C_PAREN) && p.not_eof() {
		var expr *Node
		if len(nodes) > 0 {
			if p.isAt(O_BRACE) {
				e, r := p.parse_object_or_destructuring()
				can_parse_fn = r
				expr = e
			} else if p.isAt(O_BRACKET) {
				e, r := p.parse_array_or_destructuring()
				can_parse_fn = r
				expr = e
			} else {
				expr = p.parseExpr(DEFAULT_BP)
			}
		} else {
			if p.isAt(O_BRACE) {
				e, r := p.parse_object_or_destructuring()
				can_parse_fn = r
				expr = e
			} else if p.isAt(O_BRACKET) {
				e, r := p.parse_array_or_destructuring()
				can_parse_fn = r
				expr = e
			} else {
				expr = p.parseExpr(BITOR_BP)
			}
		}
		nodes = append(nodes, expr)
		if can_parse_fn {
			can_parse_fn = lib.InSlice(valid_params, expr.Tag)
		}
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
		var body []*Node
		if p.isAt(O_BRACE) {
			body = p.parseBlock()
		} else {
			body = []*Node{{
				Tag:      T_RETURN,
				Children: []*Node{p.parseExpr(DEFAULT_BP)},
				Data:     nil,
				Loc:      end,
			}}
		}
		return &Node{
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
	return &Node{
		Tag:      T_GROUPING,
		Data:     nil,
		Children: nodes,
		Loc:      p.loc(start, end),
	}
}

func (p *ModuleParser) parseCompareList(list []*Node) *Node {
	for p.isAt(BITOR) && p.not_eof() {
		p.next()
		list = append(list, p.parseExpr(BITOR_BP))
	}
	p.expect(C_PAREN)
	return &Node{
		Tag:      T_COMPARE_LIST,
		Children: list,
		Data:     nil,
		Loc:      p.loc(list[0].Loc, list[len(list)-1].Loc),
	}
}

func parse_string(p *ModuleParser) *Node {
	loc := p.currLoc()
	tk := p.expect(STRING)
	s := p.src(tk)
	return &Node{
		Tag:  T_STRING,
		Data: s[1 : len(s)-1],
		Loc:  loc,
	}
}

func StringNode(str string, loc Loc) *Node {
	return &Node{
		Tag:  T_STRING,
		Data: str,
		Loc:  loc,
	}
}

func parse_number(p *ModuleParser) *Node {
	loc := p.currLoc()
	tk := p.expect(NUMBER)
	return &Node{
		Tag:  T_NUMBER,
		Data: lib.ParseNumber(p.src(tk)),
		Loc:  loc,
	}
}

func NumberNode(num float64, loc Loc) *Node {
	return &Node{
		Tag:  T_NUMBER,
		Data: num,
		Loc:  loc,
	}
}

func parse_prop_name(p *ModuleParser) *Node {
	loc := p.currLoc()
	var tk Token
	switch {
	case p.isAt(KEYWORD):
		tk = token(IDENTIFIER, p.eat().loc)
	default:
		tk = p.expect(IDENTIFIER)
	}
	return &Node{
		Tag:  T_IDENT,
		Data: Identifier(p.src(tk)),
		Loc:  loc,
	}
}

func parse_ident(p *ModuleParser) *Node {
	loc := p.currLoc()
	tk := p.expect(IDENTIFIER)
	return &Node{
		Tag:  T_IDENT,
		Data: Identifier(p.src(tk)),
		Loc:  loc,
	}
}

func (p *ModuleParser) parseError(errname ErrorName, additionals ...string) {
	name := "SyntaxError"
	message := ""
	// Use a string builder for efficient string concatenation and to
	// make it safe for concurrent parser errors
	b := lib.NewStringBuilder()
	b.WriteString("ParseError: ")
	switch errname {
	case UnexpectedToken:
		message = lib.Sprintf("Unexpected token reached: %s%s%s", p.at(0).tag.Lexeme(), lib.EOL, lib.DebugMsg(p.path, p.currLoc()))
	case SyntaxError:
		message = lib.Sprintf("%s%s%s", additionals[0], lib.EOL, lib.DebugMsg(p.path, p.currLoc()))
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
	// GlobalParser.errors.Push(b.String())
	panic(b.String())
}

func (p *ModuleParser) not_eof() bool {
	return p.atType(0) != EOF
}

func (p *ModuleParser) loc(startLoc, end Loc) Loc {
	return Loc{
		Line:  startLoc.Line,
		Col:   startLoc.Col,
		Start: startLoc.Start,
		End:   end.End,
	}
}

func (p *ModuleParser) currLoc() Loc {
	return p.at(0).loc
}

func (p *ModuleParser) src(tk Token) string {
	return p.lexer.src(tk)
}

func (p *ModuleParser) locSrc(loc Loc) string {
	return p.lexer.src(token(0, loc))
}

func (p *ModuleParser) at(idx int) Token {
	l := len(p.tokens)
	idx = int(p.tokenIndex) + idx
	if idx >= l || idx < 0 {
		lib.Panic("unexpected behaviour")
	}
	return p.tokens[idx]
}

func (p *ModuleParser) atType(idx int) TokenTag {
	return p.at(idx).tag
}

func (p *ModuleParser) next() {
	p.tokenIndex++
}

func (p *ModuleParser) eat() Token {
	defer p.next()
	return p.at(0)
}

func (p *ModuleParser) isAt(t TokenTag) bool {
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
	case KEYWORD:
		return currT > keyword_start && currT < keyword_end
	}
	return currT == t
}

func (p *ModuleParser) notAt(t TokenTag) bool {
	return !p.isAt(t)
}

func (p *Parser) debug(n *Node) {
	lib.Printf("Node \x1b[32m%s\x1b[0m [%d:%d, %d:%d]%s", n.Tag.Name(), n.Loc.Line, n.Loc.Col, n.Loc.Start, n.Loc.End, lib.EOL)
}

func (p *Parser) PrintNodes(count uint) {
	lib.Println(lib.Yellow("----------- AST Nodes ----------"))
	for _, n := range p.Program.Main.Body {
		p.debug(n)
	}
	i := uint(0)
	for _, m := range p.Program.Modules {
		if i < count && i < uint(len(p.Program.Modules)) {
			i++
			for _, n := range m.Body {
				p.debug(n)
			}
		} else {
			break
		}
	}
}

// func (p *Parser)
