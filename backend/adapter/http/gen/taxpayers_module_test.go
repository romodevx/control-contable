package gen_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleTaxpayers = "module_taxpayers"

// m2Routes are the 11 M2 operations of MODULE-taxpayers, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m2Routes = map[string]string{
	"listRegimes":           "GET /v1/regimes",
	"listTaxpayers":         "GET /v1/taxpayers",
	"createTaxpayer":        "POST /v1/taxpayers",
	"getTaxpayer":           "GET /v1/taxpayers/{taxpayerId}",
	"updateTaxpayer":        "PATCH /v1/taxpayers/{taxpayerId}",
	"deleteTaxpayer":        "DELETE /v1/taxpayers/{taxpayerId}",
	"archiveTaxpayer":       "POST /v1/taxpayers/{taxpayerId}/archive",
	"unarchiveTaxpayer":     "POST /v1/taxpayers/{taxpayerId}/unarchive",
	"addTaxpayerRegime":     "POST /v1/taxpayers/{taxpayerId}/regimes",
	"updateTaxpayerRegime":  "PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}",
	"archiveTaxpayerRegime": "DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}",
}

// TestFlagMap_everyM2OperationIsGatedByModuleTaxpayers (M2-11): every M2
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_taxpayers, and module_taxpayers owns nothing else. A route
// missing from here would ship ungated or hidden behind another module's flag.
func TestFlagMap_everyM2OperationIsGatedByModuleTaxpayers(t *testing.T) {
	for operationID, pattern := range m2Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleTaxpayers, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleTaxpayers, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleTaxpayers {
			owned = append(owned, operationID)
		}
	}
	slices.Sort(owned)
	want := make([]string, 0, len(m2Routes))
	for operationID := range m2Routes {
		want = append(want, operationID)
	}
	slices.Sort(want)
	assert.Equal(t, want, owned, "module_taxpayers owns exactly the M2 operations")
}
