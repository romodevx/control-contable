package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// LIVA 5 fr. IV (MODULE-withholdings §4.2): the IVA a taxpayer withheld from
// a supplier is creditable "en la declaración de pago mensual siguiente a la
// declaración en la que se haya efectuado el entero de la retención".
// Withheld in June, remitted with the June return (by 17 July), credited in
// the July return. The withheld-credit tests name their helpers with a wc
// prefix: every strategy test shares package domain_test.

// wcWithheld is a month whose expenses withheld IVA, creditable in full.
func wcWithheld(c money.Cents) domain.MonthSums {
	return domain.MonthSums{Expenses: domain.ExpenseSums{CreditableIVAWithheld: c}}
}

func TestYearSumsCreditableWithheldPriorMonth(t *testing.T) {
	y, err := domain.NewYearSums(time.July,
		domain.MonthEntry{Month: time.June, Sums: wcWithheld(4266_67)},
		domain.MonthEntry{Month: time.July, Sums: wcWithheld(1_00)},
	)
	require.NoError(t, err)

	assert.Equal(t, money.Cents(4266_67), y.CreditableWithheldPriorMonth(time.July), "June's, credited in July")
	assert.Zero(t, y.CreditableWithheldPriorMonth(time.June), "May withheld nothing")
	assert.Zero(t, y.CreditableWithheldPriorMonth(time.January), "no previous December was given")
	assert.Zero(t, y.CreditableWithheldPriorMonth(time.August), "after the period")
	assert.Zero(t, y.CreditableWithheldPriorMonth(0))

	december := y.WithPriorDecember(321_00)
	assert.Equal(t, money.Cents(321_00), december.CreditableWithheldPriorMonth(time.January),
		"January credits the previous fiscal year's December (LIVA is monthly, not annual)")
	assert.Equal(t, y.Month(time.June), december.Month(time.June), "the months are untouched")
	assert.Zero(t, y.CreditableWithheldPriorMonth(time.January), "WithPriorDecember returns a copy")
}

func TestMonthSumsAddSumsTheWithheldCreditableIVA(t *testing.T) {
	sum, err := wcWithheld(2_00).Add(wcWithheld(3_00))
	require.NoError(t, err)
	assert.Equal(t, money.Cents(5_00), sum.Expenses.CreditableIVAWithheld)

	_, err = wcWithheld(math.MaxInt64).Add(wcWithheld(1))
	require.ErrorIs(t, err, money.ErrOverflow)
}

// wcJuly is WH-11's July for a persona moral (MODULE-withholdings US-W6):
// no July records, and June withheld 4,266.67 of IVA from a PF consultant.
func wcJuly(t *testing.T, key domain.StrategyKey) domain.Input {
	t.Helper()
	ys, err := domain.NewYearSums(time.July, domain.MonthEntry{Month: time.June, Sums: domain.MonthSums{
		Expenses: domain.ExpenseSums{
			DeductiblePaid: cents(t, "40000.00"), CreditableIVAPaid: cents(t, "2133.33"),
			CreditableIVAWithheld: cents(t, "4266.67"),
			ISRWithheldToRemit:    cents(t, "4000.00"), IVAWithheldToRemit: cents(t, "4266.67"),
		},
	}})
	require.NoError(t, err)
	return domain.Input{
		Key: key, Period: period(t, "2026-07"), Sums: ys,
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(true)}),
	}
}

func TestComputeIVACreditsTheWithheldIVAOfThePriorMonth(t *testing.T) {
	l := domain.NewLedger()
	got, err := domain.ComputeIVA(l, domain.IVAInput{
		Sums: domain.MonthSums{
			Income:   domain.IncomeSums{ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 10000_00, IVA: 1600_00}}},
			Expenses: domain.ExpenseSums{CreditableIVAPaid: 100_00, CreditableIVAWithheld: 999_00},
		},
		WithheldPriorMonth: cents(t, "4266.67"),
	})
	require.NoError(t, err)

	assert.Equal(t, cents(t, "4366.67"), got.Components.CreditablePaid,
		"the paid part of the month plus the part withheld the month before; this month's withheld part waits")
	assert.Equal(t, cents(t, "-2766.67"), got.Net)
	assert.Equal(t, []domain.CPATopic{domain.TopicWithheldIVACreditTiming}, got.CPATopics)

	lines, err := l.Lines()
	require.NoError(t, err)
	assert.Equal(t, []domain.LineCode{
		domain.LineIVAChargedRate16, domain.LineIVACharged, domain.LineIVACreditable,
		domain.LineIVACreditableWithheldPriorMonth, domain.LineIVAWithheldByMoral, domain.LineIVANet,
	}, codes(lines))
	assert.Equal(t, money.Cents(100_00), lineByCode(t, lines, domain.LineIVACreditable).Amount, "the paid part")
	prior := lineByCode(t, lines, domain.LineIVACreditableWithheldPriorMonth)
	assert.Equal(t, domain.OpSubtract, prior.Operation)
	assert.Equal(t, cents(t, "4266.67"), prior.Amount)
	assert.Equal(t, "records", prior.Source.String())
	assert.Equal(t, []domain.Citation{{Law: "LIVA", Article: "5", Note: "fracción IV"}}, prior.Basis)
	assert.True(t, prior.NeedsCPAReview, "assumes the withholding was remitted on time (OQ-W4)")
}

func TestComputeIVAWithoutAPriorWithholdingAddsNoLine(t *testing.T) {
	l := domain.NewLedger()
	got, err := domain.ComputeIVA(l, domain.IVAInput{Sums: domain.MonthSums{
		Expenses: domain.ExpenseSums{CreditableIVAPaid: 2133_33, CreditableIVAWithheld: 4266_67},
	}})
	require.NoError(t, err)

	assert.Equal(t, money.Cents(2133_33), got.Components.CreditablePaid, "June credits only the paid part")
	assert.Empty(t, got.CPATopics)
	lines, err := l.Lines()
	require.NoError(t, err)
	assert.NotContains(t, codes(lines), domain.LineIVACreditableWithheldPriorMonth)
}

func TestComputeIVAMixedActivityCountsThePriorWithheldCredit(t *testing.T) {
	got, err := domain.ComputeIVA(domain.NewLedger(), domain.IVAInput{
		Sums:               domain.MonthSums{Income: domain.IncomeSums{ByRate: domain.IVAByRate{ExemptBase: 100_00}}},
		WithheldPriorMonth: 1_00,
	})
	require.NoError(t, err)
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}, got.Disclaimers,
		"the withheld IVA is credited in full although the month has exempt acts")
}

func TestComputeIVAPriorWithheldCreditOverflows(t *testing.T) {
	_, err := domain.ComputeIVA(domain.NewLedger(), domain.IVAInput{
		Sums:               domain.MonthSums{Expenses: domain.ExpenseSums{CreditableIVAPaid: math.MaxInt64}},
		WithheldPriorMonth: 1,
	})
	require.ErrorIs(t, err, money.ErrOverflow)
}

// Every persona moral strategy credits in July the IVA withheld in June
// (WH-11-JUL); the June figures of the paid part stay in June.
func TestPMStrategiesCreditTheIVAWithheldTheMonthBefore(t *testing.T) {
	strategies := []domain.Strategy{
		domain.NewGeneralPMStrategy(), domain.NewRESICOPMStrategy(), domain.NewStrategy622PM(), domain.NewStrategy603PM(),
	}
	for _, s := range strategies {
		t.Run(s.Key().String(), func(t *testing.T) {
			est, err := s.Estimate(wcJuly(t, s.Key()), params2026(t))
			require.NoError(t, err)

			require.NotNil(t, est.IVA)
			assert.Equal(t, cents(t, "4266.67"), est.IVA.CreditablePaid)
			assert.Contains(t, est.CPATopics, domain.TopicWithheldIVACreditTiming)
			assert.Equal(t, cents(t, "4266.67"), lineByCode(t, est.Lines, domain.LineIVACreditableWithheldPriorMonth).Amount)
			assert.Zero(t, est.WithheldToRemit, "June's withholdings are remitted in June's return, not July's")
		})
	}
}

// January credits the previous December's withheld IVA: the credit crosses
// the fiscal year.
func TestGeneralPMCreditsDecembersWithheldIVAInJanuary(t *testing.T) {
	ys, err := domain.NewYearSums(time.January)
	require.NoError(t, err)
	in := domain.Input{
		Key:    domain.StrategyKey{Regime: "601", PersonType: fiscal.PersonMoral},
		Period: period(t, "2027-01"), Sums: ys.WithPriorDecember(cents(t, "5333.33")),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, params2026(t))
	require.NoError(t, err)

	require.NotNil(t, est.IVA)
	assert.Equal(t, cents(t, "5333.33"), est.IVA.CreditablePaid)
}

// 625 under the definitive option takes no creditable IVA (LIVA 18-M ⚠️):
// the withheld part of the month before is shown as not applied with the
// paid part, never credited.
func TestPlatformDefinitiveDoesNotCreditThePriorWithheldIVA(t *testing.T) {
	ys, err := domain.NewYearSums(time.July,
		domain.MonthEntry{Month: time.June, Sums: wcWithheld(300_00)},
		domain.MonthEntry{Month: time.July, Sums: domain.MonthSums{Expenses: domain.ExpenseSums{CreditableIVAPaid: 100_00}}},
	)
	require.NoError(t, err)
	in := domain.Input{
		Key:    domain.StrategyKey{Regime: fiscal.RegimePlatforms, PersonType: fiscal.PersonFisica},
		Period: period(t, "2026-07"), Sums: ys,
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
			"definitive_option": domain.BoolOption(true), "activities": domain.ListOption([]string{"goods_services"}),
			"rfc_given_to_platform": domain.BoolOption(true),
		}),
	}

	est, err := domain.NewStrategy625PF().Estimate(in, params2026(t))
	require.NoError(t, err)

	require.NotNil(t, est.IVA)
	assert.Zero(t, est.IVA.CreditablePaid)
	assert.NotContains(t, codes(est.Lines), domain.LineIVACreditableWithheldPriorMonth)
	assert.Equal(t, money.Cents(400_00), lineByCode(t, est.Lines, "iva.creditable_not_applied").Amount)
}
