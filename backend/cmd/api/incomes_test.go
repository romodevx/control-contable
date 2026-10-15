package main

import (
	"context"
	"errors"
	"io"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	incomeshttp "github.com/romodevx/control-contable/backend/internal/incomes/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxpayersdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// stubQueries is the taxpayers module's read side.
type stubQueries struct {
	profile taxpayers.TaxpayerProfile
	err     error
	checked []string
}

func (s *stubQueries) CheckRegime(_ context.Context, _, taxpayerID, regime, period string) error {
	s.checked = append(s.checked, taxpayerID+"|"+regime+"|"+period)
	return s.err
}

func (s *stubQueries) Profile(context.Context, string, string) (taxpayers.TaxpayerProfile, error) {
	return s.profile, s.err
}

// The adapter hands incomes the payee facts the withholding rules read: the
// person type, the archive, and the 625 options.
func TestIncomesTaxpayers_Profile(t *testing.T) {
	q := &stubQueries{profile: taxpayers.TaxpayerProfile{
		TaxpayerID: "t", PersonType: taxpayersdomain.PersonTypeFisica, Archived: true,
		Regimes: []taxpayers.RegimeProfile{
			{Code: "612", Options: map[string]taxpayersdomain.OptionValue{"iva_border_region": taxpayersdomain.BoolOption(false)},
				Window: taxpayersdomain.Window{Start: period(t, "2025-02"), End: period(t, "2026-06")}},
			{Code: "625", Window: taxpayersdomain.Window{Start: period(t, "2026-03")}, Options: map[string]taxpayersdomain.OptionValue{
				"activities":            taxpayersdomain.ListOption([]string{"lodging", "goods_services"}),
				"rfc_given_to_platform": taxpayersdomain.BoolOption(true),
			}},
		},
	}}
	got, err := incomesTaxpayers{queries: q}.Profile(t.Context(), "u", "t")
	require.NoError(t, err)
	assert.Equal(t, incomes.TaxpayerProfile{
		PersonType: fiscal.PersonFisica, Archived: true,
		Regimes: []incomes.RegimeProfile{
			{Code: "612", StartPeriod: "2025-02", EndPeriod: "2026-06"},
			{Code: "625", StartPeriod: "2026-03", PlatformActivities: []domain.PlatformActivity{domain.ActivityLodging, domain.ActivityGoodsServices}, RFCGivenToPlatform: true},
		},
	}, got)
}

func TestIncomesTaxpayers_PassesErrorsThrough(t *testing.T) {
	refusal := errors.New("404")
	q := &stubQueries{err: refusal}
	adapter := incomesTaxpayers{queries: q}
	_, err := adapter.Profile(t.Context(), "u", "t")
	require.ErrorIs(t, err, refusal)
	require.ErrorIs(t, adapter.CheckRegime(t.Context(), "u", "t", "612", "2026-09"), refusal)
	assert.Equal(t, []string{"t|612|2026-09"}, q.checked)
}

// The composition root binds the rate catalog once, over the embedded
// parameters (the boot fails on a set without the IVA rates).
func TestNewRateCatalog(t *testing.T) {
	rates, err := newRateCatalog()
	require.NoError(t, err)
	_, ok := rates.Rate(fiscal.KeyIVAGeneral, mustDate(t, "2026-09-10"))
	assert.True(t, ok)
	var _ incomes.RateCatalog = rates
}

func mustDate(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// stubIncomeUseCases must not be reached by the platform tests; the
// contract tests wire the real ones.
type stubIncomeUseCases struct{}

var errStubIncomes = errors.New("stub: incomes not wired")

func (stubIncomeUseCases) Create(context.Context, string, incomes.CreateInput) (incomes.WriteResult, error) {
	return incomes.WriteResult{}, errStubIncomes
}

func (stubIncomeUseCases) Update(context.Context, string, string, string, incomes.Patch) (incomes.WriteResult, error) {
	return incomes.WriteResult{}, errStubIncomes
}

func (stubIncomeUseCases) Get(context.Context, string, string, string) (incomes.Income, error) {
	return incomes.Income{}, errStubIncomes
}

func (stubIncomeUseCases) List(context.Context, string, incomes.ListQuery) (incomes.Page, error) {
	return incomes.Page{}, errStubIncomes
}

func (stubIncomeUseCases) Delete(context.Context, string, string, string) error {
	return errStubIncomes
}

func (stubIncomeUseCases) Restore(context.Context, string, string, string) (incomes.WriteResult, error) {
	return incomes.WriteResult{}, errStubIncomes
}

func (stubIncomeUseCases) Defaults(context.Context, string, incomes.DefaultsQuery) (incomes.Defaults, error) {
	return incomes.Defaults{}, errStubIncomes
}

func (stubIncomeUseCases) Preview(context.Context, string, incomes.PreviewQuery) (incomes.Preview, error) {
	return incomes.Preview{}, errStubIncomes
}

// stubIncomesModule has stub use cases and no write budget.
func stubIncomesModule() incomesModule {
	return incomesModule{handler: incomeshttp.NewHandler(stubIncomeUseCases{})}
}

func TestNewHandler_refusesToServeWithoutTheIncomesHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t), now: time.Now,
		taxpayers: stubTaxpayersModule(), expenses: stubExpensesModule(), taxEngine: stubTaxEngineModule(),
		logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no incomes handler")
}

// The incomes module needs the shared unit of work, the taxpayers queries
// and the rate catalog; it hands the router its write and preview budgets.
func TestNewIncomesModule(t *testing.T) {
	rates, err := newRateCatalog()
	require.NoError(t, err)
	_, err = newIncomesModule(incomesWiring{taxpayers: &stubQueries{}, rates: rates})
	require.ErrorIs(t, err, errIncomesWiring)
	_, err = newIncomesModule(incomesWiring{tx: stubTx{}, rates: rates})
	require.ErrorIs(t, err, errIncomesWiring)
	_, err = newIncomesModule(incomesWiring{tx: stubTx{}, taxpayers: &stubQueries{}})
	require.ErrorIs(t, err, errIncomesWiring)

	m, err := newIncomesModule(incomesWiring{tx: stubTx{}, taxpayers: &stubQueries{}, rates: rates})
	require.NoError(t, err)
	assert.NotNil(t, m.handler)
	assert.NotNil(t, m.incomes)
	assert.ElementsMatch(t, []string{"createIncome", "updateIncome", "deleteIncome", "restoreIncome", "previewIncomeWithholdings"},
		slices.Collect(maps.Keys(m.userLimits)))
}

// stubTx is a unit of work that is never entered.
type stubTx struct{}

func (stubTx) ReadWrite(context.Context, string, tx.Work) error { return errStubIncomes }
func (stubTx) ReadOnly(context.Context, string, tx.Work) error  { return errStubIncomes }
