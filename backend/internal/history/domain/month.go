// Package domain holds the history's rules (M8, MODULE-history §2): which
// months a fiscal year lists, the taxpayer-level status, when a viewed
// month is "changed since viewed", the year totals and the disclaimer
// union. It does no tax math: every figure comes from the tax engine.
package domain

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidMonth reports a month that is not "YYYY-MM".
var ErrInvalidMonth = errors.New("history: month is not YYYY-MM")

// monthsPerYear is the number of calendar months in a year.
const monthsPerYear = 12

// monthLayout is the contract Period ("YYYY-MM").
const monthLayout = "2006-01"

// Month is a calendar month. It is comparable: equal months are ==.
type Month struct {
	year  int
	month time.Month
}

// NewMonth returns the month m of year.
func NewMonth(year int, m time.Month) Month { return Month{year: year, month: m} }

// MonthOf returns the calendar month of t, in t's location.
func MonthOf(t time.Time) Month { return NewMonth(t.Year(), t.Month()) }

// ParseMonth parses "YYYY-MM" (four-digit year, two-digit month).
func ParseMonth(raw string) (Month, error) {
	t, err := time.Parse(monthLayout, raw)
	if err != nil || len(raw) != len(monthLayout) || raw[0] < '0' || raw[0] > '9' {
		return Month{}, fmt.Errorf("%w: %q", ErrInvalidMonth, raw)
	}
	return MonthOf(t), nil
}

// Year returns the month's year.
func (m Month) Year() int { return m.year }

// Month returns the month of the year.
func (m Month) Month() time.Month { return m.month }

// String returns "YYYY-MM".
func (m Month) String() string { return fmt.Sprintf("%04d-%02d", m.year, int(m.month)) }

// Before reports whether m is earlier than o.
func (m Month) Before(o Month) bool { return m.index() < o.index() }

// index numbers months consecutively across years.
func (m Month) index() int { return m.year*monthsPerYear + int(m.month) - 1 }

// YearMonths lists the months of fiscalYear History shows, newest first
// (§2.1, OQ-H5): December .. January for a year before current's, current
// .. January for current's year. The use case refuses a future year.
func YearMonths(fiscalYear int, current Month) []Month {
	last := time.December
	if fiscalYear == current.Year() {
		last = current.Month()
	}
	out := make([]Month, 0, int(last))
	for m := last; m >= time.January; m-- {
		out = append(out, NewMonth(fiscalYear, m))
	}
	return out
}
