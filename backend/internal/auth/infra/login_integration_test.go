//go:build integration

package infra_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// newLogin wires usecase.Login over the stack's database as cmd/api does,
// hashing under params (the "current policy" of this test).
func (s *stack) newLogin(t *testing.T, params crypto.Argon2idParams) *usecase.Login {
	t.Helper()
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	passwords, err := crypto.NewPasswordHasher(params)
	require.NoError(t, err)
	access, err := infra.NewAccessTokenSigner(tokenSeed, "k1")
	require.NoError(t, err)
	login, err := usecase.NewLogin(usecase.LoginDeps{
		System: txm, Tx: txm, Clock: clock.NewSystem(), Entropy: rand.Reader, Index: s.index,
		Passwords: passwords, Credentials: infra.NewCredentialStore(), Accounts: infra.NewAccountStore(s.keyring),
		Refresh: infra.NewRefreshTokenStore(), Access: access, Failures: s.limits.Bucket(usecase.LoginFailuresPolicy), Ceiling: s.limits.Bucket(usecase.LoginCeilingPolicy),
		Logger: slog.New(slog.DiscardHandler),
	})
	require.NoError(t, err)
	return login
}

// signUp creates a verified account through register + verifyEmail.
func (s *stack) signUp(t *testing.T, email, phrase string) string {
	t.Helper()
	require.NoError(t, s.register.Execute(t.Context(), registration(email)))
	session, err := s.verify.Execute(t.Context(), verifying(email, s.inbox.lastCode(t, email), phrase, ptr("Ana Gómez")))
	require.NoError(t, err)
	return session.Account.ID.String()
}

func TestLogin_rightCredentialsStartANewFamily(t *testing.T) {
	s := newStack(t)
	userID := s.signUp(t, anaEmail, firstSample)
	login := s.newLogin(t, testArgon2)

	session, err := login.Execute(t.Context(), usecase.LoginInput{Email: " Ana.Gomez@Example.MX", Password: firstSample})

	require.NoError(t, err)
	assert.Equal(t, userID, session.Account.ID.String())
	assert.Equal(t, anaEmail, session.Account.Email.String(), "decrypted from the sealed column")
	assert.Equal(t, "Ana Gómez", session.Account.Name.String())
	assert.WithinDuration(t, time.Now(), session.Account.CreatedAt, time.Minute)
	assert.NotEmpty(t, session.Tokens.Access)
	assert.Equal(t, userID, session.Tokens.Refresh.UserID().String())

	families := s.familiesOf(t, userID)
	assert.Len(t, families, 2, "sign-up's family and login's new one")
}

// familiesOf lists the user's refresh token families (as the user, RLS).
func (s *stack) familiesOf(t *testing.T, userID string) []string {
	t.Helper()
	var families []string
	s.asUser(t, userID, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(t.Context(), "SELECT DISTINCT family_id::text FROM refresh_tokens WHERE user_id = $1", userID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var f string
			if err := rows.Scan(&f); err != nil {
				return err
			}
			families = append(families, f)
		}
		return rows.Err()
	})
	return families
}

func TestLogin_failuresAreOneGenericError(t *testing.T) {
	s := newStack(t)
	userID := s.signUp(t, anaEmail, firstSample)
	const pending = "pendiente@example.mx"
	require.NoError(t, s.register.Execute(t.Context(), registration(pending)))
	login := s.newLogin(t, testArgon2)

	for name, in := range map[string]usecase.LoginInput{
		"wrong password":             {Email: anaEmail, Password: secondSample},
		"unknown email":              {Email: "nadie@example.mx", Password: firstSample},
		"pending (unverified) email": {Email: pending, Password: firstSample},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := login.Execute(t.Context(), in)
			assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
		})
	}
	assert.Len(t, s.familiesOf(t, userID), 1, "no session was started")
}

func TestLogin_suspendedAccountCannotSignIn(t *testing.T) {
	s := newStack(t)
	userID := s.signUp(t, anaEmail, firstSample)
	// An operator decision: users is FORCE RLS with no policy for cc_owner,
	// so the fixture lifts FORCE for its one statement, in one transaction.
	s.ownerExec(t, "BEGIN; ALTER TABLE users NO FORCE ROW LEVEL SECURITY; "+
		"UPDATE users SET status = 'suspended' WHERE id = '"+userID+"'; "+
		"ALTER TABLE users FORCE ROW LEVEL SECURITY; COMMIT")
	var status string
	require.NoError(t, s.owner.QueryRowContext(t.Context(),
		"SELECT relforcerowsecurity::text FROM pg_class WHERE relname = 'users'").Scan(&status))
	require.Equal(t, "true", status, "FORCE is back on")

	_, err := s.newLogin(t, testArgon2).Execute(t.Context(), usecase.LoginInput{Email: anaEmail, Password: firstSample})

	assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
}

// The policy rose since sign-up: the next login stores the password under
// the new parameters, and the password keeps working.
func TestLogin_rehashesWhenThePolicyRose(t *testing.T) {
	s := newStack(t)
	userID := s.signUp(t, anaEmail, firstSample)
	before := s.readUser(t, userID).passwordHash
	raised := testArgon2
	raised.Memory += 1024
	raised.Time++

	_, err := s.newLogin(t, raised).Execute(t.Context(), usecase.LoginInput{Email: anaEmail, Password: firstSample})
	require.NoError(t, err)

	after := s.readUser(t, userID).passwordHash
	assert.NotEqual(t, before, after)
	assert.True(t, strings.HasPrefix(after, "$argon2id$v=19$m=9216,t=2,p=1$"), after)
	hasher, err := crypto.NewPasswordHasher(raised)
	require.NoError(t, err)
	ok, err := hasher.Verify(firstSample, after)
	require.NoError(t, err)
	assert.True(t, ok)
	needs, err := hasher.NeedsRehash(after)
	require.NoError(t, err)
	assert.False(t, needs)

	_, err = s.newLogin(t, raised).Execute(t.Context(), usecase.LoginInput{Email: anaEmail, Password: firstSample})
	require.NoError(t, err)
	assert.Equal(t, after, s.readUser(t, userID).passwordHash, "already current: not rehashed again")
}

// US-A2a "comparable latency": an unknown email pays an argon2id
// verification like a wrong password does. With a deliberately costly
// policy the gap without the dummy verification would be two orders of
// magnitude; the bound below only has to tell "one verification" from
// "none", so scheduler noise cannot flip it.
func TestLogin_unknownEmailCostsAsMuchAsAWrongPassword(t *testing.T) {
	s := newStack(t)
	s.signUp(t, anaEmail, firstSample)
	costly := crypto.Argon2idParams{Memory: 32 * 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}
	login := s.newLogin(t, costly)
	// Store the password under the costly policy too, so both paths verify
	// at the same cost.
	_, err := login.Execute(t.Context(), usecase.LoginInput{Email: anaEmail, Password: firstSample})
	require.NoError(t, err)

	const runs = 5
	fastest := func(in usecase.LoginInput) time.Duration {
		var times []time.Duration
		for range runs {
			start := time.Now()
			_, err := login.Execute(t.Context(), in)
			times = append(times, time.Since(start))
			require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
		}
		return slices.Min(times)
	}
	wrongPassword := fastest(usecase.LoginInput{Email: anaEmail, Password: secondSample})
	unknownEmail := fastest(usecase.LoginInput{Email: "nadie@example.mx", Password: secondSample})

	t.Logf("fastest of %d: wrong password %v, unknown email %v", runs, wrongPassword, unknownEmail)
	assert.Greater(t, unknownEmail, wrongPassword/2, "an unknown email answers measurably faster")
}

func TestCredentialStore_lookup(t *testing.T) {
	s := newStack(t)
	userID := s.signUp(t, anaEmail, firstSample)
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	hash, err := s.index.Hash(anaEmail)
	require.NoError(t, err)
	unknown, err := s.index.Hash("nadie@example.mx")
	require.NoError(t, err)
	store := infra.NewCredentialStore()

	var (
		creds        usecase.Credentials
		found, other bool
	)
	require.NoError(t, txm.System(t.Context(), func(ctx context.Context) error {
		var err error
		if creds, found, err = store.Lookup(ctx, hash); err != nil {
			return err
		}
		_, other, err = store.Lookup(ctx, unknown)
		return err
	}))

	require.True(t, found)
	assert.Equal(t, userID, creds.UserID.String())
	assert.True(t, creds.Active)
	assert.True(t, strings.HasPrefix(creds.PasswordHash, "$argon2id$"))
	assert.False(t, other)
}

// Find and ReplacePasswordHash run under RLS: another user's row is not
// found and not changed.
func TestAccountStore_isolatedPerUser(t *testing.T) {
	s := newStack(t)
	ana := s.signUp(t, anaEmail, firstSample)
	jose := s.signUp(t, "jose@example.mx", secondSample)
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	store := infra.NewAccountStore(s.keyring)
	anaID, err := domain.ParseUserID(ana)
	require.NoError(t, err)
	joseHash := s.readUser(t, jose).passwordHash

	err = txm.ReadWrite(t.Context(), jose, func(ctx context.Context) error {
		_, err := store.Find(ctx, anaID)
		return err
	})
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)

	anaHash := s.readUser(t, ana).passwordHash
	require.NoError(t, txm.ReadWrite(t.Context(), jose, func(ctx context.Context) error {
		return store.ReplacePasswordHash(ctx, anaID, anaHash, joseHash)
	}))
	assert.Equal(t, anaHash, s.readUser(t, ana).passwordHash, "RLS: not jose's to change")
}

// A stale old hash (the password changed meanwhile) leaves the row alone.
func TestAccountStore_replacePasswordHashIsGuardedByTheOldHash(t *testing.T) {
	s := newStack(t)
	ana := s.signUp(t, anaEmail, firstSample)
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	anaID, err := domain.ParseUserID(ana)
	require.NoError(t, err)
	current := s.readUser(t, ana).passwordHash
	replacement, err := s.passwords.Hash(secondSample)
	require.NoError(t, err)
	store := infra.NewAccountStore(s.keyring)

	require.NoError(t, txm.ReadWrite(t.Context(), ana, func(ctx context.Context) error {
		return store.ReplacePasswordHash(ctx, anaID, "$argon2id$v=19$m=1,t=1,p=1$c3RhbGU$c3RhbGU", replacement)
	}))
	assert.Equal(t, current, s.readUser(t, ana).passwordHash)

	require.NoError(t, txm.ReadWrite(t.Context(), ana, func(ctx context.Context) error {
		return store.ReplacePasswordHash(ctx, anaID, current, replacement)
	}))
	assert.Equal(t, replacement, s.readUser(t, ana).passwordHash)
}
