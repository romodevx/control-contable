package taxpayershttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// writeBudget is the per-user write budget of MODULE-taxpayers S-10: one
// bucket per user shared by every taxpayer and regime mutation (the same
// scope), 30 per hour (ADR-0005 limiter).
var writeBudget = ratelimit.PerWindow("taxpayer_writes", 30, time.Hour)

// UserLimits are the per-user limits of the module by operationId, keyed by
// the authenticated user. cmd/api hands them to the router, which applies
// them right after authentication. Reads are not limited here (the edge
// limiter covers them).
var UserLimits = map[string]ratelimit.Policy{
	"createTaxpayer":        writeBudget,
	"updateTaxpayer":        writeBudget,
	"deleteTaxpayer":        writeBudget,
	"archiveTaxpayer":       writeBudget,
	"unarchiveTaxpayer":     writeBudget,
	"addTaxpayerRegime":     writeBudget,
	"updateTaxpayerRegime":  writeBudget,
	"archiveTaxpayerRegime": writeBudget,
}
