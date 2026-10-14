# Mexican annual return (declaración anual), ejercicio 2026: estimation rules

> **Purpose.** Research basis for the app's **estimated** annual ISR return. Every screen that uses it must say it is an estimate and that the user's accountant files the real return. Not tax advice.
> **Cut-off date:** 2026-10-15. **Fiscal year estimated:** 2026 (filed in 2027).
> **Extends** `docs/research/tax-regimes-mx-2026.md` (the "monthly doc"). Tariffs, UMA, RESICO rates, regime catalog and the monthly algorithms live there and are referenced, not repeated.
> **Status legend:** ✅ read in the primary source in this session · ⚠️ not confirmed, derived, or depends on a rule not yet published (do not hard-code without CPA review).

## 0. Headline findings

| # | Finding | Basis | Status |
|---|---|---|---|
| 1 | The annual return is an **ISR-only** return. **IVA has no annual return**: IVA is computed per calendar month (quarterly for small rent, art. 5-F) and paid by day 17; the law has no annual IVA computation. | LIVA art. 5-D ¶1–3 | ✅ |
| 2 | The PF annual return **adds up the income of Capítulos I, III, IV, V, VI, VIII, IX plus the utilidad gravable of Cap. II Sección I (612)**, subtracts the art. 151 personal deductions, and applies the annual tariff. **RESICO PF (Cap. II Sección IV) is not in that list**: it is a separate computation. | LISR art. 152 ¶1; RMF 2026 rule 3.13.8 | ✅ |
| 3 | **RESICO PF taxpayers are relieved of the annual return** for their RESICO income: RMF 2026 rule 3.13.7 makes the monthly payments **definitive**. The law (art. 113-F) still describes an annual computation, so the app can show it as an informative reconciliation only. | LISR 113-F; RMF 2026 rule 3.13.7; SAT minisitio | ✅ rule text · ⚠️ RMF 2027 must keep the rule for ejercicio 2026 |
| 4 | The 2026 annual tariff (Anexo 8 RMF 2026, C.II) is **identical to the December accumulated table** already stored for regime 612. | Anexo 8 RMF 2026, C.II (verified again here) | ✅ |
| 5 | Deadlines for ejercicio 2026: **PM by Wednesday 31 March 2027** (601, 626 PM, 622, 623, 624), **PF during April 2027, last day Friday 30 April 2027**, **603 non-profits informative return by Monday 15 February 2027**. | LISR arts. 9 ¶3, 76 fr. V, 150 ¶1, 113-F, 86 fr. III | ✅ law · ⚠️ weekdays computed, no extension published yet |
| 6 | Several facilities used every year are **written only for ejercicio 2025** in RMF 2026 (six installments, rule 3.17.3; salary-only relief, rule 3.17.11). Their 2026 equivalents will be in RMF 2027 (expected DOF late December 2026). | RMF 2026 rules 3.17.3, 3.17.11 | ⚠️ |

---

# PART A: Who files an annual return (obligation table)

## A.1 Personas físicas

| Key | Regime | Annual ISR return for ejercicio 2026? | Deadline | Notes / informative-only parts | Basis |
|---|---|---|---|---|---|
| 605 | Sueldos y salarios | **Only if** any of: (a) other accumulable income; (b) told the employer in writing they will file; (c) stopped working before 31 Dec, or **two or more employers at the same time**; (d) salary from abroad or from a payer that does not withhold; (e) **salary income > $400,000**. SAT also lists indemnización / jubilación income. Otherwise the employer's annual calculation (art. 97) closes the year. | 30 Apr 2027 | Optional filing allowed (e.g., to claim personal deductions → refund). RMF 2026 rule 3.17.11 relieves single-employer salary earners with no tax due and nominal interest ≤ $20,000 **for ejercicio 2025**; ⚠️ equivalent rule for 2026 pending RMF 2027. | LISR 98 fr. III a)–e); 150 ¶2; RMF 3.17.11; SAT minisitio |
| 605 + 614 | Sueldos + intereses only | Optional when Cap. I + Cap. VI income ≤ **$400,000** and real interest ≤ **$100,000** with bank withholding applied. | 30 Apr 2027 | | LISR 150 ¶2 |
| 606 | Arrendamiento | **Yes.** Rental income (after the art. 115 deductions or the 35 % blind deduction) is accumulated in the art. 152 computation. | 30 Apr 2027 | | LISR 150 ¶1, 152 ¶1 (Cap. III) |
| 607 | Enajenación de bienes | **Yes** for the taxable gain (Cap. IV is in art. 152). The gain is spread over the years of ownership (art. 120) and the notary's provisional payment is credited. Exempt sale of the home (art. 93 fr. XIX a)) is **informative** when total income > $500,000. | 30 Apr 2027 | ⚠️ Gain computation (inflation-adjusted cost, art. 121–124) not modelled. Explanation-only. | LISR 150 ¶1 and ¶3, 152 ¶1 |
| 608 | Demás ingresos | **Yes** (Cap. IX is in art. 152). | 30 Apr 2027 | Explanation-only. | LISR 152 ¶1 |
| 611 | Dividendos | **Yes**, when the person must file for other reasons or chooses to: dividends are accumulated (Cap. VIII) and the corporate tax is creditable (art. 140). The additional 10 % on dividends is a definitive tax, not part of the annual. | 30 Apr 2027 | ⚠️ gross-up and credit not read in this session. Explanation-only. | LISR 140, 152 ¶1 |
| 612 | Actividades empresariales y profesionales | **Yes, always.** Annual tax per art. 152 on the utilidad gravable of art. 109. | 30 Apr 2027 | | LISR 109 ¶1, 150, 152 |
| 614 | Intereses | **Yes** when real interest > $100,000 (rule 3.17.2), or when filing for other income. Real interest is accumulated; the 0.90 % bank withholding is credited. | 30 Apr 2027 | | LISR 135, 136, 150 ¶2, 152; RMF 3.17.2 |
| 615 | Premios | **No** for the prize itself: the withholding is a **definitive** tax and Cap. VII is not in art. 152. **Informative**: must be reported when total income > $500,000. | — | | LISR 150 ¶1 and ¶3, 152 ¶2 |
| 616 | Sin obligaciones fiscales | **No.** | — | | — |
| 610 | Residentes en el extranjero | **No** (definitive withholding, Título V). | — | | LISR 153 |
| 621 | RIF (legacy) | **Generally no**: RIF bimonthly payments are definitive (art. 111, text in force until 2021). The SAT lists "**RIF con coeficiente de utilidad**" among those who **must** file. | 30 Apr 2027 | ⚠️ The coefficient option and its rule were not read. App: "ask your CPA". | LISR 111 (2021 text) via Transitorio 2º fr. IX DOF 12-11-2021; SAT minisitio |
| 625 | Plataformas, **definitive** option (art. 113-B) | **No** for platform income: tax paid as definitive withholding is excluded from the annual (art. 150 ¶1, 152 ¶2). If the person also has salary, the salary rules of 605 apply. **Informative** if total income > $500,000. | — | | LISR 113-B, 150, 152 ¶2 |
| 625 | Plataformas, **non-definitive** (default) | **Yes.** Platform withholdings are **provisional payments** (113-A ¶3), credited in the annual. The SAT groups platform income under "actividades empresariales" for the annual. | 30 Apr 2027 | ⚠️ Art. 152 names Cap. II **Sección I** but not Sección III (platforms). The app treats non-definitive platform income like 612 (income − authorized deductions), as the SAT does in its annual-return guide and RMF 12.3.13 ¶2 does for monthly payments. CPA to confirm. | LISR 113-A ¶3, 150; RMF 12.3.6, 12.3.13; SAT minisitio |
| 626 | RESICO PF | **Relieved** for RESICO income: monthly payments are definitive (RMF 3.13.7). Exceptions that **may** opt to file: co-owners under rules 3.13.20/3.13.21. The law's annual computation (113-F) is shown only as an informative reconciliation. **Salary / interest income of a RESICO taxpayer**: its own annual, computed separately (rule 3.13.8), under the 605/614 rules. | (30 Apr 2027 if filed) | ⚠️ 113-I expels a RESICO taxpayer who "no presente su declaración anual"; rule 3.13.7 neutralizes this for 2026 only if RMF 2027 keeps it. | LISR 113-E, 113-F, 113-I; RMF 2026 rules 3.13.7, 3.13.8; SAT minisitio |

**Income > $500,000 rule (all PF who file):** whoever files and had **total** income above $500,000, including exempt and definitively taxed income, must also report exempt items under art. 93 fr. XVII (loans, gifts, prizes per art. 90), XIX a) (sale of home), XXII (inheritances) and art. 138 prizes. Informative only, no tax effect (LISR 150 ¶3) ✅.

## A.2 Personas morales

| Key | Regime | Annual return? | Deadline | Notes | Basis |
|---|---|---|---|---|---|
| 601 | General de Ley | **Yes.** Resultado fiscal × 30 %; also determines PTU base and next year's profit coefficient. | 31 Mar 2027 | Large taxpayers may also owe ISSIF / dictamen (CFF 32-A, 32-H) ⚠️ not researched. | LISR 9, 76 fr. V, 14 fr. I |
| 626 | RESICO PM | **Yes.** Art. 9 computation on cash-flow income and deductions. | 31 Mar 2027 | | LISR 212; SAT press release 002-2026 |
| 603 | Fines no lucrativos | **Informative** annual return (remanente distribuible and each member's share). ISR only on specific items (art. 79 last ¶, 80). | **15 Feb 2027** | Explanation-only. | LISR 86 fr. III; SAT press release 002-2026 ("hasta el 16 de febrero" for 2025, which was a Monday) |
| 622 | AGAPES PM | **Yes.** Art. 9 calculation with the art. 74 exemption (20 UMA per partner, cap 200 UMA) and 30 % reduction. | 31 Mar 2027 | Explanation-only (see monthly doc §B.1). | LISR 74, 76 fr. V |
| 620 | Cooperativas de producción (optional regime) | **Yes, per partner**: the cooperative computes each partner's tax under art. 109 and the art. 152 tariff, and may defer it until distribution. | Per art. 194 | Explanation-only. | LISR 194 fr. I ✅ |
| 623 | Grupos de sociedades | **Yes**: each company computes its art. 9 result; the integradora computes the integrated result and deferral factor. | 31 Mar 2027 | Explanation-only. | LISR 64 ✅ (first fractions read) |
| 624 | Coordinados | **Yes**: the coordinado files for itself and, unless members opted out, computes members' tax (art. 109 + art. 152 tariff). | 31 Mar 2027 | Explanation-only. | LISR 72, 73 ✅ |
| 610 | Residentes en el extranjero | No. | — | | Título V |

**Dates and extensions (ejercicio 2026).** Law: PM "dentro de los tres meses siguientes" to year-end → 31 Mar 2027; PF "en el mes de abril" → 30 Apr 2027. If the last day is not a business day it moves to the next business day (CFF art. 12 ⚠️ not re-read). No extension for ejercicio 2026 is published as of 2026-10-15. For ejercicio 2025 the SAT announced no extension (press releases 002-2026, 24-2026). The day-17 RFC-digit grace of decree DOF 26-12-2013 art. 5.1 applies **only to monthly/day-17 filings**, not to the annual ✅ (read: "a más tardar el día 17 del mes siguiente").

**Paying a balance due (PF).** Up to **six monthly installments** if the return is filed and the first installment paid by the deadline (rule 3.17.3, written for ejercicio 2025; last installment by 30 Sep). ⚠️ The 2026 version will come in RMF 2027. **Refunds:** a saldo a favor ≥ $10,000 requires e.firma; below that the contraseña is enough (SAT press release 24-2026) ✅. Balance in favor can be refunded or offset; for offset it is updated for inflation (art. 152 ¶ "En los casos…") ✅.

---

# PART B: Annual calculation algorithms

## B.1 PF general algorithm (art. 152), used by 605, 606, 612, 625 non-definitive, 614, and their combinations

| Step | Rule | Basis |
|---|---|---|
| 1 | **Salaries (Cap. I):** accumulable salary income (gross minus exempt portions of art. 93: aguinaldo, prima vacacional, PTU, overtime caps…). Take it from the **CFDI de nómina / constancia** totals, not from gross pay. | LISR 94, 93 ⚠️ exempt caps not re-read |
| 2 | **612 (Cap. II Sec. I):** utilidad fiscal = income **collected** in the year − authorized deductions **paid** in the year (incl. depreciation, art. 104); − PTU paid in the year; − prior-year losses (10 years, inflation-updated) = **utilidad gravable**. A loss in 612 stays in 612 (carry-forward) and **cannot reduce other chapters**. | LISR 109 ¶1–2, fr. I |
| 3 | **625 non-definitive:** platform income − authorized deductions, combined with step 2 (⚠️ see A.1). | 113-A ¶3; SAT minisitio |
| 4 | **606 (Cap. III):** annual rent collected − (art. 115 actual deductions **or** 35 % blind deduction + predial). | LISR 115, 152 |
| 5 | **614 (Cap. VI):** real interest. **607/608/611:** explanation-only (Part A). | LISR 134, 135, 152 |
| 6 | **Sum** of steps 1–5 = *ingresos acumulables*. | LISR 152 ¶1 |
| 7 | **Minus personal deductions** (art. 151, §B.2), applying each cap and the global cap. | LISR 151 |
| 8 | **Minus colegiaturas** (stimulus, decree DOF 26-12-2013 arts. 1.8–1.10), outside the global cap. | Decree art. 1.8, 1.10 |
| 9 | Result ≤ 0 → tax 0. Otherwise apply the **2026 annual tariff** (monthly doc §A.2.2, December column = Anexo 8 C.II): `ISR = cuota fija + (base − límite inferior) × %`. | LISR 152; Anexo 8 C.II |
| 10 | **Credits:** provisional payments made in the year (612/606/625 self-paid), **ISR withheld on salaries** (art. 96: "pagos provisionales a cuenta del impuesto anual"), **10 % withheld by PM** on fees (art. 106) and rent (art. 116), **platform withholdings** (113-A ¶3), bank interest withholding (art. 135), and foreign tax / dividend credits (arts. 5, 140). | LISR 152 fr. I–II; 96 ¶1; 106; 113-A; 116; 135 |
| 11 | `saldo = ISR (step 9) − credits (step 10)`. > 0 = **saldo a cargo**; < 0 = **saldo a favor** (refund or offset only of tax actually paid or withheld). | LISR 152 |
| — | **Excluded from the sum:** exempt income and income taxed definitively (625 definitive, 615 prizes, RESICO PF). | LISR 152 ¶2 |

**Annual 2026 tariff (Anexo 8 RMF 2026, C.II, DOF 28-12-2025), re-verified verbatim in this session:**

| Lower limit | Upper limit | Fixed fee | % on excess |
|---|---|---|---|
| 0.01 | 10,135.11 | 0.00 | 1.92 |
| 10,135.12 | 86,022.11 | 194.59 | 6.40 |
| 86,022.12 | 151,176.19 | 5,051.37 | 10.88 |
| 151,176.20 | 175,735.66 | 12,140.13 | 16.00 |
| 175,735.67 | 210,403.69 | 16,069.64 | 17.92 |
| 210,403.70 | 424,353.97 | 22,282.14 | 21.36 |
| 424,353.98 | 668,840.14 | 67,981.92 | 23.52 |
| 668,840.15 | 1,276,925.98 | 125,485.07 | 30.00 |
| 1,276,925.99 | 1,702,567.97 | 307,910.81 | 32.00 |
| 1,702,567.98 | 5,107,703.92 | 444,116.23 | 34.00 |
| 5,107,703.93 | en adelante | 1,601,862.46 | 35.00 |

**Consequence for 612-only taxpayers:** because the annual tariff equals the December accumulated table, a 612-only taxpayer whose 12 provisional payments were computed on the same figures gets an annual result of ≈ 0 **before** personal deductions. Personal deductions, late-year adjustments (depreciation, PTU, losses) and other income are what move the annual result (example AN-01).

## B.2 Personal deductions (art. 151) and colegiaturas, ejercicio 2026

UMA 2026 annual = **$42,794.64** (monthly doc §0). "Salario mínimo … elevado al año" in art. 151 is read as **UMA** under the desindexation decree (DOF 27-01-2016, Transitorio 3º) ⚠️ derived, not stated in art. 151.

| # | Concept | Who can be the beneficiary | Requirements | Own cap | In global cap? | Basis |
|---|---|---|---|---|---|---|
| I | Medical, dental, psychology, nutrition fees (licensed professionals), hospital expenses | Taxpayer, spouse/concubine, ascendants and descendants in direct line whose own income < 1 annual UMA ⚠️ | CFDI; paid by **nominative cheque, transfer, credit/debit/service card** (cash not deductible) | none | **Yes** | 151 fr. I ¶1 |
| I bis | Medical expenses for incapacity/disability ≥ 50 % with official certificate | same | CFDI must state the link to the incapacity | none | **No** | 151 fr. I ¶3–4 |
| II | Funeral expenses | same | CFDI | ≤ 1 annual UMA = **$42,794.64** ⚠️ | Yes | 151 fr. II |
| III | Donations to authorized donees | — | Donee on SAT list | **7 % of prior-year (2025) ingresos acumulables** before personal deductions; donations to government ≤ 4 % within the 7 % | Yes | 151 fr. III ¶3 |
| IV | **Real** mortgage interest (home purchase), credit ≤ 750,000 UDIs | Taxpayer | Bank's constancia states the real interest; UDI method allowed (rule 3.17.4) | none beyond the 750k UDI loan cap | Yes | 151 fr. IV; RMF 3.17.4 |
| V | Voluntary / complementary retirement contributions, personal retirement plans | Taxpayer | Authorized institution (rule 3.17.5); constancia (rule 3.17.7) | **10 % of the year's ingresos acumulables**, max **5 annual UMA = $213,973.20** ⚠️ | **No** (explicitly excluded) | 151 fr. V; last ¶ |
| VI | Health insurance premiums (GMM) | Taxpayer, spouse/concubine, ascendants, descendants | CFDI | none | Yes | 151 fr. VI |
| VII | Compulsory school transport | Descendants | Separated on the CFDI; electronic payment | none | Yes | 151 fr. VII |
| VIII | Local payroll tax paid on salaries | Taxpayer | Local rate ≤ 5 % | none | Yes | 151 fr. VIII |
| Dec. | **Colegiaturas** (preescolar → bachillerato; tuition only, not enrolment fees) | Taxpayer, spouse/concubine, ascendants/descendants with income < 1 annual UMA ⚠️ | Private school with RVOE; CFDI separating tuition; electronic payment | Per **student** per year: preescolar **$14,200**, primaria **$12,900**, secundaria **$19,900**, profesional técnico **$17,100**, bachillerato **$24,500**. Two levels in one year → the higher cap. Minus scholarships. | **No** | Decree DOF 26-12-2013 arts. 1.8–1.10 ✅; ⚠️ no later amendment of the amounts found |

**Global cap (art. 151 last ¶):** sum of fr. I–IV, VI–VIII ≤ **min(5 annual UMA = $213,973.20 ; 15 % of the taxpayer's total income, including exempt income)**. Not capped: fr. V, incapacity medical expenses, colegiaturas.

**Apply caps in this order:** (1) each concept's own cap (funeral, donations 7 %, retirement 10 %/5 UMA, colegiatura per student/level); (2) global cap on the capped group; (3) add the uncapped items.

## B.3 RESICO PF annual (arts. 113-E, 113-F) — informative reconciliation only

| Step | Rule | Basis |
|---|---|---|
| 1 | Annual income **collected**, backed by CFDI, excluding IVA, **no deductions** (not even personal deductions). | 113-F ¶1 |
| 2 | Annual rate on the **whole** amount (non-marginal): ≤ 300,000 → 1.00 %; ≤ 600,000 → 1.10 %; ≤ 1,000,000 → 1.50 %; ≤ 2,500,000 → 2.00 %; ≤ 3,500,000 → 2.50 %. | 113-F table ✅ |
| 3 | Minus ISR paid in the monthly returns and the **1.25 %** withheld by PM (113-J). | 113-F ¶2 |
| 4 | **Result is informative.** Rule 3.13.7: monthly payments are **definitive** and the taxpayer is relieved of the annual return for RESICO income. The app must not present a "saldo a favor" as refundable or a "saldo a cargo" as payable. | RMF 2026 rule 3.13.7 |
| — | Salary and interest income of a RESICO taxpayer: separate annual under §B.1 (rule 3.13.8). Whether art. 151 personal deductions in that separate return can use RESICO income for the 15 % cap ⚠️ CPA. | RMF 3.13.8 |
| — | Mid-year exit to 612/606: rule 3.13.5 lets the taxpayer either keep the RESICO months definitive and file the annual for the remaining months only, or recompute from January under art. 106 crediting RESICO payments ⚠️ explanation-only. | RMF 3.13.5 |

The annual and monthly results can differ either way, because the rate is non-marginal: income bunched in a few months tends to make the monthly total higher; income just over an annual threshold makes the annual figure higher (example AN-03).

## B.4 Regime 625 platforms

| Situation | Annual treatment |
|---|---|
| Definitive option (113-B: only platform income, plus salary/interest allowed, prior-year ≤ $300,000; 5-year lock-in) | No annual for platform income. App shows "your platform withholdings were definitive". |
| Non-definitive (default, or > $300,000) | Platform income (excluding IVA) − authorized deductions → enters §B.1 step 3. Credit **all withholdings** (2.5 % goods/services in 2026 per LIF 2026 art. 25 fr. VI; 2.1 % transport/delivery; 4 % lodging) and any self-paid provisional payments (rule 12.3.6/12.3.13). |
| Leaving the definitive option | Effective from the **next** fiscal year (rule 12.3.9); cannot be re-elected (113-B). |

## B.5 PM 601 annual (arts. 9, 14, 76)

| Step | Rule | Basis |
|---|---|---|
| 1 | Ingresos acumulables of the year (**accrual** basis, incl. the accumulable annual inflation adjustment ⚠️). | LISR 9 fr. I, 16–18 |
| 2 | − Authorized deductions (incl. depreciation arts. 31–38, deductible inflation adjustment ⚠️; non-deductible exempt payroll per 28 fr. XXX). | LISR 25, 28 |
| 3 | − **PTU paid** in the year → **utilidad fiscal**. | LISR 9 fr. I |
| 4 | − Prior-year tax losses pending (10 years, updated) → **resultado fiscal**. | LISR 9 fr. II, 57 |
| 5 | × **30 %** = ISR del ejercicio. | LISR 9 ¶1 |
| 6 | − Provisional payments of the year; − interest withholding (art. 54); − creditable taxes (arts. 5, 10). | LISR 14 fr. III, 9 |
| 7 | = saldo a cargo / a favor. Due 31 Mar 2027. | LISR 9 ¶3, 76 fr. V |
| 8 | **PTU to distribute (2027):** 10 % × renta gravable = ingresos acumulables − deducciones, **without** subtracting PTU paid or losses, and subtracting from income the amounts made non-deductible by 28 fr. XXX. | LISR 9 ¶4–5; PTU 10 % ⚠️ (monthly doc §0) |
| 9 | **Profit coefficient for 2027 provisional payments** = utilidad fiscal 2026 ÷ ingresos nominales 2026 (ingresos acumulables **excluding** the accumulable inflation adjustment). The app should show it as "suggested CU for 2027", ⚠️ rounding practice 4 decimals. | LISR 14 fr. I, 14 ¶ "Los ingresos nominales…" |

## B.6 RESICO PM annual (arts. 207–212)

Same art. 9 structure, but income is **collected** and deductions **paid** (cash flow, art. 207/208), investments at art. 209 rates. Credit provisional payments and arts. 5 and 10 taxes (212 fr. I–II). PTU base = "la utilidad fiscal que resulte de conformidad a lo señalado en el artículo 9" (212 ¶2) ⚠️ whether PTU paid is subtracted for that base is a CPA point. No profit coefficient (RESICO PM provisional payments are cash-flow, art. 211). Due 31 Mar 2027.

## B.7 IVA

**No annual IVA return.** LIVA art. 5-D: "El impuesto se calculará por cada mes de calendario" and the monthly payment is the difference for the month (quarterly definitive for small rent, art. 5-F). The app's annual screen should show only an **IVA summary of the 12 months** (paid, credited, balance carried) as information, labelled "no se presenta declaración anual de IVA" ✅.

---

# PART C: Inputs the app must add

Already captured (per monthly doc): income and expenses per regime and month, IVA, PM withholdings, monthly provisional payments computed, profit coefficient (601).

| # | Input | Needed for | Source document for the user | MVP? |
|---|---|---|---|---|
| 1 | **Provisional payments actually paid** per month (amount filed, not just the app's estimate) | All credits (B.1 step 10, B.5 step 6) | Acuses de declaración | **Yes** |
| 2 | **Withholdings received**: ISR 10 % (fees/rent), 1.25 % RESICO, platform withholdings, bank interest withholding | Credits | CFDI de retenciones | **Yes** (partly derivable from income records) |
| 3 | **Salary income**: accumulable income, exempt income, ISR withheld, per employer; flag "stopped working before 31 Dec" | 605 obligation test and B.1 step 1 | CFDI de nómina / constancia de sueldos | **Yes** |
| 4 | **Personal deductions** by category (§B.2): amount, date, beneficiary relationship, payment method (cash flag), CFDI UUID | B.1 step 7 | CFDI from doctor, hospital, insurer, bank, AFORE, donee | **Yes** |
| 5 | Colegiaturas per student: level, amount tuition-only, scholarship | B.1 step 8 | School CFDI | **Yes** |
| 6 | Real mortgage interest (bank's annual figure) and retirement contributions (constancia) | §B.2 IV, V | Bank / AFORE constancia | **Yes** (as typed amounts) |
| 7 | **Prior-year (2025) ingresos acumulables** | 7 % donation cap | Previous annual return | Yes (optional; without it, donations shown unverified) |
| 8 | Total income including exempt (for the 15 % cap and the > $500k informative test) | §B.2 global cap; A.1 | User | Yes |
| 9 | **PTU paid in 2026** | 612 (art. 109), PM | Payroll | Yes (optional) |
| 10 | **Prior-year tax losses pending** (amount, year of origin) | 612, PM | Previous returns | Yes (typed; app does not inflation-update ⚠️) |
| 11 | **Depreciation of investments** (annual deduction amount) | 612, PM | CPA's fixed-asset schedule | ⚠️ typed amount only; no asset register in MVP |
| 12 | **Annual inflation adjustment** (accumulable/deductible) | PM 601 | CPA | ⚠️ **Out of MVP scope** (requires monthly debt/credit balances and INPC); typed override only |
| 13 | 625 definitive-option flag and prior-year platform total | B.4 | SAT aviso 4/PLT | Yes |
| 14 | Interest income (nominal, real, withholding) | 614 | Bank constancia | Optional |
| 15 | Foreign tax credits, dividend credits, 607/608/611 figures | Explanation-only | — | No |

---

# PART D: Worked examples (table-driven test candidates)

All MXN, 2026 annual tariff from §B.1, rounding only at the end (half-up to cents). Computed with a script against the Anexo 8 C.II table.

| ID | Case | Inputs | Computation | Expected |
|---|---|---|---|---|
| **AN-01** | **PF 612 only**, with medical expenses | Collected 900,000; deductions paid 400,000 → utilidad 500,000. Personal deductions: medical 40,000 (cap min(213,973.20; 15 % × 900,000 = 135,000) → OK). Provisional payments 55,773.86; 10 % withheld by PM on 300,000 = 30,000. | Base 460,000 → row 7: 67,981.92 + (460,000 − 424,353.98) × 23.52 % = **76,365.86**. Credits 85,773.86 (= December tariff on 500,000). | **Saldo a favor 9,408.00** (= 40,000 × 23.52 %). Without the deduction: 0.00. |
| **AN-02** | **PF 612 + 606 + sueldos**, several personal deductions | Salary: accumulable 360,000 (gross 380,000, 20,000 exempt), ISR withheld 54,235.91. 612: collected 400,000 − deductions 150,000 = 250,000; provisional 20,000; 10 % withheld 15,000. 606: rent 240,000 to a PF, blind 35 % = 84,000 + predial 6,000 → 150,000; provisional 18,000. Personal: medical 30,000; GMM 25,000; real mortgage interest 40,000; donations 10,000 (2025 acumulables 700,000 → 7 % cap 49,000 OK); retirement plan 50,000; primaria tuition 30,000. | Acumulables 760,000. Global cap: total income 1,020,000 × 15 % = 153,000 < 213,973.20 → capped group 105,000 OK. Retirement: min(50,000; 10 % × 760,000; 213,973.20) = 50,000 (outside cap). Colegiatura: 12,900. Base 760,000 − 105,000 − 50,000 − 12,900 = 592,100 → row 7: **107,435.78**. Credits 54,235.91 + 20,000 + 15,000 + 18,000 = 107,235.91. | **Saldo a cargo 199.87** |
| **AN-03** | **RESICO PF**, informative reconciliation | (a) Jan–Jun 20,000/month (1.00 %), Jul–Dec 60,000/month (1.50 %); monthly paid incl. retentions 1,200 + 5,400 = 6,600. (b) 11 months × 25,000 (1.00 %) + 1 month × 25,001 (1.10 %): monthly 2,750 + 275.01 = 3,025.01. | (a) Annual 480,000 → 1.10 % = 5,280.00. (b) Annual 300,001 → 1.10 % = 3,300.01. | (a) difference −1,320.00; (b) +275.00. **Both informative only**: rule 3.13.7 makes the monthly payments definitive (no refund, no balance due). |
| **AN-04** | **625 non-definitive** (goods), platform income only | Platform sales 480,000 excl. IVA (> 300,000 → no definitive option); withheld 2.5 % = 12,000; deductible purchases 300,000; no other income, no personal deductions. | Base 180,000 → row 5: 16,069.64 + (180,000 − 175,735.67) × 17.92 % = **16,833.81**. | **Saldo a cargo 4,833.81** ⚠️ treatment per A.1 |
| **AN-05** | **605, two employers at once** (obligated by 98 fr. III c)) | Employer A accumulable 300,000, ISR 40,000; employer B 150,000, ISR 8,000. | Base 450,000 → row 7: **74,013.86**; credits 48,000. | **Saldo a cargo 26,013.86** (each employer withheld as if it were the only one). |
| **AN-06** | **PM 601** | Acumulables = nominales 10,000,000; deductions 8,500,000; PTU paid 120,000; losses pending 200,000; provisional payments 330,000; interest withholding 2,000. | Utilidad fiscal 1,380,000; resultado fiscal 1,180,000 × 30 % = 354,000. | **Saldo a cargo 22,000.00.** PTU base 1,500,000 → **PTU 150,000.00**. **CU 2027 = 0.1380**. |
| **AN-07** | **RESICO PM** | Collected 6,000,000; deductions paid 4,800,000; PTU paid 50,000; losses 0; provisional payments 360,000. | Utilidad fiscal 1,150,000 × 30 % = 345,000. | **Saldo a favor 15,000.00.** PTU base per 212 ¶2: 1,150,000 → 115,000 ⚠️ (1,200,000 → 120,000 if PTU paid is not subtracted; CPA). |
| **AN-08** | Personal-deduction cap binds | 612 utilidad 300,000 = total income 300,000; medical 60,000; GMM 20,000. | Cap min(213,973.20; 45,000) = 45,000 → base 255,000 → row 6: 22,282.14 + (255,000 − 210,403.70) × 21.36 % = 31,807.91. | Deduction applied **45,000** (not 80,000); ISR **31,807.91** before credits. |

---

# PART E: MVP recommendation

**Estimate (numbers shown, labelled "Estimación"):**

1. PF **612**, **606**, **625 non-definitive**, **605** (salary from constancia) and their combinations, via §B.1, with personal deductions §B.2 (all eight categories plus colegiaturas, caps enforced) and credits for provisional payments and withholdings. Output: base, ISR, credits, saldo a cargo / a favor, deadline 30 Apr 2027, and the 605 "are you obligated?" test (98 fr. III, 150 ¶2).
2. PM **601** (§B.5) with user-typed depreciation, PTU paid, losses and an inflation-adjustment override; outputs saldo, **PTU to distribute** and **suggested CU 2027**.
3. **RESICO PM** (§B.6).

**Explanation-only (no number, or a clearly informative one):**

- **RESICO PF**: "Your monthly payments are definitive; you are relieved of the annual return (RMF 3.13.7)", plus the informative 113-F reconciliation. If the user also has salary, estimate the salary annual separately.
- **625 definitive**, **615**, **616**, **610**: "no annual for this income"; show the > $500,000 informative-reporting warning when it applies.
- **603** (informative return by 15 Feb), **607**, **608**, **611**, **614** (beyond simple real interest), **621**, **622**, **620**, **623**, **624**: describe the obligation and the deadline; "your CPA computes it".
- **IVA**: 12-month summary, "no annual IVA return".

**Mandatory disclaimer (annual screen):** "Estimación informativa de la declaración anual 2026. No sustituye la declaración que presenta tu contador ante el SAT. Basada en LISR arts. 9, 109, 113-F, 150–152, Anexo 8 RMF 2026 y RMF 2026; las reglas para el ejercicio 2026 pueden cambiar con la RMF 2027."

**⚠️ Items for the CPA:**

1. RMF 2027: whether rule 3.13.7 (RESICO PF relief), 3.17.3 (installments) and 3.17.11 (salary relief) are renewed for ejercicio 2026, and any deadline change.
2. Non-definitive 625 income: confirm it is computed with Sección I deductions inside art. 152 (art. 152 names only Sección I).
3. Reading "salario mínimo elevado al año" in art. 151 (funeral cap, dependants' income, retirement cap) as UMA.
4. Donations: 7 % base = 2025 acumulables "antes de aplicar las deducciones" personales; and whether donations also count inside the global cap (text says yes).
5. Salary exempt portions (art. 93) and the employment subsidy: the app trusts the CFDI/constancia totals.
6. Loss carry-forward and PTU inflation updates (arts. 57, 109 fr. I): the app takes typed, already-updated figures.
7. Annual inflation adjustment and depreciation for PM/612: out of MVP; typed overrides only.
8. RESICO PM PTU base (AN-07) and profit-coefficient rounding (AN-06).
9. RIF "con coeficiente de utilidad" taxpayers (SAT says they must file).
10. RESICO PF + salary: personal deductions and the 15 % cap in the separate salary annual (rule 3.13.8).
11. ISSIF / dictamen obligations for larger PM (CFF 32-A, 32-H) — not researched.

---

# PART F: Sources (accessed 2026-10-15)

| # | Source | What was read | URL |
|---|---|---|---|
| F1 | LISR, current text (last reform DOF 01-04-2024) | arts. 9, 14, 64, 72–73, 76 fr. V, 86 fr. III, 96, 98, 106, 109, 113-A, 113-B, 113-E, 113-F, 113-I, 113-J, 150, 151, 152, 194, 212 | https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf |
| F2 | LIVA, current text | art. 5-D, 5-F | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIVA.pdf |
| F3 | Anexo 8 RMF 2026 (DOF 28-12-2025), C.II annual tariff 2026 | full table | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/anexos/Anexo-8-RMF-2026_DOF-28122025.pdf |
| F4 | RMF 2026 (DOF 28-12-2025) | rules 3.13.1–3.13.19, 3.17.1–3.17.12, 12.3.3–12.3.17 | https://dof.gob.mx (RMF 2026, 28-12-2025); SAT copy: https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/ |
| F5 | Decreto que compila diversos beneficios fiscales (DOF 26-12-2013) | arts. 1.8–1.10 (colegiaturas), 5.1 (RFC-digit grace) | https://dof.gob.mx/nota_detalle.php?codigo=5328028&fecha=26/12/2013 |
| F6 | LIF 2026 (DOF 07-11-2025), art. 25 fr. VI | platform withholding 2.5 % (via monthly doc) | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIF_2026.pdf |
| F7 | SAT minisitio "Declaración Anual 2025 Personas — quiénes deben presentarla" | obligation list, RESICO relief, RIF con coeficiente | https://www.sat.gob.mx/minisitio/DeclaracionAnual/Personas/quienes_deben_presentarla.html |
| F8 | SAT press release 24-2026 (1 Apr 2026) | April deadline, refunds ≥ $10,000 need e.firma, RESICO relief | https://www.gob.mx/sat/prensa/inicia-declaracion-anual-2025-de-personas-con-mas-facilidades-para-contribuyentes-24-2026?idiom=es |
| F9 | SAT press release 002-2026 | PM deadline 31 March; non-profits 16 Feb 2026 | https://www.gob.mx/sat/prensa/invita-sat-a-empresas-a-presentar-su-declaracion-anual-002-2026?idiom=es |
| F10 | INEGI UMA 2026 (via monthly doc) | annual UMA $42,794.64 | https://www.inegi.org.mx/contenidos/saladeprensa/boletines/2026/uma/uma2026.pdf |

**Not retrieved / not re-read in this session:** CFF art. 12 (deadline roll-over), LISR arts. 93 (salary exemptions), 120–124 (gain on sale), 134–136 (real interest), 140 (dividends), 57 (losses update), RLISR art. 257; desindexation decree DOF 27-01-2016; RMF 2027 (not yet published).
