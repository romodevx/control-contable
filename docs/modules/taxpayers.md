# Module M2: Taxpayers & Onboarding

| | |
|---|---|
| Flag | `module_taxpayers` (OFF by default) |
| Spec | [MODULE-taxpayers.md](../project-specs/MODULE-taxpayers.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Taxpayers`, `x-module: module_taxpayers` (frozen by M2-01) |
| Status | Contract frozen; implementation in progress (PLAN M2-02..M2-12). The full page lands with M2-15 |

## Errors

Every error uses the single envelope `ErrorResponse`. `error.code` is the stable value to
branch on; `details[].field` is the dotted path inside the body (`regimes.0.options.rif_year`)
or the bare parameter name (`taxpayerId`, `end_period`), and `details[].code` carries the
domain rule (`rfc_check_digit`, `taxpayer_archived`, ...; see each operation in the contract).
An RFC is never echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_taxpayers -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `listRegimes` | `GET /v1/regimes` | 400 | `validation_failed` | `person_type` is missing or not `fisica`/`moral`. |
| `listRegimes` | `GET /v1/regimes` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listRegimes` | `GET /v1/regimes` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `listRegimes` | `GET /v1/regimes` | 429 | `rate_limited` | Rate limit exceeded. |
| `listRegimes` | `GET /v1/regimes` | 500 | `internal_error` | Unexpected server error. |
| `listTaxpayers` | `GET /v1/taxpayers` | 400 | `validation_failed` | Input does not match the contract. |
| `listTaxpayers` | `GET /v1/taxpayers` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `listTaxpayers` | `GET /v1/taxpayers` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `listTaxpayers` | `GET /v1/taxpayers` | 429 | `rate_limited` | Rate limit exceeded. |
| `listTaxpayers` | `GET /v1/taxpayers` | 500 | `internal_error` | Unexpected server error. |
| `createTaxpayer` | `POST /v1/taxpayers` | 400 | `validation_failed` | Input does not match the contract, or an RFC, regime, option or period rule (`details[].code`: see the operation description). |
| `createTaxpayer` | `POST /v1/taxpayers` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `createTaxpayer` | `POST /v1/taxpayers` | 402 | `entitlement_required` | Trial ended or cap reached and no active plan (read-only access). |
| `createTaxpayer` | `POST /v1/taxpayers` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `createTaxpayer` | `POST /v1/taxpayers` | 409 | `conflict`, `rfc_already_registered` | The RFC is already registered in the caller's account (active or archived), or the client `id` exists outside the caller's scope (generic, reveals nothing). |
| `createTaxpayer` | `POST /v1/taxpayers` | 422 | `invalid_regime_combination`, `taxpayer_limit_reached` | Two regimes in the body overlap in time and are incompatible, or the user already has 10 taxpayers (archived included). |
| `createTaxpayer` | `POST /v1/taxpayers` | 429 | `rate_limited` | Rate limit exceeded. |
| `createTaxpayer` | `POST /v1/taxpayers` | 500 | `internal_error` | Unexpected server error. |
| `getTaxpayer` | `GET /v1/taxpayers/{taxpayerId}` | 400 | `validation_failed` | Input does not match the contract. |
| `getTaxpayer` | `GET /v1/taxpayers/{taxpayerId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getTaxpayer` | `GET /v1/taxpayers/{taxpayerId}` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `getTaxpayer` | `GET /v1/taxpayers/{taxpayerId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `getTaxpayer` | `GET /v1/taxpayers/{taxpayerId}` | 500 | `internal_error` | Unexpected server error. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 400 | `validation_failed` | Input does not match the contract, or the corrected RFC is invalid. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 402 | `entitlement_required` | Trial ended or cap reached and no active plan (read-only access). |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 409 | `rfc_already_registered` | The corrected RFC is already registered in the caller's account. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 422 | `business_rule_violation` | The taxpayer is archived; unarchive it first. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateTaxpayer` | `PATCH /v1/taxpayers/{taxpayerId}` | 500 | `internal_error` | Unexpected server error. |
| `deleteTaxpayer` | `DELETE /v1/taxpayers/{taxpayerId}` | 400 | `validation_failed` | `confirm=true` is missing, or the input does not match the contract. |
| `deleteTaxpayer` | `DELETE /v1/taxpayers/{taxpayerId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deleteTaxpayer` | `DELETE /v1/taxpayers/{taxpayerId}` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `deleteTaxpayer` | `DELETE /v1/taxpayers/{taxpayerId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `deleteTaxpayer` | `DELETE /v1/taxpayers/{taxpayerId}` | 500 | `internal_error` | Unexpected server error. |
| `archiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/archive` | 400 | `validation_failed` | Input does not match the contract. |
| `archiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/archive` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `archiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/archive` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `archiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/archive` | 429 | `rate_limited` | Rate limit exceeded. |
| `archiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/archive` | 500 | `internal_error` | Unexpected server error. |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 400 | `validation_failed` | Input does not match the contract. |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 402 | `entitlement_required` | Trial ended or cap reached and no active plan (read-only access). |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 429 | `rate_limited` | Rate limit exceeded. |
| `unarchiveTaxpayer` | `POST /v1/taxpayers/{taxpayerId}/unarchive` | 500 | `internal_error` | Unexpected server error. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 400 | `validation_failed` | Input does not match the contract, or a regime, option or period rule. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 402 | `entitlement_required` | Trial ended or cap reached and no active plan (read-only access). |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 409 | `conflict` | The regime is already active for this taxpayer. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 422 | `business_rule_violation`, `invalid_regime_combination` | The regime overlaps an incompatible one, or the taxpayer is archived. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 429 | `rate_limited` | Rate limit exceeded. |
| `addTaxpayerRegime` | `POST /v1/taxpayers/{taxpayerId}/regimes` | 500 | `internal_error` | Unexpected server error. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 400 | `validation_failed` | Input does not match the contract, or an option fails its definition. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 402 | `entitlement_required` | Trial ended or cap reached and no active plan (read-only access). |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 422 | `business_rule_violation` | The taxpayer is archived; unarchive it first. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateTaxpayerRegime` | `PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 500 | `internal_error` | Unexpected server error. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 400 | `validation_failed` | `end_period` is missing, malformed or out of range. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 409 | `conflict`, `regime_has_records` | The regime has records and `confirm=true` was not sent, or the regime has already ended. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 422 | `business_rule_violation` | Records exist after `end_period`, it is the last active regime, or the taxpayer is archived. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 429 | `rate_limited` | Rate limit exceeded. |
| `archiveTaxpayerRegime` | `DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_taxpayers -->
