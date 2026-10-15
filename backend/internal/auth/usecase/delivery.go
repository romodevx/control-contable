package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// Delivery policy of the auth email outbox.
const (
	// DeliveryInterval is the time between two delivery runs: a sign-up code
	// should arrive within seconds.
	DeliveryInterval = 5 * time.Second
	// DeliveryRunTimeout bounds one run (at most its interval: runs of a job
	// never overlap).
	DeliveryRunTimeout = DeliveryInterval
	// DeliveryLease is how long a claim holds an email. It is far longer than
	// a run, so no other instance claims an email while a run is sending it;
	// an email whose run died is retried when its lease ends.
	DeliveryLease = 2 * time.Minute
	// MaxDeliveryAttempts mirrors the cap inside auth_email_outbox_claim: the
	// email is dropped when its last attempt fails.
	MaxDeliveryAttempts = 5
	// MaxDeliveryBatch mirrors the claim's upper bound.
	MaxDeliveryBatch = 100
	// DefaultDeliveryBatch emails per claim.
	DefaultDeliveryBatch = 20
	// DefaultDeliveryBatches claims per run at most; the rest waits for the next run.
	DefaultDeliveryBatches = 5

	deliveryJobName = "auth_email_delivery"
	// firstRetryDelay doubles on every failed attempt, up to maxRetryDelay.
	firstRetryDelay = 30 * time.Second
	maxRetryDelay   = 5 * time.Minute
	// settleTimeout bounds the bookkeeping of one email after its send, on a
	// context the run's cancellation does not reach.
	settleTimeout = 5 * time.Second
)

// ErrInvalidDeliveryConfig reports an EmailDeliveryJob built with missing or
// out-of-range deps.
var ErrInvalidDeliveryConfig = errors.New("auth: invalid email delivery configuration")

// DeliveryBackoff is the wait before attempt+1 after attempt failed:
// 30 s, 1 min, 2 min, 4 min, then 5 min.
func DeliveryBackoff(attempt int) time.Duration {
	delay := firstRetryDelay
	for i := 1; i < attempt && delay < maxRetryDelay; i++ {
		delay *= 2
	}
	return min(delay, maxRetryDelay)
}

// DeliveryDeps are what an EmailDeliveryJob needs.
type DeliveryDeps struct {
	System tx.SystemManager
	Queue  EmailQueue
	// Sender is the provider-agnostic port (the dev log sender until a
	// provider is chosen, OB2).
	Sender     EmailSender
	Logger     *slog.Logger
	BatchSize  int
	MaxBatches int
}

// EmailDeliveryJob is the worker job (worker.Job) that sends the queued auth
// emails: it claims a batch (one short System unit of work), sends each email
// outside any transaction, then deletes it or schedules its retry (one unit of
// work each). It runs whatever module_auth says: an email queued before the
// flag was turned off is still delivered.
//
// Safe under a worker split brain: claims are leased under a token
// (auth_email_outbox_claim, SKIP LOCKED), so two instances never send the
// same email while a lease holds. Only counts are logged, never an address,
// a code or which kind of email it was.
type EmailDeliveryJob struct {
	deps DeliveryDeps
}

// NewEmailDeliveryJob validates deps and builds the job.
func NewEmailDeliveryJob(deps DeliveryDeps) (*EmailDeliveryJob, error) {
	switch {
	case deps.System == nil, deps.Queue == nil, deps.Sender == nil, deps.Logger == nil:
		return nil, fmt.Errorf("%w: nil dependency", ErrInvalidDeliveryConfig)
	case deps.BatchSize < 1 || deps.BatchSize > MaxDeliveryBatch:
		return nil, fmt.Errorf("%w: batch size %d not in [1, %d]", ErrInvalidDeliveryConfig, deps.BatchSize, MaxDeliveryBatch)
	case deps.MaxBatches < 1:
		return nil, fmt.Errorf("%w: max batches %d is not positive", ErrInvalidDeliveryConfig, deps.MaxBatches)
	}
	return &EmailDeliveryJob{deps: deps}, nil
}

// Name implements worker.Job.
func (j *EmailDeliveryJob) Name() string { return deliveryJobName }

// Interval implements worker.Job.
func (j *EmailDeliveryJob) Interval() time.Duration { return DeliveryInterval }

// deliveryTally counts what a run did with its emails.
type deliveryTally struct {
	delivered, retried, dropped, lost int
}

// Run implements worker.Job: claims until a batch comes back short,
// MaxBatches is reached, a claim fails or ctx ends. A failed send is the
// email's problem (retried or dropped), never the run's; a failed bookkeeping
// write fails the run after its batch.
func (j *EmailDeliveryJob) Run(ctx context.Context) error {
	var counts deliveryTally
	defer j.report(ctx, &counts)
	for range j.deps.MaxBatches {
		if err := ctx.Err(); err != nil {
			return err
		}
		batch, err := j.claim(ctx)
		if err != nil {
			return fmt.Errorf("auth email delivery: claim: %w", err)
		}
		if err := j.deliverBatch(ctx, batch, &counts); err != nil {
			return err
		}
		if len(batch) < j.deps.BatchSize {
			return nil
		}
	}
	return nil
}

func (j *EmailDeliveryJob) claim(ctx context.Context) ([]ClaimedEmail, error) {
	var batch []ClaimedEmail
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		batch, err = j.deps.Queue.Claim(ctx, j.deps.BatchSize, DeliveryLease)
		return err
	})
	return batch, err
}

// deliverBatch sends and settles each email of batch. It stops sending when
// ctx ends; the unsent ones are retried when their lease ends.
func (j *EmailDeliveryJob) deliverBatch(ctx context.Context, batch []ClaimedEmail, counts *deliveryTally) error {
	var errs []error
	for _, email := range batch {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		if err := j.deliver(ctx, email, counts); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// deliver sends one email and records the outcome.
func (j *EmailDeliveryJob) deliver(ctx context.Context, email ClaimedEmail, counts *deliveryTally) error {
	if email.Unreadable {
		j.deps.Logger.ErrorContext(ctx, "auth email dropped: unreadable payload", slog.String("job", deliveryJobName))
		counts.dropped++
		return j.settle(ctx, email, j.deps.Queue.Finish, counts)
	}
	sendErr := j.send(ctx, email.Email)
	switch {
	case sendErr == nil:
		counts.delivered++
		return j.settle(ctx, email, j.deps.Queue.Finish, counts)
	case email.Attempt >= MaxDeliveryAttempts:
		j.deps.Logger.WarnContext(ctx, "auth email dropped: attempts exhausted",
			slog.String("job", deliveryJobName), slog.Int("attempt", email.Attempt), slog.Any("error", sendErr))
		counts.dropped++
		return j.settle(ctx, email, j.deps.Queue.Finish, counts)
	default:
		j.deps.Logger.WarnContext(ctx, "auth email not sent: will retry",
			slog.String("job", deliveryJobName), slog.Int("attempt", email.Attempt), slog.Any("error", sendErr))
		counts.retried++
		retry := func(ctx context.Context, id, claim string) (bool, error) {
			return j.deps.Queue.Retry(ctx, id, claim, DeliveryBackoff(email.Attempt))
		}
		return j.settle(ctx, email, retry, counts)
	}
}

// send dispatches on the kind. EmailSender implementations never put the
// address or the code in their errors.
func (j *EmailDeliveryJob) send(ctx context.Context, email OutboundEmail) error {
	switch email.Kind {
	case EmailVerificationCode:
		return j.deps.Sender.SendVerificationCode(ctx, email.To, email.Code)
	case EmailSignUpAttemptNotice:
		return j.deps.Sender.SendSignUpAttemptNotice(ctx, email.To)
	default:
		return fmt.Errorf("unknown email kind %d", email.Kind)
	}
}

// settle runs op (Finish or Retry) for email in its own unit of work, on a
// context the run's cancellation does not reach: an email already sent must
// be recorded, or it would be sent again when its lease ends.
func (j *EmailDeliveryJob) settle(ctx context.Context, email ClaimedEmail, op func(context.Context, string, string) (bool, error), counts *deliveryTally) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), settleTimeout)
	defer cancel()
	var held bool
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		held, err = op(ctx, email.ID, email.Claim)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth email delivery: settle: %w", err)
	}
	if !held {
		counts.lost++
	}
	return nil
}

// report logs the run's counts (metrics only, no PII). Silent when the run
// found nothing, so an idle worker does not log every 5 seconds.
func (j *EmailDeliveryJob) report(ctx context.Context, counts *deliveryTally) {
	if *counts == (deliveryTally{}) {
		return
	}
	j.deps.Logger.InfoContext(ctx, "auth email delivery finished",
		slog.String("job", deliveryJobName),
		slog.Int("delivered", counts.delivered), slog.Int("retried", counts.retried),
		slog.Int("dropped", counts.dropped), slog.Int("lost", counts.lost))
}
