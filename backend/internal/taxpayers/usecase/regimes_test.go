package usecase_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	taxpayerA = "0199d4a2-0000-7000-8000-0000000000a1"
	taxpayerB = "0199d4a2-0000-7000-8000-0000000000b1"
)

func TestAddRegime_Adds(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "612", window: window(t, "2026-01", "")})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "606", Start: "2026-03"})

	require.NoError(t, err)
	assert.Equal(t, "606", got.Code)
	assert.Equal(t, "Regime 606", got.Name)
	assert.Equal(t, "full", got.Treatment)
	assert.Equal(t, window(t, "2026-03", ""), got.Window)
	assert.Equal(t, map[string]domain.OptionValue{"deduction_mode": domain.StringOption("blind_35")}, got.Options)
	assert.Empty(t, got.Warnings)
	assert.Contains(t, h.store.locks, fmt.Sprintf("taxpayer:%d", usecase.LockUpdate), "regime writes of one taxpayer are serialized")
	assert.Contains(t, h.logs.String(), `"msg":"taxpayer_regime.added"`)
	assert.Contains(t, h.logs.String(), `"regime_code":"606"`)
}

func TestAddRegime_DefaultsTheStartToTheCurrentMonth(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "612", window: window(t, "2026-01", "")})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "606"})

	require.NoError(t, err)
	assert.Equal(t, window(t, "2026-10", ""), got.Window)
}

func TestAddRegime_IncompatibleOverlappingRegimeIs422(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "612", window: window(t, "2026-01", "")})

	_, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "626", Start: "2026-01"})

	msg := requireAppErr(t, err, apperr.CodeInvalidRegimeCombination, usecase.ErrIncompatibleRegimes,
		[2]string{"regime_code", usecase.DetailIncompatible})
	assert.Equal(t, "El RESICO sustituye al 612.", msg)
	assert.NotContains(t, h.store.regimes, taxpayerA+"/626")
}

func TestAddRegime_ExitFromRESICOTheNextMonthIsAllowed(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "626", window: window(t, "2026-01", "2026-06")})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "612", Start: "2026-07"})

	require.NoError(t, err)
	assert.Equal(t, window(t, "2026-07", ""), got.Window)
}

func TestAddRegime_PairUnderCPAReviewWarns(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "605", window: window(t, "2026-01", "")})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "607"})

	require.NoError(t, err)
	assert.Equal(t, []usecase.Warning{{Code: usecase.WarningRegimeNeedsCPAReview, Field: "regime_code", Message: "Pendiente de revisión por un contador."}}, got.Warnings)
	assert.True(t, h.store.regimes[taxpayerA+"/607"].ack)
}

func TestAddRegime_Refusals(t *testing.T) {
	tests := []struct {
		name     string
		readOnly bool
		taxpayer string
		in       usecase.RegimeInput
		code     apperr.Code
		target   error
		fields   [][2]string
	}{
		{"read-only user", true, taxpayerA, usecase.RegimeInput{Code: "606"}, apperr.CodeEntitlementRequired, access.ErrReadOnly, nil},
		{"another user's taxpayer", false, taxpayerB, usecase.RegimeInput{Code: "606"}, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound, nil},
		{"unknown taxpayer id", false, "not-a-uuid", usecase.RegimeInput{Code: "606"}, apperr.CodeNotFound, nil, nil},
		{"archived taxpayer", false, "0199d4a2-0000-7000-8000-0000000000a9", usecase.RegimeInput{Code: "606"},
			apperr.CodeBusinessRuleViolation, usecase.ErrTaxpayerArchived, [][2]string{{"taxpayerId", usecase.DetailTaxpayerArchived}}},
		{"moral-only code for a física", false, taxpayerA, usecase.RegimeInput{Code: "601"}, apperr.CodeValidationFailed,
			usecase.ErrRegimeNotForPersonType, [][2]string{{"regime_code", usecase.DetailNotForPersonType}}},
		{"already active", false, taxpayerA, usecase.RegimeInput{Code: "612"}, apperr.CodeConflict, usecase.ErrRegimeAlreadyActive, nil},
		{"bad options", false, taxpayerA, usecase.RegimeInput{Code: "625", Options: map[string]domain.OptionValue{
			"activities": domain.ListOption(nil), "extra": domain.BoolOption(true)}}, apperr.CodeValidationFailed, domain.ErrOptionValue,
			[][2]string{{"options.extra", usecase.DetailUnknownOption}, {"options.activities", usecase.DetailInvalidOption}}},
		{"start before the first parameter year", false, taxpayerA, usecase.RegimeInput{Code: "606", Start: "2025-12"}, apperr.CodeValidationFailed,
			usecase.ErrPeriodOutOfRange, [][2]string{{"start_period", usecase.DetailOutOfRange}}},
		{"start in the future", false, taxpayerA, usecase.RegimeInput{Code: "606", Start: "2026-11"}, apperr.CodeValidationFailed,
			usecase.ErrPeriodOutOfRange, [][2]string{{"start_period", usecase.DetailOutOfRange}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "612", window: window(t, "2026-01", "")})
			h.store.seedTaxpayer(userA, "0199d4a2-0000-7000-8000-0000000000a9", domain.PersonTypeFisica, true,
				seedRegime{code: "612", window: window(t, "2026-01", "")})
			h.store.seedTaxpayer(userB, taxpayerB, domain.PersonTypeFisica, false)
			h.access.readOnly = tt.readOnly
			before := len(h.store.regimes)

			_, err := h.regimes.Add(t.Context(), userA, tt.taxpayer, tt.in)

			requireAppErr(t, err, tt.code, tt.target, tt.fields...)
			assert.Len(t, h.store.regimes, before)
		})
	}
}

func TestAddRegime_ReopensAnEndedRegimeKeepingItsStart(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false,
		seedRegime{code: "612", window: window(t, "2025-01", "")},
		seedRegime{code: "606", window: window(t, "2024-01", "2024-06"), options: map[string]domain.OptionValue{
			"deduction_mode": domain.StringOption("blind_35")}})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "606", Start: "2026-02", Options: map[string]domain.OptionValue{
		"deduction_mode": domain.StringOption("actual")}})

	require.NoError(t, err)
	assert.Equal(t, window(t, "2024-01", ""), got.Window, "OQ-5: the row reopens; its start is kept")
	assert.Equal(t, map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")}, got.Options, "body options applied")
}

func TestAddRegime_ReopenChecksCompatibilityOnTheStoredWindowNotTheBodyStart(t *testing.T) {
	// 626 held 2026-01..2026-02; 606 held 2026-03..2026-04. Reopening 606 spans
	// 2026-03..open, which never meets 626. A body start of 2026-02 must not be
	// used: it would invent an overlap with 626 (the false 23P01 of PR #57).
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false,
		seedRegime{code: "626", window: window(t, "2026-01", "2026-02")},
		seedRegime{code: "606", window: window(t, "2026-03", "2026-04")},
		seedRegime{code: "612", window: window(t, "2026-05", "")})

	got, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "606", Start: "2026-02"})

	require.NoError(t, err)
	assert.Equal(t, window(t, "2026-03", ""), got.Window)
}

func TestAddRegime_ReopenStillRefusesARealOverlap(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false,
		seedRegime{code: "606", window: window(t, "2024-01", "2024-06")},
		seedRegime{code: "626", window: window(t, "2025-01", "")})

	_, err := h.regimes.Add(t.Context(), userA, taxpayerA, usecase.RegimeInput{Code: "606"})

	requireAppErr(t, err, apperr.CodeInvalidRegimeCombination, usecase.ErrIncompatibleRegimes)
	assert.False(t, h.store.regimes[taxpayerA+"/606"].held.Window.IsOpen())
}

func TestUpdateRegimeOptions_MergesAndValidates(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "625", window: window(t, "2026-01", ""),
		options: map[string]domain.OptionValue{
			"activities":            domain.ListOption([]string{"lodging"}),
			"definitive_option":     domain.BoolOption(false),
			"rfc_given_to_platform": domain.BoolOption(true),
		}})

	got, err := h.regimes.UpdateOptions(t.Context(), userA, taxpayerA, "625", map[string]domain.OptionValue{
		"definitive_option": domain.BoolOption(true),
	})

	require.NoError(t, err)
	assert.Equal(t, map[string]domain.OptionValue{
		"activities":            domain.ListOption([]string{"lodging"}),
		"definitive_option":     domain.BoolOption(true),
		"rfc_given_to_platform": domain.BoolOption(true),
	}, got.Options)

	_, err = h.regimes.UpdateOptions(t.Context(), userA, taxpayerA, "625", map[string]domain.OptionValue{
		"activities": domain.ListOption([]string{}),
	})
	requireAppErr(t, err, apperr.CodeValidationFailed, domain.ErrOptionValue, [2]string{"options.activities", usecase.DetailInvalidOption})
}

func TestUpdateRegimeOptions_Refusals(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false, seedRegime{code: "606", window: window(t, "2026-01", ""),
		options: map[string]domain.OptionValue{"deduction_mode": domain.StringOption("blind_35")}})
	h.store.seedTaxpayer(userA, "0199d4a2-0000-7000-8000-0000000000a9", domain.PersonTypeFisica, true,
		seedRegime{code: "606", window: window(t, "2026-01", "")})
	opts := map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")}

	_, err := h.regimes.UpdateOptions(t.Context(), userA, taxpayerA, "612", opts)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrRegimeNotHeld)

	_, err = h.regimes.UpdateOptions(t.Context(), userA, "0199d4a2-0000-7000-8000-0000000000a9", "606", opts)
	requireAppErr(t, err, apperr.CodeBusinessRuleViolation, usecase.ErrTaxpayerArchived)

	_, err = h.regimes.UpdateOptions(t.Context(), userB, taxpayerA, "606", opts)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)

	h.access.readOnly = true
	_, err = h.regimes.UpdateOptions(t.Context(), userA, taxpayerA, "606", opts)
	requireAppErr(t, err, apperr.CodeEntitlementRequired, access.ErrReadOnly)
	assert.Equal(t, domain.StringOption("blind_35"), h.store.regimes[taxpayerA+"/606"].held.Options["deduction_mode"])
}

func seedTwoRegimes(t *testing.T, h *harness) {
	t.Helper()
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false,
		seedRegime{code: "606", window: window(t, "2026-01", "")},
		seedRegime{code: "612", window: window(t, "2026-01", "")})
}

func TestEndRegime_Ends(t *testing.T) {
	h := newHarness(t)
	seedTwoRegimes(t, h)
	h.access.readOnly = true // ending a regime is an archive: allowed read-only

	got, err := h.regimes.End(t.Context(), userA, taxpayerA, usecase.EndRegimeInput{Code: "606", End: "2026-09"})

	require.NoError(t, err)
	assert.Equal(t, window(t, "2026-01", "2026-09"), got.Window)
	assert.Contains(t, h.store.locks, fmt.Sprintf("regime:%d", usecase.LockUpdate))
	assert.Contains(t, h.logs.String(), `"msg":"taxpayer_regime.archived"`)
}

func TestEndRegime_WithRecordsNeedsConfirm(t *testing.T) {
	h := newHarness(t)
	seedTwoRegimes(t, h)
	h.records[taxpayerA+"/606"] = period(t, "2026-08")

	_, err := h.regimes.End(t.Context(), userA, taxpayerA, usecase.EndRegimeInput{Code: "606", End: "2026-09"})
	requireAppErr(t, err, apperr.CodeRegimeHasRecords, usecase.ErrRegimeHasRecords)
	assert.True(t, h.store.regimes[taxpayerA+"/606"].held.Window.IsOpen())

	got, err := h.regimes.End(t.Context(), userA, taxpayerA, usecase.EndRegimeInput{Code: "606", End: "2026-09", Confirm: true})
	require.NoError(t, err)
	assert.Equal(t, period(t, "2026-09"), got.Window.End)
}

func TestEndRegime_Refusals(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T, h *harness)
		in     usecase.EndRegimeInput
		code   apperr.Code
		target error
		fields [][2]string
	}{
		{"records after the end", func(t *testing.T, h *harness) { h.records[taxpayerA+"/606"] = period(t, "2026-08") },
			usecase.EndRegimeInput{Code: "606", End: "2026-07", Confirm: true}, apperr.CodeBusinessRuleViolation,
			usecase.ErrRecordsAfterEndPeriod, [][2]string{{"end_period", usecase.DetailRecordsAfterEndPeriod}}},
		{"last active regime", func(t *testing.T, h *harness) {
			require.NoError(t, h.store.EndRegime(ctxAs(t, userA), taxpayerA, "612", period(t, "2026-05")))
		}, usecase.EndRegimeInput{Code: "606", End: "2026-09"}, apperr.CodeBusinessRuleViolation,
			usecase.ErrLastActiveRegime, [][2]string{{"regimeCode", usecase.DetailLastActiveRegime}}},
		{"already ended", func(t *testing.T, h *harness) {
			require.NoError(t, h.store.EndRegime(ctxAs(t, userA), taxpayerA, "606", period(t, "2026-05")))
		}, usecase.EndRegimeInput{Code: "606", End: "2026-09"}, apperr.CodeConflict, usecase.ErrRegimeAlreadyEnded, nil},
		{"end before start", func(*testing.T, *harness) {}, usecase.EndRegimeInput{Code: "606", End: "2025-12"},
			apperr.CodeValidationFailed, usecase.ErrPeriodOutOfRange, [][2]string{{"end_period", usecase.DetailOutOfRange}}},
		{"end after the current month", func(*testing.T, *harness) {}, usecase.EndRegimeInput{Code: "606", End: "2026-11"},
			apperr.CodeValidationFailed, usecase.ErrPeriodOutOfRange, [][2]string{{"end_period", usecase.DetailOutOfRange}}},
		{"malformed end", func(*testing.T, *harness) {}, usecase.EndRegimeInput{Code: "606", End: "2026-9"},
			apperr.CodeValidationFailed, domain.ErrInvalidPeriod, [][2]string{{"end_period", usecase.DetailInvalid}}},
		{"regime not held", func(*testing.T, *harness) {}, usecase.EndRegimeInput{Code: "626", End: "2026-09"},
			apperr.CodeNotFound, usecase.ErrRegimeNotHeld, nil},
		{"archived taxpayer", func(t *testing.T, h *harness) {
			require.NoError(t, h.store.SetArchivedAt(ctxAs(t, userA), taxpayerA, today))
		}, usecase.EndRegimeInput{Code: "606", End: "2026-09"}, apperr.CodeBusinessRuleViolation,
			usecase.ErrTaxpayerArchived, [][2]string{{"taxpayerId", usecase.DetailTaxpayerArchived}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			seedTwoRegimes(t, h)
			tt.setup(t, h)
			before := h.store.regimes[taxpayerA+"/606"].held.Window

			_, err := h.regimes.End(t.Context(), userA, taxpayerA, tt.in)

			requireAppErr(t, err, tt.code, tt.target, tt.fields...)
			assert.Equal(t, before, h.store.regimes[taxpayerA+"/606"].held.Window)
		})
	}
}

func TestEndRegime_AnotherUsersTaxpayerIs404(t *testing.T) {
	h := newHarness(t)
	seedTwoRegimes(t, h)

	_, err := h.regimes.End(t.Context(), userB, taxpayerA, usecase.EndRegimeInput{Code: "606", End: "2026-09"})

	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)
}
