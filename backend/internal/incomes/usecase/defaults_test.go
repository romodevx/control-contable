package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
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
	h.catalog.rates[domain.KeyISRPlatformNoRFC] = rate(t, "0.20")
	return h
}

func allRates() []fiscal.IVARate {
	return []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}
}

func known(r money.Rate) usecase.SuggestedRate { return usecase.SuggestedRate{Rate: r, Known: true} }

// MODULE-incomes OQ-7 (getIncomeDefaults): the form's IVA rates and the
// withholding suggestions for the taxpayer, regime, counterparty and day,
// from the rate catalog; the app never hard-codes a rate.
func TestDefaults(t *testing.T) {
	tests := []struct {
		name string
		q    usecase.DefaultsQuery
		want usecase.Defaults
	}{
		{
			name: "612 paid by a persona moral: 10 % ISR and 2/3 of the IVA",
			q:    usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyMoral, On: "2026-09-12"},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16,
				ISR: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "0.10"))},
				IVA: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "0.666667"))},
			},
		},
		{
			name: "sale to the general public: nothing withheld",
			q:    usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, On: "2026-09-12"},
			want: usecase.Defaults{IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16},
		},
		{
			name: "625 with one activity: it is the default, RFC given (half the IVA)",
			q:    usecase.DefaultsQuery{TaxpayerID: platformA, Regime: "625", Counterparty: domain.CounterpartyPlatform, On: "2026-09-12"},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16, PlatformActivity: domain.ActivityLodging,
				ISR: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "0.04"))},
				IVA: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "0.5"))},
			},
		},
		{
			name: "625 with several activities and none chosen: allowed, ISR without a suggestion",
			q:    usecase.DefaultsQuery{TaxpayerID: platformMulti, Regime: "625", Counterparty: domain.CounterpartyPlatform, On: "2026-09-12"},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16,
				ISR: usecase.WithholdingDefault{Allowed: true},
				IVA: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "1"))},
			},
		},
		{
			name: "625 with the activity chosen, no RFC given: 20 % and all the IVA",
			q: usecase.DefaultsQuery{
				TaxpayerID: platformMulti, Regime: "625", Counterparty: domain.CounterpartyPlatform, On: "2026-09-12",
				PlatformActivity: domain.ActivityGoodsServices,
			},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16, PlatformActivity: domain.ActivityGoodsServices,
				ISR: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "0.20"))},
				IVA: usecase.WithholdingDefault{Allowed: true, Suggested: known(rate(t, "1"))},
			},
		},
		{
			name: "a persona moral paid by a persona moral: IVA withholding allowed, no suggestion",
			q:    usecase.DefaultsQuery{TaxpayerID: pmA, Regime: "601", Counterparty: domain.CounterpartyMoral, On: "2026-09-12"},
			want: usecase.Defaults{
				IVARatesAllowed: allRates(), SuggestedIVARate: fiscal.IVARate16,
				IVA: usecase.WithholdingDefault{Allowed: true},
			},
		},
		{
			name: "a year without the border decree: no 8 %, provisional parameters",
			q:    usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic},
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
	_, err := h.uc.Defaults(t.Context(), userA, usecase.DefaultsQuery{TaxpayerID: archivedA, Regime: "612", Counterparty: domain.CounterpartyPublic})
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
		{"no regime", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Counterparty: domain.CounterpartyPublic},
			apperr.CodeValidationFailed, "regime", "required"},
		{"an unknown counterparty", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: "vecino"},
			apperr.CodeValidationFailed, "counterparty_type", "invalid"},
		{"a malformed date", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, On: "2026-13-01"},
			apperr.CodeValidationFailed, "on", "invalid"},
		{"a future date", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, On: "2026-10-14"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"before 2022", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, On: "2021-12-31"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"before the first parameter set (pre-2026)", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, On: "2025-06-01"},
			apperr.CodeValidationFailed, "on", "out_of_range"},
		{"an activity outside 625", usecase.DefaultsQuery{
			TaxpayerID: taxpayerA, Regime: "612", Counterparty: domain.CounterpartyPublic, PlatformActivity: domain.ActivityLodging,
		}, apperr.CodeValidationFailed, "platform_activity", "not_for_regime"},
		{"a regime not held on the day", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "606", Counterparty: domain.CounterpartyPublic, On: "2026-08-31"},
			apperr.CodeBusinessRuleViolation, "regime", "regime_not_active"},
		{"a regime never held", usecase.DefaultsQuery{TaxpayerID: taxpayerA, Regime: "626", Counterparty: domain.CounterpartyPublic},
			apperr.CodeBusinessRuleViolation, "regime", "regime_not_active"},
		{"a 625 activity the taxpayer does not hold", usecase.DefaultsQuery{
			TaxpayerID: platformA, Regime: "625", Counterparty: domain.CounterpartyPlatform, PlatformActivity: domain.ActivityGoodsServices,
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
		_, err := h.uc.Defaults(t.Context(), userA, usecase.DefaultsQuery{TaxpayerID: id, Regime: "612", Counterparty: domain.CounterpartyPublic})
		requireApp(t, err, apperr.CodeNotFound)
	}
}
