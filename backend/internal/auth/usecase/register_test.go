package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

type registerEnv struct {
	uc       *usecase.Register
	system   *fakeSystem
	signUps  *fakeSignUps
	sender   *fakeSender
	index    *sha256Index
	throttle *onceThrottle
	logs     interface{ String() string }
}

func newRegisterEnv(signUps *fakeSignUps, legal fakeLegal) *registerEnv {
	logger, logs := capturedLogs()
	e := &registerEnv{
		system: &fakeSystem{}, signUps: signUps, sender: &fakeSender{},
		index: &sha256Index{}, throttle: &onceThrottle{}, logs: logs,
	}
	e.uc = usecase.NewRegister(usecase.RegisterDeps{
		System:  e.system,
		Clock:   fixedClock,
		Entropy: zeroReader{},
		Index:   e.index,
		Codes:   concatCodes{},
		Legal:   legal,
		SignUps: signUps,
		Email:   e.sender,
		Notices: e.throttle,
		Logger:  logger,
	})
	return e
}

func validRegistration() usecase.RegisterInput {
	return usecase.RegisterInput{
		Email: validEmail,
		AcceptedDocuments: []domain.DocumentVersion{
			{Kind: domain.DocumentPrivacyIntegral, Version: "v1"},
			{Kind: domain.DocumentTerms, Version: "v1"},
		},
	}
}

func TestRegister_newEmailStartsASignUpAndSendsTheCode(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{startOutcome: usecase.CodeStarted}, currentDocs)

	require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))

	require.Len(t, e.signUps.starts, 1)
	got := e.signUps.starts[0]
	assert.Equal(t, emailHashOf(normalEmail), got.emailHash, "keyed on the normalized email")
	assert.Equal(t, codeHashOf(normalEmail, "000000"), got.codeHash, "HMAC of the drawn code, never the code")
	assert.Equal(t, now.Add(usecase.CodeTTL), got.expiresAt)
	assert.Equal(t, usecase.PendingSignUp{
		Documents: []domain.DocumentVersion{
			{Kind: domain.DocumentPrivacyIntegral, Version: "v1"},
			{Kind: domain.DocumentTerms, Version: "v1"},
		},
		AcceptedAt: now,
	}, got.pending)
	assert.Equal(t, []sent{{kind: "code", to: normalEmail, code: "000000"}}, e.sender.sent)
	assert.Equal(t, 1, e.system.calls, "one unit of work")
}

// D7: a live pending sign-up turns register into a resend; the store decides
// (it never replaces a live payload) and the use case only sends what it was
// told to send.
func TestRegister_livePendingSignUpBehavesLikeAResend(t *testing.T) {
	tests := []struct {
		name     string
		outcome  usecase.CodeOutcome
		wantSent int
	}{
		{"outside the cooldown: a new code is sent", usecase.CodeResent, 1},
		{"inside the cooldown: silent, nothing sent", usecase.CodeUnchanged, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newRegisterEnv(&fakeSignUps{startOutcome: tt.outcome}, currentDocs)

			require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))

			assert.Len(t, e.sender.sent, tt.wantSent)
		})
	}
}

// D3 + no enumeration: a verified email answers like any other, creates no
// code and notifies the owner at most once per throttle window. Register
// hashes no password on any path (D8), so no branch pays for argon2.
func TestRegister_verifiedEmailNotifiesTheOwnerOnce(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{existing: true}, currentDocs)

	require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))
	require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))

	assert.Empty(t, e.signUps.starts, "no pending sign-up is created")
	assert.Equal(t, []sent{{kind: "notice", to: normalEmail}}, e.sender.sent, "one notice per window")
	assert.True(t, e.throttle.seen[hexOf(emailHashOf(normalEmail))], "throttled per email hash, never per email")
}

func TestRegister_rejectsInvalidInputBeforeTheDatabase(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*usecase.RegisterInput)
		wantErr error
	}{
		{"email", func(in *usecase.RegisterInput) { in.Email = "ana@localhost" }, domain.ErrInvalidEmail},
		{"no terms", func(in *usecase.RegisterInput) { in.AcceptedDocuments = in.AcceptedDocuments[:1] }, domain.ErrDocumentMissing},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newRegisterEnv(&fakeSignUps{startOutcome: usecase.CodeStarted}, currentDocs)
			in := validRegistration()
			tt.mutate(&in)

			err := e.uc.Execute(t.Context(), in)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Zero(t, e.system.calls, "no unit of work")
			assert.Empty(t, e.sender.sent)
		})
	}
}

func TestRegister_staleDocumentVersionIsRejectedForEveryEmail(t *testing.T) {
	for _, existing := range []bool{false, true} {
		e := newRegisterEnv(&fakeSignUps{existing: existing, startOutcome: usecase.CodeStarted},
			fakeLegal{current: map[domain.DocumentKind]string{domain.DocumentTerms: "v2", domain.DocumentPrivacyIntegral: "v1"}})

		err := e.uc.Execute(t.Context(), validRegistration())

		require.ErrorIs(t, err, domain.ErrDocumentNotCurrent, "existing=%v answers the same", existing)
		assert.Empty(t, e.signUps.starts)
		assert.Empty(t, e.sender.sent)
	}
}

// An email outage never changes the answer (no oracle); it is logged without
// the address or the code.
func TestRegister_emailFailureIsLoggedNotReturned(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{startOutcome: usecase.CodeStarted}, currentDocs)
	e.sender.err = context.DeadlineExceeded

	require.NoError(t, e.uc.Execute(t.Context(), validRegistration()))

	logs := e.logs.String()
	assert.Contains(t, logs, "verification email not sent")
	assert.NotContains(t, logs, normalEmail)
	assert.NotContains(t, logs, "000000")
}

func TestRegister_databaseErrorsPropagate(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{err: errDB}, currentDocs)

	err := e.uc.Execute(t.Context(), validRegistration())

	require.ErrorIs(t, err, errDB)
	assert.Empty(t, e.sender.sent)
}

func TestRegister_legalCatalogErrorsPropagate(t *testing.T) {
	e := newRegisterEnv(&fakeSignUps{}, fakeLegal{err: errDB})

	require.ErrorIs(t, e.uc.Execute(t.Context(), validRegistration()), errDB)
}
