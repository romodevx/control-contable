package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// The modules whose flags shape the dashboard (MODULE-dashboard §3).
const (
	ModuleTaxEngine = "module_tax_engine"
	ModuleIncomes   = "module_incomes"
	ModuleExpenses  = "module_expenses"
)

// Deps are the ports the dashboard needs. Tx must be the manager every
// other port runs on (they join its unit of work; tx.ErrForeignTransaction
// otherwise).
type Deps[D any] struct {
	Tx        tx.Manager
	Flags     Flags
	Taxpayers TaxpayerChecker
	Estimator TaxEstimator[D]
	Records   RecordReader
	Access    AccessReader
	Clock     clock.Clock
}

// Request is getDashboard's input. Period "" is the current month; Months 0
// is domain.DefaultMonths.
type Request struct {
	TaxpayerID string
	Period     string
	Months     int
}

// Figures are a month's income collected and expenses paid, before IVA;
// nil when the module is OFF (contract DashboardMonth).
type Figures struct {
	Income  *money.Cents
	Expense *money.Cents
}

// TrendTax is a month's tax figures (contract TrendTax).
type TrendTax struct {
	IVAPayable          money.Cents
	ISRPayable          money.Cents
	WithholdingsToRemit money.Cents
	TotalPayable        money.Cents
	IVAOutcome          Outcome
	Status              domain.Status
}

// TrendPoint is one month of the trend; Tax is nil when the engine is OFF.
type TrendPoint struct {
	Month domain.Month
	Figures
	Tax *TrendTax
}

// Dashboard is the Home of one taxpayer for Period (contract Dashboard).
type Dashboard[D any] struct {
	TaxpayerID string
	Period     domain.Month
	IsEmpty    bool
	// Current is the period's figures (contract month): the last point's.
	Current Figures
	// Estimate is the period's full estimate; nil when the engine is OFF.
	Estimate *D
	Trend    []TrendPoint
	// Access is the caller's access (contract AccessSummary), shaped by
	// access.Summary.View like every other screen's.
	Access access.View
	// Legend is "" exactly when Estimate is nil.
	Legend string
}

// Service is the dashboard use case (getDashboard). It is safe for
// concurrent use.
type Service[D any] struct {
	deps Deps[D]
}

// NewService returns the use case over deps; a missing port is
// ErrMissingDependency, so wiring fails the boot.
func NewService[D any](deps Deps[D]) (*Service[D], error) {
	ports := []struct {
		name    string
		missing bool
	}{
		{"Tx", deps.Tx == nil}, {"Flags", deps.Flags == nil}, {"Taxpayers", deps.Taxpayers == nil},
		{"Estimator", deps.Estimator == nil}, {"Records", deps.Records == nil}, {"Access", deps.Access == nil},
		{"Clock", deps.Clock == nil},
	}
	for _, p := range ports {
		if p.missing {
			return nil, fmt.Errorf("%w: Deps.%s", ErrMissingDependency, p.name)
		}
	}
	return &Service[D]{deps: deps}, nil
}

// modules is the effective state of the flags that shape one response.
type modules struct{ tax, incomes, expenses bool }

// Get builds the Home of req.TaxpayerID for the caller, read-only and
// derived on every read (MODULE-dashboard §2). The request is checked
// first (404 malformed id, 400 period/months, 422 before 2026-01); then,
// in one read-only unit of work: the engine's range estimate when
// module_tax_engine is ON (its profile read is the ownership 404), else the
// ownership check and the record totals of the enabled record modules
// only; the is_empty existence checks a window record does not already
// answer; and the access summary. Any failure fails the whole request.
func (s *Service[D]) Get(ctx context.Context, userID string, req Request) (Dashboard[D], error) {
	id, err := uuid.Parse(req.TaxpayerID)
	if err != nil {
		return Dashboard[D]{}, notFound(err)
	}
	period, err := s.period(req.Period)
	if err != nil {
		return Dashboard[D]{}, err
	}
	months := req.Months
	if months == 0 {
		months = domain.DefaultMonths
	}
	if months < domain.MinMonths || months > domain.MaxMonths {
		return Dashboard[D]{}, monthsError(months)
	}
	on := modules{
		tax:      s.deps.Flags.Enabled(ctx, ModuleTaxEngine),
		incomes:  s.deps.Flags.Enabled(ctx, ModuleIncomes),
		expenses: s.deps.Flags.Enabled(ctx, ModuleExpenses),
	}
	out := Dashboard[D]{TaxpayerID: id.String(), Period: period}
	window := domain.Window(period, months)
	err = s.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		return s.read(ctx, userID, on, window, &out)
	})
	if err != nil {
		return Dashboard[D]{}, hidden(err)
	}
	return out, nil
}

// hidden rewords a port's 404 (an unknown or another user's taxpayer) as
// the hidden module's, so the answer is indistinguishable from
// module_dashboard OFF (MODULE-dashboard §8); the cause stays for the logs.
func hidden(err error) error {
	if app, ok := errors.AsType[*apperr.Error](err); ok && app.Code == apperr.CodeNotFound {
		return notFound(err)
	}
	return err
}

// period resolves raw against the clock's current month.
func (s *Service[D]) period(raw string) (domain.Month, error) {
	current := domain.MonthOf(s.deps.Clock.Now())
	if raw == "" {
		return current, nil
	}
	p, err := domain.ParseMonth(raw)
	switch {
	case err != nil:
		return domain.Month{}, invalidPeriodError(err)
	case current.Before(p):
		return domain.Month{}, futurePeriodError()
	case p.Before(domain.FirstMonth):
		return domain.Month{}, parametersUnavailableError()
	}
	return p, nil
}

// read fills out inside the unit of work.
func (s *Service[D]) read(ctx context.Context, userID string, on modules, window []domain.Month, out *Dashboard[D]) error {
	from, to := window[0], window[len(window)-1]
	var (
		incomes, expenses map[domain.Month]money.Cents
		taxes             map[domain.Month]MonthTax
	)
	if on.tax {
		w, err := s.deps.Estimator.EstimateWindow(ctx, userID, out.TaxpayerID, from, to)
		if err != nil {
			return err
		}
		out.Estimate, out.Legend = &w.Current, w.Legend
		taxes, incomes, expenses = byMonth(w.Months)
	} else {
		var err error
		if incomes, expenses, err = s.recordTotals(ctx, userID, out.TaxpayerID, on, from, to); err != nil {
			return err
		}
	}
	out.Trend = trend(window, on, taxes, incomes, expenses)
	out.Current = out.Trend[len(out.Trend)-1].Figures
	empty, err := s.isEmpty(ctx, out.TaxpayerID, on, out.Trend)
	if err != nil {
		return err
	}
	out.IsEmpty = empty
	summary, err := s.deps.Access.Summary(ctx, userID)
	if err != nil {
		return fmt.Errorf("dashboard: access summary: %w", err)
	}
	out.Access = summary.View(s.deps.Clock.Now())
	return nil
}

// recordTotals checks the taxpayer and reads the enabled record modules'
// totals (the engine is OFF).
func (s *Service[D]) recordTotals(ctx context.Context, userID, taxpayerID string, on modules, from, to domain.Month) (
	incomes, expenses map[domain.Month]money.Cents, err error,
) {
	if err := s.deps.Taxpayers.CheckTaxpayer(ctx, userID, taxpayerID); err != nil {
		return nil, nil, err
	}
	if on.incomes {
		if incomes, err = s.deps.Records.IncomeCollected(ctx, taxpayerID, from, to); err != nil {
			return nil, nil, fmt.Errorf("dashboard: incomes collected: %w", err)
		}
	}
	if on.expenses {
		if expenses, err = s.deps.Records.ExpensesPaid(ctx, taxpayerID, from, to); err != nil {
			return nil, nil, fmt.Errorf("dashboard: expenses paid: %w", err)
		}
	}
	return incomes, expenses, nil
}

// byMonth indexes the engine's months and their record totals.
func byMonth(months []MonthTax) (taxes map[domain.Month]MonthTax, incomes, expenses map[domain.Month]money.Cents) {
	taxes = make(map[domain.Month]MonthTax, len(months))
	incomes = make(map[domain.Month]money.Cents, len(months))
	expenses = make(map[domain.Month]money.Cents, len(months))
	for _, m := range months {
		taxes[m.Month], incomes[m.Month], expenses[m.Month] = m, m.IncomeCollected, m.ExpensesPaid
	}
	return taxes, incomes, expenses
}

// trend builds one point per month of window, gaps as zeros, each figure
// present only when its module is ON.
func trend(window []domain.Month, on modules, taxes map[domain.Month]MonthTax, incomes, expenses map[domain.Month]money.Cents) []TrendPoint {
	out := make([]TrendPoint, 0, len(window))
	for _, m := range window {
		p := TrendPoint{Month: m}
		if on.incomes {
			p.Income = centsPtr(incomes[m])
		}
		if on.expenses {
			p.Expense = centsPtr(expenses[m])
		}
		if on.tax {
			p.Tax = trendTax(taxes[m])
		}
		out = append(out, p)
	}
	return out
}

func centsPtr(c money.Cents) *money.Cents { return &c }

// trendTax is a month's tax figures with the taxpayer-level status.
func trendTax(m MonthTax) *TrendTax {
	outcome := m.IVAOutcome
	if outcome == "" {
		outcome = OutcomeZero
	}
	return &TrendTax{
		IVAPayable: m.IVAPayable, ISRPayable: m.ISRPayable, WithholdingsToRemit: m.WithholdingsToRemit,
		TotalPayable: m.TotalPayable, IVAOutcome: outcome, Status: domain.TaxpayerStatus(m.Regimes),
	}
}

// isEmpty reports no live record of an enabled module in any period. A
// record in the window answers for its module without a read; with no
// record module ON it is false (nothing to invite the user to create).
func (s *Service[D]) isEmpty(ctx context.Context, taxpayerID string, on modules, points []TrendPoint) (bool, error) {
	checks := []struct {
		on     bool
		figure func(TrendPoint) *money.Cents
		exists func(context.Context, string) (bool, error)
	}{
		{on.incomes, func(p TrendPoint) *money.Cents { return p.Income }, s.deps.Records.HasIncomes},
		{on.expenses, func(p TrendPoint) *money.Cents { return p.Expense }, s.deps.Records.HasExpenses},
	}
	if !on.incomes && !on.expenses {
		return false, nil
	}
	for _, c := range checks {
		if !c.on {
			continue
		}
		if inWindow(points, c.figure) {
			return false, nil
		}
		found, err := c.exists(ctx, taxpayerID)
		if err != nil {
			return false, fmt.Errorf("dashboard: record existence: %w", err)
		}
		if found {
			return false, nil
		}
	}
	return true, nil
}

func inWindow(points []TrendPoint, figure func(TrendPoint) *money.Cents) bool {
	for _, p := range points {
		if c := figure(p); c != nil && *c != 0 {
			return true
		}
	}
	return false
}
