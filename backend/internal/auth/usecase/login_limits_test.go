package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// MODULE-auth §5: 5 failures / 15 min per account, keyed by the email's
// blind index whether or not an account holds it (no enumeration), checked
// before the lookup and argon2id.

func TestLogin_spendsTheEmailsBudgetBeforeAnyLookup(t *testing.T) {
	for name, email := range map[string][2]string{ // as typed, normalized
		"an account's email": {validEmail, normalEmail},
		"an unknown email":   {" Nadie@Example.MX", "nadie@example.mx"},
	} {
		t.Run(name, func(t *testing.T) {
			e := newLoginEnv(t, currentParams, true)

			_, _ = e.uc.Execute(t.Context(), login(email[0], wrongSample))

			assert.Equal(t, []string{hexOf(emailHashOf(email[1]))}, e.failures.taken(),
				"keyed by the blind index, never the email, for known and unknown addresses alike")
		})
	}
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
	assert.Equal(t, []string{hexOf(emailHashOf(normalEmail))}, e.failures.resets)
}

func TestLogin_failuresNeverRefillTheBudget(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login(validEmail, wrongSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Empty(t, e.failures.resets)
}

// A malformed address can hold no account: it is refused like any failure
// (one dummy argon2id run) and keys no bucket. The client limit still
// applies at the edge.
func TestLogin_aMalformedEmailKeysNoBucket(t *testing.T) {
	e := newLoginEnv(t, currentParams, true)

	_, err := e.uc.Execute(t.Context(), login("not an email", acceptedSample))

	require.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	assert.Empty(t, e.failures.taken())
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

func TestNewLogin_refusesToRunWithoutAFailureLimit(t *testing.T) {
	_, err := usecase.NewLogin(usecase.LoginDeps{Entropy: zeroReader{}, Passwords: &fakePasswords{}})

	assert.ErrorContains(t, err, "Failures")
}
