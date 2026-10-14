//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Account deletion (US-A4, M1-09) is one DELETE of the users row: every
// user-owned row must go with it through a foreign key ON DELETE CASCADE.
// This guard covers the tables of every later module too: a new table that
// references users with any other action (RESTRICT, NO ACTION, SET NULL)
// would make deletion fail, or keep the user's data, and fails here first.

// usersForeignKeysSQL lists every foreign key referencing users, with its
// ON DELETE action (pg_constraint.confdeltype: 'c' = CASCADE).
const usersForeignKeysSQL = `
SELECT conrelid::regclass::text, conname, confdeltype::text
FROM pg_constraint
WHERE contype = 'f' AND confrelid = 'users'::regclass
ORDER BY 1, 2`

type usersForeignKey struct {
	table, name, onDelete string
}

func TestAccountDeletion_everyForeignKeyToUsersCascades(t *testing.T) {
	db := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	rows, err := db.QueryContext(t.Context(), usersForeignKeysSQL)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var keys []usersForeignKey
	for rows.Next() {
		var k usersForeignKey
		require.NoError(t, rows.Scan(&k.table, &k.name, &k.onDelete))
		keys = append(keys, k)
	}
	require.NoError(t, rows.Err())

	tables := make([]string, 0, len(keys))
	for _, k := range keys {
		tables = append(tables, k.table)
		assert.Equalf(t, "c", k.onDelete, "%s.%s must be ON DELETE CASCADE (account deletion, US-A4)", k.table, k.name)
	}
	assert.Subset(t, tables, []string{"user_usage", "entitlements", "consents", "refresh_tokens"},
		"the M1 user-owned tables reference users")
	assert.Subset(t, tables, []string{"taxpayers", "taxpayer_regimes"},
		"the M2 user-owned tables reference users directly, not only through their parent")
	assert.Subset(t, tables, []string{"incomes", "expenses", "taxpayer_tax_inputs"},
		"the M3/M4/M5 user-owned tables reference users directly, not only through their parent")
	assert.NotContains(t, tables, "trial_registry", "the trial registry must survive account deletion")
}

// userColumnsWithoutForeignKeySQL lists every *user_id column of a public
// table that no foreign key to users covers: a table keyed by a user with
// no foreign key escapes the cascade above, and its rows outlive the
// account (the check on confdeltype never sees it).
const userColumnsWithoutForeignKeySQL = `
SELECT c.table_name::text, c.column_name::text
FROM information_schema.columns c
JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = 'public' AND t.table_type = 'BASE TABLE' AND c.column_name ~ '(^|_)user_id$'
  AND NOT EXISTS (
    SELECT 1 FROM pg_constraint k
    JOIN pg_attribute a ON a.attrelid = k.conrelid AND a.attnum = ANY (k.conkey)
    WHERE k.contype = 'f' AND k.confrelid = 'users'::regclass
      AND k.conrelid = format('public.%I', c.table_name)::regclass AND a.attname = c.column_name)
ORDER BY 1, 2`

func userColumnsWithoutForeignKey(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), userColumnsWithoutForeignKeySQL)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var found []string
	for rows.Next() {
		var table, column string
		require.NoError(t, rows.Scan(&table, &column))
		found = append(found, table+"."+column)
	}
	require.NoError(t, rows.Err())
	return found
}

func TestAccountDeletion_everyUserIDColumnHasAForeignKeyToUsers(t *testing.T) {
	db := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	assert.Empty(t, userColumnsWithoutForeignKey(t, db),
		"reference users ON DELETE CASCADE, or account deletion (US-A4) leaves these rows behind")

	// The guard itself: a user-keyed table without the foreign key is caught.
	_, err := db.ExecContext(t.Context(), `CREATE TABLE guard_probe (id int PRIMARY KEY, owner_user_id uuid NOT NULL)`)
	require.NoError(t, err)
	assert.Equal(t, []string{"guard_probe.owner_user_id"}, userColumnsWithoutForeignKey(t, db))
}

// userScopedChildForeignKeysSQL lists every foreign key from one table with a user_id
// column to another such table (a parent/child chain under one user, e.g.
// taxpayer_regimes -> taxpayers), with its ON DELETE action. Deleting the parent
// (a taxpayer, MODULE-taxpayers §2.7) or the user must take the children with it.
// Self-references (refresh_tokens.rotated_from) are not chains: the users cascade
// deletes every row of the table for that user in the same statement.
const userScopedChildForeignKeysSQL = `
SELECT k.conrelid::regclass::text, k.conname, k.confdeltype::text
FROM pg_constraint k
WHERE k.contype = 'f' AND k.confrelid <> 'users'::regclass AND k.confrelid <> k.conrelid
  AND EXISTS (SELECT 1 FROM pg_attribute a WHERE a.attrelid = k.conrelid AND a.attname = 'user_id' AND NOT a.attisdropped)
  AND EXISTS (SELECT 1 FROM pg_attribute a WHERE a.attrelid = k.confrelid AND a.attname = 'user_id' AND NOT a.attisdropped)
ORDER BY 1, 2`

// S-7 (MODULE-taxpayers): the guard covers indirect chains too.
func TestAccountDeletion_everyUserScopedChainCascades(t *testing.T) {
	db := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	rows, err := db.QueryContext(t.Context(), userScopedChildForeignKeysSQL)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var tables []string
	for rows.Next() {
		var k usersForeignKey
		require.NoError(t, rows.Scan(&k.table, &k.name, &k.onDelete))
		tables = append(tables, k.table)
		assert.Equalf(t, "c", k.onDelete, "%s.%s must be ON DELETE CASCADE", k.table, k.name)
	}
	require.NoError(t, rows.Err())
	assert.Contains(t, tables, "taxpayer_regimes", "the sweep is not vacuous")
}
