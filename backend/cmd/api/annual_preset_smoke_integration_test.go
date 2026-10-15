//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Fixture presets of the annual smoke. demo_full is the only seeded preset
// holding module_annual (MODULE-annual §6, OQ-A7), so these prove what
// applying a preset without one of its prerequisites does.
const (
	// annualOnlyPreset is the smallest preset that shows the annual
	// estimate: the flag with its prerequisites (and sign-in) only.
	annualOnlyPreset = "smoke_annual_only"
	// annualWithoutTaxEnginePreset turns the annual ON with
	// module_tax_engine OFF: MODULE-annual §6 implies it OFF.
	annualWithoutTaxEnginePreset = "smoke_annual_without_tax_engine"
	// annualWithoutTaxpayersPreset turns the annual and the engine ON with
	// module_taxpayers OFF: implied OFF directly and through the engine.
	annualWithoutTaxpayersPreset = "smoke_annual_without_taxpayers"
)

const moduleAnnualKey = "module_annual"

// annualStep is one request of the ON sweep. Its target and body may name
// {statementId} / {deductionId}, filled with the ids an earlier step's
// capture read from its response.
type annualStep struct {
	op, method, target, body string
	status                   int
	capture                  string // the placeholder this step's "id" fills
}

// annualSteps exercise every module_annual operation in an order where each
// one succeeds: the edit, delete and restore of a record follow its create.
func annualSteps(taxpayerID string) []annualStep {
	base := "/v1/taxpayers/" + taxpayerID
	return []annualStep{
		{op: "getAnnualEstimate", method: http.MethodGet, target: annualEstimateTarget(taxpayerID, 2026), status: http.StatusOK},
		{op: "getAnnualBreakdown", method: http.MethodGet,
			target: annualEstimateTarget(taxpayerID, 2026) + "/breakdown?section=pf_art152", status: http.StatusOK},
		{op: "putAnnualInputs", method: http.MethodPut, target: base + "/annual-inputs/2026",
			body:   `{"prior_year_accumulable_income_cents":50000000,"other_exempt_income_cents":0,"told_employer_will_file":false}`,
			status: http.StatusOK},
		{op: "getAnnualInputs", method: http.MethodGet, target: base + "/annual-inputs/2026", status: http.StatusOK},
		{op: "createSalaryStatement", method: http.MethodPost, target: base + "/salary-statements",
			body:   `{"fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`,
			status: http.StatusCreated, capture: "{statementId}"},
		{op: "listSalaryStatements", method: http.MethodGet, target: base + "/salary-statements?fiscal_year=2026",
			status: http.StatusOK},
		{op: "updateSalaryStatement", method: http.MethodPatch, target: base + "/salary-statements/{statementId}",
			body: `{"left_before_year_end":true}`, status: http.StatusOK},
		{op: "deleteSalaryStatement", method: http.MethodDelete, target: base + "/salary-statements/{statementId}",
			status: http.StatusNoContent},
		{op: "createPersonalDeduction", method: http.MethodPost, target: base + "/personal-deductions",
			body:   `{"category":"medical","amount_cents":120000,"paid_on":"2026-05-10","payment_method":"electronic","beneficiary":"self"}`,
			status: http.StatusCreated, capture: "{deductionId}"},
		{op: "listPersonalDeductions", method: http.MethodGet, target: base + "/personal-deductions?fiscal_year=2026",
			status: http.StatusOK},
		{op: "updatePersonalDeduction", method: http.MethodPatch, target: base + "/personal-deductions/{deductionId}",
			body: `{"amount_cents":130000}`, status: http.StatusOK},
		{op: "deletePersonalDeduction", method: http.MethodDelete, target: base + "/personal-deductions/{deductionId}",
			status: http.StatusNoContent},
		{op: "restorePersonalDeduction", method: http.MethodPost,
			target: base + "/personal-deductions/{deductionId}/restore", status: http.StatusOK},
		{op: "putFiledPayments", method: http.MethodPut, target: base + "/regimes/612/filed-payments/2026",
			body: `{"months":[{"month":3,"amount_cents":200000}]}`, status: http.StatusOK},
		{op: "getFiledPayments", method: http.MethodGet, target: base + "/regimes/612/filed-payments/2026",
			status: http.StatusOK},
	}
}

// sweepAnnualOn runs every step as token and asserts each one is served.
func sweepAnnualOn(t *testing.T, env contractEnv, taxpayerID, token string) {
	t.Helper()
	ids := map[string]string{}
	fill := func(s string) string {
		for placeholder, id := range ids {
			s = strings.ReplaceAll(s, placeholder, id)
		}
		return s
	}
	for _, step := range annualSteps(taxpayerID) {
		rec := env.do(env.send(step.method, fill(step.target), step.body, token))
		if !assert.Equal(t, step.status, rec.Code, "%s: the module answers: %s", step.op, rec.Body.String()) {
			continue
		}
		if step.capture != "" {
			var created struct {
				ID string `json:"id"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created), step.op)
			require.NotEmpty(t, created.ID, step.op)
			ids[step.capture] = created.ID
		}
	}
}

// recordingMux records the patterns the generated server registers.
type recordingMux struct {
	*http.ServeMux
	patterns *[]string
}

func (m recordingMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	*m.patterns = append(*m.patterns, pattern)
	m.ServeMux.HandleFunc(pattern, handler)
}

// routedOperations are the requests of every route the generated router
// registers for module, read from the router itself (not a hand list), each
// path filled with valid sample values.
func routedOperations(t *testing.T, module string) []gatedOperation {
	t.Helper()
	var patterns []string
	gen.HandlerFromMux(gen.NewStrictHandler(nil, nil), recordingMux{ServeMux: http.NewServeMux(), patterns: &patterns})
	var ops []gatedOperation
	for _, pattern := range patterns {
		if m, ok := gen.RouteModule(pattern); !ok || m != module {
			continue
		}
		method, path, _ := strings.Cut(pattern, " ")
		ops = append(ops, gatedOperation{module: module, method: method, path: fillPath(path)})
	}
	return ops
}

// TestPresetSmoke_annualReachableOnlyWhenThePresetIncludesIt (ANN-22): each
// preset is applied as cmd/flags applies it (owner, Postgres), then the real
// API is asked for the annual operations with a real user's TX-NET-01
// taxpayer. With module_annual, module_tax_engine and module_taxpayers in
// the preset, all 15 operations are served; otherwise every route the
// router registers for module_annual answers exactly what an unknown route
// answers, and GET /v1/flags leaves the flag out. The expectation is
// written from the spec, not read from flags.Prerequisites.
func TestPresetSmoke_annualReachableOnlyWhenThePresetIncludesIt(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.addPreset(t, annualOnlyPreset, moduleAuth, moduleTaxpayersKey, moduleTaxEngineKey, moduleAnnualKey)
	env.addPreset(t, annualWithoutTaxEnginePreset, moduleAuth, moduleTaxpayersKey, moduleIncomesKey,
		moduleExpensesKey, moduleAnnualKey)
	env.addPreset(t, annualWithoutTaxpayersPreset, moduleAuth, moduleIncomesKey, moduleExpensesKey,
		moduleTaxEngineKey, moduleAnnualKey)
	env.apply(t, annualModules, allOn(annualModules))
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)

	var stepped []string
	for _, s := range annualSteps(ana.taxpayer) {
		stepped = append(stepped, s.op)
	}
	require.ElementsMatch(t, ownedOperations(moduleAnnualKey), stepped, "the ON sweep covers every module_annual operation")
	routed := routedOperations(t, moduleAnnualKey)
	require.Len(t, routed, len(ownedOperations(moduleAnnualKey)), "the router registers one route per operation")

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		reachable := slices.Contains(preset.Flags, moduleAnnualKey) &&
			slices.Contains(preset.Flags, moduleTaxEngineKey) && slices.Contains(preset.Flags, moduleTaxpayersKey)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL

			assert.Equal(t, reachable, slices.Contains(env.enabledFlags(t), moduleAnnualKey), "GET /v1/flags")
			if reachable {
				sweepAnnualOn(t, env, ana.taxpayer, ana.token)
				return
			}
			assertAnnualHidden(t, env, routed, ana.taxpayer, ana.token)
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full", annualOnlyPreset, annualWithoutTaxEnginePreset,
		annualWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
	assert.Contains(t, presetFlags(t, presets, "demo_full"), moduleAnnualKey, "demo_full holds module_annual (OQ-A7)")
	assert.NotContains(t, presetFlags(t, presets, "demo_core"), moduleAnnualKey, "demo_core leaves it out")
}

// assertAnnualHidden checks every routed annual operation, on the user's own
// taxpayer, answers exactly what an unknown route answers.
func assertAnnualHidden(t *testing.T, env contractEnv, routed []gatedOperation, taxpayerID, token string) {
	t.Helper()
	unknownTarget := "/v1/taxpayers/" + taxpayerID + "/anual"
	for _, op := range routed {
		target := strings.Replace(op.path, sampleID, taxpayerID, 1) // {taxpayerId} comes first
		body := ""
		if op.method == http.MethodPost || op.method == http.MethodPut || op.method == http.MethodPatch {
			body = `{}`
		}
		rec := env.do(env.send(op.method, target, body, token))
		if assert.True(t, flagtest.CheckHidden(t, rec), "%s %s: hidden", op.method, target) {
			unknown := env.do(env.send(op.method, unknownTarget, body, token))
			assert.Equal(t, hiddenAnswerOf(t, unknown), hiddenAnswerOf(t, rec),
				"%s %s: indistinguishable from an unknown route", op.method, target)
		}
	}
}

func presetFlags(t *testing.T, presets []flags.Preset, name string) []string {
	t.Helper()
	i := slices.IndexFunc(presets, func(p flags.Preset) bool { return p.Name == name })
	require.GreaterOrEqual(t, i, 0, "preset %s is seeded", name)
	return presets[i].Flags
}
