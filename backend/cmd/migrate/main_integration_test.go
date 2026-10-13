//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// The template keeps only the role bootstrap: these tests drive the runner
// against an empty database, as on a fresh environment.
func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m))
}

// Tables of the first (baseline) and the latest (auth service) migration.
const (
	baselineTable = "public.feature_flags"
	latestTable   = "public.trial_registry"
)

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var found sql.NullString
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT to_regclass($1)::text", name).Scan(&found))
	return found.Valid
}

func flagsTableExists(t *testing.T, db *sql.DB) bool {
	t.Helper()
	return tableExists(t, db, baselineTable)
}

func TestExecute_UpStatusDownUp(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	var out bytes.Buffer

	require.NoError(t, execute(t.Context(), cmdUp, owner, &out))
	assert.Contains(t, out.String(), "OK")
	assert.True(t, flagsTableExists(t, owner))

	out.Reset()
	require.NoError(t, execute(t.Context(), cmdUp, owner, &out), "second up")
	assert.Contains(t, out.String(), "no migrations to run")

	out.Reset()
	require.NoError(t, execute(t.Context(), cmdStatus, owner, &out))
	assert.Contains(t, out.String(), "applied")
	assert.NotContains(t, out.String(), "pending")

	out.Reset()
	require.NoError(t, execute(t.Context(), cmdDown, owner, &out))
	assert.Contains(t, out.String(), "down")
	assert.False(t, tableExists(t, owner, latestTable), "down rolled back the latest migration")
	assert.True(t, flagsTableExists(t, owner), "and only that one")

	out.Reset()
	require.NoError(t, execute(t.Context(), cmdStatus, owner, &out))
	assert.Contains(t, out.String(), "pending")

	require.NoError(t, execute(t.Context(), cmdUp, owner, &bytes.Buffer{}), "up after down")
	assert.True(t, tableExists(t, owner, latestTable))
}

func TestExecute_RefusesARuntimeRole(t *testing.T) {
	db := pgtest.NewDatabase(t)

	err := execute(t.Context(), cmdUp, db.Connect(t, pgtest.RoleApp), &bytes.Buffer{})

	require.ErrorIs(t, err, migrations.ErrNotOwner)
	assert.False(t, flagsTableExists(t, db.Connect(t, pgtest.RoleOwner)), "nothing applied")
}
