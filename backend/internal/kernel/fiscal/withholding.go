package fiscal

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Share is a legal fraction Num/Den of a base, for the IVA withholdings a
// Rate cannot hold exactly (2/3 = 10.6667 % of the price, ADR-0002 §4). The
// zero Share means "no suggestion".
type Share struct {
	Num, Den int64
}

// The IVA withholding shares of the law. They are fractions of the IVA, not
// yearly rates: RLIVA 3 fr. I (2/3) and LIVA 18-J fr. II a) (50 %, 100 %
// without RFC).
var (
	ShareTwoThirds = Share{Num: 2, Den: 3}
	ShareHalf      = Share{Num: 1, Den: 2}
	ShareWhole     = Share{Num: 1, Den: 1}
)

// IsZero reports the zero Share.
func (s Share) IsZero() bool { return s.Den == 0 }

// Suggest returns base × Num/Den rounded once to the centavo, half away from
// zero, or no suggestion for the zero Share. Only a base beyond
// money.Exact's range fails, a parameter fault.
func (s Share) Suggest(base money.Cents) (Suggestion, error) {
	if s.IsZero() {
		return Suggestion{}, nil
	}
	e, err := base.Exact()
	if err == nil {
		e, err = e.MulRatio(s.Num, s.Den)
	}
	if err != nil {
		return Suggestion{}, fmt.Errorf("apply share %d/%d: %w: %w", s.Num, s.Den, ErrInvalidParameter, err)
	}
	return Suggestion{Cents: e.Round(), Known: true}, nil
}

// SuggestAtRate returns base × the rate key holds on day on. No key, or no
// rate in force that day, is no suggestion, never an error.
func SuggestAtRate(key RateKey, base money.Cents, on Date, catalog RateCatalog) (Suggestion, error) {
	if key == "" {
		return Suggestion{}, nil
	}
	r, ok := catalog.Rate(key, on)
	if !ok {
		return Suggestion{}, nil
	}
	c, err := ApplyRate(base, r)
	if err != nil {
		return Suggestion{}, err
	}
	return Suggestion{Cents: c, Known: true}, nil
}

// Withholdings are the ISR and IVA withheld on a record, by whichever side
// of it the taxpayer is.
type Withholdings struct {
	ISR money.Cents
	IVA money.Cents
}

// Suggestion is a suggested withholding; Known is false when there is none.
type Suggestion struct {
	Cents money.Cents
	Known bool
}

// Differs reports a known suggestion farther than IVATolerance from the
// withheld amount: the CFDI wins, but the user is told.
func (s Suggestion) Differs(withheld money.Cents) bool {
	return s.Known && !WithinTolerance(withheld, s.Cents)
}

// Suggested are the suggested withholdings of a record.
type Suggested struct {
	ISR Suggestion
	IVA Suggestion
}
