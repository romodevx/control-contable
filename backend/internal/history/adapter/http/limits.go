package historyhttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// History's own per-user budgets (MODULE-history §2.7, ADR-0005): the range
// estimate inside a history read does not spend estimate_reads.
var (
	// historyReadBudget: getHistory, 600 per hour, a burst of 600.
	historyReadBudget = ratelimit.PerWindow("history_reads", 600, time.Hour)
	// historyViewBudget: markHistoryViewed, 1,200 per hour (one every 3 s),
	// a burst of 120.
	historyViewBudget = ratelimit.Policy{Scope: "history_views", Burst: historyViewBurst, Interval: time.Hour / historyViewsPerHour}
)

const (
	historyViewsPerHour = 1200
	historyViewBurst    = 120
)

// UserLimits are the module's per-user limits by operationId, keyed by the
// authenticated user. cmd/api hands them to the router, which applies them
// right after authentication.
var UserLimits = map[string]ratelimit.Policy{
	"getHistory":        historyReadBudget,
	"markHistoryViewed": historyViewBudget,
}
