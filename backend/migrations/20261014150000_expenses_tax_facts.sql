-- Expense tax facts (M4-08): the facts the domain assessed on the record's dated parameters,
-- stored with the record like the effective deductible already is, so the tax engine's
-- LedgerReader aggregate only sums them.
-- Sources: docs/project-specs/MODULE-expenses.md §2.4 (creditable-IVA candidate), §2.5
-- (cash payments over the limit), MODULE-tax-engine.md §2.4 ("M4 stores the facts it needs").
--
-- Why stored rather than derived at read time: cash_over_limit depends on the cash-payment
-- limit valid on paid_on (RateCatalog parameter data, never in the database) and on the
-- category's electronic-payment rule; the candidate also on the category's
-- iva_creditable_default for the payer's regime and person type. Re-deriving them on every
-- estimate would need those rules and parameters inside the aggregate query, and a CPA
-- update of a rule would silently rewrite past months. The use case writes them on every
-- create and edit, from the same Assess call that sets deductible.
--
--   * cash_over_limit: paid in cash over the dated limit, or in cash for a category that
--     needs an electronic payment (LISR 27 fr. III ⚠️ CPA). Only a cash payment can be;
--   * creditable_iva_cents: the creditable-IVA candidate (LIVA 5 ⚠️ CPA): with a CFDI,
--     never on a cash payment over the limit (fr. II, whatever the override), a creditable
--     category (fr. I), deductible or RESICO PF. Never above iva_cents.
--
-- Rows written before this version get no facts (false / 0): the module is not released
-- yet (module_expenses is registered in M4-11), and an edit recomputes them.

-- +goose Up

ALTER TABLE expenses
  ADD COLUMN cash_over_limit      boolean NOT NULL DEFAULT false,
  ADD COLUMN creditable_iva_cents bigint  NOT NULL DEFAULT 0,
  ADD CONSTRAINT expenses_cash_over_limit_is_cash CHECK (NOT cash_over_limit OR payment_method = 'cash'),
  ADD CONSTRAINT expenses_creditable_iva_within_iva CHECK (creditable_iva_cents BETWEEN 0 AND iva_cents),
  ADD CONSTRAINT expenses_creditable_iva_needs_cfdi_and_no_cash_over_limit CHECK (
    creditable_iva_cents = 0 OR (has_cfdi AND NOT cash_over_limit));
COMMENT ON COLUMN expenses.cash_over_limit IS
  'Cash payment over the dated limit, or cash for an electronic-payment-only category (LISR 27 fr. III), as assessed on write.';
COMMENT ON COLUMN expenses.creditable_iva_cents IS
  'Creditable-IVA candidate (LIVA 5) the domain assessed on write; the tax engine sums it by paid_on.';

GRANT INSERT (cash_over_limit, creditable_iva_cents) ON expenses TO cc_app;
GRANT UPDATE (cash_over_limit, creditable_iva_cents) ON expenses TO cc_app;

-- +goose Down
ALTER TABLE expenses
  DROP COLUMN creditable_iva_cents,
  DROP COLUMN cash_over_limit;
