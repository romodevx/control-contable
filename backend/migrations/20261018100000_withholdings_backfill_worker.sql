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
--     app.current_user_id set (tx.Manager), so every table grant below is scoped by a
--     per-user policy, exactly like cc_app's.
--
-- The list of users to visit is the only cross-user read, and it goes through a SECURITY
-- DEFINER function (data-model "Database roles": definer functions are the only ways to read
-- across users). withholding_backfill_users() returns user ids only and is owned by
-- cc_worker_definer, a NOLOGIN, NOBYPASSRLS role from the bootstrap whose one policy shows
-- users' ids. cc_worker itself still sees no row without a user context (TestSystemAsWorker).
--
-- What cc_worker gets (nothing else):
--   * EXECUTE withholding_backfill_users(): user ids, nothing more.
--   * taxpayers: SELECT of the columns the taxpayers profile reader names (id, person type,
--     display name, masked RFC, archive state, created_at; never the RFC ciphertext, hash or
--     key id), own rows only.
--   * taxpayer_regimes: SELECT, own rows only (regime windows and 625 options).
--   * regimes, expense_categories, category_regime_rules: SELECT (global catalogs, no PII).
--   * incomes / expenses: SELECT own rows (the backfill reads a whole record, description
--     included, but never prints it), and UPDATE own rows of the facts and derived columns
--     only: never an amount, a date, the deduction verdict, the description, deleted_at or a
--     key. No INSERT, no DELETE anywhere. EXECUTE withholding_rule_ids_valid(text[]): the
--     CHECK on the rule ids calls it.
-- RLS stays ENABLE + FORCE on every table; this migration only adds policies TO cc_worker and
-- cc_worker_definer.

-- +goose Up

-- Fail early, with a clear message, on an environment whose bootstrap predates WH-08.
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_roles
    WHERE rolname = 'cc_worker_definer' AND NOT rolcanlogin AND NOT rolsuper AND NOT rolbypassrls
  ) THEN
    RAISE EXCEPTION 'role cc_worker_definer is missing or privileged'
      USING HINT = 'Create it as in infra/local/postgres/bootstrap/roles.sql (docs/runbooks/withholdings-backfill.md).';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_auth_members
    WHERE roleid = 'cc_worker_definer'::regrole AND member = 'cc_owner'::regrole
      AND set_option AND NOT inherit_option
  ) THEN
    RAISE EXCEPTION 'cc_owner must be a member of cc_worker_definer WITH INHERIT FALSE, SET TRUE'
      USING HINT = 'See infra/local/postgres/bootstrap/roles.sql.';
  END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO cc_worker_definer;
CREATE POLICY users_worker_definer_ids ON users FOR SELECT TO cc_worker_definer USING (true);
GRANT SELECT (id) ON users TO cc_worker_definer;

-- +goose StatementBegin
CREATE FUNCTION withholding_backfill_users() RETURNS SETOF uuid
  LANGUAGE sql STABLE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$ SELECT u.id FROM public.users AS u ORDER BY u.id $$;
-- +goose StatementEnd
COMMENT ON FUNCTION withholding_backfill_users() IS
  'WH-08: every user id, ordered, for cmd/backfill-withholdings. Definer: cc_worker_definer; EXECUTE: cc_worker.';
-- ALTER OWNER needs the new owner to hold CREATE on the schema: lent for this statement only.
GRANT CREATE ON SCHEMA public TO cc_worker_definer;
ALTER FUNCTION withholding_backfill_users() OWNER TO cc_worker_definer;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;
REVOKE ALL ON FUNCTION withholding_backfill_users() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION withholding_backfill_users() TO cc_worker;

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

GRANT EXECUTE ON FUNCTION withholding_rule_ids_valid(text[]) TO cc_worker;

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
REVOKE EXECUTE ON FUNCTION withholding_rule_ids_valid(text[]) FROM cc_worker;
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

-- Only its owner drops the function: cc_owner sets the role, inheriting nothing.
SET LOCAL ROLE cc_worker_definer;
DROP FUNCTION withholding_backfill_users();
RESET ROLE;
REVOKE ALL ON users FROM cc_worker_definer;
DROP POLICY users_worker_definer_ids ON users;
REVOKE USAGE ON SCHEMA public FROM cc_worker_definer;
