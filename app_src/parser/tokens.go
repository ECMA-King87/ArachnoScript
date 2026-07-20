package parser

import "aspire/are/main/lib"

type TokenTag int8

const (
	LITERAL TokenTag = iota
	literal_start
	NUMBER
	STRING
	IDENTIFIER
	literal_end

	// symbols
	SEMICOLON
	COLON
	DOT
	COMMA
	O_PAREN
	C_PAREN
	O_BRACE
	C_BRACE
	O_BRACKET
	C_BRACKET
	DOT3

	// keywords
	DECL
	decl_keyword_start
	_fn
	_class
	VARDECL
	var_decl_keyword_start
	_let
	_var
	_const
	var_decl_keyword_end
	decl_keyword_end
	// modifiers
	_static
	_public
	_private

	_extends
	_async
	_await
	_try
	_catch
	_finally

	_if
	_else
	_while
	_break
	_continue
	_do
	_for
	_typeof
	_from
	_import
	_export
	_return
	_throw
	_of
	_in
	_default
	_switch
	_case
	_super
	_new
	_match
	_ctor

	ACCESSOR
	accessor_start
	_get
	_set
	accessor_end

	_fallthrough
	_void
	_instanceof
	// globalThis
	_gt
	_yield

	BINARYOP
	binary_op_start
	PLUS   // +
	MINUS  // -
	DIVIDE // /
	TIMES  // *
	MODULO // %
	STAR2  // **
	binary_op_end

	PLUS2  // ++
	MINUS2 // --

	assignment_op_start
	EQUALS         // =
	PLUS_EQUALS    // +=
	MINUS_EQUALS   // -=
	DIVIDE_EQUALS  // /=
	TIMES_EQUALS   // *=
	MODULO_EQUALS  // %=
	STAR2_EQUALS   // **=
	NULLISH_EQUALS // ??=
	assignment_op_end

	comparison_op_start
	EQUALS2     // ==
	EQUALS3     // ===
	NOT_EQUALS  // !=
	NOT_EQUALS2 // !==
	GT_EQUALS   // >=
	LT_EQUALS   // <=
	GT          // >
	LT          // <
	comaprison_op_end

	logical_op_start
	NOT
	AND
	OR
	logical_op_end

	BITOR    //  "|"
	BITAND   // "&"
	XOR      // "^"
	BITNOT   // "~"
	SHBITL   // "<<"
	SHBITR   // ">>"
	BITCLEAR // "&^"

	BITOR_EQUALS
	BITAND_EQUALS
	XOR_EQUALS
	SHBITL_EQUALS
	SHBITR_EQUALS
	BITCLEAR_EQUALS

	NULLISH
	QUESTION
	ARROW

	LABEL

	EOF
)

type Token struct {
	tag TokenTag
	loc lib.Loc
}

var lexemes map[TokenTag]string

func initLexemes() {
	lexemes = map[TokenTag]string{
		LITERAL:    "literal",
		NUMBER:     "number",
		STRING:     "string",
		IDENTIFIER: "identifier",
		SEMICOLON:  ";",
		COLON:      ":",
		DOT:        ".",
		DOT3:       "...",
		COMMA:      ",",
		O_PAREN:    "(",
		C_PAREN:    ")",
		O_BRACE:    "{",
		C_BRACE:    "}",
		O_BRACKET:  "[",
		C_BRACKET:  "]",
		LABEL:      "label",
		EOF:        "EOF",

		PLUS:     "binary op",
		MINUS:    "binary op",
		DIVIDE:   "binary op",
		TIMES:    "binary op",
		MODULO:   "binary op",
		STAR2:    "binary op",
		BINARYOP: "binary op",

		EQUALS:         "assignment op",
		PLUS_EQUALS:    "assignment op",
		MINUS_EQUALS:   "assignment op",
		DIVIDE_EQUALS:  "assignment op",
		TIMES_EQUALS:   "assignment op",
		MODULO_EQUALS:  "assignment op",
		STAR2_EQUALS:   "assignment op",
		NULLISH_EQUALS: "assignment op",

		NOT: "!",
		AND: "&&",
		OR:  "||",

		ARROW: "=>",

		EQUALS2:     "==",  // ==
		EQUALS3:     "===", // ===
		NOT_EQUALS:  "!=",
		NOT_EQUALS2: "!==",
		GT_EQUALS:   ">=", // >=
		LT_EQUALS:   "<=", // <=
		GT:          ">",
		LT:          "<",

		PLUS2:    "++",
		MINUS2:   "--",
		NULLISH:  "??",
		QUESTION: "?",

		BITOR:    "|",
		BITAND:   "&",
		XOR:      "^",
		BITNOT:   "~",
		SHBITL:   "<<",
		SHBITR:   ">>",
		BITCLEAR: "&^",

		BITOR_EQUALS:    "|=",
		BITAND_EQUALS:   "&=",
		XOR_EQUALS:      "^=",
		SHBITL_EQUALS:   "<<=",
		SHBITR_EQUALS:   ">>=",
		BITCLEAR_EQUALS: "&^=",

		// keywords
		_fn:      "keyword",
		_class:   "keyword",
		_let:     "keyword",
		_var:     "keyword",
		_const:   "keyword",
		_static:  "keyword",
		_public:  "keyword",
		_private: "keyword",
		_extends: "keyword",
		_async:   "keyword",
		_await:   "keyword",
		_try:     "keyword",
		_catch:   "keyword",
		_finally: "keyword",
		DECL:     "declaration keyword",
		VARDECL:  "variable declaration keyword",

		_if:          "keyword",
		_else:        "keyword",
		_while:       "keyword",
		_break:       "keyword",
		_continue:    "keyword",
		_do:          "keyword",
		_for:         "keyword",
		_typeof:      "keyword",
		_from:        "keyword",
		_import:      "keyword",
		_export:      "keyword",
		_return:      "keyword",
		_throw:       "keyword",
		_of:          "keyword",
		_in:          "keyword",
		_default:     "keyword",
		_switch:      "keyword",
		_case:        "keyword",
		_super:       "keyword",
		_new:         "keyword",
		_match:       "keyword",
		_ctor:        "keyword",
		_get:         "keyword",
		_set:         "keyword",
		_fallthrough: "keyword",
		_void:        "keyword",
		_instanceof:  "keyword",
		_gt:          "keyword",
	}
	for _, t := range keywords {
		lexemes[t] = "keyword"
	}
}

var keywords = map[string]TokenTag{
	"function": _fn,
	"class":    _class,
	"private":  _private,
	"public":   _public,
	"static":   _static,
	"extends":  _extends,
	"let":      _let,
	"var":      _var,
	"const":    _const,
	"async":    _async,
	"await":    _await,
	"try":      _try,
	"catch":    _catch,
	"finally":  _finally,

	"if":          _if,
	"else":        _else,
	"while":       _while,
	"break":       _break,
	"continue":    _continue,
	"do":          _do,
	"for":         _for,
	"typeof":      _typeof,
	"from":        _from,
	"import":      _import,
	"export":      _export,
	"return":      _return,
	"throw":       _throw,
	"of":          _of,
	"in":          _in,
	"default":     _default,
	"switch":      _switch,
	"case":        _case,
	"super":       _super,
	"new":         _new,
	"match":       _match,
	"constructor": _ctor,
	"get":         _get,
	"set":         _set,
	"fallthrough": _fallthrough,
	"void":        _void,
	"instanceof":  _instanceof,
	"globalThis":  _gt,
	"yield":       _yield,
}

// var unreserved = []TokenTag{
// 	_catch,
// }

func (tag TokenTag) Lexeme() string {
	if lexemes == nil {
		initLexemes()
	}
	if l, ok := lexemes[tag]; ok {
		return l
	}
	return "\x1b[32mLexeme\x1b[0m"
}
