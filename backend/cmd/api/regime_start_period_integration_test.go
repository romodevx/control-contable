//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// M5-29 QA finding D2: a user who onboards in October must be able to give
// the regime an earlier start, then record and estimate the months before.
// Through the real composition root over Postgres: create the taxpayer with
// start_period 2026-01, record a March income, estimate March.

// onboardingDay is a mid-year onboarding: 14 October 2026, Mexico City.
var onboardingDay = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

func TestRegimeStartPeriod_midYearOnboardingRecordsAndEstimatesEarlierMonths(t *testing.T) {
	modules := []string{"module_auth", "module_taxpayers", "module_incomes", "module_tax_engine"}
	env := newContractEnv(t, pingOK{}, withIncomesDay(onboardingDay), withTaxEngineDay(onboardingDay))
	env.apply(t, modules, flagtest.Static{
		"module_auth": true, "module_taxpayers": true, "module_incomes": true, "module_tax_engine": true,
	})
	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken

	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", `{"regime_code":"612","start_period":"2026-01"}`), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	taxpayer := decodeTaxpayer(t, rec)
	require.Len(t, taxpayer.Regimes, 1)
	assert.Equal(t, "2026-01", taxpayer.Regimes[0].StartPeriod, "the start the user chose, not the current month")
	base := "/v1/taxpayers/" + taxpayer.Id.String()

	rec = env.do(env.send(http.MethodPost, base+"/incomes",
		incomeJSON(t, "", map[string]any{"invoiced_on": "2026-03-05", "collected_on": "2026-03-10"}), token))
	require.Equal(t, http.StatusCreated, rec.Code, "a March income inside the regime window: %s", rec.Body.String())

	rec = env.do(env.send(http.MethodGet, estimateTarget(taxpayer.Id.String(), "2026-03"), "", token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var estimate gen.TaxpayerEstimate
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &estimate), rec.Body.String())
	require.Len(t, estimate.Regimes, 1)
	regime := estimate.Regimes[0]
	assert.Equal(t, gen.RegimeCode("612"), regime.RegimeCode)
	require.NotNil(t, regime.IvaComponents)
	assert.EqualValues(t, 160000, regime.IvaComponents.ChargedCollectedCents, "March's income is in March's estimate")
}

func TestRegimeStartPeriod_beforeTheFirstParameterYearIsOutOfRange(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers"}, flagtest.Static{"module_auth": true, "module_taxpayers": true})
	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken

	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", `{"regime_code":"612","start_period":"2025-12"}`), token))

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Equal(t, []string{"regimes.0.start_period/out_of_range"}, details(t, rec))
}
