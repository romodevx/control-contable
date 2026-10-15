-- Withholdings, stored detail, validate step: validates the CHECKs that
-- 20261017100000_withholdings_detail added NOT VALID, in its own transaction (SHARE UPDATE
-- EXCLUSIVE: reads and writes continue while it scans; PLAN #91 note 2). Every existing row
-- passes: the checked columns are new and NULL.

-- +goose Up

ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_isr_rate_range;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_iva_basis_valid;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_withholder_valid;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_detail_together;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_detail_needs_derivation;

ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_isr_rate_range;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_iva_basis_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_withholder_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_detail_together;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_detail_needs_derivation;

-- +goose Down
-- A validated CHECK cannot be marked NOT VALID again and is strictly stronger: nothing to
-- undo; the detail step's Down drops the columns and their constraints.
SELECT 1;
