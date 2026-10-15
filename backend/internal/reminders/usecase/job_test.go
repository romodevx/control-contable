package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase/pushtest"
)

const (
	jobUser1 = "0199d4a2-1111-7000-8000-000000000001"
	jobUser2 = "0199d4a2-1111-7000-8000-000000000002"
	jobUser3 = "0199d4a2-1111-7000-8000-000000000003"

	tpAna    = "0199d4a3-2222-7000-8000-0000000000a1"
	tpTaller = "0199d4a3-2222-7000-8000-0000000000a2"
	tpRIF    = "0199d4a3-2222-7000-8000-0000000000a3"

	dev1 = "0199d4a4-3333-7000-8000-000000000001"
	dev2 = "0199d4a4-3333-7000-8000-000000000002"
	dev3 = "0199d4a4-3333-7000-8000-000000000003"
)

// The spec's US-R2 instants: 14 May 2026 10:05 (early, day 14) and 18 May
// 2026 10:05 (period 2026-04 is due Sunday 17 May, moved to Monday 18).
var (
	earlyNow = time.Date(2026, time.May, 14, 10, 5, 0, 0, clock.Location())
	dueNow   = time.Date(2026, time.May, 18, 10, 5, 0, 0, clock.Location())
)

func jobPeriod(t *testing.T, year int, month time.Month) domain.Period {
	t.Helper()
	p, err := domain.NewPeriod(year, month)
	require.NoError(t, err)
	return p
}

func jobCollapseKey() []byte { return []byte(strings.Repeat("c", 32)) }

func monthly(t *testing.T) []usecase.HeldRegime {
	return []usecase.HeldRegime{{Code: "612", Start: jobPeriod(t, 2026, time.January)}}
}

// jobFixture is one worker instance's job over a (possibly shared) jobDB.
type jobFixture struct {
	db        *jobDB
	due       *jobDue
	units     *jobUnits
	system    *jobSystem
	push      usecase.PushSender
	sender    *pushtest.Sender
	estimates *jobEstimates
	flags     jobFlags
	holidays  *fakeHolidays
	alerts    *syncAlerts
	logs      *bytes.Buffer
	now       time.Time
	tokens    map[string]string // device id -> token
	opts      func(*usecase.JobDeps)
}

func newJobFixture(t *testing.T, now time.Time) *jobFixture {
	t.Helper()
	sender := &pushtest.Sender{}
	return &jobFixture{
		db: newJobDB(), due: &jobDue{}, units: &jobUnits{}, system: &jobSystem{},
		push: sender, sender: sender, estimates: &jobEstimates{answers: map[string]usecase.EstimateSummary{}},
		flags:    jobFlags{usecase.ModuleReminders: true, usecase.ModuleTaxEngine: true},
		holidays: &fakeHolidays{years: []int{2026, 2027}}, alerts: &syncAlerts{},
		logs: &bytes.Buffer{}, now: now, tokens: map[string]string{},
	}
}

// addUser gives user enabled settings, taxpayers and one device per id,
// and makes kind of p due for them.
func (f *jobFixture) addUser(t *testing.T, user string, s domain.Settings, taxpayers []usecase.RecipientTaxpayer, devices ...string) {
	t.Helper()
	f.db.settings[user] = s
	f.db.taxpayers[user] = taxpayers
	for _, d := range devices {
		token := pushToken(t)
		f.tokens[d] = token
		f.db.devices[user] = append(f.db.devices[user], usecase.SealedTarget{DeviceID: d, Ciphertext: []byte("sealed:" + token)})
	}
}

func (f *jobFixture) dueFor(user string, p domain.Period, k domain.Kind) {
	f.due.reminders = append(f.due.reminders, usecase.DueReminder{UserID: user, Period: p, Kind: k})
}

func enabledSettings(t *testing.T, showAmounts bool) domain.Settings {
	t.Helper()
	at, err := domain.ParseSendTime("10:00")
	require.NoError(t, err)
	s, err := domain.NewSettings(true, 14, true, at, showAmounts)
	require.NoError(t, err)
	return s
}

func (f *jobFixture) deps() usecase.JobDeps {
	d := usecase.JobDeps{
		System: f.system, Units: f.units, Due: f.due, Log: f.db, Settings: f.db, Recipients: f.db,
		Opener: jobOpener{}, Push: f.push, Estimates: f.estimates, Flags: f.flags,
		Holidays: f.holidays, Alerts: f.alerts, Clock: clock.NewFixed(f.now),
		Logger: slog.New(slog.NewJSONHandler(f.logs, &slog.HandlerOptions{Level: slog.LevelDebug})), CollapseKey: jobCollapseKey(),
		Concurrency: usecase.DefaultJobConcurrency, BatchSize: usecase.DefaultJobBatchSize,
	}
	if f.opts != nil {
		f.opts(&d)
	}
	return d
}

func (f *jobFixture) job(t *testing.T) *usecase.ReminderJob {
	t.Helper()
	job, err := usecase.NewReminderJob(f.deps())
	require.NoError(t, err)
	return job
}

func (f *jobFixture) run(t *testing.T) error {
	t.Helper()
	return f.job(t).Run(t.Context())
}

func twoTaxpayers(t *testing.T) []usecase.RecipientTaxpayer {
	return []usecase.RecipientTaxpayer{
		{ID: tpAna, DisplayName: "Ana", Regimes: monthly(t)},
		{ID: tpTaller, DisplayName: "Taller Sur", Regimes: monthly(t)},
	}
}

func TestReminderJobIsAWorkerJob(t *testing.T) {
	job := newJobFixture(t, earlyNow).job(t)

	assert.Equal(t, "reminders", job.Name())
	assert.Equal(t, 5*time.Minute, job.Interval(), "§2.4: tick every 5 minutes")
}

// US-R2, early: exactly one push listing both names, through the per-user
// unit of work, and a sent log row.
func TestReminderJob_EarlyReminder(t *testing.T) {
	f := newJobFixture(t, earlyNow)
	april := jobPeriod(t, 2026, time.April)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1, dev2)
	f.dueFor(jobUser1, april, domain.KindEarly)

	require.NoError(t, f.run(t))

	require.Len(t, f.due.calls, 1)
	assert.Equal(t, dueCall{now: earlyNow, dueToday: false, after: "", limit: usecase.DefaultJobBatchSize, inSystem: true}, f.due.calls[0])
	sends := f.sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, usecase.PushNotification{
		Title:      "Tu pago mensual estimado vence el lunes 18 de mayo",
		Body:       "Revisa tu estimado de abril para: Ana, Taller Sur.",
		DeepLink:   "controlcontable://estimate?period=2026-04",
		CollapseID: domain.CollapseID(jobCollapseKey(), jobUser1, april, domain.KindEarly),
	}, sends[0].Notification)
	assert.Equal(t, []uuid.UUID{uuid.MustParse(dev1), uuid.MustParse(dev2)}, sends[0].DeviceIDs)
	row := f.db.row(t, jobUser1, april, domain.KindEarly)
	assert.Equal(t, usecase.LogSent, row.status)
	assert.Equal(t, usecase.LogOutcome{Status: usecase.LogSent, SentAt: earlyNow, DevicesOK: 2}, row.outcome)
	assert.Empty(t, f.estimates.calls, "show_amounts off: the estimator is never asked")
	for _, user := range f.units.users {
		assert.Equal(t, jobUser1, user)
	}
}

// US-R2, due date: "Hoy vence…", and due_reminder_users is asked for the
// due kind because today is the period's moved due date.
func TestReminderJob_DueDateReminder(t *testing.T) {
	f := newJobFixture(t, dueNow)
	april := jobPeriod(t, 2026, time.April)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindDue)

	require.NoError(t, f.run(t))

	require.Len(t, f.due.calls, 1)
	assert.True(t, f.due.calls[0].dueToday)
	sends := f.sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, "Hoy vence tu pago mensual estimado", sends[0].Notification.Title)
	assert.Equal(t, usecase.LogSent, f.db.row(t, jobUser1, april, domain.KindDue).status)
}

// Sunday 17 May is not the due date (moved to Monday): not a due day.
func TestReminderJob_NominalDayOnAWeekendIsNotTheDueDate(t *testing.T) {
	f := newJobFixture(t, time.Date(2026, time.May, 17, 10, 5, 0, 0, clock.Location()))

	require.NoError(t, f.run(t))

	require.Len(t, f.due.calls, 1)
	assert.False(t, f.due.calls[0].dueToday)
}

// A holiday on the due date moves it (the holidays port, not a weekday rule).
func TestReminderJob_HolidayMovesTheDueDate(t *testing.T) {
	f := newJobFixture(t, time.Date(2026, time.June, 17, 10, 5, 0, 0, clock.Location()))
	f.holidays.days = []string{"2026-06-17"}

	require.NoError(t, f.run(t))

	require.Len(t, f.due.calls, 1)
	assert.False(t, f.due.calls[0].dueToday, "17 June is a holiday: the due date is the 18th")
}

// §2.1: a due date in an unseeded year is computed weekends-only and the
// reminders_calendar_unseeded alert fires.
func TestReminderJob_UnseededYearAlertsAndFallsBackToWeekends(t *testing.T) {
	f := newJobFixture(t, time.Date(2028, time.January, 17, 10, 5, 0, 0, clock.Location()))
	f.holidays.years = []int{2026, 2027}

	require.NoError(t, f.run(t))

	assert.Equal(t, []int{2028}, f.alerts.unseeded)
	require.Len(t, f.due.calls, 1)
	assert.True(t, f.due.calls[0].dueToday, "Monday 17 Jan 2028 is the due date weekends-only")
}

// §3: module OFF (or implied OFF) → no-op: no read, no row, no send.
func TestReminderJob_FlagOffIsANoOp(t *testing.T) {
	f := newJobFixture(t, earlyNow)
	f.flags[usecase.ModuleReminders] = false
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, jobPeriod(t, 2026, time.April), domain.KindEarly)

	require.NoError(t, f.run(t))

	assert.Empty(t, f.due.calls)
	assert.Zero(t, f.db.rows())
	assert.Empty(t, f.sender.Sends())
	assert.Zero(t, f.holidays.reads)
}

// US-R5: two worker instances in the split-brain window run the same tick
// concurrently: one log row, one send. A later run that day sends nothing.
func TestReminderJob_ExactlyOnceAcrossTwoInstances(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	shared := newJobFixture(t, earlyNow)
	for i, user := range []string{jobUser1, jobUser2, jobUser3} {
		shared.addUser(t, user, enabledSettings(t, false), twoTaxpayers(t), []string{dev1, dev2, dev3}[i])
		shared.dueFor(user, april, domain.KindEarly)
	}
	instances := []*jobFixture{shared, newJobFixture(t, earlyNow)}
	instances[1].db, instances[1].due, instances[1].push, instances[1].sender = shared.db, &jobDue{reminders: shared.due.reminders}, shared.push, shared.sender

	var wg sync.WaitGroup
	for _, inst := range instances {
		job := inst.job(t)
		wg.Go(func() { assert.NoError(t, job.Run(t.Context())) })
	}
	wg.Wait()

	assert.Len(t, shared.sender.Sends(), 3, "one send per user, whichever instance claimed it")
	assert.Equal(t, 3, shared.db.rows())

	later := newJobFixture(t, earlyNow.Add(5*time.Minute))
	later.db, later.due, later.push, later.sender = shared.db, &jobDue{reminders: shared.due.reminders}, shared.push, shared.sender
	require.NoError(t, later.run(t))
	assert.Len(t, shared.sender.Sends(), 3, "a sent reminder is final")
}

// US-R5: a crash after the claim leaves a pending row; after the 10-minute
// lease the retry re-claims it (attempts 2) and sends with the same
// collapse id.
func TestReminderJob_RetryAfterLeaseKeepsTheCollapseID(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.db.log[logKey(jobUser1, april, domain.KindEarly)] = &jobLogRow{
		id: "log-crashed", status: usecase.LogPending, attempts: 1, claimedAt: earlyNow.Add(-11 * time.Minute),
	}

	require.NoError(t, f.run(t))

	row := f.db.row(t, jobUser1, april, domain.KindEarly)
	assert.Equal(t, usecase.LogSent, row.status)
	assert.Equal(t, 2, row.attempts)
	sends := f.sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, domain.CollapseID(jobCollapseKey(), jobUser1, april, domain.KindEarly), sends[0].Notification.CollapseID)
}

// A live lease (another sender is on it) is not taken over.
func TestReminderJob_LiveLeaseIsNotReclaimed(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.db.log[logKey(jobUser1, april, domain.KindEarly)] = &jobLogRow{
		id: "log-live", status: usecase.LogPending, attempts: 1, claimedAt: earlyNow.Add(-9 * time.Minute),
	}

	require.NoError(t, f.run(t))

	assert.Empty(t, f.sender.Sends())
	assert.Equal(t, 1, f.db.row(t, jobUser1, april, domain.KindEarly).attempts)
}

// §2.4 step 3: only transient errors and attempts < 3 → stays pending with
// its class; at the third attempt → failed.
func TestReminderJob_TransientFailure(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	tests := []struct {
		name        string
		attempts    int // of the crashed row; 0 = a fresh claim
		wantStatus  usecase.LogStatus
		wantAttempt int
	}{
		{"first attempt stays pending", 0, usecase.LogPending, 1},
		{"second attempt stays pending", 1, usecase.LogPending, 2},
		{"third attempt fails", 2, usecase.LogFailed, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newJobFixture(t, earlyNow)
			f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1, dev2)
			f.dueFor(jobUser1, april, domain.KindEarly)
			if tc.attempts > 0 {
				f.db.log[logKey(jobUser1, april, domain.KindEarly)] = &jobLogRow{
					id: "log-x", status: usecase.LogPending, attempts: tc.attempts, claimedAt: earlyNow.Add(-time.Hour),
				}
			}
			f.sender.SetOutcome(uuid.MustParse(dev1), usecase.PushTransient, "unavailable")
			f.sender.SetOutcome(uuid.MustParse(dev2), usecase.PushTransient, "deadline")

			require.NoError(t, f.run(t))

			row := f.db.row(t, jobUser1, april, domain.KindEarly)
			assert.Equal(t, tc.wantStatus, row.status)
			assert.Equal(t, tc.wantAttempt, row.attempts)
			assert.Equal(t, "unavailable", row.outcome.ErrorClass)
			assert.Equal(t, 2, row.outcome.DevicesFailed)
		})
	}
}

// US-R7: UNREGISTERED for one device deletes it; the other still receives.
func TestReminderJob_InvalidTokenDeletesThatDeviceOnly(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1, dev2)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.sender.SetOutcome(uuid.MustParse(dev1), usecase.PushTokenInvalid, "unregistered")

	require.NoError(t, f.run(t))

	assert.Equal(t, []string{dev1}, f.db.deleted)
	assert.Equal(t, []usecase.SealedTarget{{DeviceID: dev2, Ciphertext: []byte("sealed:" + f.tokens[dev2])}}, f.db.devices[jobUser1])
	row := f.db.row(t, jobUser1, april, domain.KindEarly)
	assert.Equal(t, usecase.LogOutcome{Status: usecase.LogSent, SentAt: earlyNow, DevicesOK: 1, DevicesFailed: 1, ErrorClass: "unregistered"}, row.outcome)
}

// Every device permanently refused → failed at once, devices kept.
func TestReminderJob_EveryDeviceRejectedFails(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1, dev2)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.sender.SetOutcome(uuid.MustParse(dev1), usecase.PushTokenInvalid, "unregistered")
	f.sender.SetOutcome(uuid.MustParse(dev2), usecase.PushRejected, "sender_id_mismatch")

	require.NoError(t, f.run(t))

	row := f.db.row(t, jobUser1, april, domain.KindEarly)
	assert.Equal(t, usecase.LogFailed, row.status)
	assert.Equal(t, 2, row.outcome.DevicesFailed)
	assert.Equal(t, []string{dev1}, f.db.deleted, "the rejected device is kept")
}

// A device whose token does not open counts as failed; the others send.
func TestReminderJob_UnreadableTokenCountsAsFailed(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.db.devices[jobUser1] = append(f.db.devices[jobUser1], usecase.SealedTarget{DeviceID: dev2, Ciphertext: []byte("garbage")})
	f.dueFor(jobUser1, april, domain.KindEarly)

	require.NoError(t, f.run(t))

	sends := f.sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, []uuid.UUID{uuid.MustParse(dev1)}, sends[0].DeviceIDs)
	assert.Equal(t, usecase.LogOutcome{Status: usecase.LogSent, SentAt: earlyNow, DevicesOK: 1, DevicesFailed: 1, ErrorClass: "token_unreadable"},
		f.db.row(t, jobUser1, april, domain.KindEarly).outcome)
}

// Whole-call send errors: a configuration or builder problem fails the
// reminder; anything else is transient.
func TestReminderJob_WholeCallSendErrors(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	tests := []struct {
		err        error
		wantStatus usecase.LogStatus
		wantClass  string
	}{
		{usecase.ErrPushDisabled, usecase.LogFailed, "push_disabled"},
		{usecase.ErrInvalidPushNotification, usecase.LogFailed, "invalid_notification"},
		{errors.New("provider exploded"), usecase.LogPending, "send_error"},
	}
	for _, tc := range tests {
		t.Run(tc.wantClass, func(t *testing.T) {
			f := newJobFixture(t, earlyNow)
			f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
			f.dueFor(jobUser1, april, domain.KindEarly)
			f.sender.Err = tc.err

			require.NoError(t, f.run(t))

			row := f.db.row(t, jobUser1, april, domain.KindEarly)
			assert.Equal(t, tc.wantStatus, row.status)
			assert.Equal(t, tc.wantClass, row.outcome.ErrorClass)
			assert.Equal(t, 1, row.outcome.DevicesFailed)
		})
	}
}

// A user whose devices vanished between the selection and the claim.
func TestReminderJob_NoDeviceLeftFails(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t))
	f.dueFor(jobUser1, april, domain.KindEarly)

	require.NoError(t, f.run(t))

	assert.Empty(t, f.sender.Sends())
	assert.Equal(t, usecase.LogOutcome{Status: usecase.LogFailed, ErrorClass: "no_device"}, f.db.row(t, jobUser1, april, domain.KindEarly).outcome)
}

// §2.4 step 6: no taxpayer due that day (a 621-only taxpayer in the
// bimester's first month) → skipped, nothing sent; and settings turned off
// since the selection → skipped too.
func TestReminderJob_Skipped(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	march := jobPeriod(t, 2026, time.March)
	rifOnly := []usecase.RecipientTaxpayer{{ID: tpRIF, DisplayName: "RIF", Regimes: []usecase.HeldRegime{{Code: "621", Start: jobPeriod(t, 2026, time.January)}}}}
	closed := []usecase.RecipientTaxpayer{{ID: tpAna, DisplayName: "Ana", Regimes: []usecase.HeldRegime{{Code: "612", Start: jobPeriod(t, 2026, time.January), End: march}}}}
	disabled := enabledSettings(t, false)
	disabled.Enabled = false
	tests := []struct {
		name      string
		now       time.Time
		period    domain.Period
		kind      domain.Kind
		settings  domain.Settings
		taxpayers []usecase.RecipientTaxpayer
	}{
		{"621 only, odd month: due next month", time.Date(2026, time.April, 17, 10, 5, 0, 0, clock.Location()), march, domain.KindDue, enabledSettings(t, false), rifOnly},
		{"regime closed before the period", earlyNow, april, domain.KindEarly, enabledSettings(t, false), closed},
		{"settings turned off since the selection", earlyNow, april, domain.KindEarly, disabled, twoTaxpayers(t)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newJobFixture(t, tc.now)
			f.addUser(t, jobUser1, tc.settings, tc.taxpayers, dev1)
			f.dueFor(jobUser1, tc.period, tc.kind)

			require.NoError(t, f.run(t))

			assert.Empty(t, f.sender.Sends())
			assert.Equal(t, usecase.LogOutcome{Status: usecase.LogSkipped}, f.db.row(t, jobUser1, tc.period, tc.kind).outcome)
		})
	}
}

// §2.4 step 6: in the bimester's second month a 621-only taxpayer is
// listed next to a monthly one; in the first month only the monthly one.
func TestReminderJob_BimonthlyTaxpayerListedInItsMonthOnly(t *testing.T) {
	rif := usecase.RecipientTaxpayer{ID: tpRIF, DisplayName: "RIF", Regimes: []usecase.HeldRegime{{Code: "621", Start: jobPeriod(t, 2026, time.January)}}}
	ana := usecase.RecipientTaxpayer{ID: tpAna, DisplayName: "Ana", Regimes: monthly(t)}
	tests := []struct {
		name     string
		now      time.Time
		period   domain.Period
		wantBody string
	}{
		{"even period (bimester end)", earlyNow, jobPeriod(t, 2026, time.April), "Revisa tu estimado de abril para: Ana, RIF."},
		{"odd period (bimester start)", time.Date(2026, time.April, 14, 10, 5, 0, 0, clock.Location()), jobPeriod(t, 2026, time.March), "Revisa tu estimado de marzo para: Ana."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newJobFixture(t, tc.now)
			f.addUser(t, jobUser1, enabledSettings(t, false), []usecase.RecipientTaxpayer{ana, rif}, dev1)
			f.dueFor(jobUser1, tc.period, domain.KindEarly)

			require.NoError(t, f.run(t))

			sends := f.sender.Sends()
			require.Len(t, sends, 1)
			assert.Equal(t, tc.wantBody, sends[0].Notification.Body)
		})
	}
}

// US-R4 / §2.5: amounts only with show_amounts AND the engine ON, from the
// Estimator; an estimator error never blocks the reminder.
func TestReminderJob_Amounts(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	tests := []struct {
		name        string
		showAmounts bool
		engineOn    bool
		estimateErr error
		wantBody    string
		wantCalls   int
	}{
		{"opted in, engine on", true, true, nil, "Ana: $4,520 estimado · Taller Sur: estimado pendiente. Cifras aproximadas.", 2},
		{"opted in, engine off", true, false, nil, "Revisa tu estimado de abril para: Ana, Taller Sur.", 0},
		{"not opted in", false, true, nil, "Revisa tu estimado de abril para: Ana, Taller Sur.", 0},
		{"estimator down", true, true, errPortDown, "Revisa tu estimado de abril para: Ana, Taller Sur.", 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newJobFixture(t, earlyNow)
			f.flags[usecase.ModuleTaxEngine] = tc.engineOn
			f.addUser(t, jobUser1, enabledSettings(t, tc.showAmounts), twoTaxpayers(t), dev1)
			f.dueFor(jobUser1, april, domain.KindEarly)
			f.estimates.answers[tpAna] = usecase.EstimateSummary{Status: usecase.EstimateComplete, TotalPayableCents: 4_520_00}
			f.estimates.answers[tpTaller] = usecase.EstimateSummary{Status: usecase.EstimateIncomplete, TotalPayableCents: 9_99}
			f.estimates.err = tc.estimateErr

			require.NoError(t, f.run(t))

			sends := f.sender.Sends()
			require.Len(t, sends, 1)
			assert.Equal(t, tc.wantBody, sends[0].Notification.Body)
			assert.Len(t, f.estimates.calls, tc.wantCalls)
			if tc.wantCalls > 0 {
				assert.Equal(t, jobUser1+"|"+tpAna+"|2026-04", f.estimates.calls[0])
			}
			assert.Equal(t, usecase.LogSent, f.db.row(t, jobUser1, april, domain.KindEarly).status)
		})
	}
}

// §2.5: a complete estimate of zero, or an informational one, shows no figure.
func TestReminderJob_AmountsOnlyForAPositiveCompleteTotal(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, true), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.estimates.answers[tpAna] = usecase.EstimateSummary{Status: usecase.EstimateComplete, TotalPayableCents: 0}
	f.estimates.answers[tpTaller] = usecase.EstimateSummary{Status: usecase.EstimateOther, TotalPayableCents: 5_00}

	require.NoError(t, f.run(t))

	sends := f.sender.Sends()
	require.Len(t, sends, 1)
	assert.Equal(t, "Revisa tu estimado de abril para: Ana, Taller Sur.", sends[0].Notification.Body)
}

// US-R5: a pending row whose day passed becomes failed / expired, never sent.
func TestReminderJob_ExpiresYesterdaysPendingRow(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, dueNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindDue)
	f.db.log[logKey(jobUser1, april, domain.KindEarly)] = &jobLogRow{
		id: "log-old", status: usecase.LogPending, attempts: 1, claimedAt: time.Date(2026, time.May, 14, 23, 59, 0, 0, clock.Location()),
	}

	require.NoError(t, f.run(t))

	old := f.db.row(t, jobUser1, april, domain.KindEarly)
	assert.Equal(t, usecase.LogFailed, old.status)
	assert.Equal(t, "expired", old.outcome.ErrorClass)
	assert.Len(t, f.sender.Sends(), 1, "only today's due reminder")
}

// Batches page through due_reminder_users by user id.
func TestReminderJob_PagesThroughDueUsers(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.opts = func(d *usecase.JobDeps) { d.BatchSize = 2 }
	for i, user := range []string{jobUser1, jobUser2, jobUser3} {
		f.addUser(t, user, enabledSettings(t, false), twoTaxpayers(t), []string{dev1, dev2, dev3}[i])
		f.dueFor(user, april, domain.KindEarly)
	}

	require.NoError(t, f.run(t))

	require.Len(t, f.due.calls, 2)
	assert.Empty(t, f.due.calls[0].after)
	assert.Equal(t, jobUser2, f.due.calls[1].after)
	assert.Len(t, f.sender.Sends(), 3)
}

// One user's failure neither stops the others nor hides: the run reports it.
func TestReminderJob_OneUserFailingDoesNotStopTheOthers(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.addUser(t, jobUser2, enabledSettings(t, false), twoTaxpayers(t), dev2)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.dueFor(jobUser2, april, domain.KindEarly)
	f.db.taxpayers[jobUser1] = nil
	f.opts = func(d *usecase.JobDeps) { d.Recipients = failingTaxpayers{jobDB: f.db, user: jobUser1} }

	err := f.run(t)

	require.ErrorIs(t, err, usecase.ErrReminderFailures)
	assert.Len(t, f.sender.Sends(), 1)
}

type failingTaxpayers struct {
	*jobDB
	user string
}

func (f failingTaxpayers) Taxpayers(ctx context.Context) ([]usecase.RecipientTaxpayer, error) {
	if userOf(ctx) == f.user {
		return nil, errPortDown
	}
	return f.jobDB.Taxpayers(ctx)
}

func TestReminderJob_DueReadFailure(t *testing.T) {
	f := newJobFixture(t, earlyNow)
	f.due.err = errPortDown

	require.ErrorIs(t, f.run(t), errPortDown)
}

func TestReminderJob_HolidaysReadFailure(t *testing.T) {
	f := newJobFixture(t, earlyNow)
	f.holidays.err = errPortDown

	require.ErrorIs(t, f.run(t), errPortDown)
	assert.Empty(t, f.due.calls)
}

// A Finish that fails after a send leaves the row pending (the lease retry
// resends with the same collapse id) and is reported.
func TestReminderJob_RecordFailureIsReported(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, false), twoTaxpayers(t), dev1)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.db.finishErr = errPortDown

	require.ErrorIs(t, f.run(t), usecase.ErrReminderFailures)
	assert.Len(t, f.sender.Sends(), 1)
	assert.Equal(t, usecase.LogPending, f.db.row(t, jobUser1, april, domain.KindEarly).status)
}

// blockingSender holds every Send until released, counting how many run at once.
type blockingSender struct {
	release chan struct{}
	started chan struct{}
	running atomic.Int32
	peak    atomic.Int32
	mu      sync.Mutex
}

func (b *blockingSender) Send(ctx context.Context, _ usecase.PushNotification, targets []usecase.PushTarget) ([]usecase.PushDelivery, error) {
	now := b.running.Add(1)
	defer b.running.Add(-1)
	b.mu.Lock()
	b.peak.Store(max(b.peak.Load(), now))
	b.mu.Unlock()
	b.started <- struct{}{}
	select {
	case <-b.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	out := make([]usecase.PushDelivery, len(targets))
	for i, target := range targets {
		out[i] = usecase.PushDelivery{DeviceID: target.DeviceID, Outcome: usecase.PushDelivered}
	}
	return out, nil
}

func manyUsers(t *testing.T, f *jobFixture, n int) {
	t.Helper()
	april := jobPeriod(t, 2026, time.April)
	for i := range n {
		user := uuid.NewSHA1(uuid.NameSpaceOID, []byte{byte(i)}).String()
		f.addUser(t, user, enabledSettings(t, false), twoTaxpayers(t), uuid.NewSHA1(uuid.NameSpaceURL, []byte{byte(i)}).String())
		f.dueFor(user, april, domain.KindEarly)
	}
}

// §2.4 step 3: bounded concurrency (8 users at once at most).
func TestReminderJob_BoundedConcurrency(t *testing.T) {
	defer goleak.VerifyNone(t)
	f := newJobFixture(t, earlyNow)
	sender := &blockingSender{release: make(chan struct{}), started: make(chan struct{}, 64)}
	f.push = sender
	manyUsers(t, f, 20)
	job := f.job(t)

	done := make(chan error, 1)
	go func() { done <- job.Run(t.Context()) }()
	for range usecase.DefaultJobConcurrency {
		<-sender.started
	}
	select {
	case <-sender.started:
		t.Fatal("a ninth send started while eight were in flight")
	case <-time.After(50 * time.Millisecond):
	}
	close(sender.release)
	for range 20 - usecase.DefaultJobConcurrency {
		<-sender.started
	}
	require.NoError(t, <-done)
	assert.Equal(t, int32(usecase.DefaultJobConcurrency), sender.peak.Load())
}

// The Runner cancels a run on timeout, lost leadership or shutdown: Run
// returns the context's error promptly and leaves no goroutine behind; the
// claimed rows stay pending for the lease retry.
func TestReminderJob_CancellationLeavesNoGoroutine(t *testing.T) {
	defer goleak.VerifyNone(t)
	f := newJobFixture(t, earlyNow)
	sender := &blockingSender{release: make(chan struct{}), started: make(chan struct{}, 64)}
	f.push = sender
	manyUsers(t, f, 20)
	job := f.job(t)
	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan error, 1)
	go func() { done <- job.Run(ctx) }()
	for range usecase.DefaultJobConcurrency {
		<-sender.started
	}
	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
	f.db.mu.Lock()
	defer f.db.mu.Unlock()
	for _, row := range f.db.log {
		assert.Equal(t, usecase.LogPending, row.status, "a cancelled send is retried after the lease")
		assert.Zero(t, row.finishedAt, "never recorded")
	}
}

// §9: logs carry counts, user ids and periods only: never a token, a
// display name or an amount.
func TestReminderJob_LogsCarryNoSecretsOrPII(t *testing.T) {
	april := jobPeriod(t, 2026, time.April)
	f := newJobFixture(t, earlyNow)
	f.addUser(t, jobUser1, enabledSettings(t, true), twoTaxpayers(t), dev1, dev2)
	f.dueFor(jobUser1, april, domain.KindEarly)
	f.estimates.answers[tpAna] = usecase.EstimateSummary{Status: usecase.EstimateComplete, TotalPayableCents: 4_520_00}
	f.estimates.err = nil
	f.sender.SetOutcome(uuid.MustParse(dev2), usecase.PushTokenInvalid, "unregistered")

	require.NoError(t, f.run(t))

	logs := f.logs.String()
	require.NotEmpty(t, logs)
	for _, secret := range []string{f.tokens[dev1], f.tokens[dev2], "Ana", "Taller", "4,520", "452000", "estimado", tpAna} {
		assert.NotContains(t, logs, secret)
	}
	assert.Contains(t, logs, `"sent":1`)
	assert.Contains(t, logs, `"devices_unregistered":1`)
}

func TestNewReminderJobValidatesDeps(t *testing.T) {
	f := newJobFixture(t, earlyNow)
	tests := []struct {
		name   string
		mutate func(*usecase.JobDeps)
	}{
		{"System", func(d *usecase.JobDeps) { d.System = nil }},
		{"Units", func(d *usecase.JobDeps) { d.Units = nil }},
		{"Due", func(d *usecase.JobDeps) { d.Due = nil }},
		{"Log", func(d *usecase.JobDeps) { d.Log = nil }},
		{"Settings", func(d *usecase.JobDeps) { d.Settings = nil }},
		{"Recipients", func(d *usecase.JobDeps) { d.Recipients = nil }},
		{"Opener", func(d *usecase.JobDeps) { d.Opener = nil }},
		{"Push", func(d *usecase.JobDeps) { d.Push = nil }},
		{"Estimates", func(d *usecase.JobDeps) { d.Estimates = nil }},
		{"Flags", func(d *usecase.JobDeps) { d.Flags = nil }},
		{"Holidays", func(d *usecase.JobDeps) { d.Holidays = nil }},
		{"Alerts", func(d *usecase.JobDeps) { d.Alerts = nil }},
		{"Clock", func(d *usecase.JobDeps) { d.Clock = nil }},
		{"Logger", func(d *usecase.JobDeps) { d.Logger = nil }},
		{"short collapse key", func(d *usecase.JobDeps) { d.CollapseKey = []byte("short") }},
		{"concurrency", func(d *usecase.JobDeps) { d.Concurrency = 0 }},
		{"batch size", func(d *usecase.JobDeps) { d.BatchSize = 0 }},
		{"batch size over the function's", func(d *usecase.JobDeps) { d.BatchSize = 1001 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := f.deps()
			tc.mutate(&d)
			_, err := usecase.NewReminderJob(d)
			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}
