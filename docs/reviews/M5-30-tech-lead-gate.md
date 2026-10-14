# M5-30: Tech-lead gate of module M5 Tax Engine

- **Date:** 2026-10-14
- **Branch:** `feature/tax-engine-qa-fixes` (PR #105 → `dev`): M5-28 flag tests, the M5-29 QA report and
  its one fix, on top of all of M5 (M5-01 … M5-27).
- **Scope:** A) PR #105. B) Holistic review of `backend/internal/taxengine/**` (domain, `taxparams`,
  strategies, registry, IVA components, estimate use case, readers, handlers, golden and property
  suites) and `mobile/lib/features/tax_engine/**`, against the M5 checklist: Open/Closed, purity,
  no floats, determinism, `EngineVersion` discipline, meaningful golden/property suites. C) The QA
  note "the 601 IVA golden was not reproduced" ([M5-29 §3](M5-29-qa.md#3-golden-figures-on-screen-vs-expected)).
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language
  (idiomatic Go wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gate (M5-31) runs separately. The regime start-period
  backdating fix (M5-29 D2) is on another branch and is not duplicated here.

## Verdict: **PASS**

No blocker is open. Two guard gaps were found and closed on this branch: a High (the no-float rule
failed open) and a Medium (`EngineVersion` did not follow figure changes inside a strategy). The
601 IVA note is a data-entry gap in the QA walk, not an engine bug: the app can capture every input
of PM-601-IVA-01, and a new end-to-end test shows it computes the golden figures from them.

PR #105: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `make -C backend check` (gofmt, golangci-lint, `go test -race ./...`) | green, 0 lint issues |
| `go test -tags=integration -run 'TestGolden_pm601\|TestPresetSmoke_taxEngine\|TestContract_estimates\|TestEstimateMonth' ./cmd/api/` | green |
| `make mobile-check` (format, analyze, tests + coverage gate) | 1,032 passed, 29 skipped |
| Float probe (domain file with `x := 0.5`, `strconv.ParseFloat`, `math.Round`, `const 1.0/3`) | `golangci-lint`: **0 issues** (forbidigo fails open); new guard: fails, naming the file |
| `TestNoFloatingPointInMoneyCode` on the shipped code | green: no float or complex expression in `taxengine/**`, `platform/money`, `kernel/**` |
| `TestEngineVersionPinsTheGoldenFigures` | green 5 of 5 runs (`-count=5`): the digest is deterministic |
| PM-601-IVA-01 through the API | green; without the collection date it reproduces the QA's figures (IVA charged 0, net −200,000.00) |
| Mobile 601 collection date | green; with `collectedOn` dropped from `toCreate` it fails (`Expected '2026-10-12', Actual null`) |

## A) PR #105

| Item | Assessment |
|---|---|
| `adapter/http/gen/tax_engine_module_test.go`: the 4 operations written by hand by operationId and ServeMux pattern, and the flag owns nothing else | Good. Same pattern as M2/M3/M4. The generated map has to agree with the spec, not only with itself. |
| `cmd/api/taxengine_preset_smoke_integration_test.go`: each seeded preset plus two fixtures (no engine; engine without `module_taxpayers`), all 4 operations with a real user and taxpayer, a guard that every owned operation is probed | Good. The expectation comes from the spec (engine **and** prerequisite), not from `flags.Prerequisites`. The report shows it can fail (prerequisite removed → red). |
| D1 fix (`home_screen.dart`): the placeholder moved inside the tax engine `FlagGate` | Correct, and test first (`46abfa6` before `0039438`). A module that is OFF is invisible (rules/module-workflow). |
| `tax_engine_flag_test.dart` | Nit, fixed in `27b8831`: `find.textContaining('estimado')` was case-sensitive, so a future "Estimado…" on home would pass. It is now `RegExp('estimad', caseSensitive: false)`. |
| QA report M5-29 | Thorough: 60 of 60 breakdown lines and 10 of 10 totals against the goldens, workarounds stated, environment problems kept apart, logs grepped for PII. One correction (§C): the 601 collection date can be entered in the app. A note now says so in the report. |
| TDD evidence | The M5-28 tests pin behavior that already existed, so they passed on the first run. The report says so and shows a mutation that turns them red. Acceptable for characterization tests. |

## B) Holistic findings

### Fixed on this branch

| # | Sev | Finding | Fix |
|---|---|---|---|
| G1 | **High** | **The "no floats" rule failed open.** `.golangci` forbidigo forbids `^float(32\|64)$` with `analyze-types`, which matches those identifiers only. A probe in `internal/taxengine/domain` with an inferred `x := 0.5`, `strconv.ParseFloat`, `math.Round` and an untyped `const 1.0 / 3` got **0 issues**. The archtest checks imports, not types. The shipped code was clean (grep and the new guard agree), but nothing stopped the next float. *ADR-0001 O7 / ADR-0002: money is int64 centavos and fixed-point `Rate`. A guard that cannot fail is not a guard (lint rules fail open; mutate the fixture).* | `ce9365e`: `TestNoFloatingPointInMoneyCode` type-checks (`go/packages`) the production code of `taxengine/...`, `platform/money/...` and `kernel/...`, and fails on any expression of float or complex type, including tuples such as `ParseFloat`'s. `TestNoFloatGuardSeesEveryFloatForm` runs it on `testdata/floatprobe` and requires every form to be reported. `golang.org/x/tools` moves from indirect to direct (test only). forbidigo stays as the fast first line. |
| G2 | **Medium** | **`EngineVersion` did not follow figure changes inside a strategy.** The records fingerprint hashes the inputs and `EngineVersion`, not the figures. `TestEngineVersionPinsTheStrategyList` ties the version to the *list* of strategies only. A strategy fix that changes a figure, with the golden updated by hand and no version bump, keeps every fingerprint the same, so M8 would never mark the month "changed since viewed". 1.0.1 itself was this kind of change (626 PM remitting its ISR withheld, same list). | `f85ef97`: `TestEngineVersionPinsTheGoldenFigures` hashes what the engine computes for every golden input (authored and generated; skipped ones too, because users still get those figures): status, coverage, ISR, IVA components, withholdings to remit, missing inputs, CPA topics, disclaimers, and every line's sequence, tax, code, operation, amount, rate, row, source and CPA flag. The digest is pinned per `EngineVersion`. Wording (labels, arguments, citations) is not hashed. |

### Checklist results (no action needed)

| Area | Result |
|---|---|
| **Open/Closed** | Adding a computation is one `Strategy` file plus one line in `taxengine.Strategies()`. `domain.Registry` is immutable: it is built once at the composition root, has no `Register` method and no package global, and refuses duplicates and blank keys at startup. Regimes without a strategy fall back to `InformationalStrategy`. The golden gate runs the composition root's own list, so a strategy is either wired and gated or neither. Shared IVA (`ComputeIVA`), totals and consolidation never switch on a regime code. The rest of what a new regime touches is data and presentation: its `params` YAML values, its goldens and the ratchet baseline, its Spanish concepts in `adapter/http/texts.go` (`conceptOverrides` is keyed by `"<regime>/<label>"`), and both `EngineVersion` pins. A coverage test fails on a label with no concept, so none of these can be forgotten silently. To be listed in the M5-32 add-a-regime guide (FU3). |
| **Purity** | `domain/**` imports the standard library, `platform/money` and `kernel/fiscal` only (depguard plus archtest, transitively). The use case reaches incomes, expenses, taxpayers, tax inputs and parameters through consumer-side ports (`usecase/ports.go`, `estimate_ports.go`, `inputs_ports.go`). GORM and SQL live in `infra/taxinputs` and in the other modules' ledger readers, wired in `cmd/api`. Handlers translate and present only. |
| **No floats** | Backend: none (G1 guard). Mobile: money is integer centavos (`formatMoney`, `spokenMoney` use `~/` and `%`). Rates stay strings end to end (`formatRatePercent` moves the decimal point on the string; `parseDecimalRate` canonicalizes text). The only `double`s in the feature are layout constants (icon sizes, rule widths, alignment). |
| **Determinism** | forbidigo forbids `time.Now`, and the estimator takes an injected `clock` (tests pin 2026-10-20). No randomness in production code. Every map iteration whose order reaches output is sorted afterwards (`EstimateDisclaimers` sorts codes and topics; `Fingerprint` sorts regimes, options and input slots). Property tests use a fixed PCG seed per case index, and one property asserts byte-identical JSON for identical inputs. The mobile feature never reads `DateTime.now` and has no regime-code switches: all text comes from the server. |
| **Golden suite** | 78 hand-authored files across 21 regime/person directories plus boundary-generated tariff-edge cases. Every research worked example has a case and every research case cites an existing example. There is no update flag: a rewriter scan fails the build, and every file keeps its hand-authored header. A skip/pending ratchet fails on growth **and** on shrink (2 skips, both ⚠️ CPA topics for M5-33; 0 pending). The reference implementation is itself checked against the research's arithmetic before it judges the engine. |
| **Property suite** | Lines reconcile to their totals; ISR is monotonic in income; withholdings to remit never enter the taxpayer's own taxes; IVA never carries forward; missing required inputs always mean `incomplete`; ISR is right at every tariff row edge; a PM's withholdings to remit are complete. The generator coverage test makes every strategy get cases. `TestRegression626PMRemitsTheISRItWithheld` keeps the M5-24 bug as a minimal case. These properties are meaningful, not tautologies. |

## C) QA note: "the 601 IVA golden was not reproduced"

Investigated with the systematic-debugging method (root cause before any fix).

**Hypothesis 1:** the engine or a reader drops 601 IVA. **Rejected.** The incomes ledger
(`incomes/infra/ledger.go`) books 601 ISR by `invoiced_on` and IVA by `collected_on` (LIVA 1-B), as
MODULE-incomes §2.4 specifies. The expenses ledger carries `creditable_iva_cents` and
`iva_withheld_cents` into `ExpenseMonthSums.IVAWithheldToRemit`. `GeneralPMStrategy` passes both to
`ComputeIVA`.

**Hypothesis 2:** the app cannot express the golden's inputs. **Rejected.** The golden needs, for
June 2026: income 2,000,000.00 + IVA 320,000.00 *invoiced and collected*; creditable IVA paid
200,000.00; IVA withheld from a PF lawyer to remit, 5,333.33 (fees 50,000.00 + IVA 8,000.00 × ⅔);
coeficiente 0.15 and zero prior payments. Each of these maps to a field the app has:

| Golden input | Where the app captures it |
|---|---|
| Collected in June (accrual 601) | Income form → *Más detalles* → **Fecha de cobro** (collapsed; for 601 the visible date is *Fecha de factura*). An income saved without it is *Por cobrar* and can be edited later (MODULE-incomes US "a 601 income gets `collected_on` set") |
| IVA withheld to remit 5,333.33 | Expense, category *Honorarios y servicios profesionales*: the form prefills ISR 10 % (5,000.00) and IVA ⅔ (5,333.33) for a persona moral |
| Creditable IVA 200,000.00 | The lawyer's 8,000.00 plus 192,000.00 from other expenses with CFDI (e.g. *Mercancía* 1,200,000.00 + IVA) |
| Coeficiente, prior payments | *Datos fiscales* of 601 |

**Root cause:** in the QA walk the 601 incomes were entered with the invoice date only (they came
from the ISR golden, which is accrual) and stayed *Por cobrar*, so IVA charged was 0. The collection
date lives in the collapsed *Más detalles* section, and the walk did not open it. That is a UX
finding (FU1), not a calculation bug.

**Proof:**
- `6168d72` `cmd/api/taxengine_golden_api_integration_test.go` enters PM-601-IVA-01 through the API
  operations and fields the app sends (no SQL) and gets IVA charged 320,000.00, creditable
  200,000.00, net **120,000.00**, and the breakdown line *IVA que retuviste a tus proveedores (a
  enterar)* **5,333.33**, all equal to the golden. Without `collected_on` the same test shows the QA's
  figures.
- `d20ed96` `income_form_test.dart`: a 601 income sends the *Fecha de cobro* typed under *Más
  detalles* as `collected_on`.

**One difference by design:** the regime's withholdings to remit are **10,333.33**, not the golden's
5,333.33. The golden isolates the IVA (`taxes: [iva, withholding]` with only the IVA withheld). In the
app, the same payment also carries the 10 % ISR a persona moral must withhold from a PF's fees
(LISR 106), which the form prefills. If the user zeroes that ISR, the app shows exactly 5,333.33.
This is not a defect. It is noted for the CPA review (FU2).

## Follow-ups

None blocks M5's merge. Added to [PLAN §M5 follow-ups](../project-specs/PLAN-control-contable.md#m5-follow-ups).

| ID | Item | Owner | Target |
|---|---|---|---|
| FU1 | 601 (and every accrual regime): *Fecha de cobro* decides the IVA but sits in the collapsed *Más detalles* section. Show it in the visible group for accrual regimes, or ask "¿Ya te pagaron?" with the date. Together with M5-29 D6 (a hint on the 601 card when income is invoiced but not collected). The ui-forms rule collapses only what the 80 % case does not need, and here the 80 % case needs it | pm + frontend-sr | M3 hardening, before `module_incomes` is enabled for PM users |
| FU2 | PM-601-IVA-01 asserts an IVA-only "withheld to remit" (5,333.33). Record in [CPA-REVIEW](CPA-REVIEW.md) that the app adds the ISR withheld from PF fees (LISR 106), so a real 601 sees 10,333.33 | qa-sr | M5-33 |
| FU3 | Add-a-regime guide: strategy file, `Strategies()` line, `params` values, goldens and ratchet baseline, `texts.go` concepts, `EngineVersion` bump with both pins (strategy list and golden figures digest) | doc-writer | M5-32 |
| — | M5-29 D2 (regime start month) | — | In flight on its own branch, not duplicated |
