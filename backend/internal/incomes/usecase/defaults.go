package usecase

import (
	"context"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Query parameter names getIncomeDefaults reports.
const (
	fieldRegime = "regime"
	fieldOn     = "on"
)

// ivaTreatments are the contract's IvaRate values in contract order.
var ivaTreatments = []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}

// DefaultsQuery is getIncomeDefaults: the path's taxpayer, the regime and
// counterparty, the day ("YYYY-MM-DD"; "" is today) and the 625 activity
// (empty: the taxpayer's only one, if any).
type DefaultsQuery struct {
	TaxpayerID       string
	Regime           fiscal.RegimeCode
	Counterparty     domain.Counterparty
	On               string
	PlatformActivity domain.PlatformActivity
}

// SuggestedRate is a suggested fraction; Known is false when there is no
// suggestion (the contract's null).
type SuggestedRate struct {
	Rate  money.Rate
	Known bool
}

// WithholdingDefault is whether a tax may be withheld and its suggested
// rate. An allowed withholding may have no suggestion.
type WithholdingDefault struct {
	Allowed   bool
	Suggested SuggestedRate
}

// Defaults is the income form's prefill (contract IncomeDefaults): the IVA
// rates valid on the day, the suggested one, the withholdings (ISR as a
// fraction of the amount, IVA of the IVA) and the resolved 625 activity.
type Defaults struct {
	IVARatesAllowed  []fiscal.IVARate
	SuggestedIVARate fiscal.IVARate
	// ISR and IVA are the deprecated suggestions (ADR-0008), served for the
	// shipped form until WH-11.
	ISR              WithholdingDefault
	IVA              WithholdingDefault
	PlatformActivity domain.PlatformActivity
	LastUsed         LastFacts
}

// Defaults answers getIncomeDefaults (MODULE-incomes OQ-7): the dated rates
// of the catalog for the taxpayer's person type, the regime and its 625
// options, the counterparty and the day. A read: allowed for a read-only
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
		if out, err = uc.defaults(profile, held, q, on); err != nil {
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
	if !q.Counterparty.Valid() {
		vs.Add(domain.FieldCounterpartyType, fiscal.ErrInvalid)
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
func (uc *Incomes) defaults(profile TaxpayerProfile, held RegimeProfile, q DefaultsQuery, on fiscal.Date) (Defaults, error) {
	activity, err := defaultActivity(held, q.PlatformActivity)
	if err != nil {
		return Defaults{}, err
	}
	payee := domain.Payee{
		PersonType: profile.PersonType, Regime: held.Code,
		PlatformActivities: held.PlatformActivities, RFCGivenToPlatform: held.RFCGivenToPlatform,
	}
	policy := domain.PolicyFor(payee, q.Counterparty, policyActivity(held, activity))
	isr := WithholdingDefault{Allowed: policy.ISR.Allowed}
	if activity != "" || held.Code != fiscal.RegimePlatforms {
		isr.Suggested = uc.suggestedRate(policy.ISR.Key, on)
	}
	allowed := uc.allowedIVARates(on)
	return Defaults{
		IVARatesAllowed: allowed, SuggestedIVARate: suggestedIVARate(allowed),
		ISR: isr, IVA: WithholdingDefault{Allowed: policy.IVA.Allowed, Suggested: shareRate(policy.IVA.Share)},
		PlatformActivity: activity,
	}, nil
}

// defaultActivity is the 625 activity the ISR suggestion is for: the chosen
// one, which the taxpayer must hold (422 activity_not_held), else the only
// one held, else none.
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

// policyActivity is the activity whose policy decides what is allowed.
// With several activities and none chosen, any held one does: the activity
// only changes the ISR rate, which then has no suggestion.
func policyActivity(held RegimeProfile, activity domain.PlatformActivity) domain.PlatformActivity {
	if activity == "" && len(held.PlatformActivities) > 0 {
		return held.PlatformActivities[0]
	}
	return activity
}

// suggestedRate is the catalog rate of key on day on; no key, or no rate
// in force that day, is no suggestion.
func (uc *Incomes) suggestedRate(key fiscal.RateKey, on fiscal.Date) SuggestedRate {
	if key == "" {
		return SuggestedRate{}
	}
	r, ok := uc.deps.Rates.Rate(key, on)
	return SuggestedRate{Rate: r, Known: ok}
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

// shareRate is a legal share as a Rate rounded half up to its six decimals
// (2/3 is 0.666667); the zero Share is no suggestion.
func shareRate(s fiscal.Share) SuggestedRate {
	if s.IsZero() {
		return SuggestedRate{}
	}
	scaled := (2*s.Num*money.RateScale + s.Den) / (2 * s.Den)
	return SuggestedRate{Rate: money.Rate(scaled), Known: true}
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
