// Package golden is the tax engine's golden harness (PLAN M5-07,
// MODULE-tax-engine §3, US-X3): hand-authored cases taken from the
// research worked examples (docs/research/tax-regimes-mx-2026.md §A.5,
// §B.5) or from a CPA, run through the registered Strategy of their
// regime and compared to the centavo.
//
//   - Case format: one YAML file per case under
//     internal/taxengine/testdata/golden/<regime>-<pf|pm>/<id>.yaml; see
//     Decode and testdata/golden/README.md. Every file starts with the
//     HandAuthoredHeader comment; a tool that re-serializes the file drops
//     it and the case no longer loads.
//   - Runner: Load reads every case, Evaluate runs one through the M5-10
//     domain.Registry and the parameter Catalog, Run makes one subtest per
//     case. A case with skip_reason (an unverified ⚠️ rule) is skipped with
//     its reason; a case whose strategy is not registered yet is
//     pending_strategy; an informational case runs through the fallback.
//     Both are reported, never silent (Summary).
//   - No update flag: expected values are never written by a program.
//     FindRewriters is the meta-check the gate runs over the source tree.
//   - Boundary generator: TariffEdges and ValidityEdges list every tariff
//     row edge and validity-date edge of a parameter file, read
//     independently of the engine's loader (RawParams). The expected
//     values come from a reference implementation in the gate's tests,
//     never from the engine.
package golden
