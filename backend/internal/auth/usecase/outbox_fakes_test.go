package usecase_test

import (
	"context"
	"sync"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// queued is one email the fake outbox accepted, as a test compares it.
type queued struct {
	kind      string // "code" or "notice"
	to        string
	code      string
	expiresAt time.Time
	// inUnit is true when it was enqueued inside the use case's unit of work.
	inUnit bool
}

// fakeOutbox records what the use cases enqueue; err makes every call fail.
type fakeOutbox struct {
	queued []queued
	err    error
}

func (f *fakeOutbox) Enqueue(ctx context.Context, email usecase.OutboundEmail, expiresAt time.Time) error {
	if f.err != nil {
		return f.err
	}
	f.queued = append(f.queued, queued{
		kind: kindName(email.Kind), to: email.To.String(), code: email.Code.Reveal(),
		expiresAt: expiresAt, inUnit: inSystem(ctx),
	})
	return nil
}

func kindName(k usecase.EmailKind) string {
	switch k {
	case usecase.EmailVerificationCode:
		return "code"
	case usecase.EmailSignUpAttemptNotice:
		return "notice"
	default:
		return "unknown"
	}
}

// settled is one Finish or Retry call of the fake queue.
type settled struct {
	op    string // "finish" or "retry"
	id    string
	claim string
	delay time.Duration
	// inUnit is true when it ran inside a System unit of work.
	inUnit bool
	// ctxErr is the context's error at the call (a cancelled run still settles).
	ctxErr error
}

// fakeQueue serves scripted claim batches and records how each email was settled.
type fakeQueue struct {
	mu        sync.Mutex
	batches   [][]usecase.ClaimedEmail
	claimErr  error
	settleErr error
	// lost makes Finish and Retry report the claim as no longer held.
	lost bool

	claims  []claimCall
	settled []settled
}

// claimCall records one Claim.
type claimCall struct {
	limit  int
	lease  time.Duration
	inUnit bool
}

func (f *fakeQueue) Claim(ctx context.Context, limit int, lease time.Duration) ([]usecase.ClaimedEmail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.claims = append(f.claims, claimCall{limit: limit, lease: lease, inUnit: inSystem(ctx)})
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	if len(f.batches) == 0 {
		return nil, nil
	}
	batch := f.batches[0]
	f.batches = f.batches[1:]
	return batch, nil
}

func (f *fakeQueue) Finish(ctx context.Context, id, claim string) (bool, error) {
	return f.settle(ctx, settled{op: "finish", id: id, claim: claim})
}

func (f *fakeQueue) Retry(ctx context.Context, id, claim string, delay time.Duration) (bool, error) {
	return f.settle(ctx, settled{op: "retry", id: id, claim: claim, delay: delay})
}

func (f *fakeQueue) settle(ctx context.Context, s settled) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s.inUnit, s.ctxErr = inSystem(ctx), ctx.Err()
	f.settled = append(f.settled, s)
	if f.settleErr != nil {
		return false, f.settleErr
	}
	return !f.lost, nil
}
