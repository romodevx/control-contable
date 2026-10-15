# MODULE-withholdings — Automatic ISR/IVA withholding derivation (cross-cutting)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-15 · MoSCoW: **Must** · Touches: M3 incomes,
M4 expenses, M5 tax engine, `internal/kernel/fiscal` (ADR-0007)
Sources: research **[withholdings-matrix-2026.md](../research/withholdings-matrix-2026.md)** (hereafter
**WM**; rule IDs `ISR-n` / `IVA-n`, goldens `WH-nn`, ambiguities `An`) · research
[tax-regimes-mx-2026.md](../research/tax-regimes-mx-2026.md) (TR) · [MODULE-incomes.md](MODULE-incomes.md)
§2.6 · [MODULE-expenses.md](MODULE-expenses.md) §2.6 · [MODULE-tax-engine.md](MODULE-tax-engine.md)
§2.6, OQ-T6 · ADR-0002, ADR-0007 · [CPA-REVIEW.md](../reviews/CPA-REVIEW.md) · PLAN §9 (WH tasks).

**Owner decisions (2026-10-15), not open:**
1. Withholdings are **derived automatically** by the server from a few facts about the operation.
   **No manual override**: the user never types an ISR or IVA withholding amount.
2. The same logic runs on the **expense side**: when the user is a persona moral paying a supplier,
   the app derives what the user must withhold and remit, conditions the deduction on it, and
   credits the withheld IVA on the legal timing.
3. Catalog = common + special operations: honorarios, servicios empresariales, arrendamiento
   (residential vs not), uso o goce de bienes muebles, enajenación de bienes, autotransporte de
   bienes (IVA 4 %), comisiones, desperdicios y reciclables (IVA 100 %), plataformas, extranjero / sin
   RFC.
4. **Subcontratación 6 % is excluded.** LIVA art. 1-A fr. IV (IVA withholding of 6 % on personnel
   services) was repealed by the DOF of 23-04-2021 (WM F6, ✅ read verbatim). Since then a
   specialized service between companies has no IVA withholding unless another case of 1-A applies
   (freight, desperdicios). The app does not offer "subcontratación" as an operation: it would show
   a withholding that no longer exists (WM WH-07).

## 1. Scope

| In | Out (v1) |
|---|---|
| Derive ISR and IVA withholding per income and per expense from the decision table (§3) | CFDI XML import (the fields of WM §3 are listed for that later feature) |
| Store the derived amounts on the record; show them read-only with their basis | Any manual entry or override of a withholding amount (owner decision 1) |
| Emisor effects on the monthly estimate (§4.1) | Payments to foreign residents (LISR 156/158 25 %, LIVA 1-A fr. III 100 %): shown as "no soportado" (WM A13) |
| Receptor effects: amounts to remit, deductibility condition, withheld IVA credited the month after remittance (§4.2, **engine change**) | Interest (LISR 54/135), servidumbres (RMF 2.7.3.7), AGAPES-exempt RESICO legend (RMF 3.13.26): no operation kind for them |
| Migration of existing records from manual to derived (§5) | Exports at IVA 0 % to foreign customers (LIVA 29): the user keeps choosing the IVA rate |
| ⚠️ warnings for every rule the CPA has not confirmed (§6) | Government as a **supplier** (it is treated as a persona moral: no withholding either way) |

Feature flag: **none of its own.** The change ships inside `module_incomes`, `module_expenses` and
`module_tax_engine`, which are OFF in every shared environment. A second flag would keep the manual
path alive next to the derived one, which the owner removed (OQ-W6).

## 2. What the forms ask (minimal questions)

The server already knows **who the user is** (person type and regime of the record: the taxpayer
profile and `regime_code`). The forms ask only what the CFDI would tell us. Each question appears
**only when its answer can change a withholding**; otherwise its default is stored silently.

### 2.1 Income form ("Nuevo ingreso"), every taxpayer

| # | Field (API) | Question (es-MX) | Options (es-MX label → value) | Shown when | Default |
|---|---|---|---|---|---|
| Q1 | `counterparty_kind` | **¿A quién le facturaste?** | "Una empresa (persona moral)" → `pm` · "Una persona con negocio o profesión (persona física con actividad)" → `pf_with_activity` · "Una persona sin negocio o público en general" → `pf_no_activity_or_public` · "Gobierno federal (o un organismo federal)" → `gov_federal` · "Gobierno estatal o municipal" → `gov_state_municipal` · "Alguien en el extranjero" → `foreign` | always (it replaces "¿Quién te pagó?") | the last value used **for this regime**; first time: none (the user must choose; OQ-W2) |
| Q2 | `operation_kind` | **¿Qué le vendiste o qué servicio diste?** | "Servicios profesionales (honorarios): asesoría, consultoría, diseño, salud…" → `professional_services` · "Otros servicios de tu negocio: reparación, mantenimiento, limpieza, comida…" → `business_services` · "Renta de un inmueble (casa, local, oficina, terreno)" → `rent_real_estate` · "Renta de equipo, maquinaria o vehículos" → `rent_movable` · "Venta de productos o mercancía" → `sale_of_goods` · "Flete: transporte de mercancía por carretera" → `ground_freight` · "Comisiones por intermediación" → `commission` · "Desperdicios o reciclables (chatarra, cartón, PET…)" → `scrap` · "Otro" → `other` | Q1 ∈ {`pm`, `gov_federal`, `gov_state_municipal`} and Q4 ≠ yes | last used for this regime; first time by regime: 612 → `professional_services`, 606 → `rent_real_estate`, 626 PF → `professional_services`, 621 → `sale_of_goods`, any PM → `business_services` |
| Q3 | `residential_rent` | **¿Es una casa o departamento para vivir?** Helper: "La renta de vivienda no lleva IVA." | Sí / No | Q2 = `rent_real_estate` (also when Q2 is hidden but defaulted to rent, i.e. 606) | No. **Sí forces `iva_rate = exempt`** (LIVA 20 fr. II) and hides the IVA rate |
| Q4 | `via_platform` | **¿Lo cobraste a través de una plataforma digital (app de transporte, hospedaje o ventas)?** | Sí / No | regime 625; for PM regimes under "Más detalles" (TR §B.2.3) | 625: Sí · PM: No |

- Q1 is **not** followed by the counterparty's regime on incomes. WM §1.1 shows every persona
  moral withholds alike (601, 626 PM, 603…: ✅) and a persona física never withholds (F1, F2), so
  the payer's regime cannot change the result. Asking it would add a question with no effect
  (OQ-W1 records the departure from the brief).
- `platform_activity` and `rfc_given_to_platform` keep coming from MODULE-incomes §2.1 and the 625
  profile; they are not asked again.
- Q4 = Sí makes the operation `platform` internally (Q2 hidden).
- "Guardar y crear otro" keeps Q1–Q4 (they are context fields), clears the amounts.

### 2.2 Expense form ("Nuevo gasto"), **personas morales only**

A persona física never withholds on domestic purchases (WM F1, F2, row R-PFA), so for PF users the
section does not exist and nothing is stored but `supplier_kind = null`. For PM users a section
"Tu proveedor" appears right after the category:

| # | Field (API) | Question (es-MX) | Options | Shown when | Default |
|---|---|---|---|---|---|
| E1 | `supplier_kind` | **¿Quién te vendió o te dio el servicio?** | "Una empresa (persona moral)" → `pm` · "Una persona física con negocio o profesión" → `pf_with_activity` · "Alguien en el extranjero" → `foreign` | always (PM) | last used for this category; first time: `pm`, except categories `servicios_profesionales`, `renta_local`, `fletes` (no preselection: the user must choose) |
| E2 | `supplier_regime` | **¿En qué régimen está tu proveedor?** Helper: "Viene en su factura, junto a su RFC." | "Actividad empresarial o profesional (612)" · "Arrendamiento (606)" · "RESICO (626)" · "Plataformas digitales (625)" · "RIF (621)" · "No sé" | E1 = `pf_with_activity` | last used for this category; else "No sé" |
| E3 | `operation_kind` | **¿Qué te vendió o qué servicio te dio?** | same list as Q2 (Q2 labels in second person: "Te dio servicios profesionales…") | E1 ∈ {`pm`, `pf_with_activity`} | from the category (table below), editable |
| E4 | `residential_rent` | **¿Es una casa o departamento para vivir?** | Sí / No | E3 = `rent_real_estate` | No; Sí forces `iva_rate = exempt` |

- "No sé" on E2 derives as **612** (the most common PF regime billing a company) and adds the
  warning `supplier_regime_unknown` ("Revisa el régimen en la factura: cambia la retención"). The
  difference matters: 612 fees withhold 10 % ISR, RESICO 1.25 %.
- E1 = `foreign` derives zero and adds `foreign_counterparty_not_supported` (§6).
- No question about the supplier being a persona moral's regime: every PM emisor behaves alike
  (WM §1.2 row "PM-any").

Category → default `operation_kind` (seed data in `category_regime_rules`, not code):
`servicios_profesionales` → `professional_services`; `renta_local` → `rent_real_estate`;
`arrendamiento_equipo` (if present) → `rent_movable`; `fletes` → `ground_freight`;
`comisiones` (if present) → `commission`; `mercancia`/inventory categories → `sale_of_goods`;
every other category → `business_services`. M4's seed may lack some codes; WH-03 adds the column and
fills what exists, the CPA reviews it with the catalog. `listExpenseCategories` exposes it as `ExpenseCategory.default_operation_kind`
(contract 0.11.1): the app reads it from there and never keeps its own copy (WH-14).

### 2.3 The derived result on screen (both forms)

A read-only card replaces the old "Retenciones" inputs and updates as the user types (debounced
`previewWithholdings`, §7):

- Income: "**Te retienen**: ISR $1,250.00 · IVA $10,666.67 — **Recibes** $104,083.33".
  Below, one line per rule in plain words: "La empresa te retiene 1.25 % de ISR porque estás en
  RESICO (LISR art. 113-J)" · "y 2/3 del IVA por servicios de persona física (LIVA art. 1-A,
  RLIVA art. 3)". When nothing is withheld: "Nadie te retiene impuestos en esta operación."
- Expense (PM): "**Debes retener**: ISR $4,000.00 · IVA $4,266.67 — **Pagas a tu proveedor**
  $38,133.33 · **Entregas al SAT** $8,266.67 a más tardar el 17 de {mes siguiente}." Plus the
  deduction note (§4.2).
- Each ⚠️ rule shows the non-blocking warning of §6 under the card.
- No field of the card is editable. A help link "¿Tu factura dice otra cantidad?" explains: "La app
  calcula la retención que marca la ley para lo que nos dijiste. Si tu factura muestra otra, revisa
  con quién te facturó o con tu contador; la estimación usa la de la ley." (consequence of owner
  decision 1; risk recorded in OQ-W3).

## 3. Decision table (derivation)

Implemented once, in **`internal/kernel/fiscal`** (`fiscal.DeriveWithholdings`), pure, table-driven,
used by M3, M4 and the preview endpoint. Rates come from `RateCatalog` on the record's booking
date (dated parameters: LIF/RMF rates change yearly); the fractions 2/3, 1/2, 1 are the kernel
`Share` constants. Every output carries the **rule IDs of WM §4.2**.

### 3.1 Inputs

| Input | Income (user = emisor) | Expense (user = receptor, PM only) |
|---|---|---|
| `EmisorPersona`, `EmisorRegime` | taxpayer person type, `regime_code` | `supplier_kind` (`pm` → PM-any; `pf_with_activity` → PF) and `supplier_regime` ("No sé" → 612) |
| `ReceptorKind` | `counterparty_kind` | always `pm` (the user) |
| `OperationKind` | `operation_kind`, or `platform` when `via_platform` | `operation_kind` |
| `Subtotal` | `amount_cents` | `amount_cents` |
| `IVA` | recorded `iva_cents` (the base of the IVA fractions) | recorded `iva_cents` |
| `ResidentialRent` | `residential_rent` | `residential_rent` |
| 625 extras | `platform_activity`, `rfc_given_to_platform` | — |

### 3.2 Rules (ISR and IVA evaluated independently; first match wins)

| Inputs (summary) | ISR rate | IVA withheld | Withholder | WM rule | Status |
|---|---|---|---|---|---|
| Receptor is PF (with or without activity) or público en general | 0 | 0 | — | ISR-0, IVA-0/IVA-1 | ✅ |
| Receptor is foreign | 0 | 0 | — | ISR-1, IVA-0 | ✅ (export IVA out of scope) |
| Operation `platform`, emisor 625 PF | by activity: 2.5 % goods/services, 2.1 % transport/delivery, 4 % lodging; **20 %** without RFC | 50 % of IVA; **100 %** without RFC | platform | ISR-3, IVA-3 | ✅ |
| Operation `platform`, emisor PM | 2.5 % | 0 | platform | ISR-2 note, TR §B.2.3 | ✅ |
| Emisor PM (any regime), receptor PM or federal, `ground_freight` | 0 | **4 % × Subtotal** | receptor | ISR-2, IVA-4 | ✅ |
| Emisor PM, receptor PM, `scrap` | 0 | **100 % of IVA** | receptor | ISR-2, IVA-5 | ✅ |
| Emisor PM, any other operation | 0 | 0 | — | ISR-2, IVA-8 | ✅ (incl. the old 6 % case, WH-07) |
| Emisor 626 PF, receptor PM, any non-platform operation | **1.25 %** | by operation (rows below) | receptor | ISR-4 | ✅ |
| Emisor 612, `professional_services`, receptor PM | **10 %** | 2/3 of IVA | receptor | ISR-5, IVA-6 | ✅ |
| Emisor 606 or 612, `rent_real_estate`, receptor PM | **10 %** (626: 1.25 % ⚠️ A3) | 2/3 of IVA; **0 if residential** (exempt) | receptor | ISR-6, IVA-6 | ✅ (626 ⚠️) |
| Emisor PF, `rent_movable`, receptor PM | 0 ⚠️ A10 (626: 1.25 %) | 2/3 of IVA | receptor | ISR-7, IVA-6 | ⚠️ |
| Emisor PF, `commission`, receptor PM | 0 ⚠️ A7 (626: 1.25 %) | 2/3 of IVA | receptor | ISR-7, IVA-6 | ⚠️ |
| Emisor PF, `ground_freight`, receptor PM or federal | 0 (626: 1.25 %) | 4 % × Subtotal | receptor | ISR-7, IVA-4 | ✅ |
| Emisor PF, `scrap`, receptor PM | 0 (626: 1.25 %) | 100 % of IVA | receptor | IVA-5 | ✅ |
| Emisor PF, `sale_of_goods` / `business_services` / `other`, receptor PM | 0 (626: 1.25 %) | 0 | receptor (626 ISR) / — | ISR-7, IVA-8 (F4, F5) | ✅ |
| Emisor 625 PF **not** via platform, receptor PM | 0 ⚠️ A6 | by operation, as any PF (rows above): 2/3 of IVA on `professional_services`, `rent_real_estate`, `rent_movable`, `commission` (LIVA 1-A fr. II a, d); 0 on goods and business services | receptor (IVA) / — | ISR-7 (A6), IVA-6 or IVA-8 | ISR ⚠️ / IVA ✅ |
| Receptor `gov_federal`: same ISR as receptor PM | as PM | PF emisor: 2/3 of IVA on professional services, and on any other operation only when Subtotal > $2,000; freight 4 %; PM emisor: 0 except freight | receptor | IVA-4, IVA-7, A1, A2 | ⚠️ (except freight ✅) |
| Receptor `gov_state_municipal`: same ISR as receptor PM | as PM | **0** | receptor (ISR) | IVA-2, A1 | ISR ⚠️ / IVA ✅ |

Rules shared by every row:
- `iva_rate ∈ {rate_0, exempt}` or residential rent → IVA withheld = 0 (there is no IVA).
- With the **8 % border rate**, 2/3 is taken of the IVA actually charged (5.3333 % of the price) ⚠️
  A8; freight stays 4 % of the price.
- `operation_kind = other` never withholds IVA; ISR only through the 626 rule.
- Amounts: `ISR = round(Subtotal × rate)`; `IVA = round(IVA × share)` or `round(Subtotal × 4 %)`;
  one rounding to the centavo, half away from zero (kernel `Share.Suggest`, ADR-0002 §4; WM A5 ⚠️).
- Invariants (property tests): `0 ≤ ISR ≤ Subtotal`, `0 ≤ IVA withheld ≤ IVA`,
  `net = Subtotal + IVA − ISR − IVA withheld ≥ 0`, output independent of input order and clock.

### 3.3 Output (stored on the record and served read-only)

```
withholding: {
  isr_cents, iva_cents,            // stored in the existing isr_withheld_cents / iva_withheld_cents
  isr_rate, iva_basis,             // e.g. "0.0125"; "two_thirds_of_iva" | "half_of_iva" | "all_of_iva" | "four_percent_of_price" | "none"
  withholder,                      // "counterparty" (income) | "you" (expense) | "platform" | "none"
  net_cents,                       // what changes hands
  rule_ids: ["ISR-4","IVA-6"],     // WM §4.2
  table_version,                   // "2026.1": bumps on any rule or rate change (§5.3)
  cpa_unverified: bool             // any ⚠️ rule matched
}
```

## 4. Effects on the monthly estimate (M5)

### 4.1 User as **emisor** (incomes): no formula changes, the inputs are now derived

The engine keeps reading the recorded `isr_withheld_cents` / `iva_withheld_cents` through
`LedgerReader`; only their origin changes (derived instead of typed). MODULE-tax-engine §2.4 and §2.6
already apply:

| Regime | Effect | Basis |
|---|---|---|
| 612 PF | YTD ISR − Σ ISR withheld by PM Jan..n | LISR 106 ¶5, WM §2.1 ✅ |
| 606 PF | monthly ISR − ISR withheld in n; negative → in favor | LISR 116 ¶5 ✅ |
| 626 PF | monthly ISR − 1.25 % withheld in n. **A negative month shows 0 to pay** and the excess as "a favor en tu declaración anual" ⚠️ A4 (WH-09 verifies the current strategy and aligns it) | LISR 113-J ¶2, 113-F |
| 625 PF | platform withholding provisional or definitive (unchanged) | LISR 113-A, 113-B ✅ |
| PM | platform 2.5 % withheld credits ISR (unchanged) | TR §B.2.3 |
| IVA (all) | **IVA a cargo = trasladado cobrado − IVA retenido al contribuyente en el mes − acreditable**, monthly, not YTD | LIVA 1-A, 5-D ✅ |

Withholding month = the income's IVA date (`collected_on`): the payer withholds when it pays
(WM §2.1 "Which month"). Accrual incomes still pending collection carry no withholding in the
estimate until collected, as today.

### 4.2 User as **receptor** (expenses of a persona moral): **required engine change**

| Item | Rule | Change? |
|---|---|---|
| Withholdings to remit | `withholdings_to_remit` = Σ ISR + Σ IVA withheld on expenses paid in n, due day 17 of n+1, never netted (RMF 4.1.11) | unchanged formula, derived inputs |
| Deductibility | The expense is deductible **only if** the withholding is withheld and remitted (LISR 27 fr. V; 210 for RESICO PM; 147 fr. VII). v1 assumes timely remittance and keeps the category's deductibility, but every expense with a withholding > 0 carries the warning `deduction_requires_remittance` ("Este gasto solo es deducible si enteras las retenciones a más tardar el 17 de {mes}") | **new warning** |
| IVA acreditable, paid part | `creditable_paid(n)` = creditable share of **(IVA − IVA withheld)** on expenses paid in n | **changed**: today the full `iva_cents` is credited in n |
| IVA acreditable, withheld part | `creditable_withheld(n)` = creditable share of the IVA **withheld in n − 1**: it becomes creditable "en la declaración de pago mensual siguiente a la declaración en la que se haya efectuado el entero de la retención" (LIVA 5 fr. IV, ✅ verbatim, WM §2.2). Withheld in June, remitted with the June return (by 17 July) → credited in the **July** return | **new** |

**This corrects the current behavior.** Today `expenses.creditable_iva_cents` equals the full IVA
of a creditable expense, so the engine credits the withheld IVA **in the same month it is
withheld**, which is what TR §B.3 said ("creditable in the month it is paid") and what OQ-T6 (b)
believed it was not doing. The law defers that part one month. The change is therefore an engine
behavior change and requires:

1. Ledger: `ExpenseMonthSums` splits `CreditableIVAPaid` into `CreditableIVAPaid` (IVA − withheld)
   and `CreditableIVAWithheld`; the expense row stores both creditable parts (§5.1).
2. The fiscal-year read extends **one month back** so January credits the previous December's
   withheld IVA (crosses the fiscal year: LIVA is monthly, not annual). No record in n − 1 → 0.
3. New IVA breakdown line `iva.creditable.withheld_prior_month` (op `subtract`, source records,
   basis LIVA 5 fr. IV) with the assumption shown as text: "Supone que enteraste la retención del
   mes anterior a tiempo." ⚠️ CPA: timing assumption (§6).
4. `domain.EngineVersion` **1.0.1 → 1.1.0** (a behavior change: estimates and
   `records_fingerprint` change for PM users with withholdings; M8 shows them as "changed", which is
   correct). Contract examples updated with the next re-freeze.
5. Goldens: new **WH-01-EST** (emisor estimate, §8 US-W1), **WH-11-JUN** and **WH-11-JUL**
   (receptor split), **PM-601-IVA-02** (July of PM-601-IVA-01: credits the 5,333.33 withheld in
   June). PM-601-IVA-01 keeps its figures: its "creditable IVA paid 200,000" is read as the paid
   part (note added in the golden's header).
6. CPA-REVIEW OQ-T6 item reworded (it is now law read verbatim; only the timing assumption and the
   proportional credit remain for the CPA).

## 5. Migration of existing records

### 5.1 Schema (versioned migration, expand → backfill → contract)

| Table | Add | Change |
|---|---|---|
| `incomes` | `counterparty_kind text` (enum Q1), `operation_kind text` (enum Q2 + `platform`), `residential_rent boolean NOT NULL DEFAULT false`, `via_platform boolean NOT NULL DEFAULT false`, `withholding_rule_ids text[] NOT NULL DEFAULT '{}'`, `withholding_table_version text`, `withholding_cpa_unverified boolean NOT NULL DEFAULT false` | `counterparty_type` dropped in the contract step; the CHECK `counterparty_type IN ('persona_moral','plataforma') OR withheld = 0` is replaced by bounds only (the domain owns applicability). `isr_withheld_cents` / `iva_withheld_cents` stay, written only by the server |
| `expenses` | `supplier_kind text NULL` (NULL for PF users), `supplier_regime text NULL`, `operation_kind text NULL`, `residential_rent boolean NOT NULL DEFAULT false`, `creditable_iva_withheld_cents bigint NOT NULL DEFAULT 0`, the same three `withholding_*` columns | `creditable_iva_cents` now holds the **paid** part; CHECK `creditable_iva_cents + creditable_iva_withheld_cents ≤ iva_cents` |

RLS is unchanged (same tables, same policies); grants add the new columns to `cc_app` INSERT/UPDATE
lists. New constraints on populated tables use `NOT VALID` + `VALIDATE` (#91 note 2). Isolation tests
re-run.

### 5.2 Data (dev and any local database; no shared environment has these modules ON)

A one-shot, idempotent Go command `cmd/backfill-withholdings` (dry-run by default) runs the kernel
derivation over every live record, so the SQL never duplicates the decision table:

| Old value | New value |
|---|---|
| income `counterparty_type = persona_moral` | `counterparty_kind = pm` |
| `persona_fisica` | `pf_with_activity` (zero withholding either way) |
| `publico_general` | `pf_no_activity_or_public` |
| `plataforma` | `counterparty_kind = pm`, `via_platform = true` |
| income `operation_kind` (new) | regime default of §2.1 Q2; 606 with `iva_rate = exempt` → `residential_rent = true` |
| expense of a **PF** user | supplier fields NULL; withholdings forced to 0 (they were already 0: `withholding_not_applicable`) |
| expense of a **PM** user with a withholding > 0 | `supplier_kind = pf_with_activity`, `supplier_regime = 612`, `operation_kind` from the category |
| expense of a PM user without withholding | `supplier_kind = pm`, `operation_kind` from the category (`fletes` then derives 4 %: reported) |
| both | derived amounts overwrite the typed ones; `creditable_iva_cents` re-split |

- The dry run prints, per table, the count of rows whose derived amounts differ from the stored ones,
  and the ids (no amounts, no descriptions: they may hold names). The operator applies it only after
  reading that report.
- Runs as the migration owner role inside one transaction per user (`SET LOCAL app.current_user_id`),
  so RLS still scopes each batch; aborts on the first error.
- Acceptable outcome for dev data: a `/qa` reset (rules/qa-browser.md) is an equal alternative; the
  command exists because §5.3 needs it anyway.

### 5.3 Re-derivation when the table changes

A CPA answer or a new year's rates bumps `table_version`. The same command re-derives rows whose
`withholding_table_version` is older, which is why it is a maintained tool and not a throwaway
script. The docs runbook (WH-16) describes it next to the yearly parameter update.

## 6. Warnings (non-blocking, append-only codes on `Income.warnings` / `Expense.warnings`)

| Code | When | Text (es-MX) |
|---|---|---|
| `withholding_cpa_unverified` | any ⚠️ rule matched (A1, A2, A3, A4, A6, A7, A8, A10) | "Esta retención aún no la confirma un contador. Úsala como aproximado." |
| `foreign_counterparty_not_supported` | Q1 or E1 = `foreign` | "Las operaciones con el extranjero pueden llevar otros impuestos que la app todavía no calcula." |
| `supplier_regime_unknown` | E2 = "No sé" | "Revisa el régimen en la factura de tu proveedor: cambia la retención." |
| `deduction_requires_remittance` | expense with a derived withholding > 0 | "Este gasto solo es deducible si enteras las retenciones a más tardar el 17 de {mes}." |
| `withholding_differs_from_rate` | **retired**: never emitted again (there is no typed amount to compare); stays in the enum (append-only) | — |

The estimate adds `needs_cpa_review` on the lines fed by records with `withholding_cpa_unverified`
(topic `withholding_rule_unverified`, appended to `CpaTopic`) and on `iva.creditable.withheld_prior_month`
(topic `withheld_iva_credit_timing`).

## 7. Contract changes (architect, WH-02; **breaking** → `0.8.0 → 0.9.0` + ADR-0008)

> **As implemented by WH-02 ([ADR-0008](../adr/0008-withholdings-are-derived.md)):** the version
> is **0.10.0** (0.9.0 went to the M8 history contract). The break is applied in two steps: 0.10.0
> adds the new fields and deprecates the old ones, and WH-06 / WH-07 remove them (0.10.x).
> Item 4 is **two operations**, `previewIncomeWithholdings` and `previewExpenseWithholdings`
> (one per module flag, with a shared `withholding_previews` budget), not one `side=`
> parameter.

1. `IncomeCreate`/`IncomeUpdate`: **remove** `isr_withheld_cents`, `iva_withheld_cents` (a client
   sending them gets `400 additionalProperties`); replace `counterparty_type` with
   `counterparty_kind` (required); add `operation_kind`, `residential_rent`, `via_platform`.
2. `ExpenseCreate`/`ExpenseUpdate`: remove the two withholding fields; add `supplier_kind`,
   `supplier_regime`, `operation_kind`, `residential_rent` (PM: `supplier_kind` required; PF: any
   value → `422 business_rule_violation` `supplier_not_applicable`).
3. `Income`, `Expense`: add the read-only `withholding` object (§3.3); keep `isr_withheld_cents`,
   `iva_withheld_cents` as read-only mirrors for M5/M6 consumers.
4. New `previewWithholdings` — `GET /v1/taxpayers/{taxpayerId}/withholdings/preview?side=income|expense&regime=&on=&...`
   returns the §3.3 object without storing anything; rate-limited like the other reads; same
   validation as create. `getIncomeDefaults` drops `isr_withholding_rate`/`iva_withholding_rate`/
   `*_allowed` (now meaningless) and adds `last_used` defaults for Q1–Q4.
5. `withholding_not_applicable` and `withholding_exceeds_base` detail codes are no longer produced by
   incomes/expenses (append-only enum: kept, documented as retired).
6. `CpaTopic` += `withholding_rule_unverified`, `withheld_iva_credit_timing`; `Warning.code` += §6.
7. ADR-0008 "Withholdings are derived, never entered" records owner decision 1, the kernel
   placement, the stored-and-versioned derivation (vs. derive-on-read) and the 0.9.0 break.

## 8. User stories (Given / When / Then)

### US-W1 Owner's case: RESICO PF bills a company (WM WH-01)
- **Given** a persona física in **626 RESICO**, `module_incomes` and `module_tax_engine` ON, and no
  other record in July 2026,
- **When** I record an income of **$100,000.00** (sin IVA), IVA 16 %, collected 2026-07-10,
  "¿A quién le facturaste?" = "Una empresa (persona moral)", "¿Qué le vendiste…?" = "Servicios
  profesionales",
- **Then** the card shows IVA **$16,000.00**, "Te retienen: ISR **$1,250.00** · IVA
  **$10,666.67**", "Recibes **$104,083.33**", `rule_ids = [ISR-4, IVA-6]`, no warning;
- **And** the July estimate shows RESICO ISR $100,000 × 2.00 % = $2,000.00 − $1,250.00 retenido =
  **$750.00** to pay, and IVA $16,000.00 − $0.00 acreditable − $10,666.67 retenido =
  **$5,333.33** to pay (golden WH-01-EST).

### US-W2 No one withholds
- **Given** a 612 PF, **when** I bill $5,000.00 + IVA to "Una persona sin negocio o público en
  general" (WH-10) **or** $20,000.00 to "Una persona con negocio o profesión" (WH-04), **then** Q2
  is not shown, ISR and IVA withheld are **0**, net $5,800.00 / $23,200.00, rule ISR-0 with IVA-0 /
  IVA-1, and the card reads "Nadie te retiene impuestos en esta operación."

### US-W3 Honorarios and rent to a company
- **Given** a 612 PF, **when** I bill $50,000.00 of servicios profesionales to a company, **then**
  ISR $5,000.00, IVA withheld $5,333.33, net $47,666.67 (WH-02).
- **Given** a 606 PF, **when** I bill $30,000.00 of commercial rent to a company, **then** ISR
  $3,000.00, IVA $3,200.00, net $28,600.00 (WH-03); **when** "¿Es una casa o departamento para
  vivir?" = Sí, **then** the IVA rate becomes Exento, IVA 0, ISR $3,000.00, net $27,000.00 (WH-03b).

### US-W4 Special operations
- **Given** a 612 PF in actividad empresarial, **when** I sell goods of $50,000.00 to a company,
  **then** no withholding (WH-05); **given** RESICO PF, **then** ISR $625.00 only (WH-05b, F5).
- **Given** a PM 601, **when** I bill $10,000.00 of freight ("Flete…") to a company, **then** IVA
  withheld **$400.00** (4 % of the price), net $11,200.00 (WH-06).
- **Given** a PM 601, **when** I bill $100,000.00 of specialized services to a company, **then** no
  withholding, net $116,000.00, and no "subcontratación" option exists (WH-07, F6).
- **Given** a 612 PF, **when** I sell $20,000.00 of scrap to a company, **then** IVA withheld
  **$3,200.00** (100 %), net $20,000.00 (WH-08).
- **Given** a 625 PF with RFC given to the platform, **when** I record $10,000.00 of goods sold via
  the platform, **then** ISR $250.00 and IVA $800.00 withheld by the platform, net $10,550.00 (WH-09).

### US-W5 Unverified cases warn
- **Given** a 612 PF, **when** I bill $30,000.00 of honorarios to "Gobierno federal", **then** ISR
  $3,000.00 and IVA $3,200.00 (MVP default 2/3) are derived, with warning
  `withholding_cpa_unverified` and `cpa_unverified: true` (WH-12, A1, A2); the estimate's ISR and IVA
  lines fed by it carry `needs_cpa_review`.
- **Given** "Gobierno estatal o municipal", **then** IVA withheld is 0 (✅ LIVA 3) and ISR 10 % with
  the warning (A1).

### US-W6 PM pays a PF consultant (WM WH-11)
- **Given** a PM in **601**, **when** I record on 2026-06-15 an expense "Honorarios y servicios
  profesionales" of $40,000.00 + IVA $6,400.00, supplier "Una persona física con negocio o
  profesión", regime 612, operation professional services,
- **Then** the card shows "Debes retener: ISR **$4,000.00** · IVA **$4,266.67** — Pagas a tu
  proveedor **$38,133.33** · Entregas al SAT **$8,266.67** a más tardar el 17 de julio", and the
  warning `deduction_requires_remittance`;
- **And** the June estimate shows withholdings to remit $8,266.67 and IVA acreditable from this
  expense **$2,133.33**; the **July** estimate credits **$4,266.67** in the line "IVA retenido el mes
  pasado (acreditable este mes)" (goldens WH-11-JUN, WH-11-JUL).

### US-W7 PF users are never asked about suppliers
- **Given** a 612 PF, **when** I open "Nuevo gasto", **then** there is no "Tu proveedor" section, and
  a request with `supplier_kind` answers `422 supplier_not_applicable`.

### US-W8 No manual amounts
- **When** any client sends `isr_withheld_cents` or `iva_withheld_cents` on create or update,
  **then** `400` (`additionalProperties`); the stored withholding is always the derived one.

### US-W9 Edit re-derives
- **Given** US-W1's income, **when** I change "¿A quién le facturaste?" to "Una persona con negocio o
  profesión", **then** ISR and IVA withheld become 0, net $116,000.00, and July's estimate becomes
  ISR $2,000.00 and IVA $16,000.00.

### US-W10 Migration
- **Given** dev data with typed withholdings, **when** the operator runs the backfill in dry-run,
  **then** it lists per table how many rows change and their ids, without amounts or descriptions;
  **when** applied, every live record has `counterparty_kind` (or supplier fields for PM expenses),
  derived amounts and `withholding_table_version = 2026.1`; a second run changes nothing.

## 9. Tests

- Kernel: one table-driven case per row of §3.2 × receptor kinds, plus WH-01..WH-12 of WM §5 to the
  centavo, plus the property invariants of §3.2 (fixed seed). Independent double entry of the table
  against WM by qa-sr (as M5-06).
- M3/M4: contract tests for the new fields and the `400` on typed amounts; flag ON/OFF; RLS isolation
  re-run on the migrated tables.
- M5: goldens of §4.2 item 5, the January case (December's withheld IVA credited in January of the
  next fiscal year), fingerprint change with `EngineVersion` 1.1.0.
- Mobile: widget tests per question visibility rule (§2), the derived card, no editable withholding
  field; Prism mock for `previewWithholdings`.
- QA (rules/qa-browser.md): wipe local DB, create the US-W1, US-W4 and US-W6 data through the
  screens, compare UI figures with the goldens.

## 10. Open questions (with defaults; owner may overrule)

| # | Question | Default |
|---|---|---|
| OQ-W1 | The brief asks for the counterparty's regime "when PM or PF with activity". On **incomes** it cannot change the result (every PM withholds alike; a PF never withholds) | **Not asked on incomes.** Asked on **expenses** only for a PF supplier (E2), where it decides 10 % vs 1.25 % |
| OQ-W2 | First-time default for "¿A quién le facturaste?" | **No preselection** the first time (a wrong guess silently changes taxes); afterwards the last value for that regime |
| OQ-W3 | The CFDI shows a different withholding (payer error, special agreement) and there is no override | **Estimate uses the law's figure**, help text explains; revisit only if users report it (CFDI import will surface the gap) |
| OQ-W4 | Withheld IVA credit (LIVA 5 fr. IV) assumes the user remitted on time | **Assume on time**, ⚠️ line text; no "did you remit?" question in v1 |
| OQ-W5 | Partial creditability of a withheld expense (category creditable in part) | Same share as today's rule (all or nothing per expense) applied to both parts |
| OQ-W6 | Separate feature flag for this change | **None**: ships inside the three module flags, all OFF in shared environments |
| OQ-W7 | Where the derivation is stored | **Stored on the record** with `table_version` and re-derived by the backfill command on a table change (estimates stay a pure read; ADR-0008) |
