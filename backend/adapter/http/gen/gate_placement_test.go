package gen_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// These tests pin the ADR-0001 note that P0-15 relies on: where a flag gate can
// sit so a disabled module answers 404 before the generated wrapper binds path or
// query parameters or the strict handler decodes the body. A 400 there would
// reveal that the disabled module exists.

const (
	disabledRoute   = "DELETE /v1/devices/{deviceId}" // uuid path parameter
	malformedPath   = "/v1/devices/not-a-uuid"
	disabledBodyURL = "/v1/devices" // POST, JSON body
)

// gatingMux is the placement ADR-0001 prescribes: it wraps every handler at
// registration time, keyed by the pattern the generated server registers, so
// the gate runs before the wrapper touches the request. An unknown pattern is a
// startup failure (fail closed), never an ungated route.
type gatingMux struct {
	*http.ServeMux
	enabled func(module string) bool
}

func (m gatingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	module, ok := gen.RouteModule(pattern)
	if !ok {
		panic(fmt.Sprintf("route %q has no x-module", pattern))
	}
	m.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if module != gen.ModulePlatform && !m.enabled(module) {
			http.NotFound(w, r)
			return
		}
		handler(w, r)
	})
}

func allModulesDisabled(string) bool { return false }

func serve(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, strings.NewReader(body)))
	return rec
}

func TestRegistrationGateAnswers404BeforeBindingOrDecoding(t *testing.T) {
	module, ok := gen.RouteModule(disabledRoute)
	require.True(t, ok)
	require.NotEqual(t, gen.ModulePlatform, module, "the probe route must belong to a flagged module")

	h := gen.HandlerFromMux(gen.NewStrictHandler(nil, nil),
		gatingMux{ServeMux: http.NewServeMux(), enabled: allModulesDisabled})

	assert.Equal(t, http.StatusNotFound, serve(h, http.MethodDelete, malformedPath, "").Code,
		"malformed path parameter on a disabled module")
	assert.Equal(t, http.StatusNotFound, serve(h, http.MethodPost, disabledBodyURL, "{not json").Code,
		"malformed body on a disabled module")
}

func TestRegistrationGateFailsClosedOnUnknownPattern(t *testing.T) {
	m := gatingMux{ServeMux: http.NewServeMux(), enabled: allModulesDisabled}
	assert.Panics(t, func() {
		gen.HandlerFromMuxWithBaseURL(gen.NewStrictHandler(nil, nil), m, "/api")
	}, "a BaseURL changes the patterns; RouteModule must not silently miss them")
}

// The placements below look natural but leak. They are pinned so a generator
// upgrade that changes the ordering is noticed.
func TestLeakyGatePlacements(t *testing.T) {
	gateOnPattern := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if module, ok := gen.RouteModule(r.Pattern); !ok || module != gen.ModulePlatform {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	t.Run("StdHTTPServerOptions.Middlewares run after path binding", func(t *testing.T) {
		h := gen.HandlerWithOptions(gen.NewStrictHandler(nil, nil), gen.StdHTTPServerOptions{
			Middlewares: []gen.MiddlewareFunc{gateOnPattern},
		})
		assert.Equal(t, http.StatusBadRequest, serve(h, http.MethodDelete, malformedPath, "").Code)
	})

	t.Run("a wrapper outside the mux sees no r.Pattern yet", func(t *testing.T) {
		var seen string
		inner := gen.Handler(gen.NewStrictHandler(nil, nil))
		outer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seen = r.Pattern
			inner.ServeHTTP(w, r)
		})
		serve(outer, http.MethodDelete, malformedPath, "")
		assert.Empty(t, seen)
	})

	t.Run("ServeMux answers 405 with Allow for a disabled route's path", func(t *testing.T) {
		h := gen.HandlerFromMux(gen.NewStrictHandler(nil, nil),
			gatingMux{ServeMux: http.NewServeMux(), enabled: allModulesDisabled})
		rec := serve(h, http.MethodPut, "/v1/devices/x", "")
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code,
			"P0-15 must turn unmatched-method responses into 404 for gated paths")
		assert.Contains(t, rec.Header().Get("Allow"), http.MethodDelete)
	})
}
