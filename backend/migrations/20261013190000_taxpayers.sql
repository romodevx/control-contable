-- Taxpayers (M2-03): the contribuyentes a user manages and the regimes each one holds.
-- Sources: docs/project-specs/MODULE-taxpayers.md §2 and §5 (DECIDED items),
-- docs/architecture/data-model.md (taxpayers, taxpayer_regimes, RLS policy outline).
--
-- Ownership integrity, enforced by the database rather than trusted to the app:
--   * both tables carry user_id with ENABLE + FORCE ROW LEVEL SECURITY and one policy per
--     operation keyed on (SELECT app_user_id()); an unset context matches no row;
--   * taxpayer_regimes references its taxpayer through the composite FK
--     (user_id, taxpayer_id, person_type) -> taxpayers (user_id, id, person_type). A child
--     can never point at another user's taxpayer, even with forged ids (FK checks bypass
--     RLS, so the FK, not the policy, is what proves the pair belongs together), and it
--     carries the taxpayer's person type, which (regime_code, person_type) -> regimes then
--     restricts to the regimes of that person type (626 resolves to RESICO PF or PM);
--   * person_type, user_id and id are immutable: cc_app has no UPDATE grant on them and a
--     trigger refuses the change for every role (MODULE-taxpayers §2.1).
-- Account deletion (US-A4) cascades through both tables: each references users directly
-- ON DELETE CASCADE (so the users-FK guard sees them), and deleting a taxpayer (§2.7)
-- cascades to its regimes. Cascades bypass RLS and need no DELETE grant on the children.

-- +goose Up

-- ---------------------------------------------------------------------------------------
-- taxpayers: one row per RFC a user manages (OD-1). The RFC is never stored in clear
-- (OD-5, platform/crypto):
--   * rfc_encrypted AES-256-GCM ciphertext framed by crypto.Keyring: key_id (2 bytes, big
--                   endian) || nonce (12) || ciphertext || tag (16). rfc_key_id repeats the
--                   prefix so key rotation selects stale rows by a plain column; a CHECK
--                   keeps the two from drifting.
--                   Row binding (P0-30 D-4, MODULE-taxpayers S-1): the GCM additional data
--                   is key_id || taxpayers.id || user_id. The use case generates the row id
--                   (uuidv7) BEFORE encrypting and inserts it explicitly; it never relies on
--                   the column default. A ciphertext copied to another row or another user
--                   therefore fails to decrypt. An RFC correction (OQ-1) re-encrypts with
--                   the same row id.
--   * rfc_hash      HMAC-SHA256(k_rfc_index, user_id || normalized RFC) (S-2): the
--                   per-user uniqueness key; including user_id keeps the same RFC of two
--                   users unlinkable.
--   * rfc_masked    the only form ever returned (S-3): PF AAAA******XX9, PM AAA******XX9.
--   * rfc_check_digit_valid  whether the SAT módulo 11 check digit matched. A mismatch is
--                   ACCEPTED (D-RFC: ~1.5% of real SAT-issued RFCs fail the published
--                   algorithm) and answered with the non-blocking rfc_check_digit_mismatch
--                   warning, so a stored RFC is NOT necessarily check-digit valid. No
--                   default: every write of the RFC states the verdict. The warning itself
--                   is still never stored nor returned on reads (contract 0.4.0).
-- The domain (M2-04) rejects bad length, alphabet, date and the generic RFCs before the
-- RFC reaches this table; the check digit is the one rule it only reports.
-- ---------------------------------------------------------------------------------------
CREATE TABLE taxpayers (
  id            uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id       uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  person_type   text        NOT NULL,
  display_name  text        NOT NULL,
  rfc_encrypted bytea       NOT NULL,
  rfc_key_id    smallint    NOT NULL,
  rfc_hash      bytea       NOT NULL,
  rfc_masked    text        NOT NULL,
  rfc_check_digit_valid boolean NOT NULL,
  archived_at   timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),
  -- US-T2: one RFC once per user (archived included); another user may hold the same RFC.
  CONSTRAINT taxpayers_user_rfc_hash_key UNIQUE (user_id, rfc_hash),
  -- Composite-FK target for user-scoped children (data-model "Ownership"). Its btree also
  -- serves the RLS predicate, the list by user, the cap count and the users FK cascade.
  CONSTRAINT taxpayers_user_id_id_key UNIQUE (user_id, id),
  -- Composite-FK target of taxpayer_regimes: carries the immutable person type.
  CONSTRAINT taxpayers_user_id_id_person_type_key UNIQUE (user_id, id, person_type),
  CONSTRAINT taxpayers_person_type_valid CHECK (person_type IN ('fisica', 'moral')),
  -- Contract: 1-80 characters, trimmed, NFC (US-T3c).
  CONSTRAINT taxpayers_display_name_valid CHECK (
    display_name <> '' AND display_name = btrim(display_name)
    AND char_length(display_name) <= 80 AND display_name IS NFC NORMALIZED),
  CONSTRAINT taxpayers_rfc_key_id_range CHECK (rfc_key_id >= 1),
  -- More than the 30-byte frame (2 key_id + 12 nonce + 16 tag), and the key_id prefix
  -- equals rfc_key_id. CASE guards get_byte against a too-short value.
  CONSTRAINT taxpayers_rfc_ciphertext_framed CHECK (
    CASE WHEN octet_length(rfc_encrypted) > 30
      THEN rfc_key_id = (get_byte(rfc_encrypted, 0) << 8) | get_byte(rfc_encrypted, 1)
      ELSE false
    END),
  -- 32 = crypto.HashSize (HMAC-SHA256).
  CONSTRAINT taxpayers_rfc_hash_size CHECK (octet_length(rfc_hash) = 32),
  -- The mask hides the six date digits and matches the person type's RFC length
  -- (física 13 = 4 letters, moral 12 = 3 letters, + date + homoclave + check digit).
  CONSTRAINT taxpayers_rfc_masked_format CHECK (
    CASE person_type
      WHEN 'fisica' THEN rfc_masked ~ '^[A-ZÑ&]{4}\*{6}[A-Z0-9]{2}[0-9A]$'
      WHEN 'moral' THEN rfc_masked ~ '^[A-ZÑ&]{3}\*{6}[A-Z0-9]{2}[0-9A]$'
      ELSE false
    END)
);
COMMENT ON TABLE taxpayers IS
  'Taxpayers (RFCs) a user manages (RLS on user_id). RFC kept as AES-GCM ciphertext + HMAC blind index + mask.';
COMMENT ON COLUMN taxpayers.rfc_encrypted IS
  'AES-256-GCM, crypto.Keyring frame. AAD = key_id || taxpayers.id || user_id (row binding, S-1).';
COMMENT ON COLUMN taxpayers.rfc_check_digit_valid IS
  'SAT módulo 11 check digit matched. false = accepted with the rfc_check_digit_mismatch warning (D-RFC).';
COMMENT ON COLUMN taxpayers.person_type IS
  'fisica | moral. Immutable: no UPDATE grant, trigger taxpayers_forbid_identity_change.';

-- taxpayers_forbid_identity_change(): BEFORE UPDATE trigger. The grants already keep cc_app
-- off id, user_id and person_type; this refuses the change for any role that is ever
-- granted more (or the owner), so the identity of a taxpayer and the person type its
-- regimes, RFC length and mask depend on never change. No EXECUTE grant: checked at
-- CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION taxpayers_forbid_identity_change() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id
     OR NEW.user_id IS DISTINCT FROM OLD.user_id
     OR NEW.person_type IS DISTINCT FROM OLD.person_type THEN
    RAISE EXCEPTION 'taxpayers.id, user_id and person_type are immutable'
      USING ERRCODE = 'P0001', TABLE = 'taxpayers',
            HINT = 'Delete the taxpayer and create it again (MODULE-taxpayers 2.1).';
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION taxpayers_forbid_identity_change() IS
  'BEFORE UPDATE trigger: taxpayers.id, user_id and person_type never change.';

CREATE TRIGGER taxpayers_forbid_identity_change
  BEFORE UPDATE ON taxpayers
  FOR EACH ROW EXECUTE FUNCTION taxpayers_forbid_identity_change();
CREATE TRIGGER taxpayers_set_updated_at
  BEFORE UPDATE ON taxpayers
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE taxpayers ENABLE ROW LEVEL SECURITY;
ALTER TABLE taxpayers FORCE ROW LEVEL SECURITY;
-- Intent: a user lists, reads, creates, renames, archives, corrects the RFC of and deletes
-- only their own taxpayers; a write can never hand a row to another user (WITH CHECK).
CREATE POLICY taxpayers_select_own ON taxpayers FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayers_insert_own ON taxpayers FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayers_update_own ON taxpayers FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayers_delete_own ON taxpayers FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
-- INSERT names the id (S-1: the use case generates it before encrypting). UPDATE covers
-- rename, archive/unarchive and the RFC correction (OQ-1: the five rfc_* columns move
-- together; the use case re-validates, re-encrypts, re-hashes and restates the check-digit
-- verdict). Never person_type,
-- user_id, id or the audit columns. cc_webhook and cc_worker get nothing.
GRANT SELECT, DELETE ON taxpayers TO cc_app;
GRANT INSERT (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
              rfc_check_digit_valid)
  ON taxpayers TO cc_app;
GRANT UPDATE (display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked, rfc_check_digit_valid, archived_at)
  ON taxpayers TO cc_app;

-- ---------------------------------------------------------------------------------------
-- taxpayer_count_for_create(): the atomic cap check of CreateTaxpayer (OD-2, S-5,
-- MODULE-taxpayers §2.8). Called inside the create transaction, before the insert:
--
--   n := SELECT taxpayer_count_for_create();   -- locks the user, then counts
--   access.Policy.CheckTaxpayerLimit(n)         -- 422 taxpayer_limit_reached at 10
--   INSERT INTO taxpayers ...                    -- same transaction
--
-- It locks the context user's row (FOR NO KEY UPDATE: conflicts with itself, so this
-- user's creates run one at a time, but not with the KEY SHARE locks that foreign-key
-- checks on users take, so other writes are not blocked), then counts every taxpayer of
-- the user, archived included (OQ-9). A concurrent create waits for the lock and, once
-- it holds it, counts with a fresh snapshot (READ COMMITTED: each statement of a volatile
-- function takes a new one), so it sees the row the first create committed: at 9, five
-- concurrent creates admit exactly one.
--
-- Fails closed:
--   * REPEATABLE READ is refused (25000): its single snapshot would count without the
--     rows committed while waiting for the lock. READ COMMITTED (the TxManager default)
--     and SERIALIZABLE (SSI aborts the write skew) are safe;
--   * no user in context, or a user RLS hides: no row to lock -> P0002.
-- The cap itself (10) stays in the domain (TaxpayerCap); this function only provides the
-- lock and the count. SECURITY INVOKER: it reads the caller's rows under the caller's RLS.
-- ---------------------------------------------------------------------------------------
-- +goose StatementBegin
CREATE FUNCTION taxpayer_count_for_create() RETURNS integer
  LANGUAGE plpgsql VOLATILE
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  owned integer;
BEGIN
  IF current_setting('transaction_isolation') = 'repeatable read' THEN
    RAISE EXCEPTION 'taxpayer_count_for_create() cannot run under REPEATABLE READ'
      USING ERRCODE = '25000',
            HINT = 'Its snapshot would miss creates committed while waiting for the lock; use READ COMMITTED.';
  END IF;
  PERFORM 1 FROM public.users u WHERE u.id = (SELECT public.app_user_id()) FOR NO KEY UPDATE;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'taxpayer_count_for_create() needs the user context'
      USING ERRCODE = 'P0002';
  END IF;
  SELECT count(*) INTO owned FROM public.taxpayers t WHERE t.user_id = (SELECT public.app_user_id());
  RETURN owned;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION taxpayer_count_for_create() IS
  'Cap check of CreateTaxpayer: locks the context user (FOR NO KEY UPDATE), returns its taxpayer count.';
GRANT EXECUTE ON FUNCTION taxpayer_count_for_create() TO cc_app;

-- ---------------------------------------------------------------------------------------
-- taxpayer_regimes: the regimes a taxpayer holds, each over a window of periods
-- [start_period, end_period] (YYYY-MM, open end = still active; MODULE-taxpayers §2.6).
-- PK (taxpayer_id, regime_code): one row per regime, so one regime never overlaps
-- itself; re-adding an archived regime reopens the same row (OQ-5), which keeps the
-- (taxpayer_id, regime_code, user_id) key that incomes/expenses (M3/M4) will reference.
--   * options: per-regime options, a JSON object validated against
--     regimes.option_defs by the use case (§2.4);
--   * needs_cpa_review_ack: the user acknowledged the catalog's non-blocking CPA-review
--     warning (regimes.needs_cpa_review or a regime_compatibility pair allowed with a
--     warning) when adding or editing this regime; written by M2-06.
-- Incompatible regimes (regime_compatibility.allowed = false) may not overlap in time:
-- the use case answers 422 invalid_regime_combination first (M2-06); the trigger below is
-- the database backstop (SQLSTATE 23P01).
-- ---------------------------------------------------------------------------------------
CREATE TABLE taxpayer_regimes (
  taxpayer_id          uuid        NOT NULL,
  regime_code          text        NOT NULL,
  user_id              uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  person_type          text        NOT NULL,
  options              jsonb       NOT NULL DEFAULT '{}',
  start_period         char(7)     NOT NULL,
  end_period           char(7),
  needs_cpa_review_ack boolean     NOT NULL DEFAULT false,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (taxpayer_id, regime_code),
  -- Composite-FK target for incomes, expenses and taxpayer_tax_inputs (data-model). Its
  -- btree leads with user_id: it serves the RLS predicate and the taxpayer FK below.
  CONSTRAINT taxpayer_regimes_user_taxpayer_regime_key UNIQUE (user_id, taxpayer_id, regime_code),
  -- The taxpayer belongs to this user and has this person type (see the header).
  CONSTRAINT taxpayer_regimes_taxpayer_fkey FOREIGN KEY (user_id, taxpayer_id, person_type)
    REFERENCES taxpayers (user_id, id, person_type) ON DELETE CASCADE,
  -- The regime exists for the taxpayer's person type. Catalog rows change only through
  -- migrations, so the referencing side needs no index of its own.
  CONSTRAINT taxpayer_regimes_regime_fkey FOREIGN KEY (regime_code, person_type)
    REFERENCES regimes (code, person_type),
  -- 2022-01 is the catalog floor (§2.6); the upper bound (current month) is a clock rule
  -- the use case checks.
  CONSTRAINT taxpayer_regimes_start_period_valid CHECK (
    start_period ~ '^[0-9]{4}-(0[1-9]|1[0-2])$' AND start_period >= '2022-01'),
  CONSTRAINT taxpayer_regimes_end_period_valid CHECK (
    end_period IS NULL OR (end_period ~ '^[0-9]{4}-(0[1-9]|1[0-2])$' AND end_period >= start_period)),
  CONSTRAINT taxpayer_regimes_options_object CHECK (jsonb_typeof(options) = 'object')
);
COMMENT ON TABLE taxpayer_regimes IS
  'Regimes per taxpayer over a period window (RLS on user_id). Composite FKs bind owner, person type and catalog.';

-- taxpayer_regimes_check_compatible(): BEFORE INSERT / UPDATE OF the period window. Raises
-- 23P01 (exclusion_violation, constraint taxpayer_regimes_compatible_periods) when the
-- new window overlaps the window of another regime of the same taxpayer that
-- regime_compatibility marks as not allowed. Two windows [s1, e1] and [s2, e2] overlap iff
-- s1 <= e2 AND s2 <= e1, an open end being +infinity ('9999-12'); YYYY-MM strings compare
-- in calendar order.
-- A declarative EXCLUDE constraint cannot express this: which pairs conflict is catalog
-- data (regime_compatibility), not a property of the row. So the trigger first locks the
-- parent taxpayer (FOR NO KEY UPDATE: serializes regime writes of one taxpayer without
-- blocking FK checks), then looks for a clash with a fresh READ COMMITTED snapshot, so
-- two concurrent adds of 612 and 626 cannot both pass. REPEATABLE READ is refused (25000),
-- as in taxpayer_count_for_create(): its snapshot predates the wait for the lock, so it
-- would miss a clashing regime committed meanwhile; SERIALIZABLE (SSI) is safe. SECURITY INVOKER: it sees only the
-- caller's rows (RLS), so it can neither leak nor be fooled by another user's data.
-- No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION taxpayer_regimes_check_compatible() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  clash_code   text;
  clash_reason text;
BEGIN
  IF current_setting('transaction_isolation') = 'repeatable read' THEN
    RAISE EXCEPTION 'taxpayer_regimes writes cannot run under REPEATABLE READ'
      USING ERRCODE = '25000',
            HINT = 'Its snapshot would miss an incompatible regime committed while waiting for the lock; use READ COMMITTED.';
  END IF;
  PERFORM 1 FROM public.taxpayers t
  WHERE t.id = NEW.taxpayer_id AND t.user_id = NEW.user_id
  FOR NO KEY UPDATE;

  SELECT o.regime_code, c.reason INTO clash_code, clash_reason
  FROM public.taxpayer_regimes o
  JOIN public.regime_compatibility c
    ON c.person_type = NEW.person_type
   AND c.regime_a = least(o.regime_code, NEW.regime_code)
   AND c.regime_b = greatest(o.regime_code, NEW.regime_code)
  WHERE o.taxpayer_id = NEW.taxpayer_id
    AND o.user_id = NEW.user_id
    AND o.regime_code <> NEW.regime_code
    AND NOT c.allowed
    AND o.start_period <= coalesce(NEW.end_period, '9999-12')
    AND NEW.start_period <= coalesce(o.end_period, '9999-12')
  ORDER BY o.regime_code
  LIMIT 1;

  IF clash_code IS NOT NULL THEN
    RAISE EXCEPTION 'regimes % and % may not overlap', least(clash_code, NEW.regime_code),
                    greatest(clash_code, NEW.regime_code)
      USING ERRCODE = '23P01', CONSTRAINT = 'taxpayer_regimes_compatible_periods',
            TABLE = 'taxpayer_regimes', DETAIL = clash_reason;
  END IF;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION taxpayer_regimes_check_compatible() IS
  'BEFORE INSERT/UPDATE trigger: incompatible regimes of one taxpayer never overlap in time (23P01).';

CREATE TRIGGER taxpayer_regimes_check_compatible
  BEFORE INSERT OR UPDATE OF start_period, end_period ON taxpayer_regimes
  FOR EACH ROW EXECUTE FUNCTION taxpayer_regimes_check_compatible();
CREATE TRIGGER taxpayer_regimes_set_updated_at
  BEFORE UPDATE ON taxpayer_regimes
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE taxpayer_regimes ENABLE ROW LEVEL SECURITY;
ALTER TABLE taxpayer_regimes FORCE ROW LEVEL SECURITY;
-- Intent: a user reads, adds and edits only the regimes of their own taxpayers. No DELETE
-- policy or grant: DELETE on a regime archives it (end_period, §2.7); rows go away only
-- with their taxpayer or user (cascade).
CREATE POLICY taxpayer_regimes_select_own ON taxpayer_regimes FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayer_regimes_insert_own ON taxpayer_regimes FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY taxpayer_regimes_update_own ON taxpayer_regimes FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- UPDATE covers option changes, archive (end_period) and reopen (end_period = NULL,
-- start_period kept, OQ-5). The keys, person type and start_period never change.
GRANT SELECT ON taxpayer_regimes TO cc_app;
GRANT INSERT (taxpayer_id, regime_code, user_id, person_type, options, start_period, end_period, needs_cpa_review_ack)
  ON taxpayer_regimes TO cc_app;
GRANT UPDATE (options, end_period, needs_cpa_review_ack) ON taxpayer_regimes TO cc_app;

-- +goose Down
DROP TABLE taxpayer_regimes;
DROP FUNCTION taxpayer_regimes_check_compatible();
DROP FUNCTION taxpayer_count_for_create();
DROP TABLE taxpayers;
DROP FUNCTION taxpayers_forbid_identity_change();
