package usecase_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func ptr[T any](v T) *T { return &v }

// seeded is a harness holding the US-I1 income of taxpayerA.
func seeded(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, us1(t)))
	require.NoError(t, err)
	h.taxpayers.calls = nil
	h.access.reserves, h.access.ensures = 0, 0
	return h
}

func TestUpdate_AmountAndIVA(t *testing.T) {
	h := seeded(t)
	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Amount: ptr(money.Cents(12_000_00)), IVA: ptr(money.Cents(1_920_00)),
	})
	require.NoError(t, err)
	assert.Equal(t, money.Cents(12_000_00), got.Income.Amount)
	assert.Equal(t, money.Cents(1_920_00), got.Income.IVA)
	assert.Equal(t, fiscal.Withholdings{ISR: 1_200_00, IVA: 1_280_00}, got.Income.Withholding.Withholdings, "re-derived")
	assert.Equal(t, got.Income.Withholding, h.store.rows[incomeID].Withholding, "and stored")
	assert.Equal(t, "Diseño de logotipo", got.Income.Description, "absent fields are unchanged")
	assert.False(t, got.Created)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, 1, h.access.ensures, "edits need write access")
	assert.Zero(t, h.access.reserves, "an edit never takes a slot")
	assert.Equal(t, []regimeCall{{taxpayerA, "612", "2026-09"}}, h.taxpayers.calls)
	assert.Contains(t, h.logs.String(), `"msg":"income.updated"`)
}

func TestUpdate_NullClearsANullableField(t *testing.T) {
	h := seeded(t)
	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Description: ptr(""), CFDIUUID: ptr(""), InvoicedOn: ptr(fiscal.Date{}),
	})
	require.NoError(t, err)
	assert.Empty(t, got.Income.Description)
	assert.True(t, got.Income.CFDIUUID.IsZero())
	assert.True(t, got.Income.Dates.InvoicedOn.IsZero())
}

func TestUpdate_AccrualIncomeGetsCollected(t *testing.T) {
	h := newHarness(t)
	in := usecase.IncomeInput{
		Regime: "601", Counterparty: fiscal.ReceptorPM, Amount: 45_000_00,
		IVARate: fiscal.IVARate16, IVA: 7_200_00, Dates: domain.Dates{InvoicedOn: date(t, "2026-09-05")},
	}
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, pmA, in))
	require.NoError(t, err)
	got, err := h.uc.Update(t.Context(), userA, pmA, incomeID, usecase.Patch{CollectedOn: ptr(date(t, "2026-10-02"))})
	require.NoError(t, err)
	assert.False(t, got.Income.PendingCollection, "it loses Por cobrar")
	assert.Equal(t, "2026-09-05", got.Income.BookedOn.String(), "booked on the invoice month")
}

func TestUpdate_Refusals(t *testing.T) {
	tests := []struct {
		name     string
		taxpayer string
		id       string
		patch    usecase.Patch
		prepare  func(*testing.T, *harness)
		code     apperr.Code
		field    []string
	}{
		{name: "read-only user", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Description: ptr("x")},
			prepare: func(_ *testing.T, h *harness) { h.access.readOnly = true }, code: apperr.CodeEntitlementRequired},
		{name: "regime not held on the date", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Regime: ptr(fiscal.RegimeCode("626"))},
			code: apperr.CodeBusinessRuleViolation, field: []string{"regime_code", "regime_not_active"}},
		{name: "another taxpayer's path", taxpayer: taxpayerA2, id: incomeID, patch: usecase.Patch{Description: ptr("x")}, code: apperr.CodeNotFound},
		{name: "unknown income", taxpayer: taxpayerA, id: generated, patch: usecase.Patch{Description: ptr("x")}, code: apperr.CodeNotFound},
		{name: "malformed income id", taxpayer: taxpayerA, id: "nope", patch: usecase.Patch{Description: ptr("x")}, code: apperr.CodeNotFound},
		{name: "malformed taxpayer id", taxpayer: "nope", id: incomeID, patch: usecase.Patch{Description: ptr("x")}, code: apperr.CodeNotFound},
		{name: "soft-deleted income", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Description: ptr("x")},
			prepare: func(t *testing.T, h *harness) {
				require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
			},
			code: apperr.CodeNotFound},
		{name: "archived taxpayer", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Description: ptr("x")},
			prepare: func(_ *testing.T, h *harness) {
				p := h.taxpayers.profiles[taxpayerA]
				p.Archived = true
				h.taxpayers.profiles[taxpayerA] = p
			}, code: apperr.CodeBusinessRuleViolation, field: []string{"taxpayerId", "taxpayer_archived"}},
		{name: "domain rule", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{IVA: ptr(money.Cents(1))},
			code: apperr.CodeBusinessRuleViolation, field: []string{"iva_cents", "iva_amount_mismatch"}},
		{name: "description too long", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Description: ptr(strings.Repeat("x", 201))},
			code: apperr.CodeValidationFailed, field: []string{"description", "too_long"}},
		{name: "description holding a BOM", taxpayer: taxpayerA, id: incomeID, patch: usecase.Patch{Description: ptr("\ufeffpago")},
			code: apperr.CodeValidationFailed, field: []string{"description", "invisible_characters"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := seeded(t)
			if tt.prepare != nil {
				tt.prepare(t, h)
			}
			updates := h.store.updates
			_, err := h.uc.Update(t.Context(), userA, tt.taxpayer, tt.id, tt.patch)
			requireApp(t, err, tt.code, tt.field...)
			assert.Equal(t, updates, h.store.updates, "nothing changes")
		})
	}
}

func TestUpdate_EveryFieldPatches(t *testing.T) {
	h := seeded(t)
	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Regime: ptr(fiscal.RegimeCode("612")), Counterparty: ptr(fiscal.ReceptorPFNoActivityOrPublic),
		Operation: ptr(fiscal.OperationSaleOfGoods), ResidentialRent: ptr(false), ViaPlatform: ptr(false),
		Amount: ptr(money.Cents(100_00)), IVARate: ptr(fiscal.IVAExempt), IVA: ptr(money.Cents(0)),
		CollectedOn: ptr(date(t, "2026-09-30")), PlatformActivity: ptr(domain.PlatformActivity("")),
		CFDIUUID: ptr(cfdiLower), Description: ptr("Venta"),
	})
	require.NoError(t, err)
	assert.Equal(t, fiscal.ReceptorPFNoActivityOrPublic, got.Income.Counterparty)
	assert.Equal(t, fiscal.OperationSaleOfGoods, got.Income.Operation)
	assert.Equal(t, fiscal.IVAExempt, got.Income.IVARate)
	assert.Equal(t, "2026-09-30", got.Income.BookedOn.String())
	assert.Equal(t, "Venta", got.Income.Description)
}

// US-W9: changing who was billed re-derives (ISR and IVA withheld become 0).
func TestUpdate_ChangingTheCounterpartyReDerives(t *testing.T) {
	h := seeded(t)
	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Counterparty: ptr(fiscal.ReceptorPFWithActivity),
	})
	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{}, got.Income.Withholding.Withholdings)
	assert.Equal(t, []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA1}, got.Income.Withholding.RuleIDs)
}

// An income stored before WH-06 is served with the backfill's facts
// (0.12.0), and the app patches only what the user changed: an edit applies
// to the served facts and derives. Answering who was billed overrides them.
func TestUpdate_ALegacyIncomeEditsOnTheServedFacts(t *testing.T) {
	h := seeded(t)
	legacy := h.store.rows[incomeID]
	legacy.Counterparty, legacy.Operation, legacy.Withholding = "", "", fiscal.WithholdingRecord{}
	h.store.rows[incomeID] = legacy

	got, err := h.uc.Update(t.Context(), userA, taxpayerA, incomeID, usecase.Patch{
		Counterparty: ptr(fiscal.ReceptorPFWithActivity),
	})
	require.NoError(t, err)
	assert.Equal(t, fiscal.ReceptorPFWithActivity, got.Income.Counterparty)
	assert.Equal(t, fiscal.OperationProfessionalServices, got.Income.Operation, "the regime default")
	assert.True(t, got.Income.Withholding.Derived())
}
