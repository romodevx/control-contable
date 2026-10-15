package fiscal

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

// The due-date rule (ADR-0007 Amendment 1, MODULE-reminders §2.1, CFF art.
// 12 ¶1 and ¶5): a deadline that falls on a Saturday, a Sunday or a holiday
// that moves due dates moves to the next day that is none of those. Not
// applied ⚠️ (research §A.4.1): the RFC-digit grace days, SAT vacation
// periods (CFF 12 ¶2 makes them business days for payments) and the Friday
// extension (CFF 12 ¶5 last sentence, OQ-R1: a Friday is kept). Every
// reader (the monthly and annual estimates, the tax calendar, the reminder
// job) calls these functions; no other copy of the rule exists.

const (
	// MonthlyDueDay is the day of the month after the period on which the
	// monthly payment is due (LISR 14, 106, 113-E, 116; LIVA 5-D).
	MonthlyDueDay = 17
	// maxDueShift bounds the move to the next business day: no run of
	// weekends and holidays is this long, so reaching it is a calendar bug.
	maxDueShift = 14
)

// ErrNoBusinessDay reports a deadline that never reached a business day
// within the bound: the holiday data is wrong (internal_error).
var ErrNoBusinessDay = Error{code: "no_business_day", class: ClassInternal}

// MovedReason is why a due date is not its nominal date: the contract's
// TaxCalendarMovedReason, or MovedNone when it was not moved.
type MovedReason string

// The moved reasons.
const (
	MovedNone    MovedReason = ""
	MovedWeekend MovedReason = "weekend"
	MovedHoliday MovedReason = "holiday"
)

// HolidaySet is the immutable set of the days that move a due date (the
// holidays rows with moves_due_date = true) and the years it covers (the
// seeded holiday_years). The zero value covers no year: weekends only.
// Membership is a hash lookup, O(1).
type HolidaySet struct {
	years map[int]struct{}
	days  map[Date]struct{}
}

// NewHolidaySet builds the set of the covered years and their days. A zero
// day or a day outside the covered years is ErrInvalid. The arguments are
// copied: changing them later does not change the set.
func NewHolidaySet(years []int, days []Date) (HolidaySet, error) {
	h := HolidaySet{years: make(map[int]struct{}, len(years)), days: make(map[Date]struct{}, len(days))}
	for _, y := range years {
		h.years[y] = struct{}{}
	}
	for _, d := range days {
		if d.IsZero() || !h.Covers(d.Year()) {
			return HolidaySet{}, fmt.Errorf("holiday %q outside the covered years: %w", d, ErrInvalid)
		}
		h.days[d] = struct{}{}
	}
	return h, nil
}

// Covers reports whether the holidays of year are known. Outside a covered
// year the set answers weekends only; the caller decides whether that is
// acceptable (the calendar refuses it, the job falls back and alerts).
func (h HolidaySet) Covers(year int) bool {
	_, ok := h.years[year]
	return ok
}

// IsHoliday reports whether d is a day that moves a due date.
func (h HolidaySet) IsHoliday(d Date) bool {
	_, ok := h.days[d]
	return ok
}

// Years returns the covered years, ascending.
func (h HolidaySet) Years() []int { return slices.Sorted(maps.Keys(h.years)) }

// reason is why d is not a business day, or MovedNone when it is one. A
// holiday on a weekend is a weekend (the contract's moved_reason).
func (h HolidaySet) reason(d Date) MovedReason {
	switch wd := d.Weekday(); {
	case wd == time.Saturday || wd == time.Sunday:
		return MovedWeekend
	case h.IsHoliday(d):
		return MovedHoliday
	default:
		return MovedNone
	}
}

// NextBusinessDay is d when it is a business day, else the next day that
// is neither a weekend nor a holiday of h, with the kind of d itself as the
// reason (the days skipped after it may be of either kind). A run longer
// than the bound is ErrNoBusinessDay; the zero Date is ErrInvalid.
func NextBusinessDay(d Date, h HolidaySet) (Date, MovedReason, error) {
	if d.IsZero() {
		return Date{}, MovedNone, fmt.Errorf("next business day of the zero date: %w", ErrInvalid)
	}
	first := h.reason(d)
	day := d
	for range maxDueShift {
		if h.reason(day) == MovedNone {
			return day, first, nil
		}
		day = day.AddDays(1)
	}
	return Date{}, MovedNone, fmt.Errorf("%s: %w within %d days", d, ErrNoBusinessDay, maxDueShift)
}

// PaysBimonthly reports a taxpayer whose payment follows the bimester: every
// regime it holds in the period (codes) is 621 (MODULE-tax-engine §2.10).
// Holding none is not bimonthly. It is MonthlyDueDate's bimonthly argument.
func PaysBimonthly(codes []RegimeCode) bool {
	for _, c := range codes {
		if c != RegimeRIF {
			return false
		}
	}
	return len(codes) > 0
}

// MonthlyDueDate is the due date of the monthly period year-month: the
// nominal day 17 of the month after it (after the bimester's second month
// when bimonthly, the 621 regime, MODULE-tax-engine §2.10), and that day
// moved by NextBusinessDay.
func MonthlyDueDate(year int, month time.Month, bimonthly bool, h HolidaySet) (nominal, due Date, reason MovedReason, err error) {
	if year < 1 || month < time.January || month > time.December {
		return Date{}, Date{}, MovedNone, fmt.Errorf("period %d-%d: %w", year, month, ErrInvalid)
	}
	offset := time.Month(1)
	if bimonthly {
		offset += month % 2 // an odd month is the bimester's first
	}
	nominal = dateOf(time.Date(year, month+offset, MonthlyDueDay, 0, 0, 0, 0, time.UTC))
	due, reason, err = NextBusinessDay(nominal, h)
	if err != nil {
		return Date{}, Date{}, MovedNone, err
	}
	return nominal, due, reason, nil
}
