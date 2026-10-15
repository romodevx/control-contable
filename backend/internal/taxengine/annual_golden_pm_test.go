package taxengine_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/payments"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The PM sections of the annual golden gate (PLAN ANN-10, ANN-11): pm_601
// and pm_resico run on the embedded parameter catalog, each case's regime
// figures mapped to pm.Input and each section to the harness's view.

// errPMGoldenInput reports a case field the PM sections do not read: a
// golden that sets one would otherwise pass while testing nothing.
var errPMGoldenInput = errors.New("annual golden: field not read by the PM sections")

// pmSectionComputer is pm.General or pm.RESICO.
type pmSectionComputer interface {
	Compute(pm.Input, pm.Params) (pm.Section, error)
}

func pmGateSections(t *testing.T) annualSections {
	t.Helper()
	cat := catalog(t)
	return annualSections{
		annual.SectionPM601:    pmGateSection(cat, "601", pm.NewGeneral()),
		annual.SectionPMRESICO: pmGateSection(cat, fiscal.RegimeRESICO, pm.NewRESICO()),
	}
}

// pmGateSection computes the section of each regime held with code by a
// persona moral, on the parameter set the catalog selects for the year.
func pmGateSection(cat *taxparams.Catalog, code fiscal.RegimeCode, s pmSectionComputer) annualSection {
	return func(in annualgolden.Input) ([]annualgolden.SectionResult, error) {
		if in.PersonType != fiscal.PersonMoral {
			return nil, nil
		}
		var out []annualgolden.SectionResult
		for _, r := range in.Regimes {
			if r.Regime != code {
				continue
			}
			sel, err := cat.Select(in.FiscalYear)
			if err != nil {
				return nil, err
			}
			pin, err := pmInput(in, r)
			if err != nil {
				return nil, err
			}
			sec, err := s.Compute(pin, sel.Params)
			if err != nil {
				return nil, err
			}
			out = append(out, pmSectionResult(sec))
		}
		return out, nil
	}
}

// pmTaxInputs are where each money tax input of a PM regime goes.
var pmTaxInputs = map[string]func(*pm.Input) *money.Cents{
	string(domain.InputPTUPaid):             func(in *pm.Input) *money.Cents { return &in.PTUPaid },
	string(domain.InputPriorYearLosses):     func(in *pm.Input) *money.Cents { return &in.PriorYearLosses },
	string(domain.InputInterestWithholding): func(in *pm.Input) *money.Cents { return &in.InterestWithholding },
	string(pm.InputAnnualDepreciation):      func(in *pm.Input) *money.Cents { return &in.Depreciation },
}

func pmInput(in annualgolden.Input, r annualgolden.RegimeInput) (pm.Input, error) {
	if r.Predial != 0 || r.ISRWithheld != 0 || len(r.Options) > 0 {
		return pm.Input{}, fmt.Errorf("regime %s predial, isr_withheld or options: %w", r.Regime, errPMGoldenInput)
	}
	out := pm.Input{FiscalYear: in.FiscalYear, Income: r.Income, Deductions: r.Deductions}
	for _, ti := range r.TaxInputs {
		field, ok := pmTaxInputs[ti.Key]
		if !ok {
			return pm.Input{}, fmt.Errorf("tax input %s: %w", ti.Key, errPMGoldenInput)
		}
		*field(&out) = ti.Money
	}
	for _, p := range in.Payments {
		if p.Regime == r.Regime {
			out.Payments.Payments = append(out.Payments.Payments, payments.Payment{
				Month: time.Month(p.Month), Amount: p.Amount, Source: payments.Source(p.Source),
			})
		}
	}
	return out, nil
}

func pmSectionResult(s pm.Section) annualgolden.SectionResult {
	lines := make([]annualgolden.ResultLine, len(s.Lines))
	for i, l := range s.Lines {
		lines[i] = annualgolden.ResultLine{
			Line:  annualgolden.Line{Code: l.Code, Op: l.Operation, Amount: l.Amount},
			Cited: len(l.Basis) > 0,
		}
	}
	return annualgolden.SectionResult{
		Kind: s.Kind, Regimes: []fiscal.RegimeCode{s.Regime}, Status: s.Status, Obligation: s.Obligation,
		MissingInputs: slices.Clone(s.MissingInputs), CPATopics: slices.Clone(s.CPATopics),
		Tax: &annualgolden.TaxFigures{
			AccumulableIncome: s.Tax.AccumulableIncome, Base: s.Tax.Base, ISR: s.Tax.ISR,
			Payments: s.Tax.Payments, Withholdings: s.Tax.Withholdings,
			Balance: annualgolden.Balance{Kind: s.Tax.Balance.Kind, Amount: s.Tax.Balance.Amount},
		},
		PMExtras: &annualgolden.PMExtrasResult{
			PTUBase: s.Extras.PTUBase, PTUToDistribute: s.Extras.PTUToDistribute, SuggestedCU: s.Extras.SuggestedCU,
		},
		Lines: lines,
	}
}

// TestPMGoldenInputRefusesFieldsItDoesNotRead: a PM golden with predial,
// isr_withheld, options or an unknown tax input fails instead of passing on
// figures the sections never saw.
func TestPMGoldenInputRefusesFieldsItDoesNotRead(t *testing.T) {
	for name, r := range map[string]annualgolden.RegimeInput{
		"predial":       {Regime: "601", Predial: 1},
		"isr withheld":  {Regime: "601", ISRWithheld: 1},
		"options":       {Regime: "601", Options: map[string]string{"x": "y"}},
		"unknown input": {Regime: "601", TaxInputs: []annualgolden.TaxInput{{Key: "coeficiente_utilidad"}}},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := pmInput(annualgolden.Input{FiscalYear: 2026}, r)

			require.ErrorIs(t, err, errPMGoldenInput)
		})
	}
}

// TestPMGateSectionsReadEveryInput maps a full PM regime and runs it, so a
// dropped tax input or payment shows up as a different figure.
func TestPMGateSectionsReadEveryInput(t *testing.T) {
	in := annualgolden.Input{
		FiscalYear: 2026, PersonType: fiscal.PersonMoral,
		Regimes: []annualgolden.RegimeInput{{
			Regime: "601", Income: 1_000_000_00, Deductions: 500_000_00,
			TaxInputs: []annualgolden.TaxInput{
				{Key: "annual_depreciation", Money: 100_000_00}, {Key: "ptu_paid", Money: 50_000_00},
				{Key: "prior_year_losses", Money: 10_000_00}, {Key: "interest_withholding", Money: 1_000_00},
			},
		}},
		Payments: []annualgolden.Payment{
			{Regime: "601", Month: 3, Amount: 20_000_00, Source: annualgolden.PaymentEstimatedByApp},
			{Regime: "626", Amount: 1, Source: annualgolden.PaymentFiled},
		},
	}

	got, err := pmGateSections(t)[annual.SectionPM601](in)

	require.NoError(t, err)
	require.Len(t, got, 1)
	// 1,000,000 − 500,000 − 100,000 − 50,000 − 10,000 = 340,000 × 30 % =
	// 102,000 − 20,000 − 1,000.
	assert.Equal(t, money.Cents(340_000_00), got[0].Tax.Base)
	assert.Equal(t, annualgolden.Balance{Kind: annual.BalanceACargo, Amount: 81_000_00}, got[0].Tax.Balance)

	none, err := pmGateSections(t)[annual.SectionPMRESICO](in)
	require.NoError(t, err)
	assert.Empty(t, none, "no 626 regime held")
}
