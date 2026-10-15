//go:build integration

package infra_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/infra"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The estimate_views store over the real migrations, as cc_app under RLS.
// Fixture ids are this suite's own; every test gets its own cloned database.
const (
	userA     = "0199d4a2-0000-7000-8000-0000000008aa"
	userB     = "0199d4a2-0000-7000-8000-0000000008bb"
	taxpayerA = "0199d4a2-8888-7000-8000-0000000000a1"
	taxpayerB = "0199d4a2-8888-7000-8000-0000000000b1"
)

// registerSQL is sign-up as cc_app does it under the new user's context;
// every stored value derives from the id (no literal looks like a secret).
const registerSQL = `
WITH v(id, digest) AS (SELECT $1::uuid, sha256(convert_to($1::text, 'UTF8')))
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
SELECT id, digest, '\x0001'::bytea || digest || digest, 1,
       '$argon2id$v=19$m=8,t=1,p=1$' || rtrim(encode(digest, 'base64'), '=') || '$' || rtrim(encode(digest, 'base64'), '='),
       'Fixture'
FROM v`

// insertTaxpayerSQL writes taxpayer $2 of user $1 (persona física).
const insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, 'fisica', 'Fixture',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')), 'GOMA******AB6', true)`

type env struct {
	app   *sql.DB
	txm   *pgdb.TxManager
	store infra.ViewStore
}

func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{app: pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp), store: infra.NewViewStore()}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	for user, taxpayer := range map[string]string{userA: taxpayerA, userB: taxpayerB} {
		require.NoError(t, pgtest.AsUser(t.Context(), e.app, user, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(t.Context(), registerSQL, user); err != nil {
				return err
			}
			_, err := tx.ExecContext(t.Context(), insertTaxpayerSQL, user, taxpayer)
			return err
		}))
	}
	return e
}

func month(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func view(t *testing.T, raw, figures string, at time.Time) domain.View {
	return domain.View{Month: month(t, raw), ViewedAt: at, Seen: domain.Seen{
		RecordsFingerprint: strings.Repeat("a", 64), FiguresFingerprint: strings.Repeat(figures, 64),
		EngineVersion: "1.0.1", ParameterSet: "2026/v1",
	}}
}

func (e *env) save(t *testing.T, user, taxpayer string, v domain.View) error {
	t.Helper()
	return e.txm.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		return e.store.SaveView(ctx, user, taxpayer, v)
	})
}

func (e *env) views(t *testing.T, user, taxpayer, from, to string) []domain.View {
	t.Helper()
	var out []domain.View
	require.NoError(t, e.txm.ReadOnly(t.Context(), user, func(ctx context.Context) error {
		var err error
		out, err = e.store.Views(ctx, taxpayer, month(t, from), month(t, to))
		return err
	}))
	return out
}

var (
	first  = time.Date(2026, time.October, 3, 12, 0, 0, 0, time.UTC)
	second = time.Date(2026, time.October, 14, 18, 30, 0, 0, time.UTC)
)

// §2.3: one upsert per (user, taxpayer, period); the last view wins.
func TestViewStore_upsertIsIdempotentAndTheLastViewWins(t *testing.T) {
	e := newEnv(t)

	require.NoError(t, e.save(t, userA, taxpayerA, view(t, "2026-03", "b", first)))
	require.NoError(t, e.save(t, userA, taxpayerA, view(t, "2026-03", "b", first)), "repeating is fine")
	require.NoError(t, e.save(t, userA, taxpayerA, view(t, "2026-03", "c", second)))

	got := e.views(t, userA, taxpayerA, "2026-01", "2026-12")
	require.Len(t, got, 1)
	assert.Equal(t, month(t, "2026-03"), got[0].Month)
	assert.Equal(t, strings.Repeat("c", 64), got[0].Seen.FiguresFingerprint)
	assert.Equal(t, strings.Repeat("a", 64), got[0].Seen.RecordsFingerprint)
	assert.Equal(t, "1.0.1", got[0].Seen.EngineVersion)
	assert.Equal(t, "2026/v1", got[0].Seen.ParameterSet)
	assert.True(t, second.Equal(got[0].ViewedAt), "got %s", got[0].ViewedAt)
}

// The read is the range asked for, one statement.
func TestViewStore_readsTheRange(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"2025-12", "2026-01", "2026-06", "2026-12", "2027-01"} {
		require.NoError(t, e.save(t, userA, taxpayerA, view(t, p, "b", first)))
	}

	got := e.views(t, userA, taxpayerA, "2026-01", "2026-12")

	var periods []string
	for _, v := range got {
		periods = append(periods, v.Month.String())
	}
	assert.ElementsMatch(t, []string{"2026-01", "2026-06", "2026-12"}, periods)
}

// §2.3 and §8: user B neither reads A's views nor writes a view of A's
// taxpayer (the composite foreign key is the 404).
func TestViewStore_isolation(t *testing.T) {
	e := newEnv(t)
	require.NoError(t, e.save(t, userA, taxpayerA, view(t, "2026-03", "b", first)))

	assert.Empty(t, e.views(t, userB, taxpayerA, "2026-01", "2026-12"), "RLS: A's view is invisible to B")
	err := e.save(t, userB, taxpayerA, view(t, "2026-03", "c", second))
	assert.ErrorIs(t, err, usecase.ErrTaxpayerNotFound)
	err = e.save(t, userA, "0199d4a2-8888-7000-8000-0000000000ff", view(t, "2026-03", "c", second))
	assert.ErrorIs(t, err, usecase.ErrTaxpayerNotFound, "an unknown taxpayer is the same 404")

	got := e.views(t, userA, taxpayerA, "2026-01", "2026-12")
	require.Len(t, got, 1)
	assert.Equal(t, strings.Repeat("b", 64), got[0].Seen.FiguresFingerprint, "A's view is untouched")
}

// An archived taxpayer can be marked (§2.5).
func TestViewStore_archivedTaxpayer(t *testing.T) {
	e := newEnv(t)
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userA, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE taxpayers SET archived_at = now() WHERE id = $1", taxpayerA)
		return err
	}))

	require.NoError(t, e.save(t, userA, taxpayerA, view(t, "2026-03", "b", first)))
}
