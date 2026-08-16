package runtime

import (
	"aspire/are/main/lib"
	"aspire/are/main/parser"
)

// type lib.File : [operating system file]
// type lib.FileInfo : [operating system file]
// type []any : [array]
// type []byte : [byte array]
// type []uint8 : [byte array]
// type float64 : [number | IEEE 754 64-bit floating point number]
// type float32 : [float 32-bit number]
// type lib.WebAssembly : [WebAssembly runtime]
// type lib.Module : [WebAssembly module]
// type lib.Instance : [WebAssembly instance]

const ArgMustBeFileErr = "Argument must be of type (raw [operating system file])."
const ArgMustBeFileStatErr = "Argument must be of type (raw [file info])."
const ArgMustBeByteSliceErr = "Argument must be of type (raw [byte array])."
const FloatArrIndexErr = "Cannot use floating point number to index array."
const ArgMustBeWASMRuntimeErr = "Argument must be of type (raw [WebAssembly runtime])."
const ArgMustBeWASMModErr = "Argument must be of type (raw [WebAssembly module])."
const ArgMustBeNumErr = "Argument must be of type number."
const ArgMustBeStrErr = "Argument must be of type string."
const ArgMustBeArrErr = "Argument must be of type array."
const ArgMustBeFunErr = "Argument must be of type function."
const ArgMustBeMacroErr = "Argument must be of type macro."
const IndexOutOfBoundsErr = "Index out of bounds."
const InvalidPropKey = "Invalid property key."
const CannotBindNonFunArg = "Cannot bind to non-function argument."
const FirstArgMustBeStrErr = "First argument must be of type string."
const FirstArgMustBeFuncErr = "First argument must be of type function."

func initMacros() {
	errorOnMismatch := func(valueType, expectedType *String, n int, ctx *EvalContext, loc *lib.Loc) {
		if valueType != expectedType {
			nth := ""
			switch n {
			case 1:
				nth = "1st"
			case 2:
				nth = "2nd"
			case 3:
				nth = "3rd"
			default:
				nth = lib.Sprint(n) + "th"
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, lib.Sprintf("Expected %s argument to be of type %s but got %s.", nth, expectedType.string, valueType.string))
		}
	}
	expectArgs := func(args Args, types []*String, ctx *EvalContext, loc *lib.Loc) {
		argsLen := len(args)
		expectedLen := len(types)
		if argsLen != expectedLen {
			ctx.raiseWithTrace(TypeError, 0, *loc, lib.Sprintf("Expected %d arguments but got %d.", expectedLen, argsLen))
		}
		for n, v := range args {
			t := types[n]
			if t != AnyType {
				errorOnMismatch(v.typeof(), t, n+1, ctx, loc)
			}
		}
	}
	getProperty := func(expectTypes []*String, obj *Object, name *String, ctx *EvalContext, path int, loc *lib.Loc, err ...string) (Value, bool) {
		_, desc, found := lookupProp(obj, name, nil)
		if !found {
			return nil, false
		}
		if !lib.InSlice(expectTypes, AnyType) && !lib.InSlice(expectTypes, desc.value.typeof()) {
			ctx.raiseWithTrace(TypeError, path, *loc, err...)
		}
		return desc.value, found
	}
	var MACROS = []*Macro{
		MK_MACRO("assert", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{BooleanType}, ctx, loc)
			arg := args[0].(Boolean)
			if !arg {
				ctx.raiseWithTrace(Error, 0, *loc, "Assertion failed: ")
			}
			return undefined
		}),
		MK_MACRO("length", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			return toInt(args[0])
		}),
		MK_MACRO("inspect", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			return NewString(ctx.Inspect(args[0]))
		}),
		MK_MACRO("to_string", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			return args[0].toString()
		}),
		// A more specialised version of to_string
		// it can handle conversion of byte/rune slices to strings.
		MK_MACRO("string_cast", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			return valToStringCast(args[0])
		}),
		MK_MACRO("number_cast", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			return valToNumberCast(args[0])
		}),
		MK_MACRO("new_byte_array", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			argsLen := len(args)
			if argsLen == 0 {
				return MK_RAW(make([]byte, 0))
			}
			switch v := args[0].(type) {
			case Number:
				if argsLen > 1 {
					arr := make([]byte, argsLen)
					for idx, v := range args {
						if v.typeof() != NumberType {
							errorOnMismatch(v.typeof(), NumberType, idx+1, ctx, loc)
						}
						arr[idx] = byte(v.(Number))
					}
					return MK_RAW(arr)
				} else {
					return MK_RAW(make([]byte, int(v)))
				}
			case *String:
				return MK_RAW([]byte(v.string))
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, "Argument is not of expected type (number | ...number | string)")
			}
			return undefined
		}),
		MK_MACRO("set_context", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{ScopeType}, ctx, loc)
			arg := args[0].(*ScopeObject)
			ctx.Scope = arg.Scope
			return undefined
		}),
		MK_MACRO("get_context", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			return &ScopeObject{ctx.Scope, null}
		}),
		MK_MACRO("random", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{NumberType}, ctx, loc)
			n := int(float64(args[0].(Number)))
			return Number(lib.Rand(n))
		}),
		// follow this pattern of naming: namespace_action_object
		MK_MACRO("os_open_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			path := args[0].(*String)
			file, err := lib.OpenFile(path.string)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return MK_RAW(file)
		}),
		MK_MACRO("fs_close_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			// Returns an error if already called.
			err := file.Close()
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return undefined
		}),
		MK_MACRO("fs_read_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			// arg, ok := args[0].(*RAW[lib.File])
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			n, err := file.Read(bytes)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return Number(n)
		}),
		MK_MACRO("fs_write_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			n, err := file.Write(bytes)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return Number(n)
		}),
		MK_MACRO("fs_stats", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			stat, err := file.Stat()
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return MK_RAW(stat)
		}),
		MK_MACRO("fs_stats_size", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Number(stat.Size())
		}),
		MK_MACRO("fs_stats_name", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return NewString(stat.Name())
		}),
		MK_MACRO("fs_stats_is_dir", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Boolean(stat.IsDir())
		}),
		MK_MACRO("fs_stats_mod_time", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Number(stat.ModTime().Unix())
		}),
		MK_MACRO("json_stringify", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, BooleanType}, ctx, loc)
			v := args[0]
			indent := args[1].(Boolean)
			json, err := JSONStringify(v, bool(indent))
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return json
		}),
		MK_MACRO("json_parse", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			text := args[0].(*String)
			v, err := JSONParse(text.string)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return v
		}),
		MK_MACRO("at", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, NumberType}, ctx, loc)
			arg := float64(args[1].(Number))
			if lib.Modulo(arg, 1) != 0 {
				if lib.ENV.StrictMode {
					ctx.raiseWithTrace(TypeError, 0, *loc, FloatArrIndexErr)
				} else {
					arg = lib.TruncFloat(arg)
				}
			}
			idx := int(arg)
			switch v := args[0].(type) {
			case *String:
				if idx < 0 || idx >= len(v.string) {
					if lib.ENV.StrictMode {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
					}
					return undefined
				}
				return NewString(string(v.string[idx]))
			case *RAW:
				val := lib.ValueOf(v.value)
				if lib.IsString(val) {
					if idx < 0 || idx >= val.Len() {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
						return undefined
					}
					return NewString(string(v.value.(string)[idx]))
				}
				if lib.IsArray(val) || lib.IsSlice(val) {
					if idx < 0 || idx >= val.Len() {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
						return undefined
					}
					return MK_RAW(val.Index(idx).Interface())
				}
			case *Array:
				if idx < 0 || idx >= v.elements.Len() {
					if lib.ENV.StrictMode {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
					}
					return undefined
				}
				return v.elements.At(idx)
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, "Argument is not indexable.")
			return undefined
		}),
		MK_MACRO("new_wasm_runtime", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			return MK_RAW(lib.NewWASMRuntime())
		}),
		MK_MACRO("wasm_compile", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			runtime, ok := args[0].(*RAW).value.(*lib.WebAssemblyRuntime)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			module, err := lib.WASMCompileModule(runtime, bytes)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return MK_RAW(module)
		}),
		MK_MACRO("wasm_instantiate", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			v := args[0].(*RAW).value
			if len(args) == 2 {
				expectArgs(args, []*String{RawType, RawType}, ctx, loc)
				bytes, ok := v.([]byte)
				if !ok {
					ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeByteSliceErr)
				}
				runtime, ok := args[1].(*RAW).value.(*lib.WebAssemblyRuntime)
				if !ok {
					ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
				}
				inst, err := lib.WASMInstantiate(runtime, bytes)
				if err != nil {
					ctx.raiseWithTrace(Error, 0, *loc, err.Error())
				}
				return MK_RAW(inst)
			}
			expectArgs(args, []*String{RawType}, ctx, loc)
			module, ok := v.(lib.Module)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeWASMModErr)
			}
			inst, err := lib.WASMInstantiateModule(module)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return MK_RAW(inst)
		}),
		MK_MACRO("wasm_get_export", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, StringType}, ctx, loc)
			inst, ok := args[0].(*RAW).value.(lib.Instance)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
			}
			name := args[1].(*String).string
			fun, found := lib.GetWASMExport(inst, name)
			if !found {
				ctx.raiseWithTrace(Error, 0, *loc, "No export function with name: ", name)
			}
			return MK_MACRO("wasm_"+name, func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
				wasmArgs := make([]uint64, len(args))
				for i := range args {
					valType := args[i].typeof()
					if valType != NumberType {
						ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
						errorOnMismatch(valType, NumberType, i+1, ctx, loc)
					}
					wasmArgs[i] = uint64(args[i].(Number))
				}
				res, err := fun.Call(wasmArgs...)
				if err != nil {
					ctx.raiseWithTrace(Error, 0, *loc, err.Error())
				}
				if len(res) == 1 {
					return Number(res[0])
				}
				arr := NewArray()
				for _, el := range res {
					arr.push(Number(el))
				}
				return arr
			})
		}),
		MK_MACRO("wat_to_wasm", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			watBytes, ok := args[0].(*RAW).value.([]byte)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			wasmBytes, err := lib.WAT2WASM(watBytes)
			if err != nil {
				ctx.raiseWithTrace(Error, 0, *loc, err.Error())
			}
			return MK_RAW(wasmBytes)
		}),
		MK_MACRO("repl", func(_ Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			REPL()
			return undefined
		}),
		MK_MACRO("parse_int", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType, NumberType}, ctx, loc)
			str := args[0].(*String)
			radix := args[1].(Number)
			if lib.Modulo(float64(radix), 0) != 0 || radix < 0 || radix > 36 {
				ctx.raiseWithTrace(Error, 0, *loc, "Second argument (Radix) must be an integer within the range 0 - 36.")
			}
			i, err := lib.ParseInt(str.string, int(radix), 64)
			if err != nil {
				return NaN
			}
			return Number(float64(i))
		}),
		MK_MACRO("parse_float", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			str := args[0].(*String)
			i, err := lib.ParseFloat(str.string, 64)
			if err != nil {
				return NaN
			}
			return Number(i)
		}),
		MK_MACRO("is_finite", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{NumberType}, ctx, loc)
			return Boolean(args[0].(Number).IsFinite())
		}),
		MK_MACRO("abs", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{NumberType}, ctx, loc)
			return Number(lib.AbsNum(float64(args[0].(Number))))
		}),
		MK_MACRO("trunc_num", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{NumberType}, ctx, loc)
			return Number(lib.TruncFloat(float64(args[0].(Number))))
		}),
		MK_MACRO("stack_trace", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{NumberType}, ctx, loc)
			n := float64(args[0].(Number))
			if n < 0 {
				ctx.raiseWithTrace(Error, 0, *loc, "Argument is negative, expected positive number.")
			}
			if lib.Modulo(n, 1) != 0 {
				ctx.raiseWithTrace(Error, 0, *loc, "Argument is not an integer.")
			}
			ctx.callstack.Pop()
			// No use of frame so we can leave it empty
			defer ctx.callstack.Push(&CallFrame{})
			return newStackTrace(int(n), ctx)
		}),
		MK_MACRO("byte", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			switch v := args[0].(type) {
			case Number:
				return &RAW{byte(lib.TruncFloat(float64(v)))}
			case *String:
				return &RAW{byte(v.string[0])}
			case *RAW:
				switch v := v.value.(type) {
				case byte:
					return &RAW{v}
				case rune:
					return &RAW{byte(v)}
				case string:
					return &RAW{byte(v[0])}
				}
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, "Cannot convert argument to byte. Expected type (number | string) but got ", args[0].typeof().string)
			return undefined
		}),
		MK_MACRO("define_property", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, AnyType, ObjectType}, ctx, loc)
			var object *Object
			switch o := args[0].(type) {
			case *Instance:
				object = o.Object
			case *Class:
				object = o.Object
			case *Function:
				object = o.Object
			case *Object:
				object = o
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, "Cannot define properties on type ", lib.Green(args[0].typeof().string), ".")
			}
			var key *String
			switch o := args[1].(type) {
			case *String, *Number, *Symbol:
				key = o.toString()
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, InvalidPropKey)
			}
			attr := args[2].(*Object)
			isConfigurable := true
			isEnumerable := true
			isWritable := true
			var getter, setter *Function
			var value Value = undefined
			configurable, configFound := getProperty([]*String{BooleanType}, attr, NewString("configurable"), ctx, 0, loc)
			if configFound {
				isConfigurable = bool(configurable.(Boolean))
			}
			writable, writeFound := getProperty([]*String{BooleanType}, attr, NewString("writable"), ctx, 0, loc, "Writable descriptor must be of type boolean.")
			if writeFound {
				isWritable = bool(writable.(Boolean))
			}
			enumerable, enumFound := getProperty([]*String{BooleanType}, attr, NewString("enumerable"), ctx, 0, loc, "Enumerable descriptor must be of type boolean.")
			if enumFound {
				isEnumerable = bool(enumerable.(Boolean))
			}
			getterM, getFound := getProperty([]*String{FunctionType}, attr, NewString("getter"), ctx, 0, loc, "Getter descriptor must be of type function.")
			if getFound {
				getter = getterM.(*Function)
			}
			setterM, setFound := getProperty([]*String{FunctionType}, attr, NewString("setter"), ctx, 0, loc, "Setter descriptor must be of type function.")
			if setFound {
				setter = setterM.(*Function)
			}
			valueP, valFound := getProperty([]*String{AnyType}, attr, NewString("value"), ctx, 0, loc)
			if valFound {
				value = valueP
			}
			if setFound || getFound {
				if valFound || writeFound {
					ctx.raiseWithTrace(TypeError, 0, *loc, "Invalid property descriptor. Cannot both specify accessors and a value or writable attribute.")
				}
			}
			propDesc, f := object.own.Get(key)
			if f && !propDesc.configurable {
				ctx.raiseWithTrace(TypeError, 0, *loc, "Cannot redefine property: ", key.string)
			}
			object.own.Set(key, PropertyDescriptor{
				value:        value,
				configurable: isConfigurable,
				enumerable:   isEnumerable,
				writable:     isWritable,
				public:       true,
				static:       false,
				getter:       getter,
				setter:       setter,
			})
			return args[0]
		}),
		MK_MACRO("object_cast", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			arg := args[0]
			return valToObjectCast(arg)
		}),
		MK_MACRO("get_own_prop_descriptor", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, AnyType}, ctx, loc)
			obj := args[0]
			p := args[1]
			var object *Object
			switch o := obj.(type) {
			case *Instance:
				object = o.Object
			case *Class:
				object = o.Object
			case *Function:
				object = o.Object
			case *Object:
				object = o
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, "Cannot convert type ", lib.Green(obj.typeof().string), " to object.")
			}
			var key *String
			switch k := p.(type) {
			case *String, *Number, *Symbol:
				key = k.toString()
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, InvalidPropKey)
			}
			_, desc, found := lookupProp(object, key, nil)
			if !found {
				return undefined
			}
			rtv := NewObject()
			rtv.set(NewString("configurable"), Boolean(desc.configurable))
			rtv.set(NewString("writable"), Boolean(desc.writable))
			rtv.set(NewString("enumerable"), Boolean(desc.enumerable))
			var getter, setter, value Value = undefined, undefined, undefined
			if desc.value != nil {
				value = desc.value
			}
			if desc.getter != nil {
				getter = desc.getter
			}
			if desc.setter != nil {
				setter = desc.setter
			}
			rtv.set(NewString("value"), value)
			rtv.set(NewString("getter"), getter)
			rtv.set(NewString("setter"), setter)
			return rtv
		}),
		MK_MACRO("new_symbol", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			description := args[0].(*String)
			return &Symbol{description.string}
		}),
		MK_MACRO("symbol_for", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			key := args[0].(*String)
			sym, found := globalSymbolReg.Get(key)
			if !found {
				newSym := &Symbol{key.string}
				globalSymbolReg.Set(key, newSym)
				return newSym
			}
			return sym
		}),
		MK_MACRO("key_for_symbol", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{SymbolType}, ctx, loc)
			sym := args[0].(*Symbol)
			found := false
			var key *String
			globalSymbolReg.Until(func(k *String, value *Symbol, _ int) bool {
				found = sym == value
				if found {
					key = k
				}
				return found
			})
			if found {
				return key
			}
			return undefined
		}),
		MK_MACRO("bind_this", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			if len(args) < 2 {
				ctx.raiseWithTrace(TypeError, 0, *loc, "At least two arguments are required.")
			}
			fn, ok := args[0].(*Function)
			if !ok {
				ctx.raiseWithTrace(TypeError, 0, *loc, FirstArgMustBeFuncErr)
			}
			arguments := args[2:]
			this := args[1]
			newFunc := NewFunction(fn.name, fn.body, fn.Loc, parser.FnDecl{
				Async:     fn.async,
				Params:    fn.params,
				Name:      parser.StringNode(fn.name, fn.Loc),
				Arrow:     fn.arrow,
				Anonymous: false,
			}, NewContext(fn.declScope), this)
			return ctx.execFrame(newFunc, arguments, *loc)
		}),
		MK_MACRO("entries", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{ObjectType}, ctx, loc)
			obj := args[0].(*Object)
			entries := NewArray()
			obj.own.ForEach(func(key *String, desc PropertyDescriptor, _ bool) {
				entry := NewArray()
				if desc.getter != nil {
					entry.push(key, ctx.execFrame(desc.getter, Args{}, desc.getter.Loc))
				} else {
					entry.push(key, desc.value)
				}
				entries.push(entry)
			})
			return entries
		}),
		MK_MACRO("from_entries", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{ArrayType}, ctx, loc)
			arr := args[0].(*Array)
			obj := NewObject()
			if arr.elements.Len() == 0 {
				return obj
			}
			arr.elements.ForEach(func(i int, v Value) {
				if arr, ok := v.(*Array); ok {
					key := toValidPropKey(arr.elements.At(0))
					val := arr.elements.At(1)
					obj.set(key, val)
				} else {
					ctx.raiseWithTrace(TypeError, 0, *loc, "Entries must be of type [(string | number | symbol), any].")
				}
			})
			return obj
		}),
		MK_MACRO("sprintf", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			if len(args) == 0 {
				ctx.raiseWithTrace(Error, 0, *loc, "Expected at least one argument of type string.")
			}
			arr := make([]any, len(args)-1)
			fmt, ok := args[0].(*String)
			if !ok {
				ctx.raiseWithTrace(Error, 0, *loc, FirstArgMustBeStrErr)
			}
			for i, s := range args[1:] {
				arr[i], _ = jsonValueFromAS(s)
			}
			return NewString(lib.Sprintf(fmt.string, arr...))
		}),
		MK_MACRO("slice", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, NumberType, NumberType}, ctx, loc)
			startIdx := float64(args[1].(Number))
			endIdx := float64(args[2].(Number))
			if lib.Modulo(startIdx, 1) != 0 {
				if lib.ENV.StrictMode {
					ctx.raiseWithTrace(TypeError, 0, *loc, FloatArrIndexErr)
				} else {
					startIdx = lib.TruncFloat(startIdx)
				}
			}
			if lib.Modulo(endIdx, 1) != 0 {
				if lib.ENV.StrictMode {
					ctx.raiseWithTrace(TypeError, 0, *loc, FloatArrIndexErr)
				} else {
					endIdx = lib.TruncFloat(endIdx)
				}
			}
			idx1 := int(startIdx)
			idx2 := int(endIdx)
			switch v := args[0].(type) {
			case *String:
				if idx1 < 0 || idx1 >= len(v.string) || idx2 < 0 || idx2 > len(v.string) || idx1 > idx2 {
					if lib.ENV.StrictMode {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
					}
					return undefined
				}
				return NewString(v.string[idx1:idx2])
			case *RAW:
				val := lib.ValueOf(v.value)
				if lib.IsString(val) {
					if idx1 < 0 || idx1 >= val.Len() || idx2 < 0 || idx2 > val.Len() || idx1 > idx2 {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
						return undefined
					}
					return NewString(v.value.(string)[idx1:idx2])
				}
				if lib.IsArray(val) || lib.IsSlice(val) {
					if idx1 < 0 || idx1 > val.Len() || idx2 < 0 || idx2 > val.Len() || idx1 > idx2 {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
						return undefined
					}
					return MK_RAW(val.Slice(idx1, idx2).Interface())
				}
			case *Array:
				if idx1 < 0 || idx1 >= v.elements.Len() || idx2 < 0 || idx2 > v.elements.Len() || idx1 > idx2 {
					if lib.ENV.StrictMode {
						ctx.raiseWithTrace(RangeError, 0, *loc, IndexOutOfBoundsErr)
					}
					return undefined
				}
				slice := v.elements.Slice(idx1, idx2)
				arr := NewArray()
				arr.elements.Copy(slice)
				return arr
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, "Argument is not indexable.")
			return undefined
		}),
		MK_MACRO("append", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			if len(args) < 2 {
				ctx.raiseWithTrace(Error, 0, *loc, "Expected at least two arguments of type (any [indexable], any).")
			}
			switch arg1 := args[0].(type) {
			case *String:
				b := lib.NewStringBuilder()
				// Concatenate strings.
				// Use string casting.
				for _, v := range args {
					b.WriteString(valToStringCast(v).string)
				}
				return NewString(b.String())
			case *RAW:
				val := lib.ValueOf(arg1.value)
				if lib.IsString(val) {
					b := lib.NewStringBuilder()
					for _, v := range args {
						b.WriteString(valToStringCast(v).string)
					}
					return &RAW{b.String()}
				}
				if lib.IsArray(val) || lib.IsSlice(val) {
					arr := arg1.value
					for i, v := range args[1:] {
						errorOnMismatch(v.typeof(), RawType, i+2, ctx, loc)
						lib.AppendValue(val, lib.ValueOf(v.(*RAW).value))
					}
					return MK_RAW(arr)
				}
			case *Array:
				arg1.push(args[1:]...)
				return arg1
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, "Argument is not indexable.")
			return undefined
		}),
		MK_MACRO("append_array", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			val := lib.ValueOf(args[0].(*RAW).value)
			if lib.IsArray(val) || lib.IsSlice(val) {
				arg2 := lib.ValueOf(args[1].(*RAW).value)
				if lib.IsArray(arg2) || lib.IsSlice(arg2) {
					lib.AppendSlice(val, arg2)
					return MK_RAW(val.Interface())
				}
			}
			ctx.raiseWithTrace(TypeError, 0, *loc, "Expected both arguments to be of type (raw [array])")
			return undefined
		}),
		MK_MACRO("string_includes", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType, StringType}, ctx, loc)
			str := args[0].(*String)
			substr := args[1].(*String)
			return Boolean(lib.StrContains(str.string, substr.string))
		}),
		MK_MACRO("raw_cast", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			v := args[0]
			return valToRawCast(v)
		}),
		MK_MACRO("default", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			v := args[0]
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
			default:
				ctx.raiseWithTrace(TypeError, 0, *loc, "Cannot convert type ", lib.Green(o.typeof().string), " to object.")
			}
			return object.defaultVal
		}),
		MK_MACRO("to_uppercase", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			return NewString(lib.ToUpperCase(args[0].(*String).string))
		}),
		MK_MACRO("to_lowercase", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			return NewString(lib.ToLowerCase(args[0].(*String).string))
		}),
		// MK_MACRO("name", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
		// 	expectArgs(args, []*String{NumberType}, ctx, loc)
		// 	return undefined
		// }),
	}
	for _, m := range MACROS {
		globalThis.init(NewString("#_"+m.name), m, ConstDecl, lib.DumbyLoc)
	}
}

func valToRawCast(arg Value) *RAW {
	switch v := arg.(type) {
	case Number, Boolean, *Macro, *Symbol, *Undefined:
		return &RAW{v}
	case *String:
		return &RAW{v.string}
	// case *RAW:
	case *Array:
		arr := make([]any, v.elements.Len())
		v.elements.ForEach(func(i int, v Value) {
			arr[i] = valToRawCast(v)
		})
		return &RAW{arr}
	case *Class, *Function:
		return &RAW{undefined}
	case *Instance:
		return valToRawCast(v.Object)
	case *Object:
		_map := map[string]any{}
		v.own.ForEach(func(key *String, desc PropertyDescriptor, _ bool) {
			// Can't use getters for now.
			_map[key.string] = desc.value
		})
		return &RAW{_map}
	case *ScopeObject:
		return valToRawCast(scopeToObject(v))
	default:
		return arg.(*RAW)
	}
}

func valToNumberCast(arg Value) Number {
	switch v := arg.(type) {
	case Number:
		return v
	case *String:
		return Number(lib.ParseNumber(v.string))
	case *RAW:
		_byte, ok := v.value.(byte)
		if ok {
			return Number(_byte)
		}
		_rune, ok := v.value.(rune)
		if ok {
			return Number(_rune)
		}
	}
	return NaN
}

func valToStringCast(arg Value) *String {
	switch v := arg.(type) {
	case Number:
		return v.toString()
	case *String:
		return v
	case *RAW:
		bytes, ok := v.value.([]byte)
		if ok {
			return NewString(string(bytes))
		}
		str, ok := v.value.(string)
		if ok {
			return NewString(str)
		}
		runes, ok := v.value.([]rune)
		if ok {
			return NewString(string(runes))
		}
		r, ok := v.value.(rune)
		if ok {
			return NewString(string(r))
		}
		b, ok := v.value.(byte)
		if ok {
			return NewString(string(b))
		}
	}
	return arg.toString()
}

func valToObjectCast(arg Value) *Object {
	switch v := arg.(type) {
	case *Object:
		return v
	case *Array:
		obj := NewObject()
		v.elements.ForEach(func(i int, v Value) {
			obj.set(NewString(lib.Sprint(i)), v)
		})
		return obj
	case *Class:
		return v.Object
	case *Function:
		return v.Object
	case *Instance:
		return v.Object
	// case *Macro: Can't convert macro. Especially for future reasons.
	// Macros aren't meant to exist at runtime but v0.2.5 is interpreted.

	// TODO: Assertain if this is a good design decision.
	case *String, *Undefined, Boolean, Number:
		obj := NewObject()
		obj.defaultVal = v
		return obj
	case *RAW:
		obj := NewObject()
		switch g := v.value.(type) {
		case map[string]any:
			for k, v := range g {
				obj.set(NewString(k), valToObjectCast(&RAW{v}))
			}
		case map[string][]any:
			for k, v := range g {
				obj.set(NewString(k), valToObjectCast(&RAW{v}))
			}
		case map[int]any:
			for k, v := range g {
				obj.set(NewString(lib.Sprint(k)), valToObjectCast(&RAW{v}))
			}
		case map[int][]any:
			for k, v := range g {
				obj.set(NewString(lib.Sprint(k)), valToObjectCast(&RAW{v}))
			}
		case []any:
			for k, v := range g {
				obj.set(NewString(lib.Sprint(k)), valToObjectCast(&RAW{v}))
			}
		default:
			obj.defaultVal = v
		}
		return obj
	case *ScopeObject:
		return scopeToObject(v)
	case *Symbol:
		obj := NewObject()
		obj.set(NewString("description"), NewString(v.description))
		return obj
	}
	return null
}
