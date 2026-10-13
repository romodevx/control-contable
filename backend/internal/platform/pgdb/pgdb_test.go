package pgdb_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// unreachableDSN points at a closed loopback port: connecting is refused at once,
// so these unit tests prove what happens before (or instead of) a round trip.
const unreachableDSN = "postgres://cc_app:s3cr3t-pw@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1" //nolint:gosec // G101: fake password of a test fixture; the port is closed.

const userA = "00000000-0000-4000-8000-00000000000a"

var _ tx.Manager = (*pgdb.TxManager)(nil)

func openUnreachable(t *testing.T) *sql.DB {
	t.Helper()
	db, err := pgdb.Open(unreachableDSN)
	require.NoError(t, err, "Open must not connect")
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	return db
}

func newUnreachableManager(t *testing.T) *pgdb.TxManager {
	t.Helper()
	m, err := pgdb.NewTxManager(openUnreachable(t))
	require.NoError(t, err)
	return m
}

func noWork(t *testing.T) tx.Work {
	return func(context.Context) error {
		t.Error("work must not run")
		return nil
	}
}

func TestUserUnitOfWorkRejectsBadUserBeforeConnecting(t *testing.T) {
	m := newUnreachableManager(t)
	variants := map[string]func(context.Context, string, tx.Work) error{
		"ReadWrite": m.ReadWrite,
		"ReadOnly":  m.ReadOnly,
	}
	ids := []struct {
		name    string
		userID  string
		wantErr error
	}{
		{"empty id fails closed", "", tx.ErrNoUser},
		{"not a uuid", "not-a-uuid", tx.ErrInvalidUser},
		{"sql injection attempt", "' OR true --", tx.ErrInvalidUser},
	}
	for variant, run := range variants {
		for _, id := range ids {
			t.Run(variant+"/"+id.name, func(t *testing.T) {
				err := run(t.Context(), id.userID, noWork(t))
				assert.ErrorIs(t, err, id.wantErr)
			})
		}
	}
}

func TestCancelledContextNeverBegins(t *testing.T) {
	m := newUnreachableManager(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	assert.ErrorIs(t, m.ReadWrite(ctx, userA, noWork(t)), context.Canceled)
	assert.ErrorIs(t, m.ReadOnly(ctx, userA, noWork(t)), context.Canceled)
	assert.ErrorIs(t, m.System(ctx, noWork(t)), context.Canceled)
}

func TestBeginFailureSkipsWork(t *testing.T) {
	m := newUnreachableManager(t)
	err := m.System(t.Context(), noWork(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pgdb: begin")
	assert.NotContains(t, err.Error(), "s3cr3t-pw")
}

func TestConnOutsideUnitOfWorkFailsClosed(t *testing.T) {
	_, err := pgdb.Conn(t.Context())
	assert.ErrorIs(t, err, pgdb.ErrNoTransaction)
}

func TestNewTxManagerRequiresPool(t *testing.T) {
	_, err := pgdb.NewTxManager(nil)
	assert.Error(t, err)
}

func TestOpen(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		db := openUnreachable(t)
		assert.Equal(t, pgdb.DefaultMaxOpenConns, db.Stats().MaxOpenConnections)
	})
	t.Run("max open conns option", func(t *testing.T) {
		db, err := pgdb.Open(unreachableDSN, pgdb.WithMaxOpenConns(2))
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, db.Close()) })
		assert.Equal(t, 2, db.Stats().MaxOpenConnections)
	})
	t.Run("malformed dsn never echoes the password", func(t *testing.T) {
		_, err := pgdb.Open("postgres://cc_app:s3cr3t-pw@127.0.0.1:notaport/db")
		require.Error(t, err)
		assert.NotContains(t, err.Error(), "s3cr3t-pw")
	})
	t.Run("empty dsn", func(t *testing.T) {
		_, err := pgdb.Open("")
		assert.Error(t, err)
	})
}

func TestPingerReportsUnreachableDatabase(t *testing.T) {
	p := pgdb.NewPinger(openUnreachable(t))
	err := p.Ping(t.Context())
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cr3t-pw")
}
