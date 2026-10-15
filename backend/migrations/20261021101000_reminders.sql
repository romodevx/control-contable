-- Reminders (M9-02): devices, reminder_settings and notification_log, the claim_push_token
-- definer function the API registers devices through, and due_reminder_users, the worker's
-- one cross-user read.
-- Sources: docs/project-specs/MODULE-reminders.md §2.2-§2.6 and §8, contract 0.13.0
-- (registerDevice, unregisterDevice, get/putReminderSettings), ADR-0001 O8
-- (notification_log), docs/architecture/data-model.md "Database roles".
--
-- All three tables are user data: ENABLE + FORCE ROW LEVEL SECURITY, policies per
-- operation keyed on (SELECT app_user_id()), user_id -> users ON DELETE CASCADE (account
-- deletion, US-R8). What each runtime role may do (nothing else is granted):
--   devices            cc_app    SELECT (no token columns), DELETE own (unregisterDevice).
--                                 No INSERT/UPDATE: registration goes through
--                                 claim_push_token, which can move a row between users.
--                      cc_worker SELECT own (the token ciphertext, to send), DELETE own
--                                 (FCM UNREGISTERED cleanup).
--   reminder_settings  cc_app    SELECT, INSERT, UPDATE own (GET / PUT upsert). No DELETE:
--                                 the row goes with the account.
--                      cc_worker SELECT own.
--   notification_log   cc_worker SELECT, INSERT, UPDATE own (claim, re-claim, record). No
--                                 DELETE (retention is the platform purge's job).
--                      cc_app    nothing.
--
-- Push token at rest (§2.3, OQ-R7), the same application envelope as the RFC:
--   * token_ciphertext AES-256-GCM framed by crypto.Keyring: key_id (2 bytes) || nonce (12)
--                      || ciphertext || tag (16); token_key_id repeats the prefix (CHECK).
--   * token_hmac       keyed HMAC-SHA256 of the token (32 bytes), UNIQUE: the claim lookup.
--   Never logged, never returned (cc_app has no SELECT on either column).
--
-- Exactly once (§2.4): UNIQUE (user_id, period, kind, channel) on notification_log decides
-- which worker sends during a leader split brain, not the scheduler. A row leaves
-- 'pending' once (sent | failed | skipped) and never changes again (trigger).
--
-- Definer functions (data-model "Database roles": the only ways across users), each owned
-- by a NOLOGIN, NOBYPASSRLS bootstrap role, search_path pinned, every name qualified:
--   * claim_push_token  owner cc_auth_definer (the API-side definer role), EXECUTE cc_app.
--     Moves an installation id or a token held by any other row to the caller. Its
--     policies on devices reach the caller's own rows (app_user_id()) plus, while the call
--     runs, the rows matching the id or HMAC it binds to app.device_claim_id /
--     app.device_claim_hmac (transaction-local, cleared before it returns). It returns
--     only whether the caller's row was created, never anything about a previous owner.
--   * due_reminder_users  owner cc_worker_definer, EXECUTE cc_worker. Returns (user_id,
--     period, kind) only. Its SELECT-only policies see enabled settings, any device's
--     user_id, live taxpayers' user_id and the log's dedupe columns.
--
-- Indexes: devices (user_id, last_seen_at) serves the RLS predicate, the LRU eviction and
-- the "has a device" probe; reminder_settings gets one partial index per job branch (early
-- day, due date); the log's UNIQUE serves the dedupe probe and the RLS predicate.

-- +goose Up

-- Fail early, with a clear message, on an environment whose bootstrap lacks a definer role.
-- +goose StatementBegin
DO $$
DECLARE
  role_name text;
BEGIN
  FOREACH role_name IN ARRAY ARRAY['cc_auth_definer', 'cc_worker_definer'] LOOP
    IF NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_roles
      WHERE rolname = role_name AND NOT rolcanlogin AND NOT rolsuper AND NOT rolbypassrls
    ) THEN
      RAISE EXCEPTION 'role % is missing or privileged', role_name
        USING HINT = 'Create it as in infra/local/postgres/bootstrap/roles.sql.';
    END IF;
    IF NOT EXISTS (
      SELECT 1 FROM pg_catalog.pg_auth_members
      WHERE roleid = role_name::regrole AND member = 'cc_owner'::regrole
        AND set_option AND NOT inherit_option
    ) THEN
      RAISE EXCEPTION 'cc_owner must be a member of % WITH INHERIT FALSE, SET TRUE', role_name
        USING HINT = 'See infra/local/postgres/bootstrap/roles.sql.';
    END IF;
  END LOOP;
END
$$;
-- +goose StatementEnd

-- ---------------------------------------------------------------------------------------
-- devices: one row per app installation (§2.3). id is the client's installation id
-- (UUIDv7, stable per install), not a server-generated key.
-- ---------------------------------------------------------------------------------------
CREATE TABLE devices (
  id               uuid        PRIMARY KEY,
  user_id          uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  platform         text        NOT NULL,
  token_ciphertext bytea       NOT NULL,
  token_key_id     smallint    NOT NULL,
  token_hmac       bytea       NOT NULL,
  app_version      text,
  created_at       timestamptz NOT NULL DEFAULT now(),
  last_seen_at     timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  -- One installation per token: the claim's lookup key.
  CONSTRAINT devices_token_hmac_key UNIQUE (token_hmac),
  CONSTRAINT devices_platform_valid CHECK (platform IN ('android', 'ios')),
  CONSTRAINT devices_token_key_id_range CHECK (token_key_id >= 1),
  -- More than the 30-byte frame (2 key_id + 12 nonce + 16 tag), and the key_id prefix
  -- equals token_key_id. CASE guards get_byte against a too-short value.
  CONSTRAINT devices_token_ciphertext_framed CHECK (
    CASE WHEN octet_length(token_ciphertext) > 30
      THEN token_key_id = (get_byte(token_ciphertext, 0) << 8) | get_byte(token_ciphertext, 1)
      ELSE false
    END),
  -- 32 = crypto.HashSize (HMAC-SHA256).
  CONSTRAINT devices_token_hmac_size CHECK (octet_length(token_hmac) = 32),
  -- Contract DeviceCreate.app_version: 1-32 of [0-9A-Za-z.+-].
  CONSTRAINT devices_app_version_format CHECK (app_version ~ '^[0-9A-Za-z.+-]{1,32}$')
);
COMMENT ON TABLE devices IS
  'App installations registered for push (M9-02). Written only by claim_push_token. Token encrypted + HMAC, never returned.';
COMMENT ON COLUMN devices.id IS 'The app''s installation id (client-generated UUIDv7, stable per install).';
COMMENT ON COLUMN devices.token_ciphertext IS 'FCM registration token, AES-256-GCM, crypto.Keyring frame. Read by the worker only.';
COMMENT ON COLUMN devices.token_hmac IS 'Keyed HMAC-SHA256 of the token: the UNIQUE claim lookup key.';
COMMENT ON COLUMN devices.last_seen_at IS 'Last registerDevice of this installation; the 11th device evicts the oldest.';

CREATE INDEX devices_user_id_last_seen_at_idx ON devices (user_id, last_seen_at);
CREATE TRIGGER devices_set_updated_at
  BEFORE UPDATE ON devices
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE devices FORCE ROW LEVEL SECURITY;
-- Intent: the API lists and unregisters (404 for a foreign or unknown id) only the
-- caller's devices; the worker reads its user's tokens to send and deletes one FCM
-- reports UNREGISTERED. Neither inserts nor updates: claim_push_token does.
CREATE POLICY devices_select_own ON devices FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY devices_delete_own ON devices FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT (id, user_id, platform, app_version, created_at, last_seen_at, updated_at), DELETE
  ON devices TO cc_app;
CREATE POLICY devices_select_worker ON devices FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY devices_delete_worker ON devices FOR DELETE TO cc_worker
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT, DELETE ON devices TO cc_worker;

-- ---------------------------------------------------------------------------------------
-- reminder_settings: one row per user once saved; GET answers the defaults until then (no
-- row is created by a read). The defaults below are the §2.2 ones.
-- ---------------------------------------------------------------------------------------
CREATE TABLE reminder_settings (
  user_id      uuid        PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  enabled      boolean     NOT NULL DEFAULT false,
  early_day    smallint             DEFAULT 14,
  on_due_date  boolean     NOT NULL DEFAULT true,
  send_time    time        NOT NULL DEFAULT '10:00',
  show_amounts boolean     NOT NULL DEFAULT false,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  -- 1-16 or NULL (no early reminder). Always before the due date (day 17 or later).
  CONSTRAINT reminder_settings_early_day_range CHECK (early_day BETWEEN 1 AND 16),
  -- HH:MM, America/Mexico_City.
  CONSTRAINT reminder_settings_send_time_minutes CHECK (send_time = date_trunc('minute', send_time)::time),
  -- 422 no_reminder_selected: "on" must send something.
  CONSTRAINT reminder_settings_reminder_selected CHECK (NOT enabled OR early_day IS NOT NULL OR on_due_date)
);
COMMENT ON TABLE reminder_settings IS
  'Reminder preferences per user (M9-02, MODULE-reminders §2.2). show_amounts is a lock-screen privacy opt-in.';
COMMENT ON COLUMN reminder_settings.early_day IS
  'Nominal day of the due month for the early reminder (sent even on a weekend); NULL = none.';
COMMENT ON COLUMN reminder_settings.send_time IS 'HH:MM, America/Mexico_City for every user.';

-- One partial index per branch of due_reminder_users (keyset on user_id inside each).
CREATE INDEX reminder_settings_early_day_idx ON reminder_settings (early_day, user_id)
  WHERE enabled AND early_day IS NOT NULL;
CREATE INDEX reminder_settings_on_due_date_idx ON reminder_settings (user_id)
  WHERE enabled AND on_due_date;
CREATE TRIGGER reminder_settings_set_updated_at
  BEFORE UPDATE ON reminder_settings
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE reminder_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE reminder_settings FORCE ROW LEVEL SECURITY;
-- Intent: the API reads and saves (INSERT ... ON CONFLICT DO UPDATE) only the caller's
-- settings and can never hand them to another user; the worker reads its user's.
CREATE POLICY reminder_settings_select_own ON reminder_settings FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY reminder_settings_insert_own ON reminder_settings FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY reminder_settings_update_own ON reminder_settings FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT, INSERT (user_id, enabled, early_day, on_due_date, send_time, show_amounts),
  UPDATE (enabled, early_day, on_due_date, send_time, show_amounts)
  ON reminder_settings TO cc_app;
CREATE POLICY reminder_settings_select_worker ON reminder_settings FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT ON reminder_settings TO cc_worker;

-- ---------------------------------------------------------------------------------------
-- notification_log: one row per (user, period, kind, channel) ever attempted (§2.4).
-- period is the monthly period the reminder is about (first day of the month), never the
-- due date, so a holidays correction cannot cause a second reminder.
-- ---------------------------------------------------------------------------------------
CREATE TABLE notification_log (
  id             uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id        uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  period         date        NOT NULL,
  kind           text        NOT NULL,
  channel        text        NOT NULL DEFAULT 'push',
  status         text        NOT NULL DEFAULT 'pending',
  attempts       smallint    NOT NULL DEFAULT 1,
  claimed_at     timestamptz NOT NULL DEFAULT now(),
  sent_at        timestamptz,
  devices_ok     smallint,
  devices_failed smallint,
  error_class    text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  -- The exactly-once key. Its btree also serves the RLS predicate and the user's history.
  CONSTRAINT notification_log_user_id_period_kind_channel_key UNIQUE (user_id, period, kind, channel),
  CONSTRAINT notification_log_period_first_of_month CHECK (extract(day FROM period) = 1),
  CONSTRAINT notification_log_period_range CHECK (period BETWEEN DATE '2022-01-01' AND DATE '2100-12-01'),
  CONSTRAINT notification_log_kind_valid CHECK (kind IN ('early', 'due')),
  -- v1 sends push only; the column leaves room for email without a migration.
  CONSTRAINT notification_log_channel_valid CHECK (channel IN ('push')),
  CONSTRAINT notification_log_status_valid CHECK (status IN ('pending', 'sent', 'failed', 'skipped')),
  -- The first claim is attempt 1; a lease re-claim adds one, up to 3 (§2.4 step 5).
  CONSTRAINT notification_log_attempts_range CHECK (attempts BETWEEN 1 AND 3),
  -- sent <=> sent_at, and a send reached at least one device.
  CONSTRAINT notification_log_sent_at_iff_sent CHECK ((status = 'sent') = (sent_at IS NOT NULL)),
  -- coalesce: a NULL count would make the CHECK pass (NULL is not false).
  CONSTRAINT notification_log_sent_reached_a_device CHECK (status <> 'sent' OR coalesce(devices_ok, 0) >= 1),
  CONSTRAINT notification_log_device_counts_range CHECK (
    coalesce(devices_ok, 0) BETWEEN 0 AND 500 AND coalesce(devices_failed, 0) BETWEEN 0 AND 500),
  -- A machine class (unavailable, deadline, expired, ...), never a message or a token.
  CONSTRAINT notification_log_error_class_format CHECK (error_class ~ '^[a-z][a-z0-9_]{0,39}$')
);
COMMENT ON TABLE notification_log IS
  'One row per reminder attempted (M9-02): UNIQUE (user_id, period, kind, channel) makes the send exactly-once. No token, no PII.';
COMMENT ON COLUMN notification_log.period IS 'Monthly period the reminder is about (first of month), not the due date.';
COMMENT ON COLUMN notification_log.claimed_at IS 'Lease start: a pending row older than 10 minutes may be re-claimed.';

-- notification_log_forbid_final_change(): BEFORE UPDATE trigger. A row leaves 'pending'
-- once; sent, failed and skipped are final, so no retry or bug can resend or rewrite the
-- record of what happened. No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION notification_log_forbid_final_change() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF OLD.status <> 'pending' THEN
    RAISE EXCEPTION 'notification_log row % is final (%)', OLD.id, OLD.status
      USING ERRCODE = 'P0001', TABLE = 'notification_log';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION notification_log_forbid_final_change() IS
  'BEFORE UPDATE trigger: only pending notification_log rows change; sent, failed and skipped are final.';

CREATE TRIGGER notification_log_forbid_final_change
  BEFORE UPDATE ON notification_log
  FOR EACH ROW EXECUTE FUNCTION notification_log_forbid_final_change();
CREATE TRIGGER notification_log_set_updated_at
  BEFORE UPDATE ON notification_log
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE notification_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_log FORCE ROW LEVEL SECURITY;
-- Intent: the worker, inside one user's transaction, claims (INSERT ... ON CONFLICT DO
-- NOTHING, which needs the SELECT policy), re-claims and records that user's reminders
-- only, and can never write a row for another user. The key columns are not updatable.
CREATE POLICY notification_log_select_worker ON notification_log FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY notification_log_insert_worker ON notification_log FOR INSERT TO cc_worker
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY notification_log_update_worker ON notification_log FOR UPDATE TO cc_worker
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT,
  INSERT (user_id, period, kind, channel, status, attempts, claimed_at, sent_at, devices_ok, devices_failed, error_class),
  UPDATE (status, attempts, claimed_at, sent_at, devices_ok, devices_failed, error_class)
  ON notification_log TO cc_worker;

-- ---------------------------------------------------------------------------------------
-- claim_push_token: registerDevice's single write (§2.3).
-- ---------------------------------------------------------------------------------------
-- Claim scope helpers: what the running claim bound (NULL when unset or cleared, so the
-- scoped policies below match nothing). SECURITY INVOKER, executable by cc_auth_definer only.
-- +goose StatementBegin
CREATE FUNCTION device_claim_id() RETURNS uuid
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT nullif(current_setting('app.device_claim_id', true), '')::uuid $$;
-- +goose StatementEnd
COMMENT ON FUNCTION device_claim_id() IS
  'Installation id bound by the running claim_push_token; NULL outside one. Policy key.';

-- +goose StatementBegin
CREATE FUNCTION device_claim_hmac() RETURNS bytea
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT decode(nullif(current_setting('app.device_claim_hmac', true), ''), 'hex') $$;
-- +goose StatementEnd
COMMENT ON FUNCTION device_claim_hmac() IS
  'Token HMAC bound by the running claim_push_token; NULL outside one. Policy key.';

REVOKE ALL ON FUNCTION device_claim_id(), device_claim_hmac() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION device_claim_id(), device_claim_hmac(), app_user_id() TO cc_auth_definer;

-- Intent: while claim_push_token runs (as cc_auth_definer) it sees and deletes the rows
-- holding the installation id or token it bound (any user: the claim's purpose), and
-- inserts / refreshes only the caller's own row. Outside a claim the bound values are
-- NULL and only the caller's own rows match; no runtime role is a member of the role.
CREATE POLICY devices_claim_select ON devices FOR SELECT TO cc_auth_definer
  USING (user_id = (SELECT app_user_id()) OR id = (SELECT device_claim_id())
         OR token_hmac = (SELECT device_claim_hmac()));
CREATE POLICY devices_claim_delete ON devices FOR DELETE TO cc_auth_definer
  USING (user_id = (SELECT app_user_id()) OR id = (SELECT device_claim_id())
         OR token_hmac = (SELECT device_claim_hmac()));
CREATE POLICY devices_claim_insert ON devices FOR INSERT TO cc_auth_definer
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY devices_claim_update ON devices FOR UPDATE TO cc_auth_definer
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT (id, user_id, token_hmac, last_seen_at),
  INSERT (id, user_id, platform, token_ciphertext, token_key_id, token_hmac, app_version),
  UPDATE (platform, token_ciphertext, token_key_id, token_hmac, app_version, last_seen_at),
  DELETE
  ON devices TO cc_auth_definer;

-- claim_push_token(...) registers installation p_installation_id for the caller and returns
-- true when the caller's row was created (201: new, or claimed from another row) and false
-- when it was refreshed (200: same user, same installation).
--   1. p_user must be the transaction's user (app.current_user_id): the function never
--      writes for anyone else, even if the API passed a wrong id. Otherwise 42501.
--   2. Serializes the caller's registrations (transaction advisory lock on the user), so
--      the device cap counts a stable set.
--   3. Deletes every other row holding this installation id or this token HMAC: another
--      user's (shared phone, US-R6) or the caller's stale installation (reinstall).
--   4. Upserts the caller's row. A concurrent claim of the same installation by another
--      user that commits in between makes the upsert match nothing: 40001, retry.
--      A concurrent insert of the same token raises 23505: also a retry.
--   5. Evicts the caller's least recently seen devices beyond the cap (10, OQ-R6).
--   6. Clears the bound scope. Nothing about a previous owner leaves the function.
-- +goose StatementBegin
CREATE FUNCTION claim_push_token(
  p_user uuid, p_installation_id uuid, p_platform text, p_token_ciphertext bytea,
  p_token_key_id smallint, p_token_hmac bytea, p_app_version text
) RETURNS boolean
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  device_cap CONSTANT integer := 10;
  hmac_size  CONSTANT integer := 32;
  existed    boolean;
BEGIN
  IF p_user IS NULL OR p_user IS DISTINCT FROM public.app_user_id() THEN
    RAISE EXCEPTION 'claim_push_token registers devices for the transaction''s user only'
      USING ERRCODE = 'insufficient_privilege';
  END IF;
  IF p_installation_id IS NULL OR p_token_hmac IS NULL OR octet_length(p_token_hmac) <> hmac_size THEN
    RAISE EXCEPTION 'installation id and a % byte token HMAC are required', hmac_size
      USING ERRCODE = 'invalid_parameter_value';
  END IF;

  PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended('devices:' || p_user::text, 0));
  PERFORM pg_catalog.set_config('app.device_claim_id', p_installation_id::text, true);
  PERFORM pg_catalog.set_config('app.device_claim_hmac', pg_catalog.encode(p_token_hmac, 'hex'), true);

  DELETE FROM public.devices AS d
  WHERE (d.id = p_installation_id OR d.token_hmac = p_token_hmac)
    AND NOT (d.id = p_installation_id AND d.user_id = p_user);

  SELECT EXISTS (SELECT 1 FROM public.devices AS d WHERE d.id = p_installation_id AND d.user_id = p_user)
  INTO existed;

  INSERT INTO public.devices AS d (id, user_id, platform, token_ciphertext, token_key_id, token_hmac, app_version)
  VALUES (p_installation_id, p_user, p_platform, p_token_ciphertext, p_token_key_id, p_token_hmac, p_app_version)
  ON CONFLICT (id) DO UPDATE
    SET platform = p_platform, token_ciphertext = p_token_ciphertext, token_key_id = p_token_key_id,
        token_hmac = p_token_hmac, app_version = p_app_version, last_seen_at = pg_catalog.now()
    WHERE d.user_id = p_user;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'installation claimed concurrently; retry'
      USING ERRCODE = 'serialization_failure';
  END IF;

  DELETE FROM public.devices AS d
  WHERE d.user_id = p_user
    AND d.id IN (SELECT e.id FROM public.devices AS e WHERE e.user_id = p_user
                 ORDER BY e.last_seen_at DESC, e.id DESC OFFSET device_cap);

  PERFORM pg_catalog.set_config('app.device_claim_id', '', true);
  PERFORM pg_catalog.set_config('app.device_claim_hmac', '', true);
  RETURN NOT existed;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION claim_push_token(uuid, uuid, text, bytea, smallint, bytea, text) IS
  'M9-02: registers an installation for app_user_id(), moving the id or token from any other row; true = created. Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- due_reminder_users: who gets a reminder now (§2.4 steps 2 and 5).
-- ---------------------------------------------------------------------------------------
-- Intent of every policy below: cc_worker_definer reads, across users, only what the
-- selection needs (column grants: no token, no name, no RFC). No other role is named.
CREATE POLICY reminder_settings_due_reminders ON reminder_settings FOR SELECT TO cc_worker_definer
  USING (enabled);
GRANT SELECT (user_id, enabled, early_day, on_due_date, send_time) ON reminder_settings TO cc_worker_definer;
CREATE POLICY devices_due_reminders ON devices FOR SELECT TO cc_worker_definer USING (true);
GRANT SELECT (user_id) ON devices TO cc_worker_definer;
CREATE POLICY taxpayers_due_reminders ON taxpayers FOR SELECT TO cc_worker_definer
  USING (archived_at IS NULL);
GRANT SELECT (user_id, archived_at) ON taxpayers TO cc_worker_definer;
CREATE POLICY notification_log_due_reminders ON notification_log FOR SELECT TO cc_worker_definer USING (true);
GRANT SELECT (user_id, period, kind, channel, status, attempts, claimed_at) ON notification_log TO cc_worker_definer;

-- due_reminder_users(p_now, p_due_today, p_after, p_limit) returns up to p_limit
-- (user_id, period, kind) rows, ordered by user_id after p_after (keyset batches), of the
-- users whose reminder of today is due by p_now:
--   * today is p_now's date in America/Mexico_City; period is the month before today's
--     month (every early day, 1-16, and every moved due date, 17-30, of a period falls in
--     the month after it);
--   * p_due_today says whether today is the (moved) due date. The worker computes it with
--     the fiscal kernel (M9-04): the CFF 12 rule is not duplicated in SQL. true -> kind
--     'due' for users with on_due_date; false -> kind 'early' for users whose early_day is
--     today's day of month (the two never meet: early_day <= 16 < 17 <= the due day);
--   * enabled, send_time <= the local time of p_now (a reminder whose day has passed is
--     never returned: it is always today's reminder), >= 1 device, >= 1 live taxpayer;
--   * no notification_log row for (user, period, kind, 'push'), or a 'pending' one whose
--     10-minute lease expired with attempts < 3 (the retry, which the worker re-claims
--     atomically).
-- +goose StatementBegin
CREATE FUNCTION due_reminder_users(p_now timestamptz, p_due_today boolean, p_after uuid, p_limit integer)
  RETURNS TABLE (user_id uuid, period date, kind text)
  LANGUAGE plpgsql STABLE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_limit    CONSTANT integer  := 1000;
  max_attempts CONSTANT integer  := 3;
  lease        CONSTANT interval := interval '10 minutes';
  zone         CONSTANT text     := 'America/Mexico_City';
  local_now    timestamp;
  v_period     date;
  v_kind       text;
BEGIN
  IF p_now IS NULL OR p_due_today IS NULL OR p_limit IS NULL OR p_limit < 1 OR p_limit > max_limit THEN
    RAISE EXCEPTION 'p_now and p_due_today are required and p_limit must be between 1 and %', max_limit
      USING ERRCODE = 'invalid_parameter_value';
  END IF;
  local_now := p_now AT TIME ZONE zone;
  v_period := (date_trunc('month', local_now) - interval '1 month')::date;
  v_kind := CASE WHEN p_due_today THEN 'due' ELSE 'early' END;

  IF p_due_today THEN
    RETURN QUERY
    SELECT s.user_id, v_period, v_kind
    FROM public.reminder_settings AS s
    WHERE s.enabled AND s.on_due_date
      AND s.send_time <= local_now::time
      AND (p_after IS NULL OR s.user_id > p_after)
      AND EXISTS (SELECT 1 FROM public.devices AS d WHERE d.user_id = s.user_id)
      AND EXISTS (SELECT 1 FROM public.taxpayers AS t WHERE t.user_id = s.user_id AND t.archived_at IS NULL)
      AND NOT EXISTS (
        SELECT 1 FROM public.notification_log AS l
        WHERE l.user_id = s.user_id AND l.period = v_period AND l.kind = v_kind AND l.channel = 'push'
          AND NOT (l.status = 'pending' AND l.claimed_at < p_now - lease AND l.attempts < max_attempts))
    ORDER BY s.user_id
    LIMIT p_limit;
  ELSE
    RETURN QUERY
    SELECT s.user_id, v_period, v_kind
    FROM public.reminder_settings AS s
    WHERE s.enabled AND s.early_day = extract(day FROM local_now)::smallint
      AND s.send_time <= local_now::time
      AND (p_after IS NULL OR s.user_id > p_after)
      AND EXISTS (SELECT 1 FROM public.devices AS d WHERE d.user_id = s.user_id)
      AND EXISTS (SELECT 1 FROM public.taxpayers AS t WHERE t.user_id = s.user_id AND t.archived_at IS NULL)
      AND NOT EXISTS (
        SELECT 1 FROM public.notification_log AS l
        WHERE l.user_id = s.user_id AND l.period = v_period AND l.kind = v_kind AND l.channel = 'push'
          AND NOT (l.status = 'pending' AND l.claimed_at < p_now - lease AND l.attempts < max_attempts))
    ORDER BY s.user_id
    LIMIT p_limit;
  END IF;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION due_reminder_users(timestamptz, boolean, uuid, integer) IS
  'M9-02: (user_id, period, kind) whose reminder of today is due by p_now and not yet claimed (or whose lease expired); keyset by user_id. Definer: cc_worker_definer; EXECUTE: cc_worker.';

-- Grants first: once a definer role owns a function, cc_owner (not inheriting) may not
-- grant on it. ALTER OWNER needs the new owner to hold CREATE on the schema: lent for the
-- transfer only.
REVOKE ALL ON FUNCTION claim_push_token(uuid, uuid, text, bytea, smallint, bytea, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION claim_push_token(uuid, uuid, text, bytea, smallint, bytea, text) TO cc_app;
REVOKE ALL ON FUNCTION due_reminder_users(timestamptz, boolean, uuid, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION due_reminder_users(timestamptz, boolean, uuid, integer) TO cc_worker;

GRANT CREATE ON SCHEMA public TO cc_auth_definer;
ALTER FUNCTION claim_push_token(uuid, uuid, text, bytea, smallint, bytea, text) OWNER TO cc_auth_definer;
REVOKE CREATE ON SCHEMA public FROM cc_auth_definer;
GRANT CREATE ON SCHEMA public TO cc_worker_definer;
ALTER FUNCTION due_reminder_users(timestamptz, boolean, uuid, integer) OWNER TO cc_worker_definer;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

-- +goose Down

-- Only their owners drop the functions: cc_owner sets the role, inheriting nothing.
SET LOCAL ROLE cc_worker_definer;
DROP FUNCTION due_reminder_users(timestamptz, boolean, uuid, integer);
RESET ROLE;
SET LOCAL ROLE cc_auth_definer;
DROP FUNCTION claim_push_token(uuid, uuid, text, bytea, smallint, bytea, text);
RESET ROLE;

REVOKE ALL ON taxpayers FROM cc_worker_definer;
DROP POLICY taxpayers_due_reminders ON taxpayers;
REVOKE EXECUTE ON FUNCTION app_user_id() FROM cc_auth_definer;

DROP TABLE notification_log;
DROP FUNCTION notification_log_forbid_final_change();
DROP TABLE reminder_settings;
DROP TABLE devices;
DROP FUNCTION device_claim_hmac();
DROP FUNCTION device_claim_id();
