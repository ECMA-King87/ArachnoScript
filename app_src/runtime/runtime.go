package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
)

type (
	Value interface {
		typeof() string
		toString() string
		inspect(uint, ValueInspector) string
	}
	Boolean bool
	Number  float64
	String  struct {
		string
	}
	Undefined struct{}
	Array     struct {
		hash     uint64
		elements *lib.Array[Value]
	}
	PropertyDescriptor struct {
		value                                       Value
		configurable, enumerable, writeable, public bool
		getter                                      *Function
		setter                                      *Function
	}
	Object struct {
		hash uint64
		// key: comparable values (number | string | symbol), value: any
		own   *lib.Map[*String, PropertyDescriptor]
		proto *Object
	}
	Instance struct {
		name string
		*Object
	}
	Function struct {
		name         string
		params, body []parser.Node
		async        bool
		arrow        bool
		declScope    *Scope
		*Object
		parser.Loc
	}
	Args          []Value
	MacroCallback func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value
	Macro         struct {
		name  string
		macro MacroCallback
	}
	Class struct {
		name string
		*Object
	}
	Symbol struct {
		description string
	}
	ScopeObject struct {
		*Scope
		*Object
	}
	Pointer struct {
		ptr int
		*Scope
		// ptr lib.Pointer
	}
	// RAW[T any] struct {
	// 	value T
	// }
	RAW struct {
		value any
	}
	// RAW any
)

const (
	// comparable types
	NumberType    = "number"
	StringType    = "string"
	SymbolType    = "symbol"
	UndefinedType = "undefined"
	BooleanType   = "boolean"
	//
	ObjectType   = "object"
	FunctionType = "function"
	ClassType    = "class"
	MacroType    = "macro"
	ArrayType    = "array"
	InstanceType = "instance"
	ScopeType    = "scope"
	RawType      = "raw"
)

var ComparableTypes = []string{NumberType, StringType, SymbolType, BooleanType, UndefinedType}

// func toValidPropKey(v Value) string {
// 	return v.toString()
// }

// ############# Number #############

var (
	Infinity = Number(lib.Infinity)
	NaN      = Number(lib.NaN)
)

func (n Number) typeof() string {
	return NumberType
}

func (n Number) IsNaN() bool {
	return n != n
}

func (n Number) IsFinite() bool {
	return !n.IsNaN() &&
		n != Infinity &&
		n != -Infinity
}

func (n Number) toString() string {
	if n == Infinity {
		return "Infinity"
	} else if n.IsNaN() {
		return "NaN"
	}
	return lib.Sprint(n)
}

func (n Number) inspect(uint, ValueInspector) string {
	return lib.Yellow(n.toString())
}

// ############# String #############
var GlobalStringIntern = &lib.InternTable[string, *String]{
	Table: make(map[string]*String, 4096),
}

var emptyString String

func NewString(s string) *String {
	if s == "" {
		return &emptyString
	}
	GlobalStringIntern.Mu.RLock()
	if v, ok := GlobalStringIntern.Table[s]; ok {
		GlobalStringIntern.Mu.RUnlock()
		return v
	}
	GlobalStringIntern.Mu.RUnlock()
	GlobalStringIntern.Mu.Lock()
	defer GlobalStringIntern.Mu.Unlock()
	// double check
	if v, ok := GlobalStringIntern.Table[s]; ok {
		return v
	}
	str := &String{s}
	GlobalStringIntern.Table[s] = str
	return str
}

func (str *String) typeof() string {
	return StringType
}

func (str *String) toString() string {
	return str.string
}

func (str *String) inspect(d uint, _ ValueInspector) string {
	if d == 0 {
		return str.string
	}
	return lib.Green(lib.Quote(str.string))
}

// ############# Boolean #############

var True Boolean = true
var False Boolean = false

func (Boolean) typeof() string {
	return BooleanType
}

func (b Boolean) toString() string {
	return lib.Sprint(b)
}

func (b Boolean) inspect(uint, ValueInspector) string {
	return lib.Yellow(b.toString())
}

// ############# Undefined #############

var undefined *Undefined = &Undefined{}

func (*Undefined) typeof() string {
	return UndefinedType
}

func (*Undefined) toString() string {
	return "undefined"
}

func (*Undefined) inspect(uint, ValueInspector) string {
	return lib.Bold(lib.White("undefined"))
}

// ############# Null #############

// A nil *Object value is null.
var null *Object

// ############# Object #############

func (obj *Object) typeof() string {
	return ObjectType
}

func (obj *Object) toString() string {
	return "[object]"
}

func (obj *Object) inspect(d uint, inspector ValueInspector) string {
	if obj == nil {
		return lib.Bold(lib.White("null"))
	}
	if obj.own.Len() == 0 {
		return "{}"
	}
	thisPtr := lib.PointerOf(obj)
	// Circular reference
	// if inspector.cyclic[thisPtr] {
	if _, ok := inspector.stack[thisPtr]; ok {
		return lib.Cyan("[Circular: *" + lib.Sprint(inspector.ids[thisPtr]) + "]")
	}
	id := 0
	if v, ok := inspector.ids[thisPtr]; ok {
		id = v
	}
	b := lib.NewStringBuilder()
	if id > 0 {
		b.WriteString(lib.Cyan("<ref *" + lib.Sprint(id) + "> "))
	}
	b.WriteRune('{')
	indent := obj.own.Len() > 3
	prefix := " "
	if indent {
		prefix = lib.EOL + lib.Repeat(" ", int(d+1))
	}
	inspector.stack[thisPtr] = struct{}{}
	defer delete(inspector.stack, thisPtr)
	obj.own.ForEach(func(key *String, pd PropertyDescriptor, last bool) {
		b.WriteString(prefix)
		name := key.string
		if isValidIdentifier(name) {
			b.WriteString(name)
		} else {
			b.WriteString(lib.Green(lib.Quote(name)))
		}
		b.WriteString(": ")
		b.WriteString(pd.value.inspect(d+1, inspector))
		if last {
			if indent {
				b.WriteString(lib.EOL)
			} else {
				b.WriteRune(' ')
			}
		} else {
			b.WriteRune(',')
		}
	})
	b.WriteRune('}')
	return b.String()
}

func (obj *Object) set(key *String, val Value) {
	owner, desc, found := lookupProp(obj, key, nil)
	if found {
		desc.value = val
		owner.own.Set(key, desc)
	} else {
		obj.own.Set(key, DefaultPropDesc(val))
	}
	hashValue(obj)
}

func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		if i == 0 {
			if !(lib.IsAlpha(r) || r == '_' || r == '$') {
				return false
			}
		} else {
			if !(lib.IsAlpha(r) ||
				lib.IsDigit(r) ||
				r == '_' ||
				r == '$') {
				return false
			}
		}
	}

	return true
}

func NewObject() *Object {
	return &Object{
		hash:  0,
		own:   lib.NewMap[*String, PropertyDescriptor](),
		proto: null,
	}
}

func DefaultPropDesc(value Value) PropertyDescriptor {
	return PropertyDescriptor{
		value:        value,
		configurable: true,
		enumerable:   true,
		public:       true,
		writeable:    true,
		getter:       nil,
		setter:       nil,
	}
}

func PropDesc(value Value,
	configurable, enumerable, writeable, public bool,
	getter, setter *Function) PropertyDescriptor {
	return PropertyDescriptor{
		value:        value,
		configurable: configurable,
		enumerable:   enumerable,
		writeable:    writeable,
		public:       public,
		getter:       getter,
		setter:       setter,
	}
}

// ############# Instance #############

func (inst *Instance) typeof() string {
	return InstanceType
}

func (inst *Instance) toString() string {
	return "[object]"
}

func (inst *Instance) inspect(d uint, Inspector ValueInspector) string {
	return lib.Green(inst.name) + " " + inst.Object.inspect(d, Inspector)
}

// ############# Symbol #############

func (sym *Symbol) typeof() string {
	return SymbolType
}

func (sym *Symbol) toString() string {
	return "Symbol(" + sym.description + ")"
}

func (sym *Symbol) inspect(d uint, _ ValueInspector) string {
	return lib.Green(sym.toString())
}

// ############# ScopeObject #############

func (so *ScopeObject) typeof() string {
	return ScopeType
}

func (so *ScopeObject) toString() string {
	return "[object (Scope)]"
}

func (so *ScopeObject) inspect(d uint, Inspector ValueInspector) string {
	return scopeToObject(so).inspect(d, Inspector)
}

func scopeToObject(so *ScopeObject) *Object {
	if so.Object == nil {
		so.Object = NewObject()
	}
	so.names.ForEach(func(name string, ptr int, _ bool) {
		k := NewString(name)
		if ptr < 0 || ptr >= len(so.memory) {
			so.Object.own.Set(k, DefaultPropDesc(undefined))
		} else {
			so.Object.own.Set(k, DefaultPropDesc(so.memory[ptr]))
		}
	})
	return so.Object
}

// ############# Function #############

func (fn *Function) typeof() string {
	return FunctionType
}

func (fn *Function) toString() string {
	return "[function " + fn.name + "]"
}

func (fn *Function) inspect(d uint, _ ValueInspector) string {
	return lib.Cyan(fn.toString())
}

// ############# Class #############

func (class *Class) typeof() string {
	return ClassType
}

func (class *Class) toString() string {
	return "[class " + class.name + "]"
}

func (class *Class) inspect(d uint, _ ValueInspector) string {
	return lib.Cyan(class.toString())
}

// ############# Macro #############

func (m *Macro) typeof() string {
	return MacroType
}

func (m *Macro) toString() string {
	return "[macro]"
}

func (m *Macro) inspect(d uint, _ ValueInspector) string {
	return lib.Magenta("[macro " + m.name + "]")
}

func (m *Macro) call(args Args, ctx *EvalContext, loc *lib.Loc) Value {
	return m.macro(args, ctx, loc, m)
}

func MK_MACRO(name string, f MacroCallback) *Macro {
	return &Macro{
		name:  name,
		macro: f,
	}
}

// ############## Array ##############

func (arr *Array) typeof() string {
	return ArrayType
}

func (arr *Array) toString() string {
	return "[array]"
}

func (arr *Array) inspect(d uint, Inspector ValueInspector) string {
	arrLen := arr.elements.Len()
	if arrLen == 0 {
		return "[]"
	}

	b := lib.NewStringBuilder()
	els := make([]string, 0, arrLen)
	// the 2nd, 3rd and 4th columns
	maxCol := [4]int{0, 0, 0, 0}

	thisPtr := lib.PointerOf(arr)
	Inspector.stack[thisPtr] = struct{}{}
	defer delete(Inspector.stack, thisPtr)

	if Inspector.cyclic[thisPtr] {
		id := Inspector.ids[thisPtr]
		b.WriteString(lib.Cyan("<ref *" + lib.Sprint(id) + "> "))
	}

	shouldFormatPretty := true

	arr.elements.ForEach(func(i int, v Value) {
		var str string
		vPtr := lib.PointerOf(v)
		if _, ok := Inspector.stack[vPtr]; ok {
			id := 0
			if v, ok := Inspector.ids[vPtr]; ok {
				id = v
			}
			str = lib.Cyan("[Circular: *" + lib.Sprint(id) + "]")
		} else {
			str = v.inspect(d+1, Inspector)
		}
		els = append(els, str)
		l := visibleLen(str)
		if l > 20 {
			shouldFormatPretty = false
		} else {
			col := i % 4
			maxCol[col] = max(l, maxCol[col])
		}
	})
	shouldIndent := arrLen > 5
	b.WriteString("[ ")
	if shouldIndent {
		b.WriteString(lib.EOL)
	}
	for i, str := range els {
		col := i % 4
		strLen := visibleLen(str)
		if shouldIndent {
			if shouldFormatPretty && col == 0 {
				b.WriteString(lib.Repeat("  ", int(d+1)))
			} else {
				b.WriteString(" ")
			}
			if shouldFormatPretty {
				padLen := maxCol[col] - strLen
				if padLen > 0 {
					b.WriteString(lib.Repeat(" ", padLen))
				}
			}
			b.WriteString(str)
			if i != arrLen-1 {
				b.WriteString(",")
			}
			if (shouldFormatPretty && col == 3) || (!shouldFormatPretty && shouldIndent) || i == arrLen-1 {
				b.WriteString(lib.EOL)
			}
		} else {
			b.WriteString(str)
			if i != arrLen-1 {
				b.WriteString(", ")
			} else {
				b.WriteString(" ")
			}
		}
	}
	if shouldIndent && d > 0 {
		b.WriteString(lib.Repeat("  ", int(d)))
	}
	b.WriteString("]")
	return b.String()
}

func (arr *Array) push(el Value) Number {
	arr.elements.Push(el)
	hashValue(arr)
	return Number(arr.elements.Len())
}

func NewArray() *Array {
	return &Array{
		hash:     0,
		elements: &lib.Array[Value]{},
	}
}

func visibleLen(str string) int {
	return len(lib.StripANSI(str))
}

// ###############################################
// ################### Pointer ###################
// ###############################################

func (m *Pointer) typeof() string {
	return m.value().typeof()
}

func (m *Pointer) toString() string {
	return m.value().toString()
}

func (m *Pointer) inspect(d uint, Inspector ValueInspector) string {
	return m.value().inspect(d, Inspector)
}

func (m *Pointer) value() Value {
	return m.memory[m.ptr]
}

// ###############################################
// ################### RAW ###################
// ###############################################

func (v *RAW) typeof() string {
	return RawType
}

func (v *RAW) toString() string {
	return lib.Sprint(v.value)
}

func (v *RAW) inspect(uint, ValueInspector) string {
	return v.toString()
}

func MK_RAW(v any) *RAW {
	return &RAW{v}
}

func toInt(v Value) Number {
	switch v := v.(type) {
	case *Pointer:
		return toInt(v.value())
	case Number:
		return v
	case *String:
		strlen := len(v.string)
		if strlen == 1 {
			return Number((v.string)[0])
		}
		return Number(strlen)
	case *Function, *Class, *Macro, *Instance, *Symbol, *ScopeObject:
		return 1
	case *Object:
		if v.hash > 0 {
			return 1
		}
	case *Array:
		return Number(v.elements.Len())
	case Boolean:
		if v {
			return 1
		}
	case *Undefined:
	case *RAW:
		switch v := v.value.(type) {
		case []any:
			return Number(len(v))
		case map[any]any:
			return Number(len(v))
		case chan any:
			return Number(len(v))
		case <-chan any:
			return Number(len(v))
		case string:
			return Number(len(v))
		}
	}
	return 0
}
