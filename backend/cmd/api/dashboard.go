package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/cmd/internal/estimates"
	dashboardhttp "github.com/romodevx/control-contable/backend/internal/dashboard/adapter/http"
	dashdomain "github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	dashusecase "github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// errEstimateWindow reports a range estimate that does not answer one
// estimate per month of the window asked for (a wiring bug, 500).
var errEstimateWindow = errors.New("dashboard: the range estimate does not cover the window")

// dashboardHandler serves module_dashboard (M6-03); the estimate it embeds
// is the tax engine's presented body.
type dashboardHandler = dashboardhttp.Handler[taxenginehttp.EstimateBody]

// dashboardWiring is what the dashboard needs from the composition root.
// tx must be the taxpayers module's manager: the estimator, the profile
// read and the access policy all join the dashboard's unit of work.
type dashboardWiring struct {
	tx        tx.Manager
	flags     dashusecase.Flags
	taxpayers tpusecase.TaxpayerProfileReader
	estimator *taxusecase.Estimator
	limits    *ratelimitpg.Store
	// clock is the dashboard's "today" (default period, trial days left);
	// nil is the system clock.
	clock clock.Clock
}

// dashboardModule is the dashboard's HTTP side and its per-user budget.
type dashboardModule struct {
	handler    *dashboardHandler
	userLimits map[string]router.Limiter
}

// newDashboardModule wires the dashboard use case (M6-02) over the shared
// unit of work.
func newDashboardModule(w dashboardWiring) (dashboardModule, error) {
	now := w.clock
	if now == nil {
		now = clock.NewSystem()
	}
	if w.estimator == nil || w.taxpayers == nil {
		return dashboardModule{}, fmt.Errorf("%w: dashboard wiring", dashusecase.ErrMissingDependency)
	}
	svc, err := dashusecase.NewService(dashusecase.Deps[taxenginehttp.EstimateBody]{
		Tx:        w.tx,
		Flags:     w.flags,
		Taxpayers: dashboardTaxpayers{reader: w.taxpayers},
		Estimator: dashboardEstimates{estimator: w.estimator},
		Records:   dashboardRecords{incomes: incomesinfra.NewLedgerReader(), expenses: expensesinfra.NewLedgerReader()},
		// The access decision is the wall clock's, as every record module's
		// policy; only the dashboard's own "today" may be pinned (tests).
		Access: access.NewPolicy(w.tx, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Clock:  now,
	})
	if err != nil {
		return dashboardModule{}, err
	}
	limits := make(map[string]router.Limiter, len(dashboardhttp.UserLimits))
	for op, policy := range dashboardhttp.UserLimits {
		limits[op] = w.limits.Bucket(policy)
	}
	return dashboardModule{handler: dashboardhttp.NewHandler(svc), userLimits: limits}, nil
}

// dashboardTaxpayers adapts the taxpayers module's profile reader to the
// dashboard's ownership check (its 404 passes as is).
type dashboardTaxpayers struct {
	reader tpusecase.TaxpayerProfileReader
}

var _ dashusecase.TaxpayerChecker = dashboardTaxpayers{}

// CheckTaxpayer implements dashusecase.TaxpayerChecker.
func (a dashboardTaxpayers) CheckTaxpayer(ctx context.Context, userID, taxpayerID string) error {
	_, err := a.reader.Profile(ctx, userID, taxpayerID)
	return err
}

// rangeEstimator is the tax engine's range estimate (*taxusecase.Estimator).
type rangeEstimator interface {
	EstimateRange(ctx context.Context, userID, taxpayerID string, from, to domain.Period) ([]taxusecase.TaxpayerEstimate, error)
}

// dashboardEstimates adapts the tax engine's range estimate to the
// dashboard's TaxEstimator port: each month's totals, regime statuses and
// record totals, and the last month presented as getTaxpayerEstimate
// answers it.
type dashboardEstimates struct {
	estimator rangeEstimator
}

var _ dashusecase.TaxEstimator[taxenginehttp.EstimateBody] = dashboardEstimates{}

// EstimateWindow implements dashusecase.TaxEstimator.
func (a dashboardEstimates) EstimateWindow(ctx context.Context, userID, taxpayerID string, from, to dashdomain.Month) (
	dashusecase.TaxWindow[taxenginehttp.EstimateBody], error,
) {
	var out dashusecase.TaxWindow[taxenginehttp.EstimateBody]
	first, err := domain.NewPeriod(from.Year(), from.Month())
	if err != nil {
		return out, err
	}
	last, err := domain.NewPeriod(to.Year(), to.Month())
	if err != nil {
		return out, err
	}
	estimates, err := a.estimator.EstimateRange(ctx, userID, taxpayerID, first, last)
	if err != nil {
		return out, err
	}
	want := 0
	for m := from; !to.Before(m); m = m.AddMonths(1) {
		want++
	}
	if len(estimates) != want {
		return out, fmt.Errorf("%w: %d estimates for %s..%s", errEstimateWindow, len(estimates), from, to)
	}
	for _, e := range estimates {
		out.Months = append(out.Months, monthTaxOf(e))
	}
	current := estimates[len(estimates)-1]
	out.Current, out.Legend = taxenginehttp.PresentEstimate(current), current.Legend
	return out, nil
}

// monthTaxOf is one estimate as a trend month.
func monthTaxOf(e taxusecase.TaxpayerEstimate) dashusecase.MonthTax {
	statuses := make([]dashdomain.Status, 0, len(e.Regimes))
	for _, r := range e.Regimes {
		statuses = append(statuses, dashdomain.Status(r.Status))
	}
	return dashusecase.MonthTax{
		Month:               dashdomain.NewMonth(e.Period.Year(), e.Period.Month()),
		IVAPayable:          e.Totals.IVA.Payable(),
		ISRPayable:          e.Totals.ISR.Payable(),
		WithholdingsToRemit: e.Totals.WithholdingsToRemit,
		TotalPayable:        e.Totals.TotalPayable,
		IVAOutcome:          dashusecase.Outcome(e.Totals.IVA.Outcome()),
		Regimes:             statuses,
		IncomeCollected:     e.Records.IncomeCollected,
		ExpensesPaid:        e.Records.ExpensesPaid,
	}
}

// incomeRecords is the incomes module's aggregate reader with its
// existence check (incomesinfra.LedgerReader).
type incomeRecords interface {
	estimates.IncomeSumsReader
	HasLive(ctx context.Context, taxpayerID string) (bool, error)
}

// expenseRecords is the expenses module's (expensesinfra.LedgerReader).
type expenseRecords interface {
	estimates.ExpenseSumsReader
	HasLive(ctx context.Context, taxpayerID string) (bool, error)
}

// dashboardRecords adapts the incomes and expenses ledger readers to the
// dashboard's RecordReader (the engine-OFF figures and is_empty): the same
// one-statement range aggregate the engine reads, summed per month over
// every regime.
type dashboardRecords struct {
	incomes  incomeRecords
	expenses expenseRecords
}

var _ dashusecase.RecordReader = dashboardRecords{}

// IncomeCollected implements dashusecase.RecordReader: amount_cents by
// collected_on (the rows' Cash), before IVA.
func (a dashboardRecords) IncomeCollected(ctx context.Context, taxpayerID string, from, to dashdomain.Month) (map[dashdomain.Month]money.Cents, error) {
	rows, err := a.incomes.IncomeSums(ctx, incomesinfra.LedgerQuery{
		TaxpayerID: taxpayerID, FirstYear: from.Year(), FiscalYear: to.Year(), Through: to.Month(),
	})
	if err != nil {
		return nil, err
	}
	totals := monthTotals{from: from, to: to, sums: map[dashdomain.Month]money.Cents{}}
	for _, r := range rows {
		totals.add(r.Year, r.Month, r.Cash)
	}
	return totals.result()
}

// ExpensesPaid implements dashusecase.RecordReader: amount_cents by
// paid_on, every deductibility class, before IVA.
func (a dashboardRecords) ExpensesPaid(ctx context.Context, taxpayerID string, from, to dashdomain.Month) (map[dashdomain.Month]money.Cents, error) {
	rows, err := a.expenses.ExpenseSums(ctx, expensesinfra.LedgerQuery{
		TaxpayerID: taxpayerID, FirstYear: from.Year(), FiscalYear: to.Year(), Through: to.Month(),
	})
	if err != nil {
		return nil, err
	}
	totals := monthTotals{from: from, to: to, sums: map[dashdomain.Month]money.Cents{}}
	for _, r := range rows {
		totals.add(r.Year, r.Month, r.DeductiblePaid, r.NonDeductiblePaid, r.CashOverLimitPaid)
	}
	return totals.result()
}

// HasIncomes implements dashusecase.RecordReader.
func (a dashboardRecords) HasIncomes(ctx context.Context, taxpayerID string) (bool, error) {
	return a.incomes.HasLive(ctx, taxpayerID)
}

// HasExpenses implements dashusecase.RecordReader.
func (a dashboardRecords) HasExpenses(ctx context.Context, taxpayerID string) (bool, error) {
	return a.expenses.HasLive(ctx, taxpayerID)
}

// monthTotals sums amounts per month of [from, to], overflow-checked.
type monthTotals struct {
	from, to dashdomain.Month
	sums     map[dashdomain.Month]money.Cents
	errs     []error
}

func (m *monthTotals) add(year int, month time.Month, amounts ...money.Cents) {
	key := dashdomain.NewMonth(year, month)
	if key.Before(m.from) || m.to.Before(key) {
		return
	}
	for _, c := range amounts {
		sum, err := m.sums[key].Add(c)
		m.errs = append(m.errs, err)
		m.sums[key] = sum
	}
}

func (m *monthTotals) result() (map[dashdomain.Month]money.Cents, error) {
	if err := errors.Join(m.errs...); err != nil {
		return nil, fmt.Errorf("dashboard: month totals: %w", err)
	}
	return m.sums, nil
}
