package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Registry maps each StrategyKey to its strategy (ADR-0002 §1). It is built
// once, at the composition root, and never changes: there is no Register
// method and no package-level registry. The zero value has no strategies, so
// every regime is informational.
type Registry struct {
	strategies map[StrategyKey]Strategy
}

// NewRegistry registers strategies by their keys. A nil strategy or a key
// with a blank regime code or an unknown person type fails with
// ErrInvalidStrategy; two strategies under one key fail with
// ErrDuplicateStrategy, so a wiring mistake stops the process at startup.
func NewRegistry(strategies ...Strategy) (Registry, error) {
	byKey := make(map[StrategyKey]Strategy, len(strategies))
	for i, s := range strategies {
		if err := checkStrategy(s); err != nil {
			return Registry{}, fmt.Errorf("strategy %d: %w", i, err)
		}
		key := s.Key()
		if _, seen := byKey[key]; seen {
			return Registry{}, fmt.Errorf("%s: %w", key, ErrDuplicateStrategy)
		}
		byKey[key] = s
	}
	return Registry{strategies: byKey}, nil
}

func checkStrategy(s Strategy) error {
	if s == nil {
		return fmt.Errorf("nil: %w", ErrInvalidStrategy)
	}
	key := s.Key()
	if strings.TrimSpace(string(key.Regime)) == "" || !key.PersonType.Valid() {
		return fmt.Errorf("key %q: %w", key, ErrInvalidStrategy)
	}
	return nil
}

// Lookup returns key's strategy and true, or the InformationalStrategy of
// key and false when none is registered: a catalog regime always has one.
func (r Registry) Lookup(key StrategyKey) (Strategy, bool) {
	if s, ok := r.strategies[key]; ok {
		return s, true
	}
	return NewInformationalStrategy(key), false
}

// Estimate runs in.Key's strategy (or the informational fallback) for the
// period. Before running it, it checks in.Inputs against the strategy's
// RequiredInputs for the period, given the months the regime was active;
// whatever is absent goes to MissingInputs, which makes the status
// incomplete. The strategy still computes, with zero for each missing input
// (ADR-0002 §2, MODULE-tax-engine §2.2), so the user sees a labelled figure.
func (r Registry) Estimate(in Input, activeMonths []time.Month, p *taxparams.Params) (RegimeEstimate, error) {
	s, _ := r.Lookup(in.Key)
	facts := Facts{ActiveMonths: activeMonths, Inputs: in.Inputs}
	missing, err := MissingInputs(s.RequiredInputs(in.Options, in.Period, facts), in.Inputs)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", in.Key, in.Period, err)
	}
	est, err := s.Estimate(in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("%s %s: %w", in.Key, in.Period, err)
	}
	if est.Key != in.Key {
		return RegimeEstimate{}, fmt.Errorf("estimate %s for input %s: %w", est.Key, in.Key, ErrStrategyMismatch)
	}
	est.MissingInputs = missing
	return est, nil
}
