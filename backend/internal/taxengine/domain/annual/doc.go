// Package annual is the pure core of the annual estimate (MODULE-annual
// §2, PLAN ANN-07): the spec's section, status, obligation, balance, CPA
// topic and disclaimer enums, and the personal-deductions calculator of
// LISR art. 151 and the colegiaturas decree (DOF 26-12-2013 arts. 1.8–1.10).
//
// Placement (ADR-0009): the annual estimate is a sub-context of the tax
// engine module; this is its pure domain. It imports only the money type,
// the tax engine's domain (the Ledger steps its notes become) and the
// stdlib, and nothing in the monthly engine imports it. The legal limits (UMA, caps, tuition amounts) are read through
// the LimitsReader port; the 2026 parameter file implements it in ANN-04.
//
//   - ComputeDeductions: eligibility (cash is not deductible for the
//     categories the law restricts), each category's own cap, the global
//     cap min(5 UMA anuales, 15 % of total income) on the capped group, and
//     the uncapped categories (retirement after its own cap, disability
//     medical expenses, tuition). Every exclusion and every cap that binds is
//     a Note, rendered as an informational "Tope aplicado" breakdown line.
//
// Money is int64 centavos and fixed-point rates; there is no float and no
// clock in this package.
package annual
