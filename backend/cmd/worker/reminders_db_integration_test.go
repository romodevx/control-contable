//go:build integration

package main

import (
	"database/sql"
	"testing"

	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// reminderDatabases is a fresh migrated database as cc_app (fixtures),
// cc_worker (the job) and cc_owner (the flags).
func reminderDatabases(t *testing.T) (app, worker, owner *sql.DB) {
	t.Helper()
	db := pgtest.NewDatabase(t)
	return db.Connect(t, pgtest.RoleApp), db.Connect(t, pgtest.RoleWorker), db.Connect(t, pgtest.RoleOwner)
}
