# ANN-05: independent verification of the annual-return parameters

- **Reviewed:** `backend/internal/taxengine/params/data/2026/v1.yaml` as of commit `916a4b3` (ANN-04, branch `feature/annual-params`)
- **Date:** 2026-10-14
- **Result:** **0 value mismatches** across 62 compared fields (18 scalars plus the 11-row Anexo 8 C.II tariff × 4 fields). The deliberately altered copies fail both the diff and the Go spot tests. Five findings about naming, citations and scope are listed under [Findings](#findings). None of them changes a value.

## Method (double entry, different editions)

This review was written by the same agent that wrote ANN-04, in the same PR, so the
independence comes from the **sources**: no value in the second copy comes from a document
the first copy was transcribed from.

| What | ANN-04 transcribed from | ANN-05 second copy read from |
|---|---|---|
| LISR 151 fr. II, III, V; 150 ¶1 and ¶3; 98 fr. III e); 76 fr. V; 86 fr. III | Cámara de Diputados consolidated PDF (última reforma DOF 01-04-2024) | **DOF HTML edition of the original LISR, 11-12-2013** (nota 5325373, *tercera sección*) |
| LISR 151 last paragraph (5 UMA / 15 %) | same consolidated PDF | **DOF HTML of the reform decree of 12-11-2021** (nota 5635286), the last decree that reformed it ("Párrafo reformado DOF 18-11-2015, 12-11-2021") |
| Colegiaturas caps (decree art. 1.10) | DOF HTML (nota 5328028) | **DOF printed edition PDF**, *edición matutina* 26-12-2013 (`26122013-MAT.pdf`) |
| UMA annual 2026 | INEGI notice in the DOF, 09-01-2026 | **INEGI press bulletin PDF** `uma2026.pdf` |
| Anexo 8 C.II, annual tariff for 2026 | SAT's PDF copy of the DOF edition | **DOF HTML edition** (nota 5777219, 28-12-2025) |

1. Every document was downloaded again with `curl` (HTTP 200) into a scratch directory outside
   the repository. Each was converted to text (`pdftotext -layout`, or an HTML-to-text
   converter for DOF pages).
2. A script built the second copy **only by matching the statutory wording** in those texts,
   for example `no exceda del (\d+)% de los ingresos acumulables` or
   `excedan del equivalente a (\w+) salarios mínimos`. It did not hand-copy the numbers, and
   it stops if any phrase is missing. Spelled-out multiples ("cinco", "un") map to integers,
   and percentages are divided by 100.
3. **Deadlines.** The law states a period, not a date. The second copy checks that the period
   wording is present ("en el mes de abril del año siguiente"; "dentro de los tres meses
   siguientes a la fecha en que termine dicho ejercicio"; "a más tardar el día 15 de febrero
   de cada año") and takes the last calendar day of that period for fiscal year 2026.
4. The script then read the ANN-04 keys from `v1.yaml` and compared them field by field as
   decimals. For C.II it compared lower, upper, fixed fee and rate on all 11 rows.

## Counts compared

| Group | Keys | Fields | Mismatches |
|---|---|---|---|
| LISR 151 limits (UMA multiples and shares) | 7 | 7 | 0 |
| Colegiaturas caps (5 levels) | 5 | 5 | 0 |
| Thresholds $400,000 / $500,000 | 2 | 2 | 0 |
| Deadlines PF / PM / 603 informative | 3 | 3 | 0 |
| UMA annual in force 31-12-2026 | 1 | 1 | 0 |
| Anexo 8 C.II (`isr.art152.annual`) | 1 | 44 | 0 |
| **Total** | **19** | **62** | **0** |

The 2013 DOF edition matched for every paragraph not reformed since 2013: fr. II, the fr. III
7 % / 4 % paragraph, the fr. V first paragraph, 98 fr. III e), 150 and 76 fr. V. For 151's
last paragraph the 2013 wording ("cuatro salarios mínimos … 10 %") is obsolete. That is why
the 12-11-2021 decree text is the source there. It reads "cinco veces el valor anual de la
Unidad de Medida y Actualización, o del 15%", which matches the file.

## The altered copies fail (the check is not vacuous)

1. **Second copy altered.** I set the primaria cap to 12,990.00 and the C.II row 3 fixed fee to
   12,140.31 (digits swapped). The diff reported exactly those 2 mismatches and exited 1.
2. **`v1.yaml` altered.** In the working tree I set primaria to 12,990.00 and
   `annual.deadline.pm` to 2027-03-30.
   - The diff reported exactly those 2 mismatches.
   - `go test ./internal/taxengine/params/` failed in `TestEmbedded2026DeductionLimits`
     (`primaria: 1299000` vs `1290000`) and in `TestEmbedded2026AnnualDeadlines/annual.deadline.pm`.
   - I then restored the file (`git checkout`), and the suite was green again.

## Sources (fetched 2026-10-14)

| Document | URL | SHA-256 |
|---|---|---|
| LISR consolidated, última reforma DOF 01-04-2024 (first copy) | https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf | `803c2875c09c751c1966b8eb4dca99f460b237da4ef67ee04b4fa325d8c2f786` |
| Decree DOF 26-12-2013, HTML (first copy) | https://dof.gob.mx/nota_detalle.php?codigo=5328028&fecha=26/12/2013 | `ca1483129ffa0de609cdbc6b3f21b6d6d5772bca16df626ddea03aa267c08399` * |
| Anexo 8 RMF 2026, SAT PDF (first copy) | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/anexos/Anexo-8-RMF-2026_DOF-28122025.pdf | `69ebaf8a78a3bde5d631ad2a994e81ec6d07e2d4e39085c8c948d138f782c1f4` |
| INEGI UMA notice, DOF 09-01-2026 (first copy) | https://dof.gob.mx/nota_detalle.php?codigo=5778072&fecha=09/01/2026 | `81bd2a4cd25755a93e53390aa54a2749a8fcc0f981b693e752489b6f4b494566` * |
| LISR as published, DOF 11-12-2013, 3a sección (second copy) | https://dof.gob.mx/nota_detalle.php?codigo=5325373&fecha=11/12/2013 | `4135a3ae27e0c0775c718527c722bd2ea0c94002ab131afb0ea03f1cebf5be2e` * |
| LISR reform decree, DOF 12-11-2021 (second copy) | https://dof.gob.mx/nota_detalle.php?codigo=5635286&fecha=12/11/2021 | `a50b7f54d745ad7d75530e449d458705316e096876352765b2c3507a1696e606` * |
| DOF printed edition 26-12-2013 matutina (second copy) | https://dof.gob.mx/abrirPDF.php?archivo=26122013-MAT.pdf&anio=2013&repo=repositorio/ | `00c59bf0e44348478ff59861b87a2b9f15e2b1c58b091cfcc7b1a357a369c932` |
| Anexo 8 RMF 2026, DOF HTML (second copy) | https://dof.gob.mx/nota_detalle.php?codigo=5777219&fecha=28/12/2025 | `2e784d1461abf5d41772d64b2598468b2c182606dcbf9206e508e7c25902b2ed` * |
| INEGI UMA 2026 bulletin (second copy) | https://www.inegi.org.mx/contenidos/saladeprensa/boletines/2026/uma/uma2026.pdf | `9c9c4ff3da9b34852d5e3b8165c50c346ff3ef7ba74737489f6d1f544c353e1c` |

\* A DOF HTML page includes the day's exchange-rate sidebar, so its hash changes from day to
day even when the note does not. The PDFs are stable. The diputados.gob.mx LISR PDF changes
whenever the LISR is reformed.

## Findings

1. **Key name.** PLAN ANN-04 and MODULE-annual (two places) name the annual tariff
   `isr.annual_152`, but the file has always held Anexo 8 C.II as `isr.art152.annual` (M5-04,
   verified 0-mismatch by M5-06). The runbook, the drift test and the December-equality test
   all use that name.
   - **Fixed:** I kept the existing key rather than add a duplicate table, and changed
     MODULE-annual to say `isr.art152.annual`. PLAN is not edited. I re-read the table against
     both editions: 0 mismatches.
2. **Wrong article in PR #127.** `annual.Limits.TuitionCaps` is documented as "decree DOF
   26-12-2013 **art. 1.9**". The cap table is in **art. 1.10**; art. 1.9 is about payment means
   and proof. The YAML cites art. 1.10.
   - **Fixed in this PR** (tech-lead review): the cap note cites art. 1.10 and the cash
     exclusion art. 1.9, pinned by `TestTuitionNotesCiteTheDecreeArticleOfTheirRule`; the comments in `limits.go` and `deduction.go` say art. 1.10.
3. **"Salario mínimo" vs UMA.** Art. 151 fr. II (funeral) and fr. V (retirement) still say
   "salario mínimo general … elevado al año"; only the last paragraph says UMA. The file stores
   the multiples as printed (1 and 5). Applying them to the UMA is the consumer's reading, and
   it is already CPA topic `uma_for_minimum_wage` (PR #127).
4. **PTU 10 % not added.** PLAN ANN-04 lists "PTU 10 %". The rate is set by the Comisión
   Nacional para la PTU resolution in the DOF, not by the LISR. I could not retrieve that
   official text in this session: the DOF day indexes I tried did not list it, and the SAT
   PTU manual timed out. The rule is never to type a value from memory, so the key is
   deferred. ANN-10 (`pm_601`) needs it before its PTU line. Follow-up: fetch the resolution
   from the DOF and add `ptu.rate` with its citation. Until then the rate is provisional ⚠️: CPA-REVIEW (Rates and Annual items), topic `ptu_rate`, golden AN-06-PTU skipped.
5. **Later amendments to the colegiaturas decree.** Art. 1.10 is compared against the 2013
   publication, in two editions. I found no later DOF amendment of the amounts, but I did not
   read every decree that amends the compilation decree. Secondary sources dated 2023 and
   2026 still quote the same five amounts.

## Additional checks

- **Deadlines** are valid 2026-01-01 → 2026-12-31 only. A 2027 estimate, which falls back to
  this set, therefore gets `ErrNotInForce` instead of a 2027-04-30 date that belongs to the
  wrong year. The tests cover both edges.
- **UMA for the deduction caps** is read on 31 December (42,794.64). It is not the January
  value carried over from 2025 (41,273.52). `TestReadDeductionLimitsUsesTheUMAInForceAtYearEnd`
  pins this.
- **Continuity.** `tariff_continuity_2026_test.go` (PR #76) already bounds C.II at 1 centavo
  of drift and checks that it equals the December art. 106 table. Both pass on this data.
