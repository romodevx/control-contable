// Package pgdb is the PostgreSQL adapter of the platform (ADR-0001): the
// connection pool (pgx v5 through database/sql), the TxManager that implements
// the tx.Manager port with GORM on that same pool, and the readiness Pinger
// behind /healthz.
//
// Repositories in a module's infra/ obtain the transaction of the running unit
// of work with Conn(ctx). Outside a unit of work Conn fails with
// ErrNoTransaction, so no query can ever run without the RLS context.
//
// Use cases never import this package (archtest rule usecase-platform-adapter);
// they depend on internal/platform/tx and cmd/ wires the TxManager in.
package pgdb

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Pool defaults for one API or worker instance. Cloud SQL's smallest tiers
// allow about 25 connections; two instances of each binary stay inside that.
const (
	DefaultMaxOpenConns    = 10
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 30 * time.Minute
	DefaultConnMaxIdleTime = 5 * time.Minute
)

// ErrInvalidDSN reports a connection string that cannot be parsed. The cause is
// left out on purpose: parse errors quote the string, password included.
var ErrInvalidDSN = errors.New("pgdb: database url is not a valid PostgreSQL connection string")

// PoolOption customizes Open.
type PoolOption func(*sql.DB)

// WithMaxOpenConns overrides DefaultMaxOpenConns.
func WithMaxOpenConns(n int) PoolOption {
	return func(db *sql.DB) { db.SetMaxOpenConns(n) }
}

// Open parses dsn and returns a pgx-backed pool. It does not connect: a
// database that is down at boot makes /healthz report "degraded" instead of
// keeping the process from starting.
func Open(dsn string, opts ...PoolOption) (*sql.DB, error) {
	if dsn == "" {
		return nil, ErrInvalidDSN
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, ErrInvalidDSN
	}
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(DefaultMaxOpenConns)
	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)
	db.SetConnMaxIdleTime(DefaultConnMaxIdleTime)
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}
