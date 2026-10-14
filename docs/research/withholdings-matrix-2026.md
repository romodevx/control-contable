# Withholdings matrix 2026: ISR and IVA retenciones by emisor × receptor × operation

> **Purpose.** Lets the app derive ISR and IVA withholdings (retenciones) from **who issues** the CFDI, **who receives** it and **what is being sold**, so income and expense estimates match what is actually collected or paid. Informational only; not tax advice.
> **Extends** `docs/research/tax-regimes-mx-2026.md` (hereafter **TR**). Regime catalog, tariffs, RESICO rates and platform rates are **not repeated** here; TR section numbers are cited instead.
> **Cut-off date:** 2026-10-15. **Fiscal year:** 2026.
> **Legend:** ✅ read verbatim from the primary source in this session · ⚠️ inference, practice, or not confirmed (needs CPA review before hard-coding) · ⛔ rule no longer in force.

**Sources read verbatim in this session** (full list in §8): LIVA arts. 1, 1-A, 3, 5 fr. IV, 5-D, 14, 32 fr. V (Diputados, last reform DOF 12-11-2021); RLIVA arts. 3–4 (last reform DOF 25-09-2014); LISR arts. 27 fr. V, 54, 86, 106, 113-E/113-F, 113-J, 116, 134–135, 147 fr. VII, 156, 158, 210 (last reform DOF 01-04-2024); LIF 2026 art. 24; RMF 2026 (DOF 28-12-2025) rules 2.7.1.23, 2.7.3.5, 2.7.3.7, 2.7.5.4, 3.13.13, 3.13.14, 3.13.18, 3.13.19, 3.13.26, 4.1.1–4.1.4, 4.1.11; SAT Anexo 20 Guía de llenado CFDI 4.0; SAT catalogs `catCFDI_V_4_20261001` (c_RegimenFiscal, c_UsoCFDI, c_ObjetoImp, c_Impuesto, c_TasaOCuota).

---

## 0. The six legal facts that drive the whole matrix

| # | Fact | Basis | Status |
|---|---|---|---|
| F1 | **Only personas morales** must withhold ISR from PF on professional fees (10 %), real-estate rent (10 %) and RESICO PF activity (1.25 %). A PF payer (612, 626, 606, 625…) **never** withholds these. | LISR 106 ¶5 ("Cuando los contribuyentes presten servicios profesionales a las personas morales, éstas deberán retener… 10%"); 116 ¶5 ("pagos que efectúen las personas morales… 10%"); 113-J ("a personas morales, estas últimas deberán retener… 1.25%") | ✅ |
| F2 | **Only personas morales** must withhold IVA under LIVA 1-A fr. II: (a) independent **personal** services or use/enjoyment of goods from a **PF**; (b) **desperdicios** bought from anyone; (c) **autotransporte terrestre de bienes** from PF **or PM**; (d) **commissions** from a PF. | LIVA 1-A fr. II a)–d) | ✅ |
| F3 | The IVA withholding is reduced by regulation to **2/3 of the IVA** for PF personal services, commissions and use/enjoyment of goods, and to **4 % of the price** for ground freight. Desperdicios have **no reduction → 100 %**. | RLIVA 3 fr. I a)–c), fr. II; LIVA 1-A last ¶ | ✅ |
| F4 | "Personal" services = services that **are not actividad empresarial**. A PF in actividad empresarial selling goods or providing non-professional services (repairs, maintenance, catering…) is **not** subject to 2/3 IVA withholding nor to the 10 % ISR. | LIVA 14 last ¶ ("…tiene la característica de personal, cuando se trate de las actividades… que no tengan la naturaleza de actividad empresarial"); LISR 106 ¶5 limited to "servicios profesionales" | ✅ (classification of a given service ⚠️) |
| F5 | The **1.25 % RESICO** withholding covers **all** RESICO PF activity billed to a PM — **including sales of goods** — not only services. | LISR 113-J ("realicen actividades empresariales, profesionales u otorguen el uso o goce temporal de bienes, a personas morales") | ✅ |
| F6 | The old **6 % IVA withholding on personnel services (subcontratación)**, LIVA 1-A fr. IV, was **repealed** ("Derogada DOF 23-04-2021"). No IVA withholding applies to specialized services/works in 2026 unless another case of 1-A applies. | LIVA 1-A fr. IV (text: "(Se deroga)… Derogada DOF 23-04-2021") | ✅ ⛔ |

Consequences: **PF → PF: no withholding** (except the foreign-resident case of 1-A fr. III, which concerns the payer buying from a non-resident). **PM → PM: no withholding** except **ground freight (4 % IVA)** and **desperdicios (100 % IVA)**. **Público en general: never**.

---

## 1. Matrix

### 1.1 Receptor (payer) columns

| Code | Receptor | Withholds? |
|---|---|---|
| **R-PM** | Any persona moral: **601, 626 RESICO PM, 603, 620, 622, 623, 624** (and 610 PM with EP) | **Yes**: LISR 106/116/113-J and LIVA 1-A speak of "personas morales" without regime distinction. RESICO PM (Título VII Cap. XII) and 603 (Título III; LISR 86 fr. V "retener y enterar el impuesto a cargo de terceros") **withhold exactly like a 601**. ✅ |
| **R-FED** | Federación and its **organismos descentralizados** | IVA: yes, broader scope (§1.4). ISR: ⚠️ yes in practice (LISR 86 ¶ "sólo tendrán las obligaciones de retener y enterar el impuesto…"). |
| **R-EST** | Estados, municipios, CDMX and their descentralizados | IVA: **no** ("no efectuarán la retención", LIVA 3 ¶3) ✅. ISR: ⚠️ yes in practice (LISR 86). |
| **R-PFA** | PF **with** activity (612, 626 PF, 606, 621, 625) | **No** ISR/IVA withholding on domestic purchases (F1, F2). Only LIVA 1-A fr. III (buying tangible goods / renting goods from a non-resident without EP). ✅ |
| **R-PG** | PF **without** activity / **público en general** (RFC `XAXX010101000`, regime 616, UsoCFDI S01) | **Never** withholds. ✅ |
| **R-EXT** | Foreign resident (RFC `XEXX010101000`, regime 616 in the CFDI) | Does not withhold Mexican tax. Exports by the emisor may be IVA 0 % (LIVA 29) ⚠️ out of MVP scope. |

### 1.2 Emisor (seller / provider) rows

PF: **612-P** (actividad profesional / honorarios), **612-E** (actividad empresarial), **606** (arrendamiento inmuebles), **626-PF** (RESICO PF), **625** (plataformas), **621** (RIF legacy). PM: **601, 626-PM, 603, 620, 622, 623, 624** behave identically **as emisores** (no ISR withholding on PM income except platforms and interest; IVA only freight/desperdicios) → one row **PM-any**.

Cell format: `ISR % on subtotal / IVA withheld` — who withholds is always the **receptor** (payer), at the moment it **pays** (LIVA 1-A ¶ "en el momento en el que pague el precio… sobre el monto de lo efectivamente pagado"; LISR "sobre el monto de los pagos que les efectúen"). `—` = no withholding.

### 1.3 Matrix by operation (receptor = R-PM; other columns in §1.4)

Base for ISR = subtotal **without IVA**. IVA withheld is expressed as fraction of IVA and as % of subtotal at 16 %.

| Emisor ↓ / Operation → | Servicios profesionales (honorarios) | Servicios empresariales (no profesionales) | Arrendamiento inmueble | Uso/goce bienes muebles | Enajenación de bienes | Autotransporte terrestre de bienes | Comisiones | Desperdicios / reciclables |
|---|---|---|---|---|---|---|---|---|
| **612-P** | **10 %** (106 ¶5) / **2/3 = 10.6667 %** (1-A II a; RLIVA 3 I a) ✅ | n/a | (goes to 606) | ISR — ⚠️ / IVA 2/3 (1-A II a; RLIVA 3 I c) | — / — | — / **4 %** | ⚠️ ISR — / IVA **2/3** (1-A II d; RLIVA 3 I b) | — / **100 %** (1-A II b) |
| **612-E** | n/a | — / — (F4) ✅ | (606) | — / **2/3** (RLIVA 3 I c) ✅ | — / — ✅ | — / **4 %** (1-A II c; RLIVA 3 II) ✅ | — ⚠️ / **2/3** | — / **100 %** ✅ |
| **606** | n/a | n/a | **10 %** (116 ¶5) / **2/3** (RLIVA 3 I c); residential rent is IVA-exempt (LIVA 20 fr. II) → IVA — ✅ | n/a | n/a | n/a | n/a | n/a |
| **626-PF** | **1.25 %** (113-J) / **2/3** ✅ | **1.25 %** / — (F4) ✅ | **1.25 %** / **2/3** (residential: —) ✅ ⚠️ see A3 | **1.25 %** / **2/3** | **1.25 %** (F5) / — ✅ | **1.25 %** / **4 %** ✅ | **1.25 %** / **2/3** | **1.25 %** / **100 %** ✅ |
| **625** (sale through a platform) | Platform withholds, not the PM: ISR 2.5 % goods & services / 2.1 % transport & delivery / 4 % lodging; IVA 50 %; no RFC: ISR 20 %, IVA 100 % (TR §A.2.6; LISR 113-A, 113-C fr. IV; LIF 2026 art. 25 fr. VI; LIVA 18-J fr. II a) ✅. If the 625 PF bills a PM **directly**, outside the platform → ⚠️ A6. | ← | ← | ← | ← | ← | ← | ← |
| **621** RIF | n/a (RIF excludes professional activity, TR §E) | — / — | n/a | — / 2/3 ⚠️ | — / — | — / 4 % | — / 2/3 | — / 100 % |
| **PM-any** | — / — | — / — | — / — | — / — | — / — | — / **4 %** (1-A II c "prestados por personas físicas o morales") ✅ | — / — (1-A II d only PF) ✅ | — / **100 %** (1-A II b, no provider restriction) ✅ |

Other operation kinds (any emisor):

| Operation | ISR | IVA | Basis | Status |
|---|---|---|---|---|
| Servicios de personal / subcontratación (old 6 %) | — | — (repealed) | LIVA 1-A fr. IV derogada DOF 23-04-2021 | ✅ ⛔ |
| Fletes = autotransporte terrestre de **bienes** | as row above | **4 %** of the price (not of the IVA) | RLIVA 3 fr. II | ✅ |
| Passenger transport (not "bienes") | — | — (not in 1-A) | LIVA 1-A fr. II c) says "de bienes" | ✅ |
| Plataformas digitales (625 / PM selling via app) | 2.5 % / 2.1 % / 4 % (20 % w/o RFC), withheld by the **platform** | 50 % (100 % w/o RFC) PF only | TR §A.2.6, §B.2.3 | ✅ |
| Intereses paid by the **financial system** | **0.90 % annual on capital** (provisional) | Interest IVA rules out of scope | LISR 54, 135 ¶1; LIF 2026 art. 24 | ✅ |
| Intereses from non-financial payers (art. 134 ¶2 títulos) | 20 % of nominal interest | — | LISR 135 ¶1 | ✅ (MVP: out of scope) |
| Donativos | — | — | No withholding provision found | ⚠️ |
| Uso/goce of land for oil/power infrastructure (servidumbres) paid by PM to PF | **20 % definitive** | **100 %** | RMF 2.7.3.7 ¶2 | ✅ (MVP: out of scope) |
| Desperdicios — buyer exclusively in scrap trading | — | 100 % withheld; immediate refund facility | LIVA 1-A II b; RMF 4.1.1, 4.1.2 (definition) | ✅ |
| RESICO PF exclusively AGAPES, exempt income | **no 1.25 %** if CFDI "Descripcion" states the 113-E ¶9 exemption | IVA per general rules | RMF 3.13.26 | ✅ |
| Payment to a **foreign resident** for services rendered **in Mexico** | **25 %** of gross (treaty may reduce) | No 1-A withholding; recipient self-assesses IVA on imported services | LISR 156; LIVA 24 fr. V | ✅ / ⚠️ treaty & IVA timing |
| Rent of real estate in Mexico paid to a foreign resident | **25 %** of gross | — | LISR 158 | ✅ |
| Purchase or rent of **tangible goods** from a foreign resident without EP (any PF or PM payer) | — | **100 %** | LIVA 1-A fr. III | ✅ |

### 1.4 Receptor columns other than R-PM

| Emisor / operation | R-FED (Federación + descentralizados) | R-EST (estados, municipios) | R-PFA | R-PG | R-EXT |
|---|---|---|---|---|---|
| 612-P honorarios | ISR 10 % ⚠️ / IVA withheld on **any amount** (RMF 4.1.3 ¶2); fraction: ⚠️ **A2** (literal: 100 %; practice 2/3) | ISR 10 % ⚠️ / IVA — | — / — | — / — | — / — |
| 606 rent | ISR 10 % ⚠️ / IVA ⚠️ A2 if price > $2,000 | ISR 10 % ⚠️ / — | — / — | — / — | — |
| 626-PF any | ISR 1.25 % ⚠️ / IVA per A2 | 1.25 % ⚠️ / — | — / — | — / — | — |
| PF goods (612-E, 626, 621) | ISR — (626: 1.25 % ⚠️) / IVA **withheld** if price > $2,000 (LIVA 3 ¶3 "adquieran bienes… de personas físicas"; RMF 4.1.3) ⚠️ fraction | — / — | — / — | — / — | — |
| Autotransporte (PF or PM) | IVA **4 %** (LIVA 3 ¶3; RLIVA 4) ✅ | — | — | — | — |
| PM-any other | — / — | — / — | — / — | — / — | — |

---

## 2. Effects on the monthly estimate

### 2.1 When the app user is the **emisor** (income side)

| Item | Rule | Basis | Status |
|---|---|---|---|
| Cash actually received | `subtotal + IVA trasladado − ISR retenido − IVA retenido` | definitions above | ✅ |
| 612 provisional ISR | YTD computation; then **subtract YTD 10 % ISR withheld** — "será acreditable contra el impuesto a pagar que resulte en los pagos provisionales" | LISR 106 ¶5 (TR §A.2.3 step 7) | ✅ |
| 606 ISR | Monthly computation minus the 10 % withheld ("podrá acreditarse") | LISR 116 ¶5 (TR §A.2.4 step 4) | ✅ |
| 626 RESICO PF ISR | Monthly ISR = rate × income collected; the 1.25 % "será considerado en el pago mensual" → **subtract**. The monthly payment is **not** final: the annual return (113-F) subtracts monthly payments **and** 113-J withholdings again. A negative monthly result is **not refunded monthly**; treat as 0 for the month and carry to the annual ⚠️ (A4). | LISR 113-J ¶2; 113-F ("podrán disminuir… el que les retuvieron conforme al artículo 113-J") | ✅ / ⚠️ negative month |
| 625 platform ISR | Withholding is provisional, or **definitive** if the 113-B option applies (TR §A.2.6) | LISR 113-A, 113-B | ✅ |
| IVA payable | `IVA a cargo = IVA trasladado cobrado − IVA acreditable pagado − IVA retenido al contribuyente en el mes`, monthly, not YTD | LIVA 1 last ¶, 5-D ¶3 ("disminuirá… el impuesto que se le hubiere retenido en dicho mes") | ✅ |
| Which month | IVA counts when **actually collected** (LIVA 1-B); withholdings occur when the payer **pays** (LIVA 1-A; LISR 106/116/113-J). With PPD invoices, use the complemento de pago date. | LIVA 1-A, 1-B | ✅ |
| Proof of withholding | The PF's own CFDI showing the retention may serve as the withholding certificate (no separate CFDI de retenciones needed) | LISR 106 ¶6, 116 ¶6; LIVA 32 fr. V ¶3; RMF 3.13.13 (RESICO), 2.7.5.4 ¶2 | ✅ |

### 2.2 When the app user is the **receptor** (expense side; mainly PM users, incl. RESICO PM)

| Item | Rule | Basis | Status |
|---|---|---|---|
| Cash paid to supplier | `subtotal + IVA − ISR retenido − IVA retenido` | | ✅ |
| Remit (enterar) ISR withheld | By **day 17** of the following month. RESICO withholding: declaration "R14 Retenciones por servicios profesionales/Régimen Simplificado de Confianza". | LISR 106 ¶5, 116 ¶5 ("conjuntamente con las señaladas en el artículo 96" → day 17), 113-J ¶1; RMF 3.13.14 | ✅ |
| Remit IVA withheld | With the monthly IVA payment or by **day 17** of the following month, via "IVA retenciones"; **no** credit, offset or reduction may be applied against it. | LIVA 1-A ¶ "El retenedor…"; RMF 4.1.11 | ✅ |
| The payer replaces the supplier | The withholder "sustituirá al enajenante, prestador… en la obligación de pago y entero" | LIVA 1-A | ✅ |
| Deductibility | Expense is deductible **only if** withholding and remittance obligations are met: PM 601 (27 fr. V), RESICO PM (210 → "se estará a lo dispuesto en las fracciones aplicables del artículo 27"), PF 612 (147 fr. VII) | LISR 27 fr. V, 210, 147 fr. VII | ✅ |
| IVA acreditable — paid portion | The IVA the payer **actually paid** to the supplier (IVA − IVA retenido) is creditable in the month of payment | LIVA 5 fr. III (paid) | ✅ |
| IVA acreditable — withheld portion | The withheld IVA becomes creditable **in the monthly return following the one in which it was remitted** ("podrá ser acreditado en la declaración de pago mensual siguiente a la declaración en la que se haya efectuado el entero de la retención"). Example: withheld in **June**, remitted with the June return (filed by 17 July) → creditable in the **July** return. | LIVA 5 fr. IV ✅ | ✅ **Corrects TR §B.3**, which says "creditable in the month it is paid" |
| Quarterly 606 PF lessors who withhold under 1-A fr. III | Remit and credit quarterly | RMF 4.1.4 | ✅ (edge) |
| Monthly informative | IVA withholdings info by day 17 (LIVA 32 fr. V); RESICO PF/PM relieved of DIOT | RMF 3.13.17 | ✅ |

---

## 3. CFDI 4.0 fields that identify each case

| Field | What it tells the app | Rules (Anexo 20 guía; catalog 20261001) | Status |
|---|---|---|---|
| `Emisor/RegimenFiscal` | Emisor regime → row of the matrix (612 alone does **not** distinguish P vs E: ask the operation kind) | c_RegimenFiscal: 601, 603, 620, 622, 623, 624 = PM only; 605–616, 621, 625 = PF only; 610, 626 = both | ✅ |
| RFC length | 12 chars = PM, 13 = PF | CFF/RFC format (well-known; not re-read) | ⚠️ |
| `Receptor/Rfc` | `XAXX010101000` = público en general (no withholding); `XEXX010101000` = foreign resident | RMF 2.7.1.23 | ✅ |
| `Receptor/RegimenFiscalReceptor` | Receptor regime → column; for `XEXX010101000` must be **616** | Anexo 20 guía, campo RegimenFiscalReceptor | ✅ |
| `Receptor/UsoCFDI` | Must be compatible with RegimenFiscalReceptor. **S01** = no fiscal effects (público en general); **G01** mercancías, **G03** gastos en general; D01–D10 are PF personal deductions (receptor PF without withholding duty). | c_UsoCFDI | ✅ |
| `Concepto/ObjetoImp` | 01 no objeto; **02 objeto, taxes broken down per concept (required to show retenciones)**; 03 objeto sin desglose; 04 objeto, no causa; 05 IVA crédito PODEBI; 06 objeto del IVA, no traslado; 07/08 no traslado IVA | c_ObjetoImp; guía: if 02 "se deben desglosar los Impuestos a nivel de Concepto" | ✅ |
| `Concepto/Impuestos/Retenciones/Retencion` | `Impuesto` **001 ISR**, **002 IVA**; `TipoFactor` Tasa; `TasaOCuota` — ISR range **0.000000–0.350000**, IVA retention range **0.000000–0.160000** (so 0.100000, 0.012500, 0.106667, 0.040000, 0.160000) | c_Impuesto; c_TasaOCuota (Rango ISR 0–0.35, Retención Sí; Rango IVA 0–0.16, Retención Sí) | ✅ (exact 6th decimal of 2/3 ⚠️ A5) |
| `Concepto/ClaveProdServ` | Can hint operation kind (e.g. freight, rent) but is not authoritative | Anexo 20 Apéndice 3 | ⚠️ |
| `MetodoPago` PUE / PPD + complemento de pagos | Month in which collection, payment and withholding happen | LIVA 1-A, 1-B | ✅ |
| CFDI de retenciones e información de pagos (v2.0) | Separate document; optional when the income CFDI already shows the withholdings | RMF 2.7.5.4; LISR 106 ¶6 | ✅ |

**Minimal questions the app must ask** when no XML is imported (each maps to one input of §4):
1. My person type and regime (already in the taxpayer profile).
2. Counterparty: persona moral / persona física con actividad / persona física sin actividad o público en general / gobierno federal / gobierno estatal o municipal / extranjero. (+ counterparty regime only if PF: 626 vs other.)
3. Operation kind (picker with the §4 enum; default inferred from my regime: 612-P→honorarios, 606→arrendamiento, 612-E→enajenación).
4. Subtotal and IVA rate (16 / 8 / 0 / exento). For rent: residential or not (residential = IVA exempt).
5. Only for 625: through the platform? RFC given to platform?

---

## 4. Decision table (implementable)

### 4.1 Inputs

```
EmisorPersona   ∈ {PF, PM}
EmisorRegime    ∈ {601,603,606,612,620,621,622,623,624,625,626}
ReceptorKind    ∈ {PM_PRIVATE, GOV_FEDERAL, GOV_STATE_MUNICIPAL, PF_WITH_ACTIVITY, PF_NO_ACTIVITY_OR_PUBLIC, FOREIGN}
                  // derive: Rfc==XAXX010101000 → PF_NO_ACTIVITY_OR_PUBLIC; XEXX010101000 → FOREIGN;
                  //         RFC len 12 → PM_PRIVATE unless user flags government (⚠️ no CFDI field identifies government)
OperationKind   ∈ {PROFESSIONAL_SERVICES, BUSINESS_SERVICES, RENT_REAL_ESTATE, RENT_MOVABLE, SALE_OF_GOODS,
                   GROUND_FREIGHT, COMMISSION, SCRAP, PLATFORM, OTHER}
Subtotal        decimal, excluding IVA
IVARate         ∈ {0.16, 0.08, 0, EXEMPT}
ResidentialRent bool (only RENT_REAL_ESTATE)
```

### 4.2 Rules (evaluate ISR and IVA independently; first match wins within each)

**ISR withholding rate**

| ID | Condition | ISR rate | Withholder | Basis |
|---|---|---|---|---|
| ISR-0 | ReceptorKind ∈ {PF_WITH_ACTIVITY, PF_NO_ACTIVITY_OR_PUBLIC} | 0 | — | F1 ✅ |
| ISR-1 | ReceptorKind = FOREIGN | 0 | — | foreign payer has no Mexican duty ✅ |
| ISR-2 | EmisorPersona = PM | 0 (PLATFORM: 2.5 % by platform, TR §B.2.3) | — | No LISR provision for PM-to-PM withholding on these operations ✅ |
| ISR-3 | OperationKind = PLATFORM | per TR §A.2.6 (2.5 / 2.1 / 4 / 20 %) | platform | LISR 113-A, 113-C; LIF 2026 art. 25 fr. VI ✅ |
| ISR-4 | EmisorRegime = 626 (PF) and ReceptorKind ∈ {PM_PRIVATE, GOV_*} and operation ≠ PLATFORM | **1.25 %** (0 if AGAPES-exempt with legend) | receptor | LISR 113-J; RMF 3.13.26 ✅ (GOV ⚠️) |
| ISR-5 | EmisorRegime = 612 and OperationKind = PROFESSIONAL_SERVICES and ReceptorKind ∈ {PM_PRIVATE, GOV_*} | **10 %** | receptor | LISR 106 ¶5 ✅ (GOV ⚠️) |
| ISR-6 | EmisorRegime = 606 (or 612 billing real-estate rent) and OperationKind = RENT_REAL_ESTATE and ReceptorKind ∈ {PM_PRIVATE, GOV_*} | **10 %** | receptor | LISR 116 ¶5 ✅ (GOV ⚠️) |
| ISR-7 | otherwise | 0 | — | 612-E goods/services, commissions, freight: no ISR provision ✅ (commission ⚠️ A7) |

**IVA withholding** (`IVA = Subtotal × IVARate`; if IVARate ∈ {0, EXEMPT} or ResidentialRent → **0**, there is no IVA to withhold)

| ID | Condition | IVA withheld | Withholder | Basis |
|---|---|---|---|---|
| IVA-0 | ReceptorKind ∈ {PF_NO_ACTIVITY_OR_PUBLIC, FOREIGN} | 0 | — | ✅ |
| IVA-1 | ReceptorKind = PF_WITH_ACTIVITY | 0 (domestic supplier) | — | 1-A fr. II only PM; fr. IV repealed ✅ |
| IVA-2 | ReceptorKind = GOV_STATE_MUNICIPAL | 0 | — | LIVA 3 ¶3 last sentence ✅ |
| IVA-3 | OperationKind = PLATFORM and EmisorPersona = PF | 50 % of IVA (100 % w/o RFC) | platform | LIVA 18-J II a ✅ |
| IVA-4 | OperationKind = GROUND_FREIGHT and ReceptorKind ∈ {PM_PRIVATE, GOV_FEDERAL} | **4 % × Subtotal** | receptor | LIVA 1-A II c; RLIVA 3 II, 4 ✅ |
| IVA-5 | OperationKind = SCRAP and ReceptorKind = PM_PRIVATE | **100 % of IVA** | receptor | LIVA 1-A II b ✅ |
| IVA-6 | EmisorPersona = PF and OperationKind ∈ {PROFESSIONAL_SERVICES, RENT_REAL_ESTATE, RENT_MOVABLE, COMMISSION} and ReceptorKind = PM_PRIVATE | **2/3 × IVA** (= 10.6667 % of subtotal at 16 %; 5.3333 % at 8 % ⚠️ A8) | receptor | LIVA 1-A II a, d; RLIVA 3 I ✅ |
| IVA-7 | EmisorPersona = PF and ReceptorKind = GOV_FEDERAL and (OperationKind = PROFESSIONAL_SERVICES or Subtotal > 2,000) | ⚠️ **A2**: 2/3 × IVA (MVP default) | receptor | LIVA 3 ¶3; RMF 4.1.3 |
| IVA-8 | otherwise (PF goods / business services to PM; any PM supplier) | 0 | — | F4 ✅ |

**Outputs**: `ISRRate, ISRWithheld = round2(Subtotal × ISRRate)`, `IVAWithheld = round2(...)`, `Withholder ∈ {RECEPTOR, PLATFORM, NONE}`, `Net = Subtotal + IVA − ISRWithheld − IVAWithheld`, `RuleIDs[]`, `Warnings[]` (any ⚠️ rule adds a CPA-review warning).

```go
// Sketch only; the real implementation follows TDD with §5 as table-driven goldens.
func Withholdings(in Input) Output {
    isr := isrRule(in)   // ISR-0..ISR-7
    iva := ivaRule(in)   // IVA-0..IVA-8
    return Output{ISR: isr, IVA: iva, Net: in.Subtotal.Add(in.IVA()).Sub(isr.Amount).Sub(iva.Amount)}
}
```

Rounding: round each amount to cents per CFDI concept, half-up ⚠️ (A5). Never derive 2/3 as `0.106667 × Subtotal` and expect equality with `IVA × 2 / 3` beyond a cent.

---

## 5. Worked examples (goldens)

All MXN, IVA 16 % unless stated.

| ID | Case | Subtotal | IVA | ISR ret. | IVA ret. | Net cash | Rules |
|---|---|---|---|---|---|---|---|
| WH-01 | **Owner's case**: RESICO PF (626) → PM (García Lorca), services | 100,000.00 | 16,000.00 | **1,250.00** (1.25 %) | **10,666.67** (2/3) | **104,083.33** | ISR-4, IVA-6 |
| WH-02 | 612 honorarios → PM 601 | 50,000.00 | 8,000.00 | 5,000.00 | 5,333.33 | 47,666.67 | ISR-5, IVA-6 |
| WH-03 | 606 commercial rent → PM | 30,000.00 | 4,800.00 | 3,000.00 | 3,200.00 | 28,600.00 | ISR-6, IVA-6 |
| WH-03b | 606 **residential** rent → PM | 30,000.00 | 0 (exempt) | 3,000.00 | 0 | 27,000.00 | ISR-6, IVA exempt |
| WH-04 | 612 lawyer → PF 612 doctor (PF → PF) | 20,000.00 | 3,200.00 | 0 | 0 | 23,200.00 | ISR-0, IVA-1 |
| WH-05 | 612-E sells goods → PM 601 | 50,000.00 | 8,000.00 | 0 | 0 | 58,000.00 | ISR-7, IVA-8 |
| WH-05b | **RESICO PF** sells goods → PM | 50,000.00 | 8,000.00 | 625.00 | 0 | 57,375.00 | ISR-4, IVA-8 |
| WH-06 | PM carrier → PM, ground freight | 10,000.00 | 1,600.00 | 0 | 400.00 (4 %) | 11,200.00 | ISR-2, IVA-4 |
| WH-06b | PF 612-E carrier → PM, freight | 10,000.00 | 1,600.00 | 0 | 400.00 | 11,200.00 | ISR-7, IVA-4 |
| WH-07 | PM → PM specialized services (old 6 % case) | 100,000.00 | 16,000.00 | 0 | **0** (fr. IV repealed 2021) | 116,000.00 | ISR-2, IVA-8 |
| WH-08 | PF sells scrap → PM | 20,000.00 | 3,200.00 | 0 (612-E) | 3,200.00 (100 %) | 20,000.00 | IVA-5 |
| WH-09 | 625 PF sells goods via platform, RFC given | 10,000.00 | 1,600.00 | 250.00 (2.5 %, platform) | 800.00 (50 %) | 10,550.00 | ISR-3, IVA-3 |
| WH-10 | 612 honorarios → público en general (XAXX010101000) | 5,000.00 | 800.00 | 0 | 0 | 5,800.00 | ISR-0, IVA-0 |

**Emisor-side estimate (WH-01 owner, month with only that invoice, no expenses):** RESICO ISR = 100,000 × **2.00 %** (≤ 208,333.33, TR §A.2.5) = 2,000.00 − 1,250.00 withheld = **750.00 to pay**. IVA = 16,000.00 − 0 acreditable − 10,666.67 withheld = **5,333.33 to pay**.

**Receptor-side (WH-11): PM 601 pays a PF 612 consultant** subtotal 40,000.00 in **June**: pays supplier 40,000 + 6,400 − 4,000 − 4,266.67 = **38,133.33**. Remits by **17 July**: ISR 4,000.00 + IVA 4,266.67 (no offsets). IVA acreditable: **2,133.33 in June** (paid portion) + **4,266.67 in the July return** (month after the return in which it was remitted, LIVA 5 fr. IV). Deduction of 40,000 valid only if the withholdings are remitted (LISR 27 fr. V).

**WH-12: 612 honorarios to the Federación** subtotal 30,000.00: ISR 3,000.00 ⚠️; IVA 4,800 → withheld 3,200.00 under the MVP default (2/3) ⚠️ A2 (literal reading would be 4,800.00).

---

## 6. Ambiguities ⚠️ for CPA confirmation

| ID | Question | Why ambiguous | MVP default |
|---|---|---|---|
| A1 | Does the government (Federación, estados, municipios) withhold ISR 10 % / 1.25 %? | LISR 106/116/113-J say "personas morales"; art. 86 ¶ says government only has the obligation "de retener y enterar el impuesto… cuando… estén obligados a ello". | Yes, same as PM, with warning. |
| A2 | IVA fraction withheld by the **Federación** from PF (services, rent, goods) | LIVA 3 ¶3 orders retention "en los términos del artículo 1o.-A" (100 %); the 2/3 reduction of RLIVA 3 addresses "personas morales"; RLIVA 4 reduces only freight for the Federación. | 2/3 for services/rent; goods: 100 % only above $2,000 (RMF 4.1.3) — both flagged. |
| A3 | RESICO PF real-estate rent to PM: 1.25 % (113-J) instead of 10 % (116)? | 113-J explicitly covers "otorguen el uso o goce temporal de bienes"; 116 is for Capítulo III taxpayers. | 1.25 %. |
| A4 | RESICO PF month where 1.25 % withheld > monthly ISR | 113-J/113-E don't say whether the excess carries to later months or only to the annual return (113-F). | Monthly payable floors at 0; excess shown as "a favor (anual)". |
| A5 | Exact 2/3 rate in CFDI (`0.106666` vs `0.106667`) and rounding mode | Catalog is a range; guía gives no 2/3 example. | Compute `IVA × 2 / 3`, round half-up to cents; accept either rate when importing XML. |
| A6 | 625 PF billing a PM **directly** (not through the platform) | 113-A last ¶ covers direct collections by self-payment; whether the PM must also withhold under 106/113-J is unclear. | No PM withholding, warning. |
| A7 | PF commission agent (comisionista): any ISR withholding? | IVA 2/3 is explicit (1-A II d); ISR 106 ¶5 covers "servicios profesionales" only. | ISR 0, warning. |
| A8 | IVA 2/3 under the **8 % border** rate | RLIVA 3 says 2/3 "del impuesto que se les traslade" → 5.3333 %; decrees don't address it. | 2/3 of the IVA actually charged. |
| A9 | Classification "profesional" vs "empresarial" for a PF service | Depends on whether the activity requires a professional title / LIVA 14 last ¶; not visible in CFDI. | Ask the user; default from regime and ClaveProdServ. |
| A10 | Rent of **movable** goods by a PF 612: ISR withholding? | 116 covers inmuebles only; muebles rent is actividad empresarial. | ISR 0, IVA 2/3. |
| A11 | Donativos | No withholding provision found; Anexo 7 not retrieved. | 0. |
| A12 | SAT Anexo 7 RMF 2026 criterios normativos on withholding | Download from sat.gob.mx timed out this session. | Re-check before release. |
| A13 | Foreign-resident payments (treaties, IVA on imported services timing) | Out of MVP scope. | Show "not supported". |

**Correction to TR to apply in a follow-up (not edited here):** TR §B.3 states withheld IVA "becomes creditable for the PM in the month it is paid"; LIVA 5 fr. IV says it is creditable **in the monthly return following the one in which the withholding was remitted**. TR §A.3 "Withholding by a PM" should also add commissions, desperdicios (100 %) and note that the 6 % case was repealed.

---

## 7. What the engineers should capture (summary)

Per invoice: emisor person type + regime, receptor kind (PM / gov federal / gov state-municipal / PF with activity / PF without activity-público en general / foreign), receptor regime when PF (626 vs other), operation kind, subtotal, IVA rate, residential flag (rent), platform flag (625). Outputs per §4. Withholdings feed: emisor ISR credit (612/606/626), emisor IVA `− retenido`, receptor "retenciones por enterar" due day 17, receptor IVA acreditable split (paid now / withheld next month after entero).

---

## 8. Sources (accessed 2026-10-14/15)

| # | Source | URL |
|---|---|---|
| W1 | LIVA, current text (last reform DOF 12-11-2021): arts. 1, 1-A, 1-B, 3, 5, 5-D, 14, 18-J, 20, 24, 32 | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIVA.pdf |
| W2 | Reglamento LIVA (last reform DOF 25-09-2014): arts. 3, 4 | https://www.diputados.gob.mx/LeyesBiblio/regley/Reg_LIVA_250914.pdf |
| W3 | LISR, current text (last reform DOF 01-04-2024): arts. 27, 54, 86, 106, 113-A–113-J, 116, 134, 135, 147, 156, 158, 210 | https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf |
| W4 | Ley de Ingresos de la Federación 2026 (DOF 07-11-2025): arts. 24, 25 fr. VI | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIF_2026.pdf |
| W5 | Resolución Miscelánea Fiscal 2026 (DOF 28-12-2025): rules 2.7.1.23, 2.7.3.5, 2.7.3.7, 2.7.5.4, 3.13.13, 3.13.14, 3.13.17–3.13.19, 3.13.26, 4.1.1–4.1.4, 4.1.11 | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/rmf/RMF_2026-DOF-28122025.pdf ⚠️ (text read from a copy downloaded in an earlier session; URL path not re-verified) |
| W6 | SAT, Anexo 20 Guía de llenado de los CFDI (4.0) | http://omawww.sat.gob.mx/tramitesyservicios/Paginas/documentos/Anexo_20_Guia_de_llenado_CFDI.pdf |
| W7 | SAT CFDI 4.0 catalogs, version 20261001 (c_RegimenFiscal, c_UsoCFDI, c_ObjetoImp, c_Impuesto, c_TasaOCuota) | http://omawww.sat.gob.mx/tramitesyservicios/Paginas/documentos/catCFDI_V_4_20261001.xls |
| W8 | SAT, Formato de factura (Anexo 20) page | https://wwwmat.sat.gob.mx/consultas/35025/formato-de-factura-electronica-(anexo-20) |
| — | Not retrieved: Anexo 7 RMF 2026 (criterios normativos) — sat.gob.mx timed out | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/anexos/Anexo-7-RMF-2026_DOF-28122025.pdf ⚠️ |
