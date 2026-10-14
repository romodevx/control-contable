# MODULE-annual — Annual estimate (`module_annual`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-15 · MoSCoW: **Should** · Depends on: M2
(taxpayers, regimes catalog, `TaxpayerProfileReader`), M5 (`Estimator`, `LedgerReader` aggregates,
`taxpayer_tax_inputs`, params loader, `Ledger`/`Line`, rounding), P0 (access policy, rate limits)
Sources: owner decisions 2026-10-15 (below) · research
[`docs/research/annual-return-2026.md`](../research/annual-return-2026.md) (Parts A–E; cited as **AR §x**) ·
[`docs/research/tax-regimes-mx-2026.md`](../research/tax-regimes-mx-2026.md) (monthly doc) ·
[`docs/research/withholdings-matrix-2026.md`](../research/withholdings-matrix-2026.md) ·
[MODULE-tax-engine.md](MODULE-tax-engine.md) §2.2–§2.11 · [MODULE-history.md](MODULE-history.md) §10 ·
MODULE-withholdings.md (being written in parallel; referenced, not duplicated) · ADR-0002, ADR-0007.

**Owner decisions (2026-10-15), binding on this spec:**

1. Estimate the **declaración anual** for **every regime obligated to file** it, and **always** label it
   an estimate: **"Esto es un aproximado; tu contador te dará las cifras reales."**
2. Capture **personas físicas' personal deductions by category**, with the legal caps applied
   automatically.
3. Capture **salary income + ISR withheld** from the constancia (regime 605), used **only** by the
   annual estimate.

## 1. Scope

For the **active taxpayer** and one **fiscal year** (2026 in v1, filed in 2027), the module answers:
**"Roughly how much will I owe, or get back, in my annual return, and by when?"**

| Covered | How (AR Part E) |
|---|---|
| PF **612, 606, 625 non-definitive, 605** and any combination | **Estimated**, aggregated in **one** art. 152 section (AR §B.1), with personal deductions (AR §B.2) |
| PM **601** | **Estimated** per regime (AR §B.5) + PTU base and suggested coeficiente for next year (informative) |
| PM **RESICO (626 PM)** | **Estimated** per regime (AR §B.6) |
| PF **RESICO (626 PF)** | **Informative check only** (113-F reconciliation; rule 3.13.7 makes monthly payments definitive) |
| 625 definitive, 615, 616, 610 | "No annual for this income" (explanation) |
| 603, 607, 608, 611, 614, 620, 621, 622, 623, 624 | Explanation + deadline only, "tu contador lo calcula" |
| IVA | No annual IVA return (LIVA 5-D): a fixed explanation, no figure |

It also stores the **annual-only inputs** (§2.6): personal deductions, salary constancias, provisional
payments actually filed, and a few per-year facts. It never files, never stores an estimate (derived on
every read, like M5, ADR-0002 §5) and does **not** change any monthly figure.

| Op | Endpoint (provisional; architect freezes in ANN-02) |
|---|---|
| getAnnualEstimate | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}` |
| getAnnualBreakdown | `GET /v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown?section=` |
| getAnnualInputs / putAnnualInputs | `GET\|PUT /v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}` |
| listSalaryStatements / createSalaryStatement | `GET\|POST /v1/taxpayers/{taxpayerId}/salary-statements?fiscal_year=` |
| updateSalaryStatement / deleteSalaryStatement | `PATCH\|DELETE /v1/taxpayers/{taxpayerId}/salary-statements/{statementId}` |
| listPersonalDeductions / createPersonalDeduction | `GET\|POST /v1/taxpayers/{taxpayerId}/personal-deductions?fiscal_year=&category=` |
| update / delete / restore personal deduction | `PATCH\|DELETE /v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}`, `POST …/{deductionId}/restore` |
| getFiledPayments / putFiledPayments | `GET\|PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}` |

Ports consumed (consumer-side interfaces, no import of M3/M4/M5 internals beyond published ports and the
fiscal kernel, ADR-0007):
- `LedgerReader` (M3-07, M4-08): the 12 `MonthSums` per regime of the fiscal year (income by basis,
  deductible expenses paid, `isr_withheld_cents` recorded on incomes). One aggregate pass.
- `Estimator` (M5-22, range form from M6-02): the app's **monthly estimates** for the year, used only as
  the **default** provisional payments (§2.5).
- `TaxInputReader` (M5): `ptu_paid`, `prior_year_losses`, `interest_withholding`, `ytd_before_app_*`,
  `prior_provisional_payments`, and the new `annual_depreciation` key (§2.6).
- `TaxpayerProfileReader`, `RegimeCatalogReader` (M2), `Clock`, `HolidayCalendar` (M5 §2.10),
  `access.Policy` (P0-16) for writes.

Withholdings **received** are read as **recorded on the incomes** (`isr_withheld_cents`). How they get
there (typed, or derived from emisor × receptor) is MODULE-withholdings' job; this module only sums them.

## 2. Domain rules

### 2.1 What one annual estimate contains

```
AnnualEstimate(taxpayer, fiscal_year)
├── legend                       const (§2.8), always
├── disclaimers[]                §2.8, deterministic order
├── year_status                  closed | in_progress (+ data_through: YYYY-MM)
├── deadline                     PF 2027-04-30 / PM 2027-03-31 (603: 2027-02-15), §2.9
├── sections[]                   §2.2, fixed order
│   ├── kind                     pf_art152 | pf_resico_check | pm_601 | pm_resico | explanation
│   ├── regimes[]                regime codes the section covers
│   ├── status, missing_inputs[] §2.3
│   ├── obligation               obligated | optional | not_obligated | relieved   (§2.4)
│   ├── tax: AnnualTax | null    base, isr, credits{payments, withholdings, salary_isr}, balance
│   │                            balance = {kind: a_cargo | a_favor | cero, amount_cents}
│   ├── pm_extras | null         601: ptu_base, ptu_to_distribute, suggested_cu_next_year; RESICO PM: ptu_base
│   ├── explanation | null       text (explanation sections and informative notes)
│   ├── cpa_review_pending, cpa_topics[]   §2.10
│   └── parameter_set {fiscal_year, version}
├── informative_income_over_500k   bool (PF, AR §A.1 "> $500,000")
└── engine_version
```

The **headline** the app shows is the `pf_art152` section's balance (PF) or the sum of PM sections'
balances (a PM holds one of 601 / RESICO PM at a time; if both appear in one year, each is shown on its
own, never netted). Informative sections never feed the headline.

### 2.2 Sections: how regimes are grouped

| Section | Who | Contents | Algorithm |
|---|---|---|---|
| `pf_art152` | PF holding any of **605, 606, 612, 625 non-definitive** in the year | **One** aggregated computation: Σ chapter results − personal deductions → annual tariff → − credits | AR §B.1 steps 1–11 |
| `pf_resico_check` | PF holding **626** in the year | 113-F informative reconciliation; **never** a payable/refundable balance | AR §B.3 |
| `pm_601` | PM 601 | art. 9 annual + PTU + suggested CU | AR §B.5 |
| `pm_resico` | PM 626 | art. 9 on cash flow + PTU base | AR §B.6 |
| `explanation` | one per other regime held (table in §1) | text + deadline, no figure | AR Part A |

Rules:
- **Per-chapter netting (PF)**: each chapter contributes its own result; a **612 loss stays in 612**
  (contributes 0 and is reported as "pérdida del ejercicio, amortizable en años siguientes"), it never
  reduces salary or rent (LISR 109). Same for 606 (contributes `max(0, …)`).
- **625 non-definitive** is combined with 612 as business income (platform income − deductions), with
  CPA topic `platform_nondefinitive_annual` (AR §A.1 ⚠️).
- **RESICO PF + salary**: salary goes to `pf_art152` alone (rule 3.13.8); RESICO income never enters
  art. 152 and is not reduced by personal deductions (113-F has no deductions).
- **625 definitive option** (M5 tax input `definitive_option = true`): excluded from art. 152; an
  `explanation` section says "tus retenciones de plataforma fueron definitivas".
- A regime is in a section when its window `[start_period, end_period]` intersects the fiscal year
  (archived regimes included for their active months, MODULE-taxpayers §2.7). Only the months in the
  window count.
- 614 (intereses) is **explanation-only** in v1 (AR Part C #14 optional); its income does not enter
  art. 152 (OQ-A6).

### 2.3 Status per section

| Status | When | Figure | UI |
|---|---|---|---|
| `complete` | Every required input present (§2.6) | Yes | The balance |
| `incomplete` | A required input is missing; computed with the documented fallback (column "Missing →" in §2.6) | Yes, marked "Incompleto" | Amber chip + CTA "Completa tus datos" to the first missing input |
| `informational` | `pf_resico_check`, `explanation`, or a section the law does not tax annually | `pf_resico_check`: informative difference; else none | Text; no balance styling as payable/refundable |

- Precedence `informational` > `incomplete` > `complete`; `missing_inputs` is filled whatever the status.
- **Year in progress is not a status**: when `fiscal_year` is the current year, every section is computed
  with records through the last closed month (`data_through`), carries `year_status = in_progress` and
  the `year_in_progress` disclaimer. December figures are never extrapolated (OQ-A3).
- **Default provisional payments are not "missing"**: months without a confirmed filed payment use the
  app's monthly estimate (§2.5) and raise the `payments_estimated` disclaimer, not `incomplete`, unless
  that month's monthly estimate was itself `incomplete`.

### 2.4 Obligation (PF)

The `pf_art152` section carries `obligation`:

| Value | When (AR §A.1) |
|---|---|
| `obligated` | Holds 606, 612 or 625 non-definitive with income in the year; **or** salary with any of: other accumulable income; ≥ 2 constancias marked `simultaneous`; any constancia marked `left_before_year_end`; any constancia marked `payer_not_withholding`; `told_employer_will_file = true`; salary income (accumulable + exempt) > **$400,000** ⚠️ |
| `optional` | Salary only (± interest within LISR 150 ¶2 limits) and none of the above: "No estás obligado, pero puedes presentarla para deducir tus gastos personales" (a refund is the usual reason) |
| `not_obligated` | No income in any art. 152 chapter |

`pf_resico_check` carries `relieved` (rule 3.13.7). PM sections: always `obligated`. The salary relief rule
3.17.11 is written for ejercicio 2025 only; until RMF 2027 is out the app does **not** apply it
(`rmf_2027_pending` disclaimer).

### 2.5 Credits and provisional payments (default = the app's own estimates)

Credits in `pf_art152` (AR §B.1 step 10) and PM sections (§B.5 step 6, §B.6):

| Credit | Source |
|---|---|
| Provisional ISR **paid** per month, per regime (612, 606, 625 self-paid, 601, RESICO PM) | Precedence per month: **(1)** `annual_filed_payments` value the user confirmed or typed → **(2)** M5 tax input `prior_provisional_payments` for that month (already "what you paid") → **(3)** the app's monthly estimate for that regime and month, `max(0, isr)` |
| ISR withheld **from** the taxpayer: 10 % fees/rent (PM payer), platform withholdings | Σ `isr_withheld_cents` on the year's incomes of the regimes in the section |
| ISR withheld on salary | Σ `isr_withheld_cents` of the constancias |
| Interest withholding (PM) | M5 tax input `interest_withholding` |
| RESICO PF 1.25 % and monthly payments | Only inside `pf_resico_check` (informative) |

- Each payment line has `source`: `filed` (1), `tax_input` (2) or `estimated_by_app` (3). Any (3) →
  disclaimer `payments_estimated`: "Usamos lo que la app estimó para {n} meses; confírmalo con tus acuses
  para un resultado más preciso."
- **Confirming** (UX): the filed-payments screen shows the 12 months prefilled with the app estimate; "Confirmar
  todos" stores those values as filed (source 1); editing one month stores the typed value.
- A **saldo a favor** that depends on any `estimated_by_app` payment carries the note "Solo se devuelve lo
  efectivamente pagado o retenido" (LISR 152).
- Saldo a favor ≥ **$10,000**: info note "Para solicitar la devolución necesitas e.firma" (AR §A.2 dates).

### 2.6 Inputs

**New tables (all RLS ENABLE + FORCE, keyed on `user_id`, FK to `taxpayers`, cascade on deletion):**

`personal_deductions` (PF only; soft delete + restore, undo per ui-forms rule):

| Field | Type | Notes |
|---|---|---|
| `category` | enum | `medical`, `medical_disability`, `funeral`, `donation`, `mortgage_real_interest`, `retirement_contribution`, `health_insurance`, `school_transport`, `local_payroll_tax`, `tuition` (AR §B.2 fr. I–VIII + decree) |
| `amount_cents` | int64 > 0 | As on the CFDI / constancia (tuition: tuition only, no enrolment) |
| `paid_on` | date in the fiscal year | Annual constancias (mortgage, retirement): default 31-dic |
| `payment_method` | `electronic` \| `cash` | `cash` ⇒ **not deductible** for `medical`, `medical_disability`, `school_transport`, `tuition` (excluded with a visible line) |
| `beneficiary` | `self` \| `spouse` \| `ascendant` \| `descendant` | Not `self` → user attests the dependant's income < 1 UMA anual (checkbox, not verified) ⚠️ |
| `donee_type` | `authorized` \| `government` | `donation` only |
| `student_label`, `tuition_level`, `scholarship_cents` | text ≤ 40 / enum `preescolar`, `primaria`, `secundaria`, `profesional_tecnico`, `bachillerato` / int64 ≥ 0 | `tuition` only; the cap is per student |
| `cfdi_uuid` | UUID, optional | Format-validated only; never sent to the SAT |

`salary_statements` (PF; one row per employer constancia): `employer_label` (text ≤ 60, optional; no
employer RFC in v1), `accumulable_cents`, `exempt_cents`, `isr_withheld_cents` (≥ 0), `simultaneous`,
`left_before_year_end`, `payer_not_withholding` (booleans, default false). Figures are the constancia's
totals; the app never recomputes art. 93 exemptions or the employment subsidy (AR Part E CPA #5).

`annual_filed_payments`: `(taxpayer_id, regime_code, fiscal_year, month)` unique, `amount_cents ≥ 0`.
PUT replaces the year's set for that regime; absent month = not confirmed.

`annual_inputs` (one row per taxpayer + fiscal year): `prior_year_accumulable_income_cents` (nullable;
2025 acumulables before personal deductions, for the 7 % donation cap), `other_exempt_income_cents`
(default 0; for the 15 % cap's "total income"), `told_employer_will_file` (bool).

**Reused M5 tax inputs** (per regime + fiscal year, `taxpayer_tax_inputs`): `ptu_paid`,
`prior_year_losses` (typed, already inflation-updated; the app does not update ⚠️), `interest_withholding`,
`ytd_before_app_*`, `prior_provisional_payments`. **New key** `annual_depreciation` (money, optional) for
612 PF, 601 PM, 626 PM: the year's depreciation from the CPA's fixed-asset schedule ⚠️ typed only, no
asset register (forward catalog migration on `regimes.required_inputs`).

**Required inputs and fallbacks:**

| Section | Required | Missing → |
|---|---|---|
| `pf_art152` | A salary constancia for each year the taxpayer holds 605 | `incomplete`, salary = 0, missing `salary_statement` |
| `pf_art152` | `prior_year_accumulable_income` **when** any `donation` exists | `incomplete`; donations **excluded** (cap unverifiable), missing key named |
| `pm_601` | none beyond M5's (`coeficiente_utilidad` only matters for default payments) | — |
| any | A month whose default payment comes from an `incomplete` monthly estimate | `incomplete`, missing `filed_payment:<regime>:<YYYY-MM>` |

**Annual inflation adjustment (601) is out of scope** (owner): no input; disclaimer
`inflation_adjustment_excluded` on `pm_601` and `pm_resico`.

### 2.7 Algorithms (pure, `(Input, Params)` only, like M5 §2.11)

**PF art. 152 (`pf_art152`)**, AR §B.1:

1. Salary = Σ `accumulable_cents` of constancias.
2. 612 = income collected − deductible expenses paid (+ `ytd_before_app_*`) − `annual_depreciation` −
   `ptu_paid` − `prior_year_losses`; `max(0, …)`. 625 non-definitive adds platform income − deductions here.
3. 606 = rent collected − (blind 35 % + predial, or actual deductions, per the regime option); `max(0, …)`.
4. Ingresos acumulables = 1 + 2 + 3.
5. Personal deductions (below). 6. Base = 4 − 5; `≤ 0 → ISR 0`.
7. ISR = annual tariff (Anexo 8 C.II, `isr.annual_152` in params; identical to the December art. 106
   table, asserted by test) — `cuota fija + (base − límite inferior) × %`, binary search (M5 §2.4).
8. Credits (§2.5). 9. Balance = 7 − 8 → `a_cargo` / `a_favor` / `cero`.

**Personal deductions** (AR §B.2), order fixed:

1. Drop ineligible rows (cash for the four cash-restricted categories), each with a visible line.
2. Own caps: funeral ≤ 1 UMA anual ($42,794.64) per year ⚠️; donations ≤ 7 % × `prior_year_accumulable_income`,
   of which `government` ≤ 4 %; retirement ≤ min(10 % × ingresos acumulables, 5 UMA anuales = $213,973.20);
   tuition per student = min(Σ tuition − scholarships, cap of the **highest** level that student had in the
   year: preescolar $14,200, primaria $12,900, secundaria $19,900, profesional técnico $17,100,
   bachillerato $24,500).
3. **Global cap** on the capped group (medical, funeral, donations, mortgage real interest, health
   insurance, school transport, local payroll tax): min(**5 UMA anuales = $213,973.20**, **15 % × total
   income**). Total income = gross income of the art. 152 chapters (612/625 collected, 606 rent, salary
   accumulable + exempt) + `other_exempt_income`. RESICO income **excluded** by default ⚠️ (OQ-A4).
4. Add the uncapped items: `medical_disability`, `retirement_contribution` (after its own cap), `tuition`.

Every cap that binds produces a line "Tope aplicado: {concept}, de {x} a {y}" with its legal basis.

**RESICO PF check (`pf_resico_check`)**, AR §B.3: annual income collected (no deductions) × the
non-marginal 113-F rate (`isr.resico_pf.annual`, already in params) − Σ monthly ISR (app estimates or
filed) − 1.25 % withheld → **difference shown as "diferencia informativa"**, never as payable or
refundable. Text: "Tus pagos mensuales de RESICO son definitivos; no presentas anual por estos ingresos
(regla 3.13.7 RMF 2026)."

**PM 601 (`pm_601`)**, AR §B.5: ingresos acumulables (accrual, `invoiced_on`) − deductions −
`annual_depreciation` = utilidad before PTU; − `ptu_paid` = **utilidad fiscal**; − `prior_year_losses` =
**resultado fiscal** (`max(0, …)`) × 30 % − payments − `interest_withholding` → balance. Extras:
**PTU base** = ingresos acumulables − deductions (without PTU paid or losses; the 28 fr. XXX adjustment is
not modelled ⚠️) → **PTU to distribute** = 10 %; **suggested CU next year** = utilidad fiscal ÷ ingresos
nominales, rounded half-up to 4 decimals ⚠️ (`cu_rounding`); 0 when utilidad ≤ 0. Labelled "Sugerido
para tus pagos de {fy+1}; confírmalo con tu contador".

**PM RESICO (`pm_resico`)**, AR §B.6: same shape on cash flow (collected / paid); PTU base = utilidad
fiscal per 212 ¶2 (PTU paid subtracted) ⚠️ `resico_pm_ptu_base`; no CU.

**Money**: centavos in API/storage, `Exact` inside, one rounding per result (half away from zero),
`Ajuste por redondeo` lines, no floats (M5 §2.7). Strategies append lines through the M5 `Ledger`, so the
breakdown is the computation trace (M5 §2.5 line model, citations mandatory).

### 2.8 Legend and disclaimers

- **Legend** (frozen const, every annual response, rendered from the server field):
  **"Esto es un aproximado; tu contador te dará las cifras reales."**
- **Disclaimers** (`Disclaimer {code, text, severity}`, texts versioned with params):

| Code | Severity | When | Text (es-MX, v1) |
|---|---|---|---|
| `annual_estimate` | info | Always, first | "Estimación informativa de la declaración anual {fy}. No sustituye la declaración que presenta tu contador ante el SAT. Basada en LISR arts. 9, 109, 113-F, 150–152, Anexo 8 RMF {fy} y RMF {fy}." (AR Part E) |
| `cpa_review_pending` | strong | Any section uses a ⚠️ rule (§2.10) | "Este cálculo usa reglas que un contador aún no valida: {topics}." |
| `year_in_progress` | warning | `fiscal_year` = current year | "El año aún no termina: calculamos con tus registros hasta {mes}." |
| `payments_estimated` | warning | Any `estimated_by_app` payment | §2.5 |
| `incomplete_inputs` | warning | Any section `incomplete` | "Faltan datos; usamos ceros o dejamos fuera lo que no pudimos verificar." |
| `deductions_as_entered` | info | Any personal deduction applied | "Tomamos tus deducciones personales como las registraste; el SAT las valida con tus CFDI." |
| `resico_pf_definitive` | info | `pf_resico_check` present | §2.7 text |
| `informative_income` | info | PF total income incl. exempt > $500,000 | "Además debes informar en tu anual préstamos, donativos, herencias y premios que recibiste (LISR 150)." |
| `rmf_2027_pending` | info | Always for fiscal year 2026 | "Las facilidades para la anual {fy} (pago en parcialidades, exenciones, RESICO) se publican en la RMF {fy+1}; pueden cambiar." |
| `inflation_adjustment_excluded` | info | `pm_601` / `pm_resico` | "No incluimos el ajuste anual por inflación ni actualizamos pérdidas: pídeselos a tu contador." |
| `no_annual_iva` | info | Always | "El IVA no tiene declaración anual: se paga mes con mes." |

Order: `annual_estimate` first; then strong, warning, info; then by code.

### 2.9 Deadlines

From params (`annual.deadline.pf`, `.pm`, `.nonprofit_informative`), because the SAT can extend them:
**PF 2027-04-30**, **PM 2027-03-31**, **603 informative 2027-02-15**; moved to the next business day by the
M5 calendar (CFF 12 ⚠️, "sujeto a confirmación" note). The RFC-digit grace does **not** apply to the annual
(AR §A.2). Each section shows its own deadline; the taxpayer header shows the earliest pending one.

### 2.10 CPA topics (⚠️, M5 §2.9 mechanics: strong banner, line icon, golden `skip` until sign-off)

| Topic | Triggers |
|---|---|
| `platform_nondefinitive_annual` | 625 non-definitive in `pf_art152` (AR CPA #2) |
| `uma_for_minimum_wage` | A funeral, retirement or dependant rule uses UMA (CPA #3) |
| `deduction_cap_resico_income` | PF with RESICO and any capped deduction (CPA #10) |
| `salary_obligation_basis` | The $400,000 salary test decides `obligated` |
| `cu_rounding` | `pm_601` suggested CU |
| `resico_pm_ptu_base` | `pm_resico` PTU base |
| `resico_pf_relief_renewal` | `pf_resico_check` (rule 3.13.7 must be renewed in RMF 2027) |

### 2.11 Determinism and performance

Pure strategies; same inputs + params + engine version ⇒ byte-identical JSON. One `LedgerReader` pass for
the year, one `Estimator` range call (12 months) only when a default payment is needed, O(rows) deduction
pass. Target: `getAnnualEstimate` **p95 < 500 ms** for 3 regimes, 10,000 records, 200 deductions; pure
compute < 5 ms (benchmark).

## 3. Golden examples (CI hard gate, AR Part D)

| ID | Section | Expected | Gate |
|---|---|---|---|
| AN-01 | `pf_art152` 612 only, medical 40,000 | Base 460,000 → ISR **76,365.86**; credits 85,773.86 → **a favor 9,408.00** | run |
| AN-02 | `pf_art152` 612 + 606 + salary, six deduction kinds | Acumulables 760,000; capped group 105,000 (cap 153,000); retirement 50,000; tuition 12,900; base 592,100 → ISR **107,435.78** − 107,235.91 → **a cargo 199.87** | run |
| AN-03 | `pf_resico_check` (a)/(b) | Differences **−1,320.00** / **+275.00**, `informational`, no balance | run |
| AN-04 | `pf_art152` 625 non-definitive | ISR **16,833.81**, **a cargo 4,833.81** | **skip** (`platform_nondefinitive_annual`) until ANN-27 |
| AN-05 | `pf_art152` two simultaneous employers | ISR **74,013.86** − 48,000 → **a cargo 26,013.86**; `obligated` | run |
| AN-06 | `pm_601` | Resultado 1,180,000 × 30 % = 354,000 → **a cargo 22,000.00**; PTU **150,000.00**; CU **0.1380** | run (CU asserted; topic shown) |
| AN-07 | `pm_resico` | 345,000 − 360,000 → **a favor 15,000.00**; PTU base 1,150,000 | balance run; PTU assertion **skip** (`resico_pm_ptu_base`) |
| AN-08 | `pf_art152` cap binds | Deduction applied **45,000** (not 80,000), ISR **31,807.91** | run |

Plus generated cases: each cap at its boundary (±1 centavo), every tuition level, tariff-row edges of
`isr.annual_152`, cash exclusion per category, and an AN-01 variant with all payments `estimated_by_app`.

## 4. User stories (Given / When / Then)

### US-A1 PF annual with a refund (AN-01)
Given a PF 612 taxpayer with 2026 records (collected 900,000, deductions 400,000), filed payments
55,773.86, PM withholdings 30,000 and one electronic medical deduction of 40,000,
When she opens "Declaración anual 2026",
Then she sees **Saldo a favor $9,408.00**, deadline 30 abr 2027, the legend, `obligation = obligated`, and
the breakdown shows base 460,000, tariff row 7 and each credit with its source.

### US-A2 Several regimes, salary and deductions (AN-02)
Given a PF with 605 (constancia: accumulable 360,000, exempt 20,000, ISR 54,235.91), 612 and 606, and the
six deductions of AN-02 including a primaria tuition of 30,000,
When he opens the annual estimate,
Then one `pf_art152` section shows **Saldo a cargo $199.87**, tuition capped at 12,900 with a "Tope
aplicado" line, and retirement shown outside the global cap.

### US-A3 Global cap binds (AN-08)
Given a PF 612 with total income 300,000 and medical 60,000 + GMM 20,000,
When the estimate is computed,
Then personal deductions applied are **45,000**, a line explains "15 % de tus ingresos totales", and ISR
before credits is 31,807.91.

### US-A4 Two employers (AN-05)
Given two constancias marked `simultaneous` (300,000 / 40,000 and 150,000 / 8,000),
When the estimate is computed,
Then `obligation = obligated` with reason "Tuviste dos patrones al mismo tiempo" and **Saldo a cargo
$26,013.86**. Given a single constancia under $400,000 and no other income, Then `obligation = optional`
and the estimate still shows (a refund is possible).

### US-A5 RESICO PF informative (AN-03)
Given a PF in 626 only,
When she opens the annual,
Then the section is `informational`, says monthly payments are definitive, shows the 113-F difference as
"diferencia informativa" (−1,320.00 in case a), and offers no payable or refundable balance.

### US-A6 Platform seller (AN-04)
Given a PF 625 without the definitive option, sales 480,000, withheld 12,000, purchases 300,000,
Then **Saldo a cargo $4,833.81** with the strong CPA banner (`platform_nondefinitive_annual`).
Given the definitive option instead, Then an `explanation` section says the withholdings were definitive.

### US-A7 PM 601 (AN-06)
Given a 601 PM with the AN-06 figures (PTU paid and losses as tax inputs),
Then **Saldo a cargo $22,000.00**, deadline 31 mar 2027, and informative cards "PTU a repartir
$150,000.00" and "Coeficiente de utilidad sugerido para 2027: 0.1380".

### US-A8 RESICO PM (AN-07)
Given a RESICO PM with the AN-07 figures, Then **Saldo a favor $15,000.00**, PTU base with the ⚠️ icon.

### US-A9 Default payments and confirmation
Given a 612 taxpayer with no filed payments,
When she opens the annual,
Then credits use the app's monthly estimates, each line labelled "Estimado por la app", plus the
`payments_estimated` warning; When she taps "Confirmar pagos" and "Confirmar todos", Then the values are
stored as filed, the warning disappears and the figure is unchanged.

### US-A10 Personal deductions CRUD
Given a PF, When she adds a deduction (category first, then only that category's fields), Then it appears in
the list grouped by category with the running "aplicado / registrado" amounts; When she deletes it, Then an
Undo toast (no auto-dismiss) restores it; When the payment method is cash for a medical expense, Then the
form warns "En efectivo no es deducible" and the estimate shows it excluded.

### US-A11 Explanation-only regimes and deadlines
Given a PM 603, Then one `explanation` section: "Presentas una declaración informativa a más tardar el 15
de febrero de 2027; tu contador la prepara", no figure.

### US-A12 Year in progress
Given today is 2026-10-15, When the user opens 2026, Then figures use records through September 2026,
the `year_in_progress` warning shows and nothing is extrapolated. 2027 → `400 out_of_range`.

### US-A13 Flags, isolation, read-only
Given `module_annual` OFF (or `module_tax_engine` OFF), Then every annual route returns `404` and the
entry is hidden. Given another user's taxpayer id, Then `404`. Given a read-only user (trial ended), Then
reads work and writes return `402`.

## 5. Errors → ErrorCode (single envelope)

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Malformed id / body / year, unknown field, bad enum, amount < 0, `paid_on` outside the year, bad UUID | 400 | `validation_failed` | field codes |
| Future `fiscal_year` | 400 | `validation_failed` | `out_of_range` |
| No / invalid token | 401 | `unauthorized` | |
| Read-only user writing | 402 | `entitlement_required` | |
| Flag OFF (or implied), foreign or unknown taxpayer / row | 404 | `not_found` | |
| `fiscal_year` < 2026 | 422 | `business_rule_violation` | `parameters_unavailable` |
| Personal deduction / constancia on a PM | 422 | `business_rule_violation` | `person_type_not_allowed` |
| Filed payments for a regime not held that year, or a month outside its window | 422 | `business_rule_violation` | `regime_not_active` |
| Write on an archived taxpayer | 422 | `business_rule_violation` | `taxpayer_archived` |
| Rate limit | 429 | `rate_limited` | |

Amounts, constancia figures and deduction details are never logged or echoed in errors.

## 6. Flags

- `module_annual`: default OFF, registered by migration (ANN-22), **prerequisites `module_tax_engine` and
  `module_taxpayers`** (`flags.Prerequisites`); OFF or implied OFF ⇒ all routes `404`, absent from
  `GET /v1/flags`, entry hidden. Data rows are kept and cascade on deletion.
- Presets: `demo_full` (OQ-A7). Not in `demo_core`.
- `module_incomes` / `module_expenses` OFF: figures unaffected (the ledger is still read, MODULE-incomes OQ-8).
- Suites: (`annual`, `tax_engine`) ∈ {ON/ON, ON/OFF, OFF/ON}.

## 7. Accessibility

- Balance card reads as one semantics node: "Declaración anual 2026, estimación. Saldo a favor, 9,408
  pesos. Vence el 30 de abril de 2027. Incompleto." Balance kind in **words**, never colour alone.
- Chips (Incompleto, Informativo, Año en curso, ⚠️ Contador) are text + icon; contrast ≥ 4.5:1 text, ≥ 3:1 icons.
- Reading order: taxpayer → year → strong CPA banner → legend → headline balance → sections → disclaimers.
- Targets ≥ 44 × 44; input text ≥ 16 px; 200 % text scaling without clipping; reduce-motion respected.
- Forms follow `rules/ui-forms-and-modals.md`: the personal-deduction form has up to 7 fields → **its own
  screen**; constancia (≤ 6) → sheet; filed payments → inline editable rows (12 rows, 1 field each) with
  "Confirmar todos". Money inputs `inputmode=decimal`, accept "1,234.50" and "1234.5". Required errors on
  submit only; error summary on top with links.

## 8. UX notes

- Entry "Anual" for the active taxpayer (placement: design in ANN-17). Year picker defaults to the
  previous year from January to the PF deadline, else the current year.
- Headline card: balance + deadline + "Ver cómo se calculó"; under it "Datos para tu anual" with three
  rows and counters: Deducciones personales (PF), Constancias de sueldos (PF), Pagos provisionales
  presentados (per regime).
- Breakdown reuses the M5-26 widgets (lines, citations, source chips, rounding adjustment).
- Formal impersonal copy, es-MX money format; "Saldo a cargo" / "Saldo a favor" / "Sin saldo".

## 9. Security (gate ANN-25)

- Constancias and deductions are financial + family data (beneficiary relation, student label): RLS
  ENABLE + FORCE on all four tables, isolation tests per table, cascade on user/taxpayer deletion,
  `Cache-Control: no-store`, never logged. `student_label` and `employer_label` are free text: length
  bounds, Unicode normalization (ADR-0004), no PII required (labels may be aliases).
- No CFDI or employer RFC fetched or stored beyond the optional UUID string.
- Server-side flag enforcement; engine OFF ⇒ annual OFF.
- Rate limits: annual reads counted in the `estimate_reads` budget (one annual = 1 read + 1 for the
  12-month range when used; M5 FU5 precedent).

## 10. Contract changes for ANN-02 (architect)

1. New schemas: `AnnualEstimate`, `AnnualSection`, `AnnualTax`, `AnnualBalance`, `PmAnnualExtras`,
   `AnnualBreakdown`; reuse M5 `Line`, `Disclaimer`, `ParameterSet`, `Citation`.
2. New: `PersonalDeduction` (+ create/patch), `SalaryStatement` (+ create/patch), `AnnualInputs`,
   `FiledPayments` (12 nullable months + per-month `app_estimate_cents` and `source`).
3. Every operation `x-module: module_annual`; errors per §5; examples AN-01, AN-03, AN-06 and an
   `in_progress` year.
4. `annual_depreciation` added to the tax-inputs key enum (additive, M5 contract).
5. ADR-0008: annual engine placement (recommended: pure strategies in `internal/taxengine/annual`,
   reusing params, `Ledger`, `Exact` and tariff lookup; inputs, use case and HTTP in `internal/annual`).

## 11. Out of scope

- Filing, the SAT pre-filled return, CFDI download/validation of deductions.
- Annual inflation adjustment (601), inflation update of losses/PTU, an asset register (depreciation is typed).
- 607, 608, 611, 614 figures, foreign tax credits, dividends gross-up, AGAPES, groups, coordinados, RIF.
- Paying in installments (rule 3.17.3 for 2026 not published), refund requests, offsets.
- Annual IVA figures (no such return); a 12-month IVA summary is a later Could.
- ISSIF / dictamen, PM informative annexes; 603 remanente.
- Recomputing salary exemptions or the employment subsidy (constancia totals are trusted).
- Mid-year RESICO exit recomputation (rule 3.13.5): explanation only.

## 12. Decisions (coordinator, 2026-10-15) — all DECIDED

| # | Question | Decision |
|---|---|---|
| OQ-A1 | Unify M5 `prior_provisional_payments` and `annual_filed_payments` into one table? | **DECIDED:** **No in v1**: precedence filed > tax input > app estimate (§2.5); unify later if both are used |
| OQ-A2 | Salary constancias for a PF without 605 registered? | **DECIDED:** **Allowed** for any PF; the screen suggests adding 605 |
| OQ-A3 | Extrapolate the current year to December? | **DECIDED:** **No**: records to date + `year_in_progress` warning |
| OQ-A4 | RESICO income in the 15 % "total income" for the cap? | **DECIDED:** **Excluded** (conservative) + CPA topic `deduction_cap_resico_income`; legal reasoning in [CPA-REVIEW](../reviews/CPA-REVIEW.md#annual-estimate-module_annual-module-annual) |
| OQ-A5 | Donations without 2025 acumulables | **DECIDED:** **Excluded** and section `incomplete` |
| OQ-A6 | 614 interest in art. 152 | **DECIDED:** **Explanation only** in v1 |
| OQ-A7 | Presets including `module_annual` | **DECIDED:** **`demo_full` only** |
| OQ-A8 | Default year shown | **DECIDED:** Previous year Jan → 30 abr, else current year |
| OQ-A9 | Free tier: are annual inputs counted in trial record caps? | **DECIDED:** **No**; writes need `pro`/trial like `putTaxInputs` (402 when read-only) |
