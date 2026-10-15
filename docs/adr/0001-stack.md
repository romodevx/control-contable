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
                      (sole exception: x/text/unicode/norm, ADR-0004)
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
| API client generation | **swagger_parser** (pure Dart, supports OAS 3.1) → retrofit + json_serializable/freezed | swagger_parser 1.45.0, retrofit 4.10.0, retrofit_generator 10.2.11, freezed 4.0.2 |
| Token storage | flutter_secure_storage (Keychain / Keystore) | 11.2.0 |
| IAP | RevenueCat `purchases_flutter` (`appUserID` = our user UUID) | 10.15.2 |
| Push | firebase_messaging | 16.7.0 |
| Flags | one `FlagsProvider` (Riverpod) fed by `GET /v1/flags` + a `useFlag` helper. UX only | — |
| Release | `--obfuscate --split-debug-info`. Symbols are uploaded to crash reporting and never bundled | — |

### Context7 verification (2026-10-13)

P0-21. Each pinned Dart/Flutter library was checked in the Context7 MCP, and each pin was
re-checked against `pub.dev/api/packages/<name>` the same day. Every pin above is still the
latest release, except retrofit_generator, deliberately held at 10.2.11 (two-week rule, row below). Toolchain: Flutter 3.47.7, Dart 3.13.5. No API change forces a decision
change. The owner confirmed **Riverpod** (O5 closed, OB4 resolved).

| Library | Context7 ID | Pin = latest? | API checked | Notes for the first user |
|---|---|---|---|---|
| flutter_riverpod / riverpod_generator | `/rrousselgit/riverpod` (docs at v3.0.2) | 3.4.3 / 4.0.9, yes | `@riverpod` class notifiers (`extends _$X`, `build()`), `ProviderScope(overrides: [...overrideWithValue(...)])`, `ProviderContainer.test()` | Use `ProviderContainer.test()` in unit tests (auto-dispose replaces the 2.x `createContainer` helper). Scoped providers need `@Dependencies` or an override, enforced by riverpod_lint: add it in P0-22 |
| go_router | `/websites/pub_dev_packages_go_router` | 18.0.2, yes | `GoRouter(routes:)`, `MaterialApp.router(routerConfig:)`, top-level `redirect` | Context7 coverage is thin (15 snippets, no `redirect` detail). Read the pub.dev API docs for `redirect`/`refreshListenable` before wiring the auth and flag guards |
| dio | `/cfug/dio` | 5.11.1, yes | `InterceptorsWrapper`, `QueuedInterceptor`, `DioException.response`, `RequestOptions.copyWith` | Refresh-on-401 goes in a `QueuedInterceptor` so concurrent 401s refresh once. Never retry the re-auth/password endpoints through it (a wrong password must not be sent twice) |
| swagger_parser | `/carapacik/swagger_parser` | 1.45.0, yes | `swagger_parser.yaml` keys (`json_serializer: freezed`, `use_freezed3`, `unknown_enum_value`, `root_client`); `OAS.v3_1`; `type: [T, "null"]` → `T?` | Spiked against our contract: see [ADR-0003](0003-dart-client-generation.md). Set `use_freezed3: true` (default is `false`, Freezed 2 syntax) |
| freezed / json_serializable | `/rrousselgit/freezed` | freezed 4.0.2, freezed_annotation 3.1.0, json_serializable 6.14.1, yes | `@freezed sealed/abstract class X with _$X`, `fromJson` via `=>` factory | Since v3 every class needs `sealed` or `abstract`. `fromJson` is generated only when the factory uses `=>` |
| retrofit (+ retrofit_generator) | via swagger_parser docs | retrofit 4.10.0, yes. retrofit_generator latest is 10.2.12 (published today) | Generated clients compile and analyze clean | **Pin retrofit_generator 10.2.11** (2026-09-04): 10.2.12 is under two weeks old. The spike passed on both |
| purchases_flutter | `/websites/pub_dev_purchases_flutter` | 10.15.2, yes | `Purchases.configure(PurchasesConfiguration(apiKey)..appUserID = uuid)`, `Purchases.logIn(appUserID)` → `LogInResult` | Configure with our user UUID at sign-in; call `logIn` when the user changes on the same device, never leave an anonymous RevenueCat id behind |
| firebase_messaging | `/firebase/flutterfire` | 16.7.0, yes | `requestPermission()`, `getToken()`, `onTokenRefresh`, top-level `@pragma('vm:entry-point')` `onBackgroundMessage` | On Apple, call `getToken()` only after `getAPNSToken()` is non-null. FlutterFire `main` already deprecates `getToken`/`onTokenRefresh` in favour of `register()` + `onRegistered` (FID). Not in 16.7.0; M9 must re-check before building device registration |

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
- **Spike result (P0-21)**: GO with swagger_parser, with two conditions. See
  [ADR-0003](0003-dart-client-generation.md).

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
- **Not adopted.** The edge flood limiter lives in process (token bucket per IP). The auth
  limits (per operation, email, family, account) moved to PostgreSQL in
  [ADR-0005](0005-rate-limits-in-postgres.md) (M1-10).

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

## Context7 verification (2026-10-13)

P0-02. Each pinned Go library was checked in the Context7 MCP (it was available this time)
and each pin was re-checked against `proxy.golang.org/<module>/@latest`. Every pin above is
still the latest release. No API change forces a decision change.

| Library | Context7 ID | Pin = latest? | API checked | Notes for the first user |
|---|---|---|---|---|
| oapi-codegen v2 | `/oapi-codegen/oapi-codegen` | v2.8.0, yes | config `generate: {std-http-server, strict-server, models, embedded-spec}`; OAS 3.1 (`type: [T, "null"]`, `oneOf`+`const` enums, webhooks) | `strict-server` must sit next to a server generator (`std-http-server`). Unused types are pruned by default. Implicit `additionalProperties` are ignored. ~~Use `compatibility.schema-merging-behavior: v2`~~ (corrected in P0-06: no such key in v2.8.0; see below) |
| goose v3 | `/pressly/goose` (docs at v3.28.0) | v3.29.0, yes | `goose.NewProvider(goose.DialectPostgres, *sql.DB, fs.FS, opts...)`, `p.Up(ctx)`, `p.HasPending`, `*goose.PartialError` | Use the **Provider** API (no globals) instead of `SetBaseFS` + `goose.Up`. v3.29.0 is two days old: Context7 docs trail it by one minor; check the changelog before P0 migrations |
| GORM + driver | `/websites/gorm_io` | gorm v1.31.2, driver v1.6.3, yes | `gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}))`; `WithContext` = `Session{Context}`; `Exec` with `?` placeholders | Build GORM on the same `*sql.DB` (pgx stdlib) that goose uses, so one pool. Set RLS with `SELECT set_config('app.current_user_id', ?, true)` inside the tx (`SET LOCAL` cannot take a bind parameter) |
| pgx v5 | `/jackc/pgx/v5.11.0` | v5.11.0, yes | `stdlib` driver auto-registered as `"pgx"` / `"pgx/v5"`; `stdlib.OpenDBFromPool(pool)` | `OpenDBFromPool` forces max idle conns to 0 and closing the `*sql.DB` does not close the pool. The source now has a `sql_go1.27.go` file: Go 1.27 is supported |
| testcontainers-go | `/testcontainers/testcontainers-go` | v0.44.0, yes | `postgres.Run(ctx, "postgres:18", WithDatabase/WithUsername/WithPassword/WithInitScripts, postgres.BasicWaitStrategies())`; `Snapshot`/`Restore` | `Restore` drops and re-creates the DB from a template: fast per-test isolation for the RLS suites. `WithSQLDriver("pgx")` is needed for snapshots with pgx. Custom `wait.ForSQL` URL funcs take a `network.Port` |
| golang-jwt v5 | `/golang-jwt/jwt` | v5.3.1, yes | `jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(ed25519.PrivateKey)`; `ParseWithClaims(..., WithValidMethods([]string{"EdDSA"}), WithIssuer, WithAudience, WithExpirationRequired(), WithLeeway)` | Always pass `WithValidMethods` (alg-confusion guard) and `WithExpirationRequired`. Embed `jwt.RegisteredClaims` in the custom claims type |
| libopenapi-validator | `/pb33f/libopenapi-validator` | v0.16.1, yes | `validator.NewValidator(libopenapi.Document, opts...) (Validator, []error)`; `ValidateHttpResponse(req, resp)`; `ValidateHttpRequestResponse` | Takes a parsed `libopenapi.Document`, not bytes; returns a slice of build errors; call `Release()` when done. v0.16.1 is two days old (pre-1.0): pin exactly and read release notes on each bump |
| Firebase Admin Go | `/firebase/firebase-admin-go` | v4.22.0, yes | `firebase.NewApp(ctx, *Config)`, `app.Messaging(ctx)`, `client.Send`, `SendEachForMulticast` (≤ 500 targets), `messaging.IsUnregistered(err)` | `Messaging` fails without a project ID. Use `IsUnregistered` to delete stale device tokens (M9). Batch sends go through `SendEach*` (per-target `BatchResponse`) |

Also checked: golangci-lint v2 configuration (`/golangci/golangci-lint`) for depguard
`rules.<name>.{files,allow,deny}` and forbidigo `forbid[].pattern` (v2 renamed `p` to
`pattern`). The resulting `backend/.golangci.yml` was proven against probe files: depguard
rejects `net/http` in `domain/` and `database/sql` in `usecase/`, and forbidigo rejects
`float64` and `time.Now` in `domain/` but not in `infra/`.

## oapi-codegen and OpenAPI 3.1 (P0-06, 2026-10-13)

Re-checked in Context7 (`/oapi-codegen/oapi-codegen`) and against the v2.8.0 source while
wiring `make generate`. Config: `backend/adapter/http/gen/oapi-codegen.yaml`; the tool is pinned
by the `tool` directive in `backend/go.mod` (`go tool oapi-codegen`).

| Caveat | Effect on our contract | Workaround |
|---|---|---|
| `compatibility.schema-merging-behavior` does not exist in v2.8.0 (the P0-02 note above was wrong) | Config fails to parse | Dropped. v2.8.0 already merges `allOf` correctly by default; the `old-*` flags only opt back into legacy output |
| Multi-type unions (`type: [string, boolean, 'null']`) error with `unhandled Schema type` when used as a property. The README says they become `any`; that holds only at top level | `RegimeOptionDef.default` broke generation | Codegen-only **OpenAPI Overlay** (`adapter/http/gen/overlay.yaml`) adds `x-go-type: any` to that property. The contract is unchanged; every new overlay action must be listed here. Prefer `oneOf` or a single type in new schemas |
| `[T, "null"]` nullability, `oneOf` + `const` enums | Generate fine | None |
| An optional nullable property becomes a plain pointer: an absent key and `null` both decode to `nil` (M4-07) | A merge-patch `null` (`ExpenseUpdate.cfdi_uuid`, `deductible`, `description`; `IncomeUpdate.invoiced_on`, `collected_on`, `platform_activity`, `cfdi_uuid`, `description`, M3-06; `PersonalDeductionUpdate.donee_type`, `student_label`, `tuition_level`, `scholarship_cents`, `cfdi_uuid` and `SalaryStatementUpdate.employer_label`, ANN-15) could never clear the field | Overlay actions set `x-go-type: nullable.Nullable[T]` (`github.com/oapi-codegen/nullable`) with `x-go-type-skip-optional-pointer` on those fourteen properties. Only the Go type changes; the schemas and the request validation stay the contract's. The global `output-options.nullable-type` is not used: it would retype every nullable response field |
| A **required** response header becomes a value field the strict response writes unconditionally (`w.Header().Set(name, fmt.Sprint(v))`); an optional one becomes a `*T` written only when non-nil (contract 0.7.0) | `Cache-Control: no-store` (`components.headers.NoStore`) is required on every response but written once, by `middleware.NoStore`; as a required field every handler would have to fill it, and a forgotten one would overwrite the middleware's value with an empty header | An overlay action sets `required: false` on `components.headers.NoStore` for codegen only: the generated `CacheControl` fields are nil pointers no handler sets. The wire contract keeps `required: true`, and the contract tests (libopenapi-validator, which reads the contract, not the overlay) fail a response without it. Other headers (`WWW-Authenticate`, `Retry-After`, the legal documents' `Cache-Control`) stay required in Go |
| The embedded spec Go-normalizes `operationId` (`getHealth` → `GetHealth`) | `GetSwagger()` drifts from the contract | `compatibility.preserve-original-operation-id-casing-in-embedded-spec: true` |
| Strict middlewares receive the Go-normalized name and run **after** the request body is decoded. `StdHTTPServerOptions.Middlewares` run after the wrapper binds path and query parameters. A wrapper outside the mux runs before `r.Pattern` is set. The `ServeMux` answers 405 + `Allow` for another method on a registered path | Each of those placements answers 400 or 405 (not 404) for a disabled module, revealing it | The generated flag map exposes `RouteModule(pattern)` (`"METHOD /path"` → `x-module`, empty `BaseURL`) besides `OperationModule(operationId)`. P0-15 gates **at registration**: a `gen.ServeMux` passed to `HandlerFromMux` wraps each handler by the pattern it registers, panics at startup on a pattern `RouteModule` does not know (fail closed, e.g. a non-empty `BaseURL`), and turns the mux's 404/405 for gated paths into a uniform 404. Pinned by `adapter/http/gen/gate_placement_test.go` |

Drift: `make codegen-check` regenerates and fails if `adapter/http/gen` differs from what is
committed. `cmd/flagmapgen` fails generation when an operation lacks `operationId` or a valid
`x-module`, and `adapter/http/gen/contract_test.go` re-checks the embedded spec, the map, the
`StrictServerInterface` and the ServeMux patterns.

## HTTP server and request validation (P0-17, 2026-10-13)

Re-checked in Context7 (`/getkin/kin-openapi`) and against the kin-openapi v0.142.0 source
while wiring `cmd/api`. `go.mod` pinned v0.142.0, which govulncheck flags (GO-2026-6112,
nil-pointer panic reachable from `openapi3filter.ValidateRequest`); P0-17 bumps it to v0.149.0,
the version the table above names.

| Finding | Effect | Decision |
|---|---|---|
| The per-operation pipeline must run after routing but before the generated wrapper (P0-06 note) | A gate or validator outside the mux would answer 400/405 for a disabled module | `adapter/http/router` is the `gen.ServeMux`: it wraps each handler as it is registered with **flag gate → authn → kin-openapi validation**. A route with no `x-module` or no operation in the contract fails `router.New` at startup. Unmatched requests (unknown path, other method) get the same 404 envelope as a disabled module, never the mux's 405 + `Allow` |
| Gate before authn (C4 listed authn first) | With authn first an anonymous caller gets 401 from an enabled module and 404 from a disabled one | Gate first; C4 updated |
| For an OAS 3.1 document, `openapi3filter` validates with santhosh-tekuri/jsonschema (2020-12) and does **not** assert `format` (uuid, email, date); there is no option to turn it on | A malformed uuid/email/date passes kin validation | The generated binding still rejects them (typed `openapi_types.UUID`/`Email`/`Date`); `router` maps those errors to the same 400 envelope (`code: format`). Types, required, pattern, length, enum and range come from kin |
| 2020-12 errors are strings (`at '/password': minLength: …`) that quote the rejected value | Echoing them would leak PII | `FieldError.field` and `.code` are parsed from the location and keyword; `message` is generic |
| kin compiles the operation schema on every request | ~42 µs per validated POST (M4 Pro, `BenchmarkValidatedRequest`) | Acceptable; revisit if profiling says otherwise |
| Rate limit (O9) | — | In-process GCRA token bucket per client IP (IPv6 by /64), `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST`; `TRUSTED_PROXY_HOPS` picks the X-Forwarded-For entry (0 = TCP peer) |

### `/healthz`: liveness vs readiness (P0-17 review, 2026-10-13)

`getHealth` is `x-frozen`, so its path, parameters and responses cannot change. Decision:

- **`/healthz` is the liveness and startup probe.** It answers 200 whenever the process can
  serve; the body (`status`, `database`) is the readiness *report*. The DB ping is bounded at
  500 ms, under the 1 s default `timeoutSeconds` of a Cloud Run HTTP probe, so a hung database
  never fails liveness and never triggers a restart loop.
- **No status-code readiness gate in P0.** Every instance shares one Cloud SQL database:
  pulling instances out of rotation when it is down serves no one better, it only swaps our
  503 envelope for Cloud Run's. Handlers keep answering the 503 envelope (C4 failure modes).
- **Alerting reads the body**: a Cloud Monitoring uptime check on `/healthz` with a content
  match on `"database":"unreachable"`.
- **If a status-code readiness signal is ever needed** (Cloud Run readiness probes, a load
  balancer health check), it is a *new* platform operation (e.g. `GET /readyz`, 200/503, `x-module:
  platform`, `security: []`) added to the contract with its own ADR row. That is additive: no
  frozen operation changes. The 503 that `getHealth` already declares stays reserved.
