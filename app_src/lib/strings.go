package lib

import (
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
)

type uint128 [2]uint64

func Checksum128(bytes []byte) string {
	h := fnv.New128a()
	h.Write(bytes)
	return Sprintf("%x", h.Sum(nil))
}

func FNV1a64(s string) uint64 {
	var h uint64 = 0xcbf29ce484222325
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 0x100000001b3
	}
	return h
}

func NewStringBuilder() strings.Builder {
	return strings.Builder{}
}

func ToUpperCase(str string) string {
	return strings.ToUpper(str)
}

func ToLowerCase(str string) string {
	return strings.ToLower(str)
}

func IsAlpha(b rune) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || IsDigit(b) || b == '_' || b == '#'
}

func IsDigit(b rune) bool {
	return b >= '0' && b <= '9'
}

func IsHex(b rune) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}

func IsOctal(b rune) bool {
	return b >= '0' && b <= '7'
}

func IsBinary(b rune) bool {
	return b == '0' || b == '1'
}

func ParseNumber(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		i, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return NaN
		}
		return float64(i)
	}
	return f
}

func Quote(str string) string {
	return strconv.Quote(str)
}

// tries to parse the string, if it fails
// Unquote returns the string as is
func Unquote(str string) string {
	if ps, err := strconv.Unquote(str); err == nil {
		return ps
	}
	return str
}

func PadString(str string, pad string, length int, pre bool) string {
	str_len := len(str)
	if str_len >= length {
		return str
	}
	padded_str := ""
	padding := strings.Repeat(pad, length-str_len)
	if pre {
		padded_str = padding + str
	} else {
		padded_str = str + padding
	}
	return padded_str
}

func FormatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

func TrimStr(str string) string {
	return strings.Trim(str, " \t\r\n")
}

func SplitStr(str, sep string) []string {
	return strings.Split(str, sep)
}

func Repeat(str string, count int) string {
	return strings.Repeat(str, count)
}

func Match(pattern string, str string) bool {
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

// ANSI regular expression pattern targeting control sequences
const ansiPattern = "[\u001B\u009B]\\[[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-mA-ORZcf-nqry=><]"

var re = regexp.MustCompile(ansiPattern)

func StripANSI(str string) string {
	return re.ReplaceAllString(str, "")
}

func ReplaceStr(str, old, new string, n int) string {
	return strings.Replace(str, old, new, n)
}
