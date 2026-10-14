# MODULE-tax-engine — M5 Tax Engine (`module_tax_engine`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-13 · MoSCoW: **Must** · Depends on: M2 (taxpayers,
regimes catalog, `TaxpayerProfileReader`), M3-07 / M4-08 (`LedgerReader` aggregates), P0 (access policy)
Sources: PRD §5, §6 (US-T4, US-X1..X3, US-H1), §8, §9 · MODULES.md M5 · PLAN M5-01..M5-33 ·
ADR-0002 (engine design, params, rounding) · research `docs/research/tax-regimes-mx-2026.md` (§0, Part A,
Part B, Part C, Part D) · `docs/reviews/CPA-REVIEW.md` · MODULE-taxpayers §2.3–§2.7 · MODULE-incomes
§2.1, §2.2, §2.5, §2.6, OQ-4, OQ-8 · MODULE-expenses §2.1, §2.3–§2.6 · catalog migration
`backend/migrations/20261013180000_regimes_catalog.sql` (`required_inputs`, `option_defs`).

ADR-0002 is the design authority. This file states **what the user gets and how we know it is
right**; where it adds to the ADR, the addition is listed in §9 (contract) or §11 (open questions).

## 1. Scope

For the **active taxpayer** and one **period** (`YYYY-MM`), M5 returns an *informational estimate* of:

- **monthly IVA**, consolidated per taxpayer (one IVA result, not one per regime, §2.6);
- **provisional ISR per regime** (monthly or year-to-date, as the law says for that regime, §2.4);
- **withholdings to remit** (ISR/IVA the taxpayer withheld from suppliers; PM only), shown apart;
- per regime: a status (§2.2), the missing inputs and a breakdown whose lines are the computation
  trace, each with its legal basis (§2.5);
- totals, the due date, the mandatory legend and the disclaimers (§2.3).

M5 also stores the per-fiscal-year **tax inputs** the strategies need (§2.8). It never files, never
stores an estimate (estimates are derived on every read, ADR-0002 §5) and reads records only through
ports; it imports no M3/M4 code (archtest).

| Op | Endpoint | Contract |
|---|---|---|
| getTaxpayerEstimate | `GET /v1/taxpayers/{taxpayerId}/estimates?period=YYYY-MM` | exists; changes in §9 |
| getEstimateBreakdown | `GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown?regime=` | exists; changes in §9 |
| getTaxInputs / putTaxInputs | `GET\|PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs?fiscal_year=` | exists; changes in §9 |

Ports M5 consumes (declared consumer-side in `internal/taxengine/usecase`):
- `LedgerReader` (M3-07 incomes, M4-08 expenses): pre-aggregated `MonthSums` per `(month, regime)` for
  the fiscal year up to the period, in one indexed `GROUP BY` per table. Non-deleted records only.
  Reads data even when `module_incomes` / `module_expenses` are OFF (MODULE-incomes OQ-8, DECIDED).
- `TaxpayerProfileReader` (M2): person type, regimes with `[start_period, end_period]`, options,
  archived flag.
- `RegimeCatalogReader` (M2): `treatment`, `required_inputs`, `explanation`, `needs_cpa_review` per
  `(code, person_type)`.
- `TaxInputReader` (M5's own repository, `taxpayer_tax_inputs`).
- `HolidayCalendar` (P0 `holidays` table) and `Clock` (America/Mexico_City).
- `access.Policy` (P0-16) for `putTaxInputs` only (read-only users get `402`; reads stay open).

Ports M5 provides:
- `RateCatalog` (M5-08): allowed IVA rates and suggested withholding rates on a date, plus dated
  parameters such as `deduction.cash_payment_limit`, for M3/M4. It implements
  `fiscal.RateCatalog` (`internal/kernel/fiscal`, ADR-0007) and the expenses module's port that
  embeds it. M5 reuses the kernel's `Date`, `IVARate`, IVA rule and withholding shares rather than
  keeping a third copy.
- `Estimator` (M5-22): `EstimateMonth(ctx, user, taxpayer, period)` for M6 (dashboard), M8 (history,
  with the records fingerprint) and M9 (reminder amounts).

## 2. Domain rules

### 2.1 What one estimate contains

```
TaxpayerEstimate(taxpayer, period)
├── legend                      const, always (PRD §6)
├── disclaimers[]               §2.3, ordered by code
├── due_date                    §2.10
├── regimes[]                   one RegimeEstimate per regime held in the period, catalog order
│   ├── treatment, status, missing_inputs[], explanation
│   ├── isr: TaxResult | null   null when the regime has no ISR figure (informational, partial)
│   ├── iva_components | null   this regime's contribution to the consolidated IVA
│   ├── withheld_to_remit_cents
│   ├── cpa_review_pending, cpa_topics[]   §2.9
│   └── parameter_set {fiscal_year, version}
├── totals { iva, isr, withholdings_to_remit_cents, total_payable_rounded_cents }
├── records_fingerprint         SHA-256 (ADR-0002 §5), used by M8
└── engine_version
```

- **Which regimes**: every regime of the taxpayer whose window `[start_period, end_period]` contains
  the period, archived regimes included for periods ≤ `end_period` (MODULE-taxpayers §2.7). A
  taxpayer with no regime in the period → `200` with `regimes: []`, zero totals, and the client's
  empty state (no error: M6/M8 iterate over months).
- **Archived taxpayer**: estimates and breakdowns stay readable (MODULE-taxpayers §2.7); `putTaxInputs`
  is refused (`422 taxpayer_archived`).
- **Records in informational regimes** never change the numbers (MODULE-incomes OQ-4).

### 2.2 Status values

`RegimeEstimate.status` (contract change §9 adds `partial`):

| Status | When | Number shown | UI |
|---|---|---|---|
| `complete` | Treatment full, every input the strategy needs for **this period** is present | ISR and IVA | The amount |
| `incomplete` | A required input for this period is missing (§2.8). The strategy still computes, with **zero** for each missing input (ADR-0002 §2) | ISR and IVA, marked "incompleto" | Amount + amber "Incompleto" chip + CTA "Completa tus datos" → tax-inputs form of that regime and fiscal year, focused on the first missing key |
| `partial` | Treatment partial (621, 622, 603, 625 without the definitive option, §2.4): only some taxes are estimated | IVA (and ISR only where §2.4 says so) | Amount + "Parcial" chip + the explanation of what is not estimated |
| `informational` | No strategy, or the catalog marks the regime informational | None | Explanation text only, no amount (PRD US-X1) |

- **Precedence** when several apply: `informational` > `incomplete` > `partial` > `complete`.
  `missing_inputs` is filled whenever inputs are missing, whatever the status.
- **Required inputs depend on the period**, not only on the regime: `RequiredInputs(options, period,
  facts)` (ADR-0002 §1). Examples: prior provisional payments are not required for January; the
  coeficiente is not required in a 601 first fiscal year; `prior_year_income` is required for 625 only
  when `definitive_option = true`.
- Taxpayer-level summary for the client: `totals` carry no status; the client shows "Incompleto" on the
  totals card when any regime is `incomplete`.

### 2.3 Legend and disclaimers

- **Legend** (frozen const, every estimate and breakdown response, rendered from the server field,
  never a client copy): **"Esto es un aproximado. No todos los gastos aplican para tu giro comercial."**
- **Disclaimers** (`Disclaimer {code, text, severity}`; texts versioned with the parameter set, so they
  track the law year; `severity` is a contract change, §9):

| Code | Severity | When | Text (es-MX, v1) |
|---|---|---|---|
| `not_tax_advice` | info | Always | "Estimación informativa. No sustituye la declaración ante el SAT ni la asesoría de un contador. Cálculo basado en LISR, LIVA, LIF 2026 y Anexo 8 RMF 2026." (research Part C, mandatory) |
| `incomplete_inputs` | warning | Any regime `incomplete` | "Faltan datos para este cálculo; se usaron ceros. Complétalos para un resultado más preciso." |
| `provisional_parameters` | warning | The period's fiscal year has no published params; the latest previous year is used (ADR-0002 §3) | "Aún no se publican las tarifas de {fy}; usamos las de {fy−1}." |
| `informational_regime` | info | Any regime `informational` | "Algunos regímenes son informativos: la app no estima su impuesto." |
| `partial_regime` (**new**) | info | Any regime `partial` | "En algunos regímenes sólo estimamos parte de los impuestos." |
| `deductibility_claimed` | info | Any deductible expense entered the ISR base | "Tomamos como deducibles los gastos que marcaste así; el SAT puede no aceptarlos todos." (research Part C) |
| `no_coefficient_first_year` | info | 601 with `first_fiscal_year = true` | "En el primer ejercicio no hay pagos provisionales de ISR (LISR art. 14)." |
| `iva_mixed_activity` (**new**) | warning | A regime with creditable IVA and exempt acts: `iva_fully_taxable = false`, or exempt (`iva_rate = exempt`) collections in the month (M5-09 review) | "Tienes actividades exentas: el IVA acreditable debe ser proporcional y aquí se acreditó completo." (LIVA 5 fr. V, research Part C) |
| `cpa_review_pending` (**new**) | **strong** | Any line uses a rule still open in CPA-REVIEW (§2.9) | "Este cálculo usa reglas que un contador aún no valida: {topics}. Tómalo con mayor reserva." |
| `resico_cap_exceeded` (**new**, M5-13/M5-16) | warning | A 626 regime collected more than its annual cap in the period's fiscal year (Jan..n): PF > $3,500,000 (LISR 113-E); PM > $35,000,000, counting `ytd_before_app_income` (LISR 206 ¶2). The month is still estimated with RESICO rules | PF: "Tus ingresos cobrados en el año ya pasan de $3,500,000, el límite de RESICO. Desde el mes siguiente dejas RESICO y pagas en el régimen general de tu actividad (612 o 606); revisa con tu contador." (LISR 113-E, 113-I) · PM: "Los ingresos cobrados de la empresa en el año ya pasan de $35,000,000, el límite de RESICO. Desde el siguiente ejercicio tributa en el régimen general (601, Título II) y debe avisarlo al SAT a más tardar el 31 de enero." (LISR 206, 214) |

`not_tax_advice` sits first; then by severity (strong, warning, info), then by code. Deterministic order.

### 2.4 Regime coverage and algorithms

Strategies are keyed by `(regime_code, person_type)` (ADR-0002 §1). Coverage at launch:

| Key | Treatment | ISR (per period n) | IVA | Research |
|---|---|---|---|---|
| **612 PF** | full | **YTD**: Σ cash income Jan..n (+ `ytd_before_app_income`) − Σ deductible expenses paid Jan..n (+ `ytd_before_app_deductions`) − `ptu_paid` − `prior_year_losses` = base; base ≤ 0 → ISR 0; else **art. 106 accumulated tariff of month n** − Σ prior provisional payments (months < n) − Σ ISR withheld by PM Jan..n | monthly | A.2.2, A.2.3 |
| **606 PF** | full | **Monthly**: rent collected in n − (`blind_35`: 35 % × rent + predial paid in n; `actual`: deductible expenses paid in n) → **art. 96 monthly tariff** − ISR withheld by PM in n. Negative → `in_favor` | monthly; residential rent is exempt (record `iva_rate = exempt`) | A.2.1, A.2.4 |
| **626 PF** | full | **Monthly**: cash income in n × **non-marginal** bracket rate − ISR (1.25 %) withheld by PM in n | monthly | A.2.5 |
| **625 PF** | full with `definitive_option` and `prior_year_income ≤ $300,000`; else **partial** | Definitive: platform withholding is the ISR (no further ISR); directly collected income self-paid at the activity rate (2.1 / 4 / 2.5 %, 20 % without RFC). Partial: withholdings shown as provisional credits; ISR not estimated (explained: "calcula tu pago en 612 o RESICO y acredita estas retenciones") | definitive: platform-withheld IVA excluded as definitive; directly collected IVA at 8 % of collections ⚠️ (18-M); partial: general rules | A.2.6, A.3 |
| **601 PM** | full | **YTD accrual**: `coeficiente_utilidad` × Σ income by `invoiced_on` Jan..n − PTU (`ptu_paid` ÷ 8 × (n − 4) for n ≥ 5, capped at the profit) − `prior_year_losses` → × 30 % − Σ prior provisional payments − `interest_withholding`. First fiscal year: ISR 0, status `complete` | monthly (cash, LIVA 1-B) | B.2.1, B.3 |
| **626 PM** | full | **YTD cash**: Σ cash income − Σ deductible expenses paid (+ `ytd_before_app_*`) − `ptu_paid` − `prior_year_losses` → × 30 % − Σ prior provisional payments − `interest_withholding` ⚠️ | monthly | B.2.2 |
| **621 PF** | partial ⚠️ | **Bimonthly** (only on the bimester's second month): income − deductions for the bimester → bimonthly table (= February accumulated table) → − reduction for `rif_year` ⚠️. Odd month: no figure, explanation "El RIF se paga por bimestre; consulta {mes par}" | bimonthly ⚠️ (OQ-T7, DECIDED) | A.2.7 |
| **622 PM** | partial | Not estimated (AGAPES exemption and reduction deferred); explanation | monthly | B.1 |
| **603 PM** | partial | Not estimated (no provisional ISR; explanation) | monthly | B.1 |
| 605, 607, 608, 610 (PF/PM), 611, 614, 615, 616, 620, 623, 624 | informational | Explanation only (text from M5-19) | — | A.1, B.1, Part C |

- **YTD accumulation is per law, not per UI**: 612, 626 PM and 601 accumulate from January of the
  period's fiscal year; 606, 626 PF and 625 are monthly; **IVA is never YTD** (LIVA 5-D).
- **Income date** per regime basis (MODULE-incomes §2.2): ISR by `booked_on` (`collected_on` for cash,
  `invoiced_on` for accrual); IVA always by `collected_on`. Expenses always by `paid_on`.
- **Amounts come from the records** as entered (centavos): income `amount_cents` (before IVA),
  `iva_cents`, `isr_withheld_cents`, `iva_withheld_cents`; expense `amount_cents`, `iva_cents`,
  effective `deductible`. The engine never recomputes a record's IVA or withholding from a rate; the
  8 % border rate is therefore just data in `iva_cents`.
- **Deductible expenses** = effective `deductible = true` (MODULE-expenses §2.4). Depreciation of
  investments is out of scope (investment categories default non-deductible).
- **Tariff lookup**: binary search over rows sorted by `lower`, O(log k), k = 11. Tables are the
  published ones for the period's month and are never derived (research A.2.2).
- **Adding a regime** = one strategy file + params + goldens + one registry line; no other module
  changes (US-X3, ADR-0002 §1).

### 2.5 Breakdown lines

The breakdown is the computation trace: strategies compute **only** by appending lines through the
`Ledger` helper, so the explanation cannot drift from the math (ADR-0002 §1).

| Field | Meaning |
|---|---|
| `seq` | 1-based, fixed order per strategy (deterministic) |
| `tax` | `isr` or `iva` (`withholding` added, §9, for to-remit lines) |
| `concept` | Spanish label, e.g. "Ingresos cobrados acumulados ene–mar" |
| `operation` | `input` (a sum of records or a tax input), `add`, `subtract`, `multiply` (shows `rate`), `lookup` (tariff row: shows the row and month), `result` (subtotal), `rounding_adjustment` (§2.7), `info` (a fact with no arithmetic, e.g. "Retención definitiva") |
| `amount_cents` | The line's value rounded to the centavo for display |
| `rate` | Decimal string when the line multiplies |
| `legal_basis[]` | `{law, article, note}`; **every line except `info` and `rounding_adjustment` has ≥ 1 citation** (golden + contract test) |
| `source` (**new**) | `records`, `tax_input:<key>`, `params:<key>` or `computed`, so the user sees where each figure came from |
| `needs_cpa_review` (**new**) | `true` when the line applies an open CPA-REVIEW rule (§2.9) |

Citation catalog (minimum per strategy; the golden suite asserts them):

| Step | Citation |
|---|---|
| 612 YTD cash income / deductions / base / tariff / prior payments / PM 10 % withholding | LISR 102; LISR 103, 105; LISR 106 ¶1; Anexo 8 RMF 2026 B.VI (month n); LISR 106 ¶4; LISR 106 ¶5 |
| 606 rent / blind 35 % + predial / actual deductions / tariff / PM 10 % | LISR 116 ¶2; LISR 115 ¶2; LISR 115 fr. I–VI; Anexo 8 B.V (art. 116 table); LISR 116 ¶5 |
| 626 PF income / rate / 1.25 % | LISR 113-E ¶5; LISR 113-E (table); LISR 113-J |
| 625 rates / definitive / no RFC / IVA | LISR 113-A fr. I–III, LIF 2026 art. 25 fr. VI; LISR 113-B; LISR 113-C fr. IV; LIVA 18-J fr. II a), 18-L, 18-M |
| 601 coeficiente / YTD income / PTU / losses / 30 % / interest withholding | LISR 14 fr. I; LISR 14 fr. II; LISR 14 fr. II a); LISR 14 fr. II c); LISR 9, 14 fr. III; LISR 54, LIF 2026 art. 24 |
| 626 PM | LISR 207, 208, 211, 9 |
| 621 | LISR 111 (texto vigente hasta 31-12-2021), Transitorio Segundo fr. IX DOF 12-11-2021; Anexo 8 B.VII |
| IVA charged collected / creditable / withheld from taxpayer / consolidation | LIVA 1, 1-B; LIVA 5 fr. I–V; LIVA 1-A, RLIVA 3; LIVA 5-D |
| Withholdings to remit | LIVA 1-A; LISR 106 ¶5 / 116 ¶5 / 113-J (ISR side) |
| Due date | LISR 106, 113-E, 116, 14; LIVA 5-D; CFF 12 ⚠️ |

The breakdown header shows the parameter set (fiscal year, version) and the engine version (US-X2).

### 2.6 IVA consolidation across regimes (⚠️ pending CPA)

- Each strategy reports `IvaComponents` for its regime: `charged_collected`, `creditable_paid`,
  `withheld_from_taxpayer`, `definitive_excluded` (ADR-0002 §1).
- **Default (state until the CPA answers, CPA-REVIEW "Monthly IVA netting across regimes")**: the use
  case nets the components of **all** regimes of the taxpayer into **one** monthly IVA result:
  `IVA = Σ charged_collected − Σ definitive_excluded − Σ creditable_paid − Σ withheld_from_taxpayer`.
  Positive → `payable`; negative → `in_favor` ("saldo a favor", never shown as a payment, PRD US-X1);
  zero → `zero`. Rationale: IVA is a per-taxpayer tax (LIVA 1, 5-D), one return per RFC.
- When **two or more** regimes contribute non-zero IVA components, the consolidation line carries
  `needs_cpa_review = true` and topic `cross_regime_iva_netting` (§2.9).
- Per-regime IVA is **never** presented as a separate amount to pay; the per-regime card shows its
  components and links to the consolidated line.
- **A prior month's saldo a favor is not carried forward** in v1 (LIVA 6 allows it; out of scope,
  OQ-T5). The breakdown says so in an `info` line when the previous month was `in_favor`.
- `iva_fully_taxable = false`, **or exempt collections in the month** (the only signal for regimes
  without the option, e.g. 606 with residential and commercial rent): creditable IVA is taken in full
  (no proportion, LIVA 5 fr. V / 5-B deferred) and `iva_mixed_activity` is added, both only when the
  regime has creditable IVA. 0 % collections are taxed acts and do not trigger it. Mixing across
  regimes (one regime taxed, another exempt) is not detected per regime; M5-22 owns it.
- Two different withholdings, never confused (OQ-T6, DECIDED):
  - **IVA withheld *from* the taxpayer** by a persona moral or platform, as **recorded on the income**
    (`iva_withheld_cents`), **reduces** the taxpayer's monthly IVA (`withheld_from_taxpayer`; IVA a
    cargo = trasladado cobrado − acreditable − retenido; LIVA 1-A, 5-D; research PF-612-IVA-01). It is
    never inferred from a rate or a counterparty type: only the recorded amount counts.
  - **IVA the taxpayer (a PM) withheld *from its suppliers*** (recorded on the expense) goes to
    `withholdings_to_remit` only. It is **not** also credited as acreditable IVA (LIVA 5 fr. IV) in v1
    (conservative, matches PM-601-IVA-01; CPA item).

### 2.7 Money and rounding

- **API and storage**: `int64` centavos. **Inside the engine**: `Exact` at 10^-6 peso with 128-bit
  intermediates; no floats anywhere (`forbidigo`). Rates are fixed-point decimals; the
  `coeficiente_utilidad` is used **as given** by the user (art. 14 is silent; 4-decimal practice ⚠️).
- **Result rounding**: each `TaxResult` is rounded **once, at the end**, to the centavo, half away from
  zero (research "rounded to 2 decimals at the end only").
- **Display rounding and `Ajuste por redondeo`**: every line is displayed rounded to the centavo. If
  the displayed lines of a result do not add up to its rounded value, the engine appends one
  `rounding_adjustment` line ("Ajuste por redondeo", ±0.01…±0.0n) so every breakdown reconciles on
  screen. Never a hidden centavo. Property test: displayed lines + adjustment = result (M5-24).
  - Illustration (synthetic): lines 100.004 and 50.004 display as 100.00 and 50.00; the exact result
    150.008 rounds to 150.01 → "Ajuste por redondeo +0.01".
- **Whole-peso payment rounding (CFF art. 20)**: **not applied** pending CPA. `payable_rounded_cents =
  amount_cents` (positive part) until `rounding.payment_whole_pesos` is switched on in params (a data
  change, ADR-0002 §4). No UI copy may say "redondeado a pesos" while it is off.
- Totals: `totals.isr.amount_cents` = Σ of positive per-regime ISR (favor balances stay per regime and
  are not offset against other regimes' ISR; each regime is its own payment concept);
  `total_payable_rounded_cents = isr.payable_rounded + iva.payable_rounded (if payable) +
  withholdings_to_remit_cents`.
- Rounding happens **only** in the engine (PRD §9). Adapters and the app format, never round.

### 2.8 Tax inputs (`taxpayer_tax_inputs`, per taxpayer + regime + fiscal year)

The keys and kinds come from `regimes.required_inputs` (catalog data, M2-02); M5 validates against them.

| Regime | Key | Kind | Required for period n when… | Missing → |
|---|---|---|---|---|
| 612 PF, 626 PM, 601 PM | `prior_provisional_payments` | money **per month** (§9, OQ-T1) | n > 1: one value for every month m < n in which the regime was active | `incomplete`, missing key names the months |
| 612 PF, 626 PM | `ytd_before_app_income`, `ytd_before_app_deductions` | money | never required (optional; prompts when the regime started before the first record of the year) | — |
| 612 PF, 626 PM, 601 PM | `ptu_paid` | money | never required | — |
| 612 PF, 626 PM, 601 PM | `prior_year_losses` | money | never required | — |
| 601 PM, 626 PM | `interest_withholding` | money | never required | — |
| 601 PM | `coeficiente_utilidad` | rate, `0 ≤ cu ≤ 1` | always, unless `first_fiscal_year = true` | `incomplete` (computed with 0 → ISR 0, clearly labelled) |
| 601 PM | `first_fiscal_year` (**new**) | boolean (§9) | never (default `false`) | — |
| 625 PF | `prior_year_income` | money | `definitive_option = true` | `incomplete`; eligibility evaluated with 0 (i.e. eligible), labelled |
| 622 PM | (catalog defs as 601) | | **none in v1** (ISR not estimated); values may be saved for later | — |

- Predial for 606 `blind_35` comes from expenses of category `predial` paid in the month, not an input.
- The CTA target (`incomplete`) is `GET …/regimes/{code}/tax-inputs?fiscal_year=` with the first
  missing key; the form is per regime (M5-27). A suggested prefill for each missing prior payment is
  the app's own estimate for that month (UX only; the stored value is what the user enters).
- PUT replaces all values for `(taxpayer, regime, fiscal_year)`; amounts ≥ 0; idempotent.

### 2.9 Cases the CPA has not validated (⚠️)

Every ⚠️ rule still runs with the documented default, **but the user sees a stronger warning**:

- The affected `RegimeEstimate` has `cpa_review_pending = true` and `cpa_topics[]`; affected lines have
  `needs_cpa_review = true` (⚠️ icon next to the line).
- The response carries `cpa_review_pending` (severity `strong`): a non-dismissible amber banner **above
  the number** on the estimate card and the breakdown, naming the topics in plain Spanish.
- The same case's golden stays `skip` with its reason until M5-33 (ADR-0002 §5).

Runtime topics (append-only enum, each mapped to a CPA-REVIEW line):

| Topic | Triggers when | CPA-REVIEW item |
|---|---|---|
| `rif_reduction` | Any 621 figure | RIF reduction by year |
| `rif_bimonthly_iva` | 621 IVA computed by bimester (OQ-T7) | RIF bimonthly IVA (added with this spec) |
| `platform_definitive_iva` | 625 with the definitive option and IVA > 0 | 625 definitive IVA (18-L/18-M) |
| `cross_regime_iva_netting` | ≥ 2 regimes contribute IVA (§2.6) | Monthly IVA netting across regimes |
| `resico_pm_interest_withholding` | 626 PM with `interest_withholding > 0` | 626 PM credits art. 54 withholding |

Not escalated to the strong warning (to avoid warning fatigue; OQ-T8, DECIDED): expense-category
deductibility (covered by `deductibility_claimed`), the CFF 12 due-date move (a note under the date),
the coeficiente's decimals (used as given).

### 2.10 Period selection and due date

- `period` is **optional** on `getTaxpayerEstimate` (contract change §9): absent → the **current month**
  in America/Mexico_City (`Clock` port). The path `{period}` on the breakdown stays required.
- **Past months** are allowed back to the **first fiscal year with published parameters** (2026 in v1).
  Earlier → `422 business_rule_violation`, detail `parameters_unavailable` (2025 tables are not
  transcribed; using 2026 tables would be wrong).
- **Future months** → `400 validation_failed`, detail `out_of_range`. The current, still-open month is
  allowed: it estimates with the records so far, labelled "Mes en curso" in the UI.
- Parameter selection: the highest published version of the period's fiscal year; values resolved by
  date (`Params.At(key, date)`; period-level values use the last day of the period). A new fiscal year
  without params falls back to the previous year + `provisional_parameters` (relevant only for the
  current month in January, before the RMF publication).
- **Due date**: day 17 of the month after the period (after the bimester for 621), moved to the next
  business day using `holidays` (CFF 12 ⚠️, shown with a small "sujeto a confirmación" note). The
  RFC-digit grace days are **not** applied (research A.4 ⚠️).

### 2.11 Determinism and performance

- Strategies are pure functions of `(Input, Params)`: no clock, no randomness, no I/O, no floats, no
  map-order dependence. Same records + inputs + params + engine version → **byte-identical JSON**
  (M5-24). `records_fingerprint` = SHA-256 over the canonical aggregated input + param version +
  engine version.
- Cost: one aggregate query per table for the fiscal year (O(records) in the database, indexed on
  `(user_id, taxpayer_id, booked_on)`), O(1) strategies apart from O(log k) tariff lookups.
- **Targets** (server-side, measured in M5-23 with EXPLAIN ANALYZE and a Go benchmark):
  - `getTaxpayerEstimate` and `getEstimateBreakdown`: **p95 < 300 ms** for a taxpayer with 3 regimes and
    10,000 records in the fiscal year (same budget as record CRUD, PRD §9);
  - pure engine compute for one period: **< 2 ms** (benchmark, no I/O);
  - M6 dashboard (12 months via `Estimator`, one aggregate pass) keeps its p95 < 1 s.
- No cache in v1 (YAGNI); a fingerprint-keyed cache is the escape hatch (ADR-0002 §5).

## 3. Golden numeric examples (research, CI hard gate)

The golden suite (`internal/taxengine/testdata/golden/<regime>-<pf|pm>/<case>.yaml`) is seeded with
these research cases; **expected values are the research's, to the centavo**. ⚠️ cases ship `skip`
with reason until M5-33.

| ID | Expected | Status |
|---|---|---|
| PF-612-ISR-01 (March, base 60,000, March row 6) | ISR caused 7,150.96 → **150.96** to pay | gate |
| PF-612-ISR-02 (October, base 250,000) | caused 34,516.54 → **4,516.54** | gate |
| PF-612-ISR-03 (deductions > income) | **0.00** | gate |
| PF-612-IVA-01 (4,800 − 1,600 − 1,066.67) | **2,133.33** payable | gate |
| PF-626-ISR-01 (60,000; 20,000 from PM) | 900.00 − 250.00 = **650.00** | gate |
| PF-626-ISR-02 (25,000.00 / 25,000.01) | **250.00 / 275.00** (non-marginal) | gate |
| PF-626-IVA-01 | **5,466.67** | gate |
| PF-606-ISR-01 (rent 20,000, blind, PF tenant) | **1,076.00** | gate |
| PF-606-ISR-02 (same, PM tenant, 2,000 withheld) | **0.00** payable, **924.00** in favor | gate |
| PF-606-IVA-01 (commercial rent to PM) | **1,066.67**; residential **0** | gate |
| PF-625-01 (goods, definitive, 30,000) | ISR withheld 750.00, additional ISR **0**; IVA part ⚠️ | ISR gate, IVA skip |
| PF-625-02 (lodging 15,000) | ISR withheld **600.00** | gate |
| PF-621-01 (Sep–Oct, year 6) | **2,383.66** | skip ⚠️ |
| PM-601-ISR-01 (June, CU 0.15, 5,000,000, PTU 120,000) | 216,000 − 180,000 = **36,000.00** | gate |
| PM-601-ISR-02 (first fiscal year) | **0.00**, `no_coefficient_first_year` | gate |
| PM-601-IVA-01 | IVA **120,000.00** + to remit **5,333.33** | gate |
| PM-626-ISR-01 (March) | **50,000.00** | gate |

Plus generated cases (M5-07): each tariff's `lower`, `upper`, `upper + 0.01` for all 12 months and the
art. 96 table; validity boundaries (`rate_8` on 2026-12-31 vs 2027-01-01; UMA January vs February).

Composite cases defined by this spec (use-case level, M5-22; QA verifies the arithmetic in M5-29):

| ID | Setup | Expected |
|---|---|---|
| TX-612-TOTAL-01 | PF-612-ISR-01 + PF-612-IVA-01, same March | ISR 150.96, IVA 2,133.33, **total 2,284.29** (matches the contract example) |
| TX-NET-01 ⚠️ | PF 612 + 606, March. 612 IVA as PF-612-IVA-01 (+2,133.33); 606 commercial rent to PM: IVA 3,200.00, withheld 2,133.33, creditable 3,000.00 (−1,933.33) | Consolidated IVA **200.00** payable; `cross_regime_iva_netting` warning shown |
| TX-601-REMIT-01 | PM-601-ISR-01 + PM-601-IVA-01; the PM also withheld ISR 10 % = 5,000.00 on the 50,000 fees | to remit **10,333.33**; total **166,333.33** |
| TX-612-INC-01 | PF-612-ISR-01 without `prior_provisional_payments` for Jan–Feb | status `incomplete`, missing Jan, Feb; ISR **4,150.96** (7,150.96 − 0 − 3,000.00) labelled |
| TX-ROUND-01 | Any strategy whose displayed lines differ from the rounded result | one `Ajuste por redondeo` line; property test holds |

## 4. User stories (Given / When / Then)

### US-X1 See this month's estimate
- **Given** a PF taxpayer with 612 and the March 2026 records of PF-612-ISR-01/IVA-01 and all inputs,
  **When** I open the estimate (no `period`) on 2026-03-20, **Then** I see March: 612 ISR $150.96, IVA
  $2,133.33, total $2,284.29, due 2026-04-17, the legend and `not_tax_advice`; status `complete`.
- **Given** the same taxpayer, **When** I pick February, **Then** February is recomputed from current
  records (past months allowed).
- **Given** today is 2026-03-20, **When** I request `period=2026-04`, **Then** `400 validation_failed`
  `out_of_range`.
- **Given** IVA nets to −500.00, **When** I view it, **Then** it reads "Saldo a favor $500.00", is not
  counted in the total to pay, and no "pagar" wording appears.
- **Given** a taxpayer with 612 and 605, **When** I view the estimate, **Then** 605 shows its explanation
  with no amount, and `informational_regime` is present.

### US-X1b Incomplete inputs
- **Given** 612 in March without prior payments for January and February, **When** I view the estimate,
  **Then** 612 is `incomplete` with ISR $4,150.96 labelled "Incompleto", `missing_inputs` names
  `prior_provisional_payments` for 2026-01 and 2026-02, `incomplete_inputs` is present, and the CTA opens
  the 612 2026 tax-inputs form at that key.
- **Given** I save those inputs (Jan 2,000.00, Feb 2,000.00), **When** I return, **Then** 612 is
  `complete` with ISR $150.96.
- **Given** 612 in **January**, **When** no prior payments exist, **Then** status is `complete`.

### US-X1c Partial and per-option regimes
- **Given** 625 goods/services with `definitive_option = true` and `prior_year_income = 250,000`,
  **When** I view a month with 30,000 platform sales, **Then** ISR withheld 750.00 shows as definitive,
  additional ISR 0, and the IVA lines carry ⚠️ with the `cpa_review_pending` banner.
- **Given** 625 with `definitive_option = false`, **When** I view the estimate, **Then** status is
  `partial`, ISR is not estimated, withholdings appear as provisional credits, and the explanation
  points to 612/RESICO.
- **Given** 621 with `rif_year = 6`, **When** I view **September**, **Then** no 621 figure, with "El RIF
  se paga por bimestre; consulta octubre"; **When** I view **October** with PF-621-01 data, **Then**
  $2,383.66 with the strong ⚠️ banner (`rif_reduction`).
- **Given** a 601 PM with `first_fiscal_year = true`, **Then** ISR 0.00, status `complete`,
  `no_coefficient_first_year`.

### US-X1d Consolidated IVA and withholdings to remit
- **Given** TX-NET-01, **Then** one IVA result $200.00 payable, both regimes' components visible, and the
  `cross_regime_iva_netting` warning.
- **Given** TX-601-REMIT-01, **Then** "Retenciones a enterar $10,333.33" appears as its own line under
  the totals, is never netted against IVA, and the total is $166,333.33.

### US-X2 Breakdown
- **Given** PF-612-ISR-01, **When** I open the 612 March breakdown, **Then** the lines are those of the
  contract example (income 90,000.00 → … → ISR a pagar 150.96), each with its citation, `source`, and the
  header "Parámetros 2026 v1 · Motor 1.0.0".
- **Given** any breakdown, **Then** the displayed lines reconcile to the result, with an `Ajuste por
  redondeo` line when needed.
- **Given** a regime the taxpayer does not hold in that period, **When** I request its breakdown,
  **Then** `422 business_rule_violation` `regime_not_active`.

### US-X3 Correct for known cases
- **Given** the golden suite, **When** CI runs, **Then** every non-skipped case matches to the centavo;
  any mismatch fails the build (no update flag). Skips list their reason.
- **Given** a new param version for 2027, **When** it is added, **Then** no strategy code changes and
  2026 periods still use 2026 params.

### US-T4 Enter tax inputs
- **Given** a 601 PM, **When** I PUT `coeficiente_utilidad = "0.1500"` for 2026, **Then** 200 and the
  June estimate becomes PM-601-ISR-01's 36,000.00.
- **Given** a value `"1.5"` for the coeficiente, **Then** `400 validation_failed` `out_of_range`.
- **Given** a read-only user, **When** I PUT inputs, **Then** `402 entitlement_required`; GET still works.
- **Given** another user's taxpayer id, **Then** `404 not_found` on every M5 endpoint.

### US-X4 Flag
- **Given** `module_tax_engine` OFF, **Then** every M5 route answers `404`, M6/M8 are implied OFF, M9
  reminders omit amounts, and the app shows no estimate surface.

## 5. Errors → ErrorCode

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Bad `period` / `fiscal_year` format, bad regime code, unknown query param | 400 | `validation_failed` | `invalid` |
| Future period or fiscal year | 400 | `validation_failed` | `out_of_range` |
| Tax input: unknown key for the regime, wrong kind, both/neither value fields, negative money, rate outside [0, 1], month outside 1..12 or on a non-monthly key | 400 | `validation_failed` | `unknown_input`, `invalid`, `out_of_range` |
| Missing/invalid session | 401 | `unauthorized` | |
| Read-only user on `putTaxInputs` | 402 | `entitlement_required` | |
| Flag OFF (or `module_taxpayers` OFF), foreign/unknown taxpayer, malformed path id | 404 | `not_found` | |
| Period before the first fiscal year with params | 422 | `business_rule_violation` | `parameters_unavailable` |
| Breakdown/tax inputs for a regime not held in the period / fiscal year | 422 | `business_rule_violation` | `regime_not_active` |
| `putTaxInputs` on an archived taxpayer | 422 | `business_rule_violation` | `taxpayer_archived` |
| Per-user write budget on `putTaxInputs` (60/h, ADR-0005) | 429 | `rate_limited` | |
| Params invalid at boot | — | the process does not start (fail-fast, ADR-0002 §3) | |
| Anything else | 500 | `internal_error` | |

No new `ErrorCode` values; detail codes go into the contract descriptions (M5-01). A missing input is
**never** an error: it is the `incomplete` status. Messages are es-MX safe text and never carry amounts.

## 6. Security requirements (gate M5-31)

| # | Requirement | Test |
|---|---|---|
| S-1 | `taxpayer_tax_inputs`: ENABLE + FORCE RLS in its first migration, owner policies per operation (`user_id = app_user_id()`), composite FK `(taxpayer_id, regime_code, user_id) → taxpayer_regimes`, `ON DELETE CASCADE` from the taxpayer | `pgtest.AssertIsolation`; deletion-cascade guard lists it |
| S-2 | Estimates read only the caller's records: every port call runs inside the request's RLS transaction (`SET LOCAL app.current_user_id`) | Cross-user isolation test on GET estimates (A cannot see B's sums) |
| S-3 | Amounts, inputs and RFC never logged or traced; errors carry no amounts | Log-capture test at debug level |
| S-4 | `Cache-Control: no-store` on every M5 response | Contract test |
| S-5 | Params embedded (`go:embed`), validated at boot; no runtime path loads params from user input | Boot test with a corrupt file |
| S-6 | Strategies pure: archtest forbids `time`, `math/rand`, `os`, `net`, `database/sql`, GORM and `float*` in `internal/taxengine/domain` | archtest + `forbidigo` |
| S-7 | Path ids validated before query params; `additionalProperties: false` on `TaxInputsUpdate` | Contract tests |
| S-8 | Audit event `tax_inputs.updated` with user id, taxpayer id, regime, fiscal year only | Log-capture test |

## 7. Feature flag `module_tax_engine`

- Default OFF; registered by migration (M5-28). OFF → every M5 route `404` (server boundary).
- Implied OFF when `module_taxpayers` is OFF. M5 OFF implies M6 and M8 OFF (their routes `404`) and M9
  sends reminders without amounts.
- Independent of `module_incomes` / `module_expenses`: M5 reads their data even when their screens are
  hidden (MODULE-incomes OQ-8).
- `RateCatalog` (used by M3/M4 writes) is **not** behind the M5 flag: it is a shared port, not a route.
- Data with the flag OFF: `taxpayer_tax_inputs` still cascades on taxpayer/account deletion.
- Client (single FlagsProvider): OFF → no estimate card, no breakdown, no tax-inputs form, no CTA.
- Suites run with the flag ON and OFF (M5-23); presets `demo_core`, `demo_monetization`, `demo_full`
  include it.

## 8. UX notes (Flutter, M5-25..M5-27)

- **Estimate screen**: period picker (current month default, back to January 2026, no future), totals
  card (ISR, IVA or "Saldo a favor", Retenciones a enterar, Total a pagar, due date), then one card per
  regime with its status chip. Legend always visible on the screen, not collapsed. Strong ⚠️ banner
  above the totals when `cpa_review_pending` is present.
- **Breakdown screen**: lines in `seq` order, amount right-aligned, citation under each concept,
  `source` as a small tag, ⚠️ on `needs_cpa_review` lines, "Ajuste por redondeo" styled as a muted line.
- **Tax-inputs form**: page per regime + fiscal year (field count varies; > 6 for 612/601 → page, per
  `rules/ui-forms-and-modals.md`); monthly prior payments as an inline editable 12-row table (1–4
  fields per row rule), prefilled suggestion = the app's estimate for that month; labels in Spanish
  from `TaxInputDef.label`.
- Amounts are formatted only (es-MX, `$1,234.56`); the client never rounds or recomputes.

## 9. Contract changes for M5-01 (architect)

1. `getTaxpayerEstimate.period`: **optional** (own parameter, not the shared `PeriodQuery`); absent →
   current month (Mexico City).
2. `RegimeEstimate.status`: add `partial`. Add `cpa_review_pending: boolean`, `cpa_topics: CpaTopic[]`
   (append-only enum, §2.9). `missing_inputs`: array of `MissingInput {key, label, months?: Period[]}`
   instead of strings.
3. `Disclaimer`: add `severity: info | warning | strong`. `DisclaimerCode`: add `partial_regime`,
   `iva_mixed_activity`, `cpa_review_pending`.
4. `BreakdownLine`: add `source` and `needs_cpa_review`; `tax` adds `withholding`.
5. `TaxpayerEstimate`: add `records_fingerprint` and `parameter_set` (taxpayer level, for M8).
6. `TaxInputDef.kind`: add `money_monthly` and `boolean`; `TaxInputValue`: add `month` (1..12, for
   `money_monthly`) and `value_bool`. Catalog forward migration (M5-20/M5-21): 612/626 PM/601
   `prior_provisional_payments` → `money_monthly`; 601 gets `first_fiscal_year` (boolean, optional).
7. Error detail codes of §5 in the descriptions; `Cache-Control: no-store` on M5 responses.

ADR-0002 is amended in the same PR for items 2 and 6 (status `partial`, monthly prior payments).

**Applied by M5-01 (contract 0.6.0, frozen).** Items 1-7 are in `docs/api/openapi.yaml`; ADR-0002
Amendment 1. Precisions made while freezing (no rule changed):
- The `fiscal_year` of the tax-input operations follows the period rules: a future year → `400
  out_of_range`, before 2026 → `422 parameters_unavailable`.
- Shape rules are enforced by the schema at the edge, so they answer the JSON Schema keyword, not a
  domain code: negative money → `minimum` (`TaxInputValue.value_cents` is `NonNegativeCents`), a
  `month` outside 1..12 → `minimum`/`maximum`, an unknown body field → `additionalProperties`. The
  domain codes of §5 (`unknown_input`, `invalid`, `out_of_range`) cover what needs the catalog.
- `TaxResult.payable_rounded_cents` is `>= 0` (the positive part of `amount_cents`); the per-regime
  favor balance stays in `amount_cents` with `outcome: in_favor`.
- `DisclaimerCode` and `CpaTopic` are append-only enums (not frozen, like `ErrorCode`).
- **Append (M5-13/M5-16, PR #90):** `DisclaimerCode` gains `resico_cap_exceeded` (§2.3). In plain
  words: once a RESICO taxpayer has collected more than the cap in the year, it stops being RESICO —
  a persona física from the next month (LISR 113-E; art. 113-I lists the exits), a persona moral
  from the next fiscal year, moving to Título II (LISR 206 ¶2, 214). The engine still estimates the
  month under RESICO and says so with this warning; it never moves the taxpayer to another regime.
  Breakdown lines stay money lines: the cap is never an `info` line. Non-breaking (append-only enum;
  clients show the `text` of an unknown code); `info.version` unchanged, no `x-frozen` node changed.
- `Breakdown.disclaimers` carries the estimate's disclaimers that bear on its lines (same order):
  `cpa_review_pending` (strong) whenever the regime or any line needs CPA review, including the
  consolidated IVA line (`cross_regime_iva_netting`), and `iva_mixed_activity` whenever the
  consolidated IVA took creditable IVA in full, so §2.3/§2.9 warnings travel with every figure.
- `TaxInputDef.kind` serves `money_monthly` / `boolean` only after the catalog forward migration
  (M5-20/M5-21); the onboarding tax-inputs step (M2-09) skips kinds it does not edit, the M5-27 form
  owns them.

## 10. Out of scope

- Annual return (PF April, PM March), personal deductions (art. 151), annual PTU computation.
- Filing, payment lines (línea de captura), SAT integration, CFDI import or validation.
- Carry-forward of IVA saldo a favor between months; offsets and refunds (LIVA 6).
- Proportional creditable IVA for mixed activity (flagged only).
- Depreciation of investments; annual inflation adjustment; AGAPES exemptions (612/622).
- 606 quarterly option; 625 non-definitive ISR computation under 612/RESICO (shown as partial).
- Whole-peso payment rounding (off until CPA); RFC-digit grace days; employment subsidy.
- Fiscal years before 2026 (no transcribed params); foreign currency.
- A combined "all my taxpayers" estimate (PRD Q14).

## 11. Decisions (coordinator, 2026-10-13) — all DECIDED

| # | Question | Decision |
|---|---|---|
| OQ-T1 | Prior provisional payments: one YTD value per fiscal year (ADR-0002 as written) or one value per month? A single value is only right for one period | **DECIDED:** **Per month** (`money_monthly`); the app's own monthly estimate is offered as prefill, never assumed |
| OQ-T2 | Cross-regime IVA netting for one taxpayer (612 + 606, etc.) (CPA) | **DECIDED:** **Consolidate per taxpayer**, strong ⚠️ when ≥ 2 regimes contribute |
| OQ-T3 | Whole-peso rounding (CFF 20) (CPA) | **DECIDED:** **Off**; `payable_rounded_cents = amount_cents` |
| OQ-T4 | p95 target for estimate/breakdown | **DECIDED:** **< 300 ms** at 3 regimes × 10,000 records/year |
| OQ-T5 | Carry the previous month's IVA saldo a favor forward? | **DECIDED:** **No** in v1; an `info` line mentions it |
| OQ-T6 | (a) Does IVA withheld **from** the taxpayer, recorded on the income, reduce its monthly IVA? (b) Does IVA a PM withheld **from its suppliers** also become creditable for that PM (LIVA 5 fr. IV)? | **DECIDED: (a) Yes**, only the recorded `iva_withheld_cents`, never inferred. **(b) No** in v1: to-remit only (CPA item) |
| OQ-T7 | 621 IVA bimonthly like its ISR? (CPA) | **DECIDED:** **Bimonthly**, ⚠️ `rif_bimonthly_iva` |
| OQ-T8 | Which ⚠️ items raise the strong banner | **DECIDED:** Only §2.9 topics; deductibility, due-date move and coeficiente decimals stay info |
| OQ-T9 | Periods before 2026 | **DECIDED:** **422 `parameters_unavailable`**; transcribe 2025 only if users ask |
| OQ-T10 | Totals ISR: offset one regime's favor against another's payable? | **DECIDED:** **No**; favor balances stay per regime |
