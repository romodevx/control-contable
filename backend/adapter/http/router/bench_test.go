package router_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// BenchmarkValidatedRequest measures the per-operation pipeline for a POST
// with a JSON body: gate, contract validation, decoding and the handler.
func BenchmarkValidatedRequest(b *testing.B) {
	h := newHarness(b, setup{flags: flagSet{"module_auth": true}})
	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(validLogin))
		req.Header.Set("Content-Type", "application/json")
		h.handler.ServeHTTP(httptest.NewRecorder(), req)
	}
}
