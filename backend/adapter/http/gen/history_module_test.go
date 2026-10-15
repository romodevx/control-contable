package gen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleHistory = "module_history"

// m8Routes are the operations of MODULE-history, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m8Routes = map[string]string{
	"getHistory":        "GET /v1/taxpayers/{taxpayerId}/history",
	"markHistoryViewed": "POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed",
}

// TestFlagMap_everyM8OperationIsGatedByModuleHistory (M8-06): each M8
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_history, and module_history owns nothing else.
func TestFlagMap_everyM8OperationIsGatedByModuleHistory(t *testing.T) {
	for operationID, pattern := range m8Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleHistory, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleHistory, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleHistory {
			owned = append(owned, operationID)
		}
	}
	assert.ElementsMatch(t, []string{"getHistory", "markHistoryViewed"}, owned,
		"module_history owns exactly the M8 operations")
}
