package money

import (
	"errors"
	"math"
	"math/bits"
)

// Sentinel errors. Callers match them with errors.Is.
var (
	// ErrOverflow reports a result outside the int64 range of its type.
	ErrOverflow = errors.New("money: overflow")
	// ErrInvalidDivisor reports a ratio whose denominator is not positive.
	ErrInvalidDivisor = errors.New("money: divisor must be positive")
	// ErrSyntax reports a decimal string that is not -?digits(.digits)?.
	ErrSyntax = errors.New("money: invalid decimal syntax")
	// ErrPrecision reports more fractional digits than the type can hold.
	// Parsing never rounds silently.
	ErrPrecision = errors.New("money: too many decimal places")
)

// negativeLimit is |math.MinInt64|, the largest magnitude a negative int64 holds.
const negativeLimit = uint64(math.MaxInt64) + 1

// mulDiv returns a*b/d rounded half away from zero. The product is held in
// 128 bits (math/bits), so it is exact for every int64 a and b; only a
// quotient outside int64 is an error. d must be positive. O(1), no allocation.
func mulDiv(a, b, d int64) (int64, error) {
	if d <= 0 {
		return 0, ErrInvalidDivisor
	}
	negative := (a < 0) != (b < 0)
	hi, lo := bits.Mul64(magnitude(a), magnitude(b))
	divisor := uint64(d)
	if hi >= divisor { // the quotient needs more than 64 bits; bits.Div64 would panic
		return 0, ErrOverflow
	}
	quotient, remainder := bits.Div64(hi, lo, divisor)
	if remainder >= divisor-remainder { // 2*remainder >= divisor, without overflowing
		// hi < divisor only bounds quotient by 2^64-1, so the increment could wrap.
		// A rounded magnitude above 2^63 fits neither sign: reject it before adding.
		if quotient >= negativeLimit {
			return 0, ErrOverflow
		}
		quotient++
	}
	return withSign(quotient, negative)
}

// magnitude returns |v| as uint64; it is exact for math.MinInt64 too.
func magnitude(v int64) uint64 {
	u := uint64(v) //nolint:gosec // two's-complement reinterpretation, negated below when needed
	if v < 0 {
		return -u
	}
	return u
}

// withSign converts a magnitude back to int64, or reports ErrOverflow.
func withSign(m uint64, negative bool) (int64, error) {
	if negative {
		if m > negativeLimit {
			return 0, ErrOverflow
		}
		return -int64(m), nil //nolint:gosec // m <= 2^63; for m == 2^63 the wrap yields exactly MinInt64
	}
	if m > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return int64(m), nil
}

// checkedAdd returns a+b or ErrOverflow.
func checkedAdd(a, b int64) (int64, error) {
	sum := a + b
	if (b > 0 && sum < a) || (b < 0 && sum > a) {
		return 0, ErrOverflow
	}
	return sum, nil
}

// checkedSub returns a-b or ErrOverflow.
func checkedSub(a, b int64) (int64, error) {
	difference := a - b
	if (b > 0 && difference > a) || (b < 0 && difference < a) {
		return 0, ErrOverflow
	}
	return difference, nil
}
