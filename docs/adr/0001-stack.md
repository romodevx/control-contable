# ADR-0001 — Technology stack

| Field | Value |
|---|---|
| Status | PROPOSED (Phase 2 Design), pending owner approval of the items marked **OWNER** |
| Date | 2026-10-13 |
| Deciders | Architect; owner for the OWNER items |
| Related | [PRD](../project-specs/PRD.md), [MODULES](../project-specs/MODULES.md), [ADR-0002](0002-tax-engine.md), [C4](../architecture/c4.md), [data model](../architecture/data-model.md) |

## Context

The project is a Go modular monolith with a Flutter client, PostgreSQL with RLS, store
in-app purchases through RevenueCat, push reminders, and OpenAPI spec-first development.
The owner fixed the main stack: Go, Clean Architecture, GORM, REST, Flutter, PostgreSQL 18,
RevenueCat and per-module flags. This ADR fixes the libraries that fill that stack, with
versions, and lists the choices the owner still has to make.

Versions were checked against the registries (proxy.golang.org, pub.dev, npm, go.dev,
postgresql.org) on 2026-10-13. The Context7 MCP was **not available** in this session, so
each library's API must be checked in Context7 before its first use, as the global rules
require. Library versions are pinned in `go.mod` and `pubspec.lock`. Renovate or Dependabot
proposes upgrades.

> **Correction to the brief.** The installed Go is **go1.27.2**, not 1.25.6, and
> golangci-lint 2.14.0 is built with go1.27.1. Go supports only the two latest major
> releases (1.27 and 1.26 today), so Go 1.25 is out of security support. Decision:
> `go 1.27` + `toolchain go1.27.2`.

## Decision: backend (Go)

| Concern | Choice | Version | Notes |
|---|---|---|---|
| Language | Go | 1.27.x | `go.mod`: `go 1.27` |
| HTTP routing | **stdlib `net/http.ServeMux`** (method + path patterns) | stdlib | The routes are generated from the contract, so a router library adds little |
| Contract → server code | **oapi-codegen v2**, `std-http-server` + `strict-server` | v2.8.0 | Supports OAS 3.1 (`type: [T, "null"]`, webhooks). Generated code lives in `adapter/http/gen`, is never edited, and CI checks it is up to date |
| Request validation | oapi-codegen request validation middleware (kin-openapi) | kin-openapi v0.149.0 | Validates every input at the boundary (security rules) |
| Contract tests | **pb33f/libopenapi-validator** (native 3.1) | v0.16.1 | Checks every response in integration tests against `openapi.yaml` |
| ORM | GORM + `gorm.io/driver/postgres` (pgx v5 underneath) | gorm v1.31.2, driver v1.6.3, pgx v5.11.0 | Used in `infra/` only. `WithContext` everywhere, explicit `Preload`, never `AutoMigrate` |
| Migrations | **goose v3**, SQL files embedded with `embed.FS` | v3.29.0 | Versioned SQL, run as `cmd/migrate` with the owner role; the app role never runs DDL |
| Auth tokens | **golang-jwt/jwt v5**, EdDSA (Ed25519) access tokens | v5.3.1 | See the auth decision below |
| Password hashing | **argon2id** (`golang.org/x/crypto/argon2`) | x/crypto v0.58.0 | OWASP-recommended parameters, stored in PHC string format so they can be raised later |
| Money | `int64` centavos; in-house fixed-point `Rate` type in the domain | — | See the money decision below and ADR-0002 |
| Logging | `log/slog` (JSON) with a PII redaction handler | stdlib | Request ID in every line. The RFC and email are never logged |
| Push | Firebase Admin Go SDK (FCM HTTP v1; APNs through FCM) | firebase.google.com/go/v4 v4.22.0 | Answers PRD Q12: yes, FCM for both platforms |
| Scheduling | In-process ticker + `pg_try_advisory_lock` leader election + idempotent `notification_log` | stdlib | See the scheduling options below |
| Tests | testing + testify; testcontainers-go with a `postgres:18` image | testify v1.12.1, testcontainers-go v0.44.0 | Table-driven tests; a real PostgreSQL for integration and RLS tests |
| Lint and security | golangci-lint v2.14, govulncheck, gosec (through golangci) | — | `forbidigo` bans `float32/float64` and `time.Now` inside `domain/` (ADR-0002) |

### Backend layout (Clean Architecture, modular monolith)

```
backend/
  cmd/api/            composition root: config, DI, HTTP server
  cmd/worker/         reminders + soft-delete purge (same binary is fine; separate entrypoint)
  cmd/migrate/        goose runner (owner role)
  internal/
    platform/         P0: config, errors, flags, access policy, tx+RLS, clock, crypto
    <module>/         auth, taxpayers, incomes, expenses, taxengine, dashboard,
                      subscriptions, history, reminders
      domain/         entities, value objects, domain errors. Imports stdlib only
      usecase/        interactors + the ports they consume (small interfaces)
      adapter/http/   strict-server handlers: DTO <-> domain mapping
      infra/          GORM repositories, external clients (RevenueCat, FCM)
  migrations/         goose SQL
```

- The **Dependency Rule** is enforced in CI. A `depguard` rule forbids `gorm.io/*`,
  `net/http` and other `internal/*/infra` packages inside `domain/` and `usecase/`.
- Modules talk to each other only through **ports defined by the consumer**. For example,
  `taxengine/usecase` declares `LedgerReader`, and an adapter in `incomes/infra` and
  `expenses/infra` implements it. No module imports another module's `domain/` or `usecase/`.
- oapi-codegen produces a single `StrictServerInterface`. The composition root builds it by
  embedding one handler struct per module. Each operation carries `x-module: module_<name>`
  in the contract. A generated `operationID → flag` map drives the flag middleware, and a
  contract test fails if an operation has no `x-module`. A module whose flag is OFF
  answers **404** before its handler runs.

## Decision: mobile (Flutter)

| Concern | Choice | Version |
|---|---|---|
| SDK | Flutter stable | 3.47.7 |
| State management | **Riverpod 3** + `riverpod_generator` | flutter_riverpod 3.4.3, riverpod_generator 4.0.9 |
| Navigation | go_router | 18.0.2 |
| HTTP | dio | 5.11.1 |
| API client generation | **swagger_parser** (pure Dart, supports OAS 3.1) → retrofit + json_serializable/freezed | swagger_parser 1.45.0, retrofit 4.10.0, freezed 4.0.2 |
| Token storage | flutter_secure_storage (Keychain / Keystore) | 11.2.0 |
| IAP | RevenueCat `purchases_flutter` (`appUserID` = our user UUID) | 10.15.2 |
| Push | firebase_messaging | 16.7.0 |
| Flags | one `FlagsProvider` (Riverpod) fed by `GET /v1/flags` + a `useFlag` helper. UX only | — |
| Release | `--obfuscate --split-debug-info`. Symbols are uploaded to crash reporting and never bundled | — |

## Options considered

### O1. Router: stdlib ServeMux (chosen) vs. chi v5.4.0
- **ServeMux**: no dependency. Method and wildcard patterns have been stdlib since Go 1.22.
  oapi-codegen generates for it directly. Middleware is plain `func(http.Handler) http.Handler`.
- **chi**: nicer route groups and sub-routers. That matters little, because the routes come
  from codegen.
- Chosen: **ServeMux**, for fewer dependencies with no loss of capability.

### O2. Migrations: goose v3.29.0 (chosen) vs. golang-migrate v4.20.1
- **goose**: SQL plus optional Go migrations, a library API (`goose.Up` with `embed.FS`) and
  `-- +goose StatementBegin` blocks for policies and functions. Actively maintained.
- **golang-migrate**: has separate up/down files and a wider driver range we don't need.
  Its "dirty" state needs manual fixing after a failure.
- Chosen: **goose**.

### O3. Contract codegen: oapi-codegen (chosen) vs. ogen vs. hand-written handlers
- **ogen**: fast, with strong typing and its own router. It is opinionated (optional types
  everywhere) and has a smaller community.
- **Hand-written handlers**: they drift from the contract, which breaks the "contract is the law" rule.
- Chosen: **oapi-codegen** strict server. It has the largest community, supports 3.1 and std-http.

### O4. Dart client: swagger_parser (chosen) vs. openapi-generator `dart-dio` (CLI v2.41.0, JVM)
- `openapi-generator` needs a JVM and its OAS 3.1 support is still partial. `swagger_parser`
  is pure Dart, supports 3.1 and produces idiomatic retrofit/freezed code.
- **Risk**: we must confirm that `swagger_parser` handles our `oneOf`/`const` constructs. A
  half-day spike in P0 covers this. Fallback: openapi-generator with an overlay that
  downgrades the spec to 3.0.

### O5. Flutter state management: Riverpod 3 (chosen) vs. flutter_bloc 9.1.1
- **Riverpod**: compile-safe dependency injection, provider overrides in tests (TDD-friendly),
  async caching (`AsyncValue`) for dashboard/estimate screens, and less boilerplate.
- **Bloc**: explicit events and states, good for strict event audit trails. It needs more code
  per screen, and most screens here are CRUD plus read-only views.
- Chosen: **Riverpod**. **OWNER**: if the mobile developer already knows Bloc well, the
  team's existing skill outweighs the difference. Both are fine.

### O6. Authentication: JWT access + rotating opaque refresh (chosen) vs. opaque sessions only
- **A. JWT access (EdDSA, 15 min) + opaque refresh token (256-bit random, stored as SHA-256,
  30-day sliding, 90-day absolute limit), rotated on every use, grouped by `family_id`.**
  Reusing a rotated token revokes the whole family (US-A2). Access checks need no database
  read. Logout and account deletion revoke the family; the worst case is an access token
  living for up to 15 more minutes.
- **B. Opaque session token only** (looked up in the DB on every request): instant
  revocation, but one extra query per request on a p95 budget of 300 ms. Mobile clients do
  not benefit from statelessness either way.
- Chosen: **A**. The refresh token has the form `<user_uuid>.<secret>`, so the server can
  set the RLS user context before it looks the token up (no RLS bypass). The 15-minute
  revocation lag is acceptable for this risk level.

### O7. Money representation: int64 centavos + fixed-point Rate (chosen) vs. shopspring/decimal v1.5.0
- **int64 centavos** at the API, in the DB (`bigint`) and in the domain. Rates use a
  fixed-point `Rate` (scale 10^-6, enough for 4-decimal coeficientes and 2-decimal percent
  rates). `Cents × Rate` uses a 128-bit intermediate (`math/bits`) with explicit rounding,
  so it cannot overflow and is fully deterministic.
- **shopspring/decimal**: arbitrary precision and well known, but slower, with a
  heap-allocated big.Int and a third-party dependency inside the domain.
- Chosen: **in-house `money` package** (about 150 lines, 100% covered by unit tests). Over
  the wire, amounts are `int64` centavos and rates are decimal **strings** (`"0.1234"`).
  JSON never carries floats.

### O8. Background scheduling: ticker + advisory lock (chosen) vs. River v0.49.0 vs. platform cron
- One job type (daily reminders plus a nightly purge) does not justify a job-queue
  dependency (YAGNI). Exactly-once delivery comes from `UNIQUE (user_id, due_date, kind)` on
  `notification_log`, not from the scheduler. If more async jobs appear, adopt **River**
  (Postgres-backed, no Redis) in a new ADR.

### O9. Redis
- **Not adopted.** Rate-limit counters live in process (token bucket per IP and account). If
  we scale beyond one instance, move them to Postgres or Redis through a new ADR.

## Decisions that need the owner (OWNER)

1. **Email verification at registration.** US-A1 asks that a duplicate email return a
   generic response so accounts cannot be enumerated. That is only possible if registration
   does **not** return tokens right away.
   - *A (recommended)*: `POST /auth/register` → `202`, a 6-digit code is emailed, and
     `POST /auth/verify-email` returns the tokens. Pros: no enumeration; verified emails for
     password reset and ARCO requests. Cons: an email provider (another processor under
     LFPDPPP) and about 30 seconds more for G1.
   - *B*: `201` with tokens; a duplicate gets `409`. Pros: simpler and faster onboarding.
     Cons: email enumeration (only rate limiting mitigates it), and emails are unverified.
   - The contract is drafted with **A**.
2. **Hosting and data residency.**
   - *Recommended*: GCP Cloud Run + Cloud SQL for PostgreSQL 18 in `northamerica-south1`
     (Querétaro, MX), with Secret Manager. Firebase is GCP, so there is one vendor. Data
     stays in Mexico, which simplifies the LFPDPPP narrative.
   - *Alternatives*: Fly.io or Render + Neon (cheaper, but data outside MX), or AWS
     (`mx-central-1`).
   - Cost and the processor list are owner calls.
3. **Riverpod vs. Bloc** (O5), depending on who builds the mobile app.
4. **Email provider** (only if 1A): e.g. Amazon SES, Postmark or Resend. It goes on the
   processor list.

## Consequences

- (+) The contract drives server stubs, request validation, contract tests and the Dart
  client. One source of truth.
- (+) Small, mainstream dependencies, and no Redis or queue at v1.
- (+) The Dependency Rule and the float/time bans are machine-checked, not left to review.
- (−) Generated code must be regenerated on every contract change (CI checks drift).
- (−) Access-token revocation can lag by up to 15 minutes.
- (−) Library APIs were not checked in Context7 in this session. That check is a P0 entry criterion.
- Upgrades: Go minor releases monthly. Re-evaluate the Go major version each February and
  August (Go 1.28 is due around 2027-02).
