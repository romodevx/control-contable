package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// platformMulti is a 625 física with two activities and no RFC given to the
// platform.
const platformMulti = "0199d4a2-1111-7000-8000-0000000000a6"

func (h *harness) withPlatformMulti(t *testing.T) *harness {
	h.taxpayers.profiles[platformMulti] = usecase.TaxpayerProfile{
		PersonType: fiscal.PersonFisica,
		Regimes: []usecase.RegimeProfile{{
			Code:               "625",
			PlatformActivities: []domain.PlatformActivity{domain.ActivityTransportDelivery, domain.ActivityGoodsServices},
		}},
	}
	h.taxpayers.windows[platformMulti+"|625"] = [2]string{"2022-01", ""}
	return h
}

func allRates() []fiscal.IVARate {
	return []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}
}

// MODULE-incomes OQ-7 (getIncomeDefaults): the form's IVA rates for the
// taxpayer, regime and day, from the rate catalog; the app never hard-codes
// a rate. Withholdings are derived (ADR-0008): since 0.12.0 (WH-11) there
// is no counterparty to ask about and no withholding rate to suggest.
func TestDefaults(t *testing.T) {
	tests := []struct {
		name string
		q    usecase.DefaultsQuery
		want usecase.Defaults
	}{
		{
			name: "612: every rate of the day, 16 % suggested",
			q:    usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2026-09-12"},
			want: usecase.Defaults{IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16},
		},
		{
			name: "625 with one activity: it is the default",
			q:    usecase.DefaultsQuery{TaxpayerID: platformA, Regime: "625", On: "2026-09-12"},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16, PlatformActivity: domain.ActivityLodging,
			},
		},
		{
			name: "625 with several activities and none chosen: no default activity",
			q:    usecase.DefaultsQuery{TaxpayerID: platformMulti, Regime: "625", On: "2026-09-12"},
			want: usecase.Defaults{IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16},
		},
		{
			name: "625 with the activity chosen",
			q: usecase.DefaultsQuery{
				TaxpayerID: platformMulti, Regime: "625", On: "2026-09-12", PlatformActivity: domain.ActivityGoodsServices,
			},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16, PlatformActivity: domain.ActivityGoodsServices,
			},
		},
		{
			name: "a persona moral",
			q:    usecase.DefaultsQuery{TaxpayerID: pmA, Regime: "601", On: "2026-09-12"},
			want: usecase.Defaults{IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16},
		},
		{
			name: "a year without the border decree: no 8 %, provisional parameters",
			q:    usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612"},
			want: usecase.Defaults{
				IVARatesAllowed:  []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate0, fiscal.IVAExempt},
				SuggestedIVARate: fiscal.IVARate16,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t).withPlatformMulti(t)
			if tt.q.On == "" {
				h = h.at(today.AddDate(0, 4, 0)) // "on" defaults to today: February 2027
			}

			got, err := h.uc.Defaults(t.Context(), userA, tt.q)

			require.NoError(t, err)
			got.LastUsed = usecase.LastFacts{} // TestDefaults_LastUsed
			assert.Equal(t, tt.want, got)
			assert.Equal(t, []string{"ro"}, h.tx.modes, "a read: allowed read-only")
			assert.Zero(t, h.access.ensures)
		})
	}
}

// An archived taxpayer still prefills (reads are allowed), and the regime
// window is read from the profile, never through CheckRegime (no lock in a
// read-only unit of work).
func TestDefaults_ArchivedTaxpayerAndNoRegimeLock(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Defaults(t.Context(), userA, usecase.DefaultsQuery{TaxpayerID: archivedA, Regime: "612"})
	require.NoError(t, err)
	assert.Empty(t, h.taxpayers.calls)
}

func TestDefaults_Refusals(t *testing.T) {
	tests := []struct {
		name         string
		q            usecase.DefaultsQuery
		code         apperr.Code
		field, cause string
	}{
		{"no regime", usecase.DefaultsQuery{TaxpayerID: taxpayerA},
			apperr.CodeValidationFailed, "regime", "required"},
		{"a malformed date", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2026-13-01"},
			apperr.CodeValidationFailed, "on", "invalid"},
		{"a future date", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2026-10-14"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"before 2022", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2021-12-31"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"before the first parameter set (pre-2026)", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2025-06-01"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"an activity outside 625", usecase.DefaultsQuery{
			TaxpayerID: taxpayerA, Regime: "612", PlatformActivity: domain.ActivityLodging,
		}, apperr.CodeValidationFailed, "platform_activity", "not_for_regime"},
		{"a regime not held on the day", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "606", On: "2026-08-31"},
			apperr.CodeBusinessRuleViolation, "regime", "regime_not_active"},
		{"a regime never held", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "626"},
			apperr.CodeBusinessRuleViolation, "regime", "regime_not_active"},
		{"a 625 activity the taxpayer does not hold", usecase.DefaultsQuery{
			TaxpayerID: platformA, Regime: "625", PlatformActivity: domain.ActivityGoodsServices,
		}, apperr.CodeBusinessRuleViolation, "platform_activity", "activity_not_held"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			_, err := h.uc.Defaults(t.Context(), userA, tt.q)
			requireApp(t, err, tt.code, tt.field, tt.cause)
		})
	}
}

func TestDefaults_NotFound(t *testing.T) {
	h := newHarness(t)
	for _, id := range []string{"not-a-uuid", "0199d4a2-1111-7000-8000-0000000000ff"} {
		_, err := h.uc.Defaults(t.Context(), userA, usecase.DefaultsQuery{TaxpayerID: id, Regime: "612"})
		requireApp(t, err, apperr.CodeNotFound)
	}
}

// IncomeLastUsed (MODULE-withholdings §2.1): the first time no counterparty
// is preselected (OQ-W2) and the operation is the regime's; afterwards the
// facts of the newest income in the regime.
func TestDefaults_LastUsed(t *testing.T) {
	q := usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", On: "2026-09-12"}
	h := newHarness(t)

	got, err := h.uc.Defaults(t.Context(), userA, q)
	require.NoError(t, err)
	assert.Equal(t, usecase.LastFacts{Operation: fiscal.OperationProfessionalServices}, got.LastUsed)

	in := us1(t)
	in.Counterparty, in.Operation = fiscal.ReceptorGovFederal, fiscal.OperationCommission
	_, err = h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	got, err = h.uc.Defaults(t.Context(), userA, q)
	require.NoError(t, err)
	assert.Equal(t, usecase.LastFacts{Counterparty: fiscal.ReceptorGovFederal, Operation: fiscal.OperationCommission}, got.LastUsed)

	platform := q
	platform.TaxpayerID, platform.Regime = platformA, "625"
	got, err = h.uc.Defaults(t.Context(), userA, platform)
	require.NoError(t, err)
	assert.Equal(t, usecase.LastFacts{Operation: fiscal.OperationOther, ViaPlatform: true}, got.LastUsed, "625 starts via a platform")

	h.store.lastFactsErr = errors.New("connection reset")
	_, err = h.uc.Defaults(t.Context(), userA, q)
	require.Error(t, err)
}
