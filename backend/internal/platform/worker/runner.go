package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
)

// Defaults of a Runner.
const (
	// DefaultHeartbeat is how often leadership is campaigned for or verified:
	// the longest a lost leadership goes unnoticed, and the longest a standby
	// waits before taking over.
	DefaultHeartbeat = 10 * time.Second
	// DefaultElectionTimeout bounds one TryLead.
	DefaultElectionTimeout = 5 * time.Second
	// DefaultResignTimeout bounds the Resign at shutdown.
	DefaultResignTimeout = 5 * time.Second
)

// Elector is the leader-election port (pgdb.AdvisoryLeader in production).
type Elector interface {
	// TryLead acquires leadership, or confirms it is still held. false with a
	// nil error means another instance leads. An error means leadership is
	// not held (any lock this instance had is gone).
	TryLead(ctx context.Context) (bool, error)
	// Resign gives leadership up and releases the Elector's resources.
	Resign(ctx context.Context) error
}

// ErrJobPanicked wraps the value a panicking job raised.
var ErrJobPanicked = errors.New("worker: job panicked")

// Runner schedules the jobs of a Registry on the instance that leads.
type Runner struct {
	entries         []entry
	elector         Elector
	heartbeat       time.Duration
	electionTimeout time.Duration
	resignTimeout   time.Duration
	newTicker       NewTicker
	clock           clock.Clock
	logger          *slog.Logger
	lead            leadership
}

// Option customizes NewRunner.
type Option func(*Runner)

// WithHeartbeat sets how often leadership is campaigned for (DefaultHeartbeat).
func WithHeartbeat(d time.Duration) Option { return func(r *Runner) { r.heartbeat = d } }

// WithTicker replaces SystemTicker (tests fire ticks by hand).
func WithTicker(f NewTicker) Option { return func(r *Runner) { r.newTicker = f } }

// WithClock replaces the system clock used to time runs.
func WithClock(c clock.Clock) Option { return func(r *Runner) { r.clock = c } }

// WithLogger sets the logger (logging.New in cmd/); the default discards.
func WithLogger(l *slog.Logger) Option { return func(r *Runner) { r.logger = l } }

// NewRunner builds a Runner over the jobs registered in reg so far.
func NewRunner(reg *Registry, elector Elector, opts ...Option) (*Runner, error) {
	if reg == nil || elector == nil {
		return nil, errors.New("worker: NewRunner needs a registry and an elector")
	}
	r := &Runner{
		entries:         append([]entry(nil), reg.entries...),
		elector:         elector,
		heartbeat:       DefaultHeartbeat,
		electionTimeout: DefaultElectionTimeout,
		resignTimeout:   DefaultResignTimeout,
		newTicker:       SystemTicker,
		clock:           clock.NewSystem(),
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.heartbeat <= 0 {
		return nil, fmt.Errorf("worker: heartbeat %s is not positive", r.heartbeat)
	}
	return r, nil
}

// Leading reports whether this instance currently holds leadership.
func (r *Runner) Leading() bool {
	_, ok := r.lead.current()
	return ok
}

// Run campaigns for leadership and schedules every job until ctx is done.
// Shutdown cancels the running jobs, waits for them to return and only then
// resigns, so the next leader never overlaps this one's work. It returns the
// Resign error, if any.
func (r *Runner) Run(ctx context.Context) error {
	r.logger.InfoContext(ctx, "worker started", slog.Any("jobs", r.names()))
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return r.campaign(gctx) })
	for _, e := range r.entries {
		g.Go(func() error { return r.schedule(gctx, e) })
	}
	_ = g.Wait() // the loops only return nil, once ctx is done
	r.lead.lose()

	resignCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.resignTimeout)
	defer cancel()
	if err := r.elector.Resign(resignCtx); err != nil {
		r.logger.WarnContext(ctx, "worker stopped; resign failed", slog.Any("error", err))
		return fmt.Errorf("worker: resign: %w", err)
	}
	r.logger.InfoContext(ctx, "worker stopped")
	return nil
}

// campaign tries to lead now and on every heartbeat until ctx is done.
func (r *Runner) campaign(ctx context.Context) error {
	t := r.newTicker(r.heartbeat)
	defer t.Stop()
	for {
		r.elect(ctx)
		select {
		case <-ctx.Done():
			r.lead.lose()
			return nil
		case <-t.C():
		}
	}
}

// elect runs one election. Shutdown does not cancel it, only electionTimeout
// bounds it: a lock check aborted mid-query would cost the elector its session,
// so Run's Resign would find nothing to unlock and the lock would outlive Run
// until PostgreSQL noticed the closed session, delaying the standby's takeover.
func (r *Runner) elect(ctx context.Context) {
	tryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.electionTimeout)
	defer cancel()
	leading, err := r.elector.TryLead(tryCtx)
	if err != nil {
		r.logger.WarnContext(ctx, "leader election failed", slog.Any("error", err))
		leading = false
	}
	if leading && r.lead.acquire(ctx) {
		r.logger.InfoContext(ctx, "leadership acquired")
	}
	if !leading && r.lead.lose() {
		r.logger.WarnContext(ctx, "leadership lost")
	}
}

// schedule runs e on every tick of its interval while this instance leads.
// Runs of one job never overlap: ticks that arrive during a run are dropped.
func (r *Runner) schedule(ctx context.Context, e entry) error {
	t := r.newTicker(e.job.Interval())
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C():
			r.runIfLeader(ctx, e)
		}
	}
}

func (r *Runner) runIfLeader(ctx context.Context, e entry) {
	leaderCtx, ok := r.lead.current()
	if !ok || leaderCtx.Err() != nil {
		r.logger.DebugContext(ctx, "job skipped: not the leader", slog.String("job", e.job.Name()))
		return
	}
	runCtx, cancel := context.WithTimeout(leaderCtx, e.timeout)
	defer cancel()
	started := r.clock.Now()
	err := runRecovering(runCtx, e.job)
	r.logOutcome(ctx, e.job.Name(), r.clock.Now().Sub(started), err, runCtx.Err())
}

// runRecovering runs job, turning a panic into an error wrapping ErrJobPanicked.
func runRecovering(ctx context.Context, job Job) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%w: %v\n%s", ErrJobPanicked, v, debug.Stack())
		}
	}()
	return job.Run(ctx)
}

// logOutcome logs one run. Only the job name, the duration and the error go
// out, through the redacting logger: no user data.
//
// The outcome is read from the job's error. A run counts as timed out or
// cancelled only when that error is the run context's own error: the context
// can be cancelled (shutdown, lost leadership) after the job already returned
// an unrelated error, and that failure must still be logged as one.
func (r *Runner) logOutcome(ctx context.Context, name string, took time.Duration, err, ctxErr error) {
	attrs := []any{slog.String("job", name), slog.Duration("duration", took)}
	switch {
	case err == nil:
		r.logger.InfoContext(ctx, "job finished", attrs...)
	case errors.Is(err, ErrJobPanicked):
		r.logger.ErrorContext(ctx, "job panicked", append(attrs, slog.Any("error", err))...)
	case causedBy(err, ctxErr, context.DeadlineExceeded):
		r.logger.WarnContext(ctx, "job timed out", append(attrs, slog.Any("error", err))...)
	case causedBy(err, ctxErr, context.Canceled):
		r.logger.InfoContext(ctx, "job cancelled", append(attrs, slog.Any("error", err))...)
	default:
		r.logger.ErrorContext(ctx, "job failed", append(attrs, slog.Any("error", err))...)
	}
}

// causedBy reports whether the job's error is the run context's error
// target: the job returned (or wrapped) it and the context really ended so.
func causedBy(jobErr, ctxErr, target error) bool {
	return errors.Is(jobErr, target) && errors.Is(ctxErr, target)
}

func (r *Runner) names() []string {
	names := make([]string, 0, len(r.entries))
	for _, e := range r.entries {
		names = append(names, e.job.Name())
	}
	return names
}

// leadership is the current term: a context that lives while this instance
// leads and is cancelled the moment it stops leading.
type leadership struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

// acquire starts a term under parent; false if one is already running.
func (l *leadership) acquire(parent context.Context) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		return false
	}
	l.ctx, l.cancel = context.WithCancel(parent)
	return true
}

// lose ends the running term, cancelling its runs; false if there was none.
func (l *leadership) lose() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel == nil {
		return false
	}
	l.cancel()
	l.ctx, l.cancel = nil, nil
	return true
}

// current returns the running term's context.
func (l *leadership) current() (context.Context, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.ctx, l.cancel != nil
}
