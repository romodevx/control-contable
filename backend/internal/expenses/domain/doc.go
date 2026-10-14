// Package domain holds the expenses module's pure rules (MODULE-expenses §2):
//
//   - Draft and Assess: an expense as the user entered it, checked for the
//     paying taxpayer and its category rule, with the derived facts the use
//     case stores (effective deductible and its source), the creditable-IVA
//     candidate, totals and non-blocking warnings.
//   - ResolveDeductibility: the category rule's hint adjusted by the regime
//     options (606 blind 35 %, 625 definitive option, RESICO PF).
//   - The cash-payment rule (LISR 27 fr. III ⚠️ CPA): a cash payment above
//     the dated limit, or any cash payment of a category that requires an
//     electronic payment (fuel), is not deductible by default and warns.
//   - ExpectedIVA and CheckIVA: the IVA treatment and its tolerance, the same
//     rules as incomes (MODULE-incomes §2.5).
//   - MayWithhold and SuggestWithholdings: only a persona moral withholds
//     from its suppliers (LIVA 1-A); suggestions are UX only.
//
// Rates and the cash-payment limit come from the RateCatalog port (ADR-0002
// §3), implemented by the tax engine (M5-08). The category rule is
// catalog data (category_regime_rules) the use case reads and passes in.
// Money is money.Cents and every product rounds once, half away from zero.
// There is no float and no clock: "today" is a parameter.
//
// The date, IVA, CFDI and error types repeat the incomes module's on
// purpose: modules are vertical slices that never import each other, and
// internal/platform takes no business logic (P0-29 FU4).
//
// Errors: a client mistake is a Violations list of (field, Error) whose Error
// value is the contract details[].code; Class tells the adapter which
// ErrorCode it maps to. A parameter fault (ErrInvalidParameter) is a server
// error, never a Violation. Non-deductibility is never an error: it is a
// warning plus deductible = false. No error or warning carries an amount, a
// description or a CFDI UUID.
//
// The package imports the standard library and internal/platform/money only.
package domain
