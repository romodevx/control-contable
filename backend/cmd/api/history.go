package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	historyhttp "github.com/romodevx/control-contable/backend/internal/history/adapter/http"
	histdomain "github.com/romodevx/control-contable/backend/internal/history/domain"
	historyinfra "github.com/romodevx/control-contable/backend/internal/history/infra"
	histusecase "github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// historyHandler serves module_history (M8-04).
type historyHandler = historyhttp.Handler

// historyWiring is what History needs from the composition root. tx must
// be the taxpayers module's manager: the estimator (and its profile read)
// and the views store all join History's unit of work.
type historyWiring struct {
	tx        tx.Manager
	flags     histusecase.Flags
	estimator rangeEstimator
	limits    *ratelimitpg.Store
	// clock is History's "today" (current month, viewed_at); nil is the
	// system clock.
	clock clock.Clock
}

// historyModule is History's HTTP side and its per-user budgets.
type historyModule struct {
	handler    *historyHandler
	userLimits map[string]router.Limiter
}

// newHistoryModule wires the history use case (M8-03) over the shared unit
// of work.
func newHistoryModule(w historyWiring) (historyModule, error) {
	if w.estimator == nil || w.limits == nil {
		return historyModule{}, fmt.Errorf("%w: history wiring", histusecase.ErrMissingDependency)
	}
	now := w.clock
	if now == nil {
		now = clock.NewSystem()
	}
	svc, err := histusecase.NewService(histusecase.Deps{
		Tx: w.tx, Flags: w.flags, Estimator: historyEstimates{estimator: w.estimator},
		Views: historyinfra.NewViewStore(), Clock: now,
	})
	if err != nil {
		return historyModule{}, err
	}
	limits := make(map[string]router.Limiter, len(historyhttp.UserLimits))
	for op, policy := range historyhttp.UserLimits {
		limits[op] = w.limits.Bucket(policy)
	}
	return historyModule{handler: historyhttp.NewHandler(svc), userLimits: limits}, nil
}

// historyEstimates adapts the tax engine's range estimate to History's
// Estimator port: one EstimateRange for the months (one aggregate pass,
// MODULE-history §2.6), each month reduced to what History shows and
// compares. Disclaimer texts are the ones getTaxpayerEstimate renders.
type historyEstimates struct {
	estimator rangeEstimator
}

var _ histusecase.Estimator = historyEstimates{}

// EstimateMonths implements histusecase.Estimator.
func (a historyEstimates) EstimateMonths(ctx context.Context, userID, taxpayerID string, from, to histdomain.Month) (
	histusecase.Estimates, error,
) {
	first, err := domain.NewPeriod(from.Year(), from.Month())
	if err != nil {
		return histusecase.Estimates{}, err
	}
	last, err := domain.NewPeriod(to.Year(), to.Month())
	if err != nil {
		return histusecase.Estimates{}, err
	}
	estimates, err := a.estimator.EstimateRange(ctx, userID, taxpayerID, first, last)
	if err != nil {
		return histusecase.Estimates{}, err
	}
	want := (to.Year()-from.Year())*monthsPerYear + int(to.Month()-from.Month()) + 1
	if len(estimates) != want || estimates[0].Period != first || estimates[want-1].Period != last {
		return histusecase.Estimates{}, fmt.Errorf("%w: %d estimates for %s..%s", histusecase.ErrEstimateCoverage, len(estimates), from, to)
	}
	out := histusecase.Estimates{Months: make([]histusecase.MonthEstimate, 0, len(estimates)), Legend: estimates[0].Legend}
	for _, e := range estimates {
		out.Months = append(out.Months, historyMonthOf(e))
	}
	return out, nil
}

// monthsPerYear is the number of calendar months in a year.
const monthsPerYear = 12

// historyMonthOf is one estimate as History reads it.
func historyMonthOf(e taxusecase.TaxpayerEstimate) histusecase.MonthEstimate {
	statuses := make([]histdomain.Status, 0, len(e.Regimes))
	for _, r := range e.Regimes {
		statuses = append(statuses, histdomain.Status(r.Status))
	}
	iva := e.Totals.IVA
	set := histdomain.ParameterSet{FiscalYear: e.ParameterSet.FiscalYear, Version: e.ParameterSet.Version}
	return histusecase.MonthEstimate{
		Month:   histdomain.NewMonth(e.Period.Year(), e.Period.Month()),
		Regimes: statuses,
		Tax: histdomain.Tax{
			ISRPayable: e.Totals.ISR.Payable(), IVAPayable: iva.Payable(), IVAOutcome: string(iva.Outcome()),
			IVAInFavor: max(-iva.Amount, 0), WithholdingsToRemit: e.Totals.WithholdingsToRemit, TotalPayable: e.Totals.TotalPayable,
		},
		DueDate: e.DueDate.String(),
		CPAReviewPending: slices.ContainsFunc(e.Regimes, func(r taxusecase.RegimeResult) bool {
			return r.CPAReviewPending()
		}),
		Seen: histdomain.Seen{
			RecordsFingerprint: e.RecordsFingerprint, FiguresFingerprint: e.FiguresFingerprint,
			EngineVersion: e.EngineVersion, ParameterSet: set.String(),
		},
		ParameterSet: set,
		Disclaimers:  historyDisclaimers(e),
		HasIncomes:   e.Records.HasIncomes,
		HasExpenses:  e.Records.HasExpenses,
	}
}

// historyDisclaimers are the estimate's disclaimers with the texts
// getTaxpayerEstimate renders for them.
func historyDisclaimers(e taxusecase.TaxpayerEstimate) []histdomain.Disclaimer {
	presented := taxenginehttp.PresentEstimate(e).Disclaimers
	out := make([]histdomain.Disclaimer, 0, len(presented))
	for _, d := range presented {
		out = append(out, histdomain.Disclaimer{Code: d.Code, Severity: histdomain.Severity(d.Severity), Text: d.Text})
	}
	return out
}
