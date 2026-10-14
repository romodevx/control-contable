# WH-05: independent verification of the withholding parameters

- **Reviewed:** branch `feature/withholdings-params`, `backend/internal/taxengine/params/data/2026/v1.yaml` (keys `isr.withholding.fees_pm`, `isr.withholding.rent_pm`, `isr.withholding.resico_pf`, `iva.withholding.freight_pm`, `iva.withholding.scrap_pm`; sources `liva`, `rliva`)
- **Reviewer:** QA (qa-sr), 2026-10-14
- **Result:** 0 mismatches across 25 compared fields. Nothing in the file needed a fix. Two non-blocking notes on wording (see Recommendations).

## Method (double entry)

1. I did not read the values, `basis` entries or comments of the keys under review before building my own copy. I read only the key names and the field layout (`key`, `rate`, `valid_from`, `valid_to`, `basis[{source, article}]`; `sources[].id/url/dof_date/reference`).
2. I downloaded LIVA, RLIVA and LISR again myself (`curl -L`), hashed them, and extracted text with `pdftotext -layout`. I found each rate in the legal text myself. I used `docs/research/withholdings-matrix-2026.md` only as a pointer to article numbers.
3. I wrote my second copy to a scratch JSON outside the repo: key, rate as a decimal fraction, `valid_from` 2026-01-01, `valid_to` none, and the article(s) each key must cite. For `liva` and `rliva` I recorded `url` and `dof_date`.
4. A script loaded `v1.yaml` with a minimal YAML reader (PyYAML is not installed) and compared field by field. Rates were compared as `Decimal`, dates as strings, `valid_to` as present/absent, sources by `url` and `dof_date`. For citations it checked at article level: some `basis` entry with the right `source` must match the article (e.g. `106`, `116`, `113-J`, `1o.-A … II … c)`, `3 fr. II`).
5. Its first run reported 1 mismatch. The cause was my own pattern for RLIVA 3 fr. II, which was anchored at the start of the string. The file reads "Art. 3 fr. II", which is correct. I fixed the pattern, not the data, and the rerun was clean.

## Counts compared

| Group | Keys / sources | Fields per item | Fields |
|---|---|---|---|
| ISR withholding (LISR 106, 116, 113-J) | 3 keys | rate, valid_from, valid_to, 1 citation | 12 |
| IVA withholding, freight (LIVA 1o.-A II c), RLIVA 3 II) | 1 key | rate, valid_from, valid_to, 2 citations | 5 |
| IVA withholding, scrap (LIVA 1o.-A II b)) | 1 key | rate, valid_from, valid_to, 1 citation | 4 |
| Sources `liva`, `rliva` | 2 sources | url, dof_date | 4 |
| **Total** | 5 keys, 2 sources | | **25 fields, 0 mismatches** |

My copy against `v1.yaml`:

| Key | Mine | Legal text | Cited in `v1.yaml` |
|---|---|---|---|
| isr.withholding.fees_pm | 0.10 | LISR 106, penultimate ¶: "la tasa del 10% sobre el monto de los pagos … sin deducción alguna" | Art. 106, penúltimo párrafo |
| isr.withholding.rent_pm | 0.10 | LISR 116, 5th (penultimate) ¶: "la tasa del 10% sobre el monto de los mismos, sin deducción alguna" | Art. 116 |
| isr.withholding.resico_pf | 0.0125 | LISR 113-J: "la tasa del 1.25% sobre el monto de los pagos … sin considerar el impuesto al valor agregado" | Art. 113-J |
| iva.withholding.freight_pm | 0.04 | RLIVA 3 fr. II: "por el 4% del valor de la contraprestación pagada efectivamente"; obligation in LIVA 1o.-A fr. II c) | RLIVA Art. 3 fr. II; LIVA Art. 1o.-A fr. II inciso c) |
| iva.withholding.scrap_pm | 1 | LIVA 1o.-A fr. II b): "retención del impuesto que se les traslade"; RLIVA 3 reduces only fr. I and fr. II cases | LIVA 1o.-A fr. II b) and último párrafo; RLIVA Art. 3 |

All five keys have `valid_from: 2026-01-01` and no `valid_to`. That is correct: these rates are permanent law (LISR, LIVA, RLIVA), not yearly LIF or decree rates. None of the three documents carries a sunset for these articles.

## Sources (all fetched 2026-10-14, HTTP 200, no redirect)

| `sources[].id` | URL | SHA-256 | Última Reforma | `dof_date` in file |
|---|---|---|---|---|
| liva | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIVA.pdf | `f2b5b1770441a6a4e48f3ade408f9ff7a3e9c81bd5d529e52484f49e4a8e91a6` | DOF 12-11-2021 | 2021-11-12 (match) |
| rliva | https://www.diputados.gob.mx/LeyesBiblio/regley/Reg_LIVA_250914.pdf | `4f46c1f92fb5682281a7679836e517de95975786fa5326848a331c9eb946cc5f` | DOF 25-09-2014 | 2014-09-25 (match) |
| lisr (cross-check only) | https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf | `803c2875c09c751c1966b8eb4dca99f460b237da4ef67ee04b4fa325d8c2f786` | DOF 01-04-2024 | not under review |

## Specific checks

- **Base of each rate.**
  - `fees_pm`, `rent_pm`: a rate on the gross payment ("sin deducción alguna").
  - `resico_pf`: a rate on the payment without IVA.
  - `freight_pm`: a rate on the **price** ("valor de la contraprestación pagada efectivamente"), not a share of the IVA. 4 % of the price equals 25 % of a 16 % IVA, so the consumer must apply 0.04 to the subtotal, never to the IVA.
  - `scrap_pm`: a share of the **IVA** charged (the whole of it). 1 applied to the price would be wrong.
  - The yaml comment block above the two IVA keys states this difference in base explicitly. `ratecatalog/doc.go` line 29 also says "100 % of the IVA".
- **2/3 share stays in code: correct.** RLIVA 3 fr. I ("las dos terceras partes del impuesto que se les traslade") is a fraction of the IVA that cannot be written as a terminating decimal. It lives in `fiscal.ShareTwoThirds = Share{Num: 2, Den: 3}` (`backend/internal/kernel/fiscal/withholding.go:20`). The expenses and incomes policies use it, and no yaml key duplicates it. The caveat from M5-06 still applies: a future RLIVA reform needs a code change, not a params version bump.
- **Reciclables.** Neither LIVA nor RLIVA contains the word "reciclable" (grep over both extracted texts). LIVA 1o.-A fr. II b) covers only "desperdicios para ser utilizados como insumo de su actividad industrial o para su comercialización". RLIVA's only "desperdicios" mention (art. 47, waste from temporary imports) is unrelated. Any definition that brings recyclables in comes from SAT secondary rules (RMF / criterios normativos), which **I did not fetch**. The yaml comment "Desperdicios (and reciclables bought as such)" is an interpretation, not legal text. The rate is unaffected, because the law makes it 100 % of the IVA either way.
- **Loader.** `go test -count=1 ./internal/taxengine/params/ ./internal/taxengine/infra/ratecatalog/` passes: `params` ok 0.233s, `ratecatalog` ok 0.396s.

## Harness self-test

To show that a clean result is not a vacuous pass, I ran three altered copies of my second copy against the unchanged `v1.yaml`:

1. `scrap_pm` 1 → 0.6667 and `freight_pm` 0.04 → 0.4:
   ```
   MISMATCH iva.withholding.freight_pm.rate: yaml=0.04 mine=0.4
   MISMATCH iva.withholding.scrap_pm.rate: yaml=1 mine=0.6667
   fields compared: 25; mismatches: 2
   ```
2. `rliva` dof_date 2014-09-25 → 2014-09-24, and the `resico_pf` citation pattern 113-J → 113-I:
   ```
   MISMATCH isr.withholding.resico_pf.basis: no citation of LISR 113-J … [lisr]; yaml cites [('lisr', 'Art. 113-J')]
   MISMATCH sources.rliva.dof_date: yaml=2014-09-25 mine=2014-09-24
   fields compared: 25; mismatches: 2
   ```
   (The label still reads 113-J because only the pattern was altered.)
3. Every `valid_to` none → 2026-12-31: 5 `valid_to` mismatches, one per key.

Each run reported exactly the alterations made, and nothing else.

## Mismatches found / fixed

None in `v1.yaml`. The only mismatch seen during the work was in my own harness (step 5 of Method), and I fixed it there.

## Recommendations (non-blocking)

1. Cite `rent_pm` as "Art. 116, penúltimo párrafo" (it is the 5th of 6 paragraphs), to match the precision of `fees_pm`'s "Art. 106, penúltimo párrafo".
2. Reword the `scrap_pm` comment so "reciclables" is not presented as part of LIVA 1o.-A fr. II b). Either cite the SAT rule or criterion that brings recyclables under "desperdicios" as a `sources` entry, or drop the parenthetical. Today the claim has no source in the file.
3. Keep the SHA-256 values above in the review trail. All three URLs serve living consolidated texts.

## Author follow-up

The author applied recommendations 1 and 2 after this review, on the same branch. `rent_pm` now cites "Art. 116, penúltimo párrafo"; the author re-checked that the 10 % paragraph is followed only by the paragraph on the optional comprobante. The `scrap_pm` comment no longer mentions "reciclables" and paraphrases only LIVA 1o.-A fr. II b) ("como insumo de su actividad industrial o para su comercialización"). No rate, date or source changed, so the 25-field comparison above still holds.

## Addendum: `iva.withholding.federal_threshold` (WH-04 review, PR #124)

- **Reviewed:** the key added to `v1.yaml` for IVA-7 (`fiscal.KeyIVAFederalWithholdingThreshold`), new source `rmf-2026`, and the `liva` reference. **Reviewer:** tech-lead, 2026-10-14. **Result:** 0 mismatches across 9 fields (value, `valid_from`, `valid_to`, 2 citations, source `url`, `dof_date`, `reference`, `liva` `dof_date`).
- **Method.** I fetched both texts myself (`curl -L`, HTTP 200, no redirect), extracted them with `pdftotext -layout`, and read the rule before writing the value. The research doc was only a pointer to the rule number.

| Source | URL | SHA-256 | Edition |
|---|---|---|---|
| `rmf-2026` | https://www.sat.gob.mx/minisitio/NormatividadRMFyRGCE/documentos2026/rmf/rmf/RMF_2026-DOF-28122025.pdf | `6f71ce0cccf430171fe858c5d244b1c321fa68306f962d44ec4977858a52919a` | DOF 28-12-2025, 666 pages |
| `liva` | https://www.diputados.gob.mx/LeyesBiblio/pdf/LIVA.pdf | `f2b5b1770441a6a4e48f3ade408f9ff7a3e9c81bd5d529e52484f49e4a8e91a6` (same as above) | DOF 12-11-2021 |

- **RMF 2026 rule 4.1.3, first ¶:** "la Federación y sus organismos descentralizados, no estarán a lo establecido por el citado artículo, por las erogaciones que efectúen por concepto de adquisición de bienes o prestación de servicios, siempre que el monto del precio o de la contraprestación pactada no rebase la cantidad de $2,000.00". So the value is `2000.00`, and the comparison is "not above": exactly $2,000.00 is not withheld. The kernel implements this as `Subtotal <= threshold → IVA-8`.
- **Second ¶:** "No será aplicable … tratándose de los servicios personales independientes y de autotransporte terrestre de bienes … independientemente del monto". The kernel matches: professional services never compare against the threshold, and freight is decided earlier by IVA-4.
- **LIVA 3o., third ¶** gives the Federación's duty to withhold under 1o.-A, and only the Federación's: "Los Estados, el Distrito Federal y los Municipios … no efectuarán la retención". The kernel matches with IVA-2.
- **Validity.** An RMF lasts one fiscal year, so `valid_to: 2026-12-31`. This is the same pattern as the LIF 2026 keys. In 2027 the derivation returns `rate_not_valid_on_date` until the RMF 2027 is transcribed. That is deliberate: the value must not be guessed.
- **Tests and mutation.** `TestEmbedded2026Thresholds`, `TestEmbedded2026AmountsThatLapseAtYearEnd`, `ratecatalog.TestIVAWithholdingFederalThreshold*`, `TestNewRequiresTheFederalThreshold` and `TestDeriveWithholdingsOverThe2026Parameters` compare the file against the values above. Each mutation failed the suite and was then reverted: the yaml value set to `2000.01`, the boot `RequireMoney` dropped, and the kernel comparison hard-coded to `2_000_00`.
- **Open question for the CPA (A2, non-blocking).** The relief in ¶1 covers "adquisición de bienes o prestación de servicios". LIVA treats *uso o goce temporal* (rent) as a separate act (art. 1o. fr. III), so read literally, rent paid by the Federación is withheld at any amount. The research (WM §1 row "606 rent") applies the $2,000 threshold to rent, and the kernel follows the research. The difference matters only for rent of $2,000.00 or less paid by the Federación. It is added to CPA item A2.
