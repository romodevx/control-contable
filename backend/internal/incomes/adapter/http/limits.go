package incomeshttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// writeBudget is the per-user write budget of module_incomes (M3-06, as
// module_expenses M4-07 and MODULE-taxpayers S-10): one bucket per user
// shared by every income mutation, on the ADR-0005 limiter. 300 per hour
// lets a user type a month of invoices in one sitting, with edits and
// undos, and stops a scripted flood; its own scope keeps it apart from the
// expense and taxpayer budgets.
var writeBudget = ratelimit.PerWindow("income_writes", 300, time.Hour)

// previewBudget is the per-user budget of the withholding previews
// (ADR-0008 §4): 1200 per hour, burst 1200, apart from the write budgets.
// The scope is the one previewExpenseWithholdings spends, so both previews
// share one bucket (ADR-0005: a scope is a bucket).
var previewBudget = ratelimit.PerWindow("withholding_previews", 1200, time.Hour)

// UserLimits are the per-user limits of the module by operationId, keyed by
// the authenticated user. cmd/api hands them to the router, which applies
// them right after authentication. Other reads are not limited here (the
// edge limiter covers them).
var UserLimits = map[string]ratelimit.Policy{
	"createIncome":  writeBudget,
	"updateIncome":  writeBudget,
	"deleteIncome":  writeBudget,
	"restoreIncome": writeBudget,

	"previewIncomeWithholdings": previewBudget,
}
