package fiscal_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// seededHolidays is the M9-03 seed (research §A.4.1, read 2026-10-15): the
// CFF 12 ¶1 days of 2026 and 2027 plus the RMF 2.1.6 Holy Thursday and
// Friday, every one with moves_due_date = true.
var seededHolidays = []string{
	"2026-01-01", "2026-02-02", "2026-03-16", "2026-04-02", "2026-04-03",
	"2026-05-01", "2026-05-05", "2026-09-16", "2026-11-16", "2026-12-25",
	"2027-01-01", "2027-02-01", "2027-03-15", "2027-03-25", "2027-03-26",
	"2027-05-01", "2027-05-05", "2027-09-16", "2027-11-15", "2027-12-25",
}

func holidaySet(t *testing.T, years []int, days ...string) fiscal.HolidaySet {
	t.Helper()
	dates := make([]fiscal.Date, 0, len(days))
	for _, raw := range days {
		dates = append(dates, date(t, raw))
	}
	h, err := fiscal.NewHolidaySet(years, dates)
	require.NoError(t, err)
	return h
}

func seeded(t *testing.T) fiscal.HolidaySet {
	t.Helper()
	return holidaySet(t, []int{2026, 2027}, seededHolidays...)
}

// TestMonthlyDueDateGolden pins the rule (MODULE-reminders §2.1, ADR-0007
// Amendment 1) on every period whose due date falls in the seeded years,
// 2025-12..2027-11, plus 2027-12 (due in unseeded 2028: weekends only).
// Expected values are research §A.4.1's: seven weekend moves, no holiday
// on a day 17, and the four Fridays kept (Friday extension not applied,
// OQ-R1).
func TestMonthlyDueDateGolden(t *testing.T) {
	type entry struct {
		year    int
		month   time.Month
		nominal string
		due     string
		reason  fiscal.MovedReason
	}
	weekend := fiscal.MovedWeekend
	golden := []entry{
		{2025, time.December, "2026-01-17", "2026-01-19", weekend},
		{2026, time.January, "2026-02-17", "2026-02-17", fiscal.MovedNone},
		{2026, time.February, "2026-03-17", "2026-03-17", fiscal.MovedNone},
		{2026, time.March, "2026-04-17", "2026-04-17", fiscal.MovedNone}, // Friday kept
		{2026, time.April, "2026-05-17", "2026-05-18", weekend},
		{2026, time.May, "2026-06-17", "2026-06-17", fiscal.MovedNone},
		{2026, time.June, "2026-07-17", "2026-07-17", fiscal.MovedNone}, // Friday kept
		{2026, time.July, "2026-08-17", "2026-08-17", fiscal.MovedNone},
		{2026, time.August, "2026-09-17", "2026-09-17", fiscal.MovedNone},
		{2026, time.September, "2026-10-17", "2026-10-19", weekend},
		{2026, time.October, "2026-11-17", "2026-11-17", fiscal.MovedNone},
		{2026, time.November, "2026-12-17", "2026-12-17", fiscal.MovedNone},
		{2026, time.December, "2027-01-17", "2027-01-18", weekend},
		{2027, time.January, "2027-02-17", "2027-02-17", fiscal.MovedNone},
		{2027, time.February, "2027-03-17", "2027-03-17", fiscal.MovedNone},
		{2027, time.March, "2027-04-17", "2027-04-19", weekend},
		{2027, time.April, "2027-05-17", "2027-05-17", fiscal.MovedNone},
		{2027, time.May, "2027-06-17", "2027-06-17", fiscal.MovedNone},
		{2027, time.June, "2027-07-17", "2027-07-19", weekend},
		{2027, time.July, "2027-08-17", "2027-08-17", fiscal.MovedNone},
		{2027, time.August, "2027-09-17", "2027-09-17", fiscal.MovedNone}, // Friday kept
		{2027, time.September, "2027-10-17", "2027-10-18", weekend},
		{2027, time.October, "2027-11-17", "2027-11-17", fiscal.MovedNone},
		{2027, time.November, "2027-12-17", "2027-12-17", fiscal.MovedNone}, // Friday kept
		{2027, time.December, "2028-01-17", "2028-01-17", fiscal.MovedNone}, // unseeded year: weekends only
	}
	h := seeded(t)
	for _, g := range golden {
		t.Run(time.Date(g.year, g.month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01"), func(t *testing.T) {
			nominal, due, reason, err := fiscal.MonthlyDueDate(g.year, g.month, false, h)

			require.NoError(t, err)
			assert.Equal(t, g.nominal, nominal.String(), "nominal")
			assert.Equal(t, g.due, due.String(), "due")
			assert.Equal(t, g.reason, reason, "moved reason")
		})
	}
}

// MODULE-tax-engine §2.10: a 621-only taxpayer pays after the bimester, so
// both of its months share the second month's due date.
func TestMonthlyDueDateBimonthly(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month time.Month
		want  string
	}{
		{"first month of Sep-Oct", 2026, time.September, "2026-11-17"},
		{"second month of Sep-Oct", 2026, time.October, "2026-11-17"},
		{"first month of Nov-Dec crosses the year", 2026, time.November, "2027-01-18"},
		{"second month of Nov-Dec crosses the year", 2026, time.December, "2027-01-18"},
		{"first month of Jan-Feb", 2027, time.January, "2027-03-17"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, due, _, err := fiscal.MonthlyDueDate(tt.year, tt.month, true, seeded(t))

			require.NoError(t, err)
			assert.Equal(t, tt.want, due.String())
		})
	}
}

// One rule for "this taxpayer pays by bimester" (MODULE-tax-engine §2.10):
// every regime held in the period is 621. The estimate's due_date and the
// reminders both read it here, so they cannot drift (M9-13 gate).
func TestPaysBimonthly(t *testing.T) {
	tests := []struct {
		name  string
		codes []fiscal.RegimeCode
		want  bool
	}{
		{"no regime held", nil, false},
		{"621 only", []fiscal.RegimeCode{fiscal.RegimeRIF}, true},
		{"621 twice", []fiscal.RegimeCode{fiscal.RegimeRIF, fiscal.RegimeRIF}, true},
		{"621 with a monthly regime", []fiscal.RegimeCode{fiscal.RegimeRIF, fiscal.RegimeBusiness}, false},
		{"monthly only", []fiscal.RegimeCode{fiscal.RegimeRESICO}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fiscal.PaysBimonthly(tt.codes))
		})
	}
	assert.Equal(t, fiscal.RegimeCode("621"), fiscal.RegimeRIF)
}

func TestMonthlyDueDateRejectsABadPeriod(t *testing.T) {
	for _, m := range []time.Month{0, 13} {
		_, _, _, err := fiscal.MonthlyDueDate(2026, m, false, fiscal.HolidaySet{})

		require.ErrorIs(t, err, fiscal.ErrInvalid)
	}
	_, _, _, err := fiscal.MonthlyDueDate(0, time.January, false, fiscal.HolidaySet{})
	require.ErrorIs(t, err, fiscal.ErrInvalid)
}

// CFF 12 ¶1, ¶5: the move skips Saturdays, Sundays and holidays; the reason
// is the kind of the nominal day (the skipped days after it may differ).
func TestNextBusinessDay(t *testing.T) {
	tests := []struct {
		name     string
		holidays []string
		day      string
		want     string
		reason   fiscal.MovedReason
	}{
		{"a business day stays", nil, "2026-04-17", "2026-04-17", fiscal.MovedNone},
		{"Saturday moves to Monday", nil, "2026-10-17", "2026-10-19", fiscal.MovedWeekend},
		{"Sunday moves to Monday", nil, "2026-05-17", "2026-05-18", fiscal.MovedWeekend},
		{"a holiday moves to the next day", []string{"2026-11-17"}, "2026-11-17", "2026-11-18", fiscal.MovedHoliday},
		{"a weekend then a holiday Monday", []string{"2026-10-19"}, "2026-10-17", "2026-10-20", fiscal.MovedWeekend},
		{"a holiday Friday then the weekend", []string{"2026-04-17"}, "2026-04-17", "2026-04-20", fiscal.MovedHoliday},
		{"a holiday Saturday is a weekend", []string{"2027-05-01"}, "2027-05-01", "2027-05-03", fiscal.MovedWeekend},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := holidaySet(t, []int{2026, 2027}, tt.holidays...)

			got, reason, err := fiscal.NextBusinessDay(date(t, tt.day), h)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
			assert.Equal(t, tt.reason, reason)
		})
	}
}

// The annual deadlines of fiscal years 2025 and 2026 (LISR 150, 76 fr. V,
// 86 fr. III, read off the parameter set), moved by the same rule.
func TestNextBusinessDayAnnualDeadlinesGolden(t *testing.T) {
	tests := []struct {
		legal  string
		want   string
		reason fiscal.MovedReason
	}{
		{"2026-04-30", "2026-04-30", fiscal.MovedNone},    // FY 2025 PF
		{"2026-03-31", "2026-03-31", fiscal.MovedNone},    // FY 2025 PM
		{"2026-02-15", "2026-02-16", fiscal.MovedWeekend}, // FY 2025 nonprofit informative
		{"2027-04-30", "2027-04-30", fiscal.MovedNone},    // FY 2026 PF (a Friday, kept)
		{"2027-03-31", "2027-03-31", fiscal.MovedNone},    // FY 2026 PM
		{"2027-02-15", "2027-02-15", fiscal.MovedNone},    // FY 2026 nonprofit informative
	}
	h := seeded(t)
	for _, tt := range tests {
		t.Run(tt.legal, func(t *testing.T) {
			got, reason, err := fiscal.NextBusinessDay(date(t, tt.legal), h)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
			assert.Equal(t, tt.reason, reason)
		})
	}
}

// A run of non-business days longer than the bound is a calendar bug,
// never an endless loop or a silently wrong date.
func TestNextBusinessDayIsBounded(t *testing.T) {
	days := make([]string, 0, 31)
	for d := 1; d <= 31; d++ {
		days = append(days, time.Date(2026, time.December, d, 0, 0, 0, 0, time.UTC).Format(time.DateOnly))
	}
	h := holidaySet(t, []int{2026}, days...)

	_, _, err := fiscal.NextBusinessDay(date(t, "2026-12-01"), h)
	require.ErrorIs(t, err, fiscal.ErrNoBusinessDay)

	_, _, _, err = fiscal.MonthlyDueDate(2026, time.November, false, h)
	require.ErrorIs(t, err, fiscal.ErrNoBusinessDay)
}

func TestNextBusinessDayRejectsTheZeroDate(t *testing.T) {
	_, _, err := fiscal.NextBusinessDay(fiscal.Date{}, fiscal.HolidaySet{})

	require.ErrorIs(t, err, fiscal.ErrInvalid)
}

// Outside a covered year the set answers weekends only; the caller decides
// whether that is acceptable (MODULE-reminders §2.1).
func TestHolidaySetCoverage(t *testing.T) {
	h := seeded(t)

	assert.True(t, h.Covers(2026))
	assert.True(t, h.Covers(2027))
	assert.False(t, h.Covers(2028))
	assert.False(t, fiscal.HolidaySet{}.Covers(2026), "the zero set covers nothing")
	assert.True(t, h.IsHoliday(date(t, "2026-09-16")))
	assert.False(t, h.IsHoliday(date(t, "2026-09-17")))
	assert.Equal(t, []int{2026, 2027}, h.Years())
}

func TestNewHolidaySetRejects(t *testing.T) {
	tests := []struct {
		name  string
		years []int
		days  []fiscal.Date
	}{
		{"a day outside the covered years", []int{2026}, []fiscal.Date{date(t, "2027-01-01")}},
		{"the zero date", []int{2026}, []fiscal.Date{{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fiscal.NewHolidaySet(tt.years, tt.days)

			require.ErrorIs(t, err, fiscal.ErrInvalid)
		})
	}
}

// The set is a value: changing the caller's slices after building it does
// not change what it answers.
func TestHolidaySetIsImmutable(t *testing.T) {
	years := []int{2026}
	days := []fiscal.Date{date(t, "2026-09-16")}
	h, err := fiscal.NewHolidaySet(years, days)
	require.NoError(t, err)

	years[0], days[0] = 2030, date(t, "2026-09-17")

	assert.True(t, h.Covers(2026))
	assert.True(t, h.IsHoliday(date(t, "2026-09-16")))
	assert.False(t, h.IsHoliday(date(t, "2026-09-17")))
}

func TestDateArithmetic(t *testing.T) {
	d := date(t, "2026-12-31")

	assert.Equal(t, "2027-01-01", d.AddDays(1).String())
	assert.Equal(t, "2026-12-30", d.AddDays(-1).String())
	assert.Equal(t, time.Thursday, d.Weekday())
	assert.True(t, fiscal.Date{}.AddDays(1).IsZero(), "the zero date stays zero")
}
