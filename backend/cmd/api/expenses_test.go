package main

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expusecase "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// stubExpenseUseCases must not be reached by the platform tests; the
// contract tests wire the real ones.
type stubExpenseUseCases struct{}

var errStubExpenses = errors.New("stub: expenses not wired")

func (stubExpenseUseCases) Create(context.Context, expusecase.CreateInput) (expusecase.CreateResult, error) {
	return expusecase.CreateResult{}, errStubExpenses
}

func (stubExpenseUseCases) Update(context.Context, expusecase.UpdateInput) (expusecase.Expense, error) {
	return expusecase.Expense{}, errStubExpenses
}

func (stubExpenseUseCases) Get(context.Context, string, string, string) (expusecase.Expense, error) {
	return expusecase.Expense{}, errStubExpenses
}

func (stubExpenseUseCases) List(context.Context, expusecase.ListInput) (expusecase.Page, error) {
	return expusecase.Page{}, errStubExpenses
}

func (stubExpenseUseCases) Delete(context.Context, string, string, string) error {
	return errStubExpenses
}

func (stubExpenseUseCases) Restore(context.Context, string, string, string) (expusecase.Expense, error) {
	return expusecase.Expense{}, errStubExpenses
}

func (stubExpenseUseCases) Categories(context.Context, expusecase.CategoriesInput) (expusecase.CategoryList, error) {
	return expusecase.CategoryList{}, errStubExpenses
}

// stubExpensesModule has stub use cases and no write budget.
func stubExpensesModule() expensesModule {
	return expensesModule{handler: expenseshttp.NewHandler(stubExpenseUseCases{})}
}

func TestNewHandler_refusesToServeWithoutTheExpensesHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t), now: time.Now,
		taxpayers: stubTaxpayersModule(), logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no expenses handler")
}

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

func fiscalDate(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// The taxpayers profile becomes the expense rules' view: person type,
// archived state, each regime window ("YYYY-MM", "" = open) and the two
// options that change a category's hint.
func TestTaxpayerProfiles_mapsTheTaxpayersProfile(t *testing.T) {
	profiles := taxpayerProfiles{reader: fakeProfiles{profile: tpusecase.TaxpayerProfile{
		TaxpayerID: "tp", PersonType: tpdomain.PersonTypeFisica, Archived: true,
		Regimes: []tpusecase.RegimeProfile{
			{Code: "612", Window: tpdomain.Window{Start: period(t, "2026-01")}},
			{Code: "606", Options: map[string]tpdomain.OptionValue{"deduction_mode": tpdomain.StringOption("blind_35")},
				Window: tpdomain.Window{Start: period(t, "2025-02"), End: period(t, "2026-06")}},
			{Code: "625", Options: map[string]tpdomain.OptionValue{"definitive_option": tpdomain.BoolOption(true)},
				Window: tpdomain.Window{Start: period(t, "2026-03")}},
		},
	}}}

	got, err := profiles.Profile(t.Context(), "user", "tp")

	require.NoError(t, err)
	assert.Equal(t, expusecase.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Archived: true,
		Regimes: []expusecase.RegimeHold{
			{Code: "612", StartPeriod: "2026-01"},
			{Code: "606", StartPeriod: "2025-02", EndPeriod: "2026-06", Options: expdomain.Options{DeductionMode: expdomain.DeductionBlind35}},
			{Code: "625", StartPeriod: "2026-03", Options: expdomain.Options{DefinitiveOption: true}},
		},
	}, got)
}

func TestTaxpayerProfiles_passesTheClientError(t *testing.T) {
	want := errors.New("404")
	profiles := taxpayerProfiles{reader: fakeProfiles{err: want}}

	_, err := profiles.Profile(t.Context(), "user", "tp")

	assert.ErrorIs(t, err, want)
}

// The tax engine's RateCatalog over the embedded parameters serves the
// expenses Parameters port: the dated facts of a covered day, the previous
// year's (provisional) for a year without its own set, and
// ErrParametersUnavailable before the first set (pre-2026: 400 out_of_range).
func TestExpenseParameters_overTheEmbeddedSets(t *testing.T) {
	params, err := newExpenseParameters()
	require.NoError(t, err)

	got, err := params.On(fiscalDate(t, "2026-09-12"))
	require.NoError(t, err)
	assert.Equal(t, expusecase.DatedParameters{
		IVARatesAllowed:  []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt},
		CashPaymentLimit: 200000,
	}, got)

	next, err := params.On(fiscalDate(t, "2027-01-10"))
	require.NoError(t, err)
	assert.True(t, next.Provisional, "2027 has no published set: the 2026 values, provisional")

	_, err = params.On(fiscalDate(t, "2025-12-31"))
	require.ErrorIs(t, err, expusecase.ErrParametersUnavailable)
}

// fakeLatest answers a canned latest paid_on.
type fakeLatest struct {
	day        fiscal.Date
	err        error
	taxpayerID string
	regime     fiscal.RegimeCode
}

func (f *fakeLatest) LatestPaidOn(_ context.Context, taxpayerID string, regime fiscal.RegimeCode) (fiscal.Date, error) {
	f.taxpayerID, f.regime = taxpayerID, regime
	return f.day, f.err
}

// M2-14 G-1: the taxpayers module's record lookup reads the expenses (the
// incomes join it in M3): the latest period holding a live expense of the
// regime, the zero Period when none.
func TestRegimeRecords_latestExpensePeriod(t *testing.T) {
	store := &fakeLatest{day: fiscalDate(t, "2026-09-20")}
	records := regimeRecords{expenses: store}

	got, err := records.LatestRecordPeriod(t.Context(), "tp", "612")

	require.NoError(t, err)
	assert.Equal(t, period(t, "2026-09"), got)
	assert.Equal(t, "tp", store.taxpayerID)
	assert.Equal(t, fiscal.RegimeCode("612"), store.regime)

	store.day = fiscal.Date{}
	got, err = records.LatestRecordPeriod(t.Context(), "tp", "612")
	require.NoError(t, err)
	assert.True(t, got.IsZero())

	store.err = errors.New("db down")
	_, err = records.LatestRecordPeriod(t.Context(), "tp", "612")
	assert.ErrorIs(t, err, store.err)
}
