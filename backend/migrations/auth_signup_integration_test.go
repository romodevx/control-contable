//go:build integration

package migrations_test

import (
	"database/sql"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Auth sign-up migration (M1-05): email_verification_start / _resend (D7 and the shared
// 60 s cooldown), code_sent_at, the upsert revocation and the nullable users.name.

const (
	outcomeStarted   = "started"
	outcomeResent    = "resent"
	outcomeUnchanged = "unchanged"

	callStart  = "SELECT email_verification_start($1, $2, $3::jsonb, now() + interval '15 minutes')"
	callResend = "SELECT email_verification_resend($1, $2, now() + interval '15 minutes')"

	// pastCooldown moves the last code out of the 60 s cooldown.
	pastCooldown = "UPDATE email_verifications SET code_sent_at = now() - interval '61 seconds', created_at = now() - interval '61 seconds'"
	// expireAll makes every pending sign-up expired.
	expireAll = "UPDATE email_verifications SET created_at = now() - interval '1 hour', code_sent_at = now() - interval '1 hour', expires_at = now() - interval '1 second'"
)

func start(t *testing.T, app *sql.DB, emailHash, codeHash []byte, pending string) string {
	t.Helper()
	var outcome string
	require.NoError(t, app.QueryRowContext(t.Context(), callStart, emailHash, codeHash, pending).Scan(&outcome))
	return outcome
}

func resend(t *testing.T, app *sql.DB, emailHash, codeHash []byte) string {
	t.Helper()
	var outcome string
	require.NoError(t, app.QueryRowContext(t.Context(), callResend, emailHash, codeHash).Scan(&outcome))
	return outcome
}

func ownerExec(t *testing.T, owner *sql.DB, query string) {
	t.Helper()
	_, err := owner.ExecContext(t.Context(), query)
	require.NoError(t, err)
}

func TestEmailVerificationStart_NewSignUpStarts(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("111111")

	assert.Equal(t, outcomeStarted, start(t, app, email, code, `{"owner": "ana"}`))

	got := consume(t, app, email, code)
	assert.Equal(t, outcomeVerified, got.outcome)
	assert.JSONEq(t, `{"owner": "ana"}`, got.pending.String)
}

// D7: whoever registers second for a live pending sign-up cannot touch its payload.
func TestEmailVerificationStart_LivePendingIsNeverOverwritten(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app, owner := db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleOwner)
	email, first, second := blindIndex("ana@x"), blindIndex("111111"), blindIndex("222222")
	require.Equal(t, outcomeStarted, start(t, app, email, first, `{"owner": "ana"}`))

	t.Run("inside the cooldown nothing changes", func(t *testing.T) {
		assert.Equal(t, outcomeUnchanged, start(t, app, email, second, `{"owner": "attacker"}`))
	})

	t.Run("after the cooldown it acts as a resend", func(t *testing.T) {
		require.Equal(t, outcomeInvalid, consume(t, app, email, blindIndex("bad")).outcome, "spend an attempt")
		ownerExec(t, owner, pastCooldown)

		assert.Equal(t, outcomeResent, start(t, app, email, second, `{"owner": "attacker"}`))

		var attempts int
		require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT attempts FROM email_verifications").Scan(&attempts))
		assert.Zero(t, attempts, "a resend resets the per-code budget")
		assert.Equal(t, outcomeInvalid, consume(t, app, email, first).outcome, "the old code is replaced")
		got := consume(t, app, email, second)
		require.Equal(t, outcomeVerified, got.outcome)
		assert.JSONEq(t, `{"owner": "ana"}`, got.pending.String, "the payload is the first registrant's")
	})
}

func TestEmailVerificationStart_ExpiredPendingIsReplaced(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app, owner := db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleOwner)
	email, code := blindIndex("ana@x"), blindIndex("222222")
	require.Equal(t, outcomeStarted, start(t, app, email, blindIndex("111111"), `{"owner": "old"}`))
	ownerExec(t, owner, expireAll)

	assert.Equal(t, outcomeStarted, start(t, app, email, code, `{"owner": "new"}`))

	got := consume(t, app, email, code)
	require.Equal(t, outcomeVerified, got.outcome)
	assert.JSONEq(t, `{"owner": "new"}`, got.pending.String)
}

// Concurrent first registrations: exactly one payload wins and nobody replaces it.
func TestEmailVerificationStart_ConcurrentFirstCallsKeepOnePayload(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	email := blindIndex("race@x")
	const racers = 8

	outcomes := make([]string, racers)
	var wg sync.WaitGroup
	for i := range racers {
		wg.Go(func() {
			var outcome string
			err := app.QueryRowContext(t.Context(), callStart, email, blindIndex(strconv.Itoa(i)), `{"racer": `+strconv.Itoa(i)+`}`).Scan(&outcome)
			assert.NoError(t, err)
			outcomes[i] = outcome
		})
	}
	wg.Wait()

	winner := -1
	for i, o := range outcomes {
		if o == outcomeStarted {
			assert.Equal(t, -1, winner, "only one call starts the sign-up")
			winner = i
			continue
		}
		assert.Equal(t, outcomeUnchanged, o, "the others hit the cooldown")
	}
	require.GreaterOrEqual(t, winner, 0)
	got := consume(t, app, email, blindIndex(strconv.Itoa(winner)))
	require.Equal(t, outcomeVerified, got.outcome)
	assert.JSONEq(t, `{"racer": `+strconv.Itoa(winner)+`}`, got.pending.String)
}

func TestEmailVerificationResend_OnlyLivePendingSignUpsOutsideTheCooldown(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app, owner := db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleOwner)
	email, first, second := blindIndex("ana@x"), blindIndex("111111"), blindIndex("222222")

	assert.Equal(t, outcomeUnchanged, resend(t, app, email, first), "no pending sign-up")

	require.Equal(t, outcomeStarted, start(t, app, email, first, `{"owner": "ana"}`))
	assert.Equal(t, outcomeUnchanged, resend(t, app, email, second), "inside the cooldown")
	assert.Equal(t, outcomeVerified, consume(t, app, email, first).outcome, "the first code still works")

	require.Equal(t, outcomeStarted, start(t, app, email, first, `{"owner": "ana"}`))
	ownerExec(t, owner, pastCooldown)
	assert.Equal(t, outcomeResent, resend(t, app, email, second))
	assert.Equal(t, outcomeUnchanged, resend(t, app, email, first), "the resend restarted the cooldown")
	got := consume(t, app, email, second)
	require.Equal(t, outcomeVerified, got.outcome)
	assert.JSONEq(t, `{"owner": "ana"}`, got.pending.String)

	require.Equal(t, outcomeStarted, start(t, app, email, first, `{}`))
	ownerExec(t, owner, expireAll)
	assert.Equal(t, outcomeUnchanged, resend(t, app, email, second), "an expired sign-up must register again")
}

func TestEmailVerificationStartAndResend_RejectBadArguments(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("123456")
	bad := []struct {
		name  string
		query string
		args  []any
	}{
		{"start: email hash not 32 bytes", callStart, []any{[]byte("a"), code, `{}`}},
		{"start: code hash not 32 bytes", callStart, []any{email, []byte("123456"), `{}`}},
		{"start: pending not an object", callStart, []any{email, code, `[]`}},
		{"start: expiry beyond one hour", "SELECT email_verification_start($1, $2, '{}', now() + interval '61 minutes')", []any{email, code}},
		{"resend: code hash not 32 bytes", callResend, []any{email, []byte("123456")}},
		{"resend: expiry in the past", "SELECT email_verification_resend($1, $2, now() - interval '1 second')", []any{email, code}},
	}
	for _, b := range bad {
		t.Run(b.name, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), b.query, b.args...)
			requireSQLState(t, err, sqlstateInvalidParameterValue)
		})
	}
}

// D7 defense in depth: the API role cannot replace a pending payload any more.
func TestEmailVerificationUpsert_NoLongerExecutableByTheApp(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)

	_, err := app.ExecContext(t.Context(), callUpsert, blindIndex("ana@x"), blindIndex("1"), `{}`, "15 minutes")

	requireSQLState(t, err, sqlstateInsufficientPrivilege)
}

func TestAuthCheckCodeArgs_NotExecutableByRuntimeRoles(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			_, err := db.Connect(t, role).ExecContext(t.Context(),
				"SELECT auth_check_code_args($1, now() + interval '1 minute')", blindIndex("x"))
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

func TestUsers_NameIsOptionalButNeverBlank(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	insertNamed := func(userID string, name any) error {
		_, err := execAsUser(t, app, userID, `
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')),
        '\x0001'::bytea || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8')),
        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', $2)`, userID, name)
		return err
	}

	require.NoError(t, insertNamed(userB, nil), "no name")
	requireSQLState(t, insertNamed(userC, "   "), sqlstateCheckViolation)
}
