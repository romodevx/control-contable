package money

const (
	// ExactPerPeso is the number of Exact units in one peso (10^-6 peso scale).
	ExactPerPeso = 1_000_000
	// ExactPerCent is the number of Exact units in one centavo (10^-4 centavo).
	ExactPerCent = ExactPerPeso / 100
	// halfCent is the rounding threshold of Exact.Round.
	halfCent = ExactPerCent / 2
)

// Exact is an intermediate amount inside the tax engine, at 10^-6 peso scale
// (ADR-0002 §4). Engine arithmetic stays in Exact and rounds to Cents once,
// at the end, with Round.
type Exact int64

// Add returns e+o, or ErrOverflow.
func (e Exact) Add(o Exact) (Exact, error) {
	sum, err := checkedAdd(int64(e), int64(o))
	return Exact(sum), err
}

// Sub returns e-o, or ErrOverflow.
func (e Exact) Sub(o Exact) (Exact, error) {
	difference, err := checkedSub(int64(e), int64(o))
	return Exact(difference), err
}

// MulRate returns e × r rounded half away from zero at 10^-6 peso, through a
// 128-bit intermediate. Only a result outside int64 is an error (ErrOverflow).
func (e Exact) MulRate(r Rate) (Exact, error) {
	product, err := mulDiv(int64(e), int64(r), RateScale)
	return Exact(product), err
}

// MulRatio returns e × num / den rounded half away from zero, for fractions a
// Rate cannot hold exactly (the 2/3 IVA withholding). den must be positive
// (ErrInvalidDivisor); a result outside int64 is ErrOverflow.
func (e Exact) MulRatio(num, den int64) (Exact, error) {
	product, err := mulDiv(int64(e), num, den)
	return Exact(product), err
}

// Round returns e rounded to the centavo, half away from zero: the one
// rounding step of ADR-0002 §4. It cannot overflow.
func (e Exact) Round() Cents {
	whole, remainder := e/ExactPerCent, e%ExactPerCent
	switch {
	case remainder >= halfCent:
		whole++
	case remainder <= -halfCent:
		whole--
	}
	return Cents(whole)
}
