package usecase_test

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Backfill fixture ids (WH-08).
const (
	consultant = "0199d4a3-0e0e-7000-8000-000000000001" // US-W6, typed before WH-07
	subscribed = "0199d4a3-0e0e-7000-8000-000000000002"
	personal   = "0199d4a3-0e0e-7000-8000-000000000003"
	older      = "0199d4a3-0e0e-7000-8000-000000000004"
	upToDate   = "0199d4a3-0e0e-7000-8000-000000000005"
	unknownCat = "0199d4a3-0e0e-7000-8000-000000000006"
)

// Underived implements usecase.BackfillStore: the ctx user's rows, deleted
// ones included, ordered by id.
func (f *fakeExpenses) Underived(ctx context.Context, version string, lock usecase.Lock) ([]usecase.StoredExpense, error) {
	if f.underivedErr != nil {
		return nil, f.underivedErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.backfillLocks = append(f.backfillLocks, lock)
	var out []usecase.StoredExpense
	for _, row := range f.rows {
		if row.owner == currentUser(ctx) && (row.e.Withholding.TableVersion != version || row.e.Withholding.Withholder == "") {
			out = append(out, row.e)
		}
	}
	slices.SortFunc(out, func(a, b usecase.StoredExpense) int { return strings.Compare(a.ID, b.ID) })
	return out, nil
}

// SaveDerivation implements usecase.BackfillStore.
func (f *fakeExpenses) SaveDerivation(ctx context.Context, e usecase.StoredExpense) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved++
	_, err := f.Update(ctx, e)
	return err
}

func newBackfill(h *harness) *usecase.Backfill {
	return usecase.NewBackfill(usecase.BackfillDeps{
		Tx: h.tx, Store: h.expenses, Taxpayers: h.taxpayers, Categories: fakeCategories{},
		Parameters: h.parameters, Withholdings: fakeWithholdings{}, Clock: clock.NewFixed(today),
	})
}

// seedLegacy stores userA's expense as written before WH-07: no supplier
// facts, typed withholdings, the full IVA creditable, no derivation.
func seedLegacy(t *testing.T, h *harness, id, taxpayer string, regime fiscal.RegimeCode, category domain.CategoryCode, amount, iva, isr, ivaWithheld money.Cents) {
	t.Helper()
	h.expenses.rows[id] = storedRow{owner: userA, e: usecase.StoredExpense{
		ID: id, TaxpayerID: taxpayer, Regime: regime, Category: category, Amount: amount,
		IVARate: fiscal.IVARate16, IVA: iva, PaymentMethod: domain.PaymentTransfer, PaidOn: date(t, "2026-06-15"),
		HasCFDI: true, Deductible: true, CreditableIVA: iva,
		Withholding: fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: isr, IVA: ivaWithheld}},
	}}
}

// US-W6: a PM 601 paid a PF consultant $40,000.00 + IVA $6,400.00 and typed
// the withholdings before WH-07.
func seedConsultant(t *testing.T, h *harness) {
	t.Helper()
	seedLegacy(t, h, consultant, taxpayerM, "601", "servicios_profesionales", 40_000_00, 6_400_00, 4_000_00, 4_266_67)
}

func TestExpensesBackfill_DryRunWritesNothing(t *testing.T) {
	h := newHarness(t)
	seedConsultant(t, h)
	before := maps.Clone(h.expenses.rows)

	got, err := newBackfill(h).Run(t.Context(), userA, false)

	require.NoError(t, err)
	assert.Equal(t, []usecase.Rederived{{ID: consultant, FactsAssumed: true}}, got.Rederived, "the typed amounts were right")
	assert.Equal(t, before, h.expenses.rows)
	assert.Zero(t, h.expenses.saved)
	assert.Equal(t, []string{"ro"}, h.tx.modes)
	assert.Equal(t, []usecase.Lock{usecase.LockNone}, h.expenses.backfillLocks)
}

func TestExpensesBackfill_ApplyDerivesUSW6(t *testing.T) {
	h := newHarness(t)
	seedConsultant(t, h)

	_, err := newBackfill(h).Run(t.Context(), userA, true)

	require.NoError(t, err)
	e := h.expenses.rows[consultant].e
	assert.Equal(t, domain.Supplier{
		Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612, Operation: fiscal.OperationProfessionalServices,
	}, e.Supplier)
	assert.Equal(t, money.Cents(4_000_00), e.Withholding.ISR)
	assert.Equal(t, money.Cents(4_266_67), e.Withholding.IVA)
	assert.Equal(t, fiscal.WithholdingTableVersion, e.Withholding.TableVersion)
	assert.Equal(t, money.Cents(2_133_33), e.CreditableIVA, "the paid part")
	assert.Equal(t, money.Cents(4_266_67), e.CreditableIVAWithheld, "credited the month after")
	assert.True(t, e.Deductible, "the deduction verdict is kept")
	assert.Equal(t, []string{"rw"}, h.tx.modes)
	assert.Equal(t, []usecase.Lock{usecase.LockUpdate}, h.expenses.backfillLocks)
}

func TestExpensesBackfill_LegacySuppliers(t *testing.T) {
	h := newHarness(t)
	seedLegacy(t, h, subscribed, taxpayerM, "601", "software", 500_00, 80_00, 0, 0)
	seedLegacy(t, h, personal, taxpayerA, fiscal.RegimeBusiness, "software", 500_00, 80_00, 0, 0)

	got, err := newBackfill(h).Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Len(t, got.Rederived, 2)
	pm := h.expenses.rows[subscribed].e
	assert.Equal(t, domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationBusinessServices}, pm.Supplier)
	assert.Zero(t, pm.Withholding.ISR+pm.Withholding.IVA)
	assert.Equal(t, money.Cents(80_00), pm.CreditableIVA)
	pf := h.expenses.rows[personal].e
	assert.Equal(t, domain.Supplier{}, pf.Supplier, "a persona física is never asked about suppliers")
	assert.Equal(t, fiscal.WithholdingTableVersion, pf.Withholding.TableVersion)
	assert.Equal(t, []usecase.Rederived{{ID: subscribed, FactsAssumed: true}, {ID: personal}}, got.Rederived)
}

// Security review of #139: a persona moral's expense deleted before WH-07
// had no supplier facts, so its restore answered 400. The backfill derives
// deleted rows too; the restore then succeeds.
func TestExpensesBackfill_DeletedLegacyExpenseCanBeRestored(t *testing.T) {
	h := newHarness(t)
	seedConsultant(t, h)
	row := h.expenses.rows[consultant]
	row.e.Deleted = true
	h.expenses.rows[consultant] = row
	_, err := h.uc.Restore(t.Context(), userA, taxpayerM, consultant)
	requireAppErr(t, err, apperr.CodeValidationFailed, domain.FieldSupplierKind, "required")

	_, err = newBackfill(h).Run(t.Context(), userA, true)
	require.NoError(t, err)
	assert.True(t, h.expenses.rows[consultant].e.Deleted, "still deleted")

	restored, err := h.uc.Restore(t.Context(), userA, taxpayerM, consultant)
	require.NoError(t, err)
	assert.Equal(t, money.Cents(4_000_00), restored.Withholding.ISR)
}

func TestExpensesBackfill_ReDerivesOlderTableVersionsOnly(t *testing.T) {
	h := newHarness(t)
	seedConsultant(t, h)
	old := h.expenses.rows[consultant]
	delete(h.expenses.rows, consultant)
	old.e.ID = older
	old.e.Supplier = domain.Supplier{
		Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612, Operation: fiscal.OperationProfessionalServices,
	}
	old.e.Withholding.ISR, old.e.Withholding.IVA = 1, 1
	old.e.Withholding.TableVersion, old.e.Withholding.Withholder = "2025.9", fiscal.WithholderReceptor
	h.expenses.rows[older] = old
	fresh := old
	fresh.e.ID, fresh.e.Withholding.TableVersion = upToDate, fiscal.WithholdingTableVersion
	h.expenses.rows[upToDate] = fresh

	got, err := newBackfill(h).Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Equal(t, []usecase.Rederived{{ID: older, AmountsChanged: true}}, got.Rederived)
	assert.Equal(t, money.Cents(4_000_00), h.expenses.rows[older].e.Withholding.ISR)
	assert.Equal(t, money.Cents(1), h.expenses.rows[upToDate].e.Withholding.ISR)

	again, err := newBackfill(h).Run(t.Context(), userA, true)
	require.NoError(t, err)
	assert.Empty(t, again.Rederived, "a second run changes nothing")
	assert.Equal(t, 1, h.expenses.saved)
}

func TestExpensesBackfill_RowsTheRulesRefuseNeedUserInput(t *testing.T) {
	h := newHarness(t)
	seedLegacy(t, h, unknownCat, taxpayerM, "601", "otros", 500_00, 80_00, 0, 0)
	seedLegacy(t, h, subscribed, taxpayerM, "601", "software", 500_00, 80_00, 0, 0)
	row := h.expenses.rows[subscribed]
	row.e.PaidOn = date(t, "2026-12-01") // after today
	h.expenses.rows[subscribed] = row

	got, err := newBackfill(h).Run(t.Context(), userA, true)

	require.NoError(t, err)
	assert.Empty(t, got.Rederived)
	assert.Equal(t, []string{subscribed, unknownCat}, got.NeedsInput)
	assert.Zero(t, h.expenses.saved)
}

func TestExpensesBackfill_Failures(t *testing.T) {
	boom := errors.New("boom")
	t.Run("store", func(t *testing.T) {
		h := newHarness(t)
		h.expenses.underivedErr = boom
		_, err := newBackfill(h).Run(t.Context(), userA, false)
		require.ErrorIs(t, err, boom)
	})
	t.Run("profile", func(t *testing.T) {
		h := newHarness(t)
		seedConsultant(t, h)
		h.taxpayers.byID[taxpayerM].owner = userB
		_, err := newBackfill(h).Run(t.Context(), userA, false)
		require.Error(t, err)
	})
	t.Run("category store", func(t *testing.T) {
		h := newHarness(t)
		seedLegacy(t, h, subscribed, taxpayerM, "999", "software", 500_00, 80_00, 0, 0)
		_, err := newBackfill(h).Run(t.Context(), userA, false)
		require.ErrorIs(t, err, errCategoryStore)
	})
	t.Run("save", func(t *testing.T) {
		h := newHarness(t)
		seedConsultant(t, h)
		h.expenses.saveErr = boom
		_, err := newBackfill(h).Run(t.Context(), userA, true)
		require.ErrorIs(t, err, boom)
	})
}
