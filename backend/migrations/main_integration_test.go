//go:build integration

package migrations_test

import (
	"os"
	"testing"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

// Every test database is a clone of a template migrated by the real goose
// migrations, exactly as `cmd/migrate up` would leave it.
func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}
