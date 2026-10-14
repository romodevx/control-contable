//go:build integration

package main

import (
	"context"
	"database/sql"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/worker"
	"github.com/romodevx/control-contable/backend/internal/platform/worker/workertest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// Distinct intervals: workertest keys its manual tickers by interval.
const (
	testHeartbeat   = 3 * time.Second
	testJobInterval = time.Minute
)

// countingJob is the test-only job: it reports each run on runs.
type countingJob struct{ runs chan struct{} }

func (countingJob) Name() string            { return "test_counter" }
func (countingJob) Interval() time.Duration { return testJobInterval }

func (j countingJob) Run(context.Context) error {
	j.runs <- struct{}{}
	return nil
}

// instance is one worker process: its own cc_worker pool, runner and ticks.
type instance struct {
	runner *worker.Runner
	ticks  *workertest.Ticks
	job    countingJob
	cancel context.CancelFunc
	done   chan error

	stopOnce sync.Once
	stopErr  error
}

func startInstance(t *testing.T, database *pgtest.Database) *instance {
	t.Helper()
	in := &instance{
		ticks: workertest.NewTicks(),
		job:   countingJob{runs: make(chan struct{}, 8)},
		done:  make(chan error, 1),
	}
	var reg worker.Registry
	require.NoError(t, reg.Register(in.job, time.Second))
	runner, err := newRunner(database.Connect(t, pgtest.RoleWorker), &reg, logging.New(io.Discard, nil),
		worker.WithHeartbeat(testHeartbeat), worker.WithTicker(in.ticks.NewTicker))
	require.NoError(t, err)
	in.runner = runner
	ctx, cancel := context.WithCancel(context.Background())
	in.cancel = cancel
	go func() { in.done <- runner.Run(ctx) }()
	t.Cleanup(func() { assert.NoError(t, in.stop(t)) })
	return in
}

// beat settles one election: the second tick is received only after the first
// was handled.
func (in *instance) beat(t *testing.T) {
	t.Helper()
	in.ticks.Tick(t, testHeartbeat)
	in.ticks.Tick(t, testHeartbeat)
}

// mustRun ticks the job once and waits for the run.
func (in *instance) mustRun(t *testing.T) {
	t.Helper()
	in.ticks.Tick(t, testJobInterval)
	workertest.Await(t, in.job.runs, "the leader to run the job")
}

// mustSkip ticks the job twice (the second tick is received only once the
// first was handled) and checks that the first tick ran nothing.
func (in *instance) mustSkip(t *testing.T) {
	t.Helper()
	in.ticks.Tick(t, testJobInterval)
	in.ticks.Tick(t, testJobInterval)
	assert.Empty(t, in.job.runs, "a standby never runs jobs")
}

// stop cancels the runner and waits, bounded, for Run to return; it is
// idempotent.
func (in *instance) stop(t *testing.T) error {
	t.Helper()
	in.stopOnce.Do(func() {
		in.cancel()
		in.stopErr = workertest.Await(t, in.done, "the worker to stop")
	})
	return in.stopErr
}

// leaderOf returns the instance that leads and the one that does not.
func leaderOf(t *testing.T, a, b *instance) (leader, standby *instance) {
	t.Helper()
	a.beat(t)
	b.beat(t)
	require.NotEqual(t, a.runner.Leading(), b.runner.Leading(), "exactly one instance leads")
	if a.runner.Leading() {
		return a, b
	}
	return b, a
}

// leaderBackend returns the pid of the session holding the worker's lock.
func leaderBackend(t *testing.T, db *sql.DB) int {
	t.Helper()
	var pid int
	require.NoError(t, db.QueryRowContext(t.Context(), `
		SELECT pid FROM pg_locks
		WHERE locktype = 'advisory' AND granted AND objsubid = 1
		  AND classid::bigint = ($1::bigint >> 32) AND objid::bigint = ($1::bigint & 4294967295)`,
		pgdb.WorkerLeaderLockKey).Scan(&pid))
	return pid
}

func TestWorker_twoInstancesOnlyOneRunsTheJobs(t *testing.T) {
	database := pgtest.NewDatabase(t)
	leader, standby := leaderOf(t, startInstance(t, database), startInstance(t, database))

	leader.mustRun(t)
	standby.mustSkip(t)
}

// The leader dies without resigning (crash, network loss): PostgreSQL ends its
// session and the lock with it, the standby takes over at its next heartbeat,
// and the old leader stops running jobs at its own next heartbeat.
func TestWorker_whenTheLeaderDiesTheStandbyTakesOver(t *testing.T) {
	database := pgtest.NewDatabase(t)
	killer := database.Connect(t, pgtest.RoleWorker) // same role: may terminate its backends
	leader, standby := leaderOf(t, startInstance(t, database), startInstance(t, database))

	var terminated bool
	require.NoError(t, killer.QueryRowContext(t.Context(),
		"SELECT pg_terminate_backend($1, 5000)", leaderBackend(t, killer)).Scan(&terminated))
	require.True(t, terminated)

	standby.beat(t)
	require.True(t, standby.runner.Leading(), "the standby takes over")
	leader.beat(t)
	require.False(t, leader.runner.Leading(), "the old leader notices its lost session")
	standby.mustRun(t)
	leader.mustSkip(t)
}

// A graceful shutdown (deploy, scale-in) resigns, so the standby takes over at
// its next heartbeat instead of waiting for a dead session to be noticed.
func TestWorker_aStoppedLeaderHandsOver(t *testing.T) {
	database := pgtest.NewDatabase(t)
	leader, standby := leaderOf(t, startInstance(t, database), startInstance(t, database))

	require.NoError(t, leader.stop(t))

	standby.beat(t)
	require.True(t, standby.runner.Leading(), "the standby takes over at its first heartbeat after the resign")
	standby.mustRun(t)
}

// The worker never runs as a privileged role outside dev: the schema owner is
// refused at startup, cc_worker is accepted.
func TestGuardRuntimeRole_overPostgres(t *testing.T) {
	database := pgtest.NewDatabase(t)
	owner := database.Connect(t, pgtest.RoleOwner)
	workerPool := database.Connect(t, pgtest.RoleWorker)
	for _, env := range []config.Environment{config.EnvStaging, config.EnvProduction} {
		require.ErrorIs(t, guardRuntimeRole(t.Context(), env, owner), pgdb.ErrPrivilegedRole, env)
		require.NoError(t, guardRuntimeRole(t.Context(), env, workerPool), env)
	}
	assert.NoError(t, guardRuntimeRole(t.Context(), config.EnvDev, owner), "dev is not checked")
}
