package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// legacyWithholding is the typed amount an income stored before WH-06
// carries: not a derivation.
var legacyWithholding = fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: 7}}

// seededLegacy is seeded's US-I1 income as stored before WH-06 and not
// yet derived by the WH-08 backfill: no facts, only its counterparty_type
// (persona_moral) and a typed amount.
func seededLegacy(t *testing.T) *harness {
	t.Helper()
	h := seeded(t)
	legacy := h.store.rows[incomeID]
	legacy.Counterparty, legacy.Operation, legacy.Withholding = "", "", legacyWithholding
	h.store.rows[incomeID] = legacy
	return h
}

// The facts a legacy 612 persona_moral income is served with: the
// backfill's own §5.2 mapping (counterparty_type, then the regime default).
func assertAssumedFacts(t *testing.T, got usecase.Income) {
	t.Helper()
	assert.Equal(t, fiscal.ReceptorPM, got.Counterparty)
	assert.Equal(t, fiscal.OperationProfessionalServices, got.Operation)
	assert.False(t, got.ViaPlatform)
}

// An income the WH-08 backfill has not derived yet is still served (no
// deploy order, ADR-0008 Amendment 2): its facts are the backfill's
// mapping and its withholding is left as stored, never derived on a read.
func TestGet_ServesALegacyIncomeWithTheBackfillFacts(t *testing.T) {
	h := seededLegacy(t)

	got, err := h.uc.Get(t.Context(), userA, taxpayerA, incomeID)

	require.NoError(t, err)
	assertAssumedFacts(t, got)
	assert.Equal(t, legacyWithholding, got.Withholding, "a read never derives: the estimate still uses the stored amounts")
	assert.Equal(t, fiscal.ReceptorKind(""), h.store.rows[incomeID].Counterparty, "a read writes nothing")
}

func TestList_ServesALegacyIncomeWithTheBackfillFacts(t *testing.T) {
	h := seededLegacy(t)

	got, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09"})

	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assertAssumedFacts(t, got.Items[0])
	assert.Equal(t, legacyWithholding, got.Items[0].Withholding)
}

// A patch applies to the facts the read served: the app sends only what
// the user changed, so an amount-only edit of a legacy income keeps the
// served facts and derives the withholding (editing derives it).
func TestUpdate_PatchesALegacyIncomeOnTheServedFacts(t *testing.T) {
	h := seededLegacy(t)

	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Amount: ptr(money.Cents(12_000_00)), IVA: ptr(money.Cents(1_920_00)),
	})

	require.NoError(t, err)
	assertAssumedFacts(t, got.Income)
	assert.True(t, got.Income.Withholding.Derived())
	assert.Equal(t, fiscal.ReceptorPM, h.store.rows[incomeID].Counterparty, "the write stores the facts")
}

// Restoring a legacy income derives it on the facts a read serves, as a
// backfill run before the restore would.
func TestRestore_DerivesALegacyIncomeOnTheServedFacts(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
	legacy := h.store.rows[incomeID]
	legacy.Counterparty, legacy.Operation, legacy.Withholding = "", "", legacyWithholding
	h.store.rows[incomeID] = legacy

	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, incomeID)

	require.NoError(t, err)
	assertAssumedFacts(t, got.Income)
	assert.True(t, got.Income.Withholding.Derived())
	assert.Equal(t, got.Income.Withholding, h.store.rows[incomeID].Withholding)
}
