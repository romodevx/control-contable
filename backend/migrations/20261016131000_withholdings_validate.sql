-- Withholdings, validate step (WH-03): validates the CHECKs that
-- 20261016130000_withholdings_expand added NOT VALID on incomes and expenses.
--
-- Its own migration, so its own transaction: VALIDATE CONSTRAINT takes SHARE UPDATE
-- EXCLUSIVE, which lets reads and writes continue while it scans the table. Run in the
-- expand transaction it would scan under that migration's ACCESS EXCLUSIVE lock
-- (MODULE-withholdings §5.1, PLAN #91 note 2). Every existing row passes: the checked
-- columns are new and hold their defaults.

-- +goose Up

ALTER TABLE incomes VALIDATE CONSTRAINT incomes_counterparty_kind_valid;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_operation_kind_valid;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_residential_rent_is_exempt;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_rule_ids_valid;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_table_version_format;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_derived_together;
ALTER TABLE incomes VALIDATE CONSTRAINT incomes_withholding_cpa_unverified_needs_derivation;

ALTER TABLE expenses VALIDATE CONSTRAINT expenses_supplier_kind_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_supplier_regime_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_supplier_regime_iff_pf_with_activity;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_operation_kind_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_residential_rent_is_exempt;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_creditable_iva_withheld_within_withheld;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_creditable_iva_split_within_iva;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_creditable_iva_withheld_needs_cfdi_and_no_cash_over_limit;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_rule_ids_valid;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_table_version_format;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_derived_together;
ALTER TABLE expenses VALIDATE CONSTRAINT expenses_withholding_cpa_unverified_needs_derivation;

-- +goose Down
-- PostgreSQL cannot mark a validated constraint NOT VALID again, and a validated CHECK is
-- strictly stronger than the expand step's: nothing to undo. Rolling back further, the
-- expand step's Down drops the columns and their constraints.
SELECT 1;
