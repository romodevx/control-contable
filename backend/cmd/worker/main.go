// Command worker is the composition root of the background worker (ADR-0001
// O8, docs/architecture/c4.md): it loads the configuration, checks that it
// runs as a plain runtime role (cc_worker), registers the jobs and runs them
// on the instance that holds the leader lock (pg_try_advisory_lock on a
// dedicated connection) until SIGINT/SIGTERM. Shutdown cancels the running
// jobs, waits for them and then resigns, so a standby takes over at once.
//
// Jobs: the records purge (M9-08, every hour, whatever the module flags say)
// and the reminders (M9-06, every 5 minutes, a no-op while module_reminders
// is OFF), which open the push tokens with the API's PUSH_TOKEN_* keys and
// send through the push sender (M9-07) built at startup from FCM_PROJECT_ID /
// FCM_CREDENTIALS_JSON, disabled when no project is set. Every job runs its
// database work through tx.Manager's
// System unit of work (no user context; SECURITY DEFINER functions such as
// due_reminder_users) or a per-user unit of work.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/worker"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	retentioninfra "github.com/romodevx/control-contable/backend/internal/retention/infra"
	retention "github.com/romodevx/control-contable/backend/internal/retention/usecase"
)

// roleCheckTimeout bounds the startup check of the runtime database role.
const roleCheckTimeout = 10 * time.Second

func main() {
	if err := runUntilSignal(); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func runUntilSignal() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return run(ctx, os.LookupEnv, os.Stdout)
}

// run loads the configuration and runs the jobs until ctx is done.
func run(ctx context.Context, lookup config.Lookup, logOut io.Writer) error {
	cfg, err := config.LoadWorker(ctx, lookup, config.NewEnvSecrets(lookup))
	if err != nil {
		return err
	}
	keys, err := config.LoadReminders(ctx, lookup, config.NewEnvSecrets(lookup), cfg.Env, config.Auth{}, config.Taxpayers{})
	if err != nil {
		return err
	}
	logger := logging.New(logOut, cfg.LogLevel)
	push, err := newPushSender(ctx, cfg.Push, logger)
	if err != nil {
		return err
	}
	pool, err := pgdb.Open(cfg.DatabaseURL.Reveal())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closeDatabase(pool, logger)
	if err := guardRuntimeRole(ctx, cfg.Env, pool); err != nil {
		return err
	}
	reg, err := newRegistry(pool, logger, push, keys)
	if err != nil {
		return err
	}
	runner, err := newRunner(pool, reg, logger)
	if err != nil {
		return err
	}
	return runner.Run(ctx)
}

// recordsPurgeTimeout bounds one run of the records purge (its passes are
// short transactions; a run cut short resumes on the next tick).
const recordsPurgeTimeout = 10 * time.Minute

// scheduledJob is a production job with the timeout each of its runs gets.
type scheduledJob struct {
	job     worker.Job
	timeout time.Duration
}

// newRegistry registers the production jobs (newJobs) over pool (cc_worker),
// on the system clock.
func newRegistry(pool *sql.DB, logger *slog.Logger, push reminders.PushSender, keys config.Reminders) (*worker.Registry, error) {
	jobs, err := newJobs(pool, logger, push, keys, clock.NewSystem())
	if err != nil {
		return nil, err
	}
	reg := &worker.Registry{}
	for _, j := range jobs {
		if err := reg.Register(j.job, j.timeout); err != nil {
			return nil, err
		}
	}
	return reg, nil
}

// newJobs builds the production jobs over pool (cc_worker), in registration
// order. Module jobs check their flag on every run and skip while the module
// is effectively OFF (feature_flags is readable by cc_worker); the records
// purge never does (MODULE-reminders §3). push is the reminders' sender
// (M9-07) and keys the push token keys the API seals with, for the reminder
// job (M9-06); now is its clock. In dev without keys the reminder job is not
// built (logged); outside dev config requires them.
func newJobs(pool *sql.DB, logger *slog.Logger, push reminders.PushSender, keys config.Reminders, now clock.Clock) ([]scheduledJob, error) {
	units, err := pgdb.NewTxManager(pool)
	if err != nil {
		return nil, err
	}
	purge, err := retention.NewPurgeJob(retention.PurgeDeps{
		System: units, Purger: retentioninfra.NewPurger(), Logger: logger,
		BatchSize: retention.DefaultBatchSize, MaxBatches: retention.DefaultMaxBatches,
	})
	if err != nil {
		return nil, err
	}
	jobs := []scheduledJob{{job: purge, timeout: recordsPurgeTimeout}}
	if !hasReminderKeys(keys) {
		logger.Warn("reminders disabled", slog.String("reason", "PUSH_TOKEN_ENC_KEY / PUSH_TOKEN_INDEX_PEPPER are not set"))
		return jobs, nil
	}
	job, err := newReminderJob(pool, units, logger, push, keys, now)
	if err != nil {
		return nil, err
	}
	return append(jobs, scheduledJob{job: job, timeout: remindersTimeout}), nil
}

// newRunner builds the Runner over pool (cc_worker); the leader lock is held
// on one connection of that pool while this instance leads.
func newRunner(pool *sql.DB, reg *worker.Registry, logger *slog.Logger, opts ...worker.Option) (*worker.Runner, error) {
	elector, err := pgdb.NewAdvisoryLeader(pool, pgdb.WorkerLeaderLockKey)
	if err != nil {
		return nil, err
	}
	return worker.NewRunner(reg, elector, append([]worker.Option{worker.WithLogger(logger)}, opts...)...)
}

// guardRuntimeRole refuses to run, outside dev, unless the pool is logged in
// as a plain runtime role that RLS binds (never cc_owner, a superuser or a
// BYPASSRLS role); the same check as the API's. Outside dev a database that is
// down at boot keeps the worker from starting (fail closed). Dev is not
// checked, so the local stack can start before its database.
func guardRuntimeRole(ctx context.Context, env config.Environment, pool *sql.DB) error {
	if env == config.EnvDev {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, roleCheckTimeout)
	defer cancel()
	if err := pgdb.CheckRuntimeRole(ctx, pool); err != nil {
		return fmt.Errorf("refusing to start: runtime database role: %w", err)
	}
	return nil
}

func closeDatabase(pool *sql.DB, logger *slog.Logger) {
	if err := pool.Close(); err != nil {
		logger.Warn("close database pool", slog.Any("error", err))
	}
}
