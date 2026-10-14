package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dashboardhttp "github.com/romodevx/control-contable/backend/internal/dashboard/adapter/http"
	dashdomain "github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	dashusecase "github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const dashTaxpayer = "0199d4a2-1111-7000-8000-0000000da5a1"

func dashMonth(t *testing.T, raw string) dashdomain.Month {
	t.Helper()
	m, err := dashdomain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

// fakeRangeEstimator records the range and answers estimates.
type fakeRangeEstimator struct {
	got []string
	out []taxusecase.TaxpayerEstimate
	err error
}

func (f *fakeRangeEstimator) EstimateRange(_ context.Context, userID, taxpayerID string, from, to domain.Period) ([]taxusecase.TaxpayerEstimate, error) {
	f.got = append(f.got, userID+"|"+taxpayerID+"|"+from.String()+".."+to.String())
	return f.out, f.err
}

func estimateOf(t *testing.T, period string, iva, isr money.Cents, toRemit money.Cents, statuses ...domain.Status) taxusecase.TaxpayerEstimate {
	t.Helper()
	ivaResult := domain.NewTaxResult(money.Exact(iva) * money.ExactPerCent)
	isrResult := domain.NewTaxResult(money.Exact(isr) * money.ExactPerCent)
	est := taxusecase.TaxpayerEstimate{
		TaxpayerID: dashTaxpayer, Period: taxPeriod(t, period), Legend: domain.Legend,
		Totals: domain.Totals{IVA: ivaResult, ISR: isrResult, WithholdingsToRemit: toRemit,
			TotalPayable: ivaResult.Payable() + isrResult.Payable() + toRemit},
		Records: taxusecase.RecordTotals{IncomeCollected: 10_00, ExpensesPaid: 4_00},
	}
	for _, s := range statuses {
		est.Regimes = append(est.Regimes, taxusecase.RegimeResult{Status: s})
	}
	return est
}

func TestDashboardEstimates_mapsEachMonth(t *testing.T) {
	last := estimateOf(t, "2027-01", 2_133_33, 150_96, 0, domain.StatusComplete, domain.StatusIncomplete)
	source := &fakeRangeEstimator{out: []taxusecase.TaxpayerEstimate{
		estimateOf(t, "2026-12", -500_00, 0, 125_00, domain.StatusInformational), last,
	}}

	got, err := dashboardEstimates{estimator: source}.EstimateWindow(t.Context(), "user", dashTaxpayer,
		dashMonth(t, "2026-12"), dashMonth(t, "2027-01"))

	require.NoError(t, err)
	assert.Equal(t, []string{"user|" + dashTaxpayer + "|2026-12..2027-01"}, source.got)
	assert.Equal(t, []dashusecase.MonthTax{
		{
			Month: dashMonth(t, "2026-12"), WithholdingsToRemit: 125_00, TotalPayable: 125_00,
			IVAOutcome: dashusecase.OutcomeInFavor, Regimes: []dashdomain.Status{dashdomain.StatusInformational},
			IncomeCollected: 10_00, ExpensesPaid: 4_00,
		},
		{
			Month: dashMonth(t, "2027-01"), IVAPayable: 2_133_33, ISRPayable: 150_96, TotalPayable: 2_284_29,
			IVAOutcome: dashusecase.OutcomePayable, Regimes: []dashdomain.Status{dashdomain.StatusComplete, dashdomain.StatusIncomplete},
			IncomeCollected: 10_00, ExpensesPaid: 4_00,
		},
	}, got.Months)
	assert.Equal(t, taxenginehttp.PresentEstimate(last), got.Current, "the hero is the last month, presented as getTaxpayerEstimate")
	assert.Equal(t, domain.Legend, got.Legend)
}

func TestDashboardEstimates_passesErrorsOn(t *testing.T) {
	boom := errors.New("boom")
	_, err := dashboardEstimates{estimator: &fakeRangeEstimator{err: boom}}.EstimateWindow(t.Context(), "u", dashTaxpayer,
		dashMonth(t, "2026-01"), dashMonth(t, "2026-02"))
	require.ErrorIs(t, err, boom)

	_, err = dashboardEstimates{estimator: &fakeRangeEstimator{}}.EstimateWindow(t.Context(), "u", dashTaxpayer,
		dashMonth(t, "2026-01"), dashMonth(t, "2026-02"))
	require.ErrorIs(t, err, errEstimateWindow, "a range that does not cover the window is a bug, never a short trend")
}

// fakeIncomeRecords and fakeExpenseRecords record the range queries.
type fakeIncomeRecords struct {
	got  []incomesinfra.LedgerQuery
	rows []incomesinfra.IncomeMonthSums
	has  bool
	err  error
}

func (f *fakeIncomeRecords) IncomeSums(_ context.Context, q incomesinfra.LedgerQuery) ([]incomesinfra.IncomeMonthSums, error) {
	f.got = append(f.got, q)
	return f.rows, f.err
}

func (f *fakeIncomeRecords) HasLive(context.Context, string) (bool, error) { return f.has, f.err }

type fakeExpenseRecords struct {
	got  []expensesinfra.LedgerQuery
	rows []expensesinfra.ExpenseMonthSums
	has  bool
	err  error
}

func (f *fakeExpenseRecords) ExpenseSums(_ context.Context, q expensesinfra.LedgerQuery) ([]expensesinfra.ExpenseMonthSums, error) {
	f.got = append(f.got, q)
	return f.rows, f.err
}

func (f *fakeExpenseRecords) HasLive(context.Context, string) (bool, error) { return f.has, f.err }

// The records the window shows: per month, over every regime; months of
// the read's first January before the window are left out.
func TestDashboardRecords_sumsTheWindowByMonth(t *testing.T) {
	incomes := &fakeIncomeRecords{rows: []incomesinfra.IncomeMonthSums{
		{Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.March, Cash: 99_00},
		{Regime: fiscal.RegimeBusiness, Year: 2026, Month: time.November, Cash: 100_00, Accrual: 7_00},
		{Regime: fiscal.RegimeRent, Year: 2026, Month: time.November, Cash: 50_00},
		{Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.February, Cash: 20_00},
	}}
	expenses := &fakeExpenseRecords{rows: []expensesinfra.ExpenseMonthSums{
		{Regime: fiscal.RegimeBusiness, Year: 2027, Month: time.January, DeductiblePaid: 3_00, NonDeductiblePaid: 2_00, CashOverLimitPaid: 1_00},
	}}
	records := dashboardRecords{incomes: incomes, expenses: expenses}
	from, to := dashMonth(t, "2026-04"), dashMonth(t, "2027-03")

	gotIncomes, err := records.IncomeCollected(t.Context(), dashTaxpayer, from, to)
	require.NoError(t, err)
	gotExpenses, err := records.ExpensesPaid(t.Context(), dashTaxpayer, from, to)
	require.NoError(t, err)

	want := incomesinfra.LedgerQuery{TaxpayerID: dashTaxpayer, FirstYear: 2026, FiscalYear: 2027, Through: time.March}
	assert.Equal(t, []incomesinfra.LedgerQuery{want}, incomes.got)
	assert.Equal(t, []expensesinfra.LedgerQuery{{TaxpayerID: dashTaxpayer, FirstYear: 2026, FiscalYear: 2027, Through: time.March}}, expenses.got)
	assert.Equal(t, map[dashdomain.Month]money.Cents{
		dashMonth(t, "2026-11"): 150_00, dashMonth(t, "2027-02"): 20_00,
	}, gotIncomes, "collected only (accrual is not cash); March 2026 is before the window")
	assert.Equal(t, map[dashdomain.Month]money.Cents{dashMonth(t, "2027-01"): 6_00}, gotExpenses, "every deductibility class")
}

func TestDashboardRecords_existenceAndErrors(t *testing.T) {
	records := dashboardRecords{incomes: &fakeIncomeRecords{has: true}, expenses: &fakeExpenseRecords{}}
	has, err := records.HasIncomes(t.Context(), dashTaxpayer)
	require.NoError(t, err)
	assert.True(t, has)
	has, err = records.HasExpenses(t.Context(), dashTaxpayer)
	require.NoError(t, err)
	assert.False(t, has)

	boom := errors.New("boom")
	broken := dashboardRecords{incomes: &fakeIncomeRecords{err: boom}, expenses: &fakeExpenseRecords{err: boom}}
	_, err = broken.IncomeCollected(t.Context(), dashTaxpayer, dashMonth(t, "2026-01"), dashMonth(t, "2026-02"))
	require.ErrorIs(t, err, boom)
	_, err = broken.ExpensesPaid(t.Context(), dashTaxpayer, dashMonth(t, "2026-01"), dashMonth(t, "2026-02"))
	require.ErrorIs(t, err, boom)
}

var errStubDashboard = errors.New("stub: dashboard not wired")

type stubDashboards struct{}

func (stubDashboards) Get(context.Context, string, dashusecase.Request) (dashusecase.Dashboard[taxenginehttp.EstimateBody], error) {
	return dashusecase.Dashboard[taxenginehttp.EstimateBody]{}, errStubDashboard
}

// stubDashboardModule has a stub use case and no budget; contract tests
// wire the real module over the real tax engine.
func stubDashboardModule() dashboardModule {
	return dashboardModule{handler: dashboardhttp.NewHandler[taxenginehttp.EstimateBody](stubDashboards{})}
}

func TestNewHandler_refusesToServeWithoutTheDashboardHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), incomes: stubIncomesModule(), expenses: stubExpensesModule(),
		taxEngine: stubTaxEngineModule(), now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no dashboard handler")
}

func TestNewDashboardModule_needsTheEngineAndTheTaxpayers(t *testing.T) {
	_, err := newDashboardModule(dashboardWiring{})
	require.ErrorIs(t, err, dashusecase.ErrMissingDependency)
}

// MODULE-dashboard §3: the route is hidden when module_dashboard is OFF or
// its prerequisite module_taxpayers is OFF (implied OFF), whatever the
// engine; with both ON it is routed past the gate (here to authentication).
// module_tax_engine is not a prerequisite: it degrades the response.
func TestDashboardRoute_needsTheTaxpayersFlagToo(t *testing.T) {
	const target = "/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e/dashboard"
	for _, tc := range []struct {
		name    string
		enabled []string
		visible bool
	}{
		{"both ON", []string{"module_taxpayers", "module_dashboard"}, true},
		{"both ON, engine OFF", []string{"module_taxpayers", "module_dashboard"}, true},
		{"dashboard OFF", []string{"module_taxpayers", "module_tax_engine"}, false},
		{"taxpayers OFF", []string{"module_dashboard", "module_tax_engine"}, false},
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

// fakeDashProfiles is the taxpayers module's profile reader.
type fakeDashProfiles struct{ err error }

func (f fakeDashProfiles) Profile(context.Context, string, string) (tpusecase.TaxpayerProfile, error) {
	return tpusecase.TaxpayerProfile{}, f.err
}

func TestDashboardTaxpayers_checksThroughTheProfile(t *testing.T) {
	require.NoError(t, dashboardTaxpayers{reader: fakeDashProfiles{}}.CheckTaxpayer(t.Context(), "u", dashTaxpayer))
	boom := errors.New("not found")
	require.ErrorIs(t, dashboardTaxpayers{reader: fakeDashProfiles{err: boom}}.CheckTaxpayer(t.Context(), "u", dashTaxpayer), boom)
}
