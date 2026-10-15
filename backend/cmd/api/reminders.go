package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	remindersusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// remindersHandler serves the module_reminders operations: getTaxCalendar
// (M9-04), registerDevice, unregisterDevice, getReminderSettings and
// putReminderSettings (M9-05).
type remindersHandler = remindershttp.Handler

// remindersWiring is what the reminders module needs from the composition
// root: the holidays (a boot snapshot over the reminders' own reader,
// loadHolidays), the push token keys, the cc_app unit of work (the record
// modules' shared manager), the per-user budgets and the logger the
// calendar alerts go to.
type remindersWiring struct {
	env      config.Environment
	keys     config.Reminders
	tx       tx.Manager // cc_app
	holidays remindersusecase.HolidayCalendar
	limits   *ratelimitpg.Store
	logger   *slog.Logger
}

// remindersModule is the reminders module's HTTP side and its budgets.
type remindersModule struct {
	handler    *remindersHandler
	userLimits map[string]router.Limiter
}

// newRemindersModule wires the tax calendar over the holidays snapshot and
// the devices and settings use cases over the cc_app unit of work.
func newRemindersModule(w remindersWiring) (remindersModule, error) {
	if w.limits == nil || w.logger == nil || w.tx == nil {
		return remindersModule{}, fmt.Errorf("%w: reminders wiring", remindersusecase.ErrMissingDependency)
	}
	calendar, err := remindersusecase.NewCalendar(remindersusecase.CalendarDeps{
		Holidays: w.holidays, Alerts: calendarAlerts{logger: w.logger},
	})
	if err != nil {
		return remindersModule{}, err
	}
	vault, err := newPushTokenVault(w)
	if err != nil {
		return remindersModule{}, err
	}
	devices, err := remindersusecase.NewDevices(remindersusecase.DevicesDeps{Tx: w.tx, Store: store.NewDevices(), Sealer: vault})
	if err != nil {
		return remindersModule{}, err
	}
	settings, err := remindersusecase.NewSettings(remindersusecase.SettingsDeps{Tx: w.tx, Store: store.NewSettings()})
	if err != nil {
		return remindersModule{}, err
	}
	return remindersModule{
		handler:    remindershttp.NewHandler(remindershttp.UseCases{Calendar: calendar, Devices: devices, Settings: settings}),
		userLimits: bucketsOf(w.limits, remindershttp.UserLimits),
	}, nil
}

// newPushTokenVault builds the push token vault (MODULE-reminders §2.3,
// OQ-R7) from the configured keys. In dev a missing key becomes an
// ephemeral one, logged as such: tokens sealed with it cannot be opened,
// nor claimed by their HMAC, after a restart (the app's next
// registerDevice seals them again). Outside dev config requires both keys.
func newPushTokenVault(w remindersWiring) (*pushtoken.Vault, error) {
	key, err := keyOrEphemeral(w.env, w.logger, w.keys.PushTokenKey, "PUSH_TOKEN_ENC_KEY", crypto.KeySize)
	if err != nil {
		return nil, err
	}
	pepper, err := keyOrEphemeral(w.env, w.logger, w.keys.PushTokenIndexPepper, "PUSH_TOKEN_INDEX_PEPPER", crypto.MinPepperSize)
	if err != nil {
		return nil, err
	}
	keyID := crypto.KeyID(w.keys.PushTokenKeyID) // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	keyring, err := crypto.NewKeyring(keyID, map[crypto.KeyID][]byte{keyID: key})
	if err != nil {
		return nil, fmt.Errorf("reminders: push token keyring: %w", err)
	}
	index, err := crypto.NewTokenIndex(pepper)
	if err != nil {
		return nil, fmt.Errorf("reminders: push token index: %w", err)
	}
	return pushtoken.NewVault(keyring, index), nil
}

// unseededEvent is the log event the reminders_calendar_unseeded alert
// matches (MODULE-reminders §9): a due date computed weekends-only because
// the year's holidays are not seeded.
const unseededEvent = "reminders_calendar_unseeded"

// calendarAlerts reports the calendar's operational signals as structured
// warnings (no user data: the year only).
type calendarAlerts struct {
	logger *slog.Logger
}

var _ remindersusecase.Alerts = calendarAlerts{}

// CalendarUnseeded implements remindersusecase.Alerts.
func (a calendarAlerts) CalendarUnseeded(ctx context.Context, year int) {
	a.logger.WarnContext(ctx, "holidays not seeded: due dates computed weekends-only",
		slog.String("event", unseededEvent), slog.Int("year", year))
}
