//go:build integration

package pgdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/worker"
)

var _ worker.Elector = (*pgdb.AdvisoryLeader)(nil)

// Leader election of the worker (P0-27): pg_try_advisory_lock on a dedicated
// cc_worker connection. Each instance has its own pool, as separate processes do.

func newLeader(t *testing.T, database *pgtest.Database) *pgdb.AdvisoryLeader {
	t.Helper()
	l, err := pgdb.NewAdvisoryLeader(database.Connect(t, pgtest.RoleWorker), pgdb.WorkerLeaderLockKey)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, l.Resign(context.Background())) })
	return l
}

func tryLead(t *testing.T, l *pgdb.AdvisoryLeader) bool {
	t.Helper()
	leading, err := l.TryLead(t.Context())
	require.NoError(t, err)
	return leading
}

// holders lists the backends holding the worker's leader lock.
func holders(t *testing.T, db *sql.DB) []int {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT pid FROM pg_locks
		WHERE locktype = 'advisory' AND granted AND objsubid = 1
		  AND classid::bigint = ($1::bigint >> 32) AND objid::bigint = ($1::bigint & 4294967295)`,
		pgdb.WorkerLeaderLockKey)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	var pids []int
	for rows.Next() {
		var pid int
		require.NoError(t, rows.Scan(&pid))
		pids = append(pids, pid)
	}
	require.NoError(t, rows.Err())
	return pids
}

func TestAdvisoryLeader_onlyOneInstanceLeads(t *testing.T) {
	database := pgtest.NewDatabase(t)
	a, b := newLeader(t, database), newLeader(t, database)
	observer := database.Connect(t, pgtest.RoleWorker)

	require.True(t, tryLead(t, a))
	assert.False(t, tryLead(t, b))
	assert.True(t, tryLead(t, a), "the leader confirms its lock on every heartbeat")
	assert.False(t, tryLead(t, b))

	assert.Len(t, holders(t, observer), 1, "confirming never stacks the lock")
}

func TestAdvisoryLeader_resignHandsLeadershipOver(t *testing.T) {
	database := pgtest.NewDatabase(t)
	a, b := newLeader(t, database), newLeader(t, database)
	observer := database.Connect(t, pgtest.RoleWorker)
	require.True(t, tryLead(t, a))

	require.NoError(t, a.Resign(t.Context()))

	assert.Empty(t, holders(t, observer), "the lock never goes back to the pool")
	assert.True(t, tryLead(t, b))
	assert.False(t, tryLead(t, a))
}

// The leader's connection dies (crash, network, failover): its session lock is
// gone with it, a standby takes over, and the old leader notices at its next
// heartbeat instead of believing it still leads.
func TestAdvisoryLeader_lostConnectionLosesLeadership(t *testing.T) {
	database := pgtest.NewDatabase(t)
	a, b := newLeader(t, database), newLeader(t, database)
	killer := database.Connect(t, pgtest.RoleWorker) // same role: may terminate its backends
	require.True(t, tryLead(t, a))
	pids := holders(t, killer)
	require.Len(t, pids, 1)

	var terminated bool
	require.NoError(t, killer.QueryRowContext(t.Context(),
		"SELECT pg_terminate_backend($1, 5000)", pids[0]).Scan(&terminated))
	require.True(t, terminated)

	assert.True(t, tryLead(t, b), "the standby takes over")
	leading, err := a.TryLead(t.Context())
	require.ErrorIs(t, err, pgdb.ErrLeadershipLost)
	assert.False(t, leading)
	assert.False(t, tryLead(t, a), "the old leader reconnects as a standby")
	assert.Len(t, holders(t, killer), 1)
}

// pg_locks shows a bigint key as two unsigned 32-bit oids, so a key with the
// high bit set (negative) must still be confirmed, not lost on every heartbeat.
func TestAdvisoryLeader_confirmsANegativeKey(t *testing.T) {
	database := pgtest.NewDatabase(t)
	const negativeKey int64 = -0x1122334455667788
	l, err := pgdb.NewAdvisoryLeader(database.Connect(t, pgtest.RoleWorker), negativeKey)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, l.Resign(context.Background())) })

	require.True(t, tryLead(t, l))
	assert.True(t, tryLead(t, l), "the leader confirms a negative key's lock")
}

func TestAdvisoryLeader_aStandbyHoldsNoConnection(t *testing.T) {
	database := pgtest.NewDatabase(t)
	a := newLeader(t, database)
	pool := database.Connect(t, pgtest.RoleWorker)
	b, err := pgdb.NewAdvisoryLeader(pool, pgdb.WorkerLeaderLockKey)
	require.NoError(t, err)
	require.True(t, tryLead(t, a))

	assert.False(t, tryLead(t, b))

	assert.Zero(t, pool.Stats().InUse, "a standby returns its connection to the pool")
	require.NoError(t, b.Resign(t.Context()), "resigning without leading is a no-op")
}
