package taxengine_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// regimeRESICO is excluded from the global cap's total income (OQ-A4).
const regimeRESICO = fiscal.RegimeCode("626")

// researchLimits2026 are the deduction limits of the research
// (docs/research/annual-return-2026.md §B.2). ANN-04 puts them in the
// parameter file; this test then reads them from there.
type researchLimits2026 struct{ t *testing.T }

func (r researchLimits2026) DeductionLimits(int) (annual.Limits, error) {
	return annual.Limits{
		AnnualUMA:     mustCents(r.t, "42794.64"),
		GlobalCapUMAs: 5, GlobalCapIncomeShare: mustRate(r.t, "0.15"),
		FuneralCapUMAs: 1,
		DonationShare:  mustRate(r.t, "0.07"), GovernmentDonationShare: mustRate(r.t, "0.04"),
		RetirementIncomeShare: mustRate(r.t, "0.10"), RetirementCapUMAs: 5,
		TuitionCaps: map[annual.TuitionLevel]money.Cents{
			annual.LevelPreescolar: mustCents(r.t, "14200.00"), annual.LevelPrimaria: mustCents(r.t, "12900.00"),
			annual.LevelSecundaria: mustCents(r.t, "19900.00"), annual.LevelProfesionalTecnico: mustCents(r.t, "17100.00"),
			annual.LevelBachillerato: mustCents(r.t, "24500.00"),
		},
	}, nil
}

func mustRate(t *testing.T, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

// totalIncome is the global cap's "ingresos totales" (MODULE-annual §2.7):
// gross income of the art. 152 chapters, salary accumulable + exempt, and
// other exempt income; RESICO income excluded.
func totalIncome(in annualgolden.Input) money.Cents {
	total := in.AnnualInputs.OtherExemptIncome
	for _, r := range in.Regimes {
		if r.Regime != regimeRESICO {
			total += r.Income
		}
	}
	for _, s := range in.SalaryStatements {
		total += s.Accumulable + s.Exempt
	}
	return total
}

// TestGoldenAnnualDeductionsMatchTheCalculator checks, before the
// pf_art152 strategy exists, that every PF case's expected personal
// deductions and its deductions-note lines are what the ANN-07 calculator
// computes from the case's own inputs: an authoring slip in a seeded
// golden fails now, not when ANN-08 lands.
func TestGoldenAnnualDeductionsMatchTheCalculator(t *testing.T) {
	calc := annual.NewCalculator(researchLimits2026{t})
	checked := 0
	for _, c := range annualCases(t) {
		for _, s := range c.Want.Sections {
			if s.Kind != annual.SectionPFArt152 || s.Tax == nil || s.Tax.PersonalDeductions == nil || s.Tax.AccumulableIncome == nil {
				continue
			}
			t.Run(c.ID, func(t *testing.T) {
				got, err := calc.Compute(annual.DeductionsInput{
					FiscalYear: c.Input.FiscalYear, Deductions: c.Input.PersonalDeductions,
					AccumulableIncome: *s.Tax.AccumulableIncome, TotalIncome: totalIncome(c.Input),
					PriorYearAccumulableIncome: c.Input.AnnualInputs.PriorYearAccumulableIncome,
				})

				require.NoError(t, err)
				assert.Equal(t, *s.Tax.PersonalDeductions, got.Applied)
				for _, want := range s.Lines {
					if !slices.Contains([]annual.NoteKind{annual.NoteCapApplied, annual.NoteExcludedCash, annual.NoteExcludedUnverified}, annual.NoteKind(want.Code)) {
						continue
					}
					found := slices.ContainsFunc(got.Notes, func(n annual.Note) bool {
						return string(n.Kind) == string(want.Code) && n.To == want.Amount
					})
					assert.True(t, found, "%s: no %s note to %s in %+v", c.ID, want.Code, want.Amount, got.Notes)
				}
			})
			checked++
		}
	}
	assert.GreaterOrEqual(t, checked, 5, "AN-01, AN-02, AN-04, AN-05 and AN-08 state their personal deductions")
}
