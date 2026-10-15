//go:build integration

package migrations_test

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Records purge migration (M9-08): purge_soft_deleted, the SECURITY DEFINER function
// owned by cc_worker_definer that the worker runs to delete soft-deleted records 30 days
// after their deletion, and the dead auth leftovers (refresh tokens, pending sign-ups,
// verification failure windows) one day after they stopped mattering.

const (
	recordsPurgeVersion = 20261020120000

	workerDefinerRole = "cc_worker_definer"
	callPurge         = "SELECT target, purged FROM purge_soft_deleted($1)"

	// purgeBatch is a batch size larger than any fixture, so one call drains everything.
	purgeBatch = 100
	// maxPurgeBatch mirrors the upper bound inside purge_soft_deleted.
	maxPurgeBatch = 5000

	purgeTargetIncomes            = "incomes"
	purgeTargetExpenses           = "expenses"
	purgeTargetPersonalDeductions = "personal_deductions"
	purgeTargetRefreshTokens      = "refresh_tokens"
	purgeTargetVerifications      = "email_verifications"
	purgeTargetFailures           = "email_verification_failures"
)

// purgeTargets are every target purge_soft_deleted reports, in its order (notification_log
// since M9-06, reminders_job_integration_test.go; auth_email_outbox since the email
// outbox, auth_email_outbox_integration_test.go).
var purgeTargets = []string{
	purgeTargetIncomes, purgeTargetExpenses, purgeTargetPersonalDeductions,
	purgeTargetRefreshTokens, purgeTargetVerifications, purgeTargetFailures, purgeTargetNotificationLog,
	purgeTargetEmailOutbox,
}

// softDeletable is one user table with a deleted_at column and how a row of it is written.
type softDeletable struct {
	name   string
	insert func(t *testing.T, app *sql.DB, userID, id string)
}

var softDeletables = []softDeletable{
	{purgeTargetIncomes, func(t *testing.T, app *sql.DB, userID, id string) {
		t.Helper()
		require.NoError(t, insertIncome(t, app, userID, validIncome(userID, taxpayerOf[userID], id)))
	}},
	{purgeTargetExpenses, func(t *testing.T, app *sql.DB, userID, id string) {
		t.Helper()
		seedExpense(t, app, userID, expenseColumns{"id": quote(id)})
	}},
	{purgeTargetPersonalDeductions, func(t *testing.T, app *sql.DB, userID, id string) {
		t.Helper()
		require.NoError(t, personalDeductions.insert(t, app, userID, personalDeductions.row(userID, expenseColumns{"id": quote(id)})))
	}},
}

// recordID is a distinct uuid per (table, user, slot), so survivors can be named.
func recordID(table int, user string, slot int) string {
	return fmt.Sprintf("00000000-0000-7000-8%03d-%s%02d", table, user[len(user)-10:], slot)
}

// softDeleteAgo soft-deletes row id of userID's table as cc_app does, dated age ago.
func softDeleteAgo(t *testing.T, app *sql.DB, table, userID, id, age string) {
	t.Helper()
	affected, err := execAsUser(t, app, userID,
		"UPDATE "+table+" SET deleted_at = now() - $2::interval WHERE id = $1", id, age)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)
}

// purge calls purge_soft_deleted as cc_worker (no user context) and returns its counts.
func purge(t *testing.T, worker *sql.DB, batch any) map[string]int {
	t.Helper()
	rows, err := worker.QueryContext(t.Context(), callPurge, batch)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	counts := map[string]int{}
	var order []string
	for rows.Next() {
		var target string
		var n int
		require.NoError(t, rows.Scan(&target, &n))
		counts[target] = n
		order = append(order, target)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, purgeTargets, order, "one row per target, in a stable order")
	return counts
}

// idsOf lists userID's ids of table, as cc_app sees them (deleted rows included).
func idsOf(t *testing.T, app *sql.DB, table, userID string) []string {
	t.Helper()
	var ids []string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(t.Context(), fmt.Sprintf("SELECT id::text FROM %s ORDER BY id", table))
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rows.Err()
	}))
	return ids
}

// asOwnerWithoutForce runs query as cc_owner with the table's FORCE RLS lifted for one
// transaction (fixtures only: backdating columns no runtime role may write).
func asOwnerWithoutForce(t *testing.T, owner *sql.DB, table, query string, args ...any) {
	t.Helper()
	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE "+table+" NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), query, args...)
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE "+table+" FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

// The function holds the cross-user DELETE, so it is definer-owned, pins its search_path
// and is executable by cc_worker alone; the definer role itself stays unprivileged.
func TestPurgeSoftDeleted_DefinerOwnedPinnedAndWorkerOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	var definer bool
	var fnOwner, config string
	require.NoError(t, owner.QueryRowContext(t.Context(), `
SELECT p.prosecdef, pg_get_userbyid(p.proowner), array_to_string(p.proconfig, ',')
FROM pg_proc p WHERE p.oid = 'purge_soft_deleted(integer)'::regprocedure`).Scan(&definer, &fnOwner, &config))
	assert.True(t, definer, "SECURITY DEFINER")
	assert.Equal(t, workerDefinerRole, fnOwner)
	assert.Equal(t, "search_path=pg_catalog, public, pg_temp", config)

	executors := queryStrings(t, owner, `
SELECT r.rolname FROM pg_roles r
WHERE r.rolname LIKE 'cc\_%' AND r.rolname <> $1
  AND has_function_privilege(r.oid, 'purge_soft_deleted(integer)', 'EXECUTE') ORDER BY 1`, workerDefinerRole)
	assert.Equal(t, []string{"cc_worker"}, executors)

	var login, bypass bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT rolcanlogin, rolbypassrls FROM pg_roles WHERE rolname = $1", workerDefinerRole).Scan(&login, &bypass))
	assert.False(t, login, "NOLOGIN")
	assert.False(t, bypass, "NOBYPASSRLS")
	var schemaCreate bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT has_schema_privilege($1, 'public', 'CREATE')", workerDefinerRole).Scan(&schemaCreate))
	assert.False(t, schemaCreate, "CREATE on public is lent for the ownership transfer only")

	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook} {
		_, err := db.Connect(t, role).ExecContext(t.Context(), callPurge, purgeBatch)
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	}
}

// The definer's reach is exactly SELECT + DELETE on the purged tables, through policies
// TO cc_worker_definer only; the worker and the app still cannot delete a record directly.
// (Its read-only due_reminder_users policies, M9-02, are pinned by the reminders suite.)
func TestPurgeSoftDeleted_NoReachBeyondThePurge(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	policies := queryStrings(t, owner, `
SELECT tablename || ' ' || cmd FROM pg_policies
WHERE $1 = ANY (roles) AND tablename <> 'users' AND policyname LIKE '%\_purge\_%' ORDER BY 1`, workerDefinerRole)
	var want []string
	for _, target := range []string{
		purgeTargetEmailOutbox, purgeTargetFailures, purgeTargetVerifications, purgeTargetExpenses, purgeTargetIncomes,
		purgeTargetNotificationLog, purgeTargetPersonalDeductions, purgeTargetRefreshTokens,
	} {
		want = append(want, target+" DELETE", target+" SELECT")
	}
	assert.Equal(t, want, policies)

	writes := queryStrings(t, owner, `
SELECT c.relname || ' ' || a.privilege_type
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE n.nspname = 'public' AND a.grantee = $1::regrole AND a.privilege_type NOT IN ('SELECT', 'DELETE')`,
		workerDefinerRole)
	assert.Empty(t, writes, "the definer role inserts, updates and truncates nothing")

	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWorker} {
		conn := db.Connect(t, role)
		for _, target := range []string{purgeTargetIncomes, purgeTargetExpenses, purgeTargetPersonalDeductions} {
			_, err := conn.ExecContext(t.Context(), fmt.Sprintf("DELETE FROM %s", target))
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		}
	}
}

// Soft-deleted records go 30 days after deletion, for every user; live rows and rows
// deleted more recently stay, and cc_app's own-rows isolation is unchanged afterwards.
func TestPurgeSoftDeleted_PurgesOnlyRecordsDeletedOver30DaysAgo(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	worker := db.Connect(t, pgtest.RoleWorker)

	const (
		slotLive = iota
		slotRecent
		slotOld
	)
	survivors := map[string]map[string][]string{}
	for i, tbl := range softDeletables {
		survivors[tbl.name] = map[string][]string{}
		for _, user := range []string{userA, userB} {
			for _, slot := range []int{slotLive, slotRecent, slotOld} {
				tbl.insert(t, app, user, recordID(i, user, slot))
			}
			softDeleteAgo(t, app, tbl.name, user, recordID(i, user, slotRecent), "29 days")
			softDeleteAgo(t, app, tbl.name, user, recordID(i, user, slotOld), "31 days")
			survivors[tbl.name][user] = []string{recordID(i, user, slotLive), recordID(i, user, slotRecent)}
		}
	}

	counts := purge(t, worker, purgeBatch)
	for _, tbl := range softDeletables {
		assert.Equal(t, 2, counts[tbl.name], "%s: one 31-day-old row of each user", tbl.name)
		for _, user := range []string{userA, userB} {
			assert.Equal(t, survivors[tbl.name][user], idsOf(t, app, tbl.name, user), "%s of %s", tbl.name, user)
		}
	}

	t.Run("app isolation still holds", func(t *testing.T) {
		pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
			Table: incomesTable, UserA: userA, UserB: userB, Insert: insertOwnIncome,
		})
	})
	t.Run("the worker still sees no record without a user", func(t *testing.T) {
		for _, tbl := range []string{purgeTargetIncomes, purgeTargetExpenses} {
			var n int
			require.NoError(t, pgtest.WithoutUser(t.Context(), worker, func(tx *sql.Tx) error {
				return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+tbl).Scan(&n)
			}))
			assert.Zero(t, n, tbl)
		}
	})
}

// Dead sessions and finished sign-up state go one day after they stopped mattering:
// refresh tokens past their family's absolute expiry or revoked, pending sign-ups past
// their expiry, failure windows that ended. Live and freshly dead rows stay.
func TestPurgeSoftDeleted_PurgesDeadAuthState(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	owner := db.Connect(t, pgtest.RoleOwner)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA, userB)

	const (
		live         = "00000000-0000-4000-8000-0000000001a1"
		rotated      = "00000000-0000-4000-8000-0000000001a2"
		successor    = "00000000-0000-4000-8000-0000000001a3"
		revokedNow   = "00000000-0000-4000-8000-0000000001a4"
		revokedOld   = "00000000-0000-4000-8000-0000000001a5"
		expiredOld   = "00000000-0000-4000-8000-0000000001b1"
		liveFamily   = "00000000-0000-4000-8000-0000000002a1"
		familyRecent = "00000000-0000-4000-8000-0000000002a4"
		familyOld    = "00000000-0000-4000-8000-0000000002a5"
		familyB      = "00000000-0000-4000-8000-0000000002b1"
	)
	issue := func(user, id, family string, parent any) {
		_, err := execAsUser(t, app, user, insertSessionSQL, user, "secret-"+id, family, id, parent)
		require.NoError(t, err)
	}
	issue(userA, live, liveFamily, nil)
	issue(userA, revokedNow, familyRecent, nil)
	issue(userA, revokedOld, familyOld, nil)
	issue(userB, expiredOld, familyB, nil)
	issue(userA, rotated, liveFamily, nil)
	_, err := execAsUser(t, app, userA, "UPDATE refresh_tokens SET rotated_at = now() WHERE id = $1", rotated)
	require.NoError(t, err)
	issue(userA, successor, liveFamily, rotated)
	_, err = execAsUser(t, app, userA,
		"UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'logout' WHERE id = $1", revokedNow)
	require.NoError(t, err)
	// Two days in the past: issued, (revoked,) and for B's family past the absolute expiry.
	asOwnerWithoutForce(t, owner, purgeTargetRefreshTokens, `
UPDATE refresh_tokens SET created_at = now() - interval '2200 hours', expires_at = now() - interval '1500 hours',
                          absolute_expires_at = now() - interval '48 hours'
WHERE id = $1`, expiredOld)
	asOwnerWithoutForce(t, owner, purgeTargetRefreshTokens, `
UPDATE refresh_tokens SET created_at = now() - interval '72 hours', expires_at = now() + interval '600 hours',
                          absolute_expires_at = now() + interval '2000 hours',
                          revoked_at = now() - interval '48 hours', revoke_reason = 'logout'
WHERE id = $1`, revokedOld)

	// email_verifications and email_verification_failures are ENABLE (not FORCE): the owner
	// writes them directly.
	for _, q := range []string{
		`INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at, created_at)
		 VALUES (sha256('live'), sha256('1'), '{}', now() + interval '15 minutes', now())`,
		`INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at, created_at)
		 VALUES (sha256('expired-now'), sha256('1'), '{}', now() - interval '1 hour', now() - interval '2 hours')`,
		`INSERT INTO email_verifications (email_hash, code_hash, pending, expires_at, created_at)
		 VALUES (sha256('expired-old'), sha256('1'), '{}', now() - interval '25 hours', now() - interval '26 hours')`,
		`INSERT INTO email_verification_failures (email_hash, failures, window_started_at)
		 VALUES (sha256('open-window'), 3, now() - interval '23 hours')`,
		`INSERT INTO email_verification_failures (email_hash, failures, window_started_at)
		 VALUES (sha256('closed-window'), 3, now() - interval '25 hours')`,
	} {
		_, err := owner.ExecContext(t.Context(), q)
		require.NoError(t, err)
	}

	counts := purge(t, worker, purgeBatch)
	assert.Equal(t, 2, counts[purgeTargetRefreshTokens], "A's family revoked 2 days ago and B's expired one")
	assert.Equal(t, 1, counts[purgeTargetVerifications])
	assert.Equal(t, 1, counts[purgeTargetFailures])

	assert.Equal(t, []string{live, rotated, successor, revokedNow}, idsOf(t, app, purgeTargetRefreshTokens, userA))
	assert.Empty(t, idsOf(t, app, purgeTargetRefreshTokens, userB))
	assert.Equal(t, []string{"expired-now", "live"}, queryStrings(t, owner, `
SELECT CASE email_hash WHEN sha256('live') THEN 'live' WHEN sha256('expired-now') THEN 'expired-now' ELSE 'other' END
FROM email_verifications ORDER BY 1`))
	var openWindows int
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT count(*) FROM email_verification_failures WHERE email_hash = sha256('open-window')").Scan(&openWindows))
	assert.Equal(t, 1, openWindows)
}

// Purging a parent whose successor stays (the successor is not yet past the cutoff)
// runs the self FK's ON DELETE SET NULL (rotated_from) as the table owner: the call
// succeeds, the successor survives with only its link cleared, and its revocation is
// untouched (refresh_tokens_history_is_final).
func TestPurgeSoftDeleted_ParentPurgeKeepsSuccessorAndItsHistory(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	owner := db.Connect(t, pgtest.RoleOwner)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA)

	const (
		parent    = "00000000-0000-4000-8000-0000000003a1"
		successor = "00000000-0000-4000-8000-0000000003a2"
		family    = "00000000-0000-4000-8000-0000000004a1"
	)
	_, err := execAsUser(t, app, userA, insertSessionSQL, userA, "secret-"+parent, family, parent, nil)
	require.NoError(t, err)
	_, err = execAsUser(t, app, userA, "UPDATE refresh_tokens SET rotated_at = now() WHERE id = $1", parent)
	require.NoError(t, err)
	_, err = execAsUser(t, app, userA, insertSessionSQL, userA, "secret-"+successor, family, successor, parent)
	require.NoError(t, err)
	_, err = execAsUser(t, app, userA,
		"UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'reuse_detected' WHERE id = $1", successor)
	require.NoError(t, err)
	asOwnerWithoutForce(t, owner, purgeTargetRefreshTokens, `
UPDATE refresh_tokens SET created_at = now() - interval '72 hours', expires_at = now() + interval '600 hours',
                          absolute_expires_at = now() + interval '2000 hours',
                          revoked_at = now() - interval '48 hours', revoke_reason = 'reuse_detected'
WHERE id = $1`, parent)

	counts := purge(t, worker, purgeBatch)
	assert.Equal(t, 1, counts[purgeTargetRefreshTokens], "only the parent is past the cutoff")
	assert.Equal(t, []string{successor}, idsOf(t, app, purgeTargetRefreshTokens, userA))

	var linked bool
	var reason string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(),
			"SELECT rotated_from IS NOT NULL, revoke_reason FROM refresh_tokens WHERE id = $1", successor).
			Scan(&linked, &reason)
	}))
	assert.False(t, linked, "the link to the purged parent is cleared")
	assert.Equal(t, "reuse_detected", reason, "the successor's revocation is kept")
}

// Bounded batches: a call deletes at most p_batch_size rows per target, so the worker
// loops; once drained, a re-run deletes nothing (idempotent).
func TestPurgeSoftDeleted_BoundedBatchesAndIdempotentRerun(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	worker := db.Connect(t, pgtest.RoleWorker)
	const old = 3
	for slot := range old {
		id := recordID(0, userA, slot)
		softDeletables[0].insert(t, app, userA, id)
		softDeleteAgo(t, app, purgeTargetIncomes, userA, id, "40 days")
	}

	assert.Equal(t, 2, purge(t, worker, 2)[purgeTargetIncomes])
	assert.Equal(t, 1, purge(t, worker, 2)[purgeTargetIncomes])
	for _, n := range purge(t, worker, 2) {
		assert.Zero(t, n)
	}
	assert.Empty(t, idsOf(t, app, purgeTargetIncomes, userA))
}

func TestPurgeSoftDeleted_RejectsBadBatchSizes(t *testing.T) {
	worker := pgtest.NewDatabase(t).Connect(t, pgtest.RoleWorker)
	for _, batch := range []any{nil, 0, -1, maxPurgeBatch + 1} {
		_, err := worker.ExecContext(t.Context(), callPurge, batch)
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	}
	_, err := worker.ExecContext(t.Context(), callPurge, maxPurgeBatch)
	require.NoError(t, err)
}

// Split brain: two worker leaders may purge at once for a few seconds. Each row is
// deleted once; the later call waits on the earlier one's row locks and then skips them.
func TestPurgeSoftDeleted_ConcurrentRunsDeleteEachRowOnce(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedIncomeOwners(t, db)
	const rows = 5
	for slot := range rows {
		id := recordID(0, userB, slot)
		softDeletables[0].insert(t, app, userB, id)
		softDeleteAgo(t, app, purgeTargetIncomes, userB, id, "31 days")
	}
	first := db.Connect(t, pgtest.RoleWorker)
	second := db.Connect(t, pgtest.RoleWorker)

	tx, err := first.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	var firstCount int
	require.NoError(t, tx.QueryRowContext(t.Context(),
		"SELECT purged FROM purge_soft_deleted($1) WHERE target = $2", purgeBatch, purgeTargetIncomes).Scan(&firstCount))

	var wg sync.WaitGroup
	var secondCount int
	var secondErr error
	wg.Go(func() {
		secondErr = second.QueryRowContext(t.Context(),
			"SELECT purged FROM purge_soft_deleted($1) WHERE target = $2", purgeBatch, purgeTargetIncomes).Scan(&secondCount)
	})
	require.NoError(t, tx.Commit())
	wg.Wait()

	require.NoError(t, secondErr)
	assert.Equal(t, rows, firstCount)
	assert.Zero(t, secondCount)
}

// Down removes the function, its cutoff helpers, the definer's policies and grants and
// the new indexes; Up restores them.
func TestRecordsPurgeMigration_DownRemovesEverything(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), recordsPurgeVersion-1)
	require.NoError(t, err)
	assert.Empty(t, queryStrings(t, owner,
		"SELECT proname::text FROM pg_proc WHERE proname IN ('purge_soft_deleted', 'purge_records_cutoff', 'purge_auth_cutoff')"))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT tablename || ' ' || policyname FROM pg_policies WHERE $1 = ANY (roles) AND tablename <> 'users'", workerDefinerRole))
	assert.Empty(t, queryStrings(t, owner, `
SELECT c.relname FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE a.grantee = $1::regrole AND c.relname <> 'users'`, workerDefinerRole))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT indexname::text FROM pg_indexes WHERE indexname IN ('refresh_tokens_absolute_expires_at_idx', 'refresh_tokens_revoked_at_idx')"))
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), owner, incomesTable))

	require.NoError(t, migrations.Up(t.Context(), owner))
	worker := db.Connect(t, pgtest.RoleWorker)
	assert.Len(t, purge(t, worker, purgeBatch), len(purgeTargets))
}

// The purge's scans use indexes: the deleted_at partial indexes of the record tables and
// the two refresh_tokens indexes this migration adds. Planned as the definer role, so the
// plan includes its policies' quals (cc_owner itself is denied every row by FORCE RLS).
func TestPurgeSoftDeleted_ScansUseIndexes(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	scans := []struct{ index, query string }{
		{"incomes_deleted_at_idx", "SELECT id FROM incomes WHERE deleted_at < now() - interval '30 days' ORDER BY deleted_at LIMIT 100"},
		{"expenses_deleted_at_idx", "SELECT id FROM expenses WHERE deleted_at < now() - interval '30 days' ORDER BY deleted_at LIMIT 100"},
		{"personal_deductions_deleted_at_idx", "SELECT id FROM personal_deductions WHERE deleted_at < now() - interval '30 days' ORDER BY deleted_at LIMIT 100"},
		{"refresh_tokens_absolute_expires_at_idx", "SELECT id FROM refresh_tokens WHERE absolute_expires_at < now() - interval '1 day' ORDER BY absolute_expires_at LIMIT 100"},
		{"refresh_tokens_revoked_at_idx", "SELECT id FROM refresh_tokens WHERE revoked_at < now() - interval '1 day' ORDER BY revoked_at LIMIT 100"},
		{"email_verifications_expires_at_idx", "SELECT email_hash FROM email_verifications WHERE expires_at < now() - interval '1 day' ORDER BY expires_at LIMIT 100"},
	}
	for _, scan := range scans {
		t.Run(scan.index, func(t *testing.T) {
			tx, err := owner.BeginTx(t.Context(), nil)
			require.NoError(t, err)
			defer func() { _ = tx.Rollback() }()
			_, err = tx.ExecContext(t.Context(), "SET LOCAL ROLE "+workerDefinerRole)
			require.NoError(t, err)
			_, err = tx.ExecContext(t.Context(), "SET LOCAL enable_seqscan = off")
			require.NoError(t, err)
			var plan string
			rows, err := tx.QueryContext(t.Context(), fmt.Sprintf("EXPLAIN %s", scan.query))
			require.NoError(t, err)
			for rows.Next() {
				var line string
				require.NoError(t, rows.Scan(&line))
				plan += line + "\n"
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			assert.Contains(t, plan, scan.index)
		})
	}
}
