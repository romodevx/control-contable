package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrNotIsolated reports an RLS breach: a check saw, changed or wrote across users.
	ErrNotIsolated = errors.New("pgtest: rls isolation breached")
	// ErrVacuous reports a check that would pass only because the owner has no rows.
	// Seed the owner's rows first.
	ErrVacuous = errors.New("pgtest: vacuous isolation proof (owner has no rows)")
)

// SQLSTATE codes that count as the database refusing access.
const (
	sqlstateInsufficientPrivilege = "42501" // RLS WITH CHECK, row_security=off, missing grant
	sqlstateUndefinedObject       = "42704" // current_setting() of a never-set parameter
	sqlstateInvalidTextRepr       = "22P02" // ''::uuid after a transaction-local context reverted
)

// failClosedCodes are the errors an unset user context may raise instead of
// returning zero rows: both outcomes fail closed.
var failClosedCodes = []string{sqlstateInsufficientPrivilege, sqlstateUndefinedObject, sqlstateInvalidTextRepr}

// Table names a user-scoped table and the column its RLS policy keys on.
// Name is unqualified and resolved through the search_path.
type Table struct {
	Name        string
	OwnerColumn string
}

func (t Table) String() string { return t.Name }

// IsolationSpec describes one table for AssertIsolation. Insert writes one row owned
// by ownerID inside tx; it is called under each user's own context to seed rows, and
// under UserA's context with UserB's id to prove the cross-user insert is rejected.
type IsolationSpec struct {
	Table  Table
	UserA  string
	UserB  string
	Insert func(ctx context.Context, tx *sql.Tx, ownerID string) error
}

// AssertIsolation seeds one row for each user, then runs every mandatory RLS check
// (data-model.md) with A as the viewer and B as the owner. db must be logged in as the
// runtime role under test (normally cc_app). Each failed check is reported separately.
func AssertIsolation(t testing.TB, db *sql.DB, spec IsolationSpec) {
	t.Helper()
	ctx := t.Context()
	for _, owner := range []string{spec.UserA, spec.UserB} {
		err := AsUser(ctx, db, owner, func(tx *sql.Tx) error { return spec.Insert(ctx, tx, owner) })
		if err != nil {
			t.Fatalf("pgtest: seed %s as its owner %s: %v", spec.Table, owner, err)
		}
	}
	crossInsert := func(ctx context.Context, tx *sql.Tx) error { return spec.Insert(ctx, tx, spec.UserB) }
	checks := []struct {
		name string
		err  error
	}{
		{"rls enabled and forced", CheckRLSForced(ctx, db, spec.Table)},
		{"cannot see", CheckCannotSee(ctx, db, spec.Table, spec.UserA, spec.UserB)},
		{"cannot update", CheckCannotUpdate(ctx, db, spec.Table, spec.UserA, spec.UserB)},
		{"cannot delete", CheckCannotDelete(ctx, db, spec.Table, spec.UserA, spec.UserB)},
		{"insert for another user rejected", CheckInsertRejected(ctx, db, spec.UserA, crossInsert)},
		{"unset context fails closed", CheckUnsetContextFailsClosed(ctx, db, spec.Table, spec.UserB)},
		{"no bypass", CheckNoBypass(ctx, db, spec.Table)},
	}
	for _, c := range checks {
		if c.err != nil {
			t.Errorf("%s: %s: %v", spec.Table, c.name, c.err)
		}
	}
}

// CheckRLSForced returns ErrNotIsolated unless the table has ENABLE and FORCE ROW
// LEVEL SECURITY (FORCE keeps the policy on for the table owner too).
func CheckRLSForced(ctx context.Context, db *sql.DB, t Table) error {
	var enabled, forced sql.NullBool
	err := db.QueryRowContext(ctx,
		"SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE oid = to_regclass($1)",
		QuoteIdent(t.Name)).Scan(&enabled, &forced)
	if err != nil {
		return fmt.Errorf("pgtest: read rls flags of %s: %w", t, err)
	}
	if !enabled.Bool || !forced.Bool {
		return fmt.Errorf("%w: %s has ENABLE=%t FORCE=%t ROW LEVEL SECURITY", ErrNotIsolated, t, enabled.Bool, forced.Bool)
	}
	return nil
}

// CheckCannotSee returns ErrNotIsolated if viewer can read any row of owner.
func CheckCannotSee(ctx context.Context, db *sql.DB, t Table, viewer, owner string) error {
	if err := requireOwnerRows(ctx, db, t, owner); err != nil {
		return err
	}
	var seen int
	err := probe(ctx, db, viewer, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, countOwnedSQL(t), owner).Scan(&seen)
	})
	if err != nil {
		return fmt.Errorf("pgtest: select %s as %s: %w", t, viewer, err)
	}
	if seen > 0 {
		return fmt.Errorf("%w: %s sees %d row(s) of %s in %s", ErrNotIsolated, viewer, seen, owner, t)
	}
	return nil
}

// CheckCannotUpdate returns ErrNotIsolated if viewer can update any row of owner.
// A permission error also passes: the role cannot update at all.
func CheckCannotUpdate(ctx context.Context, db *sql.DB, t Table, viewer, owner string) error {
	col := QuoteIdent(t.OwnerColumn)
	query := "UPDATE " + QuoteIdent(t.Name) + " SET " + col + " = " + col + " WHERE " + col + " = $1"
	return checkNoRowsAffected(ctx, db, t, viewer, owner, "update", query)
}

// CheckCannotDelete returns ErrNotIsolated if viewer can delete any row of owner.
// A permission error also passes: the role cannot delete at all.
func CheckCannotDelete(ctx context.Context, db *sql.DB, t Table, viewer, owner string) error {
	query := "DELETE FROM " + QuoteIdent(t.Name) + " WHERE " + QuoteIdent(t.OwnerColumn) + " = $1"
	return checkNoRowsAffected(ctx, db, t, viewer, owner, "delete", query)
}

// CheckInsertRejected runs insert under viewer's context and returns ErrNotIsolated
// if it succeeds. insert must write a row owned by someone else. Only an RLS or
// privilege refusal (42501) passes; any other error is returned, since a constraint
// failure would prove nothing about isolation.
func CheckInsertRejected(ctx context.Context, db *sql.DB, viewer string, insert func(context.Context, *sql.Tx) error) error {
	err := probe(ctx, db, viewer, func(tx *sql.Tx) error { return insert(ctx, tx) })
	if err == nil {
		return fmt.Errorf("%w: %s inserted a row for another user", ErrNotIsolated, viewer)
	}
	if hasSQLState(err, sqlstateInsufficientPrivilege) {
		return nil
	}
	return fmt.Errorf("pgtest: cross-user insert failed for a reason other than RLS: %w", err)
}

// CheckUnsetContextFailsClosed returns ErrNotIsolated if a transaction without a user
// context can read any row of t. Zero rows or a fail-closed error both pass.
func CheckUnsetContextFailsClosed(ctx context.Context, db *sql.DB, t Table, owner string) error {
	if err := requireOwnerRows(ctx, db, t, owner); err != nil {
		return err
	}
	var seen int
	err := probe(ctx, db, "", func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM "+QuoteIdent(t.Name)).Scan(&seen)
	})
	if hasSQLState(err, failClosedCodes...) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("pgtest: select %s without context: %w", t, err)
	}
	if seen > 0 {
		return fmt.Errorf("%w: %d row(s) of %s visible without a user context", ErrNotIsolated, seen, t)
	}
	return nil
}

// CheckNoBypass returns ErrNotIsolated if the logged-in role is SUPERUSER or
// BYPASSRLS, or can read t after `SET row_security = off`.
func CheckNoBypass(ctx context.Context, db *sql.DB, t Table) error {
	var role string
	var bypass bool
	err := db.QueryRowContext(ctx,
		"SELECT rolname, rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user").Scan(&role, &bypass)
	if err != nil {
		return fmt.Errorf("pgtest: read role attributes: %w", err)
	}
	if bypass {
		return fmt.Errorf("%w: role %s is SUPERUSER or BYPASSRLS", ErrNotIsolated, role)
	}
	err = probe(ctx, db, "", func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "SET LOCAL row_security = off"); err != nil {
			return err
		}
		//nolint:gosec // G202: the identifier is quoted by QuoteIdent.
		_, err := tx.ExecContext(ctx, "SELECT 1 FROM "+QuoteIdent(t.Name)+" LIMIT 1")
		return err
	})
	if hasSQLState(err, sqlstateInsufficientPrivilege) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("pgtest: row_security=off probe on %s: %w", t, err)
	}
	return fmt.Errorf("%w: role %s reads %s with row_security = off", ErrNotIsolated, role, t)
}

func checkNoRowsAffected(ctx context.Context, db *sql.DB, t Table, viewer, owner, verb, query string) error {
	if err := requireOwnerRows(ctx, db, t, owner); err != nil {
		return err
	}
	var affected int64
	err := probe(ctx, db, viewer, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, query, owner)
		if err != nil {
			return err
		}
		affected, err = res.RowsAffected()
		return err
	})
	if hasSQLState(err, sqlstateInsufficientPrivilege) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("pgtest: %s %s as %s: %w", verb, t, viewer, err)
	}
	if affected > 0 {
		return fmt.Errorf("%w: %s can %s %d row(s) of %s in %s", ErrNotIsolated, viewer, verb, affected, owner, t)
	}
	return nil
}

// requireOwnerRows returns ErrVacuous unless owner sees at least one of its own rows.
func requireOwnerRows(ctx context.Context, db *sql.DB, t Table, owner string) error {
	var own int
	err := probe(ctx, db, owner, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, countOwnedSQL(t), owner).Scan(&own)
	})
	if err != nil {
		return fmt.Errorf("pgtest: count own rows of %s in %s: %w", owner, t, err)
	}
	if own == 0 {
		return fmt.Errorf("%w: %s owns no rows in %s", ErrVacuous, owner, t)
	}
	return nil
}

func countOwnedSQL(t Table) string {
	return "SELECT count(*) FROM " + QuoteIdent(t.Name) + " WHERE " + QuoteIdent(t.OwnerColumn) + " = $1"
}

func hasSQLState(err error, codes ...string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && slices.Contains(codes, pgErr.Code)
}
