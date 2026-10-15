package gen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleAnnual = "module_annual"

// annualRoutes are the operations of MODULE-annual §4, written by hand: the
// generated map must agree with the spec, not merely with itself.
var annualRoutes = map[string]string{
	"getAnnualEstimate":        "GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}",
	"getAnnualBreakdown":       "GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown",
	"getAnnualInputs":          "GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}",
	"putAnnualInputs":          "PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}",
	"listSalaryStatements":     "GET /v1/taxpayers/{taxpayerId}/salary-statements",
	"createSalaryStatement":    "POST /v1/taxpayers/{taxpayerId}/salary-statements",
	"updateSalaryStatement":    "PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}",
	"deleteSalaryStatement":    "DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}",
	"listPersonalDeductions":   "GET /v1/taxpayers/{taxpayerId}/personal-deductions",
	"createPersonalDeduction":  "POST /v1/taxpayers/{taxpayerId}/personal-deductions",
	"updatePersonalDeduction":  "PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}",
	"deletePersonalDeduction":  "DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}",
	"restorePersonalDeduction": "POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore",
	"getFiledPayments":         "GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}",
	"putFiledPayments":         "PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}",
}

// TestFlagMap_everyAnnualOperationIsGatedByModuleAnnual (ANN-22): each
// annual operation, by operationId and by the ServeMux pattern the gate
// reads, is owned by module_annual, and module_annual owns nothing else.
func TestFlagMap_everyAnnualOperationIsGatedByModuleAnnual(t *testing.T) {
	for operationID, pattern := range annualRoutes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleAnnual, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleAnnual, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleAnnual {
			owned = append(owned, operationID)
		}
	}
	want := make([]string, 0, len(annualRoutes))
	for operationID := range annualRoutes {
		want = append(want, operationID)
	}
	assert.ElementsMatch(t, want, owned, "module_annual owns exactly the annual operations")
}
