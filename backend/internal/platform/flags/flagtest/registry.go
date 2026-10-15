package flagtest

import "slices"

// moduleKeys is the MODULES.md flag registry the migrations seed, sorted by
// key: M1–M9 plus module_annual (MODULE-annual §6, ANN-22). It is the one
// list the flag suites compare the database against; a new module's flag
// migration adds its key here and every count follows from len.
var moduleKeys = []string{
	"module_annual", "module_auth", "module_dashboard", "module_expenses", "module_history", "module_incomes",
	"module_reminders", "module_subscriptions", "module_tax_engine", "module_taxpayers",
}

// demoCoreKeys is the MODULES.md demo_core preset, sorted by key.
var demoCoreKeys = []string{
	"module_auth", "module_dashboard", "module_expenses", "module_incomes", "module_tax_engine", "module_taxpayers",
}

// demoMonetizationKeys is demo_core + module_subscriptions, sorted by key.
var demoMonetizationKeys = []string{
	"module_auth", "module_dashboard", "module_expenses", "module_incomes", "module_subscriptions",
	"module_tax_engine", "module_taxpayers",
}

// ModuleKeys returns a copy of every registered module flag key, sorted.
// demo_full holds all of them.
func ModuleKeys() []string { return slices.Clone(moduleKeys) }

// DemoCoreKeys returns a copy of the demo_core preset's keys, sorted.
func DemoCoreKeys() []string { return slices.Clone(demoCoreKeys) }

// DemoMonetizationKeys returns a copy of the demo_monetization preset's keys, sorted.
func DemoMonetizationKeys() []string { return slices.Clone(demoMonetizationKeys) }
