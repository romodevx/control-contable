package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

func newResend(signUps *fakeSignUps, outbox *fakeOutbox, system *fakeSystem) *usecase.ResendVerification {
	return newLimitedResend(signUps, outbox, system, &fakeLimiter{limit: 100})
}

func newLimitedResend(signUps *fakeSignUps, outbox *fakeOutbox, system *fakeSystem, emails usecase.RateLimiter) *usecase.ResendVerification {
	logger, _ := capturedLogs()
	return usecase.NewResendVerification(usecase.ResendDeps{
		System:  system,
		Clock:   fixedClock,
		Entropy: zeroReader{},
		Index:   &sha256Index{},
		Codes:   concatCodes{},
		SignUps: signUps,
		Outbox:  outbox,
		Emails:  emails,
		Logger:  logger,
	})
}

func TestResendVerification(t *testing.T) {
	tests := []struct {
		name     string
		outcome  usecase.CodeOutcome
		wantSent []queued
	}{
		{"a live pending sign-up gets a new code", usecase.CodeResent, []queued{{kind: "code", to: normalEmail, code: "000000", expiresAt: now.Add(usecase.CodeTTL), inUnit: true}}},
		{"no pending sign-up, verified email or cooldown: silent", usecase.CodeUnchanged, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signUps, outbox := &fakeSignUps{resendOutcome: tt.outcome}, &fakeOutbox{}

			require.NoError(t, newResend(signUps, outbox, &fakeSystem{}).Execute(t.Context(), validEmail))

			require.Len(t, signUps.resends, 1)
			assert.Equal(t, emailHashOf(normalEmail), signUps.resends[0].emailHash)
			assert.Equal(t, codeHashOf(normalEmail, "000000"), signUps.resends[0].codeHash)
			assert.Equal(t, now.Add(usecase.CodeTTL), signUps.resends[0].expiresAt)
			assert.Equal(t, tt.wantSent, outbox.queued, "queued in the unit of work, never sent in the request")
		})
	}
}

func TestResendVerification_invalidEmail(t *testing.T) {
	system := &fakeSystem{}

	err := newResend(&fakeSignUps{}, &fakeOutbox{}, system).Execute(t.Context(), "not-an-email")

	require.ErrorIs(t, err, domain.ErrInvalidEmail)
	assert.Zero(t, system.calls)
}

func TestResendVerification_enqueueFailurePropagates(t *testing.T) {
	outbox := &fakeOutbox{err: errDB}

	err := newResend(&fakeSignUps{resendOutcome: usecase.CodeResent}, outbox, &fakeSystem{}).Execute(t.Context(), validEmail)

	require.ErrorIs(t, err, errDB)
}

func TestResendVerification_databaseErrorsPropagate(t *testing.T) {
	outbox := &fakeOutbox{}

	err := newResend(&fakeSignUps{err: errDB}, outbox, &fakeSystem{}).Execute(t.Context(), validEmail)

	require.ErrorIs(t, err, errDB)
	assert.Empty(t, outbox.queued)
}
