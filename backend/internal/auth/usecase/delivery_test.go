package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

const deliveryBatch = 3

type deliveryEnv struct {
	job    *usecase.EmailDeliveryJob
	system *fakeSystem
	queue  *fakeQueue
	sender *fakeSender
	logs   interface{ String() string }
}

func newDeliveryEnv(t *testing.T, batches ...[]usecase.ClaimedEmail) *deliveryEnv {
	t.Helper()
	logger, logs := capturedLogs()
	e := &deliveryEnv{system: &fakeSystem{}, queue: &fakeQueue{batches: batches}, sender: &fakeSender{}, logs: logs}
	job, err := usecase.NewEmailDeliveryJob(usecase.DeliveryDeps{
		System: e.system, Queue: e.queue, Sender: e.sender, Logger: logger,
		BatchSize: deliveryBatch, MaxBatches: 4,
	})
	require.NoError(t, err)
	e.job = job
	return e
}

func mustEmail(t *testing.T, raw string) domain.Email {
	t.Helper()
	email, err := domain.NewEmail(raw)
	require.NoError(t, err)
	return email
}

func mustCode(t *testing.T, digits string) domain.VerificationCode {
	t.Helper()
	code, err := domain.ParseVerificationCode(digits)
	require.NoError(t, err)
	return code
}

func codeEmail(t *testing.T, id string, attempt int) usecase.ClaimedEmail {
	t.Helper()
	return usecase.ClaimedEmail{
		ID: id, Claim: "claim-" + id, Attempt: attempt,
		Email: usecase.OutboundEmail{Kind: usecase.EmailVerificationCode, To: mustEmail(t, normalEmail), Code: mustCode(t, "482913")},
	}
}

func noticeEmail(t *testing.T, id string) usecase.ClaimedEmail {
	t.Helper()
	return usecase.ClaimedEmail{
		ID: id, Claim: "claim-" + id, Attempt: 1,
		Email: usecase.OutboundEmail{Kind: usecase.EmailSignUpAttemptNotice, To: mustEmail(t, "luis@example.mx")},
	}
}

func TestEmailDeliveryJob_isAWorkerJob(t *testing.T) {
	e := newDeliveryEnv(t)

	assert.Equal(t, "auth_email_delivery", e.job.Name())
	assert.Equal(t, usecase.DeliveryInterval, e.job.Interval())
	assert.Less(t, usecase.DeliveryRunTimeout, usecase.DeliveryLease,
		"a run ends long before its claims' lease, so no other instance takes them mid-send")
	assert.LessOrEqual(t, usecase.DeliveryRunTimeout, usecase.DeliveryInterval)
}

func TestEmailDeliveryJob_deliversAndFinishesEachClaimedEmail(t *testing.T) {
	e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", 1), noticeEmail(t, "b")})

	require.NoError(t, e.job.Run(t.Context()))

	assert.Equal(t, []sent{
		{kind: "code", to: normalEmail, code: "482913"},
		{kind: "notice", to: "luis@example.mx"},
	}, e.sender.sent)
	assert.Equal(t, []settled{
		{op: "finish", id: "a", claim: "claim-a", inUnit: true},
		{op: "finish", id: "b", claim: "claim-b", inUnit: true},
	}, e.queue.settled, "deleted on delivery, by the claim that holds it")
	require.Len(t, e.queue.claims, 1, "a short batch ends the run")
	assert.Equal(t, claimCall{limit: deliveryBatch, lease: usecase.DeliveryLease, inUnit: true}, e.queue.claims[0])
}

// Bounded retries with exponential backoff; the last failed attempt drops the
// email (it would expire with its code anyway).
func TestEmailDeliveryJob_failedSendIsRetriedThenDropped(t *testing.T) {
	tests := []struct {
		attempt int
		want    settled
	}{
		{1, settled{op: "retry", id: "a", claim: "claim-a", delay: 30 * time.Second, inUnit: true}},
		{2, settled{op: "retry", id: "a", claim: "claim-a", delay: time.Minute, inUnit: true}},
		{3, settled{op: "retry", id: "a", claim: "claim-a", delay: 2 * time.Minute, inUnit: true}},
		{4, settled{op: "retry", id: "a", claim: "claim-a", delay: 4 * time.Minute, inUnit: true}},
		{usecase.MaxDeliveryAttempts, settled{op: "finish", id: "a", claim: "claim-a", inUnit: true}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt %d", tt.attempt), func(t *testing.T) {
			e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", tt.attempt)})
			e.sender.err = errors.New("provider unavailable")

			require.NoError(t, e.job.Run(t.Context()), "a failed send is the email's problem, not the run's")

			assert.Equal(t, []settled{tt.want}, e.queue.settled)
		})
	}
}

func TestDeliveryBackoff_isExponentialAndCapped(t *testing.T) {
	tests := map[int]time.Duration{
		0: 30 * time.Second, 1: 30 * time.Second, 2: time.Minute, 3: 2 * time.Minute,
		4: 4 * time.Minute, 5: 5 * time.Minute, 40: 5 * time.Minute,
	}
	for attempt, want := range tests {
		assert.Equal(t, want, usecase.DeliveryBackoff(attempt), "attempt %d", attempt)
	}
}

func TestEmailDeliveryJob_unreadableEmailIsDroppedUnsent(t *testing.T) {
	e := newDeliveryEnv(t, []usecase.ClaimedEmail{{ID: "a", Claim: "claim-a", Attempt: 1, Unreadable: true}})

	require.NoError(t, e.job.Run(t.Context()))

	assert.Empty(t, e.sender.sent)
	assert.Equal(t, []settled{{op: "finish", id: "a", claim: "claim-a", inUnit: true}}, e.queue.settled)
	assert.Contains(t, e.logs.String(), `"dropped":1`)
}

func TestEmailDeliveryJob_drainsFullBatchesUpToTheRunLimit(t *testing.T) {
	full := func(prefix string) []usecase.ClaimedEmail {
		return []usecase.ClaimedEmail{noticeEmail(t, prefix+"1"), noticeEmail(t, prefix+"2"), noticeEmail(t, prefix+"3")}
	}
	e := newDeliveryEnv(t, full("a"), full("b"), full("c"), full("d"), full("e"))

	require.NoError(t, e.job.Run(t.Context()))

	assert.Len(t, e.queue.claims, 4, "MaxBatches claims at most; the rest waits for the next run")
	assert.Len(t, e.sender.sent, 12)
}

func TestEmailDeliveryJob_claimErrorFailsTheRun(t *testing.T) {
	e := newDeliveryEnv(t)
	e.queue.claimErr = errDB

	require.ErrorIs(t, e.job.Run(t.Context()), errDB)
	assert.Empty(t, e.sender.sent)
}

// A claim lost to another instance (its lease ran out) is not an error: the
// holder settles it. Bookkeeping errors fail the run after the batch.
func TestEmailDeliveryJob_settling(t *testing.T) {
	t.Run("lost claim", func(t *testing.T) {
		e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", 1)})
		e.queue.lost = true

		require.NoError(t, e.job.Run(t.Context()))
		assert.Contains(t, e.logs.String(), `"lost":1`)
	})
	t.Run("settle error", func(t *testing.T) {
		e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", 1), codeEmail(t, "b", 1)})
		e.queue.settleErr = errDB

		require.ErrorIs(t, e.job.Run(t.Context()), errDB)
		assert.Len(t, e.sender.sent, 2, "one bookkeeping failure does not hold back the batch")
	})
}

// A run cut short (timeout, lost leadership, shutdown) still records what it
// already sent, so a delivered email is not sent again after the lease.
func TestEmailDeliveryJob_cancelledRunStillSettlesWhatItSent(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", 1), codeEmail(t, "b", 1)})
	e.sender.onSend = cancel

	err := e.job.Run(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.Len(t, e.sender.sent, 1, "nothing more is sent once the run is cancelled")
	require.Len(t, e.queue.settled, 1)
	assert.Equal(t, "a", e.queue.settled[0].id)
	assert.NoError(t, e.queue.settled[0].ctxErr, "settled on a context the cancellation does not reach")
}

// No address, code or email kind in the job's logs, on success or failure.
func TestEmailDeliveryJob_logsCountsOnly(t *testing.T) {
	e := newDeliveryEnv(t, []usecase.ClaimedEmail{codeEmail(t, "a", 1), codeEmail(t, "b", usecase.MaxDeliveryAttempts), noticeEmail(t, "c")})
	e.sender.err = errors.New("provider unavailable")

	require.NoError(t, e.job.Run(t.Context()))

	logs := e.logs.String()
	for _, secret := range []string{normalEmail, "luis@example.mx", "482913"} {
		assert.NotContains(t, logs, secret)
	}
	assert.Contains(t, logs, `"retried":2`)
	assert.Contains(t, logs, `"dropped":1`)
}

func TestNewEmailDeliveryJob_validatesItsDeps(t *testing.T) {
	logger, _ := capturedLogs()
	valid := usecase.DeliveryDeps{
		System: &fakeSystem{}, Queue: &fakeQueue{}, Sender: &fakeSender{}, Logger: logger,
		BatchSize: 10, MaxBatches: 1,
	}
	tests := map[string]func(*usecase.DeliveryDeps){
		"no system":       func(d *usecase.DeliveryDeps) { d.System = nil },
		"no queue":        func(d *usecase.DeliveryDeps) { d.Queue = nil },
		"no sender":       func(d *usecase.DeliveryDeps) { d.Sender = nil },
		"no logger":       func(d *usecase.DeliveryDeps) { d.Logger = nil },
		"zero batch":      func(d *usecase.DeliveryDeps) { d.BatchSize = 0 },
		"batch too large": func(d *usecase.DeliveryDeps) { d.BatchSize = usecase.MaxDeliveryBatch + 1 },
		"zero batches":    func(d *usecase.DeliveryDeps) { d.MaxBatches = 0 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			deps := valid
			mutate(&deps)
			_, err := usecase.NewEmailDeliveryJob(deps)
			require.ErrorIs(t, err, usecase.ErrInvalidDeliveryConfig)
		})
	}
	_, err := usecase.NewEmailDeliveryJob(valid)
	require.NoError(t, err)
}
