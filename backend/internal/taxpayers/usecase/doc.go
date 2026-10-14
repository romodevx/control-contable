// Package usecase holds the taxpayers module's application logic
// (MODULE-taxpayers, M2-05 and M2-06):
//
//   - Taxpayers: create (onboarding, one transaction), list, get, update
//     (rename, RFC correction), archive, unarchive and delete.
//   - Regimes: add (or reopen) a regime, change its options, end it.
//   - Queries: the read-side ports other modules consume,
//     TaxpayerRegimeChecker (M3, M4) and TaxpayerProfileReader (M5).
//
// Every operation runs in one unit of work of the calling user (tx.Manager),
// so Row Level Security scopes each read and write: another user's taxpayer
// does not exist (404). Writes lock what they check first (Lock), so the
// check and the write see the same row.
//
// Order of checks (MODULE-taxpayers §2.10, contract createTaxpayer): the
// access policy (402 for a read-only user) runs before the cap and before
// any business rule; archive, end-regime and delete never ask it. An
// idempotent create retry answers before the policy: it writes nothing.
//
// The RFC is never in clear outside the request that carries it (S-1..S-4):
// the use case validates it (domain.RFC), generates the row id, then seals it
// bound to that id and the owner (RFCProtector); no read path selects or
// decrypts the ciphertext, since every response carries only rfc_masked.
// Logs and errors never contain it.
//
// Errors: every client-facing failure is an *apperr.Error carrying the
// contract code and FieldError details, wrapping a sentinel of this package
// or a domain error (match with errors.Is). Anything else is an
// infrastructure failure, wrapped with the operation's name (500).
package usecase
