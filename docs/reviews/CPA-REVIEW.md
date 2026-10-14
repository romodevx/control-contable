# CPA review checklist (fiscal year 2026)

Every item the team could not confirm against a primary source, in one place, for the
owner's accountant (PRD Q11). Tick an item when the CPA confirms or corrects it, and fix
the source it points to in the same PR. Until an item is ticked, the app treats it as
unverified: catalog rows carry `needs_cpa_review` (shown to users as a non-blocking
warning), and golden cases stay skipped (ADR-0002 §5).

Sources: [research](../research/tax-regimes-mx-2026.md) (R), [ADR-0002](../adr/0002-tax-engine.md) (ADR),
catalog migration `backend/migrations/20261013180000_regimes_catalog.sql` (CAT).

## Rates, tables and parameters

- [ ] PTU = 10 % of taxable income (Sexta Comisión, DOF 18-09-2020): secondary source only. R §0, R Part D #11
- [ ] PTU payment due within 60 days after the annual return (May 30, LFT art. 122): not re-fetched. R §B.4
- [ ] RIF (621) ISR reduction by year in RIF, 100 % to 10 % for years 1-10 (repealed art. 111): not re-fetched. R §A.2.7, ADR §2, CAT `regimes` 621
- [ ] 606 quarterly option threshold: 10 monthly minimum wages read as UMA ($35,662.20/month) per the desindexation decree; quarterly table not transcribed. R §A.2.4 step 5, ADR §2 (606 row)
- [ ] 601 profit coefficient rounding: art. 14 is silent; practice is 4 decimals; the app uses the user's figure as given. R §B.2.1, ADR §4
- [ ] January 2026 monthly UMA = 2025 UMA ($113.14 × 30.4 = $3,439.46): check against the INEGI/DOF 2025 publication. ADR §3

## Rules and procedures

- [ ] 625 definitive IVA: under LIVA 18-L/18-M, with the definitive option the 50 % platform withholding is final and directly collected IVA is paid at 8 %. R §A.2.6, R PF-625-01, ADR §1, CAT `regimes` 625
- [ ] 625 direct collections at 8 % (LIVA 18-M ¶2, text DOF 09-12-2019, unchanged in the 12-11-2021 compilation): the app applies 8 % to the consideration before IVA (half of the 16 % IVA charged), infers the direct share as the month's 16 % IVA minus 2× (with RFC) or 1× (without RFC) the platform withholding, leaves border-8 % direct IVA as is, and also applies it when the platform had no RFC. Confirm the base and those three choices. MODULE-tax-engine §2.4 row 625, PR #94
- [ ] RMF 2026 rule numbers for RESICO (Title 3.13) and platform facilities (3.11.x): the RMF text was not retrieved. R Part D, CAT `regimes` 625
- [ ] Rental IVA withholding: RLIVA art. 3 fr. I c) 2/3 used; a secondary source claims RMF rule 4.1.4 sets 100 %. R Part D — Withholdings matrix §2.2 read RMF 4.1.4: it concerns quarterly 606 lessors withholding under LIVA 1-A fr. III (foreign residents), not the 2/3 fraction; the app keeps 2/3. Confirm.
- [ ] Due-date move to the next business day under CFF art. 12. ADR §4
- [ ] Day-17 grace by 6th RFC digit (+1 to +5 business days, DOF 26-12-2013 art. 5.1): not re-verified, not applied. R §A.4, ADR §4
- [ ] 2026 employment subsidy decree: not verified. R Part C
- [ ] Payment rounding to whole pesos (CFF art. 20): not covered; off until confirmed (`rounding.payment_whole_pesos`). ADR §4
- [ ] Monthly IVA netting across regimes for a PF holding 612 + 606 (one consolidated IVA result). ADR §1
- [ ] 626 PM credits the art. 54 interest withholding like 601 (input `interest_withholding` added to 626 PM). R §B.2.2 step 4, CAT `regimes` 626 PM
- [ ] RIF (621) rounding of the reduced ISR, golden PF-621-01 (bimester Sep-Oct 2026, year 6, base 40,000.00, bimonthly row 6, 50 % reduction). The research and a single final rounding (ADR-0002 §4) differ by one centavo; confirm which the law means:
  - Research (round the caused ISR, then halve): 3,713.68 + (40,000.00 − 35,067.29) × 21.36 % = 3,713.68 + 1,053.626856 = 4,767.306856 → **4,767.31**; 4,767.31 × 50 % = 2,383.655 → **2,383.66**.
  - Single rounding (halve the unrounded caused ISR, round once): 4,767.306856 × 50 % = 2,383.653428 → **2,383.65**.
  - The golden keeps the research's 2,383.66 and stays skipped until this item and the RIF reduction % above are ticked. R §A.5 PF-621-01, golden `backend/internal/taxengine/testdata/golden/621-pf/PF-621-01.yaml`, ADR §4
- [ ] RIF (621) IVA paid by bimester like its ISR (old LIVA 5-E), not monthly; research A.3 lists 621 under the monthly rule. MODULE-tax-engine OQ-T7
- [ ] IVA a PM withheld **from its suppliers** (recorded on the expense): **rule now read verbatim** (LIVA 5 fr. IV, withholdings matrix §2.2): creditable in the monthly return after the one in which it was remitted; the app will credit month n − 1's withheld IVA in month n (PLAN WH-09). Confirm only the assumption that the user remitted on time, and that the all-or-nothing creditable share of the expense applies to both the paid and the withheld part. Distinct from IVA withheld **from** the taxpayer and recorded on the income, which v1 subtracts from the monthly IVA (LIVA 1-A, 5-D). MODULE-tax-engine OQ-T6, MODULE-withholdings §4.2, OQ-W4, OQ-W5

## Regime catalog (`regimes`, needs_cpa_review = true)

- [ ] 603 PM (fines no lucrativos): periodicity "annual" is inferred; research only says "generally no provisional ISR". CAT `regimes` 603
- [ ] 620 PM (cooperativas que difieren): payment periodicity not stated. CAT `regimes` 620
- [ ] 621 PF (RIF): reduction table, see above. CAT `regimes` 621
- [ ] 623 PM (grupos de sociedades): payment periodicity not stated. CAT `regimes` 623
- [ ] 624 PM (coordinados): payment periodicity not stated. CAT `regimes` 624
- [ ] 625 PF (plataformas): 18-M definitive IVA and RMF platform rules, see above. CAT `regimes` 625
- [ ] `regimes.explanation` texts (M5-19) for 605, 607, 608, 610 PF/PM, 611, 614, 615, 616, 620, 623, 624, 621, 622, 603: written at article level from the research and the catalog's legal basis, not re-read against the law text. Confirm in particular: 605 annual-return cases (LISR 98 fr. III, $400,000); 607 per-operation payment within 15 days and the notary's role (LISR 126); 610 withholding as definitive, normally no annual return (LISR 153); 611 10 % definitive withholding plus accumulation with credit (LISR 140); 614 retention definitive when real interest ≤ $100,000 and no other income (LISR 135); 615 retention as definitive (LISR 138); 616 change of regime notice (CFF 27); 620 deferral until distribution (LISR 194-199); 603 ISR only on items such as non-deductible expenses or income outside its purpose (LISR 79-80). Migration `20261014170000_regime_explanations.sql`

## Regime compatibility (`regime_compatibility`, allowed with a warning until confirmed)

Each pair below is currently **allowed** and shows a "pending CPA confirmation" warning.
For each, the CPA answers: may one taxpayer hold both regimes at the same time?

- [ ] PF 621 (RIF) with 605, 606, 607, 608, 611, 614, 615, 625 (8 pairs). Art. 111 (2016 text) admits Capítulos I, III and VI income besides, and fr. I excludes shareholders: 605, 606, 614 look allowed; 607, 608, 611, 615, 625 look excluded. R Part E, CAT
- [ ] PF 626 (RESICO) with 607, 608, 615 (3 pairs; art. 113-E ¶1 "únicamente" and ¶6 only name salary and interest, so they look excluded). R Part E, CAT
- [ ] PM 601 × 603 (art. 80: a Título III entity owes Título II tax on some income), 601 × 623 (art. 59 ¶4: a Cap. VI company stays under Título II), 601 × 624, 603 × 620, 603 × 622, 603 × 623, 603 × 624, 623 × 624 (8 pairs; no text found either way). R Part E, CAT
- [x] Settled from the law in M2-13 FU1 (refused, no longer under review): PF 612 × 621, 621 × 626; PM 601 × 620, 620 × 622, 620 × 623, 620 × 624, 620 × 626, 601 × 622, 622 × 623, 622 × 624. R Part E, migration `20261014120000_regime_compatibility_reconcile.sql`. The CPA may still contest a citation.
- [ ] PF Título IV pairs among 605, 606, 607, 608, 611, 612, 614, 615, 625 are seeded as allowed (separate chapters, combined in the annual return) by inference, not by an explicit rule. CAT
- [ ] 616 (sin obligaciones) and 610 (residentes en el extranjero) are seeded as excluding every other regime by definition. CAT

## Expense categories (`expense_categories`, `category_regime_rules`, all rows needs_cpa_review = true)

Seeded in M4-02 from [MODULE-expenses §2.2](../project-specs/MODULE-expenses.md) (OQ-E3, OQ-E4),
migration `backend/migrations/20261013210000_expense_categories.sql` (EXP). Every category and
every (category, regime) rule waits for this review; the legal basis is article level only.

- [ ] The 26 categories and their per-regime hint (D / P / N / ISR not applicable) for the business regimes, 606 (actual deductions) and 626 PF. MODULE-expenses §2.2, EXP
- [ ] Conservative defaults: investments (equipo de cómputo, mobiliario, automóvil, construcciones), restaurantes, gastos personales and "otros" start non-deductible. OQ-E3, EXP
- [ ] Cash rule (LISR art. 27 fr. III): cash above $2,000.00 on the total with IVA is not deductible; fuel paid in cash never is. Not in the research. OQ-E4, §2.5, EXP `cash_payment_rule`
- [ ] 620 PM (cooperativas que difieren) gets the business rows although the regime catalog marks it informational. EXP
- [ ] The profit-coefficient note on 601 and 622 (regimes requiring `coeficiente_utilidad`): expenses do not lower the 601/622 provisional ISR, only IVA and the annual return. EXP
- [ ] `iva_creditable_default`: true where the expense is (or, for RESICO PF, would be under the business column) ISR-deductible, per LIVA art. 5 fr. I; partial cases assume IVA creditable in the deductible proportion. R §A.3, EXP
- [ ] Creditable IVA follows `iva_creditable_default`, never the absence of ISR deductions (LIVA art. 5 fr. I, "estrictamente indispensable"): a never-deductible category (e.g. `gastos_personales`) yields **0** creditable IVA in every regime, RESICO PF included, and a user override to deductible does not change that. PR #67, `expenses/domain.creditableIVA`, EXP
- [ ] A cash payment over the $2,000.00 limit yields **0** creditable IVA even when the user overrides it to deductible (LIVA art. 5 fr. II, payment other than cash above the threshold): the override stands for the ISR display only, and the expense carries the warning `iva_not_creditable_cash_payment`. Confirm the threshold is the LISR art. 27 fr. III amount on the total with IVA. PR #67, MODULE-expenses §2.4–§2.5
- [ ] 606 rows: only servicios profesionales, agua (servicios básicos), construcciones, sueldos, IMSS, intereses, seguros, mantenimiento and predial count (LISR art. 115). R §A.2.4, EXP
- [ ] Booking date: an expense books on `paid_on` in every regime (no invoice date is stored). Correct for PF (LISR arts. 103/105, 115, 113-E) and RESICO PM (arts. 208, 211), and for creditable IVA (LIVA art. 5 fr. III). Título II PM (601; 620/622 unless cash basis applies, see the income-basis item) deduct on an accrual basis, except payments that are income of a PF, deductible only when paid (LISR art. 27 fr. VIII); v1 is monthly only and the 601/622 provisional ISR uses the profit coefficient (art. 14), so no monthly figure depends on it. Confirm; an annual PM computation would need an accrual date. Research §B.2.1 (income only), EXP
- [ ] 603 (fines no lucrativos) gets the business rows, but it has no provisional ISR: non-deductible expenses become remanente distribuible (LISR art. 79) rather than lowering a base. Is the D / P / N hint the right message for 603? R §B.1, EXP
- [ ] Copy: the coefficient note on 622 rows reads "En los pagos provisionales del 601…" (spec text verbatim); should 622 name its own regime? MODULE-expenses §2.2, EXP

## Income basis (`regimes.income_basis`, M3-02)

- [ ] 620, 622 and 624 are seeded `accrual` with 601, but under LISR they look cash basis: 620 cooperativas que difieren (arts. 194-199, income deferred until distributed), 622 AGAPES (art. 74, flujo de efectivo) and 624 coordinados (arts. 72-73, flujo de efectivo). Confirm per regime; a change re-derives `incomes.booked_on` in the same migration. MODULE-incomes §2.2, `20261013200000_incomes.sql`

## Withholdings derivation ([MODULE-withholdings](../project-specs/MODULE-withholdings.md), [withholdings matrix](../research/withholdings-matrix-2026.md) = WM)

Withholdings are derived automatically, with no manual override (owner, 2026-10-15), so each item
below is a figure the user cannot correct. Records matching a ⚠️ rule carry
`withholding_cpa_unverified`; ticking an item bumps the derivation `table_version` and re-derives
(MODULE-withholdings §5.3).

- [ ] A1: the Federación, estados and municipios withhold ISR like a persona moral (10 % fees and rent, 1.25 % RESICO). LISR 86 vs 106/116/113-J. WM §6 A1
- [ ] A2: IVA fraction the Federación withholds from a PF: 2/3 (RLIVA 3) on professional services; on any other operation only above $2,000 (RMF 4.1.3). The literal LIVA 3 ¶3 reading is 100 %. WM §6 A2, WH-12
- [ ] A3: RESICO PF renting real estate to a PM is withheld 1.25 % (113-J), not 10 % (116). WM §6 A3
- [ ] A4: RESICO PF month where the 1.25 % withheld exceeds the monthly ISR: shown as 0 to pay, excess "a favor en la anual" (no monthly refund or carry). WM §6 A4
- [ ] A5: 2/3 computed as IVA × 2 / 3 rounded half away from zero to the centavo (not `0.106667 × price`). WM §6 A5
- [ ] A6: a 625 PF billing a PM directly (outside the platform): no PM withholding. WM §6 A6
- [ ] A7: PF commission agent: no ISR withholding (only IVA 2/3). WM §6 A7
- [ ] A8: under the 8 % border rate, the 2/3 IVA withholding is 2/3 of the IVA charged (5.3333 % of the price); freight stays 4 % of the price. WM §6 A8
- [ ] A9: "Servicios profesionales" vs "otros servicios de tu negocio" is the user's answer (professional fees → 10 % ISR + 2/3 IVA for 612; business services → nothing). Check the es-MX wording of both options in MODULE-withholdings §2.1 Q2. WM §6 A9
- [ ] A10: PF renting movable goods to a PM: no ISR withholding, IVA 2/3. WM §6 A10
- [ ] A11: donativos have no withholding (no operation kind offered). WM §6 A11
- [ ] A12: SAT Anexo 7 RMF 2026 criterios normativos not retrieved; re-check before release. WM §6 A12
- [ ] A13: payments to and from foreign residents are not derived (zero + warning `foreign_counterparty_not_supported`). WM §6 A13
- [ ] Deduction conditioned on remittance (LISR 27 fr. V, 147 fr. VII, 210): v1 keeps the category's deductibility, assumes timely remittance, and warns `deduction_requires_remittance`. MODULE-withholdings §4.2
- [ ] "No sé" on the supplier's regime derives as 612 (10 % ISR on fees) with a warning. MODULE-withholdings §2.2 E2
- [ ] Category → operation kind defaults for expenses (servicios profesionales → honorarios, renta → inmueble, fletes → autotransporte, others → servicios empresariales). MODULE-withholdings §2.2

## Annual estimate (`module_annual`, MODULE-annual)

- [ ] **RESICO PF income in the 15 % personal-deductions cap** (topic `deduction_cap_resico_income`, MODULE-annual §2.7 and OQ-A4). LISR art. 151 last ¶ caps the capped personal deductions at min(5 UMA anuales; 15 % "del total de los ingresos del contribuyente, incluyendo aquellos por los que no se pague el impuesto"). Read on its own, that wording could include RESICO PF income. Three things point the other way: (1) RESICO PF income is taxed by itself under arts. 113-E/113-F, with no deductions of any kind, and rule 3.13.7 RMF 2026 makes its monthly payments definitive; (2) rule 3.13.8 says a RESICO taxpayer with salary or interest files the annual for those chapters **separately**, so the art. 152 return the deductions reduce holds only that other income; (3) art. 152 ¶1 lists the chapters it accumulates and Cap. II Sección IV is not one of them. **Default applied: RESICO income excluded** from "total income". This is the conservative choice: it gives a lower cap and never overstates a refund. Confirm, or say whether the cap base should include RESICO income collected (AR Part E CPA #10).
