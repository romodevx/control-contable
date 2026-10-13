package contracttest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

// spyT records what the harness reports instead of failing the real test, so
// the failure messages themselves are under test.
type spyT struct {
	testing.TB
	errors []string
	fatal  bool
}

func (s *spyT) Helper() {}

func (s *spyT) Errorf(format string, args ...any) {
	s.errors = append(s.errors, fmt.Sprintf(format, args...))
}

func (s *spyT) Error(args ...any) { s.errors = append(s.errors, fmt.Sprint(args...)) }

func (s *spyT) Fatalf(format string, args ...any) {
	s.Errorf(format, args...)
	s.FailNow()
}

func (s *spyT) Fatal(args ...any) {
	s.Error(args...)
	s.FailNow()
}

func (s *spyT) FailNow() {
	s.fatal = true
	runtime.Goexit()
}

func (s *spyT) Fail() { s.errors = append(s.errors, "Fail()") }

func (s *spyT) Failed() bool { return len(s.errors) > 0 }

func (s *spyT) report() string { return strings.Join(s.errors, "\n") }

// run calls fn with a spy on its own goroutine, so FailNow (Goexit) ends fn
// and not the calling test.
func run(t *testing.T, fn func(tb testing.TB)) *spyT {
	t.Helper()
	spy := &spyT{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(spy)
	}()
	<-done
	return spy
}

// respond is a handler answering a fixed status, headers and body.
func respond(status int, contentType, body string, headers ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for i := 0; i+1 < len(headers); i += 2 {
			w.Header().Set(headers[i], headers[i+1])
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

const (
	jsonType      = "application/json"
	healthOK      = `{"status":"ok","database":"ok","version":"0.1.0"}`
	notFoundBody  = `{"error":{"code":"not_found","message":"Not found.","request_id":"req_1"}}`
	rateLimitBody = `{"error":{"code":"rate_limited","message":"Too many requests. Try again later.","request_id":"req_1"}}`
)

func getHealth() *http.Request { return httptest.NewRequest(http.MethodGet, "/healthz", nil) }

func TestLoad_readsTheRepositoryContract(t *testing.T) {
	c := contracttest.Load(t)

	assert.Contains(t, c.OperationIDs(), "getHealth")
	assert.Contains(t, c.OperationIDs(), "listEnabledFlags")
	assert.Contains(t, c.OperationIDs(), "getMe")
	assert.Same(t, c, contracttest.Load(t), "the contract is parsed once per test binary")
}

func TestServe_conformingResponsePasses(t *testing.T) {
	c := contracttest.Load(t)

	spy := run(t, func(tb testing.TB) {
		rec := c.ServeOp(tb, contracttest.Op("getHealth"), respond(http.StatusOK, jsonType, healthOK), getHealth())
		assert.Equal(t, http.StatusOK, rec.Code, "the recorder is handed back for the test's own assertions")
		assert.JSONEq(t, healthOK, rec.Body.String())
	})

	assert.Empty(t, spy.report())
}

// Each violation must fail with a report naming the operation, what is wrong
// and where, plus the response itself (a readable diff, not a bare boolean).
func TestServe_reportsEveryKindOfViolation(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    []string
	}{
		{
			name:    "body breaks the schema (enum)",
			handler: respond(http.StatusOK, jsonType, `{"status":"great","database":"ok","version":"0.1.0"}`),
			want:    []string{"getHealth", "GET /healthz", "200", "status", `"great"`},
		},
		{
			name:    "required property missing",
			handler: respond(http.StatusOK, jsonType, `{"status":"ok","database":"ok"}`),
			want:    []string{"getHealth", "version"},
		},
		{
			name:    "undeclared property (strict mode)",
			handler: respond(http.StatusOK, jsonType, `{"status":"ok","database":"ok","version":"1","debug":"leak"}`),
			want:    []string{"getHealth", "debug"},
		},
		{
			name:    "status not declared by the operation",
			handler: respond(http.StatusTeapot, jsonType, notFoundBody),
			want:    []string{"getHealth", "418", "declared: 200, 429, 500, 503"},
		},
		{
			name:    "content type not declared",
			handler: respond(http.StatusOK, "text/plain", "ok"),
			want:    []string{"getHealth", "text/plain"},
		},
		{
			name:    "declared header breaks its schema",
			handler: respond(http.StatusTooManyRequests, jsonType, rateLimitBody, "Retry-After", "0"),
			want:    []string{"getHealth", "429", "Retry-After"},
		},
		{
			name:    "body is not JSON",
			handler: respond(http.StatusOK, jsonType, `{"status":`),
			want:    []string{"getHealth", `{"status":`},
		},
	}
	c := contracttest.Load(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := run(t, func(tb testing.TB) { c.Serve(tb, tt.handler, getHealth()) })

			require.NotEmpty(t, spy.errors, "the violation must fail the test")
			for _, want := range tt.want {
				assert.Contains(t, spy.report(), want)
			}
			assert.Contains(t, spy.report(), "docs/api/openapi.yaml")
		})
	}
}

// Strict mode rejects undeclared headers on a response that declares some
// (429 declares Retry-After), except the edge-wide headers every response
// carries; those are transport, not per-operation contract.
func TestServe_edgeHeadersAreNotOperationContract(t *testing.T) {
	c := contracttest.Load(t)
	h := respond(http.StatusTooManyRequests, jsonType, rateLimitBody,
		"Retry-After", "1", "X-Content-Type-Options", "nosniff", "X-Request-Id", "req_1")

	spy := run(t, func(tb testing.TB) { c.Serve(tb, h, getHealth()) })

	assert.Empty(t, spy.report())
}

func TestServe_undeclaredHeaderFailsInStrictMode(t *testing.T) {
	c := contracttest.Load(t)
	h := respond(http.StatusTooManyRequests, jsonType, rateLimitBody, "Retry-After", "1", "X-Debug-Trace", "on")

	spy := run(t, func(tb testing.TB) { c.Serve(tb, h, getHealth()) })

	assert.Contains(t, spy.report(), "X-Debug-Trace")
}

// Strict mode holds every response to its declared headers, including the
// ones that declare none. libopenapi-validator's default ignore list lets
// Cache-Control, ETag, Vary, Last-Modified (and Set-Cookie, always) through
// undeclared, and each of them changes what a client does.
func TestServe_undeclaredSemanticHeadersFail(t *testing.T) {
	tests := []struct {
		name, header, value string
	}{
		{name: "custom header on a response declaring none", header: "X-Debug-Trace", value: "on"},
		{name: "caching", header: "Cache-Control", value: "max-age=3600"},
		{name: "validator", header: "Etag", value: `"v1"`},
		{name: "cache key", header: "Vary", value: "Cookie"},
		{name: "freshness", header: "Last-Modified", value: "Mon, 12 Oct 2026 00:00:00 GMT"},
		{name: "cookie", header: "Set-Cookie", value: "session=x; HttpOnly"},
	}
	c := contracttest.Load(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := respond(http.StatusOK, jsonType, healthOK, tt.header, tt.value)

			spy := run(t, func(tb testing.TB) { c.Serve(tb, h, getHealth()) })

			require.NotEmptyf(t, spy.errors, "an undeclared %s must fail", tt.header)
			assert.Contains(t, spy.report(), tt.header)
		})
	}
}

func TestServe_pathNotInTheContractFails(t *testing.T) {
	c := contracttest.Load(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/not-a-route", nil)

	spy := run(t, func(tb testing.TB) { c.Serve(tb, respond(http.StatusNotFound, jsonType, notFoundBody), req) })

	require.NotEmpty(t, spy.errors)
	assert.Contains(t, spy.report(), "/v1/not-a-route")
}

func TestServeOp_requestMustRouteToTheNamedOperation(t *testing.T) {
	c := contracttest.Load(t)

	spy := run(t, func(tb testing.TB) {
		c.ServeOp(tb, contracttest.Op("listEnabledFlags"), respond(http.StatusOK, jsonType, healthOK), getHealth())
	})

	require.NotEmpty(t, spy.errors)
	assert.Contains(t, spy.report(), "GET /healthz is getHealth, not listEnabledFlags")
}

func TestServeOp_unknownOperationFails(t *testing.T) {
	c := contracttest.Load(t)

	spy := run(t, func(tb testing.TB) {
		c.ServeOp(tb, contracttest.Op("noSuchOperation"), respond(http.StatusOK, jsonType, healthOK), getHealth())
	})

	require.NotEmpty(t, spy.errors)
	assert.Contains(t, spy.report(), `"noSuchOperation" is not an operationId in docs/api/openapi.yaml`)
}

// A disabled module's 404 envelope validates against the gated operation's
// declared 404 (NotFound) response.
func TestServe_hiddenModuleEnvelopeConforms(t *testing.T) {
	c := contracttest.Load(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)

	spy := run(t, func(tb testing.TB) { c.Serve(tb, respond(http.StatusNotFound, jsonType, notFoundBody), req) })

	assert.Empty(t, spy.report())
}

func TestValidate_checksAnAlreadyRecordedResponse(t *testing.T) {
	c := contracttest.Load(t)
	req := getHealth()
	rec := httptest.NewRecorder()
	respond(http.StatusOK, jsonType, `{"status":"ok"}`).ServeHTTP(rec, req)

	spy := run(t, func(tb testing.TB) { c.Validate(tb, req, rec) })

	require.NotEmpty(t, spy.errors)
	assert.Contains(t, spy.report(), "database")
}
