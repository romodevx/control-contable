package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	taxpayersdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// stubQueries is the taxpayers module's read side.
type stubQueries struct {
	profile taxpayers.TaxpayerProfile
	err     error
	checked []string
}

func (s *stubQueries) CheckRegime(_ context.Context, _, taxpayerID, regime, period string) error {
	s.checked = append(s.checked, taxpayerID+"|"+regime+"|"+period)
	return s.err
}

func (s *stubQueries) Profile(context.Context, string, string) (taxpayers.TaxpayerProfile, error) {
	return s.profile, s.err
}

// The adapter hands incomes the payee facts the withholding rules read: the
// person type, the archive, and the 625 options.
func TestIncomesTaxpayers_Profile(t *testing.T) {
	q := &stubQueries{profile: taxpayers.TaxpayerProfile{
		TaxpayerID: "t", PersonType: taxpayersdomain.PersonTypeFisica, Archived: true,
		Regimes: []taxpayers.RegimeProfile{
			{Code: "612", Options: map[string]taxpayersdomain.OptionValue{"iva_border_region": taxpayersdomain.BoolOption(false)},
				Window: taxpayersdomain.Window{Start: period(t, "2025-02"), End: period(t, "2026-06")}},
			{Code: "625", Window: taxpayersdomain.Window{Start: period(t, "2026-03")}, Options: map[string]taxpayersdomain.OptionValue{
				"activities":            taxpayersdomain.ListOption([]string{"lodging", "goods_services"}),
				"rfc_given_to_platform": taxpayersdomain.BoolOption(true),
			}},
		},
	}}
	got, err := incomesTaxpayers{queries: q}.Profile(t.Context(), "u", "t")
	require.NoError(t, err)
	assert.Equal(t, incomes.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Archived: true,
		Regimes: []incomes.RegimeProfile{
			{Code: "612", StartPeriod: "2025-02", EndPeriod: "2026-06"},
			{Code: "625", StartPeriod: "2026-03", PlatformActivities: []domain.PlatformActivity{domain.ActivityLodging, domain.ActivityGoodsServices}, RFCGivenToPlatform: true},
		},
	}, got)
}

func TestIncomesTaxpayers_PassesErrorsThrough(t *testing.T) {
	refusal := errors.New("404")
	q := &stubQueries{err: refusal}
	adapter := incomesTaxpayers{queries: q}
	_, err := adapter.Profile(t.Context(), "u", "t")
	require.ErrorIs(t, err, refusal)
	require.ErrorIs(t, adapter.CheckRegime(t.Context(), "u", "t", "612", "2026-09"), refusal)
	assert.Equal(t, []string{"t|612|2026-09"}, q.checked)
}

// The composition root binds the rate catalog once, over the embedded
// parameters (the boot fails on a set without the IVA rates).
func TestNewRateCatalog(t *testing.T) {
	rates, err := newRateCatalog()
	require.NoError(t, err)
	_, ok := rates.Rate(fiscal.KeyIVAGeneral, mustDate(t, "2026-09-10"))
	assert.True(t, ok)
	var _ incomes.RateCatalog = rates
}

func mustDate(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}
