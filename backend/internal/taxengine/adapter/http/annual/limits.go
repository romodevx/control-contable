package annualhttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
)

// The annual budgets (ADR-0009 "Budgets", ADR-0005). The estimate, the
// breakdown and the filed payments (which show the app's estimates) spend
// the tax engine's estimate_reads, one more when the Estimator range runs
// (Deps.Range); the other reads spend annual_reads and every write
// annual_writes.
var (
	// RangePolicy is the estimate_reads policy the extra read spends.
	RangePolicy = taxenginehttp.EstimateReadBudget
	// annualReadBudget: the inputs' reads, 600 per hour.
	annualReadBudget = ratelimit.PerWindow("annual_reads", annualReadsPerHour, time.Hour)
	// annualWriteBudget: every input write (delete and restore included),
	// 300 per hour.
	annualWriteBudget = ratelimit.PerWindow("annual_writes", annualWritesPerHour, time.Hour)
)

const (
	annualReadsPerHour  = 600
	annualWritesPerHour = 300
)

// UserLimits are the module's per-user limits by operationId, keyed by the
// authenticated user. cmd/api hands them to the router, which applies them
// right after authentication.
var UserLimits = map[string]ratelimit.Policy{
	"getAnnualEstimate":        RangePolicy,
	"getAnnualBreakdown":       RangePolicy,
	"getFiledPayments":         RangePolicy,
	"putFiledPayments":         RangePolicy,
	"getAnnualInputs":          annualReadBudget,
	"listSalaryStatements":     annualReadBudget,
	"listPersonalDeductions":   annualReadBudget,
	"putAnnualInputs":          annualWriteBudget,
	"createSalaryStatement":    annualWriteBudget,
	"updateSalaryStatement":    annualWriteBudget,
	"deleteSalaryStatement":    annualWriteBudget,
	"createPersonalDeduction":  annualWriteBudget,
	"updatePersonalDeduction":  annualWriteBudget,
	"deletePersonalDeduction":  annualWriteBudget,
	"restorePersonalDeduction": annualWriteBudget,
}
