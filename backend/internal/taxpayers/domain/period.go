package domain

import (
	"cmp"
	"fmt"
	"time"
)

// Period layout: "YYYY-MM" (contract Period), a calendar month.
const (
	periodLength   = len("2006-01")
	periodSep      = '-'
	periodYearLen  = 4
	periodMonthPos = periodYearLen + 1
	monthsInYear   = 12
)

// Period is a calendar month, the tax period every regime window and record
// is keyed by. The zero value is "no period": an open window end.
type Period struct {
	year  int
	month time.Month
}

// ParsePeriod parses "YYYY-MM" exactly (ASCII digits, month 01-12, year
// 0001-9999), else ErrInvalidPeriod. It neither trims nor accepts a day.
func ParsePeriod(raw string) (Period, error) {
	if len(raw) != periodLength || raw[periodYearLen] != periodSep {
		return Period{}, ErrInvalidPeriod
	}
	year, ok := asciiNumber(raw[:periodYearLen])
	if !ok || year == 0 {
		return Period{}, ErrInvalidPeriod
	}
	month, ok := asciiNumber(raw[periodMonthPos:])
	if !ok || month < 1 || month > monthsInYear {
		return Period{}, ErrInvalidPeriod
	}
	return Period{year: year, month: time.Month(month)}, nil
}

// PeriodOf returns the month of instant in instant's own location: the
// caller passes a Mexico City time (platform/clock) for "the current month".
func PeriodOf(instant time.Time) Period {
	return Period{year: instant.Year(), month: instant.Month()}
}

// IsZero reports whether p is the zero Period.
func (p Period) IsZero() bool { return p.year == 0 }

// String returns "YYYY-MM", or "" for the zero Period.
func (p Period) String() string {
	if p.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d", p.year, int(p.month))
}

// Compare returns -1, 0 or +1 as p is before, equal to or after q.
func (p Period) Compare(q Period) int {
	if c := cmp.Compare(p.year, q.year); c != 0 {
		return c
	}
	return cmp.Compare(p.month, q.month)
}

// Before reports whether p is an earlier month than q.
func (p Period) Before(q Period) bool { return p.Compare(q) < 0 }

// After reports whether p is a later month than q.
func (p Period) After(q Period) bool { return p.Compare(q) > 0 }

// Window is the span of periods a taxpayer holds a regime:
// [Start, End], where a zero End means still open (+infinity).
type Window struct {
	Start Period
	End   Period
}

// IsOpen reports whether the window has no end.
func (w Window) IsOpen() bool { return w.End.IsZero() }

// Contains reports whether p falls inside the window.
func (w Window) Contains(p Period) bool {
	return !p.Before(w.Start) && (w.IsOpen() || !p.After(w.End))
}

// Overlaps reports whether the two windows share at least one period:
// s1 <= e2 and s2 <= e1, an open end being +infinity. Adjacent windows
// (one ends in June, the other starts in July) do not overlap.
func (w Window) Overlaps(other Window) bool {
	return w.startsNoLaterThanEndOf(other) && other.startsNoLaterThanEndOf(w)
}

func (w Window) startsNoLaterThanEndOf(other Window) bool {
	return other.IsOpen() || !w.Start.After(other.End)
}

// asciiNumber parses a string of ASCII digits only.
func asciiNumber(digits string) (int, bool) {
	const base = 10
	n := 0
	for i := range len(digits) {
		c := digits[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*base + int(c-'0')
	}
	return n, true
}
