package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// ErrRegimeNotInPeriod reports that the taxpayer does not hold the regime in
// the breakdown's month (422 regime_not_active on regime).
var ErrRegimeNotInPeriod = errors.New("taxengine: regime not held in the period")

// fieldRegime is the breakdown's regime query parameter.
const fieldRegime = "regime"

// The breakdown's regime_not_active messages (contract example).
const (
	msgRegimeNotInPeriod      = "The taxpayer does not hold this regime in that period."
	msgRegimeNotInPeriodField = "No tienes este régimen en ese mes."
)

// RegimeBreakdown is the computation trace of one regime in one month
// (contract Breakdown, MODULE-tax-engine §2.5): the regime's estimate, its
// lines in display order, and the disclaimers that bear on them.
type RegimeBreakdown struct {
	TaxpayerID   string
	PersonType   fiscal.PersonType
	Period       domain.Period
	Regime       RegimeResult
	ParameterSet taxparams.SetID
	// Lines are the regime's ISR and IVA lines, the consolidated IVA's when
	// the regime has IVA, then its withholdings to remit; Seq is 1..n.
	Lines                 []domain.Line
	Legend                string
	Disclaimers           []domain.Disclaimer
	ProvisionalParameters bool
	EngineVersion         string
}

// Breakdown returns the trace of regimeCode in rawPeriod for the caller's
// taxpayer (getEstimateBreakdown). It is the month's estimate seen from one
// regime, so the two never disagree: the estimate's refusals come first
// (404 taxpayer, 400 period, 422 parameters_unavailable), then a regime not
// held in that month is 422 regime_not_active.
func (uc *Estimator) Breakdown(ctx context.Context, userID, taxpayerID, rawPeriod, regimeCode string) (RegimeBreakdown, error) {
	est, err := uc.EstimateMonth(ctx, userID, taxpayerID, rawPeriod)
	if err != nil {
		return RegimeBreakdown{}, err
	}
	i := slices.IndexFunc(est.Regimes, func(r RegimeResult) bool { return r.Key.Regime == fiscal.RegimeCode(regimeCode) })
	if i < 0 {
		return RegimeBreakdown{}, regimeNotInPeriodError()
	}
	regime := est.Regimes[i]
	shown := domain.ConsolidatedIVA{}
	if regime.IVA != nil {
		shown = est.IVA
	}
	return RegimeBreakdown{
		TaxpayerID:   est.TaxpayerID,
		PersonType:   est.PersonType,
		Period:       est.Period,
		Regime:       regime,
		ParameterSet: est.ParameterSet,
		Lines:        breakdownLines(regime.Lines, shown.Lines),
		Legend:       est.Legend,
		Disclaimers: domain.EstimateDisclaimers(domain.DisclaimerFacts{
			Regimes: []domain.RegimeEstimate{regime.RegimeEstimate}, IVA: shown, ProvisionalParameters: est.ProvisionalParameters,
		}),
		ProvisionalParameters: est.ProvisionalParameters,
		EngineVersion:         est.EngineVersion,
	}, nil
}

// breakdownLines orders the regime's own lines (ISR, then its IVA
// components, in the strategy's order), the consolidated IVA lines, and
// the regime's withholdings to remit last, numbering them 1..n.
func breakdownLines(regime, consolidated []domain.Line) []domain.Line {
	out := make([]domain.Line, 0, len(regime)+len(consolidated))
	var remit []domain.Line
	for _, l := range regime {
		if l.Tax == domain.TaxWithholding {
			remit = append(remit, l)
			continue
		}
		out = append(out, l)
	}
	out = append(append(out, consolidated...), remit...)
	for i := range out {
		out[i].Seq = i + 1
	}
	return out
}

func regimeNotInPeriodError() error {
	return apperr.Wrap(ErrRegimeNotInPeriod, apperr.CodeBusinessRuleViolation, msgRegimeNotInPeriod).
		WithFields(apperr.FieldError{Field: fieldRegime, Code: DetailRegimeNotActive, Message: msgRegimeNotInPeriodField})
}
