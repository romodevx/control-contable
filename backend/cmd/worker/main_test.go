package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
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
		"APP_ENV":      "staging",
		"DATABASE_URL": "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1",
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

func TestNewRegistry_hasNoProductionJobsYet(t *testing.T) {
	reg, err := newRegistry()

	require.NoError(t, err)
	assert.Empty(t, reg.Names(), "M9 registers the reminder and purge jobs")
}
