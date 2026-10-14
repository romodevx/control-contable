package domain

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidMonth reports a month that is not "YYYY-MM".
var ErrInvalidMonth = errors.New("dashboard: month is not YYYY-MM")

// monthsPerYear is the number of calendar months in a year.
const monthsPerYear = 12

// monthLayout is the contract Period ("YYYY-MM").
const monthLayout = "2006-01"

// Month is a calendar month, the dashboard's period and trend point. It is
// comparable: equal months are ==.
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

// AddMonths returns the month n months after m (before it when n < 0).
func (m Month) AddMonths(n int) Month {
	i := m.index() + n
	return NewMonth(i/monthsPerYear, time.Month(i%monthsPerYear+1))
}

// index numbers months consecutively across years.
func (m Month) index() int { return m.year*monthsPerYear + int(m.month) - 1 }
