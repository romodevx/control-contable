package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// startFamilySQL stores a family's first refresh token. Expiries come from
// the database clock, the same one the refresh_tokens_expiry_window CHECK
// compares them with, so app/database clock skew cannot break the insert.
const startFamilySQL = `INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, absolute_expires_at)
VALUES (?, ?, ?, now() + make_interval(secs => ?), now() + make_interval(secs => ?))`

// lockRowSQL reads the presented token's row and locks it until the
// transaction ends. RLS limits it to the unit of work's user, the one the
// token's prefix names. Expiry is judged by the database clock.
const lockRowSQL = `SELECT id::text AS id, family_id::text AS family, token_hash,
  rotated_at IS NOT NULL AS rotated, revoked_at IS NOT NULL AS revoked,
  (expires_at <= now() OR absolute_expires_at <= now()) AS expired
FROM refresh_tokens WHERE token_hash = ? FOR UPDATE`

// rotateSQL claims the locked parent and inserts its successor in one
// statement: same user and family, rotated_from = parent, the family's
// absolute expiry, and a sliding expiry that never passes it. rotated_at
// uses clock_timestamp(), not now(): the parent may have been issued by a
// transaction that began after this one (refresh_tokens_rotated_after_issue).
// The claim repeats the live conditions, so a parent that is no longer
// claimable inserts nothing.
const rotateSQL = `WITH parent AS (
  UPDATE refresh_tokens SET rotated_at = clock_timestamp()
  WHERE id = ?::uuid AND rotated_at IS NULL AND revoked_at IS NULL
  RETURNING id, user_id, family_id, absolute_expires_at)
INSERT INTO refresh_tokens (user_id, family_id, token_hash, rotated_from, expires_at, absolute_expires_at)
SELECT user_id, family_id, ?, id, least(now() + make_interval(secs => ?), absolute_expires_at), absolute_expires_at
FROM parent
RETURNING floor(extract(epoch FROM expires_at - now()))::bigint AS lifetime_seconds`

// revokeFamilySQL revokes the live rest of the family of the token with
// the given hash, as the unit of work's user (RLS; the (user_id, family_id)
// index serves it). Rows already revoked keep their first reason (the
// history trigger forbids changing it). revoked_at uses clock_timestamp():
// a concurrent rotation may have issued a token after this transaction
// began (refresh_tokens_revoked_after_issue). RevokeFamily repeats it until
// it changes nothing (see there).
const revokeFamilySQL = `UPDATE refresh_tokens SET revoked_at = clock_timestamp(), revoke_reason = ?
WHERE revoked_at IS NULL
  AND family_id IN (SELECT family_id FROM refresh_tokens WHERE token_hash = ?)`

// errNotClaimable reports a rotation whose parent was no longer live:
// impossible under the row lock Lock takes, so a server fault.
var errNotClaimable = errors.New("auth infra: refresh token not claimable")

// RefreshTokenStore implements usecase.RefreshTokens and
// usecase.RefreshRotation over refresh_tokens, in the unit of work in ctx.
type RefreshTokenStore struct{}

// NewRefreshTokenStore returns the store.
func NewRefreshTokenStore() RefreshTokenStore { return RefreshTokenStore{} }

var (
	_ usecase.RefreshTokens   = RefreshTokenStore{}
	_ usecase.RefreshRotation = RefreshTokenStore{}
)

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

type lockedRow struct {
	ID        string
	Family    string
	TokenHash []byte
	Rotated   bool
	Revoked   bool
	Expired   bool
}

// Lock implements usecase.RefreshRotation.
func (RefreshTokenStore) Lock(ctx context.Context, hash domain.TokenHash) (usecase.StoredRefreshToken, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredRefreshToken{}, false, err
	}
	var rows []lockedRow
	if err := db.Raw(lockRowSQL, bytea(hash.Bytes())).Scan(&rows).Error; err != nil {
		return usecase.StoredRefreshToken{}, false, fmt.Errorf("auth infra: lock refresh token: %w", err)
	}
	if len(rows) == 0 {
		return usecase.StoredRefreshToken{}, false, nil
	}
	row := rows[0]
	stored, err := domain.TokenHashFromBytes(row.TokenHash)
	if err != nil {
		return usecase.StoredRefreshToken{}, false, fmt.Errorf("auth infra: stored token hash: %w", err)
	}
	return usecase.StoredRefreshToken{
		ID: row.ID, Family: row.Family, Hash: stored, Rotated: row.Rotated, Revoked: row.Revoked, Expired: row.Expired,
	}, true, nil
}

// Rotate implements usecase.RefreshRotation.
func (RefreshTokenStore) Rotate(ctx context.Context, parent usecase.StoredRefreshToken, successor domain.RefreshToken, sliding time.Duration) (time.Duration, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	var lifetimes []int64
	err = db.Raw(rotateSQL, parent.ID, bytea(successor.Hash().Bytes()), int64(sliding/time.Second)).
		Scan(&lifetimes).Error
	if err != nil {
		return 0, fmt.Errorf("auth infra: rotate refresh token: %w", err)
	}
	if len(lifetimes) == 0 {
		return 0, errNotClaimable
	}
	return time.Duration(lifetimes[0]) * time.Second, nil
}

// RevokeFamily implements usecase.FamilyRevoker.
//
// One UPDATE is not enough under READ COMMITTED: when it waits on a row a
// concurrent rotation locked, it resumes once that rotation commits but
// never sees the successor the rotation inserted, which would stay live.
// So the statement repeats until it revokes nothing: each pass sees what
// was committed before it began, and the rows already revoked stay locked
// by this transaction, so any later rotation of them waits and then finds
// them revoked. It ends once no rotation of the family is in flight.
func (RefreshTokenStore) RevokeFamily(ctx context.Context, hash domain.TokenHash, reason usecase.RevokeReason) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	for {
		res := db.Exec(revokeFamilySQL, string(reason), bytea(hash.Bytes()))
		if res.Error != nil {
			return fmt.Errorf("auth infra: revoke refresh token family: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return nil
		}
	}
}
