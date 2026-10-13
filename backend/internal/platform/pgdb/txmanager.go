package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// setUserContext is the first statement of every transaction (data-model "RLS
// policy outline"): set_config with is_local = true is SET LOCAL that takes a
// bind parameter. GORM rewrites ? to $1.
const setUserContext = "SELECT set_config('app.current_user_id', ?, true)"

// systemUserContext is the value System sets: app_user_id() maps the empty string to NULL, which
// matches no row. Setting it explicitly, rather than leaving the session alone,
// keeps System fail-closed whatever state the pooled connection is in.
const systemUserContext = ""

// ErrNoTransaction reports database access outside a unit of work.
var ErrNoTransaction = errors.New("pgdb: no transaction in context; run the work through tx.Manager")

// TxManager implements tx.Manager on a GORM handle over a pgx pool.
type TxManager struct {
	db *gorm.DB
}

var _ tx.Manager = (*TxManager)(nil)

// NewTxManager builds a TxManager on pool (normally from Open), the same pool
// the rest of the process uses.
func NewTxManager(pool *sql.DB) (*TxManager, error) {
	if pool == nil {
		return nil, errors.New("pgdb: nil pool")
	}
	db, err := openGorm(pool)
	if err != nil {
		return nil, err
	}
	return &TxManager{db: db}, nil
}

// openGorm wraps pool without connecting. Every write already runs in our own
// transaction, so GORM's implicit per-write transaction is skipped.
func openGorm(pool *sql.DB) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
		Logger:                 logger.Discard,
	})
	if err != nil {
		return nil, fmt.Errorf("pgdb: gorm: %w", err)
	}
	return db, nil
}

// unitOfWorkKey is the context key of the running unit of work.
type unitOfWorkKey struct{}

// unitOfWork is the transaction a ctx carries, with the context it was opened for.
type unitOfWork struct {
	owner    *TxManager // the manager (pool, so database role) that opened it
	db       *gorm.DB
	userID   string // canonical UUID; systemUserContext for System
	readOnly bool
}

// ReadWrite implements tx.Manager.
func (m *TxManager) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	return m.runAsUser(ctx, userID, false, work)
}

// ReadOnly implements tx.Manager.
func (m *TxManager) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	return m.runAsUser(ctx, userID, true, work)
}

// System implements tx.Manager.
func (m *TxManager) System(ctx context.Context, work tx.Work) error {
	return m.run(ctx, systemUserContext, false, work)
}

func (m *TxManager) runAsUser(ctx context.Context, userID string, readOnly bool, work tx.Work) error {
	canonical, err := canonicalUserID(userID)
	if err != nil {
		return err
	}
	return m.run(ctx, canonical, readOnly, work)
}

// canonicalUserID rejects an empty or malformed id and returns the lower-case
// canonical spelling, so the same user always compares equal.
func canonicalUserID(userID string) (string, error) {
	if userID == "" {
		return "", tx.ErrNoUser
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", tx.ErrInvalidUser
	}
	return id.String(), nil
}

func (m *TxManager) run(ctx context.Context, userID string, readOnly bool, work tx.Work) error {
	if outer, ok := ctx.Value(unitOfWorkKey{}).(*unitOfWork); ok {
		return m.join(ctx, outer, userID, readOnly, work)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.begin(ctx, userID, readOnly, work)
}

// join runs work inside the outer transaction, as long as that cannot widen or
// switch what the outer transaction may do. The outer transaction must be this
// manager's: another manager's runs on another pool, as another database role.
func (m *TxManager) join(ctx context.Context, outer *unitOfWork, userID string, readOnly bool, work tx.Work) error {
	if outer.owner != m {
		return tx.ErrForeignTransaction
	}
	if outer.userID != userID {
		return tx.ErrUserMismatch
	}
	if outer.readOnly && !readOnly {
		return tx.ErrReadOnly
	}
	return work(ctx)
}

// begin opens the transaction, sets the user context and settles the outcome:
// commit on success, rollback on error, panic or cancellation. database/sql
// also rolls back on its own as soon as ctx is cancelled.
func (m *TxManager) begin(ctx context.Context, userID string, readOnly bool, work tx.Work) (err error) {
	gtx := m.db.WithContext(ctx).Begin(&sql.TxOptions{ReadOnly: readOnly})
	if gtx.Error != nil {
		return fmt.Errorf("pgdb: begin: %w", gtx.Error)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		rbErr := rollback(gtx)
		if p := recover(); p != nil {
			panic(p)
		}
		err = errors.Join(err, rbErr)
	}()
	if err := gtx.Exec(setUserContext, userID).Error; err != nil {
		return fmt.Errorf("pgdb: set user context: %w", err)
	}
	inner := context.WithValue(ctx, unitOfWorkKey{}, &unitOfWork{owner: m, db: gtx, userID: userID, readOnly: readOnly})
	if err := work(inner); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := gtx.Commit().Error; err != nil {
		return fmt.Errorf("pgdb: commit: %w", err)
	}
	committed = true
	return nil
}

// rollback ends gtx, ignoring the transaction being already over (database/sql
// rolls back by itself when the context is cancelled).
func rollback(gtx *gorm.DB) error {
	if err := gtx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
		return fmt.Errorf("pgdb: rollback: %w", err)
	}
	return nil
}

// Conn returns the transaction of the unit of work running in ctx, bound to
// ctx. Repositories call it at the start of every method. Outside a unit of
// work it returns ErrNoTransaction: there is no way to query without the RLS
// context.
func Conn(ctx context.Context) (*gorm.DB, error) {
	uow, ok := ctx.Value(unitOfWorkKey{}).(*unitOfWork)
	if !ok {
		return nil, ErrNoTransaction
	}
	return uow.db.WithContext(ctx), nil
}
