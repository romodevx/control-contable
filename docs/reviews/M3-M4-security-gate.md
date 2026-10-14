# M3-13 / M4-14 · Security gate: M3 Incomes and M4 Expenses

| Field | Value |
|---|---|
| Scope | M3 and M4 as built on `origin/dev` @ `7a60bcc` (2026-10-14) plus PR #103 (head `144c70a`: M3-10/M4-11 flag tests, M3-11/M4-12 QA fixes, open). `backend/internal/{incomes,expenses}/**`, `internal/kernel/fiscal`, migrations `20261013200000_incomes.sql` (incl. the `incomes_derive_booked_on` trigger), `20261013210000_expense_categories.sql`, `20261013211000_expenses.sql`, `20261014150000_expenses_tax_facts.sql` (`cash_over_limit`, `creditable_iva_cents`), the ledger readers (M3-07, M4-08), `cmd/api` wiring (`incomes.go`, `expenses.go`, `records.go`, `ledger*.go`), the handlers and write budgets, idempotency, and mobile `lib/features/{incomes,expenses}/**` + `core/records/**` incl. the session memory cleared on sign-out |
| Reviewer | security |
| Date | 2026-10-14 |
| Method | `/cso --scope records` (gstack-cso 3.0.0, run `1791994842698-109531dda7af8185`, daily static, finished **partial**: 0 supported findings; helper-mediated read of the income write/manage use cases, the regime-records adapter and both record lookups; no scanners). Then a manual review of the rest of the scope; `go test -race` of incomes, expenses, kernel and `cmd/api`; `go test -tags=integration` of both modules, `migrations` and the `cmd/api` regime-end race; and an end-to-end probe of the API built from **PR #103's head** against an **isolated local** postgres (compose project `cc-gate-m3m4`, own container `cc-gate-m3m4-postgres` and volume `cc-gate-m3m4-pgdata`, verified with `docker compose config`; DB on 5447, API on 3047). The prerequisite check was repeated with an API built from `dev`. Flags were turned OFF afterwards, the stack and its volume destroyed. The shared local stack (`control-contable-postgres`, 5434) and other agents' processes were not touched |
| Verdict | **PASS.** No supported finding of severity medium or above. `module_incomes` and `module_expenses` stay **OFF** in every shared environment until the preconditions in §5 hold; the first one is that **PR #103 is merged** |

## 1. Evidence

Probe totals: 68 of 70 checks of the main e2e probe passed on the first run, plus 6 of 6 in
the regime-end and prerequisite probes. The two misses: one probe bug (the regime-end check used
a taxpayer with a single regime, so it got "needs at least one active regime"; re-run with
612 + 606, below) and one real gap (no `Cache-Control` on error envelopes, G-1).

| Check | Result |
|---|---|
| Unit (`-race`) | `internal/incomes/...`, `internal/expenses/...`, `internal/kernel/...`, `cmd/api`: green |
| Integration | both modules' `infra`, `migrations` (RLS isolation, composite FKs, column grants, cascades, booked_on trigger, CHECKs), `cmd/api` regime-end race (now 6 subtests, §3): green |
| **IDOR, 404 parity** (22 pairs) | user B with A's income/expense id under B's own taxpayer (get, update, delete, restore × 2 modules) and with A's taxpayer id on every op (incl. list, create, defaults, categories) vs a random uuid: all `404`, bodies identical minus `request_id`. A's income and expense rows byte-identical afterwards; B spent no trial slot |
| **Idempotency scoping** | B posting A's income id / expense id → bare `409 conflict` (no `details`), B's usage stays `0/0` (the reserved slot rolls back). A's same body twice → `201` then `200`, usage unchanged. Same id, other fields → `409 id_reused`; same id, A's other taxpayer → `409 id_reused`. Ids are per table: an income id reused as an expense id is a new expense (`201`) |
| **Trial cap race** | 8 concurrent income creates → exactly 5×`201` + 3×`402`; same for expenses; `user_usage` `5/5`, 5 + 5 rows. 6th income → `402` with `incomes_remaining` |
| **Restore, no double count** | at 5/5: delete → `204`, new create still `402` (delete frees nothing), restore → `200`, restore again → `200`; usage stays `5/5` |
| **Double submit at the cap** | user at 4/5, four concurrent creates with **one** client id → `[201, 200, 200, 200]`, never `402`; usage `5/0` (PR #83 fix holds over HTTP) |
| **Read-only** (trial moved to the past, LOCAL superuser) | incomes and expenses alike: create `402` (no details), update `402`, get `200`, list `200`, delete `204`, restore `200` (contract 0.6.1) |
| **Derived / identity columns** | create with `booked_on`, `user_id`, `taxpayer_id`, `deleted_at` (income) or `cash_over_limit`, `creditable_iva_cents`, `user_id` (expense) → `400`; PATCH with `booked_on`, `id`, `creditable_iva_cents`, `cash_over_limit` → `400` (`additionalProperties: false`). A $3,000 cash expense with `deductible: true` stores `cash_over_limit = true`, `creditable_iva_cents = 0` (LIVA 5 fr. II) |
| **RLS as `cc_app`** | in B's context: 0 of A's incomes/expenses visible, cross-user `UPDATE` → `UPDATE 0` on both tables; no context → 0 rows; `DELETE` → permission denied (soft delete only); `UPDATE booked_on` / `UPDATE user_id` → permission denied (column grants) |
| **Cascade** | `DELETE /v1/me` → `204`; the user's incomes and expenses `0/0` |
| **Regime end vs records** | income invoiced 2026-09, collected 2026-10: ending 612 at 2026-09 → `422 records_after_end_period` (the incomes lookup uses `greatest(invoiced_on, collected_on)`); at 2026-10 without confirm → `409 regime_has_records`; with confirm → `200` (M2-14 G-1 closed for both modules) |
| **Regime-end race** | `TestRegimeEndRace_recordWrites`: create, restore and **update** (new, §3) of incomes and expenses wait on the end path's locks and then read the ended window (`422 regime_not_active`), no live record outside it. The update rows were mutation-checked (skip the window check in `Incomes.Update` → `income_update` fails) |
| **Write budgets** | `income_writes` 300/h per user: 300 updates `200`, the 301st `429` with `Retry-After`; an expense create by the same user still `201` (own scope `expense_writes`); reads still `200` |
| **Prerequisite flags** | `module_taxpayers` OFF, the other three ON. **PR #103 build:** list/create incomes and expenses and categories all `404`; `cmd/flags` reports both modules "hidden until module_taxpayers is ON". **`dev` build:** the same calls answer `200`/`201` (an income and an expense were **created**), so the defect PR #103 F1 fixes is live on `dev` |
| **No PII in logs** | create with a marker description, a marker CFDI UUID and a distinctive amount, plus an IVA mismatch carrying them: the API log at `LOG_LEVEL=debug` (373 lines, 329 audit events) has **0** descriptions, **0** UUIDs (either case), **0** amounts, **0** passwords. Audit events carry only `user_id`, `taxpayer_id`, `income_id`/`expense_id`, `request_id`. Error bodies never echo the values |
| `Cache-Control` | `no-store` on every 200/201 record body; **absent on error envelopes** (G-1) |
| Mobile | record data is never persisted on the device (no `SharedPreferences`/storage use in `features/{incomes,expenses}` or `core/records`); list providers are `autoDispose` and keyed by taxpayer; the one `developer.log` logs the failure type only. Session memory (last payer, payment method, category usage, Undo stacks) is reset on sign-out by `forgetOnSignOut` in PR #103 (D4, test first); on `dev` it still leaks across accounts on one device |

`govulncheck` is not installed on this machine; the CI `govulncheck` job (required check) is
the evidence for dependencies on this PR.

## 2. Threat model (STRIDE), M3/M4 as built

Actors: authenticated user; another authenticated user (curious or hostile); a read-only user
(trial ended); an anonymous caller; an operator or attacker with the `cc_app` credential or a
backup; someone holding the phone after the owner signed out.
Assets: income/expense amounts, descriptions and CFDI folios (financial PII), the derived tax
facts (`booked_on`, `deductible`, `cash_over_limit`, `creditable_iva_cents`) the M5 estimate
sums, the trial counters, the regime windows, the ownership `user → taxpayer → record`.
Entry points: 14 operations under `/v1/taxpayers/{id}/{incomes,expenses,expense-categories}`
(bearer, flag-gated, prerequisite `module_taxpayers`), the ledger readers (M5, same unit of
work), the `cc_app` role, the device session.

| | Threat | Control as built | Residual |
|---|---|---|---|
| **S** | Act on another user's record with a stolen/guessed id | user id only from the verified token; RLS FORCE on both tables (`user_id = app_user_id()`); `findLive`/`findOwned` also compare `taxpayer_id`; 404 parity on all ids (§1) | — |
| **S** | Claim another user's client id on create | PK collision under RLS → bare `409`, slot rolled back | A `409` tells the caller some user holds that uuid; client uuids are unguessable, nothing about the owner leaks. Accepted (PR #82) |
| **T** | Write the derived tax facts (`booked_on`, `cash_over_limit`, `creditable_iva_cents`) to inflate deductions | not in the contract (`400`); `booked_on` has no grant (trigger only, refuses instead of guessing, PR #64); the two expense facts are set only from `domain.Assess`; DB CHECKs bind them (`cash_over_limit ⇒ cash`, `0 ≤ creditable ≤ iva`, `creditable > 0 ⇒ has_cfdi ∧ ¬cash_over_limit`) | `cc_app` can write the expense facts (needed); category creditability is enforced by the domain only (the DB cannot see the catalog). Needs API compromise. Accepted |
| **T** | Move a record onto a foreign taxpayer / regime | composite FK `(user_id, taxpayer_id, regime_code) → taxpayer_regimes`; no UPDATE grant on `user_id`/`taxpayer_id`/`id` | An immutability trigger on those columns (PR #64 FU2) is defense in depth only; deferred |
| **T** | End a regime while records still live after it (orphan records, wrong estimate) | `regimeRecords` = later of `max(paid_on)` and `max(greatest(invoiced_on, collected_on))`; end path locks taxpayer + regime rows; every create/update/restore checks the window under `FOR SHARE` on both | — |
| **T** | Beat the trial cap with concurrent or repeated creates | `ReserveRecord` (`UPDATE … WHERE n < 5`) last in the create tx; client id advisory lock before the lookup; restore never reserves; delete never refunds (trigger) | — |
| **R** | Deny a record change | audit `income.*` / `expense.*` created/updated/deleted/restored, ids only | Retention is a deploy concern (P-6) |
| **I** | Amounts, descriptions, folios in logs, errors or caches | ids-only audit; fixed es-MX messages; pg errors never reach the wire (`httperr`); `no-store` on every record body | Error envelopes lack `no-store` (G-1); they carry no record data |
| **I** | Cross-user aggregates through the ledger readers | readers run on the caller's unit of work as `cc_app` under RLS; another user's taxpayer reads empty (tested) | The estimate reads records even when `module_incomes`/`module_expenses` are OFF (MODULE-incomes OQ-8, by design; own data only). Info |
| **I** | Next account on the device sees the previous one's choices | PR #103 D4: session memory invalidated on sign-out; no record data persisted | Open until #103 merges (P-1) |
| **D** | Flood writes / burn the DB | 300/h per user per module, fail-closed limiter; edge per-IP limiter; trial cap | — |
| **E** | Read-only user keeps creating | `EnsureWritable` on create/update; restore allowed by decision (creates nothing) | — |
| **E** | Records reachable while their parent module is OFF | `flags.Prerequisites` declares `module_taxpayers` for both (PR #103 F1) | **Open on `dev`** until #103 merges (P-1) |

## 3. Findings of this gate

No supported finding of severity medium or above (cso run: 0 supported findings; manual review
and e2e: none on the PR #103 build). The prerequisite-flag defect is **High** but was found and
fixed by M3-11/M4-12 QA (PR #103 F1); this gate verified the fix end to end and makes its merge
precondition P-1.

| ID | Severity | Item |
|---|---|---|
| G-1 | Low | S-11 asks `Cache-Control: no-store` on record responses **and errors**. Bodies have it; the shared error envelope (`platform/httperr.Write`) does not, and the frozen contract declares no `Cache-Control` on error responses (the contract tests reject undeclared headers). Error bodies carry only codes and fixed messages (verified), so nothing sensitive is cacheable. Same item as M1-17 G-3 / P0-30 D-6 |
| G-2 | Info | `Delete` reads the archive flag through `Profile` without a lock (PR #83 FU): a delete racing an archive can land after it. Soft delete only, own data. `Restore` is serialized through `CheckRegime` |
| G-3 | Info | The M5 estimate reads incomes/expenses regardless of their flags (OQ-8). Own data only; the presentation rule is M5's |
| G-4 | Info | The dev email adapter logs OTP codes (`APP_ENV=dev` only, by design); shared environments must not run `dev` (M1-17 P-3) |

Fixed in this gate (test first, mutation-checked):
- **PR #97 FU-3**: the regime-end race test now covers `updateIncome` and `updateExpense` moving a date past the end (`cmd/api/records_race_integration_test.go`).
- **PR #97 FU-1**: the `records.go` and race-test comments now name the real serializer (the taxpayers row first, then the regime row), not only the regime row.

## 4. Carried M3/M4 follow-ups

Sources: MODULE-incomes §5 (S-1 … S-12), MODULE-expenses §5 (SE-1 … SE-3), M2-14 G-1/P-5, and
the reviews on PRs #64, #65, #66, #67, #82, #83, #85, #91, #93, #96, #97, #99, #103.

### Resolved

| Source | Item | Evidence |
|---|---|---|
| S-1, SE-1, #64, #65 | RLS ENABLE + FORCE, per-op policies, no DELETE, column grants; catalogs read-only | migration tests; e2e RLS (§1) |
| S-2, SE-2 | Composite FK, CHECKs (amounts, withholdings, dates, CFDI, payment method) | migration tests |
| S-3, S-11 (bodies) | No mass assignment, `additionalProperties: false`, `no-store` on record bodies | e2e derived-column matrix |
| S-4, #82, #83 | Atomic trial slot, restore never reserves, double submit at the cap replays | e2e race + double submit |
| S-5, S-6, #93, #97 | No amounts/description/folio in logs; ids-only audit | e2e log scan |
| S-7 | 300/h per user per module | e2e 429 |
| S-8 | Repositories never Commit/Rollback/SET | archtest |
| S-9 | Deletion cascade covers records | e2e + `account_deletion_integration_test.go` |
| S-10, SE-3 | Foreign ids `404`, foreign idempotency ids bare `409`; categories taxpayer-scoped | e2e IDOR matrix |
| M2-14 G-1 / P-5, #93, #96 | `NoRecords` replaced; incomes lookup uses both dates | `cmd/api/records.go`; e2e regime end |
| #93 FU-1 | Regime-end race (create/restore) | `TestRegimeEndRace_recordWrites` (#97, mutation-checked) |
| #97 FU-1, FU-3 | Lock comment; update rows in the race test | this PR |
| #64 blocker | `booked_on` trigger refuses instead of guessing | #64 `5415cf8` |
| #66 F1, #85 | Read-only users may restore (contract 0.6.1); Undo for everyone | e2e read-only row |
| #66 F2 | Merge-patch `null` on both sides | `nullable.Nullable` overlay (#93, ADR-0001), Dart hand-written map (#85, ADR-0003) |
| #66 F3, #82 | Idempotent compare for expenses on the override and normalized withholdings | `sameRequest` (`expenses/usecase/create.go`) + tests |
| #66 F4, #85 FU, #99 | Stacked Undo partial failure reported ("N de M"); toasts queued behind a pending Undo | `stacked_undo_test` (#99) |
| #67 LIVA 5 fr. I/II | Creditable IVA 0 for never-deductible categories and cash over the limit | domain tests; e2e cash row |
| #82 heads-up | One warning code `provisional_parameters` for both modules | contract `WarningCode` |
| #83 FU (flag) | Handlers flag-gated, all 14 ops hidden when OFF | `AssertHidden` (#93, #97); preset smoke (#103) |
| #91 note 3 | Estimate vs `module_expenses` flag | decided: OQ-8 reads regardless (G-3) |
| #103 F1 | `module_taxpayers` prerequisite for both modules | #103 `08d8db5`/`7e57970`; e2e (§1). **Pending merge** |
| #103 D4 | Session memory cleared on sign-out | #103 `f7f366d`/`e4380a7`. **Pending merge** |

### Deferred, with owner

| Source | Item | Owner | Target |
|---|---|---|---|
| G-1, M1-17 G-3, P0-30 D-6, S-11 (errors) | Declare `Cache-Control: no-store` on the shared error responses in the contract and set it in `httperr.Write` | architect + backend-sr | next contract re-freeze (with #97 FU-2) |
| #97 FU-2 | `out_of_range` descriptions still say "before 2022-01-01"; the served floor is 2026 (CONTRACT.md only) | architect | next contract re-freeze |
| G-2, #83 FU | `FOR SHARE` on the taxpayer row in `Delete` (incomes and expenses) | backend-sr | next records touch |
| #83 FU | Remove the PK-collision retry loop once the advisory lock is proven alone | backend-sr | next records touch |
| #64 FU2 | Immutability trigger on `user_id`/`taxpayer_id` of records (grants already block `cc_app`) | db-sr | M9 hardening |
| #91 note 2 | New constraints on populated tables use `NOT VALID` + `VALIDATE` | db-sr | standing rule |
| #64 FU1, #65, #67, #91 note 1 | Income basis of 620/622/624, booking-date basis, LIVA 5 rules, override on cash over limit | owner + CPA | before launch (`docs/reviews/CPA-REVIEW.md`) |
| #93 FU-2 | Inject the clock into `access.NewPolicy` in `newExpensesModule` | backend-sr | next records touch |
| #67 nit | `fiscal.MinRecordDate` is an exported mutable `var` | backend-sr | next kernel touch |
| #66 F5, F6 | `PeriodTotals` bound; es-MX vs English example messages | architect | next contract re-freeze |
| #85 FU | Form date labels keep the first regime's basis when a PM switches 601 ↔ 626 | frontend-sr | M3 UX polish |
| #99 FU 1–4 | Failed delete hidden behind a pending Undo; Prism mock suites not in CI; "5 gastos" hard-coded; withholding suggestions not prefilled | frontend-sr (1, 3, 4), devops-sr (2) | M3/M4 UX polish; CI backlog |
| #103 D-1 … D-9 | QA low/info items (report §4 of `M3-M4-QA.md`) | pm + frontend-sr | M3/M4 UX polish |

## 5. Preconditions to enable `module_incomes` / `module_expenses` in staging or production

All must hold; each is verified by the person flipping the flag and recorded in the deploy PR.

1. **P-1 PR #103 merged** with green checks: it declares the `module_taxpayers` prerequisite (without it both modules answer, and **write**, with taxpayers OFF; verified on `dev`) and clears session memory on sign-out (D4). Re-run its preset smoke in the target environment.
2. **P-2 `module_taxpayers` preconditions** (M2-14 §5) and **`module_auth` preconditions** (M1-17 §5) hold in the same environment. M2-14 P-5 is satisfied by this gate (records lookup wired).
3. **P-3 Gates**: M3-12 / M4-13 (tech-lead) PASS, this gate PASS; REL-09 (deploy security gate) PASS.
4. **P-4 `APP_ENV` is not `dev`** (the dev email adapter logs OTP codes, G-4).
5. **P-5 Limiter healthy**: the per-user write budgets fail closed, so a limiter outage turns every record write into `429`; alert on `rate limit unavailable`.
6. **P-6 Logs retained** with the `income.*` / `expense.*` audit events (ids only).
7. **P-7 Mobile release build** passes `mobile-artifact-check` (no source maps, no internal URLs) and uses an HTTPS API base URL.
8. **P-8 CPA review** of the items in `docs/reviews/CPA-REVIEW.md` before the estimate built on these records is shown as anything but provisional (product precondition, not a security one).

## 6. Notes

- The cso helper run is recorded as **partial**: its 600 s daily budget covered the snapshot
  inventory and a helper-mediated read of the core income write path and the records lookups;
  the rest of the review was manual and is listed in the Method row.
- The isolated stack used its own compose project, container name **and** volume name (checked
  with `docker compose config`); a first attempt interrupted by an API rate limit was destroyed
  and recreated. Only the API processes started for this gate were stopped. Local flags were
  left OFF; stack and volume removed.
- Test secrets (passwords, keys) were generated at runtime into a mode-0600 file outside the
  repository and never committed.
