//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Reminder job migrations (M9-06): cc_worker reads a user's tax inputs inside that
// user's transaction (the reminder's amounts are the Estimator's), and the platform
// purge deletes notification_log rows after 13 months (MODULE-reminders §2.6).
const (
	remindersJobReadsVersion   = 20261022100000
	notificationLogPurgeVersion = 20261022101000

	purgeTargetNotificationLog = "notification_log"
)

// The worker reads the tax inputs of the transaction's user only, and writes none.
func TestTaxInputs_WorkerReadsTheUsersOwnOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := seedTaxInputOwners(t, db)
	require.NoError(t, insertTaxInput(t, app, userA, monthlyPayment(userA, 1)))
	require.NoError(t, insertTaxInput(t, app, userB, monthlyPayment(userB, 1)))
	require.NoError(t, insertTaxInput(t, app, userB, monthlyPayment(userB, 2)))
	worker := db.Connect(t, pgtest.RoleWorker)

	for user, want := range map[string]int{userA: 1, userB: 2} {
		var n int
		var owners []string
		require.NoError(t, pgtest.AsUser(t.Context(), worker, user, func(tx *sql.Tx) error {
			if err := tx.QueryRowContext(t.Context(), "SELECT count(*) FROM taxpayer_tax_inputs").Scan(&n); err != nil {
				return err
			}
			rows, err := tx.QueryContext(t.Context(), "SELECT DISTINCT user_id::text FROM taxpayer_tax_inputs")
			if err != nil {
				return err
			}
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var owner string
				if err := rows.Scan(&owner); err != nil {
					return err
				}
				owners = append(owners, owner)
			}
			return rows.Err()
		}))
		assert.Equal(t, want, n, user)
		assert.Equal(t, []string{user}, owners)
	}
	var none int
	require.NoError(t, worker.QueryRowContext(t.Context(), "SELECT count(*) FROM taxpayer_tax_inputs").Scan(&none))
	assert.Zero(t, none, "no user context: no row")

	for _, query := range []string{
		"UPDATE taxpayer_tax_inputs SET value_cents = 1",
		"DELETE FROM taxpayer_tax_inputs",
		`INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month, value_cents)
		 VALUES ('` + userA + `', '` + taxpayerA + `', '612', 2026, 'prior_provisional_payments', 3, 1)`,
	} {
		_, err := execAsUser(t, worker, userA, query)
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	}
	_, err := execAsUser(t, db.Connect(t, pgtest.RoleWebhook), userA, "SELECT 1 FROM taxpayer_tax_inputs")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
}

func TestRemindersJobReadsMigration_Down(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), remindersJobReadsVersion-1)
	require.NoError(t, err)
	_, err = execAsUser(t, db.Connect(t, pgtest.RoleWorker), userA, "SELECT 1 FROM taxpayer_tax_inputs")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
	assert.Empty(t, queryStrings(t, owner,
		"SELECT policyname::text FROM pg_policies WHERE policyname = 'taxpayer_tax_inputs_select_worker'"))

	require.NoError(t, migrations.Up(t.Context(), owner))
	_, err = execAsUser(t, db.Connect(t, pgtest.RoleWorker), userA, "SELECT 1 FROM taxpayer_tax_inputs")
	require.NoError(t, err)
}

// backdateLogSQL moves a log row's creation into the past, as the owner (FORCE RLS
// lifted in the same transaction: the owner has no policy of its own).
const backdateLogSQL = `UPDATE notification_log SET created_at = now() - $2::interval
WHERE user_id = $1 AND period = $3::date AND kind = $4`

func backdateLog(t *testing.T, owner *sql.DB, user, period, kind, age string) {
	t.Helper()
	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE notification_log NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	res, err := tx.ExecContext(t.Context(), backdateLogSQL, user, age, period, kind)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE notification_log FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), owner, pgtest.Table{Name: "notification_log", OwnerColumn: "user_id"}))
}

// §2.6 / §8: the platform purge deletes reminder log rows created more than 13 months
// ago, of every user, and keeps the younger ones; a second pass deletes nothing.
func TestPurgeSoftDeleted_PurgesNotificationLogAfter13Months(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	const (
		periodOld   = "2025-01-01"
		periodYoung = "2025-03-01"
	)
	for _, user := range []string{userA, userB} {
		for _, period := range []string{periodOld, periodYoung} {
			_, err := execAsUser(t, worker, user, insertLogSQL, user, period, kindEarly, "pending", 1, earlyNow)
			require.NoError(t, err)
		}
		backdateLog(t, owner, user, periodOld, kindEarly, "13 months 1 day")
		backdateLog(t, owner, user, periodYoung, kindEarly, "12 months 29 days")
	}

	counts := purge(t, worker, purgeBatch)

	assert.Equal(t, 2, counts[purgeTargetNotificationLog])
	for _, user := range []string{userA, userB} {
		var periods []string
		require.NoError(t, pgtest.AsUser(t.Context(), worker, user, func(tx *sql.Tx) error {
			rows, err := tx.QueryContext(t.Context(), "SELECT period::text FROM notification_log ORDER BY period")
			if err != nil {
				return err
			}
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var p string
				if err := rows.Scan(&p); err != nil {
					return err
				}
				periods = append(periods, p)
			}
			return rows.Err()
		}))
		assert.Equal(t, []string{periodYoung}, periods, user)
	}
	assert.Zero(t, purge(t, worker, purgeBatch)[purgeTargetNotificationLog])
}

// A small batch bounds each pass; the worker's loop drains the rest.
func TestPurgeSoftDeleted_NotificationLogBatchIsBounded(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	periods := []string{"2024-01-01", "2024-02-01", "2024-03-01"}
	for _, period := range periods {
		_, err := execAsUser(t, worker, userA, insertLogSQL, userA, period, kindDue, "pending", 1, earlyNow)
		require.NoError(t, err)
		backdateLog(t, owner, userA, period, kindDue, "2 years")
	}

	assert.Equal(t, 2, purge(t, worker, 2)[purgeTargetNotificationLog])
	assert.Equal(t, 1, purge(t, worker, 2)[purgeTargetNotificationLog])
}

// The definer's new reach on notification_log is SELECT + DELETE past the cutoff, and
// the runtime roles still delete no log row directly.
func TestNotificationLogPurge_Reach(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	assert.Equal(t, []string{"notification_log_purge_delete DELETE", "notification_log_purge_select SELECT"},
		queryStrings(t, owner, `
SELECT policyname || ' ' || cmd FROM pg_policies
WHERE tablename = 'notification_log' AND $1 = ANY (roles) AND policyname LIKE '%\_purge\_%' ORDER BY 1`, workerDefinerRole))
	assert.Equal(t, []string{"notification_log DELETE"}, queryStrings(t, owner, `
SELECT c.relname || ' ' || a.privilege_type
FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.relname = 'notification_log' AND a.grantee = $1::regrole ORDER BY 1`, workerDefinerRole),
		"table-level: DELETE only (SELECT is per column)")
	var cutoffDefiner bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT prosecdef FROM pg_proc WHERE proname = 'purge_notification_log_cutoff'").Scan(&cutoffDefiner))
	assert.False(t, cutoffDefiner, "the cutoff helper is SECURITY INVOKER")
	assert.Equal(t, []string{workerDefinerRole}, queryStrings(t, owner, `
SELECT r.rolname FROM pg_roles r
WHERE r.rolname LIKE 'cc\_%' AND has_function_privilege(r.oid, 'purge_notification_log_cutoff()', 'EXECUTE') ORDER BY 1`))

	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWorker} {
		_, err := db.Connect(t, role).ExecContext(t.Context(), "DELETE FROM notification_log")
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	}
}

// Down restores purge_soft_deleted as M9-08 shipped it and removes the definer's log
// reach (keeping the due_reminder_users read); Up restores the purge.
func TestNotificationLogPurgeMigration_Down(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), notificationLogPurgeVersion-1)
	require.NoError(t, err)
	worker := db.Connect(t, pgtest.RoleWorker)
	assert.Equal(t, purgeTargets[:len(purgeTargets)-1], queryStrings(t, worker, "SELECT target FROM purge_soft_deleted(1)"))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT policyname::text FROM pg_policies WHERE policyname LIKE 'notification_log\\_purge\\_%'"))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT proname::text FROM pg_proc WHERE proname = 'purge_notification_log_cutoff'"))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT indexname::text FROM pg_indexes WHERE indexname = 'notification_log_created_at_idx'"))
	assert.Empty(t, queryStrings(t, owner, `
SELECT c.relname FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.relname = 'notification_log' AND a.grantee = $1::regrole`, workerDefinerRole))
	assert.Empty(t, dueUsers(t, worker, earlyNow, false, nil, dueLimit), "due_reminder_users still reads the log")

	require.NoError(t, migrations.Up(t.Context(), owner))
	assert.Equal(t, purgeTargets, queryStrings(t, worker, "SELECT target FROM purge_soft_deleted(1)"))
}
