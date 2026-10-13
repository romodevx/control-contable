package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// startFamilySQL stores a family's first refresh token. Expiries come from
// the database clock, the same one the refresh_tokens_expiry_window CHECK
// compares them with, so app/database clock skew cannot break the insert.
const startFamilySQL = `INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, absolute_expires_at)
VALUES (?, ?, ?, now() + make_interval(secs => ?), now() + make_interval(secs => ?))`

// RefreshTokenStore implements usecase.RefreshTokens.
type RefreshTokenStore struct{}

// NewRefreshTokenStore returns the store.
func NewRefreshTokenStore() RefreshTokenStore { return RefreshTokenStore{} }

// StartFamily implements usecase.RefreshTokens. Only the token's SHA-256 is
// stored.
func (RefreshTokenStore) StartFamily(ctx context.Context, token domain.RefreshToken, sliding, absolute time.Duration) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	family, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("auth infra: family id: %w", err)
	}
	err = db.Exec(startFamilySQL, token.UserID().String(), family.String(), bytea(token.Hash().Bytes()),
		int64(sliding/time.Second), int64(absolute/time.Second)).Error
	if err != nil {
		return fmt.Errorf("auth infra: store refresh token: %w", err)
	}
	return nil
}
