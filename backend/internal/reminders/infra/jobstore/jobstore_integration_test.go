//go:build integration

package jobstore_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/holidays"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/jobstore"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase/pushtest"
)

// The reminder job's store over the real migrations, as cc_worker under
// RLS, and the whole job end to end with a fake push sender. Fixture ids are
// this suite's own.
const (
	userA = "0199d4a2-0000-7000-8000-000000000a06"
	userB = "0199d4a2-0000-7000-8000-000000000b06"

	taxpayerA1 = "0199d4a3-0000-7000-8000-000000000a61"
	taxpayerA2 = "0199d4a3-0000-7000-8000-000000000a62"
	taxpayerA3 = "0199d4a3-0000-7000-8000-000000000a63" // archived
	taxpayerB1 = "0199d4a3-0000-7000-8000-000000000b61"

	deviceA1 = "0199d4a4-0000-7000-8000-000000000a71"
	deviceA2 = "0199d4a4-0000-7000-8000-000000000a72"
	deviceB1 = "0199d4a4-0000-7000-8000-000000000b71"

	tokenKey = crypto.KeyID(1)
)

// 14 May 2026 10:05 Mexico City: the early reminder (day 14) of period
// 2026-04, due Monday 18 May.
var earlyNow = time.Date(2026, time.May, 14, 10, 5, 0, 0, clock.Location())

const (
	insertUserSQL = `
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')),
        '\x0001'::bytea || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8')),
        1, '$argon2id$v=19$m=8,t=1,p=1$' || rtrim(encode(sha256(convert_to($1::text, 'UTF8')), 'base64'), '=') || '$' ||
           rtrim(encode(sha256(convert_to($1::text, 'UTF8')), 'base64'), '='), 'Fixture')`
	insertUsageSQL = "INSERT INTO user_usage (user_id) VALUES ($1)"
	insertTrialSQL = `INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at)
VALUES ($1, 'trial', 'registration', now(), now() + interval '7 days')`
	// insertTaxpayerSQL derives every RFC column from the taxpayer id. Taxpayers
	// are inserted in switcher order (created_at is the transaction's now()).
	insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, 'fisica', $3,
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')), 'GOMA******AB6', true)`
	insertRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period, end_period)
VALUES ($1, $2, $3, 'fisica', $4, $5)`
	insertSettingsSQL = `
INSERT INTO reminder_settings (user_id, enabled, early_day, on_due_date, send_time, show_amounts)
VALUES ($1, true, 14, true, '10:00', false)`
)

type env struct {
	app, worker *sql.DB
	units       *pgdb.TxManager // cc_worker
	vault       *pushtoken.Vault
	tokens      map[string]string
	store       jobstore.Store
}

func exec(t *testing.T, db *sql.DB, user, query string, args ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), db, user, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), query, args...)
		return err
	}))
}

// newEnv seeds users A and B: A with two live taxpayers (612, then a 621 one
// created later), an archived one and two devices; B with one taxpayer and
// one device. Both enabled with the defaults.
func newEnv(t *testing.T) *env {
	t.Helper()
	app, worker := databases(t)
	e := &env{app: app, worker: worker, tokens: map[string]string{}, store: jobstore.New()}
	var err error
	e.units, err = pgdb.NewTxManager(worker)
	require.NoError(t, err)
	key, pepper := make([]byte, crypto.KeySize), make([]byte, crypto.MinPepperSize)
	_, _ = rand.Read(key)
	_, _ = rand.Read(pepper)
	keyring, err := crypto.NewKeyring(tokenKey, map[crypto.KeyID][]byte{tokenKey: key})
	require.NoError(t, err)
	index, err := crypto.NewTokenIndex(pepper)
	require.NoError(t, err)
	e.vault = pushtoken.NewVault(keyring, index)

	for _, user := range []string{userA, userB} {
		exec(t, app, user, insertUserSQL, user)
		exec(t, app, user, insertUsageSQL, user)
		exec(t, app, user, insertTrialSQL, user)
		exec(t, app, user, insertSettingsSQL, user)
	}
	e.taxpayer(t, userA, taxpayerA1, "Ana", "612", "2026-01")
	e.taxpayer(t, userA, taxpayerA2, "Taller Sur", "621", "2026-01")
	e.taxpayer(t, userA, taxpayerA3, "Archivado", "612", "2026-01")
	exec(t, app, userA, "UPDATE taxpayers SET archived_at = now() WHERE id = $1", taxpayerA3)
	e.taxpayer(t, userB, taxpayerB1, "Beto", "612", "2026-01")
	e.device(t, userA, deviceA1)
	e.device(t, userA, deviceA2)
	e.device(t, userB, deviceB1)
	return e
}

func (e *env) taxpayer(t *testing.T, user, id, name, regime, start string) {
	t.Helper()
	exec(t, e.app, user, insertTaxpayerSQL, user, id, name)
	exec(t, e.app, user, insertRegimeSQL, user, id, regime, start, nil)
}

func (e *env) device(t *testing.T, user, id string) {
	t.Helper()
	raw := "fcm:" + rand.Text() + rand.Text()
	token, err := domain.NewPushToken(raw)
	require.NoError(t, err)
	sealed, err := e.vault.Seal(token, usecase.TokenBinding{InstallationID: id, UserID: user})
	require.NoError(t, err)
	appUnits, err := pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	require.NoError(t, appUnits.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		_, err := store.NewDevices().Claim(ctx, user, usecase.SealedDevice{
			InstallationID: id, Platform: domain.PlatformAndroid, Token: sealed,
		})
		return err
	}))
	e.tokens[id] = raw
}

func (e *env) as(t *testing.T, user string, work func(ctx context.Context) error) {
	t.Helper()
	require.NoError(t, e.units.ReadWrite(t.Context(), user, work))
}

func april(t *testing.T) domain.Period {
	t.Helper()
	p, err := domain.NewPeriod(2026, time.April)
	require.NoError(t, err)
	return p
}

func (e *env) due(t *testing.T, now time.Time, dueToday bool, after string, limit int) []usecase.DueReminder {
	t.Helper()
	var out []usecase.DueReminder
	require.NoError(t, e.units.System(t.Context(), func(ctx context.Context) error {
		var err error
		out, err = e.store.DueReminders(ctx, now, dueToday, after, limit)
		return err
	}))
	return out
}

// §2.4 step 2: the due_reminder_users read, paged by user id.
func TestDueReminders(t *testing.T) {
	e := newEnv(t)
	want := []usecase.DueReminder{
		{UserID: userA, Period: april(t), Kind: domain.KindEarly},
		{UserID: userB, Period: april(t), Kind: domain.KindEarly},
	}

	assert.Equal(t, want, e.due(t, earlyNow, false, "", 10))
	assert.Equal(t, want[:1], e.due(t, earlyNow, false, "", 1))
	assert.Equal(t, want[1:], e.due(t, earlyNow, false, userA, 10))
	assert.Empty(t, e.due(t, earlyNow.Add(-10*time.Minute), false, "", 10), "09:55: before the send time")
	due := e.due(t, time.Date(2026, time.May, 18, 10, 0, 0, 0, clock.Location()), true, "", 10)
	require.Len(t, due, 2)
	assert.Equal(t, domain.KindDue, due[0].Kind)
}

func (e *env) claim(t *testing.T, user string, r usecase.DueReminder, now time.Time) (usecase.LogClaim, bool) {
	t.Helper()
	var (
		c  usecase.LogClaim
		ok bool
	)
	e.as(t, user, func(ctx context.Context) error {
		var err error
		c, ok, err = e.store.Claim(ctx, r, now, now.Add(-10*time.Minute))
		return err
	})
	return c, ok
}

// §2.4 steps 3 and 5: one claim per reminder, even when two workers race; a
// re-claim only after the lease, once, up to three attempts.
func TestClaim_ExactlyOnceAndLease(t *testing.T) {
	e := newEnv(t)
	r := usecase.DueReminder{UserID: userA, Period: april(t), Kind: domain.KindEarly}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		claimed []usecase.LogClaim
	)
	for range 4 {
		wg.Go(func() {
			c, ok := e.claim(t, userA, r, earlyNow)
			if ok {
				mu.Lock()
				claimed = append(claimed, c)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	require.Len(t, claimed, 1, "the UNIQUE key decides who sends")
	assert.Equal(t, 1, claimed[0].Attempts)
	assert.Equal(t, []usecase.DueReminder{{UserID: userB, Period: april(t), Kind: domain.KindEarly}},
		e.due(t, earlyNow.Add(time.Minute), false, "", 10), "a claimed reminder is no longer due")

	_, ok := e.claim(t, userA, r, earlyNow.Add(9*time.Minute))
	assert.False(t, ok, "the lease is live")

	later := earlyNow.Add(11 * time.Minute)
	var reclaimed []usecase.LogClaim
	for range 3 {
		wg.Go(func() {
			c, ok := e.claim(t, userA, r, later)
			if ok {
				mu.Lock()
				reclaimed = append(reclaimed, c)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	require.Len(t, reclaimed, 1, "two workers racing to re-claim produce one retry")
	assert.Equal(t, usecase.LogClaim{ID: claimed[0].ID, Attempts: 2}, reclaimed[0])

	third, ok := e.claim(t, userA, r, later.Add(11*time.Minute))
	require.True(t, ok)
	assert.Equal(t, 3, third.Attempts)
	_, ok = e.claim(t, userA, r, later.Add(30*time.Minute))
	assert.False(t, ok, "no fourth attempt")

	_, ok = e.claim(t, userB, usecase.DueReminder{UserID: userB, Period: april(t), Kind: domain.KindEarly}, earlyNow)
	assert.True(t, ok, "another user's reminder is another row")
}

func (e *env) finish(t *testing.T, user string, c usecase.LogClaim, o usecase.LogOutcome) bool {
	t.Helper()
	var recorded bool
	e.as(t, user, func(ctx context.Context) error {
		var err error
		recorded, err = e.store.Finish(ctx, c, o)
		return err
	})
	return recorded
}

type logRow struct {
	status, class                  string
	attempts                       int
	sentAt                         sql.NullTime
	devicesOK, devicesFailed       sql.NullInt64
}

func (e *env) logRow(t *testing.T, user string, kind domain.Kind) logRow {
	t.Helper()
	var r logRow
	var class sql.NullString
	require.NoError(t, pgtest.AsUser(t.Context(), e.worker, user, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), `SELECT status, error_class, attempts, sent_at, devices_ok, devices_failed
FROM notification_log WHERE kind = $1`, string(kind)).Scan(&r.status, &class, &r.attempts, &r.sentAt, &r.devicesOK, &r.devicesFailed)
	}))
	r.class = class.String
	return r
}

// Finish records the attempt the caller holds, once; a stale attempt or a
// final row records nothing.
func TestFinish(t *testing.T) {
	e := newEnv(t)
	r := usecase.DueReminder{UserID: userA, Period: april(t), Kind: domain.KindEarly}
	c, ok := e.claim(t, userA, r, earlyNow)
	require.True(t, ok)

	pending := usecase.LogOutcome{Status: usecase.LogPending, DevicesFailed: 2, ErrorClass: "unavailable"}
	assert.True(t, e.finish(t, userA, c, pending))
	assert.Equal(t, logRow{status: "pending", class: "unavailable", attempts: 1,
		devicesOK: sql.NullInt64{Int64: 0, Valid: true}, devicesFailed: sql.NullInt64{Int64: 2, Valid: true}}, e.logRow(t, userA, domain.KindEarly))

	assert.False(t, e.finish(t, userA, usecase.LogClaim{ID: c.ID, Attempts: 2}, pending), "not the attempt held")
	assert.False(t, e.finish(t, userB, c, pending), "another user's row is invisible")

	sent := usecase.LogOutcome{Status: usecase.LogSent, SentAt: earlyNow, DevicesOK: 1, DevicesFailed: 1, ErrorClass: "unregistered"}
	assert.True(t, e.finish(t, userA, c, sent))
	got := e.logRow(t, userA, domain.KindEarly)
	assert.Equal(t, "sent", got.status)
	assert.True(t, got.sentAt.Time.Equal(earlyNow))
	assert.False(t, e.finish(t, userA, c, sent), "sent is final")
}

// §2.4 step 5: a pending row whose day passed becomes failed / expired; the
// caller's other rows and other users' rows stay.
func TestExpireStale(t *testing.T) {
	e := newEnv(t)
	yesterday := earlyNow.Add(-24 * time.Hour)
	_, ok := e.claim(t, userA, usecase.DueReminder{UserID: userA, Period: april(t), Kind: domain.KindEarly}, yesterday)
	require.True(t, ok)
	_, ok = e.claim(t, userB, usecase.DueReminder{UserID: userB, Period: april(t), Kind: domain.KindEarly}, yesterday)
	require.True(t, ok)
	_, ok = e.claim(t, userA, usecase.DueReminder{UserID: userA, Period: april(t), Kind: domain.KindDue}, earlyNow)
	require.True(t, ok)

	var n int
	startOfDay := time.Date(2026, time.May, 14, 0, 0, 0, 0, clock.Location())
	e.as(t, userA, func(ctx context.Context) error {
		var err error
		n, err = e.store.ExpireStale(ctx, startOfDay)
		return err
	})

	assert.Equal(t, 1, n)
	assert.Equal(t, logRow{status: "failed", class: "expired", attempts: 1}, e.logRow(t, userA, domain.KindEarly))
	assert.Equal(t, "pending", e.logRow(t, userA, domain.KindDue).status)
	assert.Equal(t, "pending", e.logRow(t, userB, domain.KindEarly).status, "RLS: only the caller's rows")
}

// The caller's live taxpayers in the switcher's order with their regimes,
// and their devices; never another user's.
func TestTaxpayersAndTargets(t *testing.T) {
	e := newEnv(t)
	jan, err := domain.NewPeriod(2026, time.January)
	require.NoError(t, err)

	var (
		taxpayers []usecase.RecipientTaxpayer
		targets   []usecase.SealedTarget
	)
	e.as(t, userA, func(ctx context.Context) error {
		if taxpayers, err = e.store.Taxpayers(ctx); err != nil {
			return err
		}
		targets, err = e.store.Targets(ctx)
		return err
	})

	assert.Equal(t, []usecase.RecipientTaxpayer{
		{ID: taxpayerA1, DisplayName: "Ana", Regimes: []usecase.HeldRegime{{Code: "612", Start: jan}}},
		{ID: taxpayerA2, DisplayName: "Taller Sur", Regimes: []usecase.HeldRegime{{Code: "621", Start: jan}}},
	}, taxpayers)
	require.Len(t, targets, 2)
	for _, target := range targets {
		token, err := e.vault.Open(target.Ciphertext, usecase.TokenBinding{InstallationID: target.DeviceID, UserID: userA})
		require.NoError(t, err)
		assert.Equal(t, e.tokens[target.DeviceID], token.Reveal())
	}
}

// UNREGISTERED cleanup deletes the caller's device only.
func TestDeleteDevices(t *testing.T) {
	e := newEnv(t)
	var n int
	e.as(t, userA, func(ctx context.Context) error {
		var err error
		n, err = e.store.DeleteDevices(ctx, []string{deviceA1, deviceB1, "not-a-uuid"})
		return err
	})

	assert.Equal(t, 1, n)
	e.as(t, userB, func(ctx context.Context) error {
		targets, err := e.store.Targets(ctx)
		assert.Len(t, targets, 1, "B's device stays")
		return err
	})
}

type flagsOn struct{}

func (flagsOn) Enabled(context.Context, string) bool { return true }

type noEstimates struct{}

func (noEstimates) MonthlyEstimate(context.Context, string, string, domain.Period) (usecase.EstimateSummary, error) {
	return usecase.EstimateSummary{}, nil
}

type alerts struct{}

func (alerts) CalendarUnseeded(context.Context, int) {}

func (e *env) job(t *testing.T, sender usecase.PushSender, now time.Time) *usecase.ReminderJob {
	t.Helper()
	settings := store.NewSettings()
	job, err := usecase.NewReminderJob(usecase.JobDeps{
		System: e.units, Units: e.units, Due: e.store, Log: e.store, Settings: settings, Recipients: e.store,
		Opener: e.vault, Push: sender, Estimates: noEstimates{}, Flags: flagsOn{},
		Holidays: holidays.NewReader(), Alerts: alerts{}, Clock: clock.NewFixed(now),
		Logger: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), CollapseKey: bytes.Repeat([]byte{7}, 32),
		Concurrency: usecase.DefaultJobConcurrency, BatchSize: usecase.DefaultJobBatchSize,
	})
	require.NoError(t, err)
	return job
}

// US-R2, US-R5, US-R7 end to end: two worker instances in the split-brain
// window run the same tick; each user gets exactly one push, listing the
// taxpayers due that month in the switcher's order; an UNREGISTERED device
// is deleted; a later run that day sends nothing.
func TestReminderJob_EndToEnd(t *testing.T) {
	e := newEnv(t)
	sender := &pushtest.Sender{}
	sender.SetOutcome(uuid.MustParse(deviceA2), usecase.PushTokenInvalid, "unregistered")

	var wg sync.WaitGroup
	for range 2 {
		job := e.job(t, sender, earlyNow)
		wg.Go(func() { assert.NoError(t, job.Run(t.Context())) })
	}
	wg.Wait()

	sends := sender.Sends()
	require.Len(t, sends, 2, "one per user")
	bodies := map[string]string{}
	for _, s := range sends {
		bodies[s.DeviceIDs[0].String()] = s.Notification.Body
		assert.Equal(t, "Tu pago mensual estimado vence el lunes 18 de mayo", s.Notification.Title)
	}
	assert.Equal(t, "Revisa tu estimado de abril para: Ana, Taller Sur.", bodies[deviceA1], "621 in its bimester's second month")
	assert.Equal(t, "Revisa tu estimado de abril para: Beto.", bodies[deviceB1])
	a := e.logRow(t, userA, domain.KindEarly)
	assert.Equal(t, "sent", a.status)
	assert.EqualValues(t, 1, a.devicesOK.Int64)
	assert.EqualValues(t, 1, a.devicesFailed.Int64)
	e.as(t, userA, func(ctx context.Context) error {
		targets, err := e.store.Targets(ctx)
		require.Len(t, targets, 1)
		assert.Equal(t, deviceA1, targets[0].DeviceID)
		return err
	})

	require.NoError(t, e.job(t, sender, earlyNow.Add(5*time.Minute)).Run(t.Context()))
	assert.Len(t, sender.Sends(), 2, "a sent reminder is final")
}
