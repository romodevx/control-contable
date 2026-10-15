package usecase_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

const otherUserID = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d02"

// deviceID is the push installation id the app sends with logout.
const deviceID = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23"

type logoutEnv struct {
	uc      *usecase.Logout
	tx      *fakeTx
	store   *fakeRotation
	devices *fakeDevices
	logs    interface{ String() string }
}

func newLogoutEnv() *logoutEnv {
	logger, logs := capturedLogs()
	e := &logoutEnv{tx: &fakeTx{}, store: newFakeRotation(), devices: &fakeDevices{}, logs: logs}
	e.uc = usecase.NewLogout(usecase.LogoutDeps{Tx: e.tx, Tokens: e.store, Devices: e.devices, Logger: logger})
	return e
}

// deviceDelete is one DeleteDevice call: the id and the unit of work's user.
type deviceDelete struct {
	installationID string
	txUser         string
}

// fakeDevices records the push devices logout removes.
type fakeDevices struct {
	deletes []deviceDelete
	err     error
}

func (f *fakeDevices) DeleteDevice(ctx context.Context, installationID string) (bool, error) {
	f.deletes = append(f.deletes, deviceDelete{installationID: installationID, txUser: txUser(ctx)})
	return f.err == nil, f.err
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

// M9 security gate F1: the push device goes with the session, in the same
// unit of work as the revocation and as the caller, so a phone that changes
// hands never keeps the previous user's reminders.
func TestLogout_removesThePushDeviceInTheSameUnitOfWork(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)
	e.store.add(token, storedToken{family: "f"})

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: deviceID})

	require.NoError(t, err)
	assert.Equal(t, 1, e.tx.calls, "one transaction for both")
	assert.Equal(t, []deviceDelete{{installationID: deviceID, txUser: newUserID}}, e.devices.deletes)
	assert.True(t, e.store.row(token).revoked)
	assert.NotContains(t, e.logs.String(), deviceID, "the installation id is not logged")
}

func TestLogout_withoutADeviceIDRemovesNoDevice(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)

	require.NoError(t, e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal()}))
	assert.Empty(t, e.devices.deletes)
}

// The device is the bearer's, whatever the refresh token: another user's
// token is still not looked up, but the caller's own device is removed.
func TestLogout_anotherUsersTokenStillRemovesTheCallersDevice(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, otherUserID, 1)
	e.store.add(token, storedToken{})

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: deviceID})

	require.NoError(t, err)
	assert.Equal(t, []deviceDelete{{installationID: deviceID, txUser: newUserID}}, e.devices.deletes)
	assert.False(t, e.store.row(token).revoked, "another user's token is never touched")
	assert.Empty(t, e.store.revokes)
}

// An unknown or another user's device removes nothing and is still a
// success (no oracle); the canonical lower-case id is what is looked up.
func TestLogout_unknownDeviceIsASuccess(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)
	e.uc = usecase.NewLogout(usecase.LogoutDeps{Tx: e.tx, Tokens: e.store, Devices: missingDevices{}})

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: deviceID})

	require.NoError(t, err)
}

func TestLogout_canonicalizesTheDeviceID(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{
		UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: strings.ToUpper(deviceID),
	})

	require.NoError(t, err)
	require.Len(t, e.devices.deletes, 1)
	assert.Equal(t, deviceID, e.devices.deletes[0].installationID)
}

func TestLogout_malformedDeviceIDIsRejectedBeforeAnyWork(t *testing.T) {
	e := newLogoutEnv()
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: "not-a-uuid"})

	require.ErrorIs(t, err, usecase.ErrInvalidDeviceID)
	assert.Zero(t, e.tx.calls)
}

// Atomic: a failed device removal rolls the revocation back and is an
// error, so the app can retry the whole logout.
func TestLogout_deviceFailureRollsTheLogoutBack(t *testing.T) {
	e := newLogoutEnv()
	e.devices.err = errDB
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: deviceID})

	require.ErrorIs(t, err, errDB)
	require.ErrorIs(t, e.tx.workErr, errDB, "the unit of work rolls back")
	assert.NotContains(t, e.logs.String(), `"event":"logout"`)
}

// Without the reminders module wired, a device id is accepted and ignored.
func TestLogout_withoutADeviceStoreIgnoresTheDeviceID(t *testing.T) {
	e := newLogoutEnv()
	e.uc = usecase.NewLogout(usecase.LogoutDeps{Tx: e.tx, Tokens: e.store})
	token := tokenOf(t, newUserID, 1)

	err := e.uc.Execute(t.Context(), usecase.LogoutInput{UserID: newUserID, RefreshToken: token.Reveal(), DeviceID: deviceID})

	require.NoError(t, err)
}

type missingDevices struct{}

func (missingDevices) DeleteDevice(context.Context, string) (bool, error) { return false, nil }
