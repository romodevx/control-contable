// Package domain holds the incomes module's pure rules (MODULE-incomes §2):
//
//   - Draft and Assess: an income as the user read it from the CFDI or the
//     bank statement, checked against the payee's regime, and the derived
//     facts the use case stores (booking date, "Por cobrar", the 625
//     activity) with its totals and non-blocking warnings.
//   - IncomeBasis.DateFor: which date each tax uses. ISR follows the
//     regime's basis (accrual → invoiced_on, cash → collected_on); IVA is
//     always cash (LIVA 1-B).
//   - PolicyFor and WithholdingPolicy.Suggest: which withholdings a
//     counterparty may make for a regime and person type, and the suggested
//     amounts (ISR at a dated catalog rate, IVA as a legal share of the IVA).
//
// The concepts incomes shares with expenses and the tax engine (Date, the
// CFDI UUID, the IVA treatment and its ±IVATolerance check, the withholding
// shares, RateCatalog, the error model) come from the fiscal shared kernel,
// internal/kernel/fiscal (ADR-0007). Every rate comes from
// fiscal.RateCatalog (ADR-0002 §3), implemented by the tax engine (M5-08).
// There is no float and no clock: "today" is a parameter.
//
// Errors: a client mistake is a fiscal.Violations list whose codes are the
// contract details[].code; this package adds its own codes (ErrNotForRegime,
// ErrActivityNotHeld). A parameter fault (fiscal.ErrInvalidParameter) is a
// server error, never a Violation. No error or warning carries an amount, a
// description or a CFDI UUID.
//
// The package imports the standard library, internal/kernel/fiscal and
// internal/platform/money only.
package domain
