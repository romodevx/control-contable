# P0-29: Tech-lead gate of the backend platform foundation

- **Date:** 2026-10-13
- **Scope:** A) PR #27 (`feature/p0-qa-fixes` → `dev`): flag-load timeout, RLS guard tests,
  `make api` / `make worker`, QA report P0-28. B) Holistic review of `backend/` on that
  branch, which contains all of `dev` (P0-03 … P0-28).
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language
  (idiomatic Go wins ties), SOLID, the rules in `~/.claude/rules/`.

## Verdict: **PASS**

Nothing blocks M1. The one blocker I found (the Dependency Rule check missed indirect
imports) is fixed on this branch with a test. Two smaller defects are fixed the same way.
The follow-ups below are not blockers, but M1 must respect the ones marked **M1**.

PR #27: **approve once CI is green.** The flag fix is correct and its tests now fail
deterministically when either half of the fix is reverted.

## Evidence

| Check | Result |
|---|---|
| `make check` (fmt-check, golangci-lint v2, `go test -race`) | green, 0 lint issues |
| `make codegen-check` | no drift |
| `make contract-lint`, `make contract-lint-test` | valid; 16/16 rule cases (1 new) |
| Mutation: revert "share the failed load" in `flags/service.go` | `TestWaitersShareAFailedLoad` fails |
| Mutation: drop `context.WithoutCancel` in `reload` | `TestLoadSurvivesTheCancellation…` fails |
| Mutation: old redocly x-module pattern | new rule case fails (15 pass / 1 fail) |
| `deadcode -test -tags integration ./...` | no unreachable code (outside `gen/`) |
| Gate scan (test/CI relaxations in the PR diff) | none |

Integration tests (testcontainers) were not re-run locally: no change here touches SQL or
migrations. CI runs them.

## A) PR #27

| Item | Assessment |
|---|---|
| `flags.Service` load bound (`DefaultLoadTimeout` 2 s, `WithLoadTimeout`) | Correct. The read is detached from the cancellation of the caller that started it (`context.WithoutCancel`) and bounded by its own timeout, so a waiter never inherits another client's disconnect. A caller that started a load still waits ≤ 2 s after its own disconnect, which is acceptable. 2 s fits well inside `WriteTimeout` (30 s). |
| Shared failure (`loads` / `failed` seq) | Correct. `seen` is read before the semaphore, and the holder bumps `loads` before releasing it, so only callers that queued behind a failed load reuse its error. Callers that arrive later retry. Nit: see FU6. |
| `TestEmailVerifications_DenyAllForEveryRuntimeRoleDespiteAGrant` | Good. It tests the policy layer under a mistaken grant: defence in depth. |
| `TestSchemaPublic_EveryTableHasRowLevelSecurity` | Good. The allow-list documents a reason for each table, and the "not vacuous" subtest proves the sweep can fail. |
| `make api` / `make worker` | Fine. LOCAL-guarded and no password echo. |
| **Defect (fixed, 5cc2736):** `TestWaitersShareAFailedLoad` slept 50 ms and tolerated up to 7 loads. `TestLoadSurvivesTheCancellation…` passed whether or not the second caller was queued (vacuous on a slow scheduler). | Both now use `testing/synctest`: `synctest.Wait` proves the queueing, and both assert exactly one load. Testing rules: "Deterministic only". |

## B) Holistic findings

### Fixed on this branch (failing test first)

| # | Sev | Finding | Fix |
|---|---|---|---|
| F1 | **High** | **archtest did not enforce the Dependency Rule transitively.** The usecase rules deny frameworks by import path and platform adapters by name (`pgdb`, `pgtest`), so a usecase could import `internal/platform/flags/flagsgorm` (GORM), `internal/platform/access/accessgorm` (GORM), `internal/platform/httperr` or anything under the top-level `adapter/http/` tree (net/http), or a neutral helper that imports one of them. depguard checks direct imports only. Clean Architecture, Dependency Rule. | a9aa267: `Check` builds the module's import graph and reports `inner-transitive-framework` for any domain/usecase import whose in-module closure reaches net/http, database/sql, GORM or pgx. Packages that already police themselves (domain, usecase, pure platform) are not blamed twice. O(V+E), memoized. Fixture `testdata/transitive` covers a direct adapter, the top-level adapter tree, a 2-hop chain and a clean negative. |
| F2 | Medium | Two flag tests were non-deterministic (see A). | 5cc2736 |
| F3 | Low | **Three grammars for one flag key.** `redocly.yaml` and `cmd/flagmapgen` accepted `module_[a-z][a-z_]*`, while the `feature_flags` CHECK and `flags.ValidKey` require `module_[a-z][a-z0-9]*(_[a-z0-9]+)*`. `x-module: module_auth_` passed the contract but could never be enabled (silent 404 forever), and `module_m2` was refused by the contract although the database accepts it. DRY: one source of truth. | 3036faa: flagmapgen calls `flags.ValidKey`. The redocly pattern mirrors it, with a new rule case in `tools/contract/test.js`. |

### Follow-ups (not fixed here)

| # | Sev | Finding | Owner / when |
|---|---|---|---|
| FU1 | **Medium (M1)** | **No per-request deadline.** `http.Server.WriteTimeout` does not cancel `r.Context()`. With a hung database (the D1 scenario), every M1 handler doing a `tx.Manager` call blocks until the client leaves. `statement_timeout` (15 s on `cc_app`) does not help when the server never answers. This is the same class as D1, one layer down. Add a request-deadline middleware (≈10 s, under `WriteTimeout` 30 s) in `adapter/http/middleware`, and a test with a hanging `tx.Manager` fake. | Backend, M1, before the first DB-backed handler |
| FU2 | **Medium (M1)** | `cmd/api` `notBuilt` embeds a nil `gen.StrictServerInterface`. A flag turned ON for an unbuilt module panics, and every request logs a stack trace and answers 500 (QA O1), so presets cannot be demoed. Replace it with an explicit stub (404 like a hidden module, or 501 with a contract change). Do not depend on `Recover` for control flow. | Backend + Architect, M1 |
| FU3 | Medium | `internal/platform/httperr` imports `net/http`. It is the HTTP edge of `apperr` and belongs in `adapter/http/httperr`. F1 now stops a usecase from importing it, but the placement still misleads. It is a mechanical move (3 importers). Do it after P0-30 merges to avoid conflicts. | Backend, M1 |
| FU4 | Medium | `internal/platform/access` (+ `accessgorm`) holds business rules (trial caps, entitlements, taxpayer cap). That is a bounded context sitting in the shared kernel, and modules import platform freely, so the cross-module rule never sees it. Decide by ADR: keep it as a documented shared-kernel port (and say so in `archtest/doc.go`), or promote it to `internal/access/{domain,usecase,infra}`. Either way, **no new business logic goes into `platform/`**. | Architect, before M1-07 wires it |
| FU5 | Low | Third-party imports of a usecase are a denylist (`frameworkPrefixes`), and third-party internals are not followed. Once the first module exists, make the usecase third-party rule an allowlist (e.g. `github.com/google/uuid`, `golang.org/x/sync`). | Tech Lead, M1 |
| FU6 | Low | `flags.Service` keeps `failed` after a later successful load. A waiter that slept through a failed load, then a successful load, then an `Invalidate`, returns the old error. It fails closed and the window is tiny. Clear `failed` on success. | Backend, any |
| FU7 | Low | `httperr.Response`/`Body`/`FieldError` re-declare the contract's `ErrorResponse` by hand next to the generated `gen.ErrorResponse`. Contract tests guard the drift, but there are two definitions of one envelope. Prefer the generated types at the edge. | Backend, with FU3 |
| FU8 | Low | `money` has no native fuzz target. Add `FuzzParseCents`/`FuzzParseRate` (round-trip `String`, no panic, `ErrOverflow` beyond range) before the M3 tax engine depends on it. | Backend, before M3 |
| FU9 | Info | `TestDefaultLoadTimeoutIsShort` asserts a constant (a change detector). It is kept on purpose as the documented budget; do not copy the pattern. | n/a |
| QA F3/F4 | Info | Shared compose names across worktrees; misleading `migrations:` prefix in `cmd/flags`. | DevOps (F3), Backend (F4) |

### What I checked and found sound

- **Clean Architecture.** `domain` is an allowlist (stdlib minus frameworks, own domain, `platform/money`). Pure platform obeys the domain rule. Layer rules within a module are allowlists (`usecase` → `domain`/`usecase` only). `cmd/` is the only composition root. The real repo has zero violations, including the new transitive rule.
- **money.** `mulDiv` holds the product in 128 bits (`bits.Mul64`/`Div64`). It guards the `Div64` panic (`hi >= d`) and the round-up wrap past 2^63 *before* incrementing. Its rounding is half away from zero via `r >= d-r` (no overflow), and it handles `MinInt64` exactly. `Exact.Round` cannot overflow (truncating division first). `parseFixed` is strict (`-?d+(.d+)?`), never rounds on input (`ErrPrecision`) and is overflow-checked per digit. Tests are table-driven, with a seeded cross-check against `math/big` and benchmarks. No floats; forbidigo bans them and `time.Now` in money/domain.
- **Idiomatic Go.** Errors wrap with `%w`, use sentinels in `tx`/`money`/`flags`, and match through `errors.Is`/`errors.AsType`. `context.Context` is the first parameter everywhere and propagates. Ports are small (`flags.Reader` 1 method, `tx.Manager` 3) and defined by the consumer side. Options use functional options. Goroutines are owned (`errgroup` in the worker runner; `goleak` in tests).
- **Unit of work (`pgdb.TxManager`).** `SET LOCAL` through `set_config(..., true)` with a bind parameter. Nested calls join and never widen (other user, RW-in-RO and a foreign manager are each a typed error). Panic, error and cancellation roll back. `Conn` outside a unit of work is `ErrNoTransaction`, so there is no path that queries without RLS context.
- **Package naming/cohesion.** Short, lower-case, no `util`/`common`. Test helpers follow the `httptest` convention (`pgtest`, `flagtest`, `workertest`). The exceptions are FU3 and FU4.
- **Tests.** Table-driven throughout. Integration tests use real PostgreSQL through testcontainers, with RLS isolation per table, the never-set and reverted session states, and a catalog sweep. `flags/service_test` was the only sleep-based test (fixed). The only randomness is seeded.

## What M1 must respect

1. **Layering is now enforced transitively.** A usecase depends on `tx.Manager` and its own consumer-side ports, never on `*gorm`/`pgdb`/`flagsgorm`/`accessgorm`/`adapter/http/...`, not even through a helper. Repositories live in `internal/<module>/infra` and call `pgdb.Conn(ctx)` at the top of every method.
2. **Every DB call has a deadline the client does not control** (FU1). Detach background work with `context.WithoutCancel` plus a timeout, as `flags.reload` does. Keep every budget under `WriteTimeout`.
3. **Every new table: ENABLE + FORCE RLS in its first migration.** The catalog sweep now fails otherwise. An `unforcedTables` entry needs the security gate. Each user-scoped table also gets `pgtest.AssertIsolation`.
4. **Errors.** Domain sentinels map to contract codes through one `apperr.Mapper` per module adapter. `apperr.Error.Message` never carries PII. The wire envelope only comes from `httperr.Write`.
5. **Money.** Use `money.Cents` at storage/API and `money.Exact` inside calculations, with one `Round` at the end. Never use floats or `strconv.ParseFloat`.
6. **Flags.** `x-module` keys follow `flags.ValidKey`. New module routes are registered via the generated flag map. Replace `notBuilt` before the first preset demo (FU2).
7. **TDD and determinism.** Write the failing test first. Concurrency tests use `testing/synctest`, never `time.Sleep`. Use table-driven tests and testify.
8. **No business logic in `internal/platform`** until the FU4 ADR decides where `access` lives.
