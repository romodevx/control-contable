# Module M4: Expenses

| | |
|---|---|
| Flag | `module_expenses` (OFF by default; requires `module_taxpayers`) |
| Spec | [MODULE-expenses.md](../project-specs/MODULE-expenses.md) (shares money, idempotency, regime window, IVA, delete/restore and access rules with [MODULE-incomes](../project-specs/MODULE-incomes.md)) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Expenses`, `x-module: module_expenses` (frozen by M4-01 in 0.5.0, amended in 0.6.0 and 0.6.1) |
| Gates | [M4-13 tech lead](../reviews/M3-M4-tech-lead-gate.md) PASS, [M4-14 security](../reviews/M3-M4-security-gate.md) PASS, [M4-12 QA](../reviews/M3-M4-QA.md) |
| Operations | [Incomes and expenses runbook](../runbooks/incomes-expenses.md) |
| Code | `backend/internal/expenses/` (domain, usecase, infra, adapter), `backend/internal/kernel/fiscal/`, `backend/cmd/api/expenses.go`, `records.go`, `ledger.go`, `backend/migrations/20261013210000_expense_categories.sql`, `20261013211000_expenses.sql`, `20261014150000_expenses_tax_facts.sql`, `mobile/lib/features/expenses/`, `mobile/lib/core/records/` |

## What it does

The user records what the **active taxpayer** spent: a category from a server catalog, the
amount before IVA, IVA, how it was paid, whether there is a CFDI, and the payment date. The
server tells the user whether that category usually lowers ISR in their regime, and the user
may override it. Expenses feed the tax engine (M5) through the ledger reader (M4-08):
deductions, creditable IVA and the withholdings the taxpayer must remit.

Expenses are records like [incomes](incomes.md): the same idempotent client id, merge-patch
edit, soft delete, Undo that never consumes a trial slot, read-only rules and archived-taxpayer
freeze. This page covers what is specific to expenses and links to the incomes page for the
rest.

| Operation | Endpoint | Read-only user |
|---|---|---|
| `listExpenseCategories` | `GET /v1/taxpayers/{taxpayerId}/expense-categories?regime=&on=` | allowed |
| `listExpenses` | `GET /v1/taxpayers/{taxpayerId}/expenses?period=YYYY-MM&regime=&cursor=&limit=` | allowed |
| `createExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses` | `402` |
| `getExpense` | `GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` | allowed |
| `updateExpense` | `PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` (merge-patch) | `402` |
| `deleteExpense` | `DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}` (soft) | allowed |
| `restoreExpense` | `POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore` (Undo) | allowed |

Every operation needs a bearer token; foreign ids answer `404 not_found`. Money is `int64` MXN
centavos.

## Record lifecycle

Identical to incomes: see the [state diagram and rules](incomes.md#record-lifecycle). The app's
toast reads "Gasto eliminado · Deshacer". On edit, the default deductibility is recomputed when
the category, regime, payment method, amount, `has_cfdi` or `paid_on` change; a user override
is kept as is.

## Booking date: `paid_on`

Expenses have **one date**, `paid_on`, and it is the booking date for every regime in scope:
deductions and creditable IVA are on a paid basis (no accrual for expenses). It decides the
list month, the estimate month and the regime window. There is no `booked_on` column.

| `paid_on` | Answer |
|---|---|
| Missing | `400`, `paid_on` / `required` |
| After today (America/Mexico_City) | `400`, `paid_on` / `out_of_range` |
| **Before 2026-01-01** | `400`, `paid_on` / `out_of_range` (no tax parameter set before 2026) |
| Outside the regime's window | `422`, `regime_code` / `regime_not_active` (checked before the parameters) |
| In a year without published parameters | Accepted, warning `provisional_parameters` on `paid_on` |

> [!NOTE]
> The contract (0.6.1) says "before 2022-01-01" for `paid_on` and for the categories `on`
> parameter. The served floor is **2026-01-01**. Contract follow-up for the architect in the
> [PLAN M3/M4 follow-ups](../project-specs/PLAN-control-contable.md#m3m4-follow-ups) (#97 FU-2).

## IVA

Same rules as [incomes](incomes.md#iva): `iva_cents` within ±100 centavos of
`amount_cents × rate` for `rate_16`/`rate_8`, `0` for `rate_0`/`exempt`, and the rate must be
allowed on `paid_on` (`iva_rates_allowed` in the categories response). M4 stores the IVA paid;
whether it is creditable is a stored tax fact (`creditable_iva_cents`, below) that M5 reads.

## Categories and deductibility

`listExpenseCategories` is **taxpayer-scoped**: the hints depend on the taxpayer's person type
and regime options, so a global list could not tell RESICO PF from RESICO PM or apply 606 and
625 options. It returns, for one `regime` on the date `on` (default today):

- `items[]`: `code`, `name` (es-MX), `deductibility`, `default_deductible`, `note`,
  `legal_basis` (article level, e.g. `LISR art. 103`);
- `iva_rates_allowed` on that date;
- `cash_payment_limit_cents` (dated parameter data, `200000` in 2026);
- `legend`, the mandatory text (below).

The catalog (26 categories × regimes) is data written by migrations only; the full table is in
[MODULE-expenses §2.2](../project-specs/MODULE-expenses.md).
**Every row is pending CPA review** ([CPA-REVIEW.md](../reviews/CPA-REVIEW.md)).

| `deductibility` (the hint) | Meaning | `default_deductible` |
|---|---|---|
| `deductible` | Normally lowers ISR in this regime | `true` |
| `partially_deductible` | Has caps or conditions (restaurants, vehicles, investments by depreciation) | `true` or `false`; conservative `false` for investments, restaurants and "otros" |
| `not_deductible` | Does not lower ISR here (e.g. `gastos_personales`: annual return only) | `false` |
| `isr_not_applicable` | ISR does not take deductions in this regime | `false` |

Regime options change the hint (domain rule, not catalog data):

| Taxpayer | Effect |
|---|---|
| 606 with `deduction_mode = blind_35` | Every category becomes `isr_not_applicable` except `predial` (added to the 35 %) |
| 625 with `definitive_option = true` | Every category becomes `isr_not_applicable` |
| 626 persona física (RESICO PF) | Always `isr_not_applicable`; the IVA paid may still be creditable |
| Informational regimes | `isr_not_applicable` |

### Effective `deductible`

```text
default    = rule.default_deductible (after the options above)
             AND has_cfdi
             AND NOT cash_over_limit
deductible = the user's boolean if they sent one, else default
```

- `deductible_source` is `category_default` or `user`. `deductible: null` in a `PATCH` clears
  the override.
- `deductible: true` means "the user claims it lowers ISR"; M5 adds a disclaimer for it.
- An override to `true` against a `not_deductible` hint adds the warning
  `category_not_deductible`; against `isr_not_applicable`, `not_isr_deductible_in_regime`.
  Neither is an error.
- `has_cfdi: false` adds the warning `no_cfdi` and makes the default `false`. A `cfdi_uuid`
  with `has_cfdi: false` is `400 has_cfdi_mismatch`. An unknown category is
  `400 unknown_category`.

## Cash payments over $2,000

LISR 27 fr. III, served as dated parameter data (`deduction.cash_payment_limit`), never
hard-coded. **Pending CPA review** (OQ-E4).

`cash_over_limit` is true when `payment_method = cash` and either `amount_cents + iva_cents`
is above `cash_payment_limit_cents` (the total **with** IVA), or the category is `combustible`
(fuel paid in cash is never deductible, whatever the amount).

| Situation | `deductible` | Warnings | Creditable IVA |
|---|---|---|---|
| Cash over the limit, no override | `false` (default) | `cash_over_threshold` on `payment_method` | 0 |
| Cash over the limit, user overrides to `true` | `true` (`user`) | `cash_over_threshold` + **`iva_not_creditable_cash_payment`** on `payment_method` | 0: the ISR claim stands, the IVA of a cash payment over the limit is not creditable (LIVA 5 fr. II), whatever the override |
| Cash at or under the limit (not fuel) | category default | none | as usual |

The server returns `iva_not_creditable_cash_payment` only with the override. The app shows its
note as soon as the user picks cash over the limit (QA D-1, deferred product decision).

The stored tax facts, set only by the domain and never by a client:

| Column | Meaning |
|---|---|
| `cash_over_limit` | The cash rule above |
| `creditable_iva_cents` | The creditable-IVA candidate M5 sums: the IVA when there is a CFDI, the payment is not cash over the limit, the category's rule allows IVA credit, and the expense is deductible (or the payer is RESICO PF); otherwise 0 ⚠️ CPA |

## Withholdings the taxpayer made

`isr_withheld_cents` and `iva_withheld_cents` on an expense are what the taxpayer withheld
**from its supplier** (retenciones a proveedores), which M5 adds to the withholdings to remit.

- Only a **persona moral** may withhold. A persona física sending a value above 0 gets
  `422 withholding_not_applicable`.
- Bounds as incomes: ISR ≤ `amount_cents`, IVA ≤ `iva_cents` (`422 withholding_exceeds_base`).
- The form shows the section for morales only, collapsed, open by default for
  `servicios_profesionales`, `renta_local` and `fletes`. Suggestions (UX only): ISR 10 % and
  IVA 2/3 on fees and rent, IVA 4 % of the price on freight.

## Mandatory legend

The exact text **"Esto es un aproximado. No todos los gastos aplican para tu giro
comercial."** travels as `legend` on every expense, every list page (`ExpensePage.legend`, even
an empty month) and the categories response. The app renders the server field under the
category and its hint (always visible) and in the list header; it never keeps its own copy.

## Warnings

| Code | Field | When |
|---|---|---|
| `cash_over_threshold` | `payment_method` | Cash over the limit, or cash fuel |
| `iva_not_creditable_cash_payment` | `payment_method` | Cash over the limit **and** an override to `deductible: true` (0.6.0) |
| `no_cfdi` | `has_cfdi` | `has_cfdi: false` |
| `category_not_deductible` | `deductible` | Override to `true` against a `not_deductible` hint |
| `not_isr_deductible_in_regime` | `deductible` | Override to `true` against `isr_not_applicable` |
| `duplicate_cfdi_uuid` | `cfdi_uuid` | Another live expense of the taxpayer has the folio |
| `provisional_parameters` | `paid_on` | The year's parameters are not published yet |

Non-deductibility is never an error: it is a warning plus `deductible: false`.

## Trial cap, access and idempotency

As [incomes](incomes.md#trial-cap-and-access), with the expense counter:

- **5 expenses per user** across all taxpayers, independent of the 5 incomes. The 6th answers
  `402 entitlement_required`, `expenses_remaining` / `trial_cap_reached`.
- Deleting never frees a slot; restore (Undo) never consumes one.
- Read-only users may list, get, read categories, delete and restore; create and edit answer
  `402`.
- Idempotent client `id`: same table as [incomes](incomes.md#idempotent-client-ids). The
  same-fields comparison includes the override and the normalized withholdings.
- Write budget: 300 expense mutations per user per hour (scope `expense_writes`).

### Try it locally

Continue the [M3 walkthrough](incomes.md#try-it-locally) (same `API`, `AUTH` and
`TAXPAYER_ID`, a física with 612), with `module_expenses` ON.

```sh
# 1. Categories resolved for 612 on the payment date, with the cash limit and the legend.
curl -s "$API/v1/taxpayers/$TAXPAYER_ID/expense-categories?regime=612&on=2026-10-06" -H "$AUTH" \
  | jq -c '{papeleria: (.items[] | select(.code == "papeleria")), cash_payment_limit_cents, legend}'

# 2. 2,000.00 + IVA 320.00 of papelería paid in cash: 201, over the limit, not deductible.
EXPENSE_ID=$(uuidgen | tr 'A-Z' 'a-z')
curl -s -X POST "$API/v1/taxpayers/$TAXPAYER_ID/expenses" -H "$AUTH" -H 'Content-Type: application/json' \
  -d "$(jq -nc --arg id "$EXPENSE_ID" '{id: $id, regime_code: "612", category_code: "papeleria",
    amount_cents: 200000, iva_rate: "rate_16", iva_cents: 32000, payment_method: "cash",
    has_cfdi: true, paid_on: "2026-10-06"}')" \
  | jq -c '{deductible, deductible_source, warnings: [.warnings[].code]}'

# 3. The user claims it anyway: the override stands for ISR, its IVA is not creditable.
curl -s -X PATCH "$API/v1/taxpayers/$TAXPAYER_ID/expenses/$EXPENSE_ID" -H "$AUTH" \
  -H 'Content-Type: application/json' -d '{"deductible": true}' \
  | jq -c '{deductible, deductible_source, warnings: [.warnings[].code]}'

# 4. deductible: null clears the override; the default applies again.
curl -s -X PATCH "$API/v1/taxpayers/$TAXPAYER_ID/expenses/$EXPENSE_ID" -H "$AUTH" \
  -H 'Content-Type: application/json' -d '{"deductible": null}' \
  | jq -c '{deductible, deductible_source}'

# 5. The month, with the legend the list header shows.
curl -s "$API/v1/taxpayers/$TAXPAYER_ID/expenses?period=2026-10" -H "$AUTH" \
  | jq -c '{items: (.items | length), totals, legend}'
```

Output, in order:

```text
{"papeleria":{"code":"papeleria","deductibility":"deductible","default_deductible":true,"legal_basis":"LISR art. 103","name":"Papelería y artículos de oficina","note":null},"cash_payment_limit_cents":200000,"legend":"Esto es un aproximado. No todos los gastos aplican para tu giro comercial."}
{"deductible":false,"deductible_source":"category_default","warnings":["cash_over_threshold"]}
{"deductible":true,"deductible_source":"user","warnings":["cash_over_threshold","iva_not_creditable_cash_payment"]}
{"deductible":false,"deductible_source":"category_default"}
{"items":1,"totals":{"amount_cents":200000,"count":1,"isr_withheld_cents":0,"iva_cents":32000,"iva_withheld_cents":0},"legend":"Esto es un aproximado. No todos los gastos aplican para tu giro comercial."}
```

$2,000.00 + $320.00 = $2,320.00 is above the $2,000.00 limit, so step 2 is over it even though
the subtotal is exactly the limit. Delete and Undo work as in the M3 walkthrough (step 7).

## List

As [incomes](incomes.md#list), with `paid_on` as the booking date: `period` (required) and
`regime` filters, keyset pagination on `(paid_on DESC, id DESC)`, `totals` for the whole
filtered month, plus `legend`.

## Flag behaviour (`module_expenses`)

- OFF (the default): all 7 operations, **the categories endpoint included**, answer
  `404 not_found`, even with `module_incomes` ON. The app shows no Gastos tab.
- **Prerequisite `module_taxpayers`**: with it OFF, expenses are hidden whatever their own flag
  says (`hidden` in `make -C backend flags`).
- Independent of `module_incomes`. Hidden data still counts when a regime is ended, still
  cascades on deletion, and is still read by the M5 estimate (OQ-8).
- Suites: `backend/adapter/http/gen/records_module_test.go`,
  `backend/cmd/api/records_preset_smoke_integration_test.go`,
  `mobile/test/features/expenses/*_flag_test.dart`. Member of `demo_core`,
  `demo_monetization` and `demo_full`.
- **Do not enable it in staging or production** until every precondition in
  [runbook §5](../runbooks/incomes-expenses.md#5-enable-module_incomes-and-module_expenses-in-a-shared-environment) holds.

```sh
make -C backend flags ARGS="enable module_expenses"   # needs module_taxpayers ON to take effect
make -C backend flags ARGS="disable module_expenses"
```

## Errors

Every error uses the single envelope `ErrorResponse`; `details[].code` carries the rule
(`unknown_category`, `has_cfdi_mismatch`, `withholding_not_applicable`, ...). Amounts,
descriptions and CFDI folios are never echoed. Record bodies carry `Cache-Control: no-store`;
error responses do not yet (deferred, as for [incomes](incomes.md#errors)).

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

## Decisions

Full text in [MODULE-expenses §10](../project-specs/MODULE-expenses.md#10-decisions-coordinator-2026-10-13--all-decided).

| # | Decision | Effect |
|---|---|---|
| OQ-E1 | Taxpayer-scoped categories endpoint | Replaces the global `GET /v1/expense-categories` |
| OQ-E2 | The form is its own page | 13 inputs; 6 visible at first |
| OQ-E3 | Ship the category seed, CPA review before launch | Investments, restaurants, personal and "otros" default to non-deductible |
| OQ-E4 | Cash limit $2,000 on the total with IVA; fuel only electronic | Dated parameter data, CPA to confirm |
| OQ-E5 | Expense warnings use the generic `Warning` (`field`, append-only codes) | One warning shape across modules |
| OQ-E6 | First payment method is *Tarjeta de débito* | Then the last one used, cleared on sign-out |

## API reference

```sh
make api-docs                 # build/api-docs/index.html, rendered from docs/api/openapi.yaml
make api-error-tables         # regenerate the error table on this page after a contract change
```

## Known limits

Sources as in [incomes](incomes.md#known-limits). Specific to expenses:

- The category catalog, the cash rule and the creditable-IVA rule are pending CPA review.
- The IVA note for cash over the limit shows before the override in the app, while the server
  warns only with it (QA D-1).
- No depreciation schedules, per-item caps (cars, restaurants, viáticos), proportional IVA
  credit, DIOT or user-defined categories in v1.
