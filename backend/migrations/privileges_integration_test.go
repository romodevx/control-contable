//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// queryStrings returns the first column of every row of query.
func queryStrings(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), query)
	require.NoError(t, err)
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		out = append(out, s)
	}
	require.NoError(t, rows.Err())
	return out
}

// Nothing in schema public is granted to PUBLIC (grantee 0): every privilege is per role.
// A NULL ACL means the built-in default, which for functions includes EXECUTE to PUBLIC.
func TestSchemaPublic_NothingGrantedToPublic(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	const query = `
SELECT c.relname || ': ' || a.privilege_type
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
CROSS JOIN LATERAL pg_catalog.aclexplode(coalesce(c.relacl, pg_catalog.acldefault('r', c.relowner))) a
WHERE n.nspname = 'public' AND a.grantee = 0
UNION ALL
SELECT p.oid::regprocedure::text || ': ' || a.privilege_type
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
CROSS JOIN LATERAL pg_catalog.aclexplode(coalesce(p.proacl, pg_catalog.acldefault('f', p.proowner))) a
WHERE n.nspname = 'public' AND a.grantee = 0
UNION ALL
SELECT 'schema public: ' || a.privilege_type
FROM pg_catalog.pg_namespace n
CROSS JOIN LATERAL pg_catalog.aclexplode(n.nspacl) a
WHERE n.nspname = 'public' AND a.grantee = 0`

	assert.Empty(t, queryStrings(t, owner, query))
}

// Runtime roles never hold DDL rights: the runner, as cc_owner, is the only DDL path.
func TestSchemaPublic_RuntimeRolesCannotCreate(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			_, err := db.Connect(t, role).ExecContext(t.Context(), "CREATE TABLE public.rogue (id int)")
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// A SECURITY DEFINER function in schema public must pin its search_path.
func TestSchemaPublic_DefinerFunctionsPinSearchPath(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	const query = `
SELECT p.oid::regprocedure::text
FROM pg_catalog.pg_proc p
JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'public' AND p.prosecdef
  AND NOT EXISTS (
    SELECT 1 FROM pg_catalog.unnest(p.proconfig) AS s(setting) WHERE s.setting LIKE 'search_path=%')`

	assert.Empty(t, queryStrings(t, owner, query))
}
