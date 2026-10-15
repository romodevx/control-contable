// Package usecase holds the records retention job (M9-08): the worker's
// periodic purge of soft-deleted records 30 days after their deletion and of
// dead auth state (expired or revoked sessions, expired sign-ups, ended
// verification-failure windows).
//
// The rules (what goes, when, and who may delete it) live in the database:
// the SECURITY DEFINER function purge_soft_deleted, owned by
// cc_worker_definer, computes its own cutoffs and its policies match only
// rows past them. This job only drives it: bounded passes, each in its own
// short System unit of work (no user context), until a pass comes back short
// of the batch size or the per-run pass limit is reached. It runs whatever
// the module flags say: retention is a legal promise, not a feature.
//
// Idempotent and safe under a worker split brain: a second, overlapping run
// finds the rows gone (the database re-checks every row it deletes), and a
// re-run purges nothing new. Only counts are logged, never an id or a value.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	// jobName identifies the job in the worker's logs.
	jobName = "records_purge"
	// jobInterval is the time between two runs. Hourly keeps dead sessions and
	// sign-ups from lingering much past their one-day grace; a run that finds
	// nothing costs a few index probes.
	jobInterval = time.Hour

	// MaxBatchSize mirrors the bound purge_soft_deleted enforces.
	MaxBatchSize = 5000
	// DefaultBatchSize rows per target per pass: a pass stays well inside the
	// worker's 60 s statement timeout and holds few row locks.
	DefaultBatchSize = 1000
	// DefaultMaxBatches passes per run at most; the rest waits for the next run.
	DefaultMaxBatches = 50
)

// ErrInvalidConfig reports a PurgeJob built with missing or out-of-range deps.
var ErrInvalidConfig = errors.New("retention: invalid purge job configuration")

// TargetCount is how many rows one pass deleted from one target.
type TargetCount struct {
	Target string
	Purged int
}

// Purger runs one bounded purge pass inside the unit of work in ctx.
type Purger interface {
	PurgeBatch(ctx context.Context, batchSize int) ([]TargetCount, error)
}

// PurgeDeps are what a PurgeJob needs.
type PurgeDeps struct {
	System     tx.SystemManager
	Purger     Purger
	Logger     *slog.Logger
	BatchSize  int
	MaxBatches int
}

// PurgeJob is the worker job (worker.Job) that drives the purge.
type PurgeJob struct {
	deps PurgeDeps
}

// NewPurgeJob validates deps and builds the job.
func NewPurgeJob(deps PurgeDeps) (*PurgeJob, error) {
	switch {
	case deps.System == nil, deps.Purger == nil, deps.Logger == nil:
		return nil, fmt.Errorf("%w: nil dependency", ErrInvalidConfig)
	case deps.BatchSize < 1 || deps.BatchSize > MaxBatchSize:
		return nil, fmt.Errorf("%w: batch size %d not in [1, %d]", ErrInvalidConfig, deps.BatchSize, MaxBatchSize)
	case deps.MaxBatches < 1:
		return nil, fmt.Errorf("%w: max batches %d is not positive", ErrInvalidConfig, deps.MaxBatches)
	}
	return &PurgeJob{deps: deps}, nil
}

// Name implements worker.Job.
func (j *PurgeJob) Name() string { return jobName }

// Interval implements worker.Job.
func (j *PurgeJob) Interval() time.Duration { return jobInterval }

// Run implements worker.Job: passes until one comes back short, MaxBatches is
// reached, a pass fails or ctx ends. The counts of the committed passes are
// logged in every case.
func (j *PurgeJob) Run(ctx context.Context) error {
	var totals tally
	more := false
	for range j.deps.MaxBatches {
		if err := ctx.Err(); err != nil {
			j.report(ctx, totals, more)
			return err
		}
		counts, err := j.pass(ctx)
		if err != nil {
			j.report(ctx, totals, more)
			return fmt.Errorf("records purge pass: %w", err)
		}
		totals.add(counts)
		more = j.filled(counts)
		if !more {
			break
		}
	}
	j.report(ctx, totals, more)
	return nil
}

// pass runs one purge pass in its own System unit of work, so each pass
// commits on its own and holds its row locks briefly.
func (j *PurgeJob) pass(ctx context.Context) ([]TargetCount, error) {
	var counts []TargetCount
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		counts, err = j.deps.Purger.PurgeBatch(ctx, j.deps.BatchSize)
		return err
	})
	return counts, err
}

// filled reports whether any target deleted a full batch, so rows may remain.
func (j *PurgeJob) filled(counts []TargetCount) bool {
	for _, c := range counts {
		if c.Purged >= j.deps.BatchSize {
			return true
		}
	}
	return false
}

// report logs the run's counts per target (metrics: counts only, no PII).
func (j *PurgeJob) report(ctx context.Context, totals tally, more bool) {
	purged := make([]any, 0, len(totals.order))
	for _, target := range totals.order {
		purged = append(purged, slog.Int(target, totals.purged[target]))
	}
	j.deps.Logger.InfoContext(ctx, "records purge finished",
		slog.String("job", jobName), slog.Int("passes", totals.passes), slog.Bool("more", more),
		slog.Group("purged", purged...))
}

// tally accumulates pass counts per target, keeping the targets' first-seen order.
type tally struct {
	passes int
	order  []string
	purged map[string]int
}

func (t *tally) add(counts []TargetCount) {
	if t.purged == nil {
		t.purged = make(map[string]int, len(counts))
	}
	t.passes++
	for _, c := range counts {
		if _, seen := t.purged[c.Target]; !seen {
			t.order = append(t.order, c.Target)
		}
		t.purged[c.Target] += c.Purged
	}
}
