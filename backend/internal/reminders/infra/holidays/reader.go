// Package holidays is the reminders module's reader of the global holidays
// tables (migration 20261021100000, M9-03): the seeded years and the days
// that move a due date, as the fiscal kernel's HolidaySet. It holds no
// pool: it runs on the unit of work in ctx (cc_app for the calendar,
// cc_worker for the job; SELECT only). The tax engine keeps its own reader
// of the same tables: adapters stay consumer-side (ADR-0007 Decision 6,
// Amendment 1), so either module is removable without the other.
package holidays

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// holidaysQuery is the one read, a single statement: each seeded year with
// the days of it that move a due date, a year with none as one row with a
// NULL day. Both tables are a handful of rows per seeded year.
const holidaysQuery = `SELECT y.year, h.day
FROM holiday_years y
LEFT JOIN holidays h ON h.year = y.year AND h.moves_due_date
ORDER BY y.year, h.day`

// holidayRow is one row of holidaysQuery.
type holidayRow struct {
	Year int        `gorm:"column:year"`
	Day  *time.Time `gorm:"column:day"`
}

// Reader implements usecase.HolidayCalendar.
type Reader struct{}

var _ usecase.HolidayCalendar = Reader{}

// NewReader returns a Reader.
func NewReader() Reader { return Reader{} }

// Holidays implements usecase.HolidayCalendar.
func (Reader) Holidays(ctx context.Context) (fiscal.HolidaySet, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return fiscal.HolidaySet{}, err
	}
	var rows []holidayRow
	if err := db.Raw(holidaysQuery).Scan(&rows).Error; err != nil {
		return fiscal.HolidaySet{}, fmt.Errorf("reminders holidays: %w", err)
	}
	years := make([]int, 0, len(rows))
	days := make([]fiscal.Date, 0, len(rows))
	for _, r := range rows {
		years = append(years, r.Year)
		if r.Day == nil {
			continue
		}
		d, err := fiscal.ParseDate(r.Day.Format(time.DateOnly))
		if err != nil {
			return fiscal.HolidaySet{}, fmt.Errorf("reminders holidays: day %s: %w", r.Day, err)
		}
		days = append(days, d)
	}
	set, err := fiscal.NewHolidaySet(years, days)
	if err != nil {
		return fiscal.HolidaySet{}, fmt.Errorf("reminders holidays: %w", err)
	}
	return set, nil
}
