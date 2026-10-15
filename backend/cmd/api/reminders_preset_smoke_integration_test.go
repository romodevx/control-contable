//go:build integration

package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Fixture presets of the reminders smoke. demo_full is the only seeded
// preset holding module_reminders (MODULE-reminders §3, OQ-R8); demo_core
// is the OFF/ON/ON case of the §3 suites, demo_full the ON/ON/ON one.
const (
	// remindersOnlyPreset is the smallest preset that shows the reminders:
	// the flag with its prerequisite (and sign-in) only. The engine is OFF:
	// reminders without amounts (§3 ON/ON/OFF).
	remindersOnlyPreset = "smoke_reminders_only"
	// remindersWithoutTaxpayersPreset turns the reminders and the engine ON
	// with module_taxpayers OFF: §3 implies it OFF (ON/OFF/·).
	remindersWithoutTaxpayersPreset = "smoke_reminders_without_taxpayers"
)

const moduleRemindersKey = "module_reminders"

// remindersSmokeDevice is this suite's installation id.
const remindersSmokeDevice = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0c01"

// remindersUnknownTarget is a route no contract declares.
const remindersUnknownTarget = "/v1/recordatorios"

// remindersStep is one request of the ON sweep.
type remindersStep struct {
	op, method, target, body string
	status                   int
}

// remindersSteps exercise every module_reminders operation in an order where
// each one succeeds: the device is unregistered after it is registered, so
// every preset run starts from no device.
func remindersSteps() []remindersStep {
	return []remindersStep{
		{op: "registerDevice", method: http.MethodPost, target: "/v1/devices",
			body: deviceBody(remindersSmokeDevice, "android", fcmToken()), status: http.StatusCreated},
		{op: "getReminderSettings", method: http.MethodGet, target: "/v1/reminder-settings", status: http.StatusOK},
		{op: "putReminderSettings", method: http.MethodPut, target: "/v1/reminder-settings", body: defaultsJSON,
			status: http.StatusOK},
		{op: "getTaxCalendar", method: http.MethodGet, target: "/v1/tax-calendar?year=2026", status: http.StatusOK},
		{op: "unregisterDevice", method: http.MethodDelete, target: "/v1/devices/" + remindersSmokeDevice,
			status: http.StatusNoContent},
	}
}

// TestPresetSmoke_remindersReachableOnlyWhenThePresetIncludesIt (M9-11):
// each preset is applied as cmd/flags applies it (owner, Postgres), then
// the real API is asked for the reminders operations by a signed-in user.
// With module_reminders and module_taxpayers in the preset, all five
// operations are served (/v1/devices, /v1/devices/{id},
// /v1/reminder-settings, /v1/tax-calendar); otherwise every route the
// router registers for module_reminders answers exactly what an unknown
// route answers, and GET /v1/flags leaves the flag out. The expectation is
// written from the spec, not read from flags.Prerequisites.
func TestPresetSmoke_remindersReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.addPreset(t, remindersOnlyPreset, moduleAuth, moduleTaxpayersKey, moduleRemindersKey)
	env.addPreset(t, remindersWithoutTaxpayersPreset, moduleAuth, moduleTaxEngineKey, moduleRemindersKey)
	env.apply(t, remindersModules, allOn(remindersModules))
	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken

	var stepped []string
	for _, s := range remindersSteps() {
		stepped = append(stepped, s.op)
	}
	require.ElementsMatch(t, ownedOperations(moduleRemindersKey), stepped, "the ON sweep covers every module_reminders operation")
	routed := routedOperations(t, moduleRemindersKey)
	require.Len(t, routed, len(ownedOperations(moduleRemindersKey)), "the router registers one route per operation")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		reachable := slices.Contains(preset.Flags, moduleRemindersKey) && slices.Contains(preset.Flags, moduleTaxpayersKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, reachable, slices.Contains(env.enabledFlags(t), moduleRemindersKey), "GET /v1/flags")
			if reachable {
				sweepRemindersOn(t, env, token)
				return
			}
			assertRemindersHidden(t, env, routed, token)
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full", remindersOnlyPreset, remindersWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
	assert.Contains(t, presetFlags(t, presets, "demo_full"), moduleRemindersKey, "demo_full holds module_reminders (OQ-R8)")
	assert.NotContains(t, presetFlags(t, presets, "demo_core"), moduleRemindersKey, "demo_core leaves it out")
	assert.NotContains(t, presetFlags(t, presets, "demo_monetization"), moduleRemindersKey,
		"demo_monetization leaves it out")
}

// sweepRemindersOn runs every step as token and asserts each one is served.
func sweepRemindersOn(t *testing.T, env contractEnv, token string) {
	t.Helper()
	for _, step := range remindersSteps() {
		rec := env.do(env.send(step.method, step.target, step.body, token))
		assert.Equal(t, step.status, rec.Code, "%s: the module answers: %s", step.op, rec.Body.String())
	}
}

// assertRemindersHidden checks every routed reminders operation answers
// exactly what an unknown route answers, a valid body included.
func assertRemindersHidden(t *testing.T, env contractEnv, routed []gatedOperation, token string) {
	t.Helper()
	bodies := map[string]string{}
	for _, s := range remindersSteps() {
		bodies[s.method+" "+strings.Split(s.target, "?")[0]] = s.body
	}
	for _, op := range routed {
		target := op.path
		if op.method == http.MethodGet && strings.HasSuffix(target, "/tax-calendar") {
			target += "?year=2026"
		}
		body := bodies[op.method+" "+op.path]
		rec := env.do(env.send(op.method, target, body, token))
		if assert.True(t, flagtest.CheckHidden(t, rec), "%s %s: hidden", op.method, target) {
			unknown := env.do(env.send(op.method, remindersUnknownTarget, body, token))
			assert.Equal(t, hiddenAnswerOf(t, unknown), hiddenAnswerOf(t, rec),
				"%s %s: indistinguishable from an unknown route", op.method, target)
		}
	}
}
