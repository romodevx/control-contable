package pgdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

const (
	nestUserA = "00000000-0000-4000-8000-00000000000a"
	nestUserB = "00000000-0000-4000-8000-00000000000b"
)

// inUnitOfWork returns a ctx that looks like one inside a running unit of work, so
// the nesting rules can be proven without a database. The integration suite proves
// the same rules against PostgreSQL.
func inUnitOfWork(t *testing.T, userID string, readOnly bool) (context.Context, *gorm.DB) {
	t.Helper()
	pool, err := Open("postgres://cc_app@127.0.0.1:1/control_contable?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, pool.Close()) })
	db, err := openGorm(pool)
	require.NoError(t, err)
	state := &unitOfWork{db: db, userID: userID, readOnly: readOnly}
	return context.WithValue(context.Background(), unitOfWorkKey{}, state), db
}

func TestNestedUnitOfWork(t *testing.T) {
	m := &TxManager{}
	tests := []struct {
		name     string
		outer    string // "" = System
		outerRO  bool
		run      func(context.Context, tx.Work) error
		wantErr  error
		wantJoin bool
	}{
		{"same user read-write joins", nestUserA, false,
			func(ctx context.Context, w tx.Work) error { return m.ReadWrite(ctx, nestUserA, w) }, nil, true},
		{"same user read-only joins a read-write", nestUserA, false,
			func(ctx context.Context, w tx.Work) error { return m.ReadOnly(ctx, nestUserA, w) }, nil, true},
		{"same user read-only joins a read-only", nestUserA, true,
			func(ctx context.Context, w tx.Work) error { return m.ReadOnly(ctx, nestUserA, w) }, nil, true},
		{"uppercase spelling of the same user joins", nestUserA, false,
			func(ctx context.Context, w tx.Work) error {
				return m.ReadWrite(ctx, "00000000-0000-4000-8000-00000000000A", w)
			}, nil, true},
		{"system joins system", "", false,
			func(ctx context.Context, w tx.Work) error { return m.System(ctx, w) }, nil, true},
		{"another user is refused", nestUserA, false,
			func(ctx context.Context, w tx.Work) error { return m.ReadWrite(ctx, nestUserB, w) }, tx.ErrUserMismatch, false},
		{"system inside a user is refused", nestUserA, false,
			func(ctx context.Context, w tx.Work) error { return m.System(ctx, w) }, tx.ErrUserMismatch, false},
		{"a user inside system is refused", "", false,
			func(ctx context.Context, w tx.Work) error { return m.ReadOnly(ctx, nestUserA, w) }, tx.ErrUserMismatch, false},
		{"read-write inside read-only is refused", nestUserA, true,
			func(ctx context.Context, w tx.Work) error { return m.ReadWrite(ctx, nestUserA, w) }, tx.ErrReadOnly, false},
		{"system inside read-only is refused", "", true,
			func(ctx context.Context, w tx.Work) error { return m.System(ctx, w) }, tx.ErrReadOnly, false},
		{"empty user inside a unit of work still fails closed", nestUserA, false,
			func(ctx context.Context, w tx.Work) error { return m.ReadWrite(ctx, "", w) }, tx.ErrNoUser, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, outerDB := inUnitOfWork(t, tt.outer, tt.outerRO)
			joined := false
			err := tt.run(ctx, func(inner context.Context) error {
				joined = true
				conn, err := Conn(inner)
				require.NoError(t, err)
				assert.Same(t, outerDB.Statement.ConnPool, conn.Statement.ConnPool, "nested work must reuse the outer transaction")
				return nil
			})
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantJoin, joined)
		})
	}
}

func TestNestedWorkErrorIsReturnedUnchanged(t *testing.T) {
	ctx, _ := inUnitOfWork(t, nestUserA, false)
	sentinel := assert.AnError
	err := (&TxManager{}).ReadWrite(ctx, nestUserA, func(context.Context) error { return sentinel })
	assert.Same(t, sentinel, err)
}
