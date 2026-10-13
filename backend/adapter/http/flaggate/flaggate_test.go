package flaggate_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/flaggate"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

var _ flaggate.Checker = (*flags.Service)(nil)

const (
	devicesModule  = "module_reminders"
	deleteDevice   = "DELETE /v1/devices/{deviceId}"
	registerDevice = "POST /v1/devices"
	malformedPath  = "/v1/devices/not-a-uuid"
	healthPattern  = "GET /healthz"
	testRequestID  = "req-123"
)

// staticFlags is the gate input: module -> ON.
type staticFlags map[string]bool

func (f staticFlags) Enabled(_ context.Context, module string) bool { return f[module] }

type envelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func serve(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req = req.WithContext(requestid.With(req.Context(), testRequestID))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertNotFoundEnvelope(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Empty(t, rec.Header().Get("Allow"), "Allow would list a disabled module's methods")
	var body envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_found", body.Error.Code)
	assert.Equal(t, testRequestID, body.Error.RequestID)
	assert.NotEmpty(t, body.Error.Message)
}

// reached records whether the wrapped handler ran.
func reached(hit *bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		*hit = true
		w.WriteHeader(http.StatusNoContent)
	}
}

func TestMuxGatesEachRouteByItsModule(t *testing.T) {
	tests := []struct {
		name     string
		flags    staticFlags
		wantCode int
		wantHit  bool
	}{
		{"module ON reaches the handler", staticFlags{devicesModule: true}, http.StatusNoContent, true},
		{"module OFF answers 404", staticFlags{}, http.StatusNotFound, false},
		{"another module ON does not open this one", staticFlags{"module_auth": true}, http.StatusNotFound, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := flaggate.NewMux(tc.flags)
			var hit bool
			mux.HandleFunc(registerDevice, reached(&hit))
			require.NoError(t, mux.Err())

			rec := serve(mux, http.MethodPost, "/v1/devices", "{}")

			assert.Equal(t, tc.wantCode, rec.Code)
			assert.Equal(t, tc.wantHit, hit)
			if tc.wantCode == http.StatusNotFound {
				assertNotFoundEnvelope(t, rec)
			}
		})
	}
}

func TestMuxNeverGatesPlatformRoutes(t *testing.T) {
	mux := flaggate.NewMux(staticFlags{})
	var hit bool
	mux.HandleFunc(healthPattern, reached(&hit))

	rec := serve(mux, http.MethodGet, "/healthz", "")

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, hit)
}

func TestMuxRecordsAnUnknownPatternAsAStartupError(t *testing.T) {
	mux := flaggate.NewMux(staticFlags{})
	var hit bool
	mux.HandleFunc("GET /v1/not-in-the-contract", reached(&hit))

	require.ErrorIs(t, mux.Err(), flaggate.ErrUnknownRoute)
	rec := serve(mux, http.MethodGet, "/v1/not-in-the-contract", "")
	assert.False(t, hit, "an unknown route is never registered ungated")
	assertNotFoundEnvelope(t, rec)
}

func TestMuxAnswersTheSame404ForUnmatchedRequests(t *testing.T) {
	mux := flaggate.NewMux(staticFlags{devicesModule: true})
	var hit bool
	mux.HandleFunc(deleteDevice, reached(&hit))

	for _, tc := range []struct{ name, method, target string }{
		{"unknown path", http.MethodGet, "/v1/nope"},
		{"other method on a known path (mux would say 405)", http.MethodPut, "/v1/devices/x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertNotFoundEnvelope(t, serve(mux, tc.method, tc.target, ""))
		})
	}
	assert.False(t, hit)
}

func TestHandlerGatesBeforeTheGeneratedWrapperBindsOrDecodes(t *testing.T) {
	off, err := flaggate.Handler(gen.NewStrictHandler(nil, nil), staticFlags{})
	require.NoError(t, err)

	assertNotFoundEnvelope(t, serve(off, http.MethodDelete, malformedPath, ""))
	assertNotFoundEnvelope(t, serve(off, http.MethodPost, "/v1/devices", "{not json"))
}

func TestHandlerLetsAnEnabledModuleReachTheGeneratedWrapper(t *testing.T) {
	on, err := flaggate.Handler(gen.NewStrictHandler(nil, nil), staticFlags{devicesModule: true})
	require.NoError(t, err)

	rec := serve(on, http.MethodDelete, malformedPath, "")

	assert.Equal(t, http.StatusBadRequest, rec.Code, "the wrapper ran and rejected the malformed uuid")
}

func TestWrapGatesASingleHandler(t *testing.T) {
	var hit bool
	h := flaggate.Wrap(staticFlags{}, devicesModule, http.HandlerFunc(reached(&hit)))

	assertNotFoundEnvelope(t, serve(h, http.MethodGet, "/anything", ""))
	assert.False(t, hit)
}

type failingReader struct{}

func (failingReader) EnabledKeys(context.Context) ([]string, error) {
	return nil, errors.New("connection refused")
}

func TestGateFailsClosedWhenTheFlagsCannotBeRead(t *testing.T) {
	mux := flaggate.NewMux(flags.NewService(failingReader{}))
	var hit bool
	mux.HandleFunc(registerDevice, reached(&hit))

	assertNotFoundEnvelope(t, serve(mux, http.MethodPost, "/v1/devices", "{}"))
	assert.False(t, hit)
}
