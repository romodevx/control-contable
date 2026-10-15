// Package accesshttp presents a user's access (access.View) as the
// contract's AccessSummary. Every operation that sends it (getDashboard,
// getMe, M7's subscription status) goes through Summary, so the wire shape
// has one presenter as the computation has one source (access.Summary.View).
package accesshttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
)

// Summary returns v as the contract's AccessSummary: nil fields encode as
// the contract's nulls, and trial_ends_at is UTC to the second.
func Summary(v access.View) gen.AccessSummary {
	out := gen.AccessSummary{
		Access:            gen.AccessLevel(v.Level),
		TrialDaysLeft:     v.TrialDaysLeft,
		IncomesRemaining:  v.IncomesRemaining,
		ExpensesRemaining: v.ExpensesRemaining,
	}
	if v.TrialEndsAt != nil {
		ends := v.TrialEndsAt.UTC().Truncate(time.Second)
		out.TrialEndsAt = &ends
	}
	return out
}
