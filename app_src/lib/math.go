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
// Special cases:
// Modulo(±Inf, y) = NaN
// Modulo(NaN, y) = NaN
// Modulo(x, 0) = NaN
// Modulo(x, ±Inf) = x
// Modulo(x, NaN) = NaN
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

// Returns a non-negative pseudo-random number
func IsInf(f float64, sign int) bool {
	return math.IsInf(f, sign)
}

// Returns a non-negative pseudo-random number
func AbsNum(f float64) float64 {
	return math.Abs(f)
}

const Euler = math.E
const Pi = math.Pi
const Phi = math.Phi
const Ln2 = math.Ln2
const Ln10 = math.Ln10
const Log10E = math.Log10E
const Log2E = math.Log2E
const SqrtE = math.SqrtE
const Sqrt2 = math.Sqrt2
const SqrtPi = math.SqrtPi
const SqrtPhi = math.SqrtPhi

// const SQRT1_2 = math.Sqrt2

const MAX_VALUE = math.MaxFloat64
const MIN_VALUE = math.SmallestNonzeroFloat64
const MAX_SAFE_INT = math.MaxInt
const MIN_SAFE_INT = math.MinInt

// EPSILON = 2.220446049250313e-16
const EPSILON float64 = 2e-52

// A loose relative tolerance of 1e-6
const LOOSE_REL_TOL float64 = 1e-6

// An absolute tolerance of 1e-9
const LOOSE_ABS_TOL float64 = 1e-9

// A default relative tolerance of 1e-9
const DEFAULT_REL_TOL float64 = 1e-9

// A default absolute tolerance of 1e-12
const DEFAULT_ABS_TOL float64 = 1e-12

// A strict relative tolerance of 1e-12
const STRICT_REL_TOL float64 = 1e-12

// A strict absolute tolerance of 1e-15
const STRICT_ABS_TOL float64 = 1e-15

func ApproxEqual(a, b, relTol, absTol float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}

	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return a == b
	}

	// const (
	// 	relTol = 1e-12
	// 	absTol = 1e-15
	// )

	diff := math.Abs(a - b)

	if diff <= absTol {
		return true
	}

	return diff <= relTol*math.Max(math.Abs(a), math.Abs(b))
}
