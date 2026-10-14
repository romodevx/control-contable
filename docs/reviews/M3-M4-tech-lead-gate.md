# M3-12 / M4-13: Tech-lead gate of modules M3 Incomes and M4 Expenses

- **Date:** 2026-10-14
- **Branch:** `feature/incomes-expenses-qa-fixes` (PR #103 → `dev`), which carries M3-10/M4-11 (flags),
  the F1 prerequisite fix and the QA fixes D1-D4 ([M3-M4-QA](M3-M4-QA.md)) on top of all of M3 and M4.
- **Scope:** A) PR #103. B) Holistic review of M3 and M4: `backend/internal/{incomes,expenses}/**`,
  `kernel/fiscal` and its use, the four migrations (incomes, expense categories, expenses, expense
  tax facts), handlers and composition root, the ledger readers, `mobile/lib/features/{incomes,expenses}/**`
  and the shared `core/forms`, `core/records`, `core/session` pieces. Contract 0.6.1 ↔
  [MODULE-incomes](../project-specs/MODULE-incomes.md) / [MODULE-expenses](../project-specs/MODULE-expenses.md)
  (all OQ DECIDED) ↔ code. C) The nine QA-deferred items.
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language (idiomatic Go
  wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gates (M3-13, M4-14) run separately.

## Verdict: **PASS**

No blocker. Two High defects found by this gate (both in incomes: rules that M4 got right and M3 never
received) and one hole left by the D4 fix are fixed on this branch, each with a failing test committed
before the fix. Three of the nine deferred QA items are fixed test-first; the other six have an owner and
a target module. Both modules may have their flags enabled in a shared environment once the security
gates pass.

PR #103: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `make -C backend check` (gofmt, golangci-lint, `go test -race`) | green, 0 lint issues (backend unchanged by this gate) |
| `make mobile-check` (format, analyze, tests) | analyze clean, all tests pass |
| H1 (income save closes Undo) + D-3 | `records_save_toast_test.dart` red at `fb34088` (compile: no `undo:`), green at `acfff10` |
| H2 (income stacked Undo strands the rest) | `income_list_test` "a stacked Undo restores every income it can" red at `db3cd62` (`restore:i-rent` never called), green at `6329843` |
| M1 (Undo toast survives sign-out) | `forget_toasts_on_sign_out_test.dart` red at `11d5733`, app wiring test red at `1a841f2`; green at `a3187a7` |
| D-2 | `expense_rules_test` red at `3857086`, green at `abef84b` |
| D-8 | `form_controller_test` and `income_form_test` red at `0876aa0`, green at `099299b` |
| Gate scan (test edits in the PR diff vs `origin/dev`) | 3 hunks, none relaxing: `cmd/flags` list table gains the REQUIRES column, and two `flags.Service` tests add `module_taxpayers` to sets that enable `module_incomes`, both forced by the F1 prerequisite |

## A) PR #103

| Item | Assessment |
|---|---|
| F1: `module_incomes` / `module_expenses` → `module_taxpayers` in `flags.Prerequisites` | Correct fix at the one graph the gate, `GET /v1/flags` and `cmd/flags` read (no ad hoc checks). The preset smoke probes all 14 operations under each preset plus fixtures, including "records without taxpayers", and has a guard that every owned operation is probed. |
| `records_module_test.go` | Good: 7 + 7 operations written by hand from the specs, by operationId and by ServeMux pattern. |
| Mobile flag tests | Good: the route tree is pinned to 3 locations per module; OFF asserts not-found with zero repository calls. |
| D1 semantics tap action | Right root cause (`excludeSemantics` drops the child's action) and the switcher target now meets the 48 dp guideline its own test checks. |
| D2 floating toast, D3 reveal focus | Good. `_revealFocusAfterLayout` uses `Scrollable.ensureVisible` after layout and guards `mounted`. |
| D4 forget on sign-out | `forgetOnSignOut(ref)` is the right shape (one helper, a refresh keeps state). **Incomplete (M1 below):** dropping the `StackedUndo` instance does not close the persistent Undo toast already on the root messenger. Fixed here. |
| QA report | Thorough: wipe, UI-created data, 22 flows, PII grep of the API log, environment problems kept apart. |
| TDD evidence | Each fix has its test commit first (`08d8db5`→`7e57970`, `e19af74`→`41d2d20`, `bf8cf30`→`a5d44e1`, `548a4aa`→`981a653`, `f7f366d`→`e4380a7`). |

## B) Holistic findings

### Fixed on this branch (failing test first)

| # | Sev | Finding | Fix |
|---|---|---|---|
| H1 | **High** | **Saving an income closed a pending Undo.** `showIncomeSavedToast` called `hideCurrentSnackBar()` unconditionally; the Undo toast's `closed` future then cleared the stack, so deleting an income, opening the form and saving lost the Undo. Expenses had the rule (`if (undo.pending.isEmpty)`), incomes did not. *House rule ui-forms-and-modals: the Undo toast never auto-dismisses; Clean Code ch.17 G5 (duplication): two copies of one rule drifted.* | `showRecordSavedToast` in `core/records/record_save_toast.dart` and `StackedUndo.show` hold the rule once; both modules call it. |
| H2 | **High** | **A stacked income Undo stopped at the first refused income** and threw: the other incomes in the stack were never restored and the toast was gone. M4 fixed this with `restoreEach` (PR #85 follow-up); M3 never got it. *Same G5 drift.* | `IncomeRecords.restore` returns a `RestoreReport` from `restoreEach`; the toast says "Se restauró 1 de 2 ingresos." and why. |
| M1 | Medium | **The Undo toast survived sign-out** (completes QA D4). It is `persist: true` on the root `ScaffoldMessenger`, so the next account on the device saw "2 ingresos eliminados · Deshacer" and its tap restored with the new user's token (the server answers 404, so no data crosses, but the UI offered another user's action). | `ForgetToastsOnSignOut` in `MaterialApp.builder` clears the root messenger when the session ends; a refresh keeps it. |

### What holds up

- **Dependency rule.** `incomes/domain`, `expenses/domain` and `kernel/fiscal` import only stdlib and
  `platform/money`. GORM and pgconn appear only in `infra`. Use cases depend on small consumer-side ports.
- **Shared kernel.** IVA rates and allowance, withholding rules, CFDI UUID grammar, dates, enums and
  violations live in `kernel/fiscal` and both modules use them (no copied legal rules).
- **RLS.** `incomes` and `expenses`: ENABLE + FORCE, one policy per operation, no DELETE policy or grant
  (soft delete), column-level INSERT/UPDATE grants, composite FK to `taxpayer_regimes`. The global
  catalogs are read-only for `cc_app`.
- **DB as backstop, domain as authority.** Ranges (`fiscal.MaxCents`), the CFDI grammar, the
  description rule (trim, NFC, ≤ 200 runes) and the cash-rule ⇔ hint CHECK agree with the domain;
  unknown constraint violations stay infrastructure errors by design (`constraintErrors`).
- **Flags.** Gate at mux registration, prerequisite graph in one place, client gate at router registration.
- **Mobile layering.** Generated clients only in `data/`; failures are the sealed `ApiFailure`; session
  memory now forgotten on sign-out; edit mode sends merge-patches of changed fields only, and stored
  rates/withholdings are never replaced by suggestions (tests in `income_edit_test`, `expense_edit_test`).
- **Tests.** Table-driven Go with fakes, testcontainers integration and contract tests per module;
  widget tests deterministic (fake clock, no sleeps).

### Follow-ups (not fixed here)

| # | Sev | Finding | Owner / target |
|---|---|---|---|
| FU1 | Medium | **Incomes/expenses duplication is the root cause of H1/H2.** Mobile: the delete action + Undo flow (`*_delete_action.dart`, 66 differing lines of 99), the records write class and the form-page shell are copies. Backend: cursor codec (`encodeCursor`/`decodeCursor`), `canonicalID`, description normalization, `audit`, and the usecase error tables are copies. Extract the record shell to `core/records` (mobile) and a small `internal/kernel/records` (backend) before a third record kind lands. *SOLID/DRY, Clean Code G5.* | tech-lead + frontend-sr / backend-sr, `/refactor` before M6 |
| FU2 | Low | `canonicalID` drifted: incomes rejects non-canonical spellings (`{…}`, `urn:uuid:`, no dashes), expenses accepts and canonicalizes them. Not reachable over HTTP (the generated wrapper parses path/body ids as `uuid.UUID`), but one grammar should have one validator. Folds into FU1. | backend-sr, with FU1 |
| FU3 | Low | Income *Guardar y crear otro* keeps the withholdings section's open state; expenses resets it (`_withholdingsOpen = null`). | frontend-sr, with FU1 |

## C) QA-deferred items

| QA | Decision | Status |
|---|---|---|
| D-1 IVA note before the override | Product: show the note only with the override, or have §2.5 return the warning always. The note is true; nothing is wrong on the server. | **Deferred** → pm + frontend-sr, M4 polish (before M4-15 docs) |
| D-2 fuel note in 606 | The DB CHECK `category_regime_rules_cash_rule_follows_hint` ties the cash rule to the hint, so the client derives it from `deductibility`: fuel in cash is its own risk only where the category can lower ISR. No contract change. | **Fixed** (`3857086` → `abef84b`) |
| D-3 "con un aviso" for two | ICU plural on the four headings, through the shared toast. | **Fixed** (`fb34088` → `acfff10`) |
| D-4 US-E fuel 300.00 in 612 not re-run | Covered by contract tests and now by `expense_rules_test`. | Closed (info) |
| D-5 8 % border offered with the option off | Product: hide or warn; spec §2.5 lets the user declare it per record. | **Deferred** → pm, M3/M4 polish |
| D-6 regime filter reset / empty-month copy | The reset is intended when the new month does not hold the regime; the empty-month copy is product. | **Deferred** → pm + frontend-sr, M3/M4 polish |
| D-7 612 chip clipped | Needs a short regime name; the contract has none (would be a catalog/contract change, not a client table keyed on codes). | **Deferred** → architect + pm, contract 0.7 |
| D-8 Cancel after *Guardar y crear otro* asks | `AppFormFieldModel.markPristine()`; `prepareNext` rebaselines every field, so kept context is not an unsaved change and anything typed after it is. | **Fixed** (`0876aa0` → `099299b`) |
| D-9 onboarding starts regimes this month | M2 onboarding should ask the start month (or default to Jan 1 of the year). | **Deferred** → pm, M2 follow-up |
