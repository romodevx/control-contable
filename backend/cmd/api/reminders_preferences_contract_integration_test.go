//go:build integration

package main

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of registerDevice, unregisterDevice, getReminderSettings
// and putReminderSettings (M9-05) through the real composition root over
// Postgres as cc_app (claim_push_token, RLS), with module_reminders ON and
// OFF and its prerequisite module_taxpayers OFF.

const (
	// installationA and installationB are this suite's installation ids.
	installationA = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0a01"
	installationB = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0b01"
	// devicesWriteBurst and settingsWriteBurst are §2.7's bursts.
	devicesWriteBurst  = 10
	settingsWriteBurst = 20
	defaultsJSON       = `{"enabled":false,"early_day":14,"on_due_date":true,"send_time":"10:00","show_amounts":false}`
)

// fcmToken is a token-shaped value generated per call: never a literal.
func fcmToken() string { return "fcm:" + rand.Text() + rand.Text() }

func deviceBody(id, platform, token string) string {
	return `{"id":"` + id + `","platform":"` + platform + `","push_token":"` + token + `","app_version":"1.2.0"}`
}

func (e remindersEnv) register(t *testing.T, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("registerDevice"), e.handler, e.send(http.MethodPost, "/v1/devices", body, token))
}

func (e remindersEnv) unregister(t *testing.T, id, token string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("unregisterDevice"), e.handler, e.send(http.MethodDelete, "/v1/devices/"+id, "", token))
}

func (e remindersEnv) getSettings(t *testing.T, token string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("getReminderSettings"), e.handler, e.send(http.MethodGet, "/v1/reminder-settings", "", token))
}

func (e remindersEnv) putSettings(t *testing.T, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("putReminderSettings"), e.handler, e.send(http.MethodPut, "/v1/reminder-settings", body, token))
}

func decodeDevice(t *testing.T, rec *httptest.ResponseRecorder) gen.Device {
	t.Helper()
	var d gen.Device
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &d), rec.Body.String())
	return d
}

func requireDetail(t *testing.T, rec *httptest.ResponseRecorder, code gen.ErrorCode, field, detail string) {
	t.Helper()
	body := decodeError(t, rec)
	assert.Equal(t, code, body.Error.Code)
	require.NotNil(t, body.Error.Details)
	require.NotEmpty(t, *body.Error.Details)
	assert.Equal(t, field, (*body.Error.Details)[0].Field)
	assert.Equal(t, detail, (*body.Error.Details)[0].Code)
}

// MODULE-reminders §2.3, US-R6, US-R10, PR #162 review: register, refresh,
// claim, unregister; the token is never returned and never logged.
func TestContract_reminderDevices(t *testing.T) {
	env := newRemindersEnv(t)
	bea := env.signUp(t, beaEmail, acceptableSample).Tokens.AccessToken
	tokenA, tokenB := fcmToken(), fcmToken()
	var responses []string // every body answered, checked for the tokens at the end

	flagtest.OnOff(t, "module_reminders", func(t *testing.T, f flagtest.Static, on bool) {
		env.apply(t, remindersModules, f)
		if !on {
			// OFF answers 404 before any input check, a malformed body included.
			flagtest.AssertHidden(t, env.register(t, deviceBody(installationA, "ios", tokenA), env.token))
			flagtest.AssertHidden(t, env.register(t, `{"id":"nope"}`, env.token))
			flagtest.AssertHidden(t, env.unregister(t, installationA, env.token))
			flagtest.AssertHidden(t, env.unregister(t, "nope", env.token))
			return
		}

		t.Run("201 new, then 200 refreshed, no token in the body", func(t *testing.T) {
			rec := env.register(t, deviceBody(installationA, "ios", tokenA), env.token)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			created := decodeDevice(t, rec)
			assert.Equal(t, installationA, created.Id.String())
			assert.Equal(t, gen.Ios, created.Platform)
			assert.NotContains(t, rec.Body.String(), "push_token")
			responses = append(responses, rec.Body.String())

			rec = env.register(t, deviceBody(installationA, "ios", tokenA), env.token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			refreshed := decodeDevice(t, rec)
			assert.Equal(t, created.CreatedAt, refreshed.CreatedAt)
			assert.False(t, refreshed.LastSeenAt.Before(created.LastSeenAt))
			responses = append(responses, rec.Body.String())
		})
		t.Run("400 for a malformed token or an unknown field", func(t *testing.T) {
			for _, body := range []string{
				deviceBody(installationA, "ios", tokenA+" "+tokenA),
				strings.Replace(deviceBody(installationA, "ios", tokenA), `{`, `{"token":"`+tokenA+`",`, 1),
				`{"id":"` + installationA + `","platform":"ios"}`,
			} {
				rec := env.register(t, body, env.token)

				assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
				responses = append(responses, rec.Body.String())
			}
		})
		t.Run("401 without a session", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, env.register(t, deviceBody(installationA, "ios", tokenA), "").Code)
			assert.Equal(t, http.StatusUnauthorized, env.unregister(t, installationA, "").Code)
		})
		t.Run("another user cannot remove A's device (404), A's row is intact", func(t *testing.T) {
			rec := env.unregister(t, installationA, bea)
			require.Equal(t, http.StatusNotFound, rec.Code)
			assert.Equal(t, "not_found", string(decodeError(t, rec).Error.Code))

			again := env.register(t, deviceBody(installationA, "ios", tokenA), env.token)
			assert.Equal(t, http.StatusOK, again.Code, "still A's: a refresh, not a create")
		})
		t.Run("B registering A's token claims it: 201, A's row is gone", func(t *testing.T) {
			rec := env.register(t, deviceBody(installationB, "android", tokenA), bea)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			responses = append(responses, rec.Body.String())

			assert.Equal(t, http.StatusNotFound, env.unregister(t, installationA, env.token).Code, "A lost the installation")
		})
		t.Run("204 on unregister, then 404 (done)", func(t *testing.T) {
			require.Equal(t, http.StatusCreated, env.register(t, deviceBody(installationA, "ios", tokenB), env.token).Code)

			rec := env.unregister(t, installationA, env.token)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, http.StatusNotFound, env.unregister(t, installationA, env.token).Code)
		})
		t.Run("400 for a malformed installation id", func(t *testing.T) {
			assert.Equal(t, http.StatusBadRequest, env.unregister(t, "nope", env.token).Code)
		})
	}, "module_auth", "module_taxpayers")

	for _, body := range responses {
		assert.NotContains(t, body, tokenA, "never echoed")
		assert.NotContains(t, body, tokenB, "never echoed")
	}
	assert.NotContains(t, env.logs.String(), tokenA, "never logged")
	assert.NotContains(t, env.logs.String(), tokenB, "never logged")
}

// MODULE-reminders §2.2, §6: defaults until the first save, PUT replaces
// the whole object, the 400 backstops and the 422.
func TestContract_reminderSettings(t *testing.T) {
	env := newRemindersEnv(t)
	saved := `{"enabled":true,"early_day":null,"on_due_date":true,"send_time":"08:30","show_amounts":true}`

	flagtest.OnOff(t, "module_reminders", func(t *testing.T, f flagtest.Static, on bool) {
		env.apply(t, remindersModules, f)
		if !on {
			flagtest.AssertHidden(t, env.getSettings(t, env.token))
			flagtest.AssertHidden(t, env.putSettings(t, saved, env.token))
			flagtest.AssertHidden(t, env.putSettings(t, `{}`, env.token))
			return
		}

		t.Run("200 defaults when never saved", func(t *testing.T) {
			rec := env.getSettings(t, env.token)

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.JSONEq(t, defaultsJSON, rec.Body.String())
		})
		t.Run("PUT replaces, GET answers what was saved", func(t *testing.T) {
			rec := env.putSettings(t, saved, env.token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.JSONEq(t, saved, rec.Body.String())

			assert.JSONEq(t, saved, env.getSettings(t, env.token).Body.String())
		})
		t.Run("422 no_reminder_selected", func(t *testing.T) {
			rec := env.putSettings(t, `{"enabled":true,"early_day":null,"on_due_date":false,"send_time":"10:00","show_amounts":false}`, env.token)

			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			requireDetail(t, rec, gen.ErrorCodeBusinessRuleViolation, "body", "no_reminder_selected")
			assert.JSONEq(t, saved, env.getSettings(t, env.token).Body.String(), "nothing saved")
		})
		t.Run("400: early_day omitted, out of range, bad send_time", func(t *testing.T) {
			for _, body := range []string{
				`{"enabled":true,"on_due_date":true,"send_time":"10:00","show_amounts":false}`,
				`{"enabled":true,"early_day":17,"on_due_date":true,"send_time":"10:00","show_amounts":false}`,
				`{"enabled":true,"early_day":0,"on_due_date":true,"send_time":"10:00","show_amounts":false}`,
				`{"enabled":true,"early_day":14,"on_due_date":true,"send_time":"24:00","show_amounts":false}`,
				`{"enabled":true,"early_day":14,"on_due_date":true,"send_time":"10:00","show_amounts":false,"extra":1}`,
			} {
				rec := env.putSettings(t, body, env.token)

				assert.Equal(t, http.StatusBadRequest, rec.Code, body)
			}
		})
		t.Run("another user still reads the defaults (RLS)", func(t *testing.T) {
			bea := env.signUp(t, beaEmail, acceptableSample).Tokens.AccessToken

			assert.JSONEq(t, defaultsJSON, env.getSettings(t, bea).Body.String())
		})
	}, "module_auth", "module_taxpayers")
}

// §4, US-R9: preferences, not records: a read-only user (trial ended, no
// plan) registers a device and saves settings, never a 402.
func TestContract_remindersForReadOnlyUser(t *testing.T) {
	env := newRemindersEnv(t)
	session := env.signUp(t, beaEmail, acceptableSample)
	env.makeReadOnly(t, session.User.Id.String())
	token := session.Tokens.AccessToken

	assert.Equal(t, http.StatusCreated, env.register(t, deviceBody(installationB, "android", fcmToken()), token).Code)
	assert.Equal(t, http.StatusOK, env.putSettings(t, defaultsJSON, token).Code)
	assert.Equal(t, http.StatusOK, env.getSettings(t, token).Code)
	assert.Equal(t, http.StatusNoContent, env.unregister(t, installationB, token).Code)
}

// §3: module_taxpayers OFF implies module_reminders OFF on every route.
func TestContract_reminderPreferencesImpliedOff(t *testing.T) {
	env := newRemindersEnv(t)
	env.apply(t, remindersModules, flagtest.Static{"module_auth": true, "module_reminders": true})

	flagtest.AssertHidden(t, env.register(t, deviceBody(installationA, "ios", fcmToken()), env.token))
	flagtest.AssertHidden(t, env.unregister(t, installationA, env.token))
	flagtest.AssertHidden(t, env.getSettings(t, env.token))
	flagtest.AssertHidden(t, env.putSettings(t, defaultsJSON, env.token))
}

// §2.7: devices_writes (burst 10) is one bucket for register and
// unregister; reminder_settings_writes has a burst of 20. Spent → 429 +
// Retry-After.
func TestContract_reminderPreferencesRateLimits(t *testing.T) {
	env := newRemindersEnv(t)
	for i := range devicesWriteBurst {
		req := env.send(http.MethodPost, "/v1/devices", deviceBody(installationA, "ios", fcmToken()), env.token)
		if i%2 == 1 {
			req = env.send(http.MethodDelete, "/v1/devices/"+installationA, "", env.token)
		}
		require.NotEqual(t, http.StatusTooManyRequests, env.do(req).Code, "write %d", i+1)
	}
	rec := env.unregister(t, installationA, env.token)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Equal(t, http.StatusTooManyRequests, env.register(t, deviceBody(installationA, "ios", fcmToken()), env.token).Code)

	for i := range settingsWriteBurst {
		require.Equal(t, http.StatusOK, env.do(env.send(http.MethodPut, "/v1/reminder-settings", defaultsJSON, env.token)).Code, "put %d", i+1)
	}
	rec = env.putSettings(t, defaultsJSON, env.token)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "rate_limited", string(decodeError(t, rec).Error.Code))
	assert.Equal(t, http.StatusOK, env.getSettings(t, env.token).Code, "reads have no budget of their own")
}
