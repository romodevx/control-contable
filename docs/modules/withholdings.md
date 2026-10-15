# Module: Withholdings (derived ISR and IVA retenciones)

| | |
|---|---|
| Flag | **None of its own** (OQ-W6). It ships inside `module_incomes` (income side and `previewIncomeWithholdings`), `module_expenses` (supplier side and `previewExpenseWithholdings`) and `module_tax_engine` (the withheld IVA credited the month after). All three are OFF by default |
| Spec | [MODULE-withholdings.md](../project-specs/MODULE-withholdings.md); legal research: [withholdings matrix](../research/withholdings-matrix-2026.md) (WM: rule ids `ISR-n` / `IVA-n`, goldens `WH-nn`, ambiguities `An`) |
| Design | [ADR-0008](../adr/0008-withholdings-are-derived.md) (Accepted, with Amendments 1 and 2), [ADR-0007](../adr/0007-fiscal-shared-kernel.md) (shared kernel), [ADR-0002 engine version log](../adr/0002-tax-engine.md#engine-version-log) (1.1.0) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml) **0.12.0**: tags `Incomes` and `Expenses`, schemas `Withholding` and `WithholdingPreview` (break 0.8.0 → 0.9.0 by WH-02, cleanup 0.12.0 by WH-11 / WH-12) |
| Gates | [WH-13 QA](../../.gstack/qa-reports/qa-report-withholdings-2026-10-14.md) pass after one fix; [WH-14 tech lead](../reviews/WH-14-tech-lead-gate.md) **PASS**; WH-15 security **PASS** with one Low fix ([PR #154](https://github.com/romodevx/control-contable/pull/154)) |
| Operations | [Withholdings backfill runbook](../runbooks/withholdings-backfill.md), next to the [yearly parameter update](../runbooks/tax-params-update.md#7-engineversion-and-withholdingtableversion) |
| Code | `backend/internal/kernel/fiscal/` (`derive.go`, `derive_enums.go`, `explain.go`: the only decision table), `backend/internal/incomes/`, `backend/internal/expenses/`, `backend/internal/taxengine/domain/iva.go`, `backend/cmd/backfill-withholdings/`, migrations `20261016130000_withholdings_expand.sql` .. `20261018100000_withholdings_backfill_worker.sql`, `mobile/lib/features/incomes/`, `mobile/lib/features/expenses/` |

## What changes for the user

The user **never types a withholding**. The server derives the ISR and IVA withheld from a few
facts about the operation, the same facts the CFDI would show, and stores the result with the
record (owner decisions 1 and 2, 2026-10-15). There is no manual override: when a CFDI shows a
different amount, the estimate uses the law's figure (OQ-W3).

**Income form ("Nuevo ingreso"), every taxpayer.** Each question appears only when its answer
can change a withholding; otherwise its default is stored silently:

| Question | Field | Shown when |
|---|---|---|
| Q1 "¿A quién le facturaste?" | `counterparty_kind` | Always. No preselection the first time (OQ-W2), then the last value for the regime |
| Q2 "¿Qué le vendiste o qué servicio diste?" | `operation_kind` | The payer can withhold (a company or a government) and it was not via a platform |
| Q3 "¿Es una casa o departamento para vivir?" | `residential_rent` | Q2 is a real-estate rent. "Sí" forces IVA exempt |
| Q4 "¿Lo cobraste a través de una plataforma digital?" | `via_platform` | 625 (default Sí); personas morales under "Más detalles" |

**Expense form ("Nuevo gasto"), personas morales only.** A section "Tu proveedor" after the
category: E1 who sold (`supplier_kind`: company, persona física with activity, abroad), E2 the
supplier's regime when it is a persona física (`supplier_regime`, "No sé" derives as 612 and
warns), E3 the operation (`operation_kind`, starting at the category's
`default_operation_kind` from `listExpenseCategories`), E4 residential rent. A persona física
never withholds from a domestic supplier, so the section does not exist for one.

**The derived card.** Both forms show a read-only card that updates as the user types (a
debounced call to the preview operation of the module):

- Income: "Te retienen: ISR $1,250.00 · IVA $10,666.67. Recibes $104,083.33", then one line
  per rule in plain words, or "Nadie te retiene impuestos en esta operación."
- Expense: "Debes retener: ISR $4,000.00 · IVA $4,266.67. Pagas a tu proveedor $38,133.33 ·
  Entregas al SAT $8,266.67 a más tardar el 17 de julio", plus the deduction note.
- A rule pending CPA review adds a warning under the card. The help link "¿Tu factura dice otra
  cantidad?" explains why there is no override.

Editing a record re-derives it. "Guardar y crear otro" keeps Q1..Q4 and clears the amounts.

## Derivation catalog (summary)

One pure function, `fiscal.DeriveWithholdings`, used by incomes, expenses, both previews and
the backfill. Rates come from the dated parameter set on the record's withholding date
(`collected_on`, else `invoiced_on`; `paid_on` for an expense). ISR and IVA are evaluated
independently, first match wins. The full table, with every row's status, is
[MODULE-withholdings §3.2](../project-specs/MODULE-withholdings.md#32-rules-isr-and-iva-evaluated-independently-first-match-wins);
its legal basis is [WM §4.2](../research/withholdings-matrix-2026.md#42-rules-evaluate-isr-and-iva-independently-first-match-wins-within-each).
The cases a user meets most:

| Who bills whom, for what | ISR withheld | IVA withheld | Rule ids |
|---|---|---|---|
| Anyone bills a persona física or the público en general | 0 | 0 | ISR-0, IVA-0 / IVA-1 |
| 612 bills a company for fees (honorarios) | 10 % | 2/3 of the IVA | ISR-5, IVA-6 |
| 606 or 612 rents real estate to a company | 10 % | 2/3; 0 if residential (exempt) | ISR-6, IVA-6 |
| 626 RESICO PF bills a company (any non-platform operation) | 1.25 % | as any persona física: 2/3 on fees, rents and commissions | ISR-4, IVA-6 / IVA-8 |
| 625 sells through a platform | 2.1 % / 2.5 % / 4 % by activity; 20 % without RFC | 50 %; 100 % without RFC | ISR-3, IVA-3 |
| Road freight billed to a company | as the emisor's row | 4 % of the price | IVA-4 |
| Scrap and recyclables billed to a company | as the emisor's row | 100 % of the IVA | IVA-5 |
| A persona moral bills a company (other operations) | 0 | 0 | ISR-2, IVA-8 |
| Billed abroad | 0 | 0, and the warning `foreign_counterparty_not_supported` | ISR-1 |

- **No subcontracting 6 %**: LIVA 1-A fr. IV was repealed in 2021 (owner decision 4); the
  operation does not exist in the form.
- Amounts: `ISR = round(subtotal × rate)`, `IVA = round(IVA × share)` or
  `round(subtotal × 4 %)`, one rounding to the centavo, half away from zero. Invariants
  (property tests): `0 ≤ ISR ≤ subtotal`, `0 ≤ IVA withheld ≤ IVA`, `net ≥ 0`.
- **Rules pending CPA review** (⚠️ A1, A2, A3, A4, A6, A7, A8, A10: government receptors, 626
  rent, 625 off platform, commissions, movable rent, the 8 % border rate) set
  `cpa_unverified: true` and the warning `withholding_cpa_unverified`; the estimate lines they
  feed carry `needs_cpa_review` (topic `withholding_rule_unverified`).
- Every stored result carries `table_version` (`fiscal.WithholdingTableVersion`, today
  `2026.1`). It bumps on any rule or rate change; the backfill then re-derives the older rows.

### What is stored and served

`Income.withholding` and `Expense.withholding` (schema `Withholding`) are read-only:
`isr_cents`, `iva_cents`, `isr_rate`, `iva_basis`, `withholder` (`counterparty`, `you`,
`platform`, `none`), `net_cents`, `rule_ids`, `explanations` (es-MX, one per rule),
`table_version`, `cpa_unverified` and, on an expense, `remit_due_on` (the 17th of the next
month). `isr_withheld_cents` / `iva_withheld_cents` stay on the record as read-only mirrors:
the ledgers and the estimate read them.

| Warning | When |
|---|---|
| `withholding_cpa_unverified` | A ⚠️ rule matched |
| `foreign_counterparty_not_supported` | Q1 or E1 is abroad |
| `supplier_regime_unknown` | E2 "No sé" (derived as 612) |
| `deduction_requires_remittance` | An expense with a withholding: "Este gasto solo es deducible si enteras las retenciones a más tardar el 17 de {mes}." v1 assumes timely remittance and keeps the category's deductibility |
| `withholding_differs_from_rate` | **Retired**: never emitted (no typed amount to compare); stays in the enum, codes are append-only |

## Records stored before derivation (`withholding: null`)

Incomes stored before WH-06 and expenses stored before WH-07 have typed amounts and no
derivation. Contract 0.12.0 serves them; no deploy waits for the backfill
([ADR-0008 Amendment 2 §3](../adr/0008-withholdings-are-derived.md#amendment-2-the-contract-cleanup-of-wh-11--wh-12-0120)):

- **`withholding: null`** means exactly one thing: a record stored before automatic derivation
  that the server has not derived yet. A read never derives (reads stay pure and the table
  version auditable), and the estimate keeps using the stored amounts until the record is
  derived. A row with a table version but no derived detail reads the same way on both tables
  (WH-15 fix, [#154](https://github.com/romodevx/control-contable/pull/154)); only a corrupt
  derivation (an unknown withholder) is a `500`.
- **Facts are filled on load.** A legacy income's missing facts come from its old
  `counterparty_type` with the backfill's own mapping (persona moral → `pm`, persona física →
  `pf_with_activity`, público en general → `pf_no_activity_or_public`, plataforma → `pm` via a
  platform; the operation is the regime default). So a diff-only `PATCH` applies to the facts
  the read served.
- **An edit derives it.** Saving a legacy income stores its facts and the derivation. A persona
  moral's legacy expense has no supplier facts: the form opens "Tu proveedor" empty, and saving
  without E1 answers `400 supplier_kind required` (the QA walk saw "Revisa 1 dato antes de
  guardar" linked to E1).
- **The backfill derives them in bulk**, and re-derives rows of an older `table_version`:
  [runbook](../runbooks/withholdings-backfill.md). Its "need user input" rows (a regime no
  longer held, a date out of range) stay `null` until the user edits them.
- M8 History marks the months whose figures the backfill or an edit moved ("Cambió desde que lo
  viste"). That is the intended signal.

## Effect on the monthly estimate

**Income side: no formula changed.** The engine reads the stored withheld amounts as before;
only their origin is new. For example US-W1 (626 RESICO, 100,000.00 + IVA to a company in
July): ISR 750.00, IVA 16,000.00 − 10,666.67 = 5,333.33 (golden WH-01-EST). A 626 month whose
1.25 % withholding exceeds the ISR shows 0 to pay and the excess as "a favor en tu declaración
anual" (⚠️ A4).

**Expense side (personas morales): engine 1.1.0.** What the taxpayer withholds from a supplier
is remitted with the month's return (`withholdings_to_remit`, never netted). Its creditable IVA
is split in two:

| Part | Credited | Stored in |
|---|---|---|
| IVA paid to the supplier (IVA − IVA withheld) | The month it is paid | `creditable_iva_cents` |
| IVA withheld from the supplier | **The next month**, in the return after the one that remitted it (LIVA 5 fr. IV) | `creditable_iva_withheld_cents` |

The breakdown shows the second part as "IVA retenido el mes pasado (acreditable este mes;
supone que enteraste la retención a tiempo)", flagged for CPA review (topic
`withheld_iva_credit_timing`, OQ-W4). January credits the previous December. Details:
[tax engine page](tax-engine.md#withheld-iva-is-credited-the-month-after).

### Try it locally

These requests were run on 2026-10-15 against `make -C backend api` with the `demo_full`
preset, on an isolated local stack. Get `ACCESS` from the
[M1 walkthrough](auth.md#try-it-locally) (steps 1 to 5).

```sh
API=http://localhost:3000
AUTH="Authorization: Bearer $ACCESS"

# 0. A RESICO persona física (626) and a persona moral (601), both since January 2026.
PF_ID=$(curl -s -X POST "$API/v1/taxpayers" -H "$AUTH" -H 'Content-Type: application/json' -d '{
  "person_type": "fisica", "rfc": "ROAA800101AB6", "display_name": "Ana RESICO",
  "regimes": [{"regime_code": "626", "start_period": "2026-01"}]}' | jq -r .id)
PM_ID=$(curl -s -X POST "$API/v1/taxpayers" -H "$AUTH" -H 'Content-Type: application/json' -d '{
  "person_type": "moral", "rfc": "EQA200101AB1", "display_name": "Empresa QA",
  "regimes": [{"regime_code": "601", "start_period": "2026-01"}]}' | jq -r .id)

# 1. US-W1: what a company withholds from 100,000.00 + IVA of fees (preview, nothing stored).
curl -s "$API/v1/taxpayers/$PF_ID/incomes/withholding-preview?regime=626&on=2026-07-10&amount_cents=10000000&iva_rate=rate_16&iva_cents=1600000&counterparty_kind=pm&operation_kind=professional_services" \
  -H "$AUTH" | jq -c '.withholding | {isr_cents, iva_cents, net_cents, rule_ids, withholder, table_version}'

# 2. Store it: the same derivation is stored with the record and served read-only.
curl -s -X POST "$API/v1/taxpayers/$PF_ID/incomes" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"regime_code": "626", "counterparty_kind": "pm", "operation_kind": "professional_services",
       "amount_cents": 10000000, "iva_rate": "rate_16", "iva_cents": 1600000, "collected_on": "2026-07-10"}' \
  | jq -c '{isr: .withholding.isr_cents, iva: .withholding.iva_cents, isr_withheld_cents, iva_withheld_cents}'

# 3. A typed amount is not accepted: 400 validation_failed.
curl -s -X POST "$API/v1/taxpayers/$PF_ID/incomes" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"regime_code": "626", "counterparty_kind": "pm", "amount_cents": 100000, "iva_rate": "rate_16",
       "iva_cents": 16000, "collected_on": "2026-07-11", "isr_withheld_cents": 0}' \
  | jq -c '{code: .error.code, details: [.error.details[] | {field, code}]}'

# 4. US-W6: the persona moral pays a persona física consultant (612) 40,000.00 + IVA in June.
curl -s -X POST "$API/v1/taxpayers/$PM_ID/expenses" -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"regime_code": "601", "category_code": "servicios_profesionales", "amount_cents": 4000000,
       "iva_rate": "rate_16", "iva_cents": 640000, "payment_method": "transfer", "has_cfdi": true,
       "paid_on": "2026-06-15", "supplier_kind": "pf_with_activity", "supplier_regime": "612"}' \
  | jq -c '{withholding: (.withholding | {isr_cents, iva_cents, net_cents, withholder, remit_due_on}),
            warnings: [.warnings[].code]}'

# 5. June: the withholdings to remit; July: the withheld IVA credited the month after.
curl -s "$API/v1/taxpayers/$PM_ID/estimates?period=2026-06" -H "$AUTH" \
  | jq -c '{period, withholdings_to_remit: .totals.withholdings_to_remit_cents, iva: .totals.iva}'
curl -s "$API/v1/taxpayers/$PM_ID/estimates/2026-07/breakdown?regime=601" -H "$AUTH" \
  | jq -c '.lines[] | select(.concept | startswith("IVA retenido el mes pasado"))
           | {concept, operation, amount_cents, legal_basis: [.legal_basis[] | "\(.law) \(.article)"], needs_cpa_review}'
```

Output, in order:

```text
{"isr_cents":125000,"iva_cents":1066667,"net_cents":10408333,"rule_ids":["ISR-4","IVA-6"],"withholder":"counterparty","table_version":"2026.1"}
{"isr":125000,"iva":1066667,"isr_withheld_cents":125000,"iva_withheld_cents":1066667}
{"code":"validation_failed","details":[{"field":"isr_withheld_cents","code":"additionalProperties"}]}
{"withholding":{"isr_cents":400000,"iva_cents":426667,"net_cents":3813333,"withholder":"you","remit_due_on":"2026-07-17"},"warnings":["deduction_requires_remittance"]}
{"period":"2026-06","withholdings_to_remit":826667,"iva":{"amount_cents":-213333,"payable_rounded_cents":0,"outcome":"in_favor"}}
{"concept":"IVA retenido el mes pasado (acreditable este mes; supone que enteraste la retención a tiempo)","operation":"subtract","amount_cents":426667,"legal_basis":["LIVA 5"],"needs_cpa_review":true}
```

Steps 1 and 2 are golden WH-01 (1.25 % and 2/3 of the IVA; the payer keeps 11,916.67 of
116,000.00). Step 4 is WH-11: the supplier is paid 38,133.33 and 8,266.67 goes to the SAT by
17 July. June credits only the paid part, 6,400.00 − 4,266.67 = 2,133.33 (a saldo a favor with
no income); July credits the 4,266.67 withheld in June (goldens WH-11-JUN, WH-11-JUL). The
walkthrough uses 1 of the user's 5 trial incomes and 1 of the 5 trial expenses.

## Flags, prerequisites and presets

| Flag OFF | Effect on withholdings |
|---|---|
| `module_incomes` | `previewIncomeWithholdings` and every income route answer `404 not_found`; the income form is gone |
| `module_expenses` | `previewExpenseWithholdings` and every expense route answer `404`; "Tu proveedor" is gone |
| `module_tax_engine` | No estimate; derivations are still stored by the record modules |
| `module_taxpayers` | All three are effectively OFF (prerequisite) |

- Both previews spend one per-user budget, `withholding_previews`, 1,200 per hour, separate from
  the record budgets; spent → `429 rate_limited` with `Retry-After`. Their query carries amounts
  only, never a name, an RFC or a description, and is not logged.
- Presets: `demo_core`, `demo_monetization` and `demo_full` all include incomes, expenses and
  the tax engine, so withholdings are in every preset.
- **Enable in a shared environment** only with WH-14 and WH-15 (both passed) **and** each
  module's own preconditions: the [incomes and expenses checklist](../runbooks/incomes-expenses.md#5-enable-module_incomes-and-module_expenses-in-a-shared-environment)
  and, for the estimate, the
  [M5-31 preconditions](../reviews/M5-31-security-gate.md#5-preconditions-to-enable-module_tax_engine-in-staging-or-production)
  (first the CPA sign-off). In an environment that already has records, run the
  [backfill](../runbooks/withholdings-backfill.md) after the deploy so legacy records show
  their derivation; it is not a deploy precondition.

## API reference and errors

The operations (`previewIncomeWithholdings`, `previewExpenseWithholdings`, and the
`withholding` member of the income and expense operations), schemas, parameters and examples
live only in the contract. Render them instead of copying them here:

```sh
make api-docs                 # build/api-docs/index.html, rendered from docs/api/openapi.yaml (tags Incomes, Expenses)
make mock                     # Prism mock of the contract, for the forms without a backend
```

The preview errors are in the generated error tables of
[incomes](incomes.md#errors) and [expenses](expenses.md#errors) (`make api-error-tables`
regenerates them): `400` for malformed or out-of-range facts, `422` for
`regime_not_active`, `rate_not_valid_on_date`, `iva_amount_mismatch`,
`residential_rent_exempt`, `activity_not_held`, and the hidden `404` while the module is OFF.

## Gates

| Gate | Verdict | Notes |
|---|---|---|
| WH-13 QA ([report](../../.gstack/qa-reports/qa-report-withholdings-2026-10-14.md)) | Pass after one fix | Device walk on the Android emulator against its own stack: US-W1, US-W4, US-W6 entered through the screens equal the goldens to the centavo; 15 / 15 WM goldens through the API (now `TestWithholdingsQA_goldensThroughTheAPI`); flags OFF (18 probes, all `404`); backfill dry run on a copy writes nothing, apply then a second dry run reports 0. F1 fixed: the estimate card labelled the creditable IVA "pagado" |
| WH-14 tech lead ([review](../reviews/WH-14-tech-lead-gate.md)) | **PASS** | One decision table (`fiscal/derive.go`); the database holds only the vocabulary and the app only decides which question to show; no clock, random or float in the kernel. QA O2 (legacy expense E3 empty) fixed test-first |
| WH-15 security ([PR #154](https://github.com/romodevx/control-contable/pull/154)) | **PASS**, one Low fix | A half-derived expense row read as derived and answered `500`; now `withholding: null`, as incomes. Checked: preview validation and shared budget, `cc_worker` backfill role and `withholding_backfill_users()` (ids only), column-scoped grants, report and logs without PII, flags OFF ⇒ `404` |

## Decisions

Full text in [MODULE-withholdings §10](../project-specs/MODULE-withholdings.md#10-open-questions-with-defaults-owner-may-overrule).

| # | Decision | Effect |
|---|---|---|
| Owner 1, 2 | Derived, no manual override, expense side included | No withholding input anywhere |
| Owner 4 | Subcontracting 6 % excluded (repealed 2021) | Not an operation |
| OQ-W1 | The payer's regime is never asked on incomes | It cannot change the result; asked on expenses (E2) for a persona física supplier |
| OQ-W2 | No first-time preselection for Q1 | A wrong guess would silently change taxes |
| OQ-W3 | The CFDI shows another amount | The estimate uses the law's figure; revisit with CFDI import |
| OQ-W4 | Withheld IVA credit assumes timely remittance | ⚠️ line text, no "did you remit?" question in v1 |
| OQ-W5 | Partial creditability | Same all-or-nothing rule, applied to both parts |
| OQ-W6 | No flag of its own | Ships inside the three module flags |
| OQ-W7 | Stored with `table_version` | Estimates stay a pure read; the backfill re-derives on a table change |

## Known limits and follow-ups

Tracked in the [PLAN WH follow-ups](../project-specs/PLAN-control-contable.md#wh-follow-ups).
The ones a user or an operator can notice:

- **CPA review pending (WH-17, blocked by OB5).** No ⚠️ rule is confirmed yet; each answer
  bumps `table_version` and is followed by a backfill run.
- **Onboarding start month is easy to miss** (QA O1). "¿Desde qué mes tributas…?" sits below
  every regime in onboarding step 3; a taxpayer created with the current month cannot enter
  earlier records (`422 regime_not_active`). Placement is a design-review item for
  `module_taxpayers` (frontend-sr + design review), not a withholdings defect.
- The `incomes.counterparty_type` column stays (writes mirror it) until the backfill has run in
  every environment; its drop is its own migration (WH-08 contract step).
- `IvaComponents.creditable_paid_cents` holds both creditable parts since engine 1.1.0 and has
  no description in the contract (WH-14 follow-up 1); the app labels it "IVA acreditable".
- A user whose CFDI shows a different withholding cannot correct it (OQ-W3). Foreign
  operations derive zero with a warning; export IVA is out of scope.
