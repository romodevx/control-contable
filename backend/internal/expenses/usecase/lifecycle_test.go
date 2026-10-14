package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

func isDeleted(t *testing.T, h *harness, id string) bool {
	t.Helper()
	e, err := h.expenses.Find(ctxAs(t, userA), id, usecase.LockNone)
	require.NoError(t, err)
	return e.Deleted
}

func TestDelete_SoftDeletesWithoutAnAccessCheck(t *testing.T) {
	h := seeded(t)
	h.access.readOnly = true // a read-only user may delete

	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	assert.True(t, isDeleted(t, h, expense1))
	assert.Empty(t, h.access.lastCalls)
	assert.True(t, logsMention(h, `"msg":"expense.deleted"`))
}

func TestDelete_IsIdempotent(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	assert.True(t, isDeleted(t, h, expense1))
}

func TestDelete_Refusals(t *testing.T) {
	cases := []struct {
		name               string
		user, taxpayer, id string
		archive            bool
		code               apperr.Code
		field              []string
	}{
		{"archived taxpayer", userA, taxpayerA, expense1, true, apperr.CodeBusinessRuleViolation, []string{"taxpayerId", "taxpayer_archived"}},
		{"another user", userB, taxpayerB, expense1, false, apperr.CodeNotFound, nil},
		{"another taxpayer's path", userA, taxpayerM, expense1, false, apperr.CodeNotFound, nil},
		{"unknown", userA, taxpayerA, expense2, false, apperr.CodeNotFound, nil},
		{"malformed", userA, taxpayerA, "nope", false, apperr.CodeNotFound, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := seeded(t)
			h.taxpayers.byID[taxpayerA].archived = c.archive

			err := h.uc.Delete(t.Context(), c.user, c.taxpayer, c.id)

			requireAppErr(t, err, c.code, c.field...)
			assert.False(t, isDeleted(t, h, expense1))
		})
	}
}

// DECIDED (MODULE-incomes §2.7, coordinator): restore never consumes a slot,
// and a read-only user may restore their own records.
func TestRestore_NeverReservesASlotAndAllowsReadOnly(t *testing.T) {
	for name, setup := range map[string]func(*fakeAccess){
		"at the trial cap": func(a *fakeAccess) { a.capSpent = true },
		"read-only":        func(a *fakeAccess) { a.readOnly = true },
	} {
		t.Run(name, func(t *testing.T) {
			h := seeded(t)
			require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))
			setup(h.access)

			got, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense1)

			require.NoError(t, err)
			assert.Equal(t, expense1, got.ID)
			assert.False(t, isDeleted(t, h, expense1))
			assert.Len(t, h.access.reserved, 1, "only the original create reserved")
			assert.Empty(t, h.access.lastCalls)
			assert.True(t, logsMention(h, `"msg":"expense.restored"`))
		})
	}
}

func TestRestore_ComputesWarningsAsOnCreate(t *testing.T) {
	h := seeded(t, func(in *usecase.CreateInput) { in.Draft.HasCFDI = false })
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense1)

	require.NoError(t, err)
	assert.Equal(t, []string{"no_cfdi@has_cfdi"}, warningCodes(got.Warnings))
	assert.False(t, got.Deductible)
	assert.Equal(t, domain.Legend, got.Legend)
}

func TestRestore_ALiveExpenseIsANoOp(t *testing.T) {
	h := seeded(t)

	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense1)

	require.NoError(t, err)
	assert.Equal(t, expense1, got.ID)
	assert.False(t, logsMention(h, `"msg":"expense.restored"`))
}

func TestRestore_ReChecksTheRegimeWindow(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))
	h.taxpayers.byID[taxpayerA].regimes[0].end = "2026-08"

	_, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense1)

	requireAppErr(t, err, apperr.CodeBusinessRuleViolation, "regime_code", "regime_not_active")
	assert.True(t, isDeleted(t, h, expense1))
}

func TestRestore_Refusals(t *testing.T) {
	t.Run("archived taxpayer", func(t *testing.T) {
		h := seeded(t)
		require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))
		h.taxpayers.byID[taxpayerA].archived = true

		_, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense1)

		requireAppErr(t, err, apperr.CodeBusinessRuleViolation, "taxpayerId", "taxpayer_archived")
	})
	t.Run("another user", func(t *testing.T) {
		h := seeded(t)

		_, err := h.uc.Restore(t.Context(), userB, taxpayerB, expense1)

		requireAppErr(t, err, apperr.CodeNotFound)
	})
	t.Run("purged or unknown", func(t *testing.T) {
		h := seeded(t)

		_, err := h.uc.Restore(t.Context(), userA, taxpayerA, expense2)

		requireAppErr(t, err, apperr.CodeNotFound)
	})
}
