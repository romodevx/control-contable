package incomeshttp_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	incomeshttp "github.com/romodevx/control-contable/backend/internal/incomes/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Every mutation of module_incomes spends one shared per-user budget of 300
// per hour (as module_expenses, its own scope); reads spend nothing. The
// write set is read from the contract, so a new write operation without a
// budget fails here.
func TestUserLimits_coverEveryWriteOfTheModule(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	var writes, reads []string
	for _, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			if op.Extensions["x-module"] != "module_incomes" {
				continue
			}
			if method == http.MethodGet {
				reads = append(reads, op.OperationID)
			} else {
				writes = append(writes, op.OperationID)
			}
		}
	}

	require.Len(t, writes, 4)
	require.Len(t, reads, 4)
	writeBudget := ratelimit.PerWindow("income_writes", 300, time.Hour)
	previews := ratelimit.PerWindow("withholding_previews", 1200, time.Hour)
	want := map[string]ratelimit.Policy{"previewIncomeWithholdings": previews}
	for _, id := range writes {
		want[id] = writeBudget
	}
	assert.Equal(t, want, incomeshttp.UserLimits,
		"writes share one budget; the preview spends withholding_previews (ADR-0008 §4); other reads none")
	for _, policy := range incomeshttp.UserLimits {
		require.NoError(t, policy.Validate())
	}
}
