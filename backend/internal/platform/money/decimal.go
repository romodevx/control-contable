package money

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const decimalBase = 10

// parseFixed parses the strict grammar -?digits(.digits)? into an int64 scaled
// by 10^places. More than places fractional digits is ErrPrecision: values are
// stored verbatim, never rounded on the way in. O(len(s)).
func parseFixed(s string, places int) (int64, error) {
	unsigned, negative := strings.CutPrefix(s, "-")
	whole, fraction, hasPoint := strings.Cut(unsigned, ".")
	if !isDigits(whole) || (hasPoint && !isDigits(fraction)) {
		return 0, ErrSyntax
	}
	if len(fraction) > places {
		return 0, ErrPrecision
	}
	limit := uint64(math.MaxInt64)
	if negative {
		limit = negativeLimit
	}
	scaled := whole + fraction + strings.Repeat("0", places-len(fraction))
	var m uint64
	for i := range len(scaled) {
		digit := uint64(scaled[i] - '0')
		if m > (limit-digit)/decimalBase {
			return 0, ErrOverflow
		}
		m = m*decimalBase + digit
	}
	return withSign(m, negative)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// formatFixed renders v scaled by 10^places as a plain decimal string. With
// trimZeros the fraction loses its trailing zeros (and the point, if empty).
func formatFixed(v int64, places int, trimZeros bool) string {
	scale := uint64(1)
	for range places {
		scale *= decimalBase
	}
	m := magnitude(v)
	fraction := fmt.Sprintf("%0*d", places, m%scale)
	if trimZeros {
		fraction = strings.TrimRight(fraction, "0")
	}
	var b strings.Builder
	if v < 0 {
		b.WriteByte('-')
	}
	b.WriteString(strconv.FormatUint(m/scale, decimalBase))
	if fraction != "" {
		b.WriteByte('.')
		b.WriteString(fraction)
	}
	return b.String()
}
