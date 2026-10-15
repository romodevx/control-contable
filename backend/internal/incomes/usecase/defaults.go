package usecase

import (
	"context"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// Query parameter names getIncomeDefaults reports.
const (
	fieldRegime = "regime"
	fieldOn     = "on"
)

// ivaTreatments are the contract's IvaRate values in contract order.
var ivaTreatments = []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}

// DefaultsQuery is getIncomeDefaults: the path's taxpayer, the regime, the
// day ("YYYY-MM-DD"; "" is today) and the 625 activity (empty: the
// taxpayer's only one, if any).
type DefaultsQuery struct {
	TaxpayerID       string
	Regime           fiscal.RegimeCode
	On               string
	PlatformActivity domain.PlatformActivity
}

// Defaults is the income form's prefill (contract IncomeDefaults): the IVA
// rates valid on the day, the suggested one, the resolved 625 activity and
// the withholding facts to preselect. Withholdings are derived (ADR-0008),
// so there is no withholding rate to suggest.
type Defaults struct {
	IVARatesAllowed  []fiscal.IVARate
	SuggestedIVARate fiscal.IVARate
	PlatformActivity domain.PlatformActivity
	LastUsed         LastFacts
}

// Defaults answers getIncomeDefaults (MODULE-incomes OQ-7): the dated rates
// of the catalog for the regime, its 625 activities and the day, and the
// last-used withholding facts. A read: allowed for a read-only
// user and an archived taxpayer. The regime must be held on the day (422
// regime_not_active), read from the profile's windows: a read-only unit of
// work takes no lock.
func (uc *Incomes) Defaults(ctx context.Context, userID string, q DefaultsQuery) (Defaults, error) {
	taxpayerID, ok := canonicalID(q.TaxpayerID)
	if !ok {
		return Defaults{}, taxpayerNotFound()
	}
	on, err := uc.defaultsDay(q)
	if err != nil {
		return Defaults{}, present("income defaults", err)
	}
	var out Defaults
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		profile, err := uc.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
		if err != nil {
			return err
		}
		held, ok := holding(profile, q.Regime, on.Period())
		if !ok {
			return regimeNotHeldOn(fieldRegime)
		}
		if out, err = uc.defaults(held, q, on); err != nil {
			return err
		}
		out.LastUsed, err = uc.lastUsed(ctx, taxpayerID, profile.PersonType, held.Code)
		return err
	})
	if err != nil {
		return Defaults{}, present("income defaults", err)
	}
	return out, nil
}

// lastUsed is IncomeLastUsed (MODULE-withholdings §2.1): the facts of the
// newest income in the regime, or the first-time defaults: no counterparty
// (OQ-W2), the regime's operation, no housing rent, via a platform in 625.
func (uc *Incomes) lastUsed(ctx context.Context, taxpayerID string, person fiscal.PersonType, regime fiscal.RegimeCode) (LastFacts, error) {
	last, ok, err := uc.deps.Store.LastFacts(ctx, taxpayerID, regime)
	if err != nil || ok {
		return last, err
	}
	return LastFacts{Operation: domain.DefaultOperation(person, regime), ViaPlatform: domain.DefaultViaPlatform(regime)}, nil
}

// defaultsDay validates the query (400 with every field) and returns its
// day: today when absent, else within [fiscal.MinRecordDate, today] and
// covered by a parameter set.
func (uc *Incomes) defaultsDay(q DefaultsQuery) (fiscal.Date, error) {
	today, err := uc.today()
	if err != nil {
		return fiscal.Date{}, err
	}
	var vs fiscal.Violations
	if q.Regime == "" {
		vs.Add(fieldRegime, fiscal.ErrRequired)
	}
	switch {
	case q.PlatformActivity == "":
	case !q.PlatformActivity.Valid():
		vs.Add(domain.FieldPlatformActivity, fiscal.ErrInvalid)
	case q.Regime != fiscal.RegimePlatforms:
		vs.Add(domain.FieldPlatformActivity, domain.ErrNotForRegime)
	}
	on, refusal, ok := uc.parseDefaultsDay(q.On, today)
	if !ok {
		vs.Add(fieldOn, refusal)
	}
	if len(vs) > 0 {
		return fiscal.Date{}, vs
	}
	return on, nil
}

// parseDefaultsDay returns the day of raw ("" is today), or false and the
// fiscal error that refuses it.
func (uc *Incomes) parseDefaultsDay(raw string, today fiscal.Date) (fiscal.Date, fiscal.Error, bool) {
	if raw == "" {
		return today, fiscal.Error{}, true
	}
	on, err := fiscal.ParseDate(raw)
	switch {
	case err != nil:
		return fiscal.Date{}, fiscal.ErrInvalid, false
	case on.Before(fiscal.MinRecordDate) || on.After(today):
		return fiscal.Date{}, fiscal.ErrOutOfRange, false
	}
	if _, err := uc.deps.Rates.Provisional(on); err != nil {
		return fiscal.Date{}, fiscal.ErrOutOfRange, false // before the first parameter set (2026)
	}
	return on, fiscal.Error{}, true
}

// defaults resolves the prefill for a held regime on a valid day.
func (uc *Incomes) defaults(held RegimeProfile, q DefaultsQuery, on fiscal.Date) (Defaults, error) {
	activity, err := defaultActivity(held, q.PlatformActivity)
	if err != nil {
		return Defaults{}, err
	}
	allowed := uc.allowedIVARates(on)
	return Defaults{IVARatesAllowed: allowed, SuggestedIVARate: suggestedIVARate(allowed), PlatformActivity: activity}, nil
}

// defaultActivity is the 625 activity to preselect: the chosen one, which
// the taxpayer must hold (422 activity_not_held), else the only one held,
// else none.
func defaultActivity(held RegimeProfile, chosen domain.PlatformActivity) (domain.PlatformActivity, error) {
	if held.Code != fiscal.RegimePlatforms {
		return "", nil
	}
	if chosen != "" {
		if !slices.Contains(held.PlatformActivities, chosen) {
			var vs fiscal.Violations
			vs.Add(domain.FieldPlatformActivity, domain.ErrActivityNotHeld)
			return "", vs
		}
		return chosen, nil
	}
	if len(held.PlatformActivities) == 1 {
		return held.PlatformActivities[0], nil
	}
	return "", nil
}

// allowedIVARates are the treatments fiscal.CheckIVA accepts on day on.
func (uc *Incomes) allowedIVARates(on fiscal.Date) []fiscal.IVARate {
	allowed := make([]fiscal.IVARate, 0, len(ivaTreatments))
	for _, t := range ivaTreatments {
		if _, err := fiscal.ExpectedIVA(0, t, on, uc.deps.Rates); err == nil {
			allowed = append(allowed, t)
		}
	}
	return allowed
}

// suggestedIVARate is the general 16 % while it is allowed (US-I1), else
// the first allowed treatment.
func suggestedIVARate(allowed []fiscal.IVARate) fiscal.IVARate {
	if slices.Contains(allowed, fiscal.IVARate16) || len(allowed) == 0 {
		return fiscal.IVARate16
	}
	return allowed[0]
}

// holding returns the regime the profile holds in period.
func holding(p TaxpayerProfile, code fiscal.RegimeCode, period string) (RegimeProfile, bool) {
	for _, r := range p.Regimes {
		if r.Code == code && r.StartPeriod <= period && (r.EndPeriod == "" || period <= r.EndPeriod) {
			return r, true
		}
	}
	return RegimeProfile{}, false
}
