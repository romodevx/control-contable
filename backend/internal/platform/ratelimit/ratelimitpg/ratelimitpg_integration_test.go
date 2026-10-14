//go:build integration

package ratelimitpg_test

import (
	"bytes"
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
)

// testSecret is a made-up 32-byte key, not a credential.
var testSecret = bytes.Repeat([]byte{0x5a}, 32)

var policy = ratelimit.PerWindow("login_account", 5, 15*time.Minute)

// fakeClock is the injected clock; tests move it instead of sleeping.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newStore(t *testing.T, pool *sql.DB) (*ratelimitpg.Store, *fakeClock) {
	t.Helper()
	clock := &fakeClock{now: time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)}
	store, err := ratelimitpg.NewStore(pool, testSecret, clock.Now)
	require.NoError(t, err)
	return store, clock
}

func spend(t *testing.T, b *ratelimitpg.Bucket, key string, n int) {
	t.Helper()
	for i := range n {
		allowed, _, err := b.Take(t.Context(), key)
		require.NoError(t, err)
		require.Truef(t, allowed, "event %d", i+1)
	}
}

func TestBucket_TakeAdmitsTheBurstThenAnswersTheWait(t *testing.T) {
	store, clock := newStore(t, pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp))
	b := store.Bucket(policy)
	spend(t, b, "ana", policy.Burst)

	allowed, wait, err := b.Take(t.Context(), "ana")
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 3*time.Minute, wait)

	clock.advance(3 * time.Minute)
	allowed, _, err = b.Take(t.Context(), "ana")
	require.NoError(t, err)
	assert.True(t, allowed, "one event refilled")
}

func TestBucket_ResetRefillsTheBurst(t *testing.T) {
	store, _ := newStore(t, pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp))
	b := store.Bucket(policy)
	spend(t, b, "ana", policy.Burst)

	require.NoError(t, b.Reset(t.Context(), "ana"))

	spend(t, b, "ana", policy.Burst)
}

func TestBucket_KeysAndPoliciesAreIndependent(t *testing.T) {
	store, _ := newStore(t, pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp))
	b := store.Bucket(policy)
	spend(t, b, "ana", policy.Burst)

	spend(t, b, "bea", 1)
	spend(t, store.Bucket(ratelimit.PerWindow("login_ip", 1, time.Minute)), "ana", 1)
}

// Keys never reach the database in the clear: a client address or an email
// hash is stored only as a keyed digest (security.md: no PII at rest).
func TestBucket_StoresOnlyKeyedDigests(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	store, _ := newStore(t, app)
	const ip = "203.0.113.20"
	spend(t, store.Bucket(policy), ip, 1)

	var key []byte
	require.NoError(t, app.QueryRowContext(t.Context(),
		"SELECT key_hash FROM rate_limit_buckets WHERE scope = 'login_account'").Scan(&key))
	assert.Len(t, key, 32)
	assert.NotContains(t, string(key), ip)

	other, err := ratelimitpg.NewStore(db.Connect(t, pgtest.RoleApp), bytes.Repeat([]byte{0x33}, 32), time.Now)
	require.NoError(t, err)
	spend(t, other.Bucket(policy), ip, policy.Burst) // another secret, another digest: a fresh bucket
}

// The limiter's statement runs on its own pooled connection, never in the
// caller's unit of work: a rolled-back business transaction keeps its spent
// events (a failed login counts even when the request fails later).
func TestBucket_TakeOutlivesTheCallersRollback(t *testing.T) {
	pool := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	store, _ := newStore(t, pool)
	b := store.Bucket(ratelimit.PerWindow("login_account", 1, time.Hour))
	txm, err := pgdb.NewTxManager(pool)
	require.NoError(t, err)

	errRollback := assert.AnError
	err = txm.System(t.Context(), func(ctx context.Context) error {
		allowed, _, err := b.Take(ctx, "ana")
		require.NoError(t, err)
		require.True(t, allowed)
		return errRollback
	})
	require.ErrorIs(t, err, errRollback)

	allowed, _, err := b.Take(t.Context(), "ana")
	require.NoError(t, err)
	assert.False(t, allowed, "the event spent inside the rolled-back transaction still counts")
}

func TestBucket_FailsClosedOnADatabaseError(t *testing.T) {
	pool := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	store, _ := newStore(t, pool)
	require.NoError(t, pool.Close())

	allowed, _, err := store.Bucket(policy).Take(t.Context(), "ana")

	require.Error(t, err)
	assert.False(t, allowed)
	assert.Error(t, store.Bucket(policy).Reset(t.Context(), "ana"))
}

func TestNewStore_RejectsBadArguments(t *testing.T) {
	_, err := ratelimitpg.NewStore(&sql.DB{}, []byte("short"), time.Now)
	assert.Error(t, err, "short secret")
	_, err = ratelimitpg.NewStore(nil, testSecret, time.Now)
	assert.Error(t, err, "nil pool")
	_, err = ratelimitpg.NewStore(&sql.DB{}, testSecret, nil)
	assert.Error(t, err, "nil clock")
}

func TestStore_BucketPanicsOnAnInvalidPolicy(t *testing.T) {
	store, _ := newStore(t, &sql.DB{})
	assert.Panics(t, func() { store.Bucket(ratelimit.Policy{Scope: "x"}) })
}
