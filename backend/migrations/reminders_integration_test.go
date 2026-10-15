//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Reminders migration (M9-02): devices, reminder_settings and notification_log
// (MODULE-reminders §2.3, §2.4, §2.6), the claim_push_token definer function the API
// calls to register (and move) a push token, and due_reminder_users, the worker's
// cross-user selection of who gets a reminder now.

const (
	remindersVersion = 20261021101000

	authDefinerRole = "cc_auth_definer"

	// deviceCap is the fair-use cap of §2.3: the 11th device evicts the least recently seen.
	deviceCap = 10

	// claimSQL registers installation $2 for user $1 through claim_push_token. The token
	// columns derive from the label $3 like crypto output would: a framed ciphertext
	// (key_id 1 || nonce || ct || tag) and a 32-byte HMAC. No real token is ever written.
	claimSQL = `
SELECT claim_push_token($1::uuid, $2::uuid, 'android',
       '\x0001'::bytea || sha256(convert_to($3::text, 'UTF8')) || sha256(convert_to($3::text, 'UTF8')),
       1::smallint, sha256(convert_to($3::text || '/hmac', 'UTF8')), '1.2.0')`

	insertSettingsSQL = `
INSERT INTO reminder_settings (user_id, enabled, early_day, on_due_date, send_time, show_amounts)
VALUES ($1, $2, $3, $4, $5::time, false)`

	insertLogSQL = `
INSERT INTO notification_log (user_id, period, kind, channel, status, attempts, claimed_at)
VALUES ($1, $2::date, $3, 'push', $4, $5, $6::timestamptz)`

	claimLogSQL = `
INSERT INTO notification_log (user_id, period, kind, channel, status, attempts, claimed_at)
VALUES ($1, DATE '2026-04-01', 'early', 'push', 'pending', 1, now())
ON CONFLICT (user_id, period, kind, channel) DO NOTHING RETURNING id`

	dueUsersSQL = "SELECT user_id::text, period::text, kind FROM due_reminder_users($1::timestamptz, $2, $3::uuid, $4)"

	// 14 May 2026 10:05 and 18 May 2026 10:05 in America/Mexico_City (UTC-6, no DST):
	// the early reminder (day 14) and the moved due date of period 2026-04 (US-R2).
	earlyNow   = "2026-05-14 16:05:00+00"
	periodApr  = "2026-04-01"
	kindEarly  = "early"
	kindDue    = "due"
	dueLimit   = 100
	maxDueRows = 1000
)

var (
	devicesTable          = pgtest.Table{Name: "devices", OwnerColumn: "user_id"}
	reminderSettingsTable = pgtest.Table{Name: "reminder_settings", OwnerColumn: "user_id"}
	notificationLogTable  = pgtest.Table{Name: "notification_log", OwnerColumn: "user_id"}
)

// installationOf is one installation id per user for the isolation suites.
var installationOf = map[string]string{
	userA: "0199d4a4-3c4d-7f5e-8a61-0000000000a1",
	userB: "0199d4a4-3c4d-7f5e-8a61-0000000000b1",
}

// installation is a distinct installation id per (user suffix, slot).
func installation(user string, slot int) string {
	return fmt.Sprintf("0199d4a4-3c4d-7f5e-8a61-%s%02d", user[len(user)-10:], slot)
}

func claimOwnDevice(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, claimSQL, ownerID, installationOf[ownerID], "token-of-"+ownerID)
	return err
}

func insertOwnSettings(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertSettingsSQL, ownerID, true, 14, true, "10:00")
	return err
}

func insertOwnLog(ctx context.Context, tx *sql.Tx, ownerID string) error {
	_, err := tx.ExecContext(ctx, insertLogSQL, ownerID, periodApr, kindEarly, "pending", 1, earlyNow)
	return err
}

// claim registers installation for user (token derived from label) and returns created.
func claim(t *testing.T, app *sql.DB, user, installationID, label string) bool {
	t.Helper()
	var created bool
	require.NoError(t, pgtest.AsUser(t.Context(), app, user, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), claimSQL, user, installationID, label).Scan(&created)
	}))
	return created
}

// devicesOf lists the installation ids user sees, ordered.
func devicesOf(t *testing.T, db *sql.DB, user string) []string {
	t.Helper()
	return idsOf(t, db, devicesTable.Name, user)
}

// dueUsers calls due_reminder_users as cc_worker (no user context).
func dueUsers(t *testing.T, worker *sql.DB, now string, dueToday bool, after any, limit int) []string {
	t.Helper()
	rows, err := worker.QueryContext(t.Context(), dueUsersSQL, now, dueToday, after, limit)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var user, period, kind string
		require.NoError(t, rows.Scan(&user, &period, &kind))
		out = append(out, user+" "+period+" "+kind)
	}
	require.NoError(t, rows.Err())
	return out
}

// definerFunction reads whether fn is SECURITY DEFINER, its owner and its settings.
func definerFunction(t *testing.T, owner *sql.DB, fn string) (definer bool, fnOwner, config string) {
	t.Helper()
	require.NoError(t, owner.QueryRowContext(t.Context(), `
SELECT p.prosecdef, pg_get_userbyid(p.proowner), array_to_string(p.proconfig, ',')
FROM pg_proc p WHERE p.oid = $1::regprocedure`, fn).Scan(&definer, &fnOwner, &config))
	return definer, fnOwner, config
}

// executors lists the cc_* roles (but the function's owner) that may EXECUTE fn.
func executors(t *testing.T, owner *sql.DB, fn, fnOwner string) []string {
	t.Helper()
	return queryStrings(t, owner, `
SELECT r.rolname FROM pg_roles r
WHERE r.rolname LIKE 'cc\_%' AND r.rolname <> $2
  AND has_function_privilege(r.oid, $1::regprocedure, 'EXECUTE') ORDER BY 1`, fn, fnOwner)
}

// The three tables are user data: ENABLE + FORCE RLS, and each runtime role that
// writes one is confined to its own rows (cc_app: devices through the claim, settings;
// cc_worker: the log).
func TestReminderTables_RLSIsolation(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA, userB)

	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: devicesTable, UserA: userA, UserB: userB, Insert: claimOwnDevice,
	})
	pgtest.AssertIsolation(t, app, pgtest.IsolationSpec{
		Table: reminderSettingsTable, UserA: userA, UserB: userB, Insert: insertOwnSettings,
	})
	pgtest.AssertIsolation(t, worker, pgtest.IsolationSpec{
		Table: notificationLogTable, UserA: userA, UserB: userB, Insert: insertOwnLog,
	})

	t.Run("the worker reads devices and settings of its user only", func(t *testing.T) {
		for _, table := range []pgtest.Table{devicesTable, reminderSettingsTable} {
			require.NoError(t, pgtest.CheckCannotSee(t.Context(), worker, table, userA, userB), table.Name)
			require.NoError(t, pgtest.CheckUnsetContextFailsClosed(t.Context(), worker, table, userB), table.Name)
		}
	})
	t.Run("the app never reads the log", func(t *testing.T) {
		_, err := execAsUser(t, app, userA, "SELECT 1 FROM notification_log")
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
}

// Write isolation proven on granted operations: a cross-user write affects no row
// (the isolation suite's UPDATE of the owner column is refused by a missing grant).
func TestReminderTables_CrossUserWritesAffectNoRow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA, userB)
	for _, user := range []string{userA, userB} {
		require.NoError(t, pgtest.AsUser(t.Context(), app, user, func(tx *sql.Tx) error {
			return claimOwnDevice(t.Context(), tx, user)
		}))
		require.NoError(t, pgtest.AsUser(t.Context(), app, user, func(tx *sql.Tx) error {
			return insertOwnSettings(t.Context(), tx, user)
		}))
		require.NoError(t, pgtest.AsUser(t.Context(), worker, user, func(tx *sql.Tx) error {
			return insertOwnLog(t.Context(), tx, user)
		}))
	}

	writes := []struct {
		name  string
		db    *sql.DB
		query string
	}{
		{"app updates B's settings", app, "UPDATE reminder_settings SET show_amounts = true WHERE user_id = $1"},
		{"app deletes B's device", app, "DELETE FROM devices WHERE user_id = $1"},
		{"worker deletes B's device", worker, "DELETE FROM devices WHERE user_id = $1"},
		{"worker marks B's log sent", worker,
			"UPDATE notification_log SET status = 'sent', sent_at = now(), devices_ok = 1 WHERE user_id = $1"},
	}
	for _, w := range writes {
		t.Run(w.name, func(t *testing.T) {
			affected, err := execAsUser(t, w.db, userA, w.query, userB)
			require.NoError(t, err)
			assert.Zero(t, affected)
		})
	}
	assert.Equal(t, []string{installationOf[userB]}, devicesOf(t, app, userB), "B's device is intact")
}

// Devices are written only through claim_push_token; the API reads them without the
// token and deletes its own; nobody but the worker (to send) reads the ciphertext.
func TestDevices_TokenAndWritesAreNarrow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA)
	claim(t, app, userA, installationOf[userA], "token-a")

	for _, q := range []string{
		`INSERT INTO devices (id, user_id, platform, token_ciphertext, token_key_id, token_hmac)
		 VALUES (gen_random_uuid(), app_user_id(), 'ios', '\x00'::bytea, 1, sha256('x'))`,
		"UPDATE devices SET platform = 'ios'",
		"SELECT token_ciphertext FROM devices",
		"SELECT token_hmac FROM devices",
	} {
		_, err := execAsUser(t, app, userA, q)
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	}
	var sendable int
	require.NoError(t, pgtest.AsUser(t.Context(), worker, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(),
			"SELECT count(*) FROM devices WHERE octet_length(token_ciphertext) > 30 AND token_key_id = 1").Scan(&sendable)
	}))
	assert.Equal(t, 1, sendable, "the worker reads its user's ciphertext to send")
	_, err := execAsUser(t, worker, userA, "UPDATE devices SET platform = 'ios'")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)

	affected, err := execAsUser(t, app, userA, "DELETE FROM devices WHERE id = $1", installationOf[userA])
	require.NoError(t, err)
	assert.EqualValues(t, 1, affected, "unregisterDevice deletes the caller's own row")
}

// Both functions are definer-owned by their NOLOGIN role, pin search_path and are
// executable by the one runtime role that needs them.
func TestReminderFunctions_DefinerOwnedPinnedAndNarrow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	const (
		claimFn = "claim_push_token(uuid,uuid,text,bytea,smallint,bytea,text)"
		dueFn   = "due_reminder_users(timestamptz,boolean,uuid,integer)"
	)
	cases := []struct {
		fn, owner, executor string
	}{
		{claimFn, authDefinerRole, "cc_app"},
		{dueFn, workerDefinerRole, "cc_worker"},
	}
	for _, c := range cases {
		t.Run(c.fn, func(t *testing.T) {
			definer, fnOwner, config := definerFunction(t, owner, c.fn)
			assert.True(t, definer, "SECURITY DEFINER")
			assert.Equal(t, c.owner, fnOwner)
			assert.Equal(t, "search_path=pg_catalog, public, pg_temp", config)
			assert.Equal(t, []string{c.executor}, executors(t, owner, c.fn, c.owner))
		})
	}

	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	_, err := app.ExecContext(t.Context(), dueUsersSQL, earlyNow, false, nil, dueLimit)
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
	_, err = execAsUser(t, db.Connect(t, pgtest.RoleWorker), userA, claimSQL, userA, installationOf[userA], "t")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
	_, err = execAsUser(t, db.Connect(t, pgtest.RoleWebhook), userA, claimSQL, userA, installationOf[userA], "t")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
}

// The definer roles' new reach is read-only for the worker selection, and confined to
// devices for the claim; the purge's policies are pinned by the purge suite.
func TestReminderFunctions_DefinerReach(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	assert.Equal(t, []string{
		"devices SELECT", "notification_log SELECT", "reminder_settings SELECT", "taxpayers SELECT",
	}, queryStrings(t, owner, `
SELECT tablename || ' ' || cmd FROM pg_policies
WHERE $1 = ANY (roles) AND policyname LIKE '%\_due\_reminders' ORDER BY 1`, workerDefinerRole))
	assert.Equal(t, []string{"devices DELETE", "devices INSERT", "devices SELECT", "devices UPDATE"},
		queryStrings(t, owner, `
SELECT tablename || ' ' || cmd FROM pg_policies
WHERE $1 = ANY (roles) AND tablename = 'devices' ORDER BY 1`, authDefinerRole))

	assert.Empty(t, queryStrings(t, owner, `
SELECT c.relname || ' ' || a.privilege_type
FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.relname IN ('devices', 'reminder_settings', 'notification_log', 'taxpayers')
  AND a.grantee = $1::regrole AND a.privilege_type <> 'SELECT'
  AND NOT (c.relname = 'notification_log' AND a.privilege_type = 'DELETE')`, workerDefinerRole),
		"the worker definer only reads (the log's DELETE is the 13-month purge's, M9-06)")
	assert.Empty(t, queryStrings(t, owner, `
SELECT c.relname FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.relname IN ('reminder_settings', 'notification_log') AND a.grantee = $1::regrole`, authDefinerRole),
		"the claim reaches devices only")
}

// US-R6 and §2.3: a first registration creates, a repeat refreshes; a token or an
// installation held by anyone else moves to the caller, who learns nothing about it.
func TestClaimPushToken_CreatesRefreshesAndMoves(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	phone, tablet, reinstall := installation(userA, 1), installation(userA, 2), installation(userB, 3)

	assert.True(t, claim(t, app, userA, phone, "token-phone"), "a new installation is created (201)")
	assert.False(t, claim(t, app, userA, phone, "token-phone-rotated"), "the same installation is refreshed (200)")
	assert.True(t, claim(t, app, userA, tablet, "token-tablet"))
	assert.Equal(t, []string{phone, tablet}, devicesOf(t, app, userA))

	t.Run("B logs in on A's phone: the installation moves", func(t *testing.T) {
		assert.True(t, claim(t, app, userB, phone, "token-phone-rotated"), "same 201 as a fresh device: no oracle")
		assert.Equal(t, []string{tablet}, devicesOf(t, app, userA), "A loses the phone, keeps the tablet")
		assert.Equal(t, []string{phone}, devicesOf(t, app, userB))
		require.NoError(t, pgtest.CheckCannotSee(t.Context(), app, devicesTable, userB, userA),
			"the claim leaves B unable to read A's other devices")
	})
	t.Run("B reinstalls with A's tablet token: the token moves", func(t *testing.T) {
		assert.True(t, claim(t, app, userB, reinstall, "token-tablet"))
		assert.Empty(t, devicesOf(t, app, userA))
		assert.Equal(t, []string{phone, reinstall}, devicesOf(t, app, userB))
	})
	t.Run("the same user's stale installation of the token goes", func(t *testing.T) {
		fresh := installation(userB, 4)
		assert.True(t, claim(t, app, userB, fresh, "token-tablet"))
		assert.Equal(t, []string{phone, fresh}, devicesOf(t, app, userB))
	})
	t.Run("the claim scope does not outlive the call", func(t *testing.T) {
		var id, hmac string
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			var created bool
			if err := tx.QueryRowContext(t.Context(), claimSQL, userA, installation(userA, 5), "token-5").Scan(&created); err != nil {
				return err
			}
			return tx.QueryRowContext(t.Context(), `
SELECT coalesce(current_setting('app.device_claim_id', true), ''),
       coalesce(current_setting('app.device_claim_hmac', true), '')`).Scan(&id, &hmac)
		}))
		assert.Empty(t, id)
		assert.Empty(t, hmac)
	})
}

// The 11th device evicts the least recently seen one (OQ-R6), in the same call.
func TestClaimPushToken_EvictsTheLeastRecentlySeenBeyondTheCap(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	want := make([]string, 0, deviceCap)
	for slot := range deviceCap + 1 {
		id := installation(userA, slot)
		claim(t, app, userA, id, fmt.Sprintf("token-%d", slot))
		if slot > 0 {
			want = append(want, id)
		}
	}
	assert.Equal(t, want, devicesOf(t, app, userA), "slot 0, seen first, is evicted")
}

func TestClaimPushToken_RejectsForeignOrMalformedClaims(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)

	_, err := execAsUser(t, app, userA, claimSQL, userB, installationOf[userB], "token-b")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
	err = pgtest.WithoutUser(t.Context(), app, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), claimSQL, userA, installationOf[userA], "token-a")
		return err
	})
	requireSQLState(t, err, sqlstateInsufficientPrivilege)

	const custom = "SELECT claim_push_token($1::uuid, $2::uuid, $3, '\\x0001'::bytea || sha256('c') || sha256('c'), 1::smallint, $4, $5)"
	cases := []struct {
		name     string
		id       any
		platform string
		hmac     any
		version  any
		state    string
	}{
		{"short hmac", installationOf[userA], "ios", []byte("too-short"), nil, sqlstateInvalidParameterValue},
		{"no installation id", nil, "ios", nil, nil, sqlstateInvalidParameterValue},
		{"unknown platform", installationOf[userA], "windows", nil, nil, sqlstateCheckViolation},
		{"malformed app version", installationOf[userA], "ios", nil, "1.2 beta", sqlstateCheckViolation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hmac := c.hmac
			if hmac == nil {
				hmac = make32(c.name)
			}
			_, err := execAsUser(t, app, userA, custom, userA, c.id, c.platform, hmac, c.version)
			requireSQLState(t, err, c.state)
		})
	}
}

// make32 is a 32-byte value derived from label (a stand-in HMAC, never a real one).
func make32(label string) []byte {
	out := make([]byte, 32)
	copy(out, strings.Repeat(label, 32))
	return out
}

func TestReminderSettings_DefaultsAndRules(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)

	_, err := execAsUser(t, app, userA, "INSERT INTO reminder_settings (user_id) VALUES ($1)", userA)
	require.NoError(t, err)
	var row string
	require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), `
SELECT concat_ws(' ', enabled::text, early_day, on_due_date::text, to_char(send_time, 'HH24:MI'), show_amounts::text)
FROM reminder_settings`).Scan(&row)
	}))
	assert.Equal(t, "false 14 true 10:00 false", row, "the §2.2 defaults")

	for name, set := range map[string]string{
		"early day past 16":   "early_day = 17",
		"early day zero":      "early_day = 0",
		"seconds in the time": "send_time = '10:00:30'",
		"enabled, nothing selected (no_reminder_selected)": "enabled = true, early_day = NULL, on_due_date = false",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, "UPDATE reminder_settings SET "+set)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
	affected, err := execAsUser(t, app, userA,
		"UPDATE reminder_settings SET enabled = true, early_day = NULL, on_due_date = true, send_time = '07:30'")
	require.NoError(t, err)
	assert.EqualValues(t, 1, affected, "no early reminder is fine while the due-date one is on")
}

// §2.4: the UNIQUE (user, period, kind, channel) decides who sends, not the scheduler.
func TestNotificationLog_ClaimIsExactlyOnce(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	first, second := db.Connect(t, pgtest.RoleWorker), db.Connect(t, pgtest.RoleWorker)

	tx, err := first.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", userA)
	require.NoError(t, err)
	var firstID string
	require.NoError(t, tx.QueryRowContext(t.Context(), claimLogSQL, userA).Scan(&firstID))

	var wg sync.WaitGroup
	var secondErr error
	wg.Go(func() {
		secondErr = pgtest.AsUser(t.Context(), second, userA, func(tx *sql.Tx) error {
			var id string
			return tx.QueryRowContext(t.Context(), claimLogSQL, userA).Scan(&id)
		})
	})
	require.NoError(t, tx.Commit())
	wg.Wait()
	require.ErrorIs(t, secondErr, sql.ErrNoRows, "the split-brain twin claims nothing")

	_, err = execAsUser(t, second, userA, insertLogSQL, userA, periodApr, kindEarly, "pending", 1, earlyNow)
	requireSQLState(t, err, sqlstateUniqueViolation)
	for _, other := range [][2]string{{periodApr, kindDue}, {"2026-05-01", kindEarly}} {
		_, err = execAsUser(t, second, userA, insertLogSQL, userA, other[0], other[1], "pending", 1, earlyNow)
		require.NoError(t, err, "another period or kind is another reminder")
	}
}

func TestNotificationLog_StatesAndRetries(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	seedUsers(t, app, userA)

	bad := map[string][]any{
		"period not a month start": {userA, "2026-04-02", kindEarly, "pending", 1, earlyNow},
		"unknown kind":             {userA, periodApr, "late", "pending", 1, earlyNow},
		"unknown status":           {userA, periodApr, kindEarly, "queued", 1, earlyNow},
		"a fourth attempt":         {userA, periodApr, kindEarly, "pending", 4, earlyNow},
		"sent without sent_at":     {userA, periodApr, kindEarly, "sent", 1, earlyNow},
	}
	for name, args := range bad {
		t.Run(name, func(t *testing.T) {
			_, err := execAsUser(t, worker, userA, insertLogSQL, args...)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}

	_, err := execAsUser(t, worker, userA, insertLogSQL, userA, periodApr, kindEarly, "pending", 1, "2026-05-14 15:00:00+00")
	require.NoError(t, err)
	reclaim := `UPDATE notification_log SET attempts = attempts + 1, claimed_at = $2::timestamptz
WHERE user_id = $1 AND status = 'pending' AND claimed_at < $2::timestamptz - interval '10 minutes'`
	affected, err := execAsUser(t, worker, userA, reclaim, userA, earlyNow)
	require.NoError(t, err)
	assert.EqualValues(t, 1, affected, "a stale lease is re-claimed")
	affected, err = execAsUser(t, worker, userA, reclaim, userA, earlyNow)
	require.NoError(t, err)
	assert.Zero(t, affected, "a fresh lease is not")

	_, err = execAsUser(t, worker, userA,
		"UPDATE notification_log SET status = 'sent', sent_at = now() WHERE user_id = $1", userA)
	requireSQLState(t, err, sqlstateCheckViolation)

	affected, err = execAsUser(t, worker, userA,
		"UPDATE notification_log SET status = 'sent', sent_at = now(), devices_ok = 1, devices_failed = 0 WHERE user_id = $1", userA)
	require.NoError(t, err)
	assert.EqualValues(t, 1, affected)
	_, err = execAsUser(t, worker, userA, "UPDATE notification_log SET status = 'pending', sent_at = NULL WHERE user_id = $1", userA)
	requireSQLState(t, err, sqlstateRaiseException)
	_, err = execAsUser(t, worker, userA, "UPDATE notification_log SET period = DATE '2026-05-01'")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
	_, err = execAsUser(t, worker, userA, "DELETE FROM notification_log")
	requireSQLState(t, err, sqlstateInsufficientPrivilege)
}

// dueFixture seeds users with every reason to be picked or skipped at earlyNow / dueNow.
type dueFixture struct {
	app, worker *sql.DB
	picked      string // enabled, early day 14, on due date, 10:00, device + live taxpayer
	noDevice    string
	disabled    string
	later       string // send time 11:00
	archived    string // only an archived taxpayer
	otherDay    string // early day 15, not on the due date
}

func seedDueFixture(t *testing.T) dueFixture {
	t.Helper()
	db := pgtest.NewDatabase(t)
	f := dueFixture{
		app: db.Connect(t, pgtest.RoleApp), worker: db.Connect(t, pgtest.RoleWorker),
		picked:   "00000000-0000-4000-8000-0000000009c1",
		noDevice: "00000000-0000-4000-8000-0000000009c2",
		disabled: "00000000-0000-4000-8000-0000000009c3",
		later:    "00000000-0000-4000-8000-0000000009c4",
		archived: "00000000-0000-4000-8000-0000000009c5",
		otherDay: "00000000-0000-4000-8000-0000000009c6",
	}
	type setup struct {
		enabled  bool
		early    any
		onDue    bool
		at       string
		device   bool
		archived bool
	}
	setups := map[string]setup{
		f.picked:   {true, 14, true, "10:00", true, false},
		f.noDevice: {true, 14, true, "10:00", false, false},
		f.disabled: {false, 14, true, "10:00", true, false},
		f.later:    {true, 14, true, "11:00", true, false},
		f.archived: {true, 14, true, "10:00", true, true},
		f.otherDay: {true, 15, false, "10:00", true, false},
	}
	for user, s := range setups {
		seedUsers(t, f.app, user)
		taxpayer := "00000000-0000-7000-8000-" + user[len(user)-12:]
		seedTaxpayer(t, f.app, user, taxpayer, "fisica")
		if s.archived {
			_, err := execAsUser(t, f.app, user, "UPDATE taxpayers SET archived_at = now()")
			require.NoError(t, err)
		}
		if s.device {
			claim(t, f.app, user, installation(user, 0), "token-"+user)
		}
		_, err := execAsUser(t, f.app, user, insertSettingsSQL, user, s.enabled, s.early, s.onDue, s.at)
		require.NoError(t, err)
	}
	return f
}

// §2.4 steps 2 and 5: enabled users with a device and a live taxpayer whose reminder of
// today is due by now and not yet logged (or whose claim's lease expired).
func TestDueReminderUsers_PicksWhoIsDueNow(t *testing.T) {
	f := seedDueFixture(t)
	early := f.picked + " " + periodApr + " " + kindEarly

	assert.Equal(t, []string{early}, dueUsers(t, f.worker, earlyNow, false, nil, dueLimit))
	assert.Empty(t, dueUsers(t, f.worker, "2026-05-14 15:55:00+00", false, nil, dueLimit), "09:55 local: not yet")
	assert.Equal(t, []string{
		f.picked + " " + periodApr + " " + kindDue, f.later + " " + periodApr + " " + kindDue,
	}, dueUsers(t, f.worker, "2026-05-18 17:30:00+00", true, nil, dueLimit), "the due date, 11:30 local")
	assert.Equal(t, []string{f.otherDay + " " + periodApr + " " + kindEarly},
		dueUsers(t, f.worker, "2026-05-15 16:05:00+00", false, nil, dueLimit), "day 15 picks the day-15 user")

	t.Run("keyset batches", func(t *testing.T) {
		both := dueUsers(t, f.worker, "2026-05-18 17:30:00+00", true, nil, dueLimit)
		assert.Equal(t, both[:1], dueUsers(t, f.worker, "2026-05-18 17:30:00+00", true, nil, 1))
		assert.Equal(t, both[1:], dueUsers(t, f.worker, "2026-05-18 17:30:00+00", true, f.picked, dueLimit))
	})

	// One log row of today's early reminder for the picked user, in each state (a fresh
	// fixture per state: a final row can be neither changed nor deleted).
	cases := []struct {
		name      string
		status    string
		attempts  int
		claimedAt string
		want      []string
	}{
		{"a fresh claim holds the lease", "pending", 1, "2026-05-14 16:00:00+00", nil},
		{"an expired lease is retried", "pending", 2, "2026-05-14 15:54:00+00", []string{early}},
		{"the third attempt was the last", "pending", 3, "2026-05-14 15:00:00+00", nil},
		{"sent is final", "sent", 1, "2026-05-14 15:00:00+00", nil},
		{"skipped is final", "skipped", 1, "2026-05-14 15:00:00+00", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := seedDueFixture(t)
			var sentAt any
			if c.status == "sent" {
				sentAt = c.claimedAt
			}
			_, err := execAsUser(t, g.worker, g.picked, `
INSERT INTO notification_log (user_id, period, kind, channel, status, attempts, claimed_at, sent_at, devices_ok)
VALUES ($1, $2::date, 'early', 'push', $3, $4, $5::timestamptz, $6::timestamptz, CASE WHEN $3 = 'sent' THEN 1 END)`,
				g.picked, periodApr, c.status, c.attempts, c.claimedAt, sentAt)
			require.NoError(t, err)
			assert.Equal(t, c.want, dueUsers(t, g.worker, earlyNow, false, nil, dueLimit))
		})
	}
}

func TestDueReminderUsers_RejectsBadArguments(t *testing.T) {
	worker := pgtest.NewDatabase(t).Connect(t, pgtest.RoleWorker)
	for _, args := range [][]any{
		{nil, false, nil, dueLimit},
		{earlyNow, nil, nil, dueLimit},
		{earlyNow, false, nil, 0},
		{earlyNow, false, nil, maxDueRows + 1},
		{earlyNow, false, nil, nil},
	} {
		_, err := worker.ExecContext(t.Context(), dueUsersSQL, args...)
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	}
	_, err := worker.ExecContext(t.Context(), dueUsersSQL, earlyNow, false, nil, maxDueRows)
	require.NoError(t, err)
}

// The selection's scans use indexes: the partial settings index of each branch, the
// devices and taxpayers user indexes and the log's dedupe key. Planned as the definer
// role (cc_owner is denied every row by FORCE RLS).
func TestDueReminderUsers_ScansUseIndexes(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	const exists = `
  AND EXISTS (SELECT 1 FROM devices d WHERE d.user_id = s.user_id)
  AND EXISTS (SELECT 1 FROM taxpayers p WHERE p.user_id = s.user_id AND p.archived_at IS NULL)
  AND NOT EXISTS (SELECT 1 FROM notification_log l WHERE l.user_id = s.user_id AND l.period = DATE '2026-04-01'
                  AND l.kind = 'early' AND l.channel = 'push')
ORDER BY s.user_id LIMIT 100`
	scans := []struct{ name, query string }{
		{"early", "SELECT s.user_id FROM reminder_settings s WHERE s.enabled AND s.early_day = 14 AND s.send_time <= '10:05'" + exists},
		{"due", "SELECT s.user_id FROM reminder_settings s WHERE s.enabled AND s.on_due_date AND s.send_time <= '10:05'" + exists},
	}
	// taxpayers has two (user_id, id, ...) unique indexes; either serves the probe.
	want := map[string][]string{
		"early": {"reminder_settings_early_day_idx", "devices_user_id_last_seen_at_idx", "taxpayers_user_id_id",
			"notification_log_user_id_period_kind_channel_key"},
		"due": {"reminder_settings_on_due_date_idx", "devices_user_id_last_seen_at_idx", "taxpayers_user_id_id",
			"notification_log_user_id_period_kind_channel_key"},
	}
	for _, scan := range scans {
		t.Run(scan.name, func(t *testing.T) {
			tx, err := owner.BeginTx(t.Context(), nil)
			require.NoError(t, err)
			defer func() { _ = tx.Rollback() }()
			for _, set := range []string{"SET LOCAL ROLE " + workerDefinerRole, "SET LOCAL enable_seqscan = off"} {
				_, err = tx.ExecContext(t.Context(), set)
				require.NoError(t, err)
			}
			rows, err := tx.QueryContext(t.Context(), "EXPLAIN "+scan.query)
			require.NoError(t, err)
			var plan strings.Builder
			for rows.Next() {
				var line string
				require.NoError(t, rows.Scan(&line))
				plan.WriteString(line + "\n")
			}
			require.NoError(t, rows.Err())
			require.NoError(t, rows.Close())
			for _, index := range want[scan.name] {
				assert.Contains(t, plan.String(), index)
			}
		})
	}
}

// Down removes the tables, both functions, the claim helpers and every policy and grant
// this migration gave the definer roles; Up restores them.
func TestRemindersMigration_DownRemovesEverything(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), remindersVersion-1)
	require.NoError(t, err)
	assert.Empty(t, queryStrings(t, owner,
		"SELECT relname::text FROM pg_class WHERE relname IN ('devices', 'reminder_settings', 'notification_log')"))
	assert.Empty(t, queryStrings(t, owner, `
SELECT proname::text FROM pg_proc
WHERE proname IN ('claim_push_token', 'due_reminder_users', 'device_claim_id', 'device_claim_hmac',
                  'notification_log_forbid_final_change')`))
	assert.Empty(t, queryStrings(t, owner,
		"SELECT policyname::text FROM pg_policies WHERE policyname LIKE '%\\_due\\_reminders'"))
	var canExecute bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT has_function_privilege($1, 'app_user_id()', 'EXECUTE')", authDefinerRole).Scan(&canExecute))
	assert.False(t, canExecute, "app_user_id() goes back to the runtime roles only")
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), owner, taxpayersTable))

	require.NoError(t, migrations.Up(t.Context(), owner))
	worker := db.Connect(t, pgtest.RoleWorker)
	assert.Empty(t, dueUsers(t, worker, earlyNow, false, nil, dueLimit))
}
