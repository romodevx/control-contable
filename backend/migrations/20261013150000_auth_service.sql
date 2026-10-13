-- Auth service (M1-03): the global legal_documents catalog, the service-only
-- trial_registry and the SECURITY DEFINER functions that are the only way the API
-- reaches credentials, the trial registry and pending sign-ups.
--
-- Definer design (fixes the P0-30 finding "a definer function owned by cc_owner sees no
-- users rows, because FORCE RLS has no owner policy"):
--   * The functions are owned by cc_auth_definer, a NOLOGIN, NOBYPASSRLS role created by
--     the bootstrap (cc_owner has NOCREATEROLE). It is not a table owner, so RLS always
--     applies to it, FORCE or not.
--   * Its reach is narrowed twice: column grants (only what each function reads or
--     writes) and per-table policies keyed on app.auth_email_hash, a transaction-local
--     setting each function binds to its own argument before touching a table and clears
--     afterwards. Inside a call the role sees only the row of the email being processed.
--   * EXECUTE goes to cc_app only. search_path is pinned to pg_catalog, public, pg_temp
--     (pg_temp last, so a temp table can never shadow a table) and every reference is
--     schema-qualified as well.
--   * cc_owner is a member of cc_auth_definer WITH INHERIT FALSE, SET TRUE (bootstrap):
--     ALTER ... OWNER requires that, and inheriting nothing keeps the definer policies
--     away from cc_owner. cc_owner already controls every table, so SET ROLE adds no reach.
-- HMACs (email blind index, verification code) are computed app-side with server keys;
-- the database stores and compares only 32-byte digests.

-- +goose Up

-- Fail early, with a clear message, on an environment whose bootstrap predates M1-03.
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_roles
    WHERE rolname = 'cc_auth_definer' AND NOT rolcanlogin AND NOT rolsuper AND NOT rolbypassrls
  ) THEN
    RAISE EXCEPTION 'role cc_auth_definer is missing or privileged'
      USING HINT = 'Create it as in infra/local/postgres/bootstrap/roles.sql (make db-reset locally).';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_auth_members
    WHERE roleid = 'cc_auth_definer'::regrole AND member = 'cc_owner'::regrole
      AND set_option AND NOT inherit_option
  ) THEN
    RAISE EXCEPTION 'cc_owner must be a member of cc_auth_definer WITH INHERIT FALSE, SET TRUE'
      USING HINT = 'See infra/local/postgres/bootstrap/roles.sql.';
  END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- legal_documents: versions of the privacy notices and the terms (global configuration,
-- no owner). consents (M1-02) reference (kind, version). The v1 rows below are
-- PLACEHOLDERS pending the lawyer's review (PRD Q10): their URLs use the reserved
-- .invalid TLD and review_status says so; an approved row must carry the sha256 of the
-- exact text users accept. "Current" = the latest effective_at <= now() per kind; the
-- UNIQUE (kind, effective_at) keeps that unambiguous. A handful of rows: the PK and the
-- unique btree cover every lookup.
-- ---------------------------------------------------------------------------------------
CREATE TABLE legal_documents (
  kind           text        NOT NULL,
  version        text        NOT NULL,
  url            text        NOT NULL,
  content_sha256 bytea,
  effective_at   timestamptz NOT NULL,
  review_status  text        NOT NULL DEFAULT 'pending_legal_review',
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (kind, version),
  CONSTRAINT legal_documents_kind_effective_key UNIQUE (kind, effective_at),
  -- docs/api/openapi.yaml LegalDocumentKind.
  CONSTRAINT legal_documents_kind_valid CHECK (
    kind IN ('privacy_notice_integral', 'privacy_notice_simplified', 'terms')),
  CONSTRAINT legal_documents_version_format CHECK (version ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'),
  CONSTRAINT legal_documents_url_https CHECK (url ~ '^https://[^[:space:]]+$' AND char_length(url) <= 2048),
  CONSTRAINT legal_documents_content_sha256_size CHECK (
    content_sha256 IS NULL OR octet_length(content_sha256) = 32),
  CONSTRAINT legal_documents_review_status_valid CHECK (
    review_status IN ('pending_legal_review', 'approved')),
  -- coalesce: a NULL hash would otherwise pass the CHECK.
  CONSTRAINT legal_documents_approved_has_hash CHECK (
    review_status <> 'approved' OR coalesce(octet_length(content_sha256), 0) = 32)
);
COMMENT ON TABLE legal_documents IS
  'Privacy notice / terms versions (global). Read by cc_app; written only by migrations as cc_owner.';

CREATE TRIGGER legal_documents_set_updated_at
  BEFORE UPDATE ON legal_documents
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Intent: global rows, so ENABLE (not FORCE, cc_owner writes them) with one read policy
-- for cc_app; no write policy exists for any runtime role.
ALTER TABLE legal_documents ENABLE ROW LEVEL SECURITY;
CREATE POLICY legal_documents_app_read ON legal_documents
  FOR SELECT TO cc_app
  USING (true);
GRANT SELECT ON legal_documents TO cc_app;

INSERT INTO legal_documents (kind, version, url, effective_at) VALUES
  ('privacy_notice_integral',   'v1', 'https://control-contable.invalid/legal/aviso-de-privacidad-integral',   '2026-10-13T00:00:00Z'),
  ('privacy_notice_simplified', 'v1', 'https://control-contable.invalid/legal/aviso-de-privacidad-simplificado', '2026-10-13T00:00:00Z'),
  ('terms',                     'v1', 'https://control-contable.invalid/legal/terminos',                      '2026-10-13T00:00:00Z');

-- consents (M1-02) now reference the exact version accepted. The FK supersedes M1-02's
-- provisional kind CHECK ('privacy', 'terms'), which predates the contract's three
-- LegalDocumentKind values. RESTRICT: an accepted version is evidence and is never
-- deleted or re-keyed. Referenced-side deletes are migration-only and rare, so no extra
-- index on consents (document_kind, document_version).
ALTER TABLE consents DROP CONSTRAINT consents_document_kind_valid;
ALTER TABLE consents ADD CONSTRAINT consents_legal_document_fkey
  FOREIGN KEY (document_kind, document_version) REFERENCES legal_documents (kind, version)
  ON UPDATE RESTRICT ON DELETE RESTRICT;

-- ---------------------------------------------------------------------------------------
-- trial_registry: one row per email that ever started a trial, keyed by the same HMAC
-- blind index as users.email_hash (never the email). No FK to users on purpose: it
-- survives account deletion so delete-and-re-register cannot farm trials. Append-only:
-- first_trial_at is the creation stamp, and no role may update a row (no updated_at).
-- ---------------------------------------------------------------------------------------
CREATE TABLE trial_registry (
  email_hash     bytea       PRIMARY KEY,
  first_trial_at timestamptz NOT NULL DEFAULT now(),
  -- 32 = crypto.HashSize (HMAC-SHA256).
  CONSTRAINT trial_registry_email_hash_size CHECK (octet_length(email_hash) = 32)
);
COMMENT ON TABLE trial_registry IS
  'Emails (HMAC) that already had a trial; survives account deletion. Service-only (definer function).';

-- ---------------------------------------------------------------------------------------
-- Definer row scope. auth_bind_email_hash validates a 32-byte digest and binds it to
-- app.auth_email_hash for the rest of the transaction; auth_email_hash() reads it back
-- (NULL when unset or cleared, so the policies below match nothing). Both are SECURITY
-- INVOKER and executable by cc_auth_definer only.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION auth_bind_email_hash(p_email_hash bytea) RETURNS void
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF p_email_hash IS NULL OR octet_length(p_email_hash) <> 32 THEN
    RAISE EXCEPTION 'email hash must be 32 bytes' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  PERFORM set_config('app.auth_email_hash', encode(p_email_hash, 'hex'), true);
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_bind_email_hash(bytea) IS
  'Validates a 32-byte email HMAC and scopes cc_auth_definer policies to it (transaction-local).';

-- +goose StatementBegin
CREATE FUNCTION auth_email_hash() RETURNS bytea
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT decode(nullif(current_setting('app.auth_email_hash', true), ''), 'hex') $$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_email_hash() IS
  'Email HMAC bound by the running auth definer function; NULL outside one. Policy key.';

GRANT EXECUTE ON FUNCTION auth_bind_email_hash(bytea), auth_email_hash() TO cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- Definer policies and grants. Each policy is TO cc_auth_definer only and matches only
-- the bound email's row.
-- ---------------------------------------------------------------------------------------

-- users (FORCE RLS): login reads id, password hash and status of the bound email.
CREATE POLICY users_auth_definer_lookup ON users FOR SELECT TO cc_auth_definer
  USING (email_hash = (SELECT auth_email_hash()));
GRANT SELECT (id, email_hash, password_hash, status) ON users TO cc_auth_definer;

-- trial_registry: the definer inserts the bound email and may see only that row (the
-- ON CONFLICT arbiter needs SELECT on email_hash). No UPDATE or DELETE: append-only.
-- Intent: ENABLE, not FORCE (ownerless service table, as email_verifications); no
-- runtime role has a grant or a policy, so it is deny-all to them.
ALTER TABLE trial_registry ENABLE ROW LEVEL SECURITY;
CREATE POLICY trial_registry_auth_definer_read ON trial_registry FOR SELECT TO cc_auth_definer
  USING (email_hash = (SELECT auth_email_hash()));
CREATE POLICY trial_registry_auth_definer_insert ON trial_registry FOR INSERT TO cc_auth_definer
  WITH CHECK (email_hash = (SELECT auth_email_hash()));
GRANT SELECT (email_hash), INSERT (email_hash) ON trial_registry TO cc_auth_definer;

-- email_verifications: the definer reads, writes and deletes the bound email's row.
-- attempts and created_at are not insertable: a new row always starts at 0 attempts.
CREATE POLICY email_verifications_auth_definer ON email_verifications FOR ALL TO cc_auth_definer
  USING (email_hash = (SELECT auth_email_hash()))
  WITH CHECK (email_hash = (SELECT auth_email_hash()));
GRANT SELECT, DELETE ON email_verifications TO cc_auth_definer;
GRANT INSERT (email_hash, code_hash, pending, expires_at) ON email_verifications TO cc_auth_definer;
GRANT UPDATE (code_hash, pending, attempts, expires_at) ON email_verifications TO cc_auth_definer;

-- email_verification_failures: wrong codes per email over a 24 h window. A resend (or a
-- new register) resets the per-code budget of email_verifications, so without this cap
-- register -> 5 guesses -> register ... would brute-force the 6-digit code of an inbox
-- the caller does not own. Kept apart from email_verifications so expiry, deletion or
-- replacement of the pending row never resets it; the purge job drops rows whose window
-- ended. Same posture as trial_registry: ENABLE (not FORCE), cc_auth_definer only.
CREATE TABLE email_verification_failures (
  email_hash        bytea       PRIMARY KEY,
  failures          smallint    NOT NULL DEFAULT 1,
  window_started_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT email_verification_failures_email_hash_size CHECK (octet_length(email_hash) = 32),
  CONSTRAINT email_verification_failures_failures_positive CHECK (failures > 0)
);
COMMENT ON TABLE email_verification_failures IS
  'Wrong verification codes per email HMAC in a 24 h window (cap across resends). Service-only (definer function).';
ALTER TABLE email_verification_failures ENABLE ROW LEVEL SECURITY;
CREATE POLICY email_verification_failures_auth_definer ON email_verification_failures
  FOR ALL TO cc_auth_definer
  USING (email_hash = (SELECT auth_email_hash()))
  WITH CHECK (email_hash = (SELECT auth_email_hash()));
GRANT SELECT, INSERT (email_hash), UPDATE (failures, window_started_at)
  ON email_verification_failures TO cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- auth_lookup_credentials: login. Returns the one user whose email_hash matches, with
-- the minimum login needs (no email ciphertext, no name). No row = unknown email; the
-- app answers it like a wrong password (dummy argon2 verify) to avoid enumeration.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION auth_lookup_credentials(p_email_hash bytea)
  RETURNS TABLE (id uuid, password_hash text, status text)
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  RETURN QUERY
    SELECT u.id, u.password_hash, u.status
    FROM public.users AS u
    WHERE u.email_hash = p_email_hash;
  PERFORM set_config('app.auth_email_hash', '', true);
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_lookup_credentials(bytea) IS
  'Login lookup by email HMAC: (id, argon2id PHC, status) or no row. Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- trial_registry_check_and_insert: true exactly once per email, ever (the first trial);
-- false afterwards, including after account deletion. Concurrent first calls are
-- serialized by the primary key: one inserts, the other sees the conflict. Call it in
-- the sign-up transaction, so a failed sign-up does not burn the trial.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION trial_registry_check_and_insert(p_email_hash bytea) RETURNS boolean
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  inserted bigint;
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  INSERT INTO public.trial_registry (email_hash) VALUES (p_email_hash)
    ON CONFLICT (email_hash) DO NOTHING;
  GET DIAGNOSTICS inserted = ROW_COUNT;
  PERFORM set_config('app.auth_email_hash', '', true);
  RETURN inserted = 1;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION trial_registry_check_and_insert(bytea) IS
  'True on the first trial of an email HMAC, false ever after. Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- email_verification_upsert: starts (or restarts, on resend) a pending sign-up. A
-- restart replaces the code, the pending payload and the expiry and resets attempts:
-- the old code stops working. expires_at must fall within the next hour, so a faulty
-- or compromised caller cannot park a code for long. code_hash is the app-side HMAC of
-- the 6-digit code under a server key.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION email_verification_upsert(
  p_email_hash bytea, p_code_hash bytea, p_pending jsonb, p_expires_at timestamptz
) RETURNS void
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  IF p_code_hash IS NULL OR octet_length(p_code_hash) <> 32 THEN
    RAISE EXCEPTION 'code hash must be 32 bytes' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_pending IS NULL OR jsonb_typeof(p_pending) <> 'object' THEN
    RAISE EXCEPTION 'pending must be a JSON object' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_expires_at IS NULL OR p_expires_at <= now() OR p_expires_at > now() + interval '1 hour' THEN
    RAISE EXCEPTION 'expires_at must be within the next hour' USING ERRCODE = 'invalid_parameter_value';
  END IF;

  INSERT INTO public.email_verifications AS v (email_hash, code_hash, pending, expires_at)
  VALUES (p_email_hash, p_code_hash, p_pending, p_expires_at)
  ON CONFLICT (email_hash) DO UPDATE
    SET code_hash  = excluded.code_hash,
        pending    = excluded.pending,
        attempts   = 0,
        expires_at = excluded.expires_at;
  PERFORM set_config('app.auth_email_hash', '', true);
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION email_verification_upsert(bytea, bytea, jsonb, timestamptz) IS
  'Start or restart a pending sign-up (new code, attempts = 0, expiry <= 1 h). Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- email_verification_consume: checks a code. Always returns one row:
--   verified  the code matches: the row is deleted (single use) and pending returned.
--   invalid   no pending sign-up for the email, or a wrong code (attempts + 1).
--   expired   past expires_at: the row is deleted.
--   locked    5 wrong attempts on this code (until a resend), or 15 wrong codes for the
--             email within 24 h (resends do not help): even the right code is refused.
-- The row is locked FOR UPDATE, so concurrent guesses are counted exactly.
-- CALLER CONTRACT: commit the transaction on every outcome but "verified" (do not roll
-- it back), or the attempt counter is undone. On "verified", create the user in the
-- same transaction: if that fails, the rollback restores the pending row.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION email_verification_consume(p_email_hash bytea, p_code_hash bytea)
  RETURNS TABLE (outcome text, pending jsonb)
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_attempts         CONSTANT smallint := 5;
  max_window_failures  CONSTANT smallint := 15;
  failure_window       CONSTANT interval := interval '24 hours';
  row_found            public.email_verifications%ROWTYPE;
  pending_found        boolean;
  window_failures      smallint;
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  IF p_code_hash IS NULL OR octet_length(p_code_hash) <> 32 THEN
    RAISE EXCEPTION 'code hash must be 32 bytes' USING ERRCODE = 'invalid_parameter_value';
  END IF;

  SELECT v.* INTO row_found
  FROM public.email_verifications AS v
  WHERE v.email_hash = p_email_hash
  FOR UPDATE;

  pending_found := FOUND;

  SELECT f.failures INTO window_failures
  FROM public.email_verification_failures AS f
  WHERE f.email_hash = p_email_hash AND f.window_started_at > now() - failure_window;

  IF NOT pending_found THEN
    outcome := 'invalid';
  ELSIF row_found.expires_at <= now() THEN
    DELETE FROM public.email_verifications AS v WHERE v.email_hash = p_email_hash;
    outcome := 'expired';
  ELSIF row_found.attempts >= max_attempts OR coalesce(window_failures, 0) >= max_window_failures THEN
    outcome := 'locked';
  ELSIF row_found.code_hash = p_code_hash THEN
    DELETE FROM public.email_verifications AS v WHERE v.email_hash = p_email_hash;
    outcome := 'verified';
    pending := row_found.pending;
  ELSE
    UPDATE public.email_verifications AS v SET attempts = v.attempts + 1
    WHERE v.email_hash = p_email_hash;
    -- Serialized per email by the FOR UPDATE lock above.
    INSERT INTO public.email_verification_failures AS f (email_hash) VALUES (p_email_hash)
    ON CONFLICT (email_hash) DO UPDATE
      SET failures          = CASE WHEN f.window_started_at > now() - failure_window
                                   THEN f.failures + 1 ELSE 1 END,
          window_started_at = CASE WHEN f.window_started_at > now() - failure_window
                                   THEN f.window_started_at ELSE now() END
    RETURNING f.failures INTO window_failures;
    outcome := CASE
      WHEN row_found.attempts + 1 >= max_attempts OR window_failures >= max_window_failures THEN 'locked'
      ELSE 'invalid' END;
  END IF;

  PERFORM set_config('app.auth_email_hash', '', true);
  RETURN NEXT;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION email_verification_consume(bytea, bytea) IS
  'Check a verification code: verified|invalid|expired|locked (5 attempts per code, 15 per email per 24 h, single use). Commit non-verified outcomes. Definer: cc_auth_definer; EXECUTE: cc_app.';

-- EXECUTE: cc_app only. Granted while cc_owner still owns the functions; ALTER OWNER
-- then rewrites the grantor to the new owner.
REVOKE ALL ON FUNCTION
  auth_lookup_credentials(bytea),
  trial_registry_check_and_insert(bytea),
  email_verification_upsert(bytea, bytea, jsonb, timestamptz),
  email_verification_consume(bytea, bytea)
  FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
  auth_lookup_credentials(bytea),
  trial_registry_check_and_insert(bytea),
  email_verification_upsert(bytea, bytea, jsonb, timestamptz),
  email_verification_consume(bytea, bytea)
  TO cc_app;

-- Hand the functions to the definer role. ALTER OWNER needs the new owner to hold CREATE
-- on the schema: lent for these four statements only.
GRANT CREATE ON SCHEMA public TO cc_auth_definer;
ALTER FUNCTION auth_lookup_credentials(bytea) OWNER TO cc_auth_definer;
ALTER FUNCTION trial_registry_check_and_insert(bytea) OWNER TO cc_auth_definer;
ALTER FUNCTION email_verification_upsert(bytea, bytea, jsonb, timestamptz) OWNER TO cc_auth_definer;
ALTER FUNCTION email_verification_consume(bytea, bytea) OWNER TO cc_auth_definer;
REVOKE CREATE ON SCHEMA public FROM cc_auth_definer;

-- +goose Down
-- Dropping a function takes its owner; cc_owner does not inherit it, so it acts as it.
SET LOCAL ROLE cc_auth_definer;
DROP FUNCTION email_verification_consume(bytea, bytea);
DROP FUNCTION email_verification_upsert(bytea, bytea, jsonb, timestamptz);
DROP FUNCTION trial_registry_check_and_insert(bytea);
DROP FUNCTION auth_lookup_credentials(bytea);
RESET ROLE;
REVOKE SELECT, INSERT, UPDATE, DELETE ON email_verifications FROM cc_auth_definer;
DROP POLICY email_verifications_auth_definer ON email_verifications;
REVOKE SELECT ON users FROM cc_auth_definer;
DROP POLICY users_auth_definer_lookup ON users;
DROP TABLE email_verification_failures;
DROP TABLE trial_registry;
DROP FUNCTION auth_email_hash();
DROP FUNCTION auth_bind_email_hash(bytea);
ALTER TABLE consents DROP CONSTRAINT consents_legal_document_fkey;
-- NOT VALID: rows accepted under the contract kinds may exist; new rows are checked.
ALTER TABLE consents ADD CONSTRAINT consents_document_kind_valid
  CHECK (document_kind IN ('privacy', 'terms')) NOT VALID;
DROP TABLE legal_documents;
REVOKE USAGE ON SCHEMA public FROM cc_auth_definer;
