//go:build integration

package infra_test

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
)

// Made-up sample phrases (not credentials): two acceptable ones.
const (
	firstSample  = "tlacuache montaña 2026"
	secondSample = "ajolote del lago 1987"
	anaEmail     = "ana.gomez@example.mx"
)

// expireSignUps and pastCooldown move every pending sign-up's clock (fixtures as cc_owner).
const (
	pastCooldown  = "UPDATE email_verifications SET code_sent_at = now() - interval '61 seconds', created_at = now() - interval '61 seconds'"
	expireSignUps = "UPDATE email_verifications SET created_at = now() - interval '1 hour', code_sent_at = now() - interval '1 hour', expires_at = now() - interval '1 second'"
)

func registration(email, phrase string, name *string) usecase.RegisterInput {
	return usecase.RegisterInput{
		Email: email, Password: phrase, Name: name,
		AcceptedDocuments: []domain.DocumentVersion{
			{Kind: domain.DocumentPrivacySimplified, Version: "v1"},
			{Kind: domain.DocumentTerms, Version: "v1"},
		},
	}
}

func ptr(s string) *string { return &s }

// storedUser is what verifyEmail wrote, read back as the user.
type storedUser struct {
	emailHash, emailEncrypted []byte
	keyID                     int
	passwordHash              string
	name                      sql.NullString
	verified                  bool
	consents                  []string
	usageRows                 int
	trialStart, trialEnd      sql.NullTime
	refreshHashes             [][]byte
}

func (s *stack) readUser(t *testing.T, userID string) storedUser {
	t.Helper()
	var u storedUser
	s.asUser(t, userID, func(tx *sql.Tx) error {
		ctx := t.Context()
		if err := tx.QueryRowContext(ctx, `SELECT email_hash, email_encrypted, email_key_id, password_hash, name, email_verified_at IS NOT NULL
FROM users WHERE id = $1`, userID).Scan(&u.emailHash, &u.emailEncrypted, &u.keyID, &u.passwordHash, &u.name, &u.verified); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, "SELECT document_kind || ':' || document_version FROM consents WHERE user_id = $1 ORDER BY 1", userID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				return err
			}
			u.consents = append(u.consents, c)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM user_usage WHERE user_id = $1", userID).Scan(&u.usageRows); err != nil {
			return err
		}
		err = tx.QueryRowContext(ctx, "SELECT starts_at, ends_at FROM entitlements WHERE user_id = $1 AND kind = 'trial'", userID).
			Scan(&u.trialStart, &u.trialEnd)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		hashes, err := tx.QueryContext(ctx, "SELECT token_hash FROM refresh_tokens WHERE user_id = $1", userID)
		if err != nil {
			return err
		}
		defer hashes.Close()
		for hashes.Next() {
			var h []byte
			if err := hashes.Scan(&h); err != nil {
				return err
			}
			u.refreshHashes = append(u.refreshHashes, h)
		}
		return hashes.Err()
	})
	return u
}

// US-A1a → US-A1b end to end as cc_app: nothing is written to users until the
// code is verified; then the user, consents, counters, trial and session exist.
func TestSignUp_registerThenVerifyCreatesTheAccount(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()

	require.NoError(t, s.register.Execute(ctx, registration("  Ana.Gomez@Example.MX ", firstSample, ptr("Ana Gómez"))))
	var users int
	require.NoError(t, s.owner.QueryRowContext(ctx, "SELECT count(*) FROM email_verifications").Scan(&users))
	assert.Equal(t, 1, users, "a pending sign-up, no user yet")

	session, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)

	userID := session.Account.ID.String()
	u := s.readUser(t, userID)
	wantHash, err := s.index.Hash(anaEmail)
	require.NoError(t, err)
	assert.Equal(t, wantHash, u.emailHash)
	plain, err := s.keyring.Decrypt(u.emailEncrypted, []byte("users.email_encrypted:"+userID))
	require.NoError(t, err)
	assert.Equal(t, anaEmail, string(plain), "sealed and bound to the user id")
	assert.Equal(t, 1, u.keyID)
	ok, err := s.passwords.Verify(firstSample, u.passwordHash)
	require.NoError(t, err)
	assert.True(t, ok, "the registered password")
	assert.Equal(t, "Ana Gómez", u.name.String)
	assert.True(t, u.verified)
	assert.Equal(t, []string{"privacy_notice_simplified:v1", "terms:v1"}, u.consents)
	assert.Equal(t, 1, u.usageRows)
	require.True(t, u.trialStart.Valid, "first trial")
	assert.Equal(t, access.TrialDuration, u.trialEnd.Time.Sub(u.trialStart.Time))
	require.Len(t, u.refreshHashes, 1)
	sum := sha256.Sum256([]byte(session.Tokens.Refresh.Reveal()))
	assert.Equal(t, sum[:], u.refreshHashes[0], "only the token's hash is stored")
	assert.Equal(t, "Ana Gómez", session.Account.Name.String())
	assert.WithinDuration(t, time.Now(), session.Account.CreatedAt, time.Minute)
	assert.NotEmpty(t, session.Tokens.Access)

	_, err = s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	assert.ErrorIs(t, err, usecase.ErrVerificationRejected, "a code is single use")
}

func TestSignUp_withoutANameStoresNull(t *testing.T) {
	s := newStack(t)
	require.NoError(t, s.register.Execute(t.Context(), registration(anaEmail, firstSample, nil)))

	session, err := s.verify.Execute(t.Context(), anaEmail, s.inbox.lastCode(t, anaEmail))

	require.NoError(t, err)
	assert.False(t, s.readUser(t, session.Account.ID.String()).name.Valid)
}

// D7 end to end: a second registrant cannot set the password of a live pending
// sign-up; it only triggers a resend to the address.
func TestSignUp_secondRegisterOfALivePendingKeepsTheFirstPassword(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, ptr("Ana"))))

	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, ptr("Mallory"))))
	assert.Len(t, s.inbox.codes[anaEmail], 1, "inside the cooldown: silent")

	s.ownerExec(t, pastCooldown)
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, ptr("Mallory"))))
	require.Len(t, s.inbox.codes[anaEmail], 2, "after the cooldown: a resend")

	_, err := s.verify.Execute(ctx, anaEmail, s.inbox.codes[anaEmail][0])
	require.ErrorIs(t, err, usecase.ErrVerificationRejected, "the resend replaced the code")
	session, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)

	u := s.readUser(t, session.Account.ID.String())
	ok, err := s.passwords.Verify(firstSample, u.passwordHash)
	require.NoError(t, err)
	assert.True(t, ok, "the first registrant's password")
	assert.Equal(t, "Ana", u.name.String)
}

func TestSignUp_expiredPendingIsReplacedByTheNextRegister(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, ptr("Old"))))
	s.ownerExec(t, expireSignUps)

	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, ptr("New"))))
	session, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)

	u := s.readUser(t, session.Account.ID.String())
	ok, err := s.passwords.Verify(secondSample, u.passwordHash)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "New", u.name.String)
}

// D3: a verified email creates no pending sign-up and notifies its owner.
func TestSignUp_registerOfAVerifiedEmailOnlyNotifies(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, nil)))
	_, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)

	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, nil)))
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, nil)))

	assert.Len(t, s.inbox.codes[anaEmail], 1, "no new code")
	assert.Equal(t, 1, s.inbox.notices[anaEmail], "one notice per hour")
	var pending int
	require.NoError(t, s.owner.QueryRowContext(ctx, "SELECT count(*) FROM email_verifications").Scan(&pending))
	assert.Zero(t, pending)
}

func TestSignUp_staleDocumentVersionIsRejected(t *testing.T) {
	s := newStack(t)
	in := registration(anaEmail, firstSample, nil)
	in.AcceptedDocuments[1].Version = "v0"

	err := s.register.Execute(t.Context(), in)

	require.ErrorIs(t, err, domain.ErrDocumentNotCurrent)
	assert.Empty(t, s.inbox.codes)
}

// US-A1c: resend reaches only a live pending sign-up, outside the cooldown.
func TestSignUp_resend(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()

	require.NoError(t, s.resend.Execute(ctx, anaEmail))
	assert.Empty(t, s.inbox.codes, "no pending sign-up: silent")

	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, nil)))
	require.NoError(t, s.resend.Execute(ctx, anaEmail))
	assert.Len(t, s.inbox.codes[anaEmail], 1, "cooldown: silent")

	s.ownerExec(t, pastCooldown)
	require.NoError(t, s.resend.Execute(ctx, anaEmail))
	require.Len(t, s.inbox.codes[anaEmail], 2)
	_, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	assert.NoError(t, err, "the resent code verifies")
}

// Wrong tries count and commit: after 5 the code is dead even when right.
func TestSignUp_wrongCodesLockTheCode(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, nil)))
	right := s.inbox.lastCode(t, anaEmail)
	wrong := "000000"
	if right == wrong {
		wrong = "111111"
	}

	for range 5 {
		_, err := s.verify.Execute(ctx, anaEmail, wrong)
		require.ErrorIs(t, err, usecase.ErrVerificationRejected)
	}
	_, err := s.verify.Execute(ctx, anaEmail, right)

	assert.ErrorIs(t, err, usecase.ErrVerificationRejected)
	var attempts int
	require.NoError(t, s.owner.QueryRowContext(ctx, "SELECT attempts FROM email_verifications").Scan(&attempts))
	assert.Equal(t, 5, attempts, "every wrong try was committed")
}

// An email that already had a trial (account deleted, then back) gets an
// account without a second trial.
func TestSignUp_trialRegistrySurvivesDeletion(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, firstSample, nil)))
	first, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)
	s.asUser(t, first.Account.ID.String(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", first.Account.ID.String())
		return err
	})

	require.NoError(t, s.register.Execute(ctx, registration(anaEmail, secondSample, nil)))
	again, err := s.verify.Execute(ctx, anaEmail, s.inbox.lastCode(t, anaEmail))
	require.NoError(t, err)

	u := s.readUser(t, again.Account.ID.String())
	assert.False(t, u.trialStart.Valid, "no second trial")
	assert.Equal(t, 1, u.usageRows)
}

// What one sign-up wrote stays invisible to another user (RLS as cc_app).
func TestSignUp_rowsAreIsolatedPerUser(t *testing.T) {
	s := newStack(t)
	ctx := t.Context()
	var ids []string
	for _, email := range []string{anaEmail, "jose@example.mx"} {
		require.NoError(t, s.register.Execute(ctx, registration(email, firstSample, nil)))
		session, err := s.verify.Execute(ctx, email, s.inbox.lastCode(t, email))
		require.NoError(t, err)
		ids = append(ids, session.Account.ID.String())
	}

	for _, table := range []string{"users WHERE id = $1", "consents WHERE user_id = $1", "user_usage WHERE user_id = $1",
		"entitlements WHERE user_id = $1", "refresh_tokens WHERE user_id = $1"} {
		var seen int
		s.asUser(t, ids[0], func(tx *sql.Tx) error {
			return tx.QueryRowContext(ctx, "SELECT count(*) FROM "+table, ids[1]).Scan(&seen)
		})
		assert.Zero(t, seen, table)
	}
}
