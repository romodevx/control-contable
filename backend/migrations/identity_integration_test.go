//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Identity migration (P0-11): users, user_usage, entitlements, email_verifications.

const (
	userB = "00000000-0000-4000-8000-00000000000b"
	userC = "00000000-0000-4000-8000-00000000000c"

	sqlstateUniqueViolation = "23505"
	sqlstateRaiseException  = "P0001"
)

var (
	usersTable        = pgtest.Table{Name: "users", OwnerColumn: "id"}
	userUsageTable    = pgtest.Table{Name: "user_usage", OwnerColumn: "user_id"}
	entitlementsTable = pgtest.Table{Name: "entitlements", OwnerColumn: "user_id"}
)

// insertUserSQL writes a user whose email columns derive from its id: a 32-byte blind
// index and a ciphertext framed like crypto.Keyring output (key_id 1 || nonce || ct || tag).
const insertUserSQL = `
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')),
        '\x0001'::bytea || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8')),
        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana')`

const (
	insertUsageSQL      = "INSERT INTO user_usage (user_id) VALUES ($1)"
	insertTrialSQL      = "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '7 days')"
	insertProSQL        = "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at, rc_product_id, rc_original_transaction_id) VALUES ($1, 'pro', 'revenuecat', now(), now() + interval '1 month', 'cc_monthly', 'GPA.0001')"
	insertCompSQL       = "INSERT INTO entitlements (user_id, kind, source, starts_at) VALUES ($1, 'comp', 'admin', now())"
	insertVerificationQ = "INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at) VALUES (sha256('x'), sha256('123456'), '{}', now() + interval '15 minutes')"
)

func execAs(query string) func(ctx context.Context, tx *sql.Tx, ownerID string) error {
	return func(ctx context.Context, tx *sql.Tx, ownerID string) error {
		_, err := tx.ExecContext(ctx, query, ownerID)
		return err
	}
}

// seedAs runs query (with $1 = userID) under userID's own context and commits.
func seedAs(t *testing.T, db *sql.DB, userID, query string) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), query, userID)
		return err
	}))
}

// seedUsers registers each user as cc_app does at sign-up: users + user_usage + trial.
func seedUsers(t *testing.T, app *sql.DB, ids ...string) {
	t.Helper()
	for _, id := range ids {
		seedAs(t, app, id, insertUserSQL)
		seedAs(t, app, id, insertUsageSQL)
		seedAs(t, app, id, insertTrialSQL)
	}
}

// execAsUser runs query under userID's context and returns the affected row count.
func execAsUser(t *testing.T, db *sql.DB, userID, query string, args ...any) (int64, error) {
	t.Helper()
	var affected int64
	err := pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(t.Context(), query, args...)
		if err != nil {
			return err
		}
		affected, err = res.RowsAffected()
		return err
	})
	return affected, err
}

func TestIdentityTables_AppRoleIsolation(t *testing.T) {
	specs := []pgtest.IsolationSpec{
		{Table: usersTable, Insert: execAs(insertUserSQL)},
		{Table: userUsageTable, Insert: execAs(insertUsageSQL)},
		{Table: entitlementsTable, Insert: execAs(insertTrialSQL)},
	}
	for _, spec := range specs {
		t.Run(spec.Table.Name, func(t *testing.T) {
			db := pgtest.NewDatabase(t)
			app := db.Connect(t, pgtest.RoleApp)
			if spec.Table != usersTable {
				seedAs(t, app, userA, insertUserSQL)
				seedAs(t, app, userB, insertUserSQL)
			}
			spec.UserA, spec.UserB = userA, userB

			pgtest.AssertIsolation(t, app, spec)
		})
	}
}

// The webhook writes pro entitlements for the user named by the RevenueCat event.
func TestEntitlements_WebhookRoleIsolation(t *testing.T) {
	db := pgtest.NewDatabase(t)
	seedAs(t, db.Connect(t, pgtest.RoleApp), userA, insertUserSQL)
	seedAs(t, db.Connect(t, pgtest.RoleApp), userB, insertUserSQL)

	pgtest.AssertIsolation(t, db.Connect(t, pgtest.RoleWebhook), pgtest.IsolationSpec{
		Table: entitlementsTable, UserA: userA, UserB: userB, Insert: execAs(insertProSQL),
	})
}

// The generic UPDATE probe of AssertIsolation rewrites the owner column, which these
// roles may not update at all; prove cross-user updates on the columns they may write.
func TestIdentityTables_CrossUserUpdatesMatchNoRow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	seedAs(t, db.Connect(t, pgtest.RoleWebhook), userB, insertProSQL)
	updates := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "UPDATE users SET name = 'Mallory' WHERE id = $1"},
		{pgtest.RoleApp, "UPDATE user_usage SET incomes_created = incomes_created + 1 WHERE user_id = $1"},
		{pgtest.RoleWebhook, "UPDATE entitlements SET ends_at = ends_at + interval '1 year' WHERE user_id = $1"},
		{pgtest.RoleApp, "DELETE FROM users WHERE id = $1"},
	}
	for _, u := range updates {
		t.Run(string(u.role)+": "+u.query, func(t *testing.T) {
			affected, err := execAsUser(t, db.Connect(t, u.role), userA, u.query, userB)
			require.NoError(t, err)
			assert.Zero(t, affected)
		})
	}
	t.Run("own rows are writable", func(t *testing.T) {
		for _, u := range updates[:3] {
			affected, err := execAsUser(t, db.Connect(t, u.role), userB, u.query, userB)
			require.NoError(t, err, u.query)
			assert.EqualValues(t, 1, affected, u.query)
		}
	})
}

// Least privilege: each runtime role reaches only what data-model.md grants it.
func TestIdentityTables_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	seedUsers(t, db.Connect(t, pgtest.RoleApp), userA)
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, "SELECT 1 FROM email_verifications"},
		{pgtest.RoleApp, insertVerificationQ},
		{pgtest.RoleApp, "UPDATE users SET id = id"},
		{pgtest.RoleApp, "UPDATE users SET status = 'suspended'"},
		{pgtest.RoleApp, "UPDATE users SET created_at = now()"},
		{pgtest.RoleApp, "UPDATE user_usage SET user_id = user_id"},
		{pgtest.RoleApp, "DELETE FROM user_usage"},
		{pgtest.RoleApp, "UPDATE entitlements SET ends_at = NULL"},
		{pgtest.RoleApp, "DELETE FROM entitlements"},
		{pgtest.RoleApp, "TRUNCATE users CASCADE"},
		{pgtest.RoleWebhook, "SELECT 1 FROM users"},
		{pgtest.RoleWebhook, "SELECT 1 FROM user_usage"},
		{pgtest.RoleWebhook, "SELECT 1 FROM email_verifications"},
		{pgtest.RoleWebhook, "UPDATE entitlements SET user_id = user_id"},
		{pgtest.RoleWorker, "SELECT 1 FROM users"},
		{pgtest.RoleWorker, "SELECT 1 FROM user_usage"},
		{pgtest.RoleWorker, "SELECT 1 FROM entitlements"},
		{pgtest.RoleWorker, "SELECT 1 FROM email_verifications"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, d.role), userA, d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// RLS narrows the kinds each role may write: cc_app only starts the trial, cc_webhook
// only writes pro, and comp is granted by an operator as cc_owner.
func TestEntitlements_EachRoleWritesOnlyItsKind(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedAs(t, app, userA, insertUserSQL)
	refused := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleApp, insertProSQL},
		{pgtest.RoleApp, insertCompSQL},
		{pgtest.RoleWebhook, insertTrialSQL},
		{pgtest.RoleWebhook, insertCompSQL},
		{pgtest.RoleOwner, insertTrialSQL},
		{pgtest.RoleOwner, insertProSQL},
	}
	for _, r := range refused {
		t.Run(string(r.role)+" refused: "+r.query, func(t *testing.T) {
			_, err := execAsUser(t, db.Connect(t, r.role), userA, r.query, userA)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}

	seedAs(t, app, userA, insertTrialSQL)
	seedAs(t, db.Connect(t, pgtest.RoleOwner), userA, insertCompSQL)
	affected, err := execAsUser(t, db.Connect(t, pgtest.RoleWebhook), userA,
		"UPDATE entitlements SET ends_at = NULL WHERE user_id = $1", userA)
	require.NoError(t, err)
	assert.Zero(t, affected, "cc_webhook leaves trial and comp rows alone")

	var kinds []string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(t.Context(), "SELECT kind FROM entitlements ORDER BY kind")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err != nil {
				return err
			}
			kinds = append(kinds, k)
		}
		return rows.Err()
	}))
	assert.Equal(t, []string{"comp", "trial"}, kinds, "cc_app reads every kind of its own")
}

// email_verifications has no user to key on: deny-all to every runtime role, even with
// a mistaken grant, while SECURITY DEFINER functions (owned by cc_owner) can use it.
func TestEmailVerifications_ServiceOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	var enabled, forced bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = 'email_verifications'::regclass").
		Scan(&enabled, &forced))
	assert.True(t, enabled, "ENABLE ROW LEVEL SECURITY")
	assert.False(t, forced, "not FORCE: the owner-held definer functions must reach it")

	_, err := owner.ExecContext(t.Context(), insertVerificationQ)
	require.NoError(t, err, "cc_owner (definer functions) writes")
	_, err = owner.ExecContext(t.Context(), "GRANT SELECT, UPDATE, DELETE ON email_verifications TO cc_app")
	require.NoError(t, err)

	app := db.Connect(t, pgtest.RoleApp)
	var seen int
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM email_verifications").Scan(&seen))
	assert.Zero(t, seen, "no policy for any runtime role")
	res, err := app.ExecContext(t.Context(), "DELETE FROM email_verifications")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	assert.Zero(t, affected)
}

// Account deletion (US-A4): deleting the users row removes its usage and entitlements,
// although cc_app holds no DELETE on either.
func TestUsers_DeletionCascades(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)

	affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)

	for _, table := range []pgtest.Table{userUsageTable, entitlementsTable} {
		assert.Zero(t, countOwned(t, app, userA, table), table.Name)
		assert.Equal(t, 1, countOwned(t, app, userB, table), table.Name+": other users untouched")
	}
}

// countOwned counts userID's rows of table under userID's own context.
func countOwned(t *testing.T, db *sql.DB, userID string, table pgtest.Table) int {
	t.Helper()
	var n int
	require.NoError(t, pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error {
		query := "SELECT count(*) FROM " + pgtest.QuoteIdent(table.Name) + " WHERE " + pgtest.QuoteIdent(table.OwnerColumn) + " = $1"
		return tx.QueryRowContext(t.Context(), query, userID).Scan(&n)
	}))
	return n
}

func TestUserUsage_CountersNeverDecrease(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)

	_, err := execAsUser(t, app, userA, "UPDATE user_usage SET incomes_created = 5, expenses_created = 5")
	require.NoError(t, err)
	for _, col := range []string{"incomes_created", "expenses_created"} {
		t.Run(col, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, "UPDATE user_usage SET "+col+" = "+col+" - 1")
			requireSQLState(t, err, sqlstateRaiseException)
		})
	}
}

// The trial-cap statement of data-model.md: atomic reserve, 0 rows at the cap.
func TestUserUsage_TrialCapReservation(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	const reserve = `UPDATE user_usage SET incomes_created = incomes_created + 1
		WHERE user_id = $1 AND incomes_created < 5`

	for i := range 5 {
		affected, err := execAsUser(t, app, userA, reserve, userA)
		require.NoError(t, err)
		require.EqualValues(t, 1, affected, "reservation %d", i+1)
	}
	affected, err := execAsUser(t, app, userA, reserve, userA)
	require.NoError(t, err)
	assert.Zero(t, affected, "the 6th income is refused")
}

// Every case runs as the role whose RLS policy admits the row, so the constraint under
// test is what refuses it (RLS WITH CHECK runs before CHECK constraints).
func TestIdentityTables_RejectInvalidRows(t *testing.T) {
	db := pgtest.NewDatabase(t)
	seedUsers(t, db.Connect(t, pgtest.RoleApp), userA)
	const (
		user    = "INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name) VALUES ($1, "
		okEmail = "sha256('c'), '\\x0001'::bytea || sha256('c') || sha256('c'), 1, "
		okHash  = "'$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', "
		verify  = "INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at) VALUES (sha256('y'), "
	)
	bad := []struct {
		name  string
		role  pgtest.Role
		user  string
		query string
		state string
	}{
		{"email_hash not 32 bytes", pgtest.RoleApp, userC, user + "'\\x00', '\\x0001'::bytea || sha256('c') || sha256('c'), 1, " + okHash + "'C')", sqlstateCheckViolation},
		{"ciphertext shorter than header and tag", pgtest.RoleApp, userC, user + "sha256('c'), '\\x0001'::bytea || substring(sha256('c') from 1 for 28), 1, " + okHash + "'C')", sqlstateCheckViolation},
		{"key id differs from the ciphertext prefix", pgtest.RoleApp, userC, user + "sha256('c'), '\\x0002'::bytea || sha256('c') || sha256('c'), 1, " + okHash + "'C')", sqlstateCheckViolation},
		{"key id out of range", pgtest.RoleApp, userC, user + "sha256('c'), '\\x0000'::bytea || sha256('c') || sha256('c'), 0, " + okHash + "'C')", sqlstateCheckViolation},
		{"password not an argon2id PHC string", pgtest.RoleApp, userC, user + okEmail + "'plain', 'C')", sqlstateCheckViolation},
		{"blank name", pgtest.RoleApp, userC, user + okEmail + okHash + "'  ')", sqlstateCheckViolation},
		{"duplicate email_hash", pgtest.RoleApp, userC, user + "sha256(convert_to('" + userA + "', 'UTF8')), '\\x0001'::bytea || sha256('c') || sha256('c'), 1, " + okHash + "'C')", sqlstateUniqueViolation},
		{"trial without an end", pgtest.RoleApp, userA, "INSERT INTO entitlements (user_id, kind, source, starts_at) VALUES ($1, 'trial', 'registration', now())", sqlstateCheckViolation},
		{"second trial", pgtest.RoleApp, userA, insertTrialSQL, sqlstateUniqueViolation},
		{"kind and source mismatch", pgtest.RoleOwner, userA, "INSERT INTO entitlements (user_id, kind, source, starts_at) VALUES ($1, 'comp', 'registration', now())", sqlstateCheckViolation},
		{"period ends before it starts", pgtest.RoleOwner, userA, "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'comp', 'admin', now(), now() - interval '1 day')", sqlstateCheckViolation},
		{"revenuecat ids on a comp", pgtest.RoleOwner, userA, "INSERT INTO entitlements (user_id, kind, source, starts_at, rc_product_id) VALUES ($1, 'comp', 'admin', now(), 'x')", sqlstateCheckViolation},
		{"revenuecat row without its ids", pgtest.RoleWebhook, userA, "INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'pro', 'revenuecat', now(), now() + interval '1 month')", sqlstateCheckViolation},
		{"verification code_hash not 32 bytes", pgtest.RoleOwner, "", verify + "'\\x01', '{}', now() + interval '1 hour')", sqlstateCheckViolation},
		{"verification pending not an object", pgtest.RoleOwner, "", verify + "sha256('1'), '[]', now() + interval '1 hour')", sqlstateCheckViolation},
		{"verification negative attempts", pgtest.RoleOwner, "", "INSERT INTO email_verifications (email_hash, code_hash, pending, attempts, expires_at) VALUES (sha256('y'), sha256('1'), '{}', -1, now() + interval '1 hour')", sqlstateCheckViolation},
		{"verification already expired", pgtest.RoleOwner, "", verify + "sha256('1'), '{}', now() - interval '1 hour')", sqlstateCheckViolation},
	}
	for _, b := range bad {
		t.Run(b.name, func(t *testing.T) {
			conn := db.Connect(t, b.role)
			var err error
			if b.user == "" {
				_, err = conn.ExecContext(t.Context(), b.query)
			} else {
				_, err = execAsUser(t, conn, b.user, b.query, b.user)
			}
			requireSQLState(t, err, b.state)
		})
	}
}

func TestIdentityTables_StampDefaults(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)

	var status string
	var counters int
	var stamped bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(t.Context(),
			"SELECT u.status, g.incomes_created + g.expenses_created FROM users u JOIN user_usage g ON g.user_id = u.id").
			Scan(&status, &counters); err != nil {
			return err
		}
		return tx.QueryRowContext(t.Context(),
			"UPDATE users SET name = 'Ana B' RETURNING updated_at = now()").Scan(&stamped)
	}))
	assert.Equal(t, "active", status)
	assert.Zero(t, counters, "usage starts at zero")
	assert.True(t, stamped, "the trigger stamps updated_at")
}
