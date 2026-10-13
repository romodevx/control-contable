// Command migrate applies the embedded goose migrations (package migrations) as
// cc_owner. It refuses to run as any other role.
//
// Usage:
//
//	MIGRATE_DATABASE_URL=postgres://cc_owner:...@host:5432/control_contable migrate up|down|status
//
//	up      apply every pending migration
//	down    roll back the most recent migration
//	status  list every migration as applied or pending
//
// The URL is a secret: it is read from the environment (or the secret manager that
// populates it), never from a flag, so it does not show up in process listings.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/migrations"
)

// envDatabaseURL names the owner connection string. It is distinct from the API's
// DATABASE_URL (cc_app) so a runtime process can never be pointed at DDL rights.
const envDatabaseURL = "MIGRATE_DATABASE_URL"

const driverName = "pgx"

type command string

const (
	cmdUp     command = "up"
	cmdDown   command = "down"
	cmdStatus command = "status"
)

var errUsage = errors.New("usage: migrate up|down|status (owner URL in " + envDatabaseURL + ")")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.LookupEnv, os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, lookup config.Lookup, out io.Writer) error {
	cmd, err := parseCommand(args)
	if err != nil {
		return err
	}
	dsn, err := config.NewEnvSecrets(lookup).Secret(ctx, envDatabaseURL)
	if err != nil {
		return fmt.Errorf("%s: %w", envDatabaseURL, err)
	}
	db, err := sql.Open(driverName, dsn.Reveal())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	return errors.Join(execute(ctx, cmd, db, out), db.Close())
}

func parseCommand(args []string) (command, error) {
	if len(args) != 1 {
		return "", errUsage
	}
	switch cmd := command(args[0]); cmd {
	case cmdUp, cmdDown, cmdStatus:
		return cmd, nil
	default:
		return "", fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}
}

// execute runs cmd against db after checking that db is logged in as the owner.
func execute(ctx context.Context, cmd command, db *sql.DB, out io.Writer) error {
	if err := migrations.RequireOwner(ctx, db); err != nil {
		return err
	}
	p, err := migrations.NewProvider(db)
	if err != nil {
		return err
	}
	switch cmd {
	case cmdUp:
		results, err := p.Up(ctx)
		printResults(out, results)
		return err
	case cmdDown:
		result, err := p.Down(ctx)
		if result != nil {
			printResults(out, []*goose.MigrationResult{result})
		}
		return err
	default:
		return printStatus(ctx, p, out)
	}
}

func printResults(out io.Writer, results []*goose.MigrationResult) {
	if len(results) == 0 {
		fmt.Fprintln(out, "no migrations to run")
		return
	}
	for _, r := range results {
		fmt.Fprintln(out, r.String())
	}
}

func printStatus(ctx context.Context, p *goose.Provider, out io.Writer) error {
	statuses, err := p.Status(ctx)
	if err != nil {
		return err
	}
	for _, s := range statuses {
		applied := "-"
		if s.State == goose.StateApplied {
			applied = s.AppliedAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(out, "%-8s %-20s %s\n", s.State, applied, s.Source.Path)
	}
	return nil
}
