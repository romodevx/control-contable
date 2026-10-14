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

// The 601 PM tests name their helpers with a gpm prefix: every regime
// strategy test shares package domain_test (PLAN PG-M5-S runs them in
// parallel), so a generic name would collide.

var gpmKey = domain.StrategyKey{Regime: "601", PersonType: fiscal.PersonMoral}

// gpmParams is a synthetic parameter set holding only the persona moral
// rate: the real 2026 value is transcribed in params/data/2026/v1.yaml and
// the golden gate runs the strategy against it.
func gpmParams(t testing.TB, rate string) *taxparams.Params {
	t.Helper()
	from, err := fiscal.ParseDate("2026-01-01")
	require.NoError(t, err)
	r, err := money.ParseRate(rate)
	require.NoError(t, err)
	p, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: from,
		Sources: []taxparams.Source{{
			ID: "lisr", Title: "Ley del ISR", URL: "https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf",
			DOFDate: from, Reference: "art. 9",
		}},
		Values: []taxparams.Value{{
			Key: "isr.rate.persona_moral", Kind: taxparams.KindRate, Rate: r,
			Validity: taxparams.Validity{From: from}, Basis: []taxparams.Citation{{Source: "lisr", Article: "9"}},
		}},
	})
	require.NoError(t, err)
	return p
}

func gpmRate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

// gpmAccrual is a year of accrued income through month n: amounts[i] is
// month i+1.
func gpmAccrual(t testing.TB, through time.Month, amounts ...string) domain.YearSums {
	t.Helper()
	entries := make([]domain.MonthEntry, 0, len(amounts))
	for i, a := range amounts {
		entries = append(entries, domain.MonthEntry{
			Month: time.Month(i + 1), Sums: domain.MonthSums{Income: domain.IncomeSums{Accrual: cents(t, a)}},
		})
	}
	y, err := domain.NewYearSums(through, entries...)
	require.NoError(t, err)
	return y
}

func gpmMoney(t testing.TB, key domain.InputKey, amount string) domain.TaxInputValue {
	return domain.TaxInputValue{Key: key, Kind: domain.InputMoney, Money: cents(t, amount)}
}

func gpmCoeficiente(t testing.TB, rate string) domain.TaxInputValue {
	return domain.TaxInputValue{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: gpmRate(t, rate)}
}

func gpmPriorPayments(t testing.TB, amount string, months ...time.Month) []domain.TaxInputValue {
	out := make([]domain.TaxInputValue, 0, len(months))
	for _, m := range months {
		out = append(out, domain.TaxInputValue{
			Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m, Money: cents(t, amount),
		})
	}
	return out
}

// gpmLine is the part of a line the tests assert.
type gpmLine struct {
	Tax    domain.TaxKind
	Code   domain.LineCode
	Op     domain.Operation
	Amount string
}

func gpmLines(lines []domain.Line, tax domain.TaxKind) []gpmLine {
	var out []gpmLine
	for _, l := range lines {
		if l.Tax == tax {
			out = append(out, gpmLine{l.Tax, l.Code, l.Operation, l.Amount.String()})
		}
	}
	return out
}

// pm601ISR01 is research §B.5 PM-601-ISR-01: June 2026, CU 0.1500, YTD
// accrued income 5,000,000, PTU paid 120,000, prior payments 180,000.
func pm601ISR01(t *testing.T) domain.Input {
	t.Helper()
	values := append(gpmPriorPayments(t, "36000.00", time.January, time.February, time.March, time.April, time.May),
		gpmCoeficiente(t, "0.1500"),
		gpmMoney(t, domain.InputPTUPaid, "120000.00"),
		gpmMoney(t, domain.InputPriorYearLosses, "0.00"),
	)
	return domain.Input{
		Key:    gpmKey,
		Period: period(t, "2026-06"),
		Sums: gpmAccrual(t, time.June,
			"800000.00", "800000.00", "800000.00", "800000.00", "900000.00", "900000.00"),
		Inputs: inputs(t, values...),
	}
}

func TestGeneralPMStrategyIdentity(t *testing.T) {
	s := domain.NewGeneralPMStrategy()

	assert.Equal(t, gpmKey, s.Key())
	assert.Equal(t, "601/moral", s.Key().String())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
}

func TestGeneralPMStrategyWorkedExampleISR01(t *testing.T) {
	est, err := domain.NewGeneralPMStrategy().Estimate(pm601ISR01(t), gpmParams(t, "0.30"))

	require.NoError(t, err)
	require.NotNil(t, est.ISR)
	assert.Equal(t, "36000.00", est.ISR.Amount.String())
	assert.Equal(t, domain.OutcomePayable, est.ISR.Outcome())
	assert.Equal(t, gpmKey, est.Key)
	assert.Equal(t, domain.TreatmentFull, est.Coverage)
	assert.Empty(t, est.Disclaimers)
	assert.Empty(t, est.CPATopics)
	assert.Equal(t, []gpmLine{
		{domain.TaxISR, "isr.income_ytd", domain.OpInput, "5000000.00"},
		{domain.TaxISR, "isr.coeficiente", domain.OpMultiply, "750000.00"},
		{domain.TaxISR, "isr.ptu", domain.OpSubtract, "30000.00"},
		{domain.TaxISR, "isr.prior_year_losses", domain.OpSubtract, "0.00"},
		{domain.TaxISR, "isr.base", domain.OpResult, "720000.00"},
		{domain.TaxISR, "isr.rate", domain.OpMultiply, "216000.00"},
		{domain.TaxISR, "isr.prior_payments", domain.OpSubtract, "180000.00"},
		{domain.TaxISR, "isr.interest_withholding", domain.OpSubtract, "0.00"},
		{domain.TaxISR, "isr.payable", domain.OpResult, "36000.00"},
	}, gpmLines(est.Lines, domain.TaxISR))
	assertCited(t, est.Lines)
}

func TestGeneralPMStrategyLinesCiteArticle14AndShowTheirSource(t *testing.T) {
	est, err := domain.NewGeneralPMStrategy().Estimate(pm601ISR01(t), gpmParams(t, "0.30"))
	require.NoError(t, err)

	tests := []struct {
		code     domain.LineCode
		articles []string
		source   string
	}{
		{"isr.income_ytd", []string{"14"}, "records"},
		{"isr.coeficiente", []string{"14"}, "tax_input:coeficiente_utilidad"},
		{"isr.ptu", []string{"14"}, "tax_input:ptu_paid"},
		{"isr.prior_year_losses", []string{"14"}, "tax_input:prior_year_losses"},
		{"isr.base", []string{"14"}, "computed"},
		{"isr.rate", []string{"9", "14"}, "params:isr.rate.persona_moral"},
		{"isr.prior_payments", []string{"14"}, "tax_input:prior_provisional_payments"},
		{"isr.interest_withholding", []string{"54", "24"}, "tax_input:interest_withholding"},
		{"isr.payable", []string{"14"}, "computed"},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			l := lineByCode(t, est.Lines, tt.code)

			assert.Equal(t, tt.articles, articles(l.Basis))
			assert.Equal(t, tt.source, l.Source.String())
		})
	}
	assert.Equal(t, gpmRate(t, "0.1500"), lineByCode(t, est.Lines, "isr.coeficiente").Rate)
	assert.Equal(t, gpmRate(t, "0.30"), lineByCode(t, est.Lines, "isr.rate").Rate)
}

// TestGeneralPMStrategyPTUInEqualPartsMayToDecember: LISR 14 fr. II a)
// deducts the PTU paid in equal parts from May to December, cumulatively,
// capped at the profit of the payment.
func TestGeneralPMStrategyPTUInEqualPartsMayToDecember(t *testing.T) {
	tests := []struct {
		name    string
		month   time.Month
		income  string // YTD accrued, all in the period's month
		ptu     string
		wantPTU string
		wantISR string
	}{
		{"January: no part yet", time.January, "100000.00", "80000.00", "0.00", "30000.00"},
		{"April: no part yet", time.April, "100000.00", "80000.00", "0.00", "30000.00"},
		{"May: first eighth", time.May, "100000.00", "80000.00", "10000.00", "27000.00"},
		{"August: four eighths", time.August, "100000.00", "80000.00", "40000.00", "18000.00"},
		{"December: all eight", time.December, "100000.00", "80000.00", "80000.00", "6000.00"},
		{"capped at the profit", time.December, "100000.00", "150000.00", "100000.00", "0.00"},
		{"exact share kept to the centavo", time.May, "100000.00", "1000.01", "125.00", "29962.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amounts := make([]string, tt.month)
			for i := range amounts {
				amounts[i] = "0.00"
			}
			amounts[tt.month-1] = tt.income
			p, err := domain.NewPeriod(2026, tt.month)
			require.NoError(t, err)
			in := domain.Input{
				Key: gpmKey, Period: p,
				Sums:   gpmAccrual(t, tt.month, amounts...),
				Inputs: inputs(t, gpmCoeficiente(t, "1"), gpmMoney(t, domain.InputPTUPaid, tt.ptu)),
			}

			est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

			require.NoError(t, err)
			assert.Equal(t, tt.wantPTU, lineByCode(t, est.Lines, "isr.ptu").Amount.String())
			assert.Equal(t, tt.wantISR, est.ISR.Amount.String())
		})
	}
}

// TestGeneralPMStrategyLossesAppliedUpToTheProfit: a loss larger than the
// profit after PTU is applied only up to it; the rest stays pending for
// later payments, so the base never goes below zero.
func TestGeneralPMStrategyLossesAppliedUpToTheProfit(t *testing.T) {
	tests := []struct {
		name       string
		losses     string
		wantLosses string
		wantBase   string
	}{
		{"below the profit", "40000.00", "40000.00", "50000.00"},
		{"equal to the profit", "90000.00", "90000.00", "0.00"},
		{"above the profit", "500000.00", "90000.00", "0.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := domain.Input{
				Key: gpmKey, Period: period(t, "2026-05"),
				Sums: gpmAccrual(t, time.May, "0.00", "0.00", "0.00", "0.00", "100000.00"),
				Inputs: inputs(t, gpmCoeficiente(t, "1"),
					gpmMoney(t, domain.InputPTUPaid, "80000.00"),
					gpmMoney(t, domain.InputPriorYearLosses, tt.losses)),
			}

			est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

			require.NoError(t, err)
			assert.Equal(t, tt.wantLosses, lineByCode(t, est.Lines, "isr.prior_year_losses").Amount.String())
			assert.Equal(t, tt.wantBase, lineByCode(t, est.Lines, "isr.base").Amount.String())
		})
	}
}

func TestGeneralPMStrategyCreditsPaymentsAndInterestWithholding(t *testing.T) {
	values := append(gpmPriorPayments(t, "1000.00", time.January, time.February),
		gpmCoeficiente(t, "0.2000"),
		gpmMoney(t, domain.InputInterestWithholding, "250.50"))
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-03"),
		Sums:   gpmAccrual(t, time.March, "10000.00", "10000.00", "10000.00"),
		Inputs: inputs(t, values...),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	// 0.20 × 30,000 = 6,000 × 30 % = 1,800 − 2,000 − 250.50 = −450.50: in favor.
	assert.Equal(t, "-450.50", est.ISR.Amount.String())
	assert.Equal(t, domain.OutcomeInFavor, est.ISR.Outcome())
	assert.Equal(t, "2000.00", lineByCode(t, est.Lines, "isr.prior_payments").Amount.String())
	assert.Equal(t, "250.50", lineByCode(t, est.Lines, "isr.interest_withholding").Amount.String())
}

// TestGeneralPMStrategyKeepsTheCoeficienteAsGiven: art. 14 is silent on
// its decimals (⚠️ 4-decimal practice), so the engine never rounds it.
func TestGeneralPMStrategyKeepsTheCoeficienteAsGiven(t *testing.T) {
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-01"),
		Sums:   gpmAccrual(t, time.January, "1000000.00"),
		Inputs: inputs(t, gpmCoeficiente(t, "0.123456")),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	assert.Equal(t, "123456.00", lineByCode(t, est.Lines, "isr.coeficiente").Amount.String())
	assert.Equal(t, "37036.80", est.ISR.Amount.String())
}

func TestGeneralPMStrategyFirstFiscalYearHasNoProvisionalISR(t *testing.T) {
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-06"),
		Sums: gpmAccrual(t, time.June, "0.00", "0.00", "0.00", "0.00", "0.00", "500000.00"),
		Inputs: inputs(t, domain.TaxInputValue{
			Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: true,
		}),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	require.NotNil(t, est.ISR)
	assert.Equal(t, money.Cents(0), est.ISR.Amount)
	assert.Equal(t, []gpmLine{
		{domain.TaxISR, "isr.first_fiscal_year", domain.OpInfo, "0.00"},
	}, gpmLines(est.Lines, domain.TaxISR))
	first := lineByCode(t, est.Lines, "isr.first_fiscal_year")
	assert.Equal(t, []string{"14"}, articles(first.Basis))
	assert.Equal(t, "tax_input:first_fiscal_year", first.Source.String())
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNoCoefficientFirstYear}, est.Disclaimers)
	assert.NotNil(t, est.IVA, "the IVA is monthly from the first month")
}

func TestGeneralPMStrategyFirstFiscalYearFalseComputes(t *testing.T) {
	in := pm601ISR01(t)
	values := append(gpmPriorPayments(t, "36000.00", time.January, time.February, time.March, time.April, time.May),
		gpmCoeficiente(t, "0.1500"),
		gpmMoney(t, domain.InputPTUPaid, "120000.00"),
		domain.TaxInputValue{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: false})
	in.Inputs = inputs(t, values...)

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	assert.Equal(t, "36000.00", est.ISR.Amount.String())
	assert.Empty(t, est.Disclaimers)
}

func TestGeneralPMStrategyRequiredInputs(t *testing.T) {
	firstYear := domain.TaxInputValue{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: true}
	notFirstYear := domain.TaxInputValue{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean}
	coeficiente := domain.RequiredInput{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate}
	payments := func(months ...time.Month) domain.RequiredInput {
		return domain.RequiredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: months}
	}
	tests := []struct {
		name   string
		period string
		active []time.Month
		values []domain.TaxInputValue
		want   []domain.RequiredInput
	}{
		{
			"January: only the coeficiente", "2026-01", []time.Month{time.January}, nil,
			[]domain.RequiredInput{coeficiente},
		},
		{
			"June, active all year: payments January to May", "2026-06",
			[]time.Month{time.January, time.February, time.March, time.April, time.May, time.June}, nil,
			[]domain.RequiredInput{coeficiente, payments(time.January, time.February, time.March, time.April, time.May)},
		},
		{
			"June, active from March: payments March to May", "2026-06",
			[]time.Month{time.March, time.April, time.May, time.June}, nil,
			[]domain.RequiredInput{coeficiente, payments(time.March, time.April, time.May)},
		},
		{
			"first_fiscal_year false: as any year", "2026-02", []time.Month{time.January, time.February},
			[]domain.TaxInputValue{notFirstYear},
			[]domain.RequiredInput{coeficiente, payments(time.January)},
		},
		{
			"first fiscal year: nothing", "2026-06", []time.Month{time.June},
			[]domain.TaxInputValue{firstYear}, nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := domain.Facts{ActiveMonths: tt.active, Inputs: inputs(t, tt.values...)}

			got := domain.NewGeneralPMStrategy().RequiredInputs(domain.Options{}, period(t, tt.period), facts)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGeneralPMStrategyMissingCoeficienteIsIncompleteWithZeros: through
// the registry, as the use case runs it (ADR-0002 §2).
func TestGeneralPMStrategyMissingCoeficienteIsIncompleteWithZeros(t *testing.T) {
	reg, err := domain.NewRegistry(domain.NewGeneralPMStrategy())
	require.NoError(t, err)
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-01"),
		Sums: gpmAccrual(t, time.January, "1000000.00"),
	}

	est, err := reg.Estimate(in, []time.Month{time.January}, gpmParams(t, "0.30"))

	require.NoError(t, err)
	assert.Equal(t, domain.StatusIncomplete, est.Status())
	assert.Equal(t, []domain.MissingInput{{Key: domain.InputCoeficienteUtilidad}}, est.MissingInputs)
	assert.Equal(t, money.Cents(0), est.ISR.Amount)
	assert.Equal(t, "0.00", lineByCode(t, est.Lines, "isr.coeficiente").Amount.String())
}

// TestGeneralPMStrategyIVAAndWithholdingsToRemit is PM-601-IVA-01 plus the
// ISR side of TX-601-REMIT-01: the PM also withheld 10 % ISR (5,000.00)
// on the lawyer's 50,000 fees (LISR 106 ¶5), so it remits 10,333.33.
func TestGeneralPMStrategyIVAAndWithholdingsToRemit(t *testing.T) {
	june := domain.MonthSums{
		Income: domain.IncomeSums{
			Cash: cents(t, "2000000.00"), Accrual: cents(t, "2000000.00"),
			ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "2000000.00"), IVA: cents(t, "320000.00")}},
		},
		Expenses: domain.ExpenseSums{
			CreditableIVAPaid:  cents(t, "200000.00"),
			IVAWithheldToRemit: cents(t, "5333.33"),
			ISRWithheldToRemit: cents(t, "5000.00"),
		},
	}
	sums, err := domain.NewYearSums(time.June, domain.MonthEntry{Month: time.June, Sums: june})
	require.NoError(t, err)
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-06"), Sums: sums,
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(true)}),
		Inputs:  inputs(t, gpmCoeficiente(t, "0.1500")),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	require.NotNil(t, est.IVA)
	assert.Equal(t, domain.IVAComponents{
		ChargedCollected: cents(t, "320000.00"), CreditablePaid: cents(t, "200000.00"),
	}, *est.IVA)
	assert.Equal(t, "10333.33", est.WithheldToRemit.String())
	assert.Equal(t, []gpmLine{
		{domain.TaxIVA, domain.LineIVAChargedRate16, domain.OpInput, "320000.00"},
		{domain.TaxIVA, domain.LineIVACharged, domain.OpResult, "320000.00"},
		{domain.TaxIVA, domain.LineIVACreditable, domain.OpSubtract, "200000.00"},
		{domain.TaxIVA, domain.LineIVAWithheldByMoral, domain.OpSubtract, "0.00"},
		{domain.TaxIVA, domain.LineIVANet, domain.OpResult, "120000.00"},
	}, gpmLines(est.Lines, domain.TaxIVA))
	assert.Equal(t, []gpmLine{
		{domain.TaxWithholding, domain.LineIVAWithheldToRemit, domain.OpInput, "5333.33"},
		{domain.TaxWithholding, "withholding.isr_to_remit", domain.OpAdd, "5000.00"},
		{domain.TaxWithholding, "withholding.to_remit", domain.OpResult, "10333.33"},
	}, gpmLines(est.Lines, domain.TaxWithholding))
	isrToRemit := lineByCode(t, est.Lines, "withholding.isr_to_remit")
	assert.Equal(t, []string{"106", "116", "113-J"}, articles(isrToRemit.Basis))
	assertCited(t, est.Lines)
}

// TestGeneralPMStrategyWithholdingsToRemitAreOneGroup: the ISR and IVA the
// PM withheld from its suppliers are one ledger segment, apart from its own
// ISR and IVA, and WithheldToRemit is that segment's last line: the amount
// is reported once and the breakdown shows it (ADR-0002 §1 totals rule).
func TestGeneralPMStrategyWithholdingsToRemitAreOneGroup(t *testing.T) {
	tests := []struct {
		name      string
		iva, isr  string
		want      string
		wantLines []gpmLine
	}{
		{
			name: "none: no withholding line", iva: "0.00", isr: "0.00", want: "0.00",
		},
		{
			name: "IVA only: its one line is the total", iva: "5333.33", isr: "0.00", want: "5333.33",
			wantLines: []gpmLine{{domain.TaxWithholding, domain.LineIVAWithheldToRemit, domain.OpInput, "5333.33"}},
		},
		{
			name: "ISR only: starts its own segment, never the IVA's", iva: "0.00", isr: "5000.00", want: "5000.00",
			wantLines: []gpmLine{{domain.TaxWithholding, "withholding.isr_to_remit", domain.OpInput, "5000.00"}},
		},
		{
			name: "both: IVA + ISR closed by one total", iva: "5333.33", isr: "5000.00", want: "10333.33",
			wantLines: []gpmLine{
				{domain.TaxWithholding, domain.LineIVAWithheldToRemit, domain.OpInput, "5333.33"},
				{domain.TaxWithholding, "withholding.isr_to_remit", domain.OpAdd, "5000.00"},
				{domain.TaxWithholding, "withholding.to_remit", domain.OpResult, "10333.33"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			june := domain.MonthSums{
				Income: domain.IncomeSums{
					ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "1000.00"), IVA: cents(t, "160.00")}},
				},
				Expenses: domain.ExpenseSums{
					IVAWithheldToRemit: cents(t, tt.iva),
					ISRWithheldToRemit: cents(t, tt.isr),
				},
			}
			sums, err := domain.NewYearSums(time.June, domain.MonthEntry{Month: time.June, Sums: june})
			require.NoError(t, err)
			in := domain.Input{
				Key: gpmKey, Period: period(t, "2026-06"), Sums: sums,
				Inputs: inputs(t, gpmCoeficiente(t, "0.1500")),
			}

			est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

			require.NoError(t, err)
			assert.Equal(t, tt.want, est.WithheldToRemit.String())
			assert.Equal(t, tt.wantLines, gpmLines(est.Lines, domain.TaxWithholding))
			assert.Equal(t, "160.00", lineByCode(t, est.Lines, domain.LineIVANet).Amount.String(),
				"the own IVA never absorbs a withholding to remit")
			assertCited(t, est.Lines)
		})
	}
}

// TestGeneralPMStrategyIVAIsTheMonthOnly: the ISR accumulates January..n,
// the IVA never does (LIVA 5-D).
func TestGeneralPMStrategyIVAIsTheMonthOnly(t *testing.T) {
	month := func(iva string) domain.MonthSums {
		return domain.MonthSums{Income: domain.IncomeSums{
			ByRate: domain.IVAByRate{Rate16: domain.RateSums{IVA: cents(t, iva)}},
		}}
	}
	sums, err := domain.NewYearSums(time.February,
		domain.MonthEntry{Month: time.January, Sums: month("999.00")},
		domain.MonthEntry{Month: time.February, Sums: month("160.00")})
	require.NoError(t, err)
	in := domain.Input{Key: gpmKey, Period: period(t, "2026-02"), Sums: sums}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	assert.Equal(t, "160.00", est.IVA.ChargedCollected.String())
	assert.Equal(t, money.Cents(0), est.WithheldToRemit)
}

func TestGeneralPMStrategyAddsTheMixedActivityDisclaimer(t *testing.T) {
	sums, err := domain.NewYearSums(time.March, domain.MonthEntry{Month: time.March, Sums: domain.MonthSums{
		Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "100.00")},
	}})
	require.NoError(t, err)
	in := domain.Input{
		Key: gpmKey, Period: period(t, "2026-03"), Sums: sums,
		Options: domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionIVAFullyTaxable: domain.BoolOption(false)}),
		Inputs:  inputs(t, domain.TaxInputValue{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: true}),
	}

	est, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.DisclaimerCode{
		domain.DisclaimerNoCoefficientFirstYear, domain.DisclaimerIVAMixedActivity,
	}, est.Disclaimers)
}

func TestGeneralPMStrategyFailures(t *testing.T) {
	t.Run("no persona moral rate in the parameters", func(t *testing.T) {
		empty, err := taxparams.NewParams(taxparams.Spec{
			FiscalYear: 2026, Version: 1, PublishedOn: period(t, "2026-01").End(),
			Sources: []taxparams.Source{{
				ID: "x", Title: "x", URL: "https://example.org", DOFDate: period(t, "2026-01").End(), Reference: "x",
			}},
		})
		require.NoError(t, err)

		_, err = domain.NewGeneralPMStrategy().Estimate(pm601ISR01(t), empty)

		require.ErrorIs(t, err, taxparams.ErrUnknownKey)
	})
	t.Run("rate not in force on the period's last day", func(t *testing.T) {
		in := pm601ISR01(t)
		in.Period = period(t, "2025-12")
		in.Sums = gpmAccrual(t, time.December)

		_, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

		require.ErrorIs(t, err, taxparams.ErrNotInForce)
	})
	t.Run("sums short of the period", func(t *testing.T) {
		in := pm601ISR01(t)
		in.Period = period(t, "2026-07")

		_, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

		require.ErrorIs(t, err, domain.ErrInvalidMonth)
	})
	t.Run("overflow", func(t *testing.T) {
		in := pm601ISR01(t)
		in.Inputs = inputs(t, gpmMoney(t, domain.InputPTUPaid, "90000000000000000.00"))
		in.Period = period(t, "2026-06")

		_, err := domain.NewGeneralPMStrategy().Estimate(in, gpmParams(t, "0.30"))

		require.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestGeneralPMStrategyIsDeterministic(t *testing.T) {
	s := domain.NewGeneralPMStrategy()
	p := gpmParams(t, "0.30")
	first, err := s.Estimate(pm601ISR01(t), p)
	require.NoError(t, err)

	for range 20 {
		again, err := s.Estimate(pm601ISR01(t), p)
		require.NoError(t, err)
		assert.Equal(t, first, again)
	}
}
