package golden_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// validCase is a complete hand-authored case; tests break one field at a time.
const validCase = `# golden: hand-authored from research PF-612-ISR-01. Never generated or rewritten by a tool.
id: PF-612-ISR-01
source: research:PF-612-ISR-01
pending_strategy: true
input:
  regime: "612"
  person_type: fisica
  period: "2026-03"
  options:
    iva_fully_taxable: true
    deduction_mode: actual
    activities: [lodging, goods_services]
  months:
    - month: 1
      income: {cash: "30000.00", isr_withheld: {by_moral: "1000.00"}}
      expenses: {deductible_paid: "10000.00"}
    - month: 3
      income: {cash: "60000.00", iva_collected: "4800.00", iva_withheld: {by_moral: "1066.67"}}
      expenses: {deductible_paid: "20000.00", creditable_iva_paid: "1600.00"}
  tax_inputs:
    - {key: prior_provisional_payments, kind: money_monthly, month: 1, money: "2000.00"}
    - {key: coeficiente_utilidad, kind: rate, rate: "0.15"}
    - {key: first_fiscal_year, kind: boolean, bool: false}
expected:
  status: complete
  taxes: [isr, iva]
  isr: "150.96"
  iva:
    charged_collected: "4800.00"
    creditable_paid: "1600.00"
    withheld_from_taxpayer: "1066.67"
    definitive_excluded: "0.00"
    net: "2133.33"
  lines:
    - {tax: isr, code: isr.income_ytd, amount: "90000.00"}
    - {tax: iva, code: iva.net, amount: "2133.33"}
  disclaimers: [deductibility_claimed]
`

func decode(t *testing.T, doc string) (golden.Case, error) {
	t.Helper()
	return golden.Decode([]byte(doc), "612-pf/PF-612-ISR-01.yaml")
}

func cents(t *testing.T, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func TestDecodeReadsEveryField(t *testing.T) {
	c, err := decode(t, validCase)
	require.NoError(t, err)

	assert.Equal(t, "PF-612-ISR-01", c.ID)
	assert.Equal(t, "612-pf/PF-612-ISR-01.yaml", c.File)
	assert.Equal(t, "research:PF-612-ISR-01", c.Source)
	assert.True(t, c.PendingStrategy)
	assert.False(t, c.Generated)
	assert.Empty(t, c.SkipReason)

	in := c.Input
	assert.Equal(t, domain.StrategyKey{Regime: "612", PersonType: fiscal.PersonFisica}, in.Key)
	assert.Equal(t, "2026-03", in.Period.String())
	assert.Equal(t, []time.Month{time.January, time.February, time.March}, c.ActiveMonths)

	fully, ok := in.Options.Bool("iva_fully_taxable")
	assert.True(t, ok && fully)
	mode, ok := in.Options.Text("deduction_mode")
	assert.True(t, ok)
	assert.Equal(t, "actual", mode)
	acts, ok := in.Options.List("activities")
	assert.True(t, ok)
	assert.Equal(t, []string{"lodging", "goods_services"}, acts)

	jan, mar := in.Sums.Month(time.January), in.Sums.Month(time.March)
	assert.Equal(t, cents(t, "30000.00"), jan.Income.Cash)
	assert.Equal(t, cents(t, "1000.00"), jan.Income.ISRWithheld.ByMoral)
	assert.Equal(t, cents(t, "10000.00"), jan.Expenses.DeductiblePaid)
	assert.Equal(t, domain.MonthSums{}, in.Sums.Month(time.February))
	assert.Equal(t, cents(t, "4800.00"), mar.Income.IVACollected)
	assert.Equal(t, cents(t, "1066.67"), mar.Income.IVAWithheld.ByMoral)
	assert.Equal(t, cents(t, "1600.00"), mar.Expenses.CreditableIVAPaid)

	prior, ok := in.Inputs.MonthlyMoney(domain.InputPriorProvisionalPayments, time.January)
	assert.True(t, ok)
	assert.Equal(t, cents(t, "2000.00"), prior)
	cu, ok := in.Inputs.Rate(domain.InputCoeficienteUtilidad)
	assert.True(t, ok)
	assert.Equal(t, money.Rate(150_000), cu)
	first, ok := in.Inputs.Bool(domain.InputFirstFiscalYear)
	assert.True(t, ok)
	assert.False(t, first)

	w := c.Want
	assert.Equal(t, domain.StatusComplete, w.Status)
	assert.Equal(t, []domain.TaxKind{domain.TaxISR, domain.TaxIVA}, w.Taxes)
	require.NotNil(t, w.ISR)
	assert.Equal(t, cents(t, "150.96"), *w.ISR)
	require.NotNil(t, w.IVA)
	assert.Equal(t, domain.IVAComponents{
		ChargedCollected: cents(t, "4800.00"), CreditablePaid: cents(t, "1600.00"),
		WithheldFromTaxpayer: cents(t, "1066.67"),
	}, *w.IVA)
	assert.Equal(t, []golden.LineWant{
		{Tax: domain.TaxISR, Code: "isr.income_ytd", Amount: cents(t, "90000.00")},
		{Tax: domain.TaxIVA, Code: "iva.net", Amount: cents(t, "2133.33")},
	}, w.Lines)
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed}, w.Disclaimers)
	// Absent lists are asserted empty, never skipped: an authored case
	// asserts everything.
	assert.NotNil(t, w.MissingInputs)
	assert.Empty(t, w.MissingInputs)
	assert.NotNil(t, w.CPATopics)
	assert.Empty(t, w.CPATopics)
}

func TestDecodeReadsSkipReasonAndMissingInputs(t *testing.T) {
	doc := strings.Replace(validCase, "pending_strategy: true\n",
		"skip_reason: \"⚠️ RIF reduction % not re-verified (CPA-REVIEW RIF reduction by year)\"\n", 1)
	doc = strings.Replace(doc, "  disclaimers: [deductibility_claimed]\n",
		"  missing_inputs:\n    - {key: prior_provisional_payments, months: [1, 2]}\n  cpa_topics: [rif_reduction]\n", 1)
	doc = strings.Replace(doc, "status: complete", "status: incomplete", 1)

	c, err := decode(t, doc)
	require.NoError(t, err)

	assert.Equal(t, "⚠️ RIF reduction % not re-verified (CPA-REVIEW RIF reduction by year)", c.SkipReason)
	assert.False(t, c.PendingStrategy)
	assert.Equal(t, []domain.MissingInput{{
		Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.February},
	}}, c.Want.MissingInputs)
	assert.Equal(t, []domain.CPATopic{domain.TopicRIFReduction}, c.Want.CPATopics)
	assert.Equal(t, domain.StatusIncomplete, c.Want.Status)
}

func TestDecodeReadsActiveMonthsAndAbsentTaxFigures(t *testing.T) {
	doc := strings.Replace(validCase, "  options:\n", "  active_months: [2, 3]\n  options:\n", 1)
	doc = strings.Replace(doc, "  isr: \"150.96\"\n", "", 1)
	doc = strings.Replace(doc, "  taxes: [isr, iva]", "  taxes: [isr, iva, withholding]\n  withheld_to_remit: \"5333.33\"", 1)

	c, err := decode(t, doc)
	require.NoError(t, err)

	assert.Equal(t, []time.Month{time.February, time.March}, c.ActiveMonths)
	assert.Nil(t, c.Want.ISR, "isr asserted but absent: the case expects no ISR figure")
	assert.Equal(t, cents(t, "5333.33"), c.Want.WithheldToRemit)
}

func TestDecodeRejectsWhatAHandAuthorCouldGetWrong(t *testing.T) {
	tests := []struct {
		name     string
		old, new string
		want     error
	}{
		{"no provenance header", "# golden: hand-authored", "# generated", golden.ErrNotHandAuthored},
		{"unknown field", "pending_strategy: true", "pending_strategy: true\nupdate: true", golden.ErrSyntax},
		{"unquoted amount", `isr: "150.96"`, `isr: 150.96`, golden.ErrInexactAmount},
		{"one decimal", `isr: "150.96"`, `isr: "150.9"`, golden.ErrInexactAmount},
		{"three decimals", `isr: "150.96"`, `isr: "150.960"`, golden.ErrInexactAmount},
		{"thousands separator", `cash: "30000.00"`, `cash: "30,000.00"`, golden.ErrInexactAmount},
		{"bad id", "id: PF-612-ISR-01", "id: pf 612", golden.ErrInvalidCase},
		{"no source", "source: research:PF-612-ISR-01\n", "", golden.ErrInvalidCase},
		{"bad source", "source: research:PF-612-ISR-01", "source: my-head", golden.ErrInvalidCase},
		{"blank skip reason", "pending_strategy: true", "skip_reason: \"  \"", golden.ErrInvalidCase},
		{"bad regime", `regime: "612"`, `regime: "61"`, golden.ErrInvalidCase},
		{"bad person type", "person_type: fisica", "person_type: pf", golden.ErrInvalidCase},
		{"bad period", `period: "2026-03"`, `period: "2026-13"`, golden.ErrInvalidCase},
		{"month after the period", "- month: 3", "- month: 4", golden.ErrInvalidCase},
		{"repeated month", "- month: 3", "- month: 1", golden.ErrInvalidCase},
		{"active month after the period", "  options:\n", "  active_months: [4]\n  options:\n", golden.ErrInvalidCase},
		{"option of a map", "deduction_mode: actual", "deduction_mode: {a: b}", golden.ErrInvalidCase},
		{"option list of non-strings", "[lodging, goods_services]", "[[a]]", golden.ErrInvalidCase},
		{"duplicate tax input", `rate: "0.15"}`, `rate: "0.15"}` + "\n    - {key: coeficiente_utilidad, kind: rate, rate: \"0.2\"}", golden.ErrInvalidCase},
		{"tax input with an unknown kind", "kind: boolean", "kind: flag", golden.ErrInvalidCase},
		{"tax input without its value", `kind: rate, rate: "0.15"`, "kind: rate", golden.ErrInvalidCase},
		{"tax input with two values", `kind: rate, rate: "0.15"`, `kind: rate, rate: "0.15", money: "1.00"`, golden.ErrInvalidCase},
		{"unknown status", "status: complete", "status: done", golden.ErrInvalidCase},
		{"no taxes", "taxes: [isr, iva]", "taxes: []", golden.ErrInvalidCase},
		{"unknown tax", "taxes: [isr, iva]", "taxes: [isr, iva, ieps]", golden.ErrInvalidCase},
		{"repeated tax", "taxes: [isr, iva]", "taxes: [isr, iva, isr]", golden.ErrInvalidCase},
		{"isr not asserted", "taxes: [isr, iva]", "taxes: [iva]", golden.ErrInvalidCase},
		{"withheld_to_remit not asserted", "  isr: \"150.96\"\n", "  isr: \"150.96\"\n  withheld_to_remit: \"1.00\"\n", golden.ErrInvalidCase},
		{"iva net disagrees with its components", `net: "2133.33"`, `net: "2133.34"`, golden.ErrInvalidCase},
		{"no lines", "  lines:\n    - {tax: isr, code: isr.income_ytd, amount: \"90000.00\"}\n    - {tax: iva, code: iva.net, amount: \"2133.33\"}\n", "", golden.ErrInvalidCase},
		{"line of an unasserted tax", "{tax: iva, code: iva.net", "{tax: withholding, code: iva.net", golden.ErrInvalidCase},
		{"line without a code", "code: iva.net", "code: \"\"", golden.ErrInvalidCase},
		{"unknown disclaimer", "[deductibility_claimed]", "[maybe]", golden.ErrInvalidCase},
		{"unknown cpa topic", "  disclaimers:", "  cpa_topics: [vibes]\n  disclaimers:", golden.ErrInvalidCase},
		{"missing input without months range", "  disclaimers:", "  missing_inputs: [{key: ptu_paid, months: [13]}]\n  disclaimers:", golden.ErrInvalidCase},
		{"second document", "  disclaimers: [deductibility_claimed]\n", "  disclaimers: [deductibility_claimed]\n---\nid: x\n", golden.ErrSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := strings.Replace(validCase, tt.old, tt.new, 1)
			require.NotEqual(t, validCase, doc, "the edit must apply")

			_, err := decode(t, doc)

			require.ErrorIs(t, err, tt.want)
			assert.Contains(t, err.Error(), "612-pf/PF-612-ISR-01.yaml", "errors name the file")
		})
	}
}
