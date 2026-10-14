package usecase_test

import (
	"context"
	"errors"
	"time"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	userA     = "0199d4a2-0000-7000-8000-00000000d00a"
	taxpayerA = "0199d4a2-1111-7000-8000-00000000d00a"
)

// today is 14 October 2026, noon in Mexico City: the default window has
// ten points (2026-01..2026-10).
var today = time.Date(2026, time.October, 14, 12, 0, 0, 0, clock.Location())

var errPortDown = errors.New("fake: port down")

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

func userOf(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// fakeTx runs work inline and records the modes used.
type fakeTx struct{ modes []string }

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "rw")
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.modes = append(f.modes, "ro")
	return work(context.WithValue(ctx, userKey{}, userID))
}

// fakeFlags answers the modules ON in it.
type fakeFlags map[string]bool

func (f fakeFlags) Enabled(_ context.Context, module string) bool { return f[module] }

// fakeTaxpayers is the ownership check.
type fakeTaxpayers struct {
	err   error
	calls []string
}

func (f *fakeTaxpayers) CheckTaxpayer(ctx context.Context, userID, taxpayerID string) error {
	f.calls = append(f.calls, userOf(ctx)+"|"+userID+"|"+taxpayerID)
	return f.err
}

// hero is the opaque full estimate the dashboard passes through.
type hero struct{ Period string }

// fakeEstimator answers one MonthTax per month of the window, from byMonth
// (zero figures for the others).
type fakeEstimator struct {
	byMonth map[string]usecase.MonthTax
	err     error
	calls   []string
}

func (f *fakeEstimator) EstimateWindow(ctx context.Context, userID, taxpayerID string, from, to domain.Month) (usecase.TaxWindow[hero], error) {
	f.calls = append(f.calls, userOf(ctx)+"|"+userID+"|"+taxpayerID+"|"+from.String()+".."+to.String())
	if f.err != nil {
		return usecase.TaxWindow[hero]{}, f.err
	}
	var out usecase.TaxWindow[hero]
	for m := from; !to.Before(m); m = m.AddMonths(1) {
		tax, ok := f.byMonth[m.String()]
		if !ok {
			tax = usecase.MonthTax{IVAOutcome: usecase.OutcomeZero}
		}
		tax.Month = m
		out.Months = append(out.Months, tax)
	}
	out.Current, out.Legend = hero{Period: to.String()}, "legend"
	return out, nil
}

// fakeRecords serves monthly totals and the existence checks.
type fakeRecords struct {
	incomes, expenses       map[string]money.Cents
	hasIncomes, hasExpenses bool
	err                     error
	calls                   []string
}

func (f *fakeRecords) totals(ctx context.Context, name string, sums map[string]money.Cents, from, to domain.Month) (map[domain.Month]money.Cents, error) {
	f.calls = append(f.calls, name+"|"+userOf(ctx)+"|"+from.String()+".."+to.String())
	out := map[domain.Month]money.Cents{}
	for raw, c := range sums {
		m, err := domain.ParseMonth(raw)
		if err != nil {
			return nil, err
		}
		out[m] = c
	}
	return out, f.err
}

func (f *fakeRecords) IncomeCollected(ctx context.Context, _ string, from, to domain.Month) (map[domain.Month]money.Cents, error) {
	return f.totals(ctx, "incomes", f.incomes, from, to)
}

func (f *fakeRecords) ExpensesPaid(ctx context.Context, _ string, from, to domain.Month) (map[domain.Month]money.Cents, error) {
	return f.totals(ctx, "expenses", f.expenses, from, to)
}

func (f *fakeRecords) HasIncomes(ctx context.Context, _ string) (bool, error) {
	f.calls = append(f.calls, "has incomes|"+userOf(ctx))
	return f.hasIncomes, f.err
}

func (f *fakeRecords) HasExpenses(ctx context.Context, _ string) (bool, error) {
	f.calls = append(f.calls, "has expenses|"+userOf(ctx))
	return f.hasExpenses, f.err
}

// fakeAccess answers a fixed summary.
type fakeAccess struct {
	summary access.Summary
	err     error
	users   []string
}

func (f *fakeAccess) Summary(ctx context.Context, userID string) (access.Summary, error) {
	f.users = append(f.users, userOf(ctx)+"|"+userID)
	return f.summary, f.err
}

// fixture wires the service over fakes, every module ON, a pro user.
type fixture struct {
	tx        *fakeTx
	flags     fakeFlags
	taxpayers *fakeTaxpayers
	estimator *fakeEstimator
	records   *fakeRecords
	access    *fakeAccess
	clock     clock.Clock
}

const (
	flagTax      = "module_tax_engine"
	flagIncomes  = "module_incomes"
	flagExpenses = "module_expenses"
)

func newFixture() *fixture {
	return &fixture{
		tx:        &fakeTx{},
		flags:     fakeFlags{flagTax: true, flagIncomes: true, flagExpenses: true},
		taxpayers: &fakeTaxpayers{},
		estimator: &fakeEstimator{},
		records:   &fakeRecords{},
		access:    &fakeAccess{summary: access.Summary{Status: access.Status{Level: access.LevelPro}}},
		clock:     clock.NewFixed(today),
	}
}

func (f *fixture) deps() usecase.Deps[hero] {
	return usecase.Deps[hero]{
		Tx: f.tx, Flags: f.flags, Taxpayers: f.taxpayers, Estimator: f.estimator, Records: f.records,
		Access: f.access, Clock: f.clock,
	}
}

func (f *fixture) service() *usecase.Service[hero] {
	s, err := usecase.NewService(f.deps())
	if err != nil {
		panic(err)
	}
	return s
}

func (f *fixture) get(ctx context.Context, req usecase.Request) (usecase.Dashboard[hero], error) {
	return f.service().Get(ctx, userA, req)
}

// appCode returns the contract code and the field details of err.
func appCode(err error) (apperr.Code, []apperr.FieldError) {
	app, ok := errors.AsType[*apperr.Error](err)
	if !ok {
		return "", nil
	}
	return app.Code, app.Fields
}
