// Package migrations holds the versioned goose SQL migrations, embedded in the
// binary, and the provider that applies them (ADR-0001 O2).
//
// Files are named <YYYYMMDDHHMMSS>_<snake_case>.sql (`goose create -s` is not used:
// timestamped versions avoid collisions between parallel branches). Each file has a
// `-- +goose Up` and a non-empty `-- +goose Down` section and runs in one transaction.
//
// Migrations run as cc_owner only: Up and cmd/migrate refuse any other role, so every
// object is owned by cc_owner and the runtime roles never hold DDL rights.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// OwnerRole is the only role allowed to run migrations (docs/architecture/data-model.md).
const OwnerRole = "cc_owner"

// ErrNotOwner is returned when the connection is not logged in as OwnerRole.
var ErrNotOwner = errors.New("migrations: must run as " + OwnerRole)

// FS holds every migration file.
//
//go:embed *.sql
var FS embed.FS

// NewProvider returns a goose provider over FS for db. A PostgreSQL session advisory
// lock serializes concurrent runners (e.g. two deploys racing).
func NewProvider(db *sql.DB) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("migrations: session locker: %w", err)
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, FS,
		goose.WithSessionLocker(locker),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return nil, fmt.Errorf("migrations: provider: %w", err)
	}
	return p, nil
}

// RequireOwner returns ErrNotOwner unless db is logged in as OwnerRole.
func RequireOwner(ctx context.Context, db *sql.DB) error {
	var user string
	if err := db.QueryRowContext(ctx, "SELECT current_user").Scan(&user); err != nil {
		return fmt.Errorf("migrations: read current user: %w", err)
	}
	if user != OwnerRole {
		return fmt.Errorf("%w (logged in as %q)", ErrNotOwner, user)
	}
	return nil
}

// Up applies every pending migration as OwnerRole. Its signature matches
// pgtest.Migrator, so the integration harness migrates exactly like cmd/migrate.
func Up(ctx context.Context, db *sql.DB) error {
	if err := RequireOwner(ctx, db); err != nil {
		return err
	}
	p, err := NewProvider(db)
	if err != nil {
		return err
	}
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrations: up: %w", err)
	}
	return nil
}
