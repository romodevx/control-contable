package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// MODULE-auth §5 + ADR-0005 (M1-17 lockout mitigation): failures are
// limited per (email, client network) — 5 / 15 min — so a stranger spending
// a victim's budget locks out only their own network, under a higher
// per-email ceiling that bounds distributed guessing. Both are keyed by the
// email's blind index whether or not an account holds it (no enumeration)
// and checked before the lookup and argon2id.

func TestLogin_spendsTheEmailsBudgetsBeforeAnyLookup(t *testing.T) {
	for name, email := range map[string][2]string{ // as typed, normalized
		"an account's email": {validEmail, normalEmail},
		"an unknown email":   {" Nadie@Example.MX", "nadie@example.mx"},
	} {
		t.Run(name, func(t *testing.T) {
			e := newLoginEnv(t, currentParams, true)

			_, _ = e.uc.Execute(t.Context(), login(email[0], wrongSample))

			index := hexOf(emailHashOf(email[1]))
			assert.Equal(t, []string{index + " " + testClient}, e.failures.taken(),
				"keyed by the blind index and the client network, never the email")
			assert.Equal(t, []string{index}, e.ceiling.taken(), "the per-email ceiling, for known and unknown addresses alike")
		})
	}
}

// The targeted-lockout DoS: a stranger exhausting a victim's budget from
// their own network does not stop the victim signing in from theirs.
func TestLogin_anotherNetworksFailuresDoNotLockTheAccountOut(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.failures.limit = 5
	attacker := login(validEmail, wrongSample)
	attacker.Client = "198.51.100.0/24"
	for range 10 {
		_, _ = e.uc.Execute(t.Context(), attacker)
	}

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.NoError(t, err, "the victim's network has its own budget")
}

func TestLogin_theSameNetworkIsLimitedAfterItsBudget(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.failures.limit = 5
	for range 5 {
		_, _ = e.uc.Execute(t.Context(), login(validEmail, wrongSample))
	}

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	var limited *usecase.RateLimitedError
	require.ErrorAs(t, err, &limited)
	assert.Len(t, e.ceiling.taken(), 5, "a request its network refuses spends nothing of the ceiling")
}

func TestLogin_theEmailsCeilingRefusesEveryNetwork(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.ceiling.limit = 0

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	var limited *usecase.RateLimitedError
	require.ErrorAs(t, err, &limited, "distributed guessing is bounded per email")
	assert.Equal(t, time.Minute, limited.RetryAfter)
	assert.Zero(t, e.credentials.lookups)
	assert.Empty(t, e.passwords.verified)
}

func TestLogin_anEmailOutOfBudgetIsRefusedBeforeTheLookupAndArgon2(t *testing.T) {
	for name, email := range map[string]string{
		"an account's email": validEmail,
		"an unknown email":   "nadie@example.mx",
	} {
		t.Run(name, func(t *testing.T) {
			e := newLoginEnv(t, currentParams, true)
			e.failures.limit = 0

			_, err := e.uc.Execute(t.Context(), login(email, acceptedSample))

			var limited *usecase.RateLimitedError
			require.ErrorAs(t, err, &limited, "the same answer whether or not an account holds the email")
			assert.Equal(t, time.Minute, limited.RetryAfter)
			assert.NotErrorIs(t, err, usecase.ErrInvalidCredentials)
			assert.Zero(t, e.credentials.lookups)
			assert.Empty(t, e.passwords.verified, "no argon2id run")
			assert.Empty(t, e.refresh.families, "not even the right password signs in")
			assert.Contains(t, e.logs.String(), `"event":"login_rate_limited"`)
			assert.NotContains(t, e.logs.String(), email)
		})
	}
}

// Only failures count: a successful login refills the account's budget, so
// a user who mistyped a few times is not locked out on their next mistake.
func TestLogin_successRefillsTheEmailsBudget(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.NoError(t, err)
	index := hexOf(emailHashOf(normalEmail))
	assert.Equal(t, []string{index + " " + testClient}, e.failures.resets)
	assert.Equal(t, []string{index}, e.ceiling.resets)
}

func TestLogin_failuresNeverRefillTheBudget(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, wrongSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Empty(t, e.failures.resets)
	assert.Empty(t, e.ceiling.resets)
}

// A malformed address can hold no account: it is refused like any failure
// (one dummy argon2id run) and keys no bucket. The client limit still
// applies at the edge.
func TestLogin_aMalformedEmailKeysNoBucket(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login("not an email", acceptedSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Empty(t, e.failures.taken())
	assert.Empty(t, e.ceiling.taken())
}

func TestLogin_aLimiterThatCannotDecideFailsClosed(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.failures.err = errDB

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.ErrorIs(t, err, errDB)
	assert.NotErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Zero(t, e.credentials.lookups)
	assert.Empty(t, e.refresh.families)
}

func TestLogin_aCeilingThatCannotDecideFailsClosed(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)
	e.ceiling.err = errDB

	_, err := e.uc.Execute(t.Context(), login(validEmail, acceptedSample))

	require.ErrorIs(t, err, errDB)
	assert.Zero(t, e.credentials.lookups)
}

func TestNewLogin_refusesToRunWithoutItsFailureLimits(t *testing.T) {
	_, err := usecase.NewLogin(usecase.LoginDeps{Entropy: zeroReader{}, Passwords: &fakePasswords{}, Ceiling: &fakeLimiter{}})
	assert.ErrorContains(t, err, "Failures")

	_, err = usecase.NewLogin(usecase.LoginDeps{Entropy: zeroReader{}, Passwords: &fakePasswords{}, Failures: &fakeLimiter{}})
	assert.ErrorContains(t, err, "Ceiling")
}
