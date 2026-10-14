// Package usecase holds the expenses module's use cases (M4-05, M4-06;
// MODULE-expenses, with the rules it shares from MODULE-incomes):
//
//   - Create and Update: the idempotent client id (same request ⇒ the
//     original, other content ⇒ 409 id_reused), the read-only check,
//     TaxpayerRegimeChecker (regime held on paid_on, taxpayer not archived),
//     the category rule of the taxpayer's regime and person type, the
//     domain rules on the dated parameters (Assess), the duplicate-folio
//     and provisional-parameter warnings, and on create only
//     AccessPolicy.ReserveRecord, the last step before the insert.
//   - Get, List (keyset pages with period totals), Delete (soft, allowed
//     read-only) and Restore (DECIDED: never reserves a slot, allowed
//     read-only, re-checks the regime window).
//   - Categories: the catalog resolved for a regime the taxpayer holds on a
//     date and its options, with the IVA rates and cash limit of that date
//     and the mandatory legend (listExpenseCategories).
//
// Every client-facing failure is an *apperr.Error (contract code, details
// with the domain's details[].code); anything else is an infrastructure
// failure (500). Nothing logged or returned carries an amount, a
// description or a folio fiscal: the audit events (expense.created,
// .updated, .deleted, .restored) name the ids only.
//
// The use cases depend on ports only: the unit of work (tx.Manager), the
// access policy, the taxpayers module's checker and profile reader (this
// module's own copies, wired by cmd/api), the stores in infra and the tax
// engine's RateCatalog behind Parameters.
package usecase
