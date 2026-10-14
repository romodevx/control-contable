package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The ports M2 offers M3, M4 and M5 are satisfied by Queries.
var (
	_ usecase.TaxpayerRegimeChecker = (*usecase.Queries)(nil)
	_ usecase.TaxpayerProfileReader = (*usecase.Queries)(nil)
)

func ctxAs(t *testing.T, userID string) context.Context {
	t.Helper()
	return context.WithValue(t.Context(), userKey{}, userID)
}

func TestCheckRegime(t *testing.T) {
	const archived = "0199d4a2-0000-7000-8000-0000000000a9"
	tests := []struct {
		name     string
		user     string
		taxpayer string
		code     string
		period   string
		want     apperr.Code
		target   error
		fields   [][2]string
	}{
		{name: "inside an open window", user: userA, taxpayer: taxpayerA, code: "612", period: "2026-07"},
		{name: "first month of the window", user: userA, taxpayer: taxpayerA, code: "626", period: "2026-01"},
		{name: "last month of an ended window", user: userA, taxpayer: taxpayerA, code: "626", period: "2026-06"},
		{name: "after the end", user: userA, taxpayer: taxpayerA, code: "626", period: "2026-07",
			want: apperr.CodeBusinessRuleViolation, target: usecase.ErrRegimeNotActive, fields: [][2]string{{"regime_code", usecase.DetailRegimeNotActive}}},
		{name: "before the start", user: userA, taxpayer: taxpayerA, code: "612", period: "2026-06",
			want: apperr.CodeBusinessRuleViolation, target: usecase.ErrRegimeNotActive},
		{name: "regime not held", user: userA, taxpayer: taxpayerA, code: "606", period: "2026-07",
			want: apperr.CodeBusinessRuleViolation, target: usecase.ErrRegimeNotActive},
		{name: "archived taxpayer", user: userA, taxpayer: archived, code: "612", period: "2026-07",
			want: apperr.CodeBusinessRuleViolation, target: usecase.ErrTaxpayerArchived, fields: [][2]string{{"taxpayerId", usecase.DetailTaxpayerArchived}}},
		{name: "another user's taxpayer", user: userB, taxpayer: taxpayerA, code: "612", period: "2026-07",
			want: apperr.CodeNotFound, target: usecase.ErrTaxpayerNotFound},
		{name: "malformed taxpayer id", user: userA, taxpayer: "x", code: "612", period: "2026-07", want: apperr.CodeNotFound},
		{name: "malformed period", user: userA, taxpayer: taxpayerA, code: "612", period: "2026-7",
			want: apperr.CodeValidationFailed, target: domain.ErrInvalidPeriod, fields: [][2]string{{"period", usecase.DetailInvalid}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, false,
				seedRegime{code: "626", window: window(t, "2026-01", "2026-06")},
				seedRegime{code: "612", window: window(t, "2026-07", "")})
			h.store.seedTaxpayer(userA, archived, domain.PersonTypeFisica, true, seedRegime{code: "612", window: window(t, "2026-01", "")})

			err := h.queries.CheckRegime(t.Context(), tt.user, tt.taxpayer, tt.code, tt.period)

			if tt.want == "" {
				require.NoError(t, err)
				assert.Equal(t, []string{"rw"}, h.tx.modes, "a write-side check locks, so it needs a read-write unit of work")
				assert.Equal(t, []string{
					fmt.Sprintf("taxpayer:%d", usecase.LockShare), fmt.Sprintf("regime:%d", usecase.LockShare),
				}, h.store.locks, "FOR SHARE: a concurrent archive or end-regime waits for the record write")
				return
			}
			requireAppErr(t, err, tt.want, tt.target, tt.fields...)
		})
	}
}

func TestProfile(t *testing.T) {
	h := newHarness(t)
	h.store.seedTaxpayer(userA, taxpayerA, domain.PersonTypeFisica, true,
		seedRegime{code: "626", window: window(t, "2026-01", "2026-06")},
		seedRegime{code: "606", window: window(t, "2026-07", ""), options: map[string]domain.OptionValue{
			"deduction_mode": domain.StringOption("actual")}})

	got, err := h.queries.Profile(t.Context(), userA, taxpayerA)

	require.NoError(t, err)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
	assert.Equal(t, domain.PersonTypeFisica, got.PersonType)
	assert.True(t, got.Archived, "reads of an archived taxpayer still work (past estimates)")
	require.Len(t, got.Regimes, 2)
	assert.Equal(t, "626", got.Regimes[0].Code)
	assert.Equal(t, window(t, "2026-01", "2026-06"), got.Regimes[0].Window)
	assert.Equal(t, "606", got.Regimes[1].Code)
	assert.Equal(t, "full", got.Regimes[1].Treatment)
	assert.Equal(t, domain.StringOption("actual"), got.Regimes[1].Options["deduction_mode"])
	assert.Equal(t, []string{"ro"}, h.tx.modes)

	_, err = h.queries.Profile(t.Context(), userB, taxpayerA)
	requireAppErr(t, err, apperr.CodeNotFound, usecase.ErrTaxpayerNotFound)
}
