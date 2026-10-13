package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

type verifyEnv struct {
	uc       *usecase.VerifyEmail
	tx       *fakeTx
	signUps  *fakeSignUps
	hasher   *prefixHasher
	accounts *fakeAccounts
	trials   *fakeTrials
	refresh  *fakeRefresh
}

func newVerifyEnv(signUps *fakeSignUps) *verifyEnv {
	e := &verifyEnv{
		tx: &fakeTx{}, signUps: signUps, hasher: &prefixHasher{},
		accounts: &fakeAccounts{}, trials: &fakeTrials{}, refresh: &fakeRefresh{},
	}
	e.uc = usecase.NewVerifyEmail(usecase.VerifyDeps{
		Tx:        e.tx,
		Clock:     fixedClock,
		Entropy:   zeroReader{},
		NewUserID: fixedUserID,
		Index:     &sha256Index{},
		Passwords: e.hasher,
		Codes:     concatCodes{},
		SignUps:   signUps,
		Accounts:  e.accounts,
		Trials:    e.trials,
		Refresh:   e.refresh,
		Access:    fakeAccess{},
	})
	return e
}

var acceptedAt = now.Add(-10 * time.Minute) // registered ten minutes earlier

// pendingSignUp is what register stored: the accepted documents only, no
// password and no name (D8).
func pendingSignUp() usecase.PendingSignUp {
	return usecase.PendingSignUp{
		Documents: []domain.DocumentVersion{
			{Kind: domain.DocumentPrivacyIntegral, Version: "v1"},
			{Kind: domain.DocumentTerms, Version: "v1"},
		},
		AcceptedAt: acceptedAt,
	}
}

func verification(code string) usecase.VerifyInput {
	name := "  Ana Gómez  "
	return usecase.VerifyInput{Email: validEmail, Code: code, Password: acceptedSample, Name: &name}
}

func TestVerifyEmail_rightCodeCreatesTheAccountStartsTheTrialAndASession(t *testing.T) {
	e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp(), firstTrial: true})
	userID, _ := fixedUserID()

	session, err := e.uc.Execute(t.Context(), verification("000000"))

	require.NoError(t, err)
	assert.Equal(t, newUserID, e.tx.userID, "one unit of work as the new user, so RLS admits its rows")
	require.Len(t, e.signUps.consumed, 1)
	assert.Equal(t, emailHashOf(normalEmail), e.signUps.consumed[0][0])
	assert.Equal(t, codeHashOf(normalEmail, "000000"), e.signUps.consumed[0][1])

	require.Len(t, e.accounts.created, 1)
	created := e.accounts.created[0]
	assert.Equal(t, userID, created.ID)
	assert.Equal(t, normalEmail, created.Email.String())
	assert.Equal(t, emailHashOf(normalEmail), created.EmailHash)
	assert.Equal(t, "hashed:"+acceptedSample, created.PasswordHash, "the password sent WITH the code (D8)")
	assert.Equal(t, "Ana Gómez", created.Name.String())
	assert.Equal(t, pendingSignUp().Documents, created.Consents)
	assert.Equal(t, acceptedAt, created.AcceptedAt)
	assert.Equal(t, now, created.VerifiedAt)

	assert.Equal(t, [][]byte{emailHashOf(normalEmail)}, e.signUps.trialClaims)
	assert.Equal(t, []domain.UserID{userID}, e.trials.userIDs)
	assert.Equal(t, []time.Time{now}, e.trials.started, "the trial starts on the server clock at verification")

	require.Len(t, e.refresh.families, 1)
	family := e.refresh.families[0]
	assert.Equal(t, userID, family.token.UserID())
	assert.Equal(t, usecase.RefreshSlidingWindow, family.sliding)
	assert.Equal(t, usecase.RefreshAbsoluteLifetime, family.absolute)

	assert.Equal(t, userID, session.Account.ID)
	assert.Equal(t, "access:"+newUserID+"@2026-10-13T18:30:00Z", session.Tokens.Access)
	assert.Equal(t, accessTTL, session.Tokens.AccessExpiresIn)
	assert.Equal(t, family.token.Reveal(), session.Tokens.Refresh.Reveal())
	assert.Equal(t, usecase.RefreshSlidingWindow, session.Tokens.RefreshExpiresIn)
}

func TestVerifyEmail_withoutANameKeepsItEmpty(t *testing.T) {
	e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp(), firstTrial: true})
	in := verification("000000")
	in.Name = nil

	_, err := e.uc.Execute(t.Context(), in)

	require.NoError(t, err)
	require.Len(t, e.accounts.created, 1)
	assert.True(t, e.accounts.created[0].Name.IsZero())
}

// An email already in trial_registry (deleted and came back) gets an account
// but no second trial.
func TestVerifyEmail_noSecondTrial(t *testing.T) {
	e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp(), firstTrial: false})

	_, err := e.uc.Execute(t.Context(), verification("000000"))

	require.NoError(t, err)
	assert.Len(t, e.accounts.created, 1)
	assert.Empty(t, e.trials.started)
}

// Every failed outcome is one generic error, and its unit of work COMMITS so
// the attempt counter sticks (email_verification_consume's caller contract).
// argon2 runs only for a verified code: a wrong guess costs no hash (no CPU
// amplification on a public endpoint).
func TestVerifyEmail_failedOutcomesAreRejectedAndCommitted(t *testing.T) {
	for _, outcome := range []usecase.ConsumeOutcome{usecase.ConsumeInvalid, usecase.ConsumeExpired, usecase.ConsumeLocked} {
		t.Run(outcome.String(), func(t *testing.T) {
			e := newVerifyEnv(&fakeSignUps{consume: outcome})

			_, err := e.uc.Execute(t.Context(), verification("123456"))

			require.ErrorIs(t, err, usecase.ErrVerificationRejected)
			assert.NoError(t, e.tx.workErr, "the work returned nil: committed")
			assert.Empty(t, e.accounts.created)
			assert.Empty(t, e.refresh.families)
			assert.Zero(t, e.hasher.calls, "no password hash for a rejected code")
		})
	}
}

func TestVerifyEmail_malformedEmailOrCodeIsRejectedWithoutTheDatabase(t *testing.T) {
	tests := []struct{ name, email, code string }{
		{"email the domain refuses", "ana@localhost", "123456"},
		{"code not six digits", validEmail, "12345"},
		{"code not digits", validEmail, "12345x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp()})
			in := verification(tt.code)
			in.Email = tt.email

			_, err := e.uc.Execute(t.Context(), in)

			require.ErrorIs(t, err, usecase.ErrVerificationRejected)
			assert.Zero(t, e.tx.calls)
		})
	}
}

// The password policy and the name are client errors (400), checked before
// the code: a weak password never spends one of the code's 5 tries.
func TestVerifyEmail_invalidPasswordOrNameIsAValidationErrorBeforeTheCode(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*usecase.VerifyInput)
		wantErr error
	}{
		{"common password", func(in *usecase.VerifyInput) { in.Password = commonSample }, domain.ErrPasswordCommon},
		{"short password", func(in *usecase.VerifyInput) { in.Password = shortSample }, domain.ErrPasswordTooShort},
		{"blank name", func(in *usecase.VerifyInput) { blank := "   "; in.Name = &blank }, domain.ErrInvalidName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp()})
			in := verification("000000")
			tt.mutate(&in)

			_, err := e.uc.Execute(t.Context(), in)

			require.ErrorIs(t, err, tt.wantErr)
			assert.NotErrorIs(t, err, usecase.ErrVerificationRejected)
			assert.Zero(t, e.tx.calls, "the code is not checked: no try spent")
			assert.Empty(t, e.signUps.consumed)
		})
	}
}

// Losing the race to a concurrent verification of the same email rolls back
// (the pending row comes back) and answers like a wrong code.
func TestVerifyEmail_accountAlreadyCreatedIsRejectedAndRolledBack(t *testing.T) {
	e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp(), firstTrial: true})
	e.accounts.err = usecase.ErrAccountExists

	_, err := e.uc.Execute(t.Context(), verification("000000"))

	require.ErrorIs(t, err, usecase.ErrVerificationRejected)
	assert.ErrorIs(t, e.tx.workErr, usecase.ErrAccountExists, "the work failed: rolled back")
	assert.Empty(t, e.trials.started)
}

func TestVerifyEmail_otherErrorsPropagate(t *testing.T) {
	e := newVerifyEnv(&fakeSignUps{consume: usecase.ConsumeVerified, pending: pendingSignUp()})
	e.accounts.err = errDB

	_, err := e.uc.Execute(t.Context(), verification("000000"))

	require.ErrorIs(t, err, errDB)
	assert.NotErrorIs(t, err, usecase.ErrVerificationRejected)
}
