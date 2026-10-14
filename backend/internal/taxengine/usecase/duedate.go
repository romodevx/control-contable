package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

const (
	// dueDay is the day of the month monthly payments are due (LISR 14,
	// 106, 113-E, 116; LIVA 5-D).
	dueDay = 17
	// maxDueShift bounds the move to the next business day: no run of
	// weekends and holidays is this long, so reaching it is a calendar bug.
	maxDueShift = 14
)

// errNoBusinessDay reports a due date that never reached a business day.
var errNoBusinessDay = fmt.Errorf("taxengine: no business day within %d days of the due date", maxDueShift)

// dueDate is day 17 of the month after p, moved to the next business day
// (weekends and holidays, CFF 12 ⚠️ "sujeto a confirmación"; the RFC-digit
// grace days are not applied, research A.4 ⚠️). When every regime held is
// the bimonthly 621, the payment follows the bimester (§2.10): the month
// after its second month.
func (uc *Estimator) dueDate(ctx context.Context, p domain.Period, held []RegimeHold) (fiscal.Date, error) {
	offset := time.Month(1)
	if onlyBimonthly(held) {
		offset += p.Month() % 2 // an odd month is the bimester's first
	}
	day := time.Date(p.Year(), p.Month()+offset, dueDay, 0, 0, 0, 0, time.UTC)
	for range maxDueShift {
		d, err := fiscal.ParseDate(day.Format(time.DateOnly))
		if err != nil {
			return fiscal.Date{}, fmt.Errorf("taxengine: due date: %w", err)
		}
		business, err := uc.businessDay(ctx, day, d)
		if err != nil || business {
			return d, err
		}
		day = day.AddDate(0, 0, 1)
	}
	return fiscal.Date{}, errNoBusinessDay
}

// businessDay reports whether day is neither a weekend nor a holiday.
func (uc *Estimator) businessDay(ctx context.Context, day time.Time, d fiscal.Date) (bool, error) {
	if wd := day.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return false, nil
	}
	holiday, err := uc.deps.Holidays.IsHoliday(ctx, d)
	if err != nil {
		return false, fmt.Errorf("holiday %s: %w", d, err)
	}
	return !holiday, nil
}

// onlyBimonthly reports a taxpayer whose every regime in the period is 621.
func onlyBimonthly(held []RegimeHold) bool {
	for _, h := range held {
		if h.Code != regimeRIF {
			return false
		}
	}
	return len(held) > 0
}
