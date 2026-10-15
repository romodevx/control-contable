package flags_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
)

// MODULE-tax-engine §7: module_tax_engine is implied OFF when
// module_taxpayers is OFF. The dependency is declared once, in flags.
func TestPrerequisites_declaresTheTaxEngineOnTaxpayers(t *testing.T) {
	assert.Equal(t, []string{"module_taxpayers"}, flags.Prerequisites()["module_tax_engine"])
}

// MODULE-incomes §6 and MODULE-expenses §6 (M3-10, M4-11): both record
// modules live under /v1/taxpayers/{taxpayerId}, so each is implied OFF when
// module_taxpayers is OFF. Neither requires the other.
func TestPrerequisites_declaresTheRecordModulesOnTaxpayers(t *testing.T) {
	graph := flags.Prerequisites()
	for _, module := range []string{"module_incomes", "module_expenses"} {
		assert.Equal(t, []string{"module_taxpayers"}, graph[module], module)
	}
}

// MODULE-dashboard §3: module_dashboard is implied OFF when module_taxpayers
// is OFF, and only then. The engine is NOT a prerequisite (OQ-D7): with it
// OFF Home degrades instead of disappearing.
func TestPrerequisites_declaresTheDashboardOnTaxpayersOnly(t *testing.T) {
	assert.Equal(t, []string{"module_taxpayers"}, flags.Prerequisites()["module_dashboard"])
}

// MODULE-history §3 (closes PLAN D7): History without estimates has
// nothing to show, so module_history is implied OFF when module_tax_engine
// (and, through it, module_taxpayers) is OFF.
func TestPrerequisites_declaresHistoryOnTheTaxEngine(t *testing.T) {
	graph := flags.Prerequisites()
	assert.Equal(t, []string{"module_tax_engine"}, graph["module_history"])
	on := func(m string) bool { return m == "module_history" || m == "module_tax_engine" }
	assert.Equal(t, []string{"module_taxpayers"}, flags.MissingPrerequisites(graph, "module_history", on), "transitively")
}

// MODULE-annual §6, ADR-0009 rule 2: the annual estimate is implied OFF
// when module_tax_engine or module_taxpayers is OFF.
func TestPrerequisites_declaresTheAnnualOnTheEngineAndTaxpayers(t *testing.T) {
	graph := flags.Prerequisites()
	assert.Equal(t, []string{"module_tax_engine", "module_taxpayers"}, graph["module_annual"])
	on := func(m string) bool { return m == "module_annual" || m == "module_taxpayers" }
	assert.Equal(t, []string{"module_tax_engine"}, flags.MissingPrerequisites(graph, "module_annual", on))
}

func TestPrerequisites_isAWellFormedAcyclicGraph(t *testing.T) {
	require.NoError(t, flags.CheckPrerequisites(flags.Prerequisites()))
}

func TestPrerequisites_returnsACopy(t *testing.T) {
	p := flags.Prerequisites()
	p["module_tax_engine"][0] = "module_auth"
	delete(p, "module_tax_engine")

	assert.Equal(t, []string{"module_taxpayers"}, flags.Prerequisites()["module_tax_engine"])
}

func TestCheckPrerequisites_refusesMalformedGraphs(t *testing.T) {
	for name, graph := range map[string]map[string][]string{
		"invalid module key":       {"tax_engine": {"module_taxpayers"}},
		"invalid prerequisite key": {"module_tax_engine": {"taxpayers"}},
		"self prerequisite":        {"module_a": {"module_a"}},
		"cycle":                    {"module_a": {"module_b"}, "module_b": {"module_c"}, "module_c": {"module_a"}},
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, flags.CheckPrerequisites(graph), flags.ErrInvalidPrerequisites)
		})
	}
}

func TestMissingPrerequisites(t *testing.T) {
	graph := map[string][]string{"module_c": {"module_b"}, "module_b": {"module_a"}}
	for _, tc := range []struct {
		name string
		on   []string
		want []string
	}{
		{"all ON", []string{"module_a", "module_b", "module_c"}, nil},
		{"direct prerequisite OFF", []string{"module_a", "module_c"}, []string{"module_b"}},
		{"transitive prerequisite OFF", []string{"module_b", "module_c"}, []string{"module_a"}},
		{"both OFF, sorted", []string{"module_c"}, []string{"module_a", "module_b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, flags.MissingPrerequisites(graph, "module_c", isOn(tc.on)))
		})
	}
}

func TestEffectiveModules(t *testing.T) {
	graph := map[string][]string{"module_tax_engine": {"module_taxpayers"}}
	for _, tc := range []struct {
		name string
		on   []string
		want []string
	}{
		{"both ON", []string{"module_tax_engine", "module_taxpayers"}, []string{"module_tax_engine", "module_taxpayers"}},
		{"prerequisite OFF hides the dependent", []string{"module_auth", "module_tax_engine"}, []string{"module_auth"}},
		{"dependent OFF stays OFF", []string{"module_taxpayers"}, []string{"module_taxpayers"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.want, flags.EffectiveModules(graph, tc.on))
		})
	}
}

// The service gates and lists the effective set: GET /v1/flags never lists
// a module whose routes answer 404.
func TestService_appliesThePrerequisites(t *testing.T) {
	svc, _ := newService(t, &fakeReader{keys: []string{"module_auth", "module_tax_engine"}})

	assert.False(t, svc.Enabled(t.Context(), "module_tax_engine"), "module_taxpayers is OFF")
	got, err := svc.EnabledModules(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"module_auth"}, got)
}

func isOn(keys []string) func(string) bool {
	on := make(map[string]bool, len(keys))
	for _, k := range keys {
		on[k] = true
	}
	return func(k string) bool { return on[k] }
}
