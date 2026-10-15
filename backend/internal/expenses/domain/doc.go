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
//   - Supplier withholdings (MODULE-withholdings §2.2, §4.2, ADR-0008): a
//     persona moral describes its supplier and the kernel's decision table
//     (fiscal.DeriveWithholdings) derives what it must withhold and remit;
//     a persona física never withholds and is never asked. The creditable
//     IVA splits into the share paid and the share withheld.
//
// The concepts expenses shares with incomes and the tax engine (Date, the
// CFDI UUID, the IVA treatment and its tolerance, the withholding shares,
// the rate port, the error model) come from the fiscal shared kernel,
// internal/kernel/fiscal (ADR-0007). This package's RateCatalog embeds the
// kernel's and adds the money parameters (the cash-payment limit); M5-08
// implements it. The category rule is catalog data (category_regime_rules)
// the use case reads and passes in. There is no float and no clock: "today"
// is a parameter.
//
// Errors: a client mistake is a fiscal.Violations list whose codes are the
// contract details[].code; this package adds its own (ErrUnknownCategory,
// ErrHasCFDIMismatch, ErrSupplierNotApplicable, ErrResidentialRentExempt). A
// parameter fault (fiscal.ErrInvalidParameter) is a
// server error, never a Violation. Non-deductibility is never an error: it
// is a warning plus deductible = false. No error or warning carries an
// amount, a description or a CFDI UUID.
//
// The package imports the standard library, internal/kernel/fiscal and
// internal/platform/money only.
package domain
