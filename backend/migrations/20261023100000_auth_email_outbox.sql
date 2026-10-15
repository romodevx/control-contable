-- Auth email outbox (M1-06 carry-over, pre-enable item for module_auth): register and
-- resendVerification used to send their email inside the request, so the response time
-- told whether an address holds an account (a verified one gets a notice, a new one a
-- code, and each path talks to the provider differently). From now on both only ENQUEUE
-- the email in the request's own transaction and return; the worker job
-- auth_email_delivery delivers it through the EmailSender port.
--
-- Data minimization: a row holds one opaque payload, AES-256-GCM sealed by the API
-- (AUTH_OUTBOX_KEY, the row id as additional data), with the address, the kind of email
-- and the code inside. The database never sees which kind a row is (that alone would
-- say whether the address holds an account), nor an email hash to link rows to users.
-- A row is deleted on delivery, on its last failed attempt, and once it expires (the
-- code's 15 minutes): the claim deletes expired rows on every run and the platform
-- purge (M9-08, purge_soft_deleted) deletes them too.
--
-- Access, same definer posture as the auth service (M1-03):
--   * ENABLE + FORCE ROW LEVEL SECURITY; the only policies are cc_auth_definer's (it owns
--     the four functions below) and cc_worker_definer's purge policies (expired rows
--     only). No runtime role has a table privilege.
--   * cc_app: EXECUTE auth_email_outbox_enqueue only (insert one row, nothing else).
--   * cc_worker: EXECUTE auth_email_outbox_claim / _finish / _retry only.
--
-- Worker split brain (two instances briefly leading, internal/platform/worker): a claim
-- picks due rows with FOR UPDATE SKIP LOCKED, so concurrent claims never wait on or
-- return the same row, and leases them (next_attempt_at = now() + lease) under a fresh
-- claim_token. Finish and retry are compare-and-set on that token, so a holder whose
-- lease ran out can no longer delete or reschedule the row another claim now holds.
-- Delivery is at least once by nature (a crash between send and finish resends after
-- the lease); the use case keeps every run far shorter than the lease.
--
-- Retries are bounded: a claim counts the attempt and never hands out a row past
-- max_attempts (5, mirrored by usecase.MaxDeliveryAttempts); the job drops a row whose
-- last attempt failed. Backoff is the job's (it passes the delay to retry).
--
-- Index: (next_attempt_at) for the claim's oldest-due-first scan; (expires_at) for the
-- deletes of expired rows. The table holds at most a few minutes of sign-up emails.

-- +goose Up

CREATE TABLE auth_email_outbox (
  id              uuid        PRIMARY KEY,
  payload         bytea       NOT NULL,
  attempts        integer     NOT NULL DEFAULT 0,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  claim_token     uuid,
  created_at      timestamptz NOT NULL DEFAULT now(),
  expires_at      timestamptz NOT NULL,
  CONSTRAINT auth_email_outbox_payload_size CHECK (octet_length(payload) BETWEEN 1 AND 4096),
  CONSTRAINT auth_email_outbox_attempts_range CHECK (attempts BETWEEN 0 AND 5),
  CONSTRAINT auth_email_outbox_lifetime CHECK (
    expires_at > created_at AND expires_at <= created_at + interval '1 hour')
);
COMMENT ON TABLE auth_email_outbox IS
  'Auth emails waiting for delivery (register, resend). Service-only: sealed payload, no email hash; deleted on delivery, after the last attempt or at expiry.';
COMMENT ON COLUMN auth_email_outbox.payload IS
  'AES-256-GCM (crypto.Keyring, AUTH_OUTBOX_KEY) of the kind, address and code; additional data = the row id.';
COMMENT ON COLUMN auth_email_outbox.claim_token IS
  'Token of the claim that holds the row; finish and retry compare-and-set on it.';

CREATE INDEX auth_email_outbox_next_attempt_idx ON auth_email_outbox (next_attempt_at);
CREATE INDEX auth_email_outbox_expires_at_idx ON auth_email_outbox (expires_at);

-- Intent: service-only. Nobody reads or writes rows directly; cc_auth_definer acts for
-- the API (insert) and the worker (claim, finish, retry) inside the functions below.
ALTER TABLE auth_email_outbox ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth_email_outbox FORCE ROW LEVEL SECURITY;
CREATE POLICY auth_email_outbox_definer_insert ON auth_email_outbox FOR INSERT TO cc_auth_definer
  WITH CHECK (true);
CREATE POLICY auth_email_outbox_definer_select ON auth_email_outbox FOR SELECT TO cc_auth_definer
  USING (true);
CREATE POLICY auth_email_outbox_definer_update ON auth_email_outbox FOR UPDATE TO cc_auth_definer
  USING (true) WITH CHECK (true);
CREATE POLICY auth_email_outbox_definer_delete ON auth_email_outbox FOR DELETE TO cc_auth_definer
  USING (true);
GRANT SELECT, DELETE ON auth_email_outbox TO cc_auth_definer;
GRANT INSERT (id, payload, expires_at) ON auth_email_outbox TO cc_auth_definer;
GRANT UPDATE (attempts, next_attempt_at, claim_token) ON auth_email_outbox TO cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- auth_email_outbox_enqueue: the API's only write. p_id is drawn by the API, which seals
-- the payload with it as additional data (a payload moved to another row fails to open).
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION auth_email_outbox_enqueue(p_id uuid, p_payload bytea, p_expires_at timestamptz)
  RETURNS void
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  IF p_id IS NULL THEN
    RAISE EXCEPTION 'id is required' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_payload IS NULL OR octet_length(p_payload) NOT BETWEEN 1 AND 4096 THEN
    RAISE EXCEPTION 'payload must be 1 to 4096 bytes' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_expires_at IS NULL OR p_expires_at <= now() OR p_expires_at > now() + interval '1 hour' THEN
    RAISE EXCEPTION 'expires_at must be within the next hour' USING ERRCODE = 'invalid_parameter_value';
  END IF;
  INSERT INTO public.auth_email_outbox (id, payload, expires_at) VALUES (p_id, p_payload, p_expires_at);
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_email_outbox_enqueue(uuid, bytea, timestamptz) IS
  'Queues one sealed auth email, deliverable until p_expires_at (within the hour). Definer: cc_auth_definer; EXECUTE: cc_app.';

-- ---------------------------------------------------------------------------------------
-- auth_email_outbox_claim: deletes expired rows, then leases up to p_limit due rows
-- (oldest due first, SKIP LOCKED) for p_lease_seconds under a fresh claim token, counting
-- the attempt. Rows at max_attempts are never handed out again (they expire).
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION auth_email_outbox_claim(p_limit integer, p_lease_seconds integer)
  RETURNS TABLE (id uuid, payload bytea, attempts integer, claim_token uuid)
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_attempts CONSTANT integer := 5;
  max_limit    CONSTANT integer := 100;
  min_lease    CONSTANT integer := 10;
  max_lease    CONSTANT integer := 900;
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > max_limit THEN
    RAISE EXCEPTION 'limit must be between 1 and %', max_limit USING ERRCODE = 'invalid_parameter_value';
  END IF;
  IF p_lease_seconds IS NULL OR p_lease_seconds < min_lease OR p_lease_seconds > max_lease THEN
    RAISE EXCEPTION 'lease must be between % and % seconds', min_lease, max_lease
      USING ERRCODE = 'invalid_parameter_value';
  END IF;

  DELETE FROM public.auth_email_outbox AS o
  WHERE o.id IN (SELECT e.id FROM public.auth_email_outbox AS e
                 WHERE e.expires_at <= now() ORDER BY e.expires_at LIMIT max_limit
                 FOR UPDATE SKIP LOCKED);

  RETURN QUERY
  WITH due AS (
    SELECT d.id FROM public.auth_email_outbox AS d
    WHERE d.next_attempt_at <= now() AND d.expires_at > now() AND d.attempts < max_attempts
    ORDER BY d.next_attempt_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  UPDATE public.auth_email_outbox AS o
  SET attempts = o.attempts + 1,
      next_attempt_at = now() + make_interval(secs => p_lease_seconds),
      claim_token = gen_random_uuid()
  FROM due
  WHERE o.id = due.id
  RETURNING o.id, o.payload, o.attempts, o.claim_token;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_email_outbox_claim(integer, integer) IS
  'Deletes expired rows, then leases up to p_limit due rows (SKIP LOCKED, max 5 attempts) under a fresh claim token. Definer: cc_auth_definer; EXECUTE: cc_worker.';

-- +goose StatementBegin
CREATE FUNCTION auth_email_outbox_finish(p_id uuid, p_claim_token uuid) RETURNS boolean
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
BEGIN
  DELETE FROM public.auth_email_outbox AS o WHERE o.id = p_id AND o.claim_token = p_claim_token;
  RETURN FOUND;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_email_outbox_finish(uuid, uuid) IS
  'Deletes a row its claim holder delivered (or gave up on); false when the claim is no longer held. Definer: cc_auth_definer; EXECUTE: cc_worker.';

-- +goose StatementBegin
CREATE FUNCTION auth_email_outbox_retry(p_id uuid, p_claim_token uuid, p_delay_seconds integer)
  RETURNS boolean
  LANGUAGE plpgsql
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_delay CONSTANT integer := 3600;
BEGIN
  IF p_delay_seconds IS NULL OR p_delay_seconds < 1 OR p_delay_seconds > max_delay THEN
    RAISE EXCEPTION 'delay must be between 1 and % seconds', max_delay USING ERRCODE = 'invalid_parameter_value';
  END IF;
  UPDATE public.auth_email_outbox AS o
  SET next_attempt_at = now() + make_interval(secs => p_delay_seconds), claim_token = NULL
  WHERE o.id = p_id AND o.claim_token = p_claim_token;
  RETURN FOUND;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION auth_email_outbox_retry(uuid, uuid, integer) IS
  'Releases a held row for another attempt after p_delay_seconds; false when the claim is no longer held. Definer: cc_auth_definer; EXECUTE: cc_worker.';

-- Grants before the ownership transfer (cc_owner can no longer grant afterwards).
REVOKE ALL ON FUNCTION
  auth_email_outbox_enqueue(uuid, bytea, timestamptz),
  auth_email_outbox_claim(integer, integer),
  auth_email_outbox_finish(uuid, uuid),
  auth_email_outbox_retry(uuid, uuid, integer)
  FROM PUBLIC;
GRANT EXECUTE ON FUNCTION auth_email_outbox_enqueue(uuid, bytea, timestamptz) TO cc_app;
GRANT EXECUTE ON FUNCTION
  auth_email_outbox_claim(integer, integer),
  auth_email_outbox_finish(uuid, uuid),
  auth_email_outbox_retry(uuid, uuid, integer)
  TO cc_worker;

GRANT CREATE ON SCHEMA public TO cc_auth_definer;
ALTER FUNCTION auth_email_outbox_enqueue(uuid, bytea, timestamptz) OWNER TO cc_auth_definer;
ALTER FUNCTION auth_email_outbox_claim(integer, integer) OWNER TO cc_auth_definer;
ALTER FUNCTION auth_email_outbox_finish(uuid, uuid) OWNER TO cc_auth_definer;
ALTER FUNCTION auth_email_outbox_retry(uuid, uuid, integer) OWNER TO cc_auth_definer;
REVOKE CREATE ON SCHEMA public FROM cc_auth_definer;

-- ---------------------------------------------------------------------------------------
-- Platform purge (M9-08): expired outbox rows go too, whatever their state. Intent:
-- cc_worker_definer sees and deletes only rows past their expiry.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION purge_auth_email_outbox_cutoff() RETURNS timestamptz
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT now() $$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_auth_email_outbox_cutoff() IS
  'Outbox rows that expired before this instant (now: an expired email is never sent) are purged.';
REVOKE ALL ON FUNCTION purge_auth_email_outbox_cutoff() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION purge_auth_email_outbox_cutoff() TO cc_worker_definer;

CREATE POLICY auth_email_outbox_purge_select ON auth_email_outbox FOR SELECT TO cc_worker_definer
  USING (expires_at < (SELECT purge_auth_email_outbox_cutoff()));
CREATE POLICY auth_email_outbox_purge_delete ON auth_email_outbox FOR DELETE TO cc_worker_definer
  USING (expires_at < (SELECT purge_auth_email_outbox_cutoff()));
GRANT SELECT (id, expires_at), DELETE ON auth_email_outbox TO cc_worker_definer;

GRANT CREATE ON SCHEMA public TO cc_worker_definer;
SET LOCAL ROLE cc_worker_definer;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION purge_soft_deleted(p_batch_size integer)
  RETURNS TABLE (target text, purged integer)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_batch_size CONSTANT integer     := 5000;
  records_cutoff CONSTANT timestamptz := public.purge_records_cutoff();
  auth_cutoff    CONSTANT timestamptz := public.purge_auth_cutoff();
  log_cutoff     CONSTANT timestamptz := public.purge_notification_log_cutoff();
  outbox_cutoff  CONSTANT timestamptz := public.purge_auth_email_outbox_cutoff();
BEGIN
  IF p_batch_size IS NULL OR p_batch_size < 1 OR p_batch_size > max_batch_size THEN
    RAISE EXCEPTION 'batch size must be between 1 and %', max_batch_size
      USING ERRCODE = 'invalid_parameter_value';
  END IF;

  target := 'incomes';
  DELETE FROM public.incomes AS r
  WHERE r.id IN (SELECT d.id FROM public.incomes AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'expenses';
  DELETE FROM public.expenses AS r
  WHERE r.id IN (SELECT d.id FROM public.expenses AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'personal_deductions';
  DELETE FROM public.personal_deductions AS r
  WHERE r.id IN (SELECT d.id FROM public.personal_deductions AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  -- Two index range scans (expired families, revoked families), merged, then bounded.
  target := 'refresh_tokens';
  DELETE FROM public.refresh_tokens AS r
  WHERE r.id IN (SELECT u.id FROM (
                   (SELECT a.id FROM public.refresh_tokens AS a
                    WHERE a.absolute_expires_at < auth_cutoff ORDER BY a.absolute_expires_at LIMIT p_batch_size)
                   UNION
                   (SELECT v.id FROM public.refresh_tokens AS v
                    WHERE v.revoked_at < auth_cutoff ORDER BY v.revoked_at LIMIT p_batch_size)
                 ) AS u LIMIT p_batch_size)
    AND (r.absolute_expires_at < auth_cutoff OR r.revoked_at < auth_cutoff);
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'email_verifications';
  DELETE FROM public.email_verifications AS r
  WHERE r.email_hash IN (SELECT d.email_hash FROM public.email_verifications AS d
                         WHERE d.expires_at < auth_cutoff ORDER BY d.expires_at LIMIT p_batch_size)
    AND r.expires_at < auth_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'email_verification_failures';
  DELETE FROM public.email_verification_failures AS r
  WHERE r.email_hash IN (SELECT d.email_hash FROM public.email_verification_failures AS d
                         WHERE d.window_started_at < auth_cutoff LIMIT p_batch_size)
    AND r.window_started_at < auth_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  -- notification_log (M9-06, MODULE-reminders §2.6): 13-month retention of the reminder log.
  target := 'notification_log';
  DELETE FROM public.notification_log AS r
  WHERE r.id IN (SELECT d.id FROM public.notification_log AS d
                 WHERE d.created_at < log_cutoff ORDER BY d.created_at LIMIT p_batch_size)
    AND r.created_at < log_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  -- auth_email_outbox: expired emails (sealed address and code) are never delivered.
  target := 'auth_email_outbox';
  DELETE FROM public.auth_email_outbox AS r
  WHERE r.id IN (SELECT d.id FROM public.auth_email_outbox AS d
                 WHERE d.expires_at < outbox_cutoff ORDER BY d.expires_at LIMIT p_batch_size)
    AND r.expires_at < outbox_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_soft_deleted(integer) IS
  'M9-08: deletes up to p_batch_size rows per target past its cutoff (records 30 d after soft delete; dead auth state after 1 d; notification_log after 13 months, M9-06; expired auth_email_outbox rows); returns counts. Definer: cc_worker_definer; EXECUTE: cc_worker.';
RESET ROLE;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

-- +goose Down

-- purge_soft_deleted as 20261022101000_notification_log_purge left it.
GRANT CREATE ON SCHEMA public TO cc_worker_definer;
SET LOCAL ROLE cc_worker_definer;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION purge_soft_deleted(p_batch_size integer)
  RETURNS TABLE (target text, purged integer)
  LANGUAGE plpgsql VOLATILE
  SECURITY DEFINER
  SET search_path = pg_catalog, public, pg_temp
AS $$
DECLARE
  max_batch_size CONSTANT integer     := 5000;
  records_cutoff CONSTANT timestamptz := public.purge_records_cutoff();
  auth_cutoff    CONSTANT timestamptz := public.purge_auth_cutoff();
  log_cutoff     CONSTANT timestamptz := public.purge_notification_log_cutoff();
BEGIN
  IF p_batch_size IS NULL OR p_batch_size < 1 OR p_batch_size > max_batch_size THEN
    RAISE EXCEPTION 'batch size must be between 1 and %', max_batch_size
      USING ERRCODE = 'invalid_parameter_value';
  END IF;

  target := 'incomes';
  DELETE FROM public.incomes AS r
  WHERE r.id IN (SELECT d.id FROM public.incomes AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'expenses';
  DELETE FROM public.expenses AS r
  WHERE r.id IN (SELECT d.id FROM public.expenses AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'personal_deductions';
  DELETE FROM public.personal_deductions AS r
  WHERE r.id IN (SELECT d.id FROM public.personal_deductions AS d
                 WHERE d.deleted_at < records_cutoff ORDER BY d.deleted_at LIMIT p_batch_size)
    AND r.deleted_at < records_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  -- Two index range scans (expired families, revoked families), merged, then bounded.
  target := 'refresh_tokens';
  DELETE FROM public.refresh_tokens AS r
  WHERE r.id IN (SELECT u.id FROM (
                   (SELECT a.id FROM public.refresh_tokens AS a
                    WHERE a.absolute_expires_at < auth_cutoff ORDER BY a.absolute_expires_at LIMIT p_batch_size)
                   UNION
                   (SELECT v.id FROM public.refresh_tokens AS v
                    WHERE v.revoked_at < auth_cutoff ORDER BY v.revoked_at LIMIT p_batch_size)
                 ) AS u LIMIT p_batch_size)
    AND (r.absolute_expires_at < auth_cutoff OR r.revoked_at < auth_cutoff);
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'email_verifications';
  DELETE FROM public.email_verifications AS r
  WHERE r.email_hash IN (SELECT d.email_hash FROM public.email_verifications AS d
                         WHERE d.expires_at < auth_cutoff ORDER BY d.expires_at LIMIT p_batch_size)
    AND r.expires_at < auth_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  target := 'email_verification_failures';
  DELETE FROM public.email_verification_failures AS r
  WHERE r.email_hash IN (SELECT d.email_hash FROM public.email_verification_failures AS d
                         WHERE d.window_started_at < auth_cutoff LIMIT p_batch_size)
    AND r.window_started_at < auth_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;

  -- notification_log (M9-06, MODULE-reminders §2.6): 13-month retention of the reminder log.
  target := 'notification_log';
  DELETE FROM public.notification_log AS r
  WHERE r.id IN (SELECT d.id FROM public.notification_log AS d
                 WHERE d.created_at < log_cutoff ORDER BY d.created_at LIMIT p_batch_size)
    AND r.created_at < log_cutoff;
  GET DIAGNOSTICS purged = ROW_COUNT;
  RETURN NEXT;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_soft_deleted(integer) IS
  'M9-08: deletes up to p_batch_size rows per target past its cutoff (records 30 d after soft delete; dead auth state after 1 d; notification_log after 13 months, M9-06); returns counts. Definer: cc_worker_definer; EXECUTE: cc_worker.';
RESET ROLE;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

REVOKE SELECT (id, expires_at), DELETE ON auth_email_outbox FROM cc_worker_definer;
DROP POLICY auth_email_outbox_purge_delete ON auth_email_outbox;
DROP POLICY auth_email_outbox_purge_select ON auth_email_outbox;
DROP FUNCTION purge_auth_email_outbox_cutoff();

-- The functions belong to cc_auth_definer: only it may drop them.
SET LOCAL ROLE cc_auth_definer;
DROP FUNCTION auth_email_outbox_retry(uuid, uuid, integer);
DROP FUNCTION auth_email_outbox_finish(uuid, uuid);
DROP FUNCTION auth_email_outbox_claim(integer, integer);
DROP FUNCTION auth_email_outbox_enqueue(uuid, bytea, timestamptz);
RESET ROLE;
DROP TABLE auth_email_outbox;
