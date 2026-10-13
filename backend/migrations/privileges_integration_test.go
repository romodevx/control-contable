//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// queryStrings returns the first column of every row of query (with args bound).
func queryStrings(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), query, args...)
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

// unforcedTables are the tables allowed to skip FORCE ROW LEVEL SECURITY, each with
// RLS still ENABLED: global data with no user to key on, or service-only rows the
// owner's definer functions must reach. Adding a table here needs the security gate.
var unforcedTables = map[string]string{
	"feature_flags":               "global module flags (baseline)",
	"flag_presets":                "global presets (baseline)",
	"flag_preset_flags":           "global presets (baseline)",
	"email_verifications":         "service-only; reached by cc_auth_definer through its own policy",
	"trial_registry":              "service-only; reached by cc_auth_definer through its own policy",
	"email_verification_failures": "service-only; reached by cc_auth_definer through its own policy",
	"legal_documents":             "global legal document versions (M1-03)",
}

// rlsGapsQuery lists every table in public that lacks ENABLE, or lacks FORCE without
// an allow-list entry ($1). goose_db_version is the runner's own bookkeeping.
const rlsGapsQuery = `
SELECT c.relname || CASE WHEN NOT c.relrowsecurity THEN ': RLS not enabled' ELSE ': RLS not forced' END
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p') AND c.relname <> 'goose_db_version'
  AND (NOT c.relrowsecurity OR (NOT c.relforcerowsecurity AND NOT c.relname = ANY($1)))
ORDER BY 1`

func rlsGaps(t *testing.T, owner *sql.DB) []string {
	t.Helper()
	allowed := make([]string, 0, len(unforcedTables))
	for name := range unforcedTables {
		allowed = append(allowed, name)
	}
	rows, err := owner.QueryContext(t.Context(), rlsGapsQuery, allowed)
	require.NoError(t, err)
	defer rows.Close()
	var gaps []string
	for rows.Next() {
		var g string
		require.NoError(t, rows.Scan(&g))
		gaps = append(gaps, g)
	}
	require.NoError(t, rows.Err())
	return gaps
}

// P0-28: every table a migration adds is fail-closed by default. A new user-scoped
// table without ENABLE + FORCE fails here even if its own test forgot AssertIsolation.
func TestSchemaPublic_EveryTableHasRowLevelSecurity(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	assert.Empty(t, rlsGaps(t, owner))

	t.Run("the sweep is not vacuous", func(t *testing.T) {
		for _, ddl := range []string{
			"CREATE TABLE qa_no_rls (user_id uuid)",
			"CREATE TABLE qa_not_forced (user_id uuid); ALTER TABLE qa_not_forced ENABLE ROW LEVEL SECURITY",
		} {
			_, err := owner.ExecContext(t.Context(), ddl)
			require.NoError(t, err)
		}
		assert.Equal(t, []string{"qa_no_rls: RLS not enabled", "qa_not_forced: RLS not forced"}, rlsGaps(t, owner))
	})
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
