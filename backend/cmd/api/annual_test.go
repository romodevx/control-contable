package main

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
)

// stubAnnualModule is the annual handler with no use cases: the flag gate
// answers before it in these tests.
func stubAnnualModule() annualModule {
	return annualModule{handler: annualhttp.NewHandler(annualhttp.Deps{})}
}

func TestNewHandler_refusesToServeWithoutTheAnnualHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), incomes: stubIncomesModule(), expenses: stubExpensesModule(),
		taxEngine: stubTaxEngineModule(), dashboard: stubDashboardModule(), history: stubHistoryModule(),
		now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no annual handler")
}

func TestNewAnnualModule_needsTheEngineAndTheUnitOfWork(t *testing.T) {
	_, err := newAnnualModule(annualWiring{})
	require.ErrorIs(t, err, errAnnualWiring)
}

// MODULE-annual §6 (US-A13): every annual route is hidden while
// module_annual is OFF or implied OFF (module_tax_engine or
// module_taxpayers OFF: flags.Prerequisites, through the real flags
// service); with the three ON it is routed past the gate (to
// authentication).
func TestAnnualRoutes_needTheEngineAndTaxpayersFlags(t *testing.T) {
	const base = "/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e"
	for _, tc := range []struct {
		name    string
		enabled []string
		visible bool
	}{
		{"all ON", []string{"module_taxpayers", "module_tax_engine", "module_annual"}, true},
		{"annual OFF", []string{"module_taxpayers", "module_tax_engine"}, false},
		{"engine OFF", []string{"module_taxpayers", "module_annual"}, false},
		{"taxpayers OFF", []string{"module_tax_engine", "module_annual"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newFlaggedHandler(t, testConfig(), pingOK{}, flags.NewService(keysReader{keys: tc.enabled}))
			for _, req := range []*http.Request{
				newRequest(http.MethodGet, base+"/annual-estimates/2026", ""),
				newRequest(http.MethodPut, base+"/regimes/612/filed-payments/2026", `{}`),
			} {
				rec := serveRecorded(h, req)
				if !tc.visible {
					flagtest.AssertHidden(t, rec)
				} else {
					assert.Equal(t, http.StatusUnauthorized, rec.Code, "routed past the gate")
				}
			}
		})
	}
}
