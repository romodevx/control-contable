package annualgolden

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// Case is one annual golden: a taxpayer's year and what the annual
// estimate must return for it.
type Case struct {
	ID     string
	File   string // path under the annual golden root
	Source string // "research:AN-xx", "cpa:<ref>" or "spec:<ref>"
	// SkipReason names the unverified (⚠️) rule the expected values depend
	// on: the case is reported as skipped and never run.
	SkipReason string
	// PendingStrategy marks a case whose sections are not registered yet.
	// The runner fails a stale marker, so it cannot hide a case once its
	// sections land.
	PendingStrategy bool
	Input           Input
	Want            Expected
}

// Input is one taxpayer's fiscal year as the annual estimate reads it
// (MODULE-annual §2.6): per-regime annual figures, provisional payments,
// salary constancias, personal deductions and the annual inputs.
type Input struct {
	FiscalYear         int
	PersonType         fiscal.PersonType
	Regimes            []RegimeInput
	Payments           []Payment
	SalaryStatements   []SalaryStatement
	PersonalDeductions []annual.Deduction
	AnnualInputs       AnnualInputs
}

// RegimeInput is one regime's year: income on the regime's basis (collected
// for PF and RESICO, accrual for 601), authorized deductions, predial (606),
// ISR withheld from the taxpayer, regime options and money tax inputs
// (ptu_paid, prior_year_losses, annual_depreciation, interest_withholding).
type RegimeInput struct {
	Regime      fiscal.RegimeCode
	Income      money.Cents
	Deductions  money.Cents
	Predial     money.Cents
	ISRWithheld money.Cents
	Options     map[string]string
	TaxInputs   []TaxInput
}

// TaxInput is one money tax input of a regime.
type TaxInput struct {
	Key   string
	Money money.Cents
}

// PaymentSource is where a provisional payment comes from (MODULE-annual
// §2.5 precedence).
type PaymentSource string

// The payment sources.
const (
	PaymentFiled          PaymentSource = "filed"
	PaymentTaxInput       PaymentSource = "tax_input"
	PaymentEstimatedByApp PaymentSource = "estimated_by_app"
)

// Valid reports a spec payment source.
func (s PaymentSource) Valid() bool {
	return s == PaymentFiled || s == PaymentTaxInput || s == PaymentEstimatedByApp
}

// Payment is a provisional ISR payment of a regime. Month 0 is the year's
// total as the source states it (the research gives totals, not months).
type Payment struct {
	Regime fiscal.RegimeCode
	Month  int
	Amount money.Cents
	Source PaymentSource
}

// SalaryStatement is one employer's constancia (MODULE-annual §2.6).
type SalaryStatement struct {
	Employer            string
	Accumulable         money.Cents
	Exempt              money.Cents
	ISRWithheld         money.Cents
	Simultaneous        bool
	LeftBeforeYearEnd   bool
	PayerNotWithholding bool
}

// AnnualInputs are the taxpayer's per-year annual inputs.
type AnnualInputs struct {
	PriorYearAccumulableIncome *money.Cents
	OtherExemptIncome          money.Cents
	ToldEmployerWillFile       bool
}

// Expected is what a case asserts. Sections are the estimate's sections,
// in order; a nil list field or nil pointer is not asserted.
type Expected struct {
	InformativeIncomeOver500k *bool
	Disclaimers               []annual.DisclaimerCode // compared as a set
	Sections                  []SectionWant
}

// SectionWant is one expected section.
type SectionWant struct {
	Kind          annual.SectionKind
	Regimes       []fiscal.RegimeCode // compared as a set
	Status        annual.Status
	Obligation    annual.Obligation // "" is not asserted
	MissingInputs []string          // compared as a set
	CPATopics     []annual.CPATopic // compared as a set
	Tax           *TaxWant
	Difference    *money.Cents // pf_resico_check informative difference
	PMExtras      *PMExtras
	// Lines must appear in the section's breakdown in this relative order
	// (by code, operation and amount); other lines may sit between them.
	Lines []Line
}

// TaxWant is a section's expected figures; nil fields are not asserted.
type TaxWant struct {
	AccumulableIncome  *money.Cents
	PersonalDeductions *money.Cents
	Base               *money.Cents
	ISR                *money.Cents
	Payments           *money.Cents
	Withholdings       *money.Cents
	SalaryISR          *money.Cents
	Balance            *Balance
}

// Balance is a section's balance: its sign and its (non-negative) amount.
type Balance struct {
	Kind   annual.BalanceKind
	Amount money.Cents
}

// PMExtras are a PM section's expected extras; nil fields are not asserted.
type PMExtras struct {
	PTUBase         *money.Cents
	PTUToDistribute *money.Cents
	SuggestedCU     *money.Rate
}

// Line is one breakdown line as a golden sees it.
type Line struct {
	Code   domain.LineCode
	Op     domain.Operation
	Amount money.Cents
}

// Result is the annual estimate as the harness compares it. The engine's
// adapter (ANN-14) maps its estimate to this view, so the harness does not
// depend on the engine's types.
type Result struct {
	InformativeIncomeOver500k bool
	Disclaimers               []annual.DisclaimerCode
	Sections                  []SectionResult
}

// SectionResult is one computed section.
type SectionResult struct {
	Kind          annual.SectionKind
	Regimes       []fiscal.RegimeCode
	Status        annual.Status
	Obligation    annual.Obligation
	MissingInputs []string
	CPATopics     []annual.CPATopic
	Tax           *TaxFigures
	Difference    *money.Cents
	PMExtras      *PMExtrasResult
	Lines         []Line
}

// TaxFigures are a section's computed figures.
type TaxFigures struct {
	AccumulableIncome  money.Cents
	PersonalDeductions money.Cents
	Base               money.Cents
	ISR                money.Cents
	Payments           money.Cents
	Withholdings       money.Cents
	SalaryISR          money.Cents
	Balance            Balance
}

// PMExtrasResult are a PM section's computed extras.
type PMExtrasResult struct {
	PTUBase         money.Cents
	PTUToDistribute money.Cents
	SuggestedCU     money.Rate
}
