package access

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// Client-safe messages (contract PaymentRequired / taxpayer_limit_reached).
const (
	msgTrialEnded  = "Your trial has ended. Subscribe to keep adding records."
	msgTaxpayerCap = "You have reached the maximum number of taxpayers. Contact support to raise it."
)

// trialCap is, per kind, the client message and the AccessSummary field of a
// trial-cap refusal.
var trialCap = map[RecordKind]struct{ field, message string }{
	RecordIncome: {"incomes_remaining",
		"You have used every income of your trial. Subscribe to keep adding incomes."},
	RecordExpense: {"expenses_remaining",
		"You have used every expense of your trial. Subscribe to keep adding expenses."},
}

// trialCapError is the 402 for the kind whose trial cap is spent. It names the
// kind in the message and in details, since the caps are independent.
func trialCapError(kind RecordKind) error {
	refusal := trialCap[kind]
	return apperr.Wrap(fmt.Errorf("%w: %s", ErrTrialCapReached, kind), apperr.CodeEntitlementRequired, refusal.message).
		WithFields(apperr.FieldError{Field: refusal.field, Code: DetailTrialCapReached, Message: refusal.message})
}

// DetailTrialCapReached is the details[].code of a trial-cap 402. Its field
// names the AccessSummary counter that hit zero, so the client knows which
// cap (income or expense) refused the create; the other may still have slots.
const DetailTrialCapReached = "trial_cap_reached"

// EntitlementReader is the port to a user's entitlement rows.
type EntitlementReader interface {
	Entitlements(ctx context.Context, userID string) ([]Entitlement, error)
}

// UsageCounter is the port to the trial-cap counters.
type UsageCounter interface {
	// Reserve adds one to the kind's counter only while it is below limit
	// (Unlimited: always), in one atomic statement of the ctx's transaction,
	// and reports whether it did. A missing counter row is ErrUsageMissing.
	Reserve(ctx context.Context, userID string, kind RecordKind, limit int) (bool, error)
	// Usage returns the counters.
	Usage(ctx context.Context, userID string) (Usage, error)
}

// Policy is the AccessPolicy the income and expense use cases consume. It is
// safe for concurrent use.
type Policy struct {
	txm          tx.Manager
	entitlements EntitlementReader
	usage        UsageCounter
	clock        clock.Clock
}

// NewPolicy wires a Policy. Its methods run in a unit of work of txm: called
// from inside the caller's unit of work (same ctx) they join it, so the
// reservation commits or rolls back with the record it pays for.
func NewPolicy(txm tx.Manager, entitlements EntitlementReader, usage UsageCounter, clk clock.Clock) *Policy {
	return &Policy{txm: txm, entitlements: entitlements, usage: usage, clock: clk}
}

// ReserveRecord takes one creation slot of kind for userID, on create only:
// a restore never calls it (no double count). It returns a 402 entitlement_required *apperr.Error wrapping
// ErrReadOnly or ErrTrialCapReached when the user may not create; the slot is
// given back only if the surrounding transaction rolls back.
func (p *Policy) ReserveRecord(ctx context.Context, userID string, kind RecordKind) error {
	if !kind.Valid() {
		return fmt.Errorf("%w: %q", ErrUnknownRecordKind, kind)
	}
	return p.txm.ReadWrite(ctx, userID, func(ctx context.Context) error {
		status, err := p.status(ctx, userID)
		if err != nil {
			return err
		}
		if !status.Writable() {
			return apperr.Wrap(ErrReadOnly, apperr.CodeEntitlementRequired, msgTrialEnded)
		}
		reserved, err := p.usage.Reserve(ctx, userID, kind, limitFor(status))
		if err != nil {
			return fmt.Errorf("access: reserve %s: %w", kind, err)
		}
		if !reserved {
			return trialCapError(kind)
		}
		return nil
	})
}

// EnsureWritable refuses edits of a read-only user (402). Edits create
// nothing, so they reserve no slot and are allowed at the trial cap.
func (p *Policy) EnsureWritable(ctx context.Context, userID string) error {
	return p.txm.ReadOnly(ctx, userID, func(ctx context.Context) error {
		status, err := p.status(ctx, userID)
		if err != nil {
			return err
		}
		if !status.Writable() {
			return apperr.Wrap(ErrReadOnly, apperr.CodeEntitlementRequired, msgTrialEnded)
		}
		return nil
	})
}

// CheckTaxpayerLimit refuses a new taxpayer when the user already has
// existing taxpayers (422). The fair-use cap applies at every level, trial
// included (P0-16 decision, PRD §7 "Taxpayers in trial"), so it needs no
// entitlement lookup. The count belongs to
// the taxpayers module; it must lock what it counts so two concurrent creates
// cannot both pass.
func (p *Policy) CheckTaxpayerLimit(_ context.Context, _ string, existing int) error {
	if existing < TaxpayerCap {
		return nil
	}
	return apperr.Wrap(ErrTaxpayerLimitReached, apperr.CodeTaxpayerLimitReached, msgTaxpayerCap)
}

// Summary is the access state shown to the client (contract AccessSummary).
type Summary struct {
	Status Status
	Usage  Usage
}

// Remaining returns how many records of kind the user may still create and
// whether a limit applies at all (false for pro and comp).
func (s Summary) Remaining(kind RecordKind) (int, bool) {
	if s.Status.Unlimited() {
		return 0, false
	}
	if !s.Status.Writable() {
		return 0, true
	}
	return max(TrialRecordCap-s.Usage.Of(kind), 0), true
}

// Summary returns the user's access level and counters.
func (p *Policy) Summary(ctx context.Context, userID string) (Summary, error) {
	var summary Summary
	err := p.txm.ReadOnly(ctx, userID, func(ctx context.Context) error {
		status, err := p.status(ctx, userID)
		if err != nil {
			return err
		}
		usage, err := p.usage.Usage(ctx, userID)
		if err != nil {
			return fmt.Errorf("access: read usage: %w", err)
		}
		summary = Summary{Status: status, Usage: usage}
		return nil
	})
	return summary, err
}

func (p *Policy) status(ctx context.Context, userID string) (Status, error) {
	entitlements, err := p.entitlements.Entitlements(ctx, userID)
	if err != nil {
		return Status{}, fmt.Errorf("access: read entitlements: %w", err)
	}
	return Decide(entitlements, p.clock.Now()), nil
}

func limitFor(status Status) int {
	if status.Unlimited() {
		return Unlimited
	}
	return TrialRecordCap
}
