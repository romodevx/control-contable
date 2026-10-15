# WH-14: Tech-lead gate of the withholdings module

- **Date:** 2026-10-14
- **Branch:** `feature/withholdings-qa-fixes` (PR #152 → `dev`): the WH-13 QA (goldens through the API,
  the "IVA acreditable" label fix, the QA report), on top of all of WH already on `dev`: kernel
  `fiscal.DeriveWithholdings` #124, params #123, migration #136, forms #137, engine 1.1.0 #138, use
  cases #139, backfill #142, contract cleanup 0.12.0 #146. `origin/dev` merged in (3 commits behind).
- **Scope:** A) PR #152. B) The PLAN WH-14 criteria over the module: kernel purity (no I/O, no clock),
  SOLID, no second copy of the decision table in SQL or Dart, determinism, no floats, a coherent
  legacy-`null` path. C) The two QA follow-ups (O1, O2).
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language (idiomatic Go
  wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gate (WH-15) runs at the same time on its own branch.

## Verdict: **PASS**

No blocker. The decision table lives once, in `internal/kernel/fiscal/derive.go`; the database holds
only the vocabulary (enum CHECKs, rule-id shape) and the mobile app only decides which question to show,
while every amount, rule id and explanation it displays comes from the server preview. One small
legacy-path defect (QA O2) was fixed on this branch with a failing test first. O1 stays a recorded
follow-up for design review. Three non-blocking follow-ups below.

PR #152: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `go list -deps ./internal/kernel/fiscal` | stdlib + `internal/platform/money` only; archtest `RuleKernelImports` guards it |
| `grep time.Now\|rand\.\|float` over the kernel, incomes, expenses, taxengine, `cmd/backfill-withholdings` (non-test) | none; the only clock is the injected day (`withIncomesDay`, `withExpensesDay`) |
| Floats on the wire / in the DB | `isr_rate` is a decimal string; `withholding_isr_rate numeric(7,6)`; Dart `double` only in layout constants |
| Kernel ↔ contract vocabulary | `cmd/api/withholding_codes_test.go`: CounterpartyKind, OperationKind (minus `platform`), WithholdingBasis, every rule id against `WithholdingRuleId` |
| `go test` kernel, incomes, expenses, taxengine, `cmd/...` | green |
| `make mobile-check` | green: 1,382 passed, 39 skipped |
| New O2 test before the fix | **red**: E3 `null`, expected `commission` (the category's served default) |

## A) PR #152

| Item | Assessment |
|---|---|
| `withholdings_qa_integration_test.go` | Good. Every WM golden WH-01..WH-12 enters through `createIncome`/`createExpense` on the real composition root as `cc_app`, asserts ISR/IVA/net/rule ids/`cpa_unverified`, and asserts the preview equals what was stored (one table, two entry points). WH-01-EST, WH-05b and WH-11-JUN/JUL close the loop through the estimate. `UseNumber` keeps query amounts integral. |
| F1 "IVA acreditable" | Right fix: since engine 1.1.0 `creditable_paid_cents` is paid this month + withheld last month, so "pagado" was false. Regression test pins both the label and the amount. |
| QA report | Honest about the environment (native PG instead of Docker), wipe, what went through the UI, the one SQL step (comp entitlement). |

## B) Module criteria

| Criterion | Result |
|---|---|
| Kernel purity | Pass. `DeriveWithholdings(in, catalog)` is a pure function: rates come through the `WithholdingCatalog` port, the date is an input, errors are typed (`Violations`, `ErrInvalidParameter`, `ErrRateNotValidOnDate`). |
| SOLID | Pass. ISR and IVA rules are separate first-match functions; rates are data (`params/data/2026/v1.yaml`) behind a consumer-side interface; each module only maps its facts to `WithholdingInput`. |
| One decision table | Pass. SQL: enum CHECKs and `withholding_rule_ids_valid` (shape only), the category → `operation_kind` seed served as `default_operation_kind`. Dart: question visibility (spec §2.1/§2.2) and preselection; no rate, share or rule. |
| Determinism | Pass. No clock, no randomness, map lookups only (no map iteration in output order); `ordered` fixes the ambiguity order. |
| No floats | Pass (see Evidence). |
| Legacy `null` path | Pass after O2. Incomes: read and backfill share `legacyFacts`; a read never derives. Expenses: the edit requires E1 (runbook §1) and now shows the server's E3 default. |

## C) QA follow-ups

- **O2, legacy expense E3 shows no answer: fixed.** `ExpenseSupplierFacts.showServerDefault` (called
  from `ExpenseFormModel.applyCatalog` when editing) shows the chosen category's served
  `default_operation_kind` for a stored expense without an operation, marked pristine (it is what the
  server already applies when `operation_kind` is omitted, so it is not an edit) and still moved by a
  category change. No copy of the seed in the app. Test:
  `expense_supplier_test.dart` "a legacy expense without an operation shows the category default…".
- **O1, onboarding start month far below the regimes: follow-up, not fixed.** It belongs to
  module_taxpayers (MODULE-taxpayers §160), not withholdings, and moving a field in onboarding is a
  design-review decision. Owner: frontend-sr + design review.

## Follow-ups (non-blocking)

1. **Contract text for `IvaComponents.creditable_paid_cents`.** The name says "paid" but since engine
   1.1.0 it also holds last month's withheld IVA (`taxengine/domain/iva.go`), and the schema has no
   description. Add one in the next contract patch (architect); the UI already says "IVA acreditable".
2. **Sibling presenters.** `incomes/adapter/http/withholding.go` and `expenses/adapter/http/withholding.go`
   both build `gen.Withholding`; the expense copy rejects an unknown withholder, the income copy maps it
   to `""`. Extract one presenter or add the same guard.
3. **SQL enum drift.** Kernel ↔ contract enums have drift tests; the SQL CHECK lists
   (`operation_kind`, `counterparty_kind`, `supplier_kind`) have rejection tests only. Add one test that
   every contract value is accepted by the CHECKs.
