//go:build integration

package migrations_test

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// baselineObjects are what the baseline creates; down must remove every one.
var baselineObjects = []string{
	"to_regprocedure('public.app_user_id()')",
	"to_regprocedure('public.set_updated_at()')",
	"to_regclass('public.feature_flags')",
	"to_regclass('public.flag_presets')",
	"to_regclass('public.flag_preset_flags')",
}

func remainingObjects(t *testing.T, db *sql.DB) []string {
	t.Helper()
	var left []string
	for _, obj := range baselineObjects {
		var found sql.NullString
		require.NoError(t, db.QueryRowContext(t.Context(), "SELECT "+obj+"::text").Scan(&found))
		if found.Valid {
			left = append(left, found.String)
		}
	}
	return left
}

func assertAllApplied(t *testing.T, p *goose.Provider) {
	t.Helper()
	statuses, err := p.Status(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, statuses)
	for _, s := range statuses {
		assert.Equal(t, goose.StateApplied, s.State, s.Source.Path)
	}
}

func TestUp_IsANoOpOnAnUpToDateDatabase(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	results, err := p.Up(t.Context())

	require.NoError(t, err)
	assert.Empty(t, results)
	assertAllApplied(t, p)
}

// up → down to zero → up → up: the down sections undo everything and the
// migrations re-apply cleanly on the result.
func TestMigrations_RoundTripDownToZeroAndBackUp(t *testing.T) {
	owner := pgtest.NewDatabase(t).Connect(t, pgtest.RoleOwner)
	p, err := migrations.NewProvider(owner)
	require.NoError(t, err)

	_, err = p.DownTo(t.Context(), 0)
	require.NoError(t, err)
	assert.Empty(t, remainingObjects(t, owner), "down left objects behind")

	require.NoError(t, migrations.Up(t.Context(), owner))
	assertAllApplied(t, p)
	assert.Len(t, remainingObjects(t, owner), len(baselineObjects))

	require.NoError(t, migrations.Up(t.Context(), owner), "a second up is a no-op")
	var flags int
	require.NoError(t, owner.QueryRowContext(t.Context(), "SELECT count(*) FROM feature_flags").Scan(&flags))
	assert.Equal(t, len(moduleFlags), flags, "seeds applied exactly once")
}

func TestUp_RefusesEveryRoleButTheOwner(t *testing.T) {
	db := pgtest.NewDatabase(t)
	for _, role := range []pgtest.Role{pgtest.RoleApp, pgtest.RoleWebhook, pgtest.RoleWorker} {
		t.Run(string(role), func(t *testing.T) {
			err := migrations.Up(t.Context(), db.Connect(t, role))
			require.ErrorIs(t, err, migrations.ErrNotOwner)
		})
	}
}
