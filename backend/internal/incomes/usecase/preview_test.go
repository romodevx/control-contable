package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
)

// previewUS1 is US-I1 typed into the form: 612 fees to a company.
func previewUS1() usecase.PreviewQuery {
	return usecase.PreviewQuery{
		TaxpayerID: taxpayerA, Regime: "612", On: "2026-09-12", Amount: 10_000_00, IVARate: fiscal.IVARate16,
		IVA: 1_600_00, Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationProfessionalServices,
	}
}

// The preview derives what the create would store, in a read-only unit of
// work that stores nothing.
func TestPreview_DerivesWithoutStoring(t *testing.T) {
	h := newHarness(t)

	got, err := h.uc.Preview(t.Context(), userA, previewUS1())

	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67}, got.Withholding.Withholdings)
	assert.Equal(t, []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, got.Withholding.RuleIDs)
	assert.Equal(t, fiscal.ReceptorPM, got.Counterparty)
	assert.EqualValues(t, 10_000_00, got.Amount)
	assert.EqualValues(t, 1_600_00, got.IVA)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, []string{"ro"}, h.tx.modes)
	assert.Zero(t, h.store.inserts+h.store.updates)
	assert.Zero(t, h.access.ensures, "a read: allowed read-only")
	assert.Empty(t, h.taxpayers.calls, "no window lock")

	create := us1(t)
	created, err := h.uc.Create(t.Context(), userA, usecase.CreateInput{ID: incomeID, TaxpayerID: taxpayerA, Income: create})
	require.NoError(t, err)
	assert.Equal(t, created.Income.Withholding, got.Withholding, "the same derivation as the create")
}

func TestPreview_DefaultsAndWarnings(t *testing.T) {
	t.Run("625 defaults: via the platform, the only activity", func(t *testing.T) {
		h := newHarness(t)
		got, err := h.uc.Preview(t.Context(), userA, usecase.PreviewQuery{
			TaxpayerID: platformA, Regime: "625", On: "2026-09-12", Amount: 15_000_00, IVARate: fiscal.IVARate16,
			IVA: 2_400_00, Counterparty: fiscal.ReceptorPFNoActivityOrPublic,
		})
		require.NoError(t, err)
		assert.Equal(t, fiscal.Withholdings{ISR: 600_00, IVA: 1_200_00}, got.Withholding.Withholdings)
		assert.Equal(t, domain.ActivityLodging, got.PlatformActivity)
		assert.Equal(t, fiscal.WithholderPlatform, got.Withholding.Withholder)
	})
	t.Run("an unverified rule warns, only the withholding warnings", func(t *testing.T) {
		h := newHarness(t).at(time.Date(2027, time.February, 1, 12, 0, 0, 0, clock.Location()))
		q := previewUS1()
		q.Counterparty, q.On = fiscal.ReceptorGovFederal, "2027-01-05"
		got, err := h.uc.Preview(t.Context(), userA, q)
		require.NoError(t, err)
		assert.Equal(t, []domain.Warning{{Code: domain.WarningWithholdingCPAUnverified, Field: domain.FieldWithholding}},
			got.Warnings, "no provisional_parameters in a preview")
		assert.True(t, got.Withholding.CPAUnverified)
	})
	t.Run("on defaults to today", func(t *testing.T) {
		h := newHarness(t)
		q := previewUS1()
		q.On = ""
		_, err := h.uc.Preview(t.Context(), userA, q)
		require.NoError(t, err)
	})
	t.Run("an archived taxpayer previews", func(t *testing.T) {
		h := newHarness(t)
		q := previewUS1()
		q.TaxpayerID = archivedA
		_, err := h.uc.Preview(t.Context(), userA, q)
		require.NoError(t, err)
	})
}

func TestPreview_Refusals(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(*usecase.PreviewQuery)
		code  apperr.Code
		field []string
	}{
		{"malformed taxpayer id", func(q *usecase.PreviewQuery) { q.TaxpayerID = "x" }, apperr.CodeNotFound, nil},
		{"unknown taxpayer", func(q *usecase.PreviewQuery) { q.TaxpayerID = "0199d4a2-1111-7000-8000-0000000000ff" },
			apperr.CodeNotFound, nil},
		{"no regime", func(q *usecase.PreviewQuery) { q.Regime = "" }, apperr.CodeValidationFailed, []string{"regime", "required"}},
		{"a future day", func(q *usecase.PreviewQuery) { q.On = "2026-10-14" }, apperr.CodeValidationFailed, []string{"on", "out_of_range"}},
		{"before the first parameter set", func(q *usecase.PreviewQuery) { q.On = "2025-06-01" },
			apperr.CodeValidationFailed, []string{"on", "out_of_range"}},
		{"a malformed day", func(q *usecase.PreviewQuery) { q.On = "2026-13-01" }, apperr.CodeValidationFailed, []string{"on", "invalid"}},
		{"a regime not held on the day", func(q *usecase.PreviewQuery) { q.Regime, q.On = "606", "2026-08-31" },
			apperr.CodeBusinessRuleViolation, []string{"regime", "regime_not_active"}},
		{"no counterparty kind", func(q *usecase.PreviewQuery) { q.Counterparty = "" },
			apperr.CodeValidationFailed, []string{"counterparty_kind", "required"}},
		{"via a platform on a 612", func(q *usecase.PreviewQuery) { q.ViaPlatform = ptr(true) },
			apperr.CodeValidationFailed, []string{"via_platform", "not_for_regime"}},
		{"IVA off the rate", func(q *usecase.PreviewQuery) { q.IVA = 1_500_00 },
			apperr.CodeBusinessRuleViolation, []string{"iva_cents", "iva_amount_mismatch"}},
		{"housing rent with IVA", func(q *usecase.PreviewQuery) {
			q.Operation, q.ResidentialRent = fiscal.OperationRentRealEstate, true
		}, apperr.CodeBusinessRuleViolation, []string{"iva_rate", "residential_rent_exempt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			q := previewUS1()
			tt.edit(&q)
			_, err := h.uc.Preview(t.Context(), userA, q)
			requireApp(t, err, tt.code, tt.field...)
			assert.Zero(t, h.store.inserts+h.store.updates)
		})
	}
}

func TestPreview_NoUserNoClock(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Preview(t.Context(), "", previewUS1())
	require.Error(t, err)
}
