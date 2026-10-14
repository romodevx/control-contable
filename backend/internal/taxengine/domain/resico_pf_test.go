package domain_test

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

var resicoPFKey = domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}

// params2026 is the embedded 2026 parameter set: the strategies read the
// published RESICO table, rates and caps, never a copy.
func params2026(t testing.TB) *taxparams.Params {
	t.Helper()
	c, err := params.Embedded()
	require.NoError(t, err)
	sel, err := c.Select(2026)
	require.NoError(t, err)
	return sel.Params
}

// resicoInput builds a 2026 input of key for month n with the given months.
func resicoInput(t testing.TB, key domain.StrategyKey, n time.Month, months ...domain.MonthEntry) domain.Input {
	t.Helper()
	p, err := domain.NewPeriod(2026, n)
	require.NoError(t, err)
	sums, err := domain.NewYearSums(n, months...)
	require.NoError(t, err)
	return domain.Input{Key: key, Period: p, Sums: sums}
}

func cashMonth(t testing.TB, m time.Month, cash string) domain.MonthEntry {
	t.Helper()
	return domain.MonthEntry{Month: m, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: cents(t, cash)}}}
}

type lineView struct {
	Code   domain.LineCode
	Op     domain.Operation
	Amount string
}

func isrLines(lines []domain.Line) []lineView {
	var out []lineView
	for _, l := range lines {
		if l.Tax == domain.TaxISR {
			out = append(out, lineView{Code: l.Code, Op: l.Operation, Amount: l.Amount.String()})
		}
	}
	return out
}

func TestRESICOPFIdentity(t *testing.T) {
	s := domain.NewRESICOPFStrategy()

	assert.Equal(t, resicoPFKey, s.Key())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
	assert.Empty(t, s.RequiredInputs(domain.Options{}, domain.Period{}, domain.Facts{ActiveMonths: []time.Month{1, 2, 3}}),
		"RESICO PF is monthly and definitive in practice: no prior payments (LISR 113-E)")
}

// TestRESICOPFWorkedExamples are PF-626-ISR-01 and PF-626-ISR-02
// (docs/research/tax-regimes-mx-2026.md §A.5): the rate of the month's
// bracket applies to the WHOLE income (not marginal), minus the 1.25 %
// withheld by personas morales as recorded.
func TestRESICOPFWorkedExamples(t *testing.T) {
	tests := []struct {
		name     string
		cash     string
		withheld string
		wantRate string
		wantISR  string
		want     []lineView
	}{
		{
			name: "PF-626-ISR-01", cash: "60000.00", withheld: "250.00", wantRate: "0.015", wantISR: "650.00",
			want: []lineView{
				{domain.LineRESICOPFIncomeMonth, domain.OpInput, "60000.00"},
				{domain.LineRESICOPFRate, domain.OpMultiply, "900.00"},
				{domain.LineRESICOPFWithheldByMoral, domain.OpSubtract, "250.00"},
				{domain.LineRESICOPFPayable, domain.OpResult, "650.00"},
			},
		},
		{
			name: "PF-626-ISR-02 a, upper edge of row 1", cash: "25000.00", wantRate: "0.01", wantISR: "250.00",
			want: []lineView{
				{domain.LineRESICOPFIncomeMonth, domain.OpInput, "25000.00"},
				{domain.LineRESICOPFRate, domain.OpMultiply, "250.00"},
				{domain.LineRESICOPFPayable, domain.OpResult, "250.00"},
			},
		},
		{
			name: "PF-626-ISR-02 b, one centavo into row 2", cash: "25000.01", wantRate: "0.011", wantISR: "275.00",
			want: []lineView{
				{domain.LineRESICOPFIncomeMonth, domain.OpInput, "25000.01"},
				{domain.LineRESICOPFRate, domain.OpMultiply, "275.00"},
				{domain.LineRESICOPFPayable, domain.OpResult, "275.00"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			month := domain.MonthSums{Income: domain.IncomeSums{Cash: cents(t, tt.cash)}}
			if tt.withheld != "" {
				month.Income.ISRWithheld.ByMoral = cents(t, tt.withheld)
			}
			in := resicoInput(t, resicoPFKey, time.March, domain.MonthEntry{Month: time.March, Sums: month})

			est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

			require.NoError(t, err)
			require.NotNil(t, est.ISR)
			assert.Equal(t, tt.wantISR, est.ISR.Amount.String())
			assert.Equal(t, tt.want, isrLines(est.Lines))
			assert.Equal(t, tt.wantRate, lineByCode(t, est.Lines, domain.LineRESICOPFRate).Rate.String())
			assert.Equal(t, domain.StatusComplete, est.Status())
			assert.Equal(t, domain.TreatmentFull, est.Coverage)
			assert.Empty(t, est.Disclaimers, "no deductions in RESICO PF: nothing is claimed deductible")
			assertCited(t, est.Lines)
		})
	}
}

func TestRESICOPFLinesCiteTheLaw(t *testing.T) {
	month := domain.MonthSums{Income: domain.IncomeSums{
		Cash: cents(t, "60000.00"), ISRWithheld: domain.Withheld{ByMoral: cents(t, "250.00")},
	}}
	in := resicoInput(t, resicoPFKey, time.March, domain.MonthEntry{Month: time.March, Sums: month})

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	tests := []struct {
		code    domain.LineCode
		article string
		source  string
	}{
		{domain.LineRESICOPFIncomeMonth, "113-E", "records"},
		{domain.LineRESICOPFRate, "113-E", "params:isr.resico_pf.monthly"},
		{domain.LineRESICOPFWithheldByMoral, "113-J", "records"},
		{domain.LineRESICOPFPayable, "113-E", "computed"},
	}
	for _, tt := range tests {
		l := lineByCode(t, est.Lines, tt.code)
		assert.Equal(t, "LISR", l.Basis[0].Law, tt.code)
		assert.Contains(t, articles(l.Basis), tt.article, tt.code)
		assert.Equal(t, tt.source, l.Source.String(), tt.code)
		assert.False(t, l.NeedsCPAReview, tt.code)
	}
	assert.Equal(t, []string{"3"}, lineByCode(t, est.Lines, domain.LineRESICOPFRate).Args, "60,000 is in row 3 of the monthly table")
}

// TestRESICOPFIsMonthly: only the period's month is taxed; earlier months
// were their own (definitive) payments.
func TestRESICOPFIsMonthly(t *testing.T) {
	in := resicoInput(t, resicoPFKey, time.March,
		cashMonth(t, time.January, "200000.00"), cashMonth(t, time.March, "10000.00"))

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Equal(t, "100.00", est.ISR.Amount.String())
}

func TestRESICOPFWithoutIncomeIsZero(t *testing.T) {
	in := resicoInput(t, resicoPFKey, time.March)

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	require.NotNil(t, est.ISR)
	assert.Equal(t, domain.OutcomeZero, est.ISR.Outcome())
	assert.Equal(t, []lineView{
		{domain.LineRESICOPFIncomeMonth, domain.OpInput, "0.00"},
		{domain.LineRESICOPFPayable, domain.OpResult, "0.00"},
	}, isrLines(est.Lines), "no income, no bracket: the table starts at 0.01")
	assert.Equal(t, domain.StatusComplete, est.Status())
}

// TestRESICOPFWithholdingAboveTheTax: 1.25 % withheld on the whole income
// is more than the 1 % of row 1, so the month is in favor.
func TestRESICOPFWithholdingAboveTheTax(t *testing.T) {
	month := domain.MonthSums{Income: domain.IncomeSums{
		Cash: cents(t, "20000.00"), ISRWithheld: domain.Withheld{ByMoral: cents(t, "250.00")},
	}}
	in := resicoInput(t, resicoPFKey, time.March, domain.MonthEntry{Month: time.March, Sums: month})

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Equal(t, "-50.00", est.ISR.Amount.String())
	assert.Equal(t, domain.OutcomeInFavor, est.ISR.Outcome())
	assert.Zero(t, est.ISR.Payable())
}

// TestRESICOPFAnnualCap: past 3,500,000 collected in the year the taxpayer
// leaves RESICO from the next month (LISR 113-E): the month is still
// computed, with an info line saying so.
func TestRESICOPFAnnualCap(t *testing.T) {
	tests := []struct {
		name    string
		january string
		wantCap bool
	}{
		{name: "at the cap", january: "3400000.00", wantCap: false},
		{name: "one centavo above", january: "3400000.01", wantCap: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := resicoInput(t, resicoPFKey, time.February,
				cashMonth(t, time.January, tt.january), cashMonth(t, time.February, "100000.00"))

			est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

			require.NoError(t, err)
			require.NotNil(t, est.ISR)
			assert.Equal(t, "2000.00", est.ISR.Amount.String(), "100,000 × 2 %")
			assert.Equal(t, domain.StatusComplete, est.Status())
			assert.Equal(t, tt.wantCap, containsCode(est.Lines, domain.LineRESICOPFCapExceeded))
			if tt.wantCap {
				l := lineByCode(t, est.Lines, domain.LineRESICOPFCapExceeded)
				assert.Equal(t, domain.OpInfo, l.Operation)
				assert.Equal(t, "3500000.01", l.Amount.String(), "the year's collected income")
				assert.Contains(t, articles(l.Basis), "113-E")
			}
		})
	}
}

// TestRESICOPFMonthAboveTheTable: a month above 3,500,000 has no RESICO
// rate (the monthly table ends there): no ISR figure, a partial estimate
// that says why, and the IVA still computed.
func TestRESICOPFMonthAboveTheTable(t *testing.T) {
	month := domain.MonthSums{Income: domain.IncomeSums{
		Cash:   cents(t, "3500000.01"),
		ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "3500000.01"), IVA: cents(t, "560000.00")}},
	}}
	in := resicoInput(t, resicoPFKey, time.January, domain.MonthEntry{Month: time.January, Sums: month})

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	assert.Nil(t, est.ISR)
	assert.Equal(t, domain.TreatmentPartial, est.Coverage)
	assert.Equal(t, domain.StatusPartial, est.Status())
	assert.Equal(t, domain.LabelRESICOPFAboveTable, est.Explanation)
	assert.Equal(t, []lineView{{domain.LineRESICOPFCapExceeded, domain.OpInfo, "3500000.01"}}, isrLines(est.Lines))
	require.NotNil(t, est.IVA)
	assert.Equal(t, "560000.00", est.IVA.ChargedCollected.String())
}

// TestRESICOPFIVAIsTheSharedCalculator is PF-626-IVA-01: RESICO PF has no
// IVA regime of its own (research A.3), so the shared components apply.
func TestRESICOPFIVAIsTheSharedCalculator(t *testing.T) {
	month := domain.MonthSums{
		Income: domain.IncomeSums{
			Cash:        cents(t, "60000.00"),
			ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "60000.00"), IVA: cents(t, "9600.00")}},
			IVAWithheld: domain.Withheld{ByMoral: twoThirds(t, cents(t, "3200.00"))},
			ISRWithheld: domain.Withheld{ByMoral: cents(t, "250.00")},
		},
		Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "2000.00"), IVAWithheldToRemit: cents(t, "160.00")},
	}
	in := resicoInput(t, resicoPFKey, time.March, domain.MonthEntry{Month: time.March, Sums: month})

	est, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

	require.NoError(t, err)
	require.NotNil(t, est.IVA)
	assert.Equal(t, domain.IVAComponents{
		ChargedCollected: cents(t, "9600.00"), CreditablePaid: cents(t, "2000.00"),
		WithheldFromTaxpayer: cents(t, "2133.33"),
	}, *est.IVA)
	assert.Equal(t, "5466.67", lineByCode(t, est.Lines, domain.LineIVANet).Amount.String())
	assert.Equal(t, cents(t, "160.00"), est.WithheldToRemit)
	assert.Equal(t, "650.00", est.ISR.Amount.String())
}

func TestRESICOPFPropagatesErrors(t *testing.T) {
	t.Run("a period outside the sums", func(t *testing.T) {
		in := resicoInput(t, resicoPFKey, time.March)
		in.Sums = domain.YearSums{}

		_, err := domain.NewRESICOPFStrategy().Estimate(in, params2026(t))

		require.ErrorIs(t, err, domain.ErrInvalidMonth)
	})
	t.Run("parameters without the RESICO table", func(t *testing.T) {
		_, err := domain.NewRESICOPFStrategy().Estimate(resicoInput(t, resicoPFKey, time.March), &taxparams.Params{})

		require.Error(t, err)
	})
}

// TestRESICOPFEffectiveRateIsNonDecreasing is the property of ADR-0002 §5:
// more income never lowers the RESICO PF rate (nor the tax), across every
// row edge and a sweep of the whole table.
func TestRESICOPFEffectiveRateIsNonDecreasing(t *testing.T) {
	p := params2026(t)
	table, err := p.Table("isr.resico_pf.monthly", mustPeriod(t, time.January).End())
	require.NoError(t, err)

	var amounts []money.Cents
	for _, r := range table.Rows {
		amounts = append(amounts, r.Lower-1, r.Lower, r.Lower+1, r.Upper-1, r.Upper)
	}
	const sweepSteps = 2000
	ceiling := table.Rows[len(table.Rows)-1].Upper
	for i := 1; i <= sweepSteps; i++ {
		amounts = append(amounts, ceiling/sweepSteps*money.Cents(i))
	}
	slices.Sort(amounts)

	var prevRate money.Rate
	var prevISR money.Exact
	for _, amount := range amounts {
		if amount <= 0 {
			continue
		}
		in := resicoInput(t, resicoPFKey, time.January,
			domain.MonthEntry{Month: time.January, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: amount}}})
		est, err := domain.NewRESICOPFStrategy().Estimate(in, p)
		require.NoError(t, err, amount)
		require.NotNil(t, est.ISR, amount)
		r := lineByCode(t, est.Lines, domain.LineRESICOPFRate).Rate

		require.GreaterOrEqual(t, r, prevRate, "rate decreased at %s", amount)
		require.GreaterOrEqual(t, est.ISR.Exact, prevISR, "tax decreased at %s", amount)
		prevRate, prevISR = r, est.ISR.Exact
	}
}

func mustPeriod(t testing.TB, m time.Month) domain.Period {
	t.Helper()
	p, err := domain.NewPeriod(2026, m)
	require.NoError(t, err)
	return p
}

func containsCode(lines []domain.Line, code domain.LineCode) bool {
	for _, l := range lines {
		if l.Code == code {
			return true
		}
	}
	return false
}
