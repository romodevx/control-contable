package fiscal

import (
	"cmp"
	"time"
)

// Date layout: "YYYY-MM-DD" (contract format: date), a calendar day with no
// time zone (ADR-0002 §4).
const (
	dateLayout   = "2006-01-02"
	periodLayout = "2006-01"
)

// Date is a calendar day. The zero value means "no date" (a null field).
type Date struct {
	year  int
	month time.Month
	day   int
}

// MinRecordDate is the structural lower bound of a record date (MODULE-incomes
// §2.2), the same 2022-01-01 the database CHECKs hold as a backstop. It is
// not the served floor: a record date also needs a tax parameter set, and the
// first one is 2026, so the use cases answer out_of_range before 2026-01-01
// (contract 0.7.0, PR #97 FU-2).
var MinRecordDate = Date{year: 2022, month: time.January, day: 1}

// ParseDate parses "YYYY-MM-DD" exactly into a real calendar day (year
// 0001-9999), else ErrInvalid.
func ParseDate(raw string) (Date, error) {
	if len(raw) != len(dateLayout) {
		return Date{}, ErrInvalid
	}
	t, err := time.Parse(dateLayout, raw)
	if err != nil || t.Year() == 0 {
		return Date{}, ErrInvalid
	}
	return Date{year: t.Year(), month: t.Month(), day: t.Day()}, nil
}

// IsZero reports whether d is "no date".
func (d Date) IsZero() bool { return d.year == 0 }

// Year returns the calendar (and fiscal, CFF 11) year of d, or 0 for the
// zero Date.
func (d Date) Year() int { return d.year }

// String returns "YYYY-MM-DD", or "" for the zero Date.
func (d Date) String() string { return d.format(dateLayout) }

// Period returns the "YYYY-MM" tax period of d, or "" for the zero Date.
func (d Date) Period() string { return d.format(periodLayout) }

func (d Date) format(layout string) string {
	if d.IsZero() {
		return ""
	}
	return d.civil().Format(layout)
}

// civil is d at midnight UTC: only for formatting, never for "now".
func (d Date) civil() time.Time { return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC) }

// Compare returns -1, 0 or +1 as d is before, equal to or after o.
func (d Date) Compare(o Date) int {
	if c := cmp.Compare(d.year, o.year); c != 0 {
		return c
	}
	if c := cmp.Compare(d.month, o.month); c != 0 {
		return c
	}
	return cmp.Compare(d.day, o.day)
}

// Before reports whether d is an earlier day than o.
func (d Date) Before(o Date) bool { return d.Compare(o) < 0 }

// After reports whether d is a later day than o.
func (d Date) After(o Date) bool { return d.Compare(o) > 0 }
