package profiles_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/cmd/internal/profiles"
	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incdomain "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

func period(t *testing.T, raw string) tpdomain.Period {
	t.Helper()
	p, err := tpdomain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

type stubReader struct {
	profile taxpayers.TaxpayerProfile
	err     error
}

func (s stubReader) Profile(context.Context, string, string) (taxpayers.TaxpayerProfile, error) {
	return s.profile, s.err
}

func profile(t *testing.T) taxpayers.TaxpayerProfile {
	return taxpayers.TaxpayerProfile{
		TaxpayerID: "tp", PersonType: tpdomain.PersonTypeFisica, Archived: true,
		Regimes: []taxpayers.RegimeProfile{
			{Code: "606", Options: map[string]tpdomain.OptionValue{"deduction_mode": tpdomain.StringOption("blind_35")},
				Window: tpdomain.Window{Start: period(t, "2025-02"), End: period(t, "2026-06")}},
			{Code: "625", Window: tpdomain.Window{Start: period(t, "2026-03")}, Options: map[string]tpdomain.OptionValue{
				"activities":            tpdomain.ListOption([]string{"lodging"}),
				"rfc_given_to_platform": tpdomain.BoolOption(true),
				"definitive_option":     tpdomain.BoolOption(true),
			}},
		},
	}
}

func TestIncomesReader(t *testing.T) {
	got, err := profiles.IncomesReader{Reader: stubReader{profile: profile(t)}}.Profile(t.Context(), "u", "tp")

	require.NoError(t, err)
	assert.Equal(t, incomes.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Archived: true,
		Regimes: []incomes.RegimeProfile{
			{Code: "606", StartPeriod: "2025-02", EndPeriod: "2026-06"},
			{Code: "625", StartPeriod: "2026-03", PlatformActivities: []incdomain.PlatformActivity{incdomain.ActivityLodging}, RFCGivenToPlatform: true},
		},
	}, got)
}

func TestExpensesReader(t *testing.T) {
	got, err := profiles.ExpensesReader{Reader: stubReader{profile: profile(t)}}.Profile(t.Context(), "u", "tp")

	require.NoError(t, err)
	assert.Equal(t, expenses.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Archived: true,
		Regimes: []expenses.RegimeHold{
			{Code: "606", StartPeriod: "2025-02", EndPeriod: "2026-06", Options: expdomain.Options{DeductionMode: expdomain.DeductionBlind35}},
			{Code: "625", StartPeriod: "2026-03", Options: expdomain.Options{DefinitiveOption: true}},
		},
	}, got)
}

func TestReadersPassErrorsThrough(t *testing.T) {
	refusal := errors.New("404")
	_, err := profiles.IncomesReader{Reader: stubReader{err: refusal}}.Profile(t.Context(), "u", "tp")
	require.ErrorIs(t, err, refusal)
	_, err = profiles.ExpensesReader{Reader: stubReader{err: refusal}}.Profile(t.Context(), "u", "tp")
	require.ErrorIs(t, err, refusal)
}
