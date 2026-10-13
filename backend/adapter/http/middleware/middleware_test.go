package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

var generatedID = regexp.MustCompile(`^req_[A-Z2-7]{26}$`)

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) httperr.Response {
	t.Helper()
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var env httperr.Response
	dec := json.NewDecoder(rec.Body)
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&env))
	return env
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name     string
		inbound  string
		wantSame bool
	}{
		{name: "generated when absent"},
		{name: "kept when well formed", inbound: "req_01J9Z6Q4X8", wantSame: true},
		{name: "kept with dots and dashes", inbound: "a.b-c_d", wantSame: true},
		{name: "replaced when it could inject log lines", inbound: "x\nlevel=error"},
		{name: "replaced when too long", inbound: strings.Repeat("a", 65)},
		{name: "replaced when it has spaces", inbound: "a b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen string
			h := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				seen = requestid.From(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			if tt.inbound != "" {
				req.Header.Set(middleware.RequestIDHeader, tt.inbound)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, seen, rec.Header().Get(middleware.RequestIDHeader), "context and response agree")
			if tt.wantSame {
				assert.Equal(t, tt.inbound, seen)
			} else {
				assert.Regexp(t, generatedID, seen)
			}
		})
	}
}

func TestRequestID_isUnique(t *testing.T) {
	ids := make(map[string]struct{})
	h := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ids[requestid.From(r.Context())] = struct{}{}
	}))
	for range 100 {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Len(t, ids, 100)
}

const secretInPanic = "AAAA010101AAA ana@example.com"

func TestRecover_writesTheInternalErrorEnvelope(t *testing.T) {
	var logs bytes.Buffer
	logger := logging.New(&logs, nil)
	h := middleware.RequestID(middleware.Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(secretInPanic)
	})))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	body := rec.Body.String()
	assert.NotContains(t, body, "AAAA010101AAA", "the panic value never reaches the client")
	assert.NotContains(t, body, "goroutine", "no stack trace in the response")
	env := decodeEnvelope(t, rec)
	assert.Equal(t, apperr.CodeInternalError, env.Error.Code)
	assert.Equal(t, apperr.InternalMessage, env.Error.Message)
	assert.Equal(t, rec.Header().Get(middleware.RequestIDHeader), env.Error.RequestID)

	assert.Contains(t, logs.String(), `"msg":"panic recovered"`)
	assert.Contains(t, logs.String(), env.Error.RequestID, "the log line correlates with the response")
	assert.NotContains(t, logs.String(), "AAAA010101AAA", "the logger redacts PII in the panic value")
	assert.NotContains(t, logs.String(), "ana@example.com")
}

func TestRecover_afterHeadersAreSentDoesNotWriteTwice(t *testing.T) {
	h := middleware.Recover(logging.New(&bytes.Buffer{}, nil))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("partial"))
		panic("late")
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "partial", rec.Body.String(), "no envelope appended to a started body")
}

func TestRecover_rethrowsErrAbortHandler(t *testing.T) {
	h := middleware.Recover(logging.New(&bytes.Buffer{}, nil))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestRecover_keepsResponseControllerFeatures(t *testing.T) {
	var flushErr error
	h := middleware.Recover(logging.New(&bytes.Buffer{}, nil))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flushErr = http.NewResponseController(w).Flush()
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.NoError(t, flushErr, "the wrapper unwraps to the real writer")
}

type stubLimiter struct {
	allow      bool
	retryAfter time.Duration
	keys       []string
}

func (s *stubLimiter) Allow(key string) (bool, time.Duration) {
	s.keys = append(s.keys, key)
	return s.allow, s.retryAfter
}

func TestRateLimit(t *testing.T) {
	tests := []struct {
		name           string
		allow          bool
		retryAfter     time.Duration
		wantStatus     int
		wantRetryAfter string
	}{
		{name: "allowed passes through", allow: true, wantStatus: http.StatusNoContent},
		{name: "denied rounds Retry-After up", retryAfter: 1500 * time.Millisecond, wantStatus: http.StatusTooManyRequests, wantRetryAfter: "2"},
		{name: "sub-second wait is still at least 1", retryAfter: time.Millisecond, wantStatus: http.StatusTooManyRequests, wantRetryAfter: "1"},
		{name: "whole seconds are exact", retryAfter: 3 * time.Second, wantStatus: http.StatusTooManyRequests, wantRetryAfter: "3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &stubLimiter{allow: tt.allow, retryAfter: tt.retryAfter}
			reached := false
			h := middleware.RequestID(middleware.RateLimit(limiter, middleware.ClientIP(0))(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					reached = true
					w.WriteHeader(http.StatusNoContent)
				})))
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.RemoteAddr = "203.0.113.7:4242"
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, []string{"203.0.113.7"}, limiter.keys)
			assert.Equal(t, tt.allow, reached)
			if tt.allow {
				return
			}
			assert.Equal(t, tt.wantRetryAfter, rec.Header().Get("Retry-After"))
			env := decodeEnvelope(t, rec)
			assert.Equal(t, apperr.CodeRateLimited, env.Error.Code)
			assert.NotEmpty(t, env.Error.RequestID)
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		hops       int
		remoteAddr string
		xff        []string
		want       string
	}{
		{name: "peer address without proxies", remoteAddr: "198.51.100.1:5000", want: "198.51.100.1"},
		{name: "X-Forwarded-For ignored without trusted proxies", remoteAddr: "198.51.100.1:5000", xff: []string{"1.2.3.4"}, want: "198.51.100.1"},
		{name: "one trusted hop takes the last entry", hops: 1, remoteAddr: "10.0.0.1:80", xff: []string{"6.6.6.6, 203.0.113.9"}, want: "203.0.113.9"},
		{name: "spoofed leading entries are skipped", hops: 2, remoteAddr: "10.0.0.1:80", xff: []string{"6.6.6.6, 203.0.113.9, 10.1.1.1"}, want: "203.0.113.9"},
		{name: "repeated headers are one list", hops: 1, remoteAddr: "10.0.0.1:80", xff: []string{"6.6.6.6", "203.0.113.9"}, want: "203.0.113.9"},
		{name: "fewer entries than hops falls back to the peer", hops: 2, remoteAddr: "10.0.0.1:80", xff: []string{"203.0.113.9"}, want: "10.0.0.1"},
		{name: "unparsable entry falls back to the peer", hops: 1, remoteAddr: "10.0.0.1:80", xff: []string{"not-an-ip"}, want: "10.0.0.1"},
		{name: "IPv6 is keyed by its /64", remoteAddr: "[2001:db8:1:2:aaaa::1]:443", want: "2001:db8:1:2::/64"},
		{name: "IPv4-mapped IPv6 is plain IPv4", remoteAddr: "[::ffff:198.51.100.1]:443", want: "198.51.100.1"},
		{name: "unparsable peer is kept verbatim", remoteAddr: "pipe", want: "pipe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for _, v := range tt.xff {
				req.Header.Add("X-Forwarded-For", v)
			}
			assert.Equal(t, tt.want, middleware.ClientIP(tt.hops)(req))
		})
	}
}

func TestChain_runsOutermostFirst(t *testing.T) {
	var order []string
	mw := func(name string) middleware.Func {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := middleware.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	}), mw("a"), mw("b"))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, []string{"a", "b", "handler"}, order)
}
