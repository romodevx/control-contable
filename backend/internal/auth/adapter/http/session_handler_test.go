package authhttp_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestRefreshSession_answers200WithANoStorePair(t *testing.T) {
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
	assert.Equal(t, "no-store", ok.Headers.CacheControl)
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

	resp, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: refreshBody("presented")})

	require.NoError(t, err)
	assert.IsType(t, gen.Logout204Response{}, resp)
	assert.Equal(t, usecase.LogoutInput{UserID: userIDText, RefreshToken: "presented"}, l.got)
}

// The authenticator runs first; a handler reached without a principal still
// fails closed with the generic 401, never a logout for an empty user.
func TestLogout_withoutAPrincipalIs401(t *testing.T) {
	l := &fakeLogout{}

	_, err := sessionHandler(&fakeRefresher{}, l).Logout(t.Context(), gen.LogoutRequestObject{Body: refreshBody("presented")})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeUnauthorized, appErr.Code)
	assert.Zero(t, l.calls)
}

func TestLogout_malformedTokenIs401(t *testing.T) {
	l := &fakeLogout{err: usecase.ErrTokenMalformed}
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText})

	_, err := sessionHandler(&fakeRefresher{}, l).Logout(ctx, gen.LogoutRequestObject{Body: refreshBody("x")})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeUnauthorized, appErr.Code)
}
