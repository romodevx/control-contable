package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// MODULE-auth §5: register and resend share one budget per email (D7: a
// register for a live pending sign-up acts as a resend), keyed by the blind
// index, so it treats new, pending and verified addresses alike.

func TestRegister_spendsTheEmailsBudget(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{startOutcome: usecase.CodeStarted}, currentDocs)

	require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))

	assert.Equal(t, []string{hexOf(emailHashOf(normalEmail))}, e.emails.taken())
}

func TestRegister_anEmailOutOfBudgetIs429BeforeAnyWork(t *testing.T) {
	for name, signUps := range map[string]*fakeSignUps{
		"a new email":        {startOutcome: usecase.CodeStarted},
		"a verified account": {existing: true},
	} {
		t.Run(name, func(t *testing.T) {
			e := newRegisterEnv(signUps, currentDocs)
			e.emails.limit = 0

			err := e.uc.Execute(t.Context(), validRegistration())

			var limited *usecase.RateLimitedError
			require.ErrorAs(t, err, &limited, "the same answer for every kind of address")
			assert.Equal(t, time.Minute, limited.RetryAfter)
			assert.Zero(t, e.system.calls, "no database work")
			assert.Empty(t, e.outbox.queued, "nothing sent")
		})
	}
}

// A request the contract would reject anyway spends nothing: the budget
// counts sign-up attempts, not typos.
func TestRegister_anInvalidRequestSpendsNothing(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{}, currentDocs)
	in := validRegistration()
	in.Email = "not-an-email"

	require.Error(t, e.uc.Execute(t.Context(), in))

	assert.Empty(t, e.emails.taken())
}

func TestRegister_aLimiterThatCannotDecideFailsClosed(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{startOutcome: usecase.CodeStarted}, currentDocs)
	e.emails.err = errDB

	err := e.uc.Execute(t.Context(), validRegistration())

	require.ErrorIs(t, err, errDB)
	assert.Zero(t, e.system.calls)
	assert.Empty(t, e.outbox.queued)
}

// US-A1c: the resend limit is silent, like its cooldown: the same success,
// no database work, nothing sent.
func TestResendVerification_anEmailOutOfBudgetIsSilent(t *testing.T) {
	signUps, outbox, system := &fakeSignUps{resendOutcome: usecase.CodeResent}, &fakeOutbox{}, &fakeSystem{}
	emails := &fakeLimiter{limit: 0}

	err := newLimitedResend(signUps, outbox, system, emails).Execute(t.Context(), validEmail)

	require.NoError(t, err)
	assert.Equal(t, []string{hexOf(emailHashOf(normalEmail))}, emails.taken())
	assert.Zero(t, system.calls)
	assert.Empty(t, signUps.resends)
	assert.Empty(t, outbox.queued)
}

func TestResendVerification_aLimiterThatCannotDecideFailsClosed(t *testing.T) {
	signUps, outbox, system := &fakeSignUps{resendOutcome: usecase.CodeResent}, &fakeOutbox{}, &fakeSystem{}

	err := newLimitedResend(signUps, outbox, system, &fakeLimiter{err: errDB}).Execute(t.Context(), validEmail)

	require.ErrorIs(t, err, errDB)
	assert.Empty(t, outbox.queued)
}

// A use case with no limit wired is a wiring error, never a silent default.
func TestConstructors_refuseToRunWithoutTheirLimit(t *testing.T) {
	assert.PanicsWithValue(t, "auth: NewRegister needs an Emails RateLimiter", func() {
		usecase.NewRegister(usecase.RegisterDeps{})
	})
	assert.PanicsWithValue(t, "auth: NewResendVerification needs an Emails RateLimiter", func() {
		usecase.NewResendVerification(usecase.ResendDeps{})
	})
	assert.PanicsWithValue(t, "auth: NewRefreshSession needs a Rotations RateLimiter", func() {
		usecase.NewRefreshSession(usecase.RefreshDeps{})
	})
}

// MODULE-auth §5: each token family has its own rotation budget, checked
// once the presented token is proven live (its family cannot be forged, so
// nobody can spend another user's budget with made-up tokens).
func TestRefresh_spendsTheFamilysBudgetOnlyForALiveToken(t *testing.T) {
	e := newRefreshEnv(t)
	parent := e.issued(t, storedToken{family: "family-a"})

	_, err := e.uc.Execute(t.Context(), parent.Reveal())
	require.NoError(t, err)
	_, err = e.uc.Execute(t.Context(), parent.Reveal()) // reuse: refused before the budget
	require.ErrorIs(t, err, usecase.ErrTokenReused)

	assert.Equal(t, []string{"family-a"}, e.rotations.taken())
}

func TestRefresh_aFamilyOutOfBudgetIs429AndRotatesNothing(t *testing.T) {
	e := newRefreshEnv(t)
	e.rotations.limit = 0
	parent := e.issued(t, storedToken{family: "family-a"})

	_, err := e.uc.Execute(t.Context(), parent.Reveal())

	var limited *usecase.RateLimitedError
	require.ErrorAs(t, err, &limited)
	assert.Equal(t, time.Minute, limited.RetryAfter)
	assert.False(t, e.store.row(parent).rotated, "the token stays live: the client retries after the wait")
	assert.False(t, e.store.row(parent).revoked)
}

func TestRefresh_aLimiterThatCannotDecideFailsClosed(t *testing.T) {
	e := newRefreshEnv(t)
	e.rotations.err = errDB
	parent := e.issued(t, storedToken{})

	_, err := e.uc.Execute(t.Context(), parent.Reveal())

	require.ErrorIs(t, err, errDB)
	assert.False(t, e.store.row(parent).rotated)
}
