# Module M6: Dashboard

| | |
|---|---|
| Flag | `module_dashboard` (OFF by default; implied OFF when `module_taxpayers` is OFF) |
| Spec | [MODULE-dashboard.md](../project-specs/MODULE-dashboard.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Dashboard`, `x-module: module_dashboard` (frozen by M6-01, 0.8.0) |
| Status | Contract frozen; implementation in progress (PLAN M6-02..M6-09). The full page lands with the module docs task |

## Contract highlights

- `getDashboard` (`GET /v1/taxpayers/{taxpayerId}/dashboard`): `period` optional (absent =
  current month, America/Mexico_City), `months` 6..12, default 12. The trend window is
  `[max(period - (months - 1), 2026-01), period]`, oldest first, one point per month (zeros
  for a month with no record), so the last point is always `period`.
- `month` holds income **collected** (`collected_on`) and expenses **paid** (`paid_on`) in
  `period`, before IVA. `trend[]` holds the same per month plus `tax` (IVA, ISR, withholdings
  to remit, total, `iva_outcome`, taxpayer-level `status`).
- `estimate` is the same `TaxpayerEstimate` as `getTaxpayerEstimate` for `period`; the hero
  links to the estimate screen with its `period`. `legend` is present if and only if
  `estimate` is.
- Flags degrade the response, never the status code: `module_tax_engine` OFF → `estimate`
  and every `trend[].tax` are `null`, no `legend`; `module_incomes` / `module_expenses` OFF →
  their `*_cents` are `null`. Nothing is computed for a module that is OFF.
- `access` is the caller's `AccessSummary` with the server-computed `trial_days_left`; the
  same frozen schema serves M7 and the pre-wizard access check (PR #58 follow-up, M7-01).
- Open to read-only users (no `402`). Own budget `dashboard_reads`, 600 per hour (burst
  600); every response carries `Cache-Control: no-store`.

## Errors

Every error uses the single envelope `ErrorResponse`. `details[].code` carries the rule
(`out_of_range` for a future `period`, `parameters_unavailable` before 2026-01). Amounts are
never echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_dashboard -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 400 | `validation_failed` | Input does not match the contract (malformed `period` or path id, `months` not an integer in 6..12, unknown query parameter), or a future `period` (`out_of_range`). |
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 404 | `not_found` | The taxpayer does not exist, belongs to another user, or its id is malformed, or `module_dashboard` (or `module_taxpayers`) is disabled (indistinguishable on purpose). |
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 422 | `business_rule_violation` | `period` is before 2026-01, the first month any record or estimate can exist (`parameters_unavailable`). A well-formed request with nothing to show, hence `422`, as `getTaxpayerEstimate`. |
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 429 | `rate_limited` | Rate limit exceeded. |
| `getDashboard` | `GET /v1/taxpayers/{taxpayerId}/dashboard` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_dashboard -->
