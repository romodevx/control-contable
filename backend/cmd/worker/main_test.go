package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	env := map[string]string{"APP_ENV": "dev", "DATABASE_URL": "postgres://%zz"}

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
	}

	err := run(t.Context(), lookupFrom(env), &bytes.Buffer{})

	require.ErrorContains(t, err, "refusing to start")
}

// In dev the worker starts before its database; with nothing to lead it
// campaigns, logs the failure and stops cleanly on shutdown.
func TestRun_inDevStartsAndStopsWithoutADatabase(t *testing.T) {
	env := map[string]string{
		"APP_ENV":      "dev",
		"DATABASE_URL": "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
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

// The records purge (M9-08) is registered unconditionally: retention does not
// depend on any module flag (PLAN decision 5).
func TestNewRegistry_registersTheRecordsPurge(t *testing.T) {
	pool, err := pgdb.Open("postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	reg, err := newRegistry(pool, logging.New(&bytes.Buffer{}, nil), reminders.DisabledPushSender{}, config.Reminders{})

	require.NoError(t, err)
	assert.Equal(t, []string{"records_purge"}, reg.Names(), "dev without push token keys: no reminder job")
}

// With the push token keys (required outside dev) the reminder job (M9-06)
// is registered too; it checks module_reminders on every run.
func TestNewRegistry_registersTheReminders(t *testing.T) {
	pool, err := pgdb.Open("postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	reg, err := newRegistry(pool, logging.New(&bytes.Buffer{}, nil), reminders.DisabledPushSender{}, testReminderKeys(t))

	require.NoError(t, err)
	assert.Equal(t, []string{"records_purge", "reminders"}, reg.Names())
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
