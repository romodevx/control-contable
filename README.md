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
| [Tax parameter schema](docs/architecture/tax-params-schema.md) | the yearly parameter files, their validation and selection |
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

Some modules read another module's data and declare it as a **prerequisite**
(`flags.Prerequisites` in `backend/internal/platform/flags/prerequisites.go`): such a module is
effectively ON only while its own flag and every prerequisite's are ON. Otherwise its routes
answer `404` and `GET /v1/flags` omits it; `flags list` shows it as `hidden` with a `REQUIRES`
column and says which flag to turn ON. Today `module_incomes`, `module_expenses`,
`module_tax_engine`, `module_dashboard` and `module_reminders` require `module_taxpayers`; `module_history` and
`module_annual` require `module_tax_engine` (and through it `module_taxpayers`). The dashboard does not
require the engine or the record modules: each one OFF removes its part of Home.
Every seeded preset includes the prerequisites of its modules (integration-tested).

A running API sees a change within 5 s (the flag cache TTL). A flag turned ON for a module that
is not built yet changes nothing: its routes keep answering `404`.
Enabling a flag in a shared environment requires the module's tech-lead and security gates.

### Modules

| Module | Flag | Status | Docs |
|---|---|---|---|
| M1 Auth & Account | `module_auth` | Built; gates M1-16 and M1-17 passed. **OFF in shared environments** until the [enable checklist](docs/runbooks/auth.md#6-enable-module_auth-in-a-shared-environment) holds (email provider M1-06 pending) | [module](docs/modules/auth.md), [runbook](docs/runbooks/auth.md) |
| M2 Taxpayers & Onboarding | `module_taxpayers` | Built; gates M2-13 and M2-14 passed. **OFF in shared environments** until the [enable checklist](docs/runbooks/taxpayers.md#5-enable-module_taxpayers-in-a-shared-environment) holds (needs `module_auth` enabled first) | [module](docs/modules/taxpayers.md), [runbook](docs/runbooks/taxpayers.md) |
| M3 Incomes | `module_incomes` | Built; gates M3-12 and M3-13 passed. **OFF in shared environments** until the [enable checklist](docs/runbooks/incomes-expenses.md#5-enable-module_incomes-and-module_expenses-in-a-shared-environment) holds (needs `module_taxpayers`) | [module](docs/modules/incomes.md), [runbook](docs/runbooks/incomes-expenses.md) |
| M4 Expenses | `module_expenses` | Built; gates M4-13 and M4-14 passed. **OFF in shared environments** until the same [enable checklist](docs/runbooks/incomes-expenses.md#5-enable-module_incomes-and-module_expenses-in-a-shared-environment) holds (needs `module_taxpayers`) | [module](docs/modules/expenses.md), [runbook](docs/runbooks/incomes-expenses.md) |
| M5 Tax Engine | `module_tax_engine` | Built; gates M5-30 and M5-31 passed. **OFF in shared environments** until the [security gate preconditions](docs/reviews/M5-31-security-gate.md#5-preconditions-to-enable-module_tax_engine-in-staging-or-production) hold, first the CPA sign-off (M5-33, pending), and a reviewed parameter set for the year in force (needs `module_taxpayers`) | [module](docs/modules/tax-engine.md), [add a regime](docs/guides/add-a-regime.md), [yearly parameter update](docs/runbooks/tax-params-update.md) |
| M6 Dashboard | `module_dashboard` | Built; gates M6-07 and M6-08 passed. **OFF in shared environments** until the [security gate preconditions](docs/reviews/M6-08-security-gate.md#5-preconditions-to-enable-module_dashboard-in-staging-or-production) hold, among them the owner's manual TalkBack/VoiceOver pass and the staging p95 (needs `module_taxpayers`; the engine and record modules only add sections) | [module](docs/modules/dashboard.md), [runbook](docs/runbooks/dashboard.md) |
| M8 History | `module_history` | Built; gates M8-08 and M8-09 passed. **OFF in shared environments** until the owed device walk (M8-07 §5, Docker was down) and the tax engine's own preconditions hold (needs `module_tax_engine`, and through it `module_taxpayers`). Preset: `demo_full` only | [module](docs/modules/history.md) |
| AN Annual estimate | `module_annual` | Built; QA ANN-23, gates ANN-24 and ANN-25 passed. **OFF in shared environments** until the tax engine's own preconditions hold (needs `module_tax_engine`, and through it `module_taxpayers`), a reviewed parameter set for the year with the [annual items](docs/runbooks/tax-params-update.md#8-the-annual-return-module_annual-rmf-deadlines-and-rule-renewals) done, and the CPA sign-off (ANN-27) or the owner's acceptance of the ⚠️ topics. iOS not walked (OB6). Preset: `demo_full` only | [module](docs/modules/annual.md), [user guide (es)](docs/guides/annual-user-guide.md), [ADR-0009](docs/adr/0009-annual-engine-placement.md) |
| WH Withholdings (cross-cutting) | none of its own: inside `module_incomes`, `module_expenses`, `module_tax_engine` | Built; QA WH-13, gates WH-14 and WH-15 passed. Ships with those three flags and their preconditions; CPA review of the ⚠️ rules (WH-17) pending. The backfill is run after a deploy, not before | [module](docs/modules/withholdings.md), [backfill runbook](docs/runbooks/withholdings-backfill.md), [ADR-0008](docs/adr/0008-withholdings-are-derived.md) |
| M7, M9 | `module_*` | Not built | [MODULES](docs/project-specs/MODULES.md), [PLAN](docs/project-specs/PLAN-control-contable.md) |

## API reference

The reference is generated from the contract, never written by hand:

```sh
make api-docs                 # build/api-docs/index.html (Redoc, pinned Redocly CLI)
```

Open the file in a browser. The page loads the Redoc viewer script from Redocly's CDN, so it needs
network access. `build/` is git-ignored: publish the generated page, do not commit it.

Each module page's error table is generated from the contract too: `make api-error-tables`
rewrites it, and CI (`make api-error-tables-check`) fails when it is stale.

## Testing

| Suite | Command | Needs |
|---|---|---|
| Backend unit (`-race`, incl. the Dependency Rule check) | `make -C backend test` | Go |
| Tax engine golden gate, verbose (every case, skip reason, pending count) | `make -C backend golden` | Go |
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
| `contract-ok` | `contract.yml` | ruleset and generator self-tests, OpenAPI lint, module error tables match the contract |
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
