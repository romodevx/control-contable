package taxengine_test

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// seededHolidays is the holidays calendar the M9-03 migration seeds
// (research §A.4.1, read 2026-10-15): 2026 and 2027, every day moving a due
// date. The goldens run on it, so a golden that moved when noHolidays was
// replaced would fail here and force an EngineVersion bump (ADR-0007
// Amendment 1). None moves: no 2026-2027 day 17 or annual deadline meets a
// holiday.
type seededHolidays struct{}

var seededHolidayDays = []string{
	"2026-01-01", "2026-02-02", "2026-03-16", "2026-04-02", "2026-04-03",
	"2026-05-01", "2026-05-05", "2026-09-16", "2026-11-16", "2026-12-25",
	"2027-01-01", "2027-02-01", "2027-03-15", "2027-03-25", "2027-03-26",
	"2027-05-01", "2027-05-05", "2027-09-16", "2027-11-15", "2027-12-25",
}

func (seededHolidays) Holidays(context.Context) (fiscal.HolidaySet, error) {
	days := make([]fiscal.Date, 0, len(seededHolidayDays))
	for _, raw := range seededHolidayDays {
		d, err := fiscal.ParseDate(raw)
		if err != nil {
			return fiscal.HolidaySet{}, err
		}
		days = append(days, d)
	}
	return fiscal.NewHolidaySet([]int{2026, 2027}, days)
}
