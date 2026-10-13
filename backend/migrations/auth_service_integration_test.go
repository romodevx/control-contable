//go:build integration

package migrations_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Auth service migration (M1-03): legal_documents, trial_registry and the SECURITY
// DEFINER functions owned by cc_auth_definer.

const (
	definerRole = "cc_auth_definer"

	sqlstateInvalidParameterValue = "22023"

	// maxVerificationAttempts mirrors the constant inside email_verification_consume.
	maxVerificationAttempts = 5

	outcomeVerified = "verified"
	outcomeInvalid  = "invalid"
	outcomeExpired  = "expired"
	outcomeLocked   = "locked"

	callLookup  = "SELECT id::text, password_hash, status FROM auth_lookup_credentials($1)"
	callTrial   = "SELECT trial_registry_check_and_insert($1)"
	callUpsert  = "SELECT email_verification_upsert($1, $2, $3::jsonb, now() + $4::interval)"
	callConsume = "SELECT outcome, pending::text FROM email_verification_consume($1, $2)"
)

// authFunctions are every definer function of the migration, by signature.
var authFunctions = []string{
	"auth_lookup_credentials(bytea)",
	"email_verification_consume(bytea,bytea)",
	"email_verification_upsert(bytea,bytea,jsonb,timestamp with time zone)",
	"trial_registry_check_and_insert(bytea)",
}

// blindIndex stands in for the app-side HMAC: any 32-byte digest fits the columns.
func blindIndex(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

// userEmailHash is the email_hash insertUserSQL derives from a user id.
func userEmailHash(userID string) []byte { return blindIndex(userID) }

type consumeResult struct {
	outcome string
	pending sql.NullString
}

// consume calls email_verification_consume in its own committed transaction, as the
// app must: a failed attempt only counts once it is committed.
func consume(t *testing.T, app *sql.DB, emailHash, codeHash []byte) consumeResult {
	t.Helper()
	var r consumeResult
	require.NoError(t, app.QueryRowContext(t.Context(), callConsume, emailHash, codeHash).Scan(&r.outcome, &r.pending))
	return r
}

func upsertVerification(t *testing.T, app *sql.DB, emailHash, codeHash []byte, pending, ttl string) {
	t.Helper()
	_, err := app.ExecContext(t.Context(), callUpsert, emailHash, codeHash, pending, ttl)
	require.NoError(t, err)
}

func checkTrial(t *testing.T, app *sql.DB, emailHash []byte) bool {
	t.Helper()
	var first bool
	require.NoError(t, app.QueryRowContext(t.Context(), callTrial, emailHash).Scan(&first))
	return first
}

// The definer role holds the cross-user reach, so it must be unable to log in, bypass
// RLS or be reached by any runtime role; cc_owner may SET ROLE to it (the migration
// transfers ownership) but does not inherit its privileges or policies.
func TestAuthDefinerRole_Posture(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)

	var login, super, bypass, createRole, inherit bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT rolcanlogin, rolsuper, rolbypassrls, rolcreaterole, rolinherit FROM pg_roles WHERE rolname = $1",
		definerRole).Scan(&login, &super, &bypass, &createRole, &inherit))
	assert.False(t, login, "NOLOGIN")
	assert.False(t, super, "NOSUPERUSER")
	assert.False(t, bypass, "NOBYPASSRLS")
	assert.False(t, createRole, "NOCREATEROLE")
	assert.False(t, inherit, "NOINHERIT")

	members := queryStrings(t, owner, `
SELECT pg_get_userbyid(m.member) || ' inherit=' || m.inherit_option || ' set=' || m.set_option || ' admin=' || m.admin_option
FROM pg_auth_members m WHERE m.roleid = 'cc_auth_definer'::regrole ORDER BY 1`)
	assert.Equal(t, []string{"cc_owner inherit=false set=true admin=false"}, members)

	var schemaCreate bool
	require.NoError(t, owner.QueryRowContext(t.Context(),
		"SELECT has_schema_privilege($1, 'public', 'CREATE')", definerRole).Scan(&schemaCreate))
	assert.False(t, schemaCreate, "CREATE on public is lent for the ownership transfer only")
}

// Each function is SECURITY DEFINER, owned by the definer role, pins a search_path that
// ends in pg_temp and is executable by cc_app alone.
func TestAuthFunctions_DefinerOwnedPinnedAndAppOnly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)

	for _, fn := range authFunctions {
		t.Run(fn, func(t *testing.T) {
			var fnOwner string
			var definer bool
			var config []byte
			require.NoError(t, owner.QueryRowContext(t.Context(), `
SELECT pg_get_userbyid(proowner), prosecdef, array_to_json(proconfig)
FROM pg_proc WHERE oid = $1::regprocedure`, fn).Scan(&fnOwner, &definer, &config))
			assert.Equal(t, definerRole, fnOwner)
			assert.True(t, definer, "SECURITY DEFINER")
			var settings []string
			require.NoError(t, json.Unmarshal(config, &settings))
			assert.Equal(t, []string{"search_path=pg_catalog, public, pg_temp"}, settings)

			grantees := queryStrings(t, owner, `
SELECT CASE a.grantee WHEN 0 THEN 'PUBLIC' ELSE pg_get_userbyid(a.grantee) END
FROM pg_proc p CROSS JOIN LATERAL aclexplode(p.proacl) a
WHERE p.oid = $1::regprocedure AND a.privilege_type = 'EXECUTE' ORDER BY 1`, fn)
			assert.Equal(t, []string{"cc_app", definerRole}, grantees, "cc_app and the owner itself")
		})
	}

	for _, role := range []pgtest.Role{pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role)+" cannot execute", func(t *testing.T) {
			_, err := db.Connect(t, role).ExecContext(t.Context(), callTrial, blindIndex("x"))
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}
}

// The functions are the ONLY path: cc_app reaches none of the service tables, and sees
// no other user's credentials, directly.
func TestAuthServiceTables_AppCannotReachThemDirectly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)
	require.True(t, checkTrial(t, app, userEmailHash(userA)))
	upsertVerification(t, app, blindIndex("pending@x"), blindIndex("123456"), `{"name":"P"}`, "15 minutes")

	denied := []string{
		"SELECT 1 FROM trial_registry",
		"INSERT INTO trial_registry (email_hash) VALUES (sha256('z'))",
		"DELETE FROM trial_registry",
		"SELECT 1 FROM email_verifications",
		"UPDATE email_verifications SET attempts = 0",
		"INSERT INTO legal_documents (kind, version, url, effective_at) VALUES ('terms', 'v9', 'https://x.invalid/t', now())",
		"UPDATE legal_documents SET url = 'https://evil.invalid'",
		"DELETE FROM legal_documents",
	}
	for _, q := range denied {
		t.Run(q, func(t *testing.T) {
			_, err := execAsUser(t, app, userA, q)
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}

	t.Run("users of others stay invisible", func(t *testing.T) {
		var seen int
		require.NoError(t, pgtest.AsUser(t.Context(), app, userA, func(tx *sql.Tx) error {
			return tx.QueryRowContext(t.Context(), "SELECT count(*) FROM users WHERE id = $1", userB).Scan(&seen)
		}))
		assert.Zero(t, seen)
	})

	t.Run("the definer policies do not leak to cc_owner", func(t *testing.T) {
		var seen int
		require.NoError(t, db.Connect(t, pgtest.RoleOwner).QueryRowContext(t.Context(),
			"SELECT count(*) FROM users").Scan(&seen))
		assert.Zero(t, seen, "FORCE RLS still binds the owner; it does not inherit cc_auth_definer")
	})
}

func TestAuthLookupCredentials_ReturnsOnlyTheMatchingUsersMinimalColumns(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA, userB)

	var id, phc, status string
	require.NoError(t, app.QueryRowContext(t.Context(), callLookup, userEmailHash(userA)).Scan(&id, &phc, &status))
	assert.Equal(t, userA, id)
	assert.Contains(t, phc, "$argon2id$")
	assert.Equal(t, "active", status)

	var outCols string
	require.NoError(t, app.QueryRowContext(t.Context(),
		"SELECT array_to_string(proargnames[2:], ',') FROM pg_proc WHERE oid = 'auth_lookup_credentials(bytea)'::regprocedure").Scan(&outCols))
	assert.Equal(t, "id,password_hash,status", outCols, "no email ciphertext, name or other column")

	t.Run("unknown email returns no row", func(t *testing.T) {
		rows := queryStrings(t, app, "SELECT id::text FROM auth_lookup_credentials(sha256('nobody'))")
		assert.Empty(t, rows)
	})
	t.Run("malformed hash is refused", func(t *testing.T) {
		_, err := app.ExecContext(t.Context(), callLookup, []byte{1, 2, 3})
		requireSQLState(t, err, sqlstateInvalidParameterValue)
		_, err = app.ExecContext(t.Context(), "SELECT * FROM auth_lookup_credentials(NULL)")
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	})
	t.Run("the scope GUC does not outlive the call", func(t *testing.T) {
		var leftover sql.NullString
		require.NoError(t, pgtest.WithoutUser(t.Context(), app, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(t.Context(), callLookup, userEmailHash(userA)); err != nil {
				return err
			}
			return tx.QueryRowContext(t.Context(),
				"SELECT nullif(current_setting('app.auth_email_hash', true), '')").Scan(&leftover)
		}))
		assert.False(t, leftover.Valid)
	})
}

// A caller that can create temp objects (a mistaken TEMP grant) shadows users with a
// fake row and puts pg_temp first: the pinned search_path still resolves public.users.
func TestAuthLookupCredentials_SearchPathHijackFails(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	_, err := db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(),
		"GRANT TEMP ON DATABASE "+pgtest.QuoteIdent(db.Name)+" TO cc_app")
	require.NoError(t, err)

	conn, err := app.Conn(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	for _, stmt := range []string{
		"CREATE TEMP TABLE users (id uuid, email_hash bytea, password_hash text, status text)",
		"INSERT INTO pg_temp.users VALUES ('00000000-0000-4000-8000-0000000000ff', sha256(convert_to('" + userA + "', 'UTF8')), '$argon2id$attacker', 'active')",
		"CREATE TEMP TABLE trial_registry (email_hash bytea PRIMARY KEY)",
		"SET search_path = pg_temp, public",
	} {
		_, err := conn.ExecContext(t.Context(), stmt)
		require.NoError(t, err, stmt)
	}

	var id, phc, status string
	require.NoError(t, conn.QueryRowContext(t.Context(), callLookup, userEmailHash(userA)).Scan(&id, &phc, &status))
	assert.Equal(t, userA, id, "the real row, not the shadow")
	assert.NotEqual(t, "$argon2id$attacker", phc)

	var first bool
	require.NoError(t, conn.QueryRowContext(t.Context(), callTrial, blindIndex("hijack@x")).Scan(&first))
	assert.True(t, first)
	var shadowRows int
	require.NoError(t, conn.QueryRowContext(t.Context(), "SELECT count(*) FROM pg_temp.trial_registry").Scan(&shadowRows))
	assert.Zero(t, shadowRows, "the insert went to public.trial_registry")
}

// trial_registry stops trial farming: the second sign-up with the same email gets no
// trial, even after the first account was deleted (no FK to users, no cascade).
func TestTrialRegistry_BlocksASecondTrialEvenAfterAccountDeletion(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	seedUsers(t, app, userA)
	email := userEmailHash(userA)

	assert.True(t, checkTrial(t, app, email), "first trial")
	assert.False(t, checkTrial(t, app, email), "same email again")

	affected, err := execAsUser(t, app, userA, "DELETE FROM users WHERE id = $1", userA)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)
	assert.False(t, checkTrial(t, app, email), "after the account is deleted")
	assert.True(t, checkTrial(t, app, blindIndex("other@x")), "another email still gets one")

	t.Run("malformed hash is refused", func(t *testing.T) {
		_, err := app.ExecContext(t.Context(), callTrial, []byte("short"))
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	})
}

func TestEmailVerification_CorrectCodeVerifiesOnce(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("123456")
	upsertVerification(t, app, email, code, `{"name": "Ana"}`, "15 minutes")

	got := consume(t, app, email, code)
	assert.Equal(t, outcomeVerified, got.outcome)
	assert.JSONEq(t, `{"name": "Ana"}`, got.pending.String)

	again := consume(t, app, email, code)
	assert.Equal(t, outcomeInvalid, again.outcome, "single use")
	assert.False(t, again.pending.Valid)
}

func TestEmailVerification_WrongUnknownAndExpiredCodesFail(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("123456")
	upsertVerification(t, app, email, code, `{}`, "15 minutes")

	wrong := consume(t, app, email, blindIndex("654321"))
	assert.Equal(t, outcomeInvalid, wrong.outcome)
	assert.False(t, wrong.pending.Valid)
	assert.Equal(t, outcomeInvalid, consume(t, app, blindIndex("nobody@x"), code).outcome, "unknown email")

	_, err := db.Connect(t, pgtest.RoleOwner).ExecContext(t.Context(),
		"UPDATE email_verifications SET created_at = now() - interval '1 hour', expires_at = now() - interval '1 second'")
	require.NoError(t, err)
	assert.Equal(t, outcomeExpired, consume(t, app, email, code).outcome, "right code, too late")
	assert.Equal(t, outcomeInvalid, consume(t, app, email, code).outcome, "the expired row is gone")
}

func TestEmailVerification_AttemptLimitLocksEvenTheRightCode(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("123456")
	upsertVerification(t, app, email, code, `{}`, "15 minutes")

	for i := 1; i < maxVerificationAttempts; i++ {
		assert.Equal(t, outcomeInvalid, consume(t, app, email, blindIndex("bad")).outcome, "attempt %d", i)
	}
	assert.Equal(t, outcomeLocked, consume(t, app, email, blindIndex("bad")).outcome, "last attempt")
	assert.Equal(t, outcomeLocked, consume(t, app, email, code).outcome, "the right code after the limit")

	t.Run("a resend issues a new code and a fresh budget", func(t *testing.T) {
		fresh := blindIndex("111111")
		upsertVerification(t, app, email, fresh, `{"name": "Ana 2"}`, "15 minutes")
		assert.Equal(t, outcomeInvalid, consume(t, app, email, code).outcome, "the old code is replaced")
		got := consume(t, app, email, fresh)
		assert.Equal(t, outcomeVerified, got.outcome)
		assert.JSONEq(t, `{"name": "Ana 2"}`, got.pending.String)
	})
}

// Concurrent wrong guesses cannot race past the budget: the row lock serializes them.
func TestEmailVerification_ConcurrentAttemptsAreCountedExactly(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)
	email := blindIndex("ana@x")
	upsertVerification(t, app, email, blindIndex("123456"), `{}`, "15 minutes")

	const guesses = 3 * maxVerificationAttempts
	errs := make(chan error, guesses)
	for range guesses {
		go func() {
			_, err := app.ExecContext(t.Context(), callConsume, email, blindIndex("bad"))
			errs <- err
		}()
	}
	for range guesses {
		require.NoError(t, <-errs)
	}
	var attempts int
	require.NoError(t, db.Connect(t, pgtest.RoleOwner).QueryRowContext(t.Context(),
		"SELECT attempts FROM email_verifications").Scan(&attempts))
	assert.Equal(t, maxVerificationAttempts, attempts)
}

func TestEmailVerificationUpsert_RejectsBadArguments(t *testing.T) {
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	email, code := blindIndex("ana@x"), blindIndex("123456")
	bad := []struct {
		name        string
		email, code []byte
		pending     string
		ttl         string
	}{
		{"expiry in the past", email, code, `{}`, "-1 second"},
		{"expiry beyond one hour", email, code, `{}`, "61 minutes"},
		{"code hash not 32 bytes", email, []byte("123456"), `{}`, "15 minutes"},
		{"email hash not 32 bytes", []byte("a"), code, `{}`, "15 minutes"},
		{"pending not an object", email, code, `[]`, "15 minutes"},
	}
	for _, b := range bad {
		t.Run(b.name, func(t *testing.T) {
			_, err := app.ExecContext(t.Context(), callUpsert, b.email, b.code, b.pending, b.ttl)
			requireSQLState(t, err, sqlstateInvalidParameterValue)
		})
	}
	t.Run("consume refuses a malformed code hash", func(t *testing.T) {
		_, err := app.ExecContext(t.Context(), callConsume, email, []byte("123456"))
		requireSQLState(t, err, sqlstateInvalidParameterValue)
	})
}

// legal_documents is global: cc_app reads the seeded placeholders, nobody at runtime
// writes, and the worker and webhook read nothing.
func TestLegalDocuments_SeededPlaceholdersReadOnlyToTheApp(t *testing.T) {
	db := pgtest.NewDatabase(t)
	app := db.Connect(t, pgtest.RoleApp)

	docs := queryStrings(t, app,
		"SELECT kind || ' ' || version || ' ' || review_status || ' ' || (url LIKE 'https://%.invalid/%') FROM legal_documents ORDER BY kind")
	assert.Equal(t, []string{
		"privacy_notice_integral v1 pending_legal_review true",
		"privacy_notice_simplified v1 pending_legal_review true",
		"terms v1 pending_legal_review true",
	}, docs)

	for _, role := range []pgtest.Role{pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role)+" reads nothing", func(t *testing.T) {
			_, err := db.Connect(t, role).ExecContext(t.Context(), "SELECT 1 FROM legal_documents")
			requireSQLState(t, err, sqlstateInsufficientPrivilege)
		})
	}

	owner := db.Connect(t, pgtest.RoleOwner)
	invalid := []struct{ name, query string }{
		{"unknown kind", "INSERT INTO legal_documents (kind, version, url, effective_at) VALUES ('cookies', 'v1', 'https://a.invalid/c', now())"},
		{"plain http url", "INSERT INTO legal_documents (kind, version, url, effective_at) VALUES ('terms', 'v2', 'http://a.invalid/t', now())"},
		{"approved without a content hash", "INSERT INTO legal_documents (kind, version, url, effective_at, review_status) VALUES ('terms', 'v2', 'https://a.invalid/t', now(), 'approved')"},
		{"content hash not 32 bytes", "INSERT INTO legal_documents (kind, version, url, effective_at, content_sha256) VALUES ('terms', 'v2', 'https://a.invalid/t', now(), '\\x01')"},
	}
	for _, c := range invalid {
		t.Run(c.name, func(t *testing.T) {
			_, err := owner.ExecContext(t.Context(), c.query)
			requireSQLState(t, err, sqlstateCheckViolation)
		})
	}
	t.Run("two versions of a kind cannot take effect at the same instant", func(t *testing.T) {
		_, err := owner.ExecContext(t.Context(), `INSERT INTO legal_documents (kind, version, url, effective_at)
			SELECT kind, 'v2', url, effective_at FROM legal_documents WHERE kind = 'terms'`)
		requireSQLState(t, err, sqlstateUniqueViolation)
	})
}
