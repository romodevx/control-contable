package main

import (
	"bytes"
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

// unreachable is a cc_worker DSN nothing listens on.
const unreachable = "postgres://cc_worker@127.0.0.1:1/control_contable?sslmode=disable&connect_timeout=1"

func TestRun_refusesBadArguments(t *testing.T) {
	var logs bytes.Buffer
	require.Error(t, run(t.Context(), []string{"-force"}, lookupFrom(nil), &bytes.Buffer{}, &logs))
	assert.Contains(t, logs.String(), "-apply")

	err := run(t.Context(), []string{"now"}, lookupFrom(nil), &bytes.Buffer{}, &bytes.Buffer{})
	require.ErrorContains(t, err, "unexpected arguments")
}

func TestRun_refusesAnInvalidConfiguration(t *testing.T) {
	err := run(t.Context(), nil, lookupFrom(map[string]string{}), &bytes.Buffer{}, &bytes.Buffer{})

	require.ErrorContains(t, err, "APP_ENV is required")
}

func TestRun_refusesAnUnparsableDatabaseURL(t *testing.T) {
	env := map[string]string{"APP_ENV": "dev", "DATABASE_URL": "postgres://%zz"}

	err := run(t.Context(), nil, lookupFrom(env), &bytes.Buffer{}, &bytes.Buffer{})

	require.ErrorIs(t, err, pgdb.ErrInvalidDSN)
	assert.NotContains(t, err.Error(), "%zz")
}

// The role is checked even in dev: the run writes data, so a database it
// cannot verify stops it before anything is read.
func TestRun_refusesAnUnverifiedRoleInEveryEnvironment(t *testing.T) {
	for _, env := range []string{"dev", "staging"} {
		var out bytes.Buffer
		err := run(t.Context(), []string{"-apply"}, lookupFrom(map[string]string{"APP_ENV": env, "DATABASE_URL": unreachable}),
			&out, &bytes.Buffer{})

		require.ErrorContains(t, err, "refusing to run", env)
		assert.Empty(t, out.String())
	}
}
