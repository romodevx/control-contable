package taxparams_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

func TestNewCatalogRejects(t *testing.T) {
	_, err := taxparams.NewCatalog()
	require.ErrorIs(t, err, taxparams.ErrNoParameterSets)

	_, err = taxparams.NewCatalog(withYear(t, 2026, 1, false), withYear(t, 2026, 1, true))
	require.ErrorIs(t, err, taxparams.ErrDuplicateVersion)

	_, err = taxparams.NewCatalog(withYear(t, 2026, 1, false), nil)
	require.ErrorIs(t, err, taxparams.ErrNoParameterSets)
}

func TestSelect(t *testing.T) {
	catalog, err := taxparams.NewCatalog(
		// Out of order on purpose: selection never depends on input order.
		withYear(t, 2026, 2, false),
		withYear(t, 2025, 1, false),
		withYear(t, 2026, 1, false),
		withYear(t, 2028, 1, true),
	)
	require.NoError(t, err)

	tests := []struct {
		name            string
		fiscalYear      int
		wantYear        int
		wantVersion     int
		wantProvisional bool
		wantFallback    bool
		wantErr         error
	}{
		{name: "highest version of the year", fiscalYear: 2026, wantYear: 2026, wantVersion: 2},
		{name: "single version", fiscalYear: 2025, wantYear: 2025, wantVersion: 1},
		{name: "next year not published: previous year, provisional", fiscalYear: 2027, wantYear: 2026, wantVersion: 2, wantProvisional: true, wantFallback: true},
		{name: "a set marked provisional is surfaced as provisional", fiscalYear: 2028, wantYear: 2028, wantVersion: 1, wantProvisional: true},
		{name: "never more than one year back", fiscalYear: 2030, wantErr: taxparams.ErrNoParameters},
		{name: "never forward to a later year", fiscalYear: 2024, wantErr: taxparams.ErrNoParameters},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := catalog.Select(tt.fiscalYear)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.fiscalYear, sel.RequestedYear)
			assert.Equal(t, tt.wantYear, sel.Params.FiscalYear())
			assert.Equal(t, tt.wantVersion, sel.Params.Version())
			assert.Equal(t, tt.wantProvisional, sel.Provisional)
			assert.Equal(t, tt.wantFallback, sel.Fallback)
		})
	}
}

func TestSelectFallbackKeepsDateValidity(t *testing.T) {
	// A fallback set is still resolved by date: the open-ended tariff carries
	// into the unpublished year, the rate that lapsed on 2026-12-31 does not.
	catalog, err := taxparams.NewCatalog(withYear(t, 2026, 1, false))
	require.NoError(t, err)
	sel, err := catalog.Select(2027)
	require.NoError(t, err)
	require.True(t, sel.Provisional)

	jan2027 := mustDate(t, "2027-01-31")
	_, err = sel.Params.Table(keyMonthly, jan2027)
	require.NoError(t, err)
	_, err = sel.Params.Rate(keyBorderRate, jan2027)
	require.ErrorIs(t, err, taxparams.ErrNotInForce)
}

func TestCatalogVersions(t *testing.T) {
	catalog, err := taxparams.NewCatalog(withYear(t, 2026, 2, false), withYear(t, 2025, 1, false), withYear(t, 2026, 1, false))
	require.NoError(t, err)
	assert.Equal(t, []taxparams.SetID{{FiscalYear: 2025, Version: 1}, {FiscalYear: 2026, Version: 1}, {FiscalYear: 2026, Version: 2}}, catalog.Sets())
	assert.Equal(t, "2026/v2", taxparams.SetID{FiscalYear: 2026, Version: 2}.String())
}
