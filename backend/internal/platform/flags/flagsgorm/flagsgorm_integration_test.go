//go:build integration

package flagsgorm_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagsgorm"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The module flags the migrations seed, all OFF: the baseline's nine and
// module_annual (ANN-22). demo_full holds every one.
var seededKeys = []string{
	"module_annual", "module_auth", "module_dashboard", "module_expenses", "module_history", "module_incomes",
	"module_reminders", "module_subscriptions", "module_tax_engine", "module_taxpayers",
}

var demoCore = []string{
	"module_auth", "module_dashboard", "module_expenses", "module_incomes", "module_tax_engine", "module_taxpayers",
}

func open(t *testing.T, db *sql.DB) *flagsgorm.DB {
	t.Helper()
	g, err := flagsgorm.Open(db)
	require.NoError(t, err)
	return g
}

func enabledKeys(list []flags.Flag) []string {
	keys := []string{}
	for _, f := range list {
		if f.Enabled {
			keys = append(keys, f.Key)
		}
	}
	return keys
}

func allKeys(list []flags.Flag) []string {
	keys := make([]string, 0, len(list))
	for _, f := range list {
		keys = append(keys, f.Key)
	}
	return keys
}

func TestStoreListsTheSeededFlagsOFF(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))

	list, err := store.List(t.Context())

	require.NoError(t, err)
	assert.Equal(t, seededKeys, allKeys(list), "ordered by key")
	assert.Empty(t, enabledKeys(list), "flags default OFF")
	for _, f := range list {
		assert.NotEmpty(t, f.Description, f.Key)
	}
}

func TestStoreSetEnabledReturnsThePreviousState(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))

	previous, err := store.SetEnabled(t.Context(), "module_auth", true)
	require.NoError(t, err)
	assert.False(t, previous)

	previous, err = store.SetEnabled(t.Context(), "module_auth", true)
	require.NoError(t, err)
	assert.True(t, previous, "enabling twice is a no-op")

	previous, err = store.SetEnabled(t.Context(), "module_auth", false)
	require.NoError(t, err)
	assert.True(t, previous)

	list, err := store.List(t.Context())
	require.NoError(t, err)
	assert.Empty(t, enabledKeys(list))
}

func TestStoreSetEnabledOnAnUnknownFlag(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))

	_, err := store.SetEnabled(t.Context(), "module_ghost", true)

	assert.ErrorIs(t, err, flags.ErrUnknownFlag)
}

func TestStoreApplyPresetTurnsItsFlagsONAndTheRestOFF(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))
	_, err := store.SetEnabled(t.Context(), "module_reminders", true) // outside demo_core
	require.NoError(t, err)

	result, err := store.ApplyPreset(t.Context(), "demo_core")

	require.NoError(t, err)
	assert.Equal(t, demoCore, enabledKeys(result))
	assert.Equal(t, seededKeys, allKeys(result), "the result lists every flag")
	list, err := store.List(t.Context())
	require.NoError(t, err)
	assert.Equal(t, result, list, "the result is what is stored")

	result, err = store.ApplyPreset(t.Context(), "demo_full")
	require.NoError(t, err)
	assert.Equal(t, seededKeys, enabledKeys(result))
}

func TestStoreApplyPresetLeavesUnchangedRowsUntouched(t *testing.T) {
	db := pgtest.NewDatabase(t)
	owner := db.Connect(t, pgtest.RoleOwner)
	store := flagsgorm.NewStore(open(t, owner))
	// Backdate every row; the updated_at trigger is paused so it does not
	// overwrite the backdated value with now().
	_, err := owner.ExecContext(t.Context(), `ALTER TABLE feature_flags DISABLE TRIGGER feature_flags_set_updated_at`)
	require.NoError(t, err)
	_, err = owner.ExecContext(t.Context(), `UPDATE feature_flags SET updated_at = '2026-01-01T00:00:00Z'`)
	require.NoError(t, err)
	_, err = owner.ExecContext(t.Context(), `ALTER TABLE feature_flags ENABLE TRIGGER feature_flags_set_updated_at`)
	require.NoError(t, err)

	_, err = store.ApplyPreset(t.Context(), "demo_core")
	require.NoError(t, err)

	var touched []string
	rows, err := owner.QueryContext(t.Context(),
		`SELECT key FROM feature_flags WHERE updated_at > '2026-01-01T00:00:00Z' ORDER BY key`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var k string
		require.NoError(t, rows.Scan(&k))
		touched = append(touched, k)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, demoCore, touched, "only the flags whose state changed are written")
}

func TestStoreApplyPresetOnAnUnknownPresetChangesNothing(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))
	_, err := store.SetEnabled(t.Context(), "module_auth", true)
	require.NoError(t, err)

	_, err = store.ApplyPreset(t.Context(), "demo_ghost")

	require.ErrorIs(t, err, flags.ErrUnknownPreset)
	list, err := store.List(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"module_auth"}, enabledKeys(list))
}

func TestStorePresetsListsEachPresetWithItsFlags(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))

	presets, err := store.Presets(t.Context())

	require.NoError(t, err)
	require.Len(t, presets, 3)
	assert.Equal(t, "demo_core", presets[0].Name)
	assert.Equal(t, demoCore, presets[0].Flags, "flags ordered by key")
	assert.NotEmpty(t, presets[0].Description)
	assert.Equal(t, "demo_full", presets[1].Name)
	assert.Equal(t, seededKeys, presets[1].Flags)
	assert.Equal(t, "demo_monetization", presets[2].Name)
	assert.Contains(t, presets[2].Flags, "module_subscriptions")
}

func TestReaderAsTheAppRoleSeesOnlyEnabledKeys(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))
	reader := flagsgorm.NewReader(open(t, db.Connect(t, pgtest.RoleApp)))

	keys, err := reader.EnabledKeys(t.Context())
	require.NoError(t, err)
	assert.Empty(t, keys)

	_, err = store.ApplyPreset(t.Context(), "demo_core")
	require.NoError(t, err)

	keys, err = reader.EnabledKeys(t.Context())
	require.NoError(t, err)
	assert.Equal(t, demoCore, keys)
}

func TestTheAppRoleCannotToggleFlags(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleApp)))

	_, err := store.SetEnabled(t.Context(), "module_auth", true)

	assert.Error(t, err, "cc_app has SELECT only on feature_flags")
	_, err = store.Presets(t.Context())
	assert.Error(t, err, "presets are owner-only")
}

// The service over the real adapter: ON/OFF follows the table, and a role that
// cannot read the flags (cc_webhook) sees every module OFF.
func TestServiceOverPostgres(t *testing.T) {
	db := pgtest.NewDatabase(t)
	store := flagsgorm.NewStore(open(t, db.Connect(t, pgtest.RoleOwner)))
	svc := flags.NewService(flagsgorm.NewReader(open(t, db.Connect(t, pgtest.RoleApp))))

	assert.False(t, svc.Enabled(t.Context(), "module_auth"))
	_, err := store.SetEnabled(t.Context(), "module_auth", true)
	require.NoError(t, err)
	svc.Invalidate()
	assert.True(t, svc.Enabled(t.Context(), "module_auth"))

	failClosed := flags.NewService(flagsgorm.NewReader(open(t, db.Connect(t, pgtest.RoleWebhook))))
	assert.False(t, failClosed.Enabled(t.Context(), "module_auth"), "permission denied => OFF")
	_, err = failClosed.EnabledModules(t.Context())
	assert.Error(t, err)
}
