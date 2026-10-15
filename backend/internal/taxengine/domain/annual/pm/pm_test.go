package pm_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

const fiscalYear = 2026

var errParams = errors.New("params unavailable")

// fakeParams is the Params port: the 30 % rate of LISR 9, read on the date
// the strategy asks for.
type fakeParams struct {
	rates map[taxparams.Key]money.Rate
	err   error
	asked []string
}

func (f *fakeParams) Rate(key taxparams.Key, on fiscal.Date) (money.Rate, error) {
	f.asked = append(f.asked, string(key)+"@"+on.String())
	if f.err != nil {
		return 0, f.err
	}
	r, ok := f.rates[key]
	if !ok {
		return 0, fmt.Errorf("no rate %s", key)
	}
	return r, nil
}

func params(t testing.TB) *fakeParams {
	t.Helper()
	return &fakeParams{rates: map[taxparams.Key]money.Rate{"isr.rate.persona_moral": rate(t, "0.30")}}
}

func cents(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func rate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

func filed(t testing.TB, amount string) annual.ResolvedPayments {
	t.Helper()
	return annual.ResolvedPayments{Payments: []annual.Payment{{Amount: cents(t, amount), Source: annual.PaymentFiled}}}
}

// trace renders lines as "code op amount", the shape the goldens compare.
func trace(lines []domain.Line) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = fmt.Sprintf("%s %s %s", l.Code, l.Operation, l.Amount)
	}
	return out
}

func an06(t testing.TB) pm.Input {
	t.Helper()
	return pm.Input{
		FiscalYear: fiscalYear, Income: cents(t, "10000000.00"), Deductions: cents(t, "8500000.00"),
		PTUPaid: cents(t, "120000.00"), PriorYearLosses: cents(t, "200000.00"),
		InterestWithholding: cents(t, "2000.00"), Payments: filed(t, "330000.00"),
	}
}

// TestGeneralComputesAN06 is research example AN-06 (MODULE-annual §3):
// utilidad fiscal 1,380,000; resultado fiscal 1,180,000 × 30 % = 354,000;
// − 330,000 − 2,000 = a cargo 22,000. PTU base 1,500,000 → 150,000 (⚠️
// ptu_rate); suggested CU 2027 0.1380.
func TestGeneralComputesAN06(t *testing.T) {
	p := params(t)

	got, err := pm.NewGeneral().Compute(an06(t), p)

	require.NoError(t, err)
	assert.Equal(t, annual.SectionPM601, got.Kind)
	assert.Equal(t, fiscal.RegimeCode("601"), got.Regime)
	assert.Equal(t, annual.StatusComplete, got.Status)
	assert.Equal(t, annual.ObligationObligated, got.Obligation)
	assert.Empty(t, got.MissingInputs)
	assert.ElementsMatch(t, []annual.CPATopic{annual.TopicCURounding, annual.TopicPTURate}, got.CPATopics)
	assert.Equal(t, []annual.DisclaimerCode{annual.DisclaimerInflationAdjustmentExcluded}, got.Disclaimers)
	assert.Equal(t, pm.Tax{
		AccumulableIncome: cents(t, "10000000.00"), Base: cents(t, "1180000.00"), ISR: cents(t, "354000.00"),
		Payments: cents(t, "330000.00"), Withholdings: cents(t, "2000.00"),
		Balance: pm.Balance{Kind: annual.BalanceACargo, Amount: cents(t, "22000.00")},
	}, got.Tax)
	assert.Equal(t, pm.Extras{
		PTUBase: cents(t, "1500000.00"), PTUToDistribute: cents(t, "150000.00"), SuggestedCU: rate(t, "0.1380"),
	}, got.Extras)
	assert.Equal(t, []string{
		"isr.income input 10000000.00",
		"isr.deductions subtract 8500000.00",
		"isr.profit_before_ptu result 1500000.00",
		"isr.ptu_paid subtract 120000.00",
		"isr.taxable_profit result 1380000.00",
		"isr.prior_year_losses subtract 200000.00",
		"isr.base result 1180000.00",
		"isr.rate multiply 354000.00",
		"isr.annual_tax result 354000.00",
		"isr.payments_filed subtract 330000.00",
		"isr.interest_withholding subtract 2000.00",
		"isr.balance result 22000.00",
		"ptu.base input 1500000.00",
		"ptu.rate multiply 150000.00",
		"ptu.to_distribute result 150000.00",
	}, trace(got.Lines))
	assert.Equal(t, []string{"isr.rate.persona_moral@2026-12-31"}, p.asked, "the rate in force at the year's end")
}

// TestEveryComputedLineIsCited: the breakdown is the computation trace and
// every non-info line carries its legal basis (MODULE-annual §2.7); the
// provisional PTU rate and the CPA-pending PTU base are flagged.
func TestEveryComputedLineIsCited(t *testing.T) {
	in := an06(t)
	in.Depreciation = cents(t, "1.00")
	for _, s := range []interface {
		Compute(pm.Input, pm.Params) (pm.Section, error)
	}{pm.NewGeneral(), pm.NewRESICO()} {
		got, err := s.Compute(in, params(t))

		require.NoError(t, err)
		for _, l := range got.Lines {
			assert.NotEmpty(t, l.Basis, "%s %s", got.Kind, l.Code)
			assert.Equal(t, domain.TaxISR, l.Tax)
			if l.Code == pm.LinePTURate || (l.Code == pm.LinePTUBase && got.Kind == annual.SectionPMRESICO) {
				assert.True(t, l.NeedsCPAReview, "%s %s", got.Kind, l.Code)
			}
		}
	}
}

// TestBalanceIsTheShownISRLessTheCredits is the ANN-16 regression
// (MODULE-annual §2.7: balance = ISR − payments − interest withholding):
// a base of 100,000.05 × 30 % is 30,000.015, shown 30,000.02. Paying
// exactly that leaves a cero balance; it was a favor of 0.01 because the
// credits were subtracted from the unrounded ISR.
func TestBalanceIsTheShownISRLessTheCredits(t *testing.T) {
	in := pm.Input{FiscalYear: fiscalYear, Income: cents(t, "100000.05"), Payments: filed(t, "30000.02")}
	for _, s := range []interface {
		Compute(pm.Input, pm.Params) (pm.Section, error)
	}{pm.NewGeneral(), pm.NewRESICO()} {
		got, err := s.Compute(in, params(t))

		require.NoError(t, err)
		assert.Equal(t, cents(t, "30000.02"), got.Tax.ISR, "%s", got.Kind)
		assert.Equal(t, pm.Balance{Kind: annual.BalanceCero}, got.Tax.Balance, "%s", got.Kind)
		assert.Contains(t, trace(got.Lines), "isr.balance result 0.00", "%s", got.Kind)
		assert.NotContains(t, trace(got.Lines), "rounding_adjustment rounding_adjustment -0.01", "%s", got.Kind)
	}
}

// TestGeneralSubtractsTypedDepreciation: the year's depreciation from the
// CPA's schedule is an authorized deduction (annual_depreciation, typed).
func TestGeneralSubtractsTypedDepreciation(t *testing.T) {
	in := an06(t)
	in.Depreciation = cents(t, "100000.00")

	got, err := pm.NewGeneral().Compute(in, params(t))

	require.NoError(t, err)
	assert.Contains(t, trace(got.Lines), "isr.depreciation subtract 100000.00")
	assert.Equal(t, "tax_input:annual_depreciation", got.Lines[2].Source.String())
	assert.Equal(t, cents(t, "1080000.00"), got.Tax.Base)
	assert.Equal(t, cents(t, "1400000.00"), got.Extras.PTUBase, "PTU base is after depreciation")
	assert.Equal(t, rate(t, "0.128"), got.Extras.SuggestedCU, "1,280,000 / 10,000,000")
}

// TestGeneralLossYear: a negative utilidad fiscal is a loss for the year: no
// prior losses applied, base and ISR 0, no PTU and no suggested CU; the
// payments come back as a balance in favour.
func TestGeneralLossYear(t *testing.T) {
	in := pm.Input{
		FiscalYear: fiscalYear, Income: cents(t, "1000000.00"), Deductions: cents(t, "1100000.00"),
		PriorYearLosses: cents(t, "50000.00"), Payments: filed(t, "4000.00"),
	}

	got, err := pm.NewGeneral().Compute(in, params(t))

	require.NoError(t, err)
	assert.Equal(t, []string{
		"isr.income input 1000000.00",
		"isr.deductions subtract 1100000.00",
		"isr.profit_before_ptu result -100000.00",
		"isr.taxable_profit result -100000.00",
		"isr.no_profit input 0.00",
		"isr.base result 0.00",
		"isr.rate multiply 0.00",
		"isr.annual_tax result 0.00",
		"isr.payments_filed subtract 4000.00",
		"isr.balance result -4000.00",
		"ptu.base input 0.00",
		"ptu.rate multiply 0.00",
		"ptu.to_distribute result 0.00",
	}, trace(got.Lines))
	assert.Equal(t, pm.Balance{Kind: annual.BalanceAFavor, Amount: cents(t, "4000.00")}, got.Tax.Balance)
	assert.Equal(t, pm.Extras{}, got.Extras)
}

// TestGeneralLossesApplyUpToTheProfit: pending losses larger than the
// utilidad fiscal leave a zero resultado fiscal (LISR 9 fr. II, 57).
func TestGeneralLossesApplyUpToTheProfit(t *testing.T) {
	in := pm.Input{
		FiscalYear: fiscalYear, Income: cents(t, "500000.00"), Deductions: cents(t, "400000.00"),
		PriorYearLosses: cents(t, "250000.00"),
	}

	got, err := pm.NewGeneral().Compute(in, params(t))

	require.NoError(t, err)
	assert.Contains(t, trace(got.Lines), "isr.prior_year_losses subtract 100000.00")
	assert.Equal(t, money.Cents(0), got.Tax.Base)
	assert.Equal(t, pm.Balance{Kind: annual.BalanceCero}, got.Tax.Balance)
	assert.Equal(t, rate(t, "0.2"), got.Extras.SuggestedCU, "the CU is before losses")
}

// TestSuggestedCURoundsHalfUpToFourDecimals (⚠️ cu_rounding).
func TestSuggestedCURoundsHalfUpToFourDecimals(t *testing.T) {
	tests := []struct{ profit, want string }{
		{profit: "1234567.00", want: "0.1235"}, // 0.1234567
		{profit: "1234500.00", want: "0.1235"}, // exactly half
		{profit: "1234499.99", want: "0.1234"},
		{profit: "0.01", want: "0"},
		{profit: "10000000.00", want: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.profit, func(t *testing.T) {
			in := pm.Input{FiscalYear: fiscalYear, Income: cents(t, "10000000.00")}
			in.Deductions = cents(t, "10000000.00") - cents(t, tt.profit)

			got, err := pm.NewGeneral().Compute(in, params(t))

			require.NoError(t, err)
			assert.Equal(t, rate(t, tt.want), got.Extras.SuggestedCU)
		})
	}
}

func TestNoIncomeHasNoSuggestedCU(t *testing.T) {
	got, err := pm.NewGeneral().Compute(pm.Input{FiscalYear: fiscalYear}, params(t))

	require.NoError(t, err)
	assert.Equal(t, money.Rate(0), got.Extras.SuggestedCU)
	assert.Equal(t, pm.Balance{Kind: annual.BalanceCero}, got.Tax.Balance)
}

// TestPaymentsBySourceAndEstimates: one credit line per source; months the
// app estimated raise payments_estimated, and a refund that depends on them
// carries the "solo se devuelve lo efectivamente pagado" note (LISR 152);
// a month the app could not verify makes the section incomplete.
func TestPaymentsBySourceAndEstimates(t *testing.T) {
	in := pm.Input{
		FiscalYear: fiscalYear, Income: cents(t, "10000.00"),
		Payments: annual.ResolvedPayments{
			Payments: []annual.Payment{
				{Month: 1, Amount: cents(t, "1000.00"), Source: annual.PaymentFiled},
				{Month: 2, Amount: cents(t, "2000.00"), Source: annual.PaymentTaxInput},
				{Month: 3, Amount: cents(t, "3000.00"), Source: annual.PaymentEstimatedByApp},
				{Month: 4, Amount: cents(t, "4000.00"), Source: annual.PaymentEstimatedByApp},
			},
			Missing: []string{"filed_payment:601:2026-04"},
		},
	}

	got, err := pm.NewGeneral().Compute(in, params(t))

	require.NoError(t, err)
	lines := trace(got.Lines)
	assert.Contains(t, lines, "isr.payments_filed subtract 1000.00")
	assert.Contains(t, lines, "isr.payments_tax_input subtract 2000.00")
	assert.Contains(t, lines, "isr.payments_estimated subtract 7000.00")
	assert.Contains(t, lines, "isr.refund_only_paid info 0.00")
	assert.Equal(t, cents(t, "10000.00"), got.Tax.Payments)
	assert.Equal(t, pm.Balance{Kind: annual.BalanceAFavor, Amount: cents(t, "7000.00")}, got.Tax.Balance, "10,000 × 30 % − 10,000")
	assert.Equal(t, annual.StatusIncomplete, got.Status)
	assert.Equal(t, []string{"filed_payment:601:2026-04"}, got.MissingInputs)
	assert.Contains(t, got.Disclaimers, annual.DisclaimerPaymentsEstimated)
	for _, l := range got.Lines {
		if l.Code == pm.LinePaymentsEstimated {
			assert.Equal(t, []string{"2"}, l.Args, "the number of estimated months")
			assert.Equal(t, "computed", l.Source.String())
		}
	}
}

func TestRESICOComputesAN07(t *testing.T) {
	in := pm.Input{
		FiscalYear: fiscalYear, Income: cents(t, "6000000.00"), Deductions: cents(t, "4800000.00"),
		PTUPaid: cents(t, "50000.00"), Payments: filed(t, "360000.00"),
	}

	got, err := pm.NewRESICO().Compute(in, params(t))

	require.NoError(t, err)
	assert.Equal(t, annual.SectionPMRESICO, got.Kind)
	assert.Equal(t, fiscal.RegimeRESICO, got.Regime)
	assert.Equal(t, annual.StatusComplete, got.Status)
	assert.Equal(t, annual.ObligationObligated, got.Obligation)
	assert.Equal(t, []annual.CPATopic{annual.TopicRESICOPMPTUBase}, got.CPATopics)
	assert.Equal(t, pm.Tax{
		AccumulableIncome: cents(t, "6000000.00"), Base: cents(t, "1150000.00"), ISR: cents(t, "345000.00"),
		Payments: cents(t, "360000.00"),
		Balance:  pm.Balance{Kind: annual.BalanceAFavor, Amount: cents(t, "15000.00")},
	}, got.Tax)
	assert.Equal(t, pm.Extras{PTUBase: cents(t, "1150000.00")}, got.Extras, "PTU base per 212 ¶2 (⚠️); no CU, no PTU to distribute (§2.1)")
	assert.Equal(t, []string{
		"isr.income input 6000000.00",
		"isr.deductions subtract 4800000.00",
		"isr.profit_before_ptu result 1200000.00",
		"isr.ptu_paid subtract 50000.00",
		"isr.taxable_profit result 1150000.00",
		"isr.base result 1150000.00",
		"isr.rate multiply 345000.00",
		"isr.annual_tax result 345000.00",
		"isr.payments_filed subtract 360000.00",
		"isr.balance result -15000.00",
		"ptu.base info 1150000.00",
	}, trace(got.Lines))
}

// TestRESICOInterestWithholdingIsFlagged: crediting the art. 54 withholding
// in RESICO PM is the open CPA question of the monthly engine.
func TestRESICOInterestWithholdingIsFlagged(t *testing.T) {
	in := pm.Input{FiscalYear: fiscalYear, Income: cents(t, "100.00"), InterestWithholding: cents(t, "1.00")}

	got, err := pm.NewRESICO().Compute(in, params(t))

	require.NoError(t, err)
	for _, l := range got.Lines {
		if l.Code == pm.LineInterestWithholding {
			assert.True(t, l.NeedsCPAReview)
			return
		}
	}
	t.Fatal("no interest withholding line")
}

func TestRESICOLossYearHasNoPTUBase(t *testing.T) {
	in := pm.Input{FiscalYear: fiscalYear, Income: cents(t, "100.00"), Deductions: cents(t, "200.00")}

	got, err := pm.NewRESICO().Compute(in, params(t))

	require.NoError(t, err)
	assert.Equal(t, money.Cents(0), got.Extras.PTUBase)
	assert.Equal(t, money.Cents(0), got.Tax.ISR)
}

func TestComputeRefusesInvalidInput(t *testing.T) {
	tests := map[string]func(*pm.Input){
		"negative income":       func(in *pm.Input) { in.Income = -1 },
		"negative deductions":   func(in *pm.Input) { in.Deductions = -1 },
		"negative depreciation": func(in *pm.Input) { in.Depreciation = -1 },
		"negative ptu":          func(in *pm.Input) { in.PTUPaid = -1 },
		"negative losses":       func(in *pm.Input) { in.PriorYearLosses = -1 },
		"negative interest":     func(in *pm.Input) { in.InterestWithholding = -1 },
		"negative payment": func(in *pm.Input) {
			in.Payments = annual.ResolvedPayments{Payments: []annual.Payment{{Amount: -1, Source: annual.PaymentFiled}}}
		},
		"unknown payment source": func(in *pm.Input) {
			in.Payments = annual.ResolvedPayments{Payments: []annual.Payment{{Amount: 1, Source: "guessed"}}}
		},
		"no fiscal year": func(in *pm.Input) { in.FiscalYear = 0 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			in := an06(t)
			mutate(&in)

			_, err := pm.NewGeneral().Compute(in, params(t))

			require.ErrorIs(t, err, pm.ErrInvalidInput)
		})
	}
}

func TestComputeReportsParamsAndOverflowErrors(t *testing.T) {
	_, err := pm.NewRESICO().Compute(an06(t), &fakeParams{err: errParams})
	require.ErrorIs(t, err, errParams)

	in := pm.Input{FiscalYear: fiscalYear, Income: money.Cents(1 << 62), Payments: annual.ResolvedPayments{Payments: []annual.Payment{
		{Amount: money.Cents(1 << 62), Source: annual.PaymentFiled},
		{Amount: money.Cents(1 << 62), Source: annual.PaymentFiled},
	}}}
	_, err = pm.NewGeneral().Compute(in, params(t))
	require.ErrorIs(t, err, money.ErrOverflow)
}

func TestKinds(t *testing.T) {
	assert.Equal(t, annual.SectionPM601, pm.NewGeneral().Kind())
	assert.Equal(t, annual.SectionPMRESICO, pm.NewRESICO().Kind())
}

// TestComputeIsDeterministic: same input and params, same section.
func TestComputeIsDeterministic(t *testing.T) {
	first, err := pm.NewGeneral().Compute(an06(t), params(t))
	require.NoError(t, err)
	for range 3 {
		again, err := pm.NewGeneral().Compute(an06(t), params(t))
		require.NoError(t, err)
		assert.Equal(t, first, again)
	}
}
