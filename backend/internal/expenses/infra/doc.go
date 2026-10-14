// Package infra holds the expenses module's adapters behind the use case
// ports:
//
//   - Store: the expenses table (GORM over the unit of work in ctx, as
//     cc_app under RLS: another user's rows are invisible). A taken
//     primary key (23505 expenses_pkey) is usecase.ErrIDConflict; the
//     category foreign key (23503) is usecase.ErrCategoryNotFound. LockID
//     takes a transaction-scoped advisory lock per client id, so a double
//     submit waits for the first create and then finds it.
//   - CategoryCatalog: the global expense_categories and
//     category_regime_rules tables (M4-02), read-only for cc_app.
//
// Stores never begin, commit or roll back a transaction and never set
// session state: the use case's unit of work does (P0-30 D-3).
package infra
