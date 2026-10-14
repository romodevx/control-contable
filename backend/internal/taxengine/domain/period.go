package domain

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

const (
	// periodLayout is the contract Period format.
	periodLayout = "YYYY-MM"
	decimalBase  = 10
)

// Period is the tax period of an estimate: one calendar month. The zero
// value is "no period".
type Period struct {
	year  int
	month time.Month
	end   fiscal.Date
}

// ParsePeriod parses "YYYY-MM" exactly (ASCII digits, month 01-12, year
// 0001-9999), else ErrInvalidPeriod.
func ParsePeriod(raw string) (Period, error) {
	if len(raw) != len(periodLayout) || raw[4] != '-' {
		return Period{}, fmt.Errorf("%q: %w", raw, ErrInvalidPeriod)
	}
	year, okYear := digits(raw[:4])
	month, okMonth := digits(raw[5:])
	if !okYear || !okMonth || month < int(time.January) || month > int(time.December) {
		return Period{}, fmt.Errorf("%q: %w", raw, ErrInvalidPeriod)
	}
	// The last day doubles as the year check: the kernel's Date rejects 0000.
	end, err := fiscal.ParseDate(fmt.Sprintf("%s-%02d", raw, daysIn(year, time.Month(month))))
	if err != nil {
		return Period{}, fmt.Errorf("%q: %w", raw, ErrInvalidPeriod)
	}
	return Period{year: year, month: time.Month(month), end: end}, nil
}

// NewPeriod returns the period of year and month, else ErrInvalidPeriod.
func NewPeriod(year int, month time.Month) (Period, error) {
	return ParsePeriod(fmt.Sprintf("%04d-%02d", year, int(month)))
}

// digits parses an all-ASCII-digit string; ok is false for anything else.
func digits(s string) (n int, ok bool) {
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*decimalBase + int(r-'0')
	}
	return n, true
}

// daysIn returns the number of days of month in year. time.Date normalizes
// day 0 of the next month to the last day of this one; no clock is read.
func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// IsZero reports whether p is "no period".
func (p Period) IsZero() bool { return p.year == 0 }

// Year returns the period's year, which is also its fiscal year.
func (p Period) Year() int { return p.year }

// Month returns the period's month.
func (p Period) Month() time.Month { return p.month }

// End returns the period's last day: the date period-level parameters are
// resolved on (ADR-0002 §3).
func (p Period) End() fiscal.Date { return p.end }

// String returns "YYYY-MM", or "" for the zero Period.
func (p Period) String() string { return p.end.Period() }
