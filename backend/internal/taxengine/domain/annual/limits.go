package annual

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Limits are the legal limits of the personal deductions of one fiscal year
// (docs/research/annual-return-2026.md §B.2). UMA multiples are whole
// numbers of annual UMA; shares are fractions of an income.
type Limits struct {
	// AnnualUMA is the UMA "elevada al año" (2026: 42,794.64). Art. 151
	// says "salario mínimo"; reading it as UMA is CPA topic
	// uma_for_minimum_wage.
	AnnualUMA money.Cents
	// GlobalCapUMAs and GlobalCapIncomeShare: the capped group is at most
	// min(5 UMA anuales, 15 % of total income) (LISR 151 last ¶).
	GlobalCapUMAs        int64
	GlobalCapIncomeShare money.Rate
	// FuneralCapUMAs: funeral expenses up to 1 UMA anual (fr. II).
	FuneralCapUMAs int64
	// DonationShare of the prior year's acumulables (7 %), of which
	// GovernmentDonationShare (4 %) may go to government donees (fr. III).
	DonationShare           money.Rate
	GovernmentDonationShare money.Rate
	// RetirementIncomeShare (10 %) of the year's acumulables, at most
	// RetirementCapUMAs (5) UMA anuales (fr. V).
	RetirementIncomeShare money.Rate
	RetirementCapUMAs     int64
	// TuitionCaps per student and level (decree DOF 26-12-2013 art. 1.10).
	TuitionCaps map[TuitionLevel]money.Cents
}

// LimitsReader reads the deduction limits of a fiscal year. The ANN-04
// parameter set implements it; tests use fakes.
type LimitsReader interface {
	DeductionLimits(fiscalYear int) (Limits, error)
}

// Validate reports limits that cannot be applied as ErrInvalidLimits.
func (l Limits) Validate() error {
	switch {
	case l.AnnualUMA <= 0:
		return fmt.Errorf("annual UMA %s: %w", l.AnnualUMA, ErrInvalidLimits)
	case !umaMultiple(l.GlobalCapUMAs) || !umaMultiple(l.FuneralCapUMAs) || !umaMultiple(l.RetirementCapUMAs):
		return fmt.Errorf("UMA multiples must be in [1, %d]: %w", maxUMAMultiple, ErrInvalidLimits)
	case !share(l.GlobalCapIncomeShare) || !share(l.DonationShare) || !share(l.RetirementIncomeShare):
		return fmt.Errorf("income shares must be in (0, 1]: %w", ErrInvalidLimits)
	case l.GovernmentDonationShare <= 0 || l.GovernmentDonationShare > l.DonationShare:
		return fmt.Errorf("government donation share must be in (0, donation share]: %w", ErrInvalidLimits)
	}
	for _, level := range tuitionLevels {
		if l.TuitionCaps[level] <= 0 {
			return fmt.Errorf("tuition cap of %s: %w", level, ErrInvalidLimits)
		}
	}
	return nil
}

// maxUMAMultiple bounds a cap in UMA anuales far above any art. 151 cap
// (5), so n × RateScale cannot overflow.
const maxUMAMultiple = 100

func umaMultiple(n int64) bool { return n > 0 && n <= maxUMAMultiple }

func share(r money.Rate) bool { return r > 0 && r <= money.RateScale }

// umas returns n annual UMA, or ErrOverflow.
func (l Limits) umas(n int64) (money.Cents, error) {
	exact, err := l.AnnualUMA.MulRate(money.Rate(n * money.RateScale))
	if err != nil {
		return 0, fmt.Errorf("annual: %d UMA: %w", n, err)
	}
	return exact.Round(), nil
}

// shareOf returns r × income rounded half away from zero to the centavo: a
// cap is shown as a figure, so it is rounded once.
func shareOf(income money.Cents, r money.Rate) (money.Cents, error) {
	exact, err := income.MulRate(r)
	if err != nil {
		return 0, fmt.Errorf("annual: %s × %s: %w", income, r, err)
	}
	return exact.Round(), nil
}
