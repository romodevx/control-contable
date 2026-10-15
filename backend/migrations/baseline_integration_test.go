//go:build integration

package migrations_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

const (
	userA = "00000000-0000-4000-8000-00000000000a"

	sqlstateInsufficientPrivilege = "42501"
	sqlstateInvalidTextRepr       = "22P02"
	sqlstateCheckViolation        = "23514"
)

// moduleFlags is the MODULES.md flag registry: one key per product module
// M1–M9, plus module_annual (MODULE-annual §6, registered by ANN-22).
var moduleFlags = []string{
	"module_annual", "module_auth", "module_dashboard", "module_expenses", "module_history", "module_incomes",
	"module_reminders", "module_subscriptions", "module_tax_engine", "module_taxpayers",
}

// presets is the MODULES.md "Presentation presets" table.
var presets = map[string][]string{
	"demo_core": {
		"module_auth", "module_dashboard", "module_expenses", "module_incomes", "module_tax_engine",
		"module_taxpayers",
	},
	"demo_monetization": {
		"module_auth", "module_dashboard", "module_expenses", "module_incomes", "module_subscriptions",
		"module_tax_engine", "module_taxpayers",
	},
	"demo_full": moduleFlags,
}

func requireSQLState(t *testing.T, err error, code string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr, "want SQLSTATE %s, got %v", code, err)
	assert.Equal(t, code, pgErr.Code, pgErr.Message)
}

// appUserID reads app_user_id() in one transaction under userID ("" = unset).
func appUserID(t *testing.T, db *sql.DB, userID string) (sql.NullString, error) {
	t.Helper()
	var got sql.NullString
	read := func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT app_user_id()::text").Scan(&got)
	}
	if userID == "" {
		return got, pgtest.WithoutUser(t.Context(), db, read)
	}
	return got, pgtest.AsUser(t.Context(), db, userID, read)
}

func TestAppUserID_ReturnsTheTransactionUser(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)

	got, err := appUserID(t, app, userA)

	require.NoError(t, err)
	assert.Equal(t, sql.NullString{String: userA, Valid: true}, got)
}

// Both "no context" session states must yield NULL, so `user_id = app_user_id()`
// matches no row: a never-set session and one whose transaction-local value reverted to the empty string.
func TestAppUserID_IsNullWhenTheContextIsUnset(t *testing.T) {
	db := pgtest.NewDatabase(t)

	t.Run("never set", func(t *testing.T) {
		got, err := appUserID(t, db.Connect(t, pgtest.RoleApp), "")
		require.NoError(t, err)
		assert.False(t, got.Valid, "got %q", got.String)
	})
	t.Run("reverted after a transaction", func(t *testing.T) {
		app := db.Connect(t, pgtest.RoleApp)
		app.SetMaxOpenConns(1) // same session for both transactions
		_, err := appUserID(t, app, userA)
		require.NoError(t, err)

		got, err := appUserID(t, app, "")

		require.NoError(t, err)
		assert.False(t, got.Valid, "got %q", got.String)
	})
}

func TestAppUserID_FailsClosedOnAMalformedContext(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)

	_, err := appUserID(t, app, "not-a-uuid")

	requireSQLState(t, err, sqlstateInvalidTextRepr)
}

func TestAppUserID_ExecutableByEveryRuntimeRole(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			got, err := appUserID(t, db.Connect(t, role), userA)
			require.NoError(t, err)
			assert.Equal(t, userA, got.String)
		})
	}
}

func TestFeatureFlags_SeedsEveryModuleFlagOff(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	rows, err := owner.QueryContext(t.Context(), "SELECT key, enabled FROM feature_flags ORDER BY key")
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	var keys []string
	for rows.Next() {
		var (
			key     string
			enabled bool
		)
		require.NoError(t, rows.Scan(&key, &enabled))
		assert.False(t, enabled, "%s must default OFF", key)
		keys = append(keys, key)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, moduleFlags, keys)
}

func TestFlagPresets_MatchTheModulesRegistry(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	for name, want := range presets {
		t.Run(name, func(t *testing.T) {
			var got []string
			rows, err := owner.QueryContext(t.Context(),
				"SELECT flag_key FROM flag_preset_flags WHERE preset_name = $1 ORDER BY flag_key", name)
			require.NoError(t, err)
			defer func() { require.NoError(t, rows.Close()) }()
			for rows.Next() {
				var key string
				require.NoError(t, rows.Scan(&key))
				got = append(got, key)
			}
			require.NoError(t, rows.Err())
			assert.Equal(t, want, got)
		})
	}
	var n int
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT count(*) FROM flag_presets").Scan(&n))
	assert.Equal(t, len(presets), n, "no preset beyond the registry")
}

func TestFeatureFlags_RuntimeRolesAreReadOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	writes := map[string]string{
		"insert":   "INSERT INTO feature_flags (key, description) VALUES ('module_rogue', 'rogue')",
		"update":   "UPDATE feature_flags SET enabled = true",
		"delete":   "DELETE FROM feature_flags",
		"truncate": "TRUNCATE feature_flags CASCADE",
	}
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWorker} {
		conn := db.Connect(t, role)
		t.Run(string(role)+"/select", func(t *testing.T) {
			var n int
			require.NoError(t, conn.QueryRowContext(t.Context(), "SELECT count(*) FROM feature_flags").Scan(&n))
			assert.Equal(t, len(moduleFlags), n)
		})
		for verb, stmt := range writes {
			t.Run(string(role)+"/"+verb, func(t *testing.T) {
				_, err := conn.ExecContext(t.Context(), stmt)
				requireSQLState(t, err, sqlstateInsufficientPrivilege)
			})
		}
	}
}

func TestFlagTables_NoAccessBeyondTheGrants(t *testing.T) {
	db := pgtest.NewDatabase(t)
	denied := []struct {
		role  pgtest.Role
		query string
	}{
		{pgtest.RoleWebhook, "SELECT 1 FROM feature_flags"},
		{pgtest.RoleApp, "SELECT 1 FROM flag_presets"},
		{pgtest.RoleApp, "SELECT 1 FROM flag_preset_flags"},
		{pgtest.RoleWorker, "SELECT 1 FROM flag_presets"},
		{pgtest.RoleWebhook, "SELECT 1 FROM flag_presets"},
		{pgtest.RoleApp, "SELECT 1 FROM goose_db_version"},
	}
	for _, d := range denied {
		t.Run(string(d.role)+": "+d.query, func(t *testing.T) {
			_, err := db.Connect(t, d.role).ExecContext(t.Context(), d.query)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// Defense in depth: if a later migration grants a write by mistake, RLS still has
// no write policy for the runtime roles, so the write matches no row or is refused.
func TestFlagTables_RowSecurityDeniesUngrantedWrites(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	_, err := owner.ExecContext(t.Context(),
		"GRANT INSERT, UPDATE, DELETE ON feature_flags TO cc_app; GRANT SELECT ON flag_presets TO cc_app")
	require.NoError(t, err)
	app := db.Connect(t, pgtest.RoleApp)

	_, err = app.ExecContext(t.Context(), "INSERT INTO feature_flags (key, description) VALUES ('module_rogue', 'rogue')")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)

	for _, stmt := range []string{"UPDATE feature_flags SET enabled = true", "DELETE FROM feature_flags"} {
		res, err := app.ExecContext(t.Context(), stmt)
		require.NoError(t, err, stmt)
		affected, err := res.RowsAffected()
		require.NoError(t, err)
		assert.Zero(t, affected, stmt)
	}

	var presetsSeen int
	require.NoError(t, app.QueryRowContext(t.Context(), "SELECT count(*) FROM flag_presets").Scan(&presetsSeen))
	assert.Zero(t, presetsSeen, "flag_presets has no runtime read policy")
}

func TestFeatureFlags_OwnerTogglesAndUpdatedAtFollows(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, tx.Rollback()) }()

	var stamped bool
	require.NoError(t, tx.QueryRowContext(t.Context(), `
		UPDATE feature_flags SET enabled = true, updated_at = '2000-01-01'
		WHERE key = 'module_auth'
		RETURNING enabled AND updated_at = now()`).Scan(&stamped))

	assert.True(t, stamped, "the trigger stamps updated_at with the transaction time")
}

func TestFlagTables_RejectMalformedNames(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	bad := []string{
		"INSERT INTO feature_flags (key, description) VALUES ('auth', 'missing prefix')",
		"INSERT INTO feature_flags (key, description) VALUES ('module_Auth', 'not snake_case')",
		"INSERT INTO feature_flags (key, description) VALUES ('module_auth_', 'trailing underscore')",
		"INSERT INTO feature_flags (key, description) VALUES ('module_x', '')",
		"INSERT INTO flag_presets (name, description) VALUES ('Demo', 'not snake_case')",
	}
	for _, stmt := range bad {
		t.Run(stmt, func(t *testing.T) {
			_, err := owner.ExecContext(t.Context(), stmt)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
}

func TestFlagPresetFlags_ReferenceRealFlags(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	_, err := owner.ExecContext(t.Context(),
		"INSERT INTO flag_preset_flags (preset_name, flag_key) VALUES ('demo_core', 'module_ghost')")

	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "want a foreign-key violation, got %v", err)
	assert.Equal(t, "23503", pgErr.Code)
}
