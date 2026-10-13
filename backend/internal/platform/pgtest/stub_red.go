package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

var errNotImplemented = errors.New("not implemented")

const Image = ""
const bootstrapRolesSQL = ""

var (
	ErrNoUser      = errors.New("no user")
	ErrNotIsolated = errors.New("not isolated")
	ErrVacuous     = errors.New("vacuous")
)

type Role string

const (
	RoleOwner   Role = "cc_owner"
	RoleApp     Role = "cc_app"
	RoleWebhook Role = "cc_webhook"
	RoleWorker  Role = "cc_worker"
)

type Table struct{ Name, OwnerColumn string }
type IsolationSpec struct {
	Table        Table
	UserA, UserB string
	Insert       func(ctx context.Context, tx *sql.Tx, ownerID string) error
}

func QuoteIdent(s string) string                                         { return s }
func repoRoot() (string, error)                                          { return "", errNotImplemented }
func NoMigrations(context.Context, *sql.DB) error                        { return errNotImplemented }
func AsUser(context.Context, *sql.DB, string, func(*sql.Tx) error) error { return errNotImplemented }
func WithoutUser(context.Context, *sql.DB, func(*sql.Tx) error) error    { return errNotImplemented }
func CheckCannotSee(context.Context, *sql.DB, Table, string, string) error {
	return errNotImplemented
}
func CheckCannotUpdate(context.Context, *sql.DB, Table, string, string) error {
	return errNotImplemented
}
func CheckCannotDelete(context.Context, *sql.DB, Table, string, string) error {
	return errNotImplemented
}
func CheckInsertRejected(context.Context, *sql.DB, string, func(context.Context, *sql.Tx) error) error {
	return errNotImplemented
}
func CheckUnsetContextFailsClosed(context.Context, *sql.DB, Table, string) error {
	return errNotImplemented
}
func CheckNoBypass(context.Context, *sql.DB, Table) error      { return errNotImplemented }
func AssertIsolation(t testing.TB, _ *sql.DB, _ IsolationSpec) { t.Error(errNotImplemented) }

type Migrator func(context.Context, *sql.DB) error
type Option func(*options)
type options struct{}

func WithMigrator(Migrator) Option { return func(*options) {} }

type Cluster struct{}
type Database struct{ Name string }

func Start(context.Context, ...Option) (*Cluster, error) { return nil, errNotImplemented }
func Main(m *testing.M, _ ...Option) int                 { return m.Run() }
func NewDatabase(t testing.TB) *Database                 { t.Fatal(errNotImplemented); return nil }
func (*Database) Connect(t testing.TB, _ Role) *sql.DB   { t.Fatal(errNotImplemented); return nil }
