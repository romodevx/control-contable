package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"maps"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/cmd/internal/estimates"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// errNoTaxEngineTx reports wiring without the shared unit of work.
var errNoTaxEngineTx = errors.New("taxengine: the taxpayers module's unit of work is required")

// taxEngineWiring is what the tax engine module needs from the composition
// root. tx must be the taxpayers module's manager: the estimate reads the
// taxpayer profile nested in its own unit of work, which joins only through
// the same manager (tx.ErrForeignTransaction otherwise).
type taxEngineWiring struct {
	pool      *sql.DB // cc_app
	limits    *ratelimitpg.Store
	logger    *slog.Logger
	tx        tx.Manager
	taxpayers tpusecase.TaxpayerProfileReader
	// holidays are the due date's holidays, a boot snapshot (loadHolidays
	// over the tax engine's reader): no statement per estimate.
	holidays usecase.HolidayCalendar
	// clock is the estimate's "today" (the default period, the future-month
	// refusal); nil is the system clock (tests pin a day).
	clock clock.Clock
}

// taxEngineModule is the tax engine's HTTP side: the handler of its built
// operations and its per-user write budget for the router, plus the
// Estimator that the estimate and breakdown handlers (M5-23), the
// dashboard (M6), the history (M8) and the reminders (M9) call.
type taxEngineModule struct {
	handler    *taxenginehttp.Handler
	userLimits map[string]router.Limiter
	estimator  *usecase.Estimator
	// holidays is the boot snapshot the annual estimate shares.
	holidays usecase.HolidayCalendar
	// clock is the estimate's "today" (nil: the system clock); the
	// dashboard shares it, so its default period is the estimate's.
	clock clock.Clock
}

// newTaxEngineModule wires the tax-inputs use cases (M5-21) over the cc_app
// unit of work, and the estimate (M5-22) over the taxpayers module's. The
// access policy shares the use cases' TxManager: it runs nested in their
// transaction, and a nested call through another manager is refused
// (tx.ErrForeignTransaction).
func newTaxEngineModule(w taxEngineWiring) (taxEngineModule, error) {
	if w.tx == nil {
		return taxEngineModule{}, errNoTaxEngineTx
	}
	txm, err := pgdb.NewTxManager(w.pool)
	if err != nil {
		return taxEngineModule{}, err
	}
	inputs := usecase.NewInputs(usecase.InputsDeps{
		Tx:     txm,
		Access: access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Store:  taxinputs.NewStore(), Clock: clock.NewSystem(), Logger: w.logger,
	})
	estimator, err := newEstimator(w)
	if err != nil {
		return taxEngineModule{}, err
	}
	out := make(map[string]router.Limiter, len(taxenginehttp.UserLimits))
	for op, policy := range taxenginehttp.UserLimits {
		out[op] = w.limits.Bucket(policy)
	}
	return taxEngineModule{
		handler: taxenginehttp.NewHandler(inputs, estimator), userLimits: out, estimator: estimator,
		holidays: w.holidays, clock: w.clock,
	}, nil
}

// newEstimator wires the monthly Estimator (estimates.NewMonthly) over the
// shared unit of work and the holidays snapshot.
func newEstimator(w taxEngineWiring) (*usecase.Estimator, error) {
	now := w.clock
	if now == nil {
		now = clock.NewSystem()
	}
	return estimates.NewMonthly(estimates.MonthlyDeps{ // a missing port fails the boot (FU4)
		Tx: w.tx, Taxpayers: w.taxpayers, Holidays: w.holidays, Clock: now,
	})
}

// mergeLimits joins the per-operation limits of every module. An operation
// limited by two modules is a wiring error: it would get one budget
// silently.
func mergeLimits[L any](modules ...map[string]L) (map[string]L, error) {
	out := map[string]L{}
	for _, m := range modules {
		for op := range m {
			if _, dup := out[op]; dup {
				return nil, fmt.Errorf("api: operation %q has two per-user limits", op)
			}
		}
		maps.Copy(out, m)
	}
	return out, nil
}
