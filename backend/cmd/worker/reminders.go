package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/cmd/internal/estimates"
	"github.com/romodevx/control-contable/backend/cmd/internal/sealing"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/holidays"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/jobstore"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	taxdomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxholidays "github.com/romodevx/control-contable/backend/internal/taxengine/infra/holidays"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// remindersTimeout bounds one run of the reminder job: below its 5-minute
// interval (runs never overlap) and far above one batch's sends (each send
// has its own 30 s deadline in the FCM adapter).
const remindersTimeout = 4 * time.Minute

// collapseKeyLabel separates the collapse-id key from the token index that
// shares its pepper (one secret, two derived purposes).
const collapseKeyLabel = "control-contable reminders collapse id v1"

// hasReminderKeys reports whether the push token keys are set: always
// outside dev (config.LoadReminders requires them), optional in dev.
func hasReminderKeys(keys config.Reminders) bool {
	return !keys.PushTokenKey.IsZero() && !keys.PushTokenIndexPepper.IsZero()
}

// newReminderJob wires the reminder job (M9-06) over pool (cc_worker): the
// job store, the token vault with the API's keys, the push sender, the
// module flags, the holidays and the monthly Estimator for the amounts; now
// is the job's and the estimate's "today".
func newReminderJob(pool *sql.DB, units *pgdb.TxManager, logger *slog.Logger, push reminders.PushSender, keys config.Reminders, now clock.Clock) (*reminders.ReminderJob, error) {
	vault, err := newTokenVault(keys)
	if err != nil {
		return nil, err
	}
	g, err := flagsgorm.Open(pool)
	if err != nil {
		return nil, fmt.Errorf("reminders: flags: %w", err)
	}
	tpStore := tpinfra.NewStore()
	estimator, err := estimates.NewMonthly(estimates.MonthlyDeps{
		Tx:        units,
		Taxpayers: taxpayers.NewQueries(taxpayers.Deps{Tx: units, Taxpayers: tpStore, Regimes: tpStore}),
		Holidays:  taxholidays.NewReader(),
		Clock:     now,
	})
	if err != nil {
		return nil, err
	}
	jobs := jobstore.New()
	return reminders.NewReminderJob(reminders.JobDeps{
		System: units, Units: units, Due: jobs, Log: jobs, Settings: store.NewSettings(), Recipients: jobs,
		Opener: vault, Push: push, Estimates: workerEstimates{estimator: estimator},
		Flags:    flags.NewService(flagsgorm.NewReader(g), flags.WithLogger(logger)),
		Holidays: holidays.NewReader(), Alerts: calendarAlerts{logger: logger}, Clock: now, Logger: logger,
		CollapseKey: collapseKey(keys.PushTokenIndexPepper.RevealBytes()),
		Concurrency: reminders.DefaultJobConcurrency, BatchSize: reminders.DefaultJobBatchSize,
	})
}

// newTokenVault opens the tokens the API sealed (same keys, M9-05 N1), with
// the previous key too while a rotation is in flight (ADR-0012).
func newTokenVault(keys config.Reminders) (*pushtoken.Vault, error) {
	_, vault, err := newTokenKeyringAndVault(keys)
	return vault, err
}

// newTokenKeyringAndVault is newTokenVault with its keyring, which the key
// re-seal job needs for the active key id.
func newTokenKeyringAndVault(keys config.Reminders) (*crypto.Keyring, *pushtoken.Vault, error) {
	keyring, err := sealing.Keyring(keys.PushTokenKeyID, keys.PushTokenKey.RevealBytes(), keys.PushTokenPrevious)
	if err != nil {
		return nil, nil, fmt.Errorf("reminders: push token keyring: %w", err)
	}
	index, err := crypto.NewTokenIndex(keys.PushTokenIndexPepper.RevealBytes())
	if err != nil {
		return nil, nil, fmt.Errorf("reminders: push token index: %w", err)
	}
	return keyring, pushtoken.NewVault(keyring, index), nil
}

// collapseKey derives the collapse-id HMAC key from pepper.
func collapseKey(pepper []byte) []byte {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(collapseKeyLabel))
	return mac.Sum(nil)
}

// workerEstimates adapts the tax engine's monthly Estimator to the job's
// EstimateReader. The Estimator opens its own read-only unit of work for
// the user (cc_worker reads that user's records under RLS).
type workerEstimates struct {
	estimator *taxusecase.Estimator
}

var _ reminders.EstimateReader = workerEstimates{}

// MonthlyEstimate implements reminders.EstimateReader.
func (w workerEstimates) MonthlyEstimate(ctx context.Context, userID, taxpayerID string, p domain.Period) (reminders.EstimateSummary, error) {
	e, err := w.estimator.EstimateMonth(ctx, userID, taxpayerID, p.String())
	if err != nil {
		return reminders.EstimateSummary{}, err
	}
	return estimateSummary(e), nil
}

// estimateSummary is the taxpayer-level status the dashboard derives
// (dashboard/domain.TaxpayerStatus) and the rounded total.
func estimateSummary(e taxusecase.TaxpayerEstimate) reminders.EstimateSummary {
	statuses := make([]taxdomain.Status, len(e.Regimes))
	for i, r := range e.Regimes {
		statuses[i] = r.Status
	}
	out := reminders.EstimateSummary{TotalPayableCents: int64(e.Totals.TotalPayable)}
	switch {
	case slices.Contains(statuses, taxdomain.StatusIncomplete):
		out.Status = reminders.EstimateIncomplete
	case slices.Contains(statuses, taxdomain.StatusComplete) && !slices.Contains(statuses, taxdomain.StatusPartial):
		out.Status = reminders.EstimateComplete
	default:
		out.Status = reminders.EstimateOther
	}
	return out
}

// unseededEvent is the log event the reminders_calendar_unseeded alert
// matches (MODULE-reminders §9), as the API logs it.
const unseededEvent = "reminders_calendar_unseeded"

// calendarAlerts reports a due date computed weekends-only (year only).
type calendarAlerts struct {
	logger *slog.Logger
}

var _ reminders.Alerts = calendarAlerts{}

// CalendarUnseeded implements reminders.Alerts.
func (a calendarAlerts) CalendarUnseeded(ctx context.Context, year int) {
	a.logger.WarnContext(ctx, "holidays not seeded: due dates computed weekends-only",
		slog.String("event", unseededEvent), slog.Int("year", year))
}
