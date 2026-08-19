package lib

import (
	"os"
	"reflect"
	"slices"
	"sync"
)

func ExitWith1() {
	os.Exit(1)
}

type Enum uint

type Loc struct {
	// the Line and column of the token in the source
	Line, Col,
	// the position of this token
	Start,
	// the position of the token following this one (it should always be greater then start)
	End uintptr
}

var DumbyLoc = Loc{1, 1, 0, 0}

func SourceAtPosition(path string, loc Loc) string {
	return Sprintf("at (\x1b[34m%s\x1b[33m:%d:%d\x1b[0m)%s", path, loc.Line, loc.Col, EOL)
}

// path: is the cached file key for source lookup
//
// loc: the source location to highlight
func SourceWithinRange(
	path int,
	loc Loc,
) string {
	var lines []string = SplitStr(string(ReadCachedFile(path)), "\n")
	count := min(loc.End-loc.Start, 500)
	if loc.Line == 0 || loc.Line > uintptr(len(lines)) {
		return Sprintf("at (\x1b[34m%s\x1b[0m)%s", PathFromKey(path), EOL)
	}

	index := int(loc.Line - 1)
	line := lines[index]
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	leading_source := Sprintf("\x1b[33;1m%d |\x1b[0m  ", loc.Line)
	distance := max(int(loc.Col-1)+VisibleLen(leading_source), 0)

	line_source := NewStringBuilder()
	line_source.WriteString(Repeat(" ", distance))
	line_source.WriteString("\x1b[31m")
	line_source.WriteString(Repeat("^", int(count)))
	line_source.WriteString("\x1b[0m")
	line_source.WriteString(EOL)
	line_source.WriteString(Repeat(" ", distance))
	line_source.WriteString("\x1b[34m")
	line_source.WriteString(Sprintf("%d", loc.Col))
	line_source.WriteString("\x1b[0m")
	line_source.WriteString(EOL)

	return Sprintf(
		"%s%s%s%s",
		leading_source,
		line,
		EOL,
		line_source.String(),
	)
}

func DebugMsg(path int, loc Loc) string {
	return SourceWithinRange(path, loc) + Sprintf("at (%s)%s", Blue(PathFromKey(path)), EOL)
}

func VisibleLen(str string) int {
	return len(StripANSI(str))
}

const LIBPATH_ENV_KEY = "ARE_LIB_PATH"
const ARE_ROOT = "ARE_ROOT"
const ARE_CACHE_DIR = "ARE_CACHE"

func GetEnvVar(key string) (value string, exists bool) {
	return os.LookupEnv(key)
}

func SetEnvVar(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		Panic(err)
	}
}

func InSlice[T comparable](slice []T, el T) bool {
	return slices.Contains(slice, el)
}

func Grow[T any](slice []T, n uint) []T {
	return slices.Grow(slice, int(n))
}

func SameObject(a, b any) bool {
	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(b)
	if v1.Kind() == reflect.Pointer && v2.Kind() == reflect.Pointer {
		return v1.Pointer() == v2.Pointer()
	}
	if v1.Kind() == reflect.String && v2.Kind() == reflect.String {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}

func DeepEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

func IsPointer(val reflect.Value) bool {
	return val.Kind() == reflect.Pointer
}

func IsArray(val reflect.Value) bool {
	return val.Kind() == reflect.Array
}

func IsSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice
}

func IsStruct(val reflect.Value) bool {
	return val.Kind() == reflect.Struct
}

func IsString(val reflect.Value) bool {
	return val.Kind() == reflect.String
}

func IsMap(val reflect.Value) bool {
	return val.Kind() == reflect.Map
}

func IsChan(val reflect.Value) bool {
	return val.Kind() == reflect.Chan
}

func ValueOf(v any) reflect.Value {
	val := reflect.ValueOf(v)
	return val
}

// Append appends the values x to a slice s and returns the resulting slice. As in Go, each x's value must be assignable to the slice's element type.
func AppendValue(s reflect.Value, x ...reflect.Value) reflect.Value { return reflect.Append(s, x...) }

// AppendSlice appends a slice t to a slice s and returns the resulting slice. The slices s and t must have the same element type.
func AppendSlice(s reflect.Value, t reflect.Value) reflect.Value { return reflect.AppendSlice(s, t) }

// func ReflectSet(x, v any) {
// 	v1 := reflect.ValueOf(x)
// 	v2 := reflect.ValueOf(v)
// 	switch x.(type) {
// 	case bool, string, int, float32, float64,
// 		complex128, int64, int16, int8, int32,
// 		uint64, uint16, uint8, uint32:
// 		x = v
// 	default:
// 		v1.Set(v2)
// 	}
// }

func PointerOf(v any) uintptr {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		return val.Pointer()
	}
	return 0
}

type InternTable[K comparable, V any] struct {
	Mu    sync.RWMutex
	Table map[K]V
}
