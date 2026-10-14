package taxpayershttp_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
)

// MODULE-taxpayers S-10 (PR #59 gate condition 3): every mutation of the
// module spends one shared per-user budget of 30 per hour; reads spend
// nothing. The write set is read from the contract, so a new write
// operation without a budget fails here.
func TestUserLimits_coverEveryWriteOfTheModule(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	var writes, reads []string
	for _, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			if op.Extensions["x-module"] != "module_taxpayers" {
				continue
			}
			if method == http.MethodGet {
				reads = append(reads, op.OperationID)
			} else {
				writes = append(writes, op.OperationID)
			}
		}
	}

	require.Len(t, writes, 8)
	assert.ElementsMatch(t, writes, keys(taxpayershttp.UserLimits))
	for _, id := range reads {
		assert.NotContains(t, taxpayershttp.UserLimits, id, "reads spend no write budget")
	}
	want := ratelimit.PerWindow("taxpayer_writes", 30, time.Hour)
	for id, policy := range taxpayershttp.UserLimits {
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
