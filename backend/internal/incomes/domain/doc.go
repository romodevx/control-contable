// Package domain holds the incomes module's pure rules (MODULE-incomes §2):
//
//   - Draft and Assess: an income as the user read it from the CFDI or the
//     bank statement, checked against the payee's regime, and the derived
//     facts the use case stores (booking date, "Por cobrar", the 625
//     activity) with its totals and non-blocking warnings.
//   - IncomeBasis.DateFor: which date each tax uses. ISR follows the
//     regime's basis (accrual → invoiced_on, cash → collected_on); IVA is
//     always cash (LIVA 1-B).
//   - ExpectedIVA and CheckIVA: the IVA treatment (16 %, border 8 % within its
//     validity, 0 %, exempt) and the ±IVATolerance consistency check.
//   - PolicyFor and WithholdingPolicy.Suggest: which withholdings a
//     counterparty may make for a regime and person type, and the suggested
//     amounts (ISR at a dated catalog rate, IVA as a legal share of the IVA).
//
// Every rate comes from the RateCatalog port (ADR-0002 §3), implemented by
// the tax engine (M5-08); only the IVA shares (2/3, 1/2, 1/1), which a Rate
// cannot hold exactly, are legal fractions in code (ADR-0002 §4). Money is
// money.Cents and every product rounds once, half away from zero. There is no
// float and no clock: "today" is a parameter.
//
// Errors: a client mistake is a Violations list of (field, Error) whose Error
// value is the contract details[].code; Class tells the adapter which
// ErrorCode it maps to (validation_failed or business_rule_violation). A
// parameter fault (ErrInvalidParameter) is a server error, never a Violation.
// No error or warning carries an amount, a description or a CFDI UUID.
//
// The package imports the standard library and internal/platform/money only.
package domain
