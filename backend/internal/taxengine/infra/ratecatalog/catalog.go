package ratecatalog

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// keyCashPaymentLimit is the cash payment above which an expense is not
// deductible (LISR 27 fr. III), as the params files spell it.
const keyCashPaymentLimit taxparams.Key = "deduction.cash_payment_limit"

// UMAPeriod is the UMA value a rule reads, named by its params key.
type UMAPeriod taxparams.Key

// The published UMA values. Each changes on February 1: January is still
// in force with the previous year's UMA.
const (
	UMADaily   UMAPeriod = "uma.daily"
	UMAMonthly UMAPeriod = "uma.monthly"
	UMAAnnual  UMAPeriod = "uma.annual"
)

// ivaTreatments is every contract IvaRate, in contract order.
var ivaTreatments = []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}

// Catalog is the RateCatalog adapter over the published parameter sets. K
// is the consumer's amount key type (the expenses domain's AmountKey). It
// is immutable and safe for concurrent use.
type Catalog[K ~string] struct {
	sets *taxparams.Catalog
}

// Resolved is a parameter value and the set it came from. Provisional
// means the value must carry the provisional_parameters disclaimer: the
// set is marked provisional, or it is the previous year's because the
// date's fiscal year is not published yet.
type Resolved[T any] struct {
	Value       T
	Set         taxparams.SetID
	Provisional bool
}

// requiredIVARates are the IVA rates every parameter set must hold. The
// border 8 % is the one optional rate: it exists only while a decree does.
var requiredIVARates = []taxparams.Requirement{
	taxparams.RequireRate(taxparams.Key(fiscal.KeyIVAGeneral)),
	taxparams.RequireRate(taxparams.Key(fiscal.KeyIVAZero)),
}

// New returns the adapter over sets: ErrNoCatalog for a nil catalog,
// ErrMissingIVARate when a fiscal year's set (the version Select serves)
// lacks a required IVA rate, and ErrMissingWithholdingRate when it lacks a
// withholding rate, so a broken params file fails the boot.
func New[K ~string](sets *taxparams.Catalog) (*Catalog[K], error) {
	if sets == nil {
		return nil, ErrNoCatalog
	}
	if err := requireRates(sets); err != nil {
		return nil, err
	}
	return &Catalog[K]{sets: sets}, nil
}

// requireRates checks every set Select can serve: the highest version of
// each fiscal year.
func requireRates(sets *taxparams.Catalog) error {
	withholding := requiredWithholdingRates()
	var problems []error
	for _, id := range sets.Sets() {
		sel, err := sets.Select(id.FiscalYear)
		if err != nil {
			return fmt.Errorf("ratecatalog: select %d: %w", id.FiscalYear, err)
		}
		if sel.Params.Version() != id.Version {
			continue // an older version Select never serves
		}
		if err := sel.Params.Require(requiredIVARates...); err != nil {
			problems = append(problems, fmt.Errorf("%w: set %s: %w", ErrMissingIVARate, id, err))
		}
		if err := sel.Params.Require(withholding...); err != nil {
			problems = append(problems, fmt.Errorf("%w: set %s: %w", ErrMissingWithholdingRate, id, err))
		}
	}
	return errors.Join(problems...)
}

// Provisional reports whether the values a record dated on reads come from
// a provisional set (marked so, or the previous year's fallback), so the
// record can carry the provisional_parameters warning. A date no set covers
// is ErrDateNotCovered.
func (c *Catalog[K]) Provisional(on fiscal.Date) (bool, error) {
	r, err := resolve(c.sets, on, func(*taxparams.Params) (struct{}, error) { return struct{}{}, nil })
	return r.Provisional, err
}

// Rate returns the rate key holds on day on, and false when none does: the
// key is not in force that day, no set covers the day, or the key is not a
// rate. It implements fiscal.RateCatalog.
func (c *Catalog[K]) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	r, err := c.RateOn(key, on)
	return r.Value, err == nil
}

// Amount returns the money parameter key holds on day on, and false when
// none does. It implements the expenses module's RateCatalog.
func (c *Catalog[K]) Amount(key K, on fiscal.Date) (money.Cents, bool) {
	a, err := c.AmountOn(key, on)
	return a.Value, err == nil
}

// RateOn returns the rate key holds on day on: ErrDateNotCovered, or the
// taxparams lookup error (ErrNotInForce, ErrUnknownKey, ErrKindMismatch).
func (c *Catalog[K]) RateOn(key fiscal.RateKey, on fiscal.Date) (Resolved[money.Rate], error) {
	return resolve(c.sets, on, func(p *taxparams.Params) (money.Rate, error) {
		return p.Rate(taxparams.Key(key), on)
	})
}

// AmountOn returns the money parameter key holds on day on, with RateOn's
// errors.
func (c *Catalog[K]) AmountOn(key K, on fiscal.Date) (Resolved[money.Cents], error) {
	return c.money(taxparams.Key(key), on)
}

// CashPaymentLimit returns the cash payment above which an expense paid on
// day on is not deductible (LISR 27 fr. III).
func (c *Catalog[K]) CashPaymentLimit(on fiscal.Date) (Resolved[money.Cents], error) {
	return c.money(keyCashPaymentLimit, on)
}

// UMA returns the period's UMA in force on day on.
func (c *Catalog[K]) UMA(period UMAPeriod, on fiscal.Date) (Resolved[money.Cents], error) {
	return c.money(taxparams.Key(period), on)
}

func (c *Catalog[K]) money(key taxparams.Key, on fiscal.Date) (Resolved[money.Cents], error) {
	return resolve(c.sets, on, func(p *taxparams.Params) (money.Cents, error) {
		return p.Money(key, on)
	})
}

// AllowedIVARates returns the IVA treatments a record dated on may use, in
// contract order: exactly those fiscal.CheckIVA accepts through this
// catalog (16 % and 0 % while in force, 8 % only within the border
// decree's validity, exempt always). A date no set covers is
// ErrDateNotCovered.
func (c *Catalog[K]) AllowedIVARates(on fiscal.Date) (Resolved[[]fiscal.IVARate], error) {
	return resolve(c.sets, on, func(*taxparams.Params) ([]fiscal.IVARate, error) {
		allowed := make([]fiscal.IVARate, 0, len(ivaTreatments))
		for _, t := range ivaTreatments {
			if _, err := fiscal.ExpectedIVA(0, t, on, c); err == nil {
				allowed = append(allowed, t)
			}
		}
		return allowed, nil
	})
}

// resolve selects the set for on's fiscal year and reads one value from it.
func resolve[T any](sets *taxparams.Catalog, on fiscal.Date, read func(*taxparams.Params) (T, error)) (Resolved[T], error) {
	sel, err := sets.Select(on.Year())
	if err != nil {
		return Resolved[T]{}, fmt.Errorf("%w: %q: %w", ErrDateNotCovered, on, err)
	}
	v, err := read(sel.Params)
	if err != nil {
		return Resolved[T]{}, fmt.Errorf("ratecatalog: %w", err)
	}
	return Resolved[T]{
		Value:       v,
		Set:         taxparams.SetID{FiscalYear: sel.Params.FiscalYear(), Version: sel.Params.Version()},
		Provisional: sel.Provisional,
	}, nil
}
