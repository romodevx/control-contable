//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of logout's device_id (0.14.0, M9 security gate F1)
// through the real composition root over Postgres as cc_app: the push
// device goes with the session in one transaction, only the caller's, and
// with module_reminders ON or OFF.
//
// "Gone" is read through the API: the access token outlives the logout
// (only the refresh family is revoked), and registering the same
// installation again answers 201 when the caller had no row and 200 when
// the row was still the caller's.

// installationLogout is this suite's installation id.
const installationLogout = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0c01"

func logoutDeviceJSON(refreshToken, deviceID string) string {
	return `{"refresh_token":"` + refreshToken + `","device_id":"` + deviceID + `"}`
}

func (e remindersEnv) logout(t *testing.T, body, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("logout"), e.handler, e.bearerPost("/v1/auth/logout", body, accessToken))
}

// registerLogoutDevice registers installationLogout for accessToken.
func (e remindersEnv) registerLogoutDevice(t *testing.T, accessToken string) int {
	t.Helper()
	return e.register(t, deviceBody(installationLogout, "android", fcmToken()), accessToken).Code
}

func TestContract_logoutRemovesThePushDevice(t *testing.T) {
	env := newRemindersEnv(t)

	t.Run("the caller's device goes with the session", func(t *testing.T) {
		cleo := env.signUp(t, cleoEmail, acceptableSample)
		require.Equal(t, http.StatusCreated, env.registerLogoutDevice(t, cleo.Tokens.AccessToken))
		require.Equal(t, http.StatusOK, env.registerLogoutDevice(t, cleo.Tokens.AccessToken), "control: still the caller's")

		rec := env.logout(t, logoutDeviceJSON(cleo.Tokens.RefreshToken, installationLogout), cleo.Tokens.AccessToken)

		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		assert.Equal(t, http.StatusCreated, env.registerLogoutDevice(t, cleo.Tokens.AccessToken), "the row was removed")
		refreshed := env.do(env.post("/v1/auth/refresh", refreshJSON(cleo.Tokens.RefreshToken)))
		assert.Equal(t, http.StatusUnauthorized, refreshed.Code, "and the session revoked")
	})
	t.Run("another user's device id removes nothing and is still 204", func(t *testing.T) {
		require.Equal(t, http.StatusCreated, env.registerLogoutDevice(t, env.token), "claimed from cleo")
		bea := env.signUp(t, beaEmail, acceptableSample)

		rec := env.logout(t, logoutDeviceJSON(bea.Tokens.RefreshToken, installationLogout), bea.Tokens.AccessToken)

		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		assert.Equal(t, http.StatusOK, env.registerLogoutDevice(t, env.token), "still the owner's row")
	})
	t.Run("a malformed device id is 400 and revokes nothing", func(t *testing.T) {
		eli := env.signUp(t, "eli.soto@example.mx", acceptableSample)

		rec := env.logout(t, logoutDeviceJSON(eli.Tokens.RefreshToken, "not-a-uuid"), eli.Tokens.AccessToken)

		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Equal(t, gen.ErrorCodeValidationFailed, decodeError(t, rec).Error.Code)
		refreshed := env.do(env.post("/v1/auth/refresh", refreshJSON(eli.Tokens.RefreshToken)))
		assert.Equal(t, http.StatusOK, refreshed.Code, "the session is still alive")
	})
	t.Run("module_reminders OFF: logout still removes the device", func(t *testing.T) {
		fer := env.signUp(t, "fer.ramos@example.mx", acceptableSample)
		require.Equal(t, http.StatusCreated, env.registerLogoutDevice(t, fer.Tokens.AccessToken))
		env.apply(t, remindersModules, flagtest.Static{"module_auth": true, "module_taxpayers": true})

		rec := env.logout(t, logoutDeviceJSON(fer.Tokens.RefreshToken, installationLogout), fer.Tokens.AccessToken)

		require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		env.apply(t, remindersModules, allOn(remindersModules))
		assert.Equal(t, http.StatusCreated, env.registerLogoutDevice(t, fer.Tokens.AccessToken), "the row was removed")
	})
}
