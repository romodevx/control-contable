// Package access is the AccessPolicy port (P0-16, PRD §7, data-model
// "Trial-cap algorithm"): what a user may create, given their entitlements
// (trial / pro / comp), the trial clock and the trial-cap counters.
//
// Rules (owner defaults on PRD §7):
//
//   - The trial starts at registration and lasts TrialDuration (7 days); it is
//     over at exactly StartsAt + 7 days (ends_at is exclusive).
//   - In the trial a user creates at most TrialRecordCap incomes and
//     TrialRecordCap expenses, counted per user across all taxpayers. A record
//     counts when created; deleting never frees a slot, and restoring a
//     deleted record never takes one (DECIDED, MODULE-incomes §2.7): its slot
//     was spent at creation and never left the count.
//   - With an active pro or comp entitlement records are unlimited; with
//     neither, and no running trial, access is read-only: creates and edits
//     get 402 entitlement_required, while reads, deletes and restores of the
//     user's own deleted records stay (coordinator decision, M4).
//   - Every user, trial included, holds at most TaxpayerCap taxpayers (422
//     taxpayer_limit_reached). Coordinator decision on P0-16; PRD §7 "Taxpayers
//     in trial" is amended to match.
//   - A trial-cap 402 names the spent kind (message and details[].field), as the
//     caps are independent: a refused 6th income leaves expense slots.
//
// ReserveRecord is atomic: the counter is bumped by one conditional UPDATE in
// the caller's transaction, so concurrent creates never overshoot the cap,
// and a rolled-back create gives its slot back with the rest of the work.
package access

import (
	"errors"
	"time"
)

// TrialDuration is how long a trial lasts (PRD §7). Hours, not days, to match
// the entitlements CHECK (168 hours) across a DST change.
const TrialDuration = 168 * time.Hour

// TrialRecordCap is the number of incomes, and separately of expenses, a
// trial user may create (PRD §7: the 6th is refused).
const TrialRecordCap = 5

// TaxpayerCap is the fair-use maximum of taxpayers per user, trial included (PRD Q4).
const TaxpayerCap = 10

// Unlimited is the limit passed to UsageCounter.Reserve when no cap applies.
const Unlimited = -1

// RecordKind names a trial-capped record type.
type RecordKind string

// The record kinds with a trial cap.
const (
	RecordIncome  RecordKind = "income"
	RecordExpense RecordKind = "expense"
)

// Valid reports whether k is a known record kind.
func (k RecordKind) Valid() bool { return k == RecordIncome || k == RecordExpense }

// EntitlementKind is entitlements.kind.
type EntitlementKind string

// The entitlement kinds (entitlements_kind_source CHECK).
const (
	EntitlementTrial EntitlementKind = "trial"
	EntitlementPro   EntitlementKind = "pro"
	EntitlementComp  EntitlementKind = "comp"
)

// Entitlement is one row of a user's entitlements.
type Entitlement struct {
	Kind     EntitlementKind
	StartsAt time.Time
	// EndsAt is exclusive; the zero value means open-ended (comp only).
	EndsAt time.Time
}

// ActiveAt reports whether e covers now: StartsAt <= now < EndsAt.
func (e Entitlement) ActiveAt(now time.Time) bool {
	if now.Before(e.StartsAt) {
		return false
	}
	return e.EndsAt.IsZero() || now.Before(e.EndsAt)
}

// Level is the access a user has right now (contract AccessSummary.access).
type Level string

// The access levels.
const (
	LevelTrial    Level = "trial"
	LevelPro      Level = "pro"
	LevelComp     Level = "comp"
	LevelReadOnly Level = "read_only"
)

// Status is the decision for one user at one instant.
type Status struct {
	Level Level
	// TrialEndsAt is the end of the user's trial, zero without a trial row or
	// when a paid level applies (contract: trial_ends_at is null for pro).
	TrialEndsAt time.Time
}

// Writable reports whether the user may create or edit records.
func (s Status) Writable() bool { return s.Level != LevelReadOnly }

// Unlimited reports whether record creation is uncapped.
func (s Status) Unlimited() bool { return s.Level == LevelPro || s.Level == LevelComp }

// paidPrecedence orders the uncapped kinds: a real subscription is reported
// before a comp grant.
var paidPrecedence = []struct {
	kind  EntitlementKind
	level Level
}{
	{EntitlementPro, LevelPro},
	{EntitlementComp, LevelComp},
}

// Decide returns the access level the entitlements give at now. It is pure:
// the caller supplies now from its Clock.
func Decide(entitlements []Entitlement, now time.Time) Status {
	for _, paid := range paidPrecedence {
		if anyActive(entitlements, paid.kind, now) {
			return Status{Level: paid.level}
		}
	}
	trialRow, ok := find(entitlements, EntitlementTrial)
	if !ok {
		return Status{Level: LevelReadOnly}
	}
	if trialRow.ActiveAt(now) {
		return Status{Level: LevelTrial, TrialEndsAt: trialRow.EndsAt}
	}
	return Status{Level: LevelReadOnly, TrialEndsAt: trialRow.EndsAt}
}

func anyActive(entitlements []Entitlement, kind EntitlementKind, now time.Time) bool {
	for _, e := range entitlements {
		if e.Kind == kind && e.ActiveAt(now) {
			return true
		}
	}
	return false
}

func find(entitlements []Entitlement, kind EntitlementKind) (Entitlement, bool) {
	for _, e := range entitlements {
		if e.Kind == kind {
			return e, true
		}
	}
	return Entitlement{}, false
}

// Usage is a user's created-record counters (user_usage).
type Usage struct {
	Incomes  int
	Expenses int
}

// Of returns the counter of kind.
func (u Usage) Of(kind RecordKind) int {
	if kind == RecordExpense {
		return u.Expenses
	}
	return u.Incomes
}

var (
	// ErrReadOnly reports a write by a user without a running trial or plan.
	ErrReadOnly = errors.New("access: read-only access (trial over, no active plan)")
	// ErrTrialCapReached reports the 6th income or expense of a trial.
	ErrTrialCapReached = errors.New("access: trial record cap reached")
	// ErrTaxpayerLimitReached reports a taxpayer beyond TaxpayerCap.
	ErrTaxpayerLimitReached = errors.New("access: taxpayer limit reached")
	// ErrUnknownRecordKind reports a RecordKind that is neither income nor expense.
	ErrUnknownRecordKind = errors.New("access: unknown record kind")
	// ErrUsageMissing reports a user without a user_usage row (or one RLS
	// hides): a sign-up that did not create its counters. Never a 402.
	ErrUsageMissing = errors.New("access: user_usage row missing")
)
