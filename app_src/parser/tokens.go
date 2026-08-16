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

	KEYWORD
	keyword_start
	DECL
	decl_keyword_start
	K_fn
	K_class
	VARDECL
	var_decl_keyword_start
	K_let
	K_var
	K_const
	var_decl_keyword_end
	decl_keyword_end
	// modifiers
	K_static
	K_public
	K_private

	K_extends
	K_async
	K_await
	K_try
	K_catch
	K_finally

	K_if
	K_else
	K_while
	K_break
	K_continue
	K_do
	K_for
	K_typeof
	K_from
	K_import
	K_export
	K_return
	K_throw
	K_of
	K_in
	K_default
	K_switch
	K_case
	K_super
	K_new
	K_match
	K_ctor

	ACCESSOR
	accessor_start
	K_get
	K_set
	accessor_end

	K_fallthrough
	K_void
	K_instanceof
	// globalThis
	K_gt
	K_yield
	keyword_end

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
		K_fn:      "keyword",
		K_class:   "keyword",
		K_let:     "keyword",
		K_var:     "keyword",
		K_const:   "keyword",
		K_static:  "keyword",
		K_public:  "keyword",
		K_private: "keyword",
		K_extends: "keyword",
		K_async:   "keyword",
		K_await:   "keyword",
		K_try:     "keyword",
		K_catch:   "keyword",
		K_finally: "keyword",
		DECL:      "declaration keyword",
		VARDECL:   "variable declaration keyword",

		// K_if:          "keyword",
		// K_else:        "keyword",
		// K_while:       "keyword",
		// K_break:       "keyword",
		// K_continue:    "keyword",
		// K_do:          "keyword",
		// K_for:         "keyword",
		// K_typeof:      "keyword",
		// K_from:        "keyword",
		// K_import:      "keyword",
		// K_export:      "keyword",
		// K_return:      "keyword",
		// K_throw:       "keyword",
		// K_of:          "keyword",
		// K_in:          "keyword",
		// K_default:     "keyword",
		// K_switch:      "keyword",
		// K_case:        "keyword",
		// K_super:       "keyword",
		// K_new:         "keyword",
		// K_match:       "keyword",
		// K_ctor:        "keyword",
		// K_get:         "keyword",
		// K_set:         "keyword",
		// K_fallthrough: "keyword",
		// K_void:        "keyword",
		// K_instanceof:  "keyword",
		// K_gt:          "keyword",
	}
	for _, t := range keywords {
		lexemes[t] = "keyword"
	}
}

var keywords = map[string]TokenTag{
	"function": K_fn,
	"class":    K_class,
	"private":  K_private,
	"public":   K_public,
	"static":   K_static,
	"extends":  K_extends,
	"let":      K_let,
	"var":      K_var,
	"const":    K_const,
	"async":    K_async,
	"await":    K_await,
	"try":      K_try,
	"catch":    K_catch,
	"finally":  K_finally,

	"if":          K_if,
	"else":        K_else,
	"while":       K_while,
	"break":       K_break,
	"continue":    K_continue,
	"do":          K_do,
	"for":         K_for,
	"typeof":      K_typeof,
	"from":        K_from,
	"import":      K_import,
	"export":      K_export,
	"return":      K_return,
	"throw":       K_throw,
	"of":          K_of,
	"in":          K_in,
	"default":     K_default,
	"switch":      K_switch,
	"case":        K_case,
	"super":       K_super,
	"new":         K_new,
	"match":       K_match,
	"constructor": K_ctor,
	"get":         K_get,
	"set":         K_set,
	"fallthrough": K_fallthrough,
	"void":        K_void,
	"instanceof":  K_instanceof,
	"globalThis":  K_gt,
	"yield":       K_yield,
}

// var unreserved = []TokenTag{
// 	K_catch,
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
