// Package floatprobe holds one float of each form forbidigo misses, so
// TestNoFloatGuardSeesEveryFloatForm can show the type-checked guard
// catches them. It lives under testdata: never built into the binary.
package floatprobe

import (
	"math"
	"strconv"
)

// Inferred is an untyped float literal inferred as float64.
func Inferred() int64 {
	x := 0.5
	return int64(x * 100)
}

// Parsed reads and rounds a float without naming its type.
func Parsed(s string) int64 {
	v, _ := strconv.ParseFloat(s, 64)
	return int64(math.Round(v))
}

// Constant is an untyped float constant expression.
func Constant() int64 {
	const third = 1.0 / 3
	return int64(third * 300)
}
