package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// Tokens is a new token pair (contract TokenPair).
type Tokens struct {
	Access           string
	AccessExpiresIn  time.Duration
	Refresh          domain.RefreshToken
	RefreshExpiresIn time.Duration
}

// Session is a signed-in account (contract AuthSession).
type Session struct {
	Account Account
	Tokens  Tokens
}

// sessionIssuer starts a token family: a stored refresh token and a signed
// access token. Login (M1-07) reuses it.
type sessionIssuer struct {
	entropy io.Reader
	refresh RefreshTokens
	access  AccessTokens
}

// issue must run inside the unit of work of userID: the refresh token row is
// RLS-scoped to it.
func (s sessionIssuer) issue(ctx context.Context, userID domain.UserID, now time.Time) (Tokens, error) {
	refresh, err := domain.NewRefreshToken(userID, s.entropy)
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: new refresh token: %w", err)
	}
	if err := s.refresh.StartFamily(ctx, refresh, RefreshSlidingWindow, RefreshAbsoluteLifetime); err != nil {
		return Tokens{}, fmt.Errorf("auth: store refresh token: %w", err)
	}
	access, accessTTL, err := s.access.Sign(userID, now)
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: sign access token: %w", err)
	}
	return Tokens{Access: access, AccessExpiresIn: accessTTL, Refresh: refresh, RefreshExpiresIn: RefreshSlidingWindow}, nil
}
