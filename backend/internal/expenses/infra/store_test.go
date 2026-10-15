package infra_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
)

// The writes are built from one allowlist (PR #136 note 1): the database's
// column grants would not stop a generic map that echoes request fields,
// so the columns a write may name are pinned here. Adding a column is a
// deliberate edit of this list.
func TestWritesNameOnlyTheAllowlistedColumns(t *testing.T) {
	allowlist := []string{
		"regime_code", "category_code", "description", "amount_cents", "iva_rate", "iva_cents", "payment_method",
		"has_cfdi", "cfdi_uuid", "deductible", "deductible_overridden", "cash_over_limit", "creditable_iva_cents",
		"isr_withheld_cents", "iva_withheld_cents", "paid_on", "supplier_kind", "supplier_regime", "operation_kind",
		"residential_rent", "creditable_iva_withheld_cents", "withholding_rule_ids", "withholding_table_version",
		"withholding_cpa_unverified", "withholding_isr_rate", "withholding_iva_basis", "withholding_withholder",
	}
	assert.Equal(t, allowlist, infra.EditableColumnNames())

	values := infra.EditableValues(usecase.StoredExpense{})
	assert.Len(t, values, len(allowlist), "one value per column")

	insert := infra.InsertStatement()
	assert.Equal(t, 3+len(allowlist), strings.Count(before(t, insert, "RETURNING"), "?"),
		"id, user_id, taxpayer_id and the allowlist")
	assert.True(t, strings.HasPrefix(insert,
		"INSERT INTO expenses (id, user_id, taxpayer_id, "+strings.Join(allowlist, ", ")+") VALUES"))

	update := infra.UpdateStatement()
	set := before(t, update, " WHERE")
	assert.Equal(t, len(allowlist), strings.Count(set, "?"))
	for _, forbidden := range []string{"user_id", "taxpayer_id", "deleted_at", "created_at", "booked"} {
		assert.NotContains(t, set, forbidden, "an update never moves an expense or its audit columns")
	}
}

// An underived record writes NULL detail and an empty rule list, which the
// detail and derived-together CHECKs require.
func TestEditableValuesOfAnUnderivedRecord(t *testing.T) {
	values := infra.EditableValues(usecase.StoredExpense{})
	names := infra.EditableColumnNames()
	byName := map[string]any{}
	for i, n := range names {
		byName[n] = values[i]
	}

	assert.Equal(t, "{}", byName["withholding_rule_ids"])
	for _, n := range []string{"withholding_table_version", "withholding_isr_rate", "withholding_iva_basis",
		"withholding_withholder", "supplier_kind", "supplier_regime", "operation_kind"} {
		assert.Nil(t, byName[n], n)
	}
}

// before is s up to sep, which must occur.
func before(t *testing.T, s, sep string) string {
	t.Helper()
	head, _, found := strings.Cut(s, sep)
	assert.True(t, found, sep)
	return head
}
