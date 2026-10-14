package annualgolden_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// validCase is a complete annual case; tests break one field at a time.
const validCase = `# golden: hand-authored from research AN-02. Never generated or rewritten by a tool.
id: AN-02
source: research:AN-02
pending_strategy: true
input:
  fiscal_year: 2026
  person_type: fisica
  regimes:
    - regime: "612"
      income: "400000.00"
      deductions: "150000.00"
      isr_withheld: "15000.00"
    - regime: "606"
      income: "240000.00"
      predial: "6000.00"
      options: {deduction_mode: blind_35}
      tax_inputs:
        - {key: ptu_paid, money: "0.00"}
  payments:
    - {regime: "612", amount: "20000.00", source: filed}
    - {regime: "606", month: 12, amount: "18000.00", source: estimated_by_app}
  salary_statements:
    - {employer: "Empleador", accumulable: "360000.00", exempt: "20000.00", isr_withheld: "54235.91"}
  personal_deductions:
    - {category: medical, amount: "30000.00", payment_method: electronic, beneficiary: self}
    - {category: donation, amount: "10000.00", payment_method: electronic, beneficiary: self, donee_type: authorized}
    - {category: tuition, amount: "30000.00", payment_method: electronic, beneficiary: descendant, student: "Hijo 1", tuition_level: primaria, scholarship: "0.00"}
  annual_inputs:
    prior_year_accumulable_income: "700000.00"
    other_exempt_income: "0.00"
    told_employer_will_file: false
expected:
  informative_income_over_500k: true
  disclaimers: [annual_estimate, deductions_as_entered]
  sections:
    - kind: pf_art152
      regimes: ["605", "606", "612"]
      status: complete
      obligation: obligated
      cpa_topics: [uma_for_minimum_wage]
      missing_inputs: []
      tax:
        accumulable_income: "760000.00"
        personal_deductions: "167900.00"
        base: "592100.00"
        isr: "107435.78"
        credits: {payments: "38000.00", withholdings: "15000.00", salary_isr: "54235.91"}
        balance: {kind: a_cargo, amount: "199.87"}
      lines:
        - {code: deductions.cap_applied, op: info, amount: "12900.00"}
`

func cents(t *testing.T, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func centsPtr(t *testing.T, s string) *money.Cents {
	t.Helper()
	c := cents(t, s)
	return &c
}

func TestDecodeReadsEveryField(t *testing.T) {
	c, err := annualgolden.Decode([]byte(validCase), "pf/AN-02.yaml")

	require.NoError(t, err)
	assert.Equal(t, "AN-02", c.ID)
	assert.Equal(t, "pf/AN-02.yaml", c.File)
	assert.Equal(t, "research:AN-02", c.Source)
	assert.True(t, c.PendingStrategy)
	assert.Empty(t, c.SkipReason)

	in := c.Input
	assert.Equal(t, 2026, in.FiscalYear)
	assert.Equal(t, fiscal.PersonFisica, in.PersonType)
	require.Len(t, in.Regimes, 2)
	assert.Equal(t, annualgolden.RegimeInput{
		Regime: "612", Income: cents(t, "400000.00"), Deductions: cents(t, "150000.00"),
		ISRWithheld: cents(t, "15000.00"),
	}, in.Regimes[0])
	assert.Equal(t, fiscal.RegimeCode("606"), in.Regimes[1].Regime)
	assert.Equal(t, cents(t, "6000.00"), in.Regimes[1].Predial)
	assert.Equal(t, map[string]string{"deduction_mode": "blind_35"}, in.Regimes[1].Options)
	assert.Equal(t, []annualgolden.TaxInput{{Key: "ptu_paid", Money: 0}}, in.Regimes[1].TaxInputs)
	assert.Equal(t, []annualgolden.Payment{
		{Regime: "612", Amount: cents(t, "20000.00"), Source: annualgolden.PaymentFiled},
		{Regime: "606", Month: 12, Amount: cents(t, "18000.00"), Source: annualgolden.PaymentEstimatedByApp},
	}, in.Payments)
	assert.Equal(t, []annualgolden.SalaryStatement{{
		Employer: "Empleador", Accumulable: cents(t, "360000.00"), Exempt: cents(t, "20000.00"),
		ISRWithheld: cents(t, "54235.91"),
	}}, in.SalaryStatements)
	require.Len(t, in.PersonalDeductions, 3)
	assert.Equal(t, annual.Deduction{
		Category: annual.CategoryTuition, Amount: cents(t, "30000.00"), PaymentMethod: annual.PaymentElectronic,
		Beneficiary: annual.BeneficiaryDescendant, Student: "Hijo 1", TuitionLevel: annual.LevelPrimaria,
	}, in.PersonalDeductions[2])
	assert.Equal(t, annual.DoneeAuthorized, in.PersonalDeductions[1].DoneeType)
	assert.Equal(t, annualgolden.AnnualInputs{PriorYearAccumulableIncome: centsPtr(t, "700000.00")}, in.AnnualInputs)

	want := c.Want
	require.NotNil(t, want.InformativeIncomeOver500k)
	assert.True(t, *want.InformativeIncomeOver500k)
	assert.Equal(t, []annual.DisclaimerCode{annual.DisclaimerAnnualEstimate, annual.DisclaimerDeductionsAsEntered}, want.Disclaimers)
	require.Len(t, want.Sections, 1)
	s := want.Sections[0]
	assert.Equal(t, annual.SectionPFArt152, s.Kind)
	assert.Equal(t, []fiscal.RegimeCode{"605", "606", "612"}, s.Regimes)
	assert.Equal(t, annual.StatusComplete, s.Status)
	assert.Equal(t, annual.ObligationObligated, s.Obligation)
	assert.Equal(t, []annual.CPATopic{annual.TopicUMAForMinimumWage}, s.CPATopics)
	assert.Equal(t, []string{}, s.MissingInputs)
	assert.Equal(t, &annualgolden.TaxWant{
		AccumulableIncome: centsPtr(t, "760000.00"), PersonalDeductions: centsPtr(t, "167900.00"),
		Base: centsPtr(t, "592100.00"), ISR: centsPtr(t, "107435.78"),
		Payments: centsPtr(t, "38000.00"), Withholdings: centsPtr(t, "15000.00"), SalaryISR: centsPtr(t, "54235.91"),
		Balance: &annualgolden.Balance{Kind: annual.BalanceACargo, Amount: cents(t, "199.87")},
	}, s.Tax)
	assert.Equal(t, []annualgolden.Line{{Code: "deductions.cap_applied", Op: domain.OpInfo, Amount: cents(t, "12900.00")}}, s.Lines)
	assert.Nil(t, s.Difference)
	assert.Nil(t, s.PMExtras)
}

func TestDecodeReadsPMExtrasAndTheInformativeDifference(t *testing.T) {
	doc := `# golden: hand-authored
id: AN-06
source: research:AN-06
input:
  fiscal_year: 2026
  person_type: moral
  regimes:
    - {regime: "601", income: "10000000.00", deductions: "8500000.00"}
expected:
  sections:
    - kind: pm_601
      status: complete
      pm_extras: {ptu_base: "1500000.00", ptu_to_distribute: "150000.00", suggested_cu_next_year: "0.1380"}
    - kind: pf_resico_check
      status: informational
      difference: "-1320.00"
`
	c, err := annualgolden.Decode([]byte(doc), "pm/AN-06.yaml")

	require.NoError(t, err)
	require.Len(t, c.Want.Sections, 2)
	cu := money.Rate(138000)
	assert.Equal(t, &annualgolden.PMExtras{
		PTUBase: centsPtr(t, "1500000.00"), PTUToDistribute: centsPtr(t, "150000.00"), SuggestedCU: &cu,
	}, c.Want.Sections[0].PMExtras)
	assert.Equal(t, centsPtr(t, "-1320.00"), c.Want.Sections[1].Difference)
	assert.Nil(t, c.Want.Sections[0].CPATopics, "absent cpa_topics are not asserted")
	assert.Nil(t, c.Want.Disclaimers, "absent disclaimers are not asserted")
}

func TestDecodeRefusesBrokenCases(t *testing.T) {
	tests := []struct {
		name    string
		old     string
		new     string
		wantErr error
	}{
		{"missing header", "# golden: hand-authored from research AN-02.", "# generated", golden.ErrNotHandAuthored},
		{"unknown field", "pending_strategy: true", "pending_strategy: true\nupdate: true", golden.ErrSyntax},
		{"float amount", `income: "400000.00"`, `income: 400000.00`, golden.ErrInexactAmount},
		{"monthly id", "id: AN-02", "id: PF-612-ISR-01", golden.ErrInvalidCase},
		{"bad source", "source: research:AN-02", "source: research:PF-612-ISR-01", golden.ErrInvalidCase},
		{"blank skip reason", "pending_strategy: true", "pending_strategy: true\nskip_reason: \" \"", golden.ErrInvalidCase},
		{"no fiscal year", "fiscal_year: 2026", "fiscal_year: 0", golden.ErrInvalidCase},
		{"bad person type", "person_type: fisica", "person_type: ambos", golden.ErrInvalidCase},
		{"bad regime", `regime: "612"`, `regime: "61"`, golden.ErrInvalidCase},
		{"negative income", `income: "400000.00"`, `income: "-1.00"`, golden.ErrInvalidCase},
		{"payment month 13", "month: 12", "month: 13", golden.ErrInvalidCase},
		{"payment source", "source: filed", "source: guessed", golden.ErrInvalidCase},
		{"payment of a regime not in the case", `{regime: "612", amount`, `{regime: "626", amount`, golden.ErrInvalidCase},
		{"tax input without key", "key: ptu_paid", "key: \" \"", golden.ErrInvalidCase},
		{"invalid deduction", "category: medical", "category: gym", golden.ErrInvalidCase},
		{"unknown section", "kind: pf_art152", "kind: pf_612", golden.ErrInvalidCase},
		{"unknown status", "status: complete", "status: partial", golden.ErrInvalidCase},
		{"unknown obligation", "obligation: obligated", "obligation: maybe", golden.ErrInvalidCase},
		{"unknown topic", "cpa_topics: [uma_for_minimum_wage]", "cpa_topics: [rif_reduction]", golden.ErrInvalidCase},
		{"unknown disclaimer", "disclaimers: [annual_estimate,", "disclaimers: [not_tax_advice,", golden.ErrInvalidCase},
		{"unknown balance kind", "kind: a_cargo", "kind: zero", golden.ErrInvalidCase},
		{"negative balance", `amount: "199.87"}`, `amount: "-199.87"}`, golden.ErrInvalidCase},
		{"line op", "op: info", "op: guess", golden.ErrInvalidCase},
		{"line without code", "code: deductions.cap_applied", "code: \"\"", golden.ErrInvalidCase},
		{"second document", "        - {code: deductions.cap_applied, op: info, amount: \"12900.00\"}\n",
			"        - {code: deductions.cap_applied, op: info, amount: \"12900.00\"}\n---\nid: X\n", golden.ErrSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Contains(t, validCase, tt.old)
			doc := strings.Replace(validCase, tt.old, tt.new, 1)

			_, err := annualgolden.Decode([]byte(doc), "pf/AN-02.yaml")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDecodeRefusesAnEmptyOrRepeatedSectionList(t *testing.T) {
	head, _, found := strings.Cut(validCase, "  sections:")
	require.True(t, found)
	tests := []struct {
		name string
		tail string
	}{
		{"empty", "  sections: []\n"},
		{"repeated kind", "  sections:\n    - {kind: pm_601, status: complete}\n    - {kind: pm_601, status: complete}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := annualgolden.Decode([]byte(head+tt.tail), "pf/AN-02.yaml")

			assert.ErrorIs(t, err, golden.ErrInvalidCase)
		})
	}
}
