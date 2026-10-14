package authhttp

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

// Refresher rotates a refresh token (usecase.RefreshSession).
type Refresher interface {
	Execute(ctx context.Context, presented string) (usecase.Tokens, error)
}

// SignOut revokes a refresh token family (usecase.Logout).
type SignOut interface {
	Execute(ctx context.Context, in usecase.LogoutInput) error
}

// SessionHandler implements the session operations of
// gen.StrictServerInterface (M1-08): refreshSession and logout. cmd/api
// embeds it in the API server next to Handler.
type SessionHandler struct {
	refresh Refresher
	logout  SignOut
}

// NewSessionHandler returns the handler over the use cases; both are
// required.
func NewSessionHandler(refresh Refresher, logout SignOut) *SessionHandler {
	return &SessionHandler{refresh: refresh, logout: logout}
}

// RefreshSession implements gen.StrictServerInterface (POST /v1/auth/refresh).
// Every refusal is the one generic 401 (US-A2b).
func (h *SessionHandler) RefreshSession(ctx context.Context, req gen.RefreshSessionRequestObject) (gen.RefreshSessionResponseObject, error) {
	tokens, err := h.refresh.Execute(ctx, req.Body.RefreshToken)
	if limited, ok := tooManyRequests(ctx, err); ok {
		return gen.RefreshSession429JSONResponse{TooManyRequestsJSONResponse: limited}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.RefreshSession200JSONResponse{
		Body: toTokenPair(tokens),
	}, nil
}

// Logout implements gen.StrictServerInterface (POST /v1/auth/logout) for
// the bearer token's user: 204 also when the token was already revoked or
// is not the caller's (US-A2c). Its 401 (no principal, or a malformed
// refresh token per RefreshRequest) is the contract's Unauthorized
// response, bearer challenge included.
func (h *SessionHandler) Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.Logout401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	err = h.logout.Execute(ctx, usecase.LogoutInput{UserID: userID, RefreshToken: req.Body.RefreshToken})
	if errors.Is(err, usecase.ErrTokenMalformed) || errors.Is(err, usecase.ErrInvalidAccessToken) {
		return gen.Logout401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.Logout204Response{}, nil
}
