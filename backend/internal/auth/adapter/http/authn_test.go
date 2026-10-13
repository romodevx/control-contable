package authhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

const (
	admittedValue = "opaque-value.admitted~value" // not a JWT: the verifier is a fake
	jtiValue      = "6f1c2a4e-1d1b-4c55-8a0e-5e3f1f9b7d21"
)

// fakeVerifier accepts admittedValue only and records what it was asked.
type fakeVerifier struct {
	calls []string
}

func (f *fakeVerifier) Verify(token string) (usecase.AccessClaims, error) {
	f.calls = append(f.calls, token)
	if token != admittedValue {
		return usecase.AccessClaims{}, fmt.Errorf("%w: signature", usecase.ErrInvalidAccessToken)
	}
	id, err := domain.ParseUserID(userIDText)
	if err != nil {
		return usecase.AccessClaims{}, err
	}
	return usecase.AccessClaims{UserID: id, TokenID: jtiValue, ExpiresAt: time.Now().Add(time.Minute)}, nil
}

// recordingNext is the protected handler: it records the principal it saw.
type recordingNext struct {
	called bool
	got    principal.Principal
	ok     bool
}

func (n *recordingNext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.called = true
	n.got, n.ok = principal.From(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func authenticate(t *testing.T, headers ...string) (*httptest.ResponseRecorder, *recordingNext, *fakeVerifier, *bytes.Buffer) {
	t.Helper()
	var logs bytes.Buffer
	verifier := &fakeVerifier{}
	next := &recordingNext{}
	h := authhttp.NewAuthenticator(verifier, slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))).
		Authenticate(next)
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	for _, v := range headers {
		req.Header.Add("Authorization", v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, next, verifier, &logs
}

func TestAuthenticate_aValidBearerTokenSetsThePrincipal(t *testing.T) {
	for _, scheme := range []string{"Bearer", "bearer", "BEARER"} {
		t.Run(scheme, func(t *testing.T) {
			rec, next, verifier, _ := authenticate(t, scheme+" "+admittedValue)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			require.True(t, next.called)
			require.True(t, next.ok, "the principal is in the request context")
			assert.Equal(t, principal.Principal{UserID: userIDText, TokenID: jtiValue}, next.got)
			assert.Equal(t, []string{admittedValue}, verifier.calls)
		})
	}
}

// Every failure is the same 401 envelope, whatever the reason: the client
// learns nothing about why its credentials were refused.
func TestAuthenticate_rejectsWithTheGeneric401(t *testing.T) {
	tests := []struct {
		name         string
		headers      []string
		reachesVerif bool
	}{
		{"no Authorization header", nil, false},
		{"empty header", []string{""}, false},
		{"another scheme", []string{"Basic " + admittedValue}, false},
		{"scheme only", []string{"Bearer"}, false},
		{"scheme and a space", []string{"Bearer "}, false},
		{"two spaces", []string{"Bearer  " + admittedValue}, false},
		{"token with a space", []string{"Bearer " + admittedValue + " x"}, false},
		{"token in a query-like form", []string{"Bearer access_token=" + admittedValue}, false},
		{"two Authorization headers", []string{"Bearer " + admittedValue, "Bearer " + admittedValue}, false},
		{"a token the verifier refuses", []string{"Bearer opaque-value.refused~value"}, true},
	}
	var first string
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, next, verifier, logs := authenticate(t, tt.headers...)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, next.called, "the protected handler never runs")
			assert.Equal(t, tt.reachesVerif, len(verifier.calls) == 1)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			// RFC 9110 §15.5.2 / RFC 6750 §3: a 401 names the scheme to use.
			// One fixed challenge, so it says nothing about why.
			assert.Equal(t, []string{`Bearer realm="control-contable"`}, rec.Header().Values("WWW-Authenticate"))
			var body struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "unauthorized", body.Error.Code)
			assert.Equal(t, "Authentication required.", body.Error.Message)
			if first == "" {
				first = body.Error.Message
			}
			assert.Equal(t, first, body.Error.Message)
			assert.NotContains(t, logs.String(), "opaque-value", "a token never reaches the logs")
		})
	}
}

// The refusal the authenticator writes is the contract's Unauthorized
// response of every secured operation, WWW-Authenticate header included.
func TestAuthenticate_refusalConformsToTheContract(t *testing.T) {
	c := contracttest.Load(t)
	h := middleware.RequestID(authhttp.NewAuthenticator(&fakeVerifier{}, nil).Authenticate(&recordingNext{}))
	for _, header := range []string{"", "Bearer opaque-value.refused~value"} {
		req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}

		rec := c.Serve(t, h, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}
