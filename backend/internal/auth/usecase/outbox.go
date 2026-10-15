package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// EmailKind says which auth email an OutboundEmail is.
type EmailKind int

// The auth emails.
const (
	// EmailVerificationCode carries the sign-up code (register, resend).
	EmailVerificationCode EmailKind = iota + 1
	// EmailSignUpAttemptNotice tells an account owner that someone tried to
	// register with their address (D3).
	EmailSignUpAttemptNotice
)

// OutboundEmail is one auth email waiting for delivery. Code is set only for
// EmailVerificationCode.
type OutboundEmail struct {
	Kind EmailKind
	To   domain.Email
	Code domain.VerificationCode
}

// EmailOutbox queues auth emails in the caller's unit of work (the email
// outbox, migration 20261023100000_auth_email_outbox). Register and resend
// never talk to an email provider: the time a request takes must not depend
// on whether its address holds an account (MODULE-auth §4). The worker job
// EmailDeliveryJob sends what was queued.
type EmailOutbox interface {
	// Enqueue stores email, deliverable until expiresAt (within the hour).
	// The store keeps it sealed and holds nothing that links it to a user.
	Enqueue(ctx context.Context, email OutboundEmail, expiresAt time.Time) error
}

// ClaimedEmail is a queued email leased to one delivery run.
type ClaimedEmail struct {
	// ID is the store's opaque row id.
	ID string
	// Claim is the token of this lease; Finish and Retry need it.
	Claim string
	// Attempt counts this one: 1 on the first delivery attempt.
	Attempt int
	Email   OutboundEmail
	// Unreadable is true when the stored payload could not be opened (a lost
	// key, a corrupt row): Email is zero and the job drops it unsent.
	Unreadable bool
}

// EmailQueue is the worker side of the outbox, in the unit of work in ctx.
// Finish and Retry are compare-and-set on the claim: held = false means the
// lease ran out and another run holds the email now (not an error).
type EmailQueue interface {
	// Claim leases up to limit due emails for lease.
	Claim(ctx context.Context, limit int, lease time.Duration) ([]ClaimedEmail, error)
	// Finish deletes a delivered (or abandoned) email.
	Finish(ctx context.Context, id, claim string) (held bool, err error)
	// Retry makes the email due again after delay.
	Retry(ctx context.Context, id, claim string, delay time.Duration) (held bool, err error)
}

// enqueue queues email, deliverable while its code lives (CodeTTL from now).
func enqueue(ctx context.Context, outbox EmailOutbox, email OutboundEmail, now time.Time) error {
	if err := outbox.Enqueue(ctx, email, now.Add(CodeTTL)); err != nil {
		return fmt.Errorf("auth: queue email: %w", err)
	}
	return nil
}
