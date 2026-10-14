package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"maps"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/taxinputs"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// taxEngineWiring is what the tax engine module needs from the composition
// root.
type taxEngineWiring struct {
	pool   *sql.DB // cc_app
	limits *ratelimitpg.Store
	logger *slog.Logger
}

// taxEngineModule is the tax engine's HTTP side: the handler of its built
// operations and its per-user write budget for the router.
type taxEngineModule struct {
	handler    *taxenginehttp.Handler
	userLimits map[string]router.Limiter
}

// newTaxEngineModule wires the tax-inputs use cases (M5-21) over the cc_app
// unit of work. The access policy shares the use cases' TxManager: it runs
// nested in their transaction, and a nested call through another manager
// is refused (tx.ErrForeignTransaction).
func newTaxEngineModule(w taxEngineWiring) (taxEngineModule, error) {
	txm, err := pgdb.NewTxManager(w.pool)
	if err != nil {
		return taxEngineModule{}, err
	}
	inputs := usecase.NewInputs(usecase.InputsDeps{
		Tx:     txm,
		Access: access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Store:  taxinputs.NewStore(), Clock: clock.NewSystem(), Logger: w.logger,
	})
	out := make(map[string]router.Limiter, len(taxenginehttp.UserLimits))
	for op, policy := range taxenginehttp.UserLimits {
		out[op] = w.limits.Bucket(policy)
	}
	return taxEngineModule{handler: taxenginehttp.NewHandler(inputs), userLimits: out}, nil
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
