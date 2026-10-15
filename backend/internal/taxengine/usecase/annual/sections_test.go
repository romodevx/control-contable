package annual_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// stubSection computes one empty section of its kind.
type stubSection struct{ kind adomain.SectionKind }

func (s stubSection) Kind() adomain.SectionKind { return s.kind }

func (s stubSection) Compute(annual.Year, *taxparams.Params) ([]annual.ComputedSection, error) {
	return []annual.ComputedSection{{Kind: s.kind}}, nil
}

func TestNewSectionRegistry_refusesWhatItCannotOrder(t *testing.T) {
	for name, strategies := range map[string][]annual.SectionStrategy{
		"nil":         {nil},
		"unknown":     {stubSection{kind: "pf_other"}},
		"explanation": {stubSection{kind: adomain.SectionExplanation}},
		"twice":       {stubSection{kind: adomain.SectionPM601}, stubSection{kind: adomain.SectionPM601}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := annual.NewSectionRegistry(strategies...)

			require.ErrorIs(t, err, annual.ErrInvalidSectionStrategy)
		})
	}
}

// Sections come out in the fixed MODULE-annual §2.2 order whatever the
// registration order: a new section is a new strategy and one line.
func TestSectionRegistry_computesInTheFixedOrder(t *testing.T) {
	reg, err := annual.NewSectionRegistry(
		stubSection{kind: adomain.SectionPMRESICO}, stubSection{kind: adomain.SectionPFArt152}, stubSection{kind: adomain.SectionPM601},
	)
	require.NoError(t, err)

	got, err := reg.Compute(annual.Year{}, nil)

	require.NoError(t, err)
	kinds := make([]adomain.SectionKind, len(got))
	for i, s := range got {
		kinds[i] = s.Kind
	}
	assert.Equal(t, []adomain.SectionKind{adomain.SectionPFArt152, adomain.SectionPM601, adomain.SectionPMRESICO}, kinds)
	assert.True(t, reg.Registered(adomain.SectionPM601))
	assert.False(t, reg.Registered(adomain.SectionPFRESICOCheck))
}

func realSections(t *testing.T) (annual.SectionRegistry, *taxparams.Params) {
	t.Helper()
	cat, err := params.Embedded()
	require.NoError(t, err)
	sel, err := cat.Select(2026)
	require.NoError(t, err)
	reg, err := annual.NewSectionRegistry(
		annual.NewPFArt152Section(adomain.NewPFArt152(adomain.NewCalculator(annuallimits.New(cat)))),
		annual.NewPFRESICOCheckSection(adomain.NewPFRESICOCheck()),
		annual.NewPM601Section(pm.NewGeneral()),
		annual.NewPMRESICOSection(pm.NewRESICO()),
	)
	require.NoError(t, err)
	return reg, sel.Params
}

func filedYear(regime fiscal.RegimeCode, amount money.Cents) annual.PaymentBook {
	return annual.PaymentBook{regime: {Regime: regime, Resolved: adomain.ResolvedPayments{
		Payments: []adomain.Payment{{Month: 12, Amount: amount, Source: adomain.PaymentFiled}},
	}}}
}

// The PF strategy reads the year and the book: AN-01 (MODULE-annual §3).
func TestPFArt152Section_AN01(t *testing.T) {
	reg, p := realSections(t)
	year := annual.Year{
		FiscalYear: 2026, PersonType: fiscal.PersonFisica,
		Regimes:    []adomain.RegimeYear{{Regime: "612", Income: 900_000_00, Deductions: 400_000_00, ISRWithheld: 30_000_00}},
		Deductions: []adomain.Deduction{{Category: adomain.CategoryMedical, Amount: 40_000_00, PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiarySelf}},
		Payments:   filedYear("612", 55_773_86),
	}

	got, err := reg.Compute(year, p)

	require.NoError(t, err)
	require.Len(t, got, 1)
	s := got[0]
	assert.Equal(t, adomain.SectionPFArt152, s.Kind)
	assert.True(t, s.InformativeIncomeOver500k)
	assert.True(t, s.DecidesInformativeIncome)
	assert.False(t, s.OutsideHeadline)
	require.NotNil(t, s.Tax)
	assert.Equal(t, ptr(money.Cents(40_000_00)), s.Tax.PersonalDeductions)
	assert.Equal(t, money.Cents(76_365_86), s.Tax.ISR)
	assert.Equal(t, adomain.Balance{Kind: adomain.BalanceAFavor, Amount: 9_408_00}, s.Tax.Balance)
}

// A PM section per regime held: AN-06's balance through pm_601, with the
// extras; a persona física's 626 never reaches the PM strategies.
func TestPMSections_onlyForPersonasMorales(t *testing.T) {
	reg, p := realSections(t)
	year := annual.Year{
		FiscalYear: 2026, PersonType: fiscal.PersonMoral,
		Regimes:             []adomain.RegimeYear{{Regime: "601", Income: 10_000_000_00, Deductions: 8_500_000_00, PTUPaid: 120_000_00, PriorYearLosses: 200_000_00}},
		Payments:            filedYear("601", 330_000_00),
		InterestWithholding: map[fiscal.RegimeCode]money.Cents{"601": 2_000_00},
	}

	got, err := reg.Compute(year, p)

	require.NoError(t, err)
	require.Len(t, got, 1)
	s := got[0]
	assert.Equal(t, adomain.SectionPM601, s.Kind)
	assert.Equal(t, []fiscal.RegimeCode{"601"}, s.Regimes)
	assert.Nil(t, s.Tax.PersonalDeductions)
	assert.Equal(t, money.Cents(1_180_000_00), s.Tax.Base)
	assert.Equal(t, adomain.Balance{Kind: adomain.BalanceACargo, Amount: 22_000_00}, s.Tax.Balance)
	require.NotNil(t, s.PMExtras)
	assert.Equal(t, money.Cents(1_500_000_00), s.PMExtras.PTUBase)
	require.NotNil(t, s.PMExtras.SuggestedCU)
	cu, err := money.ParseRate("0.1380")
	require.NoError(t, err)
	assert.Equal(t, cu, *s.PMExtras.SuggestedCU)
	assert.Equal(t, []annual.NoteCode{annual.NoteCUSuggestion}, s.Notes, "601 carries the CU note")
	assert.Equal(t, []annual.ReasonCode{annual.ReasonPersonaMoral}, s.ObligationReasons)
	require.Len(t, s.LineRegimes, len(s.Lines))
	for _, r := range s.LineRegimes {
		assert.Equal(t, fiscal.RegimeCode("601"), r, "a PM section's lines are its regime's")
	}

	pf := annual.Year{FiscalYear: 2026, PersonType: fiscal.PersonFisica, Regimes: []adomain.RegimeYear{{Regime: fiscal.RegimeRESICO, Income: 1}}}
	got, err = reg.Compute(pf, p)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, adomain.SectionPFRESICOCheck, got[0].Kind)
	assert.Nil(t, got[0].Tax)
	assert.NotNil(t, got[0].RESICOCheck)
	assert.True(t, got[0].OutsideHeadline, "RESICO PF payments are definitive")
	assert.NotEmpty(t, got[0].Explanation)
}

// The lines a PM section appends name its regime; the PF lines name the
// chapter or payment they belong to.
func TestLineRegimes(t *testing.T) {
	pfLines := []domain.Line{
		{Code: adomain.LineArt152BusinessIncome, Args: []string{"612"}},
		{Code: adomain.LineArt152Payments, Args: []string{"606", "filed", "12"}},
		{Code: adomain.LineArt152Salary, Args: []string{"2"}},
		{Code: adomain.LineArt152Base},
	}

	got := annual.PFLineRegimes(pfLines)

	assert.Equal(t, []fiscal.RegimeCode{"612", "606", "605", ""}, got)
}
