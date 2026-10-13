// Package flagtest runs a handler suite with a module's flag ON and OFF
// (testing rule: flag-aware suites). Typical use in a module's HTTP tests:
//
//	flagtest.OnOff(t, "module_incomes", func(t *testing.T, f flagtest.Static, on bool) {
//		h, err := router.New(api, spec, router.Options{Flags: f, Authn: authn, Logger: logger})
//		require.NoError(t, err)
//		rec := httptest.NewRecorder()
//		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/taxpayers/"+id+"/incomes", nil))
//		if !on {
//			flagtest.AssertHidden(t, rec)
//			return
//		}
//		// ... the module's own assertions
//	})
package flagtest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// notFoundCode is the ErrorResponse code of a hidden module (apperr.CodeNotFound).
const notFoundCode = "not_found"

// Static is an in-memory flag set (module -> ON). It satisfies the gate's
// Checker; a module missing from the map is OFF.
type Static map[string]bool

// Enabled reports whether module is ON.
func (s Static) Enabled(_ context.Context, module string) bool { return s[module] }

// EnabledModules returns the ON keys, sorted (the GET /v1/flags source), so a
// Static can stand in for the whole flags service.
func (s Static) EnabledModules(context.Context) ([]string, error) {
	keys := []string{}
	for k, on := range s {
		if on {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys, nil
}

// OnOff runs suite twice as subtests, "<module>=ON" then "<module>=OFF". The
// modules in alsoOn stay ON in both runs (dependencies of the module under
// test); every other module is OFF.
func OnOff(t *testing.T, module string, suite func(t *testing.T, flags Static, on bool), alsoOn ...string) {
	t.Helper()
	for _, on := range []bool{true, false} {
		flags := Static{module: on}
		for _, dep := range alsoOn {
			flags[dep] = true
		}
		t.Run(module+"="+state(on), func(t *testing.T) { suite(t, flags, on) })
	}
}

func state(on bool) string {
	if on {
		return "ON"
	}
	return "OFF"
}

// AssertHidden fails the test now unless rec is the gate's answer for a
// disabled module: 404, the standard ErrorResponse with code not_found, and no
// Allow header.
func AssertHidden(t testing.TB, rec *httptest.ResponseRecorder) {
	t.Helper()
	if !CheckHidden(t, rec) {
		t.FailNow()
	}
}

// CheckHidden is AssertHidden without stopping the test; it reports whether
// every check passed.
func CheckHidden(t assert.TestingT, rec *httptest.ResponseRecorder) bool {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	ok := assert.Equal(t, http.StatusNotFound, rec.Code, "status of a hidden module")
	ok = assert.Empty(t, rec.Header().Get("Allow"), "a hidden module never lists its methods") && ok
	if !assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "body is the ErrorResponse envelope") {
		return false
	}
	return assert.Equal(t, notFoundCode, body.Error.Code, "error code of a hidden module") && ok
}
