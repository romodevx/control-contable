package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func ptr[T any](v T) *T { return &v }

// seeded is a harness with expense1 created (software, 612, by card).
func seeded(t *testing.T, mutate ...func(*usecase.CreateInput)) *harness {
	t.Helper()
	h := newHarness(t)
	in := createInput(t, expense1)
	for _, m := range mutate {
		m(&in)
	}
	_, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)
	h.access.lastCalls, h.access.writable, h.taxpayers.checked = nil, 0, nil
	return h
}

func update(patch usecase.Patch) usecase.UpdateInput {
	return usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expense1, Patch: patch}
}

func TestUpdate_RecomputesTheDefaultWhenNotOverridden(t *testing.T) {
	h := seeded(t)

	got, err := h.uc.Update(t.Context(), update(usecase.Patch{Category: ptr(domain.CategoryCode("restaurantes"))}))

	require.NoError(t, err)
	assert.False(t, got.Deductible, "restaurantes defaults to non-deductible")
	assert.Equal(t, domain.SourceCategoryDefault, got.Source)
	assert.Equal(t, domain.CategoryCode("restaurantes"), got.Category)
	assert.Equal(t, money.Cents(500_00), got.Amount, "untouched fields stay")
	assert.Equal(t, "Suscripción anual de diseño", got.Description)
	assert.Empty(t, h.access.reserved[1:], "an edit reserves no slot")
	assert.Equal(t, []string{"writable"}, h.access.lastCalls)
	assert.True(t, logsMention(h, `"msg":"expense.updated"`))
}

func TestUpdate_KeepsAnOverrideAndClearsItWithNull(t *testing.T) {
	h := seeded(t, func(in *usecase.CreateInput) {
		in.Draft.Category, in.Draft.Deductible = "restaurantes", domain.ClaimDeductible
	})

	kept, err := h.uc.Update(t.Context(), update(usecase.Patch{Amount: ptr(money.Cents(400_00)), IVA: ptr(money.Cents(64_00))}))
	require.NoError(t, err)
	assert.True(t, kept.Deductible)
	assert.Equal(t, domain.SourceUser, kept.Source)

	cleared, err := h.uc.Update(t.Context(), update(usecase.Patch{Deductible: ptr(domain.UseDefault)}))
	require.NoError(t, err)
	assert.False(t, cleared.Deductible)
	assert.Equal(t, domain.SourceCategoryDefault, cleared.Source)
}

func TestUpdate_ClearsNullableFields(t *testing.T) {
	h := seeded(t, func(in *usecase.CreateInput) { in.Draft.CFDIUUID = uuidOf(t, cfdiUUID) })

	got, err := h.uc.Update(t.Context(), update(usecase.Patch{CFDIUUID: ptr(fiscal.CFDIUUID{}), Description: ptr("")}))

	require.NoError(t, err)
	assert.True(t, got.CFDIUUID.IsZero())
	assert.Empty(t, got.Description)
}

func TestUpdate_ReRunsEveryRuleOnTheMergedExpense(t *testing.T) {
	h := seeded(t)

	_, err := h.uc.Update(t.Context(), update(usecase.Patch{HasCFDI: ptr(false), CFDIUUID: ptr(uuidOf(t, cfdiUUID))}))
	requireAppErr(t, err, apperr.CodeValidationFailed, "cfdi_uuid", "has_cfdi_mismatch")

	_, err = h.uc.Update(t.Context(), update(usecase.Patch{Amount: ptr(money.Cents(900_00))}))
	requireAppErr(t, err, apperr.CodeBusinessRuleViolation, "iva_cents", "iva_amount_mismatch")
}

func TestUpdate_MovingToAnotherRegimeOrDateReChecksTheWindow(t *testing.T) {
	h := seeded(t)

	_, err := h.uc.Update(t.Context(), update(usecase.Patch{Regime: ptr(fiscal.RegimeRESICO)}))

	requireAppErr(t, err, apperr.CodeBusinessRuleViolation, "regime_code", "regime_not_active")
	assert.Equal(t, []string{"626@2026-09"}, h.taxpayers.checked)

	got, err := h.uc.Update(t.Context(), update(usecase.Patch{Regime: ptr(fiscal.RegimeRESICO), PaidOn: ptr(date(t, "2025-03-01"))}))
	require.NoError(t, err)
	assert.False(t, got.Deductible, "RESICO física never deducts")
}

func TestUpdate_WarnsLikeCreate(t *testing.T) {
	h := seeded(t)

	got, err := h.uc.Update(t.Context(), update(usecase.Patch{HasCFDI: ptr(false)}))

	require.NoError(t, err)
	assert.Equal(t, []string{"no_cfdi@has_cfdi"}, warningCodes(got.Warnings))
	assert.Equal(t, domain.Legend, got.Legend)
}

func TestUpdate_DuplicateFolioIgnoresTheExpenseItself(t *testing.T) {
	h := seeded(t, func(in *usecase.CreateInput) { in.Draft.CFDIUUID = uuidOf(t, cfdiUUID) })

	got, err := h.uc.Update(t.Context(), update(usecase.Patch{Description: ptr("nota")}))

	require.NoError(t, err)
	assert.Empty(t, got.Warnings)
}

func TestUpdate_Refusals(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*harness)
		in    usecase.UpdateInput
		code  apperr.Code
		field []string
	}{
		{"read-only", func(h *harness) { h.access.readOnly = true }, update(usecase.Patch{}), apperr.CodeEntitlementRequired, nil},
		{"archived", func(h *harness) { h.taxpayers.byID[taxpayerA].archived = true }, update(usecase.Patch{}),
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId", "taxpayer_archived"}},
		{"unknown expense", func(*harness) {}, usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expense2},
			apperr.CodeNotFound, nil},
		{"malformed expense id", func(*harness) {}, usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: "x"},
			apperr.CodeNotFound, nil},
		{"expense of another taxpayer of the caller", func(*harness) {},
			usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerM, ExpenseID: expense1}, apperr.CodeNotFound, nil},
		{"another user", func(*harness) {}, usecase.UpdateInput{UserID: userB, TaxpayerID: taxpayerB, ExpenseID: expense1},
			apperr.CodeNotFound, nil},
		{"deleted", func(h *harness) {
			_, err := h.expenses.SetDeleted(ctxAs(t, userA), expense1, true)
			require.NoError(t, err)
		}, update(usecase.Patch{}), apperr.CodeNotFound, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := seeded(t)
			c.setup(h)

			_, err := h.uc.Update(t.Context(), c.in)

			requireAppErr(t, err, c.code, c.field...)
		})
	}
}
