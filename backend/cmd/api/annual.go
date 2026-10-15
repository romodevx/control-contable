package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annualinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	taxholidays "github.com/romodevx/control-contable/backend/internal/taxengine/infra/holidays"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// annualHandler serves module_annual (ANN-15, ADR-0009).
type annualHandler = annualhttp.Handler

// errAnnualWiring reports annual wiring without the engine or the shared
// unit of work.
var errAnnualWiring = errors.New("annual: the tax engine's estimator and the taxpayers module's unit of work are required")

// annualWiring is what the annual estimate needs from the composition
// root. tx must be the taxpayers module's manager: the estimate reads the
// profile and runs the monthly Estimator range nested in its unit of work,
// and the access policy joins the input writes' transaction.
type annualWiring struct {
	tx        tx.Manager
	taxpayers tpusecase.TaxpayerProfileReader
	estimator *usecase.Estimator
	limits    *ratelimitpg.Store
	logger    *slog.Logger
	// clock is the annual's "today" (the year in progress, the future-year
	// refusal); nil is the system clock.
	clock clock.Clock
}

// annualModule is the annual HTTP side and its per-user budgets.
type annualModule struct {
	handler    *annualHandler
	userLimits map[string]router.Limiter
}

// newAnnualModule wires the annual estimate (ANN-14) and the annual input
// use cases (ANN-13) over the shared unit of work.
func newAnnualModule(w annualWiring) (annualModule, error) {
	if w.tx == nil || w.estimator == nil || w.taxpayers == nil || w.limits == nil {
		return annualModule{}, errAnnualWiring
	}
	now := w.clock
	if now == nil {
		now = clock.NewSystem()
	}
	estimates, err := newAnnualEstimator(w, now)
	if err != nil {
		return annualModule{}, err
	}
	store := annualinputs.NewStore()
	policy := access.NewPolicy(w.tx, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem())
	limits := make(map[string]router.Limiter, len(annualhttp.UserLimits))
	for op, p := range annualhttp.UserLimits {
		limits[op] = w.limits.Bucket(p)
	}
	return annualModule{
		handler: annualhttp.NewHandler(annualhttp.Deps{
			Estimates: estimates,
			Facts: usecaseannual.NewInputs(usecaseannual.InputsDeps{
				Tx: w.tx, Access: policy, Taxpayers: store, Store: store, Clock: now, Logger: w.logger,
			}),
			Salary: usecaseannual.NewSalaryStatements(usecaseannual.SalaryDeps{
				Tx: w.tx, Access: policy, Taxpayers: store, Store: store, Clock: now, Logger: w.logger,
			}),
			Deductions: usecaseannual.NewDeductions(usecaseannual.DeductionsDeps{
				Tx: w.tx, Access: policy, Taxpayers: store, Store: store, Clock: now, Logger: w.logger,
			}),
			Filed: usecaseannual.NewFiledPayments(usecaseannual.FiledDeps{
				Tx: w.tx, Access: policy, Taxpayers: store, Store: store, Clock: now, Logger: w.logger,
			}),
			Range: w.limits.Bucket(annualhttp.RangePolicy),
		}),
		userLimits: limits,
	}, nil
}

// newAnnualEstimator builds the annual estimate: the composition list of
// sections (taxengine.AnnualSections) on the embedded parameters, the
// monthly registry (which regimes chain their default payments) and the
// read ports over the shared unit of work.
func newAnnualEstimator(w annualWiring, now clock.Clock) (*usecaseannual.Estimator, error) {
	catalog, err := params.Embedded()
	if err != nil {
		return nil, err
	}
	monthly, err := domain.NewRegistry(taxengine.Strategies()...)
	if err != nil {
		return nil, fmt.Errorf("annual: monthly registry: %w", err)
	}
	sections, err := usecaseannual.NewSectionRegistry(taxengine.AnnualSections(annuallimits.New(catalog))...)
	if err != nil {
		return nil, fmt.Errorf("annual: section registry: %w", err)
	}
	store := annualinputs.NewStore()
	return usecaseannual.NewEstimator(usecaseannual.EstimateDeps{ // a missing port fails the boot
		Tx:       w.tx,
		Profiles: estimateProfiles{reader: w.taxpayers},
		Ledger: taxLedger{
			incomeLedger:  newIncomeLedger(incomesinfra.NewLedgerReader()),
			expenseLedger: newExpenseLedger(expensesinfra.NewLedgerReader()),
		},
		TaxInputs:         taxinputs.NewStore(),
		YearInputs:        store,
		Filed:             store,
		Monthly:           w.estimator,
		MonthlyStrategies: monthly,
		Sections:          sections,
		Holidays:          taxholidays.NewReader(),
		Params:            catalog,
		Clock:             now,
	})
}
