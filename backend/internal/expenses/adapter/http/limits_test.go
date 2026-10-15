package expenseshttp_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// Every mutation of module_expenses spends one shared per-user budget of 300
// per hour (MODULE-taxpayers S-10's limiter, its own scope); reads spend
// nothing but the withholding preview. The write set is read from the contract, so a new write
// operation without a budget fails here.
func TestUserLimits_coverEveryWriteOfTheModule(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	var writes, reads []string
	for _, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			if op.Extensions["x-module"] != "module_expenses" {
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
	const preview = "previewExpenseWithholdings"
	require.Contains(t, reads, preview)
	assert.ElementsMatch(t, append(writes, preview), keys(expenseshttp.UserLimits))
	for _, id := range reads {
		if id != preview {
			assert.NotContains(t, expenseshttp.UserLimits, id, "reads spend no budget")
		}
	}
	want := ratelimit.PerWindow("expense_writes", 300, time.Hour)
	for _, id := range writes {
		assert.Equal(t, want, expenseshttp.UserLimits[id], "%s shares the one write budget", id)
		require.NoError(t, expenseshttp.UserLimits[id].Validate())
	}
}

// The preview spends the withholding_previews budget (ADR-0008 §4: 1200 per
// hour, burst 1200), shared with previewIncomeWithholdings by its scope.
func TestUserLimits_previewSpendsTheSharedPreviewBudget(t *testing.T) {
	want := ratelimit.PerWindow("withholding_previews", 1200, time.Hour)

	assert.Equal(t, want, expenseshttp.UserLimits["previewExpenseWithholdings"])
	assert.Equal(t, want, expenseshttp.PreviewBudget)
	assert.Equal(t, 1200, want.Burst)
	require.NoError(t, want.Validate())
}

func keys(m map[string]ratelimit.Policy) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
