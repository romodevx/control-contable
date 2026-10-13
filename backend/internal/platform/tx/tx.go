// Package tx is the unit-of-work port (ADR-0001, data-model "RLS policy outline").
//
// A use case wraps every database access in a unit of work. The adapter
// (internal/platform/pgdb) runs it in one PostgreSQL transaction whose first
// statement is `SELECT set_config('app.current_user_id', $1, true)`, the
// parameterized, transaction-local form of SET LOCAL, so every Row Level
// Security policy sees the right user and the setting dies with the
// transaction: it never leaks to the next borrower of the pooled connection.
//
// Rules every implementation keeps:
//
//   - A user unit of work needs a user id. An empty id (or the nil UUID) is
//     ErrNoUser, never a transaction without context: callers cannot fail
//     open by accident.
//   - System runs with the user context explicitly empty, so user-scoped
//     tables show and accept no rows. It is for work that has no user yet or
//     spans users only through SECURITY DEFINER functions (the worker's
//     due_reminder_users / purge_soft_deleted, sign-up's definer calls).
//   - Nested calls (a unit of work started from inside another one, through
//     the ctx it received) join the outer transaction instead of opening a
//     second one: they commit or roll back with it. A nested call for another
//     user is ErrUserMismatch; a read-write call nested in a read-only one is
//     ErrReadOnly. Neither ever switches the outer transaction's context.
//     A nested call through another Manager (another pool and database role)
//     is ErrForeignTransaction: it never runs with the other role's grants.
//   - The work's error, a panic or the cancellation of ctx rolls the
//     transaction back; only a nil return on a live ctx commits.
//
// The package is framework-free so use cases can import it (archtest rule
// usecase-platform-adapter forbids them the adapter itself).
package tx

import (
	"context"
	"errors"
)

// Work is one unit of work. It must use the ctx it receives (never the
// caller's) for every database access and nested unit of work: that ctx
// carries the transaction.
type Work func(ctx context.Context) error

// Manager runs units of work as a user. Implementations are safe for
// concurrent use; a single unit of work is not (one goroutine at a time per
// transaction).
type Manager interface {
	// ReadWrite runs work in a read-write transaction as userID.
	ReadWrite(ctx context.Context, userID string, work Work) error
	// ReadOnly runs work in a READ ONLY transaction as userID. Nested inside
	// a ReadWrite unit of work it joins that transaction, so READ ONLY is
	// only a hint there: the nested work can write. Never rely on ReadOnly
	// as a security boundary.
	ReadOnly(ctx context.Context, userID string, work Work) error
}

// SystemManager runs units of work with no user context (P0-30 D-2). It is a
// separate port so that only the code that needs it (sign-up's definer calls,
// the worker's cross-user jobs) can be handed one; a use case that holds only
// a Manager cannot drop the user context. HTTP adapters import neither port
// (archtest rule adapter-unit-of-work): they call use cases.
type SystemManager interface {
	// System runs work in a read-write transaction with no user context.
	System(ctx context.Context, work Work) error
}

var (
	// ErrNoUser reports a user unit of work started without a user id.
	ErrNoUser = errors.New("tx: unit of work without a user id")
	// ErrInvalidUser reports a user id that is not a UUID.
	ErrInvalidUser = errors.New("tx: user id is not a uuid")
	// ErrUserMismatch reports a nested unit of work for a different user (or
	// System inside a user's, or the reverse) than the transaction it joins.
	ErrUserMismatch = errors.New("tx: nested unit of work for a different user context")
	// ErrReadOnly reports a read-write unit of work nested in a read-only one.
	ErrReadOnly = errors.New("tx: read-write unit of work nested in a read-only one")
	// ErrForeignTransaction reports a unit of work nested in one that another
	// Manager (another pool, so another database role) is running.
	ErrForeignTransaction = errors.New("tx: nested unit of work started through another manager")
)
