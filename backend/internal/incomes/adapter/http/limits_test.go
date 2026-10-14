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
	// 4 reads since contract 0.10.0: previewIncomeWithholdings spends its own
	// withholding_previews budget, wired by WH-06 with its handler (ADR-0008).
	require.Len(t, reads, 4)
	assert.ElementsMatch(t, writes, keys(incomeshttp.UserLimits))
	for _, id := range reads {
		assert.NotContains(t, incomeshttp.UserLimits, id, "reads spend no write budget")
	}
	want := ratelimit.PerWindow("income_writes", 300, time.Hour)
	for id, policy := range incomeshttp.UserLimits {
		assert.Equal(t, want, policy, "%s shares the one write budget", id)
		require.NoError(t, policy.Validate())
	}
}

func keys(m map[string]ratelimit.Policy) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
