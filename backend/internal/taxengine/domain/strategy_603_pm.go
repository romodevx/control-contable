package domain

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// nonProfitPMRegime is régimen 603, Personas Morales con Fines no
// Lucrativos (LISR Título III).
const nonProfitPMRegime fiscal.RegimeCode = "603"

// Strategy603PM is régimen 603 for a persona moral (research §B.1,
// MODULE-tax-engine §2.4). It is partial: the monthly IVA is estimated when
// its activities cause it; there is generally no provisional ISR (the
// entity pays ISR only on items such as non-deductible expenses or income
// outside its purpose, LISR arts. 79-80), so the ISR is explained by the
// catalog, not estimated.
type Strategy603PM struct{}

// NewStrategy603PM returns the 603 PM strategy.
func NewStrategy603PM() Strategy603PM { return Strategy603PM{} }

// Key is 603/moral.
func (Strategy603PM) Key() StrategyKey {
	return StrategyKey{Regime: nonProfitPMRegime, PersonType: fiscal.PersonMoral}
}

// Treatment is partial: IVA only.
func (Strategy603PM) Treatment() Treatment { return TreatmentPartial }

// RequiredInputs is empty: there is no provisional ISR to estimate.
func (Strategy603PM) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// Estimate computes the month's IVA and withholdings to remit; no ISR.
func (s Strategy603PM) Estimate(in Input, _ *taxparams.Params) (RegimeEstimate, error) {
	return partialPMEstimate(s.Key(), in)
}
