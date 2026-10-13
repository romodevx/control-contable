package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Image is the PostgreSQL image the harness runs: the exact tag and digest pinned in
// infra/local/compose.yaml (a unit test fails if they drift). Bump both together.
const Image = "postgres:18.6@sha256:74935e72241653ca55e0414067e6d8763aceb8a810eb51b452253ec3dcfc4336"

// Bootstrap files of the local stack, relative to the repository root. The harness
// runs them unchanged so tests see the same roles and grants as `make db-up`.
const (
	bootstrapInitScript = "infra/local/postgres/initdb/10-roles.sh"
	bootstrapRolesSQL   = "infra/local/postgres/bootstrap/roles.sql"
)

// Role is a database login role from the bootstrap (docs/architecture/data-model.md).
type Role string

// The bootstrap roles. cc_owner owns the schema and migrates; the others are
// NOSUPERUSER, NOBYPASSRLS runtime logins.
const (
	RoleOwner   Role = "cc_owner"
	RoleApp     Role = "cc_app"
	RoleWebhook Role = "cc_webhook"
	RoleWorker  Role = "cc_worker"
)

// Migrator applies the schema. It runs once per cluster, logged in as cc_owner on
// the template database that every test database is cloned from.
type Migrator func(ctx context.Context, owner *sql.DB) error

// NoMigrations is the default Migrator: the database keeps only the bootstrap.
func NoMigrations(context.Context, *sql.DB) error { return nil }

// QuoteIdent quotes a PostgreSQL identifier (table or column name) for safe
// interpolation into SQL text. Values always go through bind parameters instead.
func QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// repoRoot walks up from the working directory (the package directory under
// `go test`) to the repository root, the first directory holding the role bootstrap.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("pgtest: working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, bootstrapRolesSQL)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("pgtest: repository root not found (no " + bootstrapRolesSQL + " above the working directory)")
		}
		dir = parent
	}
}
