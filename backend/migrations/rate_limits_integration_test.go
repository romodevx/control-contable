//go:build integration

package migrations_test

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Rate-limit buckets (M1-10, ADR-0005): rate_limit_take is GCRA over one row per
// (scope, key_hash), decided by one atomic UPDATE, so every API instance shares the
// same budget. The caller passes its clock, so these tests never sleep.

const (
	callTake  = "SELECT extract(epoch FROM rate_limit_take($1, $2, $3, make_interval(secs => $4), $5))::float8"
	callReset = "SELECT rate_limit_reset($1, $2)"

	testScope  = "qa_scope"
	testBurst  = 5
	testPeriod = 180 // seconds between refills
)

var bucketEpoch = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

// take spends one event at now and returns the wait in seconds (0 = admitted).
func take(t *testing.T, db *sql.DB, scope string, key []byte, now time.Time) float64 {
	t.Helper()
	var wait float64
	require.NoError(t, db.QueryRowContext(t.Context(), callTake, scope, key, now, testPeriod, testBurst).Scan(&wait))
	return wait
}

func TestRateLimitTake_admitsTheBurstThenDenies(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	key := blindIndex("ana@x")

	for i := range testBurst {
		assert.Zerof(t, take(t, app, testScope, key, bucketEpoch), "event %d is within the burst", i+1)
	}
	assert.InDelta(t, testPeriod, take(t, app, testScope, key, bucketEpoch), 0.001,
		"the next token arrives one interval later")
	assert.InDelta(t, testPeriod-60, take(t, app, testScope, key, bucketEpoch.Add(time.Minute)), 0.001,
		"the wait shrinks with the clock")
}

func TestRateLimitTake_aDeniedEventSpendsNothing(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	key := blindIndex("ana@x")
	for range testBurst {
		require.Zero(t, take(t, app, testScope, key, bucketEpoch))
	}
	for range 10 {
		require.NotZero(t, take(t, app, testScope, key, bucketEpoch))
	}

	refilled := bucketEpoch.Add(testPeriod * time.Second)
	assert.Zero(t, take(t, app, testScope, key, refilled), "one token refilled despite the denials")
	assert.NotZero(t, take(t, app, testScope, key, refilled), "and only one")
}

func TestRateLimitTake_aFullIntervalPerTokenRefillsTheBurst(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	key := blindIndex("ana@x")
	for range testBurst {
		require.Zero(t, take(t, app, testScope, key, bucketEpoch))
	}

	later := bucketEpoch.Add(testBurst * testPeriod * time.Second)
	for i := range testBurst {
		assert.Zerof(t, take(t, app, testScope, key, later), "event %d after a full refill", i+1)
	}
	assert.NotZero(t, take(t, app, testScope, key, later))
}

func TestRateLimitTake_keysAndScopesAreIndependent(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	ana, bea := blindIndex("ana@x"), blindIndex("bea@x")
	for range testBurst {
		require.Zero(t, take(t, app, testScope, ana, bucketEpoch))
	}
	require.NotZero(t, take(t, app, testScope, ana, bucketEpoch))

	assert.Zero(t, take(t, app, testScope, bea, bucketEpoch), "another key")
	assert.Zero(t, take(t, app, "qa_other_scope", ana, bucketEpoch), "another scope")
}

func TestRateLimitReset_restoresTheWholeBurst(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	key := blindIndex("ana@x")
	for range testBurst {
		require.Zero(t, take(t, app, testScope, key, bucketEpoch))
	}

	_, err := app.ExecContext(t.Context(), callReset, testScope, key)
	require.NoError(t, err)

	for range testBurst {
		assert.Zero(t, take(t, app, testScope, key, bucketEpoch))
	}
	_, err = app.ExecContext(t.Context(), callReset, testScope, blindIndex("unknown"))
	assert.NoError(t, err, "resetting a key never seen is a no-op")
}

// The decision is one atomic UPDATE: concurrent takes from many connections
// (many API instances) admit exactly the burst, never more.
func TestRateLimitTake_concurrentTakesAdmitExactlyTheBurst(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	const callers = 24
	app.SetMaxOpenConns(callers)
	key := blindIndex("ana@x")

	var (
		wg       sync.WaitGroup
		start    = make(chan struct{})
		mu       sync.Mutex
		admitted int
	)
	for range callers {
		wg.Go(func() {
			<-start
			var wait float64
			err := app.QueryRowContext(t.Context(), callTake, testScope, key, bucketEpoch, testPeriod, testBurst).Scan(&wait)
			if !assert.NoError(t, err) || wait != 0 {
				return
			}
			mu.Lock()
			admitted++
			mu.Unlock()
		})
	}
	close(start)
	wg.Wait()

	assert.Equal(t, testBurst, admitted)
}

func TestRateLimitTake_rejectsMalformedArguments(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	key := blindIndex("ana@x")
	tests := []struct {
		name   string
		scope  string
		key    []byte
		period int
		burst  int
		state  string
	}{
		{name: "short key", scope: testScope, key: []byte("raw-ip"), period: testPeriod, burst: testBurst, state: sqlstateInvalidParameterValue},
		{name: "bad scope", scope: "Login IP", key: key, period: testPeriod, burst: testBurst, state: sqlstateInvalidParameterValue},
		{name: "zero interval", scope: testScope, key: key, period: 0, burst: testBurst, state: sqlstateInvalidParameterValue},
		{name: "zero burst", scope: testScope, key: key, period: testPeriod, burst: 0, state: sqlstateInvalidParameterValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), callTake, tt.scope, tt.key, bucketEpoch, tt.period, tt.burst)
			requireSQLState(t, err, tt.state)
		})
	}
}

// cc_app reaches the buckets only to take and reset; it cannot delete or
// rewrite them, and the other runtime roles cannot touch them at all.
// Expired rows are removed by the worker (M9), not by the API.
func TestRateLimitBuckets_privileges(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	require.Zero(t, take(t, app, testScope, blindIndex("ana@x"), bucketEpoch))

	t.Run("cc_app cannot delete or rekey a bucket", func(t *testing.T) {
		for _, stmt := range []string{
			"DELETE FROM rate_limit_buckets",
			"UPDATE rate_limit_buckets SET scope = 'other'",
			"UPDATE rate_limit_buckets SET key_hash = key_hash",
		} {
			_, err := app.ExecContext(t.Context(), stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		}
	})

	for _, role := range []pgtest.Role{pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role)+" has no access", func(t *testing.T) {
			conn := db.Connect(t, role)
			_, err := conn.ExecContext(t.Context(), "SELECT 1 FROM rate_limit_buckets")
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
			_, err = conn.ExecContext(t.Context(), callTake, testScope, blindIndex("x"), bucketEpoch, testPeriod, testBurst)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}
