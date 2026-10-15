# MODULE-history — M8 History (`module_history`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-14 · MoSCoW: **Should** · Depends on: M2
(taxpayers, active-taxpayer switcher), M5-22 (`Estimator`, incl. the range form added in M6-02),
M5-25/M5-26 (estimate and breakdown screens), P0 (access policy, rate limits)
Sources: PRD §6 (US-H1), §7 ("history stays visible" when read-only) · MODULES.md M8 · PLAN
M8-01..M8-10, D7, #98 FU1 · [MODULE-tax-engine.md](MODULE-tax-engine.md) §2.1, §2.2, §2.10, §2.11,
§7 · [MODULE-dashboard.md](MODULE-dashboard.md) §1.1, §2.2, §2.4 · [MODULE-incomes.md](MODULE-incomes.md)
§2.9 · [MODULE-expenses.md](MODULE-expenses.md) · ADR-0002 §5 (records fingerprint) · contract
scaffold `getHistory` / `markHistoryViewed` / `HistoryYear` / `HistoryMonth` in
`docs/api/openapi.yaml` (not frozen; changes in §9).

## 1. Purpose and scope

For the **active taxpayer**, History answers: **"What did the app estimate for each month of this
fiscal year, and did anything change since I last looked?"**

- A **fiscal-year list of months** (one row per month, January to December, or to the current month
  in the current year), each with the estimated total, ISR, IVA (or saldo a favor), withholdings to
  remit, the due date and the taxpayer-level status.
- A **"Cambió desde que lo viste"** marker on a month whose **displayed figures** changed since the user
  last opened it, because a late income or expense was added/edited/deleted, a tax input changed, or the
  engine or its parameters were corrected (§2.4). A change of inputs or engine that leaves every figure
  identical never marks (DECIDED OQ-H4).
- **Opening a month** shows the M5 estimate screen for that period (per-regime cards, breakdown one tap
  away) and **marks it viewed**.
- Fiscal-year **sum of the monthly estimates** (clearly not an annual return).

Every figure is **recalculated from the current records** on each read (PRD US-H1, ADR-0002 §5:
estimates are derived, never stored). History stores only *what the user saw*, never a tax figure.

| Op | Endpoint | Contract |
|---|---|---|
| getHistory | `GET /v1/taxpayers/{taxpayerId}/history?fiscal_year=YYYY` | scaffold; changes §9 |
| markHistoryViewed | `POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed` | scaffold; changes §9 |

### 1.1 Relation to other screens — summarize and link, never duplicate

| Concern | Dashboard (M6) | History (M8) | Estimate / breakdown (M5) | Ingresos / Gastos (M3/M4) |
|---|---|---|---|---|
| Period | Current month + 12-month trend | One fiscal year, every month | One period, picker | One month, booking date |
| Figures | Hero + chart | Per-month tax totals + status + changed marker | Per regime, every line | Records and their totals |
| Changed marker | No | **Yes** (only place) | No | No |
| Records | Income/expense sums | **No record figures**; links to the lists | No | Yes (owner) |

- History shows **no income/expense sums** (that is the dashboard's chart and the M3/M4 list totals).
  Each month row offers "Ver ingresos" / "Ver gastos", which open the existing M3/M4 list screens on
  that period (hidden when those flags are OFF). Note: the lists select by **booking date**
  (MODULE-incomes §2.9) while the estimate uses collected/paid/invoiced dates; the link is a
  navigation aid, not a reconciliation.
- Opening a month reuses the M5-25 estimate screen with `period` (no copied widgets); the totals row
  reuses `estimate_totals_card.dart` in compact form, like M6.
- The client never recomputes or re-rounds (MODULE-tax-engine §2.7).
- **Withholdings**: a parallel effort is redesigning income/expense withholding derivation
  (counterparty regime). History has **no withholding logic**; it shows whatever `EstimateRange`
  returns. When that work changes an estimate it bumps `EngineVersion`, and History marks the
  affected viewed months (§2.4, reason `engine`).

M8 owns **one table** (`estimate_views`) and **no tax math**. Ports (consumer-side,
`internal/history/usecase`):

| Port | Provider | Used for |
|---|---|---|
| `Estimator.EstimateRange` | M5-22 (range form, M6-02) | Every month of the fiscal year, one aggregate pass |
| `ViewStore` | M8 infra | Read the year's views (one statement); upsert one view |
| `TaxpayerReader` | M2 | Ownership + archived flag (the `404` boundary) |
| `Clock` | platform | Current month in America/Mexico_City |

## 2. Domain rules

### 2.1 Fiscal year and months

- `fiscal_year` required. `< 2026` → `422 business_rule_violation` / `parameters_unavailable` (same as
  M5 §2.10). Future year → `400 validation_failed` / `out_of_range`.
- Months returned: **January .. December** for a past year; **January .. current month** (MX time) for
  the current year. The current month is labelled "Mes en curso" and **never carries the changed
  marker** (it is live; OQ-H3).
- Months **before the taxpayer held any regime** and months with no regime in effect come back with
  status `none` (no amounts), so the list is always contiguous and the gap is explained
  ("Sin régimen en este mes"), never a hole.
- Order: **newest first** (December/current month at the top), matching the M3/M4 lists (OQ-H5).
- Year picker: 2026 .. current year. Only years with published parameters exist (M5 §2.10).

### 2.2 What one month holds

```
HistoryMonth
├── period                       YYYY-MM
├── status                       TaxpayerEstimateStatus (complete|incomplete|partial|informational|none), same rule as M6 §2.2
├── has_records                  bool: any live income or expense with a date in the month
├── tax | null                   null ⇔ status none or all regimes informational
│   ├── isr_payable_rounded_cents, iva_payable_rounded_cents, iva_outcome (payable|in_favor|zero)
│   ├── iva_in_favor_cents, withholdings_to_remit_cents, total_payable_rounded_cents
├── due_date                     from the estimate (M5 §2.10)
├── cpa_review_pending           bool: any regime pending CPA review (drives the strong banner)
├── is_current                   bool
├── records_fingerprint          the current input fingerprint (opaque; echoed on markViewed)
├── figures_fingerprint          the current result-figures hash (opaque; echoed on markViewed; §2.4)
├── engine_version, parameter_set   current ones (opaque to the client)
├── viewed_at | null             last time this user marked it viewed
├── changed_since_viewed         bool (§2.4)
└── change_reason | null         records | engine | null (§2.4)
```

`HistoryYear = { taxpayer_id, fiscal_year, months[], year_totals, legend, disclaimers[] }`.

- `year_totals`: Σ over the returned months of ISR, IVA payable, withholdings to remit and total, labelled
  **"Suma de estimados mensuales"**. It is a sum of provisional estimates, **not** the annual return
  (out of scope, §10). Saldo a favor never subtracts from it (M5 §2.6).
- `legend` (mandatory, server text) and `disclaimers` (the union across months, M5 order); `strong`
  `cpa_review_pending` shows as a non-dismissible banner when any month has it.

### 2.3 Data model — `estimate_views` (migration M8-02)

| Column | Type | Notes |
|---|---|---|
| `user_id` | uuid, FK users ON DELETE CASCADE | RLS key |
| `taxpayer_id` | uuid, FK taxpayers ON DELETE CASCADE | |
| `period` | date (first day of month), CHECK day = 1 | |
| `records_fingerprint` | char(64), CHECK `^[0-9a-f]{64}$` | input fingerprint as seen by the user |
| `figures_fingerprint` | char(64), CHECK `^[0-9a-f]{64}$` | result-figures hash as seen (drives the marker) |
| `engine_version` | text, CHECK length ≤ 32 | as seen |
| `parameter_set` | text, CHECK length ≤ 32 | as seen ("2026/v1") |
| `viewed_at` | timestamptz NOT NULL | server clock |

- PK `(user_id, taxpayer_id, period)`. The PK serves the year read
  (`WHERE user_id = $1 AND taxpayer_id = $2 AND period BETWEEN $3 AND $4`); no extra index.
- **Per user, not per taxpayer**: if shared/delegated access ever exists (G-3), each person keeps their
  own "seen" state.
- RLS: `ENABLE` + `FORCE` from the first migration; SELECT/INSERT/UPDATE/DELETE policies on
  `user_id = current_setting('app.current_user_id')::uuid`; app role is not owner and not BYPASSRLS.
  Isolation tests: user A cannot read, insert, update or delete B's rows, nor insert a row naming B's
  taxpayer (FK + taxpayers RLS) (M8-02).
- Write = one `INSERT … ON CONFLICT (user_id, taxpayer_id, period) DO UPDATE` (idempotent; the last view wins).
- Account deletion and taxpayer deletion cascade (the "estimate history" already listed in the deletion
  transaction, MODULE-auth). No tax amount is ever stored here.

### 2.4 Changed since viewed

Two hashes, both computed server-side and returned with every estimate:

- **`records_fingerprint`** (input side, ADR-0002 §5): records, tax inputs, regimes, parameter set and
  engine version.
- **`figures_fingerprint`** (result side, new): SHA-256 over a canonical text of **the figures the user
  sees** for the month: totals (ISR, IVA payable / in favor, withholdings to remit, total), each regime's
  code, status, ISR, IVA and withholdings, `cpa_review_pending`, and the due date. Excluded: versions,
  texts, explanations, citations, breakdown line order. Computed in the M5 domain next to
  `Fingerprint` (one function, golden-tested) so M5 and M8 can never disagree.

For a month with a stored view `v` and current estimate `e`:

| Condition | `changed_since_viewed` | `change_reason` | Copy |
|---|---|---|---|
| No view yet | false | null | none (never-viewed months are not "changed"; OQ-H1) |
| `v.figures = e.figures` (whatever the inputs or engine did) | false | null | "Visto el 3 oct" (secondary text) |
| figures differ, and `v.engine_version ≠ e.engine_version` or `v.parameter_set ≠ e.parameter_set` | true | `engine` | "Se actualizó el cálculo desde que lo viste" |
| figures differ otherwise | true | `records` | "Cambió desde que lo viste: hubo cambios en tus registros o datos fiscales" |
| current month | false | null | "Mes en curso" |

- **The marker is driven only by `figures_fingerprint`** (DECIDED OQ-H4): an engine or parameter bump,
  or a record edit, that yields identical figures does not mark. `records_fingerprint`,
  `engine_version` and `parameter_set` are stored **only** to explain the reason when the figures did
  change. When both causes apply, `engine` wins (one honest cause, shown once).
- YTD regimes (612, 626 PM, 601) are cumulative: a late January income legitimately changes February..
  current. Every viewed month whose figures moved is marked; that is correct, not noise.
- **False marks must not exist**: with the marker on result figures, #98 FU1 (the input fingerprint
  hashes informational-regime sums and multi-enum order) can no longer cause a mark; it is still fixed in
  M8-03 so `change_reason` stays accurate. A golden test pins that informational-only record changes,
  option reordering and an `EngineVersion`-only bump keep `figures_fingerprint`.
- A change to the `figures_fingerprint` *scheme* (new field shown, new canonical text) would mark every
  viewed month; it requires an ADR-0002 amendment and ships with a one-time re-baseline of stored views
  in the same migration (the user saw the same numbers, only the hash changed).

### 2.5 Mark viewed

- The app calls `markHistoryViewed` when the M5 estimate screen for a past (non-current) period renders
  successfully, **from any entry point** (History row, dashboard "Ver detalle" on a past month), with
  body `{ records_fingerprint, figures_fingerprint, engine_version, parameter_set }` copied from the `TaxpayerEstimate` it
  rendered. Storing *what the user saw* (not a server recompute at POST time) keeps the marker honest if
  records change between render and POST, and keeps the write cheap (no estimate).
- Server validates format (64 lower-hex; version strings ≤ 32 chars, `^[0-9A-Za-z./-]+$`), period range
  (§2.1; the current month is accepted and stored but never marked), taxpayer ownership. A wrong value
  can only mislead the caller's own marker (RLS); it is never trusted for any figure.
- Fire-and-forget on the client: failure never blocks the screen; it retries once on next open.
- Allowed for read-only users and archived taxpayers (it creates no record; no `402`, no cap).

### 2.6 One aggregate pass (performance)

Per `getHistory`, at most:

| # | Read | Notes |
|---|---|---|
| 1 | Taxpayer + regimes profile | ownership `404` here |
| 2 | Incomes `GROUP BY (month, regime)` for the fiscal year | via `EstimateRange` (Jan .. last month) |
| 3 | Expenses, same | idem |
| 4 | Tax inputs for the fiscal year | idem |
| 5 | Regime texts | idem (M5 port) |
| 6 | `estimate_views` for the year | one PK range scan |

- Never 12 × `EstimateMonth`; `EstimateRange` runs the pure engine per month (< 2 ms each).
  `has_records` comes from the same aggregate rows. A statement-count test pins the number (target
  **≤ 7**, measured and amended in M8-03 like M6 §2.4; **≤ 8 since M9-04**: the estimate's due date reads
  the `holidays` tables in one statement, ADR-0007 Amendment 1).
- **Targets**: p95 **< 1 s** for 3 regimes and 10,000 records in the fiscal year; internal goal ≤ 300 ms
  (Go benchmark + EXPLAIN ANALYZE in M8-03, k6 in REL-06). `markHistoryViewed`: one upsert, p95 < 100 ms.
- No cache in v1; `Cache-Control: no-store` on both routes.

### 2.7 Rate limit

Own per-user budgets (ADR-0005): `history_reads` **600/h, burst 600** for `getHistory`;
`history_views` **1,200/h, burst 120** for `markHistoryViewed`. `EstimateRange` inside M8 does not
spend `estimate_reads`. Spent → `429 rate_limited` + `Retry-After`.

## 3. Flags

- `module_history`: default OFF, registered by migration (M8-06); presets `demo_full` includes it
  (`demo_core` / `demo_monetization`: not, OQ-H6).
- **Prerequisite (implied OFF)**: `module_tax_engine` (and through it `module_taxpayers`). M8-04 adds
  `"module_history": {"module_tax_engine"}` to `flags.Prerequisites` (closes PLAN D7). History without
  estimates has nothing to show.
- OFF (or implied OFF) → both routes `404`, indistinguishable from an unknown taxpayer; absent from
  `GET /v1/flags`; the app hides the "Historial" entry and never calls `markHistoryViewed` (the estimate
  screen still works). `estimate_views` rows are kept (not deleted) and still cascade on deletion;
  turning the flag back on resumes the markers.
- `module_incomes` / `module_expenses` OFF: figures are unaffected (the engine still reads records,
  MODULE-incomes OQ-8); only the "Ver ingresos" / "Ver gastos" links hide, and `has_records` counts only
  enabled modules (same rule as M6 `is_empty`).
- Client: one `FlagsProvider` / `FlagGate`. Suites: (`history`, `tax_engine`) ∈ {ON/ON, ON/OFF, OFF/ON},
  plus incomes OFF and expenses OFF.

## 4. User stories (Given / When / Then)

### US-H1 Browse a fiscal year
- **Given** a 612 taxpayer with records Jan–Sep 2026 and today 2026-10-14, **When** I open Historial,
  **Then** I see October ("Mes en curso") down to January, each with total, ISR, IVA or "Saldo a favor",
  withholdings to remit when non-zero, due date and status chip; the legend sits under the list and the
  year totals read "Suma de estimados mensuales".
- **Given** 2026 selected, **Then** the year picker offers only 2026 (no earlier parameters).
- **Given** a month with no records but a YTD regime, **Then** it shows the engine's figure (may be
  non-zero or zero) and "Sin registros en este mes".
- **Given** a month before the taxpayer's first regime, **Then** it shows "Sin régimen en este mes", no amount.

### US-H2 Open a month and mark it viewed
- **When** I tap March, **Then** the M5 estimate screen opens for 2026-03 with the same totals to the
  centavo, and the app posts `markHistoryViewed` with the fingerprint it rendered.
- **When** I return, **Then** March shows "Visto el 14 oct" and no changed marker.

### US-H3 Late income marks the month
- **Given** I viewed March and April, **When** I add an income collected in March, **Then** March and
  (for 612, cumulative ISR) April show "Cambió desde que lo viste" with reason `records`.
- **When** I open March, **Then** its marker clears; April's stays until I open April.
- **Given** I delete that income again (figures return to the seen ones), **Then** the marker clears
  without opening the month.

### US-H4 Engine correction marks the month
- **Given** I viewed March under engine 1.0.1, **When** the server ships 1.0.2 and March's estimate moves,
  **Then** March shows "Se actualizó el cálculo desde que lo viste" (`engine`).
- **Given** a parameter-set change (2026/v1 → v2) that touches March, **Then** the same `engine` reason.
- **Given** an engine or parameter bump that leaves a month's displayed figures identical, **Then** the
  month is **not** marked (only `records_fingerprint` moved; DECIDED OQ-H4).
- **Given** I edit a record in a way that leaves March's figures identical (e.g. a typo in the
  description), **Then** March is not marked.

### US-H5 Never-viewed and current month
- **Given** I never opened June, **Then** June shows no marker. **Given** the current month, **Then** it
  never shows the marker, even after changes.

### US-H6 Links to records
- **When** I tap "Ver ingresos" on March, **Then** the Ingresos list opens on 2026-03 for this taxpayer.
  With `module_incomes` OFF the link is absent.

### US-H7 Read-only, archived, switch taxpayer
- **Given** a read-only user, **Then** History and the estimate screen open and marking viewed works
  (PRD §7). **Given** an archived taxpayer, **Then** History is readable. **When** I switch taxpayer,
  **Then** the list reloads with a skeleton, never the previous taxpayer's figures.

### US-H8 Flags
- **Given** `module_history` OFF, **Then** both routes `404` and "Historial" is hidden.
- **Given** `module_tax_engine` OFF, **Then** `module_history` is implied OFF (routes `404`, absent
  from flags).

### US-H9 Isolation
- **Given** user B, **When** B calls either route with A's taxpayer id, **Then** `404`, and no row of
  A's `estimate_views` is readable or writable by B (RLS test).

### US-H10 Performance
- **Given** 3 regimes and 10,000 records in 2026, **When** `getHistory` runs 1,000 times, **Then** p95 < 1 s
  with the pinned statement count.

## 5. Errors → ErrorCode (single envelope)

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Malformed `fiscal_year` / `period` / id / body, unknown query param or body field | 400 | `validation_failed` | |
| Future `fiscal_year` or `period` | 400 | `validation_failed` | `out_of_range` |
| No / invalid / expired token | 401 | `unauthorized` | |
| Flag OFF (or implied), foreign or unknown taxpayer | 404 | `not_found` | |
| `fiscal_year` / `period` before 2026 | 422 | `business_rule_violation` | `parameters_unavailable` |
| Budget spent | 429 | `rate_limited` | |
| Anything else | 500 | `internal_error` | |

- No `402` (reads and "seen" state are open to read-only users).
- No partial mode: if the estimator fails for any month, `getHistory` fails (`500`); the app shows
  "No se pudo cargar el historial" + "Reintentar".
- Amounts, fingerprints and the RFC are never logged or echoed in errors.

## 6. Accessibility

- Each month row is one focusable target (≥ 44 × 44 px) whose semantics label reads the row in order:
  "Marzo 2026. Total a pagar 2,284 pesos 29 centavos. Incompleto. Cambió desde que lo viste."
- The changed marker is **text + icon**, never colour alone; status chips carry text; saldo a favor in
  words. Contrast: text ≥ 4.5:1, marker/icon ≥ 3:1 (DIRECTIONS contrast script for any new token).
- The list is a real list (semantic headings for the year and "Suma de estimados mensuales"); the
  secondary "Ver ingresos" / "Ver gastos" actions are separate targets, reachable by screen reader and
  keyboard, not swipe-only.
- After returning from a month, focus returns to that row and its new state ("Visto …") is announced
  politely.
- Survives 200 % text scaling (amounts wrap under the month name); no animation with reduce-motion.
- Reading order: taxpayer switcher → year picker → strong CPA banner → year totals → legend → months.

## 7. UX notes (direction A)

- Entry: "Historial" in the app's main navigation for the active taxpayer (exact placement: design,
  M8-05). Year picker as a segmented control / dropdown; default the current year.
- Row: month name + "Mes en curso" tag; right-aligned tabular total; second line ISR · IVA/Saldo a favor;
  status chip; marker line in the `warning`-family token with an icon.
- Formal impersonal copy, es-MX money format. Pull to refresh; refresh on return from a month.
- Empty taxpayer (no records at all): "Aún no hay registros. El historial se llena con tu primer
  ingreso o gasto." + "Registrar ingreso" (if `module_incomes` ON).

## 8. Security (gate M8-09)

- Ownership via the taxpayer read under RLS; foreign id → `404`; `estimate_views` RLS ENABLE + FORCE.
- Client-sent fingerprint/version are format-validated, size-bounded and only ever compared, never
  executed, interpolated or used for a figure; parameterized queries only.
- Server-side flag enforcement; engine OFF ⇒ M8 OFF ⇒ no tax figure leaves the server.
- `Cache-Control: no-store`; no PII beyond the taxpayer id; rate limits §2.7.
- Gate checks: flag matrix, `404` indistinguishability, RLS isolation tests, cascade on deletion.

## 9. Contract changes for M8-01 (architect)

**Applied and frozen in contract 0.9.0 (M8-01).** Item 1: `tax` is the new `HistoryTax`
(the `TrendTax` figures plus `iva_in_favor_cents`, without `status`, which lives on the month);
`change_reason` is the named enum `HistoryChangeReason`; `engine_version` is the named
`EngineVersion` (≤ 32 chars, `^[0-9A-Za-z./-]+$`). Item 2: `year_totals` is `HistoryYearTotals`.
Item 3: `parameter_set` travels as the estimate's `ParameterSetRef` object
(`{fiscal_year, version}`, closed for input as `HistoryViewedParameterSet`), copied verbatim
from the rendered `TaxpayerEstimate`; the `"2026/v1"` text of §2.3 is the storage form only.
Item 5: `TaxpayerEstimate.figures_fingerprint` (`FiguresFingerprint`) is additive and
**optional** until M8-03 computes it (an M5 handler that does not send a required field would
fail its contract tests); M8-03 **will** make it required (tracked in PLAN M8-03, with `TaxpayerEstimate.engine_version` typed as `EngineVersion`, so a value the app echoes always passes `HistoryViewedRequest`) (a response narrowing, non-breaking for
clients). `engine_version` was already required at taxpayer level. The `400`, `404` and `422`
are inline on each operation; `500` reuses `RecordInternalError`.

1. `HistoryMonth`: replace the scaffold (`status` enum with `no_records`, `iva`/`isr` as `TaxResult`)
   with §2.2: `status: TaxpayerEstimateStatus` (frozen enum, reused), `has_records`, `tax: HistoryTax |
   null` (shape aligned with M6 `TrendTax` plus `iva_in_favor_cents`; reuse if the architect prefers),
   `due_date`, `cpa_review_pending`, `is_current`, `records_fingerprint`, `figures_fingerprint`,
   `engine_version`,
   `parameter_set`, `viewed_at` (renamed from `last_viewed_at`), `changed_since_viewed`,
   `change_reason: enum [records, engine] | null`.
2. `HistoryYear`: add `year_totals` and `disclaimers`; `legend` stays required.
3. `markHistoryViewed`: required JSON body `HistoryViewedRequest { records_fingerprint,
   figures_fingerprint (both pattern `^[0-9a-f]{64}$`), engine_version, parameter_set }`, `additionalProperties: false`; keep `204`.
4. Both operations: add `422` (`parameters_unavailable`) and `429`; `getHistory` documents the month
   range rule (§2.1) and newest-first order.
5. `TaxpayerEstimate` already carries `records_fingerprint` and `parameter_set` (M5 §9 item 5); add
   `figures_fingerprint` (additive, M5 domain function, §2.4) and confirm `engine_version` is at
   taxpayer level too (needed for item 3).
6. Examples: a year with one `records` and one `engine` change, a never-viewed month, a `none` month,
   the current month.

## 10. Out of scope

- The **annual return** (declaración anual), annual ISR adjustments, PTU, or any yearly reconciliation;
  `year_totals` is only a sum of monthly estimates. An **annual-estimate** scope is coming as a
  **separate module** (coordinator, 2026-10-14); History v1 stays monthly and does not anticipate it.
- Storing past estimates or "what you declared" (no snapshots; the marker says *that* it changed, not
  by how much). A "before → after" diff is a later Could (OQ-H2).
- Comparing years, CSV/PDF export, a combined all-taxpayers history (PRD Q14).
- Push notifications on change (M9), withholding derivation (the engine owns it).

## 11. Decisions (coordinator, 2026-10-14) — all DECIDED

| # | Question | Decision |
|---|---|---|
| OQ-H1 | Mark never-viewed months as "new"? | **DECIDED:** **No**; only viewed months can be "changed" (avoids 10 markers on first use) |
| OQ-H2 | Show how much a month changed (old → new total)? | **DECIDED:** **No** in v1 (needs stored amounts); later Could |
| OQ-H3 | Changed marker on the current month? | **DECIDED:** **Never**; it is live ("Mes en curso") |
| OQ-H4 | Engine/params bump (or record edit) that leaves a month's figures identical: mark it? | **DECIDED (changed from the proposed default):** **No.** Mark only when the displayed figures change: `figures_fingerprint` stored alongside the input fingerprint (§2.4). Reason label (`engine` vs `records`) kept when they did change |
| OQ-H5 | Order of months | **DECIDED:** **Newest first** |
| OQ-H6 | Which presets include `module_history` | **DECIDED:** **`demo_full` only** |
| OQ-H7 | Mark viewed from any entry point or only from History? | **DECIDED:** **Any entry point** while the flag is ON (the user saw the figure) |
