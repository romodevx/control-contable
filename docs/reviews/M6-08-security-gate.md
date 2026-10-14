# M6-08 · Security gate: M6 Dashboard

| Field | Value |
|---|---|
| Scope | M6 as built on `origin/dev` @ `7789182` (2026-10-14, after PRs #110, #111, #113, #114). Backend: `internal/dashboard/**` (domain window and status rules, use case, ports, handler, presenter, `dashboard_reads` budget), `cmd/api/dashboard.go` (adapters over the tax engine's `EstimateRange`, the incomes/expenses `LedgerReader` aggregates and `HasLive`, the taxpayers profile and `access.Policy.Summary`), and the router-wide undeclared-query-parameter rejection added in #114. Mobile: `lib/features/dashboard/**`. PR #116 (M6-05 flag tests, M6-06 QA, M6-07 gate fixes G1/G2) is **open** and not in the build; it adds tests and mobile fixes only, no backend behavior |
| Reviewer | security |
| Date | 2026-10-14 |
| Method | `/cso --code --diff --base 1cad880` (gstack-cso, run `1792018158267-a47c62b6dd851503`, daily static; helper-mediated reads of `router/validate.go`, `router/router.go`, `router/errors.go`, `flaggate`, the incomes ledger reader, `money/arith.go`, `tx/tx.go`). Then a manual review of the rest of the scope and STRIDE; `go test -race` of `adapter/http/...`, `internal/dashboard/...`, `cmd/api`; `go test -tags=integration` of the dashboard `cmd/api` suites; `govulncheck -test ./...`; `golangci-lint`; and an end-to-end probe of the API built from this branch against an **isolated local** postgres (compose project `cc-gate-m6`, container `cc-gate-m6-postgres`, volume `cc-gate-m6-pgdata`, checked with `docker compose config`; DB on 5449, API on 3049). The stack and its volume were destroyed afterwards. The shared local stack (`control-contable-postgres`, 5434) and other agents' processes were not touched |
| Verdict | **PASS.** One Low finding (G-1), fixed on this branch with a failing test first. No finding of severity medium or above. `module_dashboard` stays **OFF** in every shared environment until the preconditions in §5 hold |

## 1. Evidence

Probe totals: 41 of 42 checks passed on the first run; the miss was the probe's own expectation
(the 429 came at call 598 instead of 601 because the same user had spent calls earlier in the
probe; the budget refills one call every 6 s, re-checked below).

| Check | Result |
|---|---|
| Unit (`-race`) | `adapter/http/...`, `internal/dashboard/...`, `cmd/api`: green |
| Integration | `TestContract_dashboard`, `…FlagMatrix`, `…EmptyState`, `…StatementCount`, `TestDashboardRoute_needsTheTaxpayersFlagToo` and the adapter tests: green |
| **IDOR, 404 parity** (7 pairs) | user B with A's taxpayer id vs a random uuid, with no query, `period=2026-10`, `period=2026-01&months=6`, `months=12` (all `404`), `period=2099-01` (`400`), `period=2025-12` (`422`), `months=5` (`400`). Every pair: same status, bodies identical minus `request_id`. Period and months are checked before any read, the same for every id, so they reveal nothing. The foreign `404` equals the `module_dashboard` OFF `404` byte for byte (minus `request_id`) |
| **Malformed id** | `not-a-uuid`, `123`, `<uuid>x`, `%00` → `400 validation_failed` (router, field `taxpayerId`). An upper-case uuid is the same taxpayer (`200`, normalized in the body). The nil uuid → `404` |
| **Undeclared / tampered query** | `foo`, `taxpayer_id=<B's id>`, `user_id`, `access=pro` → `400`, code `additionalProperties`; duplicate `months`/`period` → `400`; `period=2026-1`, `+2026-10`, `months=abc`, `6.0`, `13` → `400`. With the module OFF every one of these is the hidden `404` (flag gate first) |
| **Read-only transaction** | `log_statement=all` on the gate DB for one request: `begin read only` → `set_config('app.current_user_id', $1, true)` → taxpayer, regimes, two ledger aggregates, tax inputs, regime texts, entitlements, usage (8 SELECTs) → `commit`. The rate-limit take runs before and outside it. No write statement |
| **Budget** | `dashboard_reads` 600/h burst 600: after the burst → `429 rate_limited` with `Retry-After`, `no-store`; refill one call per 6 s (observed). Another user still `200`; `getTaxpayerEstimate` still `200` (the range estimate inside the dashboard does not spend `estimate_reads`). The limiter runs before validation, so invalid requests also spend the budget. The edge per-IP limiter was raised on the gate API only (`RATE_LIMIT_RPS/BURST=1000`) |
| **`Cache-Control: no-store`** | on the `200` and on every error status probed (`400`, `404`, `422`, `429`) |
| **No amounts or PII in logs** (`LOG_LEVEL=debug`) | markers: the income and expense amounts, their IVA, monthly sums, the ISR figure, the description marker, both RFCs, both emails, `Bearer`/`eyJ`: **0 hits** in the API log. The dashboard logs nothing on success or on `4xx`; a `500` logs the error chain, whose dashboard and engine parts carry periods, counts and sentinels, never amounts (`money.ErrOverflow` has no operands) |
| **Flag matrix** (toggled with `cmd/flags`, 6 s TTL) | engine OFF → `estimate: null`, every `trend[].tax: null`, no `legend`; incomes OFF → `month.income_cents` and every `trend[].income_cents` `null`; expenses OFF → same for expenses; all three OFF → no figure, `is_empty: false`. `module_dashboard` OFF → owner `404` = foreign `404`, trailing-slash variant `404` with no `Location`, undeclared query / malformed id / anonymous all `404`, absent from `/v1/flags`. `module_taxpayers` OFF → the same generic `404`, absent from `/v1/flags` |
| **No leak through null vs zero** | user B holds one income and no expense. With incomes and the engine OFF, B's Home answers `is_empty: true` and no income figure anywhere, exactly as for a user with no records: the existence of the hidden module's rows does not leak. With the engine ON the tax figures still derive from incomes (by design, MODULE-incomes OQ-8; see G-3) |
| **AccessSummary** | derived from the principal's user id and the server clock only. Headers `X-User-Id`, `X-Access`, `X-Forwarded-For`, a future `Date`, and query `access=pro` (refused, `400`) change nothing. A trial account showed `trial_days_left: 7`, `trial_ends_at` = sign-up + 168 h, and the remaining counters |
| **DoS bound, 100k rows** | 50,000 incomes + 50,000 expenses on one taxpayer (inserted as LOCAL superuser past the trial cap). Sequential 200 reads: p50 140 ms, p95 220 ms (one cold first call at 5.7 s); 8 concurrent × 50: p50 273 ms, p95 433 ms. Body 4.8 kB, 10 points. `EXPLAIN ANALYZE` of the income aggregate as `cc_app` under RLS: bitmap index scan on `incomes_taxpayer_regime_idx`, 54 ms. Cost per request: 8 statements regardless of `months`, aggregates in SQL, at most 12 months × held regimes rows back; the range is bounded to [January of the window's first fiscal year, period ≤ current month] |
| **Anonymous** | `401` with the bearer challenge |
| **Dependencies** | `govulncheck -test ./...`: 0 vulnerabilities in called code (1 in a required module, not reachable). CI `govulncheck` is the required check on this PR |
| Mobile | no `print`/`log`, `SharedPreferences`, secure storage or DB in `features/dashboard`; `dashboardProvider` is `autoDispose`, keyed by taxpayer id, no automatic retry (no `429` amplification). Sections are hidden from the server's `null`s, never from a client flag alone |

## 2. Threat model (STRIDE), M6 as built

Actors: authenticated user; another authenticated user; a read-only user (trial ended); an
anonymous caller; an operator with the `cc_app` credential.
Assets: the income/expense aggregates and the tax figures of one taxpayer, the access summary
(trial dates and counters), and the inventory of flag-hidden routes.
Entry points: `getDashboard` (bearer, `module_dashboard`, prerequisite `module_taxpayers`); the
router's query and body validation, which now applies to every operation; the in-process
`EstimateRange`.

| | Threat | Control as built | Residual |
|---|---|---|---|
| **S** | Read another user's Home with a guessed/stolen taxpayer id | user id only from the verified token; one read-only unit of work with `app.current_user_id`; ownership read first (profile through the engine, or `CheckTaxpayer` when the engine is OFF); FORCE RLS on every table read; `hidden()` rewords any port's `404` as the module's (§1) | The ledger `HasLive`/sums and the profile read filter by taxpayer id and rely on RLS for the owner (G-4) |
| **S** | Influence the access banner or the default period | `AccessSummary` from `access.Policy.Summary(userID)` and the server clock; `trial_days_left` server-side (OQ-D6); no client-supplied time or level is read (§1) | — |
| **T** | Write through the dashboard | none: `Tx.ReadOnly`, only SELECTs (§1). The dashboard is never nested in a read-write unit of work | — |
| **R** | Deny having read figures | reads are not audited (as every read route) | `429` refusals are not logged either (same for every per-user budget). Info |
| **I** | Figures of a module that is OFF | figures are `nil` and their own aggregate is skipped; `is_empty` counts only enabled modules; the estimator is not called with the engine OFF (§1) | With the engine ON and a record module OFF, the estimate still reflects that module's rows (G-3, by design) |
| **I** | Existence oracle through ordering or status | flag gate → authn → budget → validation → id/period/months → ownership; period and months are refused before any read for every id; malformed id is `400` for every id (§1) | — |
| **I** | Amounts or RFC in logs, caches, responses | no logging on success/4xx; error chains without amounts; `no-store` everywhere; the body carries `taxpayer_id` and figures only, no RFC | — |
| **D** | Burn DB/CPU with dashboard reads | 8 statements per request independent of `months`; aggregates in SQL over indexed ranges; window ≤ 12 months, period ≤ current month; 600/h per user; edge per-IP limiter (§1) | Aggregate cost grows linearly with a taxpayer's own rows (54 ms at 50k); acceptable, re-measured by k6 in REL-06 |
| **D** | Turn validation errors into an amplifier | **G-1, fixed**: details capped at 20 | — |
| **E** | Dashboard reachable with its parent OFF | `flags.Prerequisites`: `module_dashboard → module_taxpayers`, verified on the binary (§1). The engine is not a prerequisite; it degrades Home (OQ-D7) | — |

## 3. Findings of this gate

| ID | Severity | Item | Status |
|---|---|---|---|
| G-1 | Low | **Validation errors were an unbounded amplifier.** `router.invalid` copied every field error into the envelope, and `undeclaredQuery` (#114) adds one per undeclared query parameter. A 58,889-byte query on the **public** `GET /v1/flags` answered 839,031 bytes (14×); a 508,941-byte body of unknown properties on the public `POST /v1/auth/register` answered 3,429,117 bytes. Bounded by the 64 KiB header cap, the 1 MiB body cap and the edge per-IP limiter, so Low. Also raised as M6-07 FU1(c) | **Fixed** on this branch: `TestValidation_capsTheDetailsList` (red first: 2,000 details), then `router.MaxFieldErrors = 20` in `invalid`, which covers both the query rule and body validation. Re-probed on the stack: the same requests answer 1,811 and 1,846 bytes, 20 details each |
| G-2 | Info | Undeclared query **names** are echoed in `details[].field` (JSON-encoded; never the value). A name is bounded by the 64 KiB header cap and now by the count cap. Not a value echo; `TestValidation_neverEchoesTheRejectedValue` still holds | Accepted |
| G-3 | Info | With `module_tax_engine` ON and `module_incomes` (or expenses) OFF, `estimate` still carries figures derived from that module's rows (e.g. `iva_components.charged_collected_cents`). This is the engine's documented behavior (MODULE-incomes OQ-8), the same as `getTaxpayerEstimate`, and only the caller's own data | Accepted by design |
| G-4 | Info | The incomes/expenses `HasLive` and sums filter by taxpayer id and rely on RLS (and the ownership read that precedes them) for the owner, like M5's G-3. Correct while every taxpayer has one owner | Carry to any delegated-access ADR |
| G-5 | Info | The trend window, statement count and indexes keep one request bounded, but the first call after start-up took 5.7 s at 100k rows (cold cache and plans); later calls stayed under 450 ms at p95 under concurrency | REL-06 k6 on staging |

## 4. Consolidated M6 follow-ups

Sources: PR reviews #110 (none), #111, #113, #114, #116 (M6-07 tech-lead gate and M6-06 QA report),
MODULE-dashboard §8 and this gate.

### Resolved

| Source | Item | Evidence |
|---|---|---|
| MODULE §8 | Ownership under RLS, foreign id `404` indistinguishable from OFF | §1 IDOR parity; `TestContract_dashboardFlagMatrix` |
| MODULE §8 | Engine OFF ⇒ no tax computation | §1 flag matrix; the estimator is not called (unit test) |
| MODULE §2.5 | Own `dashboard_reads` 600/h | §1 budget; `TestUserLimits` |
| MODULE §2.4, #114 | ≤ 8 statements, read-only | `TestContract_dashboardStatementCount`; §1 statement log |
| #111 blocker | Malformed id is `400`, not `404` | 7507dcc; §1 |
| #113 blocker | IVA/ISR colour pair ΔE | 9fc0210 |
| #116 G1, G2 | Chart bar tap action for screen readers; stale figures after a failed refresh | fixed on the #116 branch (open) |
| M6-07 FU1(c), G-1 | Cap the echoed details | this PR |
| M5 G-2, #102 FU | Validate the estimator's deps at boot | `NewService`/`newDashboardModule` refuse a missing port (M6-07 SOLID row) |

### Deferred, with owner

| Source | Item | Owner | Target |
|---|---|---|---|
| M6-07 FU1(a, b) | State the undeclared-query rule (and now the 20-detail cap) in the shared `ValidationFailed` response or `info.description`; align MODULE-tax-engine §5 (`invalid` → `additionalProperties`) | architect | next contract re-freeze |
| #111 FU1 | One `NullableNonNegativeCents` schema instead of four inline copies | architect | next additive contract bump |
| #111 FU2 | `engineOn` example not consistent with its own trend | architect | next contract re-freeze |
| #111 FU3 | Other `*NotFound` descriptions still say "malformed id ⇒ 404" | architect | next contract re-freeze |
| QA D-3 | `TrendTax` has no saldo-a-favor amount (US-D3) | architect | 0.8.x additive |
| #113 nits | `trialDaysLeft ?? 1`; hero disappearing when `legend` is null; `$0.00` for `tax == null` in the table; 200 % test for table and income/expense views | frontend-sr | M6 UX polish |
| M6-07 FU2 | ICU plural for "últimos 1 meses" | frontend-sr | M8 |
| M6-07 FU3 | `ref.watch` instead of `ref.read` in `dashboardProvider` | frontend-sr | next touch of `features/dashboard` |
| M6-07 FU4 | Read `Clock.Now()` once in `Service.Get`; `trendTax` should fail on an empty `IVAOutcome` instead of mapping it to `zero` | backend-sr | next touch of `dashboard/usecase` |
| M6-07 FU5 | MODULE §6 wording for the table; PLAN M6-02..M6-04 ticks | doc-writer | M6-09 |
| M6-07 FU6 | Manual TalkBack / VoiceOver pass on real devices | owner + qa-sr | before enabling in a shared environment (P-5) |
| M6-07 FU7 | Repo-wide guard for button nodes without a tap action | frontend-sr | M6 hardening |
| QA D-2, D-4, obs. | Copy and scroll polish | owner / pm / frontend-sr | UX polish |
| G-4 | Owner check inside the ledger and profile queries if delegated access is designed | architect | ADR for any delegated-access feature |
| G-5 | Dashboard p95 on staging (cold start included) | qa-sr | REL-06 |

## 5. Preconditions to enable `module_dashboard` in staging or production

All must hold; each is verified by the person flipping the flag and recorded in the deploy PR.

1. **P-1 This PR merged** (G-1 cap) and **PR #116 merged** (G1 screen-reader tap action and G2
   stale figures on a failed refresh), with CI green.
2. **P-2 Parent and data modules**: `module_taxpayers` preconditions (M2-14 §5). If
   `module_tax_engine` is ON in that environment, its own preconditions (M5-31 §5, including the
   CPA sign-off before production): the dashboard's hero and trend are the engine's figures.
   If `module_incomes`/`module_expenses` are ON, their preconditions (M3-M4 §5).
3. **P-3 Gates**: M6-07 (tech-lead) PASS, this gate PASS; REL-09 (deploy security gate) PASS.
4. **P-4 Limiter healthy**: per-user budgets fail closed, so a limiter outage turns Home into
   `429`; alert on `rate limit unavailable`.
5. **P-5 Accessibility**: the manual TalkBack/VoiceOver pass (M6-07 FU6) is done.
6. **P-6 `APP_ENV` is not `dev`** (the dev email adapter logs OTP codes).
7. **P-7 Mobile release build** passes `mobile-artifact-check` (no source maps, no internal URLs)
   and uses an HTTPS API base URL.
8. **P-8 Performance on staging**: REL-06 k6 confirms p95 < 1 s for the dashboard.

## 6. Notes

- The cso run is recorded as **partial**: its 600 s daily budget covered the snapshot and the
  helper-mediated reads in the Method row; the submission (one Low finding and the coverage
  records) arrived after the deadline and was refused, so the helper holds no finding or coverage
  record. The rest of the review was manual and runtime, as listed above.
- The isolated stack used its own compose project, container name **and** volume name (checked
  with `docker compose config`). Only the API processes started for this gate were stopped, by
  PID. The stack and volume were removed.
- Test secrets were generated at runtime into a mode-0600 file outside the repository, deleted
  afterwards and never committed. Bulk rows were inserted as the LOCAL superuser into the gate DB
  only.
