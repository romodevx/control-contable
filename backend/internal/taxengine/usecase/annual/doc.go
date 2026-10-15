// Package annual is the application layer of the annual estimate's inputs
// (ADR-0009, MODULE-annual §2.6, PLAN ANN-13): personal deductions (create,
// edit, soft delete, restore), salary constancias, the per-year annual
// facts and the provisional payments the user filed, plus the read ports
// the annual estimate (ANN-14) consumes.
//
// Every use case runs in one unit of work as the caller (RLS as cc_app): a
// row of another user, or of another taxpayer of the caller, is 404, never
// 403. Writes need write access (402 for a read-only user, OQ-A9) except
// delete and restore, which stay open as for records. No amount, label or
// beneficiary is ever logged: the audit events carry ids only (§5, §9).
package annual
