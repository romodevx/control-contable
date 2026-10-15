-- Withholdings backfill role (WH-08): what cmd/backfill-withholdings needs as cc_worker.
-- Sources: docs/project-specs/MODULE-withholdings.md §5.2 / §5.3, ADR-0008, PLAN WH-08 and
-- docs/runbooks/withholdings-backfill.md.
--
-- Why cc_worker and not the owner or cc_app:
--   * cc_owner is subject to FORCE ROW LEVEL SECURITY with no policy of its own, so it sees
--     no user row (and must stay the DDL-only role);
--   * cc_app is the API: giving it a cross-user list would widen every request's reach;
--   * cc_worker is the documented home of cross-user jobs (cmd/worker): LOGIN, NOBYPASSRLS,
--     not owner, 60 s statement timeout. The backfill runs one transaction per user with
--     app.current_user_id set (tx.Manager), so every grant below except the user list is
--     scoped by a per-user policy, exactly like cc_app's.
--
-- What cc_worker gets (nothing else):
--   * users: SELECT (id) only, every row (USING true): the list of users to visit. No
--     email, name, hash or status column is readable.
--   * taxpayers: SELECT of the columns the taxpayers profile reader names (id, person type,
--     display name, masked RFC, archive state, created_at; never the RFC ciphertext, hash or
--     key id), own rows only.
--   * taxpayer_regimes: SELECT, own rows only (regime windows and 625 options).
--   * regimes, expense_categories, category_regime_rules: SELECT (global catalogs, no PII).
--   * incomes / expenses: SELECT own rows (the backfill reads every column of a record,
--     description included, but never prints it), and UPDATE own rows of the facts and
--     derived columns only: never an amount, a date, the deduction verdict, the description,
--     deleted_at or a key. No INSERT, no DELETE anywhere.
-- RLS stays ENABLE + FORCE on every table; this migration only adds policies TO cc_worker.

-- +goose Up

CREATE POLICY users_select_worker ON users FOR SELECT TO cc_worker USING (true);
GRANT SELECT (id) ON users TO cc_worker;

CREATE POLICY taxpayers_select_worker ON taxpayers FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT (id, user_id, person_type, display_name, rfc_masked, archived_at, created_at) ON taxpayers TO cc_worker;

CREATE POLICY taxpayer_regimes_select_worker ON taxpayer_regimes FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT ON taxpayer_regimes TO cc_worker;

CREATE POLICY regimes_worker_read ON regimes FOR SELECT TO cc_worker USING (true);
GRANT SELECT ON regimes TO cc_worker;
CREATE POLICY expense_categories_worker_read ON expense_categories FOR SELECT TO cc_worker USING (true);
GRANT SELECT ON expense_categories TO cc_worker;
CREATE POLICY category_regime_rules_worker_read ON category_regime_rules FOR SELECT TO cc_worker USING (true);
GRANT SELECT ON category_regime_rules TO cc_worker;

CREATE POLICY incomes_select_worker ON incomes FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY incomes_update_worker ON incomes FOR UPDATE TO cc_worker
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT ON incomes TO cc_worker;
GRANT UPDATE (counterparty_type, counterparty_kind, operation_kind, residential_rent, via_platform,
              isr_withheld_cents, iva_withheld_cents, withholding_rule_ids, withholding_table_version,
              withholding_cpa_unverified, withholding_isr_rate, withholding_iva_basis, withholding_withholder)
  ON incomes TO cc_worker;

CREATE POLICY expenses_select_worker ON expenses FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY expenses_update_worker ON expenses FOR UPDATE TO cc_worker
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT ON expenses TO cc_worker;
GRANT UPDATE (creditable_iva_cents, isr_withheld_cents, iva_withheld_cents, supplier_kind, supplier_regime,
              operation_kind, residential_rent, creditable_iva_withheld_cents, withholding_rule_ids,
              withholding_table_version, withholding_cpa_unverified, withholding_isr_rate, withholding_iva_basis,
              withholding_withholder)
  ON expenses TO cc_worker;

-- +goose Down

REVOKE ALL ON expenses FROM cc_worker;
DROP POLICY expenses_update_worker ON expenses;
DROP POLICY expenses_select_worker ON expenses;
REVOKE ALL ON incomes FROM cc_worker;
DROP POLICY incomes_update_worker ON incomes;
DROP POLICY incomes_select_worker ON incomes;
REVOKE ALL ON category_regime_rules FROM cc_worker;
DROP POLICY category_regime_rules_worker_read ON category_regime_rules;
REVOKE ALL ON expense_categories FROM cc_worker;
DROP POLICY expense_categories_worker_read ON expense_categories;
REVOKE ALL ON regimes FROM cc_worker;
DROP POLICY regimes_worker_read ON regimes;
REVOKE ALL ON taxpayer_regimes FROM cc_worker;
DROP POLICY taxpayer_regimes_select_worker ON taxpayer_regimes;
REVOKE ALL ON taxpayers FROM cc_worker;
DROP POLICY taxpayers_select_worker ON taxpayers;
REVOKE ALL ON users FROM cc_worker;
DROP POLICY users_select_worker ON users;
