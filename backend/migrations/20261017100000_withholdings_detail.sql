-- Withholdings, stored detail (WH-06 / WH-07): how a derived withholding was computed, so a
-- read serves the contract's Withholding object (isr_rate, iva_basis, withholder) from the
-- stored derivation instead of deriving again.
-- Sources: docs/project-specs/MODULE-withholdings.md §3.3 (stored output), ADR-0008 §2
-- ("derive on write, store the amounts with table_version": a read never re-runs the
-- decision table, so a later rate or rule change cannot rewrite what a record shows) and
-- docs/api/openapi.yaml (Withholding, WithholdingBasis, Withholder).
--
-- Why it is not derivable from what WH-03 stores: the ISR rate of a platform sale depends
-- on whether the seller gave the platform an RFC (2.5 % or 20 %), a 625 regime option the
-- user can change after the write, and IVA-3 is half or all of the IVA for the same reason.
-- The rule ids alone do not pin the figures down.
--
-- Columns (written by the server only, like the WH-03 derivation; no request schema
-- carries them):
--   * withholding_isr_rate    the ISR rate applied (contract Withholding.isr_rate, a
--                             DecimalRate of 6 places, 0 when no ISR is withheld);
--   * withholding_iva_basis   contract WithholdingBasis;
--   * withholding_withholder  the kernel's vocabulary, 'none' | 'receptor' | 'platform'
--                             (the API words 'receptor' per side: 'counterparty' on an
--                             income, 'you' on an expense).
-- The three are set together, and only on a derived row (withholding_table_version set). A
-- derived row without them is a WH-03-era derivation (none exists: nothing derived before
-- this migration ships with WH-06 / WH-07) and is served as underived.
--
-- Lock budget as WH-03: ADD COLUMN without a default is a catalog change; the CHECKs are
-- added NOT VALID and validated by 20261017101000_withholdings_detail_validate in its own
-- transaction. RLS unchanged (same tables, same policies); cc_app gets the column grants.

-- +goose Up

SET LOCAL lock_timeout = '5s';

ALTER TABLE incomes
  ADD COLUMN withholding_isr_rate   numeric(7, 6),
  ADD COLUMN withholding_iva_basis  text,
  ADD COLUMN withholding_withholder text,
  ADD CONSTRAINT incomes_withholding_isr_rate_range CHECK (withholding_isr_rate BETWEEN 0 AND 1)
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_iva_basis_valid CHECK (withholding_iva_basis IN (
    'none', 'two_thirds_of_iva', 'half_of_iva', 'all_of_iva', 'four_percent_of_price'))
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_withholder_valid CHECK (withholding_withholder IN ('none', 'receptor', 'platform'))
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_detail_together CHECK (
    (withholding_isr_rate IS NULL) = (withholding_iva_basis IS NULL)
    AND (withholding_isr_rate IS NULL) = (withholding_withholder IS NULL))
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_detail_needs_derivation CHECK (
    withholding_isr_rate IS NULL OR withholding_table_version IS NOT NULL)
    NOT VALID;
COMMENT ON COLUMN incomes.withholding_isr_rate IS 'Derived by the server: ISR withholding rate applied (Withholding.isr_rate).';
COMMENT ON COLUMN incomes.withholding_iva_basis IS 'Derived by the server: openapi WithholdingBasis.';
COMMENT ON COLUMN incomes.withholding_withholder IS
  'Derived by the server: none | receptor | platform (receptor is the payer of the income).';

GRANT INSERT (withholding_isr_rate, withholding_iva_basis, withholding_withholder) ON incomes TO cc_app;
GRANT UPDATE (withholding_isr_rate, withholding_iva_basis, withholding_withholder) ON incomes TO cc_app;

ALTER TABLE expenses
  ADD COLUMN withholding_isr_rate   numeric(7, 6),
  ADD COLUMN withholding_iva_basis  text,
  ADD COLUMN withholding_withholder text,
  ADD CONSTRAINT expenses_withholding_isr_rate_range CHECK (withholding_isr_rate BETWEEN 0 AND 1)
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_iva_basis_valid CHECK (withholding_iva_basis IN (
    'none', 'two_thirds_of_iva', 'half_of_iva', 'all_of_iva', 'four_percent_of_price'))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_withholder_valid CHECK (withholding_withholder IN ('none', 'receptor', 'platform'))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_detail_together CHECK (
    (withholding_isr_rate IS NULL) = (withholding_iva_basis IS NULL)
    AND (withholding_isr_rate IS NULL) = (withholding_withholder IS NULL))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_detail_needs_derivation CHECK (
    withholding_isr_rate IS NULL OR withholding_table_version IS NOT NULL)
    NOT VALID;
COMMENT ON COLUMN expenses.withholding_isr_rate IS 'Derived by the server: ISR withholding rate applied (Withholding.isr_rate).';
COMMENT ON COLUMN expenses.withholding_iva_basis IS 'Derived by the server: openapi WithholdingBasis.';
COMMENT ON COLUMN expenses.withholding_withholder IS
  'Derived by the server: none | receptor | platform (receptor is the taxpayer, who must remit).';

GRANT INSERT (withholding_isr_rate, withholding_iva_basis, withholding_withholder) ON expenses TO cc_app;
GRANT UPDATE (withholding_isr_rate, withholding_iva_basis, withholding_withholder) ON expenses TO cc_app;

-- +goose Down
SET LOCAL lock_timeout = '5s';
ALTER TABLE expenses
  DROP COLUMN withholding_withholder,
  DROP COLUMN withholding_iva_basis,
  DROP COLUMN withholding_isr_rate;
ALTER TABLE incomes
  DROP COLUMN withholding_withholder,
  DROP COLUMN withholding_iva_basis,
  DROP COLUMN withholding_isr_rate;
