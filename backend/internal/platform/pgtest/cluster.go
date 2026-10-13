//go:build integration

package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Throwaway, container-local credentials. The container listens on a random host
// port and is destroyed with the test binary; nothing here guards real data.
const (
	superUser     = "postgres"
	superPassword = "pgtest-postgres" //nolint:gosec // G101: throwaway password of a disposable test container
	maintenanceDB = "postgres"
	// templateDB is the bootstrapped, migrated database every test database clones.
	templateDB = "control_contable"

	// Same initdb arguments as infra/local/compose.yaml.
	initdbArgs = "--auth-host=scram-sha-256 --auth-local=trust --encoding=UTF8 --locale=C.UTF-8"

	bootstrapRolesTarget = "/bootstrap/roles.sql" // where 10-roles.sh reads it
	readOnlyFileMode     = 0o644
	containerPort        = "5432/tcp"
	cleanupTimeout       = 30 * time.Second
	driverName           = "pgx"
)

var rolePasswords = map[Role]string{
	RoleOwner:   "pgtest-cc-owner",
	RoleApp:     "pgtest-cc-app",
	RoleWebhook: "pgtest-cc-webhook",
	RoleWorker:  "pgtest-cc-worker",
}

// runtimeRoles get CONNECT on every test database, as roles.sql grants on the real one.
var runtimeRoles = []Role{RoleApp, RoleWebhook, RoleWorker}

// Option configures Start.
type Option func(*options)

type options struct {
	migrate Migrator
}

// WithMigrator sets the schema step run once, as cc_owner, on the template database.
func WithMigrator(m Migrator) Option {
	return func(o *options) { o.migrate = m }
}

// Cluster is one bootstrapped PostgreSQL container with a migrated template database.
type Cluster struct {
	container *postgres.PostgresContainer
	endpoint  string  // host:port on the Docker host
	admin     *sql.DB // superuser on the maintenance database: creates and drops test databases
	seq       atomic.Uint64
}

// Database is one test's private copy of the migrated template.
type Database struct {
	Name    string
	cluster *Cluster
}

// shared is the package's cluster, owned by Main.
var shared *Cluster

// Main starts a cluster, runs the package's tests against it and terminates it.
// Call it from TestMain: os.Exit(pgtest.Main(m, opts...)).
func Main(m *testing.M, opts ...Option) int {
	ctx := context.Background()
	cluster, err := Start(ctx, opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	shared = cluster
	code := m.Run()
	if err := cluster.Terminate(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			code = 1
		}
	}
	return code
}

// NewDatabase returns a fresh migrated database on the cluster started by Main.
func NewDatabase(t testing.TB) *Database {
	t.Helper()
	if shared == nil {
		t.Fatal("pgtest: no cluster; call pgtest.Main from TestMain")
	}
	return shared.NewDatabase(t)
}

// Start runs the container, bootstraps the roles with the local stack's own scripts,
// applies the Migrator to the template database as cc_owner and freezes the template.
func Start(ctx context.Context, opts ...Option) (*Cluster, error) {
	o := options{migrate: NoMigrations}
	for _, opt := range opts {
		opt(&o)
	}
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	ctr, err := postgres.Run(ctx, Image,
		postgres.WithDatabase(templateDB),
		postgres.WithUsername(superUser),
		postgres.WithPassword(superPassword),
		testcontainers.WithEnv(bootstrapEnv()),
		postgres.WithInitScripts(filepath.Join(root, bootstrapInitScript)),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      filepath.Join(root, bootstrapRolesSQL),
			ContainerFilePath: bootstrapRolesTarget,
			FileMode:          readOnlyFileMode,
		}),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("pgtest: start %s: %w", Image, err), testcontainers.TerminateContainer(ctr))
	}
	c := &Cluster{container: ctr}
	if err := c.prepare(ctx, o.migrate); err != nil {
		return nil, errors.Join(err, c.Terminate(ctx))
	}
	return c, nil
}

func bootstrapEnv() map[string]string {
	return map[string]string{
		"POSTGRES_INITDB_ARGS": initdbArgs,
		"CC_DB_NAME":           templateDB,
		"CC_OWNER_PASSWORD":    rolePasswords[RoleOwner],
		"CC_APP_PASSWORD":      rolePasswords[RoleApp],
		"CC_WEBHOOK_PASSWORD":  rolePasswords[RoleWebhook],
		"CC_WORKER_PASSWORD":   rolePasswords[RoleWorker],
		"TZ":                   "UTC",
	}
}

func (c *Cluster) prepare(ctx context.Context, migrate Migrator) error {
	endpoint, err := c.container.PortEndpoint(ctx, containerPort, "")
	if err != nil {
		return fmt.Errorf("pgtest: endpoint: %w", err)
	}
	c.endpoint = endpoint
	if err := c.migrateTemplate(ctx, migrate); err != nil {
		return err
	}
	c.admin, err = c.open(ctx, superUser, superPassword, maintenanceDB)
	if err != nil {
		return err
	}
	// Nobody connects to the template again, so it can be cloned concurrently.
	_, err = c.admin.ExecContext(ctx,
		"ALTER DATABASE "+QuoteIdent(templateDB)+" WITH IS_TEMPLATE true ALLOW_CONNECTIONS false")
	if err != nil {
		return fmt.Errorf("pgtest: freeze template: %w", err)
	}
	return nil
}

func (c *Cluster) migrateTemplate(ctx context.Context, migrate Migrator) error {
	owner, err := c.open(ctx, string(RoleOwner), rolePasswords[RoleOwner], templateDB)
	if err != nil {
		return err
	}
	if err := migrate(ctx, owner); err != nil {
		return errors.Join(fmt.Errorf("pgtest: migrate template: %w", err), owner.Close())
	}
	return owner.Close()
}

// Terminate closes the admin pool and removes the container.
func (c *Cluster) Terminate(ctx context.Context) error {
	var closeErr error
	if c.admin != nil {
		closeErr = c.admin.Close()
	}
	return errors.Join(closeErr, testcontainers.TerminateContainer(c.container))
}

// NewDatabase clones the migrated template into a database owned by cc_owner,
// re-applies the bootstrap's database-level grants (CREATE DATABASE does not copy
// them) and drops it when the test ends.
func (c *Cluster) NewDatabase(t testing.TB) *Database {
	t.Helper()
	name := fmt.Sprintf("cc_test_%d", c.seq.Add(1))
	quoted := QuoteIdent(name)
	statements := []string{
		"CREATE DATABASE " + quoted + " TEMPLATE " + QuoteIdent(templateDB) + " OWNER " + QuoteIdent(string(RoleOwner)),
		"REVOKE ALL ON DATABASE " + quoted + " FROM PUBLIC",
		"GRANT CONNECT ON DATABASE " + quoted + " TO " + roleList(runtimeRoles),
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cancel()
		if _, err := c.admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoted+" WITH (FORCE)"); err != nil {
			t.Errorf("pgtest: drop %s: %v", name, err)
		}
	})
	for _, stmt := range statements {
		if _, err := c.admin.ExecContext(t.Context(), stmt); err != nil {
			t.Fatalf("pgtest: create %s: %v", name, err)
		}
	}
	return &Database{Name: name, cluster: c}
}

// Connect logs in to the database as role (with its own password, never SET ROLE)
// and closes the pool when the test ends.
func (d *Database) Connect(t testing.TB, role Role) *sql.DB {
	t.Helper()
	password, ok := rolePasswords[role]
	if !ok {
		t.Fatalf("pgtest: unknown role %q", role)
	}
	db, err := d.cluster.open(t.Context(), string(role), password, d.Name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("pgtest: close %s pool: %v", role, err)
		}
	})
	return db
}

func (c *Cluster) open(ctx context.Context, user, password, database string) (*sql.DB, error) {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     c.endpoint,
		Path:     "/" + database,
		RawQuery: "sslmode=disable",
	}
	db, err := sql.Open(driverName, dsn.String())
	if err != nil {
		return nil, fmt.Errorf("pgtest: open %s@%s: %w", user, database, err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("pgtest: connect %s@%s: %w", user, database, err), db.Close())
	}
	return db, nil
}

func roleList(roles []Role) string {
	list := ""
	for i, r := range roles {
		if i > 0 {
			list += ", "
		}
		list += QuoteIdent(string(r))
	}
	return list
}
