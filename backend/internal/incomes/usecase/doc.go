// Package usecase holds the incomes module's use cases (M3-04, M3-05,
// MODULE-incomes §2): create and update (Create, Update), read (Get, List
// with keyset pagination and period totals), soft delete and restore.
//
// Every operation runs in one unit of work (tx.Manager) as the caller, so
// RLS sees that user only, and the ports it calls from inside (the
// taxpayers' CheckRegime, the access policy) join that transaction: the
// regime's FOR SHARE locks and the trial slot commit or roll back with the
// income.
//
// Create follows the contract's order of checks: idempotency on the client
// id (a retry with the same fields replays the stored income, 200, before
// any access check; other fields, another taxpayer or a deleted income are
// 409 id_reused; another user's id is a bare 409) → read-only access (402)
// → archive and regime window (422, CheckRegime per distinct period) →
// domain rules (400/422, domain.Assess; a date before the first parameter
// set is 400 out_of_range) → trial slot (402, access.ReserveRecord) →
// insert. Update re-runs everything but the slot. Restore re-runs the
// archive and regime checks only: it never reserves a slot (OQ-10) and a
// read-only user may undo a delete (coordinator decision). Delete is
// allowed to every user; both refuse an archived taxpayer.
//
// Ports are declared here, consumer side (archtest: modules never import
// each other): Taxpayers (M2's TaxpayerRegimeChecker and profile reader,
// adapted in cmd/api), AccessPolicy (P0-16), RateCatalog (M5-08's adapter,
// with the provisional flag) and Store (infra). Client failures come back
// as *apperr.Error; anything else is an infrastructure failure (500).
// Descriptions, CFDI UUIDs and amounts are never logged (S-5); audit events
// carry ids only (S-6).
package usecase
