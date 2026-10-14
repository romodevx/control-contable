package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	historyhttp "github.com/romodevx/control-contable/backend/internal/history/adapter/http"
	histdomain "github.com/romodevx/control-contable/backend/internal/history/domain"
	histusecase "github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

func histMonth(t *testing.T, raw string) histdomain.Month {
	t.Helper()
	m, err := histdomain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

// historyEstimateOf is estimateOf with what History reads besides the
// totals: due date, fingerprints, versions, disclaimers and CPA topics.
func historyEstimateOf(t *testing.T, period, due string, est taxusecase.TaxpayerEstimate) taxusecase.TaxpayerEstimate {
	t.Helper()
	date, err := fiscal.ParseDate(due)
	require.NoError(t, err)
	est.DueDate = date
	est.Period = taxPeriod(t, period)
	est.RecordsFingerprint = strings.Repeat("a", 64)
	est.FiguresFingerprint = strings.Repeat("b", 64)
	est.EngineVersion = domain.EngineVersion
	est.ParameterSet = taxparams.SetID{FiscalYear: 2026, Version: 1}
	est.PersonType = fiscal.PersonFisica
	est.Disclaimers = []domain.Disclaimer{{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo}}
	return est
}

func TestHistoryEstimates_mapsEachMonth(t *testing.T) {
	inFavor := historyEstimateOf(t, "2026-02", "2026-03-17",
		estimateOf(t, "2026-02", -500_00, 150_96, 125_00, domain.StatusComplete, domain.StatusInformational))
	inFavor.Regimes[0].CPATopics = []domain.CPATopic{domain.TopicCrossRegimeIVANetting}
	inFavor.Records.HasIncomes = true
	none := historyEstimateOf(t, "2026-03", "2026-04-17", estimateOf(t, "2026-03", 0, 0, 0))
	none.Records = taxusecase.RecordTotals{HasExpenses: true}
	source := &fakeRangeEstimator{out: []taxusecase.TaxpayerEstimate{inFavor, none}}

	got, err := historyEstimates{estimator: source}.EstimateMonths(t.Context(), "user", dashTaxpayer,
		histMonth(t, "2026-02"), histMonth(t, "2026-03"))

	require.NoError(t, err)
	assert.Equal(t, []string{"user|" + dashTaxpayer + "|2026-02..2026-03"}, source.got, "one range estimate")
	assert.Equal(t, domain.Legend, got.Legend)
	require.Len(t, got.Months, 2)
	feb := got.Months[0]
	assert.Equal(t, histMonth(t, "2026-02"), feb.Month)
	assert.Equal(t, []histdomain.Status{histdomain.StatusComplete, histdomain.StatusInformational}, feb.Regimes)
	assert.Equal(t, histdomain.Tax{
		ISRPayable: 150_96, IVAPayable: 0, IVAOutcome: "in_favor", IVAInFavor: 500_00, WithholdingsToRemit: 125_00,
		TotalPayable: 150_96 + 125_00,
	}, feb.Tax)
	assert.Equal(t, "2026-03-17", feb.DueDate)
	assert.True(t, feb.CPAReviewPending, "any regime pending CPA review")
	assert.Equal(t, histdomain.Seen{
		RecordsFingerprint: strings.Repeat("a", 64), FiguresFingerprint: strings.Repeat("b", 64),
		EngineVersion: domain.EngineVersion, ParameterSet: "2026/v1",
	}, feb.Seen)
	assert.Equal(t, histdomain.ParameterSet{FiscalYear: 2026, Version: 1}, feb.ParameterSet)
	require.Len(t, feb.Disclaimers, 1)
	assert.Equal(t, "not_tax_advice", feb.Disclaimers[0].Code)
	assert.Equal(t, histdomain.SeverityInfo, feb.Disclaimers[0].Severity)
	assert.NotEmpty(t, feb.Disclaimers[0].Text, "the text getTaxpayerEstimate renders")
	assert.True(t, feb.HasIncomes)
	assert.False(t, feb.HasExpenses)
	mar := got.Months[1]
	assert.Empty(t, mar.Regimes)
	assert.Equal(t, "zero", mar.Tax.IVAOutcome)
	assert.False(t, mar.CPAReviewPending)
	assert.True(t, mar.HasExpenses)
}

func TestHistoryEstimates_passesErrorsOn(t *testing.T) {
	boom := errors.New("boom")
	_, err := historyEstimates{estimator: &fakeRangeEstimator{err: boom}}.EstimateMonths(t.Context(), "u", dashTaxpayer,
		histMonth(t, "2026-01"), histMonth(t, "2026-02"))
	require.ErrorIs(t, err, boom)

	_, err = historyEstimates{estimator: &fakeRangeEstimator{}}.EstimateMonths(t.Context(), "u", dashTaxpayer,
		histMonth(t, "2026-01"), histMonth(t, "2026-02"))
	require.ErrorIs(t, err, histusecase.ErrEstimateCoverage, "a range that does not cover the months is a bug")
}

// MODULE-history §3: both routes are hidden when module_history is OFF or
// implied OFF (module_tax_engine, or module_taxpayers through it); with all
// three ON they are routed past the gate (here to authentication).
func TestHistoryRoutes_needTheEngineAndTaxpayersFlags(t *testing.T) {
	const base = "/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e/history"
	for _, tc := range []struct {
		name    string
		enabled []string
		visible bool
	}{
		{"all ON", []string{"module_taxpayers", "module_tax_engine", "module_history"}, true},
		{"history OFF", []string{"module_taxpayers", "module_tax_engine"}, false},
		{"engine OFF", []string{"module_taxpayers", "module_history"}, false},
		{"taxpayers OFF", []string{"module_tax_engine", "module_history"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newFlaggedHandler(t, testConfig(), pingOK{}, flags.NewService(keysReader{keys: tc.enabled}))
			for _, req := range []*http.Request{
				newRequest(http.MethodGet, base+"?fiscal_year=2026", ""),
				newRequest(http.MethodPost, base+"/2026-03/viewed", `{}`),
			} {
				rec := serveRecorded(h, req)
				if !tc.visible {
					flagtest.AssertHidden(t, rec)
				} else {
					assert.Equal(t, http.StatusUnauthorized, rec.Code, "routed past the gate")
				}
				conforms(t, req, rec)
			}
		})
	}
}

func TestNewHandler_refusesToServeWithoutTheHistoryHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), incomes: stubIncomesModule(), expenses: stubExpensesModule(),
		taxEngine: stubTaxEngineModule(), dashboard: stubDashboardModule(), now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no history handler")
}

func TestNewHistoryModule_needsTheEngine(t *testing.T) {
	_, err := newHistoryModule(historyWiring{})
	require.ErrorIs(t, err, histusecase.ErrMissingDependency)
}

// newRequest is a request from the documentation-range client; a body is
// JSON.
func newRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.RemoteAddr = "203.0.113.10:1234"
	return req
}

func serveRecorded(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

var errStubHistory = errors.New("stub: history not wired")

type stubHistory struct{}

func (stubHistory) Get(context.Context, string, string, int) (histusecase.Year, error) {
	return histusecase.Year{}, errStubHistory
}

func (stubHistory) MarkViewed(context.Context, string, string, histusecase.ViewedRequest) error {
	return errStubHistory
}

// stubHistoryModule has a stub use case and no budget; contract tests wire
// the real module over the real tax engine.
func stubHistoryModule() historyModule {
	return historyModule{handler: historyhttp.NewHandler(stubHistory{})}
}
