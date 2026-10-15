// Package usecase holds the reminders module's use cases (MODULE-reminders):
// the tax calendar (getTaxCalendar, M9-04), devices (registerDevice,
// unregisterDevice) and reminder settings (get/putReminderSettings, M9-05).
// The due-date rule is the fiscal kernel's (ADR-0007 Amendment 1); the
// calendar only lays out a year of it. The ports are declared here, on the
// consumer side; every store port runs on the use case's unit of work.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// ErrMissingDependency reports a use case wired without one of its ports.
var ErrMissingDependency = errors.New("reminders: missing dependency")

// errUnseededYear is the cause of a calendar_unavailable refusal.
var errUnseededYear = errors.New("reminders: holidays of the year not seeded")

// DetailCalendarUnavailable is the details[].code of a year whose holidays
// are not seeded (contract getTaxCalendar, 422).
const DetailCalendarUnavailable = "calendar_unavailable"

const (
	fieldYear                = "year"
	msgCalendarUnavailable   = "The tax calendar for that year is not available yet."
	msgCalendarUnavailableES = "Aún no tenemos el calendario de ese año."
	// subjectToConfirmation is true for every entry while OQ-R1 (the CFF 12
	// ¶5 Friday extension) awaits a CPA.
	subjectToConfirmation = true
)

// HolidayCalendar answers the holidays that move a due date (the M9-03
// tables) as the kernel's HolidaySet. The API serves it from a boot
// snapshot (no statement per request); the reader in infra/holidays reads
// it inside the caller's unit of work (the boot read, the reminder job).
type HolidayCalendar interface {
	Holidays(ctx context.Context) (fiscal.HolidaySet, error)
}

// Alerts receives the operational signals of the calendar.
type Alerts interface {
	// CalendarUnseeded reports a due date computed weekends-only because
	// year's holidays are not seeded (metric reminders_calendar_unseeded).
	CalendarUnseeded(ctx context.Context, year int)
}

// CalendarDeps are the ports of the Calendar.
type CalendarDeps struct {
	Holidays HolidayCalendar
	Alerts   Alerts
}

// Calendar answers getTaxCalendar.
type Calendar struct {
	deps CalendarDeps
}

// NewCalendar returns the Calendar, or ErrMissingDependency.
func NewCalendar(deps CalendarDeps) (*Calendar, error) {
	switch {
	case deps.Holidays == nil:
		return nil, fmt.Errorf("%w: CalendarDeps.Holidays", ErrMissingDependency)
	case deps.Alerts == nil:
		return nil, fmt.Errorf("%w: CalendarDeps.Alerts", ErrMissingDependency)
	}
	return &Calendar{deps: deps}, nil
}

// Year is the calendar of one year's periods.
type Year struct {
	Year    int
	Entries []Entry
}

// Entry is one monthly period's due date (obligation monthly_isr_iva).
type Entry struct {
	Period                string // "YYYY-MM"
	Nominal               fiscal.Date
	Due                   fiscal.Date
	MovedReason           fiscal.MovedReason
	SubjectToConfirmation bool
}

// Year answers the periods year-01..year-12 (December due in year+1), the
// same for every user (MODULE-reminders §2.1): it reads no user data, so it
// opens no unit of work. year unseeded → 422
// calendar_unavailable; year+1 unseeded → December weekends-only and the
// alert. The router has checked year's range against the contract.
func (c *Calendar) Year(ctx context.Context, year int) (Year, error) {
	holidays, err := c.deps.Holidays.Holidays(ctx)
	if err != nil {
		return Year{}, fmt.Errorf("reminders: calendar %d: holidays: %w", year, err)
	}
	if !holidays.Covers(year) {
		return Year{}, apperr.Wrap(fmt.Errorf("%w: %d", errUnseededYear, year), apperr.CodeBusinessRuleViolation, msgCalendarUnavailable).
			WithFields(apperr.FieldError{Field: fieldYear, Code: DetailCalendarUnavailable, Message: msgCalendarUnavailableES})
	}
	if !holidays.Covers(year + 1) {
		c.deps.Alerts.CalendarUnseeded(ctx, year+1)
	}
	out := Year{Year: year, Entries: make([]Entry, 0, monthsPerYear)}
	for m := time.January; m <= time.December; m++ {
		nominal, due, reason, err := fiscal.MonthlyDueDate(year, m, false, holidays)
		if err != nil {
			return Year{}, fmt.Errorf("reminders: calendar %d-%02d: %w", year, m, err)
		}
		out.Entries = append(out.Entries, Entry{
			Period: fmt.Sprintf("%04d-%02d", year, int(m)), Nominal: nominal, Due: due, MovedReason: reason, SubjectToConfirmation: subjectToConfirmation,
		})
	}
	return out, nil
}

// monthsPerYear is the number of monthly periods in a year.
const monthsPerYear = 12
