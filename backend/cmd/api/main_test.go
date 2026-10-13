package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

var _ gen.StrictServerInterface = api{}

func testConfig() config.Config {
	return config.Config{
		Env:       config.EnvDev,
		HTTPAddr:  "127.0.0.1:0",
		RateLimit: config.RateLimit{PerSecond: 1, Burst: 3},
	}
}

func newTestHandler(t *testing.T, cfg config.Config, db health.Pinger) http.Handler {
	t.Helper()
	return newFlaggedHandler(t, cfg, db, flagtest.Static{})
}

func newFlaggedHandler(t *testing.T, cfg config.Config, db health.Pinger, f flagService) http.Handler {
	t.Helper()
	h, err := newHandler(cfg, deps{logger: logging.New(io.Discard, nil), db: db, flags: f, now: time.Now})
	require.NoError(t, err)
	return h
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
// ON lets the request through to the module (whose handler is not built yet).
func TestFlagGate_throughTheRealRouter(t *testing.T) {
	flagtest.OnOff(t, "module_auth", func(t *testing.T, f flagtest.Static, on bool) {
		svc := flags.NewService(keysReader{keys: enabledKeys(f)})
		req, rec := get(newFlaggedHandler(t, testConfig(), pingOK{}, svc), "/v1/me")

		if !on {
			flagtest.AssertHidden(t, rec)
			conforms(t, req, rec)
			return
		}
		assert.NotEqual(t, http.StatusNotFound, rec.Code, "an enabled module is routed past the gate")
	})
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

func TestRun_reportsMalformedDatabaseURL(t *testing.T) {
	env := map[string]string{"APP_ENV": "dev", "DATABASE_URL": "postgres://u:pw@host:notaport/db"}
	err := run(context.Background(), func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)

	assert.ErrorIs(t, err, pgdb.ErrInvalidDSN)
}

// Outside dev a role that cannot be verified is refused: the process does not start.
func TestRun_refusesToStartWhenTheRuntimeRoleCannotBeVerified(t *testing.T) {
	env := map[string]string{ //nolint:gosec // G101: fake password of a test fixture; the port is closed.
		"APP_ENV": "staging", "TRUSTED_PROXY_HOPS": "0",
		"DATABASE_URL": "postgres://cc_app:s3cr3t-pw@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
	}
	err := run(context.Background(), func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "runtime database role")
	assert.NotContains(t, err.Error(), "s3cr3t-pw")
}

func TestRun_reportsConfigErrors(t *testing.T) {
	err := run(context.Background(), func(string) (string, bool) { return "", false }, io.Discard)

	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "APP_ENV"), err.Error())
}
