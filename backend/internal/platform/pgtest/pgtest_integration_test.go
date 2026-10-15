//go:build integration

package pgtest_test

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Fixed user ids: the suite never depends on random data.
const (
	userA = "00000000-0000-4000-8000-00000000000a"
	userB = "00000000-0000-4000-8000-00000000000b"
)

var (
	notes       = pgtest.Table{Name: "notes", OwnerColumn: "user_id"}
	leakyNotes  = pgtest.Table{Name: "leaky_notes", OwnerColumn: "user_id"}
	strictNotes = pgtest.Table{Name: "strict_notes", OwnerColumn: "user_id"}

	errBoom = errors.New("boom")
)

// insertInto returns an insert step that writes one row owned by ownerID into table.
func insertInto(table pgtest.Table) func(ctx context.Context, tx *sql.Tx, ownerID string) error {
	return func(ctx context.Context, tx *sql.Tx, ownerID string) error {
		query := "INSERT INTO " + pgtest.QuoteIdent(table.Name) + " (user_id, body) VALUES ($1, 'note')" //nolint:gosec // G202: the table name is quoted; the owner id is a bind parameter
		_, err := tx.ExecContext(ctx, query, ownerID)
		return err
	}
}

func seed(t *testing.T, db *sql.DB, table pgtest.Table, ownerID string) {
	t.Helper()
	insert := insertInto(table)
	require.NoError(t, pgtest.AsUser(t.Context(), db, ownerID, func(tx *sql.Tx) error {
		return insert(t.Context(), tx, ownerID)
	}))
}

func countAs(t *testing.T, db *sql.DB, userID string, table pgtest.Table) int {
	t.Helper()
	var n int
	require.NoError(t, pgtest.AsUser(t.Context(), db, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+pgtest.QuoteIdent(table.Name)).Scan(&n)
	}))
	return n
}

func TestBootstrap_RuntimeRolesArePlainLogins(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			conn := db.Connect(t, role)
			var (
				current               string
				super, bypass, create bool
				dbOwner               string
			)
			require.NoError(t, conn.QueryRowContext(t.Context(), `
				SELECT current_user, r.rolsuper, r.rolbypassrls,
				       has_schema_privilege(current_user, 'public', 'CREATE'),
				       pg_get_userbyid(d.datdba)
				FROM pg_roles r JOIN pg_database d ON d.datname = current_database()
				WHERE r.rolname = current_user`).Scan(&current, &super, &bypass, &create, &dbOwner))
			assert.Equal(t, string(role), current, "logged in as the requested role")
			assert.False(t, super, "SUPERUSER")
			assert.False(t, bypass, "BYPASSRLS")
			assert.False(t, create, "CREATE on schema public")
			assert.Equal(t, string(pgtest.RoleOwner), dbOwner, "database owner")
		})
	}
}

// Regression (PR #5 follow-ups 2 and 3): the maintenance databases kept the
// default ACL, so PUBLIC (every runtime role) could log in to "postgres" and
// "template1" and create temp objects there. The bootstrap revokes that, and
// no bootstrap role may replicate or reach another role through membership
// (SET ROLE would hand a NOINHERIT runtime role that role's privileges). The one
// exception: cc_owner may SET ROLE to the NOLOGIN definer roles (cc_auth_definer,
// cc_worker_definer), inheriting nothing.
func TestBootstrap_RolesCannotReachMaintenanceDatabasesOrOtherRoles(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleOwner, pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			conn := db.Connect(t, role)
			var replication, memberOfAny bool
			// The allowed memberships: cc_owner may SET ROLE to the NOLOGIN definer roles
			// (it hands them the definer functions, M1-03 / WH-08) without inheriting anything.
			require.NoError(t, conn.QueryRowContext(t.Context(), `
				SELECT r.rolreplication,
				       EXISTS (SELECT 1 FROM pg_auth_members m WHERE m.member = r.oid
				               AND NOT (r.rolname = 'cc_owner'
				                        AND m.roleid IN ('cc_auth_definer'::regrole, 'cc_worker_definer'::regrole)
				                        AND NOT m.inherit_option AND NOT m.admin_option))
				FROM pg_roles r WHERE r.rolname = current_user`).Scan(&replication, &memberOfAny))
			assert.False(t, replication, "REPLICATION")
			assert.False(t, memberOfAny, "member of another role")

			for _, maintenance := range []string{"postgres", "template1"} {
				var canConnect, canTemp bool
				require.NoError(t, conn.QueryRowContext(t.Context(),
					"SELECT has_database_privilege(current_user, $1, 'CONNECT'), has_database_privilege(current_user, $1, 'TEMP')",
					maintenance).Scan(&canConnect, &canTemp))
				assert.False(t, canConnect, "CONNECT on %s", maintenance)
				assert.False(t, canTemp, "TEMP on %s", maintenance)
			}
		})
	}
}

func TestNewDatabase_KeepsBootstrapOwnershipAndRevokesPublicConnect(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	var schemaOwner string
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT pg_get_userbyid(nspowner) FROM pg_namespace WHERE nspname = 'public'").Scan(&schemaOwner))
	assert.Equal(t, string(pgtest.RoleOwner), schemaOwner)

	var publicCanConnect bool
	require.NoError(t, owner.QueryRowContext(t.Context(), `
		SELECT EXISTS (
		  SELECT 1 FROM pg_database d,
		         aclexplode(coalesce(d.datacl, acldefault('d', d.datdba))) a
		  WHERE d.datname = current_database() AND a.grantee = 0 AND a.privilege_type = 'CONNECT')`).
		Scan(&publicCanConnect))
	assert.False(t, publicCanConnect, "PUBLIC must not have CONNECT on the test database")
}

func TestMigrator_RunsAsOwnerAndEveryDatabaseInheritsTheSchema(t *testing.T) {
	assert.Equal(t, string(pgtest.RoleOwner), migratedAs)

	db := pgtest.NewDatabase(t)
	var notesTable sql.NullString
	require.NoError(t, db.Connect(t, pgtest.RoleApp).
		QueryRowContext(t.Context(), "SELECT to_regclass('public.notes')::text").Scan(&notesTable))
	assert.True(t, notesTable.Valid, "the migrated table exists in a fresh database")
}

func TestNoMigrations_IsTheDefaultAndDoesNothing(t *testing.T) {
	assert.NoError(t, pgtest.NoMigrations(t.Context(), nil))
}

func TestStart_PropagatesMigratorError(t *testing.T) {
	failing := func(context.Context, *sql.DB) error { return errBoom }
	cluster, err := pgtest.Start(t.Context(), pgtest.WithMigrator(failing))
	require.ErrorIs(t, err, errBoom)
	assert.Nil(t, cluster)
}

func TestNewDatabase_GivesEachTestItsOwnDatabase(t *testing.T) {
	first, second := pgtest.NewDatabase(t), pgtest.NewDatabase(t)
	require.NotEqual(t, first.Name, second.Name)

	seed(t, first.Connect(t, pgtest.RoleApp), notes, userA)

	assert.Equal(t, 1, countAs(t, first.Connect(t, pgtest.RoleApp), userA, notes))
	assert.Equal(t, 0, countAs(t, second.Connect(t, pgtest.RoleApp), userA, notes))
}

func TestAsUser_SetsTheContextForTheTransactionOnly(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	app.SetMaxOpenConns(1) // same session for both transactions

	var inside string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT current_setting('app.current_user_id', true)").Scan(&inside)
	}))
	assert.Equal(t, userA, inside)

	var after sql.NullString
	require.NoError(t, pgtest.WithoutUser(t.Context(), app, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(),
			"SELECT nullif(current_setting('app.current_user_id', true), '')").Scan(&after)
	}))
	assert.False(t, after.Valid, "the context must not outlive its transaction, got %q", after.String)
}

func TestAsUser_RollsBackWhenTheStepFails(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	insert := insertInto(notes)

	err := pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		require.NoError(t, insert(t.Context(), tx, userA))
		return errBoom
	})

	require.ErrorIs(t, err, errBoom)
	assert.Equal(t, 0, countAs(t, app, userA, notes))
}

func TestAsUser_RejectsAnEmptyUserID(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	err := pgtest.AsUser(t.Context(), app, "", func(*sql.Tx) error { return nil })
	require.ErrorIs(t, err, pgtest.ErrNoUser)
}

func TestAssertIsolation_PassesForAnOwnerScopedTable(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table:  notes,
		UserA:  userA,
		UserB:  userB,
		Insert: insertInto(notes),
	})
}

// check runs one isolation check with user A as the viewer and user B as the owner.
type check struct {
	name string
	run  func(ctx context.Context, db *sql.DB, table pgtest.Table) error
}

var isolationChecks = []check{
	{"rls enabled and forced", pgtest.CheckRLSForced},
	{"cannot see", func(ctx context.Context, db *sql.DB, tb pgtest.Table) error {
		return pgtest.CheckCannotSee(ctx, db, tb, userA, userB)
	}},
	{"cannot update", func(ctx context.Context, db *sql.DB, tb pgtest.Table) error {
		return pgtest.CheckCannotUpdate(ctx, db, tb, userA, userB)
	}},
	{"cannot delete", func(ctx context.Context, db *sql.DB, tb pgtest.Table) error {
		return pgtest.CheckCannotDelete(ctx, db, tb, userA, userB)
	}},
	{"insert for another user rejected", func(ctx context.Context, db *sql.DB, tb pgtest.Table) error {
		insert := insertInto(tb)
		return pgtest.CheckInsertRejected(ctx, db, userA, func(ctx context.Context, tx *sql.Tx) error {
			return insert(ctx, tx, userB)
		})
	}},
	{"unset context fails closed", func(ctx context.Context, db *sql.DB, tb pgtest.Table) error {
		return pgtest.CheckUnsetContextFailsClosed(ctx, db, tb, userB)
	}},
	{"no bypass", pgtest.CheckNoBypass},
}

func TestChecks_PassOnIsolatedTablesAndCatchLeakyOnes(t *testing.T) {
	tests := []struct {
		table pgtest.Table
		want  error
	}{
		{notes, nil},
		{strictNotes, nil},
		{leakyNotes, pgtest.ErrNotIsolated},
	}
	for _, tc := range tests {
		for _, c := range isolationChecks {
			t.Run(tc.table.Name+"/"+c.name, func(t *testing.T) {
				app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
				seed(t, app, tc.table, userB)

				err := c.run(t.Context(), app, tc.table)

				if tc.want == nil {
					require.NoError(t, err)
					return
				}
				require.ErrorIs(t, err, tc.want)
			})
		}
	}
}

func TestChecks_LeaveTheOwnersRowsUntouched(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	seed(t, app, leakyNotes, userB)
	before := snapshot(t, app, leakyNotes)

	for _, c := range isolationChecks {
		_ = c.run(t.Context(), app, leakyNotes) // breaches expected; only the side effects matter here
	}

	assert.Equal(t, before, snapshot(t, app, leakyNotes), "checks must roll back their probes")
}

// snapshot returns every row of table as text, ordered by id: a delete followed by
// a re-insert changes it even though the row count stays the same.
func snapshot(t *testing.T, db *sql.DB, table pgtest.Table) string {
	t.Helper()
	var rows string
	require.NoError(t, pgtest.WithoutUser(t.Context(), db, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(),
			"SELECT coalesce(string_agg(t::text, ';' ORDER BY id), '') FROM "+pgtest.QuoteIdent(table.Name)+" t").
			Scan(&rows)
	}))
	return rows
}

func TestChecks_RefuseAVacuousProof(t *testing.T) {
	ownerScoped := []string{"cannot see", "cannot update", "cannot delete", "unset context fails closed"}
	for _, c := range isolationChecks {
		if !slices.Contains(ownerScoped, c.name) {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp) // user B owns no rows

			err := c.run(t.Context(), app, notes)

			require.ErrorIs(t, err, pgtest.ErrVacuous)
		})
	}
}

// A policy may leak only in one of the two "no context" session states: never set
// (a newly opened connection) or reverted to ” (a pooled connection reused after a
// transaction-local set_config). The check must catch both, whatever state the
// caller's pool happens to be in.
func TestCheckUnsetContextFailsClosed_CatchesLeaksInEverySessionState(t *testing.T) {
	tables := []pgtest.Table{
		{Name: "fresh_session_leaky_notes", OwnerColumn: "user_id"},
		{Name: "reused_session_leaky_notes", OwnerColumn: "user_id"},
	}
	pools := []struct {
		name string
		open func(t *testing.T, db *pgtest.Database) (seedPool, checkPool *sql.DB)
	}{
		{"pool already used", func(t *testing.T, db *pgtest.Database) (*sql.DB, *sql.DB) {
			app := db.Connect(t, pgtest.RoleApp)
			return app, app
		}},
		{"pool never used", func(t *testing.T, db *pgtest.Database) (*sql.DB, *sql.DB) {
			return db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleApp)
		}},
	}
	for _, table := range tables {
		for _, pool := range pools {
			t.Run(table.Name+"/"+pool.name, func(t *testing.T) {
				seedPool, checkPool := pool.open(t, pgtest.NewDatabase(t))
				seed(t, seedPool, table, userB)

				err := pgtest.CheckUnsetContextFailsClosed(t.Context(), checkPool, table, userB)

				require.ErrorIs(t, err, pgtest.ErrNotIsolated)
			})
		}
	}
}
