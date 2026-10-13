package worker_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/worker"
	"github.com/romodevx/control-contable/backend/internal/platform/worker/workertest"
)

// Distinct intervals: workertest keys its manual tickers by interval.
const (
	heartbeat   = 7 * time.Second
	jobInterval = 5 * time.Minute
	jobTimeout  = time.Minute
)

// fakeElector is a scripted Elector: TryLead answers lead/err, as set by the test.
type fakeElector struct {
	mu        sync.Mutex
	lead      bool
	err       error
	tries     int
	resigned  int
	resignErr error
	onResign  func()
}

func (f *fakeElector) set(lead bool, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lead, f.err = lead, err
}

func (f *fakeElector) TryLead(context.Context) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tries++
	return f.lead && f.err == nil, f.err
}

func (f *fakeElector) Resign(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resigned++
	if f.onResign != nil {
		f.onResign()
	}
	return f.resignErr
}

func (f *fakeElector) counts() (tries, resigned int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tries, f.resigned
}

// probeJob is the test-only job: each run hands its ctx to the test on runs,
// then behaves as body says (nil body: succeed at once).
type probeJob struct {
	name string
	runs chan context.Context
	body func(ctx context.Context) error
}

func newProbeJob(body func(ctx context.Context) error) *probeJob {
	return &probeJob{name: "probe", runs: make(chan context.Context, 16), body: body}
}

func (j *probeJob) Name() string            { return j.name }
func (j *probeJob) Interval() time.Duration { return jobInterval }

func (j *probeJob) Run(ctx context.Context) error {
	j.runs <- ctx
	if j.body == nil {
		return nil
	}
	return j.body(ctx)
}

// harness runs a Runner in the background over manual ticks and a log buffer.
type harness struct {
	runner  *worker.Runner
	ticks   *workertest.Ticks
	elector *fakeElector
	logs    *syncBuffer
	cancel  context.CancelFunc
	done    chan error
}

func start(t *testing.T, elector *fakeElector, jobs ...*probeJob) *harness {
	t.Helper()
	var reg worker.Registry
	for _, j := range jobs {
		require.NoError(t, reg.Register(j, jobTimeout))
	}
	h := &harness{ticks: workertest.NewTicks(), elector: elector, logs: &syncBuffer{}, done: make(chan error, 1)}
	runner, err := worker.NewRunner(&reg, elector,
		worker.WithHeartbeat(heartbeat),
		worker.WithTicker(h.ticks.NewTicker),
		worker.WithClock(clock.NewFixed(time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC))),
		worker.WithLogger(logging.New(h.logs, slog.LevelDebug)),
	)
	require.NoError(t, err)
	h.runner = runner
	ctx, cancel := context.WithCancel(t.Context())
	h.cancel = cancel
	go func() { h.done <- runner.Run(ctx) }()
	t.Cleanup(func() { assert.NoError(t, h.stop(t)) })
	return h
}

// beat fires the heartbeat twice: the second tick is received only after the
// first election finished, so leadership is settled when beat returns.
func (h *harness) beat(t *testing.T) {
	t.Helper()
	h.ticks.Tick(t, heartbeat)
	h.ticks.Tick(t, heartbeat)
}

// stop cancels Run and waits for it; it is idempotent.
func (h *harness) stop(t *testing.T) error {
	t.Helper()
	h.cancel()
	err, ok := <-h.done // t.Context is already done in Cleanup, so no select on it
	if ok {
		close(h.done)
	}
	return err
}

func receive(t *testing.T, runs <-chan context.Context) context.Context {
	t.Helper()
	select {
	case ctx := <-runs:
		return ctx
	case <-t.Context().Done():
		t.Fatal("the job did not run")
		return nil
	}
}

func TestNewRunner_requiresARegistryAndAnElector(t *testing.T) {
	_, err := worker.NewRunner(nil, &fakeElector{})
	require.Error(t, err)
	_, err = worker.NewRunner(&worker.Registry{}, nil)
	require.Error(t, err)
	_, err = worker.NewRunner(&worker.Registry{}, &fakeElector{}, worker.WithHeartbeat(0))
	require.Error(t, err)
}

func TestRunner_aFollowerNeverRunsJobs(t *testing.T) {
	j := newProbeJob(nil)
	h := start(t, &fakeElector{lead: false}, j)
	h.beat(t)

	h.ticks.Tick(t, jobInterval)
	h.ticks.Tick(t, jobInterval)
	require.NoError(t, h.stop(t))

	assert.Empty(t, j.runs)
	assert.False(t, h.runner.Leading())
	assert.Contains(t, h.logs.String(), `"msg":"job skipped: not the leader"`)
}

func TestRunner_theLeaderRunsEachJobWithItsTimeout(t *testing.T) {
	j := newProbeJob(nil)
	h := start(t, &fakeElector{lead: true}, j)
	h.beat(t)
	require.True(t, h.runner.Leading())

	h.ticks.Tick(t, jobInterval)
	ctx := receive(t, j.runs)
	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "every run is bounded by the job's timeout")
	assert.WithinDuration(t, time.Now().Add(jobTimeout), deadline, jobTimeout)
	assert.Contains(t, h.logs.String(), `"msg":"job finished","job":"probe"`)
}

func TestRunner_campaignsAtStartAndOnEveryHeartbeat(t *testing.T) {
	e := &fakeElector{lead: true}
	h := start(t, e)

	h.ticks.Tick(t, heartbeat) // received only after the first, immediate election
	h.ticks.Tick(t, heartbeat)
	require.NoError(t, h.stop(t))

	tries, _ := e.counts()
	assert.Equal(t, 3, tries)
	assert.True(t, h.ticks.Stopped(heartbeat))
}

func TestRunner_losingLeadershipCancelsTheRunningJobAndSkipsTheNextOnes(t *testing.T) {
	j := newProbeJob(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
	e := &fakeElector{lead: true}
	h := start(t, e, j)
	h.beat(t)
	h.ticks.Tick(t, jobInterval)
	ctx := receive(t, j.runs)

	e.set(false, nil)
	h.beat(t)

	<-ctx.Done()
	require.ErrorIs(t, ctx.Err(), context.Canceled, "a job never outlives its leadership")
	assert.False(t, h.runner.Leading())
	h.ticks.Tick(t, jobInterval) // returns once the cancelled run has ended
	h.ticks.Tick(t, jobInterval)
	assert.Empty(t, j.runs)
	logs := h.logs.String()
	assert.Contains(t, logs, `"msg":"leadership acquired"`)
	assert.Contains(t, logs, `"msg":"leadership lost"`)
}

func TestRunner_anElectionErrorDropsLeadershipUntilTheNextHeartbeat(t *testing.T) {
	j := newProbeJob(nil)
	e := &fakeElector{lead: true}
	h := start(t, e, j)
	h.beat(t)
	require.True(t, h.runner.Leading())

	e.set(true, errors.New("leader connection lost"))
	h.beat(t)
	assert.False(t, h.runner.Leading(), "an unverifiable lock is a lost lock")
	h.ticks.Tick(t, jobInterval)
	h.ticks.Tick(t, jobInterval)
	assert.Empty(t, j.runs)

	e.set(true, nil)
	h.beat(t)
	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)
	assert.Contains(t, h.logs.String(), `"msg":"leader election failed"`)
}

func TestRunner_recoversAPanickingJob(t *testing.T) {
	calls := 0
	j := newProbeJob(func(context.Context) error {
		calls++
		if calls == 1 {
			panic("boom")
		}
		return nil
	})
	h := start(t, &fakeElector{lead: true}, j)
	h.beat(t)

	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)
	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)
	require.NoError(t, h.stop(t))

	assert.Equal(t, 2, calls, "the worker survives the panic and runs the job again")
	assert.Contains(t, h.logs.String(), `"msg":"job panicked","job":"probe"`)
}

func TestRunner_aJobPastItsTimeoutIsCancelled(t *testing.T) {
	var reg worker.Registry
	j := newProbeJob(func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
	require.NoError(t, reg.Register(j, time.Millisecond))
	ticks := workertest.NewTicks()
	logs := &syncBuffer{}
	runner, err := worker.NewRunner(&reg, &fakeElector{lead: true},
		worker.WithHeartbeat(heartbeat), worker.WithTicker(ticks.NewTicker),
		worker.WithLogger(logging.New(logs, nil)))
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx) }()

	ticks.Tick(t, heartbeat)
	ticks.Tick(t, jobInterval)
	runCtx := receive(t, j.runs)
	<-runCtx.Done()
	ticks.Tick(t, jobInterval) // barrier: the timed-out run has been logged
	receive(t, j.runs)
	cancel()
	require.NoError(t, <-done)

	assert.ErrorIs(t, runCtx.Err(), context.DeadlineExceeded)
	assert.Contains(t, logs.String(), `"msg":"job timed out","job":"probe"`)
}

func TestRunner_shutdownWaitsForRunningJobsBeforeResigning(t *testing.T) {
	jobReturned := make(chan struct{})
	j := newProbeJob(func(ctx context.Context) error {
		<-ctx.Done()
		close(jobReturned)
		return ctx.Err()
	})
	resignedAfterJob := false
	e := &fakeElector{lead: true}
	e.onResign = func() {
		select {
		case <-jobReturned:
			resignedAfterJob = true
		default:
		}
	}
	h := start(t, e, j)
	h.beat(t)
	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)

	require.NoError(t, h.stop(t))

	_, resigned := e.counts()
	assert.Equal(t, 1, resigned)
	assert.True(t, resignedAfterJob, "another instance may lead only once this one's jobs stopped")
	assert.True(t, h.ticks.Stopped(jobInterval))
	assert.False(t, h.runner.Leading())
	assert.Contains(t, h.logs.String(), `"msg":"worker stopped"`)
}

func TestRunner_reportsAFailedResign(t *testing.T) {
	e := &fakeElector{lead: true, resignErr: errors.New("connection reset")}
	h := start(t, e)
	h.beat(t)

	err := h.stop(t)

	require.ErrorContains(t, err, "connection reset")
}

// A job's error text goes through the redacting logger: PII inside it is masked.
func TestRunner_logsJobFailuresWithoutPII(t *testing.T) {
	j := newProbeJob(func(context.Context) error {
		return errors.New("send to ana@example.com failed")
	})
	h := start(t, &fakeElector{lead: true}, j)
	h.beat(t)

	h.ticks.Tick(t, jobInterval)
	receive(t, j.runs)
	require.NoError(t, h.stop(t))

	logs := h.logs.String()
	assert.Contains(t, logs, `"msg":"job failed","job":"probe"`)
	assert.NotContains(t, logs, "ana@example.com")
	for line := range strings.SplitSeq(strings.TrimSpace(logs), "\n") {
		assert.True(t, json.Valid([]byte(line)), line)
	}
}

// syncBuffer is a bytes.Buffer safe for the runner's goroutines.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
