// Package pm holds the annual sections of personas morales (MODULE-annual
// §2.7, PLAN ANN-10 and ANN-11): pm_601 (General) and pm_resico (RESICO),
// both the LISR art. 9 annual computation
//
//	ingresos − deducciones autorizadas − depreciación (typed)
//	= renta gravable before PTU (the PTU base of 601)
//	− PTU pagada = utilidad fiscal (the suggested coeficiente's numerator)
//	− pérdidas de ejercicios anteriores = resultado fiscal × 30 %
//	− pagos provisionales − retención de intereses = saldo.
//
// 601 reads accrued income and authorized deductions; RESICO PM the income
// collected and the deductions paid (arts. 206–215). Each strategy is pure:
// (Input, Params) only, every step appended to a domain.Ledger, so the
// breakdown is the computation and each line cites its legal basis.
//
// Placement (ADR-0009): the annual domain sub-tree of the tax engine,
// internal/taxengine/domain/annual; a package of their own keeps them apart
// from the PF sections written in parallel (PLAN PG-AN-S). The
// SectionStrategy registry is wired in ANN-14.
//
// Money is int64 centavos, Exact inside and one rounding per result; there
// is no float and no clock in this package.
package pm

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Error is a pm error; its values are constants matched with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

// ErrInvalidInput reports a negative figure, a payment with a negative
// amount or an unknown source, or no fiscal year.
const ErrInvalidInput Error = "annual pm: invalid input"

// InputAnnualDepreciation is the tax-input key of the year's depreciation,
// typed from the CPA's fixed-asset schedule (MODULE-annual §2.6, contract
// key annual_depreciation).
const InputAnnualDepreciation domain.InputKey = "annual_depreciation"

// Params is the parameter set the strategies read; *taxparams.Params
// implements it.
type Params interface {
	Rate(key taxparams.Key, on fiscal.Date) (money.Rate, error)
}

// Input is one PM regime's fiscal year. Income and Deductions are on the
// regime's basis (601: accrued acumulables and authorized deductions;
// RESICO: collected and paid). The rest are the M5 tax inputs and the
// payments already resolved by precedence (annual.ResolvePayments).
type Input struct {
	FiscalYear          int
	Income              money.Cents
	Deductions          money.Cents
	Depreciation        money.Cents // annual_depreciation, typed
	PTUPaid             money.Cents // ptu_paid
	PriorYearLosses     money.Cents // prior_year_losses, already inflation-updated
	InterestWithholding money.Cents // interest_withholding (LISR 54)
	Payments            annual.ResolvedPayments
}

// Section is one computed PM section.
type Section struct {
	Kind          annual.SectionKind
	Regime        fiscal.RegimeCode
	Status        annual.Status
	Obligation    annual.Obligation
	MissingInputs []string
	CPATopics     []annual.CPATopic
	Disclaimers   []annual.DisclaimerCode
	Tax           Tax
	Extras        Extras
	Lines         []domain.Line
}

// Tax are a section's figures, each rounded once to the centavo. Base is
// the resultado fiscal; Withholdings the interest withholding credited.
type Tax struct {
	AccumulableIncome money.Cents
	Base              money.Cents
	ISR               money.Cents
	Payments          money.Cents
	Withholdings      money.Cents
	Balance           Balance
}

// Balance is a section's balance: its sign as a kind and its amount, never
// negative.
type Balance struct {
	Kind   annual.BalanceKind
	Amount money.Cents
}

// balanceOf turns a signed balance into its kind and amount.
func balanceOf(v money.Cents) Balance {
	switch {
	case v > 0:
		return Balance{Kind: annual.BalanceACargo, Amount: v}
	case v < 0:
		return Balance{Kind: annual.BalanceAFavor, Amount: -v}
	}
	return Balance{Kind: annual.BalanceCero}
}

// Extras are the informative PM figures (MODULE-annual §2.1 pm_extras):
// the PTU base, the PTU to distribute (601 only, ⚠️ ptu_rate) and the
// suggested coeficiente de utilidad for next year (601 only, ⚠️
// cu_rounding). They never feed the balance.
type Extras struct {
	PTUBase         money.Cents
	PTUToDistribute money.Cents
	SuggestedCU     money.Rate
}
