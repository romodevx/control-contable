//go:build integration

package migrations_test

import (
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
	assert.NotContains(t, tables, "trial_registry", "the trial registry must survive account deletion")
}
