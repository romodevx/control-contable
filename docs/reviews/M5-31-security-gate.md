# M5-31 · Security gate: M5 Tax Engine

| Field | Value |
|---|---|
| Scope | M5 as built on `origin/dev` @ `3ee75d2` (2026-10-14, after PR #105). `backend/internal/taxengine/**`: the embedded parameters (`params/`, `domain/taxparams`), the strategies and registry, the estimate use case (`usecase/estimate.go`, `breakdown.go`), the readers (`infra/taxinputs`, the incomes/expenses ledger readers as wired in `cmd/api/ledger*.go`), the tax-inputs use case and handlers, the estimate and breakdown handlers, the per-user budgets (`adapter/http/limits.go`) and `cmd/api/taxengine*.go`. Mobile `lib/features/tax_engine/**`, including the prior-payment suggestion ("estimado por la app") and the tax-inputs form. Open PRs #107 (contract 0.7.0, no-store on errors) and #108 (regime start month) are **not** in the build |
| Reviewer | security |
| Date | 2026-10-14 |
| Method | `/cso --code` (gstack-cso 3.0.0, run `1792006990883-6a5151198e2dc996`, daily static; helper-mediated read of the handlers, limits, presenter, estimate and tax-inputs use cases, input validation, the `taxinputs` store, the params loader, the wiring, the contract's `TaxInputsUpdate`/`TaxInputValue`, and the mobile repository and suggestion). Then a manual review of the rest of the scope; `go test -race` of `internal/taxengine/...` and `cmd/api`; `go test -tags=integration` of `internal/taxengine/...`, the M5 `cmd/api` suites and `migrations`; and an end-to-end probe of the API built from `3ee75d2` against an **isolated local** postgres (compose project `cc-gate-m5`, own container `cc-gate-m5-postgres` and volume `cc-gate-m5-pgdata`, checked with `docker compose config`; DB on 5448, API on 3048). The stack and its volume were destroyed afterwards. The shared local stack (`control-contable-postgres`, 5434) and other agents' processes were not touched |
| Verdict | **PASS.** No supported finding of severity medium or above. `module_tax_engine` stays **OFF** in every shared environment until the preconditions in §5 hold, including the **CPA sign-off (M5-33)** |

## 1. Evidence

Probe totals: 67 of 69 checks passed. The two misses are one known gap (no `Cache-Control` on
error envelopes, G-1, fixed by PR #107, still open) and one probe bug (a read-only GET
asserted a value an earlier step had reset; the call itself answered `200`).

| Check | Result |
|---|---|
| Unit (`-race`) | `internal/taxengine/...` (9 packages, incl. golden gate, property suite, no-float guard, `EngineVersion` pins), `cmd/api`: green |
| Integration | `internal/taxengine/...` (incl. `infra/taxinputs` store and RLS), M5 `cmd/api` suites (contract, estimate, golden API, preset smoke), `migrations`: green |
| **IDOR, 404 parity** (12 pairs) | user B with A's taxpayer id vs a random uuid on: estimate (with period, default period, future period, 2025), breakdown (`regime=612`, `606`, `999`), tax-inputs GET (612, 606, fiscal year 2025) and PUT (612, 606). Every pair: same status, bodies identical minus `request_id`. A foreign taxpayer answers `404` before any regime or year check, so `regimeCode`, `regime` and `fiscal_year` reveal nothing. The future period (`400`) and the year before the parameters (`422`) are refused before any lookup, for any id. A's stored inputs were byte-identical after B's PUTs. Malformed id → `400` (router) |
| **Read-only transaction** | `log_statement=all` on the gate DB during one estimate and one breakdown: each opened `begin read only`, set `app.current_user_id` with `set_config(..., true)`, ran only SELECTs and committed. The rate-limit take runs outside that transaction. `Inputs.Get` uses `Tx.ReadOnly` too; only `Inputs.Put` is read-write |
| **PUT validation bounds** | all `400 validation_failed`: 129 values (`maxItems: 128`), negative cents, cents above `99999999999999`, month 0 and 13, money_monthly without a month, unknown key, duplicate (key, month), rate on a money key, extra property on a value, extra top-level property (`taxpayer_id`), 65-char key, SQL-looking key, string cents, float cents, missing `values`, future fiscal year, malformed JSON, a 20,001-item body. Nine months at the maximum amount store `200`, and the estimate over them answers `200` (no overflow to `500`) |
| **Read-only user `402`** (trial moved to the past, LOCAL superuser) | PUT → `402 entitlement_required`; PUT on **another user's** taxpayer → also `402` (the access check runs before the taxpayer lookup, so it is no oracle); GET tax-inputs `200`; estimate `200` |
| **Budgets** | `tax_input_writes` 60/h per user: the PUT after the budget (incl. the user's earlier PUTs) → `429` with `Retry-After`; GET tax-inputs still `200` (edge limiter only). `estimate_reads` 600/h shared by estimate and breakdown: alternating calls → `429` with `Retry-After` at ~600; another user's estimate still `200`. The edge per-IP limiter was raised on the gate API only (`RATE_LIMIT_RPS/BURST=1000`) so it would not mask the per-user budgets |
| **Estimate cost (DoS)** | One estimate = fixed number of queries: taxpayer profile, one `GROUP BY regime, month` per ledger over the fiscal year's date range (`incomes`/`expenses` live partial indexes), one tax-inputs read, one regime-text read. Rows returned are bounded by 12 months × held regimes (≤ 16), independent of how many records exist; the compute is O(regimes × months) in memory. The range is one fiscal year, January to the period, and the period must be between 2026-01 and the current month. Some 660 estimate/breakdown calls ran in about 1.6 s on the gate stack. The 600/h budget bounds the rest |
| **No PII in logs** (`LOG_LEVEL=debug`) | markers: income amount, IVA amount, description, both RFCs, a stored input amount, the maximum amount, the test emails: **0 hits** in the API log. `tax_inputs.updated` carries `user_id`, `taxpayer_id`, `regime_code`, `fiscal_year`, `request_id` only (S-8). Error bodies carry codes and fixed es-MX messages, never amounts. The estimate's profile query reads `rfc_masked` only |
| **`Cache-Control`** | `no-store` on every 200 of the four operations (estimate, breakdown, tax-inputs GET and PUT). **Absent on error envelopes** (G-1). PR #107 (contract 0.7.0, "no-store everywhere") is **still open**, so it was not verified; recheck after it merges (P-2) |
| **Params tampering** | `params.Embedded()` reads only the `go:embed` FS: no env var, path or request input reaches the loader (no `os` file access in `internal/taxengine`). Strict YAML (`KnownFields`, one document), path/header agreement, `taxparams.NewParams` validation, `NewCatalog` (validity overlaps); any problem fails `newTaxEngineModule`, which fails `main` before the server listens. The strategy registry is built once, immutable, duplicates refused. `TestLoadRejects` / `TestLoadReportsEveryBadFileWithItsPath` cover corrupt files (S-5) |
| **Flag prerequisite** | `module_taxpayers` OFF, `module_tax_engine` ON: all four operations → generic `404 not_found` (no taxpayer wording), the trailing-slash variant → `404` with no `Location` (no ServeMux redirect oracle), and the PUT wrote nothing (`0` rows with the probe's amount). `cmd/flags` reports "module_tax_engine is hidden until module_taxpayers is ON". `module_tax_engine` OFF: same five `404`s |
| **Anonymous** | `401` with the bearer challenge |
| Mobile | no `SharedPreferences`, secure storage, database or `print`/`developer.log` in `features/tax_engine`; every provider is `autoDispose` and keyed by taxpayer and period/year. The prior-payment suggestion ("estimado por la app") is shown only for empty months on screen and is filled only when the user taps *Usar*; it never writes by itself. Rates stay strings; money is integer centavos. All server text, no regime-code switches |

`govulncheck` is not installed on this machine; the CI `govulncheck` job (required check) is
the evidence for dependencies on this PR.

## 2. Threat model (STRIDE), M5 as built

Actors: authenticated user; another authenticated user (curious or hostile); a read-only user
(trial ended); an anonymous caller; an operator or attacker with the `cc_app` credential or a
backup; a contributor who edits the parameter files.
Assets: the income/expense aggregates the estimate reads, the tax inputs (prior payments,
losses, PTU, coefficient), the computed amounts and breakdown, the embedded 2026 parameters
and the strategy code that turn records into "what you owe", and the user's trust that the
figure is right.
Entry points: 4 operations (`getTaxpayerEstimate`, `getEstimateBreakdown`, `getTaxInputs`,
`putTaxInputs`; bearer, flag-gated, prerequisite `module_taxpayers`), the `Estimator` that M6,
M8 and M9 will call in-process, the `cc_app` role on `taxpayer_tax_inputs`, the parameter files
in the repository, the device session.

| | Threat | Control as built | Residual |
|---|---|---|---|
| **S** | Read another user's estimate or inputs with a guessed/stolen taxpayer id | user id only from the verified token; every read in the caller's RLS transaction (`set_config('app.current_user_id', …, true)`), FORCE RLS on `taxpayers`, `taxpayer_regimes`, `taxpayer_tax_inputs`, `incomes`, `expenses`; 404 parity on every path and query param (§1) | The `taxinputs` store's `Taxpayer`/`Regime` queries filter by id only and rely on RLS for the owner (G-3) |
| **T** | Write inputs on a foreign taxpayer or a regime not held | composite FK `(user_id, taxpayer_id, regime_code) → taxpayer_regimes`; `user_id` from the session; taxpayer then regime row locked (`FOR SHARE`, `FOR NO KEY UPDATE`); regime window and fiscal-year floor checked under the lock; kind derived by the `SECURITY INVOKER` trigger, never written | — |
| **T** | Inflate or poison inputs (negative, overflow, wrong kind, unknown key) to skew the estimate or crash it | contract bounds + `additionalProperties: false`; `CheckSubmittedInputs` against the regime's catalog definitions; `NewTaxInputs` re-validates on read; maximum legal amounts compute without overflow (§1) | Own figures only: a user can still enter wrong prior payments, which lowers their own estimate. Product risk, not a boundary |
| **T** | Change tax rates or tariffs at runtime | parameters compiled in (`go:embed`), validated at boot, fail-fast; no admin or upload path; the golden gate and `EngineVersion` digest pin the figures | A malicious or mistaken parameter change in a PR is caught by CODEOWNERS review, the golden suite and the 2026 transcription tests, not by the runtime (G-5) |
| **T** | Present unverified tax figures as authoritative | `legend` and disclaimers on every estimate, `needs_cpa_review` per regime and line, provisional-parameters notice; ⚠️ golden cases skipped until CPA sign-off | **Product risk**: the CPA items in [CPA-REVIEW](CPA-REVIEW.md) are not confirmed (P-1) |
| **R** | Deny having changed the inputs | `tax_inputs.updated` audit event after commit, ids only | Refused writes (`400`/`402`/`404`) are not audited, as in M3/M4. Retention is a deploy concern (P-7) |
| **I** | Amounts, inputs or RFC in logs, errors, caches | ids-only audit; fixed messages; pg errors never reach the wire; `no-store` on every 200 | Error envelopes lack `no-store` until PR #107 (G-1); they carry no amounts |
| **I** | Existence oracle through ordering of checks | id parse → period/year format → future refusal → parameters floor run before any DB read and are the same for every id; the read-only `402` runs before the taxpayer lookup | — |
| **D** | Burn CPU/DB with estimates | fixed query count, aggregates in SQL, one fiscal year, period bounded to `2026-01`…current month; 600/h per user shared with the breakdown; edge per-IP limiter | The mobile suggestion spends up to 6 estimate calls per tax-inputs half opened (G-4) |
| **D** | Flood input writes | 60/h per user, fail-closed limiter, `maxItems: 128` | — |
| **E** | Read-only user keeps writing inputs | `EnsureWritable` first in `Put` (`402`); reads stay open by design | — |
| **E** | Tax engine reachable while its parent module is OFF | `flags.Prerequisites` declares `module_taxpayers`; verified on the binary (§1) | Reads incomes/expenses even when those flags are OFF (MODULE-incomes OQ-8, own data only). Info |

## 3. Findings of this gate

No supported finding of severity medium or above (cso run: 0 supported findings; manual review
and e2e: none). No code change was needed, so this PR carries documentation only.

| ID | Severity | Item |
|---|---|---|
| G-1 | Low | S-4 asks `no-store` on **every** M5 response. Bodies have it; error envelopes do not (same as M3/M4 G-1, M1-17 G-3). PR #107 adds it everywhere and is still open. Error bodies carry no amounts (verified) |
| G-2 | Low | `EstimateDeps` has no nil check (PR #102 FU): a missing port panics at the first request (recovered as `500`) instead of failing the boot. Wiring is a single composition root and the boot smoke covers it, so nothing ships broken today; M6/M8/M9 will add callers |
| G-3 | Info | `infra/taxinputs` reads `taxpayers` and `taxpayer_regimes` by id only, bounded by RLS alone (the estimate's profile reader is the same). Correct while every taxpayer has a single owner. If shared access (an accountant role, delegated users) is ever added, these reads become horizontal access unless the owner check moves into the query |
| G-4 | Info | The suggestion line reads one estimate per empty month on screen (≤ 6 per half), all from the 600/h budget. The `limits.go` comment ("a user going through all twelve months … spends 48") does not count them. Worst case about 100 form openings an hour before the estimate screen answers `429` |
| G-5 | Info | The boot validates each parameter set's shape, but only the 2026 test (`TestEmbedded2026DefinesEveryKeyTheModulesRead`) checks that every key the strategies read is present. A 2027 file that leaves out a key boots, and the estimates that need the key fail closed (`500`) or the rate-catalog lookups report "not in force" (PR #77 should-fix 1) |

## 4. Carried M5 follow-ups

Sources: MODULE-tax-engine §6 (S-1 … S-8), the reviews on PRs #70, #72, #73, #76–#81,
#86–#90, #94, #95, #98, #100–#102 and #105, and the M5-30 tech-lead gate.

### Resolved

| Source | Item | Evidence |
|---|---|---|
| S-1, #80 | `taxpayer_tax_inputs` ENABLE + FORCE RLS, per-op policies, composite FK, cascade; kind trigger fails closed | migration and store integration tests; e2e IDOR (§1) |
| S-2 | Estimate reads in the caller's RLS transaction | e2e IDOR + statement log (§1) |
| S-3, S-8 | No amounts/inputs/RFC in logs; ids-only audit | e2e log scan (§1) |
| S-4 (bodies) | `no-store` on every 200 | e2e headers (§1) |
| S-5 | Params embedded, validated at boot, fail-fast | `params` load tests; wiring (§1) |
| S-6 | Strategies pure, no floats | archtest + `TestNoFloatingPointInMoneyCode` (M5-30 G1) |
| S-7 | Path ids first; `additionalProperties: false` | contract tests; e2e validation matrix (§1) |
| #70 FU1, FU3 | One `fiscal.Date`; boot calls `Embedded()` only after M5-04 | `cb81a9a`; M5-04 merged |
| #70 FU2 | Tariff continuity checks | `tariff_continuity_2026_test.go` |
| #72 FU | Breakdown disclaimers rule in the M5-23 contract tests | `TestContract_estimates` (#102) |
| #78 must-carry 1, 2 | `incomplete_inputs` whenever a regime is incomplete; "Falta" instead of $0.00 for missing inputs | `domain/disclaimers.go`; `missing_inputs_block.dart` |
| #78 nit | Typed-nil strategy panics at startup | Accepted: startup is fail-fast |
| #79 FU | 625 definitive option zeroes creditable IVA; `iva_mixed_activity` across regimes; M3 rejects exempt/0 % with IVA | `strategy_625_pf.go`; `consolidate.go`; M3 domain |
| #81 | Skip/pending ratchet fails on growth and shrink | golden ratchet (#81, M5-30) |
| #94 FU | 625 direct-collection IVA at 8 % and the shared election | recorded in CPA-REVIEW |
| #95 | `regimes.explanation` invariant enforced | `ce70beb` |
| #100 | `EngineVersion` 1.0.1 after the 626 PM fix | `e82c1d2`; M5-30 G2 digest pin |
| #105 / M5-30 G1, G2 | No-float guard fails closed; `EngineVersion` follows figures | M5-30 gate |

### Deferred, with owner

| Source | Item | Owner | Target |
|---|---|---|---|
| G-1, S-4 (errors), M3/M4 G-1 | `no-store` on error envelopes | architect + backend-sr | PR #107 (open); recheck at merge |
| G-2, #102 FU | Validate `EstimateDeps` in `NewEstimator` (fail at boot) | backend-sr | before M6 wires a second caller |
| G-3 | Owner check in the `taxinputs` and profile queries if shared access is ever designed | architect | ADR for any delegated-access feature |
| G-4 | Count the suggestion reads in the `estimate_reads` budget comment, or batch them (one YTD call) | backend-sr + frontend-sr | M5 hardening / M6 |
| G-5, #77 should-fix 1 | Generalize the "every key the modules read" check to every embedded set, at boot or in CI | backend-sr | M5-32 Dec–Jan parameter runbook, before any 2027 file |
| #77 should-fix 2 | `provisional_parameters` notice on the records forms (categories/defaults responses) | architect | next contract re-freeze |
| #98 FU1 | Fingerprint hashes informational sums and multi-enum order (false "changed" marks in M8) | backend-sr | M8 |
| #98 FU2 | Replace `noHolidays` with the holidays table (due dates) | backend-sr | M9-03/M9-04 |
| #90 FU | RESICO PF months after the cap still use RESICO rates (warning only) | owner + CPA | M5-33 |
| #101 FU 1–4 | Error-half rendering; future-year months in `paymentMonths` (server refuses the future year, `400`); `CitationDisclosure` reuse; MODULE §8 wording | frontend-sr (1–3), doc-writer (4) | M5 UX polish / M5-32 |
| #102 nit | Contract examples show `engine_version: 1.0.0` | architect | next contract re-freeze |
| PLAN FU1 | Accrual collection date out of *Más detalles* | pm + frontend-sr | before `module_incomes` for PM users |
| PLAN FU2 | PM-601 withheld-to-remit note for the CPA | qa-sr | M5-33 |
| PLAN FU3 | Add-a-regime guide with both `EngineVersion` pins | doc-writer | M5-32 |
| PR #108 | Regime start before the current month (M5-29 D2) | backend-sr | in flight; not covered by this gate |

### Product risk: CPA-unverified items

The engine is only as right as the rules in [CPA-REVIEW](CPA-REVIEW.md), and **none of them is
ticked**: PTU 10 %, the RIF reduction table and its rounding, the 606 quarterly threshold, the 625
definitive IVA and 8 % direct collections, RMF rule numbers, rental IVA withholding (2/3 vs 100 %),
the due-date rules, IVA netting across regimes, LIVA 5 fr. IV timing, the 15 regime explanations,
26 regime-compatibility pairs, every expense category rule, and the income basis of 620/622/624.
Two golden cases stay skipped (ratchet) until M5-33. The app labels these figures as estimates with
CPA-review warnings, which is honest, but a figure that the user pays from is the module's whole
value. This is not a security defect. It is the reason P-1 below exists.

## 5. Preconditions to enable `module_tax_engine` in staging or production

All must hold; each is verified by the person flipping the flag and recorded in the deploy PR.

1. **P-1 CPA sign-off (M5-33, OB5)**: the CPA has reviewed [CPA-REVIEW](CPA-REVIEW.md), the ⚠️ golden
   cases are un-skipped or have recorded decisions, and any corrected rule shipped with an
   `EngineVersion` bump. Until then the module may run in dev and staging for testing, but it is not
   enabled for real users (production).
2. **P-2 PR #107 merged** (contract 0.7.0) and `no-store` verified on an M5 error response in the
   target environment (closes G-1).
3. **P-3 `module_taxpayers` preconditions** (M2-14 §5), and those of `module_incomes` /
   `module_expenses` (M3-M4 §5) if their screens are shown; the estimate reads their data either
   way.
4. **P-4 Gates**: M5-30 (tech-lead) PASS, this gate PASS; REL-09 (deploy security gate) PASS.
5. **P-5 `APP_ENV` is not `dev`** (the dev email adapter logs OTP codes).
6. **P-6 Limiter healthy**: the per-user budgets fail closed, so a limiter outage turns every
   estimate and input write into `429`; alert on `rate limit unavailable`.
7. **P-7 Logs retained** with the `tax_inputs.updated` audit event (ids only).
8. **P-8 Mobile release build** passes `mobile-artifact-check` (no source maps, no internal URLs)
   and uses an HTTPS API base URL.
9. **P-9 Parameter set for the year in force**: before January 2027, a reviewed `2027/v1.yaml` (or
   the provisional fallback banner is accepted by the owner), with G-5 closed.

## 6. Notes

- The cso run is recorded as **partial**: its 600 s daily budget covered the snapshot and the
  helper-mediated reads listed in the Method row. The coverage submission was rejected by the
  helper's schema check, and the run was finished before a correction, so the helper holds no
  coverage record. The rest of the review was manual and runtime, and is listed above.
- The isolated stack used its own compose project, container name **and** volume name (checked
  with `docker compose config`). Only the API process started for this gate was stopped. The
  stack and volume were removed.
- Test secrets (database passwords, keys) were generated at runtime into a mode-0600 file outside
  the repository, deleted afterwards and never committed.
