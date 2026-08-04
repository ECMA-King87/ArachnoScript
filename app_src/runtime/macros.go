package runtime

import (
	"aspire/are/main/lib"
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

const ArgMustBeFileErr = "Argument of type 'raw' is not expected type operating system file."
const ArgMustBeFileStatErr = "Argument of type 'raw' is not expected type file info."
const ArgMustBeByteSliceErr = "Argument of type 'raw' is not expected byte array."
const FloatArrIndexErr = "Cannot use floating point value to index array."
const ArgMustBeWASMRuntimeErr = "Argument of type 'raw' is not expected type WebAssembly runtime type."
const ArgMustBeWASMModErr = "Argument of type 'raw' is not expected type WebAssembly module type."
const ArgMustBeNumErr = "Argument of type number."
const ArgMustBeStrErr = "Argument of type string."
const ArgMustBeArrErr = "Argument of type array."
const ArgMustBeFunErr = "Argument of type function."
const ArgMustBeMacroErr = "Argument of type macro."

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
			ctx.errorWithSource(TypeError, 0, *loc, lib.Sprintf("Expected %s argument to be of type %s but got %s.", nth, expectedType, valueType))
		}
	}
	expectArgs := func(args Args, types []*String, ctx *EvalContext, loc *lib.Loc) {
		argsLen := len(args)
		expectedLen := len(types)
		if argsLen != expectedLen {
			ctx.errorWithSource(TypeError, 0, *loc, lib.Sprintf("Expected %d arguments but got %d.", expectedLen, argsLen))
		}
		for n, v := range args {
			t := types[n]
			if t != AnyType {
				errorOnMismatch(v.typeof(), t, n+1, ctx, loc)
			}
		}
	}
	var MACROS = []*Macro{
		MK_MACRO("assert", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{BooleanType}, ctx, loc)
			arg := args[0].(Boolean)
			if !arg {
				ctx.error(Error, "Assertion failed: ", lib.EOL, lib.DebugMsg(ctx.path, *loc))
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
			switch v := args[0].(type) {
			case Number:
				return v.toString()
			case *String:
				return v
			case *RAW:
				bytes, ok := v.value.([]byte)
				if ok {
					return NewString(string(bytes))
				}
				runes, ok := v.value.([]rune)
				if ok {
					return NewString(string(runes))
				}
			}
			return args[0].toString()
		}),
		MK_MACRO("number_cast", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType}, ctx, loc)
			switch v := args[0].(type) {
			case Number:
				return v
			case *String:
				return v
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
							errorOnMismatch(v.typeof(), NumberType, idx, ctx, loc)
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
				ctx.errorWithSource(TypeError, 0, *loc, "Argument is not of expected type (number | ...number | string)")
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
				ctx.errorWithSource(Error, 0, *loc, err.Error())
			}
			return MK_RAW(file)
		}),
		MK_MACRO("fs_close_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			// Returns an error if already called.
			err := file.Close()
			if err != nil {
				ctx.errorWithSource(TypeError, 0, *loc, err.Error())
			}
			return undefined
		}),
		MK_MACRO("fs_read_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			// arg, ok := args[0].(*RAW[lib.File])
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			n, err := file.Read(bytes)
			if err != nil {
				ctx.errorWithSource(TypeError, 0, *loc, err.Error())
			}
			return Number(n)
		}),
		MK_MACRO("fs_write_file", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			n, err := file.Write(bytes)
			if err != nil {
				ctx.errorWithSource(TypeError, 0, *loc, err.Error())
			}
			return Number(n)
		}),
		MK_MACRO("fs_stats", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			file, ok := args[0].(*RAW).value.(lib.File)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileErr)
			}
			stat, err := file.Stat()
			if err != nil {
				ctx.errorWithSource(TypeError, 0, *loc, err.Error())
			}
			return MK_RAW(stat)
		}),
		MK_MACRO("fs_stats_size", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Number(stat.Size())
		}),
		MK_MACRO("fs_stats_name", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return NewString(stat.Name())
		}),
		MK_MACRO("fs_stats_is_dir", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Boolean(stat.IsDir())
		}),
		MK_MACRO("fs_stats_mod_time", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType}, ctx, loc)
			stat, ok := args[0].(*RAW).value.(lib.FileInfo)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeFileStatErr)
			}
			return Number(stat.ModTime().Unix())
		}),
		MK_MACRO("json_stringify", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, BooleanType}, ctx, loc)
			v := args[0]
			indent := args[1].(Boolean)
			json, err := JSONStringify(v, bool(indent))
			if err != nil {
				ctx.errorWithSource(Error, 0, *loc, err.Error())
			}
			return json
		}),
		MK_MACRO("json_parse", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{StringType}, ctx, loc)
			v := args[0].(*String)
			return JSONParse(v.string)
		}),
		MK_MACRO("at", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{AnyType, NumberType}, ctx, loc)
			arg := float64(args[1].(Number))
			if lib.Modulo(arg, 1) != 0 {
				if lib.ENV.StrictMode {
					ctx.errorWithSource(TypeError, 0, *loc, FloatArrIndexErr)
				} else {
					arg = lib.TruncFloat(arg)
				}
			}
			idx := int(arg)
			switch v := args[0].(type) {
			case *String:
				l := len(v.string)
				if idx < 0 {
					idx += l
				}
				if idx < l {
					return NewString(string([]byte{v.string[idx]}))
				}
			case *RAW:
				val := lib.ValueOf(v.value)
				if lib.IsString(val) {
					l := val.Len()
					if idx < 0 {
						idx += l
					}
					if idx < l {
						return NewString(string([]byte{v.value.(string)[idx]}))
					}
				}
				if lib.IsArray(val) || lib.IsSlice(val) {
					l := val.Len()
					if idx < 0 {
						idx += l
					}
					if idx < l {
						return MK_RAW(val.Index(idx).Interface())
					}
				}
			case *Array:
				if el := v.elements.At(idx); el != nil {
					return el
				} else if lib.ENV.StrictMode {
					i := idx
					if idx < 0 {
						i = v.elements.Len() - idx
					}
					ctx.errorWithSource(TypeError, 0, *loc, lib.Sprintf("Element at index %d (%d) does not exist.", idx, i))
				}
			}
			return undefined
		}),
		MK_MACRO("new_wasm_runtime", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			return MK_RAW(lib.NewWASMRuntime())
		}),
		MK_MACRO("wasm_compile", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, RawType}, ctx, loc)
			runtime, ok := args[0].(*RAW).value.(*lib.WebAssembly)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
			}
			bytes, ok := args[1].(*RAW).value.([]byte)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			module, err := lib.WASMCompileModule(runtime, bytes)
			if err != nil {
				ctx.errorWithSource(Error, 0, *loc, err.Error())
			}
			return MK_RAW(module)
		}),
		MK_MACRO("wasm_instantiate", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			v := args[0].(*RAW).value
			if len(args) == 2 {
				expectArgs(args, []*String{RawType, RawType}, ctx, loc)
				bytes, ok := v.([]byte)
				if !ok {
					ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeByteSliceErr)
				}
				runtime, ok := args[1].(*RAW).value.(*lib.WebAssembly)
				if !ok {
					ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
				}
				inst, err := lib.WASMInstantiate(runtime, bytes)
				if err != nil {
					ctx.errorWithSource(Error, 0, *loc, err.Error())
				}
				return MK_RAW(inst)
			}
			expectArgs(args, []*String{RawType}, ctx, loc)
			module, ok := v.(lib.Module)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeWASMModErr)
			}
			inst, err := lib.WASMInstantiateModule(module)
			if err != nil {
				ctx.errorWithSource(Error, 0, *loc, err.Error())
			}
			return MK_RAW(inst)
		}),
		MK_MACRO("wasm_get_export", func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
			expectArgs(args, []*String{RawType, StringType}, ctx, loc)
			inst, ok := args[0].(*RAW).value.(lib.Instance)
			if !ok {
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
			}
			name := args[1].(*String).string
			fun, found := lib.GetWASMExport(inst, name)
			if !found {
				ctx.errorWithSource(Error, 0, *loc, "No export function with name: ", name)
			}
			return MK_MACRO("wasm_"+name, func(args Args, ctx *EvalContext, loc *lib.Loc, m *Macro) Value {
				wasmArgs := make([]uint64, len(args))
				for i := range len(args) {
					valType := args[i].typeof()
					if valType != NumberType {
						ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeWASMRuntimeErr)
						errorOnMismatch(valType, NumberType, i+1, ctx, loc)
					}
					wasmArgs[i] = uint64(args[i].(Number))
				}
				res, err := fun.Call(wasmArgs...)
				if err != nil {
					ctx.errorWithSource(Error, 0, *loc, err.Error())
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
				ctx.errorWithSource(TypeError, 0, *loc, ArgMustBeByteSliceErr)
			}
			wasmBytes, err := lib.WAT2WASM(watBytes)
			if err != nil {
				ctx.errorWithSource(Error, 0, *loc, err.Error())
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
				ctx.errorWithSource(Error, 0, *loc, "Second argument (Radix) must be an integer within the range 0 - 36.")
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
				ctx.errorWithSource(Error, 0, *loc, "Argument is negative, expected positive number.")
			}
			if lib.Modulo(n, 1) != 0 {
				ctx.errorWithSource(Error, 0, *loc, "Argument is not an integer.")
			}
			if n == 0 || ctx.callstack.Len() == 0 {
				return &emptyString
			}
			b := lib.NewStringBuilder()
			fn := ctx.callstack.At(0)
			b.WriteString(lib.DebugMsg(fn.callEnvPath, fn.callEnvLoc))
			ctx.callstack.Slice(0, int(n)).ForEach(func(i int, cf *CallFrame) {
				path := lib.PathFromKey(cf.callEnvPath)
				dirPath := lib.DirOf(path)
				dir, err := lib.LocalizePath(dirPath)
				if err != nil {
					dir = dirPath
				}
				name := lib.BaseName(path) + lib.PathExt(path)
				// name := lib.BaseName(path) + "." + lib.PathExt(path)
				b.WriteString(lib.Sprintf("\tat %s (%s/%s:%d:%d)%s",
					lib.Bold(lib.Italic(cf.fn.name)), lib.Dull(dir), lib.Cyan(name),
					cf.callEnvLoc.Line, cf.callEnvLoc.Col, lib.EOL))
			})
			return NewString(b.String())
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
