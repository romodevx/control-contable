# ADR-0008 — Withholdings are derived, never entered

| Field | Value |
|---|---|
| Status | PROPOSED (WH-02; tech-lead review of the contract PR) |
| Date | 2026-10-15 |
| Related | [MODULE-withholdings](../project-specs/MODULE-withholdings.md) (owner decisions 1–4, §3, §5, §7, OQ-W6, OQ-W7), [withholdings matrix](../research/withholdings-matrix-2026.md), [ADR-0002](0002-tax-engine.md) §4, [ADR-0005](0005-rate-limits-in-postgres.md), [ADR-0007](0007-fiscal-shared-kernel.md), [CONTRACT.md](../api/CONTRACT.md), PLAN §9 (WH-01..WH-17) |
| Owner | architect |

## Context

Since M3/M4 the user types the ISR and IVA withheld on each income and expense. The
server only checks that a typed amount is allowed and below its base, and it warns
(`withholding_differs_from_rate`) when the amount differs from the suggested rate. The
owner decided on 2026-10-15 (MODULE-withholdings, decisions 1–2) that withholdings are
**derived automatically** from a few facts about the operation, on both sides (the user as
emisor and, for a persona moral, as receptor), with **no manual override**.

That decision changes the API of three frozen modules (M3, M4, M5) and creates four design
questions:

1. Where does the decision table live, so that incomes, expenses, the preview and the
   backfill share one copy?
2. Is the derived result stored on the record, or derived on every read?
3. How does the frozen contract move from typed amounts to derived ones without breaking
   the server and the app that are built on it today?
4. How is the preview exposed, given that it serves two modules with two flags?

## Options and decisions

### 1. Placement: `internal/kernel/fiscal` (chosen)

| Option | Verdict |
|---|---|
| A copy in incomes and in expenses | Rejected: the same law in two places drifts silently (the ADR-0007 argument). |
| A `withholdings` module that incomes and expenses import | Rejected: modules never import each other (archtest `cross-module`), and a module with no flag of its own contradicts OQ-W6. |
| **The shared kernel, `fiscal.DeriveWithholdings`** | **Chosen.** More than one bounded context reads it (incomes, expenses and, through the stored amounts, the tax engine) and they must agree, which is exactly ADR-0007's test for the kernel. It is a pure, table-driven function: no I/O, no clock, no floats. Rates come from `RateCatalog` on the record's date; the 2/3, 1/2 and 1 shares are the existing kernel `Share` constants. |

The kernel returns the rule ids, the `cpa_unverified` flag and the plain-words
explanation of each rule. Neither SQL nor Dart holds a copy of the table: the app shows
`Withholding.explanations` and never maps rule ids to text of its own (WH-14 checks this).

### 2. Stored with `table_version` (chosen) vs. derived on read

| Option | Verdict |
|---|---|
| Derive on every read (list, estimate, dashboard) | Rejected. Every estimate would re-run the table, and a change of rate or rule would silently rewrite past months: an estimate the user already filed would change with no record of why. M8 history depends on stored inputs and `records_fingerprint`. |
| **Derive on write, store the amounts with `table_version`** | **Chosen.** Create, update and restore derive; the estimate stays a pure read of `isr_withheld_cents` / `iva_withheld_cents` (MODULE-tax-engine is unchanged on the emisor side). A CPA answer or a new year's rates bumps `table_version`, and `cmd/backfill-withholdings` re-derives the older rows on purpose, with a dry-run report (MODULE-withholdings §5.3). |

Every derived value on the wire is read-only (`Withholding`, OpenAPI `0.10.0`). No request
accepts an amount; WH-06/WH-07 make a typed amount a `400 additionalProperties`.

### 3. Compatibility: expand now, contract in WH-06/WH-07 (chosen)

The target contract (MODULE-withholdings §7) removes `counterparty_type` and the typed
amounts from requests and makes `counterparty_kind` and `withholding` required. Applied in
one step, the regenerated server and Dart client would stop compiling across about thirty
files: the income/expense handlers, presenters, forms and their contract tests. Keeping
them compiling would mean adapter stubs that change behaviour: every income would be
stored with zero withholdings until WH-06 lands, and the forms the app ships today would
get `400` on every save.

| Option | Verdict |
|---|---|
| Hard break now, with adapter stubs | Rejected. The stubs are not "no behaviour change": the typed amounts would be dropped and the estimate in dev would lose its withholdings until WH-06. The stubs would also rewrite most of WH-06/07/11/12 by hand, outside their TDD tasks. |
| **Expand → contract, like the database migration (§5.1)** | **Chosen.** |

**0.10.0 (this ADR, WH-02) is the expand step.** It is additive on the wire:

- New schemas, frozen: `CounterpartyKind`, `OperationKind`, `SupplierKind`,
  `SupplierRegime`, `WithholdingRuleId`, `WithholdingBasis`, `Withholder`,
  `WithholdingExplanation`, `Withholding`, `WithholdingPreview`, `IncomeLastUsed`.
- New **optional** request fields (`counterparty_kind`, `operation_kind`,
  `residential_rent`, `via_platform`; for expenses `supplier_kind`, `supplier_regime`,
  `operation_kind`, `residential_rent`). New **optional** response fields with the same
  names, plus `withholding` and `IncomeDefaults.last_used`.
- `deprecated: true` and `x-deprecated-reason` (naming WH-06 or WH-07) on
  `CounterpartyType`, the `counterparty_type` parameter of `getIncomeDefaults`, the typed
  `isr_withheld_cents` / `iva_withheld_cents` of the four request bodies, and the four
  `IncomeDefaults.*_withholding_*` fields. The generators carry that through: Go gets
  `// Deprecated:`, so staticcheck SA1019 lists every remaining use, and Dart gets
  `@Deprecated`, which the analyzer flags.
- The few existing uses carry a marked suppression
  (`//nolint:staticcheck // SA1019: ... until WH-06`, and
  `// ignore: deprecated_member_use_from_same_package, deprecated until WH-11`). These are
  the work lists of WH-06 and WH-11. There is no behaviour change: the server ignores the
  new facts and omits `withholding` until WH-06/WH-07.
- The two previews are stubbed in `cmd/api/notbuilt.go` (hidden `404`) and listed in
  `pendingOperations` with WH-06 / WH-07.
- Responses keep `isr_withheld_cents` / `iva_withheld_cents` permanently, as read-only
  mirrors of `withholding` for the list totals and the estimate (§7.3).

**The contract steps (each a pre-release patch, 0.10.x, authorized by this ADR).** Each
step lands in the PR that implements the derivation behind it, so the contract never
promises behaviour the server lacks:

| Task | Contract step |
|---|---|
| WH-06 | `IncomeCreate`/`IncomeUpdate`: remove `counterparty_type`, `isr_withheld_cents`, `iva_withheld_cents`; `counterparty_kind` required on create. `Income`: remove `counterparty_type`; `counterparty_kind`, `operation_kind`, `residential_rent`, `via_platform`, `withholding` required. `getIncomeDefaults`: remove the `counterparty_type` parameter and the four `*_withholding_*` fields; `last_used` required. Remove `CounterpartyType`. Cover `previewIncomeWithholdings`. |
| WH-07 | `ExpenseCreate`/`ExpenseUpdate`: remove the typed amounts. `Expense`: `supplier_kind`, `supplier_regime`, `operation_kind` (all nullable), `residential_rent` and `withholding` required. Cover `previewExpenseWithholdings`. |
| WH-11 / WH-12 | They build against the 0.10.0 Prism mock (PG-WH-C). They never read or write a deprecated member, and they merge after WH-06 / WH-07 regenerate the client without them. |

Shipped-client risk: none. All three modules are OFF in every shared environment (OQ-W6),
and no store build reads these fields.

### 4. Two previews, one per module (chosen) vs. one `previewWithholdings?side=`

The spec sketched one operation with a `side=income|expense` parameter. The flag gate
(`adapter/http/flaggate`, P0-15) maps **one route to one module** and answers before the
query string is bound, so that a `400` cannot reveal a hidden module. A single operation
would need either:

- `x-module: platform`, which is security-reviewed (`platform_allowlist_test`) and would
  ship ungated; or
- one of the two flags for both sides, which couples the modules: with `module_incomes`
  OFF, the expense card would break. That contradicts "a disabled module is removable
  without breaking others".

**Chosen:** `previewIncomeWithholdings` (`GET /v1/taxpayers/{taxpayerId}/incomes/withholding-preview`,
`module_incomes`) and `previewExpenseWithholdings` (`GET …/expenses/withholding-preview`,
`module_expenses`). They have the same response (`WithholdingPreview`), share the query
parameters (`components.parameters.Preview*`), and share **one** per-user budget,
`withholding_previews`: 1200 per hour, burst 1200 (ADR-0005), separate from the write
budgets. The card calls the preview debounced on every change; a busy session of 60
records with about 20 previews each still fits. The budgets themselves are wired by WH-06
and WH-07 with their handlers.

### 5. Codes (append-only)

- `WarningCode` += `withholding_cpa_unverified`, `foreign_counterparty_not_supported`,
  `supplier_regime_unknown`, `deduction_requires_remittance`.
  `withholding_differs_from_rate` is **retired**: WH-06 stops sending it, and it stays in
  the enum.
- `CpaTopic` += `withholding_rule_unverified`, `withheld_iva_credit_timing` (WH-09 emits
  them; until then the app reads them as a generic topic).
- New `details[].code` values: `supplier_not_applicable` (`422`, any supplier fact from a
  persona física taxpayer, or `supplier_regime` without `pf_with_activity`),
  `residential_rent_exempt` (`422` on `iva_rate`: housing rent must be `exempt`, LIVA 20
  fr. II; the server rejects instead of silently changing the rate), and `not_for_regime`
  on `via_platform` (`400`, as on `platform_activity`). `withholding_not_applicable` and
  `withholding_exceeds_base` are retired by WH-06/WH-07.

## Consequences

- One decision table, in the kernel, under 100 % coverage and goldens WH-01..WH-12 (WH-04).
  Kernel PRs need tech-lead review (ADR-0007).
- Estimates stay pure reads. A table change is an explicit, versioned, reported
  re-derivation, never a silent rewrite of past months.
- The contract goes through 0.10.0 (expand) and two pre-release patches (contract). Each
  patch re-freezes and goes through the tech-lead `x-frozen` review. The deprecated members
  are listed by the linters until they are gone.
- Between 0.10.0 and WH-06, a client that sends only the new facts is still asked for
  `counterparty_type`. The 0.10.0 request examples send both.
- Risk carried from the owner's decision (OQ-W3): with no override, a CFDI that shows a
  different withholding is not reflected. The estimate uses the legal figure, and the help
  text explains why.

## Amendment: the contract step as implemented by WH-06 / WH-07 (0.11.3)

Three decisions refine §2 and §3 above; the wire behaviour is the one §3 promised.

1. **Retired request members answer `400`, but stay in the schema until WH-11 / WH-12.**
   Removing `isr_withheld_cents`, `iva_withheld_cents` and `counterparty_type` from the
   request schemas regenerates the Dart client without them, and the forms the app ships
   today still read them (`// ignore: deprecated_member_use…, deprecated until WH-11`):
   `mobile-ok` would fail on a backend task. So 0.11.3 keeps them in the schema, and the
   server answers any value of the typed amounts with `400 validation_failed`, detail
   `additionalProperties` on that field: exactly what a schema without them answers, so no
   client sees a change when WH-11 / WH-12 drop them with the new forms.
   `counterparty_type` stays required by the schema and is **ignored** (its stored column
   mirrors `counterparty_kind` and `via_platform` until WH-08 drops it): rejecting a field
   the schema requires would make the schema unusable. `counterparty_kind` (incomes) and
   `supplier_kind` (persona moral expenses) are required by the server (`400 required`), as
   `previewExpenseWithholdings` already documents for `supplier_kind`. `withholding` and the
   other response members stay optional in the schema for the same reason; they are sent on
   every record written since 0.11.3, and only a legacy row (before the WH-08 backfill) lacks
   `withholding`.
2. **The stored derivation includes how it was computed.** Migration
   `20261017100000_withholdings_detail` adds `withholding_isr_rate`,
   `withholding_iva_basis` and `withholding_withholder` next to the WH-03 rule ids and
   `table_version`. The rule ids alone do not determine the figures shown: a platform sale's
   ISR rate (2.5 % or 20 %) and IVA share (half or all) depend on whether the seller gave the
   platform an RFC, a 625 option the user can change after the write. Storing them keeps §2:
   a read never re-runs the table. `net_cents` is computed from the stored amounts.
3. **The kernel words the explanations** (`fiscal.Explain`, per side), from the stored
   record, so a read shows the same lines as the write and the preview. The contract
   examples are its test oracle.

## Amendment 2: the contract cleanup of WH-11 / WH-12 (0.12.0)

| Field | Value |
|---|---|
| Date | 2026-10-14 |
| Owner | architect (PLAN WH-11 / WH-12, contract cleanup) |

The new mobile forms (#137), the derivation (#139) and the backfill (#142) are merged, so the
members Amendment 1 kept only for the shipped forms go. The wire behaviour of a well-formed
request does not change; the schemas now say what the server already did.

1. **Removed.** `IncomeCreate` / `IncomeUpdate`: `counterparty_type`, `isr_withheld_cents`,
   `iva_withheld_cents`. `ExpenseCreate` / `ExpenseUpdate`: `isr_withheld_cents`,
   `iva_withheld_cents`. `getIncomeDefaults`: the `counterparty_type` parameter.
   `IncomeDefaults`: `isr_withholding_allowed`, `isr_withholding_rate`,
   `iva_withholding_allowed`, `iva_withholding_rate`. `Income`: `counterparty_type`. The
   `CounterpartyType` schema. Any of the removed body members answers
   `400 validation_failed` / `additionalProperties` from the router, so the adapters' own
   rejection of the typed amounts is gone. `counterparty_type` used to be required and
   ignored; it now answers that `400` too, the one visible change, and the app stopped
   sending it in the same PR.
2. **Required.** `IncomeCreate.counterparty_kind` (the server already answered
   `400 required`). `Income`: `counterparty_kind`, `operation_kind`, `residential_rent`,
   `via_platform`, `withholding` (nullable, item 3). `IncomeDefaults.last_used`. `Expense`:
   `supplier_kind`, `supplier_regime`, `operation_kind` (all three nullable),
   `residential_rent`, `withholding` (nullable, item 3). `isr_withheld_cents` /
   `iva_withheld_cents` stay in responses as the read-only mirrors (§7.3).
3. **A record the backfill has not derived is served; no deploy order.** The first draft
   answered `500` for a legacy record and made the WH-08 backfill a deploy precondition. That
   was rejected in review: the backfill leaves its "need user input" rows underived for good
   (a regime no longer held, a date out of range), so those records, and every list page
   holding one, would answer `500` forever and could never be opened to be fixed; and any
   environment deployed before its run would fail the same way. Instead:
   - **Facts.** The incomes use case fills a legacy income's missing facts with the
     backfill's own §5.2 mapping (`legacyFacts`: `counterparty_type`, then the regime
     default) on every load: get, list, update, restore and the create replay. A diff-only
     `PATCH` therefore applies to the facts the read served, and an edit derives and stores
     them. `counterparty_type` is `NOT NULL` with a CHECK, so the mapping is total. A legacy
     expense's supplier facts were already nullable.
   - **Withholding.** `Income.withholding` and `Expense.withholding` are required but
     nullable: `null` only for a record stored before automatic derivation that the server
     has not derived. A read never derives (§2 keeps reads pure, the table version auditable,
     and the estimate still uses the stored amounts until the record is derived), so serving
     a figure the estimate does not use was ruled out. The backfill, or an edit, derives it.
   - A record with facts but a corrupt derivation (an unknown withholder) is still a `500`.
4. **Dead code removed with the members.** The income suggestion table
   (`domain.PolicyFor`, `WithholdingPolicy.Suggest`) only fed the removed `IncomeDefaults`
   fields, so it is deleted; `fiscal.DeriveWithholdings` is the only withholding table left.
5. **Not done here.** The `incomes.counterparty_type` column stays (writes keep its mirror)
   until the backfill has run in every environment; its drop is a separate migration
   (PLAN WH-08).

Version: **0.12.0**, a minor bump. Removing request members and response fields is breaking
for a client that sent or read them; pre-release with no store build, the minor bump is the
version signal ([CONTRACT.md](../api/CONTRACT.md), "Changing a frozen node").
