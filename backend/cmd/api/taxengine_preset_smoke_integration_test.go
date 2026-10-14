//go:build integration

package main

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Fixture presets of the tax engine smoke. Every seeded preset holds
// module_tax_engine and module_taxpayers together (MODULES.md), so these
// prove what applying a preset without one of them does.
const (
	// noTaxEnginePreset keeps auth and taxpayers and leaves the engine out.
	noTaxEnginePreset = "smoke_without_tax_engine"
	// taxEngineWithoutTaxpayersPreset turns the engine ON with its
	// prerequisite OFF: MODULE-tax-engine §7 implies it OFF.
	taxEngineWithoutTaxpayersPreset = "smoke_tax_engine_without_taxpayers"
)

const moduleTaxEngineKey = "module_tax_engine"

// taxEngineProbe is one request to an M5 operation that the module answers
// with something other than 404 every time it is sent, so the same probes
// run once per preset. The PUT replaces the inputs with an empty set.
type taxEngineProbe struct{ op, method, suffix, body string }

func taxEngineProbes() []taxEngineProbe {
	return []taxEngineProbe{
		{"getTaxpayerEstimate", http.MethodGet, "/estimates?period=2026-03", ""},
		{"getEstimateBreakdown", http.MethodGet, "/estimates/2026-03/breakdown?regime=612", ""},
		{"getTaxInputs", http.MethodGet, "/regimes/612/tax-inputs?fiscal_year=2026", ""},
		{"putTaxInputs", http.MethodPut, "/regimes/612/tax-inputs?fiscal_year=2026", `{"values":[]}`},
	}
}

func taxEngineProbeOps(probes []taxEngineProbe) []string {
	ops := make([]string, 0, len(probes))
	for _, p := range probes {
		ops = append(ops, p.op)
	}
	return ops
}

// TestPresetSmoke_taxEngineReachableOnlyWhenThePresetIncludesIt (M5-28):
// each preset is applied as cmd/flags applies it (owner, Postgres), then the
// real API is asked for all 4 M5 operations with a real user and taxpayer.
// The engine answers only when the preset holds it AND module_taxpayers (its
// prerequisite); otherwise every operation answers the hidden-module 404 and
// GET /v1/flags leaves it out. The expectation is written from the spec, not
// read from flags.Prerequisites, so a missing declaration fails here.
func TestPresetSmoke_taxEngineReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.addPreset(t, noTaxEnginePreset, moduleAuth, moduleTaxpayersKey)
	env.addPreset(t, taxEngineWithoutTaxpayersPreset, moduleAuth, moduleTaxEngineKey)
	env.apply(t, []string{moduleAuth, moduleTaxpayersKey, moduleTaxEngineKey},
		flagtest.Static{moduleAuth: true, moduleTaxpayersKey: true, moduleTaxEngineKey: true})

	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", `{"regime_code":"612","start_period":"2025-01"}`), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	base := "/v1/taxpayers/" + decodeTaxpayer(t, rec).Id.String()

	probes := taxEngineProbes()
	require.ElementsMatch(t, ownedOperations(moduleTaxEngineKey), taxEngineProbeOps(probes),
		"every module_tax_engine operation is probed")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		reachable := slices.Contains(preset.Flags, moduleTaxEngineKey) && slices.Contains(preset.Flags, moduleTaxpayersKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, reachable, slices.Contains(env.enabledFlags(t), moduleTaxEngineKey), "GET /v1/flags")
			for _, p := range probes {
				rec := env.do(env.send(p.method, base+p.suffix, p.body, token))
				if reachable {
					assert.NotEqual(t, http.StatusNotFound, rec.Code, "%s: the module answers: %s", p.op, rec.Body.String())
					continue
				}
				assert.True(t, flagtest.CheckHidden(t, rec), "%s: hidden", p.op)
			}
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full", noTaxEnginePreset, taxEngineWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
}
