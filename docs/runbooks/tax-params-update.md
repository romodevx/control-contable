# Runbook: yearly tax parameter update (December–January)

Every year the tariffs, rates and thresholds the tax engine reads change on 1 January (UMA on
1 February). This runbook takes the next year's official values into
`backend/internal/taxengine/params/data/<year>/v1.yaml`, verifies them independently, and
ships them. It also covers a mid-year correction (`v<N+1>.yaml`).

| | |
|---|---|
| Owner | backend-sr (transcription), qa-sr (independent verification), tech-lead (review), owner + CPA (sign-off of anything new) |
| Schema and selection rules | [tax-params-schema.md](../architecture/tax-params-schema.md), [ADR-0002 §3](../adr/0002-tax-engine.md#3-versioned-parameters-data-not-code) |
| Precedent | M5-04 (2026 transcription, PR #73), [M5-06 verification](../reviews/M5-06-params-verification.md) (822 fields, 0 mismatches) |
| Gate | [M5-31 §5 P-9](../reviews/M5-31-security-gate.md#5-preconditions-to-enable-module_tax_engine-in-staging-or-production): before January of a year, a reviewed set for that year, or the provisional banner accepted by the owner |
| Calendar | Open the task in **mid-November** (LIF), transcribe in the **last week of December** (Anexo 8, decrees), finish in the **second week of January** (UMA) |

Commands below run from the repository root unless they start with `cd backend`.

## 1. What happens if nothing is done

The engine never stops on 1 January, and it never guesses:

| From | Behaviour without a `<year>` file | Seen by the user |
|---|---|---|
| 1 January | `Catalog.Select(year)` falls back to the **previous year's** highest version (`Fallback = Provisional = true`), **one year back only** | Every estimate of the new year carries the `provisional_parameters` warning: "Aún no se publican las tarifas de 2027; usamos las de 2026." The parameter set in the breakdown header says 2026 |
| 1 January | Values are still resolved **by date**. Tariffs and rates written without `valid_to` carry over. Values whose `valid_to` is 31 December stop resolving (`ErrNotInForce`) | A record with the 8 % border rate dated in the new year is refused (`422 rate_not_valid_on_date`), unless the decree was extended. The 625 estimate of a goods-and-services seller with the definitive option reads a lapsed LIF rate and fails closed (`500 internal_error`). The record-form suggestions that read a lapsed rate (platform withholding for goods and services or for personas morales) report it as not in force |
| 1 February | The previous year's UMA (`valid_to` 31 January of the new year) lapses | Nothing today: no production code reads the UMA yet (the rate catalog exposes it). It matters as soon as a module does |
| 1 January of the year after | No set for the year and none for the year before | `422 parameters_unavailable`: no estimate rather than a wrong one |

In the 2026 file these values end on 2026-12-31: `iva.rate.border_8` (border decree),
`isr.withholding.interest_annual` (LIF art. 24), `isr.platform.goods_services`,
`isr.platform.persona_moral`, `isr.platform.persona_moral_no_rfc` (LIF art. 25 fr. VI). The
UMA entries end on 2027-01-31. Check with:

```sh
grep -n -B4 'valid_to' backend/internal/taxengine/params/data/2026/v1.yaml | grep -E 'key:|valid_to:'
```

So the target is a merged and deployed `<year>/v1.yaml` **in the first days of January**. Until
then the provisional banner is the honest state; the lapsed values above are not.

## 2. Where the official values are published

Transcribe only from these. A secondary source (a blog, a calculator, a firm's summary) may
help you find a value, never be the value.

| What | Publication | Usually | Keys |
|---|---|---|---|
| ISR tariffs: art. 96 monthly, the **twelve** art. 106 accumulated tables, art. 116, RIF bimonthly, art. 152 annual | **Anexo 8 de la RMF** for the year, DOF ([dof.gob.mx](https://dof.gob.mx/), HTML edition) and SAT's PDF copy ([sat.gob.mx](https://www.sat.gob.mx/), RMF and anexos) | Last days of December (2026: DOF 28-12-2025) | `isr.art96.monthly`, `isr.art106.accumulated` (months 1–12), `isr.art116.monthly`, `isr.rif.bimonthly`, `isr.art152.annual` |
| Annual interest withholding rate, platform rates for goods/services and personas morales, any stimulus of the year | **Ley de Ingresos de la Federación** for the year, DOF; consolidated text at [diputados.gob.mx/LeyesBiblio](https://www.diputados.gob.mx/LeyesBiblio/) | November (2026: DOF 07-11-2025) | `isr.withholding.interest_annual` (art. 24), `isr.platform.goods_services`, `isr.platform.persona_moral`, `isr.platform.persona_moral_no_rfc` (art. 25) |
| Annual return: personal-deduction limits (LISR 151), $400,000 salary and $500,000 informative thresholds (LISR 98 fr. III e), 150 ¶3), deadlines PF / PM / 603 (LISR 150, 76 fr. V, 86 fr. III) | LISR consolidated text at diputados.gob.mx; the SAT may extend a deadline by decree or RMF rule | Only when reformed; deadlines move every year (`valid_to` = fiscal year end) | `deduction.personal.*`, `annual.salary_obligation_threshold`, `annual.informative_income_threshold`, `annual.deadline.*` |
| Colegiaturas caps per level | **Decreto que compila diversos beneficios fiscales**, DOF 26-12-2013, art. 1.10 | Only when amended | `deduction.tuition_cap.*` |
| UMA daily, monthly, annual (in force from 1 February) | **INEGI's notice in the DOF** (INEGI's own page renders values with JavaScript; the DOF notice is the legal publication) | Around 10 January (2026: DOF 09-01-2026) | `uma.daily`, `uma.monthly`, `uma.annual` |
| 8 % border IVA (north and south) | **Decreto de estímulos fiscales región fronteriza norte / sur**, or the decree that extends them, DOF | Last days of December (2026: DOF 31-12-2025, in force to 31-12-2026) | `iva.rate.border_8` |
| Statutory rates and caps (16 %, 0 %, 30 %, RESICO tables, 10 % withholdings, 113-A rates, caps) | LISR, LIVA, RLIVA consolidated texts at diputados.gob.mx; a reform appears in the DOF with the *paquete económico* (usually November) | Only when reformed | `iva.rate.*`, `isr.rate.persona_moral`, `isr.resico_pf.*`, `isr.withholding.*`, `regime.*.income_cap`, `deduction.cash_payment_limit`, ... |

Also read the year's **RMF** and the *paquete económico* reform decree for anything that
changes a rule rather than a value (a new regime option, a new withholding). A rule change
is a code change: [add-a-regime guide](../guides/add-a-regime.md) steps 2–9, not this runbook.

Save a copy of every document you transcribe from (PDF or HTML, with the fetch date) outside
the repository, and record its SHA-256 in the verification report (§4), so the evidence can
be checked again after diputados.gob.mx replaces a consolidated text.

## 3. Transcribe into `params/data/<year>/v1.yaml`

On a branch from `dev`: `feature/tax-params-<year>`.

1. Copy the previous year's file as the skeleton (keys, comments, row layout), then **replace
   every value** from the new sources. Do not keep any old value you have not re-read in a
   new source: a carried-over value must cite the new year's source or the unchanged law.
2. Header: `fiscal_year: <year>`, `version: 1`, `published_on:` the day you write it.
   `provisional: true` only if you ship before a source is published (see §6).
3. `sources[]`: one entry per document, with `id`, `title` (including "última reforma …" for a
   consolidated law and the access date), an **https** URL on an official host (`sat.gob.mx`,
   `dof.gob.mx`, `diputados.gob.mx`, `inegi.org.mx`), `dof_date` and `reference`. Every value
   and table cites a source id and an article in `basis`.
4. **Verbatim** rules (the 2026 file's header lists them):
   - amounts as printed, thousands separators removed, quoted, at most 2 decimals
     (`"844.59"`); rates as quoted fractions (`"1.92 %"` → `"0.0192"`), at most 6 decimals.
     An unquoted number is refused (`ErrInexactNumber`), so YAML never reads a float;
   - the twelve art. 106 tables are **as published**, never January × n (December 2026 row 2
     is 194.59, not 194.64). Never "correct" a fixed fee to make a tariff continuous;
     `tariff_continuity_<year>_test.go` bounds the published drift in centavos;
   - a non-terminating legal fraction (the 2/3 IVA withholding) stays in code, not in the file.
5. **Validity dates.** `valid_from` is the first day the value applies; `valid_to` only when
   the law sets an end (a LIF value: 31 December; a decree: its stated end; the UMA: 31
   January of the next year, by the UMA law art. 5). Ranges are inclusive. A key's entries
   must not overlap (`ErrOverlap`). January keeps the previous year's UMA: keep that entry in
   the old file (it already ends on 31 January) and start the new UMA on 1 February.
6. Run the loader and the boot-time validation:

   ```sh
   cd backend && go test -count=1 ./internal/taxengine/params/ ./internal/taxengine/domain/taxparams/
   ```

## 4. Independent double-entry verification (like M5-06)

A different person (qa-sr) verifies the file **without reading it first**. Follow the
[M5-06 method](../reviews/M5-06-params-verification.md#method-double-entry):

1. Read only the key names and the row layout of the new file.
2. Download every source again. Use a **different rendering** from the author's for the
   tariffs (the author parsed the SAT PDF; parse the DOF HTML edition, or the reverse), so an
   extraction error cannot hide in both copies. Type scalars and short tables by hand.
3. Write the second copy to a scratch YAML outside the repository and compare it with the
   file field by field with a script: amounts and rates as decimals, percentages divided by
   100, every `valid_from` and `valid_to`.
4. **Self-test the harness**: change two values in your copy and confirm the diff reports
   them (and every place they repeat), so a clean result is not a vacuous pass.
5. Record the result in `docs/reviews/<task>-params-verification.md`: sources with URL,
   publication date, fetch date and SHA-256; counts compared (tables, rows, fields); every
   mismatch and its fix. A mismatch is fixed with a failing spot test first (§5).
6. The verification is a review on the transcription PR; the PR does not merge without it.

## 5. Tests for the new year

Add `backend/internal/taxengine/params/embedded_<year>_test.go`, copied from
`embedded_2026_test.go`, with expected values typed **from the publications**, not from the
YAML file:

- **every key the modules read**: `TestEmbedded<year>DefinesEveryKeyTheModulesRead`. The boot
  validates each file's shape but not that each key a strategy reads exists; this test is that
  check (M5-31 G-5, follow-up FU6). Without it a missing key boots and the estimates that read
  it fail with `500`;
- every source on an official host;
- spot rows of each tariff (first, a middle row, "en adelante") and the drift bounds;
- **validity edges**: each value with an end date on its last day and the day after (the
  border rate on `<year>-12-31` and `<year+1>-01-01`), the UMA on 31 January (previous year's)
  and 1 February (new), and `Select(<year+1>)` falling back with `Provisional`.

Then update the tests that pinned "the next year is not published yet", which now fail by
design:

- `TestEmbeddedSelects2026v1` (`params/embedded_2026_test.go`): the embedded sets and the
  fallback year;
- `infra/ratecatalog/catalog_test.go`: the cases dated in the new year ("border decree
  lapsed", "LIF 2026 platform rates sunset", "daily, January 2027") now resolve from the new
  set, not provisionally;
- `golden_gate_test.go` generates tariff-edge and validity cases from `paramsFile`
  (`params/data/2026/v1.yaml`) only: extend it to the new file so every row edge of the new
  tariffs is checked against `reference_test.go`.

Year-bearing texts are code today and must change with the file:

- the `not_tax_advice` text (`adapter/http/texts.go`: "LIF 2026 y Anexo 8 RMF 2026") and its
  copies in the contract examples and MODULE-tax-engine §2.3;
- the citations that name the year in the strategies (`grep -rn '2026' backend/internal/taxengine/domain/*.go`:
  "Anexo 8 RMF 2026", "LIF 2026", the border decree's end date). A citation is wording, not a
  figure, so it does not move the golden digest; follow-up FU8 derives them from the set.

```sh
cd backend
go test -count=1 ./internal/taxengine/...
make golden
```

## 6. Provisional fallback and the `provisional: true` file

Two ways to cover January before every source is out:

| Option | When | Effect |
|---|---|---|
| **No file yet** (default) | Anexo 8 not published by the last deploy of the year | Previous year's set, by date, with the `provisional_parameters` warning. Lapsed values fail as in §1 |
| **`provisional: true` file** | The LIF and the decree are published but Anexo 8 is not (or the reverse), and the lapsed values would break estimates | The file holds what is published, verbatim; everything else carried over is cited to the previous year's source. `Selection.Provisional` stays true, so estimates and record forms keep the warning. Ship the final values as `v2.yaml` (`provisional` absent) when the rest is out |

Never put a value in a file that has not been published. The owner decides between the two
and records the decision in the deploy PR (M5-31 P-9).

## 7. `EngineVersion` and `WithholdingTableVersion`

The records fingerprint already hashes the parameter set (fiscal year and version) and
`Provisional`, so a new set moves every affected month's fingerprint by itself: **no bump is
needed for a new year's file alone**. Bump when a test says the figures changed for the same
inputs:

- `TestEngineVersionPinsTheGoldenFigures` fails when a correction (`v<N+1>`) of a year the
  goldens cover changes a figure, or when the generated cases are extended to the new year.
  Then, in one commit: bump `EngineVersion` (`domain/engine.go`, patch), pin the new digest
  (the failure prints it), and add a row to the
  [engine version log](../adr/0002-tax-engine.md#engine-version-log).
- A rule change found while reading the sources is a strategy change: follow the
  [add-a-regime guide](../guides/add-a-regime.md) step 9.

**Withholding rates.** The record forms derive withholdings from the same file
(`isr.withholding.fees_pm`, `isr.withholding.rent_pm`, `isr.withholding.resico_pf`,
`iva.withholding.freight_pm`, `iva.withholding.scrap_pm`, `iva.withholding.federal_threshold`
and the platform rates; [withholdings](../modules/withholdings.md#derivation-catalog-summary)).
Every stored derivation carries `fiscal.WithholdingTableVersion`
(`backend/internal/kernel/fiscal/derive_enums.go`). When the new year changes one of those
rates or a rule of the decision table:

1. Verify the withholding keys by independent double entry, as
   [WH-05](../reviews/WH-05-params-verification.md) did for 2026.
2. Bump `WithholdingTableVersion` in the same PR (for example `2026.1` → `2027.1`), with the
   kernel goldens (`derive_test.go`) updated.
3. After the deploy, run the [withholdings backfill](withholdings-backfill.md#4-after-a-table_version-bump)
   (dry run, owner review, `-apply`) in every environment that has records. Records written
   after the deploy already use the new table.

A new year whose withholding rates are unchanged needs no bump: rates are read by date.

## 8. Release checklist

Copy into the PR description and tick each line.

- [ ] Branch from `dev`; `make hooks` installed (gitleaks)
- [ ] `params/data/<year>/v1.yaml`: every value from a §2 source, `sources[]` complete, `valid_to` only where the law sets it
- [ ] Independent double-entry report in `docs/reviews/`, 0 open mismatches, sources hashed
- [ ] `embedded_<year>_test.go` with every key the modules read and the validity edges; the "next year" tests updated; generated edges extended
- [ ] Year-bearing texts and citations updated (`texts.go`, strategies, contract examples, MODULE §2.3)
- [ ] `EngineVersion` bumped, digest pinned and logged **if** §7 says so
- [ ] Withholding rates unchanged, or verified + `WithholdingTableVersion` bumped + [backfill](withholdings-backfill.md) run after the deploy (§7)
- [ ] Anything new or ambiguous added to [CPA-REVIEW.md](../reviews/CPA-REVIEW.md)
- [ ] `make -C backend check`, `make -C backend test-integration`, `make contract-lint`, `make api-error-tables-check` green; CI required checks green
- [ ] Tech-lead review; security gate for the deploy (REL-09)
- [ ] Deployed: the parameters are embedded (`go:embed`), so a new API build is the release. After the deploy, an estimate for January of the new year shows the new parameter set in the breakdown header and no `provisional_parameters` warning (unless §6 chose a provisional file)
- [ ] Module page and this runbook updated if anything here was wrong (stale docs are bugs)

## 9. Rollback

A wrong value is fixed **forward** with `v<N+1>.yaml` (the highest version wins); the bad
version stays in the repository as history. If a deploy must be undone before the fix, roll
back the API build: the previous build embeds the previous sets, and the year falls back
to the provisional state of §1.
