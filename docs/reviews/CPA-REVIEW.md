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
- [ ] RMF 2026 rule numbers for RESICO (Title 3.13) and platform facilities (3.11.x): the RMF text was not retrieved. R Part D, CAT `regimes` 625
- [ ] Rental IVA withholding: RLIVA art. 3 fr. I c) 2/3 used; a secondary source claims RMF rule 4.1.4 sets 100 %. R Part D
- [ ] Due-date move to the next business day under CFF art. 12. ADR §4
- [ ] Day-17 grace by 6th RFC digit (+1 to +5 business days, DOF 26-12-2013 art. 5.1): not re-verified, not applied. R §A.4, ADR §4
- [ ] 2026 employment subsidy decree: not verified. R Part C
- [ ] Payment rounding to whole pesos (CFF art. 20): not covered; off until confirmed (`rounding.payment_whole_pesos`). ADR §4
- [ ] Monthly IVA netting across regimes for a PF holding 612 + 606 (one consolidated IVA result). ADR §1
- [ ] 626 PM credits the art. 54 interest withholding like 601 (input `interest_withholding` added to 626 PM). R §B.2.2 step 4, CAT `regimes` 626 PM
- [ ] RIF (621) IVA paid by bimester like its ISR (old LIVA 5-E), not monthly; research A.3 lists 621 under the monthly rule. MODULE-tax-engine OQ-T7
- [ ] IVA a PM withheld **from its suppliers** (recorded on the expense) becomes creditable for that PM when paid (LIVA 5 fr. IV): v1 sends it to withholdings to remit only and does not credit it. Confirm timing and whether to credit. Distinct from IVA withheld **from** the taxpayer and recorded on the income, which v1 subtracts from the monthly IVA (LIVA 1-A, 5-D). MODULE-tax-engine OQ-T6

## Regime catalog (`regimes`, needs_cpa_review = true)

- [ ] 603 PM (fines no lucrativos): periodicity "annual" is inferred; research only says "generally no provisional ISR". CAT `regimes` 603
- [ ] 620 PM (cooperativas que difieren): payment periodicity not stated. CAT `regimes` 620
- [ ] 621 PF (RIF): reduction table, see above. CAT `regimes` 621
- [ ] 623 PM (grupos de sociedades): payment periodicity not stated. CAT `regimes` 623
- [ ] 624 PM (coordinados): payment periodicity not stated. CAT `regimes` 624
- [ ] 625 PF (plataformas): 18-M definitive IVA and RMF platform rules, see above. CAT `regimes` 625

## Regime compatibility (`regime_compatibility`, allowed with a warning until confirmed)

Each pair below is currently **allowed** and shows a "pending CPA confirmation" warning.
For each, the CPA answers: may one taxpayer hold both regimes at the same time?

- [ ] PF 621 (RIF) with 605, 606, 607, 608, 611, 612, 614, 615, 625, 626 (10 pairs). CAT
- [ ] PF 626 (RESICO) with 607, 608, 615 (3 pairs; art. 113-E only names salary and interest). CAT
- [ ] PM 601/603/620/622/623/624 among themselves (15 pairs), and 620 with 626 (1 pair). CAT
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
