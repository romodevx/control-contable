package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const userA = "0199d4a2-0000-7000-8000-00000000c41a"

var errPortDown = errors.New("fake: port down")

// userKey carries the fake unit of work's user.
type userKey struct{}

// fakeTx runs work inline as the user, read-only only.
type fakeTx struct{ modes []string }

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "rw")
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "ro")
	return work(context.WithValue(ctx, userKey{}, userID))
}

// fakeHolidays is the in-memory holidays table: the covered years and the
// days that move a due date. It records the user of the unit of work it
// was read in.
type fakeHolidays struct {
	years []int
	days  []string
	err   error
	users []string
}

func (f *fakeHolidays) Holidays(ctx context.Context) (fiscal.HolidaySet, error) {
	user, _ := ctx.Value(userKey{}).(string)
	f.users = append(f.users, user)
	if f.err != nil {
		return fiscal.HolidaySet{}, f.err
	}
	days := make([]fiscal.Date, 0, len(f.days))
	for _, raw := range f.days {
		d, err := fiscal.ParseDate(raw)
		if err != nil {
			return fiscal.HolidaySet{}, err
		}
		days = append(days, d)
	}
	return fiscal.NewHolidaySet(f.years, days)
}

// fakeAlerts records the unseeded years reported.
type fakeAlerts struct{ unseeded []int }

func (f *fakeAlerts) CalendarUnseeded(_ context.Context, year int) {
	f.unseeded = append(f.unseeded, year)
}

type fixture struct {
	tx       *fakeTx
	holidays *fakeHolidays
	alerts   *fakeAlerts
}

func newFixture(years ...int) *fixture {
	return &fixture{
		tx:       &fakeTx{},
		holidays: &fakeHolidays{years: years, days: []string{"2026-09-16", "2026-11-16", "2027-01-01"}},
		alerts:   &fakeAlerts{},
	}
}

func (f *fixture) calendar(t *testing.T) *usecase.Calendar {
	t.Helper()
	c, err := usecase.NewCalendar(usecase.CalendarDeps{Tx: f.tx, Holidays: f.holidays, Alerts: f.alerts})
	require.NoError(t, err)
	return c
}

// MODULE-reminders §2.1: the periods year-01..year-12, in order, December
// due in January of year+1, every entry subject to confirmation (OQ-R1).
func TestCalendarYear(t *testing.T) {
	f := newFixture(2026, 2027)

	got, err := f.calendar(t).Year(t.Context(), userA, 2026)

	require.NoError(t, err)
	assert.Equal(t, 2026, got.Year)
	require.Len(t, got.Entries, 12)
	want := []struct{ period, nominal, due string }{
		{"2026-01", "2026-02-17", "2026-02-17"}, {"2026-02", "2026-03-17", "2026-03-17"},
		{"2026-03", "2026-04-17", "2026-04-17"}, {"2026-04", "2026-05-17", "2026-05-18"},
		{"2026-05", "2026-06-17", "2026-06-17"}, {"2026-06", "2026-07-17", "2026-07-17"},
		{"2026-07", "2026-08-17", "2026-08-17"}, {"2026-08", "2026-09-17", "2026-09-17"},
		{"2026-09", "2026-10-17", "2026-10-19"}, {"2026-10", "2026-11-17", "2026-11-17"},
		{"2026-11", "2026-12-17", "2026-12-17"}, {"2026-12", "2027-01-17", "2027-01-18"},
	}
	for i, w := range want {
		e := got.Entries[i]
		assert.Equal(t, w.period, e.Period, "period %d", i)
		assert.Equal(t, w.nominal, e.Nominal.String(), w.period)
		assert.Equal(t, w.due, e.Due.String(), w.period)
		assert.True(t, e.SubjectToConfirmation, w.period)
	}
	assert.Equal(t, fiscal.MovedWeekend, got.Entries[8].MovedReason)
	assert.Equal(t, fiscal.MovedNone, got.Entries[0].MovedReason)
	assert.Equal(t, []string{"ro"}, f.tx.modes, "one read-only unit of work")
	assert.Equal(t, []string{userA}, f.holidays.users, "holidays read once, inside it")
	assert.Empty(t, f.alerts.unseeded, "year+1 is seeded")
}

// US-R3: a holiday on a day 17 moves the date with moved_reason holiday.
func TestCalendarYearHolidayMove(t *testing.T) {
	f := newFixture(2026, 2027)
	f.holidays.days = []string{"2026-11-17"}

	got, err := f.calendar(t).Year(t.Context(), userA, 2026)

	require.NoError(t, err)
	october := got.Entries[9]
	assert.Equal(t, "2026-11-18", october.Due.String())
	assert.Equal(t, fiscal.MovedHoliday, october.MovedReason)
}

// §2.1: an unseeded year is 422 calendar_unavailable on year.
func TestCalendarYearUnseeded(t *testing.T) {
	f := newFixture(2026, 2027)

	_, err := f.calendar(t).Year(t.Context(), userA, 2028)

	var ae *apperr.Error
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, apperr.CodeBusinessRuleViolation, ae.Code)
	require.Len(t, ae.Fields, 1)
	assert.Equal(t, apperr.FieldError{Field: "year", Code: usecase.DetailCalendarUnavailable, Message: "Aún no tenemos el calendario de ese año."}, ae.Fields[0])
	assert.Empty(t, f.alerts.unseeded)
}

// §2.1, US-R3: year seeded, year+1 not: December is weekends only and the
// reminders_calendar_unseeded alert fires for year+1.
func TestCalendarYearNextYearUnseeded(t *testing.T) {
	f := newFixture(2026, 2027)
	f.holidays.days = append(f.holidays.days, "2027-12-17")

	got, err := f.calendar(t).Year(t.Context(), userA, 2027)

	require.NoError(t, err)
	december := got.Entries[11]
	assert.Equal(t, "2028-01-17", december.Due.String())
	assert.Equal(t, fiscal.MovedNone, december.MovedReason)
	assert.Equal(t, "2027-12-20", got.Entries[10].Due.String(), "the covered year still reads its holidays")
	assert.Equal(t, []int{2028}, f.alerts.unseeded)
}

// Holiday data that never reaches a business day is a server fault, never
// a wrong date.
func TestCalendarYearBadHolidayData(t *testing.T) {
	f := newFixture(2026, 2027)
	f.holidays.days = nil
	for d := 17; d <= 28; d++ {
		f.holidays.days = append(f.holidays.days, fmt.Sprintf("2026-02-%02d", d))
	}
	for d := 1; d <= 6; d++ {
		f.holidays.days = append(f.holidays.days, fmt.Sprintf("2026-03-%02d", d))
	}

	_, err := f.calendar(t).Year(t.Context(), userA, 2026)

	require.ErrorIs(t, err, fiscal.ErrNoBusinessDay)
}

func TestCalendarYearHolidaysDown(t *testing.T) {
	f := newFixture(2026)
	f.holidays.err = errPortDown

	_, err := f.calendar(t).Year(t.Context(), userA, 2026)

	require.ErrorIs(t, err, errPortDown)
}

func TestNewCalendarRequiresEveryPort(t *testing.T) {
	for name, drop := range map[string]func(*usecase.CalendarDeps){
		"tx":       func(d *usecase.CalendarDeps) { d.Tx = nil },
		"holidays": func(d *usecase.CalendarDeps) { d.Holidays = nil },
		"alerts":   func(d *usecase.CalendarDeps) { d.Alerts = nil },
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(2026)
			deps := usecase.CalendarDeps{Tx: f.tx, Holidays: f.holidays, Alerts: f.alerts}
			drop(&deps)

			_, err := usecase.NewCalendar(deps)

			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}
