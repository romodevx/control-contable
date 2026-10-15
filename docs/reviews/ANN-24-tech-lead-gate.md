# ANN-24: Tech-lead gate of the annual module

- **Date:** 2026-10-15
- **Branch:** `feature/annual-qa-fixes` (PR #153 → `dev`): the ANN-23 QA (integration suites, the
  lone-missing-input semantics fix, the QA report), on top of all of AN already on `dev`: migrations
  #125, deductions #127, params #131, contract + ADR-0009 #134, PF strategies #133, PM strategies #135,
  mobile #140/#144, input use cases #141, `EstimateAnnual` + routes #145, property suite #148, flag
  #150. `origin/dev` merged in (3 commits behind).
- **Scope:** A) PR #153. B) The PLAN ANN-24 criteria over the module: OCP per section (registry),
  kernel purity, no floats, determinism, ports only, the estimate always labelled an estimate.
- **Not covered here:** the security gate (ANN-25) runs at the same time on its own branch.

## Verdict: **PASS** (after one OCP fix on this branch)

One gap against the criteria, fixed test-first here: the assembler branched on section kinds. The
rest holds. PR #153: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `go list` imports of `domain/annual/...` | stdlib, `kernel/fiscal`, `platform/money`, `taxengine/domain[/taxparams]`; `time` only for `time.Month`. archtest (`domain-imports`, `inner-transitive-framework`) green |
| `usecase/annual` imports | ports + `platform/{clock,tx,apperr,money}`, `uuid`, `x/text/norm`; no GORM, no `net/http` |
| `grep float\|time.Now\|rand.` over domain, use case, adapter, infra, `cmd/api/annual.go` (non-test) | none; the clock is the injected `platform/clock` |
| Map iteration in output order | two `range` over maps (`explain.Disclaimers`, `deductions.ownCaps`), both order-free (sorted after / per-key assignment); fingerprint is `encoding/json` |
| Estimate labelled | `annual_estimate` is always present and first (`explain.Disclaimers`), contract §AnnualEstimate states it; the app shows `EstimateLegend` |
| New guard test before the fix | **red**: 6 kind references in `estimate_assemble.go` |
| `go test ./internal/taxengine/... ./cmd/...`, `golangci-lint` | green, 0 issues; goldens and the annual property suite unchanged |

## A) PR #153

| Item | Assessment |
|---|---|
| `annual_qa_integration_test.go` | Good. Confirm-all keeps the figure and moves only payment sources, PM through the real stack, read-only and archived on every write (ADR-0009: delete and Undo stay open), the deduction bound on the real store. |
| Lone missing-input row | Right fix: `Semantics(container: true)` on `AnnualLinkRow`. The test anchors on the exact label and asserts the status text has no tap action and does not contain the CTA label. |
| QA report | Honest: own local stack, archived taxpayer via the API only, iOS not run (OB6). |

## B) Module criteria

| Criterion | Result |
|---|---|
| OCP per section | **Fixed here.** `estimate_assemble.go` named `pf_resico_check` (headline, explanation), `pf_art152` (informative income), `pm_601` (CU note) and `pm_601`/`pm_resico` (obligation reason, line regimes). A new strategy would have compiled and silently taken the PF reasons and no line regimes. Those facts are now fields of `ComputedSection` set by each strategy (`OutsideHeadline`, `Explanation`, `Notes`, `ObligationReasons`, `LineRegimes`, `DecidesInformativeIncome`). Guard: `registry_ocp_test.go` fails if any non-strategy file in the package names a computed kind. |
| Kernel purity | Pass. Strategies are pure (`Compute(Year, *Params)`); I/O is behind the use-case ports. |
| No floats | Pass. Cents and `money.Rate`; `SuggestedCU` is a fixed-point rate. |
| Determinism | Pass (see Evidence). |
| Ports only | Pass (see Evidence). |
| Labelled as estimate | Pass. |

## Follow-ups (non-blocking)

1. **Deadline per kind lives in `explain.sectionDeadlines`.** A new kind without a row gets no deadline
   silently. Either the strategy returns its deadline kind, or a test that every `computedKinds`
   entry has one.
2. **`"601"` literals.** `pmSection` and the accrual switch in `regimeYearOf` use `"601"`; add
   `fiscal.RegimeGeneralPM` beside the other regime constants.
3. **Spec vs ADR-0009 on read-only writes** (QA §6): MODULE-annual US-A13/§5 still say 402 for every
   write. Doc fix in ANN-26.
4. **Filed-payments prefill vs the monthly screen** (QA §6): the annual chains prior payments, the
   monthly screen does not; add a hint or a guide line (ANN-26).
