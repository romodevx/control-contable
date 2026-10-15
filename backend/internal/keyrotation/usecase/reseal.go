// Package usecase holds the key re-seal job (ADR-0012): the worker job that
// moves every sealed column (users' emails, taxpayers' RFCs, devices' push
// tokens) to its active key after a rotation, and reports how many rows are
// still on an old key or pepper, the readiness figure that says when a
// retired key may be removed.
//
// Each column is re-sealed in bounded passes: a System unit of work lists the
// next page of rows sealed with another key (ids only, keyset on the id), and
// every row is then re-sealed in a unit of work of its own, as its owner (RLS
// scopes the rewrite like the API's): lock the row, check it still needs it,
// open with whichever key sealed it, seal with the active key, write.
//
// There is no checkpoint table: the row's key id is the checkpoint. A
// re-sealed row leaves the list, so a run cut short anywhere (timeout,
// shutdown, lost leadership, crash) resumes at the next run, and two runs at
// once (a worker split brain) serialize on the row lock and the second finds
// the row already on the active key. A row the keyring cannot open (a key it
// lacks, a ciphertext that fails authentication) is counted, left as it is
// and passed over; the readiness count keeps showing it. Only counts are
// logged: never an id, a ciphertext or a plaintext.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	// jobName identifies the job in the worker's logs.
	jobName = "key_reseal"
	// jobInterval is the time between two runs. Outside a rotation a run costs
	// one empty page per column and the key_usage counts.
	jobInterval = 15 * time.Minute
	// RunTimeout bounds one run; a run cut short resumes at the next.
	RunTimeout = 10 * time.Minute

	// MaxBatchSize mirrors the bound the reseal_pending_* functions enforce.
	MaxBatchSize = 1000
	// DefaultBatchSize rows per page: one short System transaction per page.
	DefaultBatchSize = 200
	// DefaultMaxBatches pages per column per run at most.
	DefaultMaxBatches = 50
)

var (
	// ErrInvalidConfig reports a ResealJob built with missing or out-of-range deps.
	ErrInvalidConfig = errors.New("keyrotation: invalid reseal job configuration")
	// errUnreadable marks a row the keyring cannot open; it is counted, not fatal.
	errUnreadable = errors.New("keyrotation: sealed value cannot be opened")
)

// Ref names one sealed row: its primary key and the user whose unit of work
// may rewrite it.
type Ref struct {
	ID     string
	UserID string
}

// Store is one sealed column's persistence.
type Store interface {
	// Pending lists, in a System unit of work, at most limit rows with an id
	// above after ("" = from the start), ordered by id, whose value is
	// sealed with another key than active.
	Pending(ctx context.Context, active crypto.KeyID, after string, limit int) ([]Ref, error)
	// Lock reads ref's sealed value, locking the row until the unit of work
	// (ref's owner's) ends. found is false when the row is gone.
	Lock(ctx context.Context, ref Ref) (ciphertext []byte, found bool, err error)
	// Replace writes ref's new sealed value and its key id.
	Replace(ctx context.Context, ref Ref, ciphertext []byte, keyID crypto.KeyID) error
}

// Sealer re-seals one column's values: the module that owns the column
// binds each value to its row (the GCM additional data).
type Sealer interface {
	ActiveKeyID() crypto.KeyID
	Reseal(ciphertext []byte, ref Ref) ([]byte, error)
}

// Column is one sealed column the job holds the keys of. Target names it as
// key_usage does (table.column).
type Column struct {
	Target string
	Store  Store
	Sealer Sealer
}

// Usage is one key_usage row: how many rows of a target are on one key id
// (sealed columns) or pepper version (blind indexes).
type Usage struct {
	Target  string
	Version int
	Rows    int64
}

// UsageReader reads key_usage in a System unit of work.
type UsageReader interface {
	KeyUsage(ctx context.Context) ([]Usage, error)
}

// ResealDeps are what a ResealJob needs. Columns may be empty: the job then
// only reports readiness.
type ResealDeps struct {
	Tx         tx.Manager
	System     tx.SystemManager
	Columns    []Column
	Usage      UsageReader
	Logger     *slog.Logger
	BatchSize  int
	MaxBatches int
}

// ResealJob is the worker job (worker.Job) that re-seals and reports.
type ResealJob struct {
	deps ResealDeps
}

// NewResealJob validates deps and builds the job.
func NewResealJob(deps ResealDeps) (*ResealJob, error) {
	switch {
	case deps.Tx == nil, deps.System == nil, deps.Usage == nil, deps.Logger == nil:
		return nil, fmt.Errorf("%w: nil dependency", ErrInvalidConfig)
	case deps.BatchSize < 1 || deps.BatchSize > MaxBatchSize:
		return nil, fmt.Errorf("%w: batch size %d not in [1, %d]", ErrInvalidConfig, deps.BatchSize, MaxBatchSize)
	case deps.MaxBatches < 1:
		return nil, fmt.Errorf("%w: max batches %d is not positive", ErrInvalidConfig, deps.MaxBatches)
	}
	seen := make(map[string]bool, len(deps.Columns))
	for _, c := range deps.Columns {
		if c.Target == "" || c.Store == nil || c.Sealer == nil || seen[c.Target] {
			return nil, fmt.Errorf("%w: column %q is incomplete or repeated", ErrInvalidConfig, c.Target)
		}
		seen[c.Target] = true
	}
	return &ResealJob{deps: deps}, nil
}

// Name implements worker.Job.
func (j *ResealJob) Name() string { return jobName }

// Interval implements worker.Job.
func (j *ResealJob) Interval() time.Duration { return jobInterval }

// columnResult is one column's counts in one run.
type columnResult struct {
	target     string
	active     crypto.KeyID
	resealed   int
	skipped    int
	unreadable int
	more       bool
}

// Run implements worker.Job: every column in turn, then the readiness
// counts. A store error or the end of ctx stops the run; the counts so far
// are logged in every case.
func (j *ResealJob) Run(ctx context.Context) error {
	results := make([]columnResult, 0, len(j.deps.Columns))
	for _, c := range j.deps.Columns {
		res, err := j.resealColumn(ctx, c)
		results = append(results, res)
		if err != nil {
			j.report(ctx, results, nil)
			return fmt.Errorf("key reseal %s: %w", c.Target, err)
		}
	}
	var usage []Usage
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		usage, err = j.deps.Usage.KeyUsage(ctx)
		return err
	})
	j.report(ctx, results, usage)
	if err != nil {
		return fmt.Errorf("key usage: %w", err)
	}
	return nil
}

// resealColumn walks c's pending rows page by page, at most MaxBatches pages.
func (j *ResealJob) resealColumn(ctx context.Context, c Column) (columnResult, error) {
	res := columnResult{target: c.Target, active: c.Sealer.ActiveKeyID()}
	after := ""
	for range j.deps.MaxBatches {
		page, err := j.pending(ctx, c, res.active, after)
		if err != nil {
			return res, err
		}
		for _, ref := range page {
			if err := ctx.Err(); err != nil {
				return res, err
			}
			if err := j.resealRow(ctx, c, ref, &res); err != nil {
				return res, err
			}
		}
		res.more = len(page) == j.deps.BatchSize
		if !res.more {
			return res, nil
		}
		after = page[len(page)-1].ID
	}
	return res, nil
}

func (j *ResealJob) pending(ctx context.Context, c Column, active crypto.KeyID, after string) ([]Ref, error) {
	var page []Ref
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		page, err = c.Store.Pending(ctx, active, after, j.deps.BatchSize)
		return err
	})
	return page, err
}

// resealRow re-seals one row in its owner's unit of work and counts the
// outcome into res. Only a store or transaction error is returned.
func (j *ResealJob) resealRow(ctx context.Context, c Column, ref Ref, res *columnResult) error {
	resealed := false
	err := j.deps.Tx.ReadWrite(ctx, ref.UserID, func(ctx context.Context) error {
		ciphertext, found, err := c.Store.Lock(ctx, ref)
		if err != nil || !found {
			return err
		}
		keyID, err := crypto.KeyIDOf(ciphertext)
		if err != nil {
			return fmt.Errorf("%w: %w", errUnreadable, err)
		}
		if keyID == res.active {
			return nil
		}
		next, err := c.Sealer.Reseal(ciphertext, ref)
		if err != nil {
			return fmt.Errorf("%w: %w", errUnreadable, err)
		}
		if err := c.Store.Replace(ctx, ref, next, res.active); err != nil {
			return err
		}
		resealed = true
		return nil
	})
	switch {
	case errors.Is(err, errUnreadable):
		res.unreadable++
	case err != nil:
		return err
	case resealed:
		res.resealed++
	default:
		res.skipped++
	}
	return nil
}

// Log messages: the readiness figures are read from these lines (log-based
// metrics). Target names are attribute values, never keys: the platform
// logger redacts every attribute whose key mentions an email, a token or an
// RFC, which would hide the counts.
const (
	msgColumn   = "key reseal column"
	msgUsage    = "key usage"
	msgFinished = "key reseal finished"
)

// report logs counts only: one line per column this job holds (what this run
// did and rows_on_old_keys, the readiness figure), one per key_usage row
// (every target, held or not) and a closing line with the totals. A column
// with unreadable rows is a warning: an operator must look at it.
func (j *ResealJob) report(ctx context.Context, results []columnResult, usage []Usage) {
	var resealed, unreadable int
	var onOldKeys int64
	for _, r := range results {
		level := slog.LevelInfo
		if r.unreadable > 0 {
			level = slog.LevelWarn
		}
		old := rowsOnOldKeys(usage, r.target, r.active)
		j.deps.Logger.Log(ctx, level, msgColumn,
			slog.String("column", r.target), slog.Int("active_key_id", int(r.active)),
			slog.Int("resealed", r.resealed), slog.Int("skipped", r.skipped),
			slog.Int("unreadable", r.unreadable), slog.Bool("more", r.more),
			slog.Int64("rows_on_old_keys", old))
		resealed, unreadable, onOldKeys = resealed+r.resealed, unreadable+r.unreadable, onOldKeys+old
	}
	for _, u := range usage {
		j.deps.Logger.LogAttrs(ctx, slog.LevelInfo, msgUsage,
			slog.String("target", u.Target), slog.Int("version", u.Version), slog.Int64("rows", u.Rows))
	}
	j.deps.Logger.LogAttrs(ctx, slog.LevelInfo, msgFinished,
		slog.Int("columns", len(results)), slog.Int("resealed", resealed),
		slog.Int("unreadable", unreadable), slog.Int64("rows_on_old_keys", onOldKeys))
}

// rowsOnOldKeys sums target's rows on any key but active.
func rowsOnOldKeys(usage []Usage, target string, active crypto.KeyID) int64 {
	var n int64
	for _, u := range usage {
		if u.Target == target && u.Version != int(active) {
			n += u.Rows
		}
	}
	return n
}
