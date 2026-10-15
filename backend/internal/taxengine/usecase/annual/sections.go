package annual

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// ErrInvalidSectionStrategy reports a registry that cannot order its
// strategies: a nil one, an unknown or explanation kind, or a kind twice.
var ErrInvalidSectionStrategy = errors.New("annual: invalid section strategy")

// Year is one taxpayer's fiscal year as the section strategies read it:
// every regime held in the year (only the months in its window, through
// data_through, counted), the annual-only inputs and the provisional
// payments resolved by precedence.
type Year struct {
	FiscalYear int
	PersonType fiscal.PersonType
	// Regimes are in code order.
	Regimes []adomain.RegimeYear
	// InterestWithholding is the PM tax input interest_withholding per
	// regime (LISR 54).
	InterestWithholding map[fiscal.RegimeCode]money.Cents
	SalaryStatements    []adomain.SalaryStatement
	Deductions          []adomain.Deduction
	Inputs              adomain.PFAnnualInputs
	Payments            PaymentBook
}

// regime returns the year of code, if held.
func (y Year) regime(code fiscal.RegimeCode) (adomain.RegimeYear, bool) {
	i := slices.IndexFunc(y.Regimes, func(r adomain.RegimeYear) bool { return r.Regime == code })
	if i < 0 {
		return adomain.RegimeYear{}, false
	}
	return y.Regimes[i], true
}

// Tax are a computed section's figures, each rounded once.
// PersonalDeductions is nil for a persona moral.
type Tax struct {
	AccumulableIncome  money.Cents
	PersonalDeductions *money.Cents
	Base               money.Cents
	ISR                money.Cents
	Payments           money.Cents
	Withholdings       money.Cents
	SalaryISR          money.Cents
	Balance            adomain.Balance
}

// PMExtras are the informative PM figures (contract PmAnnualExtras): nil
// pointers are the fields RESICO PM does not have.
type PMExtras struct {
	PTUBase         money.Cents
	PTUToDistribute *money.Cents
	SuggestedCU     *money.Rate
}

// ComputedSection is one section a strategy computed, before the use case
// adds its deadline, texts, notes and reasons.
type ComputedSection struct {
	Kind    adomain.SectionKind
	Regimes []fiscal.RegimeCode
	Status  adomain.Status
	// Obligation is "" when the law does not settle it.
	Obligation adomain.Obligation
	// MissingInputs are the domain's names (salary_statement,
	// prior_year_accumulable_income, filed_payment:<regime>:<YYYY-MM>).
	MissingInputs []string
	CPATopics     []adomain.CPATopic
	Disclaimers   []adomain.DisclaimerCode
	Tax           *Tax
	PMExtras      *PMExtras
	RESICOCheck   *adomain.RESICOCheck
	Lines         []domain.Line
	// InformativeIncomeOver500k is the LISR 150 ¶3 fact pf_art152 decides.
	InformativeIncomeOver500k bool
}

// SectionStrategy computes the sections of one kind for a year: none when
// the year holds no regime of its kind. It is pure.
type SectionStrategy interface {
	Kind() adomain.SectionKind
	Compute(y Year, p *taxparams.Params) ([]ComputedSection, error)
}

// computedKinds are the kinds a strategy may compute, in the fixed order
// of an estimate (MODULE-annual §2.2); explanations are not computed.
var computedKinds = []adomain.SectionKind{
	adomain.SectionPFArt152, adomain.SectionPFRESICOCheck, adomain.SectionPM601, adomain.SectionPMRESICO,
}

// IsComputedKind reports a kind with a computation trace.
func IsComputedKind(k adomain.SectionKind) bool { return slices.Contains(computedKinds, k) }

// SectionRegistry runs the registered strategies in the fixed order. It
// is open for extension: a new section is a strategy and one registration
// line in the composition root (taxengine.AnnualSections).
type SectionRegistry struct{ strategies []SectionStrategy }

// NewSectionRegistry orders strategies by kind.
func NewSectionRegistry(strategies ...SectionStrategy) (SectionRegistry, error) {
	byKind := map[adomain.SectionKind]SectionStrategy{}
	for i, s := range strategies {
		if s == nil {
			return SectionRegistry{}, fmt.Errorf("strategy %d is nil: %w", i, ErrInvalidSectionStrategy)
		}
		k := s.Kind()
		if !IsComputedKind(k) {
			return SectionRegistry{}, fmt.Errorf("kind %q: %w", k, ErrInvalidSectionStrategy)
		}
		if _, dup := byKind[k]; dup {
			return SectionRegistry{}, fmt.Errorf("kind %q twice: %w", k, ErrInvalidSectionStrategy)
		}
		byKind[k] = s
	}
	var r SectionRegistry
	for _, k := range computedKinds {
		if s, ok := byKind[k]; ok {
			r.strategies = append(r.strategies, s)
		}
	}
	return r, nil
}

// Registered reports a kind with a strategy.
func (r SectionRegistry) Registered(k adomain.SectionKind) bool {
	return slices.ContainsFunc(r.strategies, func(s SectionStrategy) bool { return s.Kind() == k })
}

// Compute runs every strategy on y, in the fixed order.
func (r SectionRegistry) Compute(y Year, p *taxparams.Params) ([]ComputedSection, error) {
	var out []ComputedSection
	for _, s := range r.strategies {
		got, err := s.Compute(y, p)
		if err != nil {
			return nil, fmt.Errorf("annual: section %s: %w", s.Kind(), err)
		}
		out = append(out, got...)
	}
	return out, nil
}

// pfYear is the persona física's year as the PF strategies read it.
func (y Year) pfYear() adomain.PFYear {
	return adomain.PFYear{
		FiscalYear: y.FiscalYear, Regimes: y.Regimes, SalaryStatements: y.SalaryStatements,
		Deductions: y.Deductions, Inputs: y.Inputs,
	}
}

// fromDomain converts a PF strategy's section.
func fromDomain(s adomain.Section) ComputedSection {
	out := ComputedSection{
		Kind: s.Kind, Regimes: s.Regimes, Status: s.Status, Obligation: s.Obligation,
		MissingInputs: s.MissingInputs, CPATopics: s.CPATopics, Disclaimers: s.Disclaimers,
		RESICOCheck: s.RESICOCheck, Lines: s.Lines,
	}
	if s.Tax != nil {
		deductions := s.Tax.PersonalDeductions
		out.Tax = &Tax{
			AccumulableIncome: s.Tax.AccumulableIncome, PersonalDeductions: &deductions, Base: s.Tax.Base,
			ISR: s.Tax.ISR, Payments: s.Tax.Payments, Withholdings: s.Tax.Withholdings, SalaryISR: s.Tax.SalaryISR,
			Balance: s.Tax.Balance,
		}
	}
	return out
}

// pfArt152Section is the pf_art152 strategy (ANN-08).
type pfArt152Section struct{ s adomain.PFArt152 }

// NewPFArt152Section registers pf_art152.
func NewPFArt152Section(s adomain.PFArt152) SectionStrategy { return pfArt152Section{s: s} }

func (pfArt152Section) Kind() adomain.SectionKind { return adomain.SectionPFArt152 }

func (p pfArt152Section) Compute(y Year, params *taxparams.Params) ([]ComputedSection, error) {
	year := y.pfYear()
	if y.PersonType != fiscal.PersonFisica || !p.s.Applies(year) {
		return nil, nil
	}
	got, err := p.s.Compute(year, y.Payments, params)
	if err != nil {
		return nil, err
	}
	out := fromDomain(got.Section)
	out.InformativeIncomeOver500k = got.InformativeIncomeOver500k
	return []ComputedSection{out}, nil
}

// pfRESICOCheckSection is the pf_resico_check strategy (ANN-09).
type pfRESICOCheckSection struct{ s adomain.PFRESICOCheck }

// NewPFRESICOCheckSection registers pf_resico_check.
func NewPFRESICOCheckSection(s adomain.PFRESICOCheck) SectionStrategy {
	return pfRESICOCheckSection{s: s}
}

func (pfRESICOCheckSection) Kind() adomain.SectionKind { return adomain.SectionPFRESICOCheck }

func (p pfRESICOCheckSection) Compute(y Year, params *taxparams.Params) ([]ComputedSection, error) {
	year := y.pfYear()
	if y.PersonType != fiscal.PersonFisica || !p.s.Applies(year) {
		return nil, nil
	}
	got, err := p.s.Compute(year, y.Payments, params)
	if err != nil {
		return nil, err
	}
	return []ComputedSection{fromDomain(got)}, nil
}

// pmComputer is pm.General or pm.RESICO.
type pmComputer interface {
	Kind() adomain.SectionKind
	Compute(pm.Input, pm.Params) (pm.Section, error)
}

// pmSection computes a PM regime's section (ANN-10, ANN-11).
type pmSection struct {
	code     fiscal.RegimeCode
	computer pmComputer
	extras   func(pm.Extras) *PMExtras
}

// NewPM601Section registers pm_601 (régimen 601).
func NewPM601Section(g pm.General) SectionStrategy {
	return pmSection{code: "601", computer: g, extras: func(e pm.Extras) *PMExtras {
		ptu, cu := e.PTUToDistribute, e.SuggestedCU
		return &PMExtras{PTUBase: e.PTUBase, PTUToDistribute: &ptu, SuggestedCU: &cu}
	}}
}

// NewPMRESICOSection registers pm_resico (régimen 626 of a persona moral).
func NewPMRESICOSection(r pm.RESICO) SectionStrategy {
	return pmSection{code: fiscal.RegimeRESICO, computer: r, extras: func(e pm.Extras) *PMExtras {
		return &PMExtras{PTUBase: e.PTUBase}
	}}
}

func (p pmSection) Kind() adomain.SectionKind { return p.computer.Kind() }

func (p pmSection) Compute(y Year, params *taxparams.Params) ([]ComputedSection, error) {
	r, ok := y.regime(p.code)
	if y.PersonType != fiscal.PersonMoral || !ok {
		return nil, nil
	}
	got, err := p.computer.Compute(pm.Input{
		FiscalYear: y.FiscalYear, Income: r.Income, Deductions: r.Deductions, Depreciation: r.Depreciation,
		PTUPaid: r.PTUPaid, PriorYearLosses: r.PriorYearLosses, InterestWithholding: y.InterestWithholding[p.code],
		Payments: y.Payments[p.code].Resolved,
	}, params)
	if err != nil {
		return nil, err
	}
	return []ComputedSection{{
		Kind: got.Kind, Regimes: []fiscal.RegimeCode{got.Regime}, Status: got.Status, Obligation: got.Obligation,
		MissingInputs: got.MissingInputs, CPATopics: got.CPATopics, Disclaimers: got.Disclaimers,
		Tax: &Tax{
			AccumulableIncome: got.Tax.AccumulableIncome, Base: got.Tax.Base, ISR: got.Tax.ISR,
			Payments: got.Tax.Payments, Withholdings: got.Tax.Withholdings,
			Balance: adomain.Balance{Kind: got.Tax.Balance.Kind, Amount: got.Tax.Balance.Amount},
		},
		PMExtras: p.extras(got.Extras),
		Lines:    got.Lines,
	}}, nil
}

// regimeArgLines are the pf_art152 and pf_resico_check lines whose first
// argument is the regime they belong to.
var regimeArgLines = []domain.LineCode{
	adomain.LineArt152BusinessIncome, adomain.LineArt152BusinessDeductions, adomain.LineArt152Depreciation,
	adomain.LineArt152PTUPaid, adomain.LineArt152PriorYearLosses, adomain.LineArt152RentIncome,
	adomain.LineArt152RentBlind, adomain.LineArt152RentPredial, adomain.LineArt152RentDeductions,
	adomain.LineArt152RentResult, adomain.LineArt152Payments, adomain.LineArt152Withheld,
	adomain.LineRESICOCheckIncome, adomain.LineRESICOCheckPayments, adomain.LineRESICOCheckWithheld,
}

// salaryLines belong to the salary chapter (605).
var salaryLines = []domain.LineCode{adomain.LineArt152Salary, adomain.LineArt152SalaryISR}

// LineRegimes names the regime each line belongs to (contract
// AnnualBreakdownLine.regime_code; "" for a section-wide line): a PM
// section's lines are its regime's, a PF line its chapter's or payment's.
func LineRegimes(kind adomain.SectionKind, regimes []fiscal.RegimeCode, lines []domain.Line) []fiscal.RegimeCode {
	out := make([]fiscal.RegimeCode, len(lines))
	for i, l := range lines {
		switch {
		case (kind == adomain.SectionPM601 || kind == adomain.SectionPMRESICO) && len(regimes) == 1:
			out[i] = regimes[0]
		case slices.Contains(regimeArgLines, l.Code) && len(l.Args) > 0:
			out[i] = fiscal.RegimeCode(l.Args[0])
		case slices.Contains(salaryLines, l.Code):
			out[i] = adomain.RegimeSalary
		}
	}
	return out
}
