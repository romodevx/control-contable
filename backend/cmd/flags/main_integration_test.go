//go:build integration

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

func runAs(t *testing.T, db *pgtest.Database, role pgtest.Role, args ...string) (string, error) {
	t.Helper()
	cmd, err := parseCommand(args)
	require.NoError(t, err)
	var out bytes.Buffer
	err = execute(t.Context(), cmd, db.Connect(t, role), &out)
	return out.String(), err
}

func TestExecute_AsOwner(t *testing.T) {
	db := pgtest.NewDatabase(t)

	out, err := runAs(t, db, pgtest.RoleOwner, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "0 of 9 modules ON")

	out, err = runAs(t, db, pgtest.RoleOwner, "apply-preset", "demo_core")
	require.NoError(t, err)
	assert.Contains(t, out, "6 of 9 modules ON")

	out, err = runAs(t, db, pgtest.RoleOwner, "enable", "module_reminders")
	require.NoError(t, err)
	assert.Contains(t, out, "module_reminders: off -> ON")

	out, err = runAs(t, db, pgtest.RoleOwner, "disable", "module_auth")
	require.NoError(t, err)
	assert.Contains(t, out, "module_auth: ON -> off")

	out, err = runAs(t, db, pgtest.RoleOwner, "presets")
	require.NoError(t, err)
	assert.Contains(t, out, "demo_monetization")

	// What the API reads as cc_app is what the CLI wrote.
	app, err := flagsgorm.Open(db.Connect(t, pgtest.RoleApp))
	require.NoError(t, err)
	keys, err := flagsgorm.NewReader(app).EnabledKeys(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{
		"module_dashboard", "module_expenses", "module_incomes", "module_reminders", "module_tax_engine", "module_taxpayers",
	}, keys)
}

func TestExecute_UnknownFlagAndPreset(t *testing.T) {
	db := pgtest.NewDatabase(t)

	_, err := runAs(t, db, pgtest.RoleOwner, "enable", "module_ghost")
	require.ErrorIs(t, err, flags.ErrUnknownFlag)

	_, err = runAs(t, db, pgtest.RoleOwner, "apply-preset", "demo_ghost")
	require.ErrorIs(t, err, flags.ErrUnknownPreset)
}

func TestExecute_RefusesARuntimeRole(t *testing.T) {
	db := pgtest.NewDatabase(t)

	_, err := runAs(t, db, pgtest.RoleApp, "enable", "module_auth")

	require.ErrorIs(t, err, migrations.ErrNotOwner)
	out, err := runAs(t, db, pgtest.RoleOwner, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "0 of 9 modules ON", "nothing changed")
}
