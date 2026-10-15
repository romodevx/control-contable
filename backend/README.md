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
| `make cover-check` | fail if `COVER_PROFILE` (default `coverage.out`) is below `COVER_MIN`% (default 80) on business logic; `adapter/http/gen` and `cmd/` are left out (`cmd/covergate`) |
| `make vuln` | govulncheck, pinned version (same in CI) |
| `make db-up` | local PostgreSQL 18 on `127.0.0.1:5434`, roles bootstrapped, waits for healthy |
| `make db-down` | stop the local stack, keep the data volume |
| `make db-reset` | **LOCAL only**: drop the volume and re-bootstrap (refuses a remote Docker host, a non-loopback `CC_DB_HOST` or a remote `DATABASE_URL`) |
| `make migrate-up` / `migrate-down` / `migrate-status` | **LOCAL only**: run `cmd/migrate` as `cc_owner` against the local stack (apply pending / roll back the latest / list) |
| `make api` / `make worker` | **LOCAL only**: run `cmd/api` (as `cc_app`, `APP_ENV=dev`, port `API_PORT`, default 3000) / `cmd/worker` (as `cc_worker`) against the local stack, URLs built from `infra/local/.env` |
| `make db-roles-check` | log in over TCP as `cc_app`, `cc_webhook`, `cc_worker`; fail if any is SUPERUSER, BYPASSRLS, REPLICATION, a member of another role, owner, can DDL, or can CONNECT to `postgres`/`template1` |
| `make db-psql-app` | **LOCAL only**: run SQL from stdin as `cc_app` (RLS applies), `PSQL_ARGS="-v user_id=…"` for psql variables; used by the [auth runbook](../docs/runbooks/auth.md) |
| `make access-key-rotate` | **LOCAL only**: new access-token signing key and kid in `infra/local/.env`; the old pair stays verify-only (restart `make api`) |

CI (`.github/workflows/backend.yml`) runs `fmt-check`, golangci-lint v2.14.0, `codegen-check`,
`vuln`, `test`, then `test-integration GOTEST_FLAGS="-covermode=atomic -coverprofile=coverage.out"`
followed by `cover-check`, and uploads the profile as the `backend-coverage` artifact.

### Running the API

`go run ./cmd/api` (release builds: `-ldflags "-X main.version=<semver>"`). Locally, from zero:
`make db-reset && make migrate-up && make api` (then `make worker` in another shell).
Configuration is read from the environment (`internal/platform/config`):

| Variable | Default | Meaning |
|---|---|---|
| `APP_ENV` | required | `dev`, `staging` or `production` |
| `PORT` | `8080` | listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `DATABASE_URL` | required (secret) | the `cc_app` pool; today it serves the flags service (P0-15). The pool connects lazily: an unreachable database hides every flagged module (fail closed). In staging/production the API **refuses to start** unless the role is verified at boot as a plain runtime login (not SUPERUSER, not BYPASSRLS, owns nothing, not a member of such a role) |
| `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST` | `10` / `20` | token bucket per client IP (429 + `Retry-After`) |
| `TRUSTED_PROXY_HOPS` | `0` in dev; **required** in staging/production | reverse proxies appending to `X-Forwarded-For`; `0` keys by the TCP peer (behind Cloud Run/a load balancer that would be one bucket for every client) |

`GET /healthz` answers 200 while the process serves; the body says whether the database is
reachable (a bounded `SELECT 1` on the cc_app pool; `degraded`/`unreachable` when it fails). Every flagged
module answers 404 unless its flag is ON in `feature_flags` (see Feature flags). SIGINT/SIGTERM drains
in-flight requests for up to 8 s.

### Running the worker

`go run ./cmd/worker` runs the background jobs (`internal/platform/worker`, ADR-0001 O8). It
reads `APP_ENV`, `LOG_LEVEL` and the `DATABASE_URL` secret, logged in as **`cc_worker`**, and
applies the same startup role check as the API (outside dev it refuses to start as an owner,
SUPERUSER or BYPASSRLS role, or when the role cannot be verified).

- **Push (reminders, M9-07)**: `FCM_PROJECT_ID` (plain variable) turns the FCM sender on;
  unset, the worker logs `push disabled` and the reminder job sends nothing. With a project,
  the service account key is read from the `FCM_CREDENTIALS_JSON` secret (secret manager in
  shared environments, never the repository); without it the runtime's Application Default
  Credentials are used. Credentials that cannot be used stop the worker at startup. Whether
  `module_reminders` is ON is read by the job on every run, not at startup.

- **Leader election**: every instance campaigns at start and every 10 s with
  `pg_try_advisory_lock(pgdb.WorkerLeaderLockKey)` on a connection it reserves from its pool
  while it leads. The leader re-verifies the lock in `pg_locks` on each heartbeat; a dead
  connection is a lost leadership (PostgreSQL releases the session lock, a standby takes over
  at its next heartbeat). Standbys hold no connection between attempts.
- **Jobs** register in `cmd/worker` with a name, an interval and a per-run timeout. Only the
  leader runs them; each run's context is cancelled on timeout, on lost leadership and on
  shutdown. A panic is recovered and logged; logs carry the job name, outcome and duration only.
- **Shutdown** (SIGINT/SIGTERM) cancels running jobs, waits for them, then unlocks.
- Jobs stay idempotent: two instances may briefly both act as leader between a leader's
  session dying and its next heartbeat. No production job is registered yet (M9 adds them).

### Local database

`infra/local/compose.yaml` runs `postgres:18.6` pinned by digest. The first `make db-up`
writes `infra/local/.env` from `infra/local/.env.example` (git-ignored, mode 0600) with
random passwords (`infra/local/scripts/init-env.sh`). If you delete `.env` but keep the data
volume, run `make db-reset`: the roles keep the passwords they were created with. Roles, per [data model](../docs/architecture/data-model.md#database-roles):
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

### Feature flags

Every module ships behind a `module_<name>` flag in `feature_flags`, OFF by default.

- `internal/platform/flags`: `Service` (`Enabled`, `EnabledModules`, `Invalidate`) caches the
  enabled set for `flags.DefaultTTL` (5 s) and **fails closed**: if the set cannot be read,
  every module is OFF. Each read is bounded by `flags.DefaultLoadTimeout` (2 s), so a database
  that hangs still yields a prompt 404; callers queued behind a failed read share its error. `Admin` validates keys and preset names for the CLI.
- `internal/platform/flags/flagsgorm`: GORM adapter. `Reader` runs as `cc_app` (SELECT only);
  `Store` runs as `cc_owner` and applies a preset in one transaction.
- `adapter/http/flaggate`: the gate primitive (`Wrap`, `WriteNotFound`). `adapter/http/router`
  applies it to each route **as the generated server registers it**, keyed by
  `gen.RouteModule(pattern)`. A disabled module therefore answers 404 with the standard
  `ErrorResponse` before any parameter binding or body decoding. Unmatched requests get
  the same 404 (never 405 + `Allow`).
- `adapter/http/flagshttp`: `GET /v1/flags` (UX only).
- `internal/platform/flags/flagtest`: `flagtest.OnOff(t, "module_x", suite)` runs a handler
  suite with the flag ON and OFF; `flagtest.AssertHidden` checks the OFF answer.
- `cmd/flags list | enable <key> | disable <key> | apply-preset <name> | presets` runs as
  `cc_owner` from `DATABASE_URL_OWNER` and refuses any other role. Locally:
  `make flags ARGS="apply-preset demo_core"`. Running API instances see a change within
  the cache TTL.

`cmd/api` wires it: `newFlagService` builds the service over the `cc_app` pool from
`DATABASE_URL`. The service is both the router's `FlagGate` and the source of
`flagshttp.Handler`, which is embedded in the strict server.

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
