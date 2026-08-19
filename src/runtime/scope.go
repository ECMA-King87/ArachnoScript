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

type ScopeKind lib.Enum

const (
	FunctionScope ScopeKind = iota
	TryScope
	GlobalScope
	ModuleScope
	SwitchScope
	BlockScope
	ClassScope
	LoopScope
)

type Scope struct {
	memory []Value
	// A negative ptr means it is uninitialized.
	names             *lib.Map[*String, int]
	varTypes          *lib.Map[*String, DeclType]
	parent            *Scope
	worker            *Worker
	valid, in_promise bool
	of                ScopeKind

	// the scope must carry a path
	path   int
	object Value
}

func NewScope(parent *Scope, object Value, of ScopeKind, path int) *Scope {
	var worker *Worker
	if parent != nil {
		if path == 0 {
			path = parent.path
		}
		worker = parent.worker
	} else if path == 0 {
		path = -1
	}
	scope := &Scope{
		memory:     []Value{},
		names:      lib.NewMap[*String, int](),
		varTypes:   lib.NewMap[*String, DeclType](),
		parent:     parent,
		valid:      true,
		in_promise: false,
		worker:     worker,
		of:         of,
		path:       path,
		object:     object,
	}
	scope.init(ThisStr, object, ConstDecl, lib.DumbyLoc)
	return scope
}

func NewFunctionScope(function *Function, this Value) *Scope {
	return NewScope(function.declScope, this, FunctionScope, 0)
}

func (s *Scope) declare(name *String, varType DeclType, loc parser.Loc) int {
	if s.names.Has(name) {
		s.errorWithSource(SyntaxError, 0, loc, lib.Sprintf("Cannot redeclare variable '%s'.", name.string))
	}
	ptr := len(s.memory)
	s.memory = append(s.memory, undefined)
	s.names.Set(name, ptr)
	s.varTypes.Set(name, varType)
	return ptr
}

func (s *Scope) init(name *String, value Value, varType DeclType, loc parser.Loc) Value {
	if s.path == 0 && s.of != GlobalScope {
		s.raise(BuildError, "Scope path is uninitialized: "+name.string)
	}
	if s.names.Has(name) {
		s.errorWithSource(SyntaxError, 0, loc, lib.Sprintf("Cannot redeclare variable '%s'.", name.string))
	}
	ptr := len(s.memory)
	s.memory = append(s.memory, value)
	s.names.Set(name, ptr)
	s.varTypes.Set(name, varType)
	return value
}

func (s *Scope) getValue(name *String, loc *parser.Loc) Value {
	scope := s.resolve(name, s, loc)
	ptr, _ := scope.names.Get(name)
	if ptr < 0 {
		s.errorWithSource(SyntaxError, 0, *loc, lib.Sprintf("Variable '%s' used before being initialized.", name.string))
	}
	if ptr >= len(scope.memory) {
		if lib.DEBUG_MODE {
			s.errorWithSource(BuildError, 0, *loc, lib.Sprintf("Invalid memory access: index out of bounds for variable '%s'.", name.string))
		}
		return undefined
	}
	return scope.memory[ptr]
}

func (s *Scope) resolve(name *String, oscope *Scope, loc *parser.Loc) *Scope {
	if s.names.Has(name) {
		return s
	}
	if s.parent == nil {
		oscope.errorWithSource(ReferenceError, 0, *loc, lib.Sprintf("Could not resolve name \x1b[32m%s\x1b[0m, it does not exist.", name.string))
	}
	return s.parent.resolve(name, oscope, loc)
}

func (s *Scope) findScopeWith(v Value) *Scope {
	// Compare Identity, most values are pointers.
	if s.object == v {
		return s
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.findScopeWith(v)
}

func (s *Scope) scopeOf(t ScopeKind) *Scope {
	if s.of == t {
		return s
	}
	if s.parent == nil || (s.of == FunctionScope && s.parent.of != ClassScope) {
		return nil
	}
	return s.parent.scopeOf(t)
}

// Looks up the scope of which name is declared and assigns value to the memory space if found.
func (s *Scope) assignName(name *String, value Value, loc *parser.Loc) {
	declScope := s.resolve(name, s, loc)
	if t, _ := declScope.varTypes.Get(name); t == ConstDecl {
		s.errorWithSource(SyntaxError, s.path, *loc, lib.Sprintf("Cannot assign to '%s' because it is a constant.", name.string))
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
		s.raise(BuildError, "Invalid memory pointer: index out of bounds")
	}
	s.memory[ptr] = value
}

type ErrorName int

const (
	ReferenceError ErrorName = iota
	SyntaxError
	BuildError
	RangeError
	TypeError
	MathError
	Error
)

// path may be zero to use the scope's own path.
func (s *Scope) errorWithSource(name ErrorName, path int, loc parser.Loc, args ...string) {
	if path == 0 {
		path = s.path
	}
	b := lib.NewStringBuilder()
	for _, m := range args {
		b.WriteString(m)
	}
	s.error(name, path, loc, NewString(b.String()))
}

// path may be zero to use the scope's own path.
func (s *Scope) raiseWithTrace(name ErrorName, path int, loc parser.Loc, args ...string) {
	if path == 0 {
		path = s.path
	}
	b := lib.NewStringBuilder()
	for _, m := range args {
		b.WriteString(m)
	}
	s.errorWithTrace(name, path, loc, NewString(b.String()))
}

// Does not infer path. 0 means no source.
func (s *Scope) error(name ErrorName, path int, loc parser.Loc, value Value) {
	if s != nil {
		s.invalidate()
		if s.of == ModuleScope || s.parent == nil {
			b := lib.NewStringBuilder()
			errName := resolveErrName(name)
			inPromise := ""
			if s.in_promise {
				inPromise = "(in promise) "
			}
			b.WriteString(lib.Sprintf("Uncaught %s\x1b[31m%s\x1b[0m: ", inPromise, errName))
			b.WriteString(s.worker.Inspect(value))
			if path != 0 {
				b.WriteString(lib.EOL)
				b.WriteString(lib.DebugMsg(path, loc))
			}
			b.WriteString(lib.EOL)
			lib.Stdout.WriteString(b.String())
			s.worker.crash()
		} else if s.of == TryScope {
			s.invalidate()
			s.worker.stack.Push(value)
		} else {
			s.parent.error(name, path, loc, value)
		}
	}
}

// Does not infer path. 0 means no source.
func (s *Scope) errorWithTrace(name ErrorName, path int, loc parser.Loc, value Value) {
	if s != nil {
		s.invalidate()
		if s.of == ModuleScope || s.parent == nil {
			b := lib.NewStringBuilder()
			errName := resolveErrName(name)
			inPromise := ""
			if s.in_promise {
				inPromise = "(in promise) "
			}
			b.WriteString(lib.Sprintf("Uncaught %s\x1b[31m%s\x1b[0m: ", inPromise, errName))
			b.WriteString(s.worker.Inspect(value))
			b.WriteString(lib.EOL)
			ctx := NewContext(s)
			if ctx.callstack.Len() > 0 {
				b.WriteString(newStackTrace(10, ctx).string)
			} else if path != 0 {
				b.WriteString(lib.DebugMsg(path, loc))
			}
			// b.WriteString(lib.EOL)
			lib.Stdout.WriteString(b.String())
			s.worker.crash()
		} else if s.of == TryScope {
			s.invalidate()
			s.worker.stack.Push(value)
		} else {
			s.parent.errorWithTrace(name, path, loc, value)
		}
	}
}

func resolveErrName(name ErrorName) string {
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
	case RangeError:
		errName = "RangeError"
	default:
		errName = "Error"
	}
	return errName
}

// NOTE: Adds a newline at the end
// Does not support error handling.
// Used for certain errors like BuildErrors.
func (s *Scope) raise(name ErrorName, msg ...string) {
	b := lib.NewStringBuilder()
	errName := resolveErrName(name)
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
	if s.of == ModuleScope || s.parent == nil {
		lib.Stdout.WriteString(builder.String())
		s.worker.crash()
		// } else if s.of == TryScope {
		// 	s.invalidate()
		// 	s.worker.stack.Push(NewString(builder.String()))
	} else {
		s.invalidate()
		s.parent.throw(builder)
	}
}

func (s *Scope) invalidate() {
	if s != nil {
		clear(s.memory)
		s.names.Clear()
		s.varTypes.Clear()
		s.valid = false
	}
}
