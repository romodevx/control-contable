package flagtest_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

func TestModuleKeys_isASortedSetOfValidKeys(t *testing.T) {
	keys := flagtest.ModuleKeys()

	assert.NotEmpty(t, keys)
	assert.True(t, slices.IsSorted(keys), "ordered by key, as the stores list them")
	assert.Len(t, slices.Compact(slices.Clone(keys)), len(keys), "no duplicates")
	for _, k := range keys {
		assert.True(t, flags.ValidKey(k), k)
	}
}

// Every module the prerequisite graph names is a registered flag: a module
// declared in code but missing from the registry fails here, not in a count.
func TestModuleKeys_holdsEveryModuleOfThePrerequisiteGraph(t *testing.T) {
	keys := flagtest.ModuleKeys()
	for module, prereqs := range flags.Prerequisites() {
		assert.Contains(t, keys, module)
		assert.Subset(t, keys, prereqs, module)
	}
}

func TestPresetKeys_areSortedProperSubsetsOfTheRegistry(t *testing.T) {
	keys := flagtest.ModuleKeys()
	for name, preset := range map[string][]string{
		"demo_core":         flagtest.DemoCoreKeys(),
		"demo_monetization": flagtest.DemoMonetizationKeys(),
	} {
		assert.True(t, slices.IsSorted(preset), name)
		assert.Subset(t, keys, preset, name)
		assert.Less(t, len(preset), len(keys), "%s is not demo_full", name)
	}
	assert.Subset(t, flagtest.DemoMonetizationKeys(), flagtest.DemoCoreKeys(),
		"demo_monetization is demo_core + subscriptions")
}

func TestRegistry_returnsCopies(t *testing.T) {
	keys := flagtest.ModuleKeys()
	keys[0] = "module_mutated"
	core := flagtest.DemoCoreKeys()
	core[0] = "module_mutated"

	assert.NotContains(t, flagtest.ModuleKeys(), "module_mutated")
	assert.NotContains(t, flagtest.DemoCoreKeys(), "module_mutated")
}
