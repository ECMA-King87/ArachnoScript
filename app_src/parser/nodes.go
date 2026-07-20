package parser

import "aspire/are/main/lib"

type BindingPower int

const (
	DEFAULT_BP BindingPower = iota
	ASSIGNMENT_BP
	TERNARY_BP
	LOGICAL_OR_BP
	LOGICAL_AND_BP
	NULLISH_COALLESCE_BP
	BITOR_BP
	BITXOR_BP
	BITAND_BP
	EQUALITY_BP
	COMPARISON_BP
	SHIFT_BP
	BITCLEAR_BP
	ADDITIVE_BP
	MULTIPLICATIVE_BP
	UNARY_BP
	POSTFIX_BP
	CALL_BP
	MEMBER_BP
	PRIMARY_BP
)

type stmtHandler func(*Parser) Node
type nudHandler func(*Parser) Node
type ledHandler func(*Parser, Node, BindingPower) Node

// type typeNudHandler func(*Parser) Node
// type typeLedHandler func(*Parser, Node, BindingPower) Node

var bpLU map[TokenTag]BindingPower = map[TokenTag]BindingPower{}

// var typeBpLU map[TokenTag]BindingPower = map[TokenTag]BindingPower{}

var stmtLU map[TokenTag]stmtHandler = map[TokenTag]stmtHandler{}

// expects nothing to its left
var nudLU map[TokenTag]nudHandler = map[TokenTag]nudHandler{}

// expects a token to its left
var ledLU map[TokenTag]ledHandler = map[TokenTag]ledHandler{}

// var typeNudLU map[TokenTag]typeNudHandler = map[TokenTag]typeNudHandler{}
// var typeLedLU map[TokenTag]typeLedHandler = map[TokenTag]typeLedHandler{}

// func defTypeNud(k TokenTag, h typeNudHandler) {
// 	typeNudLU[k] = h
// }

// func defTypeLed(k TokenTag, h typeLedHandler, bp BindingPower) {
// 	typeBpLU[k] = bp
// 	typeLedLU[k] = h
// }

func defStmt(kind TokenTag, h stmtHandler) {
	stmtLU[kind] = h
}

func defLed(kind TokenTag, h ledHandler, bp BindingPower) {
	ledLU[kind] = h
	bpLU[kind] = bp
}

func defNud(kind TokenTag, h nudHandler) {
	nudLU[kind] = h
}

type NodeTag uint

func (t NodeTag) Name() string {
	if nodeTags == nil {
		initNodeTags()
	}
	return nodeTags[t]
}

// only initialize upon first use
var nodeTags map[NodeTag]string

// var NODETAGS_INIT = false

func initNodeTags() {
	nodeTags = map[NodeTag]string{
		T_PROGRAM:      "Program",
		T_MODULE:       "Module",
		T_IDENT:        "Identifier",
		T_NUMBER:       "Number",
		T_STRING:       "String",
		T_VARDECL:      "Var Decl",
		T_BINARY_EXP:   "Binary Expr",
		T_IFSTMT:       "If Stmt",
		T_BLOCK:        "Block",
		T_ASSIGN:       "Assignment Expr",
		T_COMPARE:      "Comparison Expr",
		T_COMPARE_LIST: "Comparison List",
		T_LOGICAL:      "Logical Expr",
		T_NOT:          "Not Expr",
		T_CALLEXPR:     "Call Expr",
		T_NEWEXPR:      "New Expr",
		T_TERNARY:      "Ternary Expr",
		T_VOID:         "Void Expr",
		T_BITNOT:       "Bitwise Not Expr",
		T_XOR:          "XOR Expr",
		T_BITAND:       "Bitwise And Expr",
		T_BITOR:        "Bitwise Or Expr",
		T_SHBITL:       "Bitwise Shift Left Expr",
		T_SHBITR:       "Bitwise Shift Right Expr",
		T_MEMBER:       "Member Expr",
		T_GROUPING:     "Grouping Expr",
		T_OBJECT_LIT:   "Object Literal",
		T_ARRAY_LIT:    "Array Literal",
		T_TYPEOFEXPR:   "Typeof Exp",
		T_INCREEXPR:    "Increment Expr",
		T_INSTANCEOF:   "Instanceof Expr",
		T_MATCHEXPR:    "Match Expr",
		T_YIELDSTMT:    "Yield Expr",
		T_FROMEXPR:     "From Expr",
		T_GT:           "globalThis",
		T_NULLISH:      "Nullish Coallesce Expr",
		T_RESTORSPREAD: "Rest Or Spread Expr",
		T_SUPER:        "Super Call",
		T_AWAIT:        "Await Expr",
		T_UNARY:        "Unary Expr",
		T_LABEL:        "Label Stmt",
		T_THROW:        "Throw Stmt",
		T_FNDECL:       "Fun Decl",
		T_WHILE:        "While Stmt",
		T_RETURN:       "Return Stmt",
		T_TFORLOOP:     "Traditional For Loop",
		T_FORLOOP:      "For Loop",
		T_IMPORT:       "Import Stmt",
		T_EXPORT:       "Export Stmt",
		T_CLASSDECL:    "Class Decl",
		T_TRYCATCH:     "Try-Catch-Finally Block",
		T_FALLTHROUGH:  "Fallthrough Stmt",
		T_BREAKSTMT:    "Break Stmt",
		T_CONTINUESTMT: "Continue Stmt",
		T_ACCESSOR:     "Function Decl",
	}
}

const (
	T_INVALID NodeTag = iota
	T_PROGRAM
	T_MODULE
	T_IDENT
	T_NUMBER
	T_STRING
	T_BINARY_EXP
	T_VARDECL
	T_IFSTMT
	T_BLOCK
	T_ASSIGN
	T_COMPARE
	T_COMPARE_LIST
	T_LOGICAL
	T_NOT
	T_CALLEXPR
	T_NEWEXPR
	T_TERNARY
	T_VOID
	T_MEMBER
	T_GROUPING
	T_OBJECT_LIT
	T_ARRAY_LIT
	T_TYPEOFEXPR
	T_INCREEXPR
	T_BITNOT
	T_XOR
	T_BITAND
	T_BITOR
	T_SHBITL
	T_SHBITR
	T_INSTANCEOF
	T_MATCHEXPR
	T_FROMEXPR
	T_GT
	T_NULLISH
	T_BITCLEAR
	T_RESTORSPREAD
	T_SUPER
	T_AWAIT
	T_YIELDSTMT
	T_UNARY
	T_LABEL
	T_THROW
	T_FNDECL
	T_RETURN
	T_WHILE
	T_TFORLOOP
	T_FORLOOP
	T_SWITCHSTMT
	T_IMPORT
	T_EXPORT
	T_CLASSDECL
	T_TRYCATCH
	T_FALLTHROUGH
	T_BREAKSTMT
	T_CONTINUESTMT
	T_ACCESSOR
)

type Program struct {
	Main    Module
	Modules []*Module
}

type Module struct {
	Path int
	Body []Node
}

type Node struct {
	Tag      NodeTag
	Children []Node
	Data     any
	Loc      lib.Loc
}

// -----------------------------------
// type Number float64
type Identifier string

type NewExpr struct {
	Operand Node
}

type TernaryExpr struct {
	Condition Node
	Then      Node
	Else      Node
}

type CallExpr struct {
	Caller Node
	Args   []Node
}

type OperandExpr struct {
	Operand Node
}

type LROpExpr struct {
	Lhs Node
	Rhs Node
	Op  TokenTag
}

type MemberExpr struct {
	Object   Node
	Member   Node
	Computed bool
}

type IncreExpr struct {
	Op      TokenTag
	Operand Node
	Pre     bool
}

type Decl struct {
	Lhs Node
	Rhs Node
	// Type TypeAnnotation
}

type NodeIndex = int
type ObjectProp struct {
	Computed bool
	Accessor bool
	Node
}

type ObjectDest struct {
	Props map[NodeIndex]ObjectDestProp
	// Keys are either (String | Identifier | Number | Computed)
	Keys []Node
}

type ObjectDestProp struct {
	Computed bool
	// Value
	Node
	Default Node
}

type ObjectLiteral struct {
	Props map[NodeIndex]ObjectProp
	// Keys are either (String | Identifier | Number | Computed)
	Keys []Node
}

type MatchExpr struct {
	Operand Node
	Cases   map[NodeIndex][]Node
	Matches []Node
	Else    []Node
}

type FromExpr struct {
	Path string
}

// type ArrayLiteral struct {
// 	Elements []Node
// }

type DeclKind int8

const (
	MutableDecl DeclKind = iota
	ConstantDecl
	HoistedDecl
)

type VarDecl struct {
	Kind  DeclKind
	Decls []Decl
}

type IfStmt struct {
	Condition Node
	ElseBlock []Node
}

type FnDecl struct {
	Async     bool
	Params    []Node
	Name      Node
	Arrow     bool
	Anonymous bool
}

type WhileLoop struct {
	Do        bool
	Body      []Node
	Condition Node
}

type TForLoop struct {
	Before    Node
	Condition Node
	AfterExec Node
}

const (
	In_Op int = iota
	Of_Op
)

type ForLoop struct {
	LHS Node
	RHS Node
	Op  int
	DeclKind
}

type SwitchStmt struct {
	Condition Node
	Cases     map[NodeIndex][]Node
	Matches   []Node
	Default   []Node
}

type ImportStmt struct {
	Namespace         string
	Named             Node // Object Literal Node
	From              string
	UseCurrentContext bool
}

type ExportStmt struct {
	Exp Node
}

type ClassDecl struct {
	Anonymous   bool
	DefaultProp Node
	Methods     []Node
	Props       []Node
	Constructor Node
	Extends     Node
	Name        string
}

type TryCatch struct {
	try     []Node
	catch   []Node
	capture Node
	finally []Node
}

type Accessor struct {
	Getter bool
	// Function
	Node
}
