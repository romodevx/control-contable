package remindershttp

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// The module's per-user budgets (MODULE-reminders §2.7, ADR-0005).
var (
	// taxCalendarReadBudget: getTaxCalendar, 120 per hour (one every 30 s),
	// a burst of 60.
	taxCalendarReadBudget = ratelimit.Policy{
		Scope: "tax_calendar_reads", Burst: taxCalendarBurst, Interval: time.Hour / taxCalendarReadsPerHour,
	}
	// devicesWriteBudget: registerDevice and unregisterDevice share one
	// bucket per user, 30 per hour (one every 2 minutes), a burst of 10.
	devicesWriteBudget = ratelimit.Policy{
		Scope: "devices_writes", Burst: devicesWriteBurst, Interval: time.Hour / devicesWritesPerHour,
	}
	// reminderSettingsWriteBudget: putReminderSettings, 60 per hour, a burst
	// of 20. getReminderSettings has no budget of its own (the edge limiter
	// bounds it).
	reminderSettingsWriteBudget = ratelimit.Policy{
		Scope: "reminder_settings_writes", Burst: reminderSettingsWriteBurst, Interval: time.Hour / reminderSettingsWritesPerHour,
	}
)

const (
	taxCalendarReadsPerHour       = 120
	taxCalendarBurst              = 60
	devicesWritesPerHour          = 30
	devicesWriteBurst             = 10
	reminderSettingsWritesPerHour = 60
	reminderSettingsWriteBurst    = 20
)

// UserLimits are the module's per-user limits by operationId, keyed by the
// authenticated user. cmd/api hands them to the router, which applies them
// right after authentication; operations with the same policy share its
// bucket.
var UserLimits = map[string]ratelimit.Policy{
	"getTaxCalendar":      taxCalendarReadBudget,
	"registerDevice":      devicesWriteBudget,
	"unregisterDevice":    devicesWriteBudget,
	"putReminderSettings": reminderSettingsWriteBudget,
}
