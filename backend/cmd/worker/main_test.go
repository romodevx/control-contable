package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authusecase "github.com/romodevx/control-contable/backend/internal/auth/usecase"
	keyrotation "github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	reminders "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

func lookupFrom(env map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestRun_refusesAnInvalidConfiguration(t *testing.T) {
	err := run(t.Context(), lookupFrom(map[string]string{}), &bytes.Buffer{})

	require.ErrorContains(t, err, "APP_ENV is required")
}

func TestRun_refusesAnUnparsableDatabaseURL(t *testing.T) {
	env := map[string]string{"APP_ENV": "dev", "DATABASE_URL": "postgres://%zz", "AUTH_OUTBOX_KEY": randomHex(t)}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorIs(t, err, pgdb.ErrInvalidDSN)
	assert.NotContains(t, err.Error(), "%zz")
}

// Outside dev the role check needs the database: one that is unreachable at
// boot keeps the worker from starting (fail closed).
func TestRun_outsideDevRefusesToStartWithAnUnverifiedRole(t *testing.T) {
	env := map[string]string{
		"APP_ENV":                 "staging",
		"DATABASE_URL":            "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
		"PUSH_TOKEN_ENC_KEY":      randomHex(t),
		"PUSH_TOKEN_INDEX_PEPPER": randomHex(t),
		"AUTH_OUTBOX_KEY":         randomHex(t),
	}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorContains(t, err, "refusing to start")
}

// In dev the worker starts before its database; with nothing to lead it
// campaigns, logs the failure and stops cleanly on shutdown.
func TestRun_inDevStartsAndStopsWithoutADatabase(t *testing.T) {
	env := map[string]string{
		"APP_ENV":         "dev",
		"DATABASE_URL":    "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
		"AUTH_OUTBOX_KEY": randomHex(t),
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var logs bytes.Buffer

	err := run(ctx, lookupFrom(env), &logs)

	require.NoError(t, err)
	assert.Contains(t, logs.String(), `"msg":"worker started"`)
	assert.Contains(t, logs.String(), `"msg":"worker stopped"`)
	assert.Contains(t, logs.String(), `"msg":"leader election failed"`)
}

// testWiring is newJobs' wiring over an unreachable pool (building the jobs
// needs no database), with the fixture outbox key and the given reminder keys.
func testWiring(t *testing.T, keys config.Reminders) jobWiring {
	t.Helper()
	pool, err := pgdb.Open("postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	return jobWiring{
		pool: pool, logger: logging.New(&bytes.Buffer{}, nil), push: reminders.DisabledPushSender{},
		reminderKeys: keys, outbox: testOutboxKeys(t), env: config.EnvDev,
		now: clock.NewFixed(time.Date(2026, time.May, 14, 10, 5, 0, 0, clock.Location())),
	}
}

// The records purge (M9-08) and the auth email delivery are registered
// unconditionally: retention does not depend on any module flag (PLAN
// decision 5), and an auth email queued before module_auth was turned off is
// still delivered.
func TestNewRegistry_registersThePurgeAndTheAuthEmailDelivery(t *testing.T) {
	reg, err := newRegistry(testWiring(t, config.Reminders{}))

	require.NoError(t, err)
	assert.Equal(t, []string{"records_purge", "auth_email_delivery", "key_reseal"}, reg.Names(), "dev without push token keys: no reminder job")
}

// With the push token keys (required outside dev) the reminder job (M9-06)
// is registered too; it checks module_reminders on every run.
func TestNewRegistry_registersTheReminders(t *testing.T) {
	reg, err := newRegistry(testWiring(t, testReminderKeys(t)))

	require.NoError(t, err)
	assert.Equal(t, []string{"records_purge", "auth_email_delivery", "reminders", "key_reseal"}, reg.Names())
}

// newJobs pairs each production job with its own per-run timeout (M9-11
// refactor: the pairing is data now, so a swap must not go unnoticed).
func TestNewJobs_pairsEachJobWithItsTimeout(t *testing.T) {
	jobs, err := newJobs(testWiring(t, testReminderKeys(t)))

	require.NoError(t, err)
	got := map[string]time.Duration{}
	for _, j := range jobs {
		got[j.job.Name()] = j.timeout
	}
	assert.Equal(t, map[string]time.Duration{
		"records_purge":       recordsPurgeTimeout,
		"auth_email_delivery": authusecase.DeliveryRunTimeout,
		"reminders":           remindersTimeout,
		"key_reseal":          keyrotation.RunTimeout,
	}, got)
}

// The worker opens the auth emails the API sealed: without AUTH_OUTBOX_KEY it
// refuses to start, in every environment (config.LoadAuthOutbox).
func TestRun_requiresTheOutboxKey(t *testing.T) {
	env := map[string]string{
		"APP_ENV":      "dev",
		"DATABASE_URL": "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
	}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorContains(t, err, "AUTH_OUTBOX_KEY")
}

// Outside dev the worker needs the push token keys to open the tokens the
// API sealed: missing ones refuse the start (config.LoadReminders).
func TestRun_outsideDevRequiresThePushTokenKeys(t *testing.T) {
	env := map[string]string{
		"APP_ENV":      "staging",
		"DATABASE_URL": "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
	}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorContains(t, err, "PUSH_TOKEN_ENC_KEY")
}
