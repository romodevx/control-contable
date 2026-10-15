package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// The record modules whose flags shape has_records (MODULE-history §3).
const (
	ModuleIncomes  = "module_incomes"
	ModuleExpenses = "module_expenses"
)

// Deps are the ports History needs. Tx must be the manager every other port
// runs on (they join its unit of work; tx.ErrForeignTransaction otherwise).
type Deps struct {
	Tx        tx.Manager
	Flags     Flags
	Estimator Estimator
	Views     ViewStore
	// Clock is "today" in Mexico City: the current month and viewed_at.
	Clock clock.Clock
}

// Month is one row of the fiscal year (contract HistoryMonth).
type Month struct {
	Month      domain.Month
	Status     domain.Status
	HasRecords bool
	// Tax is nil when the status is none or informational.
	Tax              *domain.Tax
	DueDate          string
	CPAReviewPending bool
	IsCurrent        bool
	// Seen is the current estimate's fingerprints and versions.
	Seen         domain.Seen
	ParameterSet domain.ParameterSet
	// ViewedAt is the caller's last view; nil if never viewed.
	ViewedAt *time.Time
	// Change is "" unless the month changed since viewed (§2.4).
	Change domain.ChangeReason
}

// Year is getHistory's answer (contract HistoryYear).
type Year struct {
	TaxpayerID  string
	FiscalYear  int
	Months      []Month // newest first
	Totals      domain.YearTotals
	Legend      string
	Disclaimers []domain.Disclaimer
}

// ViewedRequest is markHistoryViewed's input: the path's period and what
// the app rendered, copied from the TaxpayerEstimate.
type ViewedRequest struct {
	Period             string
	RecordsFingerprint string
	FiguresFingerprint string
	EngineVersion      string
	ParamsYear         int
	ParamsVersion      int
}

// Service is the history use case. It is safe for concurrent use.
type Service struct {
	deps Deps
}

// NewService returns the use case over deps; a missing port is
// ErrMissingDependency, so wiring fails the boot.
func NewService(deps Deps) (*Service, error) {
	ports := []struct {
		name    string
		missing bool
	}{
		{"Tx", deps.Tx == nil}, {"Flags", deps.Flags == nil}, {"Estimator", deps.Estimator == nil},
		{"Views", deps.Views == nil}, {"Clock", deps.Clock == nil},
	}
	for _, p := range ports {
		if p.missing {
			return nil, fmt.Errorf("%w: Deps.%s", ErrMissingDependency, p.name)
		}
	}
	return &Service{deps: deps}, nil
}

// Get answers getHistory for the caller's taxpayer and fiscalYear (§2):
// the id (404), the year (400 future, 422 before 2026), then, in one
// read-only unit of work, one range estimate of the year's months (its
// profile read is the ownership 404) and one read of the caller's views.
// Every figure is the engine's; nothing is stored. Any failure fails the
// whole request (no partial mode).
func (s *Service) Get(ctx context.Context, userID, taxpayerID string, fiscalYear int) (Year, error) {
	id, err := uuid.Parse(taxpayerID)
	if err != nil {
		return Year{}, notFound(err)
	}
	current := domain.MonthOf(s.deps.Clock.Now())
	switch {
	case fiscalYear > current.Year():
		return Year{}, invalidField(ErrFuture, fieldFiscalYear, DetailOutOfRange, msgFutureYear)
	case fiscalYear < domain.FirstFiscalYear:
		return Year{}, unavailable(fieldFiscalYear, msgNoYearData, msgNoYearDataField)
	}
	months := domain.YearMonths(fiscalYear, current)
	from, to := months[len(months)-1], months[0]
	var (
		estimates Estimates
		views     []domain.View
	)
	err = s.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		if estimates, err = s.deps.Estimator.EstimateMonths(ctx, userID, id.String(), from, to); err != nil {
			return hidden(err)
		}
		if views, err = s.deps.Views.Views(ctx, id.String(), from, to); err != nil {
			return fmt.Errorf("history: views: %w", err)
		}
		return nil
	})
	if err != nil {
		return Year{}, err
	}
	records := recordModules{incomes: s.deps.Flags.Enabled(ctx, ModuleIncomes), expenses: s.deps.Flags.Enabled(ctx, ModuleExpenses)}
	return build(id.String(), fiscalYear, current, months, estimates, views, records)
}

// hidden rewords a port's 404 (an unknown or another user's taxpayer) as
// History's, indistinguishable from module_history OFF (§8).
func hidden(err error) error {
	if app, ok := errors.AsType[*apperr.Error](err); ok && app.Code == apperr.CodeNotFound {
		return notFound(err)
	}
	return err
}

// recordModules is the state of the flags that shape has_records.
type recordModules struct{ incomes, expenses bool }

// build assembles the year from the engine's months (oldest first) and
// the caller's views, newest first.
func build(taxpayerID string, fiscalYear int, current domain.Month, months []domain.Month, est Estimates,
	views []domain.View, records recordModules,
) (Year, error) {
	byMonth := make(map[domain.Month]MonthEstimate, len(est.Months))
	for _, m := range est.Months {
		byMonth[m.Month] = m
	}
	viewed := make(map[domain.Month]domain.View, len(views))
	for _, v := range views {
		viewed[v.Month] = v
	}
	out := Year{TaxpayerID: taxpayerID, FiscalYear: fiscalYear, Legend: est.Legend, Months: make([]Month, 0, len(months))}
	disclaimers := make([][]domain.Disclaimer, 0, len(months))
	for _, m := range months {
		e, ok := byMonth[m]
		if !ok {
			return Year{}, fmt.Errorf("%w: %s", ErrEstimateCoverage, m)
		}
		row := monthRow(e, m == current, records)
		if v, ok := viewed[m]; ok {
			at := v.ViewedAt
			row.ViewedAt, row.Change = &at, domain.Change(&v, e.Seen, row.IsCurrent)
		}
		if row.Tax != nil {
			if err := out.Totals.Add(*row.Tax); err != nil {
				return Year{}, err
			}
		}
		out.Months = append(out.Months, row)
		disclaimers = append(disclaimers, e.Disclaimers)
	}
	out.Disclaimers = domain.UnionDisclaimers(disclaimers)
	return out, nil
}

// monthRow is one month before its view is applied.
func monthRow(e MonthEstimate, isCurrent bool, records recordModules) Month {
	row := Month{
		Month: e.Month, Status: domain.TaxpayerStatus(e.Regimes),
		HasRecords: (records.incomes && e.HasIncomes) || (records.expenses && e.HasExpenses),
		DueDate:    e.DueDate, CPAReviewPending: e.CPAReviewPending, IsCurrent: isCurrent,
		Seen: e.Seen, ParameterSet: e.ParameterSet,
	}
	if row.Status != domain.StatusNone && row.Status != domain.StatusInformational {
		tax := e.Tax
		row.Tax = &tax
	}
	return row
}

// MarkViewed answers markHistoryViewed (§2.5): it stores, for the caller,
// what the app rendered for req.Period, stamped with the server clock, in
// one upsert (idempotent: the last view wins). Checked first: the id
// (404), the period (400 format or future, 422 before 2026) and the
// formats of what was seen (400). The current month is accepted (it is
// never marked). No estimate is computed and no access check runs: a
// read-only user and an archived taxpayer may mark (it creates no record).
// A taxpayer that is not the caller's is the hidden 404.
func (s *Service) MarkViewed(ctx context.Context, userID, taxpayerID string, req ViewedRequest) error {
	id, err := uuid.Parse(taxpayerID)
	if err != nil {
		return notFound(err)
	}
	view, err := s.view(req)
	if err != nil {
		return err
	}
	err = s.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		return s.deps.Views.SaveView(ctx, userID, id.String(), view)
	})
	switch {
	case errors.Is(err, ErrTaxpayerNotFound):
		return notFound(err)
	case err != nil:
		return fmt.Errorf("history: save view: %w", err)
	}
	return nil
}

// view checks req and builds the view to store.
func (s *Service) view(req ViewedRequest) (domain.View, error) {
	now := s.deps.Clock.Now()
	period, err := domain.ParseMonth(req.Period)
	switch {
	case err != nil:
		return domain.View{}, invalidField(err, fieldPeriod, DetailPattern, msgPeriodPattern)
	case domain.MonthOf(now).Before(period):
		return domain.View{}, invalidField(ErrFuture, fieldPeriod, DetailOutOfRange, msgFuturePeriod)
	case period.Year() < domain.FirstFiscalYear:
		return domain.View{}, unavailable(fieldPeriod, msgNoPeriodData, msgNoPeriodDataField)
	}
	set, err := domain.NewParameterSet(req.ParamsYear, req.ParamsVersion)
	if err != nil {
		return domain.View{}, invalidSeen(err)
	}
	seen := domain.Seen{
		RecordsFingerprint: req.RecordsFingerprint, FiguresFingerprint: req.FiguresFingerprint,
		EngineVersion: req.EngineVersion, ParameterSet: set.String(),
	}
	if err := seen.Validate(); err != nil {
		return domain.View{}, invalidSeen(err)
	}
	return domain.View{Month: period, Seen: seen, ViewedAt: now}, nil
}

// invalidSeen is the 400 on the field a client-sent value broke; the
// value itself is never echoed.
func invalidSeen(err error) error {
	field := "body"
	if invalid, ok := errors.AsType[*domain.InvalidSeenError](err); ok {
		field = invalid.Field
	}
	return invalidField(err, field, DetailPattern, msgInvalidFormat)
}
