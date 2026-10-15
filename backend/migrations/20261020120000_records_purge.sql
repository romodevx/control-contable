-- Records purge (M9-08): purge_soft_deleted, the SECURITY DEFINER function the worker
-- runs (whatever module_reminders says) to delete what the product promises not to keep.
-- Sources: docs/project-specs/PLAN-control-contable.md M9-08 (+ SEC FU1, #33/#44, #34),
-- docs/architecture/data-model.md "Database roles", docs/runbooks/records-purge.md.
--
-- What goes, and when (cutoffs are computed by the database clock, never passed in, so
-- the caller cannot move them):
--   * incomes, expenses, personal_deductions: soft-deleted (deleted_at) more than 30 days
--     (720 h) ago. Restore / Undo is possible until then. Live rows are never touched.
--   * refresh_tokens: rows of a dead family more than a day ago: past the family's
--     absolute expiry (every token of a family shares it) or revoked (revocation is
--     family-wide). Rotated tokens of a live family stay: reuse detection needs them. A
--     family that only slid out of its 30-day window goes at its absolute expiry
--     (at most 90 days after login).
--   * email_verifications: pending sign-ups expired more than a day ago (until then
--     email_verification_consume still answers "expired", not "invalid").
--   * email_verification_failures: rows whose 24 h window ended (consume treats them as
--     a fresh window anyway).
--   rate_limit_buckets are NOT purged here: rate_limit_take would return NULL if a row
--   vanished mid-call (PLAN #47 FU2); that fix and its purge travel together.
--
-- Least privilege:
--   * cc_worker_definer (NOLOGIN, NOBYPASSRLS, bootstrap) owns the function. Its policies
--     here are TO cc_worker_definer only, FOR SELECT and FOR DELETE, and match only rows
--     past the cutoff: even a buggy function body could not delete a live record. It gets
--     SELECT on the key and cutoff columns only, DELETE, and nothing else.
--   * cc_worker gets EXECUTE on the function and nothing else. It still sees no user row
--     without a user context, and has no DELETE on any table.
--   * search_path is pinned to pg_catalog, public, pg_temp; every name is schema-qualified.
--
-- Concurrency (worker split brain: two leaders may overlap for a heartbeat): each DELETE
-- re-checks the cutoff on the row it locked (READ COMMITTED re-evaluation), so a row is
-- deleted once; the later call waits on the earlier one's row locks and then skips the
-- rows it finds gone. A restore that commits first wins the same way (deleted_at NULL).
-- Bounded: each call deletes at most p_batch_size rows per target (1..5000), each in the
-- caller's short transaction; the worker loops while a target fills its batch.
--
-- Indexes: incomes/expenses/personal_deductions already have a deleted_at partial index
-- (M3-02, M4-03, ANN-03) and email_verifications one on expires_at. refresh_tokens gets
-- (absolute_expires_at) and a partial (revoked_at); rotation changes rotated_at only, so
-- its updates stay HOT. email_verification_failures has no window index on purpose (a
-- small table whose every wrong code updates the window): the daily purge scans it.

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
      USING HINT = 'Create it as in infra/local/postgres/bootstrap/roles.sql (docs/runbooks/records-purge.md).';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_catalog.pg_auth_members
    WHERE roleid = 'cc_worker_definer'::regrole AND member = 'cc_owner'::regrole
      AND set_option AND NOT inherit_option
  ) THEN
    RAISE EXCEPTION 'cc_owner must be a member of cc_worker_definer WITH INHERIT FALSE, SET TRUE'
      USING HINT = 'See docs/runbooks/records-purge.md.';
  END IF;
END
$$;
-- +goose StatementEnd

CREATE INDEX refresh_tokens_absolute_expires_at_idx ON refresh_tokens (absolute_expires_at);
CREATE INDEX refresh_tokens_revoked_at_idx ON refresh_tokens (revoked_at) WHERE revoked_at IS NOT NULL;

-- Cutoffs: one definition each, shared by the policies and the function. SECURITY
-- INVOKER, executable by cc_worker_definer only (the role the policies apply to).
-- +goose StatementBegin
CREATE FUNCTION purge_records_cutoff() RETURNS timestamptz
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT now() - interval '720 hours' $$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_records_cutoff() IS
  'M9-08: soft-deleted records deleted before this instant (30 days ago) are purged.';

-- +goose StatementBegin
CREATE FUNCTION purge_auth_cutoff() RETURNS timestamptz
  LANGUAGE sql STABLE PARALLEL SAFE
  SET search_path = pg_catalog, pg_temp
AS $$ SELECT now() - interval '24 hours' $$;
-- +goose StatementEnd
COMMENT ON FUNCTION purge_auth_cutoff() IS
  'M9-08: dead auth state (expired or revoked sessions, expired sign-ups, ended failure windows) older than this is purged.';

REVOKE ALL ON FUNCTION purge_records_cutoff(), purge_auth_cutoff() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION purge_records_cutoff(), purge_auth_cutoff() TO cc_worker_definer;

-- Intent of every policy below: cc_worker_definer sees and deletes only rows already past
-- their cutoff, of every user (the purge is cross-user by nature). No other role is named.
CREATE POLICY incomes_purge_select ON incomes FOR SELECT TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
CREATE POLICY incomes_purge_delete ON incomes FOR DELETE TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
GRANT SELECT (id, deleted_at), DELETE ON incomes TO cc_worker_definer;

CREATE POLICY expenses_purge_select ON expenses FOR SELECT TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
CREATE POLICY expenses_purge_delete ON expenses FOR DELETE TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
GRANT SELECT (id, deleted_at), DELETE ON expenses TO cc_worker_definer;

CREATE POLICY personal_deductions_purge_select ON personal_deductions FOR SELECT TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
CREATE POLICY personal_deductions_purge_delete ON personal_deductions FOR DELETE TO cc_worker_definer
  USING (deleted_at < (SELECT purge_records_cutoff()));
GRANT SELECT (id, deleted_at), DELETE ON personal_deductions TO cc_worker_definer;

CREATE POLICY refresh_tokens_purge_select ON refresh_tokens FOR SELECT TO cc_worker_definer
  USING (absolute_expires_at < (SELECT purge_auth_cutoff()) OR revoked_at < (SELECT purge_auth_cutoff()));
CREATE POLICY refresh_tokens_purge_delete ON refresh_tokens FOR DELETE TO cc_worker_definer
  USING (absolute_expires_at < (SELECT purge_auth_cutoff()) OR revoked_at < (SELECT purge_auth_cutoff()));
GRANT SELECT (id, absolute_expires_at, revoked_at), DELETE ON refresh_tokens TO cc_worker_definer;

-- email_verifications and email_verification_failures stay ENABLE (not FORCE): ownerless
-- service tables (M1-03); these policies only add the definer's purge reach.
CREATE POLICY email_verifications_purge_select ON email_verifications FOR SELECT TO cc_worker_definer
  USING (expires_at < (SELECT purge_auth_cutoff()));
CREATE POLICY email_verifications_purge_delete ON email_verifications FOR DELETE TO cc_worker_definer
  USING (expires_at < (SELECT purge_auth_cutoff()));
GRANT SELECT (email_hash, expires_at), DELETE ON email_verifications TO cc_worker_definer;

CREATE POLICY email_verification_failures_purge_select ON email_verification_failures
  FOR SELECT TO cc_worker_definer
  USING (window_started_at < (SELECT purge_auth_cutoff()));
CREATE POLICY email_verification_failures_purge_delete ON email_verification_failures
  FOR DELETE TO cc_worker_definer
  USING (window_started_at < (SELECT purge_auth_cutoff()));
GRANT SELECT (email_hash, window_started_at), DELETE ON email_verification_failures TO cc_worker_definer;

-- purge_soft_deleted(p_batch_size): one bounded pass. Returns one row per target, in a
-- fixed order, with the number of rows it deleted (counts only: no id leaves the function).
-- +goose StatementBegin
CREATE FUNCTION purge_soft_deleted(p_batch_size integer)
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

-- Grants first: once the definer owns it, cc_owner (not inheriting) may not grant on it.
REVOKE ALL ON FUNCTION purge_soft_deleted(integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION purge_soft_deleted(integer) TO cc_worker;
-- ALTER OWNER needs the new owner to hold CREATE on the schema: lent for this statement only.
GRANT CREATE ON SCHEMA public TO cc_worker_definer;
ALTER FUNCTION purge_soft_deleted(integer) OWNER TO cc_worker_definer;
REVOKE CREATE ON SCHEMA public FROM cc_worker_definer;

-- +goose Down

-- Only its owner drops the function: cc_owner sets the role, inheriting nothing.
SET LOCAL ROLE cc_worker_definer;
DROP FUNCTION purge_soft_deleted(integer);
RESET ROLE;

REVOKE ALL ON email_verification_failures FROM cc_worker_definer;
DROP POLICY email_verification_failures_purge_delete ON email_verification_failures;
DROP POLICY email_verification_failures_purge_select ON email_verification_failures;
REVOKE ALL ON email_verifications FROM cc_worker_definer;
DROP POLICY email_verifications_purge_delete ON email_verifications;
DROP POLICY email_verifications_purge_select ON email_verifications;
REVOKE ALL ON refresh_tokens FROM cc_worker_definer;
DROP POLICY refresh_tokens_purge_delete ON refresh_tokens;
DROP POLICY refresh_tokens_purge_select ON refresh_tokens;
REVOKE ALL ON personal_deductions FROM cc_worker_definer;
DROP POLICY personal_deductions_purge_delete ON personal_deductions;
DROP POLICY personal_deductions_purge_select ON personal_deductions;
REVOKE ALL ON expenses FROM cc_worker_definer;
DROP POLICY expenses_purge_delete ON expenses;
DROP POLICY expenses_purge_select ON expenses;
REVOKE ALL ON incomes FROM cc_worker_definer;
DROP POLICY incomes_purge_delete ON incomes;
DROP POLICY incomes_purge_select ON incomes;

DROP FUNCTION purge_auth_cutoff();
DROP FUNCTION purge_records_cutoff();
DROP INDEX refresh_tokens_revoked_at_idx;
DROP INDEX refresh_tokens_absolute_expires_at_idx;
