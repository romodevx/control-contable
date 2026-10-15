// Package holidays is the tax engine's reader of the global holidays
// tables (migration 20261021100000, M9-03): the seeded years and the days
// that move a due date, as the fiscal kernel's HolidaySet. It holds no
// pool: it runs on the unit of work in ctx (cc_app, SELECT only).
package holidays

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The two reads: both tables are a handful of rows per seeded year.
const (
	yearsQuery = `SELECT year FROM holiday_years ORDER BY year`
	daysQuery  = `SELECT day FROM holidays WHERE moves_due_date ORDER BY day`
)

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
	var years []int
	if err := db.Raw(yearsQuery).Scan(&years).Error; err != nil {
		return fiscal.HolidaySet{}, fmt.Errorf("taxengine holidays: years: %w", err)
	}
	var stored []time.Time
	if err := db.Raw(daysQuery).Scan(&stored).Error; err != nil {
		return fiscal.HolidaySet{}, fmt.Errorf("taxengine holidays: days: %w", err)
	}
	days := make([]fiscal.Date, 0, len(stored))
	for _, t := range stored {
		d, err := fiscal.ParseDate(t.Format(time.DateOnly))
		if err != nil {
			return fiscal.HolidaySet{}, fmt.Errorf("taxengine holidays: day %s: %w", t, err)
		}
		days = append(days, d)
	}
	set, err := fiscal.NewHolidaySet(years, days)
	if err != nil {
		return fiscal.HolidaySet{}, fmt.Errorf("taxengine holidays: %w", err)
	}
	return set, nil
}
