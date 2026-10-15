//go:build integration

package migrations_test

import (
	"bytes"
	"context"
	"database/sql"
	"slices"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Auth email outbox (M1-06 carry-over): register and resendVerification enqueue their
// emails in the request's own transaction and a worker job delivers them, so the
// response time no longer says whether an email holds an account. The table is
// service-only: cc_app reaches it only through auth_email_outbox_enqueue, cc_worker
// only through claim / finish / retry, all SECURITY DEFINER functions owned by
// cc_auth_definer; no runtime role reads or writes it directly.

const (
	authEmailOutboxVersion = 20261023100000

	purgeTargetEmailOutbox = "auth_email_outbox"

	outboxEnqueueSQL = "SELECT auth_email_outbox_enqueue($1, $2, now() + $3::interval)"
	outboxClaimSQL   = "SELECT id::text, payload, attempts, claim_token::text FROM auth_email_outbox_claim($1, $2)"
	outboxFinishSQL  = "SELECT auth_email_outbox_finish($1, $2)"
	outboxRetrySQL   = "SELECT auth_email_outbox_retry($1, $2, $3)"

	// outboxLease is a lease far longer than any test, in seconds.
	outboxLease = 600
	// outboxMaxAttempts mirrors the cap inside auth_email_outbox_claim.
	outboxMaxAttempts = 5
	// codeLife is how long a queued email stays deliverable (the code's 15 minutes).
	codeLife = "15 minutes"
)

// sealedSample stands in for an AES-GCM sealed payload: opaque bytes to the database.
var sealedSample = bytes.Repeat([]byte{0x5a}, 64)

// claimedRow is one row of auth_email_outbox_claim.
type claimedRow struct {
	id, token string
	payload   []byte
	attempts  int
}

func enqueueEmail(t *testing.T, app *sql.DB) string {
	t.Helper()
	id := uuid.NewString()
	_, err := app.ExecContext(t.Context(), outboxEnqueueSQL, id, sealedSample, codeLife)
	require.NoError(t, err)
	return id
}

func claimEmails(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, limit int,
) ([]claimedRow, error) {
	rows, err := q.QueryContext(ctx, outboxClaimSQL, limit, outboxLease)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []claimedRow
	for rows.Next() {
		var r claimedRow
		if err := rows.Scan(&r.id, &r.payload, &r.attempts, &r.token); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func mustClaim(t *testing.T, worker *sql.DB, limit int) []claimedRow {
	t.Helper()
	rows, err := claimEmails(t.Context(), worker, limit)
	require.NoError(t, err)
	return rows
}

func callBool(t *testing.T, db *sql.DB, query string, args ...any) bool {
	t.Helper()
	var ok bool
	require.NoError(t, db.QueryRowContext(t.Context(), query, args...).Scan(&ok))
	return ok
}

// outboxFixture runs query as the owner with FORCE RLS lifted for that transaction
// only (the owner has no policy of its own) and returns the rows it touched.
func outboxFixture(t *testing.T, owner *sql.DB, query string, args ...any) int64 {
	t.Helper()
	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE auth_email_outbox NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	res, err := tx.ExecContext(t.Context(), query, args...)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE auth_email_outbox FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	var forced bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT relforcerowsecurity FROM pg_class WHERE relname = 'auth_email_outbox'").Scan(&forced))
	require.True(t, forced, "FORCE RLS restored")
	return n
}

func outboxCount(t *testing.T, owner *sql.DB) int {
	t.Helper()
	tx, err := owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "ALTER TABLE auth_email_outbox NO FORCE ROW LEVEL SECURITY")
	require.NoError(t, err)
	var n int
	require.NoError(t, tx.QueryRowContext(t.Context(), "SELECT count(*) FROM auth_email_outbox").Scan(&n))
	return n // rolled back: FORCE stays as it was
}

func TestAuthEmailOutbox_RLSEnabledAndForced(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	var enabled, forced bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT relrowsecurity, relforcerowsecurity FROM pg_class WHERE relname = 'auth_email_outbox'").Scan(&enabled, &forced))
	assert.True(t, enabled)
	assert.True(t, forced)
}

// Service-only: no runtime role touches the table directly, and even its owner sees no
// row under FORCE RLS (no owner policy). Only cc_auth_definer has policies.
func TestAuthEmailOutbox_NoDirectAccess(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	owner := db.Connect(t, pgtest.RoleOwner)
	enqueueEmail(t, app)

	statements := []string{
		"SELECT id FROM auth_email_outbox",
		"INSERT INTO auth_email_outbox (id, payload, expires_at) VALUES (gen_random_uuid(), '\\x00', now() + interval '1 minute')",
		"UPDATE auth_email_outbox SET attempts = 0",
		"DELETE FROM auth_email_outbox",
	}
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		conn := db.Connect(t, role)
		for _, stmt := range statements {
			_, err := conn.ExecContext(t.Context(), stmt)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		}
	}

	var visible int
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT count(*) FROM auth_email_outbox").Scan(&visible))
	assert.Zero(t, visible, "FORCE RLS: the owner has no policy")
	assert.Equal(t, 1, outboxCount(t, owner), "the row is there")

	assert.Equal(t, []string{authDefinerRole, workerDefinerRole}, queryStrings(t, owner, `
SELECT DISTINCT r::text FROM pg_policies, unnest(roles) AS r
WHERE tablename = 'auth_email_outbox' ORDER BY 1`), "only the definers have policies")
}

// Each runtime role runs only its side: cc_app enqueues, cc_worker delivers.
func TestAuthEmailOutbox_FunctionGrants(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	grantees := func(signature string) []string {
		return queryStrings(t, owner, `
SELECT r.rolname::text FROM pg_roles r
WHERE r.rolname LIKE 'cc\_%' AND r.rolname NOT IN ('cc_owner', 'cc_auth_definer')
  AND has_function_privilege(r.oid, $1, 'EXECUTE') ORDER BY 1`, signature)
	}
	assert.Equal(t, []string{"cc_app"}, grantees("auth_email_outbox_enqueue(uuid, bytea, timestamptz)"))
	for _, fn := range []string{
		"auth_email_outbox_claim(integer, integer)",
		"auth_email_outbox_finish(uuid, uuid)",
		"auth_email_outbox_retry(uuid, uuid, integer)",
	} {
		assert.Equal(t, []string{"cc_worker"}, grantees(fn), fn)
	}
	assert.Equal(t, []string{
		"auth_email_outbox_claim " + authDefinerRole + " true",
		"auth_email_outbox_enqueue " + authDefinerRole + " true",
		"auth_email_outbox_finish " + authDefinerRole + " true",
		"auth_email_outbox_retry " + authDefinerRole + " true",
	}, queryStrings(t, owner, `
SELECT p.proname || ' ' || pg_get_userbyid(p.proowner) || ' ' || p.prosecdef
FROM pg_proc p WHERE p.proname LIKE 'auth\_email\_outbox\_%' ORDER BY 1`))
}

func TestAuthEmailOutbox_EnqueueValidates(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)

	tests := []struct {
		name     string
		payload  []byte
		lifetime string
	}{
		{"already expired", sealedSample, "-1 second"},
		{"more than an hour", sealedSample, "61 minutes"},
		{"empty payload", []byte{}, codeLife},
		{"oversized payload", bytes.Repeat([]byte{1}, 4097), codeLife},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), outboxEnqueueSQL, uuid.NewString(), tt.payload, tt.lifetime)
			requireSQLState(t, err, sqlstateInvalidParameterValue)
		})
	}
}

// A claim leases the due rows to one caller: attempts counted, a fresh claim token,
// hidden from every other claim until the lease ends or the row is retried.
func TestAuthEmailOutbox_ClaimFinishRetry(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	first, second := enqueueEmail(t, app), enqueueEmail(t, app)

	claimed := mustClaim(t, worker, 10)
	require.Len(t, claimed, 2)
	assert.ElementsMatch(t, []string{first, second}, []string{claimed[0].id, claimed[1].id})
	for _, c := range claimed {
		assert.Equal(t, sealedSample, c.payload)
		assert.Equal(t, 1, c.attempts)
		assert.NotEmpty(t, c.token)
	}
	assert.Empty(t, mustClaim(t, worker, 10), "leased rows are not claimed twice")

	byID := map[string]claimedRow{claimed[0].id: claimed[0], claimed[1].id: claimed[1]}
	stale := uuid.NewString()
	assert.False(t, callBool(t, worker, outboxFinishSQL, first, stale), "only the claim's holder finishes it")
	assert.True(t, callBool(t, worker, outboxFinishSQL, first, byID[first].token))
	assert.False(t, callBool(t, worker, outboxFinishSQL, first, byID[first].token), "finished once")

	assert.False(t, callBool(t, worker, outboxRetrySQL, second, stale, 1))
	assert.True(t, callBool(t, worker, outboxRetrySQL, second, byID[second].token, 1))
	assert.False(t, callBool(t, worker, outboxFinishSQL, second, byID[second].token), "a retry voids the claim")
	outboxFixture(t, owner, "UPDATE auth_email_outbox SET next_attempt_at = now() - interval '1 second'")
	again := mustClaim(t, worker, 10)
	require.Len(t, again, 1)
	assert.Equal(t, second, again[0].id)
	assert.Equal(t, 2, again[0].attempts)
	assert.NotEqual(t, byID[second].token, again[0].token)
	assert.Equal(t, 1, outboxCount(t, owner), "the finished row is gone")
}

func TestAuthEmailOutbox_ExpiredLeaseIsClaimedAgain(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	id := enqueueEmail(t, app)
	held := mustClaim(t, worker, 1)
	require.Len(t, held, 1)

	outboxFixture(t, owner, "UPDATE auth_email_outbox SET next_attempt_at = now() - interval '1 second'")
	taken := mustClaim(t, worker, 1)

	require.Len(t, taken, 1)
	assert.Equal(t, id, taken[0].id)
	assert.False(t, callBool(t, worker, outboxFinishSQL, id, held[0].token), "the expired holder can no longer finish it")
	assert.True(t, callBool(t, worker, outboxFinishSQL, id, taken[0].token))
}

// Retries are bounded (the use case drops a row on its last attempt; the database
// never hands out more), and an expired email is never delivered: the claim deletes it.
func TestAuthEmailOutbox_AttemptCapAndExpiry(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	exhausted, expired := enqueueEmail(t, app), enqueueEmail(t, app)
	outboxFixture(t, owner, "UPDATE auth_email_outbox SET attempts = $2 WHERE id = $1", exhausted, outboxMaxAttempts)
	outboxFixture(t, owner, `UPDATE auth_email_outbox SET created_at = now() - interval '20 minutes',
  expires_at = now() - interval '1 second' WHERE id = $1`, expired)

	assert.Empty(t, mustClaim(t, worker, 10))
	assert.Equal(t, 1, outboxCount(t, owner), "the expired row was deleted, the exhausted one waits for its expiry")
}

func TestAuthEmailOutbox_ClaimArgumentsAreBounded(t *testing.T) {
	worker := pgtest.NewDatabase(t).Connect(t, pgtest.RoleWorker)

	for _, args := range [][2]int{{0, outboxLease}, {101, outboxLease}, {1, 9}, {1, 901}} {
		_, err := worker.ExecContext(t.Context(), "SELECT * FROM auth_email_outbox_claim($1, $2)", args[0], args[1])
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	}
	_, err := worker.ExecContext(t.Context(), outboxRetrySQL, uuid.NewString(), uuid.NewString(), 0)
	requireSQLState(t, err, sqlstateInvalidParameterValue)
}

// Worker split brain: two instances claiming at once never get the same row. A claim
// in an open transaction makes the other skip its rows (SKIP LOCKED) instead of waiting
// on them, and after the commit the lease keeps them away.
func TestAuthEmailOutbox_TwoWorkersNeverClaimTheSameRow(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	workerA, workerB := db.Connect(t, pgtest.RoleWorker), db.Connect(t, pgtest.RoleWorker)
	const queued = 30
	for range queued {
		enqueueEmail(t, app)
	}

	t.Run("an open claim is skipped, not waited on", func(t *testing.T) {
		txA, err := workerA.BeginTx(t.Context(), nil)
		require.NoError(t, err)
		defer func() { _ = txA.Rollback() }()
		heldByA, err := claimEmails(t.Context(), txA, 5)
		require.NoError(t, err)
		require.Len(t, heldByA, 5)

		txB, err := workerB.BeginTx(t.Context(), nil)
		require.NoError(t, err)
		defer func() { _ = txB.Rollback() }()
		// Waiting on A's locks would fail here instead of hanging the suite.
		_, err = txB.ExecContext(t.Context(), "SET LOCAL lock_timeout = '2s'")
		require.NoError(t, err)
		fromB, err := claimEmails(t.Context(), txB, queued)
		require.NoError(t, err, "B skips A's rows instead of waiting on them")
		require.NoError(t, txB.Commit())

		assert.Len(t, fromB, queued-5)
		for _, b := range fromB {
			assert.False(t, slices.ContainsFunc(heldByA, func(a claimedRow) bool { return a.id == b.id }))
		}
		require.NoError(t, txA.Rollback())
	})

	t.Run("concurrent claims partition the queue", func(t *testing.T) {
		outboxFixture(t, db.Connect(t, pgtest.RoleOwner),
			"UPDATE auth_email_outbox SET next_attempt_at = now() - interval '1 second', claim_token = NULL")
		var (
			mu  sync.Mutex
			ids []string
			wg  sync.WaitGroup
		)
		for _, w := range []*sql.DB{workerA, workerB, workerA, workerB} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					rows, err := claimEmails(context.Background(), w, 3)
					if !assert.NoError(t, err) || len(rows) == 0 {
						return
					}
					mu.Lock()
					for _, r := range rows {
						ids = append(ids, r.id)
					}
					mu.Unlock()
				}
			}()
		}
		wg.Wait()

		assert.Len(t, ids, queued)
		slices.Sort(ids)
		assert.Len(t, slices.Compact(ids), queued, "no row claimed twice")
	})
}

// The platform purge (M9-08) deletes expired outbox rows, whatever their state.
func TestPurgeSoftDeleted_PurgesExpiredOutboxEmails(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	worker := db.Connect(t, pgtest.RoleWorker)
	owner := db.Connect(t, pgtest.RoleOwner)
	expired, live := enqueueEmail(t, app), enqueueEmail(t, app)
	outboxFixture(t, owner, `UPDATE auth_email_outbox SET created_at = now() - interval '20 minutes',
  expires_at = now() - interval '1 second' WHERE id = $1`, expired)

	counts := purge(t, worker, purgeBatch)

	assert.Equal(t, 1, counts[purgeTargetEmailOutbox])
	assert.Equal(t, 1, outboxCount(t, owner))
	claimed := mustClaim(t, worker, 10)
	require.Len(t, claimed, 1)
	assert.Equal(t, live, claimed[0].id)
	assert.Zero(t, purge(t, worker, purgeBatch)[purgeTargetEmailOutbox])
}

func TestAuthEmailOutboxMigration_Down(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), authEmailOutboxVersion-1)
	require.NoError(t, err)
	worker := db.Connect(t, pgtest.RoleWorker)
	assert.Equal(t, purgeTargets[:len(purgeTargets)-1], queryStrings(t, worker, "SELECT target FROM purge_soft_deleted(1)"))
	assert.Empty(t, queryStrings(t, owner, "SELECT relname::text FROM pg_class WHERE relname = 'auth_email_outbox'"))
	assert.Empty(t, queryStrings(t, owner, "SELECT proname::text FROM pg_proc WHERE proname LIKE 'auth\\_email\\_outbox\\_%'"))

	require.NoError(t, migrations.Up(t.Context(), owner))
	assert.Equal(t, purgeTargets, queryStrings(t, worker, "SELECT target FROM purge_soft_deleted(1)"))
}
