package money

import "fmt"

const (
	// ratePlaces is the number of decimal places a Rate holds.
	ratePlaces = 6
	// RateScale is the Rate value of 1 (100%): rates are millionths.
	RateScale = 1_000_000
)

// Rate is a fixed-point fraction at 10^-6 precision: 0.16 is Rate(160_000).
// It covers 2-decimal percent rates and 4-decimal coeficientes de utilidad
// exactly (ADR-0001 O7). Rates travel over the wire as decimal strings.
type Rate int64

// ParseRate parses a decimal fraction string such as "0.16" or "0.1234"
// (at most six decimals). It does not accept percent notation.
func ParseRate(s string) (Rate, error) {
	v, err := parseFixed(s, ratePlaces)
	if err != nil {
		return 0, fmt.Errorf("money: parse rate %q: %w", s, err)
	}
	return Rate(v), nil
}

// String renders r in canonical form, without trailing zeros: "0.16", "1".
// ParseRate(r.String()) == r for every Rate.
func (r Rate) String() string {
	return formatFixed(int64(r), ratePlaces, true)
}
