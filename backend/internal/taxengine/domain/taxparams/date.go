package taxparams

import (
	"fmt"
	"time"
)

const (
	minYear = 1
	maxYear = 9999
)

// Date is a calendar date with no time zone (ADR-0002 §4: record dates are
// DATEs). The zero value is "no date": an open validity end.
type Date struct {
	t time.Time // UTC midnight; zero for the zero Date
}

// NewDate returns the date, or ErrInvalidDate when the parts do not name a
// calendar day in years 0001-9999 (no normalisation: 30 February fails).
func NewDate(year int, month time.Month, day int) (Date, error) {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if year < minYear || year > maxYear || t.Year() != year || t.Month() != month || t.Day() != day {
		return Date{}, fmt.Errorf("%04d-%02d-%02d: %w", year, int(month), day, ErrInvalidDate)
	}
	return Date{t: t}, nil
}

// ParseDate parses "YYYY-MM-DD" exactly, else ErrInvalidDate.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("%q: %w", s, ErrInvalidDate)
	}
	return NewDate(t.Year(), t.Month(), t.Day())
}

// IsZero reports whether d is the zero Date.
func (d Date) IsZero() bool { return d.t.IsZero() }

// Year returns d's year.
func (d Date) Year() int { return d.t.Year() }

// Compare returns -1, 0 or +1 as d is before, equal to or after o.
func (d Date) Compare(o Date) int { return d.t.Compare(o.t) }

// String returns "YYYY-MM-DD", or "" for the zero Date.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.t.Format(time.DateOnly)
}

// Validity is the inclusive range of days an entry is in force. A zero To
// means open-ended: in force until a later entry of the key replaces it
// (tariffs carry on until the next publication; a decree with a sunset, like
// the 8% border IVA, sets To).
type Validity struct {
	From Date
	To   Date
}

// IsOpenEnded reports whether v has no end.
func (v Validity) IsOpenEnded() bool { return v.To.IsZero() }

// Contains reports whether d is in [From, To].
func (v Validity) Contains(d Date) bool {
	return d.Compare(v.From) >= 0 && (v.IsOpenEnded() || d.Compare(v.To) <= 0)
}

// String renders v as "[from, to]" or "[from, open)".
func (v Validity) String() string {
	if v.IsOpenEnded() {
		return fmt.Sprintf("[%s, open)", v.From)
	}
	return fmt.Sprintf("[%s, %s]", v.From, v.To)
}

// endsBefore reports whether v ends strictly before d, so a range starting
// on d does not overlap it.
func (v Validity) endsBefore(d Date) bool {
	return !v.IsOpenEnded() && v.To.Compare(d) < 0
}
