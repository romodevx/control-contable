# ADR-0009 — The annual estimate is a sub-context of the tax engine module

| Field | Value |
|---|---|
| Status | PROPOSED (ANN-02; tech-lead review of the contract PR) |
| Date | 2026-10-15 |
| Related | [MODULE-annual](../project-specs/MODULE-annual.md) (owner decisions 1–3, §2, §5, §10, OQ-A1..OQ-A9), [annual return research](../research/annual-return-2026.md), [ADR-0001](0001-stack.md) (Dependency Rule), [ADR-0002](0002-tax-engine.md) §1–§5, [ADR-0005](0005-rate-limits-in-postgres.md), [ADR-0007](0007-fiscal-shared-kernel.md), [CONTRACT.md](../api/CONTRACT.md), `backend/internal/archtest`, PLAN §ANN (ANN-02..ANN-27) |
| Owner | architect |

MODULE-annual §10.5 named this decision "ADR-0008"; that number went to the withholdings
(WH-02), so the annual placement is ADR-0009.

## Context

`module_annual` estimates the declaración anual for every regime that files one (PF art. 152
aggregation, RESICO PF informative check, PM 601, RESICO PM, explanation-only sections). It
needs, on every read:

- the M5 domain: `Ledger`/`Line` (the breakdown is the computation trace, ADR-0002 §1),
  `money.Exact` and the one-rounding rule (ADR-0002 §4), the binary-search tariff lookup,
  the parameter loader (`isr.art152.annual`, deduction caps, deadlines; ANN-04);
- the M5 use-case ports: `LedgerReader` (12 `MonthSums` per regime), the `Estimator` range
  (the default provisional payments, §2.5) and `TaxInputReader` (`ptu_paid`,
  `prior_year_losses`, `annual_depreciation`, `interest_withholding`, ...);
- four tables of its own (ANN-03: `personal_deductions`, `salary_statements`,
  `annual_filed_payments`, `annual_inputs`).

ANN-07 already landed the pure personal-deductions calculator in
`internal/taxengine/domain/annual`, provisionally, pending this ADR.

The spec recommended pure strategies in `internal/taxengine/annual` and the inputs, use case
and HTTP in a new `internal/annual`. The archtest rule `cross-module` (no module package
imports another module) makes the second half impossible as written: a use case in
`internal/annual` cannot import the strategies or the `Ledger` it must assemble.

## Options

| Option | Verdict |
|---|---|
| A. New module `internal/annual` with its own domain, re-implementing `Ledger`, rounding and the tariff lookup | Rejected. Two copies of the rounding rule and the tariff search drift silently; the annual tariff is asserted equal to the December art. 106 table *by the same lookup* (ANN-04). |
| B. New module `internal/annual`; promote `Ledger`, `Exact`, the tariff lookup and the params loader to the shared kernel | Rejected. The kernel is for law several bounded contexts must agree on (ADR-0007); these are the tax engine's computation internals with one extra consumer. Promoting them widens the kernel's tech-lead-gated surface for no second business consumer. |
| C. New module `internal/annual` that consumes the engine through a wide consumer-owned port (`AnnualEngine.Compute(AnnualInput) AnnualResult`) wired in `cmd/` | Rejected. Every section, line, disclaimer and topic type would be declared twice (port DTOs and engine types) with a translator in `cmd/`, for a module that cannot exist without `module_tax_engine` anyway (flag prerequisite, §6). |
| **D. A sub-context inside the tax engine module** | **Chosen.** |

## Decision

The annual estimate lives in `internal/taxengine` as its own sub-tree, never mixed into the
monthly files:

| Layer | Package | Contents | Task |
|---|---|---|---|
| Domain (pure) | `internal/taxengine/domain/annual` | Enums, deductions calculator (ANN-07, already here), one strategy per section kind (`pf_art152`, `pf_resico_check`, `pm_601`, `pm_resico`, `explanation`) behind a `SectionStrategy` interface and a registry (open/closed: a new section is a new file + one registry line). `(Input, Params)` only: no clock, no I/O, no float. Appends lines through the M5 `Ledger`. | ANN-07..ANN-12 |
| Use case | `internal/taxengine/usecase/annual` | `EstimateAnnual` (one `LedgerReader` pass, at most one `Estimator` range call, payment precedence filed → tax input → app estimate, section assembly, legend, disclaimers, statuses) and the input use cases (deductions with restore, salary statements, annual inputs, filed payments; access policy `402`, person-type, archive and regime-window checks). Consumer-side ports for its four stores. | ANN-13, ANN-14 |
| Infrastructure | `internal/taxengine/infra/annualinputs` | GORM repositories of the four ANN-03 tables on the unit of work in `ctx`, as `cc_app` under RLS. | ANN-13 |
| Adapter | `internal/taxengine/adapter/http/annual` | Strict-server handlers of the 15 `x-module: module_annual` operations, presenters, budgets. | ANN-15 |

Rules that keep it a removable vertical slice:

1. **Nothing in the monthly engine imports `.../annual`.** The dependency is one-way
   (annual → monthly domain and ports). An archtest rule (ANN-14) asserts it, so deleting the
   four `annual` directories, their `cmd/` wiring and the ANN-03 migrations removes the module
   without touching M5.
2. **The flag is the HTTP boundary.** Every annual operation is `x-module: module_annual`;
   `module_annual` has prerequisites `module_tax_engine` and `module_taxpayers`
   (`flags.Prerequisites`, ANN-22). OFF or implied OFF → `404` before any input check.
3. **Derived on every read, never stored** (ADR-0002 §5): the four tables hold user inputs
   only. Same inputs + parameters + engine version → byte-identical JSON (MODULE-annual §2.11).
4. **One source per fact.** Losses, PTU paid, depreciation and interest withholding stay M5
   tax inputs (`putTaxInputs`; `annual_depreciation` is catalog data, ANN-03), mirrored
   read-only in `getAnnualInputs`. Filed payments are annual-only and do not replace
   `prior_provisional_payments` (OQ-A1: precedence, not unification).

## Contract (0.11.0, frozen by ANN-02)

New tag `Annual`, 15 operations, all `x-module: module_annual`, bearer auth,
`Cache-Control: no-store` on every response:

| Operation | Route |
|---|---|
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown?section=` |
| `getAnnualInputs` / `putAnnualInputs` | `GET`/`PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` |
| `listSalaryStatements` / `createSalaryStatement` | `GET`/`POST /v1/taxpayers/{taxpayerId}/salary-statements` |
| `updateSalaryStatement` / `deleteSalaryStatement` | `PATCH`/`DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` |
| `listPersonalDeductions` / `createPersonalDeduction` | `GET`/`POST /v1/taxpayers/{taxpayerId}/personal-deductions` |
| `updatePersonalDeduction` / `deletePersonalDeduction` | `PATCH`/`DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` |
| `getFiledPayments` / `putFiledPayments` | `GET`/`PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` |

Contract choices this ADR records:

- **Own legend.** `AnnualLegend` is a new `const` ("Esto es un aproximado; tu contador te dará
  las cifras reales."), not the monthly `Legend`: two surfaces, two frozen texts.
- **Own breakdown line.** `AnnualBreakdownLine` mirrors the frozen `BreakdownLine` but adds
  `regime_code` (per-chapter lines of the aggregated art. 152 section) and widens `source`
  with `salary_statements`, `personal_deductions`, `annual_inputs`, `filed_payment` and
  `estimated_by_app` (§2.5 labels every payment by its source). Re-freezing `BreakdownLine`
  for this would change M5's frozen surface for an annual-only need.
- **Structured missing inputs.** `AnnualMissingInput` (`key`, `target`, `regime_code`,
  `months`) instead of the spec's `filed_payment:<regime>:<YYYY-MM>` string, so the client
  routes the "Completa tus datos" CTA without parsing.
- **Reused, append-only vocabularies.** `Disclaimer`/`DisclaimerCode` and `CpaTopic` gain the
  annual codes and topics; `WarningCode` gains `cash_not_deductible` and
  `salary_regime_not_held` (OQ-A2). No new `ErrorCode`.
- **Status codes.** Future `fiscal_year` → `400 out_of_range`; a fiscal year before 2026 →
  `422 parameters_unavailable` (as M5/M8); a deduction date before 2026 or after today →
  `400 out_of_range` (the record-date rule of `BusinessDate`). `422` detail codes:
  `parameters_unavailable`, `person_type_not_allowed`, `regime_not_active`,
  `payments_not_applicable` (a held regime with no provisional ISR, e.g. 605),
  `taxpayer_archived`, `section_not_present`, `limit_reached`.
- **Access.** Creates, edits and PUTs need `trial`/`pro`/`comp` (`402` when read-only,
  OQ-A9); they never use a trial record slot. Delete and restore are allowed to a read-only
  user, as for records (CONTRACT.md 0.6.1: read-only never makes a delete one-way).
- **Bounded lists.** At most 500 live personal deductions and 20 salary statements per
  taxpayer and fiscal year (`422 limit_reached` on create, on an edit that moves a row to
  another year, and on restore: Undo must not be a way past the bound), so the unpaginated lists and the
  O(rows) deduction pass stay inside the p95 < 500 ms target (§2.11).
- **Budgets** (ADR-0005, wired by ANN-15): estimate and breakdown count in `estimate_reads`
  (+1 when the 12-month `Estimator` range runs; M5 FU5 precedent); `getFiledPayments` and
  `putFiledPayments` likewise (they show the app estimates); the other reads in
  `annual_reads` (600/h); every write in `annual_writes` (300/h).

Until ANN-15 builds the handlers, the 15 operations are `notBuilt` stubs (hidden `404`) and
listed in `pendingOperations` under ANN-15.

## Consequences

- The tax engine module grows by one sub-tree; its owners review annual PRs. The one-way
  archtest rule keeps M5 unaware of it.
- No kernel change and no M5 frozen node changes: `annual_depreciation` is catalog data
  (ANN-03), served by the existing `getTaxInputs`/`putTaxInputs`.
- The mobile client gets a separate `AnnualBreakdownLine` model; ANN-18 maps it onto the
  M5-26 widgets (same fields plus `regime_code`).
- If a second consumer of the annual figures appears outside the tax engine (for example a
  CPA export), it consumes them through a port like the dashboard does with the monthly
  estimate; that is not a reason to move the code.
