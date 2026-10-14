package usecase_test

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// fakePasswords stands in for argon2id: a hash is "argon:<params>:<password>",
// so a test can tell which parameters made it. It records every Verify (the
// expensive call whose count equalizes login timing).
type fakePasswords struct {
	params    string // the current policy's parameters
	hashErr   error
	verified  []string // the encoded hashes Verify ran against
	hashCalls int
}

func (f *fakePasswords) Hash(password string) (string, error) {
	f.hashCalls++
	if f.hashErr != nil {
		return "", f.hashErr
	}
	return "argon:" + f.params + ":" + password, nil
}

func (f *fakePasswords) Verify(password, encoded string) (bool, error) {
	f.verified = append(f.verified, encoded)
	parts := strings.SplitN(encoded, ":", 3)
	if len(parts) != 3 || parts[0] != "argon" {
		return false, errors.New("fake: malformed hash")
	}
	return parts[2] == password, nil
}

func (f *fakePasswords) NeedsRehash(encoded string) (bool, error) {
	return !strings.HasPrefix(encoded, "argon:"+f.params+":"), nil
}

// paramsOf returns the parameters an encoded fake hash was made with.
func paramsOf(encoded string) string { return strings.SplitN(encoded, ":", 3)[1] }

// fakeCredentials scripts auth_lookup_credentials by email hash.
type fakeCredentials struct {
	byHash  map[string]usecase.Credentials
	err     error
	lookups int
}

func (f *fakeCredentials) Lookup(_ context.Context, emailHash []byte) (usecase.Credentials, bool, error) {
	f.lookups++
	if f.err != nil {
		return usecase.Credentials{}, false, f.err
	}
	c, ok := f.byHash[hex.EncodeToString(emailHash)]
	return c, ok, nil
}

type replaceCall struct {
	userID        domain.UserID
	oldHash, hash string
}

// fakeLoginAccounts reads the account and records password rehashes.
type fakeLoginAccounts struct {
	account  usecase.Account
	findErr  error
	replaced []replaceCall
}

func (f *fakeLoginAccounts) Find(_ context.Context, id domain.UserID) (usecase.Account, error) {
	if f.findErr != nil {
		return usecase.Account{}, f.findErr
	}
	a := f.account
	a.ID = id
	return a, nil
}

func (f *fakeLoginAccounts) ReplacePasswordHash(_ context.Context, id domain.UserID, oldHash, newHash string) error {
	f.replaced = append(f.replaced, replaceCall{id, oldHash, newHash})
	return nil
}

const (
	currentParams = "m19456t2p1"
	olderParams   = "m8192t1p1"
	wrongSample   = "otra frase equivocada" // made-up sample phrase, not a credential
)

type loginEnv struct {
	uc          *usecase.Login
	system      *fakeSystem
	tx          *fakeTx
	passwords   *fakePasswords
	credentials *fakeCredentials
	accounts    *fakeLoginAccounts
	refresh     *fakeRefresh
	failures    *fakeLimiter
	ceiling     *fakeLimiter
	logs        interface{ String() string }
}

// newLoginEnv knows one active account for normalEmail whose stored hash
// was made with params.
func newLoginEnv(t *testing.T, params string, active bool) *loginEnv {
	t.Helper()
	userID, err := fixedUserID()
	require.NoError(t, err)
	email, err := domain.NewEmail(normalEmail)
	require.NoError(t, err)
	logger, logs := capturedLogs()
	e := &loginEnv{
		system:    &fakeSystem{},
		tx:        &fakeTx{},
		passwords: &fakePasswords{params: currentParams},
		credentials: &fakeCredentials{byHash: map[string]usecase.Credentials{
			hexOf(emailHashOf(normalEmail)): {UserID: userID, PasswordHash: "argon:" + params + ":" + acceptedSample, Active: active},
		}},
		accounts: &fakeLoginAccounts{account: usecase.Account{Email: email, CreatedAt: now}},
		refresh:  &fakeRefresh{},
		failures: &fakeLimiter{limit: 100},
		ceiling:  &fakeLimiter{limit: 100},
		logs:     logs,
	}
	e.uc, err = usecase.NewLogin(usecase.LoginDeps{
		System:      e.system,
		Tx:          e.tx,
		Clock:       fixedClock,
		Entropy:     zeroReader{},
		Index:       &sha256Index{},
		Passwords:   e.passwords,
		Credentials: e.credentials,
		Accounts:    e.accounts,
		Refresh:     e.refresh,
		Access:      fakeAccess{},
		Failures:    e.failures,
		Ceiling:     e.ceiling,
		Logger:      logger,
	})
	require.NoError(t, err)
	e.passwords.hashCalls = 0 // the dummy hash made at construction
	return e
}

// testClient is the client network every login() comes from.
const testClient = "203.0.113.0/24"

func login(email, password string) usecase.LoginInput {
	return usecase.LoginInput{Email: email, Password: password, Client: testClient}
}

func TestLogin_rightCredentialsStartANewTokenFamily(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	session, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.NoError(t, err)
	assert.Equal(t, 1, e.system.calls, "the lookup runs with no user context (definer function)")
	assert.Equal(t, newUserID, e.tx.userID, "the session is written as the user, so RLS admits its rows")
	assert.Equal(t, newUserID, session.Account.ID.String())
	assert.Equal(t, normalEmail, session.Account.Email.String())
	assert.Equal(t, "access:"+newUserID+"@2026-10-13T18:30:00Z", session.Tokens.Access)
	assert.Equal(t, accessTTL, session.Tokens.AccessExpiresIn)
	assert.Equal(t, usecase.RefreshSlidingWindow, session.Tokens.RefreshExpiresIn)
	require.Len(t, e.refresh.families, 1, "a new family")
	assert.Equal(t, session.Tokens.Refresh, e.refresh.families[0].token)
	assert.Equal(t, newUserID, e.refresh.families[0].token.UserID().String())
	assert.Equal(t, usecase.RefreshAbsoluteLifetime, e.refresh.families[0].absolute)
	assert.Empty(t, e.accounts.replaced, "current parameters: no rehash")
	assert.Zero(t, e.passwords.hashCalls)
}

// US-A2a: a wrong password, an unknown email, an account that may not sign
// in and an address that is not even valid all answer the same error after
// exactly one argon2id verification under the current parameters, so
// neither the answer nor its cost tells them apart. None reaches the user's
// unit of work or issues a token.
func TestLogin_everyFailureIsTheSameErrorAfterOneVerification(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		email    string
		password string
	}{
		{"wrong password", true, validEmail, wrongSample},
		{"unknown email", true, "nadie@example.mx", acceptedSample},
		{"suspended account, right password", false, validEmail, acceptedSample},
		{"invalid email", true, "not an email", acceptedSample},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newLoginEnv(t, currentParams, tt.active)

			_, err := e.uc.Execute(t.Context(), login(tt.email, tt.password))

			require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
			require.Len(t, e.passwords.verified, 1, "exactly one argon2id verification")
			assert.Equal(t, currentParams, paramsOf(e.passwords.verified[0]), "at the current cost")
			assert.Zero(t, e.tx.calls, "no user unit of work")
			assert.Empty(t, e.refresh.families, "no token")
			assert.NotContains(t, e.logs.String(), normalEmail)
			assert.NotContains(t, e.logs.String(), tt.password)
		})
	}
}

func TestLogin_failuresAreAudited(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, wrongSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Contains(t, e.logs.String(), `"event":"login_failed"`)
	assert.Contains(t, e.logs.String(), `"user_id":"`+newUserID+`"`)
}

func TestLogin_unknownEmailVerifiesAgainstADummyHashOfTheCurrentPolicy(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login("nadie@example.mx", acceptedSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	require.Len(t, e.passwords.verified, 1)
	needs, err := e.passwords.NeedsRehash(e.passwords.verified[0])
	require.NoError(t, err)
	assert.False(t, needs, "the dummy hash costs what a real one does")
	assert.Equal(t, 1, e.credentials.lookups)
}

// The policy rose since the password was stored: the right password is
// rehashed under the current parameters, in the user's unit of work, guarded
// by the old hash so a concurrent change is never overwritten.
func TestLogin_rehashesAPasswordStoredWithOutdatedParameters(t *testing.T) {
	e := newLoginEnv(t, olderParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.NoError(t, err)
	require.Len(t, e.accounts.replaced, 1)
	got := e.accounts.replaced[0]
	assert.Equal(t, newUserID, got.userID.String())
	assert.Equal(t, "argon:"+olderParams+":"+acceptedSample, got.oldHash)
	assert.Equal(t, "argon:"+currentParams+":"+acceptedSample, got.hash)
	assert.Equal(t, newUserID, e.tx.userID)
}

func TestLogin_aFailedRehashDoesNotFailTheLogin(t *testing.T) {
	e := newLoginEnv(t, olderParams, true)
	e.passwords.hashErr = errors.New("out of memory")

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.NoError(t, err)
	assert.Empty(t, e.accounts.replaced)
	assert.Contains(t, e.logs.String(), "rehash")
}

func TestLogin_aWrongPasswordIsNeverRehashed(t *testing.T) {
	e := newLoginEnv(t, olderParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, wrongSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Zero(t, e.passwords.hashCalls)
	assert.Empty(t, e.accounts.replaced)
}

func TestLogin_anAccountDeletedMeanwhileIsInvalidCredentials(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.accounts.findErr = usecase.ErrAccountNotFound

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
}

// A stored hash this code cannot parse is a data fault: logged, and the
// caller gets the generic answer (a 500 would tell the account exists).
func TestLogin_aMalformedStoredHashIsInvalidCredentials(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	for k, c := range e.credentials.byHash {
		c.PasswordHash = "not-a-phc-string"
		e.credentials.byHash[k] = c
	}

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Contains(t, e.logs.String(), "stored password hash")
	assert.Zero(t, e.tx.calls)
}

func TestLogin_infrastructureFailuresAreNotInvalidCredentials(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.credentials.err = errDB

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.ErrorIs(t, err, errDB)
	assert.NotErrorIs(t, err, usecase.ErrInvalidCredentials)
}

func TestNewLogin_failsWhenTheDummyHashCannotBeMade(t *testing.T) {
	_, err := usecase.NewLogin(usecase.LoginDeps{
		Entropy:   zeroReader{},
		Passwords: &fakePasswords{params: currentParams, hashErr: errors.New("boom")},
	})
	assert.Error(t, err)
}
