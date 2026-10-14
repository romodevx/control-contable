package dashboardhttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// dashboardReadBudget is getDashboard's per-user budget (MODULE-dashboard
// §2.5, OQ-D5, ADR-0005): 600 per hour, a burst of 600. It is its own
// budget: the range estimate inside a dashboard read does not spend
// estimate_reads, which guards the tax engine's routes.
var dashboardReadBudget = ratelimit.PerWindow("dashboard_reads", 600, time.Hour)

// UserLimits are the module's per-user limits by operationId, keyed by the
// authenticated user. cmd/api hands them to the router, which applies them
// right after authentication.
var UserLimits = map[string]ratelimit.Policy{
	"getDashboard": dashboardReadBudget,
}
