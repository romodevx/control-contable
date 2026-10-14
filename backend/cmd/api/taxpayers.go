package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	expinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/access/accessgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// errMissingRFCKeys reports RFC keys missing outside dev, where config
// refuses them first; it guards the wiring against a zero key anyway.
var errMissingRFCKeys = errors.New("taxpayers: RFC keys are required outside dev")

// taxpayersWiring is what the taxpayers module needs from the composition
// root.
type taxpayersWiring struct {
	env    config.Environment
	keys   config.Taxpayers
	pool   *sql.DB // cc_app
	limits *ratelimitpg.Store
	logger *slog.Logger
}

// taxpayersModule is the taxpayers module's HTTP side: the handler of its
// operations and its per-user write budget (S-10) for the router.
type taxpayersModule struct {
	handler    *taxpayershttp.Handler
	userLimits map[string]router.Limiter
	// tx is the cc_app unit of work the record modules share: CheckRegime
	// joins their transaction only through the same manager.
	tx tx.Manager
	// queries is the read side M3/M4 consume (TaxpayerRegimeChecker,
	// TaxpayerProfileReader).
	queries *usecase.Queries
}

// newTaxpayersModule wires the taxpayers use cases (M2-05, M2-06) over the
// cc_app unit of work, the RFC vault and the access policy.
func newTaxpayersModule(w taxpayersWiring) (taxpayersModule, error) {
	deps, err := newTaxpayersDeps(w)
	if err != nil {
		return taxpayersModule{}, err
	}
	return taxpayersModule{
		handler: taxpayershttp.NewHandler(taxpayershttp.UseCases{
			Taxpayers: usecase.NewTaxpayers(deps), Regimes: usecase.NewRegimes(deps), Catalog: usecase.NewCatalog(deps),
		}),
		userLimits: bucketsOf(w.limits, taxpayershttp.UserLimits),
		tx:         deps.Tx,
		queries:    usecase.NewQueries(deps),
	}, nil
}

// newTaxpayersDeps builds the taxpayers ports over a new cc_app unit of
// work.
func newTaxpayersDeps(w taxpayersWiring) (usecase.Deps, error) {
	txm, err := pgdb.NewTxManager(w.pool)
	if err != nil {
		return usecase.Deps{}, err
	}
	vault, err := newRFCVault(w)
	if err != nil {
		return usecase.Deps{}, err
	}
	store, catalog := infra.NewStore(), infra.NewCatalog()
	return usecase.Deps{
		Tx:        txm,
		Access:    access.NewPolicy(txm, accessgorm.NewStore(), accessgorm.NewStore(), clock.NewSystem()),
		Taxpayers: store, Regimes: store, Catalog: catalog, Directory: catalog,
		Records: regimeRecords{expenses: expinfra.NewStore()}, // M3-06 adds the incomes
		RFC:     vault, Clock: clock.NewSystem(), Logger: w.logger,
	}, nil
}

// bucketsOf builds the router's per-user limits from a module adapter's
// policies by operationId: the writes of one scope share one bucket per
// user.
func bucketsOf(limits *ratelimitpg.Store, policies map[string]ratelimit.Policy) map[string]router.Limiter {
	out := make(map[string]router.Limiter, len(policies))
	for op, policy := range policies {
		out[op] = limits.Bucket(policy)
	}
	return out
}

// newRFCVault builds the RFC vault (S-1, S-2) from the configured keys. In
// dev a missing key becomes an ephemeral random one, logged as such: RFCs
// sealed with it cannot be opened, nor matched as duplicates, after a
// restart. Outside dev config requires both keys.
func newRFCVault(w taxpayersWiring) (*infra.RFCVault, error) {
	key, err := keyOrEphemeral(w, w.keys.RFCKey, "RFC_ENC_KEY", crypto.KeySize)
	if err != nil {
		return nil, err
	}
	pepper, err := keyOrEphemeral(w, w.keys.RFCIndexPepper, "RFC_INDEX_PEPPER", crypto.MinPepperSize)
	if err != nil {
		return nil, err
	}
	keyID := crypto.KeyID(w.keys.RFCKeyID) // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	keyring, err := crypto.NewKeyring(keyID, map[crypto.KeyID][]byte{keyID: key})
	if err != nil {
		return nil, fmt.Errorf("taxpayers: rfc keyring: %w", err)
	}
	index, err := crypto.NewRFCIndex(pepper)
	if err != nil {
		return nil, fmt.Errorf("taxpayers: rfc index: %w", err)
	}
	return infra.NewRFCVault(keyring, index), nil
}

func keyOrEphemeral(w taxpayersWiring, secret config.Secret, name string, size int) ([]byte, error) {
	if !secret.IsZero() {
		return secret.RevealBytes(), nil
	}
	if w.env != config.EnvDev {
		return nil, fmt.Errorf("%w: %s", errMissingRFCKeys, name)
	}
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("taxpayers: ephemeral %s: %w", name, err)
	}
	w.logger.Warn("taxpayers: secret not set; using an ephemeral key until restart (dev only)",
		slog.String("variable", name))
	return key, nil
}
