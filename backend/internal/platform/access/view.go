package access

import (
	"context"
	"time"
)

// day is one trial day: the trial lasts 168 hours (TrialDuration).
const day = 24 * time.Hour

// View is a user's access as the client sees it at one instant (contract
// AccessSummary). It is the one shape every screen gets (getDashboard,
// getMe, M7's subscription status): nil fields are the contract's nulls.
type View struct {
	Level             Level
	TrialEndsAt       *time.Time
	TrialDaysLeft     *int
	IncomesRemaining  *int
	ExpensesRemaining *int
}

// View shapes s as the client sees it at now: the trial fields only while
// the trial runs; the counters for trial and read-only (zero), none for pro
// and comp (no limit applies).
func (s Summary) View(now time.Time) View {
	out := View{Level: s.Status.Level}
	if s.Status.Level == LevelTrial {
		ends := s.Status.TrialEndsAt
		days := TrialDaysLeft(now, ends)
		out.TrialEndsAt, out.TrialDaysLeft = &ends, &days
	}
	if n, limited := s.Remaining(RecordIncome); limited {
		out.IncomesRemaining = &n
	}
	if n, limited := s.Remaining(RecordExpense); limited {
		out.ExpensesRemaining = &n
	}
	return out
}

// TrialDaysLeft is the whole days left before endsAt, rounded up and at
// least 1 (contract AccessSummary.trial_days_left, OQ-D6). The caller asks
// only while the trial runs (now < endsAt); the floor keeps a clock skew
// from ever showing 0 days.
func TrialDaysLeft(now, endsAt time.Time) int {
	left := endsAt.Sub(now)
	days := int((left + day - 1) / day)
	return max(days, 1)
}

// View returns the user's access as the client sees it now, on the policy's
// clock (Summary, then Summary.View).
func (p *Policy) View(ctx context.Context, userID string) (View, error) {
	summary, err := p.Summary(ctx, userID)
	if err != nil {
		return View{}, err
	}
	return summary.View(p.clock.Now()), nil
}
