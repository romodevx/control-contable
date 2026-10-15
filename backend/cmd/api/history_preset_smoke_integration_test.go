//go:build integration

package main

import (
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Fixture presets of the history smoke. demo_full is the only seeded preset
// holding module_history (MODULE-history §3, OQ-H6), so these prove what
// applying a preset without one of its prerequisites does.
const (
	// historyOnlyPreset is the smallest preset that shows History: the flag
	// with its prerequisites and nothing else.
	historyOnlyPreset = "smoke_history_only"
	// historyWithoutTaxEnginePreset turns History ON with module_tax_engine
	// OFF: MODULE-history §3 implies it OFF.
	historyWithoutTaxEnginePreset = "smoke_history_without_tax_engine"
	// historyWithoutTaxpayersPreset turns History and the engine ON with
	// module_taxpayers OFF: implied OFF through the engine.
	historyWithoutTaxpayersPreset = "smoke_history_without_taxpayers"
)

const (
	moduleHistoryKey = "module_history"
	// historySmokePeriod is a closed month of TX-NET-01 before estToday.
	historySmokePeriod = "2026-03"
)

// TestPresetSmoke_historyReachableOnlyWhenThePresetIncludesIt (M8-06): each
// preset is applied as cmd/flags applies it (owner, Postgres), then the real
// API is asked for both M8 operations with a real user's TX-NET-01 taxpayer.
// History answers only when the preset holds module_history, module_tax_engine
// and module_taxpayers; otherwise each operation answers exactly what an
// unknown route answers, and GET /v1/flags leaves the flag out. The
// expectation is written from the spec, not read from flags.Prerequisites.
func TestPresetSmoke_historyReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.addPreset(t, historyOnlyPreset, moduleAuth, moduleTaxpayersKey, moduleTaxEngineKey, moduleHistoryKey)
	env.addPreset(t, historyWithoutTaxEnginePreset, moduleAuth, moduleTaxpayersKey, moduleIncomesKey,
		moduleExpensesKey, moduleHistoryKey)
	env.addPreset(t, historyWithoutTaxpayersPreset, moduleAuth, moduleIncomesKey, moduleExpensesKey,
		moduleTaxEngineKey, moduleHistoryKey)
	env.apply(t, historyModules, allOn(historyModules))
	h := historyEnv{contractEnv: env, c: contracttest.Load(t)}
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	viewed := h.estimate(t, ana, historySmokePeriod).body(t)
	unknownTarget := "/v1/taxpayers/" + ana.taxpayer + "/historia"

	probes := []struct {
		op, method, target, body string
		status                   int
	}{
		{"getHistory", http.MethodGet, historyTarget(ana.taxpayer, "fiscal_year=2026"), "", http.StatusOK},
		{"markHistoryViewed", http.MethodPost, viewedTarget(ana.taxpayer, historySmokePeriod), viewed, http.StatusNoContent},
	}
	var probed []string
	for _, p := range probes {
		probed = append(probed, p.op)
	}
	require.ElementsMatch(t, ownedOperations(moduleHistoryKey), probed, "every module_history operation is probed")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		reachable := slices.Contains(preset.Flags, moduleHistoryKey) &&
			slices.Contains(preset.Flags, moduleTaxEngineKey) && slices.Contains(preset.Flags, moduleTaxpayersKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, reachable, slices.Contains(env.enabledFlags(t), moduleHistoryKey), "GET /v1/flags")
			for _, p := range probes {
				rec := env.do(env.send(p.method, p.target, p.body, ana.token))
				if reachable {
					assert.Equal(t, p.status, rec.Code, "%s: the module answers: %s", p.op, rec.Body.String())
					continue
				}
				if assert.True(t, flagtest.CheckHidden(t, rec), "%s: hidden", p.op) {
					unknown := env.do(env.send(p.method, unknownTarget, p.body, ana.token))
					assert.Equal(t, hiddenAnswerOf(t, unknown), hiddenAnswerOf(t, rec),
						"%s: indistinguishable from an unknown route", p.op)
				}
			}
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full", historyOnlyPreset, historyWithoutTaxEnginePreset,
		historyWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
}
