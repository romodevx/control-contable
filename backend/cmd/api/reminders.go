package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	remindersholidays "github.com/romodevx/control-contable/backend/internal/reminders/infra/holidays"
	remindersusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// remindersHandler serves the built module_reminders operations
// (getTaxCalendar, M9-04; devices and settings follow in M9-05).
type remindersHandler = remindershttp.Handler

// remindersWiring is what the reminders module needs from the composition
// root: a unit of work for the holidays read (global tables, any user's
// transaction reads them), the per-user budgets and the logger the
// calendar alerts go to.
type remindersWiring struct {
	tx     tx.Manager
	limits *ratelimitpg.Store
	logger *slog.Logger
}

// remindersModule is the reminders module's HTTP side and its budgets.
type remindersModule struct {
	handler    *remindersHandler
	userLimits map[string]router.Limiter
}

// newRemindersModule wires the tax calendar over the holidays tables.
func newRemindersModule(w remindersWiring) (remindersModule, error) {
	if w.limits == nil || w.logger == nil {
		return remindersModule{}, fmt.Errorf("%w: reminders wiring", remindersusecase.ErrMissingDependency)
	}
	calendar, err := remindersusecase.NewCalendar(remindersusecase.CalendarDeps{
		Tx: w.tx, Holidays: remindersholidays.NewReader(), Alerts: calendarAlerts{logger: w.logger},
	})
	if err != nil {
		return remindersModule{}, err
	}
	limits := make(map[string]router.Limiter, len(remindershttp.UserLimits))
	for op, policy := range remindershttp.UserLimits {
		limits[op] = w.limits.Bucket(policy)
	}
	return remindersModule{handler: remindershttp.NewHandler(calendar), userLimits: limits}, nil
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
