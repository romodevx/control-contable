# PLAN — Control Contable (v1 MVP)

| Field | Value |
|---|---|
| Status | DRAFT (Phase 3 Planning), pending owner approval (human checkpoint before execution) |
| Date | 2026-10-13 |
| Author | eng-manager |
| Inputs | [PRD](PRD.md), [MODULES](MODULES.md), [ADR-0001](../adr/0001-stack.md), [ADR-0002](../adr/0002-tax-engine.md), [C4](../architecture/c4.md), [data model](../architecture/data-model.md), [OpenAPI](../api/openapi.yaml), [research](../research/tax-regimes-mx-2026.md) |

## 0. How to read this plan

- **Estimates are agent-executed hours**, single-point "most likely" values per task. Module and
  milestone totals add the team's **+20% buffer**. Owner wait time (reviews, decisions, store
  approvals) is **not** included: it is listed separately in §6 because it is not ours to estimate.
- These are estimates, not commitments (Clean Coder). The riskiest numbers are M5-04 (verbatim
  transcription of 12 tariff tables), M7 (store sandbox behaviour) and everything that touches
  real devices (M9 push). I re-estimate at the end of each module and record accuracy.
- **Done** column: `[ ]` → `[x]` when the task's PR is merged to `dev` (gates: when the gate passes).
- **Par** column: tasks sharing a group label may be dispatched in parallel (§3). `—` = sequential.
- **Blocked by**: owner items from §6. A blocked task does not block its siblings unless a
  dependency says so; the work-around for each blocker is in §6.
- Every task follows TDD (failing test first). "Contract freeze" tasks review, lint and freeze
  the module's paths in the already-drafted `docs/api/openapi.yaml`; they do not invent new API.

### Per-module cycle (fixed order)

`spec (PRD/MODULES, done) → contract freeze → data + RLS migration → backend TDD → Flutter →
flag registration → QA → tech-lead + security gates (concurrent reviews) → docs`.
Flutter starts from the Prism mock server (P0-26) as soon as the module's contract is frozen,
so it overlaps the backend stages. A module is DONE only when both gates pass.

### Plan-level decisions (need architect ack, flagged for the owner)

1. **`users` table moves from M1 to P0** (P0-11). `entitlements` and `user_usage` are P0 tables
   with an FK to `users`; creating `users` in P0 keeps P0 self-contained. M1 still owns all auth
   behaviour and the other auth tables.
2. **Timestamped goose migration versions** (not sequential integers), so parallel branches
   never collide on a migration number.
3. **Composition-root wiring** (one line per module in `cmd/api`) is added in each module's
   *flag registration* task, which is sequential. Parallel backend tasks therefore never edit
   the same file.
4. **All contract freezes are sequential** (architect, module order): they all edit
   `docs/api/openapi.yaml`, so they cannot run in parallel under the team rule.
5. **The soft-delete purge job** (M9-08) runs on the worker regardless of `module_reminders`.
6. **P0 gates review the backend platform only**; the mobile foundation gets its own tech-lead
   review (P0-31) so backend modules are not held up by mobile tooling.

## 1. Branch plan

- One branch = 1–2 related tasks, named `feature/<module>-<slice>` (`release/<slice>` for REL),
  **always cut from `dev`**, never from `main`. The branch for each task is in the task tables.
- PR → `dev`. Merge (squash) only after tech-lead review of the PR; the module's flag may be
  enabled in shared environments only after both module gates pass.
- QA fixes go on `feature/<module>-qa-fixes` (created only if QA finds defects; regression test first).
- Gate tasks are reviews on open PRs, marked `(PR review)`: no branch of their own.
- `main` / production: only through `release/v1.0.0` (REL-12), with **explicit owner approval**;
  merge commit for the release, tag `v1.0.0`.
- Nothing is ever committed on `dev`, `main` or `master` directly.
- The discovery docs on `feature/project-discovery` merge to `dev` first (P0-01), after the owner approves.

## 2. Tasks by module

### P0 — Foundation (platform, no flag)

Stages: contract tooling (P0-05) → data (P0-10, P0-11) → backend platform → mobile foundation → QA → gates → docs. No flag: the platform is always on.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| P0-01 | Owner approves PRD, MODULES, ADR-0001/0002, C4, data model and this plan; merge discovery docs to `dev` | tech-lead | — | 0.5 | — | `feature/project-discovery` | OB0 | [ ] |
| P0-02 | Context7 check of every pinned Go library API (oapi-codegen, goose, GORM, pgx, testcontainers-go, golang-jwt, libopenapi-validator, firebase/go); notes appended to ADR-0001 | backend-sr | P0-01 | 1.5 | — | `feature/p0-skeleton` | — | [x] |
| P0-03 | Backend skeleton: `go.mod` (go 1.27, toolchain go1.27.2), Clean Architecture layout, Makefile, golangci-lint v2 config with depguard (Dependency Rule) and forbidigo (float/time.Now in domain) | backend-sr | P0-01 | 2 | — | `feature/p0-skeleton` | — | [x] |
| P0-04 | Local stack: docker compose `postgres:18`, role bootstrap (cc_owner, cc_app NOBYPASSRLS, cc_webhook, cc_worker), `make db-up` / `make db-reset` | devops-sr | P0-03 | 2 | PG-P0-B | `feature/p0-local-stack` | — | [x] |
| P0-05 | Contract tooling: OpenAPI lint ruleset (x-module required, single error envelope, examples, security scheme) in CI; freeze platform paths (`/healthz`, `/v1/flags`, `Error`) | architect | P0-01 | 1.5 | — | `feature/p0-contract-lint` | — | [x] |
| P0-06 | oapi-codegen v2 strict std-http generation into `adapter/http/gen`, drift check, generated operationID→flag map, contract test failing on any operation without `x-module` | backend-sr | P0-03, P0-05 | 3 | PG-P0-C | `feature/p0-codegen` | — | [ ] |
| P0-07 | `platform/errors`: error envelope, domain-error→HTTP mapping (TDD) | backend-sr | P0-03 | 2 | PG-P0-B | `feature/p0-platform-errors-config` | — | [x] |
| P0-08 | `platform/config` (env + secret source port) and slog JSON handler with PII redaction (RFC, email never logged) (TDD) | backend-sr | P0-03 | 2.5 | PG-P0-B | `feature/p0-platform-errors-config` | — | [x] |
| P0-09 | Integration test harness: testcontainers-go `postgres:18`, migrations applied, role-switching + RLS assertion helpers (A cannot see B, unset context = 0 rows) | qa-sr | P0-04 | 2.5 | — | `feature/p0-test-harness` | — | [ ] |
| P0-10 | goose v3 runner `cmd/migrate` (embed.FS, timestamped versions) + baseline migration: `app_user_id()`, grants, `feature_flags`, `flag_presets` (global) | db-sr | P0-04, P0-09 | 2.5 | — | `feature/p0-migrations-base` | — | [ ] |
| P0-11 | Migration `users` (identity root, moved from M1 so P0 FKs resolve) + `entitlements` + `user_usage`, ENABLE+FORCE RLS, isolation tests | db-sr | P0-10 | 3.5 | — | `feature/p0-migrations-identity` | — | [ ] |
| P0-12 | TxManager: `set_config('app.current_user_id', $1, true)` per tx, RO/RW variants, fail-closed integration tests | backend-sr | P0-11 | 3 | — | `feature/p0-txmanager` | — | [ ] |
| P0-13 | `platform/money` (Cents, Rate 10^-6, Exact, 128-bit multiply, half-away-from-zero) + Clock port (America/Mexico_City), 100% coverage | backend-sr | P0-03 | 4 | PG-P0-B | `feature/p0-money-clock` | — | [x] |
| P0-14 | `platform/crypto`: argon2id PHC, AES-256-GCM with key_id (RFC), HMAC-SHA256 peppers (rfc_hash, email_hash) (TDD) | backend-sr | P0-03 | 3 | PG-P0-B | `feature/p0-crypto` | — | [x] |
| P0-15 | Flags service + gate middleware (OFF → 404 before handler) + `GET /v1/flags` + flag ON/OFF test helper | backend-sr | P0-06, P0-07, P0-10 | 3.5 | — | `feature/p0-flags` | — | [ ] |
| P0-16 | AccessPolicy port + infra: trial clock, atomic `ReserveRecord` (UPDATE … RETURNING), trial/pro/comp lookup, read-only after trial (TDD + integration) | backend-sr | P0-12, P0-13 | 4 | — | `feature/p0-access-policy` | — | [ ] |
| P0-17 | Composition root + middleware chain (request-id, recover, token-bucket rate limit, authn slot, flag gate, kin-openapi request validation), `/healthz` liveness/readiness | backend-sr | P0-06, P0-07, P0-08 | 3 | — | `feature/p0-http-server` | — | [ ] |
| P0-18 | Contract-test harness (libopenapi-validator) wired to integration tests; applied to `/healthz` and `/v1/flags` | qa-sr | P0-15, P0-17 | 2 | — | `feature/p0-contract-tests` | — | [ ] |
| P0-19 | Backend CI: gofmt, golangci-lint, unit, integration (testcontainers), contract, govulncheck, codegen drift, OpenAPI lint, coverage ≥ 80% gate | devops-sr | P0-06, P0-09 | 3 | — | `feature/p0-ci-backend` | — | [ ] |
| P0-20 | Flag admin CLI `cmd/flags` (list, enable, disable, apply-preset) running as cc_owner (backs the /flags skill) | backend-sr | P0-15 | 2 | — | `feature/p0-flags` | — | [ ] |
| P0-21 | Context7 check of Dart libs + swagger_parser spike on `openapi.yaml` (oneOf/const); record go/fallback decision (openapi-generator + 3.0 overlay) | frontend-sr | P0-01, P0-05 | 4.5 | PG-P0-C | `feature/p0-dart-client-spike` | — | [x] |
| P0-22 | Flutter scaffold `mobile/`: flavors dev/staging/prod, go_router, state management, dio, design tokens, es-MX l10n, analysis_options | frontend-sr | P0-21 | 4 | — | `feature/p0-mobile-scaffold` | OB4, OB6 | [ ] |
| P0-23 | Generated Dart client wiring + `make mobile-client` + drift check. ADR-0003 conditions: (a) `mobile/tool/normalize_openapi.dart` inlines path-level `parameters` into each operation, unit test first, and is the only input `make mobile-client` feeds swagger_parser; (b) CI guard `mobile/tool/check_generated_kinds.dart` runs inside `make mobile-client` and fails it when a generated Dart type's kind (enum/class/typedef) differs from its component schema or two schemas map to one Dart name; its test first feeds a fixture with an inline `oneOf` object holding a `period` enum and must fail on the `Period` collision; (c) file the path-parameter cast bug upstream (carapacik/swagger_parser) and link it in ADR-0003; drop (a) once a fixed release is pinned | frontend-sr | P0-22 | 2 | — | `feature/p0-mobile-client-flags` | — | [ ] |
| P0-24 | FlagsProvider + `useFlag` helper fed by `GET /v1/flags` (widget tests first) | frontend-sr | P0-23 | 2.5 | — | `feature/p0-mobile-client-flags` | — | [ ] |
| P0-25 | Mobile CI: `make mobile-client` (normalizer + ADR-0003 kind guard + drift check), analyze, test, release build `--obfuscate --split-debug-info`, artifact gate (no symbols / maps bundled) | devops-sr | P0-22, P0-23 | 3 | — | `feature/p0-ci-mobile` | — | [ ] |
| P0-26 | Prism mock server from `openapi.yaml` (`make mock`) so mobile work starts before backend | devops-sr | P0-05 | 1 | PG-P0-C | `feature/p0-mock-server` | — | [x] |
| P0-27 | Worker skeleton `cmd/worker`: ticker, `pg_try_advisory_lock` leader election, job registry, graceful shutdown (errgroup) | backend-sr | P0-12, P0-08 | 2.5 | — | `feature/p0-worker` | — | [ ] |
| P0-28 | QA P0: RLS fail-closed suite, flag ON/OFF on platform routes, CI green on both pipelines | qa-sr | P0-16, P0-18, P0-19, P0-20, P0-27 | 2 | — | `feature/p0-qa-fixes` | — | [ ] |
| P0-29 | Tech-lead gate P0 backend platform (canon checklist, Dependency Rule, money package review) | tech-lead | P0-28 | 2 | — | (PR review) | — | [ ] |
| P0-30 | Security gate P0 backend platform (roles, NOBYPASSRLS, crypto params, log redaction, CI scanners) | security | P0-28 | 2.5 | — | (PR review) | — | [ ] |
| P0-31 | Tech-lead review of mobile foundation (scaffold, client generation, FlagsProvider, mobile CI) | tech-lead | P0-24, P0-25 | 1.5 | — | (PR review) | — | [ ] |
| P0-32 | Docs P0: README (prereqs, setup, flags), local runbook, platform package docs | doc-writer | P0-29, P0-30 | 2 | — | `feature/p0-docs` | — | [ ] |

### M1 — Auth & Account (`module_auth`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M1-01 | Contract freeze: `/v1/auth/*`, `/v1/me`, `/v1/legal/documents` (verification flow per owner decision), lint | architect | P0-05 | 1.5 | — | `feature/auth-contract` | OB2 | [ ] |
| M1-02 | Migration `refresh_tokens` + `consents` with RLS + isolation tests | db-sr | P0-11 | 3 | — | `feature/auth-migrations` | — | [ ] |
| M1-03 | Migration `legal_documents` (global seed) + `trial_registry` and `email_verifications` (service-only) + definer fns `auth_lookup_credentials`, `trial_registry_check_and_insert`, `email_verification_upsert` / `_consume` + tests | db-sr | P0-11 | 3 | — | `feature/auth-migrations` | — | [ ] |
| M1-04 | Auth domain: email, password policy, refresh token VO `<uuid>.<secret>` (TDD) | backend-sr | P0-03 | 2 | — | `feature/auth-domain` | — | [ ] |
| M1-05 | Register + verify-email use cases: EmailSender port (dev log adapter), generic duplicate response, consents, trial start via TrialStarter (TDD) | backend-sr | M1-01, M1-02, M1-03, M1-04, P0-29, P0-30 | 4.5 | — | `feature/auth-register` | — | [ ] |
| M1-06 | Email provider adapter (SES / Postmark / Resend) behind EmailSender | backend-sr | M1-05 | 2 | — | `feature/auth-email-provider` | OB2 | [ ] |
| M1-07 | Login + EdDSA JWT (kid rotation) + authn middleware (TDD) | backend-sr | M1-05, P0-17 | 3.5 | — | `feature/auth-session` | — | [ ] |
| M1-08 | Refresh rotation, reuse detection (family revoke), logout (TDD + integration) | backend-sr | M1-07 | 3.5 | — | `feature/auth-session` | — | [ ] |
| M1-09 | `GET/PATCH/DELETE /v1/me` (cascade delete; trial_registry survives) + `GET /v1/legal/documents` | backend-sr | M1-07 | 3 | — | `feature/auth-account` | — | [ ] |
| M1-10 | Auth rate limits + contract tests for every M1 operation + flag ON/OFF suite | qa-sr | M1-08, M1-09, P0-18 | 3 | — | `feature/auth-contract-tests` | — | [ ] |
| M1-11 | Flutter register + verify-email screens with consents (widget tests first, mock server) | frontend-sr | P0-24, P0-26, M1-01 | 4 | — | `feature/auth-mobile-register` | — | [ ] |
| M1-12 | Flutter login screen, secure token storage, refresh interceptor, logout | frontend-sr | M1-11 | 4 | — | `feature/auth-mobile-session` | — | [ ] |
| M1-13 | Flutter profile + in-app account deletion | frontend-sr | M1-12 | 2.5 | — | `feature/auth-mobile-account` | — | [ ] |
| M1-14 | Register flag `module_auth` (migration + preset membership) and client gating | backend-sr | M1-10, M1-13 | 0.5 | — | `feature/auth-flag` | — | [ ] |
| M1-15 | QA M1 on emulators: register → verify → login → refresh → delete; flag OFF = 404 | qa-sr | M1-14 | 3 | — | `feature/auth-qa-fixes` | OB6 | [ ] |
| M1-16 | Tech-lead gate M1 | tech-lead | M1-15 | 1.5 | — | (PR review) | — | [ ] |
| M1-17 | Security gate M1 (enumeration, token reuse, argon2 params, rate limits) | security | M1-15 | 2.5 | — | (PR review) | — | [ ] |
| M1-18 | Docs M1 (module doc, auth runbook, generated API ref) | doc-writer | M1-16, M1-17 | 1 | — | `feature/auth-docs` | — | [ ] |

### M2 — Taxpayers & Onboarding (`module_taxpayers`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M2-01 | Contract freeze: `/v1/regimes`, `/v1/taxpayers/**`, lint | architect | P0-05 | 1.5 | — | `feature/taxpayers-contract` | — | [ ] |
| M2-02 | Migration `regimes` + `regime_compatibility` (global) seeded from research A.1/B.1 (option_defs, required_inputs) | db-sr | P0-10 | 3 | — | `feature/taxpayers-catalog-migration` | — | [ ] |
| M2-03 | Migration `taxpayers` + `taxpayer_regimes` with RLS, composite FKs, immutable person_type grant, isolation tests | db-sr | P0-11, M2-02 | 3.5 | — | `feature/taxpayers-migration` | — | [ ] |
| M2-04 | Taxpayer domain: RFC VO (PF 13 / PM 12, check digit), PersonType, regime option validation (table-driven TDD) | backend-sr | P0-03 | 3 | — | `feature/taxpayers-domain` | — | [ ] |
| M2-05 | Taxpayer use cases: create (encrypt + hash RFC, soft cap 10), list, get, patch, archive, delete | backend-sr | M2-01, M2-03, M2-04, P0-29, P0-30 | 4 | — | `feature/taxpayers-usecases` | — | [ ] |
| M2-06 | Regime assignment use cases (compatibility, options, end period) + TaxpayerRegimeChecker and TaxpayerProfileReader adapters | backend-sr | M2-05 | 4 | — | `feature/taxpayers-usecases` | — | [ ] |
| M2-07 | M2 handlers incl. `GET /v1/regimes`, contract tests, flag ON/OFF | backend-sr | M2-06, M1-07 | 3 | — | `feature/taxpayers-http` | — | [ ] |
| M2-08 | Flutter onboarding steps 1–2: persona física/moral first, then RFC | frontend-sr | M1-13, M2-01 | 3 | — | `feature/taxpayers-mobile-onboarding` | — | [ ] |
| M2-09 | Flutter onboarding steps 3–4: regimes filtered by person type + per-regime options | frontend-sr | M2-08 | 4 | — | `feature/taxpayers-mobile-onboarding` | — | [ ] |
| M2-10 | Flutter active-taxpayer switcher + taxpayer settings (edit, add/archive regime, archive) | frontend-sr | M2-09 | 4 | — | `feature/taxpayers-mobile-settings` | — | [ ] |
| M2-11 | Register flag `module_taxpayers` | backend-sr | M2-07, M2-10 | 0.5 | — | `feature/taxpayers-flag` | — | [ ] |
| M2-12 | QA M2 (onboarding PF and PM, duplicate RFC, cross-user 404, flag OFF) | qa-sr | M2-11 | 3 | — | `feature/taxpayers-qa-fixes` | OB6 | [ ] |
| M2-13 | Tech-lead gate M2 | tech-lead | M2-12 | 1.5 | — | (PR review) | — | [ ] |
| M2-14 | Security gate M2 (RFC encryption/HMAC, composite-FK isolation) | security | M2-12 | 2 | — | (PR review) | — | [ ] |
| M2-15 | Docs M2 | doc-writer | M2-13, M2-14 | 1 | — | `feature/taxpayers-docs` | — | [ ] |

### M3 — Incomes (`module_incomes`)

M3 and M4 run side by side (different tables, packages and screens). Their Flutter tasks are sequential because M4 reuses M3's list/form widgets.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M3-01 | Contract freeze: incomes paths, lint | architect | P0-05 | 1 | — | `feature/incomes-contract` | — | [ ] |
| M3-02 | Migration `incomes` with RLS, composite FK to taxpayer_regimes, partial indexes, isolation tests, EXPLAIN ANALYZE | db-sr | M2-03 | 3 | PG-M34 | `feature/incomes-migration` | — | [ ] |
| M3-03 | Income domain: amounts, dates (cash vs accrual), withholdings (TDD) | backend-sr | P0-13 | 2.5 | PG-M34 | `feature/incomes-domain` | — | [ ] |
| M3-04 | Income create/update use cases: ReserveRecord (402), regime check, RateCatalog date validity, idempotent client id | backend-sr | M3-01, M3-02, M3-03, M2-06, P0-16, M5-08 | 4 | — | `feature/incomes-usecases` | — | [ ] |
| M3-05 | Income list (keyset, period + regime filter), soft delete, restore | backend-sr | M3-04 | 3 | — | `feature/incomes-usecases` | — | [ ] |
| M3-06 | Income handlers, contract tests, flag ON/OFF, trial-cap 402 test | backend-sr | M3-05, M2-07 | 2.5 | — | `feature/incomes-http` | — | [ ] |
| M3-07 | LedgerReader income aggregates adapter (GROUP BY month, regime) + EXPLAIN ANALYZE | backend-sr | M3-02, M5-02 | 2.5 | — | `feature/incomes-ledger-reader` | — | [ ] |
| M3-08 | Flutter income list: month navigation, regime filter, delete + persistent Undo | frontend-sr | M2-10, M3-01 | 4 | — | `feature/incomes-mobile` | — | [ ] |
| M3-09 | Flutter income form as its own page (>6 fields), defaults prefilled, save-and-create-another | frontend-sr | M3-08 | 4.5 | — | `feature/incomes-mobile` | — | [ ] |
| M3-10 | Register flag `module_incomes` | backend-sr | M3-06, M3-07, M3-09 | 0.5 | — | `feature/incomes-flag` | — | [ ] |
| M3-11 | QA M3 (CRUD, undo, trial cap, flag OFF) | qa-sr | M3-10 | 2.5 | — | `feature/incomes-qa-fixes` | OB6 | [ ] |
| M3-12 | Tech-lead gate M3 | tech-lead | M3-11 | 1 | — | (PR review) | — | [ ] |
| M3-13 | Security gate M3 | security | M3-11 | 1.5 | — | (PR review) | — | [ ] |
| M3-14 | Docs M3 | doc-writer | M3-12, M3-13 | 1 | — | `feature/incomes-docs` | — | [ ] |

### M4 — Expenses (`module_expenses`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M4-01 | Contract freeze: expenses + `/v1/expense-categories`, lint | architect | P0-05 | 1 | — | `feature/expenses-contract` | — | [ ] |
| M4-02 | Migration `expense_categories` + `category_regime_rules` (global seed with legal basis) | db-sr | M2-02 | 2.5 | PG-M34 | `feature/expenses-migrations` | — | [ ] |
| M4-03 | Migration `expenses` with RLS, composite FK, partial indexes, isolation tests | db-sr | M2-03, M4-02 | 3 | — | `feature/expenses-migrations` | — | [ ] |
| M4-04 | Expense domain incl. cash-payment threshold warning (TDD) | backend-sr | P0-13 | 2.5 | PG-M34 | `feature/expenses-domain` | — | [ ] |
| M4-05 | Expense create/update use cases (ReserveRecord, regime check, RateCatalog, category rules) | backend-sr | M4-01, M4-03, M4-04, M2-06, P0-16, M5-08 | 3.5 | — | `feature/expenses-usecases` | — | [ ] |
| M4-06 | Expense list, soft delete, restore | backend-sr | M4-05 | 2.5 | — | `feature/expenses-usecases` | — | [ ] |
| M4-07 | Expense handlers + categories endpoint, contract tests, flag ON/OFF | backend-sr | M4-06, M2-07 | 3 | — | `feature/expenses-http` | — | [ ] |
| M4-08 | LedgerReader expense aggregates (by deductibility class, IVA paid, withholdings made) + EXPLAIN ANALYZE | backend-sr | M4-03, M5-02 | 2.5 | — | `feature/expenses-ledger-reader` | — | [ ] |
| M4-09 | Flutter expense list (reuses M3 list widgets) | frontend-sr | M3-09, M4-01 | 3 | — | `feature/expenses-mobile` | — | [ ] |
| M4-10 | Flutter expense form page: category with deductibility hint, payment method, cash warning | frontend-sr | M4-09 | 4.5 | — | `feature/expenses-mobile` | — | [ ] |
| M4-11 | Register flag `module_expenses` | backend-sr | M4-07, M4-08, M4-10 | 0.5 | — | `feature/expenses-flag` | — | [ ] |
| M4-12 | QA M4 | qa-sr | M4-11 | 2.5 | — | `feature/expenses-qa-fixes` | OB6 | [ ] |
| M4-13 | Tech-lead gate M4 | tech-lead | M4-12 | 1 | — | (PR review) | — | [ ] |
| M4-14 | Security gate M4 | security | M4-12 | 1.5 | — | (PR review) | — | [ ] |
| M4-15 | Docs M4 | doc-writer | M4-13, M4-14 | 1 | — | `feature/expenses-docs` | — | [ ] |

### M5 — Tax Engine (`module_tax_engine`)

Task groups:
- **Core** (M5-01…M5-10): contract, pure domain, params loader + 2026 data, tariff lookup, golden harness, RateCatalog, shared IVA components, registry.
- **One group per regime strategy** (M5-11…M5-19): 612 PF, 606 PF, 626 PF, 625 PF, 601 PM, 626 PM, partial 621 PF, partial 622/603 PM, explanation-only catalog. Each strategy is one file + its golden cases under `testdata/golden/<regime>-<pf|pm>/`; **none edits the registry or the params file** (params are transcribed once in M5-04, the registry is wired in M5-22), which is what makes them parallel-safe.
- **Golden cases** come from research §A.5 and §B.5 (PF-612-ISR-01..03, PF-612-IVA-01, PF-626-ISR-01..02, PF-626-IVA-01, PF-606-ISR-01..02, PF-606-IVA-01, PF-625-01..02, PF-621-01, PM-601-ISR-01..02, PM-601-IVA-01, PM-626-ISR-01) plus generated tariff-row and validity-date boundaries. ⚠️ cases ship as `skip` with reason until CPA sign-off (M5-33).
- **Use case + API + Flutter** (M5-20…M5-27), then flag, QA, gates, docs.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M5-01 | Contract freeze: estimates, breakdown, tax-inputs, Disclaimer codes, `legend` const, lint | architect | P0-05 | 2 | — | `feature/tax-engine-contract` | — | [ ] |
| M5-02 | Engine domain: Strategy port, StrategyKey (regime, person type), Input/MonthSums, RegimeEstimate, Line, Ledger trace helper, Citation; consumer ports LedgerReader/TaxInputReader (TDD) | backend-sr | P0-13 | 4 | — | `feature/tax-engine-domain` | — | [ ] |
| M5-03 | Params loader: YAML schema, go:embed, `Params.At(key, date)`, overlap rejection, typed fail-fast validation (contiguity, 12 tables, rates in [0,1]), version selection + provisional fallback (TDD) | backend-sr | M5-02 | 5 | — | `feature/tax-engine-params` | — | [ ] |
| M5-04 | Transcribe `params/2026/v1.yaml` verbatim: art. 96, twelve art. 106 tables, RESICO PF brackets, platform rates, UMA Jan vs Feb, IVA 16/8 (valid to 2026-12-31), sources[] | backend-sr | M5-03 | 6 | — | `feature/tax-engine-params-2026` | — | [ ] |
| M5-05 | Tariff lookup (sort.Search) + property tests (total over (0, ∞)) | backend-sr | M5-03 | 2 | — | `feature/tax-engine-params` | — | [ ] |
| M5-06 | Independent double-entry verification of `v1.yaml` against Anexo 8 / LIF / DOF (review on the M5-04 PR) | qa-sr | M5-04 | 3 | — | `feature/tax-engine-params-2026` | — | [ ] |
| M5-07 | Golden harness: case YAML format, table-driven runner, skip-with-reason for ⚠️, no update flag, CI hard gate, boundary-case generator (tariff rows, validity dates) | qa-sr | M5-02, M5-05 | 4 | — | `feature/tax-engine-golden-harness` | — | [ ] |
| M5-08 | RateCatalog port implementation (allowed IVA rates on a date) for M3/M4 | backend-sr | M5-04 | 2 | PG-M5-F | `feature/tax-engine-rate-catalog` | — | [ ] |
| M5-09 | Shared IVA components calculator (charged, creditable, withheld from taxpayer, withheld to remit, 8% by record date, 625 definitive exclusion hook) | backend-sr | M5-02, M5-03 | 3.5 | PG-M5-F | `feature/tax-engine-iva` | — | [ ] |
| M5-10 | Strategy registry + InformationalStrategy fallback + RequiredInputs → `incomplete` status | backend-sr | M5-02 | 2.5 | PG-M5-F | `feature/tax-engine-registry` | — | [ ] |
| M5-11 | Strategy 612 PF (YTD cash, deductions, PTU, losses, accumulated tariff month n, prior payments, 10% PM withholding) + IVA; goldens PF-612-ISR-01..03, PF-612-IVA-01 + 12-month boundary rows | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 6 | PG-M5-S | `feature/tax-engine-612-pf` | — | [ ] |
| M5-12 | Strategy 606 PF (blind 35% vs actual, predial, art. 96, PM 10% withholding → saldo a favor; residential exempt; quarterly → informational); goldens PF-606-ISR-01..02, PF-606-IVA-01 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 4.5 | PG-M5-S | `feature/tax-engine-606-pf` | — | [ ] |
| M5-13 | Strategy 626 PF (non-marginal bracket, 1.25% withholding); goldens PF-626-ISR-01..02, PF-626-IVA-01; property: rate non-decreasing | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3.5 | PG-M5-S | `feature/tax-engine-626-pf` | — | [ ] |
| M5-14 | Strategy 625 PF (activities 2.1/4/2.5%, 20% without RFC, definitive option, prior-year ≤ $300k → full vs partial); goldens PF-625-01 (IVA part skipped ⚠️), PF-625-02 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 4.5 | PG-M5-S | `feature/tax-engine-625-pf` | — | [ ] |
| M5-15 | Strategy 601 PM (coeficiente × YTD accrual, PTU May–Dec, losses, interest withholding, first fiscal year) + IVA + withholdings_to_remit; goldens PM-601-ISR-01..02, PM-601-IVA-01 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 6 | PG-M5-S | `feature/tax-engine-601-pm` | — | [ ] |
| M5-16 | Strategy 626 PM (YTD cash profit × 30%); golden PM-626-ISR-01 + boundaries | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3.5 | PG-M5-S | `feature/tax-engine-626-pm` | — | [ ] |
| M5-17 | Partial 621 PF (bimonthly, February accumulated table, rif_year reduction ⚠️); golden PF-621-01 skipped until CPA | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3 | PG-M5-S | `feature/tax-engine-partial` | — | [ ] |
| M5-18 | Partial 622 PM (IVA + explanation, AGAPES deferred) and 603 PM (IVA only, ISR explained) | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09, M5-15 | 2.5 | — | `feature/tax-engine-partial` | — | [ ] |
| M5-19 | Explanation-only catalog: plain-language text for 605, 607, 608, 610 (PF/PM), 611, 614, 615, 616, 620, 623, 624 (forward migration on `regimes.explanation`) + informational goldens | backend-sr | M2-02, M5-10, M5-07 | 3 | PG-M5-S | `feature/tax-engine-informational` | — | [ ] |
| M5-20 | Migration `taxpayer_tax_inputs` with RLS + isolation tests | db-sr | M2-03 | 2.5 | — | `feature/tax-engine-inputs` | — | [ ] |
| M5-21 | Tax inputs GET/PUT use case validated against `regimes.required_inputs` | backend-sr | M5-20, M5-01, M2-06 | 3 | — | `feature/tax-engine-inputs` | — | [ ] |
| M5-22 | EstimateMonth use case: readers, register all strategies at the composition root, IVA netting per taxpayer, withholdings_to_remit, final rounding + `Ajuste por redondeo`, fingerprint, disclaimers + legend | backend-sr | M5-08, M5-10, M5-11, M5-12, M5-13, M5-14, M5-15, M5-16, M5-17, M5-18, M5-19, M5-21, M3-07, M4-08 | 5 | — | `feature/tax-engine-estimate` | — | [ ] |
| M5-23 | Estimates + breakdown handlers, contract tests, flag ON/OFF (M6/M8 implied OFF), EXPLAIN ANALYZE, p95 check | backend-sr | M5-22, M2-07 | 3.5 | — | `feature/tax-engine-http` | — | [ ] |
| M5-24 | Engine property suite, fixed seed (lines + adjustment = result, monotonic in income, byte-identical JSON) | qa-sr | M5-22 | 3 | — | `feature/tax-engine-property-tests` | — | [ ] |
| M5-25 | Flutter estimate screen: per-regime + totals, legend, disclaimers, `incomplete` CTA | frontend-sr | M5-01, M2-10 | 4.5 | — | `feature/tax-engine-mobile-estimate` | — | [ ] |
| M5-26 | Flutter breakdown screen: lines, legal basis citations, rounding adjustment | frontend-sr | M5-25 | 4 | — | `feature/tax-engine-mobile-breakdown` | — | [ ] |
| M5-27 | Flutter tax-inputs form per regime (target of the `incomplete` CTA) | frontend-sr | M5-25 | 4 | — | `feature/tax-engine-mobile-inputs` | — | [ ] |
| M5-28 | Register flag `module_tax_engine` | backend-sr | M5-23, M5-24, M5-26, M5-27 | 0.5 | — | `feature/tax-engine-flag` | — | [ ] |
| M5-29 | QA M5: UI figures equal golden figures for each full regime, informational regimes, incomplete flow, flag OFF | qa-sr | M5-28 | 4 | — | `feature/tax-engine-qa-fixes` | OB6 | [ ] |
| M5-30 | Tech-lead gate M5 (OCP, purity, no floats, determinism) | tech-lead | M5-29 | 3 | — | (PR review) | — | [ ] |
| M5-31 | Security gate M5 | security | M5-29 | 2 | — | (PR review) | — | [ ] |
| M5-32 | Docs M5: engine guide, add-a-regime guide, Dec–Jan parameter update runbook | doc-writer | M5-30, M5-31 | 2.5 | — | `feature/tax-engine-docs` | — | [ ] |
| M5-33 | CPA sign-off of golden suite; un-skip ⚠️ cases or record decisions | qa-sr | M5-29 | 2 | — | `feature/tax-engine-cpa-signoff` | OB5 | [ ] |

### M6 — Dashboard (`module_dashboard`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M6-01 | Contract freeze: `/v1/taxpayers/{id}/dashboard` | architect | P0-05 | 1 | — | `feature/dashboard-contract` | — | [ ] |
| M6-02 | Dashboard use case: one aggregate pass for 12 months, Estimator, income vs expense, trial banner via AccessPolicy | backend-sr | M5-22, M6-01 | 4 | — | `feature/dashboard-backend` | — | [ ] |
| M6-03 | Dashboard handler, contract tests, flag ON/OFF, benchmark p95 < 1 s, EXPLAIN ANALYZE | backend-sr | M6-02, M5-23 | 3 | — | `feature/dashboard-backend` | — | [ ] |
| M6-04 | Flutter Home: this month's estimate, income vs expense, trend chart, trial banner, empty state | frontend-sr | M6-01, M5-25 | 5 | — | `feature/dashboard-mobile` | — | [ ] |
| M6-05 | Register flag `module_dashboard` + preset `demo_core` | backend-sr | M6-03, M6-04 | 0.5 | — | `feature/dashboard-flag` | — | [ ] |
| M6-06 | QA M6 + `demo_core` preset smoke | qa-sr | M6-05 | 2.5 | — | `feature/dashboard-qa-fixes` | OB6 | [ ] |
| M6-07 | Tech-lead gate M6 | tech-lead | M6-06 | 1 | — | (PR review) | — | [ ] |
| M6-08 | Security gate M6 | security | M6-06 | 1 | — | (PR review) | — | [ ] |
| M6-09 | Docs M6 | doc-writer | M6-07, M6-08 | 1 | — | `feature/dashboard-docs` | — | [ ] |

### M7 — Subscriptions (`module_subscriptions`)

Backend can start right after P0 + M1-07, in parallel with M5/M6. With the flag OFF, P0 trial rules still apply and demos use comp entitlements.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M7-01 | Contract freeze: `/v1/subscription`, `/sync`, `/v1/webhooks/revenuecat` | architect | P0-05 | 1 | — | `feature/subscriptions-contract` | — | [ ] |
| M7-02 | Migration `subscriptions` (RLS) + `revenuecat_events` (service-only) + cc_webhook grants + isolation tests | db-sr | P0-11 | 3 | PG-M79 | `feature/subscriptions-migration` | — | [ ] |
| M7-03 | Subscription domain: state, out-of-order guard (rc_state_at), EntitlementWriter port (TDD) | backend-sr | P0-16 | 2.5 | PG-M79 | `feature/subscriptions-domain` | — | [ ] |
| M7-04 | RevenueCat REST client (subscriber lookup) against a fake server; Context7 first | backend-sr | P0-08 | 3 | PG-M79 | `feature/subscriptions-rc-client` | OB8 (live check only) | [ ] |
| M7-05 | Webhook handler: constant-time bearer, idempotent insert, TEST/env skip, TRANSFER, 503 when RC down, cc_webhook pool | backend-sr | M7-01, M7-02, M7-03, M7-04, P0-17 | 4.5 | — | `feature/subscriptions-webhook` | — | [ ] |
| M7-06 | `GET /v1/subscription` + `POST /v1/subscription/sync`, entitlement upsert, contract tests, flag ON/OFF | backend-sr | M7-05, M1-07 | 3 | — | `feature/subscriptions-sync` | — | [ ] |
| M7-07 | Nightly reconcile job for unprocessed events (worker) | backend-sr | M7-05, P0-27 | 2 | — | `feature/subscriptions-sync` | — | [ ] |
| M7-08 | Flutter: purchases_flutter init (appUserID), paywall from offerings, restore, post-purchase sync | frontend-sr | M7-01, M1-12 | 5 | — | `feature/subscriptions-mobile-paywall` | OB8 | [ ] |
| M7-09 | Flutter: plan status in Settings + read-only mode UX after trial (402 handling) | frontend-sr | M7-08 | 3 | — | `feature/subscriptions-mobile-status` | — | [ ] |
| M7-10 | Store + RevenueCat setup: subscription products (weekly/monthly/yearly), entitlement `pro`, offerings, webhook secret in secret manager | devops-sr | P0-01 | 2.5 | — | `feature/subscriptions-store-config` | OB1, OB7, OB8 | [ ] |
| M7-11 | Register flag `module_subscriptions` + preset `demo_monetization` | backend-sr | M7-06, M7-07, M7-09 | 0.5 | — | `feature/subscriptions-flag` | — | [ ] |
| M7-12 | QA M7: sandbox purchase, renewal, cancel, restore, out-of-order and duplicate webhooks | qa-sr | M7-11, M7-10 | 4 | — | `feature/subscriptions-qa-fixes` | OB6, OB7, OB8 | [ ] |
| M7-13 | Tech-lead gate M7 | tech-lead | M7-12 | 1.5 | — | (PR review) | — | [ ] |
| M7-14 | Security gate M7 (webhook auth, replay, entitlement escalation) | security | M7-12 | 3 | — | (PR review) | — | [ ] |
| M7-15 | Docs M7 (incl. store/RC runbook) | doc-writer | M7-13, M7-14 | 1 | — | `feature/subscriptions-docs` | — | [ ] |

### M8 — History (`module_history`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M8-01 | Contract freeze: history paths | architect | P0-05 | 1 | — | `feature/history-contract` | — | [ ] |
| M8-02 | Migration `estimate_views` with RLS + isolation tests | db-sr | M2-03 | 2 | — | `feature/history-migration` | — | [ ] |
| M8-03 | History use case: fiscal-year summaries in one aggregate pass, changed-since-viewed via fingerprint, mark viewed | backend-sr | M5-22, M8-01, M8-02 | 3.5 | — | `feature/history-backend` | — | [ ] |
| M8-04 | History handlers, contract tests, flag ON/OFF (implies M5) | backend-sr | M8-03, M5-23 | 2 | — | `feature/history-backend` | — | [ ] |
| M8-05 | Flutter history list, changed marker, open breakdown (reuses M5-26) | frontend-sr | M8-01, M5-26 | 4 | — | `feature/history-mobile` | — | [ ] |
| M8-06 | Register flag `module_history` | backend-sr | M8-04, M8-05 | 0.5 | — | `feature/history-flag` | — | [ ] |
| M8-07 | QA M8 | qa-sr | M8-06 | 2 | — | `feature/history-qa-fixes` | OB6 | [ ] |
| M8-08 | Tech-lead gate M8 | tech-lead | M8-07 | 1 | — | (PR review) | — | [ ] |
| M8-09 | Security gate M8 | security | M8-07 | 1 | — | (PR review) | — | [ ] |
| M8-10 | Docs M8 | doc-writer | M8-08, M8-09 | 0.5 | — | `feature/history-docs` | — | [ ] |

### M9 — Reminders (`module_reminders`)

M9-07 (FCM) and M9-09 (Flutter push) are built against fakes; live verification waits for the Firebase project (OB9).

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M9-01 | Contract freeze: devices, reminder-settings, tax-calendar | architect | P0-05 | 1 | — | `feature/reminders-contract` | — | [ ] |
| M9-02 | Migration `devices`, `reminder_settings`, `notification_log` (RLS) + definer fns `due_reminder_users`, `claim_push_token` + isolation tests | db-sr | P0-11 | 3.5 | — | `feature/reminders-migrations` | — | [ ] |
| M9-03 | Migration `holidays` (global) seeded 2026–2027 with sources | db-sr | P0-10 | 1.5 | — | `feature/reminders-migrations` | — | [ ] |
| M9-04 | TaxCalendar domain (day 17 → next business day) + `GET /v1/tax-calendar` (TDD) | backend-sr | M9-01, M9-03, P0-13 | 3 | — | `feature/reminders-calendar` | — | [ ] |
| M9-05 | Devices + reminder settings use cases and handlers, contract tests, flag ON/OFF | backend-sr | M9-02, M1-07 | 3.5 | — | `feature/reminders-settings` | — | [ ] |
| M9-06 | Reminder job: due users, per-user tx, `notification_log` exactly-once, bounded errgroup, optional amounts when module_tax_engine ON | backend-sr | P0-27, M9-04, M9-05, M5-22 | 4.5 | — | `feature/reminders-worker` | — | [ ] |
| M9-07 | FCM sender adapter (Firebase Admin) + UNREGISTERED cleanup, fake in tests | backend-sr | P0-08 | 2.5 | PG-M79 | `feature/reminders-fcm` | OB9 (live check only) | [ ] |
| M9-08 | Soft-delete purge job (30 days) via `purge_soft_deleted` definer fn (runs regardless of reminder flag) | db-sr | P0-27, M3-02, M4-03 | 2 | — | `feature/records-purge-job` | — | [ ] |
| M9-09 | Flutter push: firebase_messaging, permission prompt, device registration | frontend-sr | M9-01, M1-12 | 3.5 | — | `feature/reminders-mobile-push` | OB9, OB7 | [ ] |
| M9-10 | Flutter reminder settings screen | frontend-sr | M9-09 | 3 | — | `feature/reminders-mobile-settings` | — | [ ] |
| M9-11 | Register flag `module_reminders` + preset `demo_full` | backend-sr | M9-06, M9-07, M9-10 | 0.5 | — | `feature/reminders-flag` | — | [ ] |
| M9-12 | QA M9 on real devices (push on iOS and Android, holiday shift, exactly-once) | qa-sr | M9-11, M9-08 | 3 | — | `feature/reminders-qa-fixes` | OB6, OB7, OB9 | [ ] |
| M9-13 | Tech-lead gate M9 | tech-lead | M9-12 | 1.5 | — | (PR review) | — | [ ] |
| M9-14 | Security gate M9 (lock-screen amounts opt-in, token claim) | security | M9-12 | 1.5 | — | (PR review) | — | [ ] |
| M9-15 | Docs M9 | doc-writer | M9-13, M9-14 | 1 | — | `feature/reminders-docs` | — | [ ] |

### REL — Release (MVP deployable)

Full-app QA (REL-08) follows rules/qa-browser.md adapted to mobile: LOCAL DB wiped with `make db-reset`, all data created through the app's screens on emulators, every preset walked.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| REL-01 | IaC staging + prod: Cloud Run (api, worker, migrate job), Cloud SQL PG 18 MX region, Secret Manager, role bootstrap | devops-sr | P0-19 | 8 | — | `release/infra-iac` | OB3 | [ ] |
| REL-02 | CD: images, migration job as cc_owner, auto-deploy staging on merge to dev, prod behind manual approval | devops-sr | REL-01 | 4 | — | `release/cd-pipeline` | OB3 | [ ] |
| REL-03 | Observability + DR: RED metrics, uptime checks, alerts, PITR, restore-drill runbook | devops-sr | REL-01 | 4 | — | `release/observability` | OB3 | [ ] |
| REL-04 | Mobile release pipeline: signing, fastlane, TestFlight + Play internal track, symbols to crash reporting | devops-sr | P0-25 | 5 | — | `release/mobile-pipeline` | OB6, OB7, OB9 | [ ] |
| REL-05 | Preset E2E smokes: demo_core, demo_monetization, demo_full | qa-sr | M1-18, M2-15, M3-14, M4-15, M5-32, M6-09, M7-15, M8-10, M9-15 | 4 | — | `release/preset-smokes` | — | [ ] |
| REL-06 | k6 load test on staging: record CRUD p95 < 300 ms, dashboard p95 < 1 s | qa-sr | REL-02, M6-03 | 3 | — | `release/load-tests` | OB3 | [ ] |
| REL-07 | Store compliance pack: privacy labels / data safety, subscription terms, restore, in-app deletion check, aviso de privacidad link | frontend-sr | M7-09, M1-13 | 3 | — | `release/store-compliance` | OB1, OB10 | [ ] |
| REL-08 | Full-app QA on emulators: local DB wiped, data created through the app screens, every flow per preset | qa-sr | REL-05 | 6 | — | `release/full-qa-fixes` | OB6 | [ ] |
| REL-09 | Security deploy gate: threat model, dependency audit, build-leak check (symbols/maps), secrets review | security | REL-08, REL-02, REL-04 | 4 | — | (PR review) | — | [ ] |
| REL-10 | Tech-lead release review | tech-lead | REL-08 | 2 | — | (PR review) | — | [ ] |
| REL-11 | Docs: final README, deploy/DR/params runbooks, CHANGELOG | doc-writer | REL-09, REL-10, REL-03, REL-06, REL-07 | 3 | — | `release/docs` | — | [ ] |
| REL-12 | Owner approval → release PR dev→main, tag v1.0.0, production deploy, store submission | tech-lead | REL-11, M5-33 | 2 | — | `release/v1.0.0` | OB0, OB5, OB10 | [ ] |

## 3. Parallel dispatch groups

Team rule: dispatch in parallel only when **3+ independent tasks** exist, with **no shared state
and no shared files**. Anything else runs sequentially.

| Group | Tasks | Launch when | Why it is safe |
|---|---|---|---|
| PG-P0-B | P0-04, P0-07 + P0-08 (one branch), P0-13, P0-14 | P0-03 merged | Disjoint dirs: compose/infra, `platform/errors`+`config`, `platform/money`, `platform/crypto` |
| PG-P0-C | P0-06, P0-21, P0-26 | P0-03 and P0-05 merged | Backend codegen vs. `mobile/` spike vs. mock-server tooling |
| PG-M34 | M3-02, M3-03, M4-02, M4-04; then the M3 and M4 backend tracks side by side | M2-03 merged | Separate tables, packages and migration files (timestamped versions) |
| PG-M5-F | M5-08, M5-09, M5-10 | M5-04 merged | RateCatalog, IVA components and registry are separate files |
| PG-M5-S | M5-11, M5-12, M5-13, M5-14, M5-15, M5-16, M5-17, M5-19 (M5-18 follows M5-15) | M5-04, M5-05, M5-06, M5-07, M5-09 merged | One strategy file + one golden folder each; params and registry are not touched (§2 M5) |
| PG-M79 | M7-02, M7-03, M7-04, M9-07 | P0 gates passed | Different modules and adapters |
| Streams | Backend (`backend/`) and mobile (`mobile/`) streams run concurrently for the whole project | Contract of the module frozen | Mobile codes against the Prism mock (P0-26); no shared files |

Always **sequential**: contract freezes (shared `openapi.yaml`), flag registrations (composition
root + presets), Flutter tasks inside one module (shared feature folder), gates after QA.

**Capacity assumption.** Effort is concentrated in backend-sr (215.5 h of 511 h). Baseline is
**2 concurrent backend-sr instances**, scaled to **3 during PG-M5-S**. With one backend-sr the
project roughly doubles in elapsed time (§5).

## 4. Totals

### Per module

| Module | Tasks | Effort (h) | With +20% buffer (h) |
|---|---|---|---|
| P0 Foundation | 32 | 80.5 | 96.6 |
| M1 Auth | 18 | 48.0 | 57.6 |
| M2 Taxpayers | 15 | 41.0 | 49.2 |
| M3 Incomes | 14 | 33.5 | 40.2 |
| M4 Expenses | 15 | 34.5 | 41.4 |
| M5 Tax Engine | 33 | 114.0 | 136.8 |
| M6 Dashboard | 9 | 19.0 | 22.8 |
| M7 Subscriptions | 15 | 39.5 | 47.4 |
| M8 History | 10 | 17.5 | 21.0 |
| M9 Reminders | 15 | 35.5 | 42.6 |
| REL Release | 12 | 48.0 | 57.6 |
| **Total** | **188** | **511.0** | **613.2** |

M5 by group: core 34.0 h · regime strategies 36.5 h (612: 6, 606: 4.5, 626 PF: 3.5, 625: 4.5,
601: 6, 626 PM: 3.5, 621: 3, 622/603: 2.5, informational: 3) · use case/API/Flutter 29.5 h ·
flag/QA/gates/docs/CPA 14.0 h.

### Per owner (effort, h)

backend-sr 215.5 · frontend-sr 89.5 · qa-sr 61.0 · db-sr 41.5 · devops-sr 32.5 ·
security 22.5 · tech-lead 21.0 · doc-writer 15.0 · architect 12.5.

## 5. Milestones (owner notifications) and critical path

Elapsed agent-hours from P0-01, from a dependency + capacity simulation (one instance per role,
backend-sr ×2 baseline; range shows backend-sr ×3). Buffer +20% included. Owner wait time excluded.

| # | Milestone | Done when | Elapsed (h), baseline | Range (h) | Notify owner with |
|---|---|---|---|---|---|
| MS1 | **Foundation done** | P0-29, P0-30 passed and P0-32 merged | **35** | 32–35 | CI links, RLS fail-closed evidence, `make db-up` instructions |
| MS2 | **Backend MVP done** | M1–M7 APIs merged with contract tests green (M1-10, M2-07, M3-06, M4-07, M5-23, M5-24, M6-03, M7-06, M7-07) | **125** | 85–125 | Golden suite report, contract-test report, p95 numbers |
| MS3 | **Mobile base done** | P0-31 review passed, M1-13 and M2-10 merged: the app registers, logs in and onboards a PF/PM taxpayer | **44** | 44 | Emulator recording, mobile CI release-build artifact check |
| MS4 | **First functional version** | `demo_core` modules M1–M6 DONE (both gates) | **148** | 106–148 | `demo_core` walkthrough; estimate vs. golden figures |
| MS5 | **MVP deployable** | REL-11 merged: M7–M9 DONE, staging live, store builds in TestFlight / Play internal, release gates passed | **172** | 133–172 | Release notes, security gate report, request for REL-12 approval |

With a single backend-sr the same milestones land at ≈ 57 / 245 / 44 / 271 / 293 h.

### Critical path

- **Dependency critical path (unlimited agents): 75 h raw → 90 h buffered** to MS5, 77 h to REL-12.
  It runs through the **mobile stream**:
  `P0-01 → P0-05 → P0-21 → P0-22 → P0-23 → P0-24 → M1-11 → M1-12 → M1-13 → M2-08 → M2-09 → M2-10 →
  M3-08 → M3-09 → M4-09 → M4-10 → M4-11 → M4-12 → M4-14 → M4-15 → REL-05 → REL-08 → REL-09 → REL-11 → REL-12`.
  Consequence: **OB4 (Riverpod vs Bloc) and OB6 (Xcode/Android Studio) sit on the critical path
  from hour ~6.**
- **Resource-constrained critical path (real one): backend-sr.** The longest backend chain is
  `P0-03 → P0-04 → P0-09 → P0-10 → P0-11 → P0-12 → P0-16 → P0-28 → P0-30 → M2-05 → M2-06 → M5-21 →
  M5-22 → M6-02 → M6-03` (47.5 h raw), but with 215 h of backend effort the binding constraint is
  backend-sr capacity, not dependencies. M5-22 (estimate use case) is the integration point
  that waits for every strategy plus both LedgerReaders: protect it.
- **Second constraint: qa-sr** (61 h, one instance). If module QA starts queuing behind M5
  goldens, add a second qa-sr instance for M5-07/M5-24.

## 6. Owner-blocking items

Work proceeds around every blocker; the "need by" hour is the baseline elapsed time when the
first blocked task would otherwise start.

| ID | Pending owner decision / action | Blocks | Work-around meanwhile | Need by |
|---|---|---|---|---|
| OB0 | Approve PRD, MODULES, ADR-0001/0002 (incl. the YAML-params and users-in-P0 deviations) and this plan; later, approve the release | P0-01 (everything), REL-12 | None: human checkpoint before execution | Now |
| OB1 | **Pricing** (PRD Q5: price points; weekly at launch or later) | M7-10, REL-07 | Paywall reads offerings from RevenueCat, so prices are never in code; sandbox products with placeholder prices | ≈ h 90 |
| OB2 | **Email verification at registration** (ADR-0001 OWNER 1: A = 202 + code, B = 201 + tokens) and **email provider** (SES / Postmark / Resend) | M1-01 (flow), M1-06 (provider), REL-01 (secret) | Contract is drafted with A; EmailSender port + dev log adapter. If B is chosen: rework M1-01/M1-05/M1-11, about +3 h | Decision ≈ h 2; provider ≈ h 100 |
| OB3 | **Hosting / data residency** (recommended GCP Cloud Run + Cloud SQL `northamerica-south1`) | REL-01, REL-02, REL-03, REL-06 (and REL-09 through REL-02) | Everything runs on the local docker stack; IaC written last | ≈ h 100 for MS5 on time |
| OB4 | **Riverpod vs Bloc** (ADR-0001 O5) | P0-22 and, through it, every Flutter task | Default to Riverpod (ADR recommendation) if no answer when P0-21 ends | **≈ h 6 (critical path)** |
| OB5 | **CPA validation** of the golden suite (PRD Q11) | M5-33, REL-12 | ⚠️ cases are skipped with reasons; dev and staging continue | Before REL-12 |
| OB6 | **Install Xcode + Android Studio** (simulators/emulators) | P0-22 platform build check, every module QA (M1-15, M2-12, M3-11, M4-12, M5-29, M6-06, M7-12, M8-07, M9-12), REL-04, REL-08 | Widget tests + CI macOS/Linux runners; QA waits | ≈ h 30 (first module QA) |
| OB7 | **Apple Developer Program + Google Play Console** accounts (Apple enrollment can take days: start now) | M7-10, M7-12, M9-09 (APNs key), M9-12, REL-04 | Build against fakes; Android-first sandbox testing if Play is ready first | ≈ h 60 |
| OB8 | **RevenueCat account** (project, API keys, webhook secret) | M7-04 (live check), M7-08 (real offerings), M7-10, M7-12 | Fake RC server in tests; paywall against mocked offerings | ≈ h 50 |
| OB9 | **Firebase project** (FCM, APNs key upload, crash reporting) | M9-07 (live check), M9-09, M9-12, REL-04 | Fake FCM sender in tests | ≈ h 65 |
| OB10 | **Lawyer for terms + aviso de privacidad** (PRD Q10; not in the brief's list but it blocks launch) | REL-07, REL-12; real content for `legal_documents` (M1-03 seeds placeholders) | Placeholder documents with versioning | Before REL-07 (≈ h 90) |

## 7. Risks to the estimate

| Risk | Affects | Mitigation |
|---|---|---|
| Transcription error in tariff tables | M5-04, every strategy | Independent double entry (M5-06) + contiguity check at load + boundary goldens |
| swagger_parser cannot handle our `oneOf`/`const` | P0-21 → all Flutter | Fallback already chosen (openapi-generator + 3.0 overlay), about +4 h |
| Library APIs not yet checked in Context7 | P0, M7, M9 | P0-02 / P0-21 are entry criteria; adapters re-check before first use |
| Store sandbox flakiness / review delays | M7-12, REL-04, REL-12 | Start OB7 now; out-of-order webhook tests do not need the store |
| CPA disagrees with a worked example | M5 strategies | Strategy isolation keeps the fix local; re-run goldens; re-estimate |
| backend-sr capacity | MS2, MS4, MS5 | 2–3 instances; PG-M5-S dispatched as one parallel wave |

## 8. Tracking

- Status per task in the tables above (`Done` column); `/status` reads this file.
- At each milestone the eng-manager records estimate vs. actual per module and updates the
  remaining estimates (Clean Coder: re-estimate, don't absorb).
