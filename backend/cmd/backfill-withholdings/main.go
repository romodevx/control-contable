// Command backfill-withholdings derives the stored ISR and IVA withholdings
// of incomes and expenses written before automatic derivation (WH-06,
// WH-07), and re-derives the ones an older decision table produced
// (MODULE-withholdings §5.2, §5.3; ADR-0008). It is a maintained tool: every
// table_version bump runs it again.
//
// Dry run by default: it prints, per table, how many records would change
// and their ids (never an amount, a description or a user id) and writes
// nothing. -apply writes. It logs in as cc_worker (DATABASE_URL, the
// worker's secret), never as the owner, and refuses a role that bypasses
// RLS. Each user is processed in a transaction of their own with
// app.current_user_id set, so RLS scopes every read and write; the first
// error stops the run. Re-running is safe: a derived record is not visited
// again. Runbook: docs/runbooks/withholdings-backfill.md.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
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
)

// roleCheckTimeout bounds the check of the database role.
const roleCheckTimeout = 10 * time.Second

func main() {
	if err := runUntilSignal(); err != nil {
		fmt.Fprintln(os.Stderr, "backfill-withholdings:", err)
		os.Exit(1)
	}
}

// runUntilSignal runs the command; SIGINT/SIGTERM cancels it (the current
// user's transaction rolls back).
func runUntilSignal() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return run(ctx, os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr)
}

// run parses args, connects as the configured role and runs the backfill,
// printing the report on out and the logs on logOut.
func run(ctx context.Context, args []string, lookup config.Lookup, out, logOut io.Writer) error {
	flags := flag.NewFlagSet("backfill-withholdings", flag.ContinueOnError)
	flags.SetOutput(logOut)
	apply := flags.Bool("apply", false, "write the derivations (default: dry run, nothing is written)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	cfg, err := config.LoadWorker(ctx, lookup, config.NewEnvSecrets(lookup))
	if err != nil {
		return err
	}
	logger := logging.New(logOut, cfg.LogLevel)
	pool, err := pgdb.Open(cfg.DatabaseURL.Reveal())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closePool(pool, logger)
	return backfill(ctx, pool, clock.NewSystem(), *apply, out, logger)
}

// backfill checks the role, runs the backfill over pool and prints the
// report. The role is checked in every environment: the run writes data.
func backfill(ctx context.Context, pool *sql.DB, now clock.Clock, apply bool, out io.Writer, logger *slog.Logger) error {
	checkCtx, cancel := context.WithTimeout(ctx, roleCheckTimeout)
	defer cancel()
	if err := pgdb.CheckRuntimeRole(checkCtx, pool); err != nil {
		return fmt.Errorf("refusing to run: database role: %w", err)
	}
	b, err := newBackfiller(pool, now)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "withholdings backfill started", slog.Bool("apply", apply))
	rep, err := b.run(ctx, apply)
	if err != nil {
		return errors.Join(fmt.Errorf("backfill stopped: %w", err), errStoppedHint)
	}
	logger.InfoContext(ctx, "withholdings backfill finished", slog.Bool("apply", apply), slog.Int("users", rep.Users))
	return rep.write(out)
}

// errStoppedHint tells the operator what a stopped run left behind.
var errStoppedHint = errors.New("the users before the failing one are done; fix the cause and run again")

func closePool(pool *sql.DB, logger *slog.Logger) {
	if err := pool.Close(); err != nil {
		logger.Warn("close database pool", slog.Any("error", err))
	}
}
