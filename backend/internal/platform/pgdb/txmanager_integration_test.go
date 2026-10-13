//go:build integration

package pgdb_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/migrations"
)

// The fail-closed suite of the TxManager (P0-12), run against the real identity
// migration as cc_app (and cc_worker for System).

const userB = "00000000-0000-4000-8000-00000000000b"

// The user-scoped identity tables (P0-11) and the column their policies key on.
var userTables = []pgtest.Table{
	{Name: "users", OwnerColumn: "id"},
	{Name: "user_usage", OwnerColumn: "user_id"},
	{Name: "entitlements", OwnerColumn: "user_id"},
}

// registerSQL is sign-up as cc_app does it, under the new user's own context: the user
// (email columns derived from the id, framed like crypto.Keyring output), its zeroed
// counters and its trial. GORM placeholders are ?.
var registerSQL = []string{
	`WITH v(id) AS (SELECT ?::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES (?)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES (?, 'trial', 'registration', now(), now() + interval '7 days')",
}

const (
	sqlstateReadOnlyTx            = "25006"
	sqlstateInsufficientPrivilege = "42501"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// env is one test's database with a TxManager logged in as cc_app.
type env struct {
	db  *sql.DB
	txm *pgdb.TxManager
}

func newEnv(t *testing.T, role pgtest.Role) env {
	t.Helper()
	return connectEnv(t, pgtest.NewDatabase(t), role)
}

func connectEnv(t *testing.T, database *pgtest.Database, role pgtest.Role) env {
	t.Helper()
	db := database.Connect(t, role)
	txm, err := pgdb.NewTxManager(db)
	require.NoError(t, err)
	return env{db: db, txm: txm}
}

func (e env) register(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		err := e.txm.ReadWrite(t.Context(), id, func(ctx context.Context) error {
			return execAll(ctx, registerSQL, id)
		})
		require.NoError(t, err, "register %s", id)
	}
}

func execAll(ctx context.Context, statements []string, arg any) error {
	conn, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	for _, stmt := range statements {
		if err := conn.Exec(stmt, arg).Error; err != nil {
			return err
		}
	}
	return nil
}

// countAll returns the rows of t visible to the running unit of work.
func countAll(ctx context.Context, t pgtest.Table) (int64, error) {
	conn, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	var n int64
	err = conn.Raw("SELECT count(*) FROM " + pgtest.QuoteIdent(t.Name)).Scan(&n).Error
	return n, err
}

// countOwned returns the rows of t owned by owner visible to the running unit of work.
func countOwned(ctx context.Context, t pgtest.Table, owner string) (int64, error) {
	conn, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	var n int64
	query := "SELECT count(*) FROM " + pgtest.QuoteIdent(t.Name) + " WHERE " + pgtest.QuoteIdent(t.OwnerColumn) + " = ?"
	err = conn.Raw(query, owner).Scan(&n).Error
	return n, err
}

// sessionUser returns app_user_id() of the running unit of work ("" when NULL).
func sessionUser(ctx context.Context) (string, error) {
	conn, err := pgdb.Conn(ctx)
	if err != nil {
		return "", err
	}
	var id sql.NullString
	err = conn.Raw("SELECT app_user_id()::text").Scan(&id).Error
	return id.String, err
}

func hasSQLState(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

func TestUserContextIsSetForTheTransaction(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	variants := map[string]func(context.Context, string, tx.Work) error{
		"ReadWrite": e.txm.ReadWrite,
		"ReadOnly":  e.txm.ReadOnly,
	}
	for name, run := range variants {
		t.Run(name, func(t *testing.T) {
			var got string
			var readOnly string
			err := run(t.Context(), "00000000-0000-4000-8000-00000000000A", func(ctx context.Context) error {
				conn, err := pgdb.Conn(ctx)
				if err != nil {
					return err
				}
				if err := conn.Raw("SHOW transaction_read_only").Scan(&readOnly).Error; err != nil {
					return err
				}
				got, err = sessionUser(ctx)
				return err
			})
			require.NoError(t, err)
			assert.Equal(t, userA, got, "canonical lower-case uuid")
			assert.Equal(t, map[bool]string{true: "on", false: "off"}[name == "ReadOnly"], readOnly)
		})
	}
}

func TestNoUserContextSeesNoRows(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	e.register(t, userA, userB)

	for _, table := range userTables {
		t.Run(table.Name+"/System", func(t *testing.T) {
			var n int64
			err := e.txm.System(t.Context(), func(ctx context.Context) error {
				user, err := sessionUser(ctx)
				if err != nil {
					return err
				}
				assert.Empty(t, user, "System has no user context")
				n, err = countAll(ctx, table)
				return err
			})
			require.NoError(t, err)
			assert.Zero(t, n)
		})
	}
}

// TestContextDiesWithTheTransaction pins a one-connection pool, so the query after
// each unit of work runs on the very session that unit of work used.
func TestContextDiesWithTheTransaction(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	e.db.SetMaxOpenConns(1)
	e.register(t, userA)
	boom := errors.New("boom")
	outcomes := map[string]func() error{
		"after commit":   func() error { return e.txm.ReadOnly(t.Context(), userA, func(context.Context) error { return nil }) },
		"after rollback": func() error { return e.txm.ReadWrite(t.Context(), userA, func(context.Context) error { return boom }) },
	}
	for name, run := range outcomes {
		t.Run(name, func(t *testing.T) {
			_ = run()
			assertPoolHasNoContext(t, e.db)
		})
	}
}

// assertPoolHasNoContext reads outside any unit of work: no user setting and no
// user-scoped row may be visible.
func assertPoolHasNoContext(t *testing.T, db *sql.DB) {
	t.Helper()
	var setting sql.NullString
	require.NoError(t, db.QueryRowContext(t.Context(),
		"SELECT current_setting('app.current_user_id', true)").Scan(&setting))
	assert.Empty(t, setting.String, "app.current_user_id outlived its transaction")
	for _, table := range userTables {
		var n int64
		require.NoError(t, db.QueryRowContext(t.Context(),
			"SELECT count(*) FROM "+pgtest.QuoteIdent(table.Name)).Scan(&n))
		assert.Zero(t, n, "%s visible outside a unit of work", table)
	}
}

func TestUserCannotSeeAnotherUser(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	e.register(t, userA, userB)

	for _, table := range userTables {
		t.Run(table.Name, func(t *testing.T) {
			var all, ofB int64
			err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
				var err error
				if all, err = countAll(ctx, table); err != nil {
					return err
				}
				ofB, err = countOwned(ctx, table, userB)
				return err
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), all, "A sees exactly its own row")
			assert.Zero(t, ofB, "A sees none of B's rows")
		})
	}
	t.Run("A cannot write for B", func(t *testing.T) {
		err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
			return execAll(ctx, registerSQL[:1], "00000000-0000-4000-8000-0000000000cc")
		})
		assert.True(t, hasSQLState(err, sqlstateInsufficientPrivilege), "RLS WITH CHECK refusal, got %v", err)
	})
}

func TestReadOnlyRefusesWrites(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		return execAll(ctx, registerSQL[:1], userA)
	})
	assert.True(t, hasSQLState(err, sqlstateReadOnlyTx), "got %v", err)
	assertNotRegistered(t, e, userA)
}

func assertNotRegistered(t *testing.T, e env, id string) {
	t.Helper()
	var n int64
	err := e.txm.ReadOnly(t.Context(), id, func(ctx context.Context) error {
		var err error
		n, err = countAll(ctx, userTables[0])
		return err
	})
	require.NoError(t, err)
	assert.Zero(t, n, "the unit of work must have been rolled back")
}

func TestOutcomes(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	writeThen := func(after func(ctx context.Context) error) tx.Work {
		return func(ctx context.Context) error {
			if err := execAll(ctx, registerSQL[:1], userA); err != nil {
				return err
			}
			return after(ctx)
		}
	}

	t.Run("error rolls back and is returned", func(t *testing.T) {
		boom := errors.New("boom")
		err := e.txm.ReadWrite(t.Context(), userA, writeThen(func(context.Context) error { return boom }))
		assert.ErrorIs(t, err, boom)
		assertNotRegistered(t, e, userA)
	})
	t.Run("panic rolls back and propagates", func(t *testing.T) {
		assert.PanicsWithValue(t, "boom", func() {
			_ = e.txm.ReadWrite(t.Context(), userA, writeThen(func(context.Context) error { panic("boom") }))
		})
		assertNotRegistered(t, e, userA)
	})
	t.Run("cancellation rolls back", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		err := e.txm.ReadWrite(ctx, userA, writeThen(func(context.Context) error {
			cancel()
			return nil
		}))
		assert.ErrorIs(t, err, context.Canceled)
		assertNotRegistered(t, e, userA)
	})
	t.Run("cancellation interrupts a running statement", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		err := e.txm.ReadWrite(ctx, userA, writeThen(func(ctx context.Context) error {
			conn, err := pgdb.Conn(ctx)
			if err != nil {
				return err
			}
			cancel()
			return conn.Exec("SELECT pg_sleep(5)").Error
		}))
		assert.Error(t, err)
		assertNotRegistered(t, e, userA)
	})
	t.Run("success commits", func(t *testing.T) {
		require.NoError(t, e.txm.ReadWrite(t.Context(), userA, writeThen(func(context.Context) error { return nil })))
		var n int64
		require.NoError(t, e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
			var err error
			n, err = countAll(ctx, userTables[0])
			return err
		}))
		assert.Equal(t, int64(1), n)
	})
}

func TestNestedUnitsOfWorkShareTheTransaction(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	txid := func(ctx context.Context) (int64, error) {
		conn, err := pgdb.Conn(ctx)
		if err != nil {
			return 0, err
		}
		var id int64
		err = conn.Raw("SELECT txid_current()").Scan(&id).Error
		return id, err
	}

	t.Run("inner sees outer's writes and rolls back with it", func(t *testing.T) {
		boom := errors.New("outer fails after inner succeeded")
		err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
			outerTx, err := txid(ctx)
			if err != nil {
				return err
			}
			if err := execAll(ctx, registerSQL[:1], userA); err != nil {
				return err
			}
			err = e.txm.ReadOnly(ctx, userA, func(inner context.Context) error {
				innerTx, err := txid(inner)
				if err != nil {
					return err
				}
				assert.Equal(t, outerTx, innerTx, "same transaction")
				n, err := countAll(inner, userTables[0])
				assert.Equal(t, int64(1), n, "inner sees the outer's uncommitted row")
				return err
			})
			require.NoError(t, err)
			return boom
		})
		assert.ErrorIs(t, err, boom)
		assertNotRegistered(t, e, userA)
	})
	t.Run("inner for another user is refused, outer context unchanged", func(t *testing.T) {
		err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
			innerErr := e.txm.ReadWrite(ctx, userB, func(context.Context) error {
				t.Error("must not run")
				return nil
			})
			assert.ErrorIs(t, innerErr, tx.ErrUserMismatch)
			user, err := sessionUser(ctx)
			assert.Equal(t, userA, user)
			return err
		})
		require.NoError(t, err)
	})
}

// TestNestedUnitOfWorkNeverJoinsAnotherManager: a process may hold one TxManager per
// role (cc_app for requests, cc_webhook or cc_worker on their own pools). A unit of
// work started through one of them inside another's must never run on the other's
// connection, where it would act with that role's grants believing it has its own.
func TestNestedUnitOfWorkNeverJoinsAnotherManager(t *testing.T) {
	database := pgtest.NewDatabase(t)
	app := connectEnv(t, database, pgtest.RoleApp)
	worker := connectEnv(t, database, pgtest.RoleWorker)
	asUserA := func(m *pgdb.TxManager) func(context.Context, tx.Work) error {
		return func(ctx context.Context, w tx.Work) error { return m.ReadWrite(ctx, userA, w) }
	}
	tests := []struct {
		name          string
		outer, nested func(context.Context, tx.Work) error
	}{
		{"system", worker.txm.System, app.txm.System},
		{"same user", asUserA(worker.txm), asUserA(app.txm)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.outer(t.Context(), func(ctx context.Context) error {
				return tt.nested(ctx, func(ctx context.Context) error {
					conn, err := pgdb.Conn(ctx)
					if err != nil {
						return err
					}
					var role string
					if err := conn.Raw("SELECT current_user").Scan(&role).Error; err != nil {
						return err
					}
					t.Errorf("cc_app work joined another manager's transaction (current_user %s)", role)
					return nil
				})
			})
			assert.ErrorIs(t, err, tx.ErrForeignTransaction)
		})
	}
}

// TestContextNeverLeaksAcrossPooledConnections runs many units of work for different
// users (and System) on a two-connection pool, sequentially and concurrently, as both
// cc_app and cc_worker. Every unit of work must see exactly its own context and rows.
func TestContextNeverLeaksAcrossPooledConnections(t *testing.T) {
	const poolSize = 2
	const rounds = 60
	e := newEnv(t, pgtest.RoleApp)
	e.db.SetMaxOpenConns(poolSize)
	e.db.SetMaxIdleConns(poolSize)
	users := []string{userA, userB, "00000000-0000-4000-8000-00000000000c"}
	e.register(t, users...)

	check := func(ctx context.Context, i int) error {
		user := ""
		if i%(len(users)+1) < len(users) {
			user = users[i%(len(users)+1)]
		}
		verify := func(ctx context.Context) error {
			got, err := sessionUser(ctx)
			if err != nil {
				return err
			}
			visible, err := countAll(ctx, userTables[0])
			if err != nil {
				return err
			}
			want := map[bool]int64{true: 0, false: 1}[user == ""]
			if got != user || visible != want {
				return fmt.Errorf("round %d as %q: saw context %q and %d user row(s), want %d", i, user, got, visible, want)
			}
			return nil
		}
		if user == "" {
			return e.txm.System(ctx, verify)
		}
		if i%2 == 0 {
			return e.txm.ReadOnly(ctx, user, verify)
		}
		return e.txm.ReadWrite(ctx, user, verify)
	}

	t.Run("sequential", func(t *testing.T) {
		for i := range rounds {
			require.NoError(t, check(t.Context(), i))
		}
	})
	t.Run("concurrent", func(t *testing.T) {
		g, ctx := errgroup.WithContext(t.Context())
		for i := range rounds {
			g.Go(func() error { return check(ctx, i) })
		}
		require.NoError(t, g.Wait())
	})
	t.Run("concurrent with failures in between", func(t *testing.T) {
		var wg sync.WaitGroup
		boom := errors.New("boom")
		for i := range rounds {
			wg.Go(func() {
				// Every other unit of work fails after setting its context: a rolled
				// back transaction must not leave its context behind either.
				err := e.txm.ReadWrite(t.Context(), users[i%len(users)], func(context.Context) error { return boom })
				assert.ErrorIs(t, err, boom)
				assert.NoError(t, check(t.Context(), i))
			})
		}
		wg.Wait()
	})
	t.Run("nothing left on any pooled connection", func(t *testing.T) {
		for range poolSize * 2 {
			assertPoolHasNoContext(t, e.db)
		}
	})
}

func TestSystemAsWorker(t *testing.T) {
	database := pgtest.NewDatabase(t)
	connectEnv(t, database, pgtest.RoleApp).register(t, userA)
	worker := connectEnv(t, database, pgtest.RoleWorker)

	err := worker.txm.System(t.Context(), func(ctx context.Context) error {
		conn, err := pgdb.Conn(ctx)
		if err != nil {
			return err
		}
		var flags int64
		if err := conn.Raw("SELECT count(*) FROM feature_flags").Scan(&flags).Error; err != nil {
			return err
		}
		assert.Positive(t, flags, "the worker reads the flags (skips jobs of disabled modules)")
		for _, table := range userTables {
			// Inside a savepoint, so a refusal does not abort the transaction.
			n, err := countInSavepoint(ctx, table)
			assert.True(t, n == 0 && (err == nil || hasSQLState(err, sqlstateInsufficientPrivilege)),
				"%s as cc_worker without a user: %d row(s), %v", table, n, err)
		}
		user, err := sessionUser(ctx)
		assert.Empty(t, user)
		return err
	})
	require.NoError(t, err)
}

func countInSavepoint(ctx context.Context, table pgtest.Table) (int64, error) {
	conn, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	if err := conn.SavePoint("probe").Error; err != nil {
		return 0, err
	}
	n, err := countAll(ctx, table)
	if rbErr := conn.RollbackTo("probe").Error; rbErr != nil {
		return n, errors.Join(err, rbErr)
	}
	return n, err
}

func TestCheckRuntimeRole(t *testing.T) {
	database := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role)+" is a plain runtime login", func(t *testing.T) {
			assert.NoError(t, pgdb.CheckRuntimeRole(t.Context(), database.Connect(t, role)))
		})
	}
	t.Run("the schema owner is refused", func(t *testing.T) {
		err := pgdb.CheckRuntimeRole(t.Context(), database.Connect(t, pgtest.RoleOwner))
		require.ErrorIs(t, err, pgdb.ErrPrivilegedRole)
		assert.Contains(t, err.Error(), "owns")
	})
}

func TestPingerReachesDatabase(t *testing.T) {
	e := newEnv(t, pgtest.RoleApp)
	assert.NoError(t, pgdb.NewPinger(e.db).Ping(t.Context()))
}
