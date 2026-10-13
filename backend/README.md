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
| `make db-roles-check` | log in over TCP as `cc_app`, `cc_webhook`, `cc_worker`; fail if any is SUPERUSER, BYPASSRLS, owner or can DDL |

### Running the API

`go run ./cmd/api` (release builds: `-ldflags "-X main.version=<semver>"`). Configuration is
read from the environment (`internal/platform/config`):

| Variable | Default | Meaning |
|---|---|---|
| `APP_ENV` | required | `dev`, `staging` or `production` |
| `PORT` | `8080` | listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `DATABASE_URL` | required (secret) | until P0-12 it is only checked for presence |
| `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST` | `10` / `20` | token bucket per client IP (429 + `Retry-After`) |
| `TRUSTED_PROXY_HOPS` | `0` | reverse proxies appending to `X-Forwarded-For`; `0` keys by the TCP peer |

`GET /healthz` answers 200 while the process serves; the body says whether the database is
reachable (`degraded`/`unreachable` until the database adapter lands in P0-12). Every flagged
module answers 404 until the flags service (P0-15) turns it on. SIGINT/SIGTERM drains
in-flight requests for up to 8 s.

### Local database

`infra/local/compose.yaml` runs `postgres:18.6` pinned by digest. The first `make db-up`
copies `infra/local/.env.example` to `infra/local/.env` (git-ignored; local placeholder
passwords only). Roles, per [data model](../docs/architecture/data-model.md#database-roles):
`cc_owner` owns the database and schema and runs migrations; `cc_app`, `cc_webhook` and
`cc_worker` get CONNECT + USAGE only and are NOBYPASSRLS. Table grants come from the
baseline migration (P0-10), so every environment gets the same ones.

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
