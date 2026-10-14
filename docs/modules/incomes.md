# Module M3: Incomes

| | |
|---|---|
| Flag | `module_incomes` (OFF by default) |
| Spec | [MODULE-incomes.md](../project-specs/MODULE-incomes.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Incomes`, `x-module: module_incomes` (frozen by M3-01, 0.5.0) |
| Status | Contract frozen; implementation in progress (PLAN M3-02..M3-14). The full page lands with M3-14 |

## Contract highlights

- Money is `int64` MXN centavos; `amount_cents` is the subtotal before IVA (`>= 1`).
- Create takes a client-generated `id` (idempotency): same `id` and fields → `200`, no second
  record, no trial slot; same `id` with other content → `409 conflict` (`id` / `id_reused`).
- `PATCH` is merge-patch on top-level fields: absent = unchanged, `null` clears a nullable field.
- Delete is soft; `restore` (Undo) **never consumes a trial slot**.
- `warnings[]` on write responses: `duplicate_cfdi_uuid`, `withholding_differs_from_rate`.
- `getIncomeDefaults` serves the allowed IVA rates and suggested withholding rates (no rate is
  hard-coded in the app).
- Record responses and the module's own errors carry `Cache-Control: no-store` (S-11).

## Errors

Every error uses the single envelope `ErrorResponse`. `error.code` is the stable value to
branch on; `details[].code` carries the domain rule (`regime_not_active`,
`iva_amount_mismatch`, `trial_cap_reached`, ...; see each operation in the contract). Amounts,
descriptions and CFDI UUIDs are never echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_incomes -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `listIncomes` | `GET /v1/taxpayers/{taxpayerId}/incomes` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `listIncomes` | `GET /v1/taxpayers/{taxpayerId}/incomes` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listIncomes` | `GET /v1/taxpayers/{taxpayerId}/incomes` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `listIncomes` | `GET /v1/taxpayers/{taxpayerId}/incomes` | 429 | `rate_limited` | Rate limit exceeded. |
| `listIncomes` | `GET /v1/taxpayers/{taxpayerId}/incomes` | 500 | `internal_error` | Unexpected server error. |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 400 | `validation_failed` | Input does not match the contract, or a date, activity or amount rule (`details[].code`: see the operation description). |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 402 | `entitlement_required` | Read-only access (trial over, no active plan), or the income trial cap is spent: `details[0]` names the cap (`incomes_remaining` / `trial_cap_reached`). Never on an idempotent retry. |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 409 | `conflict` | The client `id` is reused: by the caller with other fields, another taxpayer or a deleted income (`details`: `id` / `id_reused`), or it belongs to another user (no `details`). |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 422 | `business_rule_violation` | Well-formed but violates an income rule: regime window, archived taxpayer, IVA rate or amount, withholdings, 625 activity (`details[].code`: see `createIncome`). |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 429 | `rate_limited` | Rate limit exceeded. |
| `createIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes` | 500 | `internal_error` | Unexpected server error. |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 400 | `validation_failed` | Input does not match the contract: `on` out of range (`out_of_range`, 2022-01-01 … today), or `platform_activity` for a regime other than 625 (`not_for_regime`). |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 422 | `business_rule_violation` | The regime is not held on `on`, or the 625 activity is not among the taxpayer's. |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 429 | `rate_limited` | Rate limit exceeded. |
| `getIncomeDefaults` | `GET /v1/taxpayers/{taxpayerId}/incomes/defaults` | 500 | `internal_error` | Unexpected server error. |
| `getIncome` | `GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `getIncome` | `GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getIncome` | `GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `getIncome` | `GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getIncome` | `GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 500 | `internal_error` | Unexpected server error. |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 400 | `validation_failed` | Input does not match the contract, or a date, activity or amount rule (see `createIncome`). |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 422 | `business_rule_violation` | Well-formed but violates an income rule: regime window, archived taxpayer, IVA rate or amount, withholdings, 625 activity (`details[].code`: see `createIncome`). |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateIncome` | `PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 500 | `internal_error` | Unexpected server error. |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 422 | `business_rule_violation` | The taxpayer is archived; archive freezes its records (OQ-5). |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `deleteIncome` | `DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}` | 500 | `internal_error` | Unexpected server error. |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 422 | `business_rule_violation` | The regime no longer covers the record's dates (`regime_not_active`), or the taxpayer is archived (`taxpayer_archived`). |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 429 | `rate_limited` | Rate limit exceeded. |
| `restoreIncome` | `POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_incomes -->
