package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	expusecase "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The regime options that change an expense category's hint.
const (
	optionDeductionMode    = "deduction_mode"    // 606
	optionDefinitiveOption = "definitive_option" // 625
)

// expensesWiring is what the expenses module needs from the composition
// root. tx must be the taxpayers module's manager: the regime check joins
// the expense's unit of work (tx.ErrForeignTransaction otherwise).
type expensesWiring struct {
	tx        tx.Manager
	taxpayers *tpusecase.Queries
	rates     *rateCatalog
	limits    *ratelimitpg.Store
	clock     clock.Clock // nil = the system clock
	logger    *slog.Logger
}

// expensesModule is the expenses module's HTTP side: the handler of its
// operations and its per-user write budget for the router.
type expensesModule struct {
	handler    *expenseshttp.Handler
	userLimits map[string]router.Limiter
}

// newExpensesModule wires the expense use cases (M4-05, M4-06) over the
// shared cc_app unit of work, the access policy (trial cap), the taxpayers
// queries and the tax engine's RateCatalog over the embedded parameters.
func newExpensesModule(w expensesWiring) (expensesModule, error) {
	if w.tx == nil || w.taxpayers == nil || w.rates == nil {
		return expensesModule{}, errors.New("expenses: the shared unit of work, the taxpayers queries and the rate catalog are required")
	}
	parameters := newExpenseParameters(w.rates)
	clk := w.clock
	if clk == nil {
		clk = clock.NewSystem()
	}
	uc := expusecase.NewExpenses(expusecase.Deps{
		Tx:         w.tx,
		Access:     access.NewPolicy(w.tx, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Regimes:    w.taxpayers,
		Taxpayers:  taxpayerProfiles{reader: w.taxpayers},
		Expenses:   expinfra.NewStore(),
		Categories: expinfra.NewCategoryCatalog(),
		Parameters: parameters,
		// The kernel's view of the same catalog (fiscal.AmountKey).
		Withholdings: w.rates.Withholdings(),
		Clock:        clk,
		Logger:       w.logger,
	})
	return expensesModule{
		handler:    expenseshttp.NewHandler(uc),
		userLimits: bucketsOf(w.limits, expenseshttp.UserLimits),
	}, nil
}

// taxpayerProfiles adapts the taxpayers module's profile reader to the
// expenses TaxpayerProfiles port. Its client errors (404) pass as they are.
type taxpayerProfiles struct {
	reader tpusecase.TaxpayerProfileReader
}

var _ expusecase.TaxpayerProfiles = taxpayerProfiles{}

// Profile implements expusecase.TaxpayerProfiles.
func (a taxpayerProfiles) Profile(ctx context.Context, userID, taxpayerID string) (expusecase.TaxpayerProfile, error) {
	p, err := a.reader.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return expusecase.TaxpayerProfile{}, err
	}
	out := expusecase.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		out.Regimes = append(out.Regimes, regimeHold(r))
	}
	return out, nil
}

func regimeHold(r tpusecase.RegimeProfile) expusecase.RegimeHold {
	hold := expusecase.RegimeHold{Code: fiscal.RegimeCode(r.Code), StartPeriod: r.Window.Start.String()}
	if !r.Window.IsOpen() {
		hold.EndPeriod = r.Window.End.String()
	}
	if mode, ok := r.Options[optionDeductionMode].Text(); ok {
		hold.Options.DeductionMode = expdomain.DeductionMode(mode)
	}
	if definitive, ok := r.Options[optionDefinitiveOption].Bool(); ok {
		hold.Options.DefinitiveOption = definitive
	}
	return hold
}

// expenseParameters adapts the tax engine's RateCatalog (M5-08), bound to
// the expenses key type, to the expenses Parameters port.
type expenseParameters struct {
	*ratecatalog.Catalog[expdomain.AmountKey]
}

var _ expusecase.Parameters = expenseParameters{}

// newExpenseParameters binds the shared rate catalog to the expenses port.
func newExpenseParameters(rates *rateCatalog) expenseParameters {
	return expenseParameters{Catalog: rates}
}

// On implements expusecase.Parameters: a day no set covers (before the
// first published year) is ErrParametersUnavailable, which the use cases
// answer as 400 out_of_range on the date.
func (a expenseParameters) On(on fiscal.Date) (expusecase.DatedParameters, error) {
	limit, err := a.CashPaymentLimit(on)
	if errors.Is(err, ratecatalog.ErrDateNotCovered) {
		return expusecase.DatedParameters{}, expusecase.ErrParametersUnavailable
	}
	if err != nil {
		return expusecase.DatedParameters{}, err
	}
	rates, err := a.AllowedIVARates(on)
	if err != nil {
		return expusecase.DatedParameters{}, err
	}
	return expusecase.DatedParameters{
		IVARatesAllowed: rates.Value, CashPaymentLimit: limit.Value, Provisional: limit.Provisional || rates.Provisional,
	}, nil
}
