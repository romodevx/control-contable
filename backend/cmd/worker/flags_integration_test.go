//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase/pushtest"
)

const (
	// reminderPeriod is the period reminderDay's early reminder is about.
	reminderPeriod = "2026-04-01"
	// expiredPeriod is a reminder log row past the 13-month retention.
	expiredPeriod = "2025-01-01"
	// expiredAge backdates expiredPeriod's row past the purge cutoff.
	expiredAge = "13 months 1 day"
)

// M9-11, MODULE-reminders §3: the worker's production jobs as newJobs wires
// them, run once each over Postgres as cc_worker on a day the seeded user's
// reminder is due. The reminder job sends and logs only while
// module_reminders is effectively ON (its prerequisite module_taxpayers
// included); the platform purge is not behind any module flag and deletes
// the expired reminder log row in every case.
func TestWorkerJobs_remindersFollowTheFlagThePurgeAlwaysRuns(t *testing.T) {
	for _, tc := range []struct {
		name  string
		off   []string
		sends int
	}{
		{"module_reminders OFF", []string{"module_reminders"}, 0},
		{"module_taxpayers OFF implies it OFF", []string{"module_taxpayers"}, 0},
		{"every module OFF", flagtest.ModuleKeys(), 0},
		{"every module ON (control)", nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, workerDB, owner := reminderDatabases(t)
			setFlags(t, owner, tc.off)
			keys := testReminderKeys(t)
			seedRemindedUser(t, app, keys)
			seedExpiredLog(t, workerDB, owner)
			sender := &pushtest.Sender{}

			jobs, err := newJobs(jobWiring{
				pool: workerDB, logger: logging.New(&bytes.Buffer{}, nil), push: sender,
				reminderKeys: keys, outbox: testOutboxKeys(t), env: config.EnvDev, now: reminderDay,
			})
			require.NoError(t, err)

			var names []string
			for _, j := range jobs {
				names = append(names, j.job.Name())
				require.NoError(t, j.job.Run(t.Context()), j.job.Name())
			}
			assert.Equal(t, []string{"records_purge", "auth_email_delivery", "reminders", "key_reseal"}, names)
			assert.Len(t, sender.Sends(), tc.sends, "push sends")
			assert.Equal(t, tc.sends, countLog(t, workerDB, reminderPeriod), "reminder log rows written")
			assert.Zero(t, countLog(t, workerDB, expiredPeriod), "the purge deleted the expired row")
		})
	}
}

// setFlags turns every module ON, then the off ones OFF, as the owner.
func setFlags(t *testing.T, owner *sql.DB, off []string) {
	t.Helper()
	_, err := owner.ExecContext(t.Context(), "UPDATE feature_flags SET enabled = true")
	require.NoError(t, err)
	for _, key := range off {
		res, err := owner.ExecContext(t.Context(), "UPDATE feature_flags SET enabled = false WHERE key = $1", key)
		require.NoError(t, err)
		n, err := res.RowsAffected()
		require.NoError(t, err)
		require.EqualValues(t, 1, n, "%s is registered", key)
	}
}

// seedExpiredLog writes, as cc_worker for reminderUser, a reminder log row
// of expiredPeriod and backdates it past the purge cutoff as the owner
// (FORCE RLS lifted in that transaction only: the owner has no policy).
func seedExpiredLog(t *testing.T, workerDB, owner *sql.DB) {
	t.Helper()
	seedAsUser(t, workerDB, reminderUser,
		"INSERT INTO notification_log (user_id, period, kind) VALUES ($1, $2::date, 'early')", reminderUser, expiredPeriod)

	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE notification_log NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	res, err := tx.ExecContext(t.Context(),
		"UPDATE notification_log SET created_at = now() - $2::interval WHERE user_id = $1 AND period = $3::date",
		reminderUser, expiredAge, expiredPeriod)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE notification_log FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), owner, pgtest.Table{Name: "notification_log", OwnerColumn: "user_id"}))
}

// countLog counts reminderUser's log rows of period, as cc_worker.
func countLog(t *testing.T, workerDB *sql.DB, period string) int {
	t.Helper()
	tx, err := workerDB.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", reminderUser)
	require.NoError(t, err)
	var n int
	require.NoError(t, tx.QueryRowContext(t.Context(),
		"SELECT count(*) FROM notification_log WHERE period = $1::date", period).Scan(&n))
	return n
}
