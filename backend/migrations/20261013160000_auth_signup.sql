-- Auth sign-up (M1-05): what register / resendVerification need beyond M1-03.
--
-- D7 (MODULE-auth): a register for an email with a LIVE pending sign-up must never replace
-- its accepted documents; it behaves like a resend.
-- D8 (MODULE-auth): a pending sign-up holds NO password. The password is set at
-- verifyEmail by whoever proves inbox ownership; email_verification_start refuses a
-- payload carrying a password hash, so not even a buggy or compromised API can bring
-- the pre-registration takeover back. M1-03's
-- email_verification_upsert always replaces the whole row, so it cannot express that, and
-- nothing records when the last code was sent, so the 60 s silent resend cooldown has no
-- state. This migration:
--   * adds email_verifications.code_sent_at (the cooldown clock);
--   * adds two definer functions, the only sign-up writers the API may call from now on:
--       email_verification_start   register: start a new sign-up, replace an EXPIRED one,
--                                  or act as a resend for a live one (pending untouched);
--       email_verification_resend  resendVerification: new code for a live sign-up only;
--     both share the cooldown and answer 'unchanged' when they did nothing;
--   * revokes cc_app's EXECUTE on email_verification_upsert, so not even a compromised API
--     pool can overwrite a live pending sign-up (defense in depth for D7). The function
--     stays, owned by cc_auth_definer, for the M1-03 tests that use it as a fixture;
--   * makes users.name nullable: the contract's RegisterRequest.name is optional and
--     User.name is `[string, null]`, so an account may have no name. A present name is
--     still non-blank.
-- Same definer posture as M1-03: owned by cc_auth_definer, search_path pinned, every
-- reference schema-qualified, row scope bound to the argument's email hash.

-- +goose Up

ALTER TABLE users ALTER COLUMN name DROP NOT NULL;
ALTER TABLE users DROP CONSTRAINT users_name_present;
ALTER TABLE users ADD CONSTRAINT users_name_present CHECK (
  name IS NULL OR (btrim(name) <> '' AND char_length(name) <= 200));

ALTER TABLE email_verifications
  ADD COLUMN code_sent_at timestamptz NOT NULL DEFAULT now();
COMMENT ON COLUMN email_verifications.code_sent_at IS
  'When the current code was issued: the 60 s resend cooldown (shared by register and resend).';
-- INSERT leaves it to its default; restarts set it to now().
GRANT UPDATE (code_sent_at) ON email_verifications TO cc_auth_definer;

-- auth_check_code_args: the argument checks start and resend share with M1-03's upsert.
-- SECURITY INVOKER, executable by cc_auth_definer only.
-- +goose StatementBegin
CREATE FUNCTION auth_check_code_args(p_code_hash bytea, p_expires_at timestamptz) RETURNS void
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF p_code_hash IS NULL OR octet_length(p_code_hash) <> 32 THEN
    RAISE EXCEPTION 'code hash must be 32 bytes' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_expires_at IS NULL OR p_expires_at <= now() OR p_expires_at > now() + interval '1 hour' THEN
    RAISE EXCEPTION 'expires_at must be within the next hour' USING ERRCODE = 'invalid_parameter_value';
  END IF;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_check_code_args(bytea, timestamptz) IS
  'Validates a 32-byte code HMAC and an expiry within the next hour. Executable by cc_auth_definer only.';
REVOKE ALL ON FUNCTION auth_check_code_args(bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION auth_check_code_args(bytea, timestamptz) TO cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- email_verification_start: register. Returns
--   started    no pending sign-up, or only an expired one: the row now holds the new
--              code, pending payload and expiry (attempts 0);
--   resent     a live pending sign-up whose last code is at least 60 s old: new code and
--              expiry, attempts 0, the pending payload (accepted documents) KEPT;
--   unchanged  a live pending sign-up inside the cooldown (or one consumed concurrently):
--              nothing changed, nothing to send.
-- The insert races safely: ON CONFLICT DO NOTHING lets exactly one concurrent first call
-- create the row; every other call then locks it and takes the live-row branch, so no
-- concurrent register can replace the winner's payload.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION email_verification_start(
  p_email_hash bytea, p_code_hash bytea, p_pending jsonb, p_expires_at timestamptz
) RETURNS text
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  resend_cooldown CONSTANT interval := interval '60 seconds';
  row_found       public.email_verifications%ROWTYPE;
  outcome         text;
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  PERFORM public.auth_check_code_args(p_code_hash, p_expires_at);
  IF p_pending IS NULL OR jsonb_typeof(p_pending) <> 'object' THEN
    RAISE EXCEPTION 'pending must be a JSON object' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_pending ? 'password_hash' THEN
    RAISE EXCEPTION 'a pending sign-up holds no password (D8)' USING ERRCODE = 'invalid_parameter_value';
  END IF;

  INSERT INTO public.email_verifications (email_hash, code_hash, pending, expires_at)
  VALUES (p_email_hash, p_code_hash, p_pending, p_expires_at)
  ON CONFLICT (email_hash) DO NOTHING;

  IF FOUND THEN
    outcome := 'started';
  ELSE
    SELECT v.* INTO row_found
    FROM public.email_verifications AS v
    WHERE v.email_hash = p_email_hash
    FOR UPDATE;

    IF NOT FOUND OR (row_found.expires_at > now() AND row_found.code_sent_at > now() - resend_cooldown) THEN
      outcome := 'unchanged';
    ELSIF row_found.expires_at <= now() THEN
      UPDATE public.email_verifications AS v
      SET code_hash = p_code_hash, pending = p_pending, attempts = 0,
          expires_at = p_expires_at, code_sent_at = now()
      WHERE v.email_hash = p_email_hash;
      outcome := 'started';
    ELSE
      UPDATE public.email_verifications AS v
      SET code_hash = p_code_hash, attempts = 0, expires_at = p_expires_at, code_sent_at = now()
      WHERE v.email_hash = p_email_hash;
      outcome := 'resent';
    END IF;
  END IF;

  PERFORM set_config('app.auth_email_hash', '', true);
  RETURN outcome;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION email_verification_start(bytea, bytea, jsonb, timestamptz) IS
  'Register: started|resent|unchanged. Never replaces a live pending payload (D7); refuses a password hash in it (D8); 60 s cooldown. Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- email_verification_resend: resendVerification. 'resent' (new code and expiry, attempts
-- 0, payload kept) for a live pending sign-up outside the cooldown; 'unchanged' for none,
-- an expired one (register again) or one inside the cooldown.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION email_verification_resend(
  p_email_hash bytea, p_code_hash bytea, p_expires_at timestamptz
) RETURNS text
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  resend_cooldown CONSTANT interval := interval '60 seconds';
  row_found       public.email_verifications%ROWTYPE;
  outcome         text := 'unchanged';
BEGIN
  PERFORM public.auth_bind_email_hash(p_email_hash);
  PERFORM public.auth_check_code_args(p_code_hash, p_expires_at);

  SELECT v.* INTO row_found
  FROM public.email_verifications AS v
  WHERE v.email_hash = p_email_hash
  FOR UPDATE;

  IF FOUND AND row_found.expires_at > now() AND row_found.code_sent_at <= now() - resend_cooldown THEN
    UPDATE public.email_verifications AS v
    SET code_hash = p_code_hash, attempts = 0, expires_at = p_expires_at, code_sent_at = now()
    WHERE v.email_hash = p_email_hash;
    outcome := 'resent';
  END IF;

  PERFORM set_config('app.auth_email_hash', '', true);
  RETURN outcome;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION email_verification_resend(bytea, bytea, timestamptz) IS
  'Resend: resent|unchanged. Live pending sign-ups only, payload kept, 60 s cooldown. Definer: cc_auth_definer; EXECUTE: cc_app.';

REVOKE ALL ON FUNCTION
  email_verification_start(bytea, bytea, jsonb, timestamptz),
  email_verification_resend(bytea, bytea, timestamptz)
  FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
  email_verification_start(bytea, bytea, jsonb, timestamptz),
  email_verification_resend(bytea, bytea, timestamptz)
  TO cc_app;

-- D7 defense in depth: the API can no longer replace a pending payload at all. Only the
-- function's owner may revoke its grant; cc_owner acts as it (SET, not INHERIT).
SET LOCAL ROLE cc_auth_definer;
REVOKE EXECUTE ON FUNCTION email_verification_upsert(bytea, bytea, jsonb, timestamptz) FROM cc_app;
RESET ROLE;

-- Hand the functions to the definer role (CREATE on the schema lent for the transfer).
GRANT CREATE ON SCHEMA public TO cc_auth_definer;
ALTER FUNCTION auth_check_code_args(bytea, timestamptz) OWNER TO cc_auth_definer;
ALTER FUNCTION email_verification_start(bytea, bytea, jsonb, timestamptz) OWNER TO cc_auth_definer;
ALTER FUNCTION email_verification_resend(bytea, bytea, timestamptz) OWNER TO cc_auth_definer;
REVOKE CREATE ON SCHEMA public FROM cc_auth_definer;

-- +goose Down
-- Restoring the grant needs the function's owner, as does dropping the new functions.
SET LOCAL ROLE cc_auth_definer;
GRANT EXECUTE ON FUNCTION email_verification_upsert(bytea, bytea, jsonb, timestamptz) TO cc_app;
DROP FUNCTION email_verification_resend(bytea, bytea, timestamptz);
DROP FUNCTION email_verification_start(bytea, bytea, jsonb, timestamptz);
DROP FUNCTION auth_check_code_args(bytea, timestamptz);
RESET ROLE;
REVOKE UPDATE (code_sent_at) ON email_verifications FROM cc_auth_definer;
ALTER TABLE email_verifications DROP COLUMN code_sent_at;
-- Fails while a user has no name: set one first (no data is discarded silently).
ALTER TABLE users DROP CONSTRAINT users_name_present;
ALTER TABLE users ADD CONSTRAINT users_name_present CHECK (btrim(name) <> '' AND char_length(name) <= 200);
ALTER TABLE users ALTER COLUMN name SET NOT NULL;
