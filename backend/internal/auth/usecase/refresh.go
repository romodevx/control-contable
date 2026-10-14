package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// RefreshDeps are the ports RefreshSession needs.
type RefreshDeps struct {
	// Tx runs the rotation as the user the token's prefix names.
	Tx       tx.Manager
	Clock    clock.Clock
	Entropy  io.Reader
	Tokens   RefreshRotation
	Accounts AccountStatus
	Access   AccessTokens
	// Logger receives the token_reuse audit event and the refusal reasons;
	// nil discards them.
	Logger *slog.Logger
}

// RefreshSession rotates a refresh token (US-A2b).
//
// Concurrency policy: strict, no grace window. Rotation locks the presented
// token's row (SELECT ... FOR UPDATE) before anything else, so two refreshes
// of one token run one after the other: the first rotates it, the second
// then finds it rotated and is treated as reuse, which revokes the whole
// family, the first one's new token included. A grace window (answering the
// loser with the winner's pair, or letting a just-rotated token through for
// a few seconds) would hand a thief racing the client the very session
// reuse detection exists to end, and would need the successor's secret kept
// or re-derivable server side. The cost of strictness, a double submit
// signing the user out, is avoided where it arises: the mobile client
// serialises refreshes in a QueuedInterceptor (MODULE-auth US-A2b, security
// gate M1-02 condition 6).
type RefreshSession struct {
	deps RefreshDeps
}

// NewRefreshSession builds the use case.
func NewRefreshSession(deps RefreshDeps) *RefreshSession {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &RefreshSession{deps: deps}
}

// Execute rotates the presented token and returns a new pair. Every refusal
// wraps ErrRefreshRejected; any other error is an infrastructure failure.
//
// A refusal commits its unit of work: reuse detection's family revocation
// must outlive the 401 it answers.
func (uc *RefreshSession) Execute(ctx context.Context, presented string) (Tokens, error) {
	token, err := domain.ParseRefreshToken(presented)
	if err != nil {
		return Tokens{}, uc.refuse(ctx, domain.UserID{}, ErrTokenMalformed)
	}
	var (
		tokens  Tokens
		refusal error
	)
	err = uc.deps.Tx.ReadWrite(ctx, token.UserID().String(), func(ctx context.Context) error {
		var err error
		tokens, err = uc.rotate(ctx, token)
		if errors.Is(err, ErrRefreshRejected) {
			refusal = err
			return nil // commit: a reuse revocation must stay
		}
		return err
	})
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: refresh session: %w", err)
	}
	if refusal != nil {
		return Tokens{}, uc.refuse(ctx, token.UserID(), refusal)
	}
	return tokens, nil
}

// rotate claims the token and issues its successor. A refusal wraps
// ErrRefreshRejected.
func (uc *RefreshSession) rotate(ctx context.Context, token domain.RefreshToken) (Tokens, error) {
	stored, err := uc.claim(ctx, token)
	if err != nil {
		return Tokens{}, err
	}
	userID := token.UserID()
	active, err := uc.deps.Accounts.CanSignIn(ctx, userID)
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: read account status: %w", err)
	}
	if !active {
		return Tokens{}, ErrAccountInactive
	}
	successor, err := domain.NewRefreshToken(userID, uc.deps.Entropy)
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: new refresh token: %w", err)
	}
	lifetime, err := uc.deps.Tokens.Rotate(ctx, stored, successor, RefreshSlidingWindow)
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: rotate refresh token: %w", err)
	}
	access, accessTTL, err := uc.deps.Access.Sign(userID, uc.deps.Clock.Now())
	if err != nil {
		return Tokens{}, fmt.Errorf("auth: sign access token: %w", err)
	}
	return Tokens{Access: access, AccessExpiresIn: accessTTL, Refresh: successor, RefreshExpiresIn: lifetime}, nil
}

// claim locks the presented token and decides whether it may be rotated.
// Nothing else is read until it may (security gate M1-02, condition 1).
// Order matters: a revoked token is already dead; a rotated one is reuse
// even once expired (a stolen copy replayed late is still theft); only a
// token never rotated nor revoked can be merely expired (condition 2).
func (uc *RefreshSession) claim(ctx context.Context, token domain.RefreshToken) (StoredRefreshToken, error) {
	hash := token.Hash()
	stored, found, err := uc.deps.Tokens.Lock(ctx, hash)
	if err != nil {
		return StoredRefreshToken{}, fmt.Errorf("auth: lock refresh token: %w", err)
	}
	switch {
	case !found || !stored.Hash.Equal(hash):
		return StoredRefreshToken{}, ErrTokenUnknown
	case stored.Revoked:
		return StoredRefreshToken{}, ErrTokenRevoked
	case stored.Rotated:
		if err := uc.deps.Tokens.RevokeFamily(ctx, hash, RevokeReuseDetected); err != nil {
			return StoredRefreshToken{}, fmt.Errorf("auth: revoke reused family: %w", err)
		}
		return StoredRefreshToken{}, ErrTokenReused
	case stored.Expired:
		return StoredRefreshToken{}, ErrTokenExpired
	}
	return stored, nil
}

// refuse logs why a refresh was refused (never the token) and returns the
// refusal. Reuse is a security event: it is audited at warning level.
func (uc *RefreshSession) refuse(ctx context.Context, userID domain.UserID, refusal error) error {
	attrs := []any{slog.String("reason", refusal.Error())}
	if !userID.IsZero() {
		attrs = append(attrs, slog.String("user_id", userID.String()))
	}
	if errors.Is(refusal, ErrTokenReused) {
		uc.deps.Logger.WarnContext(ctx, "auth audit", append([]any{slog.String("event", "token_reuse")}, attrs...)...)
		return refusal
	}
	uc.deps.Logger.DebugContext(ctx, "refresh refused", attrs...)
	return refusal
}
