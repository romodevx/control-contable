//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Sessions and consents migration (M1-02): refresh_tokens and consents.

const (
	sessionsConsentsMigration = "20261013142137_sessions_consents.sql"

	sqlstateForeignKeyViolation = "23503"
)

var (
	refreshTokensTable = pgtest.Table{Name: "refresh_tokens", OwnerColumn: "user_id"}
	consentsTable      = pgtest.Table{Name: "consents", OwnerColumn: "user_id"}
)

// insertRefreshSQL issues a root token for $1: the secret's SHA-256 is derived from the
// user id plus a per-call nonce, so several tokens per user never collide.
const insertRefreshSQL = `
INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, absolute_expires_at)
VALUES ($1::text::uuid, uuidv7(), sha256(convert_to($1::text || clock_timestamp()::text || random()::text, 'UTF8')),
        now() + interval '30 days', now() + interval '90 days')`

// insertSessionSQL issues a token with a caller-chosen secret ($2) and family ($3).
const insertSessionSQL = `
INSERT INTO refresh_tokens (id, user_id, family_id, token_hash, rotated_from, expires_at, absolute_expires_at)
VALUES ($4, $1, $3, sha256(convert_to($2::text, 'UTF8')), $5, now() + interval '30 days', now() + interval '90 days')`

const insertConsentSQL = `
INSERT INTO consents (user_id, document_kind, document_version, accepted_at)
VALUES ($1, 'privacy_notice_integral', 'v1', now())`

const (
	familyA  = "00000000-0000-4000-8000-0000000000fa"
	tokenA1  = "00000000-0000-4000-8000-0000000000a1"
	tokenA2  = "00000000-0000-4000-8000-0000000000a2"
	secretA1 = "secret-a-1"
	secretA2 = "secret-a-2"
)

// lookupByHash is the refresh lookup: by token hash, under the user named in the token.
const lookupByHash = "SELECT count(*) FROM refresh_tokens WHERE token_hash = sha256(convert_to($1::text, 'UTF8'))"

func TestSessionsConsents_AppRoleIsolation(t *testing.T) {
	specs := []pgtest.IsolationSpec{
		{Table: refreshTokensTable, Insert: execAs(insertRefreshSQL)},
		{Table: consentsTable, Insert: execAs(insertConsentSQL)},
	}
	for _, spec := range specs {
		t.Run(spec.Table.Name, func(t *testing.T) {
			db := pgtest.NewDatabase(t)
			app := db.Connect(t, pgtest.RoleApp)
			seedAs(t, app, userA, insertUserSQL)
			seedAs(t, app, userB, insertUserSQL)
			spec.UserA, spec.UserB = userA, userB

			pgtest.AssertIsolation(t, app, spec)
		})
	}
}

// AssertIsolation's UPDATE probe rewrites user_id, which cc_app may not update; prove
// that cross-user writes on the columns it may write match no row, while its own do.
func TestRefreshTokens_CrossUserWritesMatchNoRow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedAs(t, app, userA, insertRefreshSQL)
	seedAs(t, app, userB, insertRefreshSQL)
	writes := []string{
		"UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'logout' WHERE user_id = $1",
		"UPDATE refresh_tokens SET rotated_at = now() WHERE user_id = $1",
		"DELETE FROM refresh_tokens WHERE user_id = $1",
	}
	for _, w := range writes {
		t.Run(w, func(t *testing.T) {
			affected, err := execAsUser(t, app, userA, w, userB)
			require.NoError(t, err)
			assert.Zero(t, affected, "A cannot write B's tokens")
		})
	}
	t.Run("own rows are writable", func(t *testing.T) {
		for _, w := range writes {
			affected, err := execAsUser(t, app, userB, w, userB)
			require.NoError(t, err, w)
			assert.EqualValues(t, 1, affected, w)
		}
	})
}

// ADR-0001 O6: the token is <user_id>.<secret>. The API sets the context to the user
// named in the token, then looks the hash up: no RLS bypass and no definer function.
func TestRefreshTokens_LookupRunsUnderTheTokensUser(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	_, err := execAsUser(t, app, userA, insertSessionSQL, userA, secretA1, familyA, tokenA1, nil)
	require.NoError(t, err)

	count := func(t *testing.T, userID string) int {
		t.Helper()
		var n int
		read := func(tx *sql.Tx) error { return tx.QueryRowContext(t.Context(), lookupByHash, secretA1).Scan(&n) }
		if userID == "" {
			require.NoError(t, pgtest.WithoutUser(t.Context(), app, read))
		} else {
			require.NoError(t, pgtest.AsUser(t.Context(), app, userID, read))
		}
		return n
	}

	assert.Equal(t, 1, count(t, userA), "the right user prefix finds the token")
	assert.Zero(t, count(t, userB), "a forged user prefix finds nothing")
	assert.Zero(t, count(t, ""), "no context finds nothing")
}

// M1-08 rotation: one atomic UPDATE claims the presented token. A second claim (reuse,
// or the loser of a concurrent refresh) matches no row and the family is revoked.
func TestRefreshTokens_RotationAndReuseDetection(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	const claim = `UPDATE refresh_tokens SET rotated_at = now()
		WHERE token_hash = sha256(convert_to($1::text, 'UTF8')) AND rotated_at IS NULL AND revoked_at IS NULL
		  AND expires_at > now() AND absolute_expires_at > now()`
	const revokeFamily = `UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'reuse_detected'
		WHERE family_id = $1 AND revoked_at IS NULL`

	_, err := execAsUser(t, app, userA, insertSessionSQL, userA, secretA1, familyA, tokenA1, nil)
	require.NoError(t, err)
	affected, err := execAsUser(t, app, userA, claim, secretA1)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected, "first use rotates")
	_, err = execAsUser(t, app, userA, insertSessionSQL, userA, secretA2, familyA, tokenA2, tokenA1)
	require.NoError(t, err, "the successor links to its parent in the same family")

	affected, err = execAsUser(t, app, userA, claim, secretA1)
	require.NoError(t, err)
	require.Zero(t, affected, "reuse claims nothing")
	affected, err = execAsUser(t, app, userA, revokeFamily, familyA)
	require.NoError(t, err)
	assert.EqualValues(t, 2, affected, "the whole family is revoked")

	affected, err = execAsUser(t, app, userA, claim, secretA2)
	require.NoError(t, err)
	assert.Zero(t, affected, "a revoked successor cannot rotate")
}

// Once set, rotated_at and the revocation are history: neither can be cleared or moved,
// so a buggy rotation cannot resurrect a token that reuse detection depends on.
func TestRefreshTokens_RotationAndRevocationAreFinal(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	_, err := execAsUser(t, app, userA, insertSessionSQL, userA, secretA1, familyA, tokenA1, nil)
	require.NoError(t, err)
	_, err = execAsUser(t, app, userA,
		"UPDATE refresh_tokens SET rotated_at = now(), revoked_at = now(), revoke_reason = 'logout'")
	require.NoError(t, err)

	for _, w := range []string{
		"UPDATE refresh_tokens SET rotated_at = NULL",
		"UPDATE refresh_tokens SET rotated_at = now() + interval '1 hour'",
		"UPDATE refresh_tokens SET revoked_at = NULL, revoke_reason = NULL",
		"UPDATE refresh_tokens SET revoke_reason = 'reuse_detected'",
	} {
		t.Run(w, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, w)
			requireSQLState(t, err, sqlstateRaiseException)
		})
	}
}

// Consents are append-only legal evidence (LFPDPPP): no role but the owner's cascade
// changes or removes them.
func TestConsents_AppendOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	seedAs(t, app, userA, insertConsentSQL)

	for _, w := range []string{
		"UPDATE consents SET document_version = '2027-01-01'",
		"UPDATE consents SET accepted_at = now()",
		"DELETE FROM consents",
		"TRUNCATE consents",
	} {
		t.Run(w, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, w)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// Least privilege: the webhook and the worker never touch sessions or consents, and
// cc_app cannot write the columns the database owns.
func TestSessionsConsents_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	seedUsers(t, db.Connect(t, pgtest.RoleApp), userA)
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE refresh_tokens SET user_id = user_id"},
		{pgtest.RoleApp, "UPDATE refresh_tokens SET family_id = family_id"},
		{pgtest.RoleApp, "UPDATE refresh_tokens SET token_hash = token_hash"},
		{pgtest.RoleApp, "UPDATE refresh_tokens SET expires_at = expires_at"},
		{pgtest.RoleApp, "UPDATE refresh_tokens SET absolute_expires_at = absolute_expires_at"},
		{pgtest.RoleApp, "UPDATE refresh_tokens SET created_at = now()"},
		{pgtest.RoleApp, "TRUNCATE refresh_tokens"},
		{pgtest.RoleApp, "UPDATE consents SET user_id = user_id"},
		{pgtest.RoleWebhook, "SELECT 1 FROM refresh_tokens"},
		{pgtest.RoleWebhook, "SELECT 1 FROM consents"},
		{pgtest.RoleWorker, "SELECT 1 FROM refresh_tokens"},
		{pgtest.RoleWorker, "DELETE FROM refresh_tokens"},
		{pgtest.RoleWorker, "SELECT 1 FROM consents"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// Every case runs as cc_app under the row's own user, so RLS admits the row and the
// constraint under test is what refuses it.
func TestSessionsConsents_RejectInvalidRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	_, err := execAsUser(t, app, userA, insertSessionSQL, userA, secretA1, familyA, tokenA1, nil)
	require.NoError(t, err)
	seedAs(t, app, userB, insertRefreshSQL)
	seedAs(t, app, userA, insertConsentSQL)
	const (
		session = "INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, absolute_expires_at, rotated_from) VALUES ($1, "
		consent = "INSERT INTO consents (user_id, document_kind, document_version, accepted_at) VALUES ($1, "
	)
	bad := []struct {
		name  string
		query string
		state string
	}{
		{"token_hash not 32 bytes", session + "uuidv7(), '\\x01', now() + interval '1 day', now() + interval '2 days', NULL)", sqlstateCheckViolation},
		{"duplicate token_hash", session + "uuidv7(), sha256(convert_to('" + secretA1 + "', 'UTF8')), now() + interval '1 day', now() + interval '2 days', NULL)", sqlstateUniqueViolation},
		{"already expired", session + "uuidv7(), sha256('n1'), now() - interval '1 second', now() + interval '2 days', NULL)", sqlstateCheckViolation},
		{"sliding expiry beyond 30 days", session + "uuidv7(), sha256('n2'), now() + interval '31 days', now() + interval '90 days', NULL)", sqlstateCheckViolation},
		{"absolute expiry beyond 90 days", session + "uuidv7(), sha256('n3'), now() + interval '1 day', now() + interval '91 days', NULL)", sqlstateCheckViolation},
		{"sliding expiry past the absolute one", session + "uuidv7(), sha256('n4'), now() + interval '2 days', now() + interval '1 day', NULL)", sqlstateCheckViolation},
		{"parent in another family", session + "uuidv7(), sha256('n5'), now() + interval '1 day', now() + interval '2 days', '" + tokenA1 + "')", sqlstateForeignKeyViolation},
		{"unknown document kind", consent + "'cookies', 'v1', now())", sqlstateForeignKeyViolation},
		{"blank document version", consent + "'terms', '  ', now())", sqlstateCheckViolation},
		{"accepted in the future", consent + "'terms', 'v1', now() + interval '1 day')", sqlstateCheckViolation},
		{"same version accepted twice", insertConsentSQL, sqlstateUniqueViolation},
	}
	for _, b := range bad {
		t.Run(b.name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, b.query, userA)
			requireSQLState(t, err, b.state)
		})
	}

	t.Run("parent owned by another user", func(t *testing.T) {
		var parentB string
		require.NoError(t, pgtest.AsUser(t.Context(), app, userB, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), "SELECT id::text FROM refresh_tokens").Scan(&parentB)
		}))
		_, err := execAsUser(t, app, userA,
			session+"'"+familyA+"', sha256('n7'), now() + interval '1 day', now() + interval '2 days', $2)", userA, parentB)
		requireSQLState(t, err, sqlstateForeignKeyViolation)
	})
}

// revoked_at and revoke_reason are set together, with a known reason, or not at all.
// cc_app cannot insert either column (tokens are born live), so the pairing is proved
// on UPDATE.
func TestRefreshTokens_RevocationReasonPairing(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	seedAs(t, app, userA, insertRefreshSQL)
	for _, w := range []string{
		"UPDATE refresh_tokens SET revoked_at = now()",
		"UPDATE refresh_tokens SET revoke_reason = 'logout'",
		"UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'stolen'",
	} {
		t.Run(w, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, w)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
}

// Account deletion (US-A4) removes the user's sessions and consents by cascade; other
// users keep theirs. Deleting a parent token keeps its successor (rotated_from → NULL).
func TestSessionsConsents_DeletionCascades(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	for _, id := range []string{userA, userB} {
		seedAs(t, app, id, insertRefreshSQL)
		seedAs(t, app, id, insertConsentSQL)
	}

	affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)
	for _, table := range []pgtest.Table{refreshTokensTable, consentsTable} {
		assert.Zero(t, countOwned(t, app, userA, table), table.Name)
		assert.Equal(t, 1, countOwned(t, app, userB, table), table.Name+": other users untouched")
	}

	t.Run("a deleted parent leaves its successor", func(t *testing.T) {
		seedUsers(t, app, userC)
		_, err := execAsUser(t, app, userC, insertSessionSQL, userC, secretA1, familyA, tokenA1, nil)
		require.NoError(t, err)
		_, err = execAsUser(t, app, userC, insertSessionSQL, userC, secretA2, familyA, tokenA2, tokenA1)
		require.NoError(t, err)

		affected, err := execAsUser(t, app, userC, "DELETE FROM refresh_tokens WHERE id = $1", tokenA1)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		var orphaned bool
		require.NoError(t, pgtest.AsUser(t.Context(), app, userC, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(),
				"SELECT rotated_from IS NULL AND user_id = $2 FROM refresh_tokens WHERE id = $1", tokenA2, userC).Scan(&orphaned)
		}))
		assert.True(t, orphaned, "successor kept, link cleared, owner intact")
	})
}

// The refresh lookup and the family revoke must use an index, not scan every token.
func TestRefreshTokens_LookupPathsAreIndexed(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	indexed := map[string]string{
		"unique hash":       "SELECT 1 FROM pg_index i JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = i.indkey[0] WHERE i.indrelid = 'refresh_tokens'::regclass AND i.indisunique AND i.indnatts = 1 AND a.attname = 'token_hash'",
		"(user_id, family)": "SELECT 1 FROM pg_index i WHERE i.indrelid = 'refresh_tokens'::regclass AND pg_get_indexdef(i.indexrelid) LIKE '%(user_id, family_id%'",
		"consents user":     "SELECT 1 FROM pg_index i WHERE i.indrelid = 'consents'::regclass AND pg_get_indexdef(i.indexrelid) LIKE '%(user_id, %'",
	}
	for name, query := range indexed {
		t.Run(name, func(t *testing.T) {
			var one int
			require.NoError(t, owner.QueryRowContext(t.Context(), query).Scan(&one))
		})
	}
}

// Down removes exactly what this migration added and up re-applies it.
func TestSessionsConsents_DownAndUp(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)
	objects := []string{
		"to_regclass('public.refresh_tokens')",
		"to_regclass('public.consents')",
		"to_regprocedure('public.refresh_tokens_history_is_final()')",
	}
	exists := func(ctx context.Context) []bool {
		out := make([]bool, len(objects))
		for i, obj := range objects {
			var found sql.NullString
			require.NoError(t, owner.QueryRowContext(ctx, "SELECT "+obj+"::text").Scan(&found))
			out[i] = found.Valid
		}
		return out
	}
	version := migrationVersion(t, p, sessionsConsentsMigration)

	_, err = p.DownTo(t.Context(), version-1)
	require.NoError(t, err)
	assert.Equal(t, []bool{false, false, false}, exists(t.Context()))
	var usersLeft sql.NullString
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT to_regclass('public.users')::text").Scan(&usersLeft))
	assert.True(t, usersLeft.Valid, "earlier migrations untouched")

	require.NoError(t, migrations.Up(t.Context(), owner))
	assert.Equal(t, []bool{true, true, true}, exists(t.Context()))
}

func migrationVersion(t *testing.T, p *goose.Provider, file string) int64 {
	t.Helper()
	for _, s := range p.ListSources() {
		if filepath.Base(s.Path) == file {
			return s.Version
		}
	}
	t.Fatalf("migration %s not embedded", file)
	return 0
}
