-- Taxpayer tax inputs (M5-20): the per-fiscal-year facts the estimate needs that no record
-- carries (prior provisional payments, coeficiente de utilidad, PTU paid, ...). Sources:
-- docs/project-specs/MODULE-tax-engine.md §2.8, §6 (S-1), §7, §9 item 6 and OQ-T1;
-- docs/api/openapi.yaml TaxInputDef / TaxInputValue / TaxInputsUpdate (frozen, 0.6.0);
-- docs/architecture/data-model.md (taxpayer_tax_inputs).
--
-- One row per stored value of (taxpayer, regime, fiscal year, input key[, month]):
--   * which keys exist, and their kind, is catalog data (regimes.required_inputs, M2-02 +
--     20261014140000_regime_tax_input_kinds.sql). kind is DERIVED from it on every write
--     by the trigger taxpayer_tax_inputs_derive_kind; cc_app cannot write it. A key the
--     regime does not define is refused (23514, constraint
--     taxpayer_tax_inputs_key_in_catalog): the database backstop of 400 unknown_input;
--   * exactly the value column of the kind is set (TaxInputValue), month only for
--     money_monthly (OQ-T1: one amount per month actually paid, never assumed):
--       money          value_cents,              month NULL
--       money_monthly  value_cents + month 1-12
--       rate           value_rate,               month NULL
--       boolean        value_bool,               month NULL
--     the database backstop of 400 invalid;
--   * money is bigint centavos in [0, NonNegativeCents.maximum]; a rate is an exact
--     numeric (never a float) in [0, 1] with at most 6 decimals (DecimalRate); numeric
--     keeps the scale it was given, so '0.1500' reads back as '0.1500';
--   * fiscal_year in [2022, 2100] (the catalog floor and FiscalYear.maximum). "Not before
--     2026" (parameters_unavailable), "not in the future" and "the regime is active in that
--     year" (regime_not_active) are clock or window rules of the use case (M5-21);
--   * one value per key and month: UNIQUE NULLS NOT DISTINCT, so an annual key (month
--     NULL) is unique too. No primary key: month is nullable by contract.
--
-- Ownership integrity, enforced by the database rather than trusted to the app (S-1):
--   * ENABLE + FORCE ROW LEVEL SECURITY and one policy per operation the app runs, keyed
--     on (SELECT app_user_id()); an unset context matches no row;
--   * the composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes makes a
--     value for another user's taxpayer, or for a regime the taxpayer never held,
--     impossible even with forged ids (FK checks bypass RLS, so the FK, not the policy,
--     proves the triple belongs together). The regime carries the person type that keys
--     the catalog (626 PF vs 626 PM);
--   * user_id, taxpayer_id, regime_code, fiscal_year, input_key and month are immutable:
--     cc_app has no UPDATE grant on them.
-- PUT replaces all values of (taxpayer, regime, fiscal year) (§2.8): DELETE then INSERT,
-- or INSERT ... ON CONFLICT ON the key, in one transaction. Two concurrent PUTs of the same
-- triple can race to a 23505; the use case serializes them by locking the taxpayer_regimes
-- row first (SELECT ... FOR NO KEY UPDATE: cc_app holds UPDATE on it, and the lock does
-- not block the FK's KEY SHARE checks).
-- Deletion: account deletion (US-A4) cascades through the users FK (the users-FK guard
-- sees this table directly); taxpayer deletion (MODULE-taxpayers §2.7) cascades
-- taxpayers -> taxpayer_regimes -> taxpayer_tax_inputs. Both hold with module_tax_engine
-- OFF (§7). Cascades bypass RLS and need no DELETE grant.

-- +goose Up

CREATE TABLE taxpayer_tax_inputs (
  user_id     uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id uuid        NOT NULL,
  regime_code text        NOT NULL,
  fiscal_year smallint    NOT NULL,
  input_key   text        NOT NULL,
  month       smallint,
  kind        text        NOT NULL,
  value_cents bigint,
  value_rate  numeric,
  value_bool  boolean,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  -- The key of a value. Its btree leads with user_id: it serves the RLS predicate, the
  -- estimate read path (taxpayer + regime + fiscal year, M5-21/M5-22), the upsert, the
  -- composite FK cascade (prefix user_id, taxpayer_id, regime_code) and the users FK
  -- cascade (prefix user_id). No other index is needed.
  CONSTRAINT taxpayer_tax_inputs_key UNIQUE NULLS NOT DISTINCT
    (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month),
  -- S-1: the regime is one this user's taxpayer holds; deleting the taxpayer (and so its
  -- regimes) deletes its inputs.
  CONSTRAINT taxpayer_tax_inputs_taxpayer_regime_fkey FOREIGN KEY (user_id, taxpayer_id, regime_code)
    REFERENCES taxpayer_regimes (user_id, taxpayer_id, regime_code) ON DELETE CASCADE,
  CONSTRAINT taxpayer_tax_inputs_fiscal_year_range CHECK (fiscal_year BETWEEN 2022 AND 2100),
  -- openapi TaxInputDef.key.
  CONSTRAINT taxpayer_tax_inputs_input_key_format CHECK (
    input_key ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$' AND char_length(input_key) <= 64),
  CONSTRAINT taxpayer_tax_inputs_kind_valid CHECK (kind IN ('money', 'money_monthly', 'rate', 'boolean')),
  -- Exactly the value column of the kind; month only for money_monthly. Every branch is
  -- built from IS [NOT] NULL, so it is never NULL (a NULL CHECK would pass).
  CONSTRAINT taxpayer_tax_inputs_value_matches_kind CHECK (
    CASE kind
      WHEN 'money' THEN value_cents IS NOT NULL AND month IS NULL AND value_rate IS NULL AND value_bool IS NULL
      WHEN 'money_monthly' THEN value_cents IS NOT NULL AND month IS NOT NULL AND value_rate IS NULL AND value_bool IS NULL
      WHEN 'rate' THEN value_rate IS NOT NULL AND month IS NULL AND value_cents IS NULL AND value_bool IS NULL
      WHEN 'boolean' THEN value_bool IS NOT NULL AND month IS NULL AND value_cents IS NULL AND value_rate IS NULL
      ELSE false
    END),
  CONSTRAINT taxpayer_tax_inputs_month_range CHECK (month BETWEEN 1 AND 12),
  -- 99999999999999 = docs/api/openapi.yaml NonNegativeCents.maximum.
  CONSTRAINT taxpayer_tax_inputs_value_cents_range CHECK (value_cents BETWEEN 0 AND 99999999999999),
  -- DecimalRate with at most 6 decimals, in [0, 1] (the coeficiente de utilidad,
  -- §2.8). NaN and Infinity compare above 1, so they are refused.
  CONSTRAINT taxpayer_tax_inputs_value_rate_range CHECK (
    value_rate BETWEEN 0 AND 1 AND scale(value_rate) <= 6)
);
COMMENT ON TABLE taxpayer_tax_inputs IS
  'Per-fiscal-year tax inputs per taxpayer + regime (RLS on user_id). Keys and kinds from regimes.required_inputs.';
COMMENT ON COLUMN taxpayer_tax_inputs.kind IS
  'money | money_monthly | rate | boolean. Derived from regimes.required_inputs by taxpayer_tax_inputs_derive_kind; never written by the app.';
COMMENT ON COLUMN taxpayer_tax_inputs.month IS
  'Month 1-12 of the fiscal year, only for money_monthly (OQ-T1); NULL for annual inputs.';
COMMENT ON COLUMN taxpayer_tax_inputs.value_rate IS
  'Exact decimal in [0, 1], at most 6 decimals (openapi DecimalRate). Never a float.';

-- taxpayer_tax_inputs_derive_kind(): BEFORE INSERT / UPDATE OF the key columns. Sets kind
-- from the definition of input_key in the catalog row of the value's regime; the regime
-- is read through the taxpayer_regimes row the composite FK will check (it carries the
-- person type that keys the catalog). SECURITY INVOKER: it reads the caller's rows under
-- the caller's RLS. It never guesses a kind: when that row is not visible it refuses the
-- write with what the later checks would say, since BEFORE triggers run first: 42501 for
-- a row of another user or a writer outside the app context (RLS WITH CHECK), 23503 for
-- the caller's own forged triple (the composite FK). A key the regime does not define is
-- 23514. No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION taxpayer_tax_inputs_derive_kind() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  inputs   jsonb;
  def_kind text;
BEGIN
  SELECT r.required_inputs INTO inputs
  FROM public.taxpayer_regimes tr
  JOIN public.regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
  WHERE tr.user_id = NEW.user_id AND tr.taxpayer_id = NEW.taxpayer_id AND tr.regime_code = NEW.regime_code;
  IF inputs IS NULL AND NEW.user_id IS DISTINCT FROM public.app_user_id() THEN
    RAISE EXCEPTION 'new row violates row-level security policy for table "taxpayer_tax_inputs"'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF inputs IS NULL THEN
    RAISE EXCEPTION 'tax input references a regime its taxpayer does not hold'
      USING ERRCODE = 'foreign_key_violation', TABLE = 'taxpayer_tax_inputs',
            CONSTRAINT = 'taxpayer_tax_inputs_taxpayer_regime_fkey';
  END IF;
  SELECT d ->> 'kind' INTO def_kind
  FROM jsonb_array_elements(inputs) AS d
  WHERE d ->> 'key' = NEW.input_key;
  IF def_kind IS NULL THEN
    RAISE EXCEPTION 'input key % is not defined for regime %', NEW.input_key, NEW.regime_code
      USING ERRCODE = 'check_violation', TABLE = 'taxpayer_tax_inputs',
            CONSTRAINT = 'taxpayer_tax_inputs_key_in_catalog';
  END IF;
  NEW.kind := def_kind;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION taxpayer_tax_inputs_derive_kind() IS
  'BEFORE INSERT/UPDATE trigger: taxpayer_tax_inputs.kind = the kind of input_key in regimes.required_inputs.';

CREATE TRIGGER taxpayer_tax_inputs_derive_kind
  BEFORE INSERT OR UPDATE OF user_id, taxpayer_id, regime_code, input_key, kind ON taxpayer_tax_inputs
  FOR EACH ROW EXECUTE FUNCTION taxpayer_tax_inputs_derive_kind();
CREATE TRIGGER taxpayer_tax_inputs_set_updated_at
  BEFORE UPDATE ON taxpayer_tax_inputs
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE taxpayer_tax_inputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE taxpayer_tax_inputs FORCE ROW LEVEL SECURITY;
-- Intent: a user reads, saves, changes and clears only the tax inputs of their own
-- taxpayers; a write can never hand a row to another user (WITH CHECK). DELETE is the
-- replace of PUT (an empty array clears the triple): inputs are not records, the audit
-- event tax_inputs.updated (S-8) is their history.
CREATE POLICY taxpayer_tax_inputs_select_own ON taxpayer_tax_inputs FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayer_tax_inputs_insert_own ON taxpayer_tax_inputs FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayer_tax_inputs_update_own ON taxpayer_tax_inputs FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayer_tax_inputs_delete_own ON taxpayer_tax_inputs FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
-- INSERT names the key and the value, never kind or the audit columns. UPDATE changes a
-- value in place (upsert); the key columns never change. cc_webhook and cc_worker get
-- nothing: no job computes an estimate yet (M9 adds its own grant and policy if it does).
GRANT SELECT, DELETE ON taxpayer_tax_inputs TO cc_app;
GRANT INSERT (user_id, taxpayer_id, regime_code, fiscal_year, input_key, month, value_cents, value_rate, value_bool)
  ON taxpayer_tax_inputs TO cc_app;
GRANT UPDATE (value_cents, value_rate, value_bool) ON taxpayer_tax_inputs TO cc_app;

-- +goose Down
DROP TABLE taxpayer_tax_inputs;
DROP FUNCTION taxpayer_tax_inputs_derive_kind();
