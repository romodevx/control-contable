# Module: Annual estimate

| | |
|---|---|
| Flag | `module_annual` (OFF by default; requires `module_tax_engine`, and through it `module_taxpayers`; registered by migration `20261019100000_register_module_annual` (ANN-22); in `demo_full` only) |
| Spec | [MODULE-annual.md](../project-specs/MODULE-annual.md) |
| Design | [ADR-0009](../adr/0009-annual-engine-placement.md) (a sub-context of the tax engine module), [ADR-0002](../adr/0002-tax-engine.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Annual`, `x-module: module_annual` (frozen by ANN-02, 0.11.0) |
| Status | Contract frozen; backend built (ANN-03..ANN-15: inputs, `EstimateAnnual`, handlers); flag registered, in `demo_full` (ANN-22); QA, gates and docs pending (ANN-23..ANN-27). The full page lands with the module docs task (ANN-26) |

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
