package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

var errStubTaxEngine = errors.New("stub: tax engine not wired")

type stubInputs struct{}

func (stubInputs) Get(context.Context, string, usecase.InputsRequest) (usecase.TaxInputsView, error) {
	return usecase.TaxInputsView{}, errStubTaxEngine
}

func (stubInputs) Put(context.Context, string, usecase.InputsRequest, []domain.SubmittedInput) (usecase.TaxInputsView, error) {
	return usecase.TaxInputsView{}, errStubTaxEngine
}

type stubEstimates struct{}

func (stubEstimates) EstimateMonth(context.Context, string, string, string) (usecase.TaxpayerEstimate, error) {
	return usecase.TaxpayerEstimate{}, errStubTaxEngine
}

func (stubEstimates) Breakdown(context.Context, string, string, string, string) (usecase.RegimeBreakdown, error) {
	return usecase.RegimeBreakdown{}, errStubTaxEngine
}

// stubTaxEngineModule has stub use cases and no write budget; contract and
// integration tests wire the real module.
func stubTaxEngineModule() taxEngineModule {
	return taxEngineModule{handler: taxenginehttp.NewHandler(stubInputs{}, stubEstimates{})}
}

func TestNewHandler_refusesToServeWithoutTheTaxEngineHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), expenses: stubExpensesModule(), now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no tax engine handler")
}

// MODULE-tax-engine §7 / §5: every M5 route (tax inputs, estimate,
// breakdown) is hidden when module_tax_engine is OFF or its prerequisite
// module_taxpayers is OFF; with both ON it is routed past the gate (here to
// authentication).
func TestTaxEngineRoutes_needTheTaxpayersFlagToo(t *testing.T) {
	for _, target := range []string{
		"/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e/regimes/612/tax-inputs?fiscal_year=2026",
		"/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e/estimates?period=2026-03",
		"/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e/estimates/2026-03/breakdown?regime=612",
	} {
		t.Run(target, func(t *testing.T) { assertTaxEngineGate(t, target) })
	}
}

func assertTaxEngineGate(t *testing.T, target string) {
	t.Helper()
	for _, tc := range []struct {
		name    string
		enabled []string
		visible bool
	}{
		{"both ON", []string{"module_taxpayers", "module_tax_engine"}, true},
		{"tax engine OFF", []string{"module_taxpayers"}, false},
		{"taxpayers OFF", []string{"module_tax_engine"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newFlaggedHandler(t, testConfig(), pingOK{}, flags.NewService(keysReader{keys: tc.enabled}))

			req, rec := get(h, target)

			if !tc.visible {
				flagtest.AssertHidden(t, rec)
			} else {
				assert.Equal(t, http.StatusUnauthorized, rec.Code, "routed past the gate")
			}
			conforms(t, req, rec)
		})
	}
}

func TestUserLimits_mergeEveryModule(t *testing.T) {
	limits, err := mergeLimits(map[string]int{"a": 1}, map[string]int{"b": 2})

	assert.NoError(t, err)
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, limits)
	_, err = mergeLimits(map[string]int{"a": 1}, map[string]int{"a": 2})
	assert.ErrorContains(t, err, `"a"`, "one operation never gets two budgets")
}
