package gen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleDashboard = "module_dashboard"

// m6Routes are the operations of MODULE-dashboard, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m6Routes = map[string]string{
	"getDashboard": "GET /v1/taxpayers/{taxpayerId}/dashboard",
}

// TestFlagMap_everyM6OperationIsGatedByModuleDashboard (M6-05): the M6
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_dashboard, and module_dashboard owns nothing else.
func TestFlagMap_everyM6OperationIsGatedByModuleDashboard(t *testing.T) {
	for operationID, pattern := range m6Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleDashboard, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleDashboard, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleDashboard {
			owned = append(owned, operationID)
		}
	}
	assert.ElementsMatch(t, []string{"getDashboard"}, owned, "module_dashboard owns exactly the M6 operation")
}
