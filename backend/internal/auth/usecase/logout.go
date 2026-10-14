package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// LogoutDeps are the ports Logout needs.
type LogoutDeps struct {
	// Tx runs the revocation as the bearer token's user.
	Tx     tx.Manager
	Tokens FamilyRevoker
	// Logger receives the logout audit event; nil discards it.
	Logger *slog.Logger
}

// LogoutInput is a logout request: the authenticated caller and the refresh
// token whose family ends.
type LogoutInput struct {
	// UserID is the bearer access token's subject.
	UserID       string
	RefreshToken string
}

// Logout revokes a refresh token family (US-A2c). It revokes, never
// deletes: the rotated tokens stay as reuse evidence until the family dies
// (security gate M1-02, condition 4).
type Logout struct {
	deps LogoutDeps
}

// NewLogout builds the use case.
func NewLogout(deps LogoutDeps) *Logout {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &Logout{deps: deps}
}

// Execute revokes the family of the caller's refresh token. It is
// idempotent: an unknown, already revoked or another user's token is a
// success, and another user's token is not even looked up, so its state is
// never revealed. A malformed token is ErrTokenMalformed (the contract's
// RefreshRequest: malformed answers 401); a caller that is not a user id is
// ErrInvalidAccessToken.
func (uc *Logout) Execute(ctx context.Context, in LogoutInput) error {
	caller, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return fmt.Errorf("%w: logout caller: %w", ErrInvalidAccessToken, err)
	}
	token, err := domain.ParseRefreshToken(in.RefreshToken)
	if err != nil {
		return ErrTokenMalformed
	}
	if token.UserID() != caller {
		return nil
	}
	err = uc.deps.Tx.ReadWrite(ctx, caller.String(), func(ctx context.Context) error {
		return uc.deps.Tokens.RevokeFamily(ctx, token.Hash(), RevokeLogout)
	})
	if err != nil {
		return fmt.Errorf("auth: logout: %w", err)
	}
	uc.deps.Logger.InfoContext(ctx, "auth audit", slog.String("event", "logout"), slog.String("user_id", caller.String()))
	return nil
}
