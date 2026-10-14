# Module M5: Tax engine

| | |
|---|---|
| Flag | `module_tax_engine` (OFF by default; implied OFF when `module_taxpayers` is OFF) |
| Spec | [MODULE-tax-engine.md](../project-specs/MODULE-tax-engine.md) |
| Design | [ADR-0002](../adr/0002-tax-engine.md) (strategies, versioned parameters, rounding) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `TaxEngine`, `x-module: module_tax_engine` (frozen by M5-01, 0.6.0) |
| Status | Contract frozen; implementation in progress (PLAN M5-02..M5-33). The full page lands with the module docs task |

## Contract highlights

- `getTaxpayerEstimate`: `period` is optional (absent = current month, America/Mexico_City).
  Months before 2026 answer `422` (`parameters_unavailable`); future months `400`
  (`out_of_range`). Estimates are derived on every read, never stored.
- One `RegimeEstimate` per regime held in the period, with `status`
  `complete | incomplete | partial | informational`. `incomplete` still computes (zero for
  each missing input) and lists `missing_inputs` (`{key, label, months?}`) for the
  "Completa tus datos" CTA. A missing input is never an error.
- Monthly IVA is consolidated per taxpayer (`totals.iva`); a negative result is a saldo a
  favor, never a payment. Withholdings to remit are their own total, never netted.
- Rules a CPA has not validated still run, but set `cpa_review_pending` / `cpa_topics` and
  the `cpa_review_pending` disclaimer with severity `strong` (non-dismissible banner).
- `Disclaimer {code, severity, text}`; `legend` is the frozen constant "Esto es un
  aproximado. No todos los gastos aplican para tu giro comercial." on every estimate and
  breakdown.
- Breakdown lines carry `legal_basis`, `source` (`records`, `tax_input:<key>`,
  `params:<key>`, `computed`) and `needs_cpa_review`; an `Ajuste por redondeo` line
  (`rounding_adjustment`) makes every breakdown reconcile on screen.
- Tax inputs are per taxpayer, regime and fiscal year, validated against the catalog's
  `required_inputs`. Kinds: `money`, `money_monthly` (prior provisional payments, one per
  month), `rate`, `boolean`. `PUT` replaces all values; read-only users get `402`, reads stay
  open.
- Every response, success or error, carries `Cache-Control: no-store` (S-4).

## Errors

Every error uses the single envelope `ErrorResponse`. `error.code` is the stable value to
branch on; `details[].code` carries the rule (`out_of_range`, `parameters_unavailable`,
`regime_not_active`, `taxpayer_archived`, `unknown_input`, `invalid`; see each operation in
the contract). Amounts and tax-input values are never echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_tax_engine -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 400 | `validation_failed` | Input does not match the contract (malformed period, fiscal year, regime code or path id, unknown query parameter), or a future period or fiscal year (`out_of_range`). |
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 404 | `not_found` | The taxpayer does not exist, belongs to another user, its id is malformed, or `module_tax_engine` (or `module_taxpayers`) is disabled (indistinguishable on purpose). |
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 422 | `business_rule_violation` | The period is before 2026, the first fiscal year with transcribed tax parameters (`parameters_unavailable`, OQ-T9). |
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 429 | `rate_limited` | Rate limit exceeded. |
| `getTaxpayerEstimate` | `GET /v1/taxpayers/{taxpayerId}/estimates` | 500 | `internal_error` | Unexpected server error. |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 400 | `validation_failed` | Input does not match the contract (malformed period, fiscal year, regime code or path id, unknown query parameter), or a future period or fiscal year (`out_of_range`). |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 404 | `not_found` | The taxpayer does not exist, belongs to another user, its id is malformed, or `module_tax_engine` (or `module_taxpayers`) is disabled (indistinguishable on purpose). |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 422 | `business_rule_violation` | The period is before 2026 (`parameters_unavailable`), or the taxpayer does not hold the regime in that period (`regime_not_active`). |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 429 | `rate_limited` | Rate limit exceeded. |
| `getEstimateBreakdown` | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown` | 500 | `internal_error` | Unexpected server error. |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 400 | `validation_failed` | Input does not match the contract (malformed period, fiscal year, regime code or path id, unknown query parameter), or a future period or fiscal year (`out_of_range`). |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 404 | `not_found` | The taxpayer does not exist, belongs to another user, its id is malformed, or `module_tax_engine` (or `module_taxpayers`) is disabled (indistinguishable on purpose). |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 422 | `business_rule_violation` | The taxpayer does not hold the regime in that fiscal year (`regime_not_active`), or the fiscal year is before 2026 (`parameters_unavailable`). |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 429 | `rate_limited` | Rate limit exceeded. |
| `getTaxInputs` | `GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 500 | `internal_error` | Unexpected server error. |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 400 | `validation_failed` | Input does not match the contract, or a value does not fit its definition (`details[].code`: `unknown_input`, `invalid`, `out_of_range`; see the operation). |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 402 | `entitlement_required` | Read-only access (trial over, no active plan). Never a trial-cap refusal. |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 404 | `not_found` | The taxpayer does not exist, belongs to another user, its id is malformed, or `module_tax_engine` (or `module_taxpayers`) is disabled (indistinguishable on purpose). |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 422 | `business_rule_violation` | The taxpayer is archived (`taxpayer_archived`), does not hold the regime in that fiscal year (`regime_not_active`), or the fiscal year is before 2026 (`parameters_unavailable`). |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 429 | `rate_limited` | Rate limit exceeded. |
| `putTaxInputs` | `PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_tax_engine -->
