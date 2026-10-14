// Package fiscal is the shared kernel of Mexican fiscal concepts (ADR-0007):
// the pure value types and rules that incomes, expenses and the tax engine
// must agree on, so that one record means the same thing to all of them.
//
//   - Date: a calendar day with no time zone (ADR-0002 §4), MinRecordDate.
//   - CFDIUUID: a normalized folio fiscal, never logged.
//   - PersonType, RegimeCode, IVARate: the contract enums every module reads.
//   - RateCatalog and RateKey: the dated-parameter port (ADR-0002 §3) and the
//     parameter names more than one bounded context uses.
//   - ExpectedIVA, CheckIVA and IVATolerance: the IVA treatment rule
//     (MODULE-incomes §2.5, shared by MODULE-expenses).
//   - Share, SuggestAtRate, Suggestion: the withholding kinds of the law (a
//     dated rate of a base, or a legal fraction of the IVA) and how a
//     suggestion is computed and compared.
//   - Error, Violation, Violations: the domain error model whose values are
//     the contract details[].code. Modules declare their own codes with
//     ValidationError and BusinessRuleError.
//
// Module-specific rules (withholding applicability, deductibility, cash
// payments, income basis) stay in each module's domain. The kernel imports
// only the standard library and internal/platform/money (archtest
// kernel-imports). Money is money.Cents, every product rounds once, half
// away from zero; there is no float and no clock.
package fiscal
