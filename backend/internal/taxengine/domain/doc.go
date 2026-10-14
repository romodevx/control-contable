// Package domain is the tax engine's pure core (ADR-0002 §1, MODULE-tax-engine
// §2): the Strategy port keyed by StrategyKey (regime code, person type), the
// Input a strategy reads (period, options, the year's MonthSums, the
// taxpayer's TaxInputs), the RegimeEstimate it returns, and the Ledger every
// strategy computes through, so the breakdown lines are the computation trace
// and cannot drift from the math.
//
//   - Period: a "YYYY-MM" tax period and its last day, the date period-level
//     parameters are resolved on.
//   - MonthSums / YearSums: the ledger aggregates of one regime per month
//     (income by collection and invoice date, IVA, withholdings; expenses
//     paid), summed year-to-date or by bimester.
//   - TaxInputs: the user's per-fiscal-year inputs, one amount per month for
//     money_monthly keys (ADR-0002 Amendment 1).
//   - Ledger: input, add, subtract, multiply, lookup, result and info steps,
//     each with its citation and source. A result rounds once, half away from
//     zero, and adds an explicit rounding_adjustment line whenever the
//     displayed lines would not add up to it.
//   - RegimeEstimate: lines, ISR TaxResult, IVA components, withholdings to
//     remit, missing inputs, CPA topics and disclaimers; Status applies the
//     precedence informational > incomplete > partial > complete.
//   - Registry: the immutable StrategyKey -> Strategy map built at the
//     composition root (duplicates fail at startup), falling back to the
//     InformationalStrategy; its Estimate checks the period's RequiredInputs
//     against the TaxInputs first, so a missing input makes it incomplete.
//
// The package is pure: no I/O, no clock, no randomness, no floats (money is
// platform/money's Cents, Exact and Rate), no map-order dependence. It reads
// the parameters through domain/taxparams and the shared fiscal kernel
// (ADR-0007) for regime codes, person types and dates.
package domain
