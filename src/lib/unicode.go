package lib

import "unicode/utf8"

func AppendRune(p []byte, r rune) []byte { return utf8.AppendRune(p, r) }

func DecodeRune(p []byte) (rune, int) {
	return utf8.DecodeRune(p)
}

func DecodeRuneInString(p string) (rune, int) {
	return utf8.DecodeRuneInString(p)
}
