package domain

import (
	"slices"
	"time"
)

// The trend window's length (contract getDashboard months: 6..12, default
// 12, OQ-D1).
const (
	MinMonths     = 6
	MaxMonths     = 12
	DefaultMonths = MaxMonths
)

// FirstMonth is the first month any record or estimate can exist: no
// point of the trend is ever before it (MODULE-dashboard §2.1).
var FirstMonth = NewMonth(2026, time.January)

// Window returns the trend's months for period: [max(period − (months −
// 1), FirstMonth), period], oldest first, one per month. months is
// MinMonths..MaxMonths (the use case checks it).
func Window(period Month, months int) []Month {
	first := period.AddMonths(1 - months)
	if first.Before(FirstMonth) {
		first = FirstMonth
	}
	out := make([]Month, 0, months)
	for m := first; !period.Before(m); m = m.AddMonths(1) {
		out = append(out, m)
	}
	return out
}

// Status is the taxpayer-level status of one month's estimate (contract
// TaxpayerEstimateStatus); a regime's status uses the same values but
// never StatusNone.
type Status string

// The statuses.
const (
	StatusComplete      Status = "complete"
	StatusIncomplete    Status = "incomplete"
	StatusPartial       Status = "partial"
	StatusInformational Status = "informational"
	StatusNone          Status = "none"
)

// TaxpayerStatus derives the taxpayer's status from its regimes' (M5
// precedence, MODULE-dashboard §2.2): none without a regime, informational
// when every regime is, else incomplete if any is, partial if any is, and
// complete.
func TaxpayerStatus(regimes []Status) Status {
	switch {
	case len(regimes) == 0:
		return StatusNone
	case !slices.ContainsFunc(regimes, func(s Status) bool { return s != StatusInformational }):
		return StatusInformational
	case slices.Contains(regimes, StatusIncomplete):
		return StatusIncomplete
	case slices.Contains(regimes, StatusPartial):
		return StatusPartial
	}
	return StatusComplete
}

// day is one trial day: the trial lasts 168 hours (access.TrialDuration).
const day = 24 * time.Hour

// TrialDaysLeft is the whole days left before endsAt, rounded up and at
// least 1 (contract AccessSummary.trial_days_left, OQ-D6). The caller asks
// only while the trial runs (now < endsAt); the floor keeps a clock skew
// from ever showing 0 days.
func TrialDaysLeft(now, endsAt time.Time) int {
	left := endsAt.Sub(now)
	days := int((left + day - 1) / day)
	return max(days, 1)
}
