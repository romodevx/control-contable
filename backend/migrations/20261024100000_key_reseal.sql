-- Key re-seal (ADR-0012, docs/runbooks/key-rotation.md): what the worker's key_reseal job
-- needs to move every sealed column to the active key, and the hash version columns a
-- later pepper rotation needs.
--
-- Sealed columns (AES-256-GCM, crypto.Keyring frame: key_id (2 bytes) || nonce || ct || tag,
-- and a <col>_key_id column the framing CHECK pins to that prefix):
--   users.email_encrypted (email_key_id), taxpayers.rfc_encrypted (rfc_key_id),
--   devices.token_ciphertext (token_key_id). auth_email_outbox.payload is sealed too, but its
--   rows live 15 minutes: a rotation keeps the retired key until they drain (no re-seal).
--
-- Cross-user reads go through SECURITY DEFINER functions owned by cc_worker_definer (the
-- NOLOGIN bootstrap role), EXECUTE for cc_worker only. They return ids and counts, never a
-- ciphertext or a hash:
--   * reseal_pending_{users,taxpayers,devices}(active key id, after id, limit): the next
--     page (keyset on id, 1..1000 rows) of rows sealed with any key but the active one, with
--     the owning user's id. The row's key id is the job's checkpoint: a re-sealed row leaves
--     the list, so a run that stops anywhere resumes by calling again.
--   * key_usage(): rows per key id of each sealed column and per hash version of each blind
--     index (the readiness figure the job logs: rows still on an old key or pepper).
-- cc_worker_definer already sees every users.id (users_worker_definer_ids) and every device
-- (devices_due_reminders); on taxpayers its policy shows live rows only, which
-- due_reminder_users relies on. The re-seal must reach archived taxpayers too, so the new
-- policy taxpayers_key_reseal applies only while app.key_reseal_scan is 'on': the functions
-- set it transaction-locally and clear it before they return, and nothing else sets it.
--
-- The rewrite itself runs as cc_worker in one transaction per row under the owner's
-- context (app.current_user_id, tx.Manager), so RLS scopes it like cc_app's own writes:
-- SELECT ... FOR UPDATE of the row, then UPDATE of the ciphertext and its key id only (the
-- framing CHECK keeps the two in step). cc_worker gets nothing else: never a hash, a hash
-- version, a name, an INSERT or a DELETE it did not already have, and still no row without a
-- user context.
--
-- Hash versions: email_hash, rfc_hash and token_hmac are HMAC-SHA256 under a pepper. Their
-- <hash>_version column (1 today) names the pepper that produced each value, so a later
-- pepper rotation can re-hash in place and look up both versions meanwhile. No runtime role
-- may write it yet: every row takes the default.

-- +goose Up

-- Fail early, with a clear message, on an environment whose bootstrap lacks the definer role.
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_roles
    WHERE rolname = 'cc_worker_definer' AND NOT rolcanlogin AND NOT rolsuper AND NOT rolbypassrls
  ) THEN
    RAISE EXCEPTION 'role cc_worker_definer is missing or privileged'
      USING HINT = 'Create it as in infra/local/postgres/bootstrap/roles.sql (docs/runbooks/records-purge.md §2).';
  END IF;
END
$$;
-- +goose StatementEnd

-- Hash versions (a constant default: no table rewrite).
ALTER TABLE users ADD COLUMN email_hash_version smallint NOT NULL DEFAULT 1
  CONSTRAINT users_email_hash_version_range CHECK (email_hash_version >= 1);
ALTER TABLE taxpayers ADD COLUMN rfc_hash_version smallint NOT NULL DEFAULT 1
  CONSTRAINT taxpayers_rfc_hash_version_range CHECK (rfc_hash_version >= 1);
ALTER TABLE devices ADD COLUMN token_hmac_version smallint NOT NULL DEFAULT 1
  CONSTRAINT devices_token_hmac_version_range CHECK (token_hmac_version >= 1);
COMMENT ON COLUMN users.email_hash_version IS 'Pepper version of email_hash (ADR-0012). 1 = AUTH_EMAIL_PEPPER.';
COMMENT ON COLUMN taxpayers.rfc_hash_version IS 'Pepper version of rfc_hash (ADR-0012). 1 = RFC_INDEX_PEPPER.';
COMMENT ON COLUMN devices.token_hmac_version IS 'Pepper version of token_hmac (ADR-0012). 1 = PUSH_TOKEN_INDEX_PEPPER.';

-- The scan flag of taxpayers_key_reseal. Not SECURITY DEFINER: the policy evaluates it as
-- cc_worker_definer.
-- +goose StatementBegin
CREATE FUNCTION key_reseal_scan() RETURNS boolean
  LANGUAGE sql STABLE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT coalesce(current_setting('app.key_reseal_scan', true), '') = 'on' $$;
-- +goose StatementEnd
COMMENT ON FUNCTION key_reseal_scan() IS
  'True only inside a key re-seal function (ADR-0012). Policy key of taxpayers_key_reseal.';
REVOKE ALL ON FUNCTION key_reseal_scan() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION key_reseal_scan() TO cc_worker_definer;

CREATE POLICY taxpayers_key_reseal ON taxpayers FOR SELECT TO cc_worker_definer
  USING ((SELECT key_reseal_scan()));
GRANT SELECT (id, rfc_key_id, rfc_hash_version) ON taxpayers TO cc_worker_definer;
GRANT SELECT (email_key_id, email_hash_version) ON users TO cc_worker_definer;
GRANT SELECT (id, token_key_id, token_hmac_version) ON devices TO cc_worker_definer;

-- +goose StatementBegin
CREATE FUNCTION reseal_pending_users(p_active_key_id smallint, p_after uuid, p_limit integer)
  RETURNS TABLE (id uuid, user_id uuid)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  IF p_active_key_id IS NULL OR p_active_key_id < 1 OR p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 1000 THEN
    RAISE EXCEPTION 'an active key id >= 1 and a limit in [1, 1000] are required';
  END IF;
  RETURN QUERY
    SELECT u.id, u.id FROM public.users AS u
    WHERE u.email_key_id <> p_active_key_id AND (p_after IS NULL OR u.id > p_after)
    ORDER BY u.id LIMIT p_limit;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION reseal_pending_taxpayers(p_active_key_id smallint, p_after uuid, p_limit integer)
  RETURNS TABLE (id uuid, user_id uuid)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  IF p_active_key_id IS NULL OR p_active_key_id < 1 OR p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 1000 THEN
    RAISE EXCEPTION 'an active key id >= 1 and a limit in [1, 1000] are required';
  END IF;
  PERFORM set_config('app.key_reseal_scan', 'on', true);
  RETURN QUERY
    SELECT t.id, t.user_id FROM public.taxpayers AS t
    WHERE t.rfc_key_id <> p_active_key_id AND (p_after IS NULL OR t.id > p_after)
    ORDER BY t.id LIMIT p_limit;
  PERFORM set_config('app.key_reseal_scan', '', true);
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION reseal_pending_devices(p_active_key_id smallint, p_after uuid, p_limit integer)
  RETURNS TABLE (id uuid, user_id uuid)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  IF p_active_key_id IS NULL OR p_active_key_id < 1 OR p_limit IS NULL OR p_limit NOT BETWEEN 1 AND 1000 THEN
    RAISE EXCEPTION 'an active key id >= 1 and a limit in [1, 1000] are required';
  END IF;
  RETURN QUERY
    SELECT d.id, d.user_id FROM public.devices AS d
    WHERE d.token_key_id <> p_active_key_id AND (p_after IS NULL OR d.id > p_after)
    ORDER BY d.id LIMIT p_limit;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION key_usage()
  RETURNS TABLE (target text, version smallint, row_count bigint)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  PERFORM set_config('app.key_reseal_scan', 'on', true);
  RETURN QUERY
    SELECT x.target, x.version, x.row_count FROM (
      SELECT 'users.email_encrypted'::text AS target, u.email_key_id AS version, count(*) AS row_count
        FROM public.users AS u GROUP BY u.email_key_id
      UNION ALL
      SELECT 'users.email_hash', u.email_hash_version, count(*) FROM public.users AS u GROUP BY u.email_hash_version
      UNION ALL
      SELECT 'taxpayers.rfc_encrypted', t.rfc_key_id, count(*) FROM public.taxpayers AS t GROUP BY t.rfc_key_id
      UNION ALL
      SELECT 'taxpayers.rfc_hash', t.rfc_hash_version, count(*) FROM public.taxpayers AS t GROUP BY t.rfc_hash_version
      UNION ALL
      SELECT 'devices.token_ciphertext', d.token_key_id, count(*) FROM public.devices AS d GROUP BY d.token_key_id
      UNION ALL
      SELECT 'devices.token_hmac', d.token_hmac_version, count(*) FROM public.devices AS d GROUP BY d.token_hmac_version
    ) AS x
    ORDER BY x.target, x.version;
  PERFORM set_config('app.key_reseal_scan', '', true);
END
$$;
-- +goose StatementEnd

COMMENT ON FUNCTION reseal_pending_users(smallint, uuid, integer) IS
  'ADR-0012: next page of users whose email is sealed with another key than the active one. Ids only.';
COMMENT ON FUNCTION reseal_pending_taxpayers(smallint, uuid, integer) IS
  'ADR-0012: next page of taxpayers (archived too) whose RFC is sealed with another key. Ids only.';
COMMENT ON FUNCTION reseal_pending_devices(smallint, uuid, integer) IS
  'ADR-0012: next page of devices whose push token is sealed with another key. Ids only.';
COMMENT ON FUNCTION key_usage() IS
  'ADR-0012: rows per key id of each sealed column and per pepper version of each blind index. Counts only.';

-- Grants first: once the definer owns them, cc_owner (not inheriting) may not grant on them.
REVOKE ALL ON FUNCTION reseal_pending_users(smallint, uuid, integer), reseal_pending_taxpayers(smallint, uuid, integer),
  reseal_pending_devices(smallint, uuid, integer), key_usage() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION reseal_pending_users(smallint, uuid, integer), reseal_pending_taxpayers(smallint, uuid, integer),
  reseal_pending_devices(smallint, uuid, integer), key_usage() TO cc_worker;
-- ALTER OWNER needs the new owner to hold CREATE on the schema: lent for these statements only.
GRANT CREATE ON SCHEMA public TO cc_worker_definer;
ALTER FUNCTION reseal_pending_users(smallint, uuid, integer) OWNER TO cc_worker_definer;
ALTER FUNCTION reseal_pending_taxpayers(smallint, uuid, integer) OWNER TO cc_worker_definer;
ALTER FUNCTION reseal_pending_devices(smallint, uuid, integer) OWNER TO cc_worker_definer;
ALTER FUNCTION key_usage() OWNER TO cc_worker_definer;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

-- cc_worker: the per-row rewrite, own rows only, the ciphertext and its key id only.
CREATE POLICY users_select_worker ON users FOR SELECT TO cc_worker
  USING (id = (SELECT app_user_id()));
CREATE POLICY users_update_worker ON users FOR UPDATE TO cc_worker
  USING (id = (SELECT app_user_id()))
  WITH CHECK (id = (SELECT app_user_id()));
GRANT SELECT (id, email_encrypted, email_key_id), UPDATE (email_encrypted, email_key_id) ON users TO cc_worker;

CREATE POLICY taxpayers_update_worker ON taxpayers FOR UPDATE TO cc_worker
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT (rfc_encrypted, rfc_key_id), UPDATE (rfc_encrypted, rfc_key_id) ON taxpayers TO cc_worker;

CREATE POLICY devices_update_worker ON devices FOR UPDATE TO cc_worker
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT UPDATE (token_ciphertext, token_key_id) ON devices TO cc_worker;

-- +goose Down

REVOKE UPDATE (token_ciphertext, token_key_id) ON devices FROM cc_worker;
DROP POLICY devices_update_worker ON devices;
REVOKE SELECT (rfc_encrypted, rfc_key_id), UPDATE (rfc_encrypted, rfc_key_id) ON taxpayers FROM cc_worker;
DROP POLICY taxpayers_update_worker ON taxpayers;
REVOKE ALL ON users FROM cc_worker;
DROP POLICY users_update_worker ON users;
DROP POLICY users_select_worker ON users;

-- Only their owner drops the functions: cc_owner sets the role, inheriting nothing.
SET LOCAL ROLE cc_worker_definer;
DROP FUNCTION key_usage();
DROP FUNCTION reseal_pending_devices(smallint, uuid, integer);
DROP FUNCTION reseal_pending_taxpayers(smallint, uuid, integer);
DROP FUNCTION reseal_pending_users(smallint, uuid, integer);
RESET ROLE;

REVOKE SELECT (id, token_key_id, token_hmac_version) ON devices FROM cc_worker_definer;
REVOKE SELECT (email_key_id, email_hash_version) ON users FROM cc_worker_definer;
REVOKE SELECT (id, rfc_key_id, rfc_hash_version) ON taxpayers FROM cc_worker_definer;
DROP POLICY taxpayers_key_reseal ON taxpayers;
DROP FUNCTION key_reseal_scan();

ALTER TABLE devices DROP COLUMN token_hmac_version;
ALTER TABLE taxpayers DROP COLUMN rfc_hash_version;
ALTER TABLE users DROP COLUMN email_hash_version;
