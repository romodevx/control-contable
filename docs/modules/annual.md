# Module: Annual estimate

| | |
|---|---|
| Flag | `module_annual` (OFF by default; requires `module_tax_engine`, and through it `module_taxpayers`; registered by migration `20261019100000_register_module_annual` (ANN-22); in `demo_full` only). See [Flags, prerequisites and presets](#flags-prerequisites-and-presets) |
| Spec | [MODULE-annual.md](../project-specs/MODULE-annual.md); legal research: [annual return 2026](../research/annual-return-2026.md) |
| Design | [ADR-0009](../adr/0009-annual-engine-placement.md) (Accepted: a sub-context of the tax engine module), [ADR-0002](../adr/0002-tax-engine.md) ([engine version log](../adr/0002-tax-engine.md#engine-version-log): 1.1.1) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml) **0.12.0**, tag `Annual`, `x-module: module_annual` (15 operations, frozen by ANN-02 in 0.11.0, unchanged since). Reference generated from it: [API reference and errors](#api-reference-and-errors) |
| Status | **Built**: ANN-01..ANN-26 done. QA ANN-23 pass, tech-lead ANN-24 **PASS**, security ANN-25 **PASS** ([Gates](#gates)). CPA sign-off of the ⚠️ goldens and topics (ANN-27, OB5) open |
| User guide | [Tu declaración anual estimada](../guides/annual-user-guide.md) (Spanish, the screens' own words) |
| Operations | [Yearly parameter update §8: the annual return](../runbooks/tax-params-update.md#8-the-annual-return-module_annual-rmf-deadlines-and-rule-renewals) (RMF of the next year, deadlines, rule renewals) |
| Code | `backend/internal/taxengine/domain/annual/` (strategies, deductions, `explain/`), `backend/internal/taxengine/usecase/annual/`, `backend/internal/taxengine/adapter/http/annual/` (handlers, budgets), `backend/internal/taxengine/infra/annualinputs/`, `backend/internal/taxengine/infra/annuallimits/`, `backend/cmd/api/annual.go`, migrations `20261016120000_regime_annual_depreciation.sql`, `20261016121000_annual_inputs.sql`, `20261017102000_salary_statements_move_year.sql`, `20261019100000_register_module_annual.sql`, `mobile/lib/features/annual/` |

## Contract highlights

- `getAnnualEstimate` (`GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}`):
  sections in fixed order (`pf_art152` aggregating 605/606/612/625 non-definitive,
  `pf_resico_check` informative only, `pm_601`, `pm_resico`, then `explanation` per other
  regime), each with `status` (`complete` / `incomplete` / `informational`),
  `missing_inputs` (with the `target` screen), `obligation`, `deadline`, `tax` (base, ISR,
  credits, `balance` `a_cargo` / `a_favor` / `cero`), PM extras (PTU base and amount,
  suggested coeficiente for next year) and CPA topics. Always the legend "Esto es un
  aproximado; tu contador te dará las cifras reales." and the disclaimers.
  `headline_sections` names the balances to show. Current year → `year_status:
  in_progress`, records through `data_through`, no extrapolation.
- `getAnnualBreakdown` (`…/breakdown?section=`): the trace, one `AnnualBreakdownLine` per
  step with legal basis and `source` (each payment line says `filed_payment`,
  `tax_input:prior_provisional_payments` or `estimated_by_app`).
- Inputs: `getAnnualInputs` / `putAnnualInputs` (prior-year income, other exempt income,
  told employer; losses, PTU paid, depreciation and interest withholding mirrored read-only
  from the M5 tax inputs); salary constancias (list, create, edit, delete; no employer RFC);
  personal deductions by category (list with registered totals, create, edit, soft delete,
  restore); filed payments per regime and year (12 months with filed → tax input → app
  estimate precedence; PUT replaces the confirmed months).
- Flags: `module_annual`, `module_tax_engine` or `module_taxpayers` OFF → every route `404`,
  indistinguishable from an unknown taxpayer.
- Access: reads open to read-only users and archived taxpayers; creates, edits and PUTs
  answer `402` when read-only (they never use a trial record slot); delete and restore are
  allowed. Budgets: `estimate_reads` (estimate, breakdown, filed payments), `annual_reads`
  600/h, `annual_writes` 300/h; every response carries `Cache-Control: no-store`.

## Engine notes (ANN-14)

- One read-only unit of work per estimate: one ledger pass for the year (through
  `data_through`), the tax inputs, the annual inputs and filed payments, and at most one
  `Estimator` range (January..`data_through`), only when some month has neither a filed
  payment nor a `prior_provisional_payments` tax input.
- **Chained defaults.** A cumulative regime's monthly estimate (612, 601, RESICO PM) is
  YTD tax minus the *typed* prior payments, so summing raw monthly estimates would pay the
  year's tax several times. The default of month *m* is therefore
  `max(0, isr(m) + Σ typed prior payments before m − Σ payments the annual credits before m)`,
  exactly the monthly figure had the credited payments been typed. A monthly estimate missing
  only the prior payments is not "incomplete" for the annual; any other missing input is
  (the month is listed under `filed_payment`).
- Sections run through an ordered registry (`taxengine.AnnualSections`); the annual golden
  gate runs AN-01..AN-08 through the use case. Nothing in the monthly engine imports the
  annual packages (`TestMonthlyEngineNeverImportsTheAnnual`).
- **Balance = ISR as shown − credits**, to the centavo (ANN-16, #148): the credits come off
  the rounded ISR, so a half-centavo ISR paid exactly is `cero`, not a favor of 0.01. Engine
  1.1.1 (#156) applied the same rule to the monthly balance
  ([log](../adr/0002-tax-engine.md#engine-version-log)).

## Mobile input screens (ANN-19..21)

Reached from "Datos para tu anual" on the annual screen and from each "Falta" link of an
`incomplete` section (`/anual/{taxpayerId}/{ejercicio}/…`, routes registered only while
`module_annual` is on). Containers follow `ui-forms-and-modals.md` by field count:

| Screen | Location | Container |
|---|---|---|
| Personal deductions: groups per category, registered vs applied (`pf_art152` `personal_deductions_cents`) and the breakdown's "Tope aplicado" lines; delete + persistent Undo (restore) | `…/deducciones` | list |
| New personal deduction: category first, then only its fields (cash toggle where cash is not deductible, beneficiary + UMA attestation, donee, tuition student/level/scholarship, optional CFDI folio) | `…/deducciones/nueva` | own screen (up to 7 fields) |
| Salary constancias: year totals, one row per employer (tap to edit), delete after a confirmation (hard delete); no employer RFC | `…/constancias` | sheet (≤ 6 fields) |
| Filed payments of one regime: 12 inline rows prefilled with the app's estimate ("Estimado por la app — confírmalo"), "Confirmar todos" | `…/pagos/{regimeCode}` | inline rows |
| Year facts (PF): prior-year income (empty = unknown), other exempt income, employer notice; the regimes' losses, PTU paid and depreciation read-only with a link to their tax-inputs form | `…/datos` | own screen |

Writes never use a trial slot; a `402` opens the paywall and marks the session read-only,
after which the add/save actions show the paywall without calling the server (UX only).
Delete and restore stay available to a read-only user.

## Flags, prerequisites and presets

| Flag OFF | Effect on the annual |
|---|---|
| `module_annual` | Every one of the 15 routes answers `404 not_found`; `GET /v1/flags` omits it; the "Declaración anual" entry and the `/anual/…` routes are gone. Rows are kept |
| `module_tax_engine` | Annual implied OFF (prerequisite, `flags.Prerequisites`): same `404`, entry hidden (History is hidden too) |
| `module_taxpayers` | Annual implied OFF through the engine |
| `module_incomes` / `module_expenses` | No effect on the figures: the ledger is still read (MODULE-incomes OQ-8) |

- The `404` of a flag OFF is indistinguishable from an unknown or foreign taxpayer, a
  malformed request included (`TestAnnualRoutes_needTheEngineAndTaxpayersFlags`).
- Presets: **`demo_full` only** (OQ-A7; `annual_preset_smoke_integration_test.go`). Not in
  `demo_core` nor `demo_monetization`.
- Budgets (ADR-0005): `estimate_reads` for estimate, breakdown and filed payments (+1 when
  the 12-month `Estimator` range runs), `annual_reads` 600/h for the other reads,
  `annual_writes` 300/h for every write; spent → `429 rate_limited` with `Retry-After`.
- **Enable in a shared environment** only when all of these hold:
  1. ANN-24 and ANN-25 passed (both have, 2026-10-15);
  2. the tax engine's own preconditions, first the CPA sign-off
     ([M5-31 §5](../reviews/M5-31-security-gate.md#5-preconditions-to-enable-module_tax_engine-in-staging-or-production)),
     since the annual cannot be ON without it;
  3. a reviewed parameter set for the fiscal year being estimated, and the annual items of the
     [yearly update §8](../runbooks/tax-params-update.md#8-the-annual-return-module_annual-rmf-deadlines-and-rule-renewals)
     done for it (RMF of the next year read, deadlines and rule renewals applied or the
     disclaimers kept);
  4. the owner accepts the open ⚠️ topics shown with the strong CPA banner, or ANN-27 has
     closed them.

```sh
make -C backend flags ARGS="enable module_annual"        # also needs module_tax_engine and module_taxpayers ON
make -C backend flags ARGS="apply-preset demo_full"      # local demo with every module
```

## API reference and errors

The 15 operations (estimate, breakdown, annual inputs, salary statements, personal
deductions with restore, filed payments), their schemas, parameters and examples live only
in the contract. Render them instead of copying them here:

```sh
make api-docs     # build/api-docs/index.html, rendered from docs/api/openapi.yaml (tag Annual)
make mock         # Prism mock of the contract, for the screens without a backend
```

The [Contract highlights](#contract-highlights) above are a reading guide, not the
reference; when they disagree, `openapi.yaml` wins and this page is the bug.

## Errors

Every error uses the single envelope `ErrorResponse`. `details[].code` carries the rule
(`out_of_range` for a future `fiscalYear`, `parameters_unavailable` before 2026,
`person_type_not_allowed`, `regime_not_active`, `payments_not_applicable`,
`taxpayer_archived`, `section_not_present`, `limit_reached`). Amounts, constancia figures
and deduction details are never logged or echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_annual -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 422 | `business_rule_violation` | The fiscal year is before 2026, the first fiscal year with transcribed tax parameters (`parameters_unavailable`). A well-formed request the engine cannot answer, hence `422`, as `getTaxpayerEstimate` and `getHistory`. |
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getAnnualEstimate` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` | 500 | `internal_error` | Unexpected server error. |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 422 | `business_rule_violation` | The fiscal year is before 2026 (`parameters_unavailable`), or the estimate of that year has no such section (`section_not_present`). |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 429 | `rate_limited` | Rate limit exceeded. |
| `getAnnualBreakdown` | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown` | 500 | `internal_error` | Unexpected server error. |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 422 | `business_rule_violation` | The fiscal year is before 2026, the first fiscal year with transcribed tax parameters (`parameters_unavailable`). A well-formed request the engine cannot answer, hence `422`, as `getTaxpayerEstimate` and `getHistory`. |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getAnnualInputs` | `GET /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 500 | `internal_error` | Unexpected server error. |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 429 | `rate_limited` | Rate limit exceeded. |
| `putAnnualInputs` | `PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` | 500 | `internal_error` | Unexpected server error. |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 422 | `business_rule_violation` | The fiscal year is before 2026, the first fiscal year with transcribed tax parameters (`parameters_unavailable`). A well-formed request the engine cannot answer, hence `422`, as `getTaxpayerEstimate` and `getHistory`. |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 429 | `rate_limited` | Rate limit exceeded. |
| `listSalaryStatements` | `GET /v1/taxpayers/{taxpayerId}/salary-statements` | 500 | `internal_error` | Unexpected server error. |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 409 | `conflict` | The client `id` is reused: by the caller with other fields, another taxpayer or a deleted row (`details`: `id` / `id_reused`), or it belongs to another user (no `details`). |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 429 | `rate_limited` | Rate limit exceeded. |
| `createSalaryStatement` | `POST /v1/taxpayers/{taxpayerId}/salary-statements` | 500 | `internal_error` | Unexpected server error. |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateSalaryStatement` | `PATCH /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 500 | `internal_error` | Unexpected server error. |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 422 | `business_rule_violation` | The taxpayer is archived; archive freezes its records (OQ-5). |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `deleteSalaryStatement` | `DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` | 500 | `internal_error` | Unexpected server error. |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 422 | `business_rule_violation` | The fiscal year is before 2026, the first fiscal year with transcribed tax parameters (`parameters_unavailable`). A well-formed request the engine cannot answer, hence `422`, as `getTaxpayerEstimate` and `getHistory`. |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 429 | `rate_limited` | Rate limit exceeded. |
| `listPersonalDeductions` | `GET /v1/taxpayers/{taxpayerId}/personal-deductions` | 500 | `internal_error` | Unexpected server error. |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 409 | `conflict` | The client `id` is reused: by the caller with other fields, another taxpayer or a deleted row (`details`: `id` / `id_reused`), or it belongs to another user (no `details`). |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 429 | `rate_limited` | Rate limit exceeded. |
| `createPersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions` | 500 | `internal_error` | Unexpected server error. |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updatePersonalDeduction` | `PATCH /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 500 | `internal_error` | Unexpected server error. |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 422 | `business_rule_violation` | The taxpayer is archived; archive freezes its records (OQ-5). |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `deletePersonalDeduction` | `DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}` | 500 | `internal_error` | Unexpected server error. |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 429 | `rate_limited` | Rate limit exceeded. |
| `restorePersonalDeduction` | `POST /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore` | 500 | `internal_error` | Unexpected server error. |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getFiledPayments` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 500 | `internal_error` | Unexpected server error. |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 400 | `validation_failed` | Input does not match the contract (malformed fiscal year, regime code, path id, query parameter or body; unknown field; bad enum; negative amount), a future fiscal year (`out_of_range`), or a field rule of the operation (`details[].code`: see it). |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 404 | `not_found` | The taxpayer, constancia or deduction does not exist, belongs to another user or is purged, or `module_annual` is effectively OFF (`module_annual`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 422 | `business_rule_violation` | Well-formed but violates an annual rule (`details[].code`, the subset each operation lists): `parameters_unavailable` (fiscal year before 2026), `person_type_not_allowed` (a persona moral writing personas físicas' data), `taxpayer_archived`, `regime_not_active` (regime not held in the year, or a month outside its window), `payments_not_applicable` (a regime with no provisional ISR payments), `limit_reached` (list bound of the year). |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 429 | `rate_limited` | Rate limit exceeded. |
| `putFiledPayments` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_annual -->


## Gates

| Gate | Verdict | Notes |
|---|---|---|
| ANN-23 QA ([report](../../.gstack/qa-reports/qa-report-annual-2026-10-14.md), [PR #153](https://github.com/romodevx/control-contable/pull/153)) | Pass, one accessibility fix | Android emulator against its own empty stack, every datum entered through the screens: PF 612 + 605 (refund, then constancia, caps, cash exclusion, donation without and with prior-year income, Confirmar todos), PM 601 (PTU, CU, typed payment), read-only (paywall on writes, delete + Undo open), `module_annual` OFF and `module_tax_engine` OFF. Every figure matched the hand arithmetic and the server. Fix: a section with one missing input was announced by TalkBack as one button (`Semantics(container: true)` on `AnnualLinkRow`, regression test first). New suites: `annual_qa_integration_test.go`, `annual_goldens_ui_test.go` (AN-01..AN-08 on screen) |
| ANN-24 tech lead ([review](../reviews/ANN-24-tech-lead-gate.md)) | **PASS** after one OCP fix | The assembler named section kinds in 6 places; the facts are now fields each strategy sets on `ComputedSection`, guarded by `registry_ocp_test.go`. Pure strategies, ports only, no floats, no clock or random, `annual_estimate` always first |
| ANN-25 security ([verdict on PR #153](https://github.com/romodevx/control-contable/pull/153#issuecomment-6186092941)) | **PASS**, no blocking finding | RLS ENABLE + FORCE and per-operation policies on the 4 tables with isolation tests, composite FKs and column grants; session context per transaction; no employer RFC, labels optional aliases (NFC, no control characters); logs carry ids only; flag matrix; input bounds (500 deductions, 20 constancias a year, under an advisory lock); every operation budgeted; `no-store`; `402` on writes, delete and restore open; no on-device persistence |

## Decisions

Full text in [MODULE-annual §12](../project-specs/MODULE-annual.md#12-decisions-coordinator-2026-10-15--all-decided)
and [ADR-0009](../adr/0009-annual-engine-placement.md).

| # | Decision | Effect |
|---|---|---|
| ADR-0009 | A sub-context of the tax engine module, one-way dependency | `module_annual` requires `module_tax_engine`; the monthly engine never imports the annual |
| OQ-A1 | Filed payments are their own table | Precedence filed → tax input `prior_provisional_payments` → app estimate |
| OQ-A2 | Constancias allowed for any PF | The screen suggests adding 605 when it is not held |
| OQ-A3 | No extrapolation of the year in progress | Records through `data_through` + `year_in_progress` |
| OQ-A4 | RESICO income out of the 15 % total income | Conservative; topic `deduction_cap_resico_income` |
| OQ-A5 | Donations without prior-year income | Excluded, section `incomplete` |
| OQ-A6 | 614 interest | Explanation only in v1 |
| OQ-A7 | Presets | `demo_full` only |
| OQ-A8 | Default year | Previous year from January to 30 April, else the current one |
| OQ-A9 + ADR-0009 | Read-only users | Creates, edits and PUTs `402`, never a trial slot; delete and restore stay open (CONTRACT.md 0.6.1) |

## Known limits and follow-ups

Tracked in the [PLAN AN follow-ups](../project-specs/PLAN-control-contable.md#an-follow-ups).
The ones a user or an operator can notice:

- **CPA sign-off pending (ANN-27, OB5).** AN-04 and the AN-07 PTU assertion are golden
  `skip`s; every ⚠️ topic (`uma_for_minimum_wage`, `ptu_rate`, `cu_rounding`,
  `salary_obligation_basis`, ...) raises the strong banner until then.
- **RMF 2027 is not published** (late December 2026). Deadlines, the RESICO PF relief
  (rule 3.13.7) and the salary relief (3.17.11, written for ejercicio 2025 only, not applied)
  may change for ejercicio 2026; `rmf_2027_pending` says so on every 2026 estimate. Handled by
  the [yearly update §8](../runbooks/tax-params-update.md#8-the-annual-return-module_annual-rmf-deadlines-and-rule-renewals).
- **iOS not walked** (OB6: Xcode pending), as in every module QA; Android only.
- **Archived taxpayer not walked on a device**: covered through the stack only
  (`TestAnnualQA_readOnlyAndArchivedOnEveryWrite`). Owed with the next device walk.
- **RESICO PF / PM, 625 and 603 not walked on a device**: pinned by the goldens and
  `annual_goldens_ui_test.dart`.
- **Filed-payments prefill vs the monthly screen** (QA §6, ANN-24 FU4): the annual chains the
  credited payments, so a month can prefill $0.00 while its monthly screen shows an ISR. Explained
  in the [user guide §3.3](../guides/annual-user-guide.md#33-pagos-provisionales-presentados-confirmados);
  an on-screen hint is a frontend follow-up.
- **Security follow-ups (ANN-25, non-blocking):** the soft-delete purge job (M9-08) must cover
  `personal_deductions`; a guard test that every `x-module: module_annual` operation has a
  budget; reject Unicode format characters (Cf: bidi overrides, zero-width) in labels.
- **Tech-lead follow-ups (ANN-24):** a completeness test for the deadline of each section
  kind; a `fiscal.RegimeGeneralPM` constant for the `"601"` literals.
- Annual inflation adjustment (601) and loss updating are out of scope
  (`inflation_adjustment_excluded`); IVA has no annual return (`no_annual_iva`).
