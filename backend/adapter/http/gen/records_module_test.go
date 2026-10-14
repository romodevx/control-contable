package gen_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// recordModuleRoutes are the 7 operations of each record module, written by
// hand from MODULE-incomes (M3) and MODULE-expenses (M4): the generated map
// must agree with the specs, not merely with itself.
var recordModuleRoutes = map[string]map[string]string{
	"module_incomes": {
		"listIncomes":       "GET /v1/taxpayers/{taxpayerId}/incomes",
		"createIncome":      "POST /v1/taxpayers/{taxpayerId}/incomes",
		"getIncomeDefaults": "GET /v1/taxpayers/{taxpayerId}/incomes/defaults",
		"getIncome":         "GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}",
		"updateIncome":      "PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}",
		"deleteIncome":      "DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}",
		"restoreIncome":     "POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore",
		// ADR-0008: one preview per module, so each is hidden with its own flag.
		"previewIncomeWithholdings": "GET /v1/taxpayers/{taxpayerId}/incomes/withholding-preview",
	},
	"module_expenses": {
		"listExpenseCategories": "GET /v1/taxpayers/{taxpayerId}/expense-categories",
		"listExpenses":          "GET /v1/taxpayers/{taxpayerId}/expenses",
		"createExpense":         "POST /v1/taxpayers/{taxpayerId}/expenses",
		"getExpense":            "GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}",
		"updateExpense":         "PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}",
		"deleteExpense":         "DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}",
		"restoreExpense":        "POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore",
		// ADR-0008: one preview per module, so each is hidden with its own flag.
		"previewExpenseWithholdings": "GET /v1/taxpayers/{taxpayerId}/expenses/withholding-preview",
	},
}

// TestFlagMap_everyRecordOperationIsGatedByItsModule (M3-10, M4-11): every
// M3 operation, by operationId and by the ServeMux pattern the gate reads,
// is owned by module_incomes, every M4 one by module_expenses, and neither
// flag owns anything else. A route missing from here would ship ungated or
// hidden behind another module's flag.
func TestFlagMap_everyRecordOperationIsGatedByItsModule(t *testing.T) {
	for module, routes := range recordModuleRoutes {
		t.Run(module, func(t *testing.T) {
			for operationID, pattern := range routes {
				got, ok := gen.OperationModule(operationID)
				assert.True(t, ok, "%s is in the contract", operationID)
				assert.Equal(t, module, got, "%s -> flag", operationID)

				got, ok = gen.RouteModule(pattern)
				assert.True(t, ok, "route %q is registered", pattern)
				assert.Equal(t, module, got, "%s -> flag", pattern)
			}

			var owned []string
			for operationID, m := range gen.OperationModules() {
				if m == module {
					owned = append(owned, operationID)
				}
			}
			slices.Sort(owned)
			assert.Equal(t, slices.Sorted(maps.Keys(routes)), owned, "%s owns exactly its operations", module)
		})
	}
}
