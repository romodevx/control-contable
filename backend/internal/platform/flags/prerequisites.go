package flags

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// ErrInvalidPrerequisites means a prerequisite graph names a malformed key or has
// a cycle.
var ErrInvalidPrerequisites = errors.New("flags: invalid module prerequisites")

// prerequisites declares, per module, the modules whose data its routes
// read: a module is effectively ON only when its own flag and every
// prerequisite's (transitively) are ON. It is the single declaration the
// route gate, GET /v1/flags and cmd/flags all read
// (docs/architecture/platform.md#flags).
var prerequisites = map[string][]string{
	// MODULE-tax-engine §7: implied OFF when module_taxpayers is OFF.
	"module_tax_engine": {"module_taxpayers"},
	// MODULE-incomes §6, MODULE-expenses §6: records live under a taxpayer.
	"module_incomes":  {"module_taxpayers"},
	"module_expenses": {"module_taxpayers"},
	// MODULE-dashboard §3: the route lives under a taxpayer. The engine is
	// not a prerequisite (OQ-D7): with it OFF the dashboard shows no tax.
	"module_dashboard": {"module_taxpayers"},
	// MODULE-history §3 (PLAN D7): History without estimates has nothing to
	// show; through the engine it also needs module_taxpayers.
	"module_history": {"module_tax_engine"},
	// MODULE-annual §6, ADR-0009: the annual estimate is a sub-context of
	// the engine and lives under a taxpayer. The flag row itself is
	// registered by migration in ANN-22.
	"module_annual": {"module_tax_engine", "module_taxpayers"},
}

// Prerequisites returns a copy of the declared module prerequisites.
func Prerequisites() map[string][]string {
	out := make(map[string][]string, len(prerequisites))
	for module, needs := range prerequisites {
		out[module] = slices.Clone(needs)
	}
	return out
}

// CheckPrerequisites returns ErrInvalidPrerequisites when graph names a key
// outside the flag grammar or has a cycle (a module requiring itself
// included). O(V+E).
func CheckPrerequisites(graph map[string][]string) error {
	const (
		unvisited = iota
		inProgress
		done
	)
	state := make(map[string]int, len(graph))
	var visit func(module string) error
	visit = func(module string) error {
		switch state[module] {
		case inProgress:
			return fmt.Errorf("%w: cycle through %s", ErrInvalidPrerequisites, module)
		case done:
			return nil
		}
		if !ValidKey(module) {
			return fmt.Errorf("%w: key %q", ErrInvalidPrerequisites, module)
		}
		state[module] = inProgress
		for _, p := range graph[module] {
			if err := visit(p); err != nil {
				return err
			}
		}
		state[module] = done
		return nil
	}
	for _, module := range slices.Sorted(maps.Keys(graph)) {
		if err := visit(module); err != nil {
			return err
		}
	}
	return nil
}

// MissingPrerequisites returns the prerequisites of module, direct and
// transitive, that on reports OFF, sorted; nil when none is. O(V+E).
func MissingPrerequisites(graph map[string][]string, module string, on func(string) bool) []string {
	visited := map[string]bool{module: true}
	var missing []string
	stack := slices.Clone(graph[module])
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[p] {
			continue
		}
		visited[p] = true
		if !on(p) {
			missing = append(missing, p)
		}
		stack = append(stack, graph[p]...)
	}
	slices.Sort(missing)
	return missing
}

// EffectiveModules returns the keys of enabled whose prerequisites are all
// in enabled too, in their original order.
func EffectiveModules(graph map[string][]string, enabled []string) []string {
	on := make(map[string]bool, len(enabled))
	for _, k := range enabled {
		on[k] = true
	}
	isOn := func(k string) bool { return on[k] }
	out := make([]string, 0, len(enabled))
	for _, k := range enabled {
		if len(MissingPrerequisites(graph, k, isOn)) == 0 {
			out = append(out, k)
		}
	}
	return out
}
