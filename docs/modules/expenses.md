# Module M4: Expenses

| | |
|---|---|
| Flag | `module_expenses` (OFF by default) |
| Spec | [MODULE-expenses.md](../project-specs/MODULE-expenses.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Expenses`, `x-module: module_expenses` (frozen by M4-01, 0.5.0) |
| Status | Contract frozen; implementation in progress (PLAN M4-02..M4-15). The full page lands with M4-15 |

## Contract highlights

- Same record mechanics as [incomes](incomes.md): centavos, idempotent client `id`, merge-patch,
  soft delete, restore without a trial slot, `Cache-Control: no-store`.
- Categories are taxpayer-scoped (`GET /v1/taxpayers/{taxpayerId}/expense-categories`): the
  hint depends on person type and regime options. The catalog is pending CPA review.
- `deductible` is the effective value; `deductible_source` says whether the user overrode it.
  `deductible: null` in a `PATCH` clears the override.
- `warnings[]` (generic `Warning`, with `field`): `cash_over_threshold`, `no_cfdi`,
  `not_isr_deductible_in_regime`, `category_not_deductible`, `duplicate_cfdi_uuid`,
  `iva_not_creditable_cash_payment` (0.6.0).
- `legend` (the mandatory "Esto es un aproximado..." text) travels on every expense, page and
  category list; the app renders the server field.

## Errors

Every error uses the single envelope `ErrorResponse`. Non-deductibility is never an error: it
is a warning plus `deductible: false`.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_expenses -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 422 | `business_rule_violation` | The taxpayer does not hold `regime` on `on`. |
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 429 | `rate_limited` | Rate limit exceeded. |
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories` | 500 | `internal_error` | Unexpected server error. |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses` | 429 | `rate_limited` | Rate limit exceeded. |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses` | 500 | `internal_error` | Unexpected server error. |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 400 | `validation_failed` | Input does not match the contract, an unknown category, a folio fiscal without CFDI, or a date out of range (`details[].code`: see the operation description). |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 402 | `entitlement_required` | Read-only access (trial over, no active plan), or the expense trial cap is spent: `details[0]` names the cap (`expenses_remaining` / `trial_cap_reached`). Never on an idempotent retry. |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 409 | `conflict` | The client `id` is reused: by the caller with other fields, another taxpayer or a deleted expense (`details`: `id` / `id_reused`), or it belongs to another user (no `details`). |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 422 | `business_rule_violation` | Well-formed but violates an expense rule: regime window, archived taxpayer, IVA rate or amount, withholdings (`details[].code`: see `createExpense`). |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 429 | `rate_limited` | Rate limit exceeded. |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | 500 | `internal_error` | Unexpected server error. |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 500 | `internal_error` | Unexpected server error. |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 400 | `validation_failed` | Input does not match the contract, or a category, CFDI or date rule (see `createExpense`). |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 422 | `business_rule_violation` | Well-formed but violates an expense rule: regime window, archived taxpayer, IVA rate or amount, withholdings (`details[].code`: see `createExpense`). |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 500 | `internal_error` | Unexpected server error. |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 422 | `business_rule_violation` | The taxpayer is archived; archive freezes its records (OQ-5). |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | 500 | `internal_error` | Unexpected server error. |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 400 | `validation_failed` | Input does not match the contract (query parameter, path or body). |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 404 | `not_found` | The taxpayer or record does not exist, belongs to another user, is deleted or purged, its id is malformed, or the module is disabled (indistinguishable on purpose). |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 422 | `business_rule_violation` | The regime no longer covers the record's dates (`regime_not_active`), or the taxpayer is archived (`taxpayer_archived`). |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 429 | `rate_limited` | Rate limit exceeded. |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_expenses -->
