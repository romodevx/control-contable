package authhttp_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// A *usecase.RateLimitedError is the contract's TooManyRequests on every
// operation that can return it: rate_limited, the request id and a
// Retry-After in whole seconds, rounded up.

var errLimited = fmt.Errorf("wrapped: %w", &usecase.RateLimitedError{RetryAfter: 90*time.Second + time.Millisecond})

func withRequestID(t *testing.T) context.Context {
	t.Helper()
	return requestid.With(t.Context(), "req_test")
}

func assertTooMany(t *testing.T, got gen.TooManyRequestsJSONResponse) {
	t.Helper()
	assert.Equal(t, 91, got.Headers.RetryAfter)
	assert.Equal(t, gen.ErrorCodeRateLimited, got.Body.Error.Code)
	assert.Equal(t, "Too many requests. Try again later.", got.Body.Error.Message)
	assert.Equal(t, "req_test", got.Body.Error.RequestId)
}

func TestLogin_aRateLimitedEmailIs429(t *testing.T) {
	resp, err := loginHandler(&fakeLogin{err: errLimited}).Login(withRequestID(t), gen.LoginRequestObject{Body: loginBody()})

	require.NoError(t, err)
	got, ok := resp.(gen.Login429JSONResponse)
	require.True(t, ok, "%T", resp)
	assertTooMany(t, got.TooManyRequestsJSONResponse)
}

func TestRegister_aRateLimitedEmailIs429(t *testing.T) {
	h := newHandler(&fakeRegister{err: errLimited}, &fakeVerify{}, &fakeResend{})

	resp, err := h.Register(withRequestID(t), gen.RegisterRequestObject{Body: &gen.RegisterJSONRequestBody{Email: "ana@example.mx"}})

	require.NoError(t, err)
	got, ok := resp.(gen.Register429JSONResponse)
	require.True(t, ok, "%T", resp)
	assertTooMany(t, got.TooManyRequestsJSONResponse)
}

func TestRefreshSession_aRateLimitedFamilyIs429(t *testing.T) {
	resp, err := sessionHandler(&fakeRefresher{err: errLimited}, &fakeLogout{}).RefreshSession(withRequestID(t),
		gen.RefreshSessionRequestObject{Body: refreshBody("x")})

	require.NoError(t, err)
	got, ok := resp.(gen.RefreshSession429JSONResponse)
	require.True(t, ok, "%T", resp)
	assertTooMany(t, got.TooManyRequestsJSONResponse)
}
