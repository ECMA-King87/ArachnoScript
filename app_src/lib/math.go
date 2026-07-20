package lib

import (
	"math"
	"math/rand"
)

var NaN = math.NaN()
var Infinity = math.Inf(0)

// Evaluates b to the power p
func Pow(b, p float64) float64 {
	return math.Pow(b, p)
}

// Evaluates a % b or a modulus b
// which is the floating-point remainder of a/b
func Modulo(a, b float64) float64 {
	return math.Mod(a, b)
}

// Returns the integer value of x.
func TruncFloat(f float64) float64 {
	return math.Trunc(f)
}

// Returns a non-negative pseudo-random number
func Rand(n int) int {
	return rand.Intn(n)
}
