-- Baseline (P0-10): RLS session helper, shared trigger, schema posture and the global
-- feature-flag registry. Runs as cc_owner (cmd/migrate refuses any other role), so every
-- object below is owned by cc_owner. Roles and database-level grants come from the
-- bootstrap (infra/local/postgres/bootstrap/roles.sql locally, the secret-managed
-- equivalent in shared environments); docs/architecture/data-model.md lists them.

-- +goose Up

-- Schema posture. Re-asserts the bootstrap so every environment converges to the same
-- least-privilege state even if its roles were created by other tooling:
--   * PUBLIC has nothing on schema public; runtime roles get USAGE only (never CREATE),
--     so the only DDL path is this runner, as cc_owner.
--   * Functions cc_owner creates are not executable by PUBLIC; each migration grants
--     EXECUTE explicitly. Table and sequence privileges are never defaulted either:
--     every migration grants exactly what each runtime role needs, per table.
-- Down does not undo these statements: they restate the bootstrap, they do not extend it.
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO cc_app, cc_webhook, cc_worker;
ALTER DEFAULT PRIVILEGES FOR ROLE cc_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;

-- app_user_id(): the user of the current transaction, set by the TxManager with
-- set_config('app.current_user_id', $1, true). Every user-scoped RLS policy compares
-- `user_id = (SELECT app_user_id())` (an initplan: evaluated once per query).
-- Fails closed in every state of the session:
--   * never set (fresh connection)        -> current_setting(..., true) IS NULL -> NULL
--   * reverted after a transaction ('')   -> nullif(..., '')                    -> NULL
--   * malformed (not a uuid)              -> the cast raises 22P02, the query fails
-- NULL never equals a user_id, so an unset context sees and writes no rows.
-- SECURITY INVOKER (default): it reads only the caller's own setting.
-- +goose StatementBegin
CREATE FUNCTION app_user_id() RETURNS uuid
  LANGUAGE sql STABLE PARALLEL SAFE
AS $$ SELECT nullif(pg_catalog.current_setting('app.current_user_id', true), '')::uuid $$;
-- +goose StatementEnd
COMMENT ON FUNCTION app_user_id() IS
  'User of the current transaction (app.current_user_id); NULL when unset, error when malformed. RLS key.';
GRANT EXECUTE ON FUNCTION app_user_id() TO cc_app, cc_webhook, cc_worker;

-- set_updated_at(): BEFORE UPDATE trigger that stamps updated_at with the transaction
-- time, so the audit column cannot drift from the write that changed the row.
-- No EXECUTE grant: PostgreSQL checks it when the trigger is created (as cc_owner),
-- not when it fires.
-- +goose StatementBegin
CREATE FUNCTION set_updated_at() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog
AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION set_updated_at() IS 'BEFORE UPDATE trigger: updated_at = transaction time.';

-- ---------------------------------------------------------------------------------------
-- Feature flags (global configuration, not user data)
--
-- Why no user-scoped RLS: these rows have no owner. Every user sees the same flag set,
-- so there is no user_id/tenant_id for a policy to key on, and FORCE ROW LEVEL SECURITY
-- would only lock out cc_owner, the one role allowed to write. Access is instead:
--   * privileges: cc_app and cc_worker get SELECT on feature_flags only (the API gate and
--     GET /v1/flags; the worker skips jobs of disabled modules). cc_webhook gets nothing:
--     its route is gated in the API process. No runtime role may read the presets.
--   * read-only RLS as defense in depth: RLS is ENABLED (not forced) with a single
--     SELECT policy for cc_app and cc_worker. Should a later migration grant a write by
--     mistake, no INSERT/UPDATE/DELETE policy exists for any runtime role, so the write
--     is refused (INSERT) or matches no row (UPDATE/DELETE). The presets tables have RLS
--     enabled with no policy at all: deny-all for everyone but their owner.
--   * writes happen only as cc_owner: seeds here, toggles through cmd/flags (P0-20).
-- ---------------------------------------------------------------------------------------

CREATE TABLE feature_flags (
  key         text        PRIMARY KEY,
  enabled     boolean     NOT NULL DEFAULT false,
  description text        NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  -- Flag keys are snake_case "module_<name>" (feature-flags rule, MODULES.md).
  CONSTRAINT feature_flags_key_format CHECK (key ~ '^module_[a-z][a-z0-9]*(_[a-z0-9]+)*$'),
  CONSTRAINT feature_flags_description_present CHECK (btrim(description) <> '')
);
COMMENT ON TABLE feature_flags IS
  'Global module flags; OFF by default. Read by cc_app/cc_worker, written only by cc_owner.';

CREATE TRIGGER feature_flags_set_updated_at
  BEFORE UPDATE ON feature_flags
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE feature_flags ENABLE ROW LEVEL SECURITY;
-- Intent: runtime roles may read every flag (global config); no write policy exists for them.
CREATE POLICY feature_flags_runtime_read ON feature_flags
  FOR SELECT TO cc_app, cc_worker
  USING (true);
GRANT SELECT ON feature_flags TO cc_app, cc_worker;

-- Presentation presets: named flag sets that demos apply through cmd/flags as cc_owner.
CREATE TABLE flag_presets (
  name        text        PRIMARY KEY,
  description text        NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT flag_presets_name_format CHECK (name ~ '^[a-z][a-z0-9]*(_[a-z0-9]+)*$'),
  CONSTRAINT flag_presets_description_present CHECK (btrim(description) <> '')
);
COMMENT ON TABLE flag_presets IS 'Named flag sets for demos (MODULES.md). Owner-only.';

CREATE TRIGGER flag_presets_set_updated_at
  BEFORE UPDATE ON flag_presets
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Preset membership as rows, not a text[] column: the foreign keys guarantee a preset
-- can only name flags that exist, and dropping a flag removes it from every preset.
CREATE TABLE flag_preset_flags (
  preset_name text        NOT NULL REFERENCES flag_presets (name) ON UPDATE CASCADE ON DELETE CASCADE,
  flag_key    text        NOT NULL REFERENCES feature_flags (key) ON UPDATE CASCADE ON DELETE CASCADE,
  created_at  timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (preset_name, flag_key)
);
COMMENT ON TABLE flag_preset_flags IS 'Flags turned ON by each preset; the rest are turned OFF. Owner-only.';
-- The PK serves lookups by preset; this btree serves the FK from feature_flags
-- (cascades on a flag key change or delete would otherwise scan the table).
CREATE INDEX flag_preset_flags_flag_key_idx ON flag_preset_flags (flag_key);

-- Intent: deny-all to every runtime role; no grant and no policy (owner-only tables).
ALTER TABLE flag_presets ENABLE ROW LEVEL SECURITY;
ALTER TABLE flag_preset_flags ENABLE ROW LEVEL SECURITY;

-- Seeds: the nine module flags of MODULES.md, all OFF.
INSERT INTO feature_flags (key, description) VALUES
  ('module_auth',          'M1 Auth & Account'),
  ('module_taxpayers',     'M2 Taxpayers & Onboarding'),
  ('module_incomes',       'M3 Incomes'),
  ('module_expenses',      'M4 Expenses'),
  ('module_tax_engine',    'M5 Tax Engine (estimate + breakdown)'),
  ('module_dashboard',     'M6 Dashboard'),
  ('module_subscriptions', 'M7 Subscriptions (RevenueCat)'),
  ('module_history',       'M8 History'),
  ('module_reminders',     'M9 Reminders (push)');

INSERT INTO flag_presets (name, description) VALUES
  ('demo_core',         'Core flow: auth, taxpayers, incomes, expenses, tax engine, dashboard'),
  ('demo_monetization', 'demo_core + subscriptions'),
  ('demo_full',         'All modules M1-M9');

INSERT INTO flag_preset_flags (preset_name, flag_key)
SELECT 'demo_core', key FROM feature_flags
WHERE key IN ('module_auth', 'module_taxpayers', 'module_incomes', 'module_expenses',
              'module_tax_engine', 'module_dashboard');

INSERT INTO flag_preset_flags (preset_name, flag_key)
SELECT 'demo_monetization', flag_key FROM flag_preset_flags WHERE preset_name = 'demo_core'
UNION ALL
SELECT 'demo_monetization', 'module_subscriptions';

INSERT INTO flag_preset_flags (preset_name, flag_key)
SELECT 'demo_full', key FROM feature_flags;

-- +goose Down
DROP TABLE flag_preset_flags;
DROP TABLE flag_presets;
DROP TABLE feature_flags;
DROP FUNCTION set_updated_at();
DROP FUNCTION app_user_id();
