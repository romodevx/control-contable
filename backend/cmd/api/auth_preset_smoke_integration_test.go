//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// noAuthPreset is a fixture preset that leaves module_auth out: no seeded
// preset does (MODULE-auth §7, every module needs a user), yet the smoke must
// prove that applying one hides the module.
const noAuthPreset = "smoke_without_auth"

// moduleAuthKey is the flag under smoke.
const moduleAuthKey = "module_auth"

// TestPresetSmoke_authReachableOnlyWhenThePresetIncludesIt (M1-14): each
// presentation preset is applied as cmd/flags applies it (owner, Postgres),
// then the real API is asked for every M1 operation. A preset with
// module_auth reaches the module; a preset without it answers 404 for all.
func TestPresetSmoke_authReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.addPreset(t, noAuthPreset, "module_taxpayers")
	env.apply(t, []string{moduleAuthKey}, flagtest.Static{moduleAuthKey: true})
	session := env.signUp(t, signUpEmail, acceptableSample)
	probes := authProbes(session.Tokens.RefreshToken)

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		withAuth := slices.Contains(preset.Flags, moduleAuthKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, withAuth, slices.Contains(env.enabledFlags(t), moduleAuthKey), "GET /v1/flags")
			for _, p := range probes {
				rec := env.do(env.probe(p, session.Tokens.AccessToken))
				if withAuth {
					assert.NotEqual(t, http.StatusNotFound, rec.Code, "%s %s: the module answers", p.op, p.name)
					continue
				}
				assert.True(t, flagtest.CheckHidden(t, rec), "%s %s: hidden", p.op, p.name)
			}
		})
	}
	assert.Subset(t, smoked, []string{"demo_core", "demo_monetization", "demo_full", noAuthPreset},
		"every seeded preset and the fixture are smoked")
}

// addPreset creates a preset of modules as the owner, for this database only.
func (e contractEnv) addPreset(t *testing.T, name string, modules ...string) {
	t.Helper()
	_, err := e.owner.ExecContext(t.Context(),
		"INSERT INTO flag_presets (name, description) VALUES ($1, 'test fixture')", name)
	require.NoError(t, err)
	for _, m := range modules {
		_, err = e.owner.ExecContext(t.Context(),
			"INSERT INTO flag_preset_flags (preset_name, flag_key) VALUES ($1, $2)", name, m)
		require.NoError(t, err)
	}
}

// enabledFlags is what GET /v1/flags answers.
func (e contractEnv) enabledFlags(t *testing.T) []string {
	t.Helper()
	rec := e.do(e.request(http.MethodGet, "/v1/flags"))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body struct {
		Enabled []string `json:"enabled"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Enabled
}
