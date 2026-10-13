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

// Table of the first (baseline) migration.
const baselineTable = "public.feature_flags"

// dbVersion returns the version goose has recorded as applied.
func dbVersion(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	p, err := migrations.NewProvider(db)
	require.NoError(t, err)
	v, err := p.GetDBVersion(t.Context())
	require.NoError(t, err)
	return v
}

// lastTwoVersions returns the versions of the two newest embedded migrations, so the
// down/up test never needs editing when a migration is added.
func lastTwoVersions(t *testing.T, db *sql.DB) (previous, latest int64) {
	t.Helper()
	p, err := migrations.NewProvider(db)
	require.NoError(t, err)
	sources := p.ListSources()
	require.GreaterOrEqual(t, len(sources), 2, "need two migrations to prove a single-step down")
	return sources[len(sources)-2].Version, sources[len(sources)-1].Version
}

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
	previous, latest := lastTwoVersions(t, owner)
	assert.Equal(t, previous, dbVersion(t, owner), "down rolled back the latest migration, and only that one")
	assert.True(t, flagsTableExists(t, owner))

	out.Reset()
	require.NoError(t, execute(t.Context(), cmdStatus, owner, &out))
	assert.Contains(t, out.String(), "pending")

	require.NoError(t, execute(t.Context(), cmdUp, owner, &bytes.Buffer{}), "up after down")
	assert.Equal(t, latest, dbVersion(t, owner))
}

func TestExecute_RefusesARuntimeRole(t *testing.T) {
	db := pgtest.NewDatabase(t)

	err := execute(t.Context(), cmdUp, db.Connect(t, pgtest.RoleApp), &bytes.Buffer{})

	require.ErrorIs(t, err, migrations.ErrNotOwner)
	assert.False(t, flagsTableExists(t, db.Connect(t, pgtest.RoleOwner)), "nothing applied")
}
