package taxenginehttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// inputWriteBudget is the per-user write budget of putTaxInputs
// (MODULE-tax-engine §5, ADR-0005): 60 per hour.
var inputWriteBudget = ratelimit.PerWindow("tax_input_writes", 60, time.Hour)

// UserLimits are the per-user limits of the module by operationId, keyed by
// the authenticated user. cmd/api hands them to the router, which applies
// them right after authentication. Reads are left to the edge limiter.
var UserLimits = map[string]ratelimit.Policy{
	"putTaxInputs": inputWriteBudget,
}
