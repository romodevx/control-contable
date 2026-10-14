// Package ratelimitpg keeps rate-limit buckets in PostgreSQL (ADR-0005), so
// every API instance spends from the same budget: one row per (scope, key)
// in rate_limit_buckets, decided by rate_limit_take's atomic GCRA UPDATE.
//
// Each call is one autocommit statement on the pool, never part of the
// caller's unit of work (it does not read the transaction in ctx): an event
// is spent even when the request's own transaction rolls back. Keys are
// stored as HMAC-SHA256 digests under a key derived from the server secret,
// so no client address or email reaches the table.
package ratelimitpg

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// keyPurpose derives the bucket-key HMAC key from the server secret, so the
// digests are unrelated to any other HMAC made with that secret.
const keyPurpose = "control-contable/rate-limit-key/v1"

const (
	// takeSQL passes and returns durations as whole microseconds (the
	// database's resolution), the wait rounded up.
	takeSQL  = "SELECT ceil(extract(epoch FROM rate_limit_take($1, $2, $3, $4::bigint * interval '1 microsecond', $5)) * 1000000)::bigint"
	resetSQL = "SELECT rate_limit_reset($1, $2)"
)

// Store creates the buckets of each policy over one pool (cc_app).
type Store struct {
	pool *sql.DB
	key  []byte
	now  func() time.Time
}

// NewStore returns the store. secret (at least crypto.MinPepperSize bytes)
// keys the bucket digests; now is the clock every decision uses.
func NewStore(pool *sql.DB, secret []byte, now func() time.Time) (*Store, error) {
	switch {
	case pool == nil:
		return nil, errors.New("ratelimitpg: nil pool")
	case now == nil:
		return nil, errors.New("ratelimitpg: nil clock")
	case len(secret) < crypto.MinPepperSize:
		return nil, fmt.Errorf("ratelimitpg: secret of %d bytes, want at least %d", len(secret), crypto.MinPepperSize)
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(keyPurpose))
	return &Store{pool: pool, key: mac.Sum(nil), now: now}, nil
}

// Bucket returns the limiter of p. It panics on an invalid policy: policies
// are constants, so that is a programming error caught at wiring.
func (s *Store) Bucket(p ratelimit.Policy) *Bucket {
	if err := p.Validate(); err != nil {
		panic(err)
	}
	return &Bucket{store: s, policy: p}
}

// Bucket enforces one policy, a bucket per key. It is safe for concurrent use.
type Bucket struct {
	store  *Store
	policy ratelimit.Policy
}

// Take spends one event of key's bucket. When none is left it returns
// allowed = false and how long until one is (nothing is spent). On an error
// allowed is false: the caller fails closed.
func (b *Bucket) Take(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error) {
	var waitMicros int64
	err = b.store.pool.QueryRowContext(ctx, takeSQL, b.policy.Scope, b.digest(key), b.store.now(),
		b.policy.Interval.Microseconds(), b.policy.Burst).Scan(&waitMicros)
	if err != nil {
		return false, 0, fmt.Errorf("ratelimitpg: take %s: %w", b.policy.Scope, err)
	}
	if waitMicros <= 0 {
		return true, 0, nil
	}
	return false, time.Duration(waitMicros) * time.Microsecond, nil
}

// Reset fills key's bucket again.
func (b *Bucket) Reset(ctx context.Context, key string) error {
	if _, err := b.store.pool.ExecContext(ctx, resetSQL, b.policy.Scope, b.digest(key)); err != nil {
		return fmt.Errorf("ratelimitpg: reset %s: %w", b.policy.Scope, err)
	}
	return nil
}

// digest is HMAC(derived key, scope NUL key): 32 bytes, as the table requires.
func (b *Bucket) digest(key string) []byte {
	mac := hmac.New(sha256.New, b.store.key)
	mac.Write([]byte(b.policy.Scope))
	mac.Write([]byte{0})
	mac.Write([]byte(key))
	return mac.Sum(nil)
}
