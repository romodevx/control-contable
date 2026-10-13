package pgdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sync"
)

// WorkerLeaderLockKey is the session advisory lock cmd/worker instances
// compete for: the ASCII bytes of "cc_worke" as a big-endian int64. Only the
// worker takes it.
const WorkerLeaderLockKey int64 = 0x63635f776f726b65

// ErrLeadershipLost reports that the leader's lock can no longer be verified:
// its connection died or the lock is gone. The lock is released (or was
// already) and the leader must stop acting as one.
var ErrLeadershipLost = errors.New("pgdb: worker leadership lost")

// lockHeldQuery confirms that this session still holds the lock. An advisory
// lock on a bigint key shows in pg_locks as classid (high 32 bits) and objid
// (low 32 bits) with objsubid 1.
const lockHeldQuery = `
SELECT EXISTS (
  SELECT 1 FROM pg_locks
  WHERE locktype = 'advisory' AND granted AND objsubid = 1 AND pid = pg_backend_pid()
    AND classid::bigint = $1 AND objid::bigint = $2)`

// AdvisoryLeader implements worker.Elector with a session-level
// pg_try_advisory_lock held on one connection reserved from the pool for as
// long as this instance leads. The lock lives exactly as long as that
// session, so a dead connection is a lost leadership: PostgreSQL releases the
// lock and a standby takes it at its next attempt. A standby holds no
// connection between attempts.
//
// A connection that leads is never returned to the pool: on loss or Resign it
// is discarded, so no pooled connection ever carries the lock.
type AdvisoryLeader struct {
	pool *sql.DB
	key  int64

	mu   sync.Mutex
	conn *sql.Conn // non-nil exactly while leading
}

// NewAdvisoryLeader competes for key on pool (logged in as cc_worker).
func NewAdvisoryLeader(pool *sql.DB, key int64) (*AdvisoryLeader, error) {
	if pool == nil {
		return nil, errors.New("pgdb: nil pool")
	}
	return &AdvisoryLeader{pool: pool, key: key}, nil
}

// TryLead implements worker.Elector: a leader verifies that its session still
// holds the lock; a standby tries to take it.
func (l *AdvisoryLeader) TryLead(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conn != nil {
		return l.confirm(ctx)
	}
	return l.acquire(ctx)
}

func (l *AdvisoryLeader) confirm(ctx context.Context) (bool, error) {
	var held bool
	err := l.conn.QueryRowContext(ctx, lockHeldQuery, l.key>>32, l.key&0xffffffff).Scan(&held)
	if err != nil || !held {
		l.discard()
		return false, errors.Join(ErrLeadershipLost, err)
	}
	return true, nil
}

func (l *AdvisoryLeader) acquire(ctx context.Context) (bool, error) {
	conn, err := l.pool.Conn(ctx)
	if err != nil {
		return false, fmt.Errorf("pgdb: leader connection: %w", err)
	}
	var won bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", l.key).Scan(&won); err != nil {
		// The lock may have been granted before the error: never pool this session.
		discardConn(conn)
		return false, fmt.Errorf("pgdb: try leader lock: %w", err)
	}
	if !won {
		// No lock taken, so the connection is clean and goes back to the pool.
		if err := conn.Close(); err != nil {
			return false, fmt.Errorf("pgdb: release standby connection: %w", err)
		}
		return false, nil
	}
	l.conn = conn
	return true, nil
}

// Resign implements worker.Elector: it unlocks, so a standby may take over as
// soon as Resign returns, then closes the leading session, which releases the
// lock even when the unlock failed. It is a no-op for a standby.
func (l *AdvisoryLeader) Resign(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.conn == nil {
		return nil
	}
	defer l.discard()
	if _, err := l.conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", l.key); err != nil {
		return fmt.Errorf("pgdb: unlock leader lock (the session is closed instead): %w", err)
	}
	return nil
}

func (l *AdvisoryLeader) discard() {
	if l.conn != nil {
		discardConn(l.conn)
		l.conn = nil
	}
}

// discardConn closes conn's physical connection instead of pooling it:
// database/sql drops a connection whose Raw callback reports ErrBadConn.
// Closing the session ends it on the server, which releases its locks.
func discardConn(conn *sql.Conn) {
	_ = conn.Raw(func(any) error { return driver.ErrBadConn })
}
