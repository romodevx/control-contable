package remindershttp_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// PR #162 review (carry-over): the push token never reaches a log line nor
// a response, on any path. These run registerDevice behind the real router
// (flag gate, authentication, contract validation, the error writer that
// logs 5xx causes) over the real use case, with the logger at DEBUG and
// its output captured.

const userHeader = "X-Test-User"

// principalAuthn admits the X-Test-User as the principal.
type principalAuthn struct{}

func (principalAuthn) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get(userHeader)
		if user == "" {
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
			return
		}
		next.ServeHTTP(w, r.WithContext(principal.With(r.Context(), principal.Principal{UserID: user})))
	})
}

// server serves the reminders handler; every other operation is the nil
// interface (never called here).
type server struct {
	*remindershttp.Handler
	others
}

type others struct{ gen.StrictServerInterface }

// lockedBuffer is a log sink safe for the router's goroutines.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type inlineTx struct{}

func (inlineTx) ReadWrite(ctx context.Context, _ string, work tx.Work) error { return work(ctx) }
func (inlineTx) ReadOnly(ctx context.Context, _ string, work tx.Work) error  { return work(ctx) }

// echoSealer "seals" by keeping the token: the worst case for a leak, so a
// log line or response carrying the sealed value would carry the token.
type echoSealer struct{}

func (echoSealer) Seal(token domain.PushToken, _ usecase.TokenBinding) (usecase.SealedToken, error) {
	return usecase.SealedToken{Ciphertext: []byte(token.Reveal()), KeyID: 1, HMAC: []byte(token.Reveal())}, nil
}

// leakyStore fails the claim with a wrapped error naming the device it was
// given (formatted with %v, as a careless adapter might).
type leakyStore struct{ fail bool }

var errStoreDown = errors.New("fake: store down")

func (s leakyStore) Claim(_ context.Context, _ string, d usecase.SealedDevice) (bool, error) {
	if s.fail {
		return false, errors.Join(errStoreDown, errors.New("claim "+d.InstallationID))
	}
	return true, nil
}

func (leakyStore) Device(_ context.Context, id string) (domain.Device, error) {
	return domain.Device{ID: id, Platform: domain.PlatformAndroid, CreatedAt: seenAt, LastSeenAt: seenAt}, nil
}

func (leakyStore) DeleteDevice(context.Context, string) (bool, error) { return true, nil }

func newLoggedRouter(t *testing.T, store usecase.DeviceStore) (http.Handler, *lockedBuffer) {
	t.Helper()
	devices, err := usecase.NewDevices(usecase.DevicesDeps{Tx: inlineTx{}, Store: store, Sealer: echoSealer{}})
	require.NoError(t, err)
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	logs := &lockedBuffer{}
	routes, err := router.New(server{Handler: remindershttp.NewHandler(remindershttp.UseCases{Devices: devices})}, spec,
		router.Options{
			Flags:  flagtest.Static{"module_reminders": true},
			Authn:  principalAuthn{},
			Logger: logging.New(logs, slog.LevelDebug),
		})
	require.NoError(t, err)
	return middleware.Chain(routes, middleware.NoStore), logs
}

func postDevice(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/devices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(userHeader, callerID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func deviceJSON(token, extra string) string {
	return `{"id":"` + installationID + `","platform":"android","push_token":"` + token + `"` + extra + `}`
}

func TestRegisterDevice_theTokenNeverReachesALogOrAResponse(t *testing.T) {
	token := pushToken(t)
	tests := []struct {
		name   string
		store  leakyStore
		body   string
		status int
		logged bool // a 5xx is logged with its cause
	}{
		{"201 registered", leakyStore{}, deviceJSON(token, ""), http.StatusCreated, false},
		{"500 store down (cause logged)", leakyStore{fail: true}, deviceJSON(token, ""), http.StatusInternalServerError, true},
		{"400 token with a space", leakyStore{}, deviceJSON(token+` `+token, ""), http.StatusBadRequest, false},
		{"400 token too long", leakyStore{}, deviceJSON(strings.Repeat(token, 4096/len(token)+1), ""), http.StatusBadRequest, false},
		{"400 unknown field", leakyStore{}, deviceJSON(token, `,"token_echo":"`+token+`"`), http.StatusBadRequest, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler, logs := newLoggedRouter(t, tc.store)

			rec := postDevice(handler, tc.body)

			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			assert.NotContains(t, rec.Body.String(), token, "never echoed in a response")
			assert.NotContains(t, logs.String(), token, "never logged")
			if tc.logged {
				assert.Contains(t, logs.String(), "request failed", "the 5xx was logged (so the check above saw a real line)")
				assert.Contains(t, logs.String(), errStoreDown.Error())
			}
		})
	}
}
