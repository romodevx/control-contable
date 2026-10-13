package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNoUser is returned by AsUser for an empty user id: an empty context would
// silently behave like an unset one. Use WithoutUser to test the unset case.
var ErrNoUser = errors.New("pgtest: empty user id")

// setUserContext is the TxManager's statement (data-model "RLS policy outline"):
// the parameterizable, transaction-local equivalent of SET LOCAL.
const setUserContext = "SELECT set_config('app.current_user_id', $1, true)"

// txEnd says how a helper transaction finishes.
type txEnd int

const (
	commit txEnd = iota
	rollback
)

// AsUser runs fn in a transaction whose app.current_user_id is userID, then
// commits. If fn fails the transaction is rolled back and fn's error returned.
func AsUser(ctx context.Context, db *sql.DB, userID string, fn func(*sql.Tx) error) error {
	if userID == "" {
		return ErrNoUser
	}
	return inTx(ctx, db, userID, commit, fn)
}

// WithoutUser runs fn in a transaction with no user context set, then commits.
func WithoutUser(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	return inTx(ctx, db, "", commit, fn)
}

// probe runs fn under userID ("" = unset) and always rolls back, so isolation
// checks leave no trace even when they find a breach.
func probe(ctx context.Context, db *sql.DB, userID string, fn func(*sql.Tx) error) error {
	return inTx(ctx, db, userID, rollback, fn)
}

func inTx(ctx context.Context, db *sql.DB, userID string, end txEnd, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pgtest: begin: %w", err)
	}
	if err := runInTx(ctx, tx, userID, fn); err != nil {
		return errors.Join(err, rollbackErr(tx))
	}
	if end == rollback {
		return rollbackErr(tx)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pgtest: commit: %w", err)
	}
	return nil
}

func runInTx(ctx context.Context, tx *sql.Tx, userID string, fn func(*sql.Tx) error) error {
	if userID != "" {
		if _, err := tx.ExecContext(ctx, setUserContext, userID); err != nil {
			return fmt.Errorf("pgtest: set user context: %w", err)
		}
	}
	return fn(tx)
}

func rollbackErr(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return fmt.Errorf("pgtest: rollback: %w", err)
	}
	return nil
}
