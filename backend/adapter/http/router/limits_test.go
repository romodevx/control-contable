package router_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

// fakeLimiter admits budget events per key, then answers wait; err fails
// every take.
type fakeLimiter struct {
	mu     sync.Mutex
	budget int
	wait   time.Duration
	err    error
	keys   []string
}

func (l *fakeLimiter) Take(_ context.Context, key string) (bool, time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.keys = append(l.keys, key)
	if l.err != nil {
		return false, 0, l.err
	}
	if l.budget == 0 {
		return false, l.wait, nil
	}
	l.budget--
	return true, 0, nil
}

func (l *fakeLimiter) taken() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.keys...)
}

type limitedHarness struct {
	handler   http.Handler
	calls     []string
	authnSeen []string
	logs      bytes.Buffer
}

func newLimitedHarness(t *testing.T, flags flagSet, limits map[string]router.Limiter) *limitedHarness {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	h := &limitedHarness{}
	var body gen.LoginJSONRequestBody
	h.handler, err = router.New(fakeAPI{calls: &h.calls, loginBody: &body}, spec, router.Options{
		Flags:           flags,
		Authn:           recordingAuthn{deny: true, paths: &h.authnSeen},
		Logger:          logging.New(&h.logs, nil),
		OperationLimits: limits,
		ClientKey:       func(r *http.Request) string { return "client:" + r.RemoteAddr },
	})
	require.NoError(t, err)
	return h
}

func (h *limitedHarness) do(method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.20:1234"
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec
}

func TestOperationLimits_answer429WithRetryAfterOnceTheClientIsOut(t *testing.T) {
	limiter := &fakeLimiter{budget: 2, wait: 90*time.Second + time.Millisecond}
	h := newLimitedHarness(t, flagSet{"module_auth": true}, map[string]router.Limiter{"login": limiter})

	for range 2 {
		assert.Equal(t, http.StatusUnauthorized, h.do(http.MethodPost, "/v1/auth/login", validLogin).Code, "the fake login answers 401")
	}
	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "91", rec.Header().Get("Retry-After"), "whole seconds, rounded up")
	assert.Equal(t, apperr.CodeRateLimited, envelope(t, rec).Code)
	assert.Equal(t, []string{"login", "login"}, h.calls, "the limited request never reaches the handler")
	assert.Equal(t, []string{"client:203.0.113.20:1234"}, limiter.taken()[:1], "keyed by the client")
}

// MODULE-auth §5: limits sit after the flag check. A disabled module answers
// the hidden 404 and spends nothing, so a limit can never reveal it.
func TestOperationLimits_runAfterTheFlagGate(t *testing.T) {
	limiter := &fakeLimiter{}
	h := newLimitedHarness(t, flagSet{}, map[string]router.Limiter{"login": limiter})

	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, limiter.taken(), "a disabled operation spends no budget")
}

// The limit runs before authentication and validation: a flood of malformed
// or anonymous requests is limited too, and costs no argon2 or JWT check.
func TestOperationLimits_runBeforeAuthenticationAndValidation(t *testing.T) {
	tests := []struct {
		name, method, target, body string
		op                         string
	}{
		{name: "schema-invalid body", method: http.MethodPost, target: "/v1/auth/login", body: `{"email":1}`, op: "login"},
		{name: "secured operation without a bearer", method: http.MethodGet, target: "/v1/me", op: "getMe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &fakeLimiter{wait: time.Second}
			h := newLimitedHarness(t, flagSet{"module_auth": true}, map[string]router.Limiter{tt.op: limiter})

			rec := h.do(tt.method, tt.target, tt.body)

			assert.Equal(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, "1", rec.Header().Get("Retry-After"))
			assert.Empty(t, h.authnSeen)
			assert.Empty(t, h.calls)
		})
	}
}

func TestOperationLimits_leaveOtherOperationsAlone(t *testing.T) {
	limiter := &fakeLimiter{}
	h := newLimitedHarness(t, flagSet{"module_auth": true}, map[string]router.Limiter{"register": limiter})

	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, limiter.taken())
}

// A limiter that cannot decide fails closed: the generic 500, the cause
// logged, the handler never run.
func TestOperationLimits_failClosedWhenTheLimiterFails(t *testing.T) {
	limiter := &fakeLimiter{err: errors.New("database down")}
	h := newLimitedHarness(t, flagSet{"module_auth": true}, map[string]router.Limiter{"login": limiter})

	rec := h.do(http.MethodPost, "/v1/auth/login", validLogin)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, apperr.CodeInternalError, envelope(t, rec).Code)
	assert.Empty(t, h.calls)
	assert.Contains(t, h.logs.String(), "database down")
}

func TestNew_rejectsBadOperationLimits(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	base := router.Options{Flags: flagSet{}, Authn: recordingAuthn{paths: new([]string)}, Logger: logging.New(&bytes.Buffer{}, nil)}

	t.Run("an operation the contract lacks", func(t *testing.T) {
		opts := base
		opts.OperationLimits = map[string]router.Limiter{"noSuchOperation": &fakeLimiter{}}
		opts.ClientKey = func(*http.Request) string { return "" }
		_, err := router.New(fakeAPI{calls: new([]string)}, spec, opts)
		assert.ErrorContains(t, err, "noSuchOperation")
	})
	t.Run("limits without a client key", func(t *testing.T) {
		opts := base
		opts.OperationLimits = map[string]router.Limiter{"login": &fakeLimiter{}}
		_, err := router.New(fakeAPI{calls: new([]string)}, spec, opts)
		assert.ErrorContains(t, err, "ClientKey")
	})
}
