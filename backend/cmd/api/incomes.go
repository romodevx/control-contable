package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	expensesdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	incomeshttp "github.com/romodevx/control-contable/backend/internal/incomes/adapter/http"
	incomesdomain "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	taxpayersdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The 625 options the income withholding rules read (regime catalog keys).
const (
	optionActivities         = "activities"
	optionRFCGivenToPlatform = "rfc_given_to_platform"
)

// rateCatalog is the tax engine's RateCatalog adapter, bound once with the
// expenses module's amount key so one value serves both record modules
// (ADR-0007 §6): incomes reads fiscal.RateCatalog and Provisional,
// expenses adds Amount.
type rateCatalog = ratecatalog.Catalog[expensesdomain.AmountKey]

// newRateCatalog builds the catalog over the parameter sets compiled into
// the binary. A set without the IVA rates fails the boot (PR #77 FU).
func newRateCatalog() (*rateCatalog, error) {
	sets, err := params.Embedded()
	if err != nil {
		return nil, err
	}
	return ratecatalog.New[expensesdomain.AmountKey](sets)
}

// errIncomesWiring reports wiring without the shared unit of work, the
// taxpayers queries or the rate catalog.
var errIncomesWiring = errors.New("incomes: the shared unit of work, the taxpayers queries and the rate catalog are required")

// incomesWiring is what the incomes module needs from the composition root.
// tx must be the taxpayers module's manager: CheckRegime joins the income's
// transaction only through it.
type incomesWiring struct {
	tx        tx.Manager
	taxpayers taxpayerQueries
	rates     *rateCatalog
	limits    *ratelimitpg.Store
	// clock is the incomes' today; accessClock decides the trial (its rows
	// start at the database's now()). nil is the system clock, which
	// production uses for both.
	clock       clock.Clock
	accessClock clock.Clock
	logger      *slog.Logger
}

// incomesModule is the incomes module's HTTP side: the handler of its
// operations and its per-user write budget for the router, over its use
// cases.
type incomesModule struct {
	handler    *incomeshttp.Handler
	userLimits map[string]router.Limiter
	incomes    *incomes.Incomes
}

// newIncomesModule wires the incomes use cases (M3-04, M3-05) and their
// handler (M3-06) over the shared cc_app unit of work, the taxpayers
// module, the access policy (trial cap) and the rate catalog.
func newIncomesModule(w incomesWiring) (incomesModule, error) {
	if w.tx == nil || w.taxpayers == nil || w.rates == nil {
		return incomesModule{}, errIncomesWiring
	}
	store := accessgorm.NewStore()
	uc := incomes.NewIncomes(incomes.Deps{
		Tx:           w.tx,
		Store:        incomesinfra.NewStore(),
		Taxpayers:    incomesTaxpayers{queries: w.taxpayers},
		Access:       access.NewPolicy(w.tx, store, store, orSystem(w.accessClock)),
		Rates:        w.rates,
		Withholdings: w.rates.Withholdings(),
		Clock:        orSystem(w.clock),
		Logger:       w.logger,
	})
	return incomesModule{
		handler:    incomeshttp.NewHandler(uc),
		userLimits: bucketsOf(w.limits, incomeshttp.UserLimits),
		incomes:    uc,
	}, nil
}

// orSystem is c, or the system clock when c is nil.
func orSystem(c clock.Clock) clock.Clock {
	if c == nil {
		return clock.NewSystem()
	}
	return c
}

// taxpayerQueries is the taxpayers module's read side (*usecase.Queries).
type taxpayerQueries interface {
	taxpayers.TaxpayerRegimeChecker
	taxpayers.TaxpayerProfileReader
}

// incomesTaxpayers adapts the taxpayers module to the incomes port: the
// modules never import each other (archtest), only the composition root
// sees both.
type incomesTaxpayers struct {
	queries taxpayerQueries
}

// CheckRegime implements incomes.Taxpayers.
func (a incomesTaxpayers) CheckRegime(ctx context.Context, userID, taxpayerID, regimeCode, period string) error {
	return a.queries.CheckRegime(ctx, userID, taxpayerID, regimeCode, period)
}

// Profile implements incomes.Taxpayers.
func (a incomesTaxpayers) Profile(ctx context.Context, userID, taxpayerID string) (incomes.TaxpayerProfile, error) {
	p, err := a.queries.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return incomes.TaxpayerProfile{}, err
	}
	out := incomes.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		held := incomes.RegimeProfile{
			Code:               fiscal.RegimeCode(r.Code),
			StartPeriod:        r.Window.Start.String(),
			PlatformActivities: activities(r.Options[optionActivities]),
			RFCGivenToPlatform: flag(r.Options[optionRFCGivenToPlatform]),
		}
		if !r.Window.IsOpen() {
			held.EndPeriod = r.Window.End.String()
		}
		out.Regimes = append(out.Regimes, held)
	}
	return out, nil
}

func activities(v taxpayersdomain.OptionValue) []incomesdomain.PlatformActivity {
	list, ok := v.List()
	if !ok {
		return nil
	}
	out := make([]incomesdomain.PlatformActivity, len(list))
	for i, a := range list {
		out[i] = incomesdomain.PlatformActivity(a)
	}
	return out
}

func flag(v taxpayersdomain.OptionValue) bool {
	b, _ := v.Bool()
	return b
}
