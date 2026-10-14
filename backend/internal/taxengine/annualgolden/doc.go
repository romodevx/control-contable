// Package annualgolden is the annual estimate's golden harness (PLAN
// ANN-06, MODULE-annual §3): hand-authored cases taken from the research
// worked examples AN-01..AN-08 (docs/research/annual-return-2026.md Part D)
// or from a CPA, run through the registered annual sections and compared to
// the centavo. It mirrors the monthly harness (package golden) and shares
// its rules: the hand-authored header, the strict YAML with quoted centavo
// strings, the error values, the skip/pending verdicts and the ratchet
// baseline.
//
//   - Case format: one YAML file per case under
//     internal/taxengine/testdata/annual_golden/<pf|pm>/<id>.yaml; see
//     Decode and testdata/annual_golden/README.md. A case is a taxpayer's
//     year (regime figures, payments, salary constancias, personal
//     deductions, annual inputs) and the sections the estimate must return
//     (lines, totals, balance, PM extras, informative difference).
//   - Runner: Load reads every case, Evaluate runs one through the
//     consumer-side Estimator and compares its Result, Run makes one subtest
//     per case. A skip_reason (an unverified ⚠️ rule) is reported skipped; a
//     case whose sections are not registered is pending_strategy; neither is
//     silent (Summary), and Summary.Ratchet holds both at the baseline.
//   - No update flag: nothing writes a golden. golden.FindRewriters, run by
//     the gate over the whole module, covers this package too.
package annualgolden
