//go:build integration

package infra_test

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The taxpayers use cases over the real stores and migration, as cc_app under
// RLS, wired as cmd/api will wire them (M2-07). Fixture ids are this suite's
// own; every test gets its own cloned database.
const (
	userA = "0199d4a2-0000-7000-8000-00000000000a"
	userB = "0199d4a2-0000-7000-8000-00000000000b"

	// Synthetic RFCs (see domain/rfc_test.go): valid check digits, except
	// rfcFisicaMismatch.
	rfcFisica         = "PEXJ900215HK8"
	rfcFisicaOther    = "AAAA010101AA0"
	rfcFisicaMismatch = "PEXJ900215HK9"
	rfcMoral          = "A&B020304XY2"
)

// Obviously fake 32-byte keys.
var (
	testRFCKey    = bytes.Repeat([]byte{0x5a}, crypto.KeySize)
	testRFCPepper = bytes.Repeat([]byte{0x6b}, crypto.MinPepperSize)
)

// today is the use cases' frozen clock: 13 October 2026 in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

// registerSQL is sign-up as cc_app does it under the new user's context: the
// user, its counters and a 7-day trial starting at the database's now().
var registerSQL = []string{
	`WITH v(id) AS (SELECT $1::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES ($1)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '168 hours')",
}

type env struct {
	app       *sql.DB
	txm       *pgdb.TxManager
	vault     *infra.RFCVault
	store     infra.Store
	logs      *bytes.Buffer
	taxpayers *usecase.Taxpayers
	regimes   *usecase.Regimes
	queries   *usecase.Queries
}

type envOption func(*usecase.Deps)

// readOnlyAccess decides access a month after the wall clock: the trial rows
// start at the database's now(), so every trial is over by then.
func readOnlyAccess(txm *pgdb.TxManager) envOption {
	return func(d *usecase.Deps) {
		later := clock.NewFixed(time.Now().Add(30 * 24 * time.Hour))
		d.Access = access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), later)
	}
}

func newEnv(t *testing.T, users ...string) *env {
	t.Helper()
	return newEnvOn(t, pgtest.NewDatabase(t), users...)
}

// newEnvOn is newEnv over an existing test database.
func newEnvOn(t *testing.T, db *pgtest.Database, users ...string) *env {
	t.Helper()
	e := &env{app: db.Connect(t, pgtest.RoleApp), logs: &bytes.Buffer{}, store: infra.NewStore()}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	keyring, err := crypto.NewKeyring(1, map[crypto.KeyID][]byte{1: testRFCKey})
	require.NoError(t, err)
	index, err := crypto.NewRFCIndex(testRFCPepper)
	require.NoError(t, err)
	e.vault = infra.NewRFCVault(keyring, index)
	for _, u := range users {
		register(t, e.app, u)
	}
	return e
}

// wire builds the use cases; options override deps (e.g. readOnlyAccess).
func (e *env) wire(t *testing.T, options ...envOption) *env {
	t.Helper()
	deps := e.deps()
	for _, o := range options {
		o(&deps)
	}
	e.taxpayers, e.regimes, e.queries = usecase.NewTaxpayers(deps), usecase.NewRegimes(deps), usecase.NewQueries(deps)
	return e
}

// deps are the real stores, catalog and policy over the env's pool.
func (e *env) deps() usecase.Deps {
	return usecase.Deps{
		Tx: e.txm, Access: access.NewPolicy(e.txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Taxpayers: e.store, Regimes: e.store, Catalog: infra.NewCatalog(), Directory: infra.NewCatalog(), Records: infra.NoRecords{},
		RFC: e.vault, Clock: clock.NewFixed(today),
		Logger: slog.New(slog.NewJSONHandler(e.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
}

func register(t *testing.T, app *sql.DB, userID string) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		for _, statement := range registerSQL {
			if _, err := tx.ExecContext(t.Context(), statement, userID); err != nil {
				return err
			}
		}
		return nil
	}))
}

// queryAs runs one query under userID's RLS context and scans one row.
func queryAs(t *testing.T, app *sql.DB, userID, query string, args []any, dest ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), query, args...).Scan(dest...)
	}))
}

// execAs runs one statement under userID's RLS context.
func execAs(ctx context.Context, app *sql.DB, userID, statement string, args ...any) error {
	return pgtest.AsUser(ctx, app, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, statement, args...)
		return err
	})
}
