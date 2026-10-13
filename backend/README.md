# Control Contable — backend

Go modular monolith (Clean Architecture). Stack and layout: [ADR-0001](../docs/adr/0001-stack.md);
components: [C4](../docs/architecture/c4.md).

## Prerequisites

- Go 1.27 (`toolchain go1.27.2` is pinned in `go.mod`)
- golangci-lint v2.14+
- Docker, for `make test-integration` (testcontainers-go, `postgres:18`)

## Usage

| Target | What it does |
|---|---|
| `make` / `make check` | `fmt-check` + `lint` + `test` |
| `make fmt` | gofmt + goimports |
| `make lint` | golangci-lint (depguard Dependency Rule, forbidigo float/`time.Now` bans in `domain/` and `platform/money`, gosec) |
| `make test` | unit tests with `-race`, including the `internal/archtest` Dependency Rule check |
| `make test-integration` | tests behind the `integration` build tag (real PostgreSQL) |
| `make generate` | regenerate `adapter/http/gen` from `../docs/api/openapi.yaml`: oapi-codegen strict std-http server (pinned via `go tool`) + operationId/route → `x-module` flag map (`cmd/flagmapgen`) |
| `make codegen-check` | regenerate and fail if the output differs from what is committed (drift check) |
| `make db-up` | local PostgreSQL 18 on `127.0.0.1:5434`, roles bootstrapped, waits for healthy |
| `make db-down` | stop the local stack, keep the data volume |
| `make db-reset` | **LOCAL only**: drop the volume and re-bootstrap (refuses a remote Docker host, a non-loopback `CC_DB_HOST` or a remote `DATABASE_URL`) |
| `make migrate-up` / `migrate-down` / `migrate-status` | **LOCAL only**: run `cmd/migrate` as `cc_owner` against the local stack (apply pending / roll back the latest / list) |
| `make db-roles-check` | log in over TCP as `cc_app`, `cc_webhook`, `cc_worker`; fail if any is SUPERUSER, BYPASSRLS, owner or can DDL |

### Local database

`infra/local/compose.yaml` runs `postgres:18.6` pinned by digest. The first `make db-up`
copies `infra/local/.env.example` to `infra/local/.env` (git-ignored; local placeholder
passwords only). Roles, per [data model](../docs/architecture/data-model.md#database-roles):
`cc_owner` owns the database and schema and runs migrations; `cc_app`, `cc_webhook` and
`cc_worker` get CONNECT + USAGE only and are NOBYPASSRLS. Table grants come from the
baseline migration (P0-10), so every environment gets the same ones.

### Migrations

`migrations/` holds goose v3 SQL files embedded in the binary, named
`<YYYYMMDDHHMMSS>_<snake_case>.sql`, each with an `Up` and a non-empty `Down` section.
`cmd/migrate up|down|status` applies them; it reads the owner URL from
`MIGRATE_DATABASE_URL` (never a flag, never the API's `DATABASE_URL`) and refuses to run as
any role but `cc_owner`. A PostgreSQL advisory lock serializes concurrent runners. Never
`AutoMigrate`.

### Integration tests

`make test-integration` runs the suites behind the `integration` build tag against a real
PostgreSQL through `internal/platform/pgtest`:

- one container per test package (`pgtest.Main` from `TestMain`), same image and digest as
  `infra/local/compose.yaml`, bootstrapped with the same `10-roles.sh` + `roles.sql`;
- a pluggable `Migrator` run once as `cc_owner` on a template database (default
  `NoMigrations`; `pgtest.WithMigrator(migrations.Up)` applies the real goose migrations);
- `pgtest.NewDatabase(t)` clones the template per test (isolated, parallel-safe, dropped at
  cleanup); `db.Connect(t, pgtest.RoleApp)` logs in as the real role;
- `pgtest.AsUser` / `WithoutUser` run a transaction with or without
  `set_config('app.current_user_id', id, true)`;
- `pgtest.AssertIsolation` runs the data model's mandatory RLS checks on one table (ENABLE +
  FORCE, A cannot see/update/delete B, cross-user insert rejected, unset context fails
  closed, no bypass). Each check refuses a vacuous proof (owner without rows).

## Layout

Directories appear with the first task that needs them (no empty placeholders):

```
cmd/{api,worker,migrate}/   composition roots
cmd/flagmapgen/             generator of the operationId → flag-key map (go:generate)
adapter/http/gen/           generated from the contract; never edited by hand
internal/platform/          P0 cross-cutting packages (no flag)
internal/<module>/          domain/ usecase/ adapter/ infra/
internal/archtest/          machine check of the Dependency Rule
migrations/                 goose SQL
```

The Dependency Rule is enforced twice: depguard (third-party and framework imports in
`domain/` and `usecase/`) and `internal/archtest` (layer and cross-module rules that
depguard cannot express). See `internal/archtest/doc.go`.
