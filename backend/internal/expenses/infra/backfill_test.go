package infra_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// The backfill (WH-08) writes the supplier facts, the derivation and the
// creditable split only: cc_worker's UPDATE grant on expenses.
func TestDerivationStatementNamesOnlyTheDerivedColumns(t *testing.T) {
	want := []string{
		"creditable_iva_cents", "isr_withheld_cents", "iva_withheld_cents", "supplier_kind", "supplier_regime",
		"operation_kind", "residential_rent", "creditable_iva_withheld_cents", "withholding_rule_ids",
		"withholding_table_version", "withholding_cpa_unverified", "withholding_isr_rate", "withholding_iva_basis",
		"withholding_withholder",
	}
	statement := infra.DerivationStatement()
	set := strings.TrimPrefix(before(t, statement, " WHERE"), "UPDATE expenses SET ")
	var names []string
	for assignment := range strings.SplitSeq(set, ", ") {
		name, _, _ := strings.Cut(assignment, " = ")
		names = append(names, name)
	}
	assert.ElementsMatch(t, want, names)
	assert.Equal(t, len(want), strings.Count(set, "?"))
	assert.True(t, strings.HasSuffix(statement, "WHERE id = ?"))

	e := usecase.StoredExpense{
		ID: "x", Supplier: domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationBusinessServices},
		Withholding: fiscal.WithholdingRecord{TableVersion: "2026.1", Withholder: fiscal.WithholderNone},
	}
	args := infra.DerivationArgs(e)
	require.Len(t, args, len(want)+1)
	assert.Equal(t, "x", args[len(args)-1], "the id comes last")
}
