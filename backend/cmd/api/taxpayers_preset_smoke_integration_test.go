//go:build integration

package main

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// noTaxpayersPreset is a fixture preset that keeps module_auth and leaves
// module_taxpayers out: every seeded preset includes it (MODULES.md), yet the
// smoke must prove that applying one hides the module.
const noTaxpayersPreset = "smoke_without_taxpayers"

// moduleTaxpayersKey is the flag under smoke.
const moduleTaxpayersKey = "module_taxpayers"

// moduleAuth stays in the fixture preset so the signed-in user stays valid.
const moduleAuth = "module_auth"

// taxpayerProbe is one request to an M2 operation that the module answers
// with something other than 404 every time it is sent, so the same probes
// can run once per preset. Destructive operations are sent in a form the
// module rejects before writing (no confirm, a malformed period).
type taxpayerProbe struct{ op, method, suffix, body string }

// taxpayerProbes target the taxpayer at /v1/taxpayers/{id}; suffixes that
// start with "!" are absolute paths.
func taxpayerProbes() []taxpayerProbe {
	return []taxpayerProbe{
		{"listRegimes", http.MethodGet, "!/v1/regimes?person_type=fisica", ""},
		{"listTaxpayers", http.MethodGet, "!/v1/taxpayers", ""},
		{"createTaxpayer", http.MethodPost, "!/v1/taxpayers", `{}`},
		{"getTaxpayer", http.MethodGet, "", ""},
		{"updateTaxpayer", http.MethodPatch, "", `{"display_name":"Ana"}`},
		{"deleteTaxpayer", http.MethodDelete, "", ""},
		{"archiveTaxpayer", http.MethodPost, "/archive", ""},
		{"unarchiveTaxpayer", http.MethodPost, "/unarchive", ""},
		{"addTaxpayerRegime", http.MethodPost, "/regimes", `{"regime_code":"606"}`},
		{"updateTaxpayerRegime", http.MethodPatch, "/regimes/612", `{"options":{}}`},
		{"archiveTaxpayerRegime", http.MethodDelete, "/regimes/612?end_period=not-a-period", ""},
	}
}

func (p taxpayerProbe) target(taxpayerID string) string {
	if len(p.suffix) > 0 && p.suffix[0] == '!' {
		return p.suffix[1:]
	}
	return "/v1/taxpayers/" + taxpayerID + p.suffix
}

// TestPresetSmoke_taxpayersReachableOnlyWhenThePresetIncludesIt (M2-11):
// each presentation preset is applied as cmd/flags applies it (owner,
// Postgres), then the real API is asked for every M2 operation with a real
// user and taxpayer. A preset with module_taxpayers reaches the module; a
// preset without it answers the hidden-module 404 for all 11.
func TestPresetSmoke_taxpayersReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.addPreset(t, noTaxpayersPreset, moduleAuth)
	env.apply(t, []string{moduleAuth, moduleTaxpayersKey},
		flagtest.Static{moduleAuth: true, moduleTaxpayersKey: true})
	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", regimeJSON("612")), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	taxpayerID := decodeTaxpayer(t, rec).Id.String()

	probes := taxpayerProbes()
	require.ElementsMatch(t, ownedOperations(moduleTaxpayersKey), probeOps(probes),
		"every module_taxpayers operation is probed")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		withTaxpayers := slices.Contains(preset.Flags, moduleTaxpayersKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, withTaxpayers, slices.Contains(env.enabledFlags(t), moduleTaxpayersKey), "GET /v1/flags")
			for _, p := range probes {
				rec := env.do(env.send(p.method, p.target(taxpayerID), p.body, token))
				if withTaxpayers {
					assert.NotEqual(t, http.StatusNotFound, rec.Code, "%s: the module answers: %s", p.op, rec.Body.String())
					continue
				}
				assert.True(t, flagtest.CheckHidden(t, rec), "%s: hidden", p.op)
			}
		})
	}
	assert.Subset(t, smoked, []string{"demo_core", "demo_monetization", "demo_full", noTaxpayersPreset},
		"every seeded preset and the fixture are smoked")
}

func ownedOperations(module string) []string {
	var owned []string
	for op, m := range gen.OperationModules() {
		if m == module {
			owned = append(owned, op)
		}
	}
	return owned
}

func probeOps(probes []taxpayerProbe) []string {
	ops := make([]string, 0, len(probes))
	for _, p := range probes {
		ops = append(ops, p.op)
	}
	return ops
}
