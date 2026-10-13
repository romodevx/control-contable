-- Identity (P0-11): the users identity root, the trial-cap counters, the entitlements and
-- the service-only pending sign-ups. users moved from M1 to P0 (PLAN decision 1) so the
-- P0 tables' foreign keys resolve; M1 owns the auth behaviour and its other tables.
--
-- RLS posture (docs/architecture/data-model.md, "RLS policy outline"):
--   * users, user_usage, entitlements are user-scoped: ENABLE + FORCE ROW LEVEL SECURITY
--     and one policy per operation and role keyed on (SELECT app_user_id()), an initplan
--     evaluated once per query. An unset context is NULL and matches no row.
--   * email_verifications is pre-account (no user to key on): ENABLE, not FORCE, and no
--     policy. Deny-all to every runtime role; reached only through cc_owner's SECURITY
--     DEFINER functions (M1-03), which FORCE would lock out.
--   * Privileges are granted per table, per role and, for writes, per column. A privilege
--     without a matching policy is still refused or matches no row.
-- Foreign keys and their cascades bypass RLS (PostgreSQL RI checks always do), so account
-- deletion removes the children without granting DELETE on them.

-- +goose Up

-- ---------------------------------------------------------------------------------------
-- users: identity root. The email is never stored in clear (ADR-0001, platform/crypto):
--   * email_hash      HMAC-SHA256 blind index of the normalized email (crypto.NewEmailIndex):
--                     login lookup and the one-account-per-email guarantee.
--   * email_encrypted AES-256-GCM ciphertext (crypto.Keyring): key_id (2 bytes, big endian)
--                     || nonce (12) || ciphertext || tag (16). email_key_id repeats the
--                     prefix so key rotation can select stale rows by a plain column
--                     and a CHECK keeps the two from drifting.
--   * password_hash   argon2id PHC string (crypto.PasswordHasher).
-- Account deletion (US-A4) is a hard DELETE that cascades to every user-scoped table;
-- no soft delete here, by design (ARCO cancellation).
-- ---------------------------------------------------------------------------------------
CREATE TABLE users (
  id                uuid        PRIMARY KEY DEFAULT uuidv7(),
  email_hash        bytea       NOT NULL,
  email_encrypted   bytea       NOT NULL,
  email_key_id      smallint    NOT NULL,
  password_hash     text        NOT NULL,
  name              text        NOT NULL,
  status            text        NOT NULL DEFAULT 'active',
  email_verified_at timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT users_email_hash_key UNIQUE (email_hash),
  -- 32 = crypto.HashSize (HMAC-SHA256).
  CONSTRAINT users_email_hash_size CHECK (octet_length(email_hash) = 32),
  CONSTRAINT users_email_key_id_range CHECK (email_key_id >= 1),
  -- More than the 30-byte frame (2 key_id + 12 nonce + 16 tag), and the key_id prefix
  -- equals email_key_id. CASE guards get_byte against a too-short value.
  CONSTRAINT users_email_ciphertext_framed CHECK (
    CASE WHEN octet_length(email_encrypted) > 30
      THEN email_key_id = (get_byte(email_encrypted, 0) << 8) | get_byte(email_encrypted, 1)
      ELSE false
    END),
  CONSTRAINT users_password_hash_phc CHECK (
    password_hash ~ '^\$argon2id\$v=19\$m=[0-9]+,t=[0-9]+,p=[0-9]+\$[A-Za-z0-9+/]+\$[A-Za-z0-9+/]+$'),
  CONSTRAINT users_name_present CHECK (btrim(name) <> '' AND char_length(name) <= 200),
  CONSTRAINT users_status_valid CHECK (status IN ('active', 'suspended'))
);
COMMENT ON TABLE users IS
  'Identity root (RLS on id). Email kept as HMAC blind index + AES-GCM ciphertext; never in clear.';

CREATE TRIGGER users_set_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE users FORCE ROW LEVEL SECURITY;
-- Intent: a user reads and changes only their own row. Registration sets the context to
-- the new id before the INSERT; login reads across users only through the definer
-- function auth_lookup_credentials (M1-03), never through these policies.
CREATE POLICY users_select_own ON users FOR SELECT TO cc_app
  USING (id = (SELECT app_user_id()));
CREATE POLICY users_insert_own ON users FOR INSERT TO cc_app
  WITH CHECK (id = (SELECT app_user_id()));
CREATE POLICY users_update_own ON users FOR UPDATE TO cc_app
  USING (id = (SELECT app_user_id()))
  WITH CHECK (id = (SELECT app_user_id()));
CREATE POLICY users_delete_own ON users FOR DELETE TO cc_app
  USING (id = (SELECT app_user_id()));
-- id, status and the audit columns are not writable by the API; status is an operator
-- decision (cc_owner). cc_webhook and cc_worker get nothing: FK checks need no grant.
GRANT SELECT, DELETE ON users TO cc_app;
GRANT INSERT (id, email_hash, email_encrypted, email_key_id, password_hash, name, email_verified_at)
  ON users TO cc_app;
GRANT UPDATE (email_hash, email_encrypted, email_key_id, password_hash, name, email_verified_at)
  ON users TO cc_app;

-- ---------------------------------------------------------------------------------------
-- user_usage: per-user created-record counters for the trial caps (PRD §7). One row per
-- user, created zeroed at sign-up in the same transaction as users, so the atomic
-- reservation `UPDATE ... WHERE incomes_created < 5 RETURNING` always finds its row.
-- Deleting a record never decrements a counter: a trigger refuses any decrease.
-- ---------------------------------------------------------------------------------------
CREATE TABLE user_usage (
  user_id          uuid        PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  incomes_created  integer     NOT NULL DEFAULT 0,
  expenses_created integer     NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT user_usage_counters_non_negative CHECK (incomes_created >= 0 AND expenses_created >= 0)
);
COMMENT ON TABLE user_usage IS 'Trial-cap counters (RLS on user_id). Monotonic; one row per user.';

-- user_usage_forbid_decrease(): BEFORE UPDATE trigger enforcing PRD §7 "deleting a record
-- does not give the trial record back". No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION user_usage_forbid_decrease() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog
AS $$
BEGIN
  IF NEW.incomes_created < OLD.incomes_created OR NEW.expenses_created < OLD.expenses_created THEN
    RAISE EXCEPTION 'user_usage counters never decrease'
      USING ERRCODE = 'P0001', HINT = 'Deleting a record does not return its trial slot (PRD 7).';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION user_usage_forbid_decrease() IS 'BEFORE UPDATE trigger: user_usage counters are monotonic.';

CREATE TRIGGER user_usage_forbid_decrease
  BEFORE UPDATE ON user_usage
  FOR EACH ROW EXECUTE FUNCTION user_usage_forbid_decrease();
CREATE TRIGGER user_usage_set_updated_at
  BEFORE UPDATE ON user_usage
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE user_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_usage FORCE ROW LEVEL SECURITY;
-- Intent: a user reads, creates and increments only their own counters. No DELETE policy
-- or grant: the row goes away only with its user (cascade).
CREATE POLICY user_usage_select_own ON user_usage FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY user_usage_insert_own ON user_usage FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY user_usage_update_own ON user_usage FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- INSERT only names the user: the counters always start at their zero defaults.
GRANT SELECT ON user_usage TO cc_app;
GRANT INSERT (user_id) ON user_usage TO cc_app;
GRANT UPDATE (incomes_created, expenses_created) ON user_usage TO cc_app;

-- ---------------------------------------------------------------------------------------
-- entitlements: what a user may do (PRD §7). One row per kind and user:
--   * trial  source registration, written by cc_app at sign-up, always ends (7 days).
--   * pro    source revenuecat, written by cc_webhook / sync from RevenueCat; carries the
--            store product and original transaction ids; always ends (renewal date).
--   * comp   source admin, granted by an operator as cc_owner (demo / QA accounts);
--            may be open-ended (ends_at NULL).
-- Each writer is limited to its own kind by RLS, so a compromised API pool cannot grant
-- itself pro, and the webhook cannot touch trial or comp rows.
-- ---------------------------------------------------------------------------------------
CREATE TABLE entitlements (
  id                         uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id                    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  kind                       text        NOT NULL,
  source                     text        NOT NULL,
  starts_at                  timestamptz NOT NULL,
  ends_at                    timestamptz,
  rc_product_id              text,
  rc_original_transaction_id text,
  created_at                 timestamptz NOT NULL DEFAULT now(),
  updated_at                 timestamptz NOT NULL DEFAULT now(),
  -- One trial / pro / comp row each; its btree also serves every lookup by user_id.
  CONSTRAINT entitlements_user_kind_key UNIQUE (user_id, kind),
  CONSTRAINT entitlements_kind_source CHECK (
    (kind, source) IN (('trial', 'registration'), ('pro', 'revenuecat'), ('comp', 'admin'))),
  CONSTRAINT entitlements_period_ordered CHECK (ends_at IS NULL OR ends_at > starts_at),
  CONSTRAINT entitlements_only_comp_open_ended CHECK (kind = 'comp' OR ends_at IS NOT NULL),
  CONSTRAINT entitlements_revenuecat_ids CHECK (
    CASE WHEN source = 'revenuecat'
      THEN coalesce(btrim(rc_product_id), '') <> '' AND coalesce(btrim(rc_original_transaction_id), '') <> ''
      ELSE rc_product_id IS NULL AND rc_original_transaction_id IS NULL
    END)
);
COMMENT ON TABLE entitlements IS
  'trial/pro/comp per user (RLS on user_id). cc_app writes trial, cc_webhook pro, cc_owner comp.';

CREATE TRIGGER entitlements_set_updated_at
  BEFORE UPDATE ON entitlements
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE entitlements ENABLE ROW LEVEL SECURITY;
ALTER TABLE entitlements FORCE ROW LEVEL SECURITY;
-- Intent: the API and the webhook read every entitlement of the context user (the access
-- policy needs all kinds). The webhook sets app.current_user_id to the event's
-- app_user_id (our user id) before touching this table.
CREATE POLICY entitlements_select_own ON entitlements FOR SELECT TO cc_app, cc_webhook
  USING (user_id = (SELECT app_user_id()));
-- Intent: the API only starts the user's own trial (registration).
CREATE POLICY entitlements_app_insert_trial ON entitlements FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()) AND kind = 'trial');
-- Intent: the webhook writes only the context user's pro row.
CREATE POLICY entitlements_webhook_insert_pro ON entitlements FOR INSERT TO cc_webhook
  WITH CHECK (user_id = (SELECT app_user_id()) AND kind = 'pro');
CREATE POLICY entitlements_webhook_update_pro ON entitlements FOR UPDATE TO cc_webhook
  USING (user_id = (SELECT app_user_id()) AND kind = 'pro')
  WITH CHECK (user_id = (SELECT app_user_id()) AND kind = 'pro');
CREATE POLICY entitlements_webhook_delete_pro ON entitlements FOR DELETE TO cc_webhook
  USING (user_id = (SELECT app_user_id()) AND kind = 'pro');
-- Intent: FORCE applies to the owner too; operators grant and revoke comp as cc_owner,
-- one user at a time under that user's context, and nothing else.
CREATE POLICY entitlements_owner_comp ON entitlements FOR ALL TO cc_owner
  USING (user_id = (SELECT app_user_id()) AND kind = 'comp')
  WITH CHECK (user_id = (SELECT app_user_id()) AND kind = 'comp');
GRANT SELECT ON entitlements TO cc_app, cc_webhook;
GRANT INSERT (user_id, kind, source, starts_at, ends_at) ON entitlements TO cc_app;
GRANT INSERT (user_id, kind, source, starts_at, ends_at, rc_product_id, rc_original_transaction_id)
  ON entitlements TO cc_webhook;
GRANT UPDATE (starts_at, ends_at, rc_product_id, rc_original_transaction_id) ON entitlements TO cc_webhook;
GRANT DELETE ON entitlements TO cc_webhook;

-- ---------------------------------------------------------------------------------------
-- email_verifications: pending sign-ups (ADR-0001 option A), one per email. Service-only:
-- no grant to any runtime role; the definer functions email_verification_upsert /
-- _consume (M1-03) are the only way in. Rows expire; the purge job deletes them.
-- pending holds the password hash, name and accepted document versions until the code
-- is verified; the email itself is not stored (the user re-sends it with the code).
-- ---------------------------------------------------------------------------------------
CREATE TABLE email_verifications (
  email_hash bytea       PRIMARY KEY,
  code_hash  bytea       NOT NULL,
  pending    jsonb       NOT NULL,
  attempts   smallint    NOT NULL DEFAULT 0,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT email_verifications_email_hash_size CHECK (octet_length(email_hash) = 32),
  -- sha256 of the 6-digit code.
  CONSTRAINT email_verifications_code_hash_size CHECK (octet_length(code_hash) = 32),
  CONSTRAINT email_verifications_pending_object CHECK (jsonb_typeof(pending) = 'object'),
  CONSTRAINT email_verifications_attempts_non_negative CHECK (attempts >= 0),
  CONSTRAINT email_verifications_expiry_after_creation CHECK (expires_at > created_at)
);
COMMENT ON TABLE email_verifications IS
  'Pending sign-ups keyed by email blind index. Service-only (definer functions); no runtime grant.';
-- Purge job range scan: DELETE ... WHERE expires_at < now().
CREATE INDEX email_verifications_expires_at_idx ON email_verifications (expires_at);

CREATE TRIGGER email_verifications_set_updated_at
  BEFORE UPDATE ON email_verifications
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Intent: deny-all to every runtime role (no grant, no policy), even after a mistaken
-- grant. Not FORCE: cc_owner's SECURITY DEFINER functions must reach the rows.
ALTER TABLE email_verifications ENABLE ROW LEVEL SECURITY;

-- +goose Down
DROP TABLE email_verifications;
DROP TABLE entitlements;
DROP TABLE user_usage;
DROP FUNCTION user_usage_forbid_decrease();
DROP TABLE users;
