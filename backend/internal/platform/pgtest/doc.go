// Package pgtest is the integration test harness: a real PostgreSQL (testcontainers-go,
// the same postgres:18 image the local stack pins), bootstrapped with the same role
// script (infra/local/postgres), plus helpers to run as a runtime role under a user
// context and to prove Row Level Security isolation.
//
// # Lifecycle
//
// One container per test binary (package), started by Main from TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrate))) }
//
// Start bootstraps the roles, runs the Migrator once as cc_owner against a template
// database, and freezes it. NewDatabase then clones that template (CREATE DATABASE …
// TEMPLATE, milliseconds) so every test gets its own migrated database, dropped at
// cleanup. Tests never share rows, so they may run in parallel and in any order.
//
// The container is not reused across packages on purpose: `go test ./...` runs
// packages as concurrent processes, and a shared container would race on the
// template and be reaped by whichever process finishes first.
//
// # Migrations
//
// The Migrator hook receives a *sql.DB logged in as cc_owner on the template. The
// default is NoMigrations; the goose runner (P0-10) plugs in here.
//
// # Running as a role
//
// Database.Connect logs in as a real role (cc_app, cc_webhook, cc_worker or cc_owner)
// with its password, never SET ROLE from a superuser. AsUser runs a transaction
// under `set_config('app.current_user_id', id, true)`, exactly like the TxManager.
//
// # Proving isolation
//
// AssertIsolation runs every check of the data-model's mandatory RLS test list against
// one table. The Check* functions are the same checks, returning ErrNotIsolated on a
// breach and ErrVacuous when the proof would pass only because the owner has no rows.
// Every probe runs in a transaction that is rolled back.
//
// Container code is behind the `integration` build tag (make test-integration).
package pgtest
