package contracttest_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

// pendingOperations are the contract operations with no contract test yet,
// each with the PLAN task that owns its contract tests. The list only
// shrinks: when a module's tests call ServeOp(contracttest.Op("<id>"), ...),
// the meta-test below fails until the id is removed from here. Adding an
// operation to the contract means adding it here (with its task) or covering
// it in the same PR.
var pendingOperations = map[string]string{
	// M7 module_subscriptions
	"getSubscription":          "M7-06",
	"syncSubscription":         "M7-06",
	"receiveRevenueCatWebhook": "M7-05",
	// M9 module_reminders
	"registerDevice":      "M9-05",
	"unregisterDevice":    "M9-05",
	"getReminderSettings": "M9-05",
	"putReminderSettings": "M9-05",
	"getTaxCalendar":      "M9-04",
}

// backendRoot is the Go module root, three levels above this package.
var backendRoot = filepath.Join("..", "..", "..")

// TestContractCoverage tracks which contract operations have a contract test:
// every operation is either covered (a ServeOp call with its Op literal
// somewhere in the backend's tests) or pending, never both and never neither.
func TestContractCoverage(t *testing.T) {
	ops := contracttest.Load(t).OperationIDs()
	covered, err := contracttest.ScanCoverage(backendRoot, ".") // this package's own tests are not contract tests
	require.NoError(t, err)

	for _, id := range ops {
		_, isPending := pendingOperations[id]
		sites, isCovered := covered[id]
		switch {
		case isCovered && isPending:
			t.Errorf("%s is covered (%v): remove it from pendingOperations", id, sites)
		case !isCovered && !isPending:
			t.Errorf("%s has no contract test: cover it with ServeOp(contracttest.Op(%q), ...) or add it to pendingOperations with its PLAN task", id, id)
		}
	}
	for id := range pendingOperations {
		assert.Containsf(t, ops, id, "pendingOperations lists %q, which is not in docs/api/openapi.yaml", id)
	}
	for id, sites := range covered {
		assert.Containsf(t, ops, id, "%v cover %q, which is not in docs/api/openapi.yaml", sites, id)
	}
	t.Logf("contract coverage: %d of %d operations covered, %d pending", len(ops)-len(pendingOperations), len(ops), len(pendingOperations))
}
