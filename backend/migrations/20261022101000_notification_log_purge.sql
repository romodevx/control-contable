-- notification_log retention (M9-06 carry-over, PLAN M9-02 N3; MODULE-reminders §2.6 and §8):
-- the platform purge (M9-08, purge_soft_deleted) also deletes reminder log rows created more
-- than 13 months ago, whatever module_reminders says (retention is a legal promise, not a
-- feature). Must exist before module_reminders is enabled in a shared environment.
--
-- Same least-privilege shape as M9-08 (docs/runbooks/records-purge.md):
--   * the cutoff is one definition (purge_notification_log_cutoff), computed by the database
--     clock, shared by the policies and the function;
--   * cc_worker_definer's new policies on notification_log are FOR SELECT and FOR DELETE and
--     match only rows past the cutoff; it gets SELECT (id, created_at) and DELETE. Its
--     due_reminder_users SELECT policy (M9-02) is unchanged; the DELETE policy alone bounds
--     what the function can remove;
--   * purge_soft_deleted keeps its owner, search_path, grants and the order of its targets;
--     notification_log is appended last. Replacing it runs as its owner, with CREATE on the
--     schema lent for that statement only (as at the ownership transfer).
--
-- Index: notification_log (created_at) for the bounded, oldest-first scan (the UNIQUE leads
-- with user_id, useless here). The table only grows by one row per user and reminder.

-- +goose Up

CREATE INDEX notification_log_created_at_idx ON notification_log (created_at);

-- +goose StatementBegin
CREATE FUNCTION purge_notification_log_cutoff() RETURNS timestamptz
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT now() - interval '13 months' $$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_notification_log_cutoff() IS
  'M9-06: notification_log rows created before this instant (13 months ago) are purged.';
REVOKE ALL ON FUNCTION purge_notification_log_cutoff() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION purge_notification_log_cutoff() TO cc_worker_definer;

-- Intent: cc_worker_definer sees and deletes only log rows past the cutoff, of every user.
CREATE POLICY notification_log_purge_select ON notification_log FOR SELECT TO cc_worker_definer
  USING (created_at < (SELECT purge_notification_log_cutoff()));
CREATE POLICY notification_log_purge_delete ON notification_log FOR DELETE TO cc_worker_definer
  USING (created_at < (SELECT purge_notification_log_cutoff()));
GRANT SELECT (id, created_at), DELETE ON notification_log TO cc_worker_definer;

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

-- +goose Down

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
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_soft_deleted(integer) IS
  'M9-08: deletes up to p_batch_size rows per target past its cutoff (records 30 d after soft delete; dead auth state after 1 d); returns counts. Definer: cc_worker_definer; EXECUTE: cc_worker.';
RESET ROLE;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

-- Only this migration's reach: the due_reminder_users read (M9-02) names other columns.
REVOKE SELECT (id, created_at), DELETE ON notification_log FROM cc_worker_definer;
DROP POLICY notification_log_purge_delete ON notification_log;
DROP POLICY notification_log_purge_select ON notification_log;
DROP FUNCTION purge_notification_log_cutoff();
DROP INDEX notification_log_created_at_idx;
