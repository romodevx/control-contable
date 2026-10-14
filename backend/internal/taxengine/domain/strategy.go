package domain

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// StrategyKey identifies a strategy: a regime code is not unique across
// person types (626 is RESICO PF and RESICO PM), so the key is both
// (ADR-0002 §1).
type StrategyKey struct {
	Regime     fiscal.RegimeCode
	PersonType fiscal.PersonType
}

// String renders "<regime>/<person type>", e.g. "626/moral".
func (k StrategyKey) String() string { return string(k.Regime) + "/" + string(k.PersonType) }

// Input is everything a strategy reads for one regime and period, assembled
// by the use case from the read ports. Strategies never see raw records.
type Input struct {
	Key     StrategyKey
	Period  Period
	Options Options
	Sums    YearSums // this regime, January..Period.Month of the fiscal year
	Inputs  TaxInputs
}

// Facts are what RequiredInputs needs beyond the options and the period.
type Facts struct {
	ActiveMonths []time.Month // months of the fiscal year the regime was held, up to the period
	Inputs       TaxInputs    // the inputs entered so far (e.g. first_fiscal_year)
}

// RequiredInput is a tax input the period needs; Months lists the months a
// money_monthly key needs (prior payments: every active month before n).
type RequiredInput struct {
	Key    InputKey
	Kind   InputKind
	Months []time.Month
}

// Strategy computes one regime's monthly estimate (ADR-0002 §1, amended).
// It must be pure: no I/O, no clock, no randomness, no floats, no map-order
// dependence; it computes only through a Ledger, so its lines are its trace.
type Strategy interface {
	Key() StrategyKey
	// Treatment is the catalog-level support level of the regime.
	Treatment() Treatment
	// RequiredInputs are the inputs period p needs: they depend on the
	// period, not only on the regime (Amendment 1).
	RequiredInputs(o Options, p Period, f Facts) []RequiredInput
	// Estimate computes the period, with zero for any missing input.
	Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error)
}
