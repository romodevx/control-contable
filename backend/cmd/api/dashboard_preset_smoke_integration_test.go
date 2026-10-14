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

// Fixture presets of the dashboard smoke. Every seeded preset holds
// module_dashboard with taxpayers, incomes, expenses and the engine
// (MODULES.md), so these prove what applying a preset without one of them
// does.
const (
	// noDashboardPreset keeps every other demo_core module: the dashboard
	// stays hidden and the app keeps its previous Home.
	noDashboardPreset = "smoke_without_dashboard"
	// dashboardWithoutTaxEnginePreset: the engine is not a prerequisite
	// (MODULE-dashboard §3, OQ-D7); the dashboard answers with no tax.
	dashboardWithoutTaxEnginePreset = "smoke_dashboard_without_tax_engine"
	// dashboardWithoutRecordsPreset: incomes and expenses OFF null their
	// figures; the dashboard still answers.
	dashboardWithoutRecordsPreset = "smoke_dashboard_without_records"
	// dashboardWithoutTaxpayersPreset turns the dashboard ON with its
	// prerequisite OFF: MODULE-dashboard §3 implies it OFF.
	dashboardWithoutTaxpayersPreset = "smoke_dashboard_without_taxpayers"
)

const moduleDashboardKey = "module_dashboard"

// TestPresetSmoke_dashboardReachableOnlyWhenThePresetIncludesIt (M6-05):
// each preset is applied as cmd/flags applies it (owner, Postgres), then the
// real API is asked for the dashboard of a real user's taxpayer with records.
// The dashboard answers only when the preset holds it AND module_taxpayers;
// otherwise the hidden-module 404 and GET /v1/flags leaves it out. When it
// answers, each optional module the preset leaves out is null in the body.
// The expectation is written from the spec, not read from flags.Prerequisites.
func TestPresetSmoke_dashboardReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.addPreset(t, noDashboardPreset, moduleAuth, moduleTaxpayersKey, moduleIncomesKey, moduleExpensesKey,
		moduleTaxEngineKey)
	env.addPreset(t, dashboardWithoutTaxEnginePreset, moduleAuth, moduleTaxpayersKey, moduleIncomesKey,
		moduleExpensesKey, moduleDashboardKey)
	env.addPreset(t, dashboardWithoutRecordsPreset, moduleAuth, moduleTaxpayersKey, moduleTaxEngineKey,
		moduleDashboardKey)
	env.addPreset(t, dashboardWithoutTaxpayersPreset, moduleAuth, moduleIncomesKey, moduleExpensesKey,
		moduleTaxEngineKey, moduleDashboardKey)
	all := flagtest.Static{}
	for _, m := range dashboardModules {
		all[m] = true
	}
	env.apply(t, dashboardModules, all)
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	target := dashboardTarget(ana.taxpayer, "")

	require.ElementsMatch(t, []string{"getDashboard"}, ownedOperations(moduleDashboardKey),
		"every module_dashboard operation is probed")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		holds := func(module string) bool {
			return slices.Contains(preset.Flags, module) && slices.Contains(preset.Flags, moduleTaxpayersKey)
		}
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			reachable := holds(moduleDashboardKey)
			assert.Equal(t, reachable, slices.Contains(env.enabledFlags(t), moduleDashboardKey), "GET /v1/flags")
			rec := env.do(env.send(http.MethodGet, target, "", ana.token))
			if !reachable {
				assert.True(t, flagtest.CheckHidden(t, rec), "getDashboard: hidden")
				return
			}
			body := decodeDashboard(t, rec)
			assert.Equal(t, holds(moduleTaxEngineKey), string(body.Estimate) != "null", "estimate iff the engine is ON")
			assert.Equal(t, holds(moduleIncomesKey), body.Month.IncomeCents != nil, "income iff incomes is ON")
			assert.Equal(t, holds(moduleExpensesKey), body.Month.ExpenseCents != nil, "expense iff expenses is ON")
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full", noDashboardPreset, dashboardWithoutTaxEnginePreset,
		dashboardWithoutRecordsPreset, dashboardWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
}
