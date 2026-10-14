# Annual estimate golden cases (PLAN ANN-06)

Hand-authored expected results for the annual estimate (MODULE-annual §3), taken from the
worked examples AN-01..AN-08 of `docs/research/annual-return-2026.md` Part D or from a CPA.
The gate is `internal/taxengine/annual_golden_gate_test.go`; it runs in `go test ./...`, so a
mismatch fails CI (`make test`), and `make golden` prints its summary. The harness is
`internal/taxengine/annualgolden`; it mirrors the monthly one (`../golden/README.md`) and
shares its rules.

**There is no update flag.** Expected values are never written by a program: change them by
hand, in a reviewed PR, citing the source. `TestGoldensAreNeverRewritten` scans the whole
module for a rewrite path, and `TestGoldenAnnualFilesKeepTheirHeader` plus the loader refuse a
file without the hand-authored header (YAML serializers drop comments).

## Layout

`<pf|pm>/<id>.yaml`, one case per file, file name = `id`. README files are allowed.
`TestGoldenAnnualSourcesAreTheResearchWorkedExamples` ties every `research:` source to a
worked example and every worked example to at least one case.

## Format

```yaml
# golden: hand-authored from <source>. Never generated or rewritten by a tool.   <- line 1, required
id: AN-03-A                       # AN-<nn> plus optional upper-case segments
source: research:AN-03            # research:AN-<nn> | cpa:<ref> | spec:<ref>
skip_reason: "⚠️ ..."             # optional: an unverified rule; the case is reported skipped
pending_strategy: true            # optional: a section the case expects is not registered yet
input:
  fiscal_year: 2026
  person_type: fisica             # fisica | moral
  regimes:                        # the regimes held in the year, with the year's figures
    - regime: "612"               # SAT code, quoted
      income: "900000.00"         # collected (PF, RESICO) or accrued (601)
      deductions: "400000.00"     # authorized deductions paid
      predial: "0.00"             # 606
      isr_withheld: "30000.00"    # withheld from the taxpayer (10 % PM, platforms)
      options: {deduction_mode: blind_35}
      tax_inputs: [{key: ptu_paid, money: "0.00"}]   # money tax inputs of the regime
  payments:                       # provisional payments, already resolved (precedence is ANN-14's)
    - {regime: "612", month: 3, amount: "1000.00", source: filed}   # month absent = the source's year total
  salary_statements:              # constancias (605)
    - {employer: "A", accumulable: "0.00", exempt: "0.00", isr_withheld: "0.00",
       simultaneous: false, left_before_year_end: false, payer_not_withholding: false}
  personal_deductions:            # annual.Deduction, validated by the domain
    - {category: tuition, amount: "30000.00", payment_method: electronic, beneficiary: descendant,
       student: "Hijo 1", tuition_level: primaria, scholarship: "0.00", donee_type: ...}
  annual_inputs: {prior_year_accumulable_income: "700000.00", other_exempt_income: "0.00",
                  told_employer_will_file: false}
expected:
  informative_income_over_500k: true      # optional
  disclaimers: [annual_estimate]          # optional, compared as a set
  sections:                               # every section of the estimate, in order
    - kind: pf_art152                     # pf_art152 | pf_resico_check | pm_601 | pm_resico | explanation
      regimes: ["612"]                    # optional, set
      status: complete                    # complete | incomplete | informational
      obligation: obligated               # optional
      missing_inputs: []                  # optional, set
      cpa_topics: []                      # optional, set
      tax:                                # optional; each figure optional
        accumulable_income, personal_deductions, base, isr: "..."
        credits: {payments, withholdings, salary_isr}
        balance: {kind: a_favor, amount: "9408.00"}   # amount never negative: the kind is the sign
      difference: "-1320.00"              # pf_resico_check informative difference
      pm_extras: {ptu_base, ptu_to_distribute, suggested_cu_next_year: "0.1380"}
      lines:                              # optional: must appear in this relative order
        - {code: deductions.cap_applied, op: info, amount: "12900.00"}
```

Amounts are quoted decimals with exactly two places; an unquoted number is a YAML float and is
refused. Unlike the monthly cases, an absent list or figure is **not asserted**: the research
examples state some figures and not others, and every asserted value must come from the source.
Line codes other than the deductions notes (`deductions.cap_applied`,
`deductions.excluded_cash`, `deductions.excluded_unverified`, ANN-07) are proposals until the
section's strategy PR fixes them, by hand, in the same PR that removes `pending_strategy`.
When a section lists `lines`, its `deductions.*` lines are compared **exactly** (every "Tope
aplicado" and exclusion, in order): a cap line the golden does not list fails. Every computed
line other than `info` and rounding must cite its legal basis, as in the monthly harness.

## Skips, pending sections and the ratchet (never silent)

- `skip_reason` marks a case whose expected value depends on a ⚠️ rule (MODULE-annual §2.10).
  It stays skipped until the CPA signs off (ANN-27). AN-07's PTU base is asserted apart, in
  `AN-07-PTU`, so the balance of AN-07 runs while its PTU base waits.
- `pending_strategy: true` marks a case whose sections are not registered. The runner **fails**
  a case with an unregistered section and no marker, and **fails** a marker once every section
  it expects is registered.
- `testdata/annual_golden_baseline.yaml` lists the allowed skipped ids and the pending count per
  section kind; the gate fails above it and below it, so it only goes down, by hand.

## Activating a section

1. Register the section (ANN-08..ANN-11) and let the gate's `gateAnnualEstimator` see it
   (ANN-14 wires the composition root's estimator; its adapter maps the estimate to
   `annualgolden.Result`).
2. Remove `pending_strategy` from its cases; fix line codes by hand if needed; lower its count in
   the baseline.
3. `go test ./internal/taxengine/...` must pass to the centavo.

## Generated cases (follow-up)

MODULE-annual §3 also asks for generated cases: each deduction cap at its boundary (±1 centavo),
every tuition level and cash exclusion per category — these run today as table tests of the
calculator (`domain/annual/deductions_test.go`, ANN-07) — and the row edges of
`isr.annual_152`, which need that table in the parameter file (ANN-04). The tariff-edge
generator joins this gate with ANN-04/ANN-08, reading the table independently of the engine as
the monthly `golden.TariffEdges` does.
