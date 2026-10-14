package domain

import "github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"

// regimeExplanationPrefix starts the label key of a regime's plain-language
// explanation, the catalog's regimes.explanation (text: M5-19).
const regimeExplanationPrefix = "regime_explanation:"

// RegimeExplanation is the label key of key's catalog explanation,
// "regime_explanation:<regime>/<person type>": the adapter renders it from
// the regimes catalog row of that (code, person type).
func RegimeExplanation(key StrategyKey) LabelKey {
	return LabelKey(regimeExplanationPrefix + key.String())
}

// InformationalStrategy is the fallback of every catalog regime without a
// registered strategy (ADR-0002 §1): no number, no required input, only the
// catalog's explanation. So every catalog regime works from day one.
type InformationalStrategy struct {
	key StrategyKey
}

// NewInformationalStrategy returns the informational strategy of key.
func NewInformationalStrategy(key StrategyKey) InformationalStrategy {
	return InformationalStrategy{key: key}
}

// Key returns the regime this fallback stands for.
func (s InformationalStrategy) Key() StrategyKey { return s.key }

// Treatment is always informational.
func (InformationalStrategy) Treatment() Treatment { return TreatmentInformational }

// RequiredInputs is empty: nothing is computed, so nothing is missing.
func (InformationalStrategy) RequiredInputs(Options, Period, Facts) []RequiredInput { return nil }

// Estimate returns the explanation and no figure: status informational.
func (s InformationalStrategy) Estimate(Input, *taxparams.Params) (RegimeEstimate, error) {
	return RegimeEstimate{
		Key:         s.key,
		Coverage:    TreatmentInformational,
		Explanation: RegimeExplanation(s.key),
	}, nil
}
