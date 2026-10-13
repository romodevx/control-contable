package migrations_test

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/migrations"
)

// goose `create` names files <YYYYMMDDHHMMSS>_<snake_case>.sql (timestamped versions).
var fileName = regexp.MustCompile(`^\d{14}_[a-z0-9]+(_[a-z0-9]+)*\.sql$`)

func sqlFiles(t *testing.T) []string {
	t.Helper()
	names, err := fs.Glob(migrations.FS, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, names, "no migrations embedded")
	return names
}

func TestFS_EmbedsOnlyTimestampedSQLFiles(t *testing.T) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	require.NoError(t, err)
	for _, e := range entries {
		assert.Regexp(t, fileName, e.Name())
	}
}

func TestFS_EveryMigrationHasUpAndDown(t *testing.T) {
	for _, name := range sqlFiles(t) {
		t.Run(name, func(t *testing.T) {
			body, err := fs.ReadFile(migrations.FS, name)
			require.NoError(t, err)
			src := string(body)
			up := strings.Index(src, "-- +goose Up")
			down := strings.Index(src, "-- +goose Down")
			require.GreaterOrEqual(t, up, 0, "missing -- +goose Up")
			require.Greater(t, down, up, "missing -- +goose Down after Up")
			assert.NotEmpty(t, strings.TrimSpace(src[down+len("-- +goose Down"):]), "empty Down section")
		})
	}
}

func TestFS_VersionsAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, name := range sqlFiles(t) {
		version := name[:14]
		if prev, dup := seen[version]; dup {
			t.Errorf("version %s used by %s and %s", version, prev, name)
		}
		seen[version] = name
	}
}

func TestNewProvider_RejectsNilDB(t *testing.T) {
	_, err := migrations.NewProvider(nil)
	require.Error(t, err)
}
