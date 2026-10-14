package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	expenses "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expensesuc "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// Every warning code a module can send must be in the contract's WarningCode enum (the single
// list, docs/api/openapi.yaml). A module adding a code without the contract fails here, not in
// the client's unknown-code fallback.
func TestModuleWarningCodesAreInTheContract(t *testing.T) {
	codes := []string{
		taxpayers.WarningRFCCheckDigitMismatch,
		taxpayers.WarningRegimeNeedsCPAReview,
		string(incomes.WarningWithholdingDiffersFromRate),
		string(incomes.WarningDuplicateCFDIUUID),
		string(expenses.WarningCashOverThreshold),
		string(expenses.WarningNoCFDI),
		string(expenses.WarningNotISRDeductibleInRegime),
		string(expenses.WarningCategoryNotDeductible),
		string(expenses.WarningDuplicateCFDIUUID),
		string(expenses.WarningIVANotCreditableCashPayment),
		string(expensesuc.WarningProvisionalParameters),
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			assert.True(t, gen.WarningCode(code).Valid(), "add %q to WarningCode in docs/api/openapi.yaml", code)
		})
	}
}
