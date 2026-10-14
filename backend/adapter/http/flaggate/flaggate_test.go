package flaggate_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	devicesModule = "module_reminders"
	testRequestID = "req-123"
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

func serve(h http.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/devices", nil)
	req = req.WithContext(requestid.With(req.Context(), testRequestID))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertNotFoundEnvelope(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "not_found", body.Error.Code)
	assert.Equal(t, testRequestID, body.Error.RequestID)
	assert.NotEmpty(t, body.Error.Message)
}

// reached records whether the wrapped handler ran.
func reached(hit *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*hit = true
		w.WriteHeader(http.StatusNoContent)
	})
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		flags   staticFlags
		module  string
		wantHit bool
	}{
		{"module ON reaches the handler", staticFlags{devicesModule: true}, devicesModule, true},
		{"module OFF answers 404", staticFlags{}, devicesModule, false},
		{"another module ON does not open this one", staticFlags{"module_auth": true}, devicesModule, false},
		{"platform is never gated", staticFlags{}, gen.ModulePlatform, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var hit bool
			rec := serve(flaggate.Wrap(tc.flags, tc.module, reached(&hit)))

			assert.Equal(t, tc.wantHit, hit)
			if !tc.wantHit {
				assertNotFoundEnvelope(t, rec)
				return
			}
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}
}

type failingReader struct{}

func (failingReader) EnabledKeys(context.Context) ([]string, error) {
	return nil, errors.New("connection refused")
}

func TestWrapFailsClosedWhenTheFlagsCannotBeRead(t *testing.T) {
	var hit bool
	rec := serve(flaggate.Wrap(flags.NewService(failingReader{}), devicesModule, reached(&hit)))

	assertNotFoundEnvelope(t, rec)
	assert.False(t, hit)
}

func TestWriteNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	flaggate.WriteNotFound(rec, req.WithContext(requestid.With(req.Context(), testRequestID)))

	assertNotFoundEnvelope(t, rec)
}
