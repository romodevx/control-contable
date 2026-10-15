# Platform packages (`backend/internal/platform`)

The P0 shared kernel: cross-cutting packages every module may import. They carry no feature
flag, since the platform is always on. This page summarises each package's responsibility and the
invariants callers may rely on. The authoritative detail is each package's Go doc comment
(`go doc ./internal/platform/<pkg>` from `backend/`). When this page and the code disagree, the
code wins and this page has a bug.

Rule (P0-29 FU4): **no new business logic goes into `platform/`**. `access` is the one exception,
kept as a documented shared port by [ADR-0011](../adr/0011-access-and-authorization.md). Business concepts that several bounded contexts must agree
on (dates, IVA treatment, CFDI UUID, withholding shares, the domain error model) go in the fiscal
shared kernel, `backend/internal/kernel/fiscal` ([ADR-0007](../adr/0007-fiscal-shared-kernel.md)).

Sealed platform data uses the rotation and re-seal procedure in
[ADR-0012](../adr/0012-encryption-key-rotation.md) and the
[key rotation runbook](../runbooks/key-rotation.md).

## Map

| Package | Responsibility | Used by |
|---|---|---|
| [`money`](#money) | Fixed-point amounts and rates, the single rounding rule | domain, tax engine |
| [`clock`](#clock) | The only reader of the wall clock, in Mexico City time | use cases, flags, access |
| [`tx`](#tx-and-pgdb) | Unit-of-work port that carries the RLS user context | use cases |
| [`pgdb`](#tx-and-pgdb) | PostgreSQL adapter: pool, `TxManager`, readiness pinger, worker leader lock, runtime-role guard | `cmd/`, module `infra/` |
| [`flags`](#flags) | Module flag service (gate + `GET /v1/flags`), admin for `cmd/flags` | router, `cmd/flags` |
| [`access`](#access) | AccessPolicy: trial clock, trial caps, entitlements, taxpayer cap | income/expense/taxpayer use cases |
| [`apperr`](#apperr-and-httperr) | Delivery-agnostic application error model (contract error codes) | use cases, adapters |
| [`httperr`](#apperr-and-httperr) | HTTP status mapping and the single error envelope. Lives in `backend/adapter/http/httperr` since WP4 | HTTP adapters |
| [`crypto`](#crypto) | argon2id passwords, AES-256-GCM keyring, HMAC blind indexes | auth, taxpayers |
| [`config`](#config-logging-requestid) | Environment and secret loading, validated up front | `cmd/` |
| [`logging`](#config-logging-requestid) | slog JSON logger with PII redaction | everything |
| [`requestid`](#config-logging-requestid) | Request correlation id in the context | middleware, logging, errors |
| [`ratelimit`](#ratelimit) | In-process keyed token bucket (GCRA) | HTTP middleware |
| [`worker`](#worker) | Leader-only job runner for `cmd/worker` | `cmd/worker` |
| [`pgtest`](#test-support) | Integration harness: real PostgreSQL, role switching, RLS assertions | integration tests |
| `flags/flagtest`, `worker/workertest` | Test helpers (flag ON/OFF suites, manual tickers) | tests |

## money

Three `int64` fixed-point types (ADR-0001 O7, ADR-0002 §4):

| Type | Scale | Use |
|---|---|---|
| `Cents` | 10^-2 peso (centavo) | storage and the API |
| `Rate` | 10^-6 | tax rates, coeficientes |
| `Exact` | 10^-6 peso | the tax engine's intermediate amounts |

Invariants:

- **No floats anywhere.** golangci-lint's forbidigo bans them in `domain/` and in `money`
  itself. The package imports only the standard library.
- **Exact products.** `Cents.MulRate`, `Exact.MulRate` and `Exact.MulRatio` multiply through a
  128-bit intermediate (`math/bits`). A result outside `int64` is `ErrOverflow`, never a silent
  wrap. `Add` and `Sub` also return `ErrOverflow`.
- **One rounding rule: half away from zero.** Products round at the 10^-6 scale. `Exact.Round`
  is the single step from `Exact` to `Cents` (ADR-0002 §4) and cannot overflow. Fractions a
  `Rate` cannot hold exactly (the 2/3 IVA withholding) go through `MulRatio(num, den)`, which
  rounds once instead of truncating a decimal approximation.
- `ParseCents` and `ParseRate` parse decimal strings, and `String` renders them back without
  loss.

## clock

The only place that reads the wall clock (`time.Now` is banned in `domain/` by forbidigo).
"Today" and the current tax period are Mexico City calendar facts, so every `Clock` returns times
in `America/Mexico_City` (`clock.ZoneName`). The zone database is embedded (`time/tzdata`), so
the binary does not depend on the host's zoneinfo. Tests inject `clock.Fixed`.

## tx and pgdb

`tx.Manager` is the unit-of-work port. Use cases import `tx`; only `cmd/` and module `infra/`
packages import `pgdb` (archtest rule `usecase-platform-adapter`).

```go
type Manager interface {
    ReadWrite(ctx context.Context, userID string, work Work) error
    ReadOnly(ctx context.Context, userID string, work Work) error
    System(ctx context.Context, work Work) error
}
```

Invariants (RLS context):

- **Every transaction sets the user context first.** `pgdb.TxManager` opens one PostgreSQL
  transaction whose first statement is `SELECT set_config('app.current_user_id', $1, true)`: the
  parameterized, transaction-local form of `SET LOCAL`. The setting dies with the transaction and
  never leaks to the next borrower of the pooled connection.
- **Fail closed.** An empty user id or the nil UUID is `tx.ErrNoUser`, never a transaction
  without context. With no context set, RLS policies show and accept zero rows.
- **No query outside a unit of work.** Repositories get the transaction with `pgdb.Conn(ctx)`.
  Outside a unit of work that call is `pgdb.ErrNoTransaction`.
- **Nesting joins, never switches.** A nested call joins the outer transaction. A different user
  is `ErrUserMismatch`, read-write inside read-only is `ErrReadOnly`, and a nested call through
  another `Manager` (another pool or role) is `ErrForeignTransaction`. `ReadOnly` nested in
  `ReadWrite` is only a hint: never treat it as a security boundary.
- **`System`** runs with the user context explicitly empty. It is for work that has no user yet
  or that spans users only through `SECURITY DEFINER` functions (sign-up, worker purges).
- **Runtime roles are verified at boot.** `pgdb.CheckRuntimeRole` refuses (outside dev) a role
  that is SUPERUSER, BYPASSRLS, an owner or a member of such a role (`ErrPrivilegedRole`). The
  database adds `statement_timeout` (15 s for `cc_app`/`cc_webhook`, 60 s for `cc_worker`).
- `pgdb.Pinger` backs `GET /healthz` readiness with a bounded `SELECT 1`.
- `pgdb.AdvisoryLeader` holds `pg_try_advisory_lock(WorkerLeaderLockKey)` on a reserved
  connection while it leads (see [worker](#worker)).

## flags

Every module's routes sit behind a `module_<name>` flag in `feature_flags`, OFF by default.

Invariants (gating):

- **Fail closed.** `Service` caches the enabled set for `DefaultTTL` (5 s). If the set cannot be
  read, every module is OFF. Each load is bounded by `DefaultLoadTimeout` (2 s) and detached from
  the caller's cancellation, so a hung database still yields a prompt 404. Callers queued behind
  a failed load share its error.
- **The server is the boundary.** `adapter/http/router` wraps each route as the generated server
  registers it, keyed by `gen.RouteModule(pattern)`. A disabled module answers 404 with the
  standard `ErrorResponse` before parameter binding, body decoding or authentication. Unmatched
  paths get the same 404, never 405.
- **One key grammar.** `flags.ValidKey` (`module_[a-z][a-z0-9]*(_[a-z0-9]+)*`) is shared by the
  database CHECK, `cmd/flagmapgen` and the Redocly `x-module` rule.
- **Prerequisites.** `flags.Prerequisites` declares, once, the modules whose data a module's
  routes read (`module_tax_engine`, `module_incomes`, `module_expenses` and `module_dashboard` → `module_taxpayers`; `module_history` → `module_tax_engine`; MODULE-tax-engine §7, MODULE-incomes §6, MODULE-expenses §6, MODULE-dashboard §3, MODULE-history §3). `Service` builds
  each snapshot from the effective set (`EffectiveModules`): a module whose prerequisite is OFF is
  OFF for the route gate (404) and absent from `GET /v1/flags`, from one snapshot. The graph is
  tested to be well-formed and acyclic (`CheckPrerequisites`); `cmd/flags list` shows the
  `REQUIRES` column and a `hidden` state; every seeded preset is prerequisite-closed. New
  dependencies are added to the map, never as a check inside a handler.
- `GET /v1/flags` is UX only. Clients never enforce anything with it.
- `flagsgorm.Reader` runs as `cc_app` (SELECT only). `flagsgorm.Store` runs as `cc_owner` for the
  CLI and applies a preset in one transaction. `Admin` validates keys and preset names.
- `flagtest.OnOff(t, "module_x", suite)` runs a handler suite with the flag ON and OFF.

## access

The AccessPolicy port (P0-16, PRD §7) and its GORM adapter `accessgorm`, which runs as `cc_app`
inside the caller's unit of work, so RLS hides other users' `entitlements` and `user_usage`.

Rules:

- The trial starts at registration and lasts `TrialDuration` (7 days). It ends at exactly
  start + 7 days (the end is exclusive).
- In the trial a user creates at most `TrialRecordCap` (5) incomes and 5 expenses, counted per
  user across all taxpayers. Deleting never frees a slot, and restoring a record never
  reserves one (the slot was spent at creation; MODULE-incomes §2.7). A refused create is `402` with detail `trial_cap_reached` naming the spent
  kind, because the caps are independent.
- With an active `pro` or `comp` entitlement, records are unlimited. With neither and no running
  trial, access is **read-only**: creates and edits get `402 entitlement_required`
  (`ErrReadOnly`), while reads, deletes and restores (Undo) of the user's own records still
  work (contract 0.6.1).
- Every user, trial included, holds at most `TaxpayerCap` (10) taxpayers
  (`422 taxpayer_limit_reached`).
- **`ReserveRecord` is atomic.** One conditional `UPDATE … RETURNING` in the caller's
  transaction bumps the counter, so concurrent creates never overshoot the cap, and a rolled-back
  create gives its slot back.

## apperr and httperr

- `apperr.Error` carries a contract `ErrorCode` (from `docs/api/openapi.yaml`) and a client-safe
  message. The cause stays wrapped for logs and **never reaches the wire**. An unknown error
  becomes `internal_error` with `apperr.InternalMessage`.
- `apperr` has no `net/http` import. Domain packages keep their own sentinel errors, and each
  module's adapter translates them with an `apperr.Mapper`.
- `httperr.Status` maps codes to HTTP statuses. `httperr.Write` writes the **single error
  envelope** (`ErrorResponse`) with the request id. Contract tests guard it against the schema.
  The package is `adapter/http/httperr` (P0-29 FU3, done in WP4). The archtest rule
  `inward-adapter` keeps it, and the rest of the `adapter/` tree, out of every package except
  adapters and `cmd/`.

## crypto

- `PasswordHasher`: argon2id in PHC string format (`DefaultArgon2idParams`, OWASP production
  parameters), constant-time verification, rehash-needed detection, inputs capped at
  `MaxPasswordBytes`.
- `Keyring`: AES-256-GCM for the RFC. Every ciphertext starts with the `KeyID` that sealed it, so
  keys rotate without a flag day.
- `BlindIndex`: HMAC-SHA256 with a pepper (`rfc_hash`, `email_hash`) over a normalised value, for
  equality lookups and unique indexes.
- The package never logs. Its errors never carry plaintext, passwords, keys or peppers, and types
  holding secrets render as `REDACTED` in fmt and slog.

## config, logging, requestid

- `config.Load` reads `APP_ENV`, `PORT`, `LOG_LEVEL`, `RATE_LIMIT_RPS`/`_BURST`,
  `TRUSTED_PROXY_HOPS` and the `DATABASE_URL` secret through a `SecretSource` port. It reports
  every problem at once. `TRUSTED_PROXY_HOPS` is required in staging and production.
  `LoadWorker` is the worker's subset. The table of variables is in
  [backend/README.md](../../backend/README.md#running-the-api).
- `logging`: slog JSON lines carrying the request id. Redaction fails closed in two layers: by
  key (`rfc`, `email`, `password`, `token`, `authorization`, `curp`, …) and by value (RFC, CURP,
  email, bearer tokens, JWTs, URL credentials). Scrubbing is pattern-based, so the rule stays:
  **log identifiers, never payloads**.
- `requestid.With`/`From` carry the correlation id that the error envelope and every log line
  share.

## ratelimit

An in-process GCRA token bucket keyed by client IP (ADR-0001 O9: counters live in the process
until the API scales past one instance). `Allow` is O(1). Idle keys are swept every
`SweepInterval` (1 min), so memory tracks recently active clients. Over the limit the middleware
answers `429` with `Retry-After`.

## worker

The job runner behind `cmd/worker`:

- Only the leader runs jobs. Leadership comes from `pgdb.AdvisoryLeader`. The runner campaigns at
  start and on every heartbeat (10 s), and an elector error counts as lost leadership.
- Each run's context is cancelled on timeout, on lost leadership and on shutdown. A job never
  outlives the leadership it started under. A panic is recovered and logged.
- Two instances may briefly both act as leader between a leader's connection dying and its next
  heartbeat. **Jobs must be idempotent.** Exactly-once comes from the data (unique keys), not from
  the scheduler.

## Test support

- `pgtest`: one `postgres:18` container per test package (the same image digest and role
  bootstrap as `infra/local`), a template database migrated once as `cc_owner`, and one cloned
  database per test. `AsUser`/`WithoutUser` run under a user context or none.
  `AssertIsolation` runs the mandatory RLS checks on a table (ENABLE + FORCE, A cannot read,
  update or delete B's rows, cross-user insert rejected, unset context fails closed) and refuses
  a vacuous proof.
- `flagtest`, `workertest`: flag ON/OFF suites and manual tickers, so no test waits on the wall
  clock.
