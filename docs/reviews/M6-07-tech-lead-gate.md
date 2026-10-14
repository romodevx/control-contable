# M6-07: Tech-lead gate of module M6 Dashboard

- **Date:** 2026-10-14
- **Branch:** `feature/dashboard-qa-fixes` (PR #116 → `dev`): the M6-05 flag tests and the M6-06 QA report, on
  top of all of M6 already on `dev` (M6-01 contract 0.8.0 #111, M6-04 Home #113, M6-02/M6-03 use case and
  handler #114).
- **Scope:** A) PR #116. B) Holistic review of `backend/internal/dashboard/**`, `backend/cmd/api/dashboard.go`,
  the tax engine's `EstimateRange` and the ledger readers' range form, the router change that refuses
  undeclared query parameters, and `mobile/lib/features/dashboard/**` (Home, hero, month card, chart, table,
  trial banner). C) A code-level accessibility audit of the chart, because the QA could not drive TalkBack
  through adb ([M6-06 §6](M6-06-qa.md#6-screen-reader)).
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language (idiomatic Go
  wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gate (M6-08) runs at the same time on another branch.

## Verdict: **PASS**

No blocker is open. Two Medium defects in the mobile Home were found and fixed on this branch, each with a
failing test first: chart bars that a screen reader could not activate (QA D-1), and a failed pull to
refresh that left the old figures on screen as if they were fresh. The backend matches the contract and
the spec, and every API call the mobile client makes declares its query parameters, so the router's new
rule cannot break the app.

PR #116: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `make -C backend check` (gofmt, golangci-lint, `go test -race ./...`) | green, 0 lint issues |
| `make mobile-check` (format, analyze, tests + coverage gate) | green: 1,141 passed, 30 skipped |
| PR #116 required checks (head `4160c2d`, before this review) | all green, integration (testcontainers) included |
| Generated Dart clients vs `openapi.yaml` 0.8.0, every operation's query parameters | 52 of 52 operations match, **0 undeclared** (script over `@GET/@POST/…` + `@Query` against path- and operation-level parameters, `$ref`s resolved). The only other `queryParameters` in `mobile/lib` are go_router locations, not API calls |
| New `dashboard_semantics_test.dart` before the fix | 6 of 7 green, **red**: a bar has no `SemanticsAction.tap` |
| New "failed pull to refresh" test before the fix | **red**: the hero stays, no error shown |

## A) PR #116

| Item | Assessment |
|---|---|
| `adapter/http/gen/dashboard_module_test.go` | Good. The operation is written by hand by operationId and by ServeMux pattern, and the flag owns nothing else. Same pattern as M2–M5. |
| `flags/prerequisites_test.go` | Good. Pins `module_dashboard → module_taxpayers` only. OQ-D7 (engine is not a prerequisite) has a test. |
| `cmd/api/dashboard_preset_smoke_integration_test.go` | Good. Every seeded preset plus four fixtures, with a real user and records, and the expected result written from the spec (flag **and** prerequisite), not read from `flags.Prerequisites`. It also checks that each optional module that is OFF is `null` in the body, and a guard fails if a new dashboard operation is not probed. The report shows it can fail. |
| `dashboard_flag_test.dart` | Good. OFF keeps the previous Home and never reads the dashboard. The `estimad` regex is case-insensitive (the M5-30 lesson was applied). |
| `onboarding_after_sign_up_test.dart` + `AuthHarness.overrides` | Good. It covers ON and OFF and the after-sign-out path. The `overrides` parameter is additive and does not change existing callers. |
| QA report M6-06 | Thorough: every Home figure was checked against what was entered and against the estimate screen, the flag matrix was walked, and environment problems are kept apart. Withdrawing the suspected defect after finding the root cause (a `KEYCODE_BACK` from the driver), and discarding the speculative fix, was the right call. §8 discloses the `pkill -f` pattern honestly; the rule stays **kill by PID only**. |
| TDD evidence | The M6-05 tests pin behavior that already existed, so they passed on the first run. The report says so and shows a mutation that turns each one red. This is acceptable for characterization tests (same standard as M5-30). |

## B) Holistic findings

### Fixed on this branch

| # | Sev | Finding | Fix |
|---|---|---|---|
| G1 | **Medium** | **A screen reader could not activate a chart bar.** `_MonthColumn` wraps a `GestureDetector` in `Semantics(button: true, excludeSemantics: true)`. `excludeSemantics` also drops the detector's tap action, so TalkBack and VoiceOver announced "button" on a node that had no action. A screen-reader user could not change the selected month (QA D-1, filed as a UX call). This is the **same defect the taxpayer switcher had in QA M3/M4 D1**, and the fix was not carried over to the new feature. *WCAG 2.2 4.1.2 Name, Role, Value: a node with the button role must be operable. Clean Code: the same bug found twice calls for a rule, so it goes to the follow-ups.* | `3deb1b3` (red test), `ca608ac`: the bar's `Semantics` carries `onTap` itself. |
| G2 | **Medium** | **A failed refresh showed stale figures as if they were fresh.** Riverpod keeps the previous value on an `AsyncError`, and `_DashboardContent` checked `value` before `error`. After a pull to refresh failed (for example, offline), Home kept the old figures and showed no error and no retry. MODULE-dashboard §5 says offline is the error state and v1 caches no figures, and the use case refuses partial answers for the same reason. | `9585eba` (red test), `d82ac40`: the error wins whenever the read is not loading. A refresh that is still running keeps the old figures, and switching taxpayer shows the skeleton (US-D7, unchanged). |

### Spec ↔ contract 0.8.0 ↔ code (MODULE-dashboard OQ-D1…D7)

| Decision | Contract | Backend | Mobile | Result |
|---|---|---|---|---|
| OQ-D1: 12 months, 6 allowed | `months` 6..12, default 12 | `domain.MinMonths/MaxMonths/DefaultMonths`; `Window` clamps to 2026-01 (table test) | omits `months` and `period` (server decides) | ✓ |
| OQ-D2: income = collected | "collected/paid before IVA" | `Records.IncomeCollected` = the ledger's `Cash` (`collected_on`); engine OFF reads the same aggregate | "Ingresos cobrados" | ✓ The flag matrix test checks exact sums with the engine ON and OFF |
| OQ-D3: OFF modules hidden | nullable `income_cents`/`expense_cents` | figures `nil` per flag; the dashboard's own aggregate is skipped when the engine is OFF | series, figure, "Diferencia" and quick action hidden from the response alone | ✓ |
| OQ-D4: withholdings segment | `withholdings_to_remit_cents` | from `Totals.WithholdingsToRemit` | third series only when any month is > 0 | ✓ |
| OQ-D5: own budget | 429 declared | `dashboard_reads` 600/h; the range estimate does not spend `estimate_reads` (per-operation limits) | — | ✓ |
| OQ-D6: server days left | `trial_days_left` nullable | `domain.TrialDaysLeft`, rounded up, minimum 1, table test | banner reads it, never the device clock | ✓ |
| OQ-D7: engine OFF degrades Home | `estimate`/`trend[].tax` nullable, `legend` optional | estimator not called, `estimate`/`tax` nil, no legend; only prerequisite is `module_taxpayers` | no tax view, no switcher, no legend | ✓ MODULE-tax-engine §7 amended (line 434) |
| §2.4 one aggregate pass | — | `EstimateRange` reads the ledgers once per table across two fiscal years; `EstimateMonth` = the range of one month; 8 statements pinned by `TestContract_dashboardStatementCount` | — | ✓ |
| §8 indistinguishable 404 | 404 "Not found." | `hidden()` rewords any port's not-found as the module's; contract test compares against flag OFF | — | ✓ |

### Router: undeclared query parameters are refused (global change in #114)

`router.validate` now answers `400 validation_failed` with `details[].code = additionalProperties` for any
query parameter that the operation (or its path item) does not declare. It runs after the flag gate and
authentication, so a module that is OFF still answers 404 first (contract test: `not-a-uuid` + `months=99`
on OFF is the hidden 404). **Every client call is declared** (Evidence, 52/52). Two documentation gaps
remain (FU1).

### Checklist results (no action needed)

| Area | Result |
|---|---|
| **Clean Architecture** | `dashboard/domain` imports the standard library only. The use case declares its ports on the consumer side (`TaxpayerChecker`, `TaxEstimator[D]`, `RecordReader`, `AccessReader`, `Flags`). The tax engine's presented body travels as the type parameter `D`, so the use case never imports the engine's HTTP adapter. Adapters for the engine, ledgers and taxpayers live in `cmd/api/dashboard.go` (composition root). archtest covers `internal/<module>` generically. |
| **SOLID** | Each type has one reason to change: domain rules (window, status precedence, days left), orchestration (`Service.Get/read`), presentation (`present.go`), and the rate policy (`limits.go`). `NewService` refuses a missing port, so the boot fails (FU4 of M5 is closed for the estimator too). |
| **Idiomatic Go** | Errors are wrapped with `%w`, and sentinels are checked with `errors.Is`/`errors.AsType`. `ctx` is the first parameter and flows into the unit of work. There are no goroutines. Overflow-checked sums use `money.Cents.Add` with `errors.Join`. `present.go` writes `null` explicitly where the generated types would omit a field, with a comment that says why. |
| **Algorithms** | The window is O(months), and the trend is built from `map[Month]` lookups in O(months). `is_empty` reuses the window's figures before it runs an `EXISTS`. The mobile scale, peak and totals are each one pass. |
| **Tests** | Domain: table-driven tests, including boundaries (days left at 0, −1 h and 1 ns; windows that cross a year). Use case: fakes per port, plus a statement-count pin. Contract: every response is validated against the spec, plus the flag matrix (8 combinations), empty state, RLS (Bea sees only her records) and amounts never logged. Mobile: 63 dashboard tests (the 7-test semantics suite included), plus contrast and target guidelines in both themes, plus 200 % text. |
| **Flags** | Server: `flaggate` at mux registration, with the prerequisite in `flags.Prerequisites`. Client: one `FlagGate(module: dashboard)` in `home_screen.dart`, and the record and engine modules are hidden from the response alone, with no scattered flag checks. |

## C) Accessibility audit (code level; the QA's TalkBack gap)

`mobile/test/features/dashboard/dashboard_semantics_test.dart` pins the semantics tree that TalkBack and
VoiceOver read. It reads the tree itself, not `find.bySemanticsLabel` (which matches any node and cannot
tell whether a label merged into a bigger one):

| Assertion | Result |
|---|---|
| The tax chart summary is its own node, labelled exactly "Impuestos estimados, últimos 3 meses. Marzo: 2,284 pesos con 29 centavos. Mes más alto: febrero, 3,500 pesos.", with no button flag and no action | ✓ |
| Under it: one bar node per month, oldest first (reading order) | ✓ |
| A bar says month, each series, total and status ("febrero de 2026: IVA 3,100 pesos, ISR 400 pesos, total 3,500 pesos, incompleto"), with the button role and selected state | ✓ |
| A screen reader selects a month through the bar's **tap action** | ✓ after G1 |
| A one-month window (period 2026-01) keeps the summary and the lone bar apart | ✓ |
| The income/expense view's summary states both series for the month and the window totals | ✓ |
| Engine OFF: the summary mentions no tax (no "impuesto", "IVA" or "ISR") | ✓ |
| Table view: each month is one heading node ("febrero de 2026, Incompleto") whose children read one sentence per figure, in column order | ✓ |

The table is a list of month blocks, not a grid. For a screen reader this is equivalent to a table: every
figure is named, and the status is given in words. It also wraps at 200 % text, which an 8-column grid on a
phone cannot do. §6's "real table" wording should be updated to match (FU5).

**Still owed: a manual pass with TalkBack (Android, real device) and VoiceOver (iOS, OB6).** Widget tests
check the tree, not the platform's traversal, gestures or speech (FU6).

## Follow-ups

None blocks M6's merge. Target: the M6-09 docs task copies these into PLAN §M6 follow-ups.

| ID | Sev | Item | Owner | Target |
|---|---|---|---|---|
| FU1 | Low | **Undeclared query parameters, documentation.** (a) The rule now applies to every operation, but the contract states it only in the dashboard and tax-engine 400 descriptions. Add one sentence to the shared `ValidationFailed` response (or `info.description`) and an ADR-0004-style note: it changes behavior for every endpoint, and third-party clients will exist one day. (b) MODULE-tax-engine §5 says `details[].code = invalid`, but the router emits `additionalProperties`; align the spec. (c) For M6-08: the refused **names** are echoed in `details[].field` without a limit on count or length. This is JSON-encoded and only the name is echoed, never the value; consider capping it | architect (+ security for c) | Next contract re-freeze |
| FU2 | Low | **"Últimos 1 meses".** `dashboardTrendTitle`, `dashboardTaxSummary` and `dashboardIncomeExpenseSummary` take `{count}` without ICU plural. A one-month window (period 2026-01, which M8 or a test can ask for) reads "últimos 1 meses". Use `{count, plural, =1{…} other{…}}` | frontend-sr | M8 (first caller with a `period`) |
| FU3 | Low | `dashboardProvider` uses `ref.read(dashboardRepositoryProvider)` inside a provider, so a rebuilt API client (base URL, session) does not reach an alive family member. Use `ref.watch` (Riverpod guidance for provider bodies) | frontend-sr | Next touch of `features/dashboard` |
| FU4 | Low | Backend nits: `Service.Get` reads `Clock.Now()` twice (period and days left), so read it once. `trendTax` turns an empty `IVAOutcome` into `zero`, which hides a wiring bug; return an error instead | backend-sr | Next touch of `dashboard/usecase` |
| FU5 | Low | MODULE-dashboard §6 "real table" → "one block per month: a heading and one sentence per figure" (what ships, see C). Also mark PLAN M6-02, M6-03 and M6-04 done (merged in #113/#114) | doc-writer | M6-09 |
| FU6 | — | **Owner:** a manual TalkBack pass on a real Android device and VoiceOver on iOS (OB6). Checklist: swipe order banner → hero → legend → month card → summary → bars → key → selection; double-tap a bar selects it (G1); "Ver como tabla" reads each month as a heading; 200 % text | owner (+ qa-sr) | Before `module_dashboard` is enabled in a shared environment |
| FU7 | — | **Guard against G1 recurring.** This is the second `excludeSemantics` that dropped a tap action (taxpayer switcher, M3/M4 D1; chart bar, here). Add a repo-wide widget-test helper (or an `accessibility_test.dart` sweep) that fails on any node with the button flag and no tap action | frontend-sr | M6 hardening |
| QA D-2 | — | "Ingresos y gastos" title and summary with one record module OFF | owner (copy) | UX polish |
| QA D-3 | — | US-D3 wants "Saldo a favor $X". Contract 0.8.0 `TrendTax` has `iva_outcome` but no amount, which is a spec ↔ contract gap. Add `iva_in_favor_cents` (additive, 0.8.x) | architect | Next contract re-freeze |
| QA D-4 | — | *usted* vs *tú* copy pass | pm (copy) | UX polish |
| QA obs. | — | Returning from the estimate screen keeps the chart's scroll, so the selected month can be off screen. The status chip shows in the income/expense view | frontend-sr | UX polish |
