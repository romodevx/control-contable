package taxenginehttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// inputWriteBudget is the per-user write budget of putTaxInputs
// (MODULE-tax-engine §5, ADR-0005): 60 per hour.
var inputWriteBudget = ratelimit.PerWindow("tax_input_writes", 60, time.Hour)

// estimateReadBudget is the per-user budget the estimate and the breakdown
// share (MODULE-tax-engine §5, ADR-0005): 600 per hour, a burst of 600
// refilled one every 6 s. Each read aggregates a fiscal year of incomes and
// expenses, so it is the module's costliest read. A user going through all
// twelve months, each with its estimate and three breakdowns, spends 48;
// a scraper is held to 10 a minute sustained.
var estimateReadBudget = ratelimit.PerWindow("estimate_reads", 600, time.Hour)

// UserLimits are the per-user limits of the module by operationId, keyed by
// the authenticated user. cmd/api hands them to the router, which applies
// them right after authentication. getTaxInputs is left to the edge
// limiter.
var UserLimits = map[string]ratelimit.Policy{
	"putTaxInputs":         inputWriteBudget,
	"getTaxpayerEstimate":  estimateReadBudget,
	"getEstimateBreakdown": estimateReadBudget,
}
