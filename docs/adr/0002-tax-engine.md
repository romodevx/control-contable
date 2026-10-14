# ADR-0002 — Tax engine: pluggable regime strategies over versioned parameter data

| Field | Value |
|---|---|
| Status | PROPOSED (Phase 2 Design) |
| Date | 2026-10-13 |
| Amended | 2026-10-13 by M5-01: status `partial`, monthly prior payments ([Amendment 1](#amendment-1-m5-01-2026-10-13)) |
| Related | [PRD §5, US-X1..X3](../project-specs/PRD.md), [MODULES M5](../project-specs/MODULES.md), [ADR-0001](0001-stack.md), [research](../research/tax-regimes-mx-2026.md) |
| Grounding | `docs/research/tax-regimes-mx-2026.md` (cut-off 2026-10-13). Items marked ⚠️ there stay ⚠️ here and must not ship as hard-coded behaviour without CPA review |

## Context

We estimate monthly IVA and provisional ISR for every SAT regime, for personas físicas (PF)
and morales (PM). Every number must be explainable line by line, with its legal basis,
and reproducible by golden tests reviewed by a CPA. Adding a regime or a fiscal year must
not touch other modules (US-X3).

The research shapes the design in five ways:

1. **Regime code is not unique.** `626` is RESICO PF (a flat rate on income, arts. 113-E
   to 113-J) *and* RESICO PM (30% on cash profit, arts. 206–215). `610` also exists for
   both person types. Strategies are therefore keyed by **(regime code, person type)**.
2. **Tables are published verbatim per month.** The art. 106 accumulated tariffs (Anexo 8
   B.VI) are twelve tables, and December is **not** 12 × January. They are stored as
   published and never derived.
3. **Parameters have validity ranges inside a fiscal year.** The 8% border IVA is valid
   until 2026-12-31. The UMA changes on 1 February, so January 2026 still uses the 2025
   UMA. Rates change through the LIF each year (platforms: 1% → 2.5% in 2026).
4. **IVA is monthly and not year-to-date. ISR depends on the regime**: year-to-date and
   cash basis (612, 626 PM), monthly (606, 626 PF), coeficiente × year-to-date accrual
   income (601), or withholding (625).
5. **Some outputs are not "your tax"**: a PM must pay in full the IVA and ISR it withheld
   from PF suppliers (LIVA 1-A; PM-601-IVA-01). That is a separate amount to remit.

## Options

1. **One calculator with `switch regime`.** Rejected: it violates OCP and every change
   touches shared code.
2. **Rules DSL stored in the database.** Rejected (YAGNI): it needs its own tooling, and
   these formulas are clearer as typed Go.
3. **One strategy per (regime, person type) in code + versioned parameters as data. Chosen.**

## Decision

### 1. Strategy port (pure domain, `internal/taxengine/domain`)

```go
type StrategyKey struct {
    Regime     RegimeCode // c_RegimenFiscal, e.g. "626"
    PersonType PersonType // Fisica | Moral
}

// Strategy computes one regime's monthly estimate. It must be pure:
// no I/O, no clock, no randomness, no floats, no map-order dependence.
type Strategy interface {
    Key() StrategyKey
    Treatment() Treatment                        // Full | Partial | Informational
    RequiredInputs(o Options, p Period, f Facts) []RequiredInput // per period (Amendment 1); drives "incomplete" + CTA (US-T4)
    Estimate(in Input, p Params) (RegimeEstimate, error)
}
```

- `Input` is assembled by the use case from read ports (`LedgerReader`,
  `TaxInputReader`, `TaxpayerProfileReader`). It contains:
  - `Period`, regime `Options`, and the person type;
  - **pre-aggregated** `MonthSums` for every month of the fiscal year up to the period:
    - income by `collected_on` (cash) and by `invoiced_on` (accrual);
    - IVA charged, IVA collected and IVA withheld from the taxpayer;
    - ISR withheld from the taxpayer, split by withholder kind (PM at 10% / 1.25%,
      platform);
    - expenses by deductibility class, with IVA paid;
    - withholdings the taxpayer made to suppliers;
  - user-supplied `TaxInputs` for the fiscal year.

  Strategies never see raw rows. Aggregation is one indexed `GROUP BY month, regime`
  query, O(n) over the taxpayer's records for the year. Strategies are O(1) apart from
  tariff lookup.
- `RegimeEstimate` holds `{isr *TaxResult, ivaComponents IvaComponents,
  withheldToRemit Cents, lines []Line, status, missingInputs, explanation}`.
  - `Line` = `{tax, concept, operation, amount Exact, rate *Rate, legalBasis []Citation}`.
  - The **breakdown is the computation trace**: strategies compute only by appending lines
    through a `Ledger` helper, so the explanation cannot drift from the math.
- **IVA is consolidated per taxpayer.** Each strategy reports its IVA *components*:
  - charged and collected;
  - creditable paid (deductible, CFDI, IVA shown separately, paid);
  - withheld from the taxpayer;
  - IVA excluded because it is definitive (625 under LIVA 18-M ⚠️).

  The use case nets the components across the taxpayer's regimes into one monthly IVA
  result (LIVA arts. 1, 1-B, 5-D), shown as payable or *saldo a favor*. Withholdings the
  taxpayer made to suppliers are added to a separate `withholdings_to_remit` total, never
  netted (LIVA 1-A). *(CPA to confirm cross-regime netting for a PF holding 612 + 606.)*
- **Totals rule (M5-22): each amount is counted exactly once.** A PM regime reports three
  disjoint things: its own ISR (`isr`, the ISR segment's last line), its own IVA
  (`ivaComponents`, netted by the use case), and the ISR **and** IVA it withheld from its
  suppliers, already summed into `withheldToRemit` (the `tax: withholding` segment: the IVA
  line, plus the ISR line and one `withholding.to_remit` total when both exist; e.g.
  5,333.33 + 5,000.00 = 10,333.33). So the use case computes
  `totals.withholdings_to_remit = Σ regime.withheldToRemit`,
  `totals.isr = Σ payable regime.isr`, `totals.iva = net(Σ ivaComponents)`, and
  `total_payable = isr + iva + withholdings_to_remit` (contract `EstimateTotals`). It never
  re-adds `withholding` lines, never adds `MonthSums.ISRWithheldToRemit` /
  `IVAWithheldToRemit` again, and never folds `withheldToRemit` into `isr` or `iva`.
- **Registry**: `map[StrategyKey]Strategy`, built at the composition root. A catalog regime
  with no registered strategy falls back to a generic `InformationalStrategy`, which
  returns the catalog's plain-language explanation and no number. So every catalog regime
  works from day one.
- **Adding a regime** = one strategy file + its params + its golden cases + one registry
  line. Compatibility and eligibility rules (which regimes may combine, RESICO exclusions,
  person types) are catalog data checked by the taxpayers module, not by the engine.

### 2. Coverage at launch (from research)

| Key | Treatment | ISR algorithm (research §) | Required tax inputs (`InputKey`) | Options |
|---|---|---|---|---|
| 612 PF | full | YTD cash income − YTD paid deductions − PTU − losses → **accumulated tariff for month n** − prior payments − YTD 10% PM withholding (A.2.3) | `prior_provisional_payments`, `ytd_before_app_income`, `ytd_before_app_deductions` (when started mid-year), `ptu_paid`, `prior_year_losses` (optional) | `iva_fully_taxable` |
| 606 PF | full | Monthly rent − (actual deductions **or** 35% blind + predial) → **art. 96 monthly tariff** − 10% PM withholding (A.2.4) | none | `deduction_mode: blind_35 \| actual`. Quarterly option `pays_quarterly`: **not computed** (table not transcribed) → informational |
| 626 PF | full | Monthly cash income × **non-marginal** rate bracket − 1.25% PM withholding (A.2.5) | none | — |
| 625 PF | full when the definitive option applies, else partial | Withholding by activity (2.1% / 4% / 2.5% in 2026; 20% without RFC) as definitive ISR; directly collected income self-paid at the same rates (A.2.6) | `prior_year_income` (≤ $300k eligibility) | `activities[]: transport_delivery \| lodging \| goods_services`, `definitive_option`, `rfc_given_to_platform` |
| 601 PM | full | **Coeficiente** × YTD accrual income − PTU (May–Dec in equal parts) − losses → × 30% − prior payments − interest withholding (B.2.1). First fiscal year: 0, status "no provisional payments" | `coeficiente_utilidad` (rate, used as given), `prior_provisional_payments`, `ptu_paid`, `prior_year_losses`, `interest_withholding`, `first_fiscal_year` (bool) | `iva_fully_taxable` |
| 626 PM | full | YTD cash income − YTD paid deductions − PTU − losses → × 30% − prior payments (B.2.2) | `prior_provisional_payments`, `ytd_before_app_*`, `ptu_paid`, `prior_year_losses` | `iva_fully_taxable` |
| 621 PF | partial | Bimonthly definitive; the bimonthly table = the February accumulated table; reduction by year in RIF ⚠️ | — | `rif_year` (1–10) |
| 622 PM | partial | As 601, plus the AGAPES exemption and reduction: **deferred**, so it shows IVA only + explanation | as 601 | — |
| 603 PM | partial | IVA only; ISR explained | — | — |
| 605, 607, 608, 610, 611, 614, 615, 616, 620, 623, 624 | informational | Explanation only (Part C) | — | — |

Missing required inputs → `status = incomplete`, the computation goes ahead with zero for
that input, and the `incomplete_inputs` disclaimer is added (US-T4). A `partial` treatment
yields `status = partial` (Amendment 1). `prior_provisional_payments` is one amount **per
month** (`money_monthly`, Amendment 1), not one year-to-date value.

### 3. Versioned parameters: data, not code

- **Source of truth**: YAML in the repo, `internal/taxengine/params/<fiscal_year>/v<N>.yaml`,
  embedded with `go:embed`. Every value is stored **verbatim** from its official source,
  and each file carries `sources[]` (Part D citations).
- **Shape**:
  ```yaml
  fiscal_year: 2026
  version: 1
  published_on: 2026-10-13
  values:                      # each entry may repeat with non-overlapping validity
    - key: iva.rate.border_8
      rate: "0.08"
      valid_from: 2026-01-01
      valid_to:   2026-12-31   # inclusive; decree extension DOF 31-12-2025
      basis: [{law: Decreto, article: "Frontera norte/sur, ext. DOF 31-12-2025"}]
    - key: uma.monthly
      money: "3439.46"         # 2025 UMA (daily 113.14 x 30.4), still in force Jan 2026
      valid_from: 2026-01-01
      valid_to:   2026-01-31
    - key: uma.monthly
      money: "3566.22"
      valid_from: 2026-02-01
      valid_to:   2027-01-31
  tables:
    - key: isr.art106.accumulated
      month: 3                 # one table per month, verbatim from Anexo 8 B.VI
      rows: [{lower: "0.01", upper: "844.59"...}]   # money as decimal strings
    - key: isr.art96.monthly   # Anexo 8 B.V; reused by 606
    - key: isr.resico_pf.monthly   # non-marginal brackets: {upper, rate}
    - key: isr.platform.withholding # by activity
  ```
  The January value is the 2025 monthly UMA (daily $113.14 × 30.4). Check it against the
  INEGI/DOF 2025 publication when the file is written.
- **Resolution** is `Params.At(key, date)`. Each value or table is resolved **by date**,
  never by fiscal year alone:
  - period-level values use the last day of the period;
  - record-level checks (for example "is `rate_8` valid on `collected_on`?") use the
    record date.

  Overlapping validity ranges for one key are rejected at load.
- **Typed, fail-fast loading**: each strategy declares a typed params struct, decoded and
  validated **at startup**. The checks:
  - tariffs are contiguous (`upper[i] + 0.01 == lower[i+1]`, as research verified for all
    twelve tables);
  - all 12 monthly art. 106 tables are present;
  - rates are within [0, 1].

  A bad file fails the boot (and CI), never an estimate.
- **Selection**: use the highest published `version` for the period's fiscal year. If that
  fiscal year is not published yet (for example January, before the RMF), use the latest
  previous year and add the `provisional_parameters` disclaimer.
- **Cross-module use**: the **allowed IVA rates on a date** come from the engine's params
  through a `RateCatalog` port that incomes and expenses consume. A `rate_8` record dated
  2027-01-05 is rejected with `422` once the decree lapses, unless a new param version
  extends it. No date is hard-coded.
- **Why files, not the `tax_parameter_sets` / `tax_tables` tables in MODULES.md (deviation)**:
  - every rate change is a reviewed PR gated by the golden suite (US-X3);
  - the engine stays pure and testable without a database;
  - old versions remain in the repo forever.

  A `ParamsSource` port keeps a DB-backed source possible later.
- **Tariff lookup**: rows sorted by `lower`, binary search (`sort.Search`), O(log k), k = 11.
  RESICO PF uses a **non-marginal** bracket: the rate applies to the whole income
  (PF-626-ISR-02: 25,000.00 → 1.0%, 25,000.01 → 1.1%).

### 4. Money and rounding

- **Storage and API**: `int64` centavos. Rates use a fixed-point `Rate` (scale 10^-6) and
  travel as decimal strings. A `coeficiente_utilidad` is used **as given** by the user
  (art. 14 does not fix its decimals; practice is 4 ⚠️).
- **Inside the engine**:
  - intermediate amounts are `Exact`, an `int64` at **10^-6 peso** (10^-4 centavo) scale.
    Products use a 128-bit intermediate (`math/bits`), so they cannot overflow below about
    9.2 × 10^12 pesos;
  - no floats anywhere (`forbidigo` in CI).
- **Rounding**:
  1. **Result rounding, to the centavo, half away from zero, at the end only.** This
     matches the research worked examples ("rounded to 2 decimals at the end only";
     e.g. 2/3 × 1,600 = 1,066.67).
  2. **Display rounding of lines.** Each breakdown line is shown rounded to the centavo.
     If the shown lines do not add up to the rounded result, the engine emits an explicit
     `Ajuste por redondeo` line, so the breakdown always reconciles visibly. It never
     hides a centavo.
  3. **Payment rounding to whole pesos (CFF art. 20)** is **not** applied by default.
     Research did not cover it. `payable_rounded_cents` is returned equal to
     `amount_cents` until the CPA confirms how SAT's forms apply art. 20. Switching it on
     is a parameter flag (`rounding.payment_whole_pesos`), not code.
- Rounding happens **only** inside the engine (PRD §9). Adapters and the UI format numbers;
  they never round.
- **Dates**: record dates are calendar `DATE`s (no time zone). "Today" and the current
  period come from a `Clock` port in `America/Mexico_City`.
- **Due date**: day 17 of the following month, moved to the next business day using the
  `holidays` table (CFF art. 12 ⚠️ to confirm). The RFC-digit grace days are **not**
  applied (research ⚠️).

### 5. Determinism and golden tests

- Strategies are pure functions of `(Input, Params)`: no `time.Now`, no randomness, no I/O,
  no floats, and lines are emitted in a fixed order. The same inputs produce byte-identical
  JSON.
- A **records fingerprint** (SHA-256 over the canonical aggregated input + param version +
  engine version) goes out with each estimate. M8 uses it to mark months "changed since
  viewed".
- **Estimates are derived, not stored** (US-H1 recomputes). A fingerprint-keyed cache is
  the escape hatch if the dashboard misses its p95 (YAGNI for v1).
- **Golden suite**: `internal/taxengine/testdata/golden/<regime>-<pf|pm>/<case>.yaml`. It is
  seeded with the research worked examples, PF-612-ISR-01..03, PF-612-IVA-01,
  PF-626-ISR-01..02, PF-626-IVA-01, PF-606-ISR-01..02, PF-606-IVA-01, PF-625-01..02,
  PF-621-01 ⚠️, PM-601-ISR-01..02, PM-601-IVA-01 and PM-626-ISR-01, plus:
  - each tariff's boundary rows (`lower`, `upper`, `upper + 0.01`) for every month;
  - one validity-boundary case per dated value (`rate_8` on 2026-12-31 vs. 2027-01-01;
    the UMA in January vs. February).

  CI runs it table-driven as a hard gate at 100% with no update flag. Cases marked ⚠️ are
  `skip`ped with their reason until the CPA signs them off (PRD Q11).
- **Property tests** with a fixed seed:
  - displayed lines + rounding adjustment = result;
  - more income never lowers ISR within a regime;
  - the RESICO PF rate is non-decreasing in income;
  - tariff lookup is total over (0, ∞).

### 6. API surface (`docs/api/openapi.yaml`)

- `TaxpayerEstimate`, `RegimeEstimate`, `TaxResult`, `EstimateTotals` (`iva`, `isr`,
  `withholdings_to_remit_cents`), `Breakdown`, `BreakdownLine`, `Citation`,
  `ParameterSetRef`, `Disclaimer {code, severity, text}` (severity added by Amendment 1).
- `legend` is a response field with `const` "Esto es un aproximado. No todos los gastos
  aplican para tu giro comercial."
- Disclaimer codes:
  - `not_tax_advice`, whose text is the research's mandatory notice: "Estimación
    informativa. No sustituye la declaración ante el SAT ni la asesoría de un contador.
    Cálculo basado en LISR, LIVA, LIF 2026 y Anexo 8 RMF 2026.";
  - `incomplete_inputs`, `provisional_parameters`, `informational_regime`,
    `deductibility_claimed`, `no_coefficient_first_year`.

  The texts are versioned with the param set, so they track the law year.

## Consequences

- (+) OCP: regimes and fiscal years are additive. Strategies are tested with no
  infrastructure. Explanations come from the computation itself.
- (+) Date-ranged values make sunset rules (the border IVA) and mid-year changes (the UMA)
  data-only.
- (−) **A recurring Dec–Jan task**: transcribe the new Anexo 8, LIF rates and decrees into
  `params/<fy+1>/v1.yaml`, then have the CPA review the golden suite. Until then the app
  runs on `provisional_parameters`.
- (−) Each ⚠️ item (RIF reduction, 625 definitive IVA under 18-M, PTU, CFF 20 rounding,
  CFF 12 due-date move) is a known open risk, tracked as a skipped golden case.

## Amendment 1 (M5-01, 2026-10-13)

Decided by the coordinator in [MODULE-tax-engine](../project-specs/MODULE-tax-engine.md) §11
(OQ-T1, OQ-T8) and frozen in the contract 0.6.0 (`docs/api/openapi.yaml`, M5-01). Both
changes are pre-release: the 0.1.0 draft shapes were never frozen and nothing was built on
them.

### A1.1 Status `partial`

`RegimeEstimate.status` is `complete | incomplete | partial | informational`.

- **Why.** The draft only had `complete | incomplete | informational`, so a partial regime
  (621, 622, 603, 625 without the definitive option) had to pretend to be `complete` while
  showing only IVA, or `informational` while showing a number. Both mislead the user.
- **Rule.** `partial` when the strategy's treatment for these options is partial: some taxes
  are estimated, `explanation` says which are not, `isr` is `null` where no ISR figure
  exists. Precedence when several apply: `informational` > `incomplete` > `partial` >
  `complete`; `missing_inputs` is filled whenever inputs are missing, whatever the status.
  The `partial_regime` disclaimer (info) is added when any regime is `partial`.
- **Strategy port.** `Treatment()` stays the catalog-level support level. The per-estimate
  status is computed by the use case from the strategy's result (`partial` vs. full for
  the options, e.g. 625 without `definitive_option`) and the missing inputs.

### A1.2 Monthly prior provisional payments (`money_monthly`)

- **Why.** §2 stored `prior_provisional_payments` as one year-to-date amount per fiscal
  year. A single value is only right for one period: the March estimate needs January +
  February, the June estimate January..May. Re-estimating a past month (allowed, US-H1)
  with a YTD value entered in June would subtract payments made after that month.
- **Rule.** `TaxInputDef.kind` gains `money_monthly` (one `NonNegativeCents` per month 1-12,
  `TaxInputValue.month`) and `boolean` (`value_bool`, for `first_fiscal_year`). For 612 PF,
  626 PM and 601 PM, `prior_provisional_payments` is `money_monthly`: the estimate of month
  n subtracts the sum of months 1..n-1, and the input is **required for period n** only
  for every month m < n in which the regime was active (none for January). A missing month
  makes the regime `incomplete` with `MissingInput {key, label, months[]}` naming the
  months.
- **Required inputs depend on the period.** `RequiredInputs(options, period, facts)`
  replaces `RequiredInputs(options)`: prior payments are not required in January, the
  coeficiente is not required in a 601 first fiscal year, `prior_year_income` is required
  for 625 only with `definitive_option = true`.
- **Prefill, never assume.** The tax-inputs form offers the app's own estimate of each month
  as a suggestion (UX only); the stored value is what the user enters.
- **Data.** The catalog CHECK `regime_required_inputs_valid` accepts the new kinds, and the
  612/626 PM/601 rows move to `money_monthly` (601 gains `first_fiscal_year`), in a forward
  migration with M5-20/M5-21. `taxpayer_tax_inputs` stores one row per `(key, month)`.

### A1.3 Also frozen with the contract (no design change)

- `Disclaimer.severity` (`info | warning | strong`) and the codes `partial_regime`,
  `iva_mixed_activity`, `cpa_review_pending`. Rules a CPA has not validated still run with
  their default, set `cpa_review_pending` / `cpa_topics` on the regime and
  `needs_cpa_review` on the line, and raise the `strong` disclaimer (OQ-T8).
- `BreakdownLine.source` and `needs_cpa_review`; `tax: withholding` for lines to remit.
- `TaxpayerEstimate.records_fingerprint` and `parameter_set` (§5, used by M8).
- `getTaxpayerEstimate.period` is optional (current month, America/Mexico_City); periods
  before 2026 answer `422 parameters_unavailable` (OQ-T9).

## Amendment 2 (M8-03, 2026-10-14)

### A2.1 Figures fingerprint

Every estimate also carries `figures_fingerprint` (contract 0.9.1, required): the SHA-256 of
a canonical text of the figures the user sees, computed by `domain.FiguresFingerprint` next
to `Fingerprint`. Scheme `figures/1`: the due date; the totals' ISR and IVA amounts,
withholdings to remit and total payable; per regime (sorted by key) its key, status,
`cpa_review_pending`, withholdings to remit, ISR amount and IVA components, an absent figure
written `-`. Versions, texts, explanations, disclaimers, breakdown lines and sub-centavo
precision are out. M8 marks a viewed month changed only when this hash moves
(MODULE-history §2.4, OQ-H4). Changing the scheme marks every viewed month: it needs a
further amendment and a one-time re-baseline of `estimate_views` in the same migration. The
golden test `TestFiguresFingerprintGolden` pins it.

### A2.2 Records fingerprint ignores what never changes a figure (#98 FU1)

An informational regime contributes only its key, treatment and active months: its record
sums, options and tax inputs are no longer hashed. A `multi_enum` option is hashed as a
sorted set. Values change once; no view was stored before M8 shipped, so nothing is
re-baselined.

## Engine version log

`domain.EngineVersion` is hashed into the records fingerprint (§5). It changes whenever a
strategy changes a figure for the same input. The contract `examples` stay at `1.0.0`
because they are only illustrations.

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-10-13 | Launch strategy list (§2). |
| 1.0.1 | 2026-10-14 | 626 PM adds the ISR it withheld from suppliers to `withheld_to_remit` (LISR 106 ¶5 / 116 ¶5 / 113-J), as 601 does. Found by the M5-24 property suite (PR #100). |
