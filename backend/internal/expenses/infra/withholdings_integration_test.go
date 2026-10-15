//go:build integration

package infra_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// feesUSW6 is US-W6 (WM WH-11) on userA's persona moral: a 612 consultant,
// 40,000.00 + IVA 6,400.00 paid on 2026-06-15; the operation comes from the
// category's default_operation_kind.
func feesUSW6(t *testing.T, n int) usecase.CreateInput {
	t.Helper()
	in := software(t, n, "2026-06-15")
	in.TaxpayerID, in.Regime = taxpayerM, "601"
	in.Draft.Category, in.Draft.Amount, in.Draft.IVA = "servicios_profesionales", 40_000_00, 6_400_00
	in.Draft.PaymentMethod = domain.PaymentTransfer
	in.Draft.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612}
	return in
}

// The derivation is stored as cc_app under RLS, column by column, and read
// back whole: rule ids, table version, detail and the creditable split.
func TestCreate_USW6StoresTheDerivation(t *testing.T) {
	e := newEnv(t)

	got, err := e.uc.Create(t.Context(), feesUSW6(t, 1))

	require.NoError(t, err)
	var (
		kind, regime, operation, ruleIDs, version, rate, basis, withholder string
		isr, iva, creditable, creditableWithheld                           int64
		residential, unverified                                            bool
	)
	queryAs(t, e.app, userA, `SELECT supplier_kind, supplier_regime, operation_kind, residential_rent,
		array_to_string(withholding_rule_ids, ','), withholding_table_version, withholding_cpa_unverified,
		withholding_isr_rate::text, withholding_iva_basis, withholding_withholder,
		isr_withheld_cents, iva_withheld_cents, creditable_iva_cents, creditable_iva_withheld_cents
		FROM expenses WHERE id = $1`, []any{expenseID(1)},
		&kind, &regime, &operation, &residential, &ruleIDs, &version, &unverified, &rate, &basis, &withholder,
		&isr, &iva, &creditable, &creditableWithheld)
	assert.Equal(t, []string{"pf_with_activity", "612", "professional_services"}, []string{kind, regime, operation})
	assert.False(t, residential)
	assert.Equal(t, "ISR-5,IVA-6", ruleIDs)
	assert.Equal(t, fiscal.WithholdingTableVersion, version)
	assert.False(t, unverified)
	assert.Equal(t, "0.100000", rate)
	assert.Equal(t, []string{"two_thirds_of_iva", "receptor"}, []string{basis, withholder})
	assert.Equal(t, []int64{4_000_00, 4_266_67, 2_133_33, 4_266_67}, []int64{isr, iva, creditable, creditableWithheld})

	read, err := e.uc.Get(t.Context(), userA, taxpayerM, expenseID(1))
	require.NoError(t, err)
	assert.Equal(t, got.Expense.Withholding, read.Withholding, "the read serves what the write stored")
	assert.Equal(t, "2026-07-17", read.RemitDueOn.String())
	assert.Equal(t, money.Cents(4_000_00), read.Withholding.ISR)
}

// A persona física stores no supplier facts and a zero, derived record.
func TestCreate_PersonaFisicaStoresAZeroDerivation(t *testing.T) {
	e := newEnv(t)

	_, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))

	require.NoError(t, err)
	var kind, operation sql.NullString
	var ruleIDs, withholder string
	queryAs(t, e.app, userA, `SELECT supplier_kind, operation_kind, array_to_string(withholding_rule_ids, ','),
		withholding_withholder FROM expenses WHERE id = $1`, []any{expenseID(1)}, &kind, &operation, &ruleIDs, &withholder)
	assert.False(t, kind.Valid || operation.Valid)
	assert.Equal(t, "ISR-0,IVA-1", ruleIDs)
	assert.Equal(t, "none", withholder)
}

// An update re-derives and the list reads the new columns (List selects
// them explicitly).
func TestUpdate_ReDerivesAndListsTheDerivation(t *testing.T) {
	e := newEnv(t)
	_, err := e.uc.Create(t.Context(), feesUSW6(t, 1))
	require.NoError(t, err)

	_, err = e.uc.Update(t.Context(), usecase.UpdateInput{
		UserID: userA, TaxpayerID: taxpayerM, ExpenseID: expenseID(1),
		Patch: usecase.Patch{SupplierKind: ptr(domain.SupplierPM)},
	})
	require.NoError(t, err)

	page, err := e.uc.List(t.Context(), usecase.ListInput{UserID: userA, TaxpayerID: taxpayerM, Period: "2026-06"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	item := page.Items[0]
	assert.Equal(t, domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationProfessionalServices}, item.Supplier)
	assert.Equal(t, []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8}, item.Withholding.RuleIDs)
	assert.Equal(t, fiscal.Withholdings{}, item.Withholding.Withholdings)
	assert.Equal(t, money.Cents(0), page.Totals.ISRWithheld)
}

func TestPreview_ReadsTheCatalogAndStoresNothing(t *testing.T) {
	e := newEnv(t)

	got, err := e.uc.Preview(t.Context(), usecase.PreviewInput{
		UserID: userA, TaxpayerID: taxpayerM, Regime: "601", On: "2026-06-15", Category: "servicios_profesionales",
		Amount: 40_000_00, IVARate: fiscal.IVARate16, IVA: 6_400_00,
		Supplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612},
	})

	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, got.Withholding.Withholdings)
	assert.Equal(t, 0, expensesCreated(t, e.app, userA))
}
