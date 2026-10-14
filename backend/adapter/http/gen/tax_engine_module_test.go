package gen_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleTaxEngine = "module_tax_engine"

// m5Routes are the 4 operations of MODULE-tax-engine, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m5Routes = map[string]string{
	"getTaxpayerEstimate":  "GET /v1/taxpayers/{taxpayerId}/estimates",
	"getEstimateBreakdown": "GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown",
	"getTaxInputs":         "GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs",
	"putTaxInputs":         "PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs",
}

// TestFlagMap_everyM5OperationIsGatedByModuleTaxEngine (M5-28): every M5
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_tax_engine, and module_tax_engine owns nothing else. A
// route missing from here would ship ungated or hidden behind another
// module's flag.
func TestFlagMap_everyM5OperationIsGatedByModuleTaxEngine(t *testing.T) {
	for operationID, pattern := range m5Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleTaxEngine, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleTaxEngine, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleTaxEngine {
			owned = append(owned, operationID)
		}
	}
	slices.Sort(owned)
	assert.Equal(t, slices.Sorted(maps.Keys(m5Routes)), owned, "module_tax_engine owns exactly the M5 operations")
}
