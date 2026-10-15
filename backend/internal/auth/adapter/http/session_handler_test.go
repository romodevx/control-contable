package authhttp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

type fakeRefresher struct {
	got    string
	tokens usecase.Tokens
	err    error
}

func (f *fakeRefresher) Execute(_ context.Context, presented string) (usecase.Tokens, error) {
	f.got = presented
	return f.tokens, f.err
}

type fakeLogout struct {
	got   usecase.LogoutInput
	calls int
	err   error
}

func (f *fakeLogout) Execute(_ context.Context, in usecase.LogoutInput) error {
	f.calls++
	f.got = in
	return f.err
}

func sessionHandler(r *fakeRefresher, l *fakeLogout) *authhttp.SessionHandler {
	return authhttp.NewSessionHandler(r, l)
}

func refreshBody(token string) *gen.RefreshSessionJSONRequestBody {
	return &gen.RefreshSessionJSONRequestBody{RefreshToken: token}
}

func logoutBody(token string) *gen.LogoutJSONRequestBody {
	return &gen.LogoutJSONRequestBody{RefreshToken: token}
}

func TestRefreshSession_answers200WithAPair(t *testing.T) {
	userID, err := domain.ParseUserID(userIDText)
	require.NoError(t, err)
	refresh, err := domain.NewRefreshToken(userID, zeroes{})
	require.NoError(t, err)
	r := &fakeRefresher{tokens: usecase.Tokens{
		Access: "jwt", AccessExpiresIn: 15 * time.Minute, Refresh: refresh, RefreshExpiresIn: 2 * time.Hour,
	}}

	resp, err := sessionHandler(r, &fakeLogout{}).RefreshSession(t.Context(),
		gen.RefreshSessionRequestObject{Body: refreshBody("presented")})

	require.NoError(t, err)
	assert.Equal(t, "presented", r.got)
	ok, isOK := resp.(gen.RefreshSession200JSONResponse)
	require.True(t, isOK, "%T", resp)
	assert.Equal(t, gen.TokenPair{
		AccessToken: "jwt", AccessExpiresIn: 900, RefreshToken: refresh.Reveal(), RefreshExpiresIn: 7200, TokenType: gen.Bearer,
	}, ok.Body)
}

func TestRefreshSession_refusalIsTheGeneric401(t *testing.T) {
	r := &fakeRefresher{err: usecase.ErrTokenReused}

	_, err := sessionHandler(r, &fakeLogout{}).RefreshSession(t.Context(),
		gen.RefreshSessionRequestObject{Body: refreshBody("presented")})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeUnauthorized, appErr.Code)
	assert.Equal(t, "Your session has ended. Sign in again.", appErr.Message)
}

func TestRefreshSession_infrastructureFailureIsNotMapped(t *testing.T) {
	r := &fakeRefresher{err: errBoom}

	_, err := sessionHandler(r, &fakeLogout{}).RefreshSession(t.Context(),
		gen.RefreshSessionRequestObject{Body: refreshBody("presented")})

	require.ErrorIs(t, err, errBoom)
	_, isApp := errors.AsType[*apperr.Error](err)
	assert.False(t, isApp, "left for the router's generic 500")
}

func TestLogout_answers204ForTheBearersUser(t *testing.T) {
	l := &fakeLogout{}
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText, TokenID: "jti"})

	resp, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: logoutBody("presented")})

	require.NoError(t, err)
	assert.IsType(t, gen.Logout204Response{}, resp)
	assert.Equal(t, usecase.LogoutInput{UserID: userIDText, RefreshToken: "presented"}, l.got)
}

// M9 security gate F1: the optional device_id reaches the use case, which
// removes that push device with the session.
func TestLogout_passesThePushDeviceID(t *testing.T) {
	l := &fakeLogout{}
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText})
	device := uuid.MustParse("0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23")
	body := logoutBody("presented")
	body.DeviceId = &device

	resp, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: body})

	require.NoError(t, err)
	assert.IsType(t, gen.Logout204Response{}, resp)
	assert.Equal(t, usecase.LogoutInput{UserID: userIDText, RefreshToken: "presented", DeviceID: device.String()}, l.got)
}

// A device id the use case refuses is a 400 naming device_id, never a 500.
func TestLogout_invalidDeviceIDIsAValidationFailure(t *testing.T) {
	l := &fakeLogout{err: usecase.ErrInvalidDeviceID}
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText})

	_, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: logoutBody("x")})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "device_id", appErr.Fields[0].Field)
	assert.Equal(t, "format", appErr.Fields[0].Code)
}

// The authenticator runs first; a handler reached without a principal still
// fails closed with the generic 401, never a logout for an empty user.
//
// Every logout 401 is the contract's Unauthorized response, whose
// WWW-Authenticate header is required (M1-16: the malformed-token path
// went out as a plain apperr 401 without the challenge).
func TestLogout_401sCarryTheBearerChallenge(t *testing.T) {
	withUser := func(t *testing.T) context.Context {
		return principal.With(t.Context(), principal.Principal{UserID: userIDText})
	}
	tests := []struct {
		name      string
		ctx       func(t *testing.T) context.Context
		logout    *fakeLogout
		wantCalls int
	}{
		{name: "without a principal", ctx: func(t *testing.T) context.Context { return t.Context() }, logout: &fakeLogout{}, wantCalls: 0},
		{name: "malformed refresh token", ctx: withUser, logout: &fakeLogout{err: usecase.ErrTokenMalformed}, wantCalls: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := sessionHandler(&fakeRefresher{}, tc.logout).Logout(tc.ctx(t), gen.LogoutRequestObject{Body: logoutBody("x")})

			require.NoError(t, err)
			unauthorized, ok := resp.(gen.Logout401JSONResponse)
			require.True(t, ok, "got %T", resp)
			assert.Equal(t, `Bearer realm="control-contable"`, unauthorized.Headers.WWWAuthenticate)
			assert.Equal(t, gen.ErrorCodeUnauthorized, unauthorized.Body.Error.Code)
			assert.Equal(t, tc.wantCalls, tc.logout.calls)
		})
	}
}

// Only the credential failures become the 401 response; anything else
// still reaches the error mapper (a 500, never a disguised 401).
func TestLogout_otherFailuresStayErrors(t *testing.T) {
	dbDown := errors.New("db down")
	l := &fakeLogout{err: dbDown}
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText})

	resp, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: logoutBody("x")})

	require.ErrorIs(t, err, dbDown)
	assert.Nil(t, resp)
}
