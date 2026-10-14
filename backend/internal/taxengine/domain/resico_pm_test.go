package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

var resicoPMKey = domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}

func cashFlowMonth(t testing.TB, m time.Month, income, deductible string) domain.MonthEntry {
	t.Helper()
	return domain.MonthEntry{Month: m, Sums: domain.MonthSums{
		Income:   domain.IncomeSums{Cash: cents(t, income)},
		Expenses: domain.ExpenseSums{DeductiblePaid: cents(t, deductible)},
	}}
}

func moneyInput(t testing.TB, key domain.InputKey, amount string) domain.TaxInputValue {
	t.Helper()
	return domain.TaxInputValue{Key: key, Kind: domain.InputMoney, Money: cents(t, amount)}
}

func priorPayment(t testing.TB, m time.Month, amount string) domain.TaxInputValue {
	t.Helper()
	return domain.TaxInputValue{
		Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m, Money: cents(t, amount),
	}
}

// pm626ISR01 is PM-626-ISR-01 (research §B.5): March, YTD collected
// 1,500,000, YTD deductions paid 1,100,000, prior payments 70,000.
func pm626ISR01(t testing.TB, extra ...domain.TaxInputValue) domain.Input {
	t.Helper()
	in := resicoInput(t, resicoPMKey, time.March,
		cashFlowMonth(t, time.January, "500000.00", "400000.00"),
		cashFlowMonth(t, time.February, "500000.00", "350000.00"),
		cashFlowMonth(t, time.March, "500000.00", "350000.00"))
	values := append([]domain.TaxInputValue{
		priorPayment(t, time.January, "35000.00"), priorPayment(t, time.February, "35000.00"),
	}, extra...)
	inputs, err := domain.NewTaxInputs(values...)
	require.NoError(t, err)
	in.Inputs = inputs
	return in
}

func TestRESICOPMIdentity(t *testing.T) {
	s := domain.NewRESICOPMStrategy()

	assert.Equal(t, resicoPMKey, s.Key())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
}

// TestRESICOPMRequiredInputs: from February on, the prior provisional
// payment of every earlier month the regime was active (MODULE-tax-engine
// §2.8); the other inputs are optional.
func TestRESICOPMRequiredInputs(t *testing.T) {
	tests := []struct {
		name   string
		month  time.Month
		active []time.Month
		want   []time.Month
	}{
		{name: "January needs nothing", month: time.January, active: []time.Month{1}},
		{name: "March, active all year", month: time.March, active: []time.Month{1, 2, 3}, want: []time.Month{1, 2}},
		{name: "March, active from February", month: time.March, active: []time.Month{2, 3}, want: []time.Month{2}},
		{name: "March, first active month", month: time.March, active: []time.Month{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.NewRESICOPMStrategy().RequiredInputs(domain.Options{}, mustPeriod(t, tt.month),
				domain.Facts{ActiveMonths: tt.active})

			if tt.want == nil {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, []domain.RequiredInput{{
				Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: tt.want,
			}}, got)
		})
	}
}

func TestRESICOPMWorkedExample(t *testing.T) {
	est, err := domain.NewRESICOPMStrategy().Estimate(pm626ISR01(t), params2026(t))

	require.NoError(t, err)
	require.NotNil(t, est.ISR)
	assert.Equal(t, "50000.00", est.ISR.Amount.String())
	assert.Equal(t, []lineView{
		{domain.LineRESICOPMIncomeYTD, domain.OpInput, "1500000.00"},
		{domain.LineRESICOPMDeductionsYTD, domain.OpSubtract, "1100000.00"},
		{domain.LineRESICOPMBase, domain.OpResult, "400000.00"},
		{domain.LineRESICOPMRate, domain.OpMultiply, "120000.00"},
		{domain.LineRESICOPMPriorPayments, domain.OpSubtract, "70000.00"},
		{domain.LineRESICOPMPayable, domain.OpResult, "50000.00"},
	}, isrLines(est.Lines))
	assert.Equal(t, "0.3", lineByCode(t, est.Lines, domain.LineRESICOPMRate).Rate.String())
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed}, est.Disclaimers)
	assert.Empty(t, est.CPATopics)
	assert.Equal(t, domain.TreatmentFull, est.Coverage)
	assertCited(t, est.Lines)
}

func TestRESICOPMLinesCiteTheLaw(t *testing.T) {
	est, err := domain.NewRESICOPMStrategy().Estimate(pm626ISR01(t), params2026(t))

	require.NoError(t, err)
	tests := []struct {
		code    domain.LineCode
		article string
		source  string
	}{
		{domain.LineRESICOPMIncomeYTD, "207", "records"},
		{domain.LineRESICOPMDeductionsYTD, "208", "records"},
		{domain.LineRESICOPMBase, "211", "computed"},
		{domain.LineRESICOPMRate, "9", "params:isr.rate.persona_moral"},
		{domain.LineRESICOPMPriorPayments, "211", "tax_input:prior_provisional_payments"},
		{domain.LineRESICOPMPayable, "211", "computed"},
	}
	for _, tt := range tests {
		l := lineByCode(t, est.Lines, tt.code)
		assert.Equal(t, "LISR", l.Basis[0].Law, tt.code)
		assert.Contains(t, articles(l.Basis), tt.article, tt.code)
		assert.Equal(t, tt.source, l.Source.String(), tt.code)
	}
}

// TestRESICOPMOptionalInputs: what the year held before the app, PTU paid
// and prior-year losses enter the base; the art. 54 interest withholding is
// credited as an open CPA question (§2.9).
func TestRESICOPMOptionalInputs(t *testing.T) {
	in := pm626ISR01(t,
		moneyInput(t, domain.InputYTDBeforeAppIncome, "100000.00"),
		moneyInput(t, domain.InputYTDBeforeAppDeductions, "40000.00"),
		moneyInput(t, domain.InputPTUPaid, "10000.00"),
		moneyInput(t, domain.InputPriorYearLosses, "50000.00"),
		moneyInput(t, domain.InputInterestWithholding, "1000.00"),
	)

	est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	// 1,500,000 + 100,000 − 1,100,000 − 40,000 − 10,000 − 50,000 = 400,000.
	assert.Equal(t, []lineView{
		{domain.LineRESICOPMIncomeYTD, domain.OpInput, "1500000.00"},
		{domain.LineRESICOPMIncomeBeforeApp, domain.OpAdd, "100000.00"},
		{domain.LineRESICOPMDeductionsYTD, domain.OpSubtract, "1100000.00"},
		{domain.LineRESICOPMDeductionsBeforeApp, domain.OpSubtract, "40000.00"},
		{domain.LineRESICOPMPTUPaid, domain.OpSubtract, "10000.00"},
		{domain.LineRESICOPMPriorYearLosses, domain.OpSubtract, "50000.00"},
		{domain.LineRESICOPMBase, domain.OpResult, "400000.00"},
		{domain.LineRESICOPMRate, domain.OpMultiply, "120000.00"},
		{domain.LineRESICOPMPriorPayments, domain.OpSubtract, "70000.00"},
		{domain.LineRESICOPMInterestWithholding, domain.OpSubtract, "1000.00"},
		{domain.LineRESICOPMPayable, domain.OpResult, "49000.00"},
	}, isrLines(est.Lines))
	assert.Equal(t, "49000.00", est.ISR.Amount.String())
	assert.Equal(t, []domain.CPATopic{domain.TopicRESICOPMInterestWithholding}, est.CPATopics)
	assert.True(t, lineByCode(t, est.Lines, domain.LineRESICOPMInterestWithholding).NeedsCPAReview)
	for _, code := range []domain.LineCode{
		domain.LineRESICOPMIncomeBeforeApp, domain.LineRESICOPMDeductionsBeforeApp, domain.LineRESICOPMPTUPaid,
		domain.LineRESICOPMPriorYearLosses, domain.LineRESICOPMInterestWithholding,
	} {
		assert.Contains(t, lineByCode(t, est.Lines, code).Source.String(), "tax_input:", code)
	}
	assertCited(t, est.Lines)
}

// TestRESICOPMLossIsAZeroBase: a cash-flow loss has no 30 % to pay; the
// base is taken as zero, so prior payments leave the year in favor.
func TestRESICOPMLossIsAZeroBase(t *testing.T) {
	in := resicoInput(t, resicoPMKey, time.February,
		cashFlowMonth(t, time.January, "100000.00", "50000.00"),
		cashFlowMonth(t, time.February, "100000.00", "200000.00"))
	inputs, err := domain.NewTaxInputs(priorPayment(t, time.January, "15000.00"))
	require.NoError(t, err)
	in.Inputs = inputs

	est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Equal(t, []lineView{
		{domain.LineRESICOPMIncomeYTD, domain.OpInput, "200000.00"},
		{domain.LineRESICOPMDeductionsYTD, domain.OpSubtract, "250000.00"},
		{domain.LineRESICOPMBase, domain.OpResult, "-50000.00"},
		{domain.LineRESICOPMNoProfit, domain.OpInput, "0.00"},
		{domain.LineRESICOPMRate, domain.OpMultiply, "0.00"},
		{domain.LineRESICOPMPriorPayments, domain.OpSubtract, "15000.00"},
		{domain.LineRESICOPMPayable, domain.OpResult, "-15000.00"},
	}, isrLines(est.Lines))
	assert.Equal(t, domain.OutcomeInFavor, est.ISR.Outcome())
	assertCited(t, est.Lines)
}

func TestRESICOPMJanuaryHasNoPriorPayments(t *testing.T) {
	in := resicoInput(t, resicoPMKey, time.January, cashFlowMonth(t, time.January, "100000.00", "0.00"))

	est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Equal(t, "30000.00", est.ISR.Amount.String())
	assert.False(t, containsCode(est.Lines, domain.LineRESICOPMPriorPayments))
	assert.Empty(t, est.Disclaimers, "no deductible expense entered the base")
}

// TestRESICOPMAnnualCap: past 35,000,000 collected in the year the entity
// leaves RESICO from the next fiscal year (LISR 206 ¶2, 214): this year is
// still computed, with an info line saying so.
func TestRESICOPMAnnualCap(t *testing.T) {
	tests := []struct {
		name    string
		income  string
		wantCap bool
	}{
		{name: "at the cap", income: "35000000.00"},
		{name: "one centavo above", income: "35000000.01", wantCap: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := resicoInput(t, resicoPMKey, time.January, cashFlowMonth(t, time.January, tt.income, "0.00"))

			est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

			require.NoError(t, err)
			require.NotNil(t, est.ISR)
			assert.Equal(t, domain.StatusComplete, est.Status())
			assert.Equal(t, tt.wantCap, containsCode(est.Lines, domain.LineRESICOPMCapExceeded))
			if tt.wantCap {
				l := lineByCode(t, est.Lines, domain.LineRESICOPMCapExceeded)
				assert.Equal(t, domain.OpInfo, l.Operation)
				assert.Equal(t, tt.income, l.Amount.String())
				assert.Contains(t, articles(l.Basis), "206")
			}
		})
	}
}

func TestRESICOPMCapCountsIncomeBeforeTheApp(t *testing.T) {
	in := resicoInput(t, resicoPMKey, time.January, cashFlowMonth(t, time.January, "30000000.00", "0.00"))
	inputs, err := domain.NewTaxInputs(moneyInput(t, domain.InputYTDBeforeAppIncome, "5000000.01"))
	require.NoError(t, err)
	in.Inputs = inputs

	est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Equal(t, "35000000.01", lineByCode(t, est.Lines, domain.LineRESICOPMCapExceeded).Amount.String())
}

func TestRESICOPMIVAIsTheSharedCalculator(t *testing.T) {
	march := domain.MonthSums{
		Income: domain.IncomeSums{
			Cash:   cents(t, "100000.00"),
			ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "100000.00"), IVA: cents(t, "16000.00")}},
		},
		Expenses: domain.ExpenseSums{
			DeductiblePaid: cents(t, "50000.00"), CreditableIVAPaid: cents(t, "8000.00"),
			IVAWithheldToRemit: cents(t, "1066.67"),
		},
	}
	in := resicoInput(t, resicoPMKey, time.March,
		cashFlowMonth(t, time.January, "999999.00", "0.00"), domain.MonthEntry{Month: time.March, Sums: march})

	est, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	require.NotNil(t, est.IVA)
	assert.Equal(t, domain.IVAComponents{ChargedCollected: cents(t, "16000.00"), CreditablePaid: cents(t, "8000.00")},
		*est.IVA, "IVA is the month's alone, never YTD (LIVA 5-D)")
	assert.Equal(t, cents(t, "1066.67"), est.WithheldToRemit)
}

func TestRESICOPMPropagatesErrors(t *testing.T) {
	t.Run("a period outside the sums", func(t *testing.T) {
		in := pm626ISR01(t)
		in.Sums = domain.YearSums{}

		_, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

		require.ErrorIs(t, err, domain.ErrInvalidMonth)
	})
	t.Run("parameters without the PM rate", func(t *testing.T) {
		_, err := domain.NewRESICOPMStrategy().Estimate(pm626ISR01(t), &taxparams.Params{})

		require.Error(t, err)
	})
	t.Run("an overflowing year", func(t *testing.T) {
		huge := money.Cents(1) << 62
		in := resicoInput(t, resicoPMKey, time.February,
			domain.MonthEntry{Month: time.January, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: huge}}},
			domain.MonthEntry{Month: time.February, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: huge}}})

		_, err := domain.NewRESICOPMStrategy().Estimate(in, params2026(t))

		require.ErrorIs(t, err, money.ErrOverflow)
	})
}
