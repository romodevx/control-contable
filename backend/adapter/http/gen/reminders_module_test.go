package gen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

const moduleReminders = "module_reminders"

// m9Routes are the operations of MODULE-reminders, written by hand: the
// generated map must agree with the spec, not merely with itself.
var m9Routes = map[string]string{
	"registerDevice":      "POST /v1/devices",
	"unregisterDevice":    "DELETE /v1/devices/{deviceId}",
	"getReminderSettings": "GET /v1/reminder-settings",
	"putReminderSettings": "PUT /v1/reminder-settings",
	"getTaxCalendar":      "GET /v1/tax-calendar",
}

// TestFlagMap_everyM9OperationIsGatedByModuleReminders (M9-01): each M9
// operation, by operationId and by the ServeMux pattern the gate reads, is
// owned by module_reminders, and module_reminders owns nothing else.
func TestFlagMap_everyM9OperationIsGatedByModuleReminders(t *testing.T) {
	for operationID, pattern := range m9Routes {
		t.Run(operationID, func(t *testing.T) {
			module, ok := gen.OperationModule(operationID)
			assert.True(t, ok, "operationId is in the contract")
			assert.Equal(t, moduleReminders, module, "operationId -> flag")

			module, ok = gen.RouteModule(pattern)
			assert.True(t, ok, "route %q is registered", pattern)
			assert.Equal(t, moduleReminders, module, "route -> flag")
		})
	}

	var owned []string
	for operationID, module := range gen.OperationModules() {
		if module == moduleReminders {
			owned = append(owned, operationID)
		}
	}
	assert.ElementsMatch(t, []string{
		"registerDevice", "unregisterDevice", "getReminderSettings", "putReminderSettings", "getTaxCalendar",
	}, owned, "module_reminders owns exactly the M9 operations")
}
