package domain

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// agapesPMRegime is régimen 622, Actividades Agrícolas, Ganaderas,
// Silvícolas y Pesqueras for personas morales (LISR art. 74).
const agapesPMRegime fiscal.RegimeCode = "622"

// Strategy622PM is régimen 622 AGAPES for a persona moral (research §B.1,
// MODULE-tax-engine §2.4). It is partial: the monthly IVA is estimated, the
// ISR is not. Its provisional ISR follows 601 plus the art. 74 exemption
// (20 UMA per partner, capped at 200) and the 30 % reduction, which need
// per-partner data the app does not collect yet (deferred, §8); the catalog
// explanation says so. No tax input is required in v1 (§2.8).
type Strategy622PM struct{}

// NewStrategy622PM returns the 622 PM strategy.
func NewStrategy622PM() Strategy622PM { return Strategy622PM{} }

// Key is 622/moral.
func (Strategy622PM) Key() StrategyKey {
	return StrategyKey{Regime: agapesPMRegime, PersonType: fiscal.PersonMoral}
}

// Treatment is partial: IVA only.
func (Strategy622PM) Treatment() Treatment { return TreatmentPartial }

// RequiredInputs is empty: the ISR is not estimated, so its inputs (kept in
// the catalog for later) are never missing.
func (Strategy622PM) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// Estimate computes the month's IVA and withholdings to remit; no ISR.
func (s Strategy622PM) Estimate(in Input, _ *taxparams.Params) (RegimeEstimate, error) {
	return partialPMEstimate(s.Key(), in)
}
