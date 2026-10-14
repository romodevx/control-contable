package expenseshttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// writeBudget is the per-user write budget of module_expenses (M4-07, as
// MODULE-taxpayers S-10): one bucket per user shared by every expense
// mutation, on the ADR-0005 limiter. 300 per hour lets a user type a month
// of receipts in one sitting (one write every 12 s for an hour, plus edits
// and undos) and stops a scripted flood; its own scope keeps it apart from
// the taxpayer and the future income budgets.
var writeBudget = ratelimit.PerWindow("expense_writes", 300, time.Hour)

// UserLimits are the per-user limits of the module by operationId, keyed by
// the authenticated user. cmd/api hands them to the router, which applies
// them right after authentication. Reads are not limited here (the edge
// limiter covers them).
var UserLimits = map[string]ratelimit.Policy{
	"createExpense":  writeBudget,
	"updateExpense":  writeBudget,
	"deleteExpense":  writeBudget,
	"restoreExpense": writeBudget,
}
