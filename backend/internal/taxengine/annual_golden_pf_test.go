package taxengine_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// pfAnnualSections is the gate's annual engine for personas físicas: the
// pf_art152 (ANN-08) and pf_resico_check (ANN-09) strategies on the
// embedded parameter sets, the deduction limits read through the ANN-04
// adapter. ANN-14 replaces it with the composition root's estimator; until
// then the other section kinds stay unregistered (pending_strategy).
type pfAnnualSections struct {
	catalog *taxparams.Catalog
	art152  annual.PFArt152
	resico  annual.PFRESICOCheck
}

func newPFAnnualSections(t *testing.T) pfAnnualSections {
	t.Helper()
	catalog, err := params.Embedded()
	require.NoError(t, err)
	return pfAnnualSections{
		catalog: catalog,
		art152:  annual.NewPFArt152(annual.NewCalculator(annuallimits.New(catalog))),
		resico:  annual.NewPFRESICOCheck(),
	}
}

func (e pfAnnualSections) Registered(k annual.SectionKind) bool {
	return k == e.art152.Kind() || k == e.resico.Kind()
}

var errPMNotRegistered = errors.New("no persona moral section is registered")

// Estimate runs every registered section that applies, in the fixed
// section order (MODULE-annual §2.2).
func (e pfAnnualSections) Estimate(in annualgolden.Input) (annualgolden.Result, error) {
	if in.PersonType != fiscal.PersonFisica {
		return annualgolden.Result{}, errPMNotRegistered
	}
	sel, err := e.catalog.Select(in.FiscalYear)
	if err != nil {
		return annualgolden.Result{}, err
	}
	year, err := pfYearOf(in)
	if err != nil {
		return annualgolden.Result{}, err
	}
	payments := goldenPayments(in.Payments)
	var out annualgolden.Result
	if e.art152.Applies(year) {
		got, err := e.art152.Compute(year, payments, sel.Params)
		if err != nil {
			return annualgolden.Result{}, err
		}
		out.InformativeIncomeOver500k = got.InformativeIncomeOver500k
		out = withSection(out, got.Section)
	}
	if e.resico.Applies(year) {
		got, err := e.resico.Compute(year, payments, sel.Params)
		if err != nil {
			return annualgolden.Result{}, err
		}
		out = withSection(out, got)
	}
	return out, nil
}

// withSection appends s in the harness view and merges its disclaimers.
func withSection(r annualgolden.Result, s annual.Section) annualgolden.Result {
	got := annualgolden.SectionResult{
		Kind: s.Kind, Regimes: s.Regimes, Status: s.Status, Obligation: s.Obligation,
		MissingInputs: s.MissingInputs, CPATopics: s.CPATopics, Difference: s.Difference,
	}
	if s.Tax != nil {
		got.Tax = &annualgolden.TaxFigures{
			AccumulableIncome: s.Tax.AccumulableIncome, PersonalDeductions: s.Tax.PersonalDeductions,
			Base: s.Tax.Base, ISR: s.Tax.ISR, Payments: s.Tax.Payments, Withholdings: s.Tax.Withholdings,
			SalaryISR: s.Tax.SalaryISR,
			Balance:   annualgolden.Balance{Kind: s.Tax.Balance.Kind, Amount: s.Tax.Balance.Amount},
		}
	}
	for _, l := range s.Lines {
		got.Lines = append(got.Lines, annualgolden.ResultLine{
			Line:  annualgolden.Line{Code: l.Code, Op: l.Operation, Amount: l.Amount},
			Cited: len(l.Basis) > 0,
		})
	}
	r.Sections = append(r.Sections, got)
	r.Disclaimers = slices.Compact(slices.Sorted(slices.Values(slices.Concat(r.Disclaimers, s.Disclaimers))))
	return r
}

// goldenPayments is the payments port over a case's resolved payments.
type goldenPayments []annualgolden.Payment

func (g goldenPayments) ProvisionalPayments(regime fiscal.RegimeCode) ([]annual.Payment, error) {
	var out []annual.Payment
	for _, p := range g {
		if p.Regime == regime {
			out = append(out, annual.Payment{Month: p.Month, Amount: p.Amount, Source: annual.PaymentSource(p.Source)})
		}
	}
	return out, nil
}

// The money tax inputs a PF regime's year carries in a golden.
const (
	goldenDepreciation    = "annual_depreciation"
	goldenPTUPaid         = string(domain.InputPTUPaid)
	goldenPriorYearLosses = string(domain.InputPriorYearLosses)
)

var errGoldenTaxInput = errors.New("tax input not read by a PF section")

// pfYearOf maps a case's year to the strategies' input.
func pfYearOf(in annualgolden.Input) (annual.PFYear, error) {
	year := annual.PFYear{
		FiscalYear: in.FiscalYear, Deductions: in.PersonalDeductions,
		Inputs: annual.PFAnnualInputs{
			PriorYearAccumulableIncome: in.AnnualInputs.PriorYearAccumulableIncome,
			OtherExemptIncome:          in.AnnualInputs.OtherExemptIncome,
			ToldEmployerWillFile:       in.AnnualInputs.ToldEmployerWillFile,
		},
	}
	for _, r := range in.Regimes {
		y, err := regimeYearOf(r)
		if err != nil {
			return annual.PFYear{}, err
		}
		year.Regimes = append(year.Regimes, y)
	}
	for _, s := range in.SalaryStatements {
		year.SalaryStatements = append(year.SalaryStatements, annual.SalaryStatement{
			Accumulable: s.Accumulable, Exempt: s.Exempt, ISRWithheld: s.ISRWithheld,
			Simultaneous: s.Simultaneous, LeftBeforeYearEnd: s.LeftBeforeYearEnd, PayerNotWithholding: s.PayerNotWithholding,
		})
	}
	return year, nil
}

func regimeYearOf(r annualgolden.RegimeInput) (annual.RegimeYear, error) {
	y := annual.RegimeYear{
		Regime: r.Regime, Income: r.Income, Deductions: r.Deductions, Predial: r.Predial, ISRWithheld: r.ISRWithheld,
	}
	for _, ti := range r.TaxInputs {
		target := map[string]*money.Cents{
			goldenDepreciation: &y.Depreciation, goldenPTUPaid: &y.PTUPaid, goldenPriorYearLosses: &y.PriorYearLosses,
		}[ti.Key]
		if target == nil {
			return annual.RegimeYear{}, fmt.Errorf("%s %s: %w", r.Regime, ti.Key, errGoldenTaxInput)
		}
		*target = ti.Money
	}
	options := map[domain.OptionKey]domain.OptionValue{}
	for k, v := range r.Options {
		key := domain.OptionKey(k)
		if key != annual.OptionPlatformDefinitive {
			options[key] = domain.TextOption(v)
			continue
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			return annual.RegimeYear{}, fmt.Errorf("%s %s: %w", r.Regime, k, err)
		}
		options[key] = domain.BoolOption(b)
	}
	y.Options = domain.NewOptions(options)
	return y, nil
}
