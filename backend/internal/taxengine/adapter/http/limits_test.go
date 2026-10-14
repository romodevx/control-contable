package taxenginehttp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
)

// MODULE-tax-engine §5: the estimate and the breakdown share one per-user
// read budget, 600 per hour (each read aggregates a year of records); the
// tax-input writes keep their own 60 per hour.
func TestUserLimits_estimateReadsShareOneBudget(t *testing.T) {
	reads := ratelimit.PerWindow("estimate_reads", 600, time.Hour)
	assert.Equal(t, map[string]ratelimit.Policy{
		"putTaxInputs":         ratelimit.PerWindow("tax_input_writes", 60, time.Hour),
		"getTaxpayerEstimate":  reads,
		"getEstimateBreakdown": reads,
	}, taxenginehttp.UserLimits)
	for op, p := range taxenginehttp.UserLimits {
		require.NoError(t, p.Validate(), op)
	}
}
