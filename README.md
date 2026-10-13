# Control Contable

A mobile app that estimates the monthly taxes a Mexican taxpayer will owe. Persona física and
small persona moral taxpayers record what they earned and spent during the month, and the app
estimates the provisional ISR and the IVA due on the 17th of the following month, with a
breakdown of how it got there. One account can manage several RFCs (your own, a spouse's, a
company's). Product scope: [PRD](docs/project-specs/PRD.md); module backlog:
[MODULES](docs/project-specs/MODULES.md).

> [!IMPORTANT]
> **Informational only. Not tax advice.** Every estimate is an approximation based on LISR,
> LIVA, LIF 2026 and Anexo 8 RMF 2026. It does not replace the filing before the SAT or the
> advice of a *contador*, and the user remains responsible for their own filings. The API
> returns this legend on every estimate (`disclaimers[]`, code `not_tax_advice`) and the app
> must show it next to every figure.

## Architecture

A Go modular monolith (Clean Architecture) on PostgreSQL 18 with Row Level Security, a
background worker, and a Flutter client. The API is spec-first: `docs/api/openapi.yaml` is the
single source of truth for the server stubs, the Dart client, the mock server and the API
reference.

```
backend/   Go API (cmd/api), worker (cmd/worker), migrations, flag CLI (cmd/flags)
mobile/    Flutter app, flavors dev/staging/prod, generated Dart client
docs/      PRD, ADRs, architecture, OpenAPI contract, runbooks, reviews
infra/     local PostgreSQL stack (docker compose, LOCAL ONLY)
tools/     pinned contract tooling (Redocly CLI, Prism), mobile release-artifact gate
```

| Read this | For |
|---|---|
| [C4 model](docs/architecture/c4.md) | system context, containers, components, data flows |
| [Data model](docs/architecture/data-model.md) | tables, database roles, RLS policies |
| [Platform packages](docs/architecture/platform.md) | what each `backend/internal/platform` package guarantees |
| [ADR-0001](docs/adr/0001-stack.md) | stack and pinned versions |
| [ADR-0002](docs/adr/0002-tax-engine.md) | tax engine design |
| [ADR-0003](docs/adr/0003-dart-client-generation.md) | Dart client generation |
| [Contract policy](docs/api/CONTRACT.md) | OpenAPI lint rules |
| [backend/README.md](backend/README.md) | every backend target, configuration variables, worker, migrations |
| [mobile/README.md](mobile/README.md) | flavors, mobile targets, release builds |

## Prerequisites

| Tool | Version | Used for |
|---|---|---|
| Go | 1.27 (`toolchain go1.27.2`, pinned in `backend/go.mod`) | API, worker, migrations, tests |
| golangci-lint | v2.14+ | `make -C backend lint` |
| Docker | with Compose v2 | local PostgreSQL (`127.0.0.1:5434`), integration tests (testcontainers) |
| Node.js | the version in [`.nvmrc`](.nvmrc) (24.21.0; at least 24.18.0) | contract lint, mock server, API reference |
| Flutter | 3.47.7 (Dart 3.13.5) | mobile app and Dart client generation |

Android needs the Android SDK; iOS needs full Xcode and CocoaPods.

## Setup from zero

Repository-wide targets (contract, mock, API reference, mobile) run from the repository root.
Backend and database targets live in `backend/Makefile`: run them with `make -C backend …`.
The full walkthrough, with expected output and troubleshooting, is the
[local development runbook](docs/runbooks/local-dev.md).

```sh
make -C backend init-env      # infra/local/.env with random local passwords (mode 0600, git-ignored)
make -C backend db-up         # PostgreSQL 18 on 127.0.0.1:5434, roles bootstrapped, waits for healthy
make -C backend migrate-up    # apply the goose migrations as cc_owner
make -C backend api           # API on http://localhost:3000 (GET /healthz)
make -C backend worker        # background worker, in a second shell
```

The database targets are **LOCAL only**: they refuse a remote Docker host or a non-loopback
database host. `make -C backend db-reset` destroys the local volume and bootstraps the roles
again.

Mobile work does not need the backend. The Prism mock serves the contract:

```sh
make mock                     # http://127.0.0.1:3001, examples from docs/api/openapi.yaml
make mobile-client            # regenerate the Dart client after a contract change
make mobile-run-mock          # run the dev flavor against the mock
```

## Feature flags

Every product module ships behind a `module_<name>` flag stored in `feature_flags`. **All flags
are OFF by default.** The server is the security boundary: a disabled module answers `404` with
the standard error envelope before any parameter binding, so it is invisible. The app reads
`GET /v1/flags` for UX only.

The flag CLI (`backend/cmd/flags`) runs as `cc_owner` and refuses any other role:

```sh
make -C backend flags                                 # list every flag and its state
make -C backend flags ARGS="enable module_auth"
make -C backend flags ARGS="disable module_auth"
make -C backend flags ARGS="presets"                  # list the presentation presets
make -C backend flags ARGS="apply-preset demo_core"
```

| Preset | Modules |
|---|---|
| `demo_core` | auth, taxpayers, incomes, expenses, tax engine, dashboard |
| `demo_monetization` | `demo_core` + subscriptions |
| `demo_full` | every module, M1 to M9 |

A running API sees a change within 5 s (the flag cache TTL). No product module is built yet, and
a flag turned ON for an unbuilt module makes its routes answer 500 (P0 follow-up FU2, fixed in
M1). Apply presets once their modules ship.
Enabling a flag in a shared environment requires the module's tech-lead and security gates.

## API reference

The reference is generated from the contract, never written by hand:

```sh
make api-docs                 # build/api-docs/index.html (Redoc, pinned Redocly CLI)
```

Open the file in a browser. The page loads the Redoc viewer script from Redocly's CDN, so it needs
network access. `build/` is git-ignored: publish the generated page, do not commit it.

## Testing

| Suite | Command | Needs |
|---|---|---|
| Backend unit (`-race`, incl. the Dependency Rule check) | `make -C backend test` | Go |
| Backend lint + format + unit | `make -C backend check` | Go, golangci-lint |
| Backend integration (real PostgreSQL, RLS isolation) | `make -C backend test-integration` | Docker |
| Contract tests (responses vs `openapi.yaml`) | `make -C backend test-contract` | Docker |
| Backend coverage gate (80 % business logic) | `make -C backend test-integration GOTEST_FLAGS="-covermode=atomic -coverprofile=coverage.out" && make -C backend cover-check` | Docker |
| Contract lint + ruleset self-test | `make contract-lint && make contract-lint-test` | Node |
| Mobile format, analyze, unit/widget | `make mobile-check` | Flutter |
| Mobile coverage gate (80 % non-generated `lib/`) | `make mobile-cover` | Flutter |
| Mobile against the mock | `make mock`, then `make mobile-test-mock` | Node, Flutter |
| Dart client drift | `make mobile-client-check` | Flutter |

## CI

Three workflows in `.github/workflows/`, each ending in one aggregate job. The aggregate jobs are
the required status checks on `dev` and `main`; a workflow whose paths did not change reports
success without running its gates.

| Required check | Workflow | Gates |
|---|---|---|
| `backend-ok` | `backend.yml` | gofmt, golangci-lint, codegen drift, govulncheck, unit tests, integration + contract tests, coverage ≥ 80 % |
| `contract-ok` | `contract.yml` | ruleset self-test, OpenAPI lint |
| `mobile-ok` | `mobile.yml` | format, analyze, tests + coverage ≥ 80 %, Dart client drift and kind guard, obfuscated release build, artifact gate (no symbols or source maps shipped) |

## Contributing

1. Branch from `dev`: `feature/…`, `bugfix/…`, `hotfix/…` or `release/…`. Never commit on
   `dev`, `main` or `master`.
2. Test first: a failing test, then the implementation, then the refactor. Every fixed bug
   leaves a regression test.
3. Contract first: change `docs/api/openapi.yaml`, run `make contract-lint`, then
   `make -C backend generate` and `make mobile-client`, and commit the generated code.
4. [Conventional commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`,
   `refactor:`, `test:`, `chore:`, `perf:`, `ci:`), small and atomic.
5. Open the PR against `dev`. It merges (squash) once `backend-ok`, `mobile-ok` and
   `contract-ok` are green and it passes tech-lead review. Modules also pass a security gate
   before their flag is enabled in a shared environment.
6. Docs are code: update the README, runbooks or package docs in the same PR. Stale docs are bugs.

Work is planned in [PLAN-control-contable.md](docs/project-specs/PLAN-control-contable.md).

## Security

Control Contable handles taxpayer data (RFC, CURP, income and expense records). Report
vulnerabilities privately as described in [SECURITY.md](SECURITY.md). Never open a public issue
for one.
