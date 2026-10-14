# PLAN — Control Contable (v1 MVP)

| Field | Value |
|---|---|
| Status | APPROVED (P0-01). P0 Foundation done (MS1 when `feature/p0-docs` merges); M1 next |
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
| P0-01 | Owner approves PRD, MODULES, ADR-0001/0002, C4, data model and this plan; merge discovery docs to `dev` | tech-lead | — | 0.5 | — | `feature/project-discovery` | OB0 | [x] |
| P0-02 | Context7 check of every pinned Go library API (oapi-codegen, goose, GORM, pgx, testcontainers-go, golang-jwt, libopenapi-validator, firebase/go); notes appended to ADR-0001 | backend-sr | P0-01 | 1.5 | — | `feature/p0-skeleton` | — | [x] |
| P0-03 | Backend skeleton: `go.mod` (go 1.27, toolchain go1.27.2), Clean Architecture layout, Makefile, golangci-lint v2 config with depguard (Dependency Rule) and forbidigo (float/time.Now in domain) | backend-sr | P0-01 | 2 | — | `feature/p0-skeleton` | — | [x] |
| P0-04 | Local stack: docker compose `postgres:18`, role bootstrap (cc_owner, cc_app NOBYPASSRLS, cc_webhook, cc_worker), `make db-up` / `make db-reset` | devops-sr | P0-03 | 2 | PG-P0-B | `feature/p0-local-stack` | — | [x] |
| P0-05 | Contract tooling: OpenAPI lint ruleset (x-module required, single error envelope, examples, security scheme) in CI; freeze platform paths (`/healthz`, `/v1/flags`, `Error`) | architect | P0-01 | 1.5 | — | `feature/p0-contract-lint` | — | [x] |
| P0-06 | oapi-codegen v2 strict std-http generation into `adapter/http/gen`, drift check, generated operationID→flag map, contract test failing on any operation without `x-module` | backend-sr | P0-03, P0-05 | 3 | PG-P0-C | `feature/p0-codegen` | — | [x] |
| P0-07 | `platform/errors`: error envelope, domain-error→HTTP mapping (TDD) | backend-sr | P0-03 | 2 | PG-P0-B | `feature/p0-platform-errors-config` | — | [x] |
| P0-08 | `platform/config` (env + secret source port) and slog JSON handler with PII redaction (RFC, email never logged) (TDD) | backend-sr | P0-03 | 2.5 | PG-P0-B | `feature/p0-platform-errors-config` | — | [x] |
| P0-09 | Integration test harness: testcontainers-go `postgres:18`, migrations applied, role-switching + RLS assertion helpers (A cannot see B, unset context = 0 rows) | qa-sr | P0-04 | 2.5 | — | `feature/p0-test-harness` | — | [x] |
| P0-10 | goose v3 runner `cmd/migrate` (embed.FS, timestamped versions) + baseline migration: `app_user_id()`, grants, `feature_flags`, `flag_presets` (global) | db-sr | P0-04, P0-09 | 2.5 | — | `feature/p0-migrations-base` | — | [x] |
| P0-11 | Migration `users` (identity root, moved from M1 so P0 FKs resolve) + `entitlements` + `user_usage`, ENABLE+FORCE RLS, isolation tests | db-sr | P0-10 | 3.5 | — | `feature/p0-migrations-identity` | — | [x] |
| P0-12 | TxManager: `set_config('app.current_user_id', $1, true)` per tx, RO/RW variants, fail-closed integration tests | backend-sr | P0-11 | 3 | — | `feature/p0-txmanager` | — | [x] |
| P0-13 | `platform/money` (Cents, Rate 10^-6, Exact, 128-bit multiply, half-away-from-zero) + Clock port (America/Mexico_City), 100% coverage | backend-sr | P0-03 | 4 | PG-P0-B | `feature/p0-money-clock` | — | [x] |
| P0-14 | `platform/crypto`: argon2id PHC, AES-256-GCM with key_id (RFC), HMAC-SHA256 peppers (rfc_hash, email_hash) (TDD) | backend-sr | P0-03 | 3 | PG-P0-B | `feature/p0-crypto` | — | [x] |
| P0-15 | Flags service + gate middleware (OFF → 404 before handler) + `GET /v1/flags` + flag ON/OFF test helper | backend-sr | P0-06, P0-07, P0-10 | 3.5 | — | `feature/p0-flags` | — | [x] |
| P0-16 | AccessPolicy port + infra: trial clock, atomic `ReserveRecord` (UPDATE … RETURNING), trial/pro/comp lookup, read-only after trial (TDD + integration) | backend-sr | P0-12, P0-13 | 4 | — | `feature/p0-access-policy` | — | [x] |
| P0-17 | Composition root + middleware chain (request-id, recover, token-bucket rate limit, authn slot, flag gate, kin-openapi request validation), `/healthz` liveness/readiness | backend-sr | P0-06, P0-07, P0-08 | 3 | — | `feature/p0-http-server` | — | [x] |
| P0-18 | Contract-test harness (libopenapi-validator) wired to integration tests; applied to `/healthz` and `/v1/flags` | qa-sr | P0-15, P0-17 | 2 | — | `feature/p0-contract-tests` | — | [x] |
| P0-19 | Backend CI: gofmt, golangci-lint, unit, integration (testcontainers), contract, govulncheck, codegen drift, OpenAPI lint, coverage ≥ 80% gate | devops-sr | P0-06, P0-09 | 3 | — | `feature/p0-ci-backend` | — | [x] |
| P0-20 | Flag admin CLI `cmd/flags` (list, enable, disable, apply-preset) running as cc_owner (backs the /flags skill) | backend-sr | P0-15 | 2 | — | `feature/p0-flags` | — | [x] |
| P0-21 | Context7 check of Dart libs + swagger_parser spike on `openapi.yaml` (oneOf/const); record go/fallback decision (openapi-generator + 3.0 overlay) | frontend-sr | P0-01, P0-05 | 4.5 | PG-P0-C | `feature/p0-dart-client-spike` | — | [x] |
| P0-22 | Flutter scaffold `mobile/`: flavors dev/staging/prod, go_router, state management, dio, design tokens, es-MX l10n, analysis_options | frontend-sr | P0-21 | 4 | — | `feature/p0-mobile-scaffold` | OB4, OB6 | [x] |
| P0-23 | Generated Dart client wiring + `make mobile-client` + drift check. ADR-0003 conditions: (a) `mobile/tool/normalize_openapi.dart` inlines path-level `parameters` into each operation, unit test first, and is the only input `make mobile-client` feeds swagger_parser; (b) CI guard `mobile/tool/check_generated_kinds.dart` runs inside `make mobile-client` and fails it when a generated Dart type's kind (enum/class/typedef) differs from its component schema or two schemas map to one Dart name; its test first feeds a fixture with an inline `oneOf` object holding a `period` enum and must fail on the `Period` collision; (c) file the path-parameter cast bug upstream (carapacik/swagger_parser) and link it in ADR-0003; drop (a) once a fixed release is pinned | frontend-sr | P0-22 | 2 | — | `feature/p0-mobile-client-flags` | — | [x] |
| P0-24 | FlagsProvider + `useFlag` helper fed by `GET /v1/flags` (widget tests first) | frontend-sr | P0-23 | 2.5 | — | `feature/p0-mobile-client-flags` | — | [x] |
| P0-25 | Mobile CI: `make mobile-client` (normalizer + ADR-0003 kind guard + drift check), analyze, test, release build `--obfuscate --split-debug-info`, artifact gate (no symbols / maps bundled) | devops-sr | P0-22, P0-23 | 3 | — | `feature/p0-ci-mobile` | — | [x] |
| P0-26 | Prism mock server from `openapi.yaml` (`make mock`) so mobile work starts before backend | devops-sr | P0-05 | 1 | PG-P0-C | `feature/p0-mock-server` | — | [x] |
| P0-27 | Worker skeleton `cmd/worker`: ticker, `pg_try_advisory_lock` leader election, job registry, graceful shutdown (errgroup) | backend-sr | P0-12, P0-08 | 2.5 | — | `feature/p0-worker` | — | [x] |
| P0-28 | QA P0: RLS fail-closed suite, flag ON/OFF on platform routes, CI green on both pipelines | qa-sr | P0-16, P0-18, P0-19, P0-20, P0-27 | 2 | — | `feature/p0-qa-fixes` | — | [x] |
| P0-29 | Tech-lead gate P0 backend platform (canon checklist, Dependency Rule, money package review) | tech-lead | P0-28 | 2 | — | (PR review) | — | [x] |
| P0-30 | Security gate P0 backend platform (roles, NOBYPASSRLS, crypto params, log redaction, CI scanners) | security | P0-28 | 2.5 | — | (PR review) | — | [x] |
| P0-31 | Tech-lead review of mobile foundation (scaffold, client generation, FlagsProvider, mobile CI) | tech-lead | P0-24, P0-25 | 1.5 | — | (PR review) | — | [x] |
| P0-32 | Docs P0: README (prereqs, setup, flags), local runbook, platform package docs | doc-writer | P0-29, P0-30 | 2 | — | `feature/p0-docs` | — | [x] |

#### P0 follow-ups

Deferred by the P0 gates, none blocking. Each lands with the task in its target column. Sources:
[P0-29 tech-lead gate](../reviews/P0-29-tech-lead-gate.md) (FU*) and
[P0-30 security gate](../reviews/P0-30-security-gate.md) (D-*).
Mobile and QA deferrals keep their own lists: [P0-31 mobile review](../reviews/P0-31-mobile-foundation.md)
(F-04..F-11) and [P0-28 QA](../reviews/P0-28-qa-foundation.md) (F3, F4).

| ID | Item | Target |
|---|---|---|
| FU1 | Request-deadline middleware (about 10 s, under `WriteTimeout`); test with a hanging `tx.Manager` | M1, before the first DB-backed handler |
| FU2 | Replace `cmd/api` `notBuilt` (nil server, 500 on a flag ON) with an explicit stub | M1, before the first preset demo |
| FU3 | Move `internal/platform/httperr` to `adapter/http/httperr` | M1 |
| FU4 | ADR: `platform/access` stays a documented shared-kernel port or becomes `internal/access` | M1, before M1-07 wires it |
| FU5 | Usecase third-party imports: denylist → allowlist in `archtest` | M1 |
| FU6 | `flags.Service` clears `failed` after a successful load | Any (backend) |
| FU7 | Use the generated `gen.ErrorResponse` at the edge instead of the hand-written envelope types | M1, with FU3 |
| FU8 | `FuzzParseCents` / `FuzzParseRate` fuzz targets for `money` | Before M3 (tax engine) |
| D-2 | Split `tx.Manager` from `tx.SystemManager`; archtest forbids `SystemManager` in `adapter/http` except sign-up | M1-05 |
| D-3 | Lint/archtest rule: repositories never `Commit`/`Rollback`/`SavePoint` or run `SET`/`set_config` | M1, first repository PR |
| D-4 | Taxpayer row id as GCM AAD for the RFC ciphertext | M2-05 (blocks its security gate) |
| D-5 | NFC normalisation before `email_hash`, if EAI addresses are accepted | M1-05 decision |
| D-6 | `Cache-Control: no-store` on error responses, contract first | Next contract revision (M1-01) |
| D-7 | Required code-owner review | When a second account or GitHub App exists |
| D-8 | npm dev-tooling alerts in `tools/contract` (Prism transitive deps) | Next Dependabot cycle |
| D-9 | Re-apply required checks `backend-ok`, `mobile-ok`, `contract-ok` on `dev` and `main` | Done: verified on both branches on 2026-10-13 |

### M1 — Auth & Account (`module_auth`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M1-01 | Contract freeze: `/v1/auth/*`, `/v1/me`, `/v1/legal/documents` (verification flow per owner decision), lint | architect | P0-05 | 1.5 | — | `feature/auth-contract` | OB2 | [x] |
| M1-02 | Migration `refresh_tokens` + `consents` with RLS + isolation tests | db-sr | P0-11 | 3 | — | `feature/auth-migrations` | — | [x] |
| M1-03 | Migration `legal_documents` (global seed) + `trial_registry` and `email_verifications` (service-only) + definer fns `auth_lookup_credentials`, `trial_registry_check_and_insert`, `email_verification_upsert` / `_consume` + tests | db-sr | P0-11 | 3 | — | `feature/auth-migrations` | — | [x] |
| M1-04 | Auth domain: email, password policy, refresh token VO `<uuid>.<secret>` (TDD) | backend-sr | P0-03 | 2 | — | `feature/auth-domain` | — | [x] |
| M1-05 | Register + verify-email use cases: EmailSender port (dev log adapter), generic duplicate response, consents, trial start via TrialStarter (TDD) | backend-sr | M1-01, M1-02, M1-03, M1-04, P0-29, P0-30 | 4.5 | — | `feature/auth-register` | — | [x] |
| M1-06 | Email provider adapter (SES / Postmark / Resend) behind EmailSender | backend-sr | M1-05 | 2 | — | `feature/auth-email-provider` | OB2 | blocked (OB2) |
| M1-07 | Login + EdDSA JWT (kid rotation) + authn middleware (TDD) | backend-sr | M1-05, P0-17 | 3.5 | — | `feature/auth-session` | — | [x] |
| M1-08 | Refresh rotation, reuse detection (family revoke), logout (TDD + integration) | backend-sr | M1-07 | 3.5 | — | `feature/auth-session` | — | [x] |
| M1-09 | `GET/PATCH/DELETE /v1/me` (cascade delete; trial_registry survives) + `GET /v1/legal/documents` | backend-sr | M1-07 | 3 | — | `feature/auth-account` | — | [x] |
| M1-10 | Auth rate limits + contract tests for every M1 operation + flag ON/OFF suite | qa-sr | M1-08, M1-09, P0-18 | 3 | — | `feature/auth-contract-tests` | — | [x] |
| M1-11 | Flutter register + verify-email screens with consents (widget tests first, mock server) | frontend-sr | P0-24, P0-26, M1-01 | 4 | — | `feature/auth-mobile-register` | — | [x] |
| M1-12 | Flutter login screen, secure token storage, refresh interceptor, logout | frontend-sr | M1-11 | 4 | — | `feature/auth-mobile-session` | — | [x] |
| M1-13 | Flutter profile + in-app account deletion | frontend-sr | M1-12 | 2.5 | — | `feature/auth-mobile-account` | — | [x] |
| M1-14 | Register flag `module_auth` (migration + preset membership) and client gating | backend-sr | M1-10, M1-13 | 0.5 | — | `feature/auth-flag` | — | [x] |
| M1-15 | QA M1 on emulators: register → verify → login → refresh → delete; flag OFF = 404 | qa-sr | M1-14 | 3 | — | `feature/auth-qa-fixes` | OB6 | [x] |
| M1-16 | Tech-lead gate M1 | tech-lead | M1-15 | 1.5 | — | (PR review) | — | [x] |
| M1-17 | Security gate M1 (enumeration, token reuse, argon2 params, rate limits) | security | M1-15 | 2.5 | — | (PR review) | — | [x] |
| M1-18 | Docs M1 (module doc, auth runbook, generated API ref) | doc-writer | M1-16, M1-17 | 1 | — | `feature/auth-docs` | — | [x] |

M1-06 stays open: the email provider (SES / Postmark / Resend) is the owner's choice (OB2).
Until then the dev log adapter sends nothing, and `module_auth` cannot be enabled in a shared
environment ([auth runbook §6](../runbooks/auth.md#6-enable-module_auth-in-a-shared-environment), P-1).

#### M1 follow-ups

Deferred by the M1 gates; none blocks M1's merge. Each lands with the task or module in its
target column. Sources: [M1-16 tech-lead gate](../reviews/M1-16-tech-lead-gate.md) (FU*) and
[M1-17 security gate](../reviews/M1-17-security-gate.md) §3–§4 (G-*, PR numbers). Items
marked **pre-enable** must land before `module_auth` is ON in a shared environment.

| ID | Item | Owner | Target |
|---|---|---|---|
| FU1 | **Pre-enable.** Mobile `guardApi` also maps `TypeError` / `CheckedFromJsonException` to `ApiUnexpected` (test: `access_token: null`) | frontend-sr | M1 hardening |
| FU2 | **Pre-enable.** `Session.readRefreshToken` returns `null` on a store exception; add `sign_out_test.dart` | frontend-sr | M1 hardening |
| FU3 | One error→HTTP mapper in `authhttp` (`ErrWrongPassword`, `ErrAccountNotFound`, dead `ErrorMapper` password rules) | backend-sr | Next touch of `authhttp` |
| FU4 | One `FailureNotice` controller + `announce` helper in `core/forms`; `SignIn` / `CompleteSignUp` use cases | frontend-sr | Before M2 forms (M2-08) |
| FU5 | Clear the "Espera N min" notice when the wait ends; countdown on verify/register `429`; round `formatCountdown` up | frontend-sr | With FU4 |
| FU6 | ADR on `access` package ownership (P0-29 FU4, overdue) | architect | Before M2 |
| FU7 | Move `platform/httperr` to `adapter/http/httperr`; dedupe the 429 and invalid-input messages (P0-29 FU3) | backend-sr | M1 refactor |
| FU8 | Inject one `clock.Clock` in `cmd/api/auth.go` | backend-sr | M1 refactor |
| FU9 | FORCE RLS (with definer policy) on `trial_registry`, `email_verifications`, `email_verification_failures`, `legal_documents`, or reword MODULE-auth §6.3; fix the stale `identity.sql` comment | db-sr | M1 refactor |
| FU10 | archtest: denylist → allowlist (P0-29 FU5) | tech-lead | M2 |
| FU11 | Measure secure-store restore latency on a low-end release build; decide Android `resetOnError` | frontend-sr + security | M1 hardening |
| #40, G-2 | Async email (outbox): provider latency and failures never reach the response | backend-sr | M1-06 (pre-enable, P-1) |
| #34 | Create `cc_auth_definer` in every shared environment before the M1 migrations | devops-sr | REL-09 (pre-enable, P-2) |
| #34, OB10 | Real legal documents replace the `.invalid` placeholders | owner + legal | Before enable (P-5) |
| #40, OB10 | Consents are those of the first registrant, not re-affirmed at verify | backend-sr + legal | With OB10 |
| #47 FU1, G-1 | `rate_limit_take` returns a wait ≤ 0 under a concurrent reset (`greatest(…, 1 µs)`) | db-sr | M9 |
| #47 FU2 | `rate_limit_take` returns NULL if the purge deletes the row mid-call | db-sr | M9 |
| #33, #44 | Purge dead refresh-token families + `(absolute_expires_at)` index | db-sr | M9 |
| #34 | Purge `email_verification_failures` and `rate_limit_buckets` past their windows | db-sr | M9 |
| #41 | Suspension revokes refresh families; operator suspend path (definer function) | db-sr + backend-sr | Before any suspend feature |
| #41 #3 | Router authenticates each operation with its declared security scheme (webhook) | backend-sr | Before M7-05 |
| P0-30 D-6, G-3 | `Cache-Control: no-store` on error responses | backend-sr | Next platform touch |
| P0-30 D-3 | Lint/archtest: repositories never Commit/Rollback/SET | tech-lead | M2 |
| P0-31 F-08/D-8 | osv-scanner for pub deps in CI; Prism dev-tool advisories | devops-sr | REL-09 |
| #43 | Deletion-cascade guard misses indirect chains | db-sr | When M2 tables land |
| — | Raise argon2id memory (e.g. 46 MiB, t = 1) once the Cloud Run size is fixed | backend-sr | REL-09 sizing |
| M1-18 | Support can't trace a failed login: show `request_id` in the app's error notice; operator email → user-id lookup without exposing PII ([auth runbook §4](../runbooks/auth.md#4-investigate-a-user-cant-log-in-without-pii)) | frontend-sr + backend-sr | M1 hardening |

### M2 — Taxpayers & Onboarding (`module_taxpayers`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M2-01 | Contract freeze: `/v1/regimes`, `/v1/taxpayers/**`, lint | architect | P0-05 | 1.5 | — | `feature/taxpayers-contract` | — | [x] |
| M2-02 | Migration `regimes` + `regime_compatibility` (global) seeded from research A.1/B.1 (option_defs, required_inputs) | db-sr | P0-10 | 3 | — | `feature/taxpayers-catalog-migration` | — | [x] |
| M2-03 | Migration `taxpayers` + `taxpayer_regimes` with RLS, composite FKs, immutable person_type grant, isolation tests | db-sr | P0-11, M2-02 | 3.5 | — | `feature/taxpayers-migration` | — | [x] |
| M2-04 | Taxpayer domain: RFC VO (PF 13 / PM 12, check digit), PersonType, regime option validation (table-driven TDD) | backend-sr | P0-03 | 3 | — | `feature/taxpayers-domain` | — | [x] |
| M2-05 | Taxpayer use cases: create (encrypt + hash RFC, soft cap 10), list, get, patch, archive, delete | backend-sr | M2-01, M2-03, M2-04, P0-29, P0-30 | 4 | — | `feature/taxpayers-usecases` | — | [x] |
| M2-06 | Regime assignment use cases (compatibility, options, end period) + TaxpayerRegimeChecker and TaxpayerProfileReader adapters | backend-sr | M2-05 | 4 | — | `feature/taxpayers-usecases` | — | [x] |
| M2-07 | M2 handlers incl. `GET /v1/regimes`, contract tests, flag ON/OFF | backend-sr | M2-06, M1-07 | 3 | — | `feature/taxpayers-http` | — | [x] |
| M2-08 | Flutter onboarding steps 1–2: persona física/moral first, then RFC | frontend-sr | M1-13, M2-01 | 3 | — | `feature/taxpayers-mobile-onboarding` | — | [x] |
| M2-09 | Flutter onboarding steps 3–4: regimes filtered by person type + per-regime options | frontend-sr | M2-08 | 4 | — | `feature/taxpayers-mobile-onboarding` | — | [x] |
| M2-10 | Flutter active-taxpayer switcher + taxpayer settings (edit, add/archive regime, archive) | frontend-sr | M2-09 | 4 | — | `feature/taxpayers-mobile-settings` | — | [x] |
| M2-11 | Register flag `module_taxpayers` | backend-sr | M2-07, M2-10 | 0.5 | — | `feature/taxpayers-flag` | — | [x] |
| M2-12 | QA M2 (onboarding PF and PM, duplicate RFC, cross-user 404, flag OFF) | qa-sr | M2-11 | 3 | — | `feature/taxpayers-qa-fixes` | OB6 | [x] |
| M2-13 | Tech-lead gate M2 | tech-lead | M2-12 | 1.5 | — | (PR review) | — | [x] |
| M2-14 | Security gate M2 (RFC encryption/HMAC, composite-FK isolation) | security | M2-12 | 2 | — | (PR review) | — | [x] |
| M2-15 | Docs M2 | doc-writer | M2-13, M2-14 | 1 | — | `feature/taxpayers-docs` | — | [x] |

#### M2 follow-ups

Deferred by the M2 gates; none blocks M2's merge. Each lands with the task or module in its
target column. Sources: [M2-13 tech-lead gate](../reviews/M2-13-tech-lead-gate.md) (FU*) and
[M2-14 security gate](../reviews/M2-14-security-gate.md) §3–§4 (G-*, PR numbers). Items
marked **pre-enable** must land before `module_taxpayers` is ON in a shared environment
([taxpayers runbook §5](../runbooks/taxpayers.md#5-enable-module_taxpayers-in-a-shared-environment)).

| ID | Item | Owner | Target |
|---|---|---|---|
| FU1 | **Pre-enable.** Spec ↔ seed drift in the compatibility matrix (621 × 612, 626 × 621, PM pairs): amend [spec §2.5](MODULE-taxpayers.md#25-compatibility-data-not-code) / OQ-6, or forward migration setting `allowed = false` | pm + db-sr | Before enable |
| FU2 | Duplicate RFC of an archived taxpayer: message should suggest unarchiving (US-T2), or amend the spec | pm / backend-sr | M2 hardening |
| FU3 | `unawaited(adopt/select/clear)` store exceptions escape; `activeTaxpayerProvider` must fall back on a store failure (§2.9) | frontend-sr | M2 hardening |
| FU4 | Block `OnboardingFlow.back()` and step-5 *Omitir* while the create is in flight | frontend-sr | M2 hardening |
| FU5 | `SemanticsService.sendAnnouncement` boilerplate (M1-16 FU4 carried); `takeAnnouncements` assertions for paywall and settings errors | frontend-sr | Before M3 forms |
| FU6 | `inline_text_editor.dart`: return focus to *Editar* after Save/Cancel | frontend-sr | M2 hardening |
| FU7 | Split `onboarding_flow.dart`: shared `RegimeDraft` with `AddRegimeFlow`, post-create flow | frontend-sr | Refactor before M3 |
| FU8, G-4 | Key `active_taxpayer_id` by user id (or amend spec §2.9) | frontend-sr | M3-08 |
| FU9 | Backend nits: `toOptions` drops the `setOption` error; `errors.Join` vs `%w` in `usecase/regimes.go`; spec §2.8/S-5 say `FOR UPDATE` (code uses `FOR NO KEY UPDATE`); one catalog query per regime; map `taxpayers_rfc_masked_format` | backend-sr | M2 refactor |
| FU10 | Regimes step uses `review_pending_with` to warn before submit | frontend-sr | M2 hardening |
| FU11 | PR #68 leftovers: probe comment, preset state, one `gen` helper for module-owned operations | qa-sr | Next touch |
| FU12 | Re-verify M1-16 FU3, FU7, FU10 | backend-sr / tech-lead | As tracked in M1 follow-ups |
| G-1 | ~~Replace `infra.NoRecords{}` with the M3/M4 latest-record lookup before any record write path is enabled (P-5)~~ **Done** (`cmd/api/records.go`; verified by the [M3/M4 security gate](../reviews/M3-M4-security-gate.md)) | backend-sr | M3-04 / M4-05 |
| G-2, #59 | **RFC key-rotation re-seal job** (previous keyring entry, resumable per-row re-seal, pepper re-hash with a hash-version column). Until then no rotation of `RFC_ENC_KEY` / `RFC_INDEX_PEPPER` ([runbook §2](../runbooks/taxpayers.md#2-do-not-rotate-rfc_enc_key-until-the-m9-re-seal-job-exists), P-4) | backend-sr + db-sr | M9 |
| G-3 | `RFCVault.Seal` stores the key id parsed from the ciphertext (`crypto.KeyIDOf`) | backend-sr | With the M9 re-seal job |
| #59 nit | Contract note: an idempotent create retry ignores a different body | architect | Next contract touch |
| #61 FU | Undo of archiving the active taxpayer restores it as active; reload the list on a taxpayer-scoped 404 | frontend-sr | M3-08 |
| #58 | `checkRate` names a missing bound as `0`/`999`; "guardado" in the review title | frontend-sr | Next mobile touch |
| #68 D1–D3 | UX polish backlog (cap step, read-only Undo, CPA copy): fixed in M2-13; keep the backlog item for review | pm + frontend-sr | M2 UX polish |
| #55, #56 | Regime-level `needs_cpa_review` exposure; `first_fiscal_year`; 625 per-option treatment | architect + backend-sr | M5-19 |
| OQ-6, OQ-7 | CPA review of the compatibility matrix and of mid-year option changes ([CPA-REVIEW.md](../reviews/CPA-REVIEW.md)) | owner + CPA | Before launch |
| M1-17 G-3, P0-30 D-6 | `Cache-Control: no-store` on error responses | backend-sr | Next platform touch |

### M3 — Incomes (`module_incomes`)

M3 and M4 run side by side (different tables, packages and screens). Their Flutter tasks are sequential because M4 reuses M3's list/form widgets.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M3-01 | Contract freeze: incomes paths, lint | architect | P0-05 | 1 | — | `feature/incomes-contract` | — | [x] |
| M3-02 | Migration `incomes` with RLS, composite FK to taxpayer_regimes, partial indexes, isolation tests, EXPLAIN ANALYZE | db-sr | M2-03 | 3 | PG-M34 | `feature/incomes-migration` | — | [x] |
| M3-03 | Income domain: amounts, dates (cash vs accrual), withholdings (TDD) | backend-sr | P0-13 | 2.5 | PG-M34 | `feature/incomes-domain` | — | [x] |
| M3-04 | Income create/update use cases: ReserveRecord (402), regime check, RateCatalog date validity, idempotent client id | backend-sr | M3-01, M3-02, M3-03, M2-06, P0-16, M5-08 | 4 | — | `feature/incomes-usecases` | — | [x] |
| M3-05 | Income list (keyset, period + regime filter), soft delete, restore | backend-sr | M3-04 | 3 | — | `feature/incomes-usecases` | — | [x] |
| M3-06 | Income handlers, contract tests, flag ON/OFF, trial-cap 402 test | backend-sr | M3-05, M2-07 | 2.5 | — | `feature/incomes-http` | — | [x] |
| M3-07 | LedgerReader income aggregates adapter (GROUP BY month, regime) + EXPLAIN ANALYZE | backend-sr | M3-02, M5-02 | 2.5 | — | `feature/incomes-ledger-reader` | — | [x] |
| M3-08 | Flutter income list: month navigation, regime filter, delete + persistent Undo | frontend-sr | M2-10, M3-01 | 4 | — | `feature/incomes-mobile` | — | [x] |
| M3-09 | Flutter income form as its own page (>6 fields), defaults prefilled, save-and-create-another | frontend-sr | M3-08 | 4.5 | — | `feature/incomes-mobile` | — | [x] |
| M3-10 | Register flag `module_incomes` | backend-sr | M3-06, M3-07, M3-09 | 0.5 | — | `feature/incomes-flag` | — | [x] |
| M3-11 | QA M3 (CRUD, undo, trial cap, flag OFF) | qa-sr | M3-10 | 2.5 | — | `feature/incomes-qa-fixes` | OB6 | [x] |
| M3-12 | Tech-lead gate M3 | tech-lead | M3-11 | 1 | — | (PR review) | — | [x] |
| M3-13 | Security gate M3 | security | M3-11 | 1.5 | — | `feature/incomes-expenses-security-gate` | — | [x] |
| M3-14 | Docs M3 | doc-writer | M3-12, M3-13 | 1 | — | `feature/incomes-docs` | — | [x] |

### M4 — Expenses (`module_expenses`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M4-01 | Contract freeze: expenses + `/v1/expense-categories`, lint | architect | P0-05 | 1 | — | `feature/expenses-contract` | — | [x] |
| M4-02 | Migration `expense_categories` + `category_regime_rules` (global seed with legal basis) | db-sr | M2-02 | 2.5 | PG-M34 | `feature/expenses-migrations` | — | [x] |
| M4-03 | Migration `expenses` with RLS, composite FK, partial indexes, isolation tests | db-sr | M2-03, M4-02 | 3 | — | `feature/expenses-migrations` | — | [x] |
| M4-04 | Expense domain incl. cash-payment threshold warning (TDD) | backend-sr | P0-13 | 2.5 | PG-M34 | `feature/expenses-domain` | — | [x] |
| M4-05 | Expense create/update use cases (ReserveRecord, regime check, RateCatalog, category rules) | backend-sr | M4-01, M4-03, M4-04, M2-06, P0-16, M5-08 | 3.5 | — | `feature/expenses-usecases` | — | [x] |
| M4-06 | Expense list, soft delete, restore | backend-sr | M4-05 | 2.5 | — | `feature/expenses-usecases` | — | [x] |
| M4-07 | Expense handlers + categories endpoint, contract tests, flag ON/OFF | backend-sr | M4-06, M2-07 | 3 | — | `feature/expenses-http` | — | [x] |
| M4-08 | LedgerReader expense aggregates (by deductibility class, IVA paid, withholdings made) + EXPLAIN ANALYZE | backend-sr | M4-03, M5-02 | 2.5 | — | `feature/expenses-ledger-reader` | — | [x] |
| M4-09 | Flutter expense list (reuses M3 list widgets) | frontend-sr | M3-09, M4-01 | 3 | — | `feature/expenses-mobile` | — | [x] |
| M4-10 | Flutter expense form page: category with deductibility hint, payment method, cash warning | frontend-sr | M4-09 | 4.5 | — | `feature/expenses-mobile` | — | [x] |
| M4-11 | Register flag `module_expenses` | backend-sr | M4-07, M4-08, M4-10 | 0.5 | — | `feature/expenses-flag` | — | [x] |
| M4-12 | QA M4 | qa-sr | M4-11 | 2.5 | — | `feature/expenses-qa-fixes` | OB6 | [x] |
| M4-13 | Tech-lead gate M4 | tech-lead | M4-12 | 1 | — | (PR review) | — | [x] |
| M4-14 | Security gate M4 | security | M4-12 | 1.5 | — | `feature/incomes-expenses-security-gate` | — | [x] |
| M4-15 | Docs M4 | doc-writer | M4-13, M4-14 | 1 | — | `feature/expenses-docs` | — | [x] |

#### M3/M4 follow-ups

Deferred by the M3/M4 gates; none blocks M3 or M4. Each lands with the task or module in its
target column. Sources: [M3-M4 security gate](../reviews/M3-M4-security-gate.md) §3–§4 (G-*,
PR numbers), [M3-M4 tech-lead gate](../reviews/M3-M4-tech-lead-gate.md) (FU*, QA decisions) and
[M3-M4 QA](../reviews/M3-M4-QA.md) §4 (D-*). Items marked **pre-enable** must hold before
`module_incomes` / `module_expenses` are ON in a shared environment
([runbook §5](../runbooks/incomes-expenses.md#5-enable-module_incomes-and-module_expenses-in-a-shared-environment)).

| ID | Item | Owner | Target |
|---|---|---|---|
| #97 FU-2 | ~~Contract text: dates before 2026 ⇒ `400 out_of_range`. `createIncome`, `updateIncome`, `getIncomeDefaults` (`on`), `createExpense`, `updateExpense` and `listExpenseCategories` (`on`) say "before 2022-01-01"; the server refuses every date before the first parameter set (2026-01-01), and a date outside the regime window answers `422 regime_not_active` first. Also fix the served message ("anterior a 2022") and `fiscal.MinRecordDate`'s doc. Documented in [incomes](../modules/incomes.md#date-limits)~~ **Done** (contract 0.7.0: `BusinessDate`, `createIncome`, `getIncomeDefaults`, `createExpense`, `OnDate`; message "nor before 2026"; the difference with the estimates' `422 parameters_unavailable` is explained in `BusinessDate`) | architect (+ backend-sr for the message) | Next contract re-freeze |
| G-1, M1-17 G-3, P0-30 D-6, S-11 | ~~`Cache-Control: no-store` on error responses: declare it on the shared error responses and set it in `platform/httperr.Write` (record bodies already have it; the M3/M4 errors declare it optional)~~ **Done** (contract 0.7.0: every response declares the required `NoStore`; `adapter/http/middleware.NoStore` sends it on every response and forces it on every error; `NoStoreOnError` removed) | architect + backend-sr | Next contract re-freeze (with #97 FU-2) |
| FU1, FU2 | Incomes/expenses duplication (root cause of tech-lead H1/H2): extract the record shell to `core/records` (mobile) and `internal/kernel/records` (backend: cursor codec, `canonicalID`, description normalization, audit, error tables) | tech-lead + frontend-sr / backend-sr | `/refactor` before M6 |
| FU3 | Income *Guardar y crear otro* keeps the withholdings section open; expenses resets it | frontend-sr | With FU1 |
| G-2, #83 FU | `FOR SHARE` on the taxpayer row in `Delete` (incomes and expenses) | backend-sr | Next records touch |
| #83 FU | Remove the PK-collision retry loop once the advisory lock is proven alone | backend-sr | Next records touch |
| #93 FU-2 | Inject the clock into `access.NewPolicy` in `newExpensesModule` | backend-sr | Next records touch |
| #67 nit | `fiscal.MinRecordDate` is an exported mutable `var` | backend-sr | Next kernel touch |
| #64 FU2 | Immutability trigger on `user_id` / `taxpayer_id` of records (grants already block `cc_app`) | db-sr | M9 hardening |
| #91 note 2 | New constraints on populated tables use `NOT VALID` + `VALIDATE` | db-sr | Standing rule |
| #66 F5, F6 | `PeriodTotals` bound; es-MX vs English example messages | architect | Next contract re-freeze |
| #85 FU | Form date labels keep the first regime's basis when a PM switches 601 ↔ 626 | frontend-sr | M3/M4 UX polish |
| #99 FU 1–4 | Failed delete hidden behind a pending Undo; Prism mock suites not in CI; "5 gastos" hard-coded; withholding suggestions not prefilled | frontend-sr (1, 3, 4), devops-sr (2) | M3/M4 UX polish; CI backlog |
| D-1 | IVA note for cash over the limit shown before the override, while the server warns only with it: show it only with the override, or have MODULE-expenses §2.5 always return the warning | pm + frontend-sr | M3/M4 UX polish |
| D-5 | 8 % border rate offered with the regime's border option off | pm | M3/M4 UX polish |
| D-6 | Regime filter reset on month change; empty-month copy when other months have records | pm + frontend-sr | M3/M4 UX polish |
| D-7 | 612 filter chip clipped: needs a short regime name in the catalog/contract | architect + pm | Contract 0.7 |
| D-9 | Onboarding starts every regime this month, so earlier records answer `422 regime_not_active`: ask the start month (or default to January) | pm | M2 follow-up |
| P-8, #64 FU1, #65, #67, #91 note 1 | **Pre-enable (product).** CPA review: income basis of 620/622/624, booking dates, category catalog, cash rule and LIVA 5 creditable IVA, override on cash over the limit ([CPA-REVIEW.md](../reviews/CPA-REVIEW.md)) | owner + CPA | Before launch |

### M5 — Tax Engine (`module_tax_engine`)

Task groups:
- **Core** (M5-01…M5-10): contract, pure domain, params loader + 2026 data, tariff lookup, golden harness, RateCatalog, shared IVA components, registry.
- **One group per regime strategy** (M5-11…M5-19): 612 PF, 606 PF, 626 PF, 625 PF, 601 PM, 626 PM, partial 621 PF, partial 622/603 PM, explanation-only catalog. Each strategy is one file + its golden cases under `testdata/golden/<regime>-<pf|pm>/`; **none edits the registry or the params file** (params are transcribed once in M5-04, the registry is wired in M5-22), which is what makes them parallel-safe.
- **Golden cases** come from research §A.5 and §B.5 (PF-612-ISR-01..03, PF-612-IVA-01, PF-626-ISR-01..02, PF-626-IVA-01, PF-606-ISR-01..02, PF-606-IVA-01, PF-625-01..02, PF-621-01, PM-601-ISR-01..02, PM-601-IVA-01, PM-626-ISR-01) plus generated tariff-row and validity-date boundaries. ⚠️ cases ship as `skip` with reason until CPA sign-off (M5-33).
- **Use case + API + Flutter** (M5-20…M5-27), then flag, QA, gates, docs.

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M5-01 | Contract freeze: estimates, breakdown, tax-inputs, Disclaimer codes, `legend` const, lint | architect | P0-05 | 2 | — | `feature/tax-engine-contract` | — | [x] |
| M5-02 | Engine domain: Strategy port, StrategyKey (regime, person type), Input/MonthSums, RegimeEstimate, Line, Ledger trace helper, Citation; consumer ports LedgerReader/TaxInputReader (TDD) | backend-sr | P0-13 | 4 | — | `feature/tax-engine-domain` | — | [x] |
| M5-03 | Params loader: YAML schema, go:embed, `Params.At(key, date)`, overlap rejection, typed fail-fast validation (contiguity, 12 tables, rates in [0,1]), version selection + provisional fallback (TDD) | backend-sr | M5-02 | 5 | — | `feature/tax-engine-params` | — | [x] |
| M5-04 | Transcribe `params/2026/v1.yaml` verbatim: art. 96, twelve art. 106 tables, RESICO PF brackets, platform rates, UMA Jan vs Feb, IVA 16/8 (valid to 2026-12-31), sources[] | backend-sr | M5-03 | 6 | — | `feature/tax-engine-params-2026` | — | [x] |
| M5-05 | Tariff lookup (sort.Search) + property tests (total over (0, ∞)) | backend-sr | M5-03 | 2 | — | `feature/tax-engine-params` | — | [x] |
| M5-06 | Independent double-entry verification of `v1.yaml` against Anexo 8 / LIF / DOF (review on the M5-04 PR) | qa-sr | M5-04 | 3 | — | `feature/tax-engine-params-2026` | — | [x] |
| M5-07 | Golden harness: case YAML format, table-driven runner, skip-with-reason for ⚠️, no update flag, CI hard gate, boundary-case generator (tariff rows, validity dates) | qa-sr | M5-02, M5-05 | 4 | — | `feature/tax-engine-golden-harness` | — | [x] |
| M5-08 | RateCatalog port implementation (allowed IVA rates on a date) for M3/M4 | backend-sr | M5-04 | 2 | PG-M5-F | `feature/tax-engine-rate-catalog` | — | [x] |
| M5-09 | Shared IVA components calculator (charged, creditable, withheld from taxpayer, withheld to remit, 8% by record date, 625 definitive exclusion hook) | backend-sr | M5-02, M5-03 | 3.5 | PG-M5-F | `feature/tax-engine-iva` | — | [x] |
| M5-10 | Strategy registry + InformationalStrategy fallback + RequiredInputs → `incomplete` status | backend-sr | M5-02 | 2.5 | PG-M5-F | `feature/tax-engine-registry` | — | [x] |
| M5-11 | Strategy 612 PF (YTD cash, deductions, PTU, losses, accumulated tariff month n, prior payments, 10% PM withholding) + IVA; goldens PF-612-ISR-01..03, PF-612-IVA-01 + 12-month boundary rows | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 6 | PG-M5-S | `feature/tax-engine-612-pf` | — | [x] |
| M5-12 | Strategy 606 PF (blind 35% vs actual, predial, art. 96, PM 10% withholding → saldo a favor; residential exempt; quarterly → informational); goldens PF-606-ISR-01..02, PF-606-IVA-01 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 4.5 | PG-M5-S | `feature/tax-engine-606-pf` | — | [x] |
| M5-13 | Strategy 626 PF (non-marginal bracket, 1.25% withholding); goldens PF-626-ISR-01..02, PF-626-IVA-01; property: rate non-decreasing | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3.5 | PG-M5-S | `feature/tax-engine-626-pf` | — | [x] |
| M5-14 | Strategy 625 PF (activities 2.1/4/2.5%, 20% without RFC, definitive option, prior-year ≤ $300k → full vs partial); goldens PF-625-01 (IVA part skipped ⚠️), PF-625-02 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 4.5 | PG-M5-S | `feature/tax-engine-625-pf` | — | [x] |
| M5-15 | Strategy 601 PM (coeficiente × YTD accrual, PTU May–Dec, losses, interest withholding, first fiscal year) + IVA + withholdings_to_remit; goldens PM-601-ISR-01..02, PM-601-IVA-01 | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 6 | PG-M5-S | `feature/tax-engine-601-pm` | — | [x] |
| M5-16 | Strategy 626 PM (YTD cash profit × 30%); golden PM-626-ISR-01 + boundaries | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3.5 | PG-M5-S | `feature/tax-engine-626-pm` | — | [x] |
| M5-17 | Partial 621 PF (bimonthly, February accumulated table, rif_year reduction ⚠️); golden PF-621-01 skipped until CPA | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09 | 3 | PG-M5-S | `feature/tax-engine-partial` | — | [x] |
| M5-18 | Partial 622 PM (IVA + explanation, AGAPES deferred) and 603 PM (IVA only, ISR explained) | backend-sr | M5-04, M5-05, M5-06, M5-07, M5-09, M5-15 | 2.5 | — | `feature/tax-engine-partial` | — | [x] |
| M5-19 | Explanation-only catalog: plain-language text for 605, 607, 608, 610 (PF/PM), 611, 614, 615, 616, 620, 623, 624 (forward migration on `regimes.explanation`) + informational goldens | backend-sr | M2-02, M5-10, M5-07 | 3 | PG-M5-S | `feature/tax-engine-informational` | — | [x] |
| M5-20 | Migration `taxpayer_tax_inputs` with RLS + isolation tests | db-sr | M2-03 | 2.5 | — | `feature/tax-engine-inputs` | — | [x] |
| M5-21 | Tax inputs GET/PUT use case validated against `regimes.required_inputs` | backend-sr | M5-20, M5-01, M2-06 | 3 | — | `feature/tax-engine-inputs` | — | [x] |
| M5-22 | EstimateMonth use case: readers, register all strategies at the composition root, IVA netting per taxpayer, withholdings_to_remit, final rounding + `Ajuste por redondeo`, fingerprint, disclaimers + legend | backend-sr | M5-08, M5-10, M5-11, M5-12, M5-13, M5-14, M5-15, M5-16, M5-17, M5-18, M5-19, M5-21, M3-07, M4-08 | 5 | — | `feature/tax-engine-estimate` | — | [x] |
| M5-23 | Estimates + breakdown handlers, contract tests, flag ON/OFF (M6/M8 implied OFF), EXPLAIN ANALYZE, p95 check | backend-sr | M5-22, M2-07 | 3.5 | — | `feature/tax-engine-http` | — | [x] |
| M5-24 | Engine property suite, fixed seed (lines + adjustment = result, monotonic in income, byte-identical JSON) | qa-sr | M5-22 | 3 | — | `feature/tax-engine-property-tests` | — | [x] |
| M5-25 | Flutter estimate screen: per-regime + totals, legend, disclaimers, `incomplete` CTA | frontend-sr | M5-01, M2-10 | 4.5 | — | `feature/tax-engine-mobile-estimate` | — | [x] |
| M5-26 | Flutter breakdown screen: lines, legal basis citations, rounding adjustment | frontend-sr | M5-25 | 4 | — | `feature/tax-engine-mobile-breakdown` | — | [x] |
| M5-27 | Flutter tax-inputs form per regime (target of the `incomplete` CTA) | frontend-sr | M5-25 | 4 | — | `feature/tax-engine-mobile-inputs` | — | [x] |
| M5-28 | Register flag `module_tax_engine` | backend-sr | M5-23, M5-24, M5-26, M5-27 | 0.5 | — | `feature/tax-engine-flag` | — | [x] |
| M5-29 | QA M5: UI figures equal golden figures for each full regime, informational regimes, incomplete flow, flag OFF | qa-sr | M5-28 | 4 | — | `feature/tax-engine-qa-fixes` | OB6 | [x] |
| M5-30 | Tech-lead gate M5 (OCP, purity, no floats, determinism) | tech-lead | M5-29 | 3 | — | (PR review) | — | [x] |
| M5-31 | Security gate M5 | security | M5-29 | 2 | — | `feature/tax-engine-security-gate` | — | [x] |
| M5-32 | Docs M5: engine guide, add-a-regime guide, Dec–Jan parameter update runbook | doc-writer | M5-30, M5-31 | 2.5 | — | `feature/tax-engine-docs` | — | [x] |
| M5-33 | CPA sign-off of golden suite; un-skip ⚠️ cases or record decisions | qa-sr | M5-29 | 2 | — | `feature/tax-engine-cpa-signoff` | OB5 | [ ] |

#### M5 follow-ups

Deferred by the M5 gates and QA; none blocks M5's merge. Each lands with the task or module in
its target column. Sources: [M5-30 tech-lead gate](../reviews/M5-30-tech-lead-gate.md#follow-ups),
[M5-31 security gate](../reviews/M5-31-security-gate.md#4-carried-m5-follow-ups) (G-*, PR numbers;
flag preconditions in its §5), [M5-29 QA](../reviews/M5-29-qa.md#5-issues) (D*),
[M5-06](../reviews/M5-06-params-verification.md#recommendations-non-blocking) and M5-32 (this
docs task: FU7–FU9). The module itself stays **OFF in shared environments** until
[M5-31 §5](../reviews/M5-31-security-gate.md#5-preconditions-to-enable-module_tax_engine-in-staging-or-production)
holds, first the CPA sign-off (M5-33).

| ID | Item | Owner | Target |
|---|---|---|---|
| FU1, D6 | **Surface the collection date.** Accrual regimes (601): show *Fecha de cobro* outside the collapsed *Más detalles* (it decides the IVA; or ask "¿Ya te pagaron?" with the date), plus a hint on the 601 estimate card when income is invoiced but not collected (IVA is 0 until then) | pm + frontend-sr | M3 hardening, before `module_incomes` for PM users |
| FU2 | CPA-REVIEW: PM-601-IVA-01's withheld-to-remit is IVA only; the app adds the LISR 106 ISR withheld from PF fees | qa-sr | M5-33 |
| FU3 | ~~Add-a-regime guide lists every touch point, including both `EngineVersion` pins~~ **Done** ([add-a-regime guide](../guides/add-a-regime.md)) | doc-writer | M5-32 |
| FU4 | M5-31 G-2: validate `EstimateDeps` in `NewEstimator` so a missing port fails the boot | backend-sr | Before M6 wires a second caller |
| FU5 | M5-31 G-4: the prior-payment suggestion spends up to 6 estimate reads per half; count it in the `estimate_reads` budget or batch it | backend-sr + frontend-sr | M5 hardening / M6 |
| FU6 | M5-31 G-5: check that every embedded parameter set defines every key the modules read (not only 2026), at boot or in CI. Until then the [parameter runbook §5](../runbooks/tax-params-update.md#5-tests-for-the-new-year) requires a per-year `TestEmbedded<year>DefinesEveryKeyTheModulesRead` | backend-sr | Before the 2027 parameter file |
| FU7 | Whole-peso payment rounding (CFF 20) is planned as the parameter `rounding.payment_whole_pesos` (ADR-0002 §4), but no file defines it and no code reads it: wire it (default off) so the CPA's answer is a data change | backend-sr | M5-33 (if the CPA confirms rounding) |
| FU8 | Year-bearing texts are code: the `not_tax_advice` text ("LIF 2026 y Anexo 8 RMF 2026", `texts.go` + contract examples) and the strategies' citations ("Anexo 8 RMF 2026", "LIF 2026", the border decree's end). Derive them from the selected parameter set's `sources[]` | backend-sr (+ architect for the examples) | Before the 2027 parameter file |
| FU9 | The generated golden edges (tariff rows, validity dates) and the figures digest read `params/data/2026/v1.yaml` only: iterate over every embedded set | qa-sr | With the 2027 parameter file |
| D3 | *Datos fiscales*: the suggestion row disappears on the first keystroke and the fields below jump ~260 px, so a second tap lands on the wrong month. Keep the row's space, or hide the suggestion on save | frontend-sr | M5 UX polish |
| D4 | Breakdown: result lines use a smaller label than input lines (inverted hierarchy); every line repeats a source chip and a *Fundamento* disclosure, so 612 takes ~7 screens. Also: the totals card shows no *Completo* chip when everything is complete | frontend-sr + design | M5 UX polish |
| D5 | RESICO expense hint repeats the same sentence twice | frontend-sr | M3/M4 UX polish |
| D7 | MODULE-tax-engine §7 "M5 OFF implies M6 and M8 OFF" is declared by the M6/M8 flag tasks | backend-sr | M6-xx / M8-xx flag tasks |
| G-3 | `infra/taxinputs` and the profile reader read taxpayers by id under RLS alone: add an owner check in the query if shared access (accountant, delegated users) is ever designed | architect | ADR for any delegated-access feature |
| #77 should-fix 2 | `provisional_parameters` notice on the records forms (categories/defaults responses) | architect | Next contract re-freeze |
| #98 FU1 | Fingerprint hashes informational sums and multi-enum order (false "changed" marks in M8) | backend-sr | M8 |
| #98 FU2 | Replace the weekends-only due-date calendar with the `holidays` table | backend-sr | M9-03/M9-04 (blocker beyond 2028 parameters) |
| #90 FU | RESICO PF months after the cap still use RESICO rates (warning only) | owner + CPA | M5-33 |
| #101 FU 1–3 | A half with an error stays open next to the opened one (up to 12 fields); `paymentMonths` offers months of a future fiscal year; `CitationDisclosure` could reuse `FormDisclosure` | frontend-sr | M5 UX polish |
| #101 FU 4 | ~~MODULE §8 still says "inline editable 12-row table"~~ **Done** (§8 describes the shipped halves of ≤ 6 and the suggestion) | doc-writer | M5-32 |
| #102 nit | Contract examples show `engine_version: 1.0.0` (current 1.0.1) | architect | Next contract re-freeze |
| #102 note | ~45 ms per request is spent outside the measured SQL (likely pgx generic plans); profile it when M6 reads 12 months | backend-sr | M6-03 |
| M5-06 rec. 1–3 | Cite the DOF edition of Anexo 8 (nota 5777219) as the primary source; keep a SHA-256 of each fetched source in the verification report (now in the runbook §2/§4); confirm the scope of LIF 2026 art. 25 fr. IX a) (platform IVA withholding on personas morales) | backend-sr (1), qa-sr (2), pm + CPA (3) | With the 2027 parameter file; M5-33 |
| D2, PR #108 | ~~A regime always starts in the current month~~ **Done** (#108: onboarding and *Agregar régimen* ask the start month, from 2026-01) | — | — |

### M6 — Dashboard (`module_dashboard`)

| ID | Task | Owner | Deps | Est (h) | Par | Branch | Blocked by | Done |
|---|---|---|---|---|---|---|---|---|
| M6-01 | Contract freeze: `/v1/taxpayers/{id}/dashboard` | architect | P0-05 | 1 | — | `feature/dashboard-contract` | — | [x] |
| M6-02 | Dashboard use case: one aggregate pass for 12 months, Estimator, income vs expense, trial banner via AccessPolicy | backend-sr | M5-22, M6-01 | 4 | — | `feature/dashboard-backend` | — | [ ] |
| M6-03 | Dashboard handler, contract tests, flag ON/OFF, benchmark p95 < 1 s, EXPLAIN ANALYZE | backend-sr | M6-02, M5-23 | 3 | — | `feature/dashboard-backend` | — | [ ] |
| M6-04 | Flutter Home: this month's estimate, income vs expense, trend chart, trial banner, empty state | frontend-sr | M6-01, M5-25 | 5 | — | `feature/dashboard-mobile` | — | [ ] |
| M6-05 | Register flag `module_dashboard` + preset `demo_core` | backend-sr | M6-03, M6-04 | 0.5 | — | `feature/dashboard-flag` | — | [x] |
| M6-06 | QA M6 + `demo_core` preset smoke | qa-sr | M6-05 | 2.5 | — | `feature/dashboard-qa-fixes` | OB6 | [x] |
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
