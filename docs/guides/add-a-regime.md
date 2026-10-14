# Guide: add a regime to the tax engine

For a backend engineer who has to make the engine compute a regime it only explains today,
or change how it computes one it already does. It lists every place the change touches;
each step has a test that fails if the step is forgotten.

| | |
|---|---|
| Design | [ADR-0002 §1–§5](../adr/0002-tax-engine.md) (Strategy port, coverage, parameters, rounding, goldens) |
| Rules | [MODULE-tax-engine §2](../project-specs/MODULE-tax-engine.md#2-domain-rules) (statuses, lines, IVA, rounding, inputs, CPA topics) |
| Module page | [Tax engine](../modules/tax-engine.md) |
| Origin | M5-30 tech-lead gate FU3 (the list of touch points) |

**Open/Closed is the point.** A regime is one strategy file, one line in `Strategies()`, its
data (parameters, catalog row, goldens) and its Spanish texts. Nothing in the use case,
the IVA netting, the totals, the handlers or the app switches on a regime code. If your
change needs one of those to know about your regime, stop and take it to the architect: it
needs an ADR.

## Before you start

1. **A source for every figure.** The rules come from the law (LISR, LIVA, LIF, RMF, Anexo
   8) and the expected figures from a worked example, the research
   ([`tax-regimes-mx-2026.md`](../research/tax-regimes-mx-2026.md)) or a CPA. No worked example,
   no strategy: write the example first (research addendum or CPA note) and get it reviewed.
2. **The key.** Strategies are keyed by `(regime code, person type)`. 626 is two strategies
   (RESICO PF and RESICO PM).
3. **What the records cannot tell.** List the tax inputs the regime needs (prior payments,
   coeficiente, prior-year income, ...) and in which periods each is required. Required
   inputs depend on the period, not only on the regime.
4. **Unverified rules.** Every rule you cannot confirm against a primary source becomes a CPA
   topic (step 8). It still runs; the user sees the strong warning.
5. A branch from `dev` (`feature/tax-engine-<code>-<pf|pm>`) and TDD: the goldens and the
   strategy's unit tests fail before the strategy exists.

## Steps

### 1. Goldens first: hand-authored, from the official example

Write one file per worked example under
`backend/internal/taxengine/testdata/golden/<code>-<pf|pm>/<ID>.yaml`, in the format of the
[golden README](../../backend/internal/taxengine/testdata/golden/README.md):

- line 1 is the header `# golden: hand-authored from <source>. Never generated or rewritten by a tool.`;
- `id` like `PF-612-ISR-01`, `source` is `research:<example id>`, `cpa:<ref>` or `spec:<ref>`;
- amounts are **quoted** decimals with two places (`"150.96"`); an unquoted number is refused;
- `expected` holds the source's figures **to the centavo**: status, the taxes asserted,
  every line in order (tax, code, op, amount), missing inputs, CPA topics, disclaimers;
- a case whose figure depends on an unverified rule gets `skip_reason: "⚠️ ..."`;
- while the strategy does not exist, mark each case `pending_strategy: true`.

**Never regenerate a golden.** There is no update flag: `TestGoldensAreNeverRewritten` fails
the build if any code can write under `internal/taxengine/`, and the loader refuses a file
without the header. If the engine disagrees with the example, the engine is wrong until a
cited source says otherwise.

Add generated boundary cases when the regime has a tariff: the gate
(`golden_gate_test.go`) builds `lower`, `upper` and `upper + 0.01` of every row from the
parameter file and checks them against `reference_test.go` (a linear scan in `math/big`,
never the engine).

### 2. The strategy file

`backend/internal/taxengine/domain/strategy_<code>_<pf|pm>.go`, with its unit tests next to
it. It implements `domain.Strategy`:

```go
type Strategy interface {
	Key() StrategyKey
	Treatment() Treatment                                     // full or partial
	RequiredInputs(o Options, p Period, f Facts) []RequiredInput
	Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error)
}
```

- **Pure.** No `time.Now`, no randomness, no I/O, no maps whose order reaches the output, no
  floats. Money is `money.Cents` or `Exact`, rates `money.Rate`. `depguard`, the archtest and
  `TestNoFloatingPointInMoneyCode` fail otherwise.
- **Compute only through the `Ledger`** (`Input`, `Add`, `Subtract`, `AddProduct`, `Scale`,
  `Lookup`, `Result`, `Info`). Every amount is a line, each with its legal basis and its `source`
  (`records`, `tax_input:<key>`, `params:<key>`, `computed`). Every line except `info` and
  `rounding_adjustment` cites at least one article.
- **IVA** goes through the shared `ComputeIVA` components; do not net IVA yourself.
  Withholdings the taxpayer made go to `withheld_to_remit`, never into its own taxes.
- **Parameters by date**: `p.Rate(key, date)`, `p.Table(key, date)`,
  `p.MonthlyTable(key, month, date)`; period-level values use the last day of the period.
  A value not in force is an error (`ErrNotInForce`), never a default.
- **Missing inputs count as zero**; `RequiredInputs` decides `incomplete`, not the strategy.
- Never round yourself. `Ledger.Result` rounds the running value once, to the centavo, and
  appends the `Ajuste por redondeo` line when the displayed lines do not add up
  (ADR-0002 §4).

Copy the shape of the closest existing strategy: `strategy_612_pf.go` (year to date, monthly
accumulated tariff), `rent_pf.go` (monthly, options), `resico_pf.go` (flat brackets),
`strategy_601_pm.go` (accrual, coeficiente), `partial_pm.go` (partial, IVA only).

### 3. Register it: one line

Add it to `Strategies()` in `backend/internal/taxengine/strategies.go`. That list is the
composition root's registry (`cmd/api/taxengine.go`) **and** the golden gate's, so a
strategy is wired and gated, or neither. `domain.NewRegistry` refuses a duplicate key at
boot. Add the key to `launchStrategyKeys` in `strategies_test.go`.

Then remove `pending_strategy` from the regime's goldens and run the gate:

```sh
cd backend
make golden                                  # every case, skip reason and pending count
go test ./internal/taxengine/...             # the gate runs in `go test ./...` too
```

Every non-skipped case must match to the centavo. Line codes are yours to name in this PR
(rename them by hand in the goldens); the `expected` totals never change without a source.

### 4. Lower the baseline ratchet

`backend/internal/taxengine/testdata/golden_baseline.yaml` lists the allowed skipped ids and
the pending count per strategy key. It is edited **by hand, only downwards**: remove the
regime's pending count, and list a skipped id only when it carries a `skip_reason`.
`TestGoldenCases` fails when a run skips an id not listed, when it has more pending cases
than allowed, **and** when it is below the baseline (so this PR must lower it).

### 5. Parameters

If the strategy reads a key the parameter file does not have (a tariff, a rate, a cap), add
it to `backend/internal/taxengine/params/data/<year>/v<N>.yaml` **verbatim** from the official
publication, citing a `sources[]` entry, with its validity dates. Never derive a published
table. A change to an existing year is a new version (`v<N+1>.yaml`), not an edit; follow
the [parameter update runbook](../runbooks/tax-params-update.md) §3–§5, including the
independent double entry.

Add the key to `TestEmbedded<year>DefinesEveryKeyTheModulesRead` in
`params/embedded_<year>_test.go`. That test is the only check that a key a strategy reads
exists: the boot validates the file's shape but not each strategy's keys (M5-31 G-5).

### 6. Catalog row (forward migration)

The use case trusts the catalog: a regime whose `regimes.treatment` is `informational`
gets the informational estimate **even if a strategy is registered**. Write a new goose
migration in `backend/migrations/` (never edit an applied one; see
`20261014160000_regime_606_pays_quarterly.sql` for the shape) that sets, for the
`(code, person_type)` row:

- `treatment` (`full` or `partial`), matching `Strategy.Treatment()`;
- `required_inputs` (the keys, kinds and labels of step "Before you start" 3);
- `option_defs` if the regime has options;
- `explanation` for a `partial` regime (the `regimes_explanation_when_not_full` constraint
  requires it).

`regimes` is global configuration (SELECT-only for `cc_app`), so the migration changes no
policy. Run `make -C backend test-integration` (catalog and tax-input migration tests).

### 7. Spanish texts

Line concepts, disclaimer texts and topic phrases are rendered by
`backend/internal/taxengine/adapter/http/texts.go`, never by the strategy or the app:

- a new label key goes in `concepts`; a regime's own wording of a shared step goes in
  `conceptOverrides`, keyed `"<code>/<label>"`;
- `TestEveryBreakdownLineHasAConcept` runs every golden through the registry and fails on a
  label with no concept.

### 8. CPA topics and CPA-REVIEW

For every rule you could not confirm:

- add an item to [CPA-REVIEW.md](../reviews/CPA-REVIEW.md) under the right section, with the
  sources it points to (research section, ADR, catalog row, PR);
- if the rule must raise the strong banner, add a `CpaTopic` (append-only) to the contract
  (architect, `make contract-lint`, `make -C backend generate`, `make mobile-client`), to
  `domain/enums.go`, and its phrase to `topicPhrases` in `texts.go`; set it on the regime and
  `needs_cpa_review` on the affected lines;
- keep the dependent goldens skipped with the reason; M5-33-style sign-off lifts them.

### 9. Bump `EngineVersion`, both pins in the same commit

A new strategy changes figures for the same records, so the records fingerprint must move or
M8 will never mark those months "changed since viewed". In one commit:

1. bump `EngineVersion` in `backend/internal/taxengine/domain/engine.go` (minor for a new
   regime, patch for a fix to an existing one);
2. pin the strategy list: add the new version to the `pinned` map of
   `TestEngineVersionPinsTheStrategyList` (`strategies_test.go`);
3. pin the figures: add the new version and its digest to `goldenFiguresDigest`
   (`engine_version_test.go`). Run the test once; the failure prints the actual digest:

   ```sh
   cd backend && go test -count=1 -run TestEngineVersionPinsTheGoldenFigures ./internal/taxengine/
   ```

4. add a row to the [engine version log](../adr/0002-tax-engine.md#engine-version-log)
   (version, date, change, PR).

A fix inside an existing strategy that changes any figure, status or line (with no change
to the list) needs steps 1, 3 and 4: the digest catches it even when the goldens were
edited by hand.

### 10. Docs

In the same PR:

- [MODULE-tax-engine §2.4](../project-specs/MODULE-tax-engine.md#24-regime-coverage-and-algorithms)
  (coverage row and algorithm), §2.5 (citations), §2.8 (inputs), §2.9 (topics), §3 (goldens);
- the coverage table of the [module page](../modules/tax-engine.md#regime-coverage);
- ADR-0002 §2 if the coverage at launch changes, and the engine version log (step 9).

## Done when

| Check | Command |
|---|---|
| Format, lint (incl. forbidigo, depguard), unit tests, archtest, golden gate, both version pins, concept coverage, no-float guard | `make -C backend check` |
| Catalog migration and tax-input tests, golden figures through the API | `make -C backend test-integration` |
| Contract unchanged or linted, error tables in sync | `make contract-lint && make api-error-tables-check` |
| Every skip has a reason and the baseline went down | `make -C backend golden` |

Then the PR goes to tech-lead review (Open/Closed, purity, no floats, determinism,
`EngineVersion`), and to the CPA for any new topic.
