// Package infra holds the taxpayers module's adapters behind the use case
// ports:
//
//   - Store: the taxpayers and taxpayer_regimes tables (GORM over the unit
//     of work in ctx, as cc_app under RLS). Database refusals become use
//     case errors: 23505 (duplicate RFC, taxpayer id or regime), 23P01 (the
//     overlap trigger), 23503 (catalog and taxpayer foreign keys) and the
//     23514 checks the use cases validate first.
//   - Catalog: the global regimes and regime_compatibility tables (M2-02).
//   - RFCVault: AES-256-GCM sealing of the RFC bound to its row and owner,
//     and the per-user HMAC blind index (S-1, S-2).
//   - NoRecords: the record lookup until M3/M4 own incomes and expenses.
//
// Stores never begin, commit or roll back a transaction and never set
// session state: the use case's unit of work does (P0-30 D-3).
package infra
