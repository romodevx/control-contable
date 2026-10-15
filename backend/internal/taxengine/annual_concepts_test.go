package taxengine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// conceptCases reach the annual lines the goldens do not: a cash
// exclusion, unverifiable donations, a 612 loss, the 606 actual mode, a
// base of zero, RESICO PF above the table, a PM loss and a refund of
// estimated payments.
func conceptCases() []annualgolden.Input {
	return []annualgolden.Input{
		{FiscalYear: 2026, PersonType: fiscal.PersonFisica,
			Regimes: []annualgolden.RegimeInput{
				{Regime: "612", Income: 100_00, Deductions: 900_00},
				{Regime: "606", Income: 1_000_00, Deductions: 2_000_00, Options: map[string]string{string(domain.OptionRentDeductionMode): "actual"}},
			},
			PersonalDeductions: []annual.Deduction{
				{Category: annual.CategoryMedical, Amount: 1_00, PaymentMethod: annual.PaymentCash, Beneficiary: annual.BeneficiarySelf},
				{Category: annual.CategoryDonation, Amount: 1_00, PaymentMethod: annual.PaymentElectronic, Beneficiary: annual.BeneficiarySelf, DoneeType: annual.DoneeAuthorized},
			},
		},
		{FiscalYear: 2026, PersonType: fiscal.PersonFisica,
			Regimes: []annualgolden.RegimeInput{{Regime: "626", Income: 4_000_000_00}}},
		{FiscalYear: 2026, PersonType: fiscal.PersonMoral,
			Regimes:  []annualgolden.RegimeInput{{Regime: "601", Income: 100_00, Deductions: 900_00}},
			Payments: []annualgolden.Payment{{Regime: "601", Amount: 50_000_00, Source: annualgolden.PaymentEstimatedByApp}}},
		{FiscalYear: 2026, PersonType: fiscal.PersonMoral,
			Regimes:  []annualgolden.RegimeInput{{Regime: "626", Income: 1_000_00, Deductions: 100_00}},
			Payments: []annualgolden.Payment{{Regime: "626", Amount: 10_00, Source: annualgolden.PaymentTaxInput}}},
	}
}

// Every line the annual sections can append has a Spanish concept in the
// adapter (contract AnnualBreakdownLine.concept): the breakdown never shows
// a key or the generic fallback. It runs every annual golden and
// conceptCases through the use case.
func TestEveryAnnualBreakdownLineHasAConcept(t *testing.T) {
	uc := newUseCaseAnnual(t)
	inputs := conceptCases()
	for _, c := range annualCases(t) {
		inputs = append(inputs, c.Input)
	}
	labels := map[domain.LabelKey]bool{}
	for _, in := range inputs {
		est, err := uc.estimate(in)
		require.NoError(t, err)
		for _, s := range est.Sections {
			for _, l := range s.Lines {
				labels[l.Label] = true
			}
		}
	}

	for label := range labels {
		assert.Truef(t, annualhttp.HasConcept(label), "annual label %q has no concept in adapter/http/annual/texts.go", label)
	}
	for _, must := range []domain.LabelKey{
		"annual_deduction_excluded_cash", "annual_deduction_excluded_unverified", "annual_business_loss",
		"annual_rent_deductions", "annual_rent_loss", "annual_no_base", "annual_resico_above_table", "isr.no_profit",
		"isr.refund_only_paid", "isr.payments_tax_input", "annual_deduction_cap_applied",
	} {
		assert.Truef(t, labels[must], "no case reaches %q: the coverage is not exhaustive", must)
	}
	t.Logf("%d distinct annual labels rendered", len(labels))
}
