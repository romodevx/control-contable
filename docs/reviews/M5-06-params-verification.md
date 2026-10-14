# M5-06: independent verification of the 2026 tax parameters

- **Reviewed:** PR #73, `backend/internal/taxengine/params/data/2026/v1.yaml` (head `486dc3d`)
- **Reviewer:** QA (qa-sr), 2026-10-14
- **Result:** 0 mismatches across 822 compared fields. Nothing in the file needed a fix.

## Method (double entry)

1. I did not read the author's extraction or the values in `v1.yaml` before building my own copy. I read only the key names and the row layout, so I would know what to compare.
2. I downloaded every official source again myself (list below).
3. **Tariffs.** I used a different rendering of Anexo 8 from the one the author used. The author parsed the SAT PDF copy. I parsed the **DOF HTML edition** (nota 5777219, 28/12/2025, "ANEXOS 4, 5, 6, 8, 15 y 25 de la RMF 2026") with my own parser, so a PDF-extraction error could not hide in both copies. Amounts were kept as printed, with thousands separators removed, and percentages were kept as printed.
4. **Scalars and RESICO tables.** I typed these by hand from LISR, LIF 2026, LIVA, RLIVA, the border decrees and the INEGI UMA notices in the DOF.
5. I wrote the second copy to a scratch YAML outside the repo. A script then compared it with `v1.yaml` field by field: every amount and rate was compared as a decimal, with each percentage divided by 100, and every `valid_to` was compared too. For the RESICO tables, which print only "Hasta", I derived each lower bound as the previous upper bound + 0.01.
6. **Harness self-test.** I changed two values in my copy (Dec row 2 fee 194.59→194.64, UMA daily 117.31→117.30) and the diff reported exactly 3 mismatches: the fee appears in both art. 106 December and art. 152. So a clean result is not a vacuous pass.

## Counts compared

| Group | Tables / values | Rows | Fields |
|---|---|---|---|
| Anexo 8 B.V (art. 96), B.VI × 12 (art. 106), art. 116 monthly, B.VII (RIF), C.II (art. 152) | 16 tables | 176 | lower, upper, fixed fee, rate |
| LISR 113-E / 113-F (RESICO PF) | 2 tables | 10 | lower (derived), upper, rate |
| Rates, thresholds, UMA | 26 values | — | value + `valid_to` |
| **Total** | 18 tables, 26 values | 186 rows | **822 fields, 0 mismatches** |

## Sources (all fetched 2026-10-14, HTTP 200)

| `sources[].id` | URL resolves to | Check |
|---|---|---|
| anexo8-rmf-2026 | SAT PDF, header "DIARIO OFICIAL, Domingo 28 de diciembre de 2025 … ANEXO 8 DE LA RMF PARA 2026" | correct document |
| — (cross-check only) | `dof.gob.mx/nota_detalle.php?codigo=5777219&fecha=28/12/2025` (DOF HTML edition) | used for my copy |
| lisr | Cámara de Diputados LISR, "Última Reforma DOF 01-04-2024" | matches title |
| lif-2026 | LIF 2026, "Nueva Ley DOF 07-11-2025"; art. 24 (0.90 %), art. 25 fr. VI (2.5 % / 2.5 % / 20 %) | correct |
| liva | LIVA, "Última Reforma DOF 12-11-2021"; art. 1o. 16 %, 2o.-A 0 %, 18-J fr. II a) 50 % / 100 % | correct |
| rliva | RLIVA, "Última Reforma DOF 25-09-2014"; art. 3 fr. I (2/3), fr. II (4 %) | correct |
| dof-frontera-norte-2018 | DOF 31/12/2018, Decreto región fronteriza norte; art. Décimo Primero, "crédito equivalente al 50% de la tasa" | correct |
| dof-frontera-sur-2020 | DOF 30/12/2020, Decreto región fronteriza sur; art. Décimo, same credit | correct |
| dof-fronteras-2025 | DOF 31/12/2025; arts. Primero and Segundo reform transitorio Primero of both decrees: "vigente hasta el 31 de diciembre de 2026" | correct |
| dof-uma-2025 | DOF 10/01/2025, INEGI: 113.14 / 3,439.46 / 41,273.52 from 2025-02-01 | correct |
| dof-uma-2026 | DOF 09/01/2026, INEGI: 117.31 / 3,566.22 / 42,794.64 from 2026-02-01 | correct |

The INEGI page `inegi.org.mx/temas/uma/` loads its values with JavaScript and shows none to a non-browser fetch. The INEGI notices published in the DOF, listed above, are the legal publication, and those are what I used.

## Specific checks

- **UMA dates.** The 2025 values apply from 2026-01-01 to 2026-01-31 and the 2026 values from 2026-02-01. The UMA law, art. 5, starts each year's values on 1 February. `valid_to: 2027-01-31` on the 2026 entries follows from that rule; it is not a printed date. The tests already check the 31 Jan / 1 Feb switch.
- **Border IVA (8 %).** `valid_to: 2026-12-31`, and the schema treats ranges as inclusive at both ends (`docs/architecture/tax-params-schema.md` §Resolution by date). The rate therefore applies through 2026-12-31 and lapses on 2027-01-01, which matches "vigente hasta el 31 de diciembre de 2026". Both edges are tested. The value 0.08 is 16 % × (1 − 50 % credit), because neither decree prints "8 %".
- **Art. 106 tables are stored as published, not derived.** In my DOF copy, December differs from 12 × January in 29 cells. Examples: row 1 upper 10,135.11 against a derived 10,135.08; row 2 fee 194.59 against 194.64; row 3 fee 5,051.37 against 5,051.40. `v1.yaml` matches the published value in all 29 cells. In the source, C.II (annual) equals December, art. 116 monthly equals B.V, and B.VII (RIF) equals February. These are facts of the source, not shortcuts in the file.
- **2/3 IVA share kept in code (`fiscal.ShareTwoThirds`): sound.**
  1. RLIVA 3 fr. I says "las dos terceras partes del impuesto que se les traslade". That is a fraction of the IVA, not a rate on the price, and 2/3 is not a terminating decimal, so a 6-decimal `rate` would be off by up to 0.00005 × the IVA amount.
  2. `Share.Suggest` multiplies the exact amount by Num/Den and rounds once to the centavo (half away from zero), so no rounding error builds up.
  3. It is a regulation constant (RLIVA has been unchanged since 2014), not a yearly parameter.
  4. `ShareHalf` and `ShareWhole` (LIVA 18-J fr. II a) are terminating decimals, so the file's "a 6-decimal rate cannot hold 2/3" reason does not apply to them. Keeping them next to the 2/3 share, as fractions of the IVA, is still the right design. The difference is the base, not the precision.
  5. Caveat: a future RLIVA/LIVA reform will not show up in a params version bump. A code change is needed, so watch for one.

## Mismatches found / fixed

None. No fixes, so no new failing tests were needed. The existing spot tests (`embedded_2026_test.go`) already pin the values that matter at each boundary.

## Recommendations (non-blocking)

1. Add the DOF edition (nota 5777219) as a second `anexo8` source, or as the primary one. The SAT PDF is a copy, and the DOF publication is the authoritative text.
2. The LISR, LIVA and LIF URLs on diputados.gob.mx point to a living consolidated text, so after a reform they will return different content. Each source title already records "última reforma …". Consider keeping a SHA-256 of each fetched document in the review trail so the evidence can be checked again later.
3. LIF 2026 art. 25 fr. IX a) adds IVA withholding by platforms on **personas morales**, under 18-J fr. II a), first paragraph. `platformPolicy` currently models personas físicas only. Confirm the scope with the module spec. This is outside `v1.yaml`.
