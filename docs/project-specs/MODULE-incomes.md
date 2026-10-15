# MODULE-incomes — M3 Incomes (`module_incomes`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-13 · MoSCoW: **Must** · Depends on: M2, P0 (access policy), M5-08 (`RateCatalog`)
Sources: PRD §2, §5, §6 (US-I1, US-I2), §7, §9 · MODULES.md M3 · PLAN M3-01..M3-14 ·
research `docs/research/tax-regimes-mx-2026.md` (§0, A.2.3–A.2.6, A.3, B.2, B.2.3, B.3, Part C) ·
ADR-0002 §1, §3, §4 · `docs/architecture/data-model.md` (incomes, user_usage, trial-cap algorithm) ·
MODULE-taxpayers.md §1 (ports), §2.6, §2.7, §2.10 · `backend/internal/platform/access` (P0-16) ·
reviews P0-28 O3, P0-29 FU8, P0-30 D-3/D-6, M1-17 #43 · `rules/ui-forms-and-modals.md`.

Sibling: [MODULE-expenses.md](MODULE-expenses.md) reuses §2.3 (idempotency), §2.6 (trial cap and
access), §2.7 (delete/undo), §5 and §6 of this file. Changes here apply there unless that file
says otherwise.

## 1. Scope

The user records, lists, edits, deletes and restores the incomes of the **active taxpayer**.
Each income belongs to exactly one taxpayer + regime the taxpayer holds on the income's dates.
M3 stores facts as the user reads them from the CFDI or bank statement; it never computes a
tax. M5 reads the incomes through `LedgerReader` (M3-07).

| Op | Endpoint | Contract |
|---|---|---|
| listIncomes | `GET /v1/taxpayers/{taxpayerId}/incomes?period=YYYY-MM&regime=&cursor=&limit=` | exists |
| createIncome | `POST /v1/taxpayers/{taxpayerId}/incomes` | exists; changes in §8 |
| getIncome / updateIncome | `GET / PATCH …/incomes/{incomeId}` | exists; changes in §8 |
| deleteIncome | `DELETE …/incomes/{incomeId}` (soft delete) | exists; add `422` |
| restoreIncome | `POST …/incomes/{incomeId}/restore` | exists; add `402`, `422` |
| getIncomeDefaults | `GET …/incomes/defaults?regime=&on=&counterparty_type=` | **new (OQ-7)** |

Ports M3 consumes (each declared consumer-side in `internal/incomes/usecase`; archtest forbids
importing another module):
- `TaxpayerRegimeChecker.CheckRegime(ctx, user, taxpayer, regime, period)` (M2): called once per
  distinct period among the income's non-null dates, **inside the income's unit of work** so its
  `FOR SHARE` locks hold until commit. Its errors (404, 422 `taxpayer_archived` /
  `regime_not_active`) pass through unchanged.
- `TaxpayerProfileReader.Profile` (M2): person type and regime options (625 `activities`,
  `rfc_given_to_platform`) for the withholding rules and defaults.
- `access` policy `ReserveRecord(ctx, user, RecordIncome)` (P0-16): read-only → `402`; trial cap
  → `402` naming `incomes_remaining`.
- `RateCatalog` (M5-08): allowed IVA rates and suggested withholding rates on a date.

Ports M3 provides:
- `LedgerReader` income aggregates (M3-07) for M5/M6.
- `RegimeRecords.LatestRecordPeriod` (M2): replaces `taxpayers/infra.NoRecords` with an adapter
  over incomes **and** expenses (wired in M3-07, extended in M4-08): the latest period of any
  non-null business date among **non-deleted** records of `(taxpayer, regime)`.

## 2. Domain rules

### 2.1 Fields

| Field | Type | Required | Rule |
|---|---|---|---|
| `id` | uuid | client SHOULD send | Idempotency key (§2.3). Server generates a uuidv7 when absent |
| `regime_code` | `RegimeCode` | yes | Held by the taxpayer on every non-null date of the income (§2.4) |
| `amount_cents` | int64 centavos | yes | **≥ 1**, subtotal **before IVA**, MXN. Upper bound `NonNegativeCents.maximum` |
| `iva_rate` | `rate_16 \| rate_8 \| rate_0 \| exempt` | yes | Valid on the IVA date (§2.5) |
| `iva_cents` | int64 centavos | yes | §2.5 consistency |
| `counterparty_type` | `persona_fisica \| persona_moral \| publico_general \| plataforma` | **yes (contract change)** | Decides which withholdings may exist (§2.6) |
| `isr_withheld_cents` | int64 centavos | no, default 0 | §2.6 |
| `iva_withheld_cents` | int64 centavos | no, default 0 | §2.6 |
| `invoiced_on` | date | accrual regimes: yes; else no | §2.4 |
| `collected_on` | date | cash regimes: yes; accrual: no (`null` = pending collection) | §2.4 |
| `platform_activity` | `transport_delivery \| lodging \| goods_services` | 625 only (**new**) | Must be one of the taxpayer's 625 `activities`; defaults to it when there is exactly one; forbidden for other regimes |
| `cfdi_uuid` | string | no (**new**) | CFDI folio fiscal: 8-4-4-4-12 hex, normalized to upper case. Duplicates are a warning (§2.8) |
| `description` | string ≤ 200 | no | Trimmed, NFC; empty → `null`. May hold a client's name: never logged |

- **Currency**: MXN only; there is no currency field. Foreign-currency income is converted by the
  user (out of scope, §7).
- Money crosses the API and DB as `int64` centavos (`money.Cents`). The server never parses a
  float (`forbidigo`); FU8 fuzz targets `FuzzParseCents`/`FuzzParseRate` land **before M3-03**.

### 2.2 Cash vs accrual: which date the estimate uses

The regime's income basis is catalog data: a new column `regimes.income_basis`
(`cash | accrual`, M3-02 forward migration), not a `switch` in code. The contract serves it as
`Regime.income_basis` and `TaxpayerRegime.income_basis` (0.16.0, ADR-0010), so clients never
mirror this table.

| Basis | Regimes | ISR date | IVA date | Required date |
|---|---|---|---|---|
| accrual | PM Título II: 601, 620, 622, 623, 624 | `invoiced_on` | `collected_on` (LIVA 1-B: IVA is always cash) | `invoiced_on` |
| cash | every PF regime, 626 PM, 603, 610 | `collected_on` | `collected_on` | `collected_on` |

- An accrual income with `collected_on = null` is **"Por cobrar"**: it counts for ISR in its
  invoice month and for IVA only when the user sets `collected_on` (edit).
- The **booking date** is the ISR date: `invoiced_on` for accrual, `collected_on` for cash. It
  decides the list period (§2.9) and is stored as `booked_on` (derived on every write, indexed).
- Either date may precede the other (advances are collected before the invoice).
- Both dates: `2022-01-01 ≤ date ≤ today` (America/Mexico_City, `Clock` port). Future dates are
  refused (`400 out_of_range`).

### 2.3 Idempotency (double submit)

The form generates the `id` once per form session and resends it on every retry; "Guardar y
crear otro" starts a new session with a new `id`. The server, **before** the access policy:

| The `id` … | Response | Slot used |
|---|---|---|
| is new | `201` | yes |
| exists for the caller, same taxpayer, not deleted, same persisted fields (canonical compare after normalization) | `200` with the stored income | no |
| exists for the caller with different fields, another taxpayer, or soft-deleted | `409 conflict`, detail `id_reused` | no |
| exists for another user (RLS hides it; PK collision on insert) | `409 conflict`, no detail, same message | no |

A retry at 5/5 in the trial therefore gets `200`, never `402`. The 409 for a foreign id is
indistinguishable from a reuse by the caller except for the detail: it reveals no ownership.

### 2.4 Regime window

- Every non-null date's period must lie inside the taxpayer regime's `[start_period, end_period]`
  and the taxpayer must not be archived: `CheckRegime` per distinct period (§1).
- Moving an income to another regime (PATCH `regime_code`) re-runs every rule for the new regime.
- Records in **informational** regimes (605, 614, …) are allowed; the form says
  "Este régimen es informativo: el ingreso no cambia tu estimación" (OQ-4).

### 2.5 IVA

- `iva_rate` must be allowed on the **IVA date** (`collected_on`, or `invoiced_on` for an accrual
  income still pending collection) by `RateCatalog`. Today `rate_8` is valid until 2026-12-31
  (border decree extension, DOF 31-12-2025); after that `422 business_rule_violation`
  `rate_not_valid_on_date`. No date is hard-coded.
- `rate_8` is a user declaration of border eligibility (research Part C); the default is `rate_16`.
- `rate_0` and `exempt` → `iva_cents` must be `0`.
- `rate_16` / `rate_8` → `|iva_cents − round(amount_cents × rate)| ≤ 100` centavos
  (`IvaTolerance`, OQ-3), computed with `money.Exact` and one `Round`. CFDIs compute IVA per
  line, so totals drift a few centavos; the tolerance absorbs it without letting a typo through.
  Outside it → `422 iva_amount_mismatch`. The client prefills `iva_cents` from the rate; the user
  may correct it under "Más detalles". This is a check, not stored rounding (PRD §9).

### 2.6 Withholdings received (retenciones)

> **Superseded (2026-10-15) by [MODULE-withholdings.md](MODULE-withholdings.md).** Withholdings are
> now **derived by the server** from `counterparty_kind`, `operation_kind`, `residential_rent` and
> `via_platform` (its §2.1, §3); the user never types them (owner decision, no override).
> `counterparty_type` is replaced by `counterparty_kind`; `isr_withheld_cents` /
> `iva_withheld_cents` become read-only outputs; the "Retenciones" inputs (§7), the suggestion
> table below, the bounds and `withholding_differs_from_rate` are retired with contract 0.9.0
> (PLAN §9, WH tasks). The text below is kept as the record of the shipped M3 behavior until WH-06
> merges.

Applicability is decided by the taxpayer's person type, the regime and `counterparty_type`
(research A.2.3–A.2.6, B.2.3). Outside the table, a withholding > 0 →
`422 withholding_not_applicable` on that field.

| Taxpayer regime | Counterparty | ISR withheld (suggested) | IVA withheld (suggested) |
|---|---|---|---|
| 612 PF | persona_moral | 10 % of `amount` (LISR 106) | 2/3 of IVA (RLIVA 3) |
| 606 PF | persona_moral | 10 % of `amount` (LISR 116) | 2/3 of IVA when IVA > 0 |
| 626 PF | persona_moral | 1.25 % of `amount` (LISR 113-J) | 2/3 of IVA (services, rent) |
| 625 PF | plataforma | by `platform_activity`: 2.1 % transport/delivery, 4 % lodging, 2.5 % goods/services (LIF 2026 art. 25-VI); **20 %** when `rfc_given_to_platform = false` | 50 % of IVA; **100 %** without RFC (LIVA 18-J) |
| any PM regime | plataforma | 2.5 % (2026, B.2.3) | — (must be 0) |
| any PM regime | persona_moral | — (must be 0) | allowed (freight 4 % case), no suggestion |
| any | persona_fisica, publico_general | — (must be 0) | — (must be 0) |

- Suggested rates come from `RateCatalog` on the booking date (LIF rates change yearly); the
  client prefills them and the user corrects them to the CFDI. The stored value is what the user
  entered.
- Bounds: `isr_withheld_cents ≤ amount_cents`, `iva_withheld_cents ≤ iva_cents`
  (`422 withholding_exceeds_base`). A value that differs from the suggestion by more than the
  §2.5 tolerance is accepted with the warning `withholding_differs_from_rate` (the CFDI wins).
- 625 incomes collected directly (not through the app) use `counterparty_type` ≠ `plataforma`;
  M5 self-pays them at the activity's rate (ADR-0002 §2).

### 2.7 Delete, undo, restore

- `DELETE` sets `deleted_at` (soft). The income leaves lists, totals, `LedgerReader` and
  `LatestRecordPeriod` at once. `204`; deleting an already-deleted income is `204` (idempotent).
- `POST …/restore` clears `deleted_at` after re-running §2.4. **DECIDED (coordinator,
  2026-10-14; contract 0.6.1): a read-only user may restore (Undo) their own deleted records**:
  no access check runs and restore never answers `402` (it gives back the user's own record
  and creates nothing). Another user's record stays `404`.
  **DECIDED (coordinator, 2026-10-13): restore never consumes a trial slot.** The slot was
  consumed once at creation and never freed, so the restored record never left the count:
  Undo works at 5/5 with no paywall. Restoring a live income is `200` (no-op). After the 30-day
  purge (M9-08) → `404`.
- **Backend requirement (M3-05/M4-05):** P0-16 `access` currently documents "restoring a deleted
  record reserves a slot like a create". The restore use cases must **not** call
  `ReserveRecord` (no `user_usage` bump) and run no access check (amended 2026-10-14: not even
  `ErrReadOnly`). Tests: at 5/5, delete + restore → `200` and `incomes_created` stays 5;
  read-only, delete + restore → `200`.
- `GET` of a soft-deleted income → `404`.
- On an **archived taxpayer**, delete and restore are refused (`422 taxpayer_archived`): archive
  freezes the data (OQ-5).
- The trial counter never goes down (`user_usage` trigger refuses decreases).

### 2.8 Warnings (non-blocking, `Income.warnings`, contract change)

| Code | Field | When |
|---|---|---|
| `duplicate_cfdi_uuid` | `cfdi_uuid` | Another live income of the same taxpayer has the UUID. Legit for partial collections (PPD with complementos de pago), so never an error (OQ-9) |
| `withholding_differs_from_rate` | `isr_withheld_cents` / `iva_withheld_cents` | §2.6 |

Warnings are computed per write, not stored, and never echo the UUID or the description.

### 2.9 List

- `period` (required) selects incomes whose **booking date** is in the month; `regime` filters.
- Keyset pagination `(booked_on DESC, id DESC)`, `limit` default 50, max 100; opaque cursor.
- `totals` cover the whole filtered period (not the page): count, amount, IVA, ISR and IVA
  withheld. Accrual incomes pending collection are listed with the badge "Por cobrar".
- p95 < 300 ms (PRD §9); `EXPLAIN ANALYZE` of the list and totals queries in M3-05.

### 2.10 Access (P0-16, read-only rule as in MODULE-taxpayers §2.10)

| Operation | Trial with slots | Trial at 5/5 incomes | Read-only (no trial, no plan) | pro / comp |
|---|---|---|---|---|
| list, get, defaults | ✓ | ✓ | ✓ | ✓ |
| create | ✓ (uses a slot) | `402` `incomes_remaining` | `402` | ✓ |
| restore a deleted income | ✓ (no slot) | ✓ (no slot) | ✓ (no slot) | ✓ |
| edit | ✓ | ✓ | `402` | ✓ |
| delete | ✓ | ✓ | ✓ | ✓ |

- Caps are **per user, across all taxpayers**, incomes and expenses **independent** (5 + 5): a
  refused 6th income leaves expense slots.
- Order inside the create transaction: idempotency lookup → `access` read-only check →
  `CheckRegime` → domain rules → `ReserveRecord` (atomic conditional `UPDATE user_usage`) →
  insert. A failed insert rolls the slot back with the transaction.

## 3. User stories (Given / When / Then)

### US-I1 Record an income
- **Given** a física with 612 active and `module_incomes` ON, **when** I open "Nuevo ingreso",
  **then** the regime is preselected (only one), the date is today, IVA is 16 %, and the
  counterparty is my last choice (first time: persona física).
- **When** I enter 10,000.00, counterparty persona moral, **then** IVA shows 1,600.00, the
  withholdings section opens prefilled with ISR 1,000.00 and IVA 1,066.67, and the total
  "Neto a recibir 9,533.33" updates live.
- **When** I save, **then** `201` with `amount_cents: 1000000`, `iva_cents: 160000`,
  `isr_withheld_cents: 100000`, `iva_withheld_cents: 106667`, and the list for the month shows it
  with the new totals.
- **Given** a moral with 601, **when** I record an income invoiced on 2026-09-05 with no
  collection date, **then** `201`; it lists under September with "Por cobrar".
- **Given** 612, **when** I omit `collected_on`, **then** `400 validation_failed`
  (`collected_on` / `required`).
- **Given** a 625 taxpayer with activities `[lodging]`, **when** I record 15,000.00 from
  plataforma, **then** `platform_activity` defaults to `lodging` and ISR suggests 600.00.
- **Given** 625 with `[transport_delivery, goods_services]`, **when** I omit
  `platform_activity`, **then** `400` (`platform_activity` / `required`); with `lodging`,
  **then** `422` (`activity_not_held`).
- **Given** counterparty persona física, **when** I send `isr_withheld_cents: 500`, **then**
  `422 business_rule_violation` (`isr_withheld_cents` / `withholding_not_applicable`).
- **When** `iva_rate: exempt` with `iva_cents: 100`, **or** rate 16 % on 10,000.00 with IVA
  1,500.00, **then** `422` (`iva_cents` / `iva_amount_mismatch`); IVA 1,600.40 → `201`.
- **Given** an income collected on 2027-01-05 with `rate_8` and no 2027 parameter extending
  it, **then** `422` (`iva_rate` / `rate_not_valid_on_date`).
- **When** `collected_on` is tomorrow or before 2022-01-01, **then** `400` (`out_of_range`).
- **When** `amount_cents: 0`, **then** `400` (`amount_cents` / `must_be_positive`).

### US-I1b Save and create another
- **When** I tap "Guardar y crear otro", **then** the income is saved, the form keeps regime,
  counterparty, IVA rate and date, clears amount, IVA, withholdings, description and UUID,
  generates a new `id`, and focuses the amount.

### US-I1c Double submit
- **Given** a slow network, **when** I tap "Guardar" twice, **then** the second tap is ignored
  while the busy state shows (the button is never disabled) and the server sees one `id`.
- **Given** the first response was lost, **when** the app retries with the same `id` and body,
  **then** `200` with the stored income, no second record and no second trial slot, even at 5/5.
- **When** the same `id` arrives with a different amount, **then** `409 conflict` (`id_reused`).

### US-I2 List and filter a month
- **Given** incomes in September across 612 and 606, **when** I open Ingresos for 2026-09,
  **then** I see them newest first with month totals; **when** I filter 606, **then** items and
  totals show only 606.
- **When** I swipe to August, **then** the August page loads; an empty month shows
  "Sin ingresos en agosto" and a "Registrar ingreso" action.
- **Given** 120 incomes in a month, **then** pages of 50 load on scroll and totals cover all 120.

### US-I3 Edit
- **When** I change the amount and IVA of an income, **then** `200`, totals update.
- **When** I move an income to a regime the taxpayer did not hold on its date, **then** `422`
  (`regime_code` / `regime_not_active`).
- **When** a 601 income gets `collected_on` set, **then** it loses "Por cobrar".

### US-I4 Delete, Undo, restore
- **When** I delete an income, **then** it disappears, totals drop, and a toast
  "Ingreso eliminado · Deshacer" stays until I dismiss it (no auto-dismiss, `role="status"`).
- **When** I tap Deshacer, **then** `POST …/restore` → `200`, the income is back in place.
- **Given** a trial user who created 5 incomes and deleted one, **when** they create another,
  **then** `402` (`incomes_remaining`): deleting never frees a slot.
- **Given** a trial user at 5/5 who deleted an income, **when** they tap Deshacer (restore),
  **then** `200`, no paywall, and `incomes_created` stays 5 (the slot never left the count).
- **Given** the regime was ended before the income's date after the delete, **when** I restore,
  **then** `422` (`regime_not_active`).

### US-I5 Trial cap and read-only
- **Given** a trial user with 5 incomes created (any taxpayers) and 2 expenses, **when** they
  create a 6th income, **then** `402 entitlement_required` with `details[0].field =
  incomes_remaining`, `code = trial_cap_reached`; **when** they create an expense, **then** `201`.
- **Given** 4/5 and 3 concurrent creates, **then** exactly one `201` and two `402`
  (integration test, real PostgreSQL, `synctest`/goroutines).
- **Given** a read-only user, **when** they list, delete or restore (Undo) their own income,
  **then** it works; **when** they create or edit, **then** `402` and nothing changes.

### US-I6 Isolation and archive
- **Given** user B's taxpayer or income id, **when** user A calls any M3 endpoint with it,
  **then** `404 not_found` (never 403).
- **Given** an archived taxpayer, **when** I list, **then** `200`; **when** I create, edit,
  delete or restore, **then** `422` (`taxpayer_archived`).
- **Given** a malformed `taxpayerId`, **then** `404` regardless of other query errors (QA O3).

## 4. Errors → ErrorCode

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Schema violation, unknown field, bad enum, bad UUID/date format | 400 | `validation_failed` | per field (`invalid`, `required`, `too_long`) |
| Missing date for the regime's basis; missing `platform_activity` | 400 | `validation_failed` | `required` |
| `amount_cents` = 0 | 400 | `validation_failed` | `must_be_positive` |
| Date in the future or before 2022-01-01 | 400 | `validation_failed` | `out_of_range` |
| `platform_activity` on a non-625 income | 400 | `validation_failed` | `not_for_regime` |
| Missing/invalid session | 401 | `unauthorized` | |
| Read-only user (create, edit; never restore) | 402 | `entitlement_required` | |
| Trial cap spent (create only; never restore) | 402 | `entitlement_required` | `trial_cap_reached` (field `incomes_remaining`) |
| Flag OFF (or `module_taxpayers` OFF), foreign/unknown taxpayer or income, purged income, malformed path id | 404 | `not_found` | |
| `id` reused with other content / other taxpayer / deleted; foreign `id` | 409 | `conflict` | `id_reused` (caller's only) |
| Regime not held on a date; taxpayer archived | 422 | `business_rule_violation` | `regime_not_active`, `taxpayer_archived` (from M2) |
| IVA rate not valid on the IVA date | 422 | `business_rule_violation` | `rate_not_valid_on_date` |
| IVA amount inconsistent with the rate | 422 | `business_rule_violation` | `iva_amount_mismatch` |
| Withholding not allowed / above its base | 422 | `business_rule_violation` | `withholding_not_applicable`, `withholding_exceeds_base` |
| 625 activity not among the taxpayer's | 422 | `business_rule_violation` | `activity_not_held` |
| Per-user write budget exceeded | 429 | `rate_limited` | |
| Anything else | 500 | `internal_error` | |

No new `ErrorCode` values. Detail codes go into the contract descriptions in M3-01. Messages
are es-MX safe text (localized by the client) and never carry amounts, descriptions or UUIDs.

## 5. Security requirements (gate M3-13)

| # | Requirement | Test |
|---|---|---|
| S-1 | `incomes`: ENABLE + FORCE RLS in its first migration; owner policies for SELECT/INSERT/UPDATE/DELETE (`user_id = app_user_id()`); `cc_app` not owner, NOBYPASSRLS | `pgtest.AssertIsolation` (A cannot read/write B), catalog sweep |
| S-2 | Composite FK `(taxpayer_id, regime_code, user_id) → taxpayer_regimes`; `ON DELETE CASCADE` from the taxpayer; DB CHECKs: `amount_cents > 0`, amounts ≥ 0, `iva_withheld_cents ≤ iva_cents`, `isr_withheld_cents ≤ amount_cents`, `iva_rate` enum, `collected_on IS NOT NULL OR invoiced_on IS NOT NULL`, `char_length(description) ≤ 200` | Cross-user composite-FK rejection; CHECK tests |
| S-3 | `user_id` from the session and `taxpayer_id` from the path only (no mass assignment); `id` immutable; `additionalProperties: false` on create/update | Contract tests |
| S-4 | Trial slot reserved atomically in the create transaction (`ReserveRecord`); restore never reserves (no double count) | Concurrency test (US-I5); restore-at-5/5 test |
| S-5 | Description, CFDI UUID and amounts never logged, traced or put in error messages | Log-capture test at debug level |
| S-6 | Audit events `income.created/updated/deleted/restored` with user id, taxpayer id and income id only | Log-capture test |
| S-7 | Per-user write budget on income mutations (default 300/h, ADR-0005 limiter) | 429 test |
| S-8 | Repositories never `Commit`/`Rollback`/`SET` (P0-30 D-3) | archtest |
| S-9 | Deletion-cascade guard covers `users → taxpayers → incomes` (M1-17 #43, indirect chain) | Guard test lists `incomes` |
| S-10 | Path ids validated before query params; foreign ids answer `404`, foreign idempotency ids `409` without detail | Contract + isolation tests |
| S-11 | `Cache-Control: no-store` on record responses and errors (P0-30 D-6: the architect folds it into the M3-01 contract revision) | Contract test |
| S-12 | Purge (M9-08) runs through the `purge_soft_deleted` definer function only (rows with `deleted_at < now() − 30 days`) | Definer test |

## 6. Feature flag `module_incomes`

- Default OFF; registered by migration (M3-10); key `module_incomes`.
- OFF → every M3 route answers `404 not_found` (server-side boundary), including `defaults`.
  `module_taxpayers` OFF implies M3 OFF (MODULE-taxpayers §6).
- Independent of `module_expenses`: either can be ON alone.
- With the flag OFF the data still exists: `LatestRecordPeriod` still counts incomes (M2 must not
  end a regime under hidden records) and account/taxpayer deletion still cascades. Whether M5
  reads incomes of a hidden module: OQ-8.
- Client (single FlagsProvider): OFF → no Ingresos tab, no "Registrar ingreso" on Home. UX only.
- Suites run with the flag ON and OFF (M3-06); presets `demo_core`, `demo_monetization`,
  `demo_full` include it.

## 7. UX notes (`rules/ui-forms-and-modals.md`)

- **Container: its own page** (`IncomeFormPage`), not a dialog. The form has 11 inputs in total,
  more than 6, so the field-count rule sends it to a page, as PLAN M3-09 says (OQ-1). Progressive
  disclosure keeps the 80 % case short:
  - Visible: Monto (sin IVA) · Tasa de IVA · Fecha de cobro (accrual: Fecha de factura, required,
    then Fecha de cobro right under it, **empty by default** = "Por cobrar", hinted "Decide el mes
    del IVA", M5-30 FU1/D6) · ¿Quién te pagó? (counterparty) · Descripción. The basis is the
    regime of the income being registered (served `income_basis`), so mixed regimes switch it. Régimen appears only when the taxpayer holds more than one
    on that date; Actividad only for 625 with more than one activity.
  - "Retenciones": collapsed; opens **prefilled** when the counterparty allows withholdings.
    **Replaced** by the read-only derived card and questions Q1–Q4 of
    [MODULE-withholdings §2.1, §2.3](MODULE-withholdings.md#21-income-form-nuevo-ingreso-every-taxpayer).
  - "Más detalles": collapsed: IVA amount (editable), the invoice date (cash regimes only),
    Folio fiscal (UUID).
- Single column; label above every field; optional fields marked "(opcional)" (fewer than the
  required ones). Amount: `inputmode="decimal"`, accepts `10000`, `10,000.00`, `$10 000.5`;
  client converts to centavos by string parsing, never floats. Dates accept `dd/mm/aaaa` with or
  without separators. Input text ≥ 16 px, targets ≥ 44 × 44 px.
- Live summary under the amount: "IVA 1,600.00 · Retenciones −2,066.67 · Neto 9,533.33".
- Validation: required-empty errors on submit only; format errors on blur once typed; failed
  submit → error summary at the top, focus moves to it, each item links to its field. Server
  errors map to the same fields.
- Footer: "Guardar" (primary, never disabled; busy state ignores repeat taps), "Guardar y crear
  otro" (secondary), "Cancelar" (text). Cancel with unsaved changes asks first.
- Trial: the page shows "Te quedan 3 de 5 ingresos de prueba" when the access summary is known
  (M7 follow-up); a `402` opens the paywall and keeps the form data.
- List: month switcher, regime filter chip, totals card; rows show date · description (or
  counterparty) · amount, the same order as the form. Delete from the row → do it, then the
  persistent Undo toast (§3 US-I4); several deletes stack into "2 ingresos eliminados ·
  Deshacer". No confirmation dialog (reversible).

## 8. Contract changes for M3-01 (architect)

Version `0.4.0 → 0.5.0` (additive, plus non-frozen schema tightening before the freeze).
1. `IncomeCreate`/`IncomeUpdate`: `additionalProperties: false`; `counterparty_type` required on
   create; `amount_cents` `minimum: 1`; `id` not accepted on update.
2. `IncomeFields`: add `cfdi_uuid` (nullable, pattern `^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$`)
   and `platform_activity` (nullable enum).
3. `Income`: add `warnings: Warning[]` (codes in §2.8) and `pending_collection: boolean`.
4. `deleteIncome`: add `422`; `restoreIncome`: add `402`, `422`.
5. New `getIncomeDefaults` (OQ-7): `{iva_rates_allowed[], suggested_iva_rate, isr_withholding_rate, iva_withholding_rate}`
   as `DecimalRate`s for `(regime, counterparty_type, on)`, from `RateCatalog`.
6. Detail codes of §4 in the operation descriptions; `Cache-Control: no-store` (S-11).

**Frozen as (M3-01, contract 0.5.0; architect notes).** `amount_cents: 0` is refused at the
edge with the schema keyword `minimum` (`must_be_positive` stays the domain's backstop code).
`getIncomeDefaults` also returns `isr_withholding_allowed` / `iva_withholding_allowed` (a
`null` rate alone cannot tell "not allowed" from "allowed, no suggestion") and the resolved
`platform_activity`, and takes an optional `platform_activity` query parameter (625 with several
activities). `warnings` is always present (`[]` on reads). PATCH is merge-patch (`null` clears).
M3/M4 errors declare `Cache-Control: no-store` (optional header, S-11). The list has no
`deleted` filter (the recently-deleted view is out of scope, §9).

Data-model deltas for M3-02 (db-sr): `cfdi_uuid text NULL`, `platform_activity text NULL`,
`booked_on date NOT NULL` (one partial index `(user_id, taxpayer_id, booked_on, id) WHERE
deleted_at IS NULL` replaces the two date indexes), `(user_id, taxpayer_id, cfdi_uuid) WHERE
cfdi_uuid IS NOT NULL AND deleted_at IS NULL`, `regimes.income_basis` (forward migration on the
M2 catalog).

## 9. Out of scope

- **CFDI XML import** or SAT download (PRD NG2; later module), OCR of receipts, bank sync (NG5).
- Foreign currency; credit notes, returns and discounts (negative amounts); linking partial
  collections to one CFDI.
- A "recently deleted" view (Could; Undo lives in the persistent toast, OQ-6); text search;
  recurring incomes; CSV import/export; attachments.
- Optimistic concurrency (ETag) on edit: last write wins in v1.
- A combined multi-taxpayer list (PRD Q14).

## 10. Decisions (coordinator, 2026-10-13) — all DECIDED

| # | Question | Default |
|---|---|---|
| OQ-1 | Form as a drawer (owner's guess: ≤ 6 fields) or a page? | **DECIDED: Page.** 11 inputs total → page by the house rule; only 5 are visible at first |
| OQ-2 | Enter the subtotal or the total with IVA? | **DECIDED: Subtotal** (contract), with a live total; a "capturé el total" toggle is Could |
| OQ-3 | IVA tolerance vs `amount × rate` | **DECIDED: ±100 centavos** (constant `IvaTolerance`) |
| OQ-4 | Allow incomes in informational regimes? | **DECIDED: Yes**, with a notice; they do not change the estimate |
| OQ-5 | Delete/restore on an archived taxpayer | **DECIDED: Refused** (`422 taxpayer_archived`) |
| OQ-6 | Undo persistence | **DECIDED: Toast never auto-dismisses, deletes stack**; recently-deleted view later |
| OQ-7 | New `getIncomeDefaults` endpoint so withholding rates are not hard-coded in the app | **DECIDED: Yes** |
| OQ-8 | Does M5 read incomes when `module_incomes` is OFF? | **DECIDED: Yes** (flags hide screens, not data); presets enable incomes with the engine |
| OQ-9 | Duplicate CFDI UUID | **DECIDED: Warning**, never an error (partial collections) |
| OQ-10 | Does restore (Undo) consume a trial slot? | **DECIDED: no** (overrides the PM default). The slot is consumed once at creation and never freed; restoring the same record does not consume another. No paywall on Undo at 5/5. Backend requirement in §2.7 |
