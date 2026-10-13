package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

const otherUserID = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d02"

type logoutEnv struct {
	uc    *usecase.Logout
	tx    *fakeTx
	store *fakeRotation
	logs  interface{ String() string }
}

func newLogoutEnv() *logoutEnv {
	logger, logs := capturedLogs()
	e := &logoutEnv{tx: &fakeTx{}, store: newFakeRotation(), logs: logs}
	e.uc = usecase.NewLogout(usecase.LogoutDeps{Tx: e.tx, Tokens: e.store, Logger: logger})
	return e
}

func tokenOf(t *testing.T, userID string, seed byte) domain.RefreshToken {
	t.Helper()
	id, err := domain.ParseUserID(userID)
	require.NoError(t, err)
	token, err := domain.NewRefreshToken(id, &seqReader{next: seed})
	require.NoError(t, err)
	return token
}

// US-A2c: logout revokes (never deletes) the caller's token family, as the
// caller, so the rotated history stays as reuse evidence.
func TestLogout_revokesTheCallersFamily(t *testing.T) {
	e := newLogoutEnv()
	first := tokenOf(t, newUserID, 1)
	e.store.add(first, storedToken{rotated: true, family: "f"})
	current := tokenOf(t, newUserID, 2)
	e.store.add(current, storedToken{family: "f"})

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: current.Reveal()})

	require.NoError(t, err)
	assert.Equal(t, newUserID, e.tx.userID, "as the bearer's user")
	assert.Equal(t, []usecase.RevokeReason{usecase.RevokeLogout}, e.store.revokes)
	for _, token := range []domain.RefreshToken{first, current} {
		row := e.store.row(token)
		require.NotNil(t, row, "revoked, not deleted")
		assert.True(t, row.revoked)
		assert.Equal(t, usecase.RevokeLogout, row.reason)
	}
	assert.Contains(t, e.logs.String(), `"event":"logout"`)
	assert.NotContains(t, e.logs.String(), current.Reveal())
}

// Idempotent and silent: an unknown or already revoked token is a success,
// and another user's token is never looked up (its state is not revealed).
func TestLogout_isIdempotentAndNeverTouchesAnotherUsersToken(t *testing.T) {
	t.Run("already revoked", func(t *testing.T) {
		e := newLogoutEnv()
		token := tokenOf(t, newUserID, 1)
		e.store.add(token, storedToken{revoked: true, reason: usecase.RevokeReuseDetected})

		require.NoError(t, e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal()}))
		assert.Equal(t, usecase.RevokeReuseDetected, e.store.row(token).reason, "the first revocation stands")
	})
	t.Run("unknown", func(t *testing.T) {
		e := newLogoutEnv()
		token := tokenOf(t, newUserID, 1)

		require.NoError(t, e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal()}))
	})
	t.Run("another user's token", func(t *testing.T) {
		e := newLogoutEnv()
		token := tokenOf(t, otherUserID, 1)
		e.store.add(token, storedToken{})

		require.NoError(t, e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal()}))
		assert.Zero(t, e.tx.calls, "no lookup at all")
		assert.False(t, e.store.row(token).revoked)
	})
}

func TestLogout_malformedTokenIsRejected(t *testing.T) {
	e := newLogoutEnv()

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: "garbage"})

	require.ErrorIs(t, err, usecase.ErrRefreshRejected)
	assert.Zero(t, e.tx.calls)
}

// The caller comes from a verified access token; anything else fails closed.
func TestLogout_withoutAValidCallerIsUnauthenticated(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: "", RefreshToken: token.Reveal()})

	require.ErrorIs(t, err, usecase.ErrInvalidAccessToken)
	assert.Zero(t, e.tx.calls)
}

func TestLogout_storeFailureIsAnError(t *testing.T) {
	e := newLogoutEnv()
	e.uc = usecase.NewLogout(usecase.LogoutDeps{Tx: e.tx, Tokens: failingRevoker{}})
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal()})

	require.ErrorIs(t, err, errDB)
}

type failingRevoker struct{}

func (failingRevoker) RevokeFamily(context.Context, domain.TokenHash, usecase.RevokeReason) error {
	return errDB
}
