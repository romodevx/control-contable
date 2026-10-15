# API contract rules

`docs/api/openapi.yaml` is the single source of truth. It is linted on every PR that touches
`docs/api/**`, `redocly.yaml` or `tools/contract/**` (`.github/workflows/contract.yml`), and locally with:

```sh
make contract-lint        # lint the contract
make contract-lint-test   # prove each rule rejects what it should (fixtures)
```

Tooling is pinned: Redocly CLI `2.62.1` in `tools/contract/package.json` + `package-lock.json`,
installed with `npm ci`. Ruleset: [`redocly.yaml`](../../redocly.yaml); custom rules:
[`tools/contract/plugin.js`](../../tools/contract/plugin.js).

Transitive security pins (`overrides` in `tools/contract/package.json`, exact versions at
least two weeks old). This tooling is dev-only and never ships in an artifact:

| Override | Why |
|---|---|
| `postman-collection` → `lodash` `4.18.1` | GHSA-r5fr-rjxr-66jc (high), GHSA-f23m-r3pf-42rh, GHSA-xxjr-mmjv-4gpg; postman-collection pins `4.17.21` exactly |
| `postman-collection` → `uuid` `11.1.1` | GHSA-w5hq-g745-h8pq; postman-collection pins `8.3.2` exactly (it only calls `v4()`, same API) |
| `js-yaml` → `argparse` `2.0.1` | Drops `sprintf-js` (GHSA-hp3w-g68c-fv3c, no patched release). argparse is required only by js-yaml 3's own CLI binary, which nothing here runs; `json-schema-ref-parser@6` needs js-yaml 3's `safeLoad`, so js-yaml itself cannot move to 4 |

Accepted risk: `postman-collection` keeps its exact pin `@faker-js/faker@5.5.3`
(GHSA-qxc2-j82w-r537, patched only in `10.5.0`). Faker 10 breaks postman-collection at load
time (`faker.address` was removed), and Prism loads it eagerly, so overriding it breaks the
mock. The vulnerable path is `faker.helpers.fake` with an attacker-controlled template;
postman-collection never calls it and Prism never reads Postman collections. Prism's own
faker resolves to the patched `10.6.0`. Drop this note when postman-collection moves to a
faker `>= 10.5.0`.

## Enforced rules

| Rule | What it enforces |
|---|---|
| `recommended-strict` | Redocly's recommended set, every rule as an error |
| `operation-operationId` (+ `-unique`, `-url-safe`) | Every operation has a unique, URL-safe `operationId` (codegen and the operationID→flag map depend on it) |
| `rule/operation-x-module` | Every operation declares `x-module`: `platform` or a `module_<name>` flag key |
| `contract/error-response-has-body` | Every `4xx`/`5xx`/`default` response (inline or `$ref`) has at least one media type |
| `rule/error-response-envelope` | Every media type of that body has a schema, and it is `$ref: '#/components/schemas/ErrorResponse'` (single error envelope) |
| `contract/body-example` | Every request/response body has an example, on the media type or on the schema it references |
| `no-invalid-media-type-examples`, `no-invalid-schema-examples` | Every example validates against its schema (no extra properties) |
| `security-defined` | Every operation is covered by a declared security scheme (global or its own) |
| `contract/public-operation-explicit` | `security: []` (public) only together with `x-public: true`, and vice versa |
| `contract/frozen-unchanged` | Any node marked `x-frozen: sha256:<hash>`, anywhere in the document, still hashes to that value |

## Reviewed conventions (not lint-enforced yet)

| Convention | Why | Backstop |
|---|---|---|
| No inline object schemas inside `oneOf`/`anyOf`: name them as components and `$ref` them | swagger_parser names an inline object's enums after the bare property, so `SubscriptionStatus.plan.period` silently overwrote the `Period` component in the Dart client ([ADR-0003](../adr/0003-dart-client-generation.md)) | Mobile kind guard in `make mobile-client` (PLAN P0-23, run in CI by P0-25) |

## Frozen contract

A frozen node is a published promise: the backend, the generated Dart client and the
contract tests already depend on it. It is marked in the contract itself:

```yaml
x-frozen: sha256:<sha256 of the node's canonical JSON, keys sorted, x-frozen excluded>
```

The hash covers the node as written (a `$ref` is hashed as the reference string, not the
target), so every schema a frozen operation exposes is frozen on its own.

### Frozen today

| Node | Frozen by | Why |
|---|---|---|
| `GET /healthz` (`getHealth`) | P0-05 | Liveness/readiness probe used by deploy health checks |
| `GET /v1/flags` (`listEnabledFlags`) | P0-05 | Feeds the mobile FlagsProvider |
| `components.schemas.ErrorResponse` | P0-05 | The single error envelope every endpoint returns |
| `components.schemas.FieldError` | P0-05 | Item of `ErrorResponse.error.details` |
| `components.schemas.HealthStatus` | P0-05 | `/healthz` body |
| `components.schemas.EnabledFlags`, `ModuleKey` | P0-05 | `/v1/flags` body |
| M1 `module_auth` operations: `register`, `verifyEmail`, `resendVerification`, `login`, `refreshSession`, `logout`, `getMe`, `updateMe`, `deleteMe`, `listLegalDocuments` | M1-01 (re-frozen by D8, 0.2.0; `deleteMe` and `listLegalDocuments` by D10, 0.3.0) | [MODULE-auth](../project-specs/MODULE-auth.md) (D1-D10); the backend (M1-05..M1-09) and the Flutter auth screens (M1-11..M1-13) build on it. D8 moved `password`/`name` from `RegisterRequest` to `VerifyEmailRequest`; D10 made a wrong `deleteMe` password `403 invalid_password` (its `401` is now the shared `Unauthorized`) and gave the legal documents `Cache-Control: public, max-age=300` (pre-release breaking changes, recorded in MODULE-auth §9) |
| `components.schemas.RegisterRequest`, `DocumentAcceptance`, `VerificationSent`, `VerifyEmailRequest`, `ResendVerificationRequest`, `LoginRequest`, `RefreshRequest`, `TokenPair`, `AuthSession`, `User`, `UserUpdate`, `DeleteAccountRequest`, `LegalDocumentKind`, `LegalDocument`, `LegalDocumentList` | M1-01 | Bodies of the M1 operations |
| `components.headers.NoStore` | M1-01 (re-frozen by 0.7.0) | `Cache-Control: no-store`, since 0.7.0 on **every** response, success or error, except the public `200` of `listLegalDocuments` (`public, max-age=300`). Written once by `middleware.NoStore`, outside routing, so a `404` written before the module is reached, a `401`, a `429` and a recovered `500` carry it too |
| `components.responses.Unauthorized`, `components.headers.WWWAuthenticate` | M1-07 (security gate, pre-release addition) | The 401 of every secured operation, with the fixed `WWW-Authenticate: Bearer realm="control-contable"` challenge RFC 9110 §15.5.2 requires; the bearer authenticator writes it |
| `components.responses.TooManyRequests` | M1-10 (pre-release change, 0.3.1; noted in the PR #9 tech-lead review) | The 429 of every operation. `Retry-After` (whole seconds, at least 1) is **required**: both the edge limiter and the per-operation and per-account limits send it, and the harness now fails a 429 without it |
| M2 `module_taxpayers` operations: `listRegimes`, `listTaxpayers`, `createTaxpayer`, `getTaxpayer`, `updateTaxpayer`, `deleteTaxpayer`, `archiveTaxpayer`, `unarchiveTaxpayer`, `addTaxpayerRegime`, `updateTaxpayerRegime`, `archiveTaxpayerRegime` | M2-01 (0.4.0) | [MODULE-taxpayers](../project-specs/MODULE-taxpayers.md) (OD-1..OD-7, OQ-1..OQ-9); the backend (M2-05..M2-07) and the Flutter onboarding/switcher (M2-08..M2-10) build on it. `unarchiveTaxpayer` and the RFC correction on `updateTaxpayer` are new (OQ-1, OQ-2); RFC rules travel as `details[].code` on `rfc`, never echoing the RFC; a check-digit mismatch is accepted with a `warnings` entry (D-RFC) |
| `components.schemas.PersonType`, `Treatment`, `RegimeCode`, `RegimeOptionDef`, `RegimeIncompatibility`, `Regime`, `RegimeList`, `RegimeOptions`, `TaxpayerRegime`, `TaxpayerRegimeCreate`, `TaxpayerRegimeUpdate`, `RfcInput`, `Warning`, `Taxpayer`, `TaxpayerCreate`, `TaxpayerUpdate`, `TaxpayerList` | M2-01 (`Regime` re-frozen by M2-07) | Bodies of the M2 operations. Request bodies are closed (`additionalProperties: false`, S-11); `RfcInput` bounds only the length so the domain reports every RFC rule with its own code. M2-07 added the optional response field `Regime.review_pending_with` (non-breaking, PR #56 follow-up): the pairs under CPA review, so the client warns before submitting |
| M3 `module_incomes` operations: `listIncomes`, `createIncome`, `getIncomeDefaults`, `getIncome`, `updateIncome`, `deleteIncome`, `restoreIncome` | M3-01 (0.5.0) | [MODULE-incomes](../project-specs/MODULE-incomes.md) §8 (OQ-1..OQ-10, all decided); the backend (M3-02..M3-07) and the Flutter list/form (M3-08, M3-09) build on it. `getIncomeDefaults` is new (OQ-7); delete gained `422`, restore `402`/`422`; restore never consumes a trial slot (OQ-10) |
| M4 `module_expenses` operations: `listExpenseCategories`, `listExpenses`, `createExpense`, `getExpense`, `updateExpense`, `deleteExpense`, `restoreExpense` | M4-01 (0.5.0) | [MODULE-expenses](../project-specs/MODULE-expenses.md) §8 (OQ-E1..OQ-E6). `listExpenseCategories` moved from the global `GET /v1/expense-categories` to `GET /v1/taxpayers/{taxpayerId}/expense-categories` (OQ-E1; it was never frozen); `ExpenseWarning` was replaced by the generic `Warning` (OQ-E5) |
| `components.schemas.IvaRate`, `Legend`, `PositiveCents`, `BusinessDate`, `CounterpartyType`, `PlatformActivity`, `CfdiUuid`, `RecordDescription`, `PeriodTotals`, `IncomeCreate`, `IncomeUpdate`, `Income`, `IncomePage`, `IncomeDefaults`, `PaymentMethod`, `DeductibleSource`, `ExpenseCreate`, `ExpenseUpdate`, `Expense`, `ExpensePage`, `Deductibility`, `ExpenseCategory`, `ExpenseCategoryList` | M3-01 / M4-01 | Bodies of the M3/M4 operations. Request bodies are closed (`additionalProperties: false`, S-3) and flat (no `allOf`, which a closed schema cannot compose). `PATCH` bodies are merge-patch on top-level fields: absent = unchanged, `null` clears a nullable field (and `deductible: null` clears the expense override), see ADR-0003 "Null on the wire" |
| `components.responses.RecordBadRequest`, `RecordReadOnly`, `RecordNotFound`, `RecordTaxpayerArchived`, `RecordRestoreRefused`, `IncomeRuleViolation`, `ExpenseRuleViolation`, `RecordInternalError` | M3-01 / M4-01 (re-frozen by 0.7.0) | The M3/M4 error responses, with `Cache-Control: no-store` (S-11, P0-30 D-6). Until 0.7.0 they referenced the optional `NoStoreOnError` header; 0.7.0 removed it and they reference the required `NoStore` like every response. The unused shared `Conflict` response was removed (every `409` is now declared inline) |
| M5 `module_tax_engine` operations: `getTaxpayerEstimate`, `getEstimateBreakdown`, `getTaxInputs`, `putTaxInputs` | M5-01 (0.6.0) | [MODULE-tax-engine](../project-specs/MODULE-tax-engine.md) §9 (OQ-T1..OQ-T10, all decided); the engine (M5-02..M5-22), the API (M5-23) and the Flutter estimate, breakdown and tax-inputs screens (M5-25..M5-27) build on it. `period` is optional on the estimate (current month); before 2026 → `422 parameters_unavailable`, a future month → `400 out_of_range`; every M5 response and error carries `Cache-Control: no-store` (S-4). Pre-release breaking changes from the 0.1.0 draft (never frozen, nothing built on them): `RegimeEstimate.missing_inputs` became `MissingInput` objects, `status` gained `partial`, `TaxResult.payable_rounded_cents` is `>= 0`, and the new fields are required (amended [ADR-0002](../adr/0002-tax-engine.md)) |
| `components.schemas.ParameterSetRef`, `TaxResult`, `MissingInput`, `RegimeEstimate`, `IvaComponents`, `EstimateTotals`, `RecordsFingerprint`, `TaxpayerEstimate`, `Citation`, `BreakdownLine`, `Breakdown`, `TaxInputDef`, `TaxInputValue`, `TaxInputs`, `TaxInputsUpdate`, `Disclaimer`, `DisclaimerSeverity` | M5-01 | Bodies of the M5 operations. `TaxInputsUpdate` and `TaxInputValue` are closed (`additionalProperties: false`, S-7). `TaxInputDef.kind` gained `money_monthly` and `boolean` (OQ-T1); the catalog CHECK `regime_required_inputs_valid` and the 612/626 PM/601 rows move to them in the forward migration of M5-20/M5-21, until then the catalog only serves `money`/`rate`. `DisclaimerCode` and `CpaTopic` are **append-only**, not frozen (like `ErrorCode`): clients show a disclaimer's `text` whatever its code |
| `components.responses.TaxEngineBadRequest`, `TaxEngineNotFound`, `TaxEngineParametersUnavailable` | M5-01 | The shared M5 errors, with `Cache-Control: no-store` (`NoStore` since 0.7.0). `402` and `500` reuse `RecordReadOnly` and `RecordInternalError`; each `422` with several detail codes is inline on its operation. The unused shared `UnprocessableEntity` response was removed |
| `createExpense`, `components.schemas.Expense`, `Warning` (re-frozen); `WarningCode` (new) | M5-01 (0.6.0, non-breaking on the wire) | Their descriptions list the expense warning `iva_not_creditable_cash_payment` on `payment_method` (PR #67, MODULE-expenses §2.5, §2.8). `Warning.code` moved from a pattern string to `$ref: WarningCode`, an **append-only** enum (not frozen, like `ErrorCode`) that is the single list of every warning code with the field it names (M2, M3, M4); the operation descriptions only repeat the subset each one can send. Same strings on the wire; the generated clients get a typed code with an unknown fallback (show `message`) |
| `restoreIncome`, `restoreExpense` (re-frozen) | M3-08 review (0.6.1, pre-release breaking change, coordinator decision) | A read-only user may restore (Undo) their own deleted records: restore runs no access check and no longer declares `402` (MODULE-incomes §2.7, §2.10; MODULE-expenses US-E3). Undo gives back the user's own record and creates nothing, so read-only never makes a delete one-way. `404` (another user's or a purged record) and `422` (`regime_not_active`, `taxpayer_archived`) are unchanged. Nothing shipped depended on the `402`: the use cases already allowed it (PR #82) and the HTTP handlers are not built yet |
| `WarningCode` (append) | M3-04 (0.6.0, append-only, no freeze touched) | Code `provisional_parameters` (shared with M4-05, same string as the disclaimer) also on `collected_on` / `invoiced_on`: an income whose dates read a provisional parameter set (PR #77 follow-up: `RateCatalog` reported `Provisional`, but records dropped it). The frozen `createIncome` / `updateIncome` / `Income` descriptions do not list it yet; the architect folds it into the next re-freeze. A date before the first parameter set answers `400` `out_of_range` on that date (M3-06, as expenses and the frozen date codes; M3-04 had sent `422` `parameters_unavailable`) |
| `getEstimateBreakdown`, `components.schemas.Breakdown` (re-frozen before any build) | M5-01 (0.6.0, additive) | `Breakdown.disclaimers` is pinned down: the estimate's disclaimers that bear on the lines, same order, with `cpa_review_pending` (strong) whenever the regime or a line (including the consolidated IVA line, `cross_regime_iva_netting`) needs CPA review, and `iva_mixed_activity` when the consolidated IVA took any regime's creditable IVA in full (MODULE-tax-engine §2.3, §2.9: the warnings are visible wherever the figures are). New named example `crossRegimeIva612` (TX-NET-01) |
| `DisclaimerCode` (append-only, not frozen): `resico_cap_exceeded` | M5-13/M5-16 (0.6.0, non-breaking) | A 626 regime above its annual income cap (PF 3.5M, LISR 113-E/113-I; PM 35M, LISR 206/214) leaves RESICO; the estimate carries this warning instead of a breakdown line (lines are money only). Clients show the `text` of an unknown code. MODULE-tax-engine §2.3, §9 |
| Every frozen operation and shared response that lacked it (`getHealth`, `listEnabledFlags`, the M1 auth operations, the M2 writes and `listRegimes`, the M3/M4/M5 operations with inline responses; `Unauthorized`, `TooManyRequests`, the `Record*` and `TaxEngine*` responses), `NoStore`, `BusinessDate` (re-frozen); `NoStoreOnError` (removed) | 0.7.0 (pre-release, non-breaking on the wire for clients) | **`Cache-Control: no-store` on every response** (M3/M4 security gate G-1, M1-17 G-3, P0-30 D-6, S-11): each response declares the required `NoStore` header and `middleware.NoStore` sends it, also on errors, which no handler can make cacheable. **Record date floor** (PR #97 FU-2): `BusinessDate`, `createIncome`, `getIncomeDefaults` and `createExpense` (and `OnDate`, not frozen) say what the server always did: a record date before 2026-01-01 answers `400 validation_failed` / `out_of_range`, because 2026 is the first fiscal year with a tax parameter set (2022 is only the database CHECK backstop). The estimates keep `422 parameters_unavailable` for a month before 2026: a well-formed period the engine has no parameters for, not invalid input. `createIncome` also lists the `provisional_parameters` warning (M3-04 note above) |
| `createTaxpayer`, `addTaxpayerRegime`, `components.schemas.TaxpayerRegimeCreate` (re-frozen) | M5-29 QA D2 (0.7.1, pre-release narrowing, wire shape unchanged) | `start_period` is allowed from `2026-01` (the first fiscal year with tax parameters) instead of `2022-01`: a month before it can hold no record (no rate covers it) nor an estimate. Earlier starts answer `400 out_of_range` on `start_period` as before. The app now sends `start_period` from a month picker in onboarding and *Agregar régimen*, so a taxpayer who starts mid-year can record and estimate January onwards. Stored rows keep the database CHECK (`2022-01`) |
| M6 `module_dashboard` operation: `getDashboard` | M6-01 (0.8.0) | [MODULE-dashboard](../project-specs/MODULE-dashboard.md) §9 (OQ-D1..OQ-D7, all decided); the use case (M6-02), handler and contract tests (M6-03) and the Flutter Home (M6-04) build on it. `period` is now optional (current month, America/Mexico_City; was the required `PeriodQuery`), before 2026-01 → `422 parameters_unavailable`, a future month → `400 out_of_range`; `months` 6..12, default 12, window clamped at 2026-01. No `402` (a read); `429` with `Retry-After` (`dashboard_reads`, 600/h). Flags degrade the body, never the status: engine OFF → `estimate` and `trend[].tax` `null`, no `legend`; incomes/expenses OFF → their `*_cents` `null`. Pre-release breaking changes from the 0.1.0 draft (never frozen, nothing built on it): top-level `income_cents`/`expense_cents` moved into `month`, `TrendPoint.tax_payable_rounded_cents` became `tax: TrendTax \| null`, `estimate` nullable, `legend` optional |
| `components.schemas.Dashboard`, `DashboardMonth`, `TrendPoint`, `TrendTax`, `TaxpayerEstimateStatus`, `AccessSummary`, `AccessLevel` | M6-01 | Bodies of `getDashboard`. `AccessSummary` gained `trial_days_left` (server clock, OQ-D6) and every field is now required (nullable), with its per-level meaning in the schema; its `access` enum became the named `AccessLevel`. It is frozen here as **the** access shape for every screen: M7's `SubscriptionStatus` already references it, and the PR #58 follow-up (the app needs access state before a create wizard, with `module_subscriptions` OFF too) is answered at M7-01 by exposing this same schema, not a new one (MODULES.md M7) |
| M8 `module_history` operations: `getHistory`, `markHistoryViewed` | M8-01 (0.9.0) | [MODULE-history](../project-specs/MODULE-history.md) §9 (OQ-H1..OQ-H7, all decided); the migration (M8-02), use case (M8-03), handlers and contract tests (M8-04) and the Flutter list (M8-05) build on it. `fiscal_year` required: before 2026 → `422 parameters_unavailable`, after the current year → `400 out_of_range`; months newest first (January .. December, or .. the current month). The changed marker is driven only by `figures_fingerprint` (OQ-H4), `change_reason` `engine` \| `records`; never-viewed and current months are never marked. `markHistoryViewed` takes a required closed body (what the app rendered) and stays an idempotent `204`. `module_history` requires `module_tax_engine` (flag OFF or implied OFF → `404`). No `402`; `429` with `Retry-After` (`history_reads` 600/h, `history_views` 1,200/h). Pre-release breaking changes from the 0.1.0 draft (never frozen, nothing built on it): `HistoryMonth.status` is `TaxpayerEstimateStatus` (`no_records` removed; `has_records` instead), `iva`/`isr` `TaxResult` replaced by `tax: HistoryTax \| null`, `last_viewed_at` renamed `viewed_at` |
| `components.schemas.HistoryYear`, `HistoryYearTotals`, `HistoryMonth`, `HistoryTax`, `HistoryChangeReason`, `HistoryViewedRequest`, `HistoryViewedParameterSet`, `FiguresFingerprint`, `EngineVersion` | M8-01 | Bodies of the M8 operations. `HistoryViewedRequest` and `HistoryViewedParameterSet` are closed (`additionalProperties: false`); fingerprints are 64 lower-case hex, only ever compared |
| `components.schemas.TaxpayerEstimate`, `RecordsFingerprint` (re-frozen) | M8-01 (0.9.0, additive, non-breaking) | `TaxpayerEstimate` gained the **optional** `figures_fingerprint` (hash of the displayed figures, computed in the M5 domain next to `RecordsFingerprint`; sent once M8-03 ships, which **will** make it required: tracked in PLAN M8-03, together with typing `TaxpayerEstimate.engine_version` as `EngineVersion`). `RecordsFingerprint`'s description now says M8 uses it only to explain the change reason. No field removed or tightened |
| Withholdings: `previewIncomeWithholdings` (`module_incomes`), `previewExpenseWithholdings` (`module_expenses`) (new); `createIncome`, `getIncomeDefaults`, `updateIncome`, `restoreIncome`, `createExpense`, `updateExpense`, `restoreExpense`, `IncomeRuleViolation`, `ExpenseRuleViolation` (re-frozen) | WH-02 (0.10.0, **breaking, in two steps**, [ADR-0008](../adr/0008-withholdings-are-derived.md)) | [MODULE-withholdings](../project-specs/MODULE-withholdings.md) §7: withholdings are **derived** by the server, stored with `table_version`, never entered. 0.10.0 is the *expand* step, additive on the wire: new optional facts (`counterparty_kind`, `operation_kind`, `residential_rent`, `via_platform`, supplier fields), the read-only `withholding` object, `last_used` defaults, and two previews (one per module, so each one hides with its own flag; shared `withholding_previews` budget, 1200/h). The members to be removed are `deprecated: true` with an `x-deprecated-reason` that names the task removing them. The *contract* steps, WH-06 (incomes) and WH-07 (expenses), remove them and make the new fields required, as pre-release patches (0.10.x) authorized by ADR-0008. New detail codes `supplier_not_applicable`, `residential_rent_exempt`; `withholding_not_applicable` / `withholding_exceeds_base` retired by WH-06/WH-07 |
| `components.schemas.ExpenseCategory`, `ExpenseCategoryList` (re-frozen) | PR #137 tech-lead review (0.11.1, additive, non-breaking) | New **optional** `ExpenseCategory.default_operation_kind`: the category's `operation_kind` from the `category_regime_rules` seed (MODULE-withholdings §2.2), the one E3 starts at and the one the server applies when `operation_kind` is omitted. The mobile form had mirrored the seed in Dart; WH-14 forbids a second copy of the table, so the server exposes it. Sent once WH-03 seeds it and WH-07 serves it, which **will** make it required (PLAN WH-07). Until then a client leaves E3 unselected and omits `operation_kind`, and the server applies the same default. No field removed or tightened |
| `components.schemas.CounterpartyKind`, `OperationKind`, `SupplierKind`, `SupplierRegime`, `WithholdingRuleId`, `WithholdingBasis`, `Withholder`, `WithholdingExplanation`, `Withholding`, `WithholdingPreview`, `IncomeLastUsed` (new); `CounterpartyType` (deprecated), `IncomeCreate`, `IncomeUpdate`, `Income`, `IncomeDefaults`, `ExpenseCreate`, `ExpenseUpdate`, `Expense` (re-frozen) | WH-02 (0.10.0) | Bodies of the withholdings change. `Withholding` is read-only everywhere; `isr_withheld_cents` / `iva_withheld_cents` stay in responses permanently, as its mirrors. `WarningCode` (append-only) gains `withholding_cpa_unverified`, `foreign_counterparty_not_supported`, `supplier_regime_unknown` and `deduction_requires_remittance`; `withholding_differs_from_rate` is retired (no longer sent after WH-06, kept in the enum). `CpaTopic` (append-only) gains `withholding_rule_unverified` and `withheld_iva_credit_timing` (WH-09). Codegen note: the `Withholding` schema made oapi-codegen prefix the `BreakdownLineTax` constants (`gen.BreakdownLineTaxWithholding`) |
| `module_annual` operations: `getAnnualEstimate`, `getAnnualBreakdown`, `getAnnualInputs`, `putAnnualInputs`, `listSalaryStatements`, `createSalaryStatement`, `updateSalaryStatement`, `deleteSalaryStatement`, `listPersonalDeductions`, `createPersonalDeduction`, `updatePersonalDeduction`, `deletePersonalDeduction`, `restorePersonalDeduction`, `getFiledPayments`, `putFiledPayments` (new) | ANN-02 (0.11.0, additive, [ADR-0009](../adr/0009-annual-engine-placement.md)) | [MODULE-annual](../project-specs/MODULE-annual.md) §10 (OQ-A1..OQ-A9, all decided); the input use cases (ANN-13), `EstimateAnnual` (ANN-14), the handlers and contract tests (ANN-15) and the Flutter screens (ANN-17..ANN-21) build on it. `module_annual` requires `module_tax_engine` and `module_taxpayers` (OFF or implied OFF → `404`). Future `fiscalYear` → `400 out_of_range`; before 2026 → `422 parameters_unavailable`; a deduction's `paid_on` follows the `BusinessDate` rule (`400 out_of_range`). Creates, edits and PUTs answer `402` when read-only and never use a trial slot (OQ-A9); delete and restore are allowed to a read-only user, as records (0.6.1). Detail codes `person_type_not_allowed`, `payments_not_applicable`, `section_not_present`, `limit_reached` (500 deductions / 20 constancias per year) are new. Budgets: `estimate_reads` (estimate, breakdown, filed payments), `annual_reads` 600/h, `annual_writes` 300/h (ANN-15). Not built yet: `notBuilt` stubs, pending under ANN-15. `annual_depreciation` needs no contract change: tax-input keys are catalog data (ANN-03), served by the frozen `getTaxInputs` |
| `components.schemas.AnnualLegend`, `AnnualYearStatus`, `AnnualSectionKind`, `AnnualComputedSectionKind`, `AnnualSectionStatus`, `AnnualObligation`, `AnnualBalanceKind`, `AnnualBalance`, `PaymentSourceCounts`, `AnnualCredits`, `AnnualTax`, `PmAnnualExtras`, `ResicoPfCheck`, `AnnualNote`, `AnnualInputTarget`, `AnnualMissingInput`, `AnnualSection`, `AnnualEstimate`, `AnnualLineSource`, `AnnualBreakdownLine`, `AnnualBreakdown`, `AnnualRegimeTaxInput`, `AnnualInputs`, `AnnualInputsUpdate`, `EmployerLabel`, `SalaryStatement`, `SalaryStatementCreate`, `SalaryStatementUpdate`, `SalaryStatementTotals`, `SalaryStatementList`, `DeductionCategory`, `DeductionPaymentMethod`, `DeductionBeneficiary`, `DoneeType`, `TuitionLevel`, `StudentLabel`, `DeductionExclusion`, `PersonalDeduction`, `PersonalDeductionCreate`, `PersonalDeductionUpdate`, `DeductionCategoryTotal`, `PersonalDeductionList`, `FiledPaymentSource`, `FiledPaymentMonth`, `FiledPayments`, `FiledPaymentEntry`, `FiledPaymentsUpdate`; `components.responses.AnnualBadRequest`, `AnnualNotFound`, `AnnualParametersUnavailable`, `AnnualRuleViolation`, `AnnualIdConflict` (new) | ANN-02 (0.11.0) | Bodies and errors of the annual operations. Request bodies are closed (`additionalProperties: false`); `PATCH` bodies are merge-patch (`null` clears a nullable field). `AnnualLegend` is its own constant, not the monthly `Legend`. `AnnualBreakdownLine` mirrors the frozen `BreakdownLine` without `tax` and with `regime_code` and the wider `AnnualLineSource` (`filed_payment`, `estimated_by_app`, `salary_statements`, ...), so M5's frozen line is untouched (ADR-0009). Append-only, not frozen: `AnnualNoteCode` (new); `DisclaimerCode` gains `annual_estimate`, `year_in_progress`, `payments_estimated`, `deductions_as_entered`, `resico_pf_definitive`, `informative_income`, `rmf_2027_pending`, `inflation_adjustment_excluded`, `no_annual_iva`; `CpaTopic` gains the eight MODULE-annual §2.10 topics (the monthly card reads them as the generic topic); `WarningCode` gains `cash_not_deductible` and `salary_regime_not_held` |

**`FieldError.field` naming (one convention).** A dotted path relative to the request body
(`password`, `accepted_documents.1.kind`), `body` for the body root, and the bare parameter
name for path/query/header errors (`taxpayerId`, `period`). It is never a `/`-prefixed JSON
Pointer, in examples, specs or handlers. `FieldError.code` is the failed JSON Schema keyword
(`required`, `minLength`, `enum`, `additionalProperties`, …) for request validation, or a
snake_case domain code from the handler (`password_policy`). The router produces both from
either kin-openapi engine (`backend/adapter/http/router/fields.go`).

`ErrorCode` is **append-only**, not frozen: modules add codes as they land. Removing or
renaming a code is a breaking change (rules below). Clients must treat unknown codes as
`internal_error`-like generic failures.

Module contracts are frozen the same way at each module's "contract freeze" task in the PLAN
(mark its operations and body schemas with `x-frozen`).

### Changing a frozen node

1. **Breaking** (removing/renaming a field, path, enum value; tightening validation; changing
   a type or status code): needs an ADR and a version bump (`info.version`; a new path
   version when shipped clients would break). Prefer adding a new field/operation instead.
2. **Non-breaking** (new optional response field, wording of a description, new example):
   allowed; the PR says why it is non-breaking.
3. In both cases, the lint error prints the new hash. Set `x-frozen` to it in the same PR;
   the tech-lead gate reviews every `x-frozen` diff.

To freeze a new node, add `x-frozen: sha256:pending`, run `make contract-lint` and copy the
`actual` hash from the error.

## Contract tests (P0-18)

Real responses from the real router are validated against this file with
[libopenapi-validator](https://github.com/pb33f/libopenapi-validator) `v0.16.1` (pinned per
ADR-0001), independently of the kin-openapi copy the server is generated from. The harness is
`backend/adapter/http/contracttest`:

```go
c := contracttest.Load(t)                                         // parses this file once per test binary
rec := c.ServeOp(t, contracttest.Op("createIncome"), handler, req) // serve + validate; counts as coverage
rec = c.Serve(t, handler, req)                                     // serve + validate; no coverage claim (e.g. flag-OFF 404)
```

- Status, declared response headers and body are checked in **strict mode**: an undeclared
  property fails, and so does any response header the matched response does not declare,
  except the transport headers listed (with the reason for each) in `transportHeaders`
  (`backend/adapter/http/contracttest/headers.go`): `Content-Type`, `Content-Length`, `Date`,
  `X-Request-Id`, `X-Content-Type-Options`. The header check is the harness's own:
  libopenapi-validator's strict mode skips responses that declare no headers and ignores
  `Set-Cookie`, `Cache-Control`, `ETag`, `Vary`, ... by default. A response that needs one of
  those declares it in this file. Formats (`uuid`, `date`, ...) are asserted.
- Strict mode also reports every key of an open map (`additionalProperties` with a schema) as
  undeclared. The keys of the contract's open maps are data, so they are exempt by path in
  `openMapPaths` (`backend/adapter/http/contracttest/contracttest.go`); their values are still
  validated. Today: `RegimeOptions` (`TaxpayerRegime.options`, keyed by the catalog's option keys).
- A failure prints the operation, the declared statuses, each violation with its JSON path and
  the response received (status, headers, pretty-printed body).
- `ServeOp` also fails when the request does not route to the named operation.

**Coverage of the contract.** `TestContractCoverage` scans the backend's Go files (every build
tag) for `contracttest.Op("<operationId>")` literals. Every operation must be either covered or
listed in `pendingOperations` (`backend/adapter/http/contracttest/coverage_test.go`) with the
PLAN task that owns its tests; never both. The list only shrinks: a module's contract-test task
covers its operations and removes them from it in the same PR. A new operation is added to the
list (or covered) in the PR that adds it to the contract.

Run them with `make test-integration` (they need Docker), or only the contract suites with
`make test-contract`. CI runs them in the backend workflow's integration job.
