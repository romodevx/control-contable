package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"maps"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
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
		handler: taxenginehttp.NewHandler(inputs, estimator), userLimits: out, estimator: estimator, clock: w.clock,
	}, nil
}

// newEstimator builds the immutable strategy registry (every strategy the
// engine ships, taxengine.Strategies; ADR-0002 §1) and the embedded
// parameters (a bad file fails the boot), and wires the estimate's read
// ports over the shared unit of work.
func newEstimator(w taxEngineWiring) (*usecase.Estimator, error) {
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	if err != nil {
		return nil, fmt.Errorf("taxengine: strategy registry: %w", err)
	}
	catalog, err := params.Embedded()
	if err != nil {
		return nil, err
	}
	now := w.clock
	if now == nil {
		now = clock.NewSystem()
	}
	return usecase.NewEstimator(usecase.EstimateDeps{ // a missing port fails the boot (FU4)
		Tx:       w.tx,
		Profiles: estimateProfiles{reader: w.taxpayers},
		Ledger: taxLedger{
			incomeLedger:  newIncomeLedger(incomesinfra.NewLedgerReader()),
			expenseLedger: newExpenseLedger(expensesinfra.NewLedgerReader()),
		},
		Inputs:     taxinputs.NewStore(),
		Texts:      taxinputs.NewStore(),
		Holidays:   noHolidays{},
		Strategies: registry,
		Params:     catalog,
		Clock:      now,
	})
}

// taxLedger is the incomes and expenses aggregate readers as the tax
// engine's one LedgerReader port.
type taxLedger struct {
	incomeLedger
	expenseLedger
}

var _ usecase.LedgerReader = taxLedger{}

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
