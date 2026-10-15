//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Holidays migration (M9-03): the global non-business days that move the day-17 due
// date (CFF art. 12, MODULE-reminders §2.1, research tax-regimes-mx-2026.md §A.4.1),
// seeded for 2026 and 2027, and the years the seed covers.

const (
	holidaysVersion = 20261021100000

	basisCFF = "cff_12"
	basisRMF = "rmf_2_1_6"
)

// seededHolidayYears are the years research §A.4.1 lists.
var seededHolidayYears = []int{2026, 2027}

// nthWeekday is the nth (1-based) given weekday of month in year.
func nthWeekday(year int, month time.Month, weekday time.Weekday, nth int) time.Time {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	offset := (int(weekday) - int(first.Weekday()) + 7) % 7
	return first.AddDate(0, 0, offset+7*(nth-1))
}

// cffHolidays are CFF art. 12 ¶1's days of year, computed from the article's words
// (the 1 December of a presidential transition falls in 2030, out of the seed).
func cffHolidays(year int) []string {
	day := func(m time.Month, d int) time.Time { return time.Date(year, m, d, 0, 0, 0, 0, time.UTC) }
	days := []time.Time{
		day(time.January, 1),
		nthWeekday(year, time.February, time.Monday, 1),
		nthWeekday(year, time.March, time.Monday, 3),
		day(time.May, 1),
		day(time.May, 5),
		day(time.September, 16),
		nthWeekday(year, time.November, time.Monday, 3),
		day(time.December, 25),
	}
	out := make([]string, 0, len(days))
	for _, d := range days {
		out = append(out, d.Format(time.DateOnly))
	}
	return out
}

// easter is Gregorian Easter Sunday (anonymous Gregorian algorithm, Meeus/Jones/Butcher).
func easter(year int) time.Time {
	a, b, c := year%19, year/100, year%100
	d, e := b/4, b%4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i, k := c/4, c%4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// holyThursdayFriday are RMF rule 2.1.6's Holy Thursday and Friday of year.
func holyThursdayFriday(year int) []string {
	sunday := easter(year)
	return []string{sunday.AddDate(0, 0, -3).Format(time.DateOnly), sunday.AddDate(0, 0, -2).Format(time.DateOnly)}
}

func TestHolidays_SeedCoversTheResearchYears(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	assert.Equal(t, []string{"2026", "2027"}, queryStrings(t, owner, "SELECT year::text FROM holiday_years ORDER BY year"))
	for _, year := range seededHolidayYears {
		assert.Equal(t, cffHolidays(year), queryStrings(t, owner,
			"SELECT day::text FROM holidays WHERE basis = $1 AND year = $2 ORDER BY day", basisCFF, year),
			"CFF 12 ¶1 days of %d", year)
		assert.Equal(t, holyThursdayFriday(year), queryStrings(t, owner,
			"SELECT day::text FROM holidays WHERE basis = $1 AND year = $2 ORDER BY day", basisRMF, year),
			"RMF 2.1.6 Holy Thursday and Friday of %d", year)
	}

	var rows, moving, sourced int
	require.NoError(t, owner.QueryRowContext(t.Context(), `
SELECT count(*), count(*) FILTER (WHERE moves_due_date),
       count(*) FILTER (WHERE source_url LIKE 'https://%' AND verified_on = DATE '2026-10-15')
FROM holidays`).Scan(&rows, &moving, &sourced))
	assert.Equal(t, 20, rows, "the 20 rows of research §A.4.1")
	assert.Equal(t, rows, moving, "every seeded day moves the due date (Holy Thursday/Friday ⚠️ per the spec)")
	assert.Equal(t, rows, sourced, "every row names its https source and the day it was read")

	assert.Empty(t, queryStrings(t, owner, "SELECT day::text FROM holidays WHERE extract(day FROM day) = 17"),
		"no 2026-2027 day 17 meets a holiday (research §A.4.1): the noHolidays swap changes no estimate date")
}

// A global catalog: every runtime role that computes a due date reads it; nobody writes it.
func TestHolidays_ReadOnlyCatalogForAppAndWorker(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			conn := db.Connect(t, role)
			for _, table := range []string{"holidays", "holiday_years"} {
				var n int
				require.NoError(t, pgtest.WithoutUser(t.Context(), conn, func(tx *sql.Tx) error {
					return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table).Scan(&n)
				}))
				assert.Positive(t, n, "%s reads %s without a user context", role, table)
			}
			for _, write := range []string{
				"INSERT INTO holiday_years (year, seeded_on, source_note) VALUES (2028, DATE '2026-10-15', 'x')",
				"UPDATE holidays SET moves_due_date = false",
				"DELETE FROM holidays",
				"DELETE FROM holiday_years",
			} {
				_, err := conn.ExecContext(t.Context(), write)
				requireSQLState(t, err, sqlstateInsufficientPrivilege)
			}
		})
	}
	t.Run("cc_webhook reads nothing", func(t *testing.T) {
		_, err := db.Connect(t, pgtest.RoleWebhook).ExecContext(t.Context(), "SELECT 1 FROM holidays")
		requireSQLState(t, err, sqlstateInsufficientPrivilege)
	})
}

// The constraints keep a later seed honest: a holiday belongs to a seeded year, cites
// a known basis and an https source, and its year is derived, never written.
func TestHolidays_ConstraintsRejectMalformedRows(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	const insert = `INSERT INTO holidays (day, name, basis, moves_due_date, source_url, verified_on)
VALUES ($1::date, 'Prueba', $2, true, $3, DATE '2026-10-15')`
	const cff = "https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf"

	_, err := owner.ExecContext(t.Context(), insert, "2028-01-01", basisCFF, cff)
	requireSQLState(t, err, sqlstateForeignKeyViolation)
	_, err = owner.ExecContext(t.Context(), insert, "2026-06-01", "decreto", cff)
	requireSQLState(t, err, sqlstateCheckViolation)
	_, err = owner.ExecContext(t.Context(), insert, "2026-06-01", basisCFF, "http://example.com/cff.pdf")
	requireSQLState(t, err, sqlstateCheckViolation)
	_, err = owner.ExecContext(t.Context(), `INSERT INTO holidays (day, year, name, basis, moves_due_date, source_url, verified_on)
VALUES (DATE '2026-06-01', 2026, 'Prueba', 'cff_12', true, $1, DATE '2026-10-15')`, cff)
	requireSQLState(t, err, sqlstateGeneratedAlways)
	_, err = owner.ExecContext(t.Context(),
		"INSERT INTO holiday_years (year, seeded_on, source_note) VALUES (2028, DATE '2026-10-15', '')")
	requireSQLState(t, err, sqlstateCheckViolation)
	_, err = owner.ExecContext(t.Context(), "DELETE FROM holiday_years WHERE year = 2026")
	requireSQLState(t, err, sqlstateForeignKeyViolation)

	_, err = owner.ExecContext(t.Context(), insert, "2026-06-01", basisCFF, cff)
	require.NoError(t, err, "a well-formed row of a seeded year is accepted")
}

func TestHolidaysMigration_DownDropsTheCatalog(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), holidaysVersion-1)
	require.NoError(t, err)
	assert.Empty(t, queryStrings(t, owner,
		"SELECT relname::text FROM pg_class WHERE relname IN ('holidays', 'holiday_years')"))

	require.NoError(t, migrations.Up(t.Context(), owner))
	assert.Len(t, queryStrings(t, owner, "SELECT day::text FROM holidays"), 20)
}
