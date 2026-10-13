// Command worker is the composition root of the background worker (ADR-0001
// O8, docs/architecture/c4.md): it loads the configuration, checks that it
// runs as a plain runtime role (cc_worker), registers the jobs and runs them
// on the instance that holds the leader lock (pg_try_advisory_lock on a
// dedicated connection) until SIGINT/SIGTERM. Shutdown cancels the running
// jobs, waits for them and then resigns, so a standby takes over at once.
//
// No production job is registered yet: M9 adds the reminders and the
// soft-delete purge. Every job runs its database work through tx.Manager's
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

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/worker"
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
	logger := logging.New(logOut, cfg.LogLevel)
	pool, err := pgdb.Open(cfg.DatabaseURL.Reveal())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closeDatabase(pool, logger)
	if err := guardRuntimeRole(ctx, cfg.Env, pool); err != nil {
		return err
	}
	reg, err := newRegistry()
	if err != nil {
		return err
	}
	runner, err := newRunner(pool, reg, logger)
	if err != nil {
		return err
	}
	return runner.Run(ctx)
}

// newRegistry registers the production jobs. Each one checks its module flag
// on every run and skips while the module is OFF (feature_flags is readable by
// cc_worker).
func newRegistry() (*worker.Registry, error) {
	return &worker.Registry{}, nil
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
