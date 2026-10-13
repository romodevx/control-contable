-- Sessions and consents (M1-02): rotating refresh tokens (ADR-0001 O6) and the LFPDPPP
-- consent record (PRD §8). Both are user-scoped: ENABLE + FORCE ROW LEVEL SECURITY and
-- one policy per operation keyed on (SELECT app_user_id()), an initplan evaluated once
-- per query. An unset context is NULL and matches no row.
--
-- Refresh lookup under RLS (no bypass, no definer function): the token is
-- `<user_id>.<secret>`. The API parses the user id, sets app.current_user_id to it, and
-- only then looks up token_hash = sha256 of the canonical token text (domain
-- RefreshToken.Hash: it binds the user id too). A forged prefix puts the lookup under a
-- user who does not own that hash, so it finds nothing.
--
-- Rotation (M1-08) is one atomic claim, which also settles concurrent refreshes:
--   UPDATE refresh_tokens SET rotated_at = now()
--   WHERE token_hash = $1 AND rotated_at IS NULL AND revoked_at IS NULL
--     AND expires_at > now() AND absolute_expires_at > now()
--   RETURNING id, family_id, absolute_expires_at;
-- 0 rows and the hash exists (rotated or revoked) = reuse: revoke the family with
--   UPDATE ... SET revoked_at = now(), revoke_reason = 'reuse_detected'
--   WHERE family_id = $1 AND revoked_at IS NULL;
--
-- Not here, by design:
--   * No FK consents -> legal_documents: that table belongs to M1-03, built in parallel.
--     The FK is added by whichever of the two lands second (see the PR).
--   * No IP address or user agent on consents, no device label on tokens: the data model
--     does not ask for them, and LFPDPPP minimization says not to keep what is not used.
--   * No cc_worker access: the data model grants it nothing here. Expired tokens are
--     deleted by cc_app under each user's context (DELETE policy below).

-- +goose Up

-- ---------------------------------------------------------------------------------------
-- refresh_tokens: one row per issued refresh token. Only the SHA-256 of the canonical
-- "<user_id>.<secret>" text (256-bit secret) is stored, never the token. A login starts
-- a family; every rotation inserts a successor in the same family with rotated_from
-- pointing at its parent and the parent's absolute expiry. Account deletion (US-A4) cascades.
-- ---------------------------------------------------------------------------------------
CREATE TABLE refresh_tokens (
  id                  uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id             uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  family_id           uuid        NOT NULL,
  token_hash          bytea       NOT NULL,
  rotated_from        uuid,
  expires_at          timestamptz NOT NULL,
  absolute_expires_at timestamptz NOT NULL,
  rotated_at          timestamptz,
  revoked_at          timestamptz,
  revoke_reason       text,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  -- Refresh lookup; its btree serves the equality probe.
  CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash),
  -- Target of the parent FK: a successor shares its parent's user and family.
  CONSTRAINT refresh_tokens_id_user_family_key UNIQUE (id, user_id, family_id),
  -- Deleting a parent keeps its successor and clears only the link (PostgreSQL 15+
  -- column list; a plain SET NULL would null user_id and family_id too).
  CONSTRAINT refresh_tokens_rotated_from_fkey FOREIGN KEY (rotated_from, user_id, family_id)
    REFERENCES refresh_tokens (id, user_id, family_id) ON DELETE SET NULL (rotated_from),
  -- 32 = sha256 output.
  CONSTRAINT refresh_tokens_token_hash_size CHECK (octet_length(token_hash) = 32),
  -- ADR-0001 O6: 30-day sliding window inside a 90-day absolute limit. Hours, not days,
  -- so the bound does not depend on the session TimeZone across a DST change. A
  -- successor inherits the family's absolute expiry, which is never later than its own
  -- created_at + 90 days.
  CONSTRAINT refresh_tokens_expiry_window CHECK (
    expires_at > created_at
    AND expires_at <= created_at + interval '720 hours'
    AND expires_at <= absolute_expires_at
    AND absolute_expires_at <= created_at + interval '2160 hours'),
  CONSTRAINT refresh_tokens_rotated_after_issue CHECK (rotated_at IS NULL OR rotated_at >= created_at),
  CONSTRAINT refresh_tokens_revoked_after_issue CHECK (revoked_at IS NULL OR revoked_at >= created_at),
  -- A revocation always says why, and only revoked tokens carry a reason. coalesce:
  -- `NULL IN (...)` is NULL, and a NULL CHECK passes.
  CONSTRAINT refresh_tokens_revocation_reason CHECK (
    (revoked_at IS NULL AND revoke_reason IS NULL)
    OR (revoked_at IS NOT NULL
        AND coalesce(revoke_reason IN ('logout', 'reuse_detected', 'password_changed', 'admin'), false)))
);
COMMENT ON TABLE refresh_tokens IS
  'Rotating refresh tokens (RLS on user_id). Stores sha256 of the token text only; grouped by family_id.';
COMMENT ON COLUMN refresh_tokens.created_at IS 'Issued at.';

-- Revoke a family in one statement (data-model.md); the leading user_id matches the RLS
-- predicate, and the same index serves "revoke every session of this user".
CREATE INDEX refresh_tokens_user_family_idx ON refresh_tokens (user_id, family_id);
-- The parent FK's ON DELETE scan probes rotated_from; without an index it seq-scans.
CREATE INDEX refresh_tokens_rotated_from_idx ON refresh_tokens (rotated_from)
  WHERE rotated_from IS NOT NULL;

-- refresh_tokens_history_is_final(): BEFORE UPDATE trigger. Once a token is rotated or
-- revoked that fact is history: reuse detection depends on it, so a buggy or hostile
-- update cannot resurrect a token. No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION refresh_tokens_history_is_final() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog
AS $$
BEGIN
  IF OLD.rotated_at IS NOT NULL AND NEW.rotated_at IS DISTINCT FROM OLD.rotated_at THEN
    RAISE EXCEPTION 'refresh token rotation is final'
      USING ERRCODE = 'P0001', HINT = 'A rotated token is never claimable again (ADR-0001 O6).';
  END IF;
  IF OLD.revoked_at IS NOT NULL
     AND (NEW.revoked_at IS DISTINCT FROM OLD.revoked_at OR NEW.revoke_reason IS DISTINCT FROM OLD.revoke_reason) THEN
    RAISE EXCEPTION 'refresh token revocation is final'
      USING ERRCODE = 'P0001', HINT = 'Filter revocations with revoked_at IS NULL.';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION refresh_tokens_history_is_final() IS
  'BEFORE UPDATE trigger: rotated_at and the revocation never change once set.';

CREATE TRIGGER refresh_tokens_history_is_final
  BEFORE UPDATE ON refresh_tokens
  FOR EACH ROW EXECUTE FUNCTION refresh_tokens_history_is_final();
CREATE TRIGGER refresh_tokens_set_updated_at
  BEFORE UPDATE ON refresh_tokens
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens FORCE ROW LEVEL SECURITY;
-- Intent: the API issues, looks up, rotates, revokes and deletes only the context user's
-- tokens. The context comes from the user id embedded in the token itself (see header).
CREATE POLICY refresh_tokens_select_own ON refresh_tokens FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY refresh_tokens_insert_own ON refresh_tokens FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY refresh_tokens_update_own ON refresh_tokens FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- Logout may delete instead of revoke, and expired tokens are cleaned up per user.
CREATE POLICY refresh_tokens_delete_own ON refresh_tokens FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
-- Tokens are born live (no rotated_at / revoked_at on INSERT). Identity, hash and expiry
-- are immutable; only the rotation and revocation columns change. cc_webhook and
-- cc_worker get nothing.
GRANT SELECT, DELETE ON refresh_tokens TO cc_app;
GRANT INSERT (id, user_id, family_id, token_hash, rotated_from, expires_at, absolute_expires_at)
  ON refresh_tokens TO cc_app;
GRANT UPDATE (rotated_at, revoked_at, revoke_reason) ON refresh_tokens TO cc_app;

-- ---------------------------------------------------------------------------------------
-- consents: which version of each legal document a user accepted, and when (LFPDPPP).
-- Append-only evidence: written once by verifyEmail (or on re-acceptance of a new
-- version), never updated, removed only with the account (cascade; PRD Q6 pending
-- legal review, OB10).
-- ---------------------------------------------------------------------------------------
CREATE TABLE consents (
  id               uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id          uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  document_kind    text        NOT NULL,
  document_version text        NOT NULL,
  accepted_at      timestamptz NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  -- One acceptance per document version; a retried verifyEmail cannot duplicate it.
  -- Its btree also serves "latest accepted version of kind X for this user".
  CONSTRAINT consents_user_document_key UNIQUE (user_id, document_kind, document_version),
  CONSTRAINT consents_document_kind_valid CHECK (document_kind IN ('privacy', 'terms')),
  CONSTRAINT consents_document_version_present CHECK (
    btrim(document_version) <> '' AND char_length(document_version) <= 32),
  -- Accepted at sign-up (pending in email_verifications) and recorded at verification,
  -- so accepted_at may precede created_at but never follows it. 5 minutes absorbs
  -- app/DB clock skew.
  CONSTRAINT consents_accepted_not_in_future CHECK (accepted_at <= created_at + interval '5 minutes')
);
COMMENT ON TABLE consents IS
  'Accepted legal document versions (RLS on user_id). Append-only; no IP or user agent kept.';

ALTER TABLE consents ENABLE ROW LEVEL SECURITY;
ALTER TABLE consents FORCE ROW LEVEL SECURITY;
-- Intent: a user records and reads only their own consents. No UPDATE or DELETE policy
-- or grant: the record is evidence and goes away only with its user (cascade).
CREATE POLICY consents_select_own ON consents FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY consents_insert_own ON consents FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
GRANT SELECT ON consents TO cc_app;
GRANT INSERT (user_id, document_kind, document_version, accepted_at) ON consents TO cc_app;

-- +goose Down
DROP TABLE consents;
DROP TABLE refresh_tokens;
DROP FUNCTION refresh_tokens_history_is_final();
