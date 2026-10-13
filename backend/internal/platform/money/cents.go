package money

import "fmt"

// centsPlaces is the number of decimal places of a peso amount in centavos.
const centsPlaces = 2

// Cents is an amount of Mexican pesos in centavos. It is the storage and API
// representation (ADR-0001 O7): bigint in PostgreSQL, int64 in JSON.
type Cents int64

// ParseCents parses a decimal peso string such as "3439.46" or "-0.05" (at
// most two decimals, no sign other than a leading '-', no separators).
func ParseCents(s string) (Cents, error) {
	v, err := parseFixed(s, centsPlaces)
	if err != nil {
		return 0, fmt.Errorf("money: parse cents %q: %w", s, err)
	}
	return Cents(v), nil
}

// Add returns c+o, or ErrOverflow.
func (c Cents) Add(o Cents) (Cents, error) {
	sum, err := checkedAdd(int64(c), int64(o))
	return Cents(sum), err
}

// Sub returns c-o, or ErrOverflow.
func (c Cents) Sub(o Cents) (Cents, error) {
	difference, err := checkedSub(int64(c), int64(o))
	return Cents(difference), err
}

// Exact converts c to the engine's intermediate scale, or reports ErrOverflow
// beyond about ±9.2 × 10^12 pesos.
func (c Cents) Exact() (Exact, error) {
	e, err := mulDiv(int64(c), ExactPerCent, 1)
	return Exact(e), err
}

// MulRate returns c × r at Exact scale, rounded half away from zero at
// 10^-6 peso. The product is computed in 128 bits, so only a result outside
// Exact's range is an error (ErrOverflow).
func (c Cents) MulRate(r Rate) (Exact, error) {
	// c centavos × r millionths = c·r × 10^-8 peso; Exact is 10^-6 peso.
	const centRateToExact = RateScale / ExactPerCent
	e, err := mulDiv(int64(c), int64(r), centRateToExact)
	return Exact(e), err
}

// String renders c as a plain peso decimal with two places ("-1066.67").
// It is for logs and tests; presentation formatting belongs to the adapters.
func (c Cents) String() string {
	return formatFixed(int64(c), centsPlaces, false)
}
