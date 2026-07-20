package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
	"strings"
)

type DeclType int8

const (
	// Must be the same as implementation in parser.
	MutableDecl DeclType = iota
	ConstDecl
	VarDecl
)

const (
	FunctionScope lib.Enum = iota
	TryCatchScope
	GlobalScope
	ModuleScope
	SwitchScope
	BlockScope
	LoopScope
)

type Scope struct {
	memory []Value
	// A negative ptr means it is uninitialized.
	names             *lib.Map[string, int]
	varTypes          *lib.Map[string, DeclType]
	parent            *Scope
	worker            *Worker
	valid, in_promise bool
	of                lib.Enum

	// the scope must carry a path
	path       int
	objectHash uint64
	// object Value
}

func NewScope(parent *Scope, object uint64, of lib.Enum, path int) *Scope {
	var worker *Worker
	if parent != nil {
		if path == -1 {
			path = parent.path
		}
		worker = parent.worker
	}
	return &Scope{
		memory:     []Value{},
		names:      lib.NewMap[string, int](),
		varTypes:   lib.NewMap[string, DeclType](),
		parent:     parent,
		valid:      true,
		in_promise: false,
		worker:     worker,
		of:         of,
		path:       path,
		objectHash: object,
	}
}

func NewFunctionScope(function *Function, this *Object) *Scope {
	return NewScope(function.declScope, hashValue(this), FunctionScope, -1)
}

func (s *Scope) declare(name string, varType DeclType, loc parser.Loc) int {
	if s.names.Has(name) {
		s.error(SyntaxError, lib.Sprintf("Cannot redeclare variable '%s'.%s%s", name, lib.EOL, lib.SourceLog(s.path, loc)))
	}
	ptr := len(s.memory)
	s.memory = append(s.memory, undefined)
	s.names.Set(name, ptr)
	s.varTypes.Set(name, varType)
	return ptr
}

func (s *Scope) init(name string, value Value, varType DeclType, loc parser.Loc) Value {
	if s.path == -1 && s.of != GlobalScope {
		s.error(BuildError, "Scope path is uninitialized: "+name)
	}
	if s.names.Has(name) {
		s.error(SyntaxError, lib.Sprintf("Cannot redeclare variable '%s'.%s%s", name, lib.EOL, lib.SourceLog(s.path, loc)))
	}
	ptr := len(s.memory)
	s.names.Set(name, ptr)
	s.memory = append(s.memory, value)
	s.varTypes.Set(name, varType)
	return value
}

func (s *Scope) getValue(name string, loc *parser.Loc) Value {
	scope := s.resolve(name, s, loc)
	ptr, _ := scope.names.Get(name)
	if ptr == -1 {
		s.error(SyntaxError, "Variable '", name, "' used before being initialized.", lib.EOL, lib.SourceLog(s.path, *loc))
	}
	if ptr < 0 || ptr >= len(scope.memory) {
		if lib.DEBUG_MODE {
			s.error(BuildError, "Invalid memory access: index out of bounds for variable '", name, "'.", lib.EOL, lib.SourceLog(s.path, *loc))
		}
		return undefined
	}
	return scope.memory[ptr]
}

func (s *Scope) resolve(name string, oscope *Scope, loc *parser.Loc) *Scope {
	if s.names.Has(name) {
		return s
	}
	if s.parent == nil {
		oscope.error(ReferenceError, "Could not resolve name \x1b[32m", name, "\x1b[0m, it does not exist.", lib.EOL, lib.SourceLog(oscope.path, *loc))
	}
	return s.parent.resolve(name, oscope, loc)
}

func (s *Scope) findScopeWith(u uint64) *Scope {
	if s.objectHash == u {
		return s
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.findScopeWith(u)
}

func (s *Scope) scopeOf(t lib.Enum) *Scope {
	if s.of == t {
		return s
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.scopeOf(t)
}

// Looks up the scope of which name is declared and assigns value to the memory space if found.
func (s *Scope) assignName(name string, value Value, loc *parser.Loc) {
	declScope := s.resolve(name, s, loc)
	if t, _ := declScope.varTypes.Get(name); t == ConstDecl {
		s.errorWithSource(SyntaxError, s.path, *loc, lib.Sprintf("Cannot assign to '%s' because it is a constant.", name))
	}
	ptr, e := declScope.names.Get(name)
	if !e && lib.DEBUG_MODE {
		lib.Panic(lib.BUG_MESSAGE)
	}
	declScope.assign(ptr, value)
}

// Assigns the value to ptr to value in the memory of s.
func (s *Scope) assign(ptr int, value Value) {
	if ptr < 0 || ptr >= len(s.memory) {
		s.error(BuildError, "Invalid memory pointer: index out of bounds")
	}
	if p, ok := s.memory[ptr].(*Pointer); ok {
		if p.Scope != nil && p.Scope != s {
			p.assign(p.ptr, value)
		}
	}
	s.memory[ptr] = value
}

type ErrorName int

const (
	ReferenceError ErrorName = iota
	SyntaxError
	BuildError
	TypeError
	MathError
	Error
)

// path can be negative to use the s's path.
func (s *Scope) errorWithSource(name ErrorName, path int, loc parser.Loc, args ...string) {
	if path < 0 {
		path = s.path
	}
	args = append(args, lib.EOL, lib.SourceLog(path, loc))
	s.error(name, args...)
}

// NOTE: Adds a newline at the end
func (s *Scope) error(name ErrorName, msg ...string) {
	b := lib.NewStringBuilder()
	errName := ""
	switch name {
	case ReferenceError:
		errName = "ReferenceError"
	case SyntaxError:
		errName = "SyntaxError"
	case TypeError:
		errName = "TypeError"
	case BuildError:
		errName = "BuildError"
	case MathError:
		errName = "MathError"
	default:
		errName = "Error"
	}
	inPromise := ""
	if s.in_promise {
		inPromise = "(in promise) "
	}
	b.WriteString(lib.Sprintf("Uncaught %s\x1b[31m%s\x1b[0m: ", inPromise, errName))
	for _, m := range msg {
		b.WriteString(m)
	}
	b.WriteString(lib.EOL)
	s.throw(&b)
}

func (s *Scope) throw(builder *strings.Builder) {
	s.invalidate()
	if s.parent == nil {
		lib.Stdout.WriteString(builder.String())
		s.worker.crash()
	}
	s.parent.throw(builder)
}

func (s *Scope) invalidate() {
	clear(s.memory)
	s.names.Clear()
	s.varTypes.Clear()
	s.valid = false
}
