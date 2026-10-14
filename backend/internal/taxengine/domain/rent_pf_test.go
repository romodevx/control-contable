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
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

var rentKey = domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}

// rentParams is the embedded 2026 parameter set: the strategy is tested
// against the published art. 116 table, not a synthetic one.
func rentParams(t *testing.T) *taxparams.Params {
	t.Helper()
	c, err := params.Embedded()
	require.NoError(t, err)
	sel, err := c.Select(2026)
	require.NoError(t, err)
	return sel.Params
}

// rentParamsWithoutTariff is a valid set holding the blind deduction rate
// and no art. 116 table.
func rentParamsWithoutTariff(t *testing.T) *taxparams.Params {
	t.Helper()
	day := func(s string) fiscal.Date {
		d, err := fiscal.ParseDate(s)
		require.NoError(t, err)
		return d
	}
	basis := []taxparams.Citation{{Source: "lisr", Article: "Art. 115"}}
	p, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: day("2026-01-01"),
		Sources: []taxparams.Source{{
			ID: "lisr", Title: "LISR", URL: "https://example.test/lisr", DOFDate: day("2025-11-07"), Reference: "Art. 115",
		}},
		Values: []taxparams.Value{{
			Key: "isr.rent.blind_deduction", Kind: taxparams.KindRate, Rate: 350_000,
			Validity: taxparams.Validity{From: day("2026-01-01")}, Basis: basis,
		}},
	})
	require.NoError(t, err)
	return p
}

func rentOptions(mode string) domain.Options {
	if mode == "" {
		return domain.Options{}
	}
	return domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionRentDeductionMode: domain.TextOption(mode),
	})
}

// rentInput is March 2026 with month's sums; February's sums are noise the
// monthly (not year-to-date) computation must ignore.
func rentInput(t *testing.T, mode string, month domain.MonthSums) domain.Input {
	t.Helper()
	noise := domain.MonthSums{
		Income:   domain.IncomeSums{Cash: cents(t, "99999.00"), ISRWithheld: domain.Withheld{ByMoral: cents(t, "9999.90")}},
		Expenses: domain.ExpenseSums{DeductiblePaid: cents(t, "5000.00"), PredialPaid: cents(t, "800.00")},
	}
	sums, err := domain.NewYearSums(time.March,
		domain.MonthEntry{Month: time.February, Sums: noise},
		domain.MonthEntry{Month: time.March, Sums: month},
	)
	require.NoError(t, err)
	return domain.Input{Key: rentKey, Period: period(t, "2026-03"), Options: rentOptions(mode), Sums: sums}
}

// rentLine is a line as the tests assert it.
type rentLine struct {
	code   domain.LineCode
	op     domain.Operation
	amount string
}

func rentLines(lines []domain.Line, tax domain.TaxKind) []rentLine {
	var out []rentLine
	for _, l := range lines {
		if l.Tax == tax {
			out = append(out, rentLine{code: l.Code, op: l.Operation, amount: l.Amount.String()})
		}
	}
	return out
}

func rentEstimate(t *testing.T, in domain.Input) domain.RegimeEstimate {
	t.Helper()
	est, err := domain.NewRentPFStrategy().Estimate(in, rentParams(t))
	require.NoError(t, err)
	return est
}

func TestRentPFStrategyDeclaration(t *testing.T) {
	s := domain.NewRentPFStrategy()

	assert.Equal(t, rentKey, s.Key())
	assert.Equal(t, domain.TreatmentFull, s.Treatment())
	assert.Empty(t, s.RequiredInputs(rentOptions("actual"), period(t, "2026-03"), domain.Facts{}),
		"606 needs no tax input: rent, deductions, predial and withholdings are records (ADR-0002 §2)")
	assert.NoError(t, rentParams(t).Require(s.ParamRequirements()...))
}

func TestRentPFStrategyISR(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		month       domain.MonthSums
		wantISR     string
		wantOutcome domain.Outcome
		wantLines   []rentLine
		disclaimers []domain.DisclaimerCode
	}{
		{
			name:    "PF-606-ISR-01 blind deduction, tenant is a PF",
			mode:    "blind_35",
			month:   domain.MonthSums{Income: domain.IncomeSums{Cash: 2_000_000}},
			wantISR: "1076.00", wantOutcome: domain.OutcomePayable,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "20000.00"},
				{"isr.blind_deduction", domain.OpSubtract, "7000.00"},
				{"isr.predial", domain.OpSubtract, "0.00"},
				{"isr.base", domain.OpResult, "13000.00"},
				{"isr.tariff_row", domain.OpLookup, "1011.68"},
				{"isr.tariff_excess", domain.OpMultiply, "64.32"},
				{"isr.caused", domain.OpResult, "1076.00"},
				{"isr.withheld_by_moral", domain.OpSubtract, "0.00"},
				{"isr.payable", domain.OpResult, "1076.00"},
			},
		},
		{
			name: "PF-606-ISR-02 the PM tenant's 10 % leaves a saldo a favor",
			mode: "blind_35",
			month: domain.MonthSums{Income: domain.IncomeSums{
				Cash: 2_000_000, ISRWithheld: domain.Withheld{ByMoral: 200_000},
			}},
			wantISR: "-924.00", wantOutcome: domain.OutcomeInFavor,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "20000.00"},
				{"isr.blind_deduction", domain.OpSubtract, "7000.00"},
				{"isr.predial", domain.OpSubtract, "0.00"},
				{"isr.base", domain.OpResult, "13000.00"},
				{"isr.tariff_row", domain.OpLookup, "1011.68"},
				{"isr.tariff_excess", domain.OpMultiply, "64.32"},
				{"isr.caused", domain.OpResult, "1076.00"},
				{"isr.withheld_by_moral", domain.OpSubtract, "2000.00"},
				{"isr.payable", domain.OpResult, "-924.00"},
			},
		},
		{
			// 20,000 − 7,000 − 1,000 = 12,000 → row 3: 420.95 + (12,000 − 7,168.52) × 10.88 %
			// = 420.95 + 525.665024 = 946.615024; the deduction of the predial is claimed.
			name: "blind deduction plus the predial paid in the month",
			mode: "blind_35",
			month: domain.MonthSums{
				Income:   domain.IncomeSums{Cash: 2_000_000},
				Expenses: domain.ExpenseSums{PredialPaid: 100_000, DeductiblePaid: 300_000},
			},
			wantISR: "946.62", wantOutcome: domain.OutcomePayable,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "20000.00"},
				{"isr.blind_deduction", domain.OpSubtract, "7000.00"},
				{"isr.predial", domain.OpSubtract, "1000.00"},
				{"isr.base", domain.OpResult, "12000.00"},
				{"isr.tariff_row", domain.OpLookup, "420.95"},
				{"isr.tariff_excess", domain.OpMultiply, "525.67"},
				{"isr.caused", domain.OpResult, "946.62"},
				{"isr.withheld_by_moral", domain.OpSubtract, "0.00"},
				{"isr.payable", domain.OpResult, "946.62"},
			},
			disclaimers: []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed},
		},
		{
			// 20,000 − 3,000 = 17,000 → row 5: 1,339.14 + (17,000 − 14,644.65) × 17.92 %
			// = 1,339.14 + 422.07872 = 1,761.21872.
			name: "actual deductions replace the blind deduction and the predial",
			mode: "actual",
			month: domain.MonthSums{
				Income:   domain.IncomeSums{Cash: 2_000_000},
				Expenses: domain.ExpenseSums{DeductiblePaid: 300_000, PredialPaid: 100_000},
			},
			wantISR: "1761.22", wantOutcome: domain.OutcomePayable,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "20000.00"},
				{"isr.actual_deductions", domain.OpSubtract, "3000.00"},
				{"isr.base", domain.OpResult, "17000.00"},
				{"isr.tariff_row", domain.OpLookup, "1339.14"},
				{"isr.tariff_excess", domain.OpMultiply, "422.08"},
				{"isr.caused", domain.OpResult, "1761.22"},
				{"isr.withheld_by_moral", domain.OpSubtract, "0.00"},
				{"isr.payable", domain.OpResult, "1761.22"},
			},
			disclaimers: []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed},
		},
		{
			name: "deductions above the rent leave no base: no tax, the withholding is in favor",
			mode: "actual",
			month: domain.MonthSums{
				Income:   domain.IncomeSums{Cash: 1_000_000, ISRWithheld: domain.Withheld{ByMoral: 100_000}},
				Expenses: domain.ExpenseSums{DeductiblePaid: 1_500_000},
			},
			wantISR: "-1000.00", wantOutcome: domain.OutcomeInFavor,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "10000.00"},
				{"isr.actual_deductions", domain.OpSubtract, "15000.00"},
				{"isr.base", domain.OpResult, "-5000.00"},
				{"isr.no_base", domain.OpInput, "0.00"},
				{"isr.caused", domain.OpResult, "0.00"},
				{"isr.withheld_by_moral", domain.OpSubtract, "1000.00"},
				{"isr.payable", domain.OpResult, "-1000.00"},
			},
			disclaimers: []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed},
		},
		{
			name:    "no rent in the month: zero",
			mode:    "blind_35",
			wantISR: "0.00", wantOutcome: domain.OutcomeZero,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "0.00"},
				{"isr.blind_deduction", domain.OpSubtract, "0.00"},
				{"isr.predial", domain.OpSubtract, "0.00"},
				{"isr.base", domain.OpResult, "0.00"},
				{"isr.no_base", domain.OpInput, "0.00"},
				{"isr.caused", domain.OpResult, "0.00"},
				{"isr.withheld_by_moral", domain.OpSubtract, "0.00"},
				{"isr.payable", domain.OpResult, "0.00"},
			},
		},
		{
			name:    "no deduction_mode: the catalog default, blind_35",
			mode:    "",
			month:   domain.MonthSums{Income: domain.IncomeSums{Cash: 2_000_000}},
			wantISR: "1076.00", wantOutcome: domain.OutcomePayable,
			wantLines: []rentLine{
				{"isr.rent_month", domain.OpInput, "20000.00"},
				{"isr.blind_deduction", domain.OpSubtract, "7000.00"},
				{"isr.predial", domain.OpSubtract, "0.00"},
				{"isr.base", domain.OpResult, "13000.00"},
				{"isr.tariff_row", domain.OpLookup, "1011.68"},
				{"isr.tariff_excess", domain.OpMultiply, "64.32"},
				{"isr.caused", domain.OpResult, "1076.00"},
				{"isr.withheld_by_moral", domain.OpSubtract, "0.00"},
				{"isr.payable", domain.OpResult, "1076.00"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			est := rentEstimate(t, rentInput(t, tt.mode, tt.month))

			require.NotNil(t, est.ISR)
			assert.Equal(t, tt.wantISR, est.ISR.Amount.String())
			assert.Equal(t, tt.wantOutcome, est.ISR.Outcome())
			assert.Equal(t, tt.wantLines, rentLines(est.Lines, domain.TaxISR))
			assert.ElementsMatch(t, tt.disclaimers, est.Disclaimers)
			assert.Equal(t, rentKey, est.Key)
			assert.Equal(t, domain.TreatmentFull, est.Coverage)
			assert.Equal(t, domain.StatusComplete, est.Status())
			assert.Empty(t, est.CPATopics)
			assertCited(t, est.Lines)
		})
	}
}

func TestRentPFStrategyLinesCiteAndSource(t *testing.T) {
	in := rentInput(t, "blind_35", domain.MonthSums{Income: domain.IncomeSums{Cash: 2_000_000}})

	lines := rentEstimate(t, in).Lines

	tests := []struct {
		code     domain.LineCode
		source   string
		articles []string
	}{
		{"isr.rent_month", "records", []string{"114", "116"}},
		{"isr.blind_deduction", "params:isr.rent.blind_deduction", []string{"115"}},
		{"isr.predial", "records", []string{"115"}},
		{"isr.base", "computed", []string{"116"}},
		{"isr.tariff_row", "params:isr.art116.monthly", []string{"116", "Tarifa art. 116"}},
		{"isr.tariff_excess", "params:isr.art116.monthly", []string{"116", "Tarifa art. 116"}},
		{"isr.caused", "computed", []string{"116"}},
		{"isr.withheld_by_moral", "records", []string{"116"}},
		{"isr.payable", "computed", []string{"116"}},
	}
	for _, tt := range tests {
		l := lineByCode(t, lines, tt.code)
		assert.Equal(t, tt.source, l.Source.String(), tt.code)
		assert.Equal(t, tt.articles, articles(l.Basis), tt.code)
	}
	blind := lineByCode(t, lines, "isr.blind_deduction")
	assert.Equal(t, []string{"0.35"}, blind.Args, "the blind deduction line names its rate")
	excess := lineByCode(t, lines, "isr.tariff_excess")
	assert.Equal(t, money.Rate(160_000), excess.Rate)
	assert.Equal(t, 4, lineByCode(t, lines, "isr.tariff_row").Row)
}

func TestRentPFStrategyIVA(t *testing.T) {
	tests := []struct {
		name      string
		month     domain.MonthSums
		want      domain.IVAComponents
		wantLines []rentLine
	}{
		{
			name: "PF-606-IVA-01 commercial rent to a PM",
			month: domain.MonthSums{Income: domain.IncomeSums{
				Cash:        2_000_000,
				ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: 2_000_000, IVA: 320_000}},
				IVAWithheld: domain.Withheld{ByMoral: 213_333},
				ISRWithheld: domain.Withheld{ByMoral: 200_000},
			}},
			want: domain.IVAComponents{ChargedCollected: 320_000, WithheldFromTaxpayer: 213_333},
			wantLines: []rentLine{
				{"iva.charged.rate_16", domain.OpInput, "3200.00"},
				{"iva.charged", domain.OpResult, "3200.00"},
				{"iva.creditable", domain.OpSubtract, "0.00"},
				{"iva.withheld.moral", domain.OpSubtract, "2133.33"},
				{"iva.regime_net", domain.OpResult, "1066.67"},
			},
		},
		{
			name: "PF-606-IVA-01 residential rent is exempt",
			month: domain.MonthSums{Income: domain.IncomeSums{
				Cash: 2_000_000, ByRate: domain.IVAByRate{ExemptBase: 2_000_000},
			}},
			want: domain.IVAComponents{},
			wantLines: []rentLine{
				{"iva.charged.rate_16", domain.OpInput, "0.00"},
				{"iva.base.exempt", domain.OpInfo, "20000.00"},
				{"iva.charged", domain.OpResult, "0.00"},
				{"iva.creditable", domain.OpSubtract, "0.00"},
				{"iva.withheld.moral", domain.OpSubtract, "0.00"},
				{"iva.regime_net", domain.OpResult, "0.00"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			est := rentEstimate(t, rentInput(t, "blind_35", tt.month))

			require.NotNil(t, est.IVA)
			assert.Equal(t, tt.want, *est.IVA)
			assert.Equal(t, tt.wantLines, rentLines(est.Lines, domain.TaxIVA))
			assert.Empty(t, est.Disclaimers)
		})
	}
}

func TestRentPFStrategyIVAMixedActivityAndRemit(t *testing.T) {
	month := domain.MonthSums{
		Income: domain.IncomeSums{
			Cash: 3_000_000,
			ByRate: domain.IVAByRate{
				Rate16: domain.RateSums{Base: 2_000_000, IVA: 320_000}, ExemptBase: 1_000_000,
			},
		},
		Expenses: domain.ExpenseSums{CreditableIVAPaid: 48_000, IVAWithheldToRemit: 10_000},
	}

	est := rentEstimate(t, rentInput(t, "actual", month))

	assert.ElementsMatch(t, []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}, est.Disclaimers,
		"commercial and residential rent in one month: creditable IVA taken in full, flagged (§2.6)")
	assert.Equal(t, money.Cents(10_000), est.WithheldToRemit)
	assert.Equal(t, []rentLine{{"withholding.iva_to_remit", domain.OpInput, "100.00"}},
		rentLines(est.Lines, domain.TaxWithholding))
}

func TestRentPFStrategyQuarterlyIsInformational(t *testing.T) {
	in := rentInput(t, "blind_35", domain.MonthSums{Income: domain.IncomeSums{Cash: 2_000_000}})
	in.Options = domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionRentDeductionMode:    domain.TextOption("blind_35"),
		domain.OptionRentQuarterlyPayment: domain.BoolOption(true),
	})

	est := rentEstimate(t, in)

	assert.Equal(t, domain.StatusInformational, est.Status())
	assert.Equal(t, domain.RegimeExplanation(rentKey), est.Explanation)
	assert.Nil(t, est.ISR)
	assert.Nil(t, est.IVA)
	assert.Empty(t, est.Lines)
}

func TestRentPFStrategyFailures(t *testing.T) {
	good := domain.MonthSums{Income: domain.IncomeSums{Cash: 2_000_000}}
	t.Run("unknown deduction_mode", func(t *testing.T) {
		_, err := domain.NewRentPFStrategy().Estimate(rentInput(t, "blind_50", good), rentParams(t))

		assert.ErrorIs(t, err, domain.ErrRentDeductionMode)
	})
	t.Run("parameters without the art. 116 table", func(t *testing.T) {
		p := rentParamsWithoutTariff(t)

		_, err := domain.NewRentPFStrategy().Estimate(rentInput(t, "blind_35", good), p)

		assert.ErrorIs(t, err, taxparams.ErrUnknownKey)
	})
	t.Run("an overflowing rent", func(t *testing.T) {
		huge := domain.MonthSums{Income: domain.IncomeSums{Cash: money.Cents(1) << 62}}

		_, err := domain.NewRentPFStrategy().Estimate(rentInput(t, "blind_35", huge), rentParams(t))

		assert.ErrorIs(t, err, money.ErrOverflow)
	})
}

func TestRentPFStrategyIsDeterministic(t *testing.T) {
	in := rentInput(t, "blind_35", domain.MonthSums{Income: domain.IncomeSums{
		Cash: 2_000_000, ISRWithheld: domain.Withheld{ByMoral: 200_000},
		ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 2_000_000, IVA: 320_000}},
	}})

	first := rentEstimate(t, in)
	for range 5 {
		assert.Equal(t, first, rentEstimate(t, in))
	}
}
