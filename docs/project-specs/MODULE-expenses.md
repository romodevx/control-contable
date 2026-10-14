# MODULE-expenses — M4 Expenses (`module_expenses`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-13 · MoSCoW: **Must** · Depends on: M2, P0 (access policy), M5-08 (`RateCatalog`)
Sources: PRD §2, §6 (US-E1, US-E2), §7, §8 · MODULES.md M4 · PLAN M4-01..M4-15 ·
research `docs/research/tax-regimes-mx-2026.md` (A.2.3–A.2.5, A.3, B.2, B.2.3, B.3, Part C) ·
ADR-0002 §1–§4 · `docs/architecture/data-model.md` (expenses, expense_categories,
category_regime_rules) · MODULE-taxpayers.md §1, §2.4, §2.7, §2.10 · **[MODULE-incomes.md](MODULE-incomes.md)**.

**Shared with M3, not repeated here** (MODULE-incomes): §2.1 money and currency rules, §2.3
idempotency, §2.4 regime window, §2.5 IVA rate validity and `iva_cents` tolerance, §2.7 delete /
undo / restore (restore never consumes a trial slot: DECIDED, M3 §2.7; archived taxpayer refuses delete and restore), §2.9
list and pagination, §2.10 access table (with `expenses_remaining`), §5 security requirements
(read `incomes` as `expenses`), §6 flag mechanics, §7 form mechanics. Expense-specific rules
follow.

## 1. Scope

The user records, lists, edits, deletes and restores the expenses of the **active taxpayer**,
picks a category from a server catalog, sees whether it usually lowers ISR for their regime,
and may override that. Expenses feed M5 (creditable IVA, deductions, withholdings to remit)
through `LedgerReader` (M4-08).

| Op | Endpoint | Contract |
|---|---|---|
| listExpenseCategories | `GET /v1/taxpayers/{taxpayerId}/expense-categories?regime=&on=` | **replaces** `GET /v1/expense-categories?regime=` (OQ-E1) |
| listExpenses | `GET /v1/taxpayers/{taxpayerId}/expenses?period=YYYY-MM&regime=&cursor=&limit=` | exists |
| createExpense | `POST /v1/taxpayers/{taxpayerId}/expenses` | exists; changes in §8 |
| getExpense / updateExpense | `GET / PATCH …/expenses/{expenseId}` | exists; changes in §8 |
| deleteExpense | `DELETE …/expenses/{expenseId}` | exists; add `422` |
| restoreExpense | `POST …/expenses/{expenseId}/restore` | exists; add `402`, `422` |

Ports: the same as M3 (`TaxpayerRegimeChecker`, `TaxpayerProfileReader`, `access`
`ReserveRecord(RecordExpense)`, `RateCatalog`), plus `RateCatalog`'s cash-payment limit (§2.5).
M4-08 extends the `LatestRecordPeriod` adapter M3-07 wired with expense dates.

## 2. Domain rules

### 2.1 Fields

| Field | Type | Required | Rule |
|---|---|---|---|
| `id` | uuid | client SHOULD send | Idempotency (M3 §2.3) |
| `regime_code` | `RegimeCode` | yes | Held on `paid_on` (M3 §2.4) |
| `category_code` | string | yes | Exists in `expense_categories` (`400 unknown_category`) |
| `amount_cents` | int64 centavos | yes | ≥ 1, subtotal **before IVA**, MXN |
| `iva_rate` | enum | yes | Valid on `paid_on` (M3 §2.5) |
| `iva_cents` | int64 centavos | yes | M3 §2.5 consistency (IVA paid, creditable candidate) |
| `payment_method` | `cash \| transfer \| card_debit \| card_credit \| check \| e_wallet \| other` | yes | Drives §2.5 |
| `paid_on` | date | yes | The only date: deductions and creditable IVA are on a paid basis for every regime in scope (research A.2.3, A.3, B.2.2). `2022-01-01 ≤ paid_on ≤ today` (Mexico City) |
| `has_cfdi` | boolean | yes, default `true` in the UI | `false` → warning `no_cfdi`, default deductible `false` |
| `cfdi_uuid` | string | no (**new**) | Same format as M3; requires `has_cfdi = true` (`400 has_cfdi_mismatch`); duplicate → warning `duplicate_cfdi_uuid` |
| `deductible` | boolean \| null | no (**new**) | `null`/omitted = use the default (§2.4); a boolean = user override |
| `isr_withheld_cents` | int64 centavos | no, default 0 | Withheld **by the taxpayer from the supplier** (§2.6) |
| `iva_withheld_cents` | int64 centavos | no, default 0 | Same; `≤ iva_cents` |
| `description` | string ≤ 200 | no | Trimmed, NFC; never logged |

The booking date is `paid_on` (list period, `booked_on`). The response adds `deductible`
(effective boolean), `deductible_source` (`category_default | user`), `warnings[]` and the
mandatory `legend`.

### 2.2 Category catalog (global, seeded in M4-02) ⚠️ CPA review before launch

`expense_categories (code, name, sort_order)` and `category_regime_rules (category_code,
regime_code, person_type, deductibility, default_deductible, note, legal_basis)` are data,
written by migrations only. `deductibility ∈ {deductible, partially_deductible, not_deductible,
isr_not_applicable}` is the hint shown to the user; `default_deductible` is the boolean the
record starts with. They differ on purpose for partial cases (conservative default).

Columns: **Business** = 612, 621, 625 (non-definitive), 601, 620, 622, 626 PM, 603 (same row
values, one rule row per regime) · **606** = arrendamiento with `deduction_mode = actual` ·
**626 PF** = RESICO PF. D = deductible, P = partially, N = not deductible, X = ISR not applicable.

| # | Code | Name (es-MX) | Business | 606 | 626 PF | Default deductible (business / 606) | Legal basis (article level) | Note shown |
|---|---|---|---|---|---|---|---|---|
| 1 | `mercancia` | Mercancía y materias primas | D | N | X | true / false | LISR 25, 103, 208 | |
| 2 | `servicios_profesionales` | Honorarios y servicios profesionales | D | D | X | true / true | LISR 25, 103, 115 | Si eres persona moral, revisa las retenciones |
| 3 | `renta_local` | Renta de local, oficina o bodega | D | N | X | true / false | LISR 25, 103 | Si eres persona moral, revisa las retenciones |
| 4 | `servicios_basicos` | Luz, agua, teléfono e internet | D | P | X | true / true | LISR 103, 115 | Arrendamiento: sólo el agua del inmueble rentado |
| 5 | `software` | Software y suscripciones | D | N | X | true / false | LISR 103 | |
| 6 | `papeleria` | Papelería y artículos de oficina | D | N | X | true / false | LISR 103 | |
| 7 | `publicidad` | Publicidad y mercadotecnia | D | N | X | true / false | LISR 103 | |
| 8 | `combustible` | Combustibles | D | N | X | true / false | LISR 27 | Sólo deducible si lo pagas con tarjeta, transferencia o monedero, sin importar el monto |
| 9 | `vehiculo_mantenimiento` | Mantenimiento y refacciones de vehículo | P | N | X | true / false | LISR 28, 36 | Tiene topes de deducción |
| 10 | `viaticos` | Viáticos y gastos de viaje | P | N | X | true / false | LISR 28 | Tiene topes y requisitos |
| 11 | `restaurantes` | Consumo en restaurantes | P | N | X | false / false | LISR 28 | Sólo una parte es deducible |
| 12 | `equipo_computo` | Equipo de cómputo (inversión) | P | N | X | false / false | LISR 31, 34 | Se deduce por depreciación anual, no completo en el mes |
| 13 | `mobiliario` | Mobiliario y equipo (inversión) | P | N | X | false / false | LISR 31, 34 | Igual que el anterior |
| 14 | `vehiculo_compra` | Compra de automóvil (inversión) | P | N | X | false / false | LISR 34, 36 | Depreciación con tope |
| 15 | `construcciones` | Construcciones y mejoras al inmueble | P | P | X | false / false | LISR 34, 115 | Se deduce por depreciación |
| 16 | `sueldos` | Sueldos y salarios pagados | P | D | X | true / true | LISR 28, 115 | Parte de las prestaciones exentas no es deducible |
| 17 | `cuotas_imss` | Cuotas IMSS e INFONAVIT patronales | D | D | X | true / true | LISR 25, 103, 115 | |
| 18 | `intereses` | Intereses pagados | D | D | X | true / true | LISR 25, 103, 115 | Sólo los intereses reales |
| 19 | `seguros` | Seguros y fianzas | D | D | X | true / true | LISR 25, 103, 115 | |
| 20 | `mantenimiento` | Mantenimiento y reparaciones | D | D | X | true / true | LISR 103, 115 | |
| 21 | `predial` | Impuesto predial | D | D | X | true / true | LISR 103, 115 | En arrendamiento se suma aun con la deducción ciega |
| 22 | `capacitacion` | Capacitación y cursos | D | N | X | true / false | LISR 103 | |
| 23 | `comisiones` | Comisiones bancarias y de plataformas | D | N | X | true / false | LISR 103 | |
| 24 | `fletes` | Fletes y paquetería | D | N | X | true / false | LISR 103 | Si eres persona moral, revisa la retención de IVA (4 %) |
| 25 | `gastos_personales` | Gastos personales (médicos, colegiaturas, donativos) | N | N | X | false / false | LISR 151 | Se aplican en tu declaración anual, no en el pago mensual |
| 26 | `otros` | Otros gastos | P | N | X | false / false | LISR 27 | Verifica que cumpla los requisitos |

- 601 (and the other coeficiente regimes) rows add the note "En los pagos provisionales del 601 el
  ISR se calcula con el coeficiente de utilidad; este gasto cuenta para tu IVA y tu declaración
  anual".
- Informational regimes get `isr_not_applicable`, default `false`, note "Este régimen es
  informativo".
- Every row is ⚠️ until the CPA signs it off (add to `docs/reviews/CPA-REVIEW.md` in M4-02); the
  basis column is article-level on purpose (fractions are not verified by the research).
- A catalog test asserts: one rule row per (category, catalog regime, person type), non-empty
  `name`/`note` where shown, `default_deductible = false` whenever `deductibility ∈
  {not_deductible, isr_not_applicable}`.

### 2.3 Options change the effective hint (domain, not catalog)

`DeductibilityResolver(category, regime, person_type, options)` — pure, table-driven TDD:
- 606 with `deduction_mode = blind_35`: every category becomes `isr_not_applicable` except
  `predial` (LISR 115 ¶2: predial is added to the 35 %).
- 625 with `definitive_option = true`: every category becomes `isr_not_applicable` (113-B).
- 626 PF: always `isr_not_applicable`; the form says "En RESICO tus gastos no reducen tu ISR,
  pero el IVA que pagaste sí puede acreditarse" (PRD US-E1).

### 2.4 Effective `deductible`

```
default   = rule.default_deductible(resolved by §2.3)
            AND has_cfdi
            AND NOT cash_over_limit            (§2.5)
deductible = override if the user sent a boolean, else default
```

- `deductible_source = user` when overridden; `null` in a PATCH clears the override.
- On edit without an override, the default is **recomputed** when category, regime, payment
  method, amount, `has_cfdi` or `paid_on` change; an override is kept as is.
- `deductible` means "the user claims it lowers ISR"; M5 adds the `deductibility_claimed`
  disclaimer (research Part C). Creditable IVA is M5's rule (LIVA 5: CFDI, IVA shown, paid,
  ISR-deductible — or would be, for RESICO PF); M4 stores the facts it needs.
- Overriding to `true` a category whose hint is `not_deductible`/`isr_not_applicable` is
  allowed with the warning `category_not_deductible` (or `not_isr_deductible_in_regime`).

### 2.5 Cash payments ⚠️ (not in the research; CPA)

- `cash_over_limit` = `payment_method = cash` AND (`amount_cents + iva_cents` > the limit, or the
  category is `combustible`). Limit: **$2,000.00** (LISR 27 fr. III), served by `RateCatalog` as
  dated parameter data (`deduction.cash_payment_limit`, ADR-0002 §3), never hard-coded.
- Effect: warning `cash_over_threshold` ("Los pagos en efectivo de más de $2,000 no son
  deducibles") and default deductible `false`. Never an error.

### 2.6 Withholdings the taxpayer made (retenciones a proveedores)

- Allowed only for **persona moral** taxpayers (research B.2.3, LIVA 1-A). A física sending a
  value > 0 → `422 withholding_not_applicable`. The form shows the section only for morales,
  collapsed, opening by default on `servicios_profesionales`, `renta_local` and `fletes`.
- Suggestions (UX only, from `RateCatalog`): ISR 10 % and IVA 2/3 on fees and rent; IVA 4 % of
  the price on freight. M5 adds them to `withholdings_to_remit`, never netted (ADR-0002 §1).
- Bounds: `isr_withheld_cents ≤ amount_cents`, `iva_withheld_cents ≤ iva_cents`.

### 2.7 Mandatory legend (owner rule)

- The exact text **"Esto es un aproximado. No todos los gastos aplican para tu giro
  comercial."** appears on the **expense form**, directly under the category and its
  deductibility hint, always visible (not inside a collapsed section), and on the expense list
  header. It is the `Legend` const in `ExpenseCategoryList.legend` and `Expense.legend`; the app
  renders the server field, never a copy of its own.
- A widget test fails if the form renders without it; a contract test asserts the const.

### 2.8 Warnings (`Expense.warnings`, codes append-only)

`cash_over_threshold`, `no_cfdi`, `not_isr_deductible_in_regime`, `category_not_deductible`
(existing) + **`duplicate_cfdi_uuid`** (new). Each carries `field` (contract change: align
`ExpenseWarning` with the generic `Warning` shape, OQ-E5).

## 3. User stories (Given / When / Then)

### US-E1 Record an expense
- **Given** a física with 612 and `module_expenses` ON, **when** I open "Nuevo gasto", **then**
  the regime is preselected, date today, IVA 16 %, payment method my last one (first time:
  tarjeta de débito), "¿Tienes factura?" = Sí, and the legend shows under the category.
- **When** I pick "Software y suscripciones", **then** the hint reads "Normalmente deducible
  para tu régimen" with its basis, and the deductible switch (under "Más detalles") is on.
- **When** I save 500.00 + IVA 80.00 paid by card, **then** `201`, `deductible: true`,
  `deductible_source: category_default`, `warnings: []`, `legend` = the const.
- **Given** RESICO PF, **when** I pick any category, **then** the hint is "No reduce tu ISR en
  RESICO" and the record saves with `deductible: false`; IVA is still stored.
- **Given** 606 with `blind_35`, **when** I pick `mantenimiento`, **then** the hint is
  "Con la deducción ciega del 35 % este gasto no cuenta"; **with** `predial`, **then** deductible.
- **When** I pay 2,320.00 in cash for papelería, **then** `201`, `deductible: false`, warning
  `cash_over_threshold`; **when** I pay 300.00 of `combustible` in cash, **then** the same.
- **When** `has_cfdi: false`, **then** `201`, `deductible: false`, warning `no_cfdi`.
- **When** I turn the switch on for `restaurantes` (`partially_deductible`), **then**
  `deductible: true`, `deductible_source: user`, no warning; **when** I do it for
  `gastos_personales` (`not_deductible`), **then** the same plus warning `category_not_deductible`.
- **When** I PATCH `deductible: null`, **then** the override is cleared and the default applies.
- **When** `cfdi_uuid` is set with `has_cfdi: false`, **then** `400` (`has_cfdi_mismatch`).
- **When** `category_code: "viajes-luna"`, **then** `400` (`category_code` / `unknown_category`).
- **Given** a moral with 601, **when** I record fees of 10,000.00 + IVA 1,600.00 with ISR
  withheld 1,000.00 and IVA withheld 1,066.67, **then** `201`; **given** a física, **then**
  `422` (`withholding_not_applicable`).

### US-E1b Categories for my regime
- **When** the form loads, **then** `GET /v1/taxpayers/{id}/expense-categories?regime=612&on=2026-09-12`
  returns every category with the hint, `default_deductible`, note and basis already resolved for
  my regime **and options**, the allowed IVA rates on that date, and the legend.
- **Given** a moral with 626, **then** the hints are those of RESICO PM (deductible), not RESICO PF.

### US-E2 List, filter, edit, delete (same AC as US-I2..US-I4)
- **Given** expenses in September, **when** I open Gastos for 2026-09, **then** items newest
  first, month totals, filter by regime; the legend sits in the list header.
- **When** I change the category of a non-overridden expense, **then** `deductible` is
  recomputed; **when** it was overridden, **then** it stays.
- **When** I delete and tap Deshacer, **then** restore → `200` with no trial slot consumed, even at 5/5 (M3 §2.7, DECIDED); `expenses_created` is unchanged.

### US-E3 Trial cap and read-only
- **Given** a trial user with 5 expenses created and 0 incomes, **when** they create a 6th
  expense, **then** `402` with `details[0].field = expenses_remaining`; **when** they create an
  income, **then** `201` (independent caps).
- **Given** 4/5 and 3 concurrent creates, **then** exactly one `201`.
- Read-only: list/get/categories/delete ✓; create/edit/restore `402`.

### US-E4 Isolation, archive, flag
- Foreign ids → `404`; archived taxpayer → reads ✓, writes `422 taxpayer_archived`.
- `module_expenses` OFF → every M4 route incl. categories `404`, even with `module_incomes` ON.

## 4. Errors → ErrorCode

The M3 table applies (field `expenses_remaining` for the cap), plus:

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Unknown category | 400 | `validation_failed` | `unknown_category` |
| `cfdi_uuid` with `has_cfdi = false` | 400 | `validation_failed` | `has_cfdi_mismatch` |
| Missing `paid_on` / `payment_method` / `has_cfdi` | 400 | `validation_failed` | `required` |
| Withholding by a física, or above its base | 422 | `business_rule_violation` | `withholding_not_applicable`, `withholding_exceeds_base` |

No new `ErrorCode` values. Non-deductibility is never an error: it is a warning plus
`deductible: false`.

## 5. Security requirements (gate M4-14)

M3 §5 S-1..S-12 with `expenses` for `incomes`, plus:

| # | Requirement | Test |
|---|---|---|
| SE-1 | `expense_categories`, `category_regime_rules`: global, ENABLE RLS with one `SELECT TO cc_app` policy, no write grant (migrations only); FK `expenses.category_code → expense_categories`; FK `(regime_code, person_type) → regimes` on the rules | Catalog sweep; write refused as `cc_app` |
| SE-2 | DB CHECKs: `cfdi_uuid IS NULL OR has_cfdi`, `payment_method` enum, amounts as M3 | CHECK tests |
| SE-3 | The categories endpoint is taxpayer-scoped and returns `404` for a foreign taxpayer (it reads the caller's regime options) | Isolation test |

## 6. Feature flag `module_expenses`

As M3 §6 with key `module_expenses` (registered in M4-11): OFF → all M4 routes `404`, the
categories endpoint included; implied OFF when `module_taxpayers` is OFF; independent of
`module_incomes`; the hidden data still counts for `LatestRecordPeriod`. Client OFF → no Gastos
tab. Presets `demo_core`, `demo_monetization`, `demo_full` include it.

## 7. UX notes

- **Container: its own page** (`ExpenseFormPage`, reusing the M3 form scaffolding, PLAN M4-10):
  13 inputs in total, so the field-count rule sends it to a page (OQ-E2).
  - Visible: Categoría (searchable list, most-used first) → hint + **legend** · Monto (sin IVA) ·
    Tasa de IVA · Fecha de pago · Forma de pago · ¿Tienes factura? (Sí/No). Régimen only with
    more than one regime on that date.
  - "Más detalles" (collapsed): IVA amount, "¿Es deducible?" switch with "Usar el valor sugerido"
    reset, Folio fiscal (UUID), Descripción.
  - "Retenciones a tu proveedor": morales only, collapsed except for fees, rent and freight.
- Warnings render inline next to their field after save and on blur for the cash rule ("Los
  pagos en efectivo de más de $2,000 no son deducibles"), never as blocking errors.
- "Guardar y crear otro" keeps regime, date, payment method and "¿Tienes factura?"; clears
  category, amounts, UUID, description, override; new `id`; focus on Categoría.
- All other mechanics (labels, validation timing, error summary, busy state, Undo toast, trial
  remaining "Te quedan N de 5 gastos de prueba") as M3 §7.

## 8. Contract changes for M4-01 (architect)

1. Replace `GET /v1/expense-categories?regime=` with
   `GET /v1/taxpayers/{taxpayerId}/expense-categories?regime=&on=` (OQ-E1). Response:
   `{items: [{code, name, deductibility, default_deductible, note, legal_basis}], iva_rates_allowed[], cash_payment_limit_cents, legend}`
   resolved for the taxpayer's person type and options.
2. `ExpenseCreate`/`ExpenseUpdate`: `additionalProperties: false`; `amount_cents` `minimum: 1`;
   add `cfdi_uuid` (nullable, M3 pattern) and `deductible` (boolean | null).
3. `Expense`: add `deductible` (boolean, required), `deductible_source` (enum); `warnings` items
   gain `field` and the code `duplicate_cfdi_uuid` (OQ-E5).
4. `deleteExpense`: add `422`; `restoreExpense`: add `402`, `422`.
5. Detail codes of §4 and M3 §4 in the descriptions; `Cache-Control: no-store` as M3.

**Frozen as (M4-01, contract 0.5.0; architect notes).** `ExpensePage` also carries `legend`
(the list header renders the server field even on an empty month). `category_code` has no
pattern, so an unknown code is always `unknown_category`. `ExpenseCategory` is flat (no per-regime
`hints[]`): the endpoint is resolved for one regime.

Data-model deltas for M4-02/M4-03 (db-sr): `expenses.cfdi_uuid text NULL`, `deductible boolean
NOT NULL`, `deductible_overridden boolean NOT NULL DEFAULT false`; `category_regime_rules.default_deductible
boolean NOT NULL`; partial indexes as M3 (`paid_on` is the booking date, so no `booked_on`
column is needed).

## 9. Out of scope

Everything in M3 §9 (CFDI XML import later, OCR, foreign currency, credit notes, recently-deleted
view, ETag, search), plus: depreciation schedules for investments (research Part C), IVA
proportional credit for mixed activity (the 612/601 `iva_fully_taxable` option covers v1),
per-item deduction caps (cars, restaurants, viáticos), personal deductions (annual), DIOT (NG3),
user-defined categories.

## 10. Decisions (coordinator, 2026-10-13) — all DECIDED

| # | Question | Default |
|---|---|---|
| OQ-E1 | Taxpayer-scoped categories endpoint instead of the global one? The global one cannot tell 626 PF from 626 PM nor apply 606/625 options | **DECIDED: Yes, replace it** (not frozen yet) |
| OQ-E2 | Form as drawer or page? | **DECIDED: Page** (13 inputs; 6 visible at first) |
| OQ-E3 | Category seed and per-regime defaults (§2.2) | **DECIDED: Ship the seed, ⚠️, CPA review before launch**; investments, restaurants, personal and "otros" default to non-deductible (conservative estimate) |
| OQ-E4 | Cash limit $2,000 on the total with IVA, and fuel always electronic (LISR 27-III) — not in the research | **DECIDED: As stated, as dated parameter data**, CPA to confirm |
| OQ-E5 | Unify `ExpenseWarning` with the generic `Warning` (`field`, append-only codes) | **DECIDED: Yes**, in M4-01 before the freeze |
| OQ-E6 | Default payment method on first use | **DECIDED: Tarjeta de débito**; afterwards the last one used |
