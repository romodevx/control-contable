-- Withholdings, expand step (WH-03): the facts the server derives ISR and IVA withholdings
-- from, and the stored derivation, on incomes and expenses.
-- Sources: docs/project-specs/MODULE-withholdings.md §2 (questions and defaults), §3.3
-- (stored output), §5.1 (schema), ADR-0008 (derived, stored with table_version; expand ->
-- contract) and docs/api/openapi.yaml 0.10.0 (CounterpartyKind, OperationKind,
-- SupplierKind, SupplierRegime, WithholdingRuleId, Withholding.table_version).
--
-- Expand only: nothing is dropped or renamed. incomes.counterparty_type, its CHECKs and the
-- typed isr_withheld_cents / iva_withheld_cents stay as they are; from WH-06 / WH-07 the
-- server writes the derived amounts into those same two columns (the read-only mirrors of
-- Withholding.isr_cents / iva_cents). Dropping counterparty_type is the contract step
-- (WH-08, after the backfill).
--
-- Lock budget. ADD COLUMN with a constant default is a catalog change (no rewrite), but it
-- still takes ACCESS EXCLUSIVE on the table: lock_timeout makes the deploy fail fast
-- instead of queueing every read behind a long transaction. Every CHECK on incomes and
-- expenses is added NOT VALID (enforced on new writes at once, no scan here) and is
-- validated by the next migration, 20261016131000_withholdings_validate, in its own
-- transaction: VALIDATE takes only SHARE UPDATE EXCLUSIVE, so reads and writes go on while
-- it scans (MODULE-withholdings §5.1, PLAN #91 note 2). Existing rows satisfy every new
-- CHECK by construction (new columns take their defaults), so validation cannot fail.
--
-- New facts (nullable until the backfill, WH-08; NOT NULL is a contract-step decision):
--   * incomes.counterparty_kind   CounterpartyKind, Q1 "¿A quién le facturaste?";
--   * operation_kind (both)       OperationKind, Q2 / E3. A platform sale is not stored as an
--                                 operation kind: it is via_platform (the contract's
--                                 OperationKind has no 'platform' value, so a stored one could
--                                 never be served);
--   * residential_rent (both)     Q3 / E4. true requires iva_rate = 'exempt' (LIVA 20 fr. II;
--                                 422 residential_rent_exempt in the use case);
--   * incomes.via_platform        Q4. Which regimes may set it (625, personas morales) is the
--                                 use case's 400 not_for_regime: 626 is both PF and PM and the
--                                 person type is not on the row;
--   * expenses.supplier_kind,     E1 / E2, persona moral taxpayers only (the use case answers
--     supplier_regime             422 supplier_not_applicable for a persona física: the person
--                                 type is not on the row). The regime is stored exactly when
--                                 the supplier is a persona física with activity, 'unknown'
--                                 for "No sé" (derived as 612 with a warning).
-- Stored derivation (written by the server only, like cash_over_limit: cc_app holds the
-- column grants because the use case writes through it, and no request schema carries
-- them, ADR-0008):
--   * withholding_rule_ids        WithholdingRuleId[] (<= 4, contract maxItems), '{}' until
--                                 derived;
--   * withholding_table_version   '2026.1' style (contract pattern); NULL until derived, and
--                                 set exactly when rule ids are;
--   * withholding_cpa_unverified  a ⚠️ rule matched; only on a derived row;
--   * expenses.creditable_iva_withheld_cents
--                                 the creditable share of the IVA withheld from the supplier,
--                                 credited the month after remittance (LIVA 5 fr. IV;
--                                 MODULE-withholdings §4.2). creditable_iva_cents becomes the
--                                 paid part (IVA − withheld) once WH-07 / WH-08 re-split it;
--                                 until then legacy rows hold the full candidate there and 0
--                                 here, which the new bounds accept.
-- RLS is unchanged: same tables, same per-operation policies on user_id; no new user_id
-- column (the users-FK guard sees nothing new).
--
-- Catalog: category_regime_rules.default_operation_kind, the E3 default per expense
-- category (MODULE-withholdings §2.2), seed data rather than code. Only persona moral rows
-- carry it: a persona física never withholds on domestic purchases, so the question does
-- not exist for it. Global configuration written only by migrations (M4-02 pattern).

-- +goose Up

SET LOCAL lock_timeout = '5s';

-- withholding_rule_ids_valid: ids is a one-dimensional list of at most 4 contract
-- WithholdingRuleId values ('ISR-n', 'IVA-n', 'IVA-NONE'), no NULL element. A function
-- because a CHECK cannot hold a subquery; cc_app needs EXECUTE to write the columns.
CREATE FUNCTION withholding_rule_ids_valid(ids text[]) RETURNS boolean
  LANGUAGE sql IMMUTABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
  RETURN cardinality(ids) <= 4
    AND (cardinality(ids) = 0 OR array_ndims(ids) = 1)
    AND NOT EXISTS (
      SELECT 1 FROM unnest(ids) AS id
      WHERE id IS NULL OR id !~ '^(ISR-[0-9]{1,2}|IVA-([0-9]{1,2}|NONE))$');
COMMENT ON FUNCTION withholding_rule_ids_valid(text[]) IS
  'True when ids is at most 4 openapi WithholdingRuleId values (CHECK of incomes/expenses.withholding_rule_ids).';
GRANT EXECUTE ON FUNCTION withholding_rule_ids_valid(text[]) TO cc_app;

-- ---------------------------------------------------------------------------------------
-- incomes
-- ---------------------------------------------------------------------------------------
ALTER TABLE incomes
  ADD COLUMN counterparty_kind          text,
  ADD COLUMN operation_kind             text,
  ADD COLUMN residential_rent           boolean NOT NULL DEFAULT false,
  ADD COLUMN via_platform               boolean NOT NULL DEFAULT false,
  ADD COLUMN withholding_rule_ids       text[]  NOT NULL DEFAULT '{}',
  ADD COLUMN withholding_table_version  text,
  ADD COLUMN withholding_cpa_unverified boolean NOT NULL DEFAULT false,
  ADD CONSTRAINT incomes_counterparty_kind_valid CHECK (counterparty_kind IN (
    'pm', 'pf_with_activity', 'pf_no_activity_or_public', 'gov_federal', 'gov_state_municipal', 'foreign'))
    NOT VALID,
  ADD CONSTRAINT incomes_operation_kind_valid CHECK (operation_kind IN (
    'professional_services', 'business_services', 'rent_real_estate', 'rent_movable', 'sale_of_goods',
    'ground_freight', 'commission', 'scrap', 'other'))
    NOT VALID,
  ADD CONSTRAINT incomes_residential_rent_is_exempt CHECK (NOT residential_rent OR iva_rate = 'exempt')
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_rule_ids_valid CHECK (withholding_rule_ids_valid(withholding_rule_ids))
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_table_version_format CHECK (
    withholding_table_version ~ '^[0-9]{4}\.[0-9]{1,3}$')
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_derived_together CHECK (
    (withholding_table_version IS NULL) = (cardinality(withholding_rule_ids) = 0))
    NOT VALID,
  ADD CONSTRAINT incomes_withholding_cpa_unverified_needs_derivation CHECK (
    NOT withholding_cpa_unverified OR withholding_table_version IS NOT NULL)
    NOT VALID;
COMMENT ON COLUMN incomes.counterparty_kind IS
  'Q1 (openapi CounterpartyKind). Replaces counterparty_type (dropped in the contract step); NULL until backfilled.';
COMMENT ON COLUMN incomes.operation_kind IS
  'Q2 (openapi OperationKind). A platform sale is via_platform, never an operation kind.';
COMMENT ON COLUMN incomes.residential_rent IS 'Q3: housing rent; requires iva_rate = exempt (LIVA 20 fr. II).';
COMMENT ON COLUMN incomes.via_platform IS 'Q4: collected through a digital platform (625, personas morales).';
COMMENT ON COLUMN incomes.withholding_rule_ids IS
  'Derived by the server (ADR-0008): decision-table rule ids of isr/iva_withheld_cents; {} until derived.';
COMMENT ON COLUMN incomes.withholding_table_version IS
  'Derived by the server: decision table + rates version that produced the withholdings; NULL until derived.';
COMMENT ON COLUMN incomes.withholding_cpa_unverified IS 'Derived by the server: a rule the CPA has not confirmed matched.';

-- Same pattern as the M3-02 grants: the use cases write through cc_app; the API never
-- accepts the derived columns (ADR-0008).
GRANT INSERT (counterparty_kind, operation_kind, residential_rent, via_platform, withholding_rule_ids,
              withholding_table_version, withholding_cpa_unverified)
  ON incomes TO cc_app;
GRANT UPDATE (counterparty_kind, operation_kind, residential_rent, via_platform, withholding_rule_ids,
              withholding_table_version, withholding_cpa_unverified)
  ON incomes TO cc_app;

-- ---------------------------------------------------------------------------------------
-- expenses
-- ---------------------------------------------------------------------------------------
ALTER TABLE expenses
  ADD COLUMN supplier_kind                 text,
  ADD COLUMN supplier_regime               text,
  ADD COLUMN operation_kind                text,
  ADD COLUMN residential_rent              boolean NOT NULL DEFAULT false,
  ADD COLUMN creditable_iva_withheld_cents bigint  NOT NULL DEFAULT 0,
  ADD COLUMN withholding_rule_ids          text[]  NOT NULL DEFAULT '{}',
  ADD COLUMN withholding_table_version     text,
  ADD COLUMN withholding_cpa_unverified    boolean NOT NULL DEFAULT false,
  ADD CONSTRAINT expenses_supplier_kind_valid CHECK (supplier_kind IN ('pm', 'pf_with_activity', 'foreign'))
    NOT VALID,
  ADD CONSTRAINT expenses_supplier_regime_valid CHECK (
    supplier_regime IN ('612', '606', '626', '625', '621', 'unknown'))
    NOT VALID,
  -- E2 is asked exactly when E1 = pf_with_activity ('unknown' records "No sé").
  ADD CONSTRAINT expenses_supplier_regime_iff_pf_with_activity CHECK (
    (supplier_kind IS NOT DISTINCT FROM 'pf_with_activity') = (supplier_regime IS NOT NULL))
    NOT VALID,
  ADD CONSTRAINT expenses_operation_kind_valid CHECK (operation_kind IN (
    'professional_services', 'business_services', 'rent_real_estate', 'rent_movable', 'sale_of_goods',
    'ground_freight', 'commission', 'scrap', 'other'))
    NOT VALID,
  ADD CONSTRAINT expenses_residential_rent_is_exempt CHECK (NOT residential_rent OR iva_rate = 'exempt')
    NOT VALID,
  -- The withheld creditable part is a share of the IVA withheld; both parts together never
  -- exceed the IVA of the expense, and like the paid part it needs a CFDI and no cash
  -- payment over the limit (LIVA 5 fr. II, III).
  ADD CONSTRAINT expenses_creditable_iva_withheld_within_withheld CHECK (
    creditable_iva_withheld_cents BETWEEN 0 AND iva_withheld_cents)
    NOT VALID,
  ADD CONSTRAINT expenses_creditable_iva_split_within_iva CHECK (
    creditable_iva_cents + creditable_iva_withheld_cents <= iva_cents)
    NOT VALID,
  ADD CONSTRAINT expenses_creditable_iva_withheld_needs_cfdi_and_no_cash_over_limit CHECK (
    creditable_iva_withheld_cents = 0 OR (has_cfdi AND NOT cash_over_limit))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_rule_ids_valid CHECK (withholding_rule_ids_valid(withholding_rule_ids))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_table_version_format CHECK (
    withholding_table_version ~ '^[0-9]{4}\.[0-9]{1,3}$')
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_derived_together CHECK (
    (withholding_table_version IS NULL) = (cardinality(withholding_rule_ids) = 0))
    NOT VALID,
  ADD CONSTRAINT expenses_withholding_cpa_unverified_needs_derivation CHECK (
    NOT withholding_cpa_unverified OR withholding_table_version IS NOT NULL)
    NOT VALID;
COMMENT ON COLUMN expenses.supplier_kind IS
  'E1 (openapi SupplierKind), persona moral taxpayers only; NULL for a persona fisica.';
COMMENT ON COLUMN expenses.supplier_regime IS
  'E2 (openapi SupplierRegime), set exactly when supplier_kind = pf_with_activity; unknown = "No sé".';
COMMENT ON COLUMN expenses.operation_kind IS 'E3 (openapi OperationKind); default from category_regime_rules.';
COMMENT ON COLUMN expenses.residential_rent IS 'E4: housing rent; requires iva_rate = exempt (LIVA 20 fr. II).';
COMMENT ON COLUMN expenses.creditable_iva_withheld_cents IS
  'Creditable share of the IVA withheld; credited the month after remittance (LIVA 5 fr. IV). creditable_iva_cents is the paid part.';
COMMENT ON COLUMN expenses.withholding_rule_ids IS
  'Derived by the server (ADR-0008): decision-table rule ids of isr/iva_withheld_cents; {} until derived.';
COMMENT ON COLUMN expenses.withholding_table_version IS
  'Derived by the server: decision table + rates version that produced the withholdings; NULL until derived.';
COMMENT ON COLUMN expenses.withholding_cpa_unverified IS 'Derived by the server: a rule the CPA has not confirmed matched.';

-- Same pattern as cash_over_limit / creditable_iva_cents (M4-08): the use cases write the
-- facts and the derivation through cc_app; the API never accepts the derived columns.
GRANT INSERT (supplier_kind, supplier_regime, operation_kind, residential_rent, creditable_iva_withheld_cents,
              withholding_rule_ids, withholding_table_version, withholding_cpa_unverified)
  ON expenses TO cc_app;
GRANT UPDATE (supplier_kind, supplier_regime, operation_kind, residential_rent, creditable_iva_withheld_cents,
              withholding_rule_ids, withholding_table_version, withholding_cpa_unverified)
  ON expenses TO cc_app;

-- ---------------------------------------------------------------------------------------
-- category_regime_rules.default_operation_kind (MODULE-withholdings §2.2). Codes the M4
-- seed lacks (arrendamiento_equipo) get their default when they are added. ⚠️ The CPA
-- reviews this mapping with the catalog (the rows keep needs_cpa_review). Small global
-- catalog: validated in this transaction.
-- ---------------------------------------------------------------------------------------
ALTER TABLE category_regime_rules ADD COLUMN default_operation_kind text;
UPDATE category_regime_rules
SET default_operation_kind = CASE category_code
  WHEN 'servicios_profesionales' THEN 'professional_services'
  WHEN 'renta_local' THEN 'rent_real_estate'
  WHEN 'arrendamiento_equipo' THEN 'rent_movable'
  WHEN 'fletes' THEN 'ground_freight'
  WHEN 'comisiones' THEN 'commission'
  WHEN 'mercancia' THEN 'sale_of_goods'
  ELSE 'business_services'
END
WHERE person_type = 'moral';
ALTER TABLE category_regime_rules
  ADD CONSTRAINT category_regime_rules_default_operation_kind_valid CHECK (default_operation_kind IN (
    'professional_services', 'business_services', 'rent_real_estate', 'rent_movable', 'sale_of_goods',
    'ground_freight', 'commission', 'scrap', 'other')),
  ADD CONSTRAINT category_regime_rules_default_operation_kind_iff_moral CHECK (
    (person_type = 'moral') = (default_operation_kind IS NOT NULL));
COMMENT ON COLUMN category_regime_rules.default_operation_kind IS
  'E3 default for a persona moral paying a supplier (MODULE-withholdings 2.2); NULL for personas fisicas.';

-- +goose Down
ALTER TABLE category_regime_rules DROP COLUMN default_operation_kind;
ALTER TABLE expenses
  DROP COLUMN withholding_cpa_unverified,
  DROP COLUMN withholding_table_version,
  DROP COLUMN withholding_rule_ids,
  DROP COLUMN creditable_iva_withheld_cents,
  DROP COLUMN residential_rent,
  DROP COLUMN operation_kind,
  DROP COLUMN supplier_regime,
  DROP COLUMN supplier_kind;
ALTER TABLE incomes
  DROP COLUMN withholding_cpa_unverified,
  DROP COLUMN withholding_table_version,
  DROP COLUMN withholding_rule_ids,
  DROP COLUMN via_platform,
  DROP COLUMN residential_rent,
  DROP COLUMN operation_kind,
  DROP COLUMN counterparty_kind;
DROP FUNCTION withholding_rule_ids_valid(text[]);
