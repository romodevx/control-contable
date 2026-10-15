package infra_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
)

// A row with a table version but without the detail columns (the detail
// CHECKs allow it: detail_needs_derivation is one-way) must read as
// underived, as incomes does (20261017100000_withholdings_detail): the API
// then serves `withholding: null` (0.12.0) instead of a 500 on every read
// of the expense and its list (WH-15 security gate).
func TestRowWithholding_DerivationWithoutDetailReadsAsUnderived(t *testing.T) {
	version := "2026.1"
	half := infra.ExpenseRow{
		ISRWithheldCents: 400000, IVAWithheldCents: 426667,
		RuleIDs: "ISR-5,IVA-6", TableVersion: &version,
	}

	got, err := infra.RowWithholding(half)

	require.NoError(t, err)
	assert.False(t, got.Derived(), "a derivation without its detail is served as underived")
	assert.Empty(t, got.RuleIDs)
}

// A complete derivation still reads back whole.
func TestRowWithholding_CompleteDerivationReadsBack(t *testing.T) {
	version, rate, basis, withholder := "2026.1", "0.100000", "two_thirds_of_iva", "receptor"
	row := infra.ExpenseRow{
		ISRWithheldCents: 400000, IVAWithheldCents: 426667, RuleIDs: "ISR-5,IVA-6",
		TableVersion: &version, ISRRate: &rate, IVABasis: &basis, Withholder: &withholder,
	}

	got, err := infra.RowWithholding(row)

	require.NoError(t, err)
	assert.True(t, got.Derived())
	assert.Equal(t, "receptor", string(got.Withholder))
	assert.Len(t, got.RuleIDs, 2)
}
