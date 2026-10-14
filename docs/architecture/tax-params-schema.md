# Tax parameters: file schema, validation and selection

| Field | Value |
|---|---|
| Implements | [ADR-0002 §3](../adr/0002-tax-engine.md) (versioned parameters: data, not code), PLAN M5-03 |
| Model | `backend/internal/taxengine/domain/taxparams` (pure: no I/O, no clock, no floats) |
| Loader | `backend/internal/taxengine/params` (`go:embed`, strict YAML) |
| Files | `backend/internal/taxengine/params/data/<fiscal_year>/v<N>.yaml` |

The real 2026 values are transcribed in M5-04 (`data/2026/v1.yaml`) and double-checked in
M5-06. The test fixtures under `params/testdata/` are **synthetic** and are not official values.

## Layout

```
internal/taxengine/params/
  data/                 # embedded with go:embed
    README.md           # the only non-parameter file allowed
    2026/
      v1.yaml           # fiscal year 2026, version 1
      v2.yaml           # a correction or a mid-year reform
```

ADR-0002 writes the path as `params/<fiscal_year>/v<N>.yaml`. The files sit one level down in
`data/` because `go:embed` needs a directory that exists before the first year is
transcribed. Any other file or directory under `data/` fails the load (`ErrUnexpectedFile`). A
file's `fiscal_year` and `version` must match its path (`ErrPathMismatch`).

## File schema

```yaml
fiscal_year: 2026           # 2000-9999
version: 1                  # >= 1; the highest version of a year wins
published_on: 2026-10-13    # the day this file was written
provisional: false          # optional; true = values not yet official (see Selection)

sources:                    # at least one; every value and table cites one
  - id: dof-2025-12-31      # unique in the file
    title: Decreto de estímulos fiscales región fronteriza
    url: https://www.dof.gob.mx/...   # absolute https URL
    dof_date: 2025-12-31    # publication date in the DOF
    reference: "Artículo primero"     # article, rule or annex section

values:                     # dated scalars; a key may repeat with non-overlapping ranges
  - key: iva.rate.border_8
    rate: "0.08"            # exactly one of rate | money | flag
    valid_from: 2026-01-01
    valid_to: 2026-12-31    # inclusive; omit for "until a later entry replaces it"
    basis: [{source: dof-2025-12-31, article: "Art. 2"}]
  - key: uma.monthly
    money: "..."            # pesos with at most 2 decimals
    valid_from: 2026-01-01
    valid_to: 2026-01-31    # January still uses the previous year's UMA
    basis: [...]
  - key: rounding.payment_whole_pesos
    flag: false
    valid_from: 2026-01-01
    basis: [...]

tables:                     # tariffs, verbatim as published
  - key: isr.art96.monthly
    kind: marginal          # fixed_fee + rate over the excess of lower (LISR 96, 106)
    valid_from: 2026-01-01  # valid_to optional, as for values
    basis: [...]
    rows:
      - {lower: "0.01", upper: "...", fixed_fee: "0.00", rate: "0.0192"}
      - {lower: "...", fixed_fee: "...", rate: "0.35"}   # no upper = "en adelante"
  - key: isr.resico_pf.monthly
    kind: flat              # non-marginal: the row's rate applies to the whole amount
    valid_from: 2026-01-01
    basis: [...]
    rows:
      - {lower: "0.01", upper: "...", rate: "0.01"}     # flat rows have no fixed_fee
  - key: isr.art106.accumulated
    kind: marginal
    month: 1                # 1-12: one table per month, never derived
    valid_from: 2026-01-01
    basis: [...]
    rows: [...]
```

### Types

| Field | Written as | Held as |
|---|---|---|
| `rate`, row `rate` | **quoted** decimal fraction, at most 6 decimals (`"0.16"`) | `money.Rate` (10^-6, exact) |
| `money`, `lower`, `upper`, `fixed_fee` | **quoted** pesos, at most 2 decimals (`"844.59"`) | `money.Cents` (int64) |
| dates | `YYYY-MM-DD`, plain or quoted | `taxparams.Date` (calendar day, no time zone) |
| `key` | dotted lower-case segments `[a-z0-9_]`, at least two | `taxparams.Key` |

An unquoted number (`rate: 0.08`, `money: 100`) is rejected (`ErrInexactNumber`). YAML would read it
as a binary float before we see it, and money never touches a float. More decimals than the type
holds are rejected, never rounded (`money.ErrPrecision`). The decoder is strict: an unknown field,
an empty file or a second document is `ErrSyntax`.

## Validation (fail-fast, at startup)

`taxparams.NewParams` checks the whole file and reports **every** problem in one error, each with
its path (`values[3] uma.monthly: ...`, `tables[5] isr.art106.accumulated rows[2]: ...`).
`params.Load` does the same across files. A bad file fails the boot and CI, never an estimate.

| Rule | Error |
|---|---|
| `fiscal_year` 2000-9999, `version` >= 1, `published_on` present | `ErrInvalidHeader` |
| Sources: at least one; id, title, reference, `dof_date` present; https URL; unique id | `ErrInvalidSource` |
| Key syntax | `ErrInvalidKey` |
| `valid_from` present; `valid_to` absent or >= `valid_from` | `ErrInvalidValidity` |
| Every value and table has a basis; each citation has an article and a known source id | `ErrInvalidBasis` |
| A value has exactly one of rate, money, flag | `ErrValueKind` (loader) |
| Table kind is `marginal` or `flat` | `ErrTableKind` (loader) |
| One key keeps one kind (rate vs money, value vs table, marginal vs flat) | `ErrKindConflict` |
| Rates in [0, 1], including tariff rates | `ErrRateOutOfRange` |
| Amounts and fixed fees >= 0 | `ErrNegativeMoney` |
| **No overlap**: two entries of one key (and month) in force on the same day | `ErrOverlap` |
| Tables: rows present; first `lower` is 0.01; `upper` >= `lower`; only the last row may omit `upper`; no `fixed_fee` on a flat table | `ErrInvalidTable` |
| **Contiguous rows**: `upper[i] + 0.01 == lower[i+1]` (no gap, no overlap) | `ErrNotContiguous` |
| `month` within 1-12 (absent = not monthly) | `ErrInvalidMonth` |
| A key is all monthly or all plain | `ErrMonthlyMixed` |
| **Exactly twelve**: each validity window of a monthly key holds months 1-12 once | `ErrMonthlyIncomplete` (a repeated month is also `ErrOverlap`) |

Regime strategies also declare what they read, for example
`taxparams.RequireMonthlyTables("isr.art106.accumulated")` or `RequireRate("iva.rate.general")`.
The composition root checks them with `Params.Require` at boot, so a missing or mistyped key
(`ErrUnknownKey`, `ErrKindMismatch`) fails the boot as well.

## Resolution by date

`Params.At(key, date)` returns the value in force on that day. The typed forms are
`Rate`, `Money`, `Flag`, `Table(key, date)` and `MonthlyTable(key, month, date)`.

- Ranges are inclusive at both ends. A missing `valid_to` means the entry stays in force until
  a later entry of the same key starts.
- Entries are sorted by `valid_from` and never overlap. The lookup is therefore a binary search
  for the last entry starting on or before the date, O(log n).
- If nothing covers the date, the result is `ErrNotInForce`. This is how the 8% border rate
  lapses on 2027-01-01 unless a new version extends it. An unknown key is `ErrUnknownKey`, and
  the wrong accessor is `ErrKindMismatch`.
- Callers pass the date. The package never reads a clock. Period-level values use the last day
  of the period, and record-level checks use the record's date (ADR-0002 §3).
- Lookups return copies, so a `Params` is immutable once built.

## Version selection and the provisional fallback

`Catalog.Select(fiscalYear)` returns a `Selection{Params, RequestedYear, Provisional, Fallback}`:

1. If the year has sets, the **highest version** is used. `Provisional` is true only if that
   file says `provisional: true`.
2. If the year has no set yet (for example January, before the new Anexo 8 is transcribed),
   the **previous year's** highest version is used with `Fallback = Provisional = true`.
   Estimates then carry the `provisional_parameters` disclaimer. Values still resolve by date
   in this case: tariffs written without a `valid_to` carry over, and a decree that lapsed on
   31 December stays lapsed.
3. The fallback goes back **one year only**, and a later year never stands in for an earlier
   one. Anything else is `ErrNoParameters`, because parameters that stale should stop the
   estimate rather than quietly produce a wrong number.

## Adding a year or a correction

1. Create `data/<year>/v1.yaml` (or `v<N+1>.yaml` for a correction). Copy the values verbatim
   and cite a source for each.
2. Leave `valid_to` empty when the law sets no end. Set it when the law does, as with a
   decree's sunset.
3. Run `make check`. The loader test and the golden suite (M5-07) gate the PR, and the CPA
   reviews the goldens.
