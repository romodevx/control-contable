//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// platformModule is the x-module of the always-on operations.
const platformModule = "platform"

// contractClientAddr is the documentation-range client of every request.
const contractClientAddr = "203.0.113.20:1234"

// gatedOperation is one flagged operation of the contract, as a request.
type gatedOperation struct {
	id, module, method, path string
}

// contractEnv is the real composition root over Postgres: the owner toggles
// flags as cmd/flags does and the API reads them as cc_app.
type contractEnv struct {
	handler http.Handler
	store   flagsgorm.Store
	flags   *flags.Service
	logs    *syncBuffer // the API's log output (the dev email adapter writes codes here)
	owner   *sql.DB     // cc_owner on the same database, for fixtures
	app     *sql.DB     // cc_app, the API's own pool, for reading rows as a user
}

// contractWiring is what a contractOption may adjust: the record modules'
// clocks, the estimate's clock and the API's cc_app pool.
type contractWiring struct {
	recordWiring
	taxEngineClock clock.Clock
	// wrapApp, when set, replaces the cc_app pool every module uses (the
	// performance test traces its statements).
	wrapApp func(t *testing.T, app *sql.DB) *sql.DB
}

// contractOption adjusts the composition of a contractEnv.
type contractOption func(*contractWiring)

// withExpensesDay pins the expense rules' clock to day (the auth, trial and
// taxpayer clocks stay the system's).
func withExpensesDay(day time.Time) contractOption {
	return func(w *contractWiring) { w.expensesClock = clock.NewFixed(day) }
}

// withIncomesDay pins the income rules' clock to day (the auth, trial and
// taxpayer clocks stay the system's).
func withIncomesDay(day time.Time) contractOption {
	return func(w *contractWiring) { w.incomesClock = clock.NewFixed(day) }
}

// withTaxEngineDay pins the estimate's "today" (the default period, the
// future-month refusal) to day.
func withTaxEngineDay(day time.Time) contractOption {
	return func(w *contractWiring) { w.taxEngineClock = clock.NewFixed(day) }
}

func newContractEnv(t *testing.T, db health.Pinger, options ...contractOption) contractEnv {
	t.Helper()
	pg := pgtest.NewDatabase(t)
	ownerDB := pg.Connect(t, pgtest.RoleOwner)
	owner, err := flagsgorm.Open(ownerDB)
	require.NoError(t, err)
	var wiring contractWiring
	for _, o := range options {
		o(&wiring)
	}
	app := pg.Connect(t, pgtest.RoleApp)
	if wiring.wrapApp != nil {
		app = wiring.wrapApp(t, app)
	}
	svc, err := newFlagService(app, nil)
	require.NoError(t, err)
	logs := &syncBuffer{}
	logger := logging.New(logs, nil)
	auth, err := newAuthModule(authWiring{
		env: config.EnvDev, keys: testAuthKeys(), pool: app, logger: logger, now: time.Now, argon2: fastArgon2,
	})
	require.NoError(t, err)
	wiring.env, wiring.keys, wiring.pool, wiring.auth, wiring.logger = config.EnvDev, testTaxpayerKeys(), app, auth, logger
	records, err := newRecordModules(wiring.recordWiring)
	require.NoError(t, err)
	taxEngine, err := newTaxEngineModule(taxEngineWiring{
		pool: app, limits: auth.limits, logger: logger,
		tx: records.taxpayers.tx, taxpayers: records.taxpayers.queries, clock: wiring.taxEngineClock,
	})
	require.NoError(t, err)
	return contractEnv{
		handler: newHandlerWithAll(t, contractConfig(), db, svc, auth, records, taxEngine, logger),
		store:   flagsgorm.NewStore(owner),
		flags:   svc,
		logs:    logs,
		owner:   ownerDB,
		app:     app,
	}
}

// contractConfig leaves the rate limit out of the way of a full contract sweep.
func contractConfig() config.Config {
	cfg := testConfig()
	cfg.RateLimit.Burst = 1 << 20
	return cfg
}

// apply makes the flags in Postgres exactly f (every other module OFF).
func (e contractEnv) apply(t *testing.T, modules []string, f flagtest.Static) {
	t.Helper()
	for _, m := range modules {
		_, err := e.store.SetEnabled(t.Context(), m, f[m])
		require.NoError(t, err, m)
	}
	e.flags.Invalidate() // another process would wait for flags.DefaultTTL
}

func (e contractEnv) request(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.RemoteAddr = contractClientAddr
	return req
}

// TestContract_platformOperations: GET /healthz and GET /v1/flags conform to
// the contract with every module's flag ON and OFF, and the flagged
// operations of an OFF module answer the contract's 404 envelope.
func TestContract_platformOperations(t *testing.T) {
	c := contracttest.Load(t)
	ops := gatedOperations(t)
	env := newContractEnv(t, pingOK{})
	modules := registeredModules(t, env, modulesOf(ops))

	for _, module := range modules {
		flagtest.OnOff(t, module, func(t *testing.T, f flagtest.Static, on bool) {
			env.apply(t, modules, f)

			rec := c.ServeOp(t, contracttest.Op("getHealth"), env.handler, env.request(http.MethodGet, "/healthz"))
			assert.Equal(t, http.StatusOK, rec.Code)

			rec = c.ServeOp(t, contracttest.Op("listEnabledFlags"), env.handler, env.request(http.MethodGet, "/v1/flags"))
			require.Equal(t, http.StatusOK, rec.Code)
			assert.ElementsMatch(t, flags.EffectiveModules(flags.Prerequisites(), enabledKeys(f)), decodeEnabled(t, rec),
				"GET /v1/flags lists exactly the effectively ON modules")

			for _, op := range ops {
				assertGate(t, c, env, op, effectiveOn(f, op.module))
			}
		})
	}
}

// assertGate checks one flagged operation: OFF is the contract's 404
// envelope, and so is ON for an operation not built yet (P0-29 FU2); a built
// operation that is ON is routed past the gate (the module's own contract
// tests take it from there).
func assertGate(t *testing.T, c *contracttest.Contract, env contractEnv, op gatedOperation, on bool) {
	t.Helper()
	req := env.request(op.method, op.path)
	if on && slices.Contains(builtOperations, op.id) {
		rec := httptest.NewRecorder()
		env.handler.ServeHTTP(rec, req)
		assert.NotEqualf(t, http.StatusNotFound, rec.Code, "%s (%s) is ON: routed past the gate", op.id, op.module)
		return
	}
	rec := c.Serve(t, env.handler, req)
	if !flagtest.CheckHidden(t, rec) {
		t.Errorf("%s %s (%s, %s on=%v) is not hidden", op.method, op.path, op.id, op.module, on)
	}
}

// effectiveOn is whether module's routes are served under f: its flag and
// every prerequisite's are ON (flags.Prerequisites, MODULE-tax-engine §7).
func effectiveOn(f flagtest.Static, module string) bool {
	isOn := func(m string) bool { return f[m] }
	return f[module] && len(flags.MissingPrerequisites(flags.Prerequisites(), module, isOn)) == 0
}

// An unreachable database degrades /healthz without breaking its contract,
// whatever the flags.
func TestContract_healthzDegradedConforms(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, unreachableDatabase(t))
	_, err := env.store.ApplyPreset(t.Context(), "demo_full")
	require.NoError(t, err)
	env.flags.Invalidate()

	rec := c.ServeOp(t, contracttest.Op("getHealth"), env.handler, env.request(http.MethodGet, "/healthz"))

	require.Equal(t, http.StatusOK, rec.Code)
	var got gen.HealthStatus
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, gen.HealthStatusStatusDegraded, got.Status)
	assert.Equal(t, gen.HealthStatusDatabaseUnreachable, got.Database)
}

// samplePathValues fill the contract's path parameters. The gate answers
// before any parameter is parsed, so only the routing shape matters; the
// values are still valid ones, so an ON module reaches its validation.
var samplePathValues = map[string]string{
	"period":     "2026-01",
	"regimeCode": "612",
}

const sampleID = "0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e"

var pathParam = regexp.MustCompile(`\{([^}]+)\}`)

// gatedOperations lists every flagged operation of the served contract.
func gatedOperations(t *testing.T) []gatedOperation {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	var ops []gatedOperation
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			module, _ := gen.OperationModule(op.OperationID)
			if module == platformModule {
				continue
			}
			ops = append(ops, gatedOperation{id: op.OperationID, module: module, method: method, path: fillPath(path)})
		}
	}
	require.NotEmpty(t, ops)
	slices.SortFunc(ops, func(a, b gatedOperation) int { return strings.Compare(a.id, b.id) })
	return ops
}

func fillPath(template string) string {
	return pathParam.ReplaceAllStringFunc(template, func(m string) string {
		if v, ok := samplePathValues[strings.Trim(m, "{}")]; ok {
			return v
		}
		return sampleID
	})
}

// unregisteredModules are contract modules whose flag is not registered yet,
// each with the PLAN task whose migration registers it. Contracts freeze
// before their module is built (spec-first), so a module can be in the
// contract with no feature_flags row. Its flag cannot be set, so it is swept
// as OFF only: in every sweep each of its operations must be hidden. The
// list only shrinks.
var unregisteredModules = map[string]string{
	"module_annual": "ANN-22",
}

// registeredModules returns the contract modules that have a flag row, and
// fails for a module that is neither registered nor in unregisteredModules,
// or that is both.
func registeredModules(t *testing.T, env contractEnv, modules []string) []string {
	t.Helper()
	rows, err := env.store.List(t.Context())
	require.NoError(t, err)
	var registered []string
	for _, module := range modules {
		isRegistered := slices.ContainsFunc(rows, func(f flags.Flag) bool { return f.Key == module })
		task, isPending := unregisteredModules[module]
		switch {
		case isRegistered && isPending:
			t.Errorf("%s is registered: remove it from unregisteredModules (%s)", module, task)
		case !isRegistered && !isPending:
			t.Errorf("%s has no flag row: register it by migration or list it in unregisteredModules with its PLAN task", module)
		case isRegistered:
			registered = append(registered, module)
		}
	}
	for module := range unregisteredModules {
		assert.Containsf(t, modules, module, "unregisteredModules lists %q, which no contract operation declares", module)
	}
	return registered
}

func modulesOf(ops []gatedOperation) []string {
	var modules []string
	for _, op := range ops {
		if !slices.Contains(modules, op.module) {
			modules = append(modules, op.module)
		}
	}
	slices.Sort(modules)
	return modules
}

func decodeEnabled(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	var body gen.EnabledFlags
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Enabled
}
