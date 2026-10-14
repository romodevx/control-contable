package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

var _ gen.StrictServerInterface = api{}

func testConfig() config.Config {
	return config.Config{
		Env:            config.EnvDev,
		HTTPAddr:       "127.0.0.1:0",
		RateLimit:      config.RateLimit{PerSecond: 1, Burst: 3},
		RequestTimeout: 10 * time.Second,
	}
}

// stubAuth is an auth use case that must not be reached by the platform
// tests; contract and integration tests wire the real ones.
type stubAuth struct{}

func (stubAuth) Execute(context.Context, usecase.RegisterInput) error {
	return errors.New("stub: not wired")
}

type stubVerify struct{}

func (stubVerify) Execute(context.Context, usecase.VerifyInput) (usecase.Session, error) {
	return usecase.Session{}, errors.New("stub: not wired")
}

type stubResend struct{}

func (stubResend) Execute(context.Context, string) error { return errors.New("stub: not wired") }

type stubLogin struct{}

func (stubLogin) Execute(context.Context, usecase.LoginInput) (usecase.Session, error) {
	return usecase.Session{}, errors.New("stub: not wired")
}

// stubAuthModule has stub use cases and the real authenticator over the
// fixture keys, so a secured operation is never reachable unauthenticated.
func stubAuthModule(t *testing.T) authModule {
	t.Helper()
	authn, err := newAuthenticator(testAuthKeys(), time.Now, nil)
	require.NoError(t, err)
	return authModule{
		handler: authhttp.NewHandler(authhttp.UseCases{Register: stubAuth{}, Verify: stubVerify{}, Resend: stubResend{}, Login: stubLogin{}}),
		authn:   authn,
	}
}

func newTestHandler(t *testing.T, cfg config.Config, db health.Pinger) http.Handler {
	t.Helper()
	return newFlaggedHandler(t, cfg, db, flagtest.Static{})
}

func newFlaggedHandler(t *testing.T, cfg config.Config, db health.Pinger, f flagService) http.Handler {
	t.Helper()
	return newHandlerWith(t, cfg, db, f, stubAuthModule(t), logging.New(io.Discard, nil))
}

func newHandlerWith(t *testing.T, cfg config.Config, db health.Pinger, f flagService, auth authModule, logger *slog.Logger) http.Handler {
	t.Helper()
	return newHandlerWithModules(t, cfg, db, f, auth, stubTaxpayersModule(), logger)
}

func newHandlerWithModules(t *testing.T, cfg config.Config, db health.Pinger, f flagService, auth authModule,
	taxpayers taxpayersModule, logger *slog.Logger) http.Handler {
	t.Helper()
	return newHandlerWithAll(t, cfg, db, f, auth, taxpayers, stubExpensesModule(), stubTaxEngineModule(), logger)
}

func newHandlerWithAll(t *testing.T, cfg config.Config, db health.Pinger, f flagService, auth authModule,
	taxpayers taxpayersModule, expenses expensesModule, taxEngine taxEngineModule, logger *slog.Logger) http.Handler {
	t.Helper()
	h, err := newHandler(cfg, deps{logger: logger, db: db, flags: f, auth: auth, taxpayers: taxpayers,
		expenses: expenses, taxEngine: taxEngine, now: time.Now})
	require.NoError(t, err)
	return h
}

// testAuthKeys are fixture keys (all one byte each), never real ones.
func testAuthKeys() config.Auth {
	key := func(b byte) config.Secret { return config.NewSecret(strings.Repeat(string(rune(b)), 32)) }
	return config.Auth{
		EmailPepper: key(0x11), EmailKey: key(0x22), EmailKeyID: 1,
		CodeKey: key(0x33), AccessTokenKey: key(0x44), AccessTokenKeyID: "k1",
	}
}

// fastArgon2 keeps password hashing cheap in tests (production: OWASP defaults).
var fastArgon2 = crypto.Argon2idParams{Memory: 8 * 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}

// syncBuffer is a bytes.Buffer safe for the concurrent writes of a logger.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// keysReader is a flags.Reader returning fixed keys or an error.
type keysReader struct {
	keys []string
	err  error
}

func (r keysReader) EnabledKeys(context.Context) ([]string, error) { return r.keys, r.err }

type pingOK struct{}

func (pingOK) Ping(context.Context) error { return nil }

// conforms validates rec against docs/api/openapi.yaml (the contract-test
// harness), as the operation req routes to.
func conforms(t *testing.T, req *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	contracttest.Load(t).Validate(t, req, rec)
}

func get(h http.Handler, target string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return req, rec
}

func TestHealthz(t *testing.T) {
	tests := []struct {
		name string
		db   health.Pinger
		want gen.HealthStatus
	}{
		{name: "database reachable", db: pingOK{},
			want: gen.HealthStatus{Status: gen.HealthStatusStatusOk, Database: gen.HealthStatusDatabaseOk, Version: version}},
		{name: "database unreachable", db: unreachableDatabase(t),
			want: gen.HealthStatus{Status: gen.HealthStatusStatusDegraded, Database: gen.HealthStatusDatabaseUnreachable, Version: version}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, rec := get(newTestHandler(t, testConfig(), tt.db), "/healthz")

			require.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader))
			conforms(t, req, rec)
			var got gen.HealthStatus
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRateLimit_answers429WithRetryAfter(t *testing.T) {
	h := newTestHandler(t, testConfig(), pingOK{})
	for range testConfig().RateLimit.Burst {
		_, rec := get(h, "/healthz")
		require.Equal(t, http.StatusOK, rec.Code)
	}

	req, rec := get(h, "/healthz")

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "1", rec.Header().Get("Retry-After"))
	conforms(t, req, rec)
}

func TestFlaggedModulesAreOffByDefault(t *testing.T) {
	req, rec := get(newTestHandler(t, testConfig(), pingOK{}), "/v1/me")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	conforms(t, req, rec)
}

// The flags service drives the real router's gate: OFF is the 404 envelope,
// ON lets the request through to the module (here to its contract validation).
func TestFlagGate_throughTheRealRouter(t *testing.T) {
	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		svc := flags.NewService(keysReader{keys: enabledKeys(f)})
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/verify-email/resend", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		newFlaggedHandler(t, testConfig(), pingOK{}, svc).ServeHTTP(rec, req)

		if !on {
			flagtest.AssertHidden(t, rec)
			conforms(t, req, rec)
			return
		}
		assert.Equal(t, http.StatusBadRequest, rec.Code, "an enabled module is routed past the gate")
		conforms(t, req, rec)
	})
}

// P0-29 FU2: an operation of an enabled module that is not built yet answers
// exactly like a hidden module: no panic, no 500, nothing about its schema.
func TestNotBuilt_isHiddenEvenWithItsModuleOn(t *testing.T) {
	svc := flags.NewService(keysReader{keys: []string{"module_auth", "module_taxpayers", "module_incomes"}})
	h := newFlaggedHandler(t, testConfig(), pingOK{}, svc)

	const taxpayer = "/v1/taxpayers/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e"
	for _, target := range []string{taxpayer + "/incomes?period=2026-09", taxpayer + "/incomes/0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9f"} {
		req, rec := get(h, target)
		flagtest.AssertHidden(t, rec)
		conforms(t, req, rec)
	}
}

func TestFlagGate_failsClosedWhenFlagsAreUnreadable(t *testing.T) {
	svc := flags.NewService(keysReader{err: errors.New("connection refused")})
	h := newFlaggedHandler(t, testConfig(), pingOK{}, svc)

	req, rec := get(h, "/v1/me")
	flagtest.AssertHidden(t, rec)
	conforms(t, req, rec)

	req, rec = get(h, "/v1/flags")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection refused")
	conforms(t, req, rec)
}

func TestListEnabledFlags_servesTheFlagsService(t *testing.T) {
	svc := flags.NewService(keysReader{keys: []string{"module_taxpayers", "module_auth"}})

	req, rec := get(newFlaggedHandler(t, testConfig(), pingOK{}, svc), "/v1/flags")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":["module_auth","module_taxpayers"]}`, rec.Body.String())
	conforms(t, req, rec)
}

func TestListEnabledFlags_isNeverGated(t *testing.T) {
	req, rec := get(newTestHandler(t, testConfig(), pingOK{}), "/v1/flags") // every module OFF

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":[]}`, rec.Body.String())
	conforms(t, req, rec)
}

func enabledKeys(f flagtest.Static) []string {
	keys := []string{}
	for k, on := range f {
		if on {
			keys = append(keys, k)
		}
	}
	return keys
}

func TestNewServer_hasTimeouts(t *testing.T) {
	srv := newServer(http.NotFoundHandler(), ":0", logging.New(io.Discard, nil))

	assert.Positive(t, srv.ReadHeaderTimeout)
	assert.Positive(t, srv.ReadTimeout)
	assert.Positive(t, srv.WriteTimeout)
	assert.Positive(t, srv.IdleTimeout)
	assert.Positive(t, srv.MaxHeaderBytes)
	assert.NotNil(t, srv.ErrorLog)
}

func TestServe_drainsInFlightRequestsOnShutdown(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := newServer(slow, ln.Addr().String(), logging.New(io.Discard, nil))
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, srv, ln, logging.New(io.Discard, nil)) }()

	status := make(chan int, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err != nil {
			status <- 0
			return
		}
		_ = resp.Body.Close()
		status <- resp.StatusCode
	}()
	<-started
	cancel()
	time.AfterFunc(50*time.Millisecond, func() { close(release) })

	assert.Equal(t, http.StatusNoContent, <-status, "the in-flight request completes")
	assert.NoError(t, <-served, "a requested shutdown is not an error")
	resp, err := http.Get("http://" + ln.Addr().String() + "/")
	if err == nil {
		_ = resp.Body.Close()
	}
	assert.Error(t, err, "the listener is closed")
}

// unreachableDatabase is the real readiness adapter over a pool whose server
// refuses connections.
func unreachableDatabase(t *testing.T) health.Pinger {
	t.Helper()
	pool, err := openDatabase(config.Config{DatabaseURL: config.NewSecret(
		"postgres://cc_app:pw@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1")})
	require.NoError(t, err, "opening the pool must not need the database")
	t.Cleanup(func() { assert.NoError(t, pool.Close()) })
	return pgdb.NewPinger(pool)
}

func TestOpenDatabase_rejectsMalformedURL(t *testing.T) {
	_, err := openDatabase(config.Config{DatabaseURL: config.NewSecret("postgres://u:pw@host:notaport/db")})
	assert.ErrorIs(t, err, pgdb.ErrInvalidDSN)
}

// withAuthKeys adds fixture auth keys (hex, all one byte) to env.
func withAuthKeys(env map[string]string) map[string]string {
	for i, name := range []string{"AUTH_EMAIL_PEPPER", "AUTH_EMAIL_KEY", "AUTH_CODE_KEY", "AUTH_ACCESS_TOKEN_KEY"} {
		env[name] = strings.Repeat(string("0123456789"[i])+"a", 32)
	}
	return env
}

func TestRun_reportsMalformedDatabaseURL(t *testing.T) {
	env := withAuthKeys(map[string]string{"APP_ENV": "dev", "DATABASE_URL": "postgres://u:pw@host:notaport/db"})
	err := run(context.Background(), func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)

	assert.ErrorIs(t, err, pgdb.ErrInvalidDSN)
}

// Outside dev a role that cannot be verified is refused: the process does not start.
func TestRun_refusesToStartWhenTheRuntimeRoleCannotBeVerified(t *testing.T) {
	env := withTaxpayersKeys(withAuthKeys(map[string]string{ //nolint:gosec // G101: fake password of a test fixture; the port is closed.
		"APP_ENV": "staging", "TRUSTED_PROXY_HOPS": "0",
		"DATABASE_URL": "postgres://cc_app:s3cr3t-pw@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
	}))
	err := run(context.Background(), func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "runtime database role")
	assert.NotContains(t, err.Error(), "s3cr3t-pw")
}

func TestRun_reportsConfigErrors(t *testing.T) {
	err := run(context.Background(), func(string) (string, bool) { return "", false }, io.Discard)

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "APP_ENV"), err.Error())
	assert.Contains(t, err.Error(), "AUTH_EMAIL_PEPPER", "the auth keys are reported in the same error")
}
