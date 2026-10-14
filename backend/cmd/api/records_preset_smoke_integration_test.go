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

// Fixture presets of the record-module smoke. Every seeded preset holds
// module_incomes, module_expenses and module_taxpayers together (MODULES.md),
// so these prove what applying a preset without one of them does.
const (
	// noRecordsPreset keeps auth and taxpayers and leaves both record modules out.
	noRecordsPreset = "smoke_without_records"
	// incomesOnlyPreset and expensesOnlyPreset prove the two modules are independent.
	incomesOnlyPreset  = "smoke_incomes_only"
	expensesOnlyPreset = "smoke_expenses_only"
	// recordsWithoutTaxpayersPreset turns both record modules ON with their
	// prerequisite OFF: MODULE-incomes §6 and MODULE-expenses §6 imply them OFF.
	recordsWithoutTaxpayersPreset = "smoke_records_without_taxpayers"
)

const (
	moduleIncomesKey  = "module_incomes"
	moduleExpensesKey = "module_expenses"
)

const (
	smokeIncomeID  = "0199d4a3-0000-7000-8000-00000000a001"
	smokeExpenseID = "0199d4a3-0000-7000-8000-00000000b001"
)

// recordProbe is one request to an M3 or M4 operation that the module
// answers with something other than 404 every time it is sent, in order:
// delete precedes restore, so a preset that reaches the module leaves the
// record live for the next one. create is sent empty (400 before writing).
type recordProbe struct{ op, method, suffix, body string }

func incomeProbes() []recordProbe {
	return []recordProbe{
		{"listIncomes", http.MethodGet, "/incomes?period=2026-09", ""},
		{"getIncomeDefaults", http.MethodGet, "/incomes/defaults?regime=612&counterparty_type=persona_moral", ""},
		{"createIncome", http.MethodPost, "/incomes", `{}`},
		{"getIncome", http.MethodGet, "/incomes/" + smokeIncomeID, ""},
		{"updateIncome", http.MethodPatch, "/incomes/" + smokeIncomeID, `{"description":"Smoke"}`},
		{"deleteIncome", http.MethodDelete, "/incomes/" + smokeIncomeID, ""},
		{"restoreIncome", http.MethodPost, "/incomes/" + smokeIncomeID + "/restore", ""},
	}
}

func expenseProbes() []recordProbe {
	return []recordProbe{
		{"listExpenseCategories", http.MethodGet, "/expense-categories?regime=612", ""},
		{"listExpenses", http.MethodGet, "/expenses?period=2026-09", ""},
		{"createExpense", http.MethodPost, "/expenses", `{}`},
		{"getExpense", http.MethodGet, "/expenses/" + smokeExpenseID, ""},
		{"updateExpense", http.MethodPatch, "/expenses/" + smokeExpenseID, `{"deductible":true}`},
		{"deleteExpense", http.MethodDelete, "/expenses/" + smokeExpenseID, ""},
		{"restoreExpense", http.MethodPost, "/expenses/" + smokeExpenseID + "/restore", ""},
	}
}

// builtOwnedOperations are module's operations minus the not-built stubs,
// which answer the hidden-module 404 whatever the preset (router NotBuilt).
// The withholding previews (contract 0.10.0) join the probes when WH-06 and
// WH-07 build them (ADR-0008).
func builtOwnedOperations(module string) []string {
	notBuilt := notBuiltOperations()
	return slices.DeleteFunc(ownedOperations(module), func(op string) bool {
		return slices.Contains(notBuilt, op)
	})
}

func recordProbeOps(probes []recordProbe) []string {
	ops := make([]string, 0, len(probes))
	for _, p := range probes {
		ops = append(ops, p.op)
	}
	return ops
}

// TestPresetSmoke_recordModulesReachableOnlyWhenThePresetIncludesThem
// (M3-10, M4-11): each preset is applied as cmd/flags applies it (owner,
// Postgres), then the real API is asked for all 7 operations of each record
// module with a real user, taxpayer, income and expense. A module answers
// only when the preset holds it AND module_taxpayers (its prerequisite);
// otherwise every operation answers the hidden-module 404 and GET /v1/flags
// leaves it out. The expectation is written from the specs, not read from
// flags.Prerequisites, so a missing declaration fails here.
func TestPresetSmoke_recordModulesReachableOnlyWhenThePresetIncludesThem(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.addPreset(t, noRecordsPreset, moduleAuth, moduleTaxpayersKey)
	env.addPreset(t, incomesOnlyPreset, moduleAuth, moduleTaxpayersKey, moduleIncomesKey)
	env.addPreset(t, expensesOnlyPreset, moduleAuth, moduleTaxpayersKey, moduleExpensesKey)
	env.addPreset(t, recordsWithoutTaxpayersPreset, moduleAuth, moduleIncomesKey, moduleExpensesKey)
	env.apply(t, []string{moduleAuth, moduleTaxpayersKey, moduleIncomesKey, moduleExpensesKey},
		flagtest.Static{moduleAuth: true, moduleTaxpayersKey: true, moduleIncomesKey: true, moduleExpensesKey: true})

	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers",
		createJSON("", "fisica", rfcAna, "Ana", `{"regime_code":"612","start_period":"2026-01"}`), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	base := "/v1/taxpayers/" + decodeTaxpayer(t, rec).Id.String()
	rec = env.do(env.send(http.MethodPost, base+"/incomes", incomeJSON(t, smokeIncomeID, nil), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	rec = env.do(env.send(http.MethodPost, base+"/expenses", expenseJSON(t, smokeExpenseID, nil), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	modules := map[string][]recordProbe{moduleIncomesKey: incomeProbes(), moduleExpensesKey: expenseProbes()}
	for module, probes := range modules {
		require.ElementsMatch(t, builtOwnedOperations(module), recordProbeOps(probes),
			"every built %s operation is probed", module)
	}

	presets, err := env.store.Presets(t.Context())
	require.NoError(t, err)
	var smoked []string
	for _, preset := range presets {
		smoked = append(smoked, preset.Name)
		t.Run(preset.Name, func(t *testing.T) {
			_, err := env.store.ApplyPreset(t.Context(), preset.Name)
			require.NoError(t, err)
			env.flags.Invalidate() // another process would wait for flags.DefaultTTL
			enabled := env.enabledFlags(t)

			for module, probes := range modules {
				reachable := slices.Contains(preset.Flags, module) && slices.Contains(preset.Flags, moduleTaxpayersKey)
				assert.Equal(t, reachable, slices.Contains(enabled, module), "GET /v1/flags lists %s", module)
				for _, p := range probes {
					rec := env.do(env.send(p.method, base+p.suffix, p.body, token))
					if reachable {
						assert.NotEqual(t, http.StatusNotFound, rec.Code, "%s: the module answers: %s", p.op, rec.Body.String())
						continue
					}
					assert.True(t, flagtest.CheckHidden(t, rec), "%s: hidden", p.op)
				}
			}
		})
	}
	assert.Subset(t, smoked, []string{
		"demo_core", "demo_monetization", "demo_full",
		noRecordsPreset, incomesOnlyPreset, expensesOnlyPreset, recordsWithoutTaxpayersPreset,
	}, "every seeded preset and fixture is smoked")
}
