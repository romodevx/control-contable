//go:build integration

package jobstore_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// databases is one test's fresh migrated database, as cc_app (fixtures) and
// cc_worker (the job).
func databases(t *testing.T) (app, worker *sql.DB) {
	t.Helper()
	db := pgtest.NewDatabase(t)
	return db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleWorker)
}
