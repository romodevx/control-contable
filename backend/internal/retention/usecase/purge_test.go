package usecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/retention/usecase"
)

// fakeSystem runs work inline, as tx.SystemManager would, counting the units of work.
type fakeSystem struct{ calls int }

func (f *fakeSystem) System(ctx context.Context, work tx.Work) error {
	f.calls++
	return work(ctx)
}

// fakePurger replays one scripted pass per call and records the batch sizes asked for.
type fakePurger struct {
	passes  [][]usecase.TargetCount
	err     error
	errAt   int // 1-based call that fails; 0 = never
	calls   int
	batches []int
	inUnit  []bool
}

type unitKey struct{}

func (f *fakePurger) PurgeBatch(ctx context.Context, batchSize int) ([]usecase.TargetCount, error) {
	f.calls++
	f.batches = append(f.batches, batchSize)
	f.inUnit = append(f.inUnit, ctx.Value(unitKey{}) != nil)
	if f.calls == f.errAt {
		return nil, f.err
	}
	if f.calls > len(f.passes) {
		return f.passes[len(f.passes)-1], nil
	}
	return f.passes[f.calls-1], nil
}

// markingSystem tags the ctx it hands to work, so the purger can tell it ran inside.
type markingSystem struct{ fakeSystem }

func (m *markingSystem) System(ctx context.Context, work tx.Work) error {
	return m.fakeSystem.System(ctx, func(ctx context.Context) error {
		return work(context.WithValue(ctx, unitKey{}, true))
	})
}

func pass(incomes, tokens int) []usecase.TargetCount {
	return []usecase.TargetCount{{Target: "incomes", Purged: incomes}, {Target: "refresh_tokens", Purged: tokens}}
}

const testBatch = 10

func newJob(t *testing.T, system tx.SystemManager, purger usecase.Purger, logs *bytes.Buffer, maxBatches int) *usecase.PurgeJob {
	t.Helper()
	job, err := usecase.NewPurgeJob(usecase.PurgeDeps{
		System: system, Purger: purger, Logger: slog.New(slog.NewJSONHandler(logs, nil)),
		BatchSize: testBatch, MaxBatches: maxBatches,
	})
	require.NoError(t, err)
	return job
}

// lastLog decodes the last JSON log line.
func lastLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n"))
	var entry map[string]any
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &entry))
	return entry
}

func TestPurgeJob_NameAndInterval(t *testing.T) {
	job := newJob(t, &fakeSystem{}, &fakePurger{passes: [][]usecase.TargetCount{pass(0, 0)}}, &bytes.Buffer{}, 3)
	assert.Equal(t, "records_purge", job.Name())
	assert.Equal(t, time.Hour, job.Interval())
}

// A pass in which any target filled its batch may have left rows behind: run another,
// each in its own unit of work, until a pass comes back short.
func TestPurgeJob_LoopsWhileATargetFillsItsBatch(t *testing.T) {
	system := &markingSystem{}
	purger := &fakePurger{passes: [][]usecase.TargetCount{pass(testBatch, 3), pass(2, testBatch), pass(0, 4)}}
	var logs bytes.Buffer

	require.NoError(t, newJob(t, system, purger, &logs, 5).Run(t.Context()))

	assert.Equal(t, 3, purger.calls)
	assert.Equal(t, 3, system.calls, "one short transaction per pass")
	assert.Equal(t, []int{testBatch, testBatch, testBatch}, purger.batches)
	assert.Equal(t, []bool{true, true, true}, purger.inUnit, "every pass runs inside the System unit of work")
	entry := lastLog(t, &logs)
	assert.Equal(t, "records purge finished", entry["msg"])
	assert.EqualValues(t, 3, entry["passes"])
	assert.Equal(t, map[string]any{"incomes": float64(12), "refresh_tokens": float64(17)}, entry["purged"])
}

// MaxBatches bounds one run; what is left goes on the next run.
func TestPurgeJob_StopsAtMaxBatches(t *testing.T) {
	purger := &fakePurger{passes: [][]usecase.TargetCount{pass(testBatch, 0)}}
	var logs bytes.Buffer

	require.NoError(t, newJob(t, &fakeSystem{}, purger, &logs, 4).Run(t.Context()))

	assert.Equal(t, 4, purger.calls)
	entry := lastLog(t, &logs)
	assert.EqualValues(t, 4, entry["passes"])
	assert.Equal(t, true, entry["more"], "rows remain for the next run")
}

// A failed pass stops the run; the counts of the passes already committed are still
// reported.
func TestPurgeJob_FailedPassStopsTheRun(t *testing.T) {
	boom := errors.New("boom")
	purger := &fakePurger{passes: [][]usecase.TargetCount{pass(testBatch, 0)}, err: boom, errAt: 2}
	var logs bytes.Buffer

	err := newJob(t, &fakeSystem{}, purger, &logs, 5).Run(t.Context())

	require.ErrorIs(t, err, boom)
	assert.Equal(t, 2, purger.calls)
	entry := lastLog(t, &logs)
	assert.EqualValues(t, 1, entry["passes"])
	assert.Equal(t, map[string]any{"incomes": float64(testBatch), "refresh_tokens": float64(0)}, entry["purged"])
}

// A cancelled run (timeout, lost leadership, shutdown) starts no further pass.
func TestPurgeJob_HonorsCancellation(t *testing.T) {
	purger := &fakePurger{passes: [][]usecase.TargetCount{pass(testBatch, 0)}}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := newJob(t, &fakeSystem{}, purger, &bytes.Buffer{}, 5).Run(ctx)

	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, purger.calls)
}

func TestNewPurgeJob_RejectsInvalidDeps(t *testing.T) {
	valid := func() usecase.PurgeDeps {
		return usecase.PurgeDeps{
			System: &fakeSystem{}, Purger: &fakePurger{}, Logger: slog.New(slog.DiscardHandler),
			BatchSize: testBatch, MaxBatches: 1,
		}
	}
	cases := map[string]func(*usecase.PurgeDeps){
		"nil system":           func(d *usecase.PurgeDeps) { d.System = nil },
		"nil purger":           func(d *usecase.PurgeDeps) { d.Purger = nil },
		"nil logger":           func(d *usecase.PurgeDeps) { d.Logger = nil },
		"zero batch":           func(d *usecase.PurgeDeps) { d.BatchSize = 0 },
		"batch over the limit": func(d *usecase.PurgeDeps) { d.BatchSize = usecase.MaxBatchSize + 1 },
		"zero max batches":     func(d *usecase.PurgeDeps) { d.MaxBatches = 0 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			deps := valid()
			mutate(&deps)
			_, err := usecase.NewPurgeJob(deps)
			require.ErrorIs(t, err, usecase.ErrInvalidConfig)
		})
	}
	_, err := usecase.NewPurgeJob(valid())
	require.NoError(t, err)
}
