# Tax engine golden cases (PLAN M5-07)

Hand-authored expected results for the tax engine, taken from the worked examples of
`docs/research/tax-regimes-mx-2026.md` §A.5 / §B.5 or from a CPA. Cases with no figure to
take from a worked example (informational regimes, a partial regime's month without a
figure) cite the spec rule they pin (`spec:<ref>`). The gate is
`internal/taxengine/golden_gate_test.go`; it runs in `go test ./...`, so a mismatch fails CI.

**There is no update flag.** Expected values are never written by a program: change them by
hand, in a reviewed PR, citing the source. `TestGoldensAreNeverRewritten` fails if any code can
regenerate a golden (an update-style flag or env var, or a file write under
`internal/taxengine/`), and the loader refuses a file without the hand-authored header (YAML
serializers drop comments).

## Layout

`<regime>-<pf|pm>/<id>.yaml`, one case per file, file name = `id`. README files are allowed.

## Format

```yaml
# golden: hand-authored from <source>. Never generated or rewritten by a tool.   <- line 1, required
id: PF-612-ISR-01                 # PF-/PM-/TX- and upper-case segments
source: research:PF-612-ISR-01    # research:<worked example id> | cpa:<ref> | spec:<ref>
skip_reason: "⚠️ ..."             # optional: an unverified rule; the case is reported skipped
pending_strategy: true            # optional: the strategy is not registered yet
input:
  regime: "612"                   # SAT code, quoted
  person_type: fisica             # fisica | moral
  period: "2026-03"
  active_months: [3]              # optional; default January..period
  options: {deduction_mode: blind_35, definitive_option: true, activities: [lodging]}
  months:                         # MonthSums per month, January..period; absent amounts are 0
    - month: 3
      income:   {cash, accrual, by_rate: {rate_16: {base, iva}, rate_8: {base, iva}, rate_0_base, exempt_base}, iva_withheld: {by_moral, by_platform}, isr_withheld: {...}}
      expenses: {deductible_paid, non_deductible_paid, cash_over_limit_paid, iva_paid: {rate_16: {base, iva}, ...}, creditable_iva_paid, creditable_iva_withheld, predial_paid, isr_withheld_to_remit, iva_withheld_to_remit}
  tax_inputs:                     # one value, matching the kind
    - {key: prior_provisional_payments, kind: money_monthly, month: 1, money: "2000.00"}
    - {key: coeficiente_utilidad, kind: rate, rate: "0.1500"}
    - {key: first_fiscal_year, kind: boolean, bool: true}
expected:
  status: complete                # complete | incomplete | partial | informational
  taxes: [isr]                    # what the case asserts: isr, iva, withholding
  isr: "150.96"                   # TaxResult amount; absent with isr asserted = no ISR figure
  iva: {charged_collected, creditable_paid, withheld_from_taxpayer, definitive_excluded, net}
  withheld_to_remit: "5333.33"    # needs withholding in taxes
  lines:                          # every line of the asserted taxes, in order ([] if none)
    - {tax: isr, code: isr.payable, op: result, amount: "150.96"}
  missing_inputs: [{key: prior_provisional_payments, months: [1, 2]}]   # absent = none
  cpa_topics: [rif_reduction]                                           # absent = none
  disclaimers: [deductibility_claimed]                                  # absent = none
  explanation: "regime_explanation:605/fisica"   # partial/informational label key; absent = not asserted
```

Amounts are quoted decimals with exactly two places (the centavo); an unquoted number is a YAML
float and is refused. `iva.net` must equal charged − excluded − creditable − withheld
(MODULE-tax-engine §2.6). Lists are compared as sets; lines in order, by tax, code, op and amount (`op` is the contract operation: a `subtract` line shows a positive amount).
Every line except `info` and `rounding_adjustment` must cite a legal basis. Missing inputs come
from `domain.Registry.Estimate` (the strategy's `RequiredInputs` against the case's `tax_inputs`).

## Skips and pending strategies (never silent)

- `skip_reason` marks a case whose expected value depends on a ⚠️ rule (ADR-0002 §5). It stays
  skipped until the CPA signs off (M5-33); the reason is printed by `go test -v` and in the summary.
- `pending_strategy: true` marks a case whose strategy does not exist yet. The runner **fails**
  a case with no strategy and no marker, and **fails** a marker once the strategy is registered.
- Generated boundary cases are pending until their strategy is registered, then they run.

- **Ratchet:** `testdata/golden_baseline.yaml` lists the allowed skipped ids and the pending count
  per strategy key. `TestGoldenCases` fails when a run has a skip not listed there or more pending
  cases for a strategy than its count, so a new golden must run instead of waiting. It also fails
  when the run is below the baseline: the PR that registers a strategy lowers its count, and the
  CPA sign-off that lifts a skip removes its id. The baseline is edited by hand, only downwards.

Run `make golden` (in `backend/`) to see the summary with every skip reason and pending count.

## Line codes

The line codes of the seeded cases (`isr.income_ytd`, `isr.tariff_row`, `iva.net`, ...) are a
proposal written before the strategies. The strategy PR (M5-11..M5-19) that registers a strategy
may rename codes and split or merge intermediate lines **by hand** in its goldens, in the same
PR that removes `pending_strategy`. The `expected` totals are the source's figures and never
change without a new source.

## Activating a strategy

1. Add it to `taxengine.Strategies()` (`internal/taxengine/strategies.go`). The gate's
   `gateStrategies` is that same list, so the composition root and the gate cannot disagree; the
   gate runs it through the real `domain.Registry` (M5-10), which also derives the missing inputs
   as the use case does. Cases expecting `informational` run through the registry's fallback with
   no strategy at all. The full checklist (catalog row, texts, `EngineVersion`) is
   `docs/guides/add-a-regime.md`.
2. Remove `pending_strategy` from its cases; adjust line codes by hand if needed.
   Lower (or remove) its count in `testdata/golden_baseline.yaml`.
3. `go test ./internal/taxengine/...` must pass to the centavo, including the generated
   tariff-edge cases of that regime.

## Generated cases

`golden_gate_test.go` also generates, from `params/data/2026/v1.yaml`:

- strategy cases at every tariff row edge (lower, upper, upper + 0.01) of 612 PF (the twelve
  accumulated tables), 606 PF (art. 116) and 626 PF (RESICO), with the income as the whole base;
- direct checks of the engine's tariff lookup at every row edge of every table, and of every
  value and table on the day before, first day, last day and day after its validity.

Their expected values come from `reference_test.go` (linear scan, `math/big`, one rounding),
never from the engine.
