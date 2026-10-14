# MODULE-dashboard — M6 Dashboard (`module_dashboard`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-14 · MoSCoW: **Must** · Depends on: M2
(taxpayers, active-taxpayer switcher), M3-07 / M4-08 (`LedgerReader`), M5-22 (`Estimator`), P0
(access policy)
Sources: PRD §5, §6 (US-D1), §7, §8, §9, Q14 · MODULES.md M6 · PLAN M6-01..M6-09, FU4, FU5 ·
[MODULE-tax-engine.md](MODULE-tax-engine.md) §1, §2.1–§2.3, §2.10, §7, §8 ·
[MODULE-incomes.md](MODULE-incomes.md) §2.2, OQ-8 · [MODULE-expenses.md](MODULE-expenses.md) ·
`docs/modules/{incomes,expenses,tax-engine}.md` · `docs/design/DIRECTIONS.md` (direction **A ·
Confianza sobria**, DECIDED 2026-10-13) · contract scaffold `getDashboard` / `Dashboard` /
`TrendPoint` / `AccessSummary` in `docs/api/openapi.yaml` (not frozen; changes in §9).

## 1. Purpose and scope

Home is the first screen after sign-in. For the **active taxpayer** it answers three questions
in one glance, read-only:

1. **How much will I probably pay this month?** A summary of this month's estimate (total,
   ISR, IVA or saldo a favor, withholdings to remit, due date, status), with the mandatory
   legend right under it.
2. **How is this month going?** Income collected vs. expenses paid this month.
3. **How has the year gone?** A 12-month trend of incomes, expenses and estimated taxes.

Plus a **trial / access banner** (days and records left, or read-only), an **empty state** for a
taxpayer with no records, and **quick actions** "Registrar ingreso" / "Registrar gasto".

| Op | Endpoint | Contract |
|---|---|---|
| getDashboard | `GET /v1/taxpayers/{taxpayerId}/dashboard?period=YYYY-MM&months=12` | scaffold exists; changes in §9 (M6-01) |

M6 owns **no tables** and **no tax math**. It composes ports:

| Port (consumer-side, `internal/dashboard/usecase`) | Provider | Used for |
|---|---|---|
| `LedgerReader` | M3-07 / M4-08 | Monthly sums for the window, one aggregate pass (§2.4) |
| `Estimator` (range form, §2.4) | M5-22 | Tax figures per month; **not called** when `module_tax_engine` is OFF |
| `access.Policy` | P0-16 | `AccessSummary` for the banner |
| `TaxpayerReader` | M2 | Ownership, archived flag (the `404` boundary) |
| `Clock` | platform | Current month in America/Mexico_City |

### 1.1 Relation to the estimate screen (M5-25) — summarize and link, never duplicate

| Concern | Home (M6) | Estimate screen (M5-25, `/estimado`) |
|---|---|---|
| Period | Current month only, no picker | Period picker (Jan 2026 .. current month) |
| Totals | Total a pagar (hero), ISR, IVA / Saldo a favor, Retenciones a enterar, due date | Same totals card |
| Per regime | **No.** One status chip for the taxpayer (§2.2) | One card per regime with status, explanation, missing inputs |
| "Completa tus datos" | Chip text + link **to the estimate screen** | CTA straight to the tax-inputs form (M5-27) |
| Breakdown, citations | **No** | Link to the breakdown (M5-26) |
| Disclaimers | Legend + the `strong` CPA banner only (§2.3); the rest one tap away | All, in order |

Tapping the hero (or "Ver detalle") opens the estimate screen for the same taxpayer and period.
Home renders the server's `TaxpayerEstimate` as is; the client never recomputes or re-rounds
(MODULE-tax-engine §2.7). Widgets that show the totals are **reused** from M5-25
(`estimate_totals_card.dart`, `estimate_notices.dart`) in a compact variant, not copied.

## 2. Domain rules

### 2.1 Period and window

- `period`: optional; absent → current month in America/Mexico_City (same rule as
  `getTaxpayerEstimate`, MODULE-tax-engine §9 item 1). Home always omits it; M8 and tests may pass it.
  Future month → `400 out_of_range`; before 2026-01 → `422 parameters_unavailable`.
- `months`: 6..12, default **12**. Trend window = `[max(period − (months − 1), 2026-01), period]`,
  oldest first, **one point per month, gaps filled with zeros** (a month with no records is a
  point, not a hole). In October 2026 the default window therefore has 10 points.
- No month before 2026-01 ever appears (records cannot be dated before 2026; no params either).

### 2.2 What the response holds

```
Dashboard(taxpayer, period)
├── period, taxpayer_id
├── is_empty            true ⇔ the taxpayer has no live (non-deleted) income and no live expense, in any period
├── month               { income_cents | null, expense_cents | null }   the `period`; null ⇔ that module OFF (§3)
├── estimate            TaxpayerEstimate | null                null ⇔ module_tax_engine effectively OFF
├── trend[]             TrendPoint per month in the window
│   ├── period, income_cents, expense_cents
│   └── tax | null      { iva_payable_rounded_cents, isr_payable_rounded_cents,
│                         withholdings_to_remit_cents, total_payable_rounded_cents,
│                         iva_outcome (payable|in_favor|zero), status }   null ⇔ engine OFF
├── access              AccessSummary (+ trial_days_left, §2.6)
└── legend              const, present iff estimate != null
```

- **Income** = Σ `amount_cents` (before IVA) of live incomes with `collected_on` in the month
  (cash basis, the same date that drives IVA). Label: "Ingresos cobrados". An income invoiced
  but not collected is not counted (601's accrual ISR still uses `invoiced_on` inside the engine;
  the Home figure is a cash view and says so in its label and in the table).
- **Expense** = Σ `amount_cents` (before IVA) of live expenses with `paid_on` in the month,
  deductible or not. Label: "Gastos pagados".
- Records in **informational regimes** count in income/expense (they are real money) but never
  in tax figures (MODULE-incomes OQ-4).
- **Archived regimes / taxpayer**: data stays readable (MODULE-taxpayers §2.7); see §2.7.
- **Taxpayer-level status** (`trend[].tax.status` and the hero chip), derived from the estimate's
  regimes with the M5 precedence: `informational` (every regime informational) > `incomplete` (any
  regime incomplete) > `partial` (any partial) > `complete`; `none` when the taxpayer holds no
  regime in that month (`regimes: []`). Computed **server-side** in the use case (one rule, unit-tested).
- Saldo a favor never adds to a bar or total: `iva_payable_rounded_cents = 0` and
  `iva_outcome = in_favor` (PRD US-X1; MODULE-tax-engine §2.6).

### 2.3 Mandatory legend and disclaimers near tax figures

- Every place Home shows a tax figure (hero, the tax view of the chart, the data-table view)
  has the legend **"Esto es un aproximado. No todos los gastos aplican para tu giro comercial."**
  in the tinted `legendBg` block **directly under it**, rendered from the server's `legend`
  field, never a client copy, never dismissible, never collapsed (DIRECTIONS "Mandatory legend").
- When `estimate.disclaimers` contains a `strong` item (`cpa_review_pending`), Home shows it as
  a compact, non-dismissible banner above the hero (same widget as M5-25), so the warning travels
  with the figure (MODULE-tax-engine §9 "warnings travel with every figure"). `warning` / `info`
  disclaimers stay on the estimate screen; the hero carries "Ver detalle".
- `estimate == null` (engine OFF): no tax figure anywhere, so no legend and no disclaimer.

### 2.4 One aggregate pass (performance)

The use case never loops 12 × `EstimateMonth`. Per request, at most:

| # | Read | Notes |
|---|---|---|
| 1 | Taxpayer + regimes profile | one read (ownership `404` here) |
| 2 | Incomes `GROUP BY (month, regime)` over `[Jan of fiscal year(window start) .. period]` | one statement; YTD regimes (612, 626 PM, 601) need January of the first fiscal year |
| 3 | Expenses `GROUP BY (month, regime)`, same range | one statement |
| 4 | Tax inputs for the ≤ 2 fiscal years in the window | one statement; skipped when engine OFF |
| 5 | Access summary | one read |
| 6 | `is_empty`: `EXISTS` on each table (or derived from 2/3 when the window already has rows) | index-only |

- M5 exposes a **range form** of the estimator: `EstimateRange(ctx, user, taxpayer, from, to)
  ([]TaxpayerEstimate, error)`, which takes the ledger sums once and runs the pure engine per month
  (< 2 ms each, MODULE-tax-engine §2.10). `EstimateMonth` becomes `EstimateRange(p, p)[0]` so the
  two can never disagree (golden suite runs through both). M5-22 owner adds it in M6-02; FU4
  (`NewEstimator` validates its deps) lands first.
- Income/expense figures come from the **same** aggregate rows the estimator reads (the
  `MonthSums` carry the `collected_on` / `paid_on` sums; if a field is missing the port grows
  additively in M6-02). No extra query for the chart.
- **Targets** (server-side, measured in M6-03 with a Go benchmark and EXPLAIN ANALYZE; re-checked by
  k6 in REL-06): **p95 < 1 s** for 3 regimes, 10,000 records per fiscal year, a window spanning two
  fiscal years. Internal goal ≤ 300 ms (same reads as one estimate plus pure compute). A
  statement-count test pins the ≤ 6 statements above, independent of `months`.
- Indexes: the existing `(user_id, taxpayer_id, collected_on)` / `(…, paid_on)` serve 2/3; M6-03
  confirms with EXPLAIN that no sequential scan appears.
- No cache in v1 (estimates are derived on every read, ADR-0002 §5). `Cache-Control: no-store`.

### 2.5 Rate limit

Own per-user budget `dashboard_reads`, **600 per hour, burst 600** (ADR-0005 mechanism), separate
from `estimate_reads`: the estimator range call inside M6 does **not** spend `estimate_reads`
(that budget guards the M5 routes). Spent → `429 rate_limited` with `Retry-After`.

### 2.6 Trial / access banner

From `access.Policy` (P0-16; PRD §7). Server adds `trial_days_left` so the client never uses its
own clock.

| `access` | Banner | Action |
|---|---|---|
| `trial` | "Prueba gratuita: quedan {d} días · {i} ingresos y {g} gastos" (`d` = whole days left, rounded up, min 1 while the trial is active; `i`/`g` = `incomes_remaining` / `expenses_remaining`) | "Ver planes" → paywall (only when `module_subscriptions` is ON; else no action) |
| `read_only` | "Modo lectura: los datos siguen visibles. Para registrar, contrate un plan." | Same as above |
| `pro`, `comp` | None | — |

- Counts are per user across all taxpayers (PRD Q2); the banner says so in its semantics label.
- Banner is persistent (not a toast), never covers the hero, never blocks scrolling.

### 2.7 Quick actions and edge states

| State | Home shows |
|---|---|
| New user / no records (`is_empty`) | Empty state replaces hero + chart: "Aún no hay ingresos ni gastos registrados. El estimado aparece con el primer registro." Primary "Registrar ingreso", secondary "Registrar gasto". No `$0.00` hero (DIRECTIONS A cons). Trial banner still shown |
| Records, none this month | Hero shows this month's estimate (likely zero, status from §2.2), month card $0 / $0, chart shows the year |
| No regime in effect this month (`status = none`) | Hero replaced by the existing `NoRegimePrompt` (M2); chart still shows past months |
| Only informational regimes | Hero shows the explanation "Este régimen es informativo: la app no estima su impuesto" + link to the estimate screen; no amount |
| Archived taxpayer | Everything readable; quick actions hidden (creates would `422`) |
| Read-only user | Quick actions stay visible; tapping opens the paywall (UX) — server answers `402` regardless |
| `module_incomes` / `module_expenses` OFF | The matching quick action, month figure and chart series are hidden; the server returns them as `null` (§3, OQ-D3). Tax figures are unaffected (the estimator still reads records, MODULE-incomes OQ-8) |

Quick actions open the existing create pages of M3/M4 prefilled with the active taxpayer; on return
Home refreshes.

## 3. Flags

- `module_dashboard`: default OFF, registered in the baseline migration (already present); presets
  `demo_core`, `demo_monetization`, `demo_full` include it.
- **Prerequisite (implied OFF)**: `module_taxpayers` only (the route lives under a taxpayer). M6-05
  adds `"module_dashboard": {"module_taxpayers"}` to `flags.Prerequisites`.
- **`module_tax_engine` is NOT a prerequisite; it degrades Home** (set by the M6 spec brief,
  2026-10-14; DECIDED OQ-D7). With the engine OFF:
  - the server does not call the estimator nor read tax inputs; `estimate = null`, every
    `trend[].tax = null`, `legend` absent. No tax figure leaves the server (security boundary is
    the server, not the client);
  - Home shows the access banner, the month's income vs. expense, the income/expense chart (no
    "Impuestos" view, no view switcher) and quick actions.
  - This **amends** MODULE-tax-engine §7 / US-X4 ("M5 OFF implies M6 OFF"), updated in this PR, and
    the M5-29 QA note asking M6 to declare `module_dashboard → module_tax_engine` (not done, by design).
- **`module_incomes` / `module_expenses` OFF hide their figures** (a disabled module is invisible,
  module-workflow rule; DECIDED OQ-D3). The server does not compute them: `month.income_cents` and
  every `trend[].income_cents` are `null` with incomes OFF (same for expenses), and the dashboard's
  own aggregate for that table is skipped when the engine does not need it either. The app hides the
  month figure, the chart series, "Diferencia" and the quick action. Both OFF: no "Ingresos y gastos"
  view; with the engine also OFF Home shows only the access banner. `is_empty` counts only the
  enabled modules' records; with both OFF it is `false` and no empty state shows.
- `module_dashboard` OFF → `getDashboard` answers `404`; the app keeps today's placeholder Home
  (`features/home/presentation/home_screen.dart`) with its existing flag-gated buttons. Inicio stays
  in the bottom bar.
- Client: one `FlagsProvider` / `FlagGate`; no scattered conditionals. Suites run with
  (`dashboard`, `tax_engine`) ∈ {ON/ON, ON/OFF, OFF/ON}, with `module_incomes` and
  `module_expenses` each OFF, and with `module_taxpayers` OFF.

## 4. User stories (Given / When / Then)

### US-D1 This month's estimate summary
- **Given** an active taxpayer with records in the current month and the engine ON, **When** Home
  opens, **Then** the hero shows Total a pagar (serif, cents smaller), ISR, IVA or "Saldo a favor"
  (`favor` token + text), Retenciones a enterar when non-zero, the due date ("Vence el 17 de
  noviembre"), and the taxpayer status chip; the legend block sits directly under it.
- **Given** any regime `incomplete`, **Then** the chip reads "Incompleto" with "Completa tus datos"
  linking to the estimate screen (not to the form).
- **Given** a `cpa_review_pending` disclaimer, **Then** the strong banner shows above the hero and
  cannot be dismissed.
- **When** the user taps the hero or "Ver detalle", **Then** the estimate screen opens for the same
  taxpayer and month and shows the same totals to the centavo.

### US-D2 Income vs. expense this month
- **Given** incomes collected and expenses paid this month, **Then** Home shows "Ingresos cobrados"
  and "Gastos pagados" (before IVA) and their difference labeled "Diferencia" (not "utilidad": it is
  not a tax base). Invoiced-but-not-collected incomes are not counted.

### US-D3 12-month trend
- **Given** records over several months, **When** Home opens, **Then** one chart shows the window
  of §2.1 with two views in a segmented control: **"Impuestos"** (default; stacked bars, IVA navy at
  the bottom, ISR teal on top, withholdings to remit as a third segment when any month has one) and
  **"Ingresos y gastos"** (paired bars per month). A visible key, the current month's value as text.
- **Given** a month with status `incomplete` or `partial`, **Then** its bar carries a marker and the
  table/semantics say "incompleto" / "parcial".
- **Given** a month with saldo a favor, **Then** its IVA segment is zero and the table says "Saldo a
  favor $X" (never a negative bar).
- **Given** the engine OFF, **Then** only the "Ingresos y gastos" view exists, with no switcher.

### US-D4 Trial banner
- **Given** a trial user with 3 days, 2 incomes and 5 expenses left, **Then** the banner reads
  "Prueba gratuita: quedan 3 días · 2 ingresos y 5 gastos".
- **Given** a read-only user, **Then** the read-only banner shows and every figure stays visible
  (PRD §7 "dashboard stays visible").
- **Given** `pro` or `comp`, **Then** no banner.

### US-D5 Empty state
- **Given** a taxpayer with no live records, **When** Home opens, **Then** the empty state shows
  with "Registrar ingreso" as primary; no hero amount, no chart.
- **Given** the user deletes their only record, **Then** Home shows the empty state again (deleting
  never frees a trial slot; the banner counts stay).

### US-D6 Quick actions
- **When** the user taps "Registrar ingreso" / "Registrar gasto", **Then** the M3/M4 create page
  opens for the active taxpayer; after saving and returning, Home shows the new figures.

### US-D7 Switch taxpayer
- **Given** two taxpayers, **When** the user switches in the header, **Then** Home reloads for the
  new one; no figure from the previous taxpayer is shown while loading (skeleton, not stale data).
  A combined "all my taxpayers" view does not exist (PRD Q14, owner decision).

### US-D8 Flags
- **Given** `module_dashboard` OFF, **Then** `getDashboard` → `404` and the placeholder Home shows.
- **Given** `module_taxpayers` OFF, **Then** `module_dashboard` is implied OFF (route `404`, absent from
  `GET /v1/flags`).
- **Given** `module_tax_engine` OFF and the dashboard ON, **Then** the response has `estimate: null`,
  `trend[].tax: null`, no legend, and the app shows income/expense only.
- **Given** `module_incomes` OFF and the dashboard ON, **Then** `month.income_cents` and every
  `trend[].income_cents` are `null`, and the app shows no income figure, series or "Registrar
  ingreso" (same for `module_expenses` and expenses).

### US-D9 Performance
- **Given** 3 regimes and 10,000 records per fiscal year across two fiscal years, **When**
  `getDashboard` runs 1,000 times, **Then** server p95 < 1 s and ≤ 6 SQL statements per request.

## 5. Errors → ErrorCode (single envelope)

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Malformed `period` / `months` / path id, unknown query param, `months` outside 6..12 | 400 | `validation_failed` | |
| Future `period` | 400 | `validation_failed` | `out_of_range` |
| No / invalid / expired token | 401 | `unauthorized` | |
| Flag OFF (or `module_taxpayers` OFF), foreign or unknown taxpayer | 404 | `not_found` | |
| `period` before 2026-01 | 422 | `business_rule_violation` | `parameters_unavailable` |
| `dashboard_reads` spent | 429 | `rate_limited` | |
| Anything else | 500 | `internal_error` | |

- No `402`: the dashboard is a read and stays open to read-only users.
- Partial failure is not a mode: if the estimator fails, the request fails (`500`); the app shows
  an error state with "Reintentar", never a Home with silently missing tax figures.
- Client states: loading (skeleton), error (message + "Reintentar", keeps last taxpayer header),
  offline (same as error; no cached figures in v1).
- Amounts are never echoed in errors or logs; the RFC is never logged (security.md).

## 6. Accessibility (charts and Home)

- **Screen-reader summary**: the chart has one `Semantics` label stating the same as the visible
  text, e.g. "Impuestos estimados, últimos 10 meses. Octubre: 2,284.29 pesos, incompleto. Mes más
  alto: marzo, 3,100 pesos." The income/expense view states both series for the current month and
  the window totals.
- **Data-table alternative**: "Ver como tabla" toggles a real table (month, Ingresos cobrados,
  Gastos pagados, IVA, ISR, Retenciones, Total, Estado), reachable by screen reader and keyboard,
  with the legend under it. Same data as the chart, from the same response.
- Each bar is focusable/tappable (target ≥ 44 × 44 px, merged per month) and announces its month
  and values; selection shows a value label, not a hover-only tooltip.
- Colour is never the only carrier: key with text labels, status markers with text, saldo a favor
  with words. Marks ≥ 3:1 against the card, chart text ≥ 4.5:1 (DIRECTIONS contrast method);
  any new token (withholdings, income, expense) is checked with the contrast script.
- Layout survives 200 % text scaling (the chart keeps its height; the table wraps); no animation
  when the platform asks to reduce motion.
- Reading order: taxpayer switcher → access banner → strong CPA banner → hero → legend → month card
  → chart (or table) → quick actions.

## 7. UX notes (direction A, Flutter M6-04)

- Hero: `heroBg` navy card, Source Serif 4 amount 44 px, tabular figures, cents smaller; chips per
  status. Legend block under it (`legendBg` / `legendText`).
- Chart: square-topped stacked bars, navy IVA under teal ISR (DIRECTIONS A). DIRECTIONS says six
  months; this spec makes it **12** (owner request; see OQ-D1). One chart only, with the view
  switcher. No pies, no gradients, no lines over bars.
- Formal impersonal copy ("Vence el …", "Registrar ingreso"), es-MX money format `$1,234.56`.
- Pull to refresh; refresh also on return from a create page and on taxpayer switch.

## 8. Security

- Ownership through the taxpayer read under RLS (`app.current_user_id`); foreign id → `404`.
- Server-side flag enforcement is the boundary; engine OFF ⇒ no tax computation at all (§3).
- `Cache-Control: no-store`; no PII in the response beyond the taxpayer id; the RFC is not returned.
- Rate limit §2.5. Security gate M6-08 checks the flag matrix, the `404` indistinguishability and
  that no tax figure leaks with the engine OFF.

## 9. Contract changes for M6-01 (architect)

1. `period`: optional own parameter (not the required `PeriodQuery`); absent → current month (MX).
2. `estimate`: `TaxpayerEstimate | null`; `legend`: not required (present iff `estimate`).
3. Replace top-level `income_cents` / `expense_cents` with `month: { income_cents, expense_cents }`
   (or keep them flat; architect's call, but documented as collected/paid before IVA). Both
   fields, and `TrendPoint.income_cents` / `expense_cents`, become nullable (`null` ⇔ the
   module's flag is effectively OFF, §3).
4. `TrendPoint`: replace `tax_payable_rounded_cents` with `tax: TrendTax | null`, `TrendTax =
   {iva_payable_rounded_cents, isr_payable_rounded_cents, withholdings_to_remit_cents,
   total_payable_rounded_cents, iva_outcome, status}`; `status` enum `complete | incomplete |
   partial | informational | none`.
5. `AccessSummary`: add `trial_days_left: integer | null` (shared with M7; additive).
6. Responses: add `422` (`parameters_unavailable`) and `429` (`rate_limited`); `months` keeps 6..12,
   default 12; describe the window rule of §2.1.
7. Examples: engine ON, engine OFF, empty, trial, read-only.

## 10. Out of scope

- A combined "all my taxpayers" view (PRD Q14, owner decision; Could, later).
- A period picker on Home, year-over-year comparison, history (M8).
- Per-regime cards, breakdown, tax-inputs editing on Home (they live in M5).
- Caching, offline figures, widgets/home-screen extensions, push (M9).
- Profit / "utilidad" figures, forecasting the rest of the year, CSV export.

## 11. Open questions (defaults apply unless the owner says otherwise)

All decided by the coordinator on 2026-10-14.

| # | Question | Decision |
|---|---|---|
| OQ-D1 | DIRECTIONS shows 6 months; the owner asked for a 12-month trend | **DECIDED:** **12** (`months=12`), 6 allowed by the API |
| OQ-D2 | Income on Home: collected (cash) or invoiced? | **DECIDED:** **Collected** (`collected_on`), labeled "Ingresos cobrados" |
| OQ-D3 | With `module_incomes`/`module_expenses` OFF, hide their figures too? | **DECIDED (changed from the proposed default):** **Hide**. The server returns `null` and does not compute them; the app hides the figure, series and quick action (§3) |
| OQ-D4 | Withholdings to remit in the tax bars? | **DECIDED:** **Yes**, third segment, only when non-zero; new token contrast-checked |
| OQ-D5 | Separate rate budget or share `estimate_reads`? | **DECIDED:** **Separate** `dashboard_reads`, 600/h |
| OQ-D6 | Trial days left: server or client clock? | **DECIDED:** **Server** (`trial_days_left`) |
| OQ-D7 | Engine OFF: hide Home (old M5 §7 rule) or show income/expense only? | **DECIDED:** **Income/expense only**; M5 §7 / US-X4 amended in this PR |
