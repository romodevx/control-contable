package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// The 622 PM and 603 PM tests name their helpers with a ppm prefix: every
// regime strategy test shares package domain_test (PLAN PG-M5-S).

// ppmStrategies are the partial personas morales whose ISR the app does not
// estimate (MODULE-tax-engine §2.4): 622 AGAPES (exemption and reduction
// deferred) and 603 fines no lucrativos (no provisional ISR).
func ppmStrategies() []domain.Strategy {
	return []domain.Strategy{domain.NewStrategy622PM(), domain.NewStrategy603PM()}
}

// ppmJune is PM-601-IVA-01's month (research §B.5) under key: IVA charged
// and collected 320,000, creditable 200,000, and 5,333.33 of IVA withheld
// from a PF lawyer to remit. May has IVA that June must not see.
func ppmJune(t *testing.T, key domain.StrategyKey, june domain.MonthSums) domain.Input {
	t.Helper()
	may := domain.MonthSums{Income: domain.IncomeSums{ByRate: domain.IVAByRate{
		Rate16: domain.RateSums{Base: cents(t, "100.00"), IVA: cents(t, "16.00")},
	}}}
	ys, err := domain.NewYearSums(time.June,
		domain.MonthEntry{Month: time.May, Sums: may}, domain.MonthEntry{Month: time.June, Sums: june})
	require.NoError(t, err)
	return domain.Input{
		Key: key, Period: period(t, "2026-06"), Sums: ys,
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(true)}),
	}
}

func ppmPM601IVA01(t *testing.T) domain.MonthSums {
	t.Helper()
	return domain.MonthSums{
		Income: domain.IncomeSums{Accrual: cents(t, "2000000.00"), ByRate: domain.IVAByRate{
			Rate16: domain.RateSums{Base: cents(t, "2000000.00"), IVA: cents(t, "320000.00")},
		}},
		Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "200000.00"), IVAWithheldToRemit: cents(t, "5333.33")},
	}
}

func TestPartialPMStrategiesIdentity(t *testing.T) {
	keys := []domain.StrategyKey{
		{Regime: "622", PersonType: fiscal.PersonMoral}, {Regime: "603", PersonType: fiscal.PersonMoral},
	}
	for i, s := range ppmStrategies() {
		t.Run(keys[i].String(), func(t *testing.T) {
			facts := domain.Facts{ActiveMonths: []time.Month{time.January, time.February, time.March}}

			assert.Equal(t, keys[i], s.Key())
			assert.Equal(t, domain.TreatmentPartial, s.Treatment())
			assert.Empty(t, s.RequiredInputs(domain.Options{}, period(t, "2026-03"), facts),
				"ISR is not estimated, so no ISR input is required (§2.8: 622 none in v1)")
		})
	}
}

// The IVA is the month's, through the shared components: PM-601-IVA-01's
// figures hold for any persona moral under the monthly LIVA rules.
func TestPartialPMStrategiesEstimateTheMonthsIVAOnly(t *testing.T) {
	for _, s := range ppmStrategies() {
		t.Run(s.Key().String(), func(t *testing.T) {
			est, err := s.Estimate(ppmJune(t, s.Key(), ppmPM601IVA01(t)), nil)
			require.NoError(t, err)

			assert.Equal(t, s.Key(), est.Key)
			assert.Nil(t, est.ISR, "ISR not estimated: explained instead")
			require.NotNil(t, est.IVA)
			assert.Equal(t, domain.IVAComponents{
				ChargedCollected: cents(t, "320000.00"), CreditablePaid: cents(t, "200000.00"),
			}, *est.IVA)
			assert.Equal(t, cents(t, "5333.33"), est.WithheldToRemit)
			assert.Equal(t, domain.TreatmentPartial, est.Coverage)
			assert.Equal(t, domain.StatusPartial, est.Status())
			assert.Equal(t, domain.RegimeExplanation(s.Key()), est.Explanation)
			assert.Empty(t, est.CPATopics)
			for _, l := range est.Lines {
				assert.NotEqual(t, domain.TaxISR, l.Tax, "no ISR line: %s", l.Code)
			}
			assertCited(t, est.Lines)
		})
	}
}

// A persona moral remits the ISR it withheld from its suppliers with the
// IVA it withheld (LISR 106 ¶5, 116 ¶5, 113-J; LIVA 1-A), as 601 does.
func TestPartialPMStrategiesRemitTheISRAndIVATheyWithheld(t *testing.T) {
	for _, s := range ppmStrategies() {
		t.Run(s.Key().String(), func(t *testing.T) {
			june := ppmPM601IVA01(t)
			june.Expenses.ISRWithheldToRemit = cents(t, "5000.00")

			est, err := s.Estimate(ppmJune(t, s.Key(), june), nil)
			require.NoError(t, err)

			assert.Equal(t, cents(t, "10333.33"), est.WithheldToRemit)
			assert.Equal(t, cents(t, "10333.33"), lineByCode(t, est.Lines, "withholding.to_remit").Amount)
		})
	}
}

func TestPartialPMStrategiesPassTheIVADisclaimersThrough(t *testing.T) {
	for _, s := range ppmStrategies() {
		t.Run(s.Key().String(), func(t *testing.T) {
			in := ppmJune(t, s.Key(), ppmPM601IVA01(t))
			in.Options = domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(false)})

			est, err := s.Estimate(in, nil)
			require.NoError(t, err)

			assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}, est.Disclaimers)
		})
	}
}

// Through the registry: partial, never incomplete, even without the 601
// inputs 622 keeps in its catalog definitions.
func TestPartialPMStrategiesThroughTheRegistry(t *testing.T) {
	r, err := domain.NewRegistry(ppmStrategies()...)
	require.NoError(t, err)
	for _, s := range ppmStrategies() {
		t.Run(s.Key().String(), func(t *testing.T) {
			months := []time.Month{time.January, time.February, time.March, time.April, time.May, time.June}

			est, err := r.Estimate(ppmJune(t, s.Key(), ppmPM601IVA01(t)), months, nil)
			require.NoError(t, err)

			assert.Equal(t, domain.StatusPartial, est.Status())
			assert.Empty(t, est.MissingInputs)
		})
	}
}
