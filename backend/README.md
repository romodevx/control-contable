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
