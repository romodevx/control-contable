// Package estimates wires the tax engine's monthly Estimator (M5-22) for the
// composition roots that compute estimates: cmd/api (the estimate, the
// dashboard, the history) and cmd/worker (the reminders' amounts, M9-06).
// The modules never import each other (archtest cross-module), so the
// adapters that translate the taxpayers, incomes and expenses modules' types
// into the tax engine's ports live here, where only composition roots see
// them.
package estimates

import (
	"fmt"

	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// Ledger is the incomes and expenses aggregate readers as the tax engine's
// one LedgerReader port.
type Ledger struct {
	IncomeLedger
	ExpenseLedger
}

var _ usecase.LedgerReader = Ledger{}

// NewLedger reads both tables on the unit of work in ctx.
func NewLedger() Ledger {
	return Ledger{
		IncomeLedger:  NewIncomeLedger(incomesinfra.NewLedgerReader()),
		ExpenseLedger: NewExpenseLedger(expensesinfra.NewLedgerReader()),
	}
}

// MonthlyDeps are what the monthly Estimator needs from a composition root.
// Tx must be the manager the taxpayers profile reader runs on, so the
// profile joins the estimate's unit of work (tx.ErrForeignTransaction
// otherwise).
type MonthlyDeps struct {
	Tx        tx.Manager
	Taxpayers tpusecase.TaxpayerProfileReader
	// Holidays are the due date's holidays (a boot snapshot in the API).
	Holidays usecase.HolidayCalendar
	// Clock is the estimate's "today" (the default period, the
	// future-month refusal).
	Clock clock.Clock
}

// NewMonthly builds the immutable strategy registry (every strategy the
// engine ships, taxengine.Strategies; ADR-0002 §1) and the embedded
// parameters (a bad file fails the boot), and wires the estimate's read
// ports over d.Tx. A missing port is usecase.ErrMissingDependency.
func NewMonthly(d MonthlyDeps) (*usecase.Estimator, error) {
	if d.Taxpayers == nil {
		return nil, fmt.Errorf("%w: MonthlyDeps.Taxpayers", usecase.ErrMissingDependency)
	}
	registry, err := domain.NewRegistry(taxengine.Strategies()...)
	if err != nil {
		return nil, fmt.Errorf("taxengine: strategy registry: %w", err)
	}
	catalog, err := params.Embedded()
	if err != nil {
		return nil, err
	}
	store := taxinputs.NewStore()
	return usecase.NewEstimator(usecase.EstimateDeps{
		Tx:         d.Tx,
		Profiles:   NewProfiles(d.Taxpayers),
		Ledger:     NewLedger(),
		Inputs:     store,
		Texts:      store,
		Holidays:   d.Holidays,
		Strategies: registry,
		Params:     catalog,
		Clock:      d.Clock,
	})
}
