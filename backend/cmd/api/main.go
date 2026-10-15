// Command api is the composition root of the HTTP API (ADR-0001): it loads
// the configuration, wires adapters into the generated strict server, builds
// the middleware chain and serves until SIGINT/SIGTERM, then drains in-flight
// requests.
//
// Middleware order (docs/architecture/c4.md): request id → recover → rate
// limit → request deadline → routing; then, per operation, flag gate (or the
// not-built 404) → authentication → contract validation → handler
// (adapter/http/router).
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	"github.com/romodevx/control-contable/backend/adapter/http/flagshttp"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// version is reported by /healthz; release builds set it with
// -ldflags "-X main.version=<semver>".
var version = "0.0.0-dev"

// http.Server limits. WriteTimeout bounds the slowest handler; shutdown
// fits inside Cloud Run's 10 s grace period after SIGTERM.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 64 << 10
	shutdownTimeout   = 8 * time.Second
	// roleCheckTimeout bounds the startup check of the runtime database role.
	roleCheckTimeout = 10 * time.Second
)

func main() {
	if err := runUntilSignal(); err != nil {
		fmt.Fprintln(os.Stderr, "api:", err)
		os.Exit(1)
	}
}

func runUntilSignal() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return run(ctx, os.LookupEnv, os.Stdout)
}

// run loads the configuration and serves until ctx is done.
func run(ctx context.Context, lookup config.Lookup, logOut io.Writer) error {
	secrets := config.NewEnvSecrets(lookup)
	cfg, err := config.Load(ctx, lookup, secrets)
	authKeys, authErr := config.LoadAuth(ctx, lookup, secrets)
	if err := errors.Join(err, authErr); err != nil {
		return err
	}
	outboxKeys, err := config.LoadAuthOutbox(ctx, lookup, secrets, authKeys)
	if err != nil {
		return err
	}
	taxpayerKeys, err := config.LoadTaxpayers(ctx, lookup, secrets, cfg.Env, authKeys)
	if err != nil {
		return err
	}
	reminderKeys, err := config.LoadReminders(ctx, lookup, secrets, cfg.Env, authKeys, taxpayerKeys)
	if err != nil {
		return err
	}
	logger := logging.New(logOut, cfg.LogLevel)
	// The runtime pool (cc_app). Opening does not connect: once started, an
	// unreachable database makes every flagged module OFF (fail closed) and
	// /healthz "degraded", not a crash loop.
	pool, err := openDatabase(cfg)
	if err != nil {
		return err
	}
	defer closeDatabase(pool, logger)
	if err := guardRuntimeRole(ctx, cfg.Env, pool); err != nil {
		return err
	}
	flagSvc, err := newFlagService(pool, logger)
	if err != nil {
		return err
	}
	auth, err := newAuthModule(authWiring{keys: authKeys, outbox: outboxKeys, pool: pool, logger: logger, now: time.Now})
	if err != nil {
		return err
	}
	records, err := newRecordModules(recordWiring{env: cfg.Env, keys: taxpayerKeys, pool: pool, auth: auth, logger: logger})
	if err != nil {
		return err
	}
	holidays, err := loadHolidaySnapshots(ctx, pool)
	if err != nil {
		return err
	}
	taxEngine, err := newTaxEngineModule(taxEngineWiring{
		pool: pool, limits: auth.limits, logger: logger,
		tx: records.taxpayers.tx, taxpayers: records.taxpayers.queries, holidays: holidays.taxEngine,
	})
	if err != nil {
		return err
	}
	dashboard, err := newDashboardModule(dashboardWiring{
		tx: records.taxpayers.tx, flags: flagSvc, taxpayers: records.taxpayers.queries, estimator: taxEngine.estimator,
		limits: auth.limits,
	})
	if err != nil {
		return err
	}
	history, err := newHistoryModule(historyWiring{
		tx: records.taxpayers.tx, flags: flagSvc, estimator: taxEngine.estimator, limits: auth.limits,
	})
	if err != nil {
		return err
	}
	annual, err := newAnnualModule(annualWiring{
		tx: records.taxpayers.tx, taxpayers: records.taxpayers.queries, estimator: taxEngine.estimator,
		holidays: taxEngine.holidays, limits: auth.limits, logger: logger,
	})
	if err != nil {
		return err
	}
	reminders, err := newRemindersModule(remindersWiring{
		env: cfg.Env, keys: reminderKeys, tx: records.taxpayers.tx, holidays: holidays.reminders, limits: auth.limits, logger: logger,
	})
	if err != nil {
		return err
	}
	handler, err := newHandler(cfg, deps{
		logger:    logger,
		db:        pgdb.NewPinger(pool),
		flags:     flagSvc,
		auth:      auth,
		taxpayers: records.taxpayers,
		incomes:   records.incomes,
		expenses:  records.expenses,
		taxEngine: taxEngine,
		dashboard: dashboard,
		history:   history,
		annual:    annual,
		reminders: reminders,
		now:       time.Now,
	})
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.HTTPAddr, err)
	}
	logger.InfoContext(ctx, "api listening", slog.String("addr", ln.Addr().String()),
		slog.String("env", string(cfg.Env)), slog.String("version", version))
	return serve(ctx, newServer(handler, cfg.HTTPAddr, logger), ln, logger)
}

// recordWiring is what the record modules (taxpayers, incomes, expenses)
// need from the composition root.
type recordWiring struct {
	env    config.Environment
	keys   config.Taxpayers
	pool   *sql.DB // cc_app
	auth   authModule
	logger *slog.Logger
	// incomesClock and expensesClock are the record rules' clocks; nil is
	// the system clock (tests pin a day).
	incomesClock  clock.Clock
	expensesClock clock.Clock
}

// recordModules are the taxpayers module and the record modules over its
// unit of work.
type recordModules struct {
	taxpayers taxpayersModule
	incomes   incomesModule
	expenses  expensesModule
}

// newRecordModules wires the incomes and expenses modules over the
// taxpayers module's cc_app unit of work and one rate catalog: their
// regime checks join it, and the taxpayers record lookup reads the
// expenses inside the taxpayers' own.
func newRecordModules(w recordWiring) (recordModules, error) {
	taxpayers, err := newTaxpayersModule(taxpayersWiring{env: w.env, keys: w.keys, pool: w.pool, limits: w.auth.limits, logger: w.logger})
	if err != nil {
		return recordModules{}, err
	}
	rates, err := newRateCatalog()
	if err != nil {
		return recordModules{}, err
	}
	incomes, err := newIncomesModule(incomesWiring{
		tx: taxpayers.tx, taxpayers: taxpayers.queries, rates: rates,
		limits: w.auth.limits, clock: w.incomesClock, logger: w.logger,
	})
	if err != nil {
		return recordModules{}, err
	}
	expenses, err := newExpensesModule(expensesWiring{
		tx: taxpayers.tx, taxpayers: taxpayers.queries, rates: rates,
		limits: w.auth.limits, clock: w.expensesClock, logger: w.logger,
	})
	if err != nil {
		return recordModules{}, err
	}
	return recordModules{taxpayers: taxpayers, incomes: incomes, expenses: expenses}, nil
}

// openDatabase opens the cc_app pool (pgx) without connecting.
func openDatabase(cfg config.Config) (*sql.DB, error) {
	pool, err := pgdb.Open(cfg.DatabaseURL.Reveal())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return pool, nil
}

// guardRuntimeRole refuses to serve, outside dev, unless the pool is logged in
// as a plain runtime role that RLS binds (never cc_owner, a superuser or a
// BYPASSRLS role). It needs the database: outside dev, a database that is down
// at boot keeps this instance from starting (fail closed) rather than serving
// with an unverified role. Dev is not checked, so the local stack can start
// before its database.
func guardRuntimeRole(ctx context.Context, env config.Environment, pool *sql.DB) error {
	if env == config.EnvDev {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, roleCheckTimeout)
	defer cancel()
	if err := pgdb.CheckRuntimeRole(ctx, pool); err != nil {
		return fmt.Errorf("refusing to start: runtime database role: %w", err)
	}
	return nil
}

func closeDatabase(pool *sql.DB, logger *slog.Logger) {
	if err := pool.Close(); err != nil {
		logger.Warn("close database pool", slog.Any("error", err))
	}
}

// deps are the adapters newHandler wires; tests substitute them.
type deps struct {
	logger    *slog.Logger
	db        health.Pinger
	flags     flagService
	auth      authModule
	taxpayers taxpayersModule
	incomes   incomesModule
	expenses  expensesModule
	taxEngine taxEngineModule
	dashboard dashboardModule
	history   historyModule
	annual    annualModule
	reminders remindersModule
	now       func() time.Time
}

// flagService is what the HTTP layer needs from the flags service: the
// router's gate and the GET /v1/flags source. *flags.Service implements it.
type flagService interface {
	router.FlagGate
	flagshttp.Lister
}

// newFlagService builds the cached, fail-closed flags service over db (cc_app).
func newFlagService(db *sql.DB, logger *slog.Logger) (*flags.Service, error) {
	g, err := flagsgorm.Open(db)
	if err != nil {
		return nil, err
	}
	return flags.NewService(flagsgorm.NewReader(g), flags.WithLogger(logger)), nil
}

// newHandler builds the full handler: the edge middlewares around the
// generated server with its per-operation pipeline.
func newHandler(cfg config.Config, d deps) (http.Handler, error) {
	spec, err := gen.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("load embedded contract: %w", err)
	}
	server := api{
		Handler:          health.New(d.db, version, d.logger),
		flagsHandler:     flagshttp.New(d.flags, d.logger),
		authHandler:      d.auth.handler,
		accountHandler:   d.auth.account,
		SessionHandler:   d.auth.sessions,
		taxpayersHandler: d.taxpayers.handler,
		incomesHandler:   d.incomes.handler,
		expensesHandler:  d.expenses.handler,
		taxEngineHandler: d.taxEngine.handler,
		dashboardHandler: d.dashboard.handler,
		historyHandler:   d.history.handler,
		annualHandler:    d.annual.handler,
		remindersHandler: d.reminders.handler,
	}
	if d.auth.authn == nil {
		return nil, errors.New("api: no authenticator: secured operations would run unauthenticated")
	}
	if d.taxpayers.handler == nil {
		return nil, errors.New("api: no taxpayers handler: module_taxpayers operations would panic")
	}
	if d.expenses.handler == nil {
		return nil, errors.New("api: no expenses handler: module_expenses operations would panic")
	}
	if d.taxEngine.handler == nil {
		return nil, errors.New("api: no tax engine handler: module_tax_engine operations would panic")
	}
	if d.incomes.handler == nil {
		return nil, errors.New("api: no incomes handler: module_incomes operations would panic")
	}
	if d.dashboard.handler == nil {
		return nil, errors.New("api: no dashboard handler: module_dashboard operations would panic")
	}
	if d.history.handler == nil {
		return nil, errors.New("api: no history handler: module_history operations would panic")
	}
	if d.annual.handler == nil {
		return nil, errors.New("api: no annual handler: module_annual operations would panic")
	}
	if d.reminders.handler == nil {
		return nil, errors.New("api: no reminders handler: module_reminders operations would panic")
	}
	userLimits, err := mergeLimits(d.taxpayers.userLimits, d.incomes.userLimits, d.expenses.userLimits, d.taxEngine.userLimits,
		d.dashboard.userLimits, d.history.userLimits, d.annual.userLimits, d.reminders.userLimits)
	if err != nil {
		return nil, err
	}
	routes, err := router.New(server, spec, router.Options{
		// The flag service answers the effective state: a module whose
		// prerequisite is OFF is OFF (flags.Prerequisites).
		Flags: d.flags,
		// One authenticator per contract security scheme: the bearer EdDSA
		// access tokens (M1-07). The RevenueCat webhook secret joins when M7
		// builds receiveRevenueCatWebhook; until then it is not built and
		// answered before authentication.
		Authenticators: map[string]router.Authenticator{router.SchemeBearer: d.auth.authn},
		Logger:         d.logger,
		NotBuilt:       notBuiltOperations(),
		// Per-client limits of each auth operation, after the flag gate
		// (MODULE-auth §5), keyed like the edge limiter.
		OperationLimits: d.auth.clientLimits,
		ClientKey:       middleware.ClientIP(cfg.TrustedProxyHops),
		// The per-user write budgets (MODULE-taxpayers S-10, incomes M3-06,
		// expenses M4-07, MODULE-tax-engine §5), keyed by the authenticated
		// user.
		UserLimits: userLimits,
	})
	if err != nil {
		return nil, err
	}
	limiter := ratelimit.New(cfg.RateLimit.PerSecond, cfg.RateLimit.Burst, d.now)
	return middleware.Chain(routes,
		middleware.RequestID,
		// Cache-Control: no-store on every response, before anything that can
		// answer (recover, limiter, router), so errors carry it too (S-11).
		middleware.NoStore,
		middleware.Recover(d.logger),
		middleware.RateLimit(limiter, middleware.ClientIP(cfg.TrustedProxyHops)),
		// The client network keys the per-(email, network) login budget
		// (ADR-0005), resolved behind the same trusted proxies.
		middleware.WithClientNetwork(middleware.ClientNetwork(cfg.TrustedProxyHops)),
		middleware.Deadline(cfg.RequestTimeout),
	), nil
}

func newServer(handler http.Handler, addr string, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
}

// serve runs srv on ln until ctx is done, then stops accepting connections
// and waits up to shutdownTimeout for in-flight requests.
func serve(ctx context.Context, srv *http.Server, ln net.Listener, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	logger.Info("api shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
