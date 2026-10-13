package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

func newResend(signUps *fakeSignUps, sender *fakeSender, system *fakeSystem) *usecase.ResendVerification {
	logger, _ := capturedLogs()
	return usecase.NewResendVerification(usecase.ResendDeps{
		System:  system,
		Clock:   fixedClock,
		Entropy: zeroReader{},
		Index:   &sha256Index{},
		Codes:   concatCodes{},
		SignUps: signUps,
		Email:   sender,
		Logger:  logger,
	})
}

func TestResendVerification(t *testing.T) {
	tests := []struct {
		name     string
		outcome  usecase.CodeOutcome
		wantSent []sent
	}{
		{"a live pending sign-up gets a new code", usecase.CodeResent, []sent{{kind: "code", to: normalEmail, code: "000000"}}},
		{"no pending sign-up, verified email or cooldown: silent", usecase.CodeUnchanged, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signUps, sender := &fakeSignUps{resendOutcome: tt.outcome}, &fakeSender{}

			require.NoError(t, newResend(signUps, sender, &fakeSystem{}).Execute(t.Context(), validEmail))

			require.Len(t, signUps.resends, 1)
			assert.Equal(t, emailHashOf(normalEmail), signUps.resends[0].emailHash)
			assert.Equal(t, codeHashOf(normalEmail, "000000"), signUps.resends[0].codeHash)
			assert.Equal(t, now.Add(usecase.CodeTTL), signUps.resends[0].expiresAt)
			assert.Equal(t, tt.wantSent, sender.sent)
		})
	}
}

func TestResendVerification_invalidEmail(t *testing.T) {
	system := &fakeSystem{}

	err := newResend(&fakeSignUps{}, &fakeSender{}, system).Execute(t.Context(), "not-an-email")

	require.ErrorIs(t, err, domain.ErrInvalidEmail)
	assert.Zero(t, system.calls)
}

func TestResendVerification_emailFailureStaysGeneric(t *testing.T) {
	sender := &fakeSender{err: errDB}

	assert.NoError(t, newResend(&fakeSignUps{resendOutcome: usecase.CodeResent}, sender, &fakeSystem{}).Execute(t.Context(), validEmail))
}

func TestResendVerification_databaseErrorsPropagate(t *testing.T) {
	sender := &fakeSender{}

	err := newResend(&fakeSignUps{err: errDB}, sender, &fakeSystem{}).Execute(t.Context(), validEmail)

	require.ErrorIs(t, err, errDB)
	assert.Empty(t, sender.sent)
}
