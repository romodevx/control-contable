# Module M8: History

| | |
|---|---|
| Flag | `module_history` (OFF by default; implied OFF when `module_tax_engine` is OFF, and through it `module_taxpayers`) |
| Spec | [MODULE-history.md](../project-specs/MODULE-history.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `History`, `x-module: module_history` (frozen by M8-01, 0.9.0) |
| Status | Contract frozen; implementation in progress (PLAN M8-02..M8-09). The full page lands with the module docs task (M8-10) |

## Contract highlights

- `getHistory` (`GET /v1/taxpayers/{taxpayerId}/history?fiscal_year=YYYY`): one row per
  month, **newest first**: January .. December for a past year, January .. the current
  month (America/Mexico_City) for the current one. A month without a regime is
  `status: none`, `tax: null`, never a hole. Each month's `tax` is the same `totals` as
  `getTaxpayerEstimate` for that `period`, to the centavo; `year_totals` is the "Suma de
  estimados mensuales", never an annual return.
- **Changed since viewed** is driven only by `figures_fingerprint` (OQ-H4): a viewed month
  whose displayed figures moved is `changed_since_viewed: true` with `change_reason`
  `engine` (engine version or parameter set differs from the one viewed; wins when both
  apply) or `records`. Never-viewed months (OQ-H1) and the current month (OQ-H3) are never
  marked.
- `markHistoryViewed` (`POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed`): body
  `HistoryViewedRequest` (`records_fingerprint`, `figures_fingerprint`, `engine_version`,
  `parameter_set`) copied from the `TaxpayerEstimate` the app rendered, from any entry point
  (OQ-H7). Idempotent upsert, `204`. The current month is accepted, never marked.
- `TaxpayerEstimate.figures_fingerprint` is new and additive (optional until M8-03, which will make it required; tracked in PLAN M8-03).
- Flags: `module_history`, `module_tax_engine` or `module_taxpayers` OFF → both routes `404`,
  indistinguishable from an unknown taxpayer. `module_incomes` / `module_expenses` OFF only
  narrow `has_records`; figures are unaffected.
- Open to read-only users and archived taxpayers (no `402`). Own budgets `history_reads`
  600/h (burst 600) and `history_views` 1,200/h (burst 120); every response carries
  `Cache-Control: no-store`.

## Errors

Every error uses the single envelope `ErrorResponse`. `details[].code` carries the rule
(`out_of_range` for a future `fiscal_year` / `period`, `parameters_unavailable` before
2026). Amounts and fingerprints are never echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_history -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 400 | `validation_failed` | Input does not match the contract (malformed `fiscal_year` or path id, unknown query parameter), or a `fiscal_year` after the current one (`out_of_range`). |
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 404 | `not_found` | The taxpayer does not exist or belongs to another user, or `module_history` is effectively OFF (`module_history`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 422 | `business_rule_violation` | `fiscal_year` is before 2026, the first fiscal year with published tax parameters (`parameters_unavailable`). A well-formed request with nothing to show, hence `422`, as `getTaxpayerEstimate`. |
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 429 | `rate_limited` | Rate limit exceeded. |
| `getHistory` | `GET /v1/taxpayers/{taxpayerId}/history` | 500 | `internal_error` | Unexpected server error. |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 400 | `validation_failed` | Input does not match the contract (malformed `period` or path id; missing, malformed or unknown body field, e.g. a fingerprint that is not 64 lower-case hex characters), or a future `period` (`out_of_range`). |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 404 | `not_found` | The taxpayer does not exist or belongs to another user, or `module_history` is effectively OFF (`module_history`, `module_tax_engine` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose). With the module ON a malformed id is `400`. |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 422 | `business_rule_violation` | `period` is before 2026-01, the first month any estimate can exist (`parameters_unavailable`). |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 429 | `rate_limited` | Rate limit exceeded. |
| `markHistoryViewed` | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_history -->
