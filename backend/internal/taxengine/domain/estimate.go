package domain

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// TaxResult is one tax figure, rounded once, at the end, to the centavo,
// half away from zero (ADR-0002 §4).
type TaxResult struct {
	Exact  money.Exact
	Amount money.Cents // negative: saldo a favor
}

// NewTaxResult rounds an exact result: the one rounding step of a tax.
func NewTaxResult(exact money.Exact) TaxResult {
	return TaxResult{Exact: exact, Amount: exact.Round()}
}

// Outcome reports payable, in favor or zero, from the rounded amount.
func (r TaxResult) Outcome() Outcome {
	switch {
	case r.Amount > 0:
		return OutcomePayable
	case r.Amount < 0:
		return OutcomeInFavor
	}
	return OutcomeZero
}

// Payable returns the positive part of the amount: what is to pay. Whole-peso
// rounding (CFF 20) is off until the CPA confirms it (OQ-T3).
func (r TaxResult) Payable() money.Cents { return max(r.Amount, 0) }

// IVAComponents are one regime's contribution to the taxpayer's monthly IVA,
// netted across regimes by the use case (§2.6); never an amount on its own.
type IVAComponents struct {
	ChargedCollected     money.Cents
	CreditablePaid       money.Cents
	WithheldFromTaxpayer money.Cents // only the IVA recorded as withheld on incomes (OQ-T6 a)
	DefinitiveExcluded   money.Cents // 625 definitive IVA (LIVA 18-M ⚠️); 0 otherwise
}

// MissingInput is a tax input the period needs and the taxpayer has not
// entered; Months lists the missing months of a money_monthly key.
type MissingInput struct {
	Key    InputKey
	Months []time.Month
}

// RegimeEstimate is one regime's estimate for the period (§2.1). A strategy
// fills the computation; the use case adds what it alone knows (the missing
// inputs from RequiredInputs, the parameter set, the consolidated IVA).
type RegimeEstimate struct {
	Key StrategyKey
	// Coverage is what this estimate covers: full, partial (only some
	// taxes, Explanation says which) or informational (no number).
	Coverage        Treatment
	ISR             *TaxResult     // nil: no ISR figure this period
	IVA             *IVAComponents // nil: no IVA figure
	WithheldToRemit money.Cents    // withheld from suppliers, to remit (PM)
	Lines           []Line
	MissingInputs   []MissingInput
	Explanation     LabelKey // partial or informational text; "" otherwise
	CPATopics       []CPATopic
	Disclaimers     []DisclaimerCode // regime-specific (deductibility_claimed, ...)
}

// Status applies the contract's precedence: informational > incomplete >
// partial > complete (ADR-0002 Amendment 1).
func (e RegimeEstimate) Status() Status {
	switch {
	case e.Coverage == TreatmentInformational:
		return StatusInformational
	case len(e.MissingInputs) > 0:
		return StatusIncomplete
	case e.Coverage == TreatmentPartial:
		return StatusPartial
	}
	return StatusComplete
}

// CPAReviewPending reports a computation using a rule no CPA has validated.
func (e RegimeEstimate) CPAReviewPending() bool { return len(e.CPATopics) > 0 }
