# Mexican tax regimes 2026: monthly ISR and IVA estimation rules

> **Purpose.** This is the research basis for the app's *informational* estimate of monthly provisional ISR and monthly IVA. It is not tax advice.
> **Cut-off date:** 2026-10-13. **Fiscal year:** 2026.
> **Status legend:** ✅ confirmed against the primary source in this session · ⚠️ secondary source or not re-verified (do not hard-code without review).

## 0. Key 2026 parameters

| Parameter | 2026 value | Legal basis | Status |
|---|---|---|---|
| Monthly ISR tariff, PF (art. 96) | 11 brackets, 1.92 %–35 % (§A.2.1) | LISR art. 96 + Anexo 8 RMF 2026 (DOF 28-12-2025), apartado B.V | ✅ updated by art. 152: cumulative inflation > 10 % |
| Accumulated monthly tariffs, art. 106 (Jan–Dec) | §A.2.2 | Anexo 8 RMF 2026, B.VI | ✅ |
| Annual tariff 2026 (arts. 97 and 152) | = December table | Anexo 8 RMF 2026, C.II | ✅ |
| RESICO PF monthly and annual rates | 1.00 %–2.50 % | LISR arts. 113-E and 113-F | ✅ unchanged since 2022 |
| Corporate ISR rate | 30 % | LISR art. 9 | ✅ |
| ISR withheld by platforms, goods and services (PF and PM) | **2.5 %** (was 1 %) | **LIF 2026 art. 25, fr. VI** (overrides LISR art. 113-A fr. III) | ✅ |
| ISR withheld by platforms, ground transport and delivery | 2.1 % | LISR art. 113-A fr. I | ✅ |
| ISR withheld by platforms, lodging | 4 % | LISR art. 113-A fr. II | ✅ |
| Annual ISR withholding rate on interest (arts. 54 and 135) | 0.90 % | LIF 2026 art. 24 | ✅ |
| UMA 2026, from 1 Feb 2026 to 31 Jan 2027 | daily $117.31 · monthly $3,566.22 · annual $42,794.64 | INEGI communiqué 1/26 (8-01-2026), DOF 9-01-2026 | ✅ (in January 2026 the 2025 UMA of $113.14 still applied) |
| IVA general rate | 16 % | LIVA art. 1 | ✅ |
| IVA in the border regions | 8 % (a credit of 50 % of the rate) | Decrees for the northern and southern border regions; extension published DOF 31-12-2025, **in force until 31-12-2026** | ✅ extension confirmed (only the transitory article was read) |
| PTU | 10 % of taxable income | Resolución de la Sexta Comisión Nacional (DOF 18-09-2020) | ⚠️ secondary sources only |

**Laws current as of 2026-10-13 (Cámara de Diputados compilations):** LISR, last reform DOF 01-04-2024. LIVA, last reform DOF 12-11-2021. CFF, last reform DOF 09-04-2026. **The 2026 package did not amend the LISR or the LIVA.** All 2026 rate changes come through the **Ley de Ingresos de la Federación 2026** (DOF 07-11-2025) and the RMF 2026.

---

# PART A: Personas físicas (individuals)

## A.1 Regime catalog

The catalog is SAT `c_RegimenFiscal` from `catCFDI_V_4_20261001.xls`, downloaded 2026-10-13. All keys have been in force since 2022-01-01, with no end date.

| Key | Regime | Who | Periodicity | Fits "income & expenses → estimate"? |
|---|---|---|---|---|
| 605 | Sueldos y salarios y asimilados | Employees | Employer withholds monthly (art. 96); annual return in April (art. 150) when required | **No.** The employer withholds the tax, so there is no self-assessed monthly payment. At most the app could show an annual estimate. |
| 606 | Arrendamiento | Lessors of real estate | Monthly provisional payment (art. 116). Quarterly optional if only rental income ≤ 10 monthly minimum wages, which are read as UMA | **Yes.** Income minus actual deductions (art. 115) or the 35 % blind deduction. Monthly, not year-to-date. |
| 607 | Enajenación o adquisición de bienes | Individuals selling real estate or other assets | Per transaction; the notary computes it (arts. 120–132) | **No.** The tax depends on a single transaction, inflation-adjusted cost and the notary. |
| 608 | Demás ingresos | Other income (art. 141 ff.) | Mostly annual, with some withholding | **No.** |
| 610 | Residentes en el extranjero sin EP | Non-residents | Definitive withholding (Título V) | **No.** |
| 611 | Dividendos | Shareholders | Annual. The 10 % withholding is definitive (art. 140) | **No.** |
| 612 | Actividades empresariales y profesionales | Businesses and independent professionals | Monthly provisional payment, year-to-date (art. 106); annual return in April | **Yes.** This is the core use case. |
| 614 | Intereses | Interest income | Withholding by the bank (0.90 % for 2026) | **No.** |
| 615 | Premios | Prize income | Withholding by the payer | **No.** |
| 616 | Sin obligaciones fiscales | Taxpayers without obligations | None | **No.** |
| 621 | Incorporación Fiscal (RIF, legacy) | Only taxpayers who stayed in RIF after the 2022 reform, until their 10 years end | **Bimonthly, definitive** (art. 111 LISR as in force until 2021) | **Partial.** The formula is simple, but the yearly reduction table and the closed population make it niche. |
| 625 | Plataformas tecnológicas | Individuals selling or providing services through apps | The platform withholds monthly. Withholding is either provisional, or definitive if they opt in (113-A, 113-B) | **Yes.** Withholding rates and the definitive option. |
| 626 | RESICO PF | Individuals with business, professional or rental income ≤ $3.5M a year | Monthly, definitive in practice (113-E), plus an annual return in April (113-F) | **Yes.** No deductions, with a rate applied to income collected. |

Codes 628/629/630, which you mentioned, are **not** in the current CFDI 4.0 `c_RegimenFiscal` sheet. 610 and 626 apply to both PF and PM.

## A.2 ISR algorithms

### A.2.1 Monthly tariff, art. 96 (used by 606 and 605; also the base for every other table)

Source: Anexo 8 RMF 2026, B.V. The art. 116 rental table (Anexo 8) is identical.

| # | Lower limit | Upper limit | Fixed fee | % on excess |
|---|---|---|---|---|
| 1 | 0.01 | 844.59 | 0.00 | 1.92 |
| 2 | 844.60 | 7,168.51 | 16.22 | 6.40 |
| 3 | 7,168.52 | 12,598.02 | 420.95 | 10.88 |
| 4 | 12,598.03 | 14,644.64 | 1,011.68 | 16.00 |
| 5 | 14,644.65 | 17,533.64 | 1,339.14 | 17.92 |
| 6 | 17,533.65 | 35,362.83 | 1,856.84 | 21.36 |
| 7 | 35,362.84 | 55,736.68 | 5,665.16 | 23.52 |
| 8 | 55,736.69 | 106,410.50 | 10,457.09 | 30.00 |
| 9 | 106,410.51 | 141,880.66 | 25,659.23 | 32.00 |
| 10 | 141,880.67 | 425,641.99 | 37,009.69 | 34.00 |
| 11 | 425,642.00 | and above | 133,488.54 | 35.00 |

**Formula:** `ISR = fixed fee + (base − lower limit) × rate`, for the row whose limits contain the base.

### A.2.2 Accumulated tariffs for regime 612, art. 106 (Anexo 8 RMF 2026, B.VI)

Within each month, every upper limit equals the next row's lower limit minus 0.01; this was verified for all 12 tables. **Store the published values verbatim.** December is *not* exactly 12× January (for example, row 2's fixed fee is 194.59, while 12 × 16.22 = 194.64), so the app must not derive the tables by multiplication.

**Lower limits**

| # | % | Jan | Feb | Mar | Apr | May | Jun | Jul | Aug | Sep | Oct | Nov | Dec |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 1.92 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 | 0.01 |
| 2 | 6.40 | 844.60 | 1,689.19 | 2,533.78 | 3,378.37 | 4,222.96 | 5,067.55 | 5,912.14 | 6,756.73 | 7,601.32 | 8,445.91 | 9,290.50 | 10,135.12 |
| 3 | 10.88 | 7,168.52 | 14,337.03 | 21,505.54 | 28,674.05 | 35,842.56 | 43,011.07 | 50,179.58 | 57,348.09 | 64,516.60 | 71,685.11 | 78,853.62 | 86,022.12 |
| 4 | 16.00 | 12,598.03 | 25,196.05 | 37,794.07 | 50,392.09 | 62,990.11 | 75,588.13 | 88,186.15 | 100,784.17 | 113,382.19 | 125,980.21 | 138,578.23 | 151,176.20 |
| 5 | 17.92 | 14,644.65 | 29,289.29 | 43,933.93 | 58,578.57 | 73,223.21 | 87,867.85 | 102,512.49 | 117,157.13 | 131,801.77 | 146,446.41 | 161,091.05 | 175,735.67 |
| 6 | 21.36 | 17,533.65 | 35,067.29 | 52,600.93 | 70,134.57 | 87,668.21 | 105,201.85 | 122,735.49 | 140,269.13 | 157,802.77 | 175,336.41 | 192,870.05 | 210,403.70 |
| 7 | 23.52 | 35,362.84 | 70,725.67 | 106,088.50 | 141,451.33 | 176,814.16 | 212,176.99 | 247,539.82 | 282,902.65 | 318,265.48 | 353,628.31 | 388,991.14 | 424,353.98 |
| 8 | 30.00 | 55,736.69 | 111,473.37 | 167,210.05 | 222,946.73 | 278,683.41 | 334,420.09 | 390,156.77 | 445,893.45 | 501,630.13 | 557,366.81 | 613,103.49 | 668,840.15 |
| 9 | 32.00 | 106,410.51 | 212,821.01 | 319,231.51 | 425,642.01 | 532,052.51 | 638,463.01 | 744,873.51 | 851,284.01 | 957,694.51 | 1,064,105.01 | 1,170,515.51 | 1,276,925.99 |
| 10 | 34.00 | 141,880.67 | 283,761.33 | 425,641.99 | 567,522.65 | 709,403.31 | 851,283.97 | 993,164.63 | 1,135,045.29 | 1,276,925.95 | 1,418,806.61 | 1,560,687.27 | 1,702,567.98 |
| 11 | 35.00 | 425,642.00 | 851,283.99 | 1,276,925.98 | 1,702,567.97 | 2,128,209.96 | 2,553,851.95 | 2,979,493.94 | 3,405,135.93 | 3,830,777.92 | 4,256,419.91 | 4,682,061.90 | 5,107,703.93 |

**Fixed fees**

| # | Jan | Feb | Mar | Apr | May | Jun | Jul | Aug | Sep | Oct | Nov | Dec |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 | 0.00 |
| 2 | 16.22 | 32.44 | 48.66 | 64.88 | 81.10 | 97.32 | 113.54 | 129.76 | 145.98 | 162.20 | 178.42 | 194.59 |
| 3 | 420.95 | 841.90 | 1,262.85 | 1,683.80 | 2,104.75 | 2,525.70 | 2,946.65 | 3,367.60 | 3,788.55 | 4,209.50 | 4,630.45 | 5,051.37 |
| 4 | 1,011.68 | 2,023.36 | 3,035.04 | 4,046.72 | 5,058.40 | 6,070.08 | 7,081.76 | 8,093.44 | 9,105.12 | 10,116.80 | 11,128.48 | 12,140.13 |
| 5 | 1,339.14 | 2,678.28 | 4,017.42 | 5,356.56 | 6,695.70 | 8,034.84 | 9,373.98 | 10,713.12 | 12,052.26 | 13,391.40 | 14,730.54 | 16,069.64 |
| 6 | 1,856.84 | 3,713.68 | 5,570.52 | 7,427.36 | 9,284.20 | 11,141.04 | 12,997.88 | 14,854.72 | 16,711.56 | 18,568.40 | 20,425.24 | 22,282.14 |
| 7 | 5,665.16 | 11,330.32 | 16,995.48 | 22,660.64 | 28,325.80 | 33,990.96 | 39,656.12 | 45,321.28 | 50,986.44 | 56,651.60 | 62,316.76 | 67,981.92 |
| 8 | 10,457.09 | 20,914.18 | 31,371.27 | 41,828.36 | 52,285.45 | 62,742.54 | 73,199.63 | 83,656.72 | 94,113.81 | 104,570.90 | 115,027.99 | 125,485.07 |
| 9 | 25,659.23 | 51,318.46 | 76,977.69 | 102,636.92 | 128,296.15 | 153,955.38 | 179,614.61 | 205,273.84 | 230,933.07 | 256,592.30 | 282,251.53 | 307,910.81 |
| 10 | 37,009.69 | 74,019.38 | 111,029.07 | 148,038.76 | 185,048.45 | 222,058.14 | 259,067.83 | 296,077.52 | 333,087.21 | 370,096.90 | 407,106.59 | 444,116.23 |
| 11 | 133,488.54 | 266,977.08 | 400,465.62 | 533,954.16 | 667,442.70 | 800,931.24 | 934,419.78 | 1,067,908.32 | 1,201,396.86 | 1,334,885.40 | 1,468,373.94 | 1,601,862.46 |

The **annual 2026 tariff** (arts. 97 and 152; Anexo 8 C.II) is **identical to the December table**. For comparison, the annual 2025 tariff ended at a lower limit of 4,511,707.38 and a fixed fee of 1,414,947.85.

### A.2.3 Regime 612: Actividades empresariales y profesionales (LISR arts. 100–110)

| Step | Rule | Basis |
|---|---|---|
| 1 | Year-to-date income **actually collected** (cash basis, art. 102), excluding IVA. | art. 102, art. 106 ¶1 |
| 2 | Minus year-to-date authorized deductions **actually paid** (art. 103: purchases, expenses, investments via depreciation, interest, IMSS employer fees, local tax), each meeting art. 105 and art. 27 (CFDI, strictly necessary, and so on). | arts. 103, 105 |
| 3 | Minus PTU paid in the year and prior-year tax losses not yet applied. | art. 106 ¶1 |
| 4 | = **Base**. If ≤ 0, the ISR is 0; the app should carry the loss forward within the year automatically because the figures are year-to-date. | |
| 5 | Apply the **accumulated tariff for month n** (§A.2.2). | art. 106 ¶2–3 |
| 6 | Minus provisional payments already made this year (months 1…n−1). | art. 106 ¶4 |
| 7 | Minus year-to-date **10 % ISR withheld** by personas morales on professional fees (`0.10 × fees billed to PM, excluding IVA`). | art. 106 ¶5 |
| 8 | Result > 0 is the payment. Result < 0 means no payment; the surplus is credited in later months or the annual return. | |

Deadline: day 17 of the following month (art. 106). Annual return: April (art. 150), using the annual tariff, personal deductions (art. 151, capped at **min(5 annual UMA = $213,973.20; 15 % of total income)**), and credit for provisional payments.

Agricultural, livestock, forestry and fishing (AGAPES) individuals in 612: income up to 40 annual UMA is exempt, and there is a 40 % reduction (art. 74, ¶11–12). This is out of scope for v1 (§A.5).

### A.2.4 Regime 606: Arrendamiento (arts. 114–118)

| Step | Rule | Basis |
|---|---|---|
| 1 | Rent **for the month** (not year-to-date), excluding IVA. | art. 116 ¶2 |
| 2a | Actual deductions: property tax (predial), maintenance, water, real interest, salaries, insurance, construction investments. | art. 115 fr. I–VI |
| 2b | **or the blind deduction: 35 % of income + predial** paid (the option replaces 2a). | art. 115 ¶2 |
| 3 | Base = income − deductions. Apply the **monthly** tariff (§A.2.1; art. 116 table in Anexo 8). | art. 116 ¶2 |
| 4 | Minus the **10 % ISR withheld** by a PM tenant (`0.10 × rent`, with no deductions). | art. 116 ¶5 |
| 5 | Quarterly option: if rental is the only income and it is ≤ 10 monthly minimum wages, read as UMA = **$35,662.20 a month** (Feb 2026 onward). In that case use the quarterly table in Anexo 8 (not transcribed here). | art. 116 ¶3; desindexation decree DOF 27-01-2016, Transitorio 3º ⚠️ |

### A.2.5 Regime 626: RESICO PF (arts. 113-E to 113-J)

| Step | Rule | Basis |
|---|---|---|
| 1 | Monthly income **actually collected**, backed by CFDI, **excluding IVA**, with **no deductions**. | 113-E ¶5 |
| 2 | Look up the rate. It is **not marginal**: the rate applies to the whole amount. | 113-E table |
| 3 | ISR = income × rate. | |
| 4 | Minus the **1.25 % withheld** by personas morales on payments to the taxpayer (excluding IVA). | 113-J |
| 5 | Payment by day 17. Annual return in April, using the annual table and crediting monthly payments and withholdings. | 113-E ¶4, 113-F |

| Monthly income collected, up to (MXN) | Rate | Annual income, up to (MXN) | Rate |
|---|---|---|---|
| 25,000.00 | 1.00 % | 300,000.00 | 1.00 % |
| 50,000.00 | 1.10 % | 600,000.00 | 1.10 % |
| 83,333.33 | 1.50 % | 1,000,000.00 | 1.50 % |
| 208,333.33 | 2.00 % | 2,500,000.00 | 2.00 % |
| 3,500,000.00 | 2.50 % | 3,500,000.00 | 2.50 % |

**Eligibility and exits (113-E):**

| Item | Rule |
|---|---|
| Income cap | Prior-year income ≤ **$3,500,000** (annualized for new taxpayers). |
| Exceeding the cap mid-year | Move to 612 or 606 **from the next month**. The taxpayer may return when prior-year income is ≤ $3.5M and they are compliant. |
| Missing **3 or more** monthly payments in a calendar year | Expelled from RESICO (113-I). An expelled taxpayer can never return to RESICO. |
| Who may not use RESICO | Shareholders of a PM, related parties (art. 90), taxpayers with preferential-regime income (REFIPRES), and those with art. 94 fr. III–VI assimilated income. |
| Combined income | Salary and interest income (Cap. I and VI) are allowed if total prior-year income ≤ $3.5M. |
| Exclusive AGAPES activity | ISR-exempt up to **$900,000** collected. |

### A.2.6 Regime 625: Plataformas tecnológicas (arts. 113-A to 113-D; LIF 2026 art. 25 fr. VI)

| Activity through the platform | ISR withheld on income excluding IVA | IVA withheld | Basis |
|---|---|---|---|
| Ground passenger transport and delivery | **2.1 %** | 50 % of the IVA | 113-A fr. I; LIVA 18-J fr. II a) |
| Lodging | **4 %** | 50 % of the IVA | 113-A fr. II; LIVA 18-J fr. II a) |
| Sale of goods and other services | **2.5 % (2026)** (1 % until 2025) | 50 % of the IVA | LIF 2026 art. 25 fr. VI; LIVA 18-J fr. II a) |
| No RFC given to the platform | **20 %** | **100 %** | 113-C fr. IV; 18-J fr. II a) |

| Option | Condition | Effect |
|---|---|---|
| Withholding as **definitive ISR** | Only platform income (plus salary or interest allowed) and **prior-year income ≤ $300,000** | No further ISR and no deductions on that income. (113-B) |
| Directly collected income (not through the app) | Total income ≤ $300,000 | Self-pay the same rates on direct collections; the payment is definitive. (113-A last ¶) |
| Withholding as **definitive IVA** | Same $300k test | Directly collected IVA is paid at **8 %** of those collections. (LIVA 18-L, 18-M) |
| Otherwise | | The withholding is a **provisional payment**. The taxpayer computes under 612 (or RESICO) and credits the withholding. |

### A.2.7 Regime 621: RIF (legacy; art. 111 LISR as in force until 31-12-2021; Transitorio Segundo fr. IX, DOF 12-11-2021)

- Bimonthly, definitive payment: base = income collected − deductions − PTU for the bimester. Apply the bimonthly table (Anexo 8 B.VII; it is **identical to the February accumulated table** in §A.2.2). Then reduce the ISR by the % for the taxpayer's year in RIF.
- ⚠️ The reduction table (100/90/80/70/60/50/40/30/20/10 % for years 1–10) comes from the repealed art. 111 and was **not re-fetched** in this session.
- RIF closed to new taxpayers in 2022, so the youngest remaining RIF taxpayers are in year 5 or 6 in 2026. Low priority.

## A.3 IVA for individuals (612, 606, 626, 625, 621)

| Concept | Rule | Basis |
|---|---|---|
| Rates | 16 % general. 0 % for unprocessed food, patent medicines, non-carbonated water, exports and others. Exempt: land, housing, medical services by licensed PF, education, residential rent, and others. | LIVA arts. 1, 2-A, 9, 15, 20, 29 |
| Border region 8 % | A 50 % credit on the 16 % rate. Applies to sales of goods, services and rentals delivered **in** the northern or southern region. Requires filing a SAT notice; new taxpayers file it with their RFC registration. Excluded: real estate, intangibles, digital content, and taxpayers on the CFF 69 / 69-B lists. In force until 31-12-2026. | Northern border decree DOF 31-12-2018, arts. 11–13; extension DOF 31-12-2025 |
| Monthly payment | IVA charged and **actually collected** in the month − creditable IVA **actually paid** in the month − IVA withheld from the taxpayer in the month. **Not year-to-date.** | LIVA arts. 1, 1-B, 5-D |
| Creditable IVA | Expense deductible for ISR, IVA shown separately on a CFDI, actually paid, and IVA withheld (by the taxpayer as payer) already paid to SAT. For mixed taxed and exempt activity, credit only in proportion. | LIVA art. 5 fr. I–V |
| Surplus | A balance in favor can be credited in later months, or a refund or offset can be requested. | LIVA art. 6 |
| Withholding by a PM | 2/3 of the IVA (= 10.6667 % of the price) on PF professional services, commissions and rentals. 4 % of the price on ground freight of goods (from PF **or** PM). **100 %** of the IVA on desperdicios (no reduction, from anyone). The old 6 % on personnel services (subcontratación, LIVA 1-A fr. IV) was **repealed** (DOF 23-04-2021) and no longer applies. Full matrix: [withholdings-matrix-2026.md](withholdings-matrix-2026.md). | LIVA 1-A fr. II a)–d), fr. IV (derogada); RLIVA art. 3 fr. I–II |
| Platforms | 50 % withheld (100 % without RFC). Definitive under 18-M if ≤ $300k. | LIVA 18-J, 18-L, 18-M |
| Quarterly IVA, rental | Follows the quarterly ISR option for rent (art. 116), definitive. | LIVA 5-F |
| RESICO PF | No special IVA regime: the general rules apply. | |

## A.4 Deadlines, annual return and limits (PF)

| Item | Rule | Basis |
|---|---|---|
| Monthly ISR and IVA | Day **17** of the following month | LISR arts. 106, 113-E, 116; LIVA 5-D |
| Grace by 6th digit of RFC (+1 to +5 business days) | Decree that compiles tax benefits, DOF 26-12-2013, art. 5.1 | ⚠️ not re-verified for 2026 |
| Annual return | **April** of the following year (PF, RESICO PF) | LISR arts. 150, 113-F |
| Personal deductions cap | min(5 × $42,794.64 = **$213,973.20**; 15 % of income) | art. 151 last ¶ |
| RESICO PF limit | $3.5M (§A.2.5) | 113-E |
| Platform definitive option | ≤ $300k | 113-B; LIVA 18-L |

## A.5 Worked examples (PF), for table-driven tests

All amounts are in MXN, rounded to 2 decimals at the end only.

| ID | Regime / month | Inputs | Computation | Expected |
|---|---|---|---|---|
| PF-612-ISR-01 | 612, **March 2026** | YTD income Jan–Mar 90,000; YTD deductions 30,000; prior payments (Jan+Feb) 4,000; YTD 10 % withholding 3,000 (on 30,000 of fees billed to PM) | Base 60,000 → March row 6: 5,570.52 + (60,000 − 52,600.93) × 21.36 % = **7,150.96** | ISR to pay **150.96** |
| PF-612-ISR-02 | 612, **October 2026** | YTD base 250,000; prior payments 30,000; withholding 0 | October row 6: 18,568.40 + (250,000 − 175,336.41) × 21.36 % = **34,516.54** | ISR to pay **4,516.54** |
| PF-612-ISR-03 | 612, any month | YTD deductions > YTD income | Base ≤ 0 | ISR **0.00** |
| PF-612-IVA-01 | 612, March | Collected in March 30,000 + IVA 4,800 (10,000 of it billed to a PM); IVA paid on expenses 1,600 | IVA withheld = 2/3 × 1,600 = 1,066.67. IVA = 4,800 − 1,600 − 1,066.67 | IVA to pay **2,133.33** |
| PF-626-ISR-01 | RESICO, any month | Collected 60,000 (20,000 from a PM) | 60,000 ≤ 83,333.33 → 1.5 % = 900.00; withholding 1.25 % × 20,000 = 250.00 | ISR to pay **650.00** |
| PF-626-ISR-02 | RESICO, boundary | Collected 25,000.00 / 25,000.01 | 1.0 % / 1.1 % (non-marginal) | **250.00 / 275.00** |
| PF-626-IVA-01 | RESICO | IVA charged 9,600; creditable 2,000; withheld 2/3 × 3,200 = 2,133.33 | 9,600 − 2,000 − 2,133.33 | IVA to pay **5,466.67** |
| PF-606-ISR-01 | Rent, monthly, blind deduction, tenant is a PF | Rent 20,000; predial 0 | Base 20,000 − 7,000 = 13,000 → row 4: 1,011.68 + (13,000 − 12,598.03) × 16 % | ISR to pay **1,076.00** |
| PF-606-ISR-02 | Same, tenant is a PM | Withholding 10 % = 2,000 | 1,076.00 − 2,000 | Pay **0.00**; credit balance **924.00** |
| PF-606-IVA-01 | Commercial rent to a PM | IVA 3,200; withheld 2/3 = 2,133.33; creditable 0 | 3,200 − 2,133.33 | **1,066.67** (residential rent: exempt → **0**) |
| PF-625-01 | Platform goods, definitive option | Sales 30,000 excluding IVA, all through the app | ISR withheld 2.5 % = 750.00; IVA 4,800, 50 % withheld = 2,400.00 | Additional ISR **0**. Additional IVA: under 18-M the 50 % withholding is definitive, **0** ⚠️ confirm 18-M effect |
| PF-625-02 | Lodging | 15,000 | 4 % | ISR withheld **600.00** |
| PF-621-01 | RIF, Sep–Oct, year 6 | Income 100,000; deductions 60,000 | Base 40,000 → bimonthly row 6: 3,713.68 + 4,932.71 × 21.36 % = 4,767.31; −50 % | **2,383.66** ⚠️ reduction % not re-verified |

---

# PART B: Personas morales (legal entities)

## B.1 Regime catalog

| Key | Regime | Who | Periodicity | Fits? |
|---|---|---|---|---|
| 601 | General de Ley PM | Most companies | Monthly provisional payment using the **profit coefficient** (art. 14); annual return by **March 31** (art. 76 fr. V) | **Yes**, if the user supplies the prior-year coefficient (or prior-year taxable profit and income). |
| 603 | PM con fines no lucrativos | Non-profits and authorized donees (Título III) | Generally no provisional ISR; ISR on distributable surplus and non-deductible expenses; IVA depends on activities | **No** for ISR; **partial** for IVA. |
| 610 | Residentes en el extranjero | Non-residents | Withholding | **No.** |
| 620 | Cooperativas de producción que difieren | arts. 194–199 | Special deferral | **No.** |
| 622 | AGAPES PM | Agricultural and fishing companies (art. 74) | Monthly (or semiannual via facilidades); exemption of 20 annual UMA per partner, capped at 200 UMA (= $855,892.80 / $8,558,928.00); ISR reduced 30 % up to 423 UMA per partner | **Partial.** Same as 601 plus the exemption and reduction. Defer. |
| 623 | Grupos de sociedades | arts. 59–71 | Group integration | **No.** |
| 624 | Coordinados | Transport coordinators (arts. 72–73) | Pay on behalf of members | **No.** |
| 626 | RESICO PM | PM owned **only by PF**, income ≤ **$35M** (art. 206) | Monthly provisional payment, cash basis (art. 211); annual by March 31 | **Yes.** |

## B.2 ISR algorithms (PM)

### B.2.1 Regime 601: General (arts. 9, 14)

| Step | Rule | Basis |
|---|---|---|
| 1 | **Profit coefficient (CU)** = taxable profit of the last 12-month fiscal year ÷ that year's nominal income. If there is none, use the last year with a coefficient, no more than 5 years back. **First fiscal year: no provisional payments** (no coefficient). | art. 14 fr. I; art. 14 ¶ on filing |
| 2 | Estimated profit = CU × **year-to-date nominal income** (accrual basis: accumulable income excluding the annual inflation adjustment). | art. 14 fr. II |
| 3 | Minus PTU paid in the year, deducted **in equal parts from May to December**, cumulatively, and capped at the profit for the payment. | art. 14 fr. II a) |
| 4 | Minus prior-year tax losses pending application. | art. 14 fr. II c) |
| 5 | × **30 %** | arts. 14 fr. III, 9 |
| 6 | Minus prior provisional payments of the year and interest withholding (art. 54; 0.90 % annual for 2026). | art. 14 fr. III; LIF 2026 art. 24 |
| — | Annual: taxable result × 30 %; **PTU 10 %** of taxable income. Due within 3 months of year-end (March 31). | arts. 9, 76 fr. V |

The rounding of the coefficient is **not stated in art. 14**. Practice uses 4 decimals ⚠️. The app should use the user's figure as given.

### B.2.2 Regime 626: RESICO PM (arts. 206–215)

| Step | Rule | Basis |
|---|---|---|
| 1 | Year-to-date income **actually collected** (cash basis). | arts. 207, 211 |
| 2 | Minus year-to-date authorized deductions **actually paid**: returns and discounts, purchases of goods, expenses, investments (special rates if investment ≤ $3M a year, art. 209), interest, IMSS fees, pension reserves. | arts. 208, 209, 211 |
| 3 | Minus PTU paid and prior-year losses. | art. 211 |
| 4 | × **30 %**, minus prior provisional payments and art. 54 withholding. | arts. 211, 9 |
| — | Annual: art. 9 tax, crediting provisional payments. PTU base = art. 9 taxable profit. | art. 212 |
| Exit | Year-to-date income > **$35M**, or the entity no longer meets the requirements → Título II (regime 601) **from the next fiscal year**. The first year after exit uses the coefficient for the activity under CFF art. 58. File a notice by January 31. | arts. 206 ¶2, 214 |
| Excluded | Partners controlling other companies or related parties, trusts and A en P, Cap. IV/VI/VII/VIII Título II, Título III, entities that previously left RESICO. | art. 206 fr. I–V |

### B.2.3 Withholding personas morales make (affects their PF suppliers)

| Payment to PF | ISR | IVA |
|---|---|---|
| Professional fees (612) | 10 % (106) | 2/3 (RLIVA 3) |
| Rent (606) | 10 % (116) | 2/3 |
| RESICO PF supplier (626) | 1.25 % (113-J) | 2/3 if services or rent |
| Freight transport (ground, of goods; also from a PM) | — | 4 % of the price (RLIVA 3 fr. II) |
| Commissions (PF) | — ⚠️ | 2/3 (LIVA 1-A II d) |
| Desperdicios (any supplier) | — | 100 % (LIVA 1-A II b) |
| Subcontratación / personnel services | — | **— (6 % repealed DOF 23-04-2021)** |
| PM selling through platforms (2026, new) | **2.5 %** withheld by the platform, creditable (20 % without RFC) | — |

## B.3 IVA (PM)

The rules are the same as §A.3 (16 %, 8 % border, 0 %, exempt, creditable IVA, day-17 monthly payment, not year-to-date). In addition, IVA that the PM withholds from others must be **paid in full** to SAT by day 17, with no offsets (LIVA 1-A; RMF 4.1.11). The IVA the PM actually paid to the supplier (IVA − IVA withheld) is creditable in the month of payment (art. 5 fr. III). The **withheld** IVA becomes creditable **in the monthly return following the one in which the withholding was remitted** (art. 5 fr. IV: "podrá ser acreditado en la declaración de pago mensual siguiente a la declaración en la que se haya efectuado el entero de la retención"). Example: withheld in June, remitted with the June return (filed by 17 July) → creditable in the **July** return.

> **Correction (2026-10-15).** This section previously said the withheld IVA "becomes creditable for the PM in the month it is paid". LIVA art. 5 fr. IV, read verbatim in [withholdings-matrix-2026.md](withholdings-matrix-2026.md) §2.2, defers it to the return after the remittance. Spec impact: [MODULE-withholdings §4.2](../project-specs/MODULE-withholdings.md).

## B.4 Deadlines and limits (PM)

| Item | Value | Basis |
|---|---|---|
| Monthly ISR and IVA | Day 17 | LISR 14, 211; LIVA 5-D |
| Annual return | March 31 | LISR 76 fr. V |
| RESICO PM cap | $35,000,000 | LISR 206 |
| PTU payment | Within 60 days after the annual return is due (May 30) ⚠️ | LFT art. 122; not re-fetched |

## B.5 Worked examples (PM)

| ID | Regime / month | Inputs | Computation | Expected |
|---|---|---|---|---|
| PM-601-ISR-01 | 601, **June 2026** | 2025 taxable profit 1,200,000 ÷ income 8,000,000 → CU 0.1500; YTD nominal income 5,000,000; PTU paid in May 120,000 (÷ 8 = 15,000 a month → 2 parts = 30,000); losses 0; prior payments 180,000 | 0.15 × 5,000,000 = 750,000 − 30,000 = 720,000 × 30 % = 216,000 | ISR to pay **36,000.00** |
| PM-601-ISR-02 | 601, first fiscal year | — | No coefficient | **0.00**, flagged "no provisional payments" |
| PM-601-IVA-01 | 601, June | IVA charged and collected 320,000; creditable IVA paid 200,000; IVA withheld from a PF lawyer (fees 50,000 → 8,000 × 2/3) 5,333.33 | 320,000 − 200,000 = 120,000; withholding paid separately | IVA **120,000.00** + withheld IVA **5,333.33** |
| PM-626-ISR-01 | RESICO PM, March | YTD collected 1,500,000; YTD deductions paid 1,100,000; prior payments 70,000 | 400,000 × 30 % = 120,000 | ISR to pay **50,000.00** |

---

# Part C: What the app cannot estimate accurately (disclaimers and risks)

| Risk | Impact | Mitigation |
|---|---|---|
| Deductibility depends on requirements (CFDI, payment method, art. 27/28 limits, non-deductible items such as 53 %/47 % payroll-exempt portions or car caps) | Over-deduction in 612, 606 and PM | Treat entered expenses as "claimed deductible"; show a disclaimer. |
| Depreciation of investments (arts. 31–38) | 612 and PM bases | v1: ask for the depreciation figure, or exclude it. |
| PM coefficient depends on the prior annual return | Wrong input → wrong estimate | Require the coefficient as an explicit input. |
| Annual inflation adjustment, tax losses, PTU caps (LFT 127) | PM / 612 | Optional inputs only. |
| IVA proportional credit for mixed taxable and exempt activity (art. 5 fr. V, 5-B) | Over-credit | Ask whether the activity is 100 % taxable; otherwise flag. |
| Border 8 % eligibility (notice filed, place of delivery, exclusions) | Wrong rate | The user declares eligibility; the default is 16 %. |
| RESICO PF expulsion (3 or more missed payments, cap), RIF reduction year | Wrong regime | Prompt annually. |
| Platform income mixed with own sales; definitive-option eligibility | 625 | Ask for the prior-year total. |
| Salary + 612 or 606 → annual return combines everything | Monthly figures are not the annual result | Monthly-only scope. |
| Tables change each year when inflation passes 10 % (art. 152); LIF rates change every year | Stale numbers | Version the tables by fiscal year; this is a data, not code, change. |
| The day-17 grace by RFC digit and the 2026 employment subsidy decree were not verified | Date and amount errors | Show "day 17" only. |
| Out of scope | 605, 607, 608, 610, 611, 614, 615, 616, 603 (ISR), 620, 623, 624, AGAPES special treatment | Show "not supported". |

**Mandatory UI disclaimer:** "Estimación informativa. No sustituye la declaración ante el SAT ni la asesoría de un contador. Cálculo basado en LISR, LIVA, LIF 2026 y Anexo 8 RMF 2026."

---

# Part D: Sources (accessed 2026-10-13)

| # | Source | URL |
|---|---|---|
| 1 | Anexo 8 RMF 2026 (DOF 28-12-2025): tariffs | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/anexos/Anexo-8-RMF-2026_DOF-28122025.pdf |
| 2 | LISR, current text (last reform DOF 01-04-2024) | https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf |
| 3 | LIVA, current text (last reform DOF 12-11-2021) | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIVA.pdf |
| 4 | CFF, current text (last reform DOF 09-04-2026) | https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf |
| 5 | Ley de Ingresos de la Federación 2026 (DOF 07-11-2025), arts. 24 and 25 fr. VI | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIF_2026.pdf |
| 6 | Reglamento LIVA, art. 3 (2/3 and 4 % withholding) | https://www.diputados.gob.mx/LeyesBiblio/regley/Reg_LIVA_250914.pdf |
| 7 | SAT CFDI 4.0 catalogs (`c_RegimenFiscal`), version 20261001 | http://omawww.sat.gob.mx/tramitesyservicios/Paginas/documentos/catCFDI_V_4_20261001.xls |
| 8 | INEGI, UMA 2026 communiqué | https://www.inegi.org.mx/contenidos/saladeprensa/boletines/2026/uma/uma2026.pdf |
| 9 | DOF 31-12-2025, decree extending the northern and southern border incentives to 31-12-2026 | https://dof.gob.mx/nota_detalle.php?codigo=5777697&fecha=31%2F12%2F2025 · SAT copy: https://www.sat.gob.mx/minisitio/EstimulosFiscalesFronteraNorteSur/documentos/Decreto_modifica_diversos_estimulos_fiscales_region_fronteriza_norte_sur.pdf |
| 10 | Northern border incentive decree (DOF 31-12-2018), arts. 11–13 IVA | https://sidof.segob.gob.mx/notas/docFuente/5547485 |
| 11 | PTU 10 %, Sexta Comisión (DOF 18-09-2020) ⚠️ secondary source | https://idconline.mx/laboral/2020/10/06/mantienen-el-mismo-porcentaje-de-ptu-a-repartir |

**Not retrieved:** the full RMF 2026 text (the SAT host timed out). The rule numbers for RESICO and platform facilities (Title 3.13, 3.11.x) and a secondary-source claim that RMF rule 4.1.4 sets 100 % IVA withholding on rentals were **not confirmed**. That claim contradicts RLIVA art. 3 fr. I c), which sets 2/3 and is the version used here. Verify against the official RMF 2026 before release.

# Part E: Regime coexistence (M2-13 FU1, accessed 2026-10-14)

Which regime pairs one taxpayer may hold at the same time, for the pairs where
MODULE-taxpayers §2.5 and the catalog seed disagreed. Texts read from the primary sources
below (✅). A pair is refused only where the text excludes it; otherwise it stays allowed
with a CPA-review warning.

| Pair | Verdict | Text (LISR unless noted) | Source |
|---|---|---|---|
| PF 621 × 612 | Refused ✅ | Art. 111 ¶1 (in force until 2021): RIF is for PF "que realicen **únicamente** actividades empresariales … por los que no se requiera … título profesional". ¶15: after RIF they "deberán tributar conforme al régimen … Sección I del Capítulo II" (612). Transitorio 2º fr. IX keeps those rules for RIF taxpayers. | E1, E2 |
| PF 621 × 626 | Refused ✅ | Art. 113-E ¶1: RESICO is for PF "que realicen **únicamente** actividades empresariales, profesionales u otorguen el uso o goce temporal de bienes"; ¶6 admits only Capítulos I and VI income besides. Transitorio 2º fr. IX: a RIF taxpayer "podrá optar por continuar" in RIF; otherwise the SAT moves it to Sección IV (RESICO). Either one, not both. | E2, D2 |
| PM 620 × 601, 622, 623, 624 | Refused ✅ | Art. 194 ¶1: production cooperatives of PF "**en lugar de** aplicar lo dispuesto en el Título II … podrán aplicar" this chapter. 601, 622 (Cap. VIII), 623 (Cap. VI) and 624 (Cap. VII) are Título II. | D2 |
| PM 620 × 626 | Refused ✅ | Art. 206 fr. IV: "No tributarán conforme a este Capítulo … Quienes tributen conforme al Capítulo VII del Título VII" (the cooperatives of art. 194). | D2 |
| PM 601 × 622 | Refused ✅ | Art. 74 ¶1: PM dedicated exclusively to agricultural, livestock, forestry or fishing activities "**Deberán** cumplir … conforme al régimen establecido en el presente Capítulo". | D2 |
| PM 622 × 623 | Refused ✅ | Art. 74 ¶2: "Lo dispuesto en este Capítulo no será aplicable a las personas morales que tributen en los términos del Capítulo VI del Título II". | D2 |
| PM 622 × 624 | Refused ✅ | Art. 74 and art. 72 ¶3 each define "exclusively" as at least 90 % of total income from their own activity; both cannot hold at once. | D2 |
| PM 601 × 603 | Allowed, CPA review | Art. 80: a Título III entity that sells goods other than fixed assets or serves non-members pays tax on that profit "en los términos del Título II" above 5 % of income. Not an exclusion. | D2 |
| PM 601 × 623 | Allowed, CPA review | Art. 59 ¶4: group companies "estarán a lo establecido en las demás disposiciones de esta Ley" (Título II) save what Cap. VI says. Not an exclusion. | D2 |
| PM 601 × 624, 603 × 620/622/623/624, 623 × 624 | Allowed, CPA review | No text found that excludes or allows the pair. | D2 |

Observations for the CPA (not changed in FU1): read like the 621 × 626 row, art. 113-E ¶1
and ¶6 would also exclude 626 PF with 607, 608 and 615; and art. 111 (2016 text: "Capítulos
I, III y VI" besides, fr. I socios) would exclude 621 with 607, 608, 611, 615 and 625. Those
pairs stay allowed with a warning until the CPA answers (CPA-REVIEW).

| # | Source | URL |
|---|---|---|
| E1 | LISR, original text (DOF 11-12-2013), art. 111 | https://www.diputados.gob.mx/LeyesBiblio/ref/lisr/LISR_orig_11dic13.pdf |
| E2 | LISR reform decree (DOF 12-11-2021): art. 113-E and Disposiciones Transitorias, art. Segundo fr. IX–XI | https://www.diputados.gob.mx/LeyesBiblio/ref/lisr/LISR_ref07_12nov21.pdf |
| E3 | LISR reform decree (DOF 18-11-2015): art. 111, Capítulos I, III y VI paragraph and fr. I exceptions | https://www.diputados.gob.mx/LeyesBiblio/ref/lisr/LISR_ref01_18nov15.pdf |
