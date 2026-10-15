package remindershttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// taxCalendarReadBudget: getTaxCalendar, 120 per hour (one every 30 s), a
// burst of 60 (MODULE-reminders §2.7, ADR-0005).
var taxCalendarReadBudget = ratelimit.Policy{
	Scope: "tax_calendar_reads", Burst: taxCalendarBurst, Interval: time.Hour / taxCalendarReadsPerHour,
}

const (
	taxCalendarReadsPerHour = 120
	taxCalendarBurst        = 60
)

// UserLimits are the module's per-user limits by operationId, keyed by the
// authenticated user. cmd/api hands them to the router, which applies them
// right after authentication.
var UserLimits = map[string]ratelimit.Policy{
	"getTaxCalendar": taxCalendarReadBudget,
}
