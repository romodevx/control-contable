package estimates

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	taxdomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// fakeProfiles answers a canned taxpayers profile.
type fakeProfiles struct {
	profile tpusecase.TaxpayerProfile
	err     error
}

func (f fakeProfiles) Profile(context.Context, string, string) (tpusecase.TaxpayerProfile, error) {
	return f.profile, f.err
}

func period(t *testing.T, raw string) tpdomain.Period {
	t.Helper()
	p, err := tpdomain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

func taxPeriod(t *testing.T, raw string) taxdomain.Period {
	t.Helper()
	p, err := taxdomain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

// The taxpayers profile becomes the estimate's view: person type, archived
// state, each regime's catalog treatment, its options (bool, enum, list)
// and its window (an open end is the zero Period).
func TestEstimateProfiles_mapsTheTaxpayersProfile(t *testing.T) {
	profiles := NewProfiles(fakeProfiles{profile: tpusecase.TaxpayerProfile{
		TaxpayerID: "tp", PersonType: tpdomain.PersonTypeFisica, Archived: true,
		Regimes: []tpusecase.RegimeProfile{
			{
				Code: "612", Treatment: "full",
				Options: map[string]tpdomain.OptionValue{"iva_fully_taxable": tpdomain.BoolOption(true)},
				Window:  tpdomain.Window{Start: period(t, "2026-01")},
			},
			{
				Code: "625", Treatment: "full",
				Options: map[string]tpdomain.OptionValue{
					"activities":    tpdomain.ListOption([]string{"lodging", "goods_services"}),
					"rfc_given":     tpdomain.StringOption("yes"),
					"ignored_empty": {},
				},
				Window: tpdomain.Window{Start: period(t, "2025-02"), End: period(t, "2026-06")},
			},
		},
	}})

	got, err := profiles.Profile(t.Context(), "user", "tp")

	require.NoError(t, err)
	assert.Equal(t, fiscal.PersonFisica, got.PersonType)
	assert.True(t, got.Archived)
	require.Len(t, got.Regimes, 2)
	business, platforms := got.Regimes[0], got.Regimes[1]
	assert.Equal(t, taxusecase.RegimeHold{
		Code: fiscal.RegimeBusiness, Treatment: taxdomain.TreatmentFull, Start: taxPeriod(t, "2026-01"),
		Options: taxdomain.NewOptions(map[taxdomain.OptionKey]taxdomain.OptionValue{
			taxdomain.OptionIVAFullyTaxable: taxdomain.BoolOption(true),
		}),
	}, business)
	assert.Equal(t, taxPeriod(t, "2025-02"), platforms.Start)
	assert.Equal(t, taxPeriod(t, "2026-06"), platforms.End)
	activities, ok := platforms.Options.List("activities")
	assert.True(t, ok)
	assert.Equal(t, []string{"lodging", "goods_services"}, activities)
	rfc, ok := platforms.Options.Text("rfc_given")
	assert.True(t, ok)
	assert.Equal(t, "yes", rfc)
	_, set := platforms.Options.Bool("ignored_empty")
	assert.False(t, set)
}

func TestEstimateProfiles_passesTheClientError(t *testing.T) {
	want := errors.New("404")

	_, err := NewProfiles(fakeProfiles{err: want}).Profile(t.Context(), "user", "tp")

	assert.ErrorIs(t, err, want)
}

// A catalog treatment outside the contract enum is a wiring bug, never an
// estimate on a guessed treatment.
func TestEstimateProfiles_refusesAnUnknownTreatment(t *testing.T) {
	profiles := NewProfiles(fakeProfiles{profile: tpusecase.TaxpayerProfile{
		PersonType: tpdomain.PersonTypeMoral,
		Regimes:    []tpusecase.RegimeProfile{{Code: "601", Treatment: "mostly", Window: tpdomain.Window{Start: period(t, "2026-01")}}},
	}})

	_, err := profiles.Profile(t.Context(), "user", "tp")

	assert.ErrorIs(t, err, errUnknownTreatment)
}
