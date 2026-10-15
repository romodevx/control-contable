package annual_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

const art152Salary = fiscal.RegimeCode("605")

// art152Params is the embedded 2026 set: the strategy runs on the
// published Anexo 8 C.II table and thresholds, not a synthetic copy.
func art152Params(t testing.TB) *taxparams.Params {
	t.Helper()
	c, err := params.Embedded()
	require.NoError(t, err)
	sel, err := c.Select(2026)
	require.NoError(t, err)
	return sel.Params
}

// art152Payments is the payments port as a map: what ANN-14 resolves by
// precedence (filed > tax input > app estimate) per regime.
type art152Payments map[fiscal.RegimeCode][]annual.Payment

func (p art152Payments) ProvisionalPayments(regime fiscal.RegimeCode) ([]annual.Payment, error) {
	return p[regime], nil
}

type art152FailingPayments struct{ err error }

func (f art152FailingPayments) ProvisionalPayments(fiscal.RegimeCode) ([]annual.Payment, error) {
	return nil, f.err
}

func art152Filed(amount money.Cents) annual.Payment {
	return annual.Payment{Amount: amount, Source: annual.PaymentFiled}
}

func art152Compute(t *testing.T, in annual.PFYear, pay annual.PaymentsReader) annual.PFArt152Result {
	t.Helper()
	got, err := annual.NewPFArt152(calculator(t)).Compute(in, pay, art152Params(t))
	require.NoError(t, err)
	return got
}

func art152Business(income, deductions, withheld money.Cents) annual.RegimeYear {
	return annual.RegimeYear{Regime: fiscal.RegimeBusiness, Income: income, Deductions: deductions, ISRWithheld: withheld}
}

func art152Constancia(accumulable, exempt, withheld money.Cents) annual.SalaryStatement {
	return annual.SalaryStatement{Accumulable: accumulable, Exempt: exempt, ISRWithheld: withheld}
}

func art152Codes(lines []domain.Line) []domain.LineCode {
	out := make([]domain.LineCode, len(lines))
	for i, l := range lines {
		out[i] = l.Code
	}
	return out
}

func art152Line(t *testing.T, lines []domain.Line, code domain.LineCode) domain.Line {
	t.Helper()
	i := slices.IndexFunc(lines, func(l domain.Line) bool { return l.Code == code })
	require.GreaterOrEqual(t, i, 0, "no %s line in %v", code, art152Codes(lines))
	return lines[i]
}

// TestPFArt152ReproducesAN01 is research AN-01: 612 only, medical 40,000.
func TestPFArt152ReproducesAN01(t *testing.T) {
	in := annual.PFYear{
		FiscalYear: 2026,
		Regimes:    []annual.RegimeYear{art152Business(pesos(t, "900000.00"), pesos(t, "400000.00"), pesos(t, "30000.00"))},
		Deductions: []annual.Deduction{electronic(annual.CategoryMedical, pesos(t, "40000.00"))},
	}
	pay := art152Payments{fiscal.RegimeBusiness: {art152Filed(pesos(t, "55773.86"))}}

	got := art152Compute(t, in, pay)

	s := got.Section
	assert.Equal(t, annual.SectionPFArt152, s.Kind)
	assert.Equal(t, []fiscal.RegimeCode{fiscal.RegimeBusiness}, s.Regimes)
	assert.Equal(t, annual.StatusComplete, s.Status)
	assert.Equal(t, annual.ObligationObligated, s.Obligation)
	assert.Empty(t, s.MissingInputs)
	assert.Empty(t, s.CPATopics)
	require.NotNil(t, s.Tax)
	assert.Equal(t, annual.Tax{
		AccumulableIncome: pesos(t, "500000.00"), PersonalDeductions: pesos(t, "40000.00"),
		Base: pesos(t, "460000.00"), ISR: pesos(t, "76365.86"),
		Payments: pesos(t, "55773.86"), Withholdings: pesos(t, "30000.00"),
		Balance: annual.Balance{Kind: annual.BalanceAFavor, Amount: pesos(t, "9408.00")},
	}, *s.Tax)
	assert.True(t, got.InformativeIncomeOver500k)
	assert.Contains(t, s.Disclaimers, annual.DisclaimerDeductionsAsEntered)
	assert.NotContains(t, s.Disclaimers, annual.DisclaimerPaymentsEstimated)
	row := art152Line(t, s.Lines, annual.LineArt152TariffRow)
	assert.Equal(t, 7, row.Row, "Anexo 8 C.II row 7")
}

func TestPFArt152LinesAreTheComputationTrace(t *testing.T) {
	in := annual.PFYear{
		FiscalYear: 2026,
		Regimes: []annual.RegimeYear{
			art152Business(pesos(t, "900000.00"), pesos(t, "400000.00"), pesos(t, "30000.00")),
		},
		SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "100000.00"), 0, pesos(t, "9000.00"))},
		Deductions:       []annual.Deduction{electronic(annual.CategoryMedical, pesos(t, "40000.00"))},
	}
	pay := art152Payments{fiscal.RegimeBusiness: {art152Filed(pesos(t, "55773.86"))}}

	got := art152Compute(t, in, pay)

	assert.Equal(t, []domain.LineCode{
		annual.LineArt152BusinessIncome, annual.LineArt152BusinessDeductions, annual.LineArt152BusinessResult,
		annual.LineArt152Salary, annual.LineArt152BusinessTaxable, annual.LineArt152AccumulableIncome,
		annual.LineArt152PersonalDeductions, annual.LineArt152Base,
		annual.LineArt152TariffRow, annual.LineArt152TariffExcess, annual.LineArt152ISR,
		annual.LineArt152Payments, annual.LineArt152Withheld, annual.LineArt152SalaryISR, annual.LineArt152Balance,
	}, art152Codes(withoutRounding(got.Section.Lines)))
	for _, l := range got.Section.Lines {
		if l.Operation != domain.OpInfo && l.Operation != domain.OpRoundingAdjustment {
			assert.NotEmpty(t, l.Basis, "%s cites its legal basis", l.Code)
		}
	}
}

func withoutRounding(lines []domain.Line) []domain.Line {
	return slices.DeleteFunc(slices.Clone(lines), func(l domain.Line) bool {
		return l.Operation == domain.OpRoundingAdjustment
	})
}

// TestPFArt152ABusinessLossStaysInItsChapter: a 612 loss contributes 0 and
// never reduces the salary (LISR 109); it is reported as a loss line.
func TestPFArt152ABusinessLossStaysInItsChapter(t *testing.T) {
	in := annual.PFYear{
		FiscalYear:       2026,
		Regimes:          []annual.RegimeYear{{Regime: art152Salary}, art152Business(pesos(t, "100000.00"), pesos(t, "160000.00"), 0)},
		SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "300000.00"), 0, pesos(t, "40000.00"))},
	}

	got := art152Compute(t, in, art152Payments{})

	require.NotNil(t, got.Section.Tax)
	assert.Equal(t, pesos(t, "300000.00"), got.Section.Tax.AccumulableIncome, "salary alone: the 60,000 loss is not netted")
	loss := art152Line(t, got.Section.Lines, annual.LineArt152BusinessLoss)
	assert.Equal(t, domain.OpInfo, loss.Operation)
	assert.Equal(t, pesos(t, "60000.00"), loss.Amount)
	assert.Equal(t, money.Cents(0), art152Line(t, got.Section.Lines, annual.LineArt152BusinessTaxable).Amount)
}

// TestPFArt152RentUsesTheBlindDeductionAndItsFloor reproduces the 606 part
// of AN-02 (240,000 - 35 % - predial 6,000 = 150,000) and floors a rent loss.
func TestPFArt152RentUsesTheBlindDeductionAndItsFloor(t *testing.T) {
	blind := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionRentDeductionMode: domain.TextOption("blind_35"),
	})
	actual := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionRentDeductionMode: domain.TextOption("actual"),
	})
	tests := []struct {
		name string
		rent annual.RegimeYear
		want money.Cents
	}{
		{"blind 35 % plus predial", annual.RegimeYear{Regime: fiscal.RegimeRent, Income: pesos(t, "240000.00"), Predial: pesos(t, "6000.00"), Options: blind}, pesos(t, "150000.00")},
		{"absent mode is blind", annual.RegimeYear{Regime: fiscal.RegimeRent, Income: pesos(t, "100000.00")}, pesos(t, "65000.00")},
		{"actual deductions", annual.RegimeYear{Regime: fiscal.RegimeRent, Income: pesos(t, "100000.00"), Deductions: pesos(t, "30000.00"), Options: actual}, pesos(t, "70000.00")},
		{"a rent loss contributes zero", annual.RegimeYear{Regime: fiscal.RegimeRent, Income: pesos(t, "10000.00"), Deductions: pesos(t, "30000.00"), Options: actual}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := art152Compute(t, annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{tt.rent}}, art152Payments{})

			require.NotNil(t, got.Section.Tax)
			assert.Equal(t, tt.want, got.Section.Tax.AccumulableIncome)
		})
	}
}

func TestPFArt152RejectsAnUnknownRentDeductionMode(t *testing.T) {
	odd := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("half")})
	in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: fiscal.RegimeRent, Income: 100, Options: odd}}}

	_, err := annual.NewPFArt152(calculator(t)).Compute(in, art152Payments{}, art152Params(t))

	require.ErrorIs(t, err, domain.ErrRentDeductionMode)
}

// TestPFArt152BusinessSubtractsDepreciationPTUAndLosses: LISR 109 fr. I.
func TestPFArt152BusinessSubtractsDepreciationPTUAndLosses(t *testing.T) {
	r := art152Business(pesos(t, "500000.00"), pesos(t, "100000.00"), 0)
	r.Depreciation, r.PTUPaid, r.PriorYearLosses = pesos(t, "20000.00"), pesos(t, "30000.00"), pesos(t, "50000.00")

	got := art152Compute(t, annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{r}}, art152Payments{})

	assert.Equal(t, pesos(t, "300000.00"), got.Section.Tax.AccumulableIncome)
	codes := art152Codes(got.Section.Lines)
	for _, c := range []domain.LineCode{annual.LineArt152Depreciation, annual.LineArt152PTUPaid, annual.LineArt152PriorYearLosses} {
		assert.Contains(t, codes, c)
	}
}

// TestPFArt152PlatformIncome: 625 non-definitive joins the business chapter
// under CPA review (AN-04 figures); the definitive option is excluded.
func TestPFArt152PlatformIncome(t *testing.T) {
	platform := annual.RegimeYear{
		Regime: fiscal.RegimePlatforms, Income: pesos(t, "480000.00"), Deductions: pesos(t, "300000.00"), ISRWithheld: pesos(t, "12000.00"),
	}
	t.Run("non-definitive is computed like 612 (AN-04)", func(t *testing.T) {
		got := art152Compute(t, annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{platform}}, art152Payments{})

		s := got.Section
		assert.Equal(t, []annual.CPATopic{annual.TopicPlatformNonDefinitiveAnnual}, s.CPATopics)
		assert.Contains(t, s.Disclaimers, annual.DisclaimerCPAReviewPending)
		assert.Equal(t, annual.ObligationObligated, s.Obligation)
		assert.Equal(t, annual.Tax{
			AccumulableIncome: pesos(t, "180000.00"), Base: pesos(t, "180000.00"), ISR: pesos(t, "16833.81"),
			Withholdings: pesos(t, "12000.00"),
			Balance:      annual.Balance{Kind: annual.BalanceACargo, Amount: pesos(t, "4833.81")},
		}, *s.Tax)
		assert.True(t, art152Line(t, s.Lines, annual.LineArt152BusinessIncome).NeedsCPAReview)
	})
	t.Run("the definitive option is not in art. 152", func(t *testing.T) {
		definitive := platform
		definitive.Options = domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
			annual.OptionPlatformDefinitive: domain.BoolOption(true),
		})
		in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{definitive}}

		assert.False(t, annual.NewPFArt152(calculator(t)).Applies(in))
		got := art152Compute(t, in, art152Payments{})
		assert.Empty(t, got.Section.Regimes)
		assert.Equal(t, annual.ObligationNotObligated, got.Section.Obligation)
	})
}

func TestPFArt152AppliesToItsChapters(t *testing.T) {
	s := annual.NewPFArt152(calculator(t))
	for _, tt := range []struct {
		regime fiscal.RegimeCode
		want   bool
	}{
		{art152Salary, true}, {fiscal.RegimeRent, true}, {fiscal.RegimeBusiness, true}, {fiscal.RegimePlatforms, true},
		{fiscal.RegimeRESICO, false}, {"614", false},
	} {
		in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: tt.regime}}}
		assert.Equal(t, tt.want, s.Applies(in), "%s", tt.regime)
	}
	assert.True(t, s.Applies(annual.PFYear{FiscalYear: 2026, SalaryStatements: []annual.SalaryStatement{{}}}),
		"constancias without 605 registered (OQ-A2)")
}

// TestPFArt152Obligation is the MODULE-annual §2.4 table (LISR 98 fr. III).
func TestPFArt152Obligation(t *testing.T) {
	salaryOnly := func(statements ...annual.SalaryStatement) annual.PFYear {
		return annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: art152Salary}}, SalaryStatements: statements}
	}
	under := art152Constancia(pesos(t, "300000.00"), pesos(t, "20000.00"), pesos(t, "30000.00"))
	flagged := func(set func(*annual.SalaryStatement)) annual.SalaryStatement {
		s := under
		set(&s)
		return s
	}
	told := salaryOnly(under)
	told.Inputs.ToldEmployerWillFile = true
	withBusiness := salaryOnly(under)
	withBusiness.Regimes = append(withBusiness.Regimes, art152Business(pesos(t, "1000.00"), 0, 0))

	tests := []struct {
		name   string
		in     annual.PFYear
		want   annual.Obligation
		topics []annual.CPATopic
	}{
		{"one employer under 400,000: optional", salaryOnly(under), annual.ObligationOptional, nil},
		{"salary exactly 400,000 is not above it", salaryOnly(art152Constancia(pesos(t, "380000.00"), pesos(t, "20000.00"), 0)), annual.ObligationOptional, nil},
		{"salary above 400,000 decides", salaryOnly(art152Constancia(pesos(t, "380000.00"), pesos(t, "20000.01"), 0)), annual.ObligationObligated, []annual.CPATopic{annual.TopicSalaryObligationBasis}},
		{"two simultaneous employers", salaryOnly(flagged(func(s *annual.SalaryStatement) { s.Simultaneous = true }), flagged(func(s *annual.SalaryStatement) { s.Simultaneous = true })), annual.ObligationObligated, nil},
		{"one simultaneous flag alone is not two employers", salaryOnly(flagged(func(s *annual.SalaryStatement) { s.Simultaneous = true })), annual.ObligationOptional, nil},
		{"left before year end", salaryOnly(flagged(func(s *annual.SalaryStatement) { s.LeftBeforeYearEnd = true })), annual.ObligationObligated, nil},
		{"payer not withholding", salaryOnly(flagged(func(s *annual.SalaryStatement) { s.PayerNotWithholding = true })), annual.ObligationObligated, nil},
		{"told the employer", told, annual.ObligationObligated, nil},
		{"other accumulable income", withBusiness, annual.ObligationObligated, nil},
		{"no income in any chapter", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: fiscal.RegimeBusiness}}}, annual.ObligationNotObligated, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := art152Compute(t, tt.in, art152Payments{})

			assert.Equal(t, tt.want, got.Section.Obligation)
			assert.Equal(t, tt.topics, got.Section.CPATopics)
		})
	}
}

// TestPFArt152MissingInputs: MODULE-annual §2.6 fallbacks make the section
// incomplete, never silently complete.
func TestPFArt152MissingInputs(t *testing.T) {
	t.Run("605 without a constancia", func(t *testing.T) {
		got := art152Compute(t, annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: art152Salary}}}, art152Payments{})

		assert.Equal(t, annual.StatusIncomplete, got.Section.Status)
		assert.Equal(t, []string{annual.MissingSalaryStatement}, got.Section.MissingInputs)
		assert.Contains(t, got.Section.Disclaimers, annual.DisclaimerIncompleteInputs)
	})
	t.Run("donations without the prior year's acumulables", func(t *testing.T) {
		in := annual.PFYear{
			FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(pesos(t, "500000.00"), 0, 0)},
			Deductions: []annual.Deduction{donation(annual.DoneeAuthorized, pesos(t, "1000.00"))},
		}

		got := art152Compute(t, in, art152Payments{})

		assert.Equal(t, annual.StatusIncomplete, got.Section.Status)
		assert.Equal(t, []string{annual.MissingPriorYearAccumulableIncome}, got.Section.MissingInputs)
		assert.Equal(t, money.Cents(0), got.Section.Tax.PersonalDeductions)
		art152Line(t, got.Section.Lines, domain.LineCode(annual.NoteExcludedUnverified))
	})
	t.Run("a default payment from an incomplete monthly estimate", func(t *testing.T) {
		in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(pesos(t, "500000.00"), 0, 0)}}
		pay := art152Payments{fiscal.RegimeBusiness: {
			{Month: 3, Amount: pesos(t, "100.00"), Source: annual.PaymentEstimatedByApp, FromIncompleteEstimate: true},
			{Month: 4, Amount: pesos(t, "100.00"), Source: annual.PaymentEstimatedByApp},
		}}

		got := art152Compute(t, in, pay)

		assert.Equal(t, annual.StatusIncomplete, got.Section.Status)
		assert.Equal(t, []string{"filed_payment:612:2026-03"}, got.Section.MissingInputs)
		assert.Contains(t, got.Section.Disclaimers, annual.DisclaimerPaymentsEstimated)
	})
}

// TestPFArt152CreditsKeepTheirSources: one payments line per regime and
// source, in precedence order, each naming where it came from (§2.5).
func TestPFArt152CreditsKeepTheirSources(t *testing.T) {
	in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{
		art152Business(pesos(t, "500000.00"), 0, 0),
		{Regime: fiscal.RegimeRent, Income: pesos(t, "100000.00")},
	}}
	pay := art152Payments{
		fiscal.RegimeBusiness: {
			{Month: 2, Amount: pesos(t, "300.00"), Source: annual.PaymentEstimatedByApp},
			{Month: 1, Amount: pesos(t, "100.00"), Source: annual.PaymentFiled},
			{Month: 3, Amount: pesos(t, "200.00"), Source: annual.PaymentTaxInput},
		},
		fiscal.RegimeRent: {{Month: 1, Amount: pesos(t, "50.00"), Source: annual.PaymentFiled}},
	}

	got := art152Compute(t, in, pay)

	var payments []domain.Line
	for _, l := range got.Section.Lines {
		if l.Code == annual.LineArt152Payments {
			payments = append(payments, l)
		}
	}
	require.Len(t, payments, 4)
	type shown struct {
		amount money.Cents
		source string
		args   []string
	}
	var gotShown []shown
	for _, l := range payments {
		gotShown = append(gotShown, shown{l.Amount, l.Source.String(), l.Args})
	}
	assert.Equal(t, []shown{
		{pesos(t, "50.00"), "tax_input:annual_filed_payments", []string{"606", "filed", "1"}},
		{pesos(t, "100.00"), "tax_input:annual_filed_payments", []string{"612", "filed", "1"}},
		{pesos(t, "200.00"), "tax_input:prior_provisional_payments", []string{"612", "tax_input", "1"}},
		{pesos(t, "300.00"), "computed", []string{"612", "estimated_by_app", "1"}},
	}, gotShown)
	assert.Equal(t, pesos(t, "650.00"), got.Section.Tax.Payments)
	assert.Contains(t, got.Section.Disclaimers, annual.DisclaimerPaymentsEstimated)
}

// TestPFArt152BaseAtOrBelowZeroOwesNoTax: LISR 152, research §B.1 step 9.
func TestPFArt152BaseAtOrBelowZeroOwesNoTax(t *testing.T) {
	in := annual.PFYear{
		FiscalYear:       2026,
		Regimes:          []annual.RegimeYear{{Regime: art152Salary}},
		SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "10000.00"), 0, pesos(t, "500.00"))},
		Deductions:       []annual.Deduction{electronic(annual.CategoryMedicalDisability, pesos(t, "20000.00"))},
	}

	got := art152Compute(t, in, art152Payments{})

	assert.Equal(t, money.Cents(0), got.Section.Tax.ISR)
	assert.Equal(t, money.Cents(0), got.Section.Tax.Base, "contract AnnualTax.base_cents: PF base ≤ 0 → 0")
	assert.Equal(t, pesos(t, "-10000.00"), art152Line(t, got.Section.Lines, annual.LineArt152Base).Amount, "the trace keeps the signed base")
	assert.Equal(t, annual.Balance{Kind: annual.BalanceAFavor, Amount: pesos(t, "500.00")}, got.Section.Tax.Balance)
	art152Line(t, got.Section.Lines, annual.LineArt152NoBase)
}

func TestPFArt152BalanceKinds(t *testing.T) {
	tests := []struct {
		name     string
		withheld string
		want     annual.Balance
	}{
		{"a cargo", "0.00", annual.Balance{Kind: annual.BalanceACargo, Amount: 19_200}},
		{"cero", "192.00", annual.Balance{Kind: annual.BalanceCero}},
		{"a favor", "200.00", annual.Balance{Kind: annual.BalanceAFavor, Amount: 800}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := annual.PFYear{
				FiscalYear:       2026,
				Regimes:          []annual.RegimeYear{{Regime: art152Salary}},
				SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "10000.00"), 0, pesos(t, tt.withheld))},
			}

			got := art152Compute(t, in, art152Payments{})

			assert.Equal(t, tt.want, got.Section.Tax.Balance, "row 1: (10,000.00 - 0.01) x 1.92 percent = 192.00")
		})
	}
}

// TestPFArt152CapOnRESICOIncome: RESICO income stays out of the global
// cap's total income (OQ-A4) and raises its CPA topic.
func TestPFArt152CapOnRESICOIncome(t *testing.T) {
	in := annual.PFYear{
		FiscalYear: 2026,
		Regimes: []annual.RegimeYear{
			{Regime: art152Salary}, {Regime: fiscal.RegimeRESICO, Income: pesos(t, "1000000.00")},
		},
		SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "200000.00"), 0, 0)},
		Deductions:       []annual.Deduction{electronic(annual.CategoryMedical, pesos(t, "50000.00"))},
	}

	got := art152Compute(t, in, art152Payments{})

	assert.Equal(t, pesos(t, "30000.00"), got.Section.Tax.PersonalDeductions, "15 % of 200,000, not of 1,200,000")
	assert.Contains(t, got.Section.CPATopics, annual.TopicDeductionCapRESICOIncome)
	assert.Equal(t, []fiscal.RegimeCode{art152Salary}, got.Section.Regimes)
	assert.True(t, got.InformativeIncomeOver500k, "LISR 150 ¶3 counts every income, RESICO included")
}

func TestPFArt152InformativeIncomeThreshold(t *testing.T) {
	for _, tt := range []struct {
		exempt string
		want   bool
	}{{"0.00", false}, {"0.01", true}} {
		in := annual.PFYear{
			FiscalYear:       2026,
			Regimes:          []annual.RegimeYear{{Regime: art152Salary}},
			SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "450000.00"), 0, 0)},
			Inputs:           annual.PFAnnualInputs{OtherExemptIncome: pesos(t, "50000.00") + pesos(t, tt.exempt)},
		}

		got := art152Compute(t, in, art152Payments{})

		assert.Equal(t, tt.want, got.InformativeIncomeOver500k, "LISR 150 ¶3: above 500,000")
	}
}

func TestPFArt152RejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		in   annual.PFYear
		pay  annual.PaymentsReader
	}{
		{"negative income", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(-1, 0, 0)}}, art152Payments{}},
		{"negative constancia", annual.PFYear{FiscalYear: 2026, SalaryStatements: []annual.SalaryStatement{art152Constancia(0, 0, -1)}}, art152Payments{}},
		{"a regime held twice", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0), art152Business(2, 0, 0)}}, art152Payments{}},
		{"negative payment", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}, art152Payments{fiscal.RegimeBusiness: {art152Filed(-1)}}},
		{"unknown payment source", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}, art152Payments{fiscal.RegimeBusiness: {{Amount: 1, Source: "guess"}}}},
		{"a month paid twice", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}, art152Payments{fiscal.RegimeBusiness: {
			{Month: 3, Amount: 1, Source: annual.PaymentFiled}, {Month: 3, Amount: 1, Source: annual.PaymentEstimatedByApp},
		}}},
		{"a year total beside monthly payments", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}, art152Payments{fiscal.RegimeBusiness: {
			art152Filed(1), {Month: 1, Amount: 1, Source: annual.PaymentTaxInput},
		}}},
		{"payment month 13", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}, art152Payments{fiscal.RegimeBusiness: {{Month: 13, Amount: 1, Source: annual.PaymentFiled}}}},
		{"invalid deduction", annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}, Deductions: []annual.Deduction{{Category: "yacht"}}}, art152Payments{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := annual.NewPFArt152(calculator(t)).Compute(tt.in, tt.pay, art152Params(t))

			require.Error(t, err)
			assert.True(t, errors.Is(err, annual.ErrInvalidYear) || errors.Is(err, annual.ErrInvalidDeduction), "%v", err)
		})
	}
}

func TestPFArt152PropagatesPortErrors(t *testing.T) {
	boom := errors.New("payments unavailable")
	in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(1, 0, 0)}}

	_, err := annual.NewPFArt152(calculator(t)).Compute(in, art152FailingPayments{err: boom}, art152Params(t))

	require.ErrorIs(t, err, boom)
}

func TestPFArt152NeedsTheAnnualTariff(t *testing.T) {
	in := annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{art152Business(pesos(t, "1000.00"), 0, 0)}}

	_, err := annual.NewPFArt152(calculator(t)).Compute(in, art152Payments{}, rentOnlyParams(t))

	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}

// rentOnlyParams is a valid set without the annual tariff.
func rentOnlyParams(t *testing.T) *taxparams.Params {
	t.Helper()
	day := func(s string) fiscal.Date {
		d, err := fiscal.ParseDate(s)
		require.NoError(t, err)
		return d
	}
	p, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: day("2026-01-01"),
		Sources: []taxparams.Source{{ID: "lisr", Title: "LISR", URL: "https://example.test/lisr", DOFDate: day("2025-11-07"), Reference: "Art. 115"}},
		Values: []taxparams.Value{{
			Key: "isr.rent.blind_deduction", Kind: taxparams.KindRate, Rate: 350_000,
			Validity: taxparams.Validity{From: day("2026-01-01")}, Basis: []taxparams.Citation{{Source: "lisr", Article: "Art. 115"}},
		}},
	})
	require.NoError(t, err)
	return p
}

// TestPFArt152IsDeterministic: same input, same section, line for line.
func TestPFArt152IsDeterministic(t *testing.T) {
	in := annual.PFYear{
		FiscalYear: 2026,
		Regimes: []annual.RegimeYear{
			{Regime: fiscal.RegimeRent, Income: pesos(t, "240000.00")}, art152Business(pesos(t, "400000.00"), pesos(t, "150000.00"), 0),
			{Regime: art152Salary},
		},
		SalaryStatements: []annual.SalaryStatement{art152Constancia(pesos(t, "360000.00"), 0, 0)},
		Deductions: []annual.Deduction{
			tuition("B", annual.LevelPrimaria, pesos(t, "30000.00"), 0), tuition("A", annual.LevelSecundaria, pesos(t, "30000.00"), 0),
		},
	}

	first := art152Compute(t, in, art152Payments{})
	for range 5 {
		assert.Equal(t, first, art152Compute(t, in, art152Payments{}))
	}
	assert.Equal(t, []fiscal.RegimeCode{art152Salary, fiscal.RegimeRent, fiscal.RegimeBusiness}, first.Section.Regimes)
}
