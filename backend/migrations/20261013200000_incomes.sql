-- Incomes (M3-02): the incomes a user records for one taxpayer + regime, and the income
-- basis (cash | accrual) of every catalog regime. Sources: docs/project-specs/
-- MODULE-incomes.md §2 and §5 (DECIDED items), docs/architecture/data-model.md (incomes,
-- RLS policy outline, key constraints).
--
-- Ownership integrity, enforced by the database rather than trusted to the app:
--   * incomes carries user_id with ENABLE + FORCE ROW LEVEL SECURITY and one policy per
--     operation the app runs, keyed on (SELECT app_user_id()); an unset context matches
--     no row;
--   * the composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes makes an
--     income of another user's taxpayer, or of a regime the taxpayer never held,
--     impossible even with forged ids (FK checks bypass RLS, so the FK, not the policy,
--     proves the triple belongs together). Whether the regime is active on the income's
--     dates is a period rule the use case checks (CheckRegime, §2.4);
--   * id, user_id and taxpayer_id are immutable: cc_app has no UPDATE grant on them.
-- Deletion: account deletion (US-A4) cascades through the users FK (the users-FK guard
-- sees incomes directly); taxpayer deletion (MODULE-taxpayers §2.7) cascades
-- taxpayers -> taxpayer_regimes -> incomes. Cascades bypass RLS and need no DELETE grant.
-- Delete from the API is soft (deleted_at, §2.7): cc_app has no DELETE grant or policy,
-- so a soft-deleted income can always be restored until the 30-day purge (M9-08, its own
-- definer function and policy).

-- +goose Up

-- ---------------------------------------------------------------------------------------
-- regimes.income_basis (§2.2): which date an income of the regime is booked on for ISR.
-- Catalog data, not a switch in code. accrual = personas morales of Título II (601, 620,
-- 622, 623, 624: ISR on invoiced_on); cash = every persona física regime, RESICO PM (626),
-- 603 and 610 (ISR on collected_on). IVA is always cash (LIVA 1-B) and needs no column.
-- Changing a regime's basis later must also re-derive incomes.booked_on in that migration.
-- ---------------------------------------------------------------------------------------
ALTER TABLE regimes ADD COLUMN income_basis text;
UPDATE regimes SET income_basis =
  CASE WHEN person_type = 'moral' AND code IN ('601', '620', '622', '623', '624') THEN 'accrual' ELSE 'cash' END;
ALTER TABLE regimes
  ALTER COLUMN income_basis SET NOT NULL,
  ADD CONSTRAINT regimes_income_basis_valid CHECK (income_basis IN ('cash', 'accrual'));
COMMENT ON COLUMN regimes.income_basis IS
  'cash | accrual: incomes book on collected_on (cash) or invoiced_on (accrual) for ISR (MODULE-incomes 2.2).';

-- ---------------------------------------------------------------------------------------
-- incomes: one row per income as the user reads it from the CFDI or bank statement (§2.1).
--   * id           the client's idempotency key (§2.3): the form generates it once per
--                  session; the server falls back to uuidv7() when absent. A retry with the
--                  same id collides on the primary key, across users too (RLS hides the
--                  row, the PK does not): the use case answers 200 or 409.
--   * money        bigint centavos, MXN only; amount_cents is the subtotal before IVA.
--                  The upper bound is the contract's NonNegativeCents.maximum.
--   * iva_rate     rate_16 | rate_8 | rate_0 | exempt. Whether a rate is valid on the
--                  income's date (rate_8 until 2026-12-31) is date-ranged parameter data
--                  the use case checks through RateCatalog, not a constraint (ADR-0002 §3).
--                  The ±100-centavo tolerance of iva_cents vs the rate is also a use-case
--                  rule (§2.5); only the zero-IVA rates are pinned here.
--   * withholdings bounded by their base (§2.6); which counterparties may withhold for a
--                  regime is a use-case rule, except that a persona física or the público
--                  en general never withholds (any regime).
--   * dates        invoiced_on / collected_on, floor 2022-01-01 (the catalog floor); "not
--                  in the future" is a clock rule of the use case. At least one is set;
--                  which one is required follows the regime's basis (below).
--   * booked_on    the ISR date that decides the list period (§2.2, §2.9): invoiced_on for
--                  an accrual regime, collected_on for a cash one. DERIVED by the trigger
--                  incomes_derive_booked_on on every write; cc_app cannot write it. A cash
--                  income without collected_on therefore fails NOT NULL on booked_on: the
--                  database backstop of the use case's 400 required.
--   * cfdi_uuid    folio fiscal, normalized to upper case. Not unique: a duplicate is a
--                  warning the use case computes (OQ-9, partial collections).
--   * description  trimmed, NFC, 1..200 characters, or NULL (empty -> NULL in the use case).
--                  May hold a client's name: never logged (S-5).
--   * deleted_at   soft delete (§2.7): set by DELETE, cleared by restore.
-- ---------------------------------------------------------------------------------------
CREATE TABLE incomes (
  id                 uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id            uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id        uuid        NOT NULL,
  regime_code        text        NOT NULL,
  counterparty_type  text        NOT NULL,
  amount_cents       bigint      NOT NULL,
  iva_rate           text        NOT NULL,
  iva_cents          bigint      NOT NULL,
  isr_withheld_cents bigint      NOT NULL DEFAULT 0,
  iva_withheld_cents bigint      NOT NULL DEFAULT 0,
  invoiced_on        date,
  collected_on       date,
  booked_on          date        NOT NULL,
  platform_activity  text,
  cfdi_uuid          text,
  description        text,
  deleted_at         timestamptz,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now(),
  -- S-2: the regime is one this user's taxpayer holds; deleting the taxpayer (and so its
  -- regimes) deletes its incomes.
  CONSTRAINT incomes_taxpayer_regime_fkey FOREIGN KEY (user_id, taxpayer_id, regime_code)
    REFERENCES taxpayer_regimes (user_id, taxpayer_id, regime_code) ON DELETE CASCADE,
  CONSTRAINT incomes_counterparty_type_valid CHECK (
    counterparty_type IN ('persona_fisica', 'persona_moral', 'publico_general', 'plataforma')),
  -- §2.1: amount >= 1 centavo; 99999999999999 = docs/api/openapi.yaml NonNegativeCents.maximum.
  CONSTRAINT incomes_amount_range CHECK (amount_cents BETWEEN 1 AND 99999999999999),
  CONSTRAINT incomes_iva_range CHECK (iva_cents BETWEEN 0 AND 99999999999999),
  CONSTRAINT incomes_iva_rate_valid CHECK (iva_rate IN ('rate_16', 'rate_8', 'rate_0', 'exempt')),
  CONSTRAINT incomes_zero_rate_has_no_iva CHECK (iva_rate IN ('rate_16', 'rate_8') OR iva_cents = 0),
  -- §2.6 bounds (422 withholding_exceeds_base in the use case).
  CONSTRAINT incomes_isr_withheld_range CHECK (isr_withheld_cents BETWEEN 0 AND amount_cents),
  CONSTRAINT incomes_iva_withheld_range CHECK (iva_withheld_cents BETWEEN 0 AND iva_cents),
  CONSTRAINT incomes_no_withholding_by_individuals CHECK (
    counterparty_type IN ('persona_moral', 'plataforma') OR (isr_withheld_cents = 0 AND iva_withheld_cents = 0)),
  CONSTRAINT incomes_has_a_date CHECK (collected_on IS NOT NULL OR invoiced_on IS NOT NULL),
  CONSTRAINT incomes_invoiced_on_floor CHECK (invoiced_on >= DATE '2022-01-01'),
  CONSTRAINT incomes_collected_on_floor CHECK (collected_on >= DATE '2022-01-01'),
  -- 625 (plataformas tecnológicas) exists for personas físicas only, so the code suffices.
  CONSTRAINT incomes_platform_activity_valid CHECK (
    platform_activity IS NULL
    OR (regime_code = '625' AND platform_activity IN ('transport_delivery', 'lodging', 'goods_services'))),
  CONSTRAINT incomes_cfdi_uuid_format CHECK (
    cfdi_uuid ~ '^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$'),
  CONSTRAINT incomes_description_valid CHECK (
    description <> '' AND description = btrim(description)
    AND char_length(description) <= 200 AND description IS NFC NORMALIZED)
);
COMMENT ON TABLE incomes IS
  'Incomes per taxpayer + regime (RLS on user_id). Soft delete; booked_on derived from regimes.income_basis.';
COMMENT ON COLUMN incomes.id IS
  'Client-generated idempotency key (MODULE-incomes 2.3); uuidv7() when the client sends none.';
COMMENT ON COLUMN incomes.booked_on IS
  'ISR date: invoiced_on (accrual regime) or collected_on (cash). Derived by incomes_derive_booked_on; never written by the app.';
COMMENT ON COLUMN incomes.description IS 'Free text that may name a client: never logged (S-5).';
COMMENT ON COLUMN incomes.cfdi_uuid IS 'CFDI folio fiscal, upper case. Not unique: duplicates are a warning (OQ-9).';

-- Indexes. Each leads with user_id, the RLS predicate; deleted rows are left out of the
-- hot ones, so a soft delete costs nothing at read time. EXPLAIN ANALYZE in the M3-02 PR.
--   * list + totals + ledger ISR aggregates (§2.9, M3-07): period range on booked_on and
--     keyset (booked_on DESC, id DESC), scanned backwards;
CREATE INDEX incomes_booked_idx ON incomes (user_id, taxpayer_id, booked_on, id)
  WHERE deleted_at IS NULL;
--   * ledger IVA aggregates (M3-07): IVA is cash for every regime, so it ranges over
--     collected_on, which differs from booked_on for accrual incomes; pending collections
--     (NULL) never count for IVA;
CREATE INDEX incomes_collected_idx ON incomes (user_id, taxpayer_id, collected_on, id)
  WHERE collected_on IS NOT NULL AND deleted_at IS NULL;
--   * duplicate CFDI UUID warning (§2.8) among the taxpayer's live incomes;
CREATE INDEX incomes_cfdi_uuid_idx ON incomes (user_id, taxpayer_id, cfdi_uuid)
  WHERE cfdi_uuid IS NOT NULL AND deleted_at IS NULL;
--   * the composite FK (cascade from taxpayer_regimes, deleted rows included), the users
--     FK cascade and LatestRecordPeriod per (taxpayer, regime);
CREATE INDEX incomes_taxpayer_regime_idx ON incomes (user_id, taxpayer_id, regime_code);
--   * the purge job (M9-08): soft-deleted rows only.
CREATE INDEX incomes_deleted_at_idx ON incomes (deleted_at) WHERE deleted_at IS NOT NULL;

-- incomes_derive_booked_on(): BEFORE INSERT / UPDATE OF the inputs. Sets booked_on from the
-- basis of the income's regime: the regime is read through the taxpayer_regimes row the
-- composite FK will check (it carries the person type that keys the catalog). SECURITY
-- INVOKER: it reads the caller's rows under the caller's RLS. When that row is not visible
-- the statement is already doomed (RLS WITH CHECK runs first for another user's row; the
-- FK fails after for a forged triple), so it keeps a placeholder date and lets those
-- report the real error. No EXECUTE grant: checked at CREATE TRIGGER time.
-- +goose StatementBegin
CREATE FUNCTION incomes_derive_booked_on() RETURNS trigger
  LANGUAGE plpgsql
  SET search_path = pg_catalog, pg_temp
AS $$
DECLARE
  basis text;
BEGIN
  SELECT r.income_basis INTO basis
  FROM public.taxpayer_regimes tr
  JOIN public.regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
  WHERE tr.user_id = NEW.user_id AND tr.taxpayer_id = NEW.taxpayer_id AND tr.regime_code = NEW.regime_code;

  NEW.booked_on := CASE basis
    WHEN 'accrual' THEN NEW.invoiced_on
    WHEN 'cash' THEN NEW.collected_on
    ELSE coalesce(NEW.collected_on, NEW.invoiced_on)
  END;
  RETURN NEW;
END
$$;
-- +goose StatementEnd
COMMENT ON FUNCTION incomes_derive_booked_on() IS
  'BEFORE INSERT/UPDATE trigger: incomes.booked_on = invoiced_on (accrual regime) or collected_on (cash).';

CREATE TRIGGER incomes_derive_booked_on
  BEFORE INSERT OR UPDATE OF regime_code, invoiced_on, collected_on, booked_on ON incomes
  FOR EACH ROW EXECUTE FUNCTION incomes_derive_booked_on();
CREATE TRIGGER incomes_set_updated_at
  BEFORE UPDATE ON incomes
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE incomes ENABLE ROW LEVEL SECURITY;
ALTER TABLE incomes FORCE ROW LEVEL SECURITY;
-- Intent: a user lists, reads, records, edits, soft-deletes and restores only their own
-- incomes; a write can never hand a row to another user (WITH CHECK). No DELETE policy:
-- see the header.
CREATE POLICY incomes_select_own ON incomes FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY incomes_insert_own ON incomes FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY incomes_update_own ON incomes FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- INSERT names the client id (§2.3), never booked_on, deleted_at or the audit columns.
-- UPDATE covers edit (moving to another regime included, §2.4), soft delete and restore;
-- never id, user_id, taxpayer_id (S-3), booked_on or the audit columns. cc_webhook and
-- cc_worker get nothing.
GRANT SELECT ON incomes TO cc_app;
GRANT INSERT (id, user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
              isr_withheld_cents, iva_withheld_cents, invoiced_on, collected_on, platform_activity, cfdi_uuid,
              description)
  ON incomes TO cc_app;
GRANT UPDATE (regime_code, counterparty_type, amount_cents, iva_rate, iva_cents, isr_withheld_cents,
              iva_withheld_cents, invoiced_on, collected_on, platform_activity, cfdi_uuid, description, deleted_at)
  ON incomes TO cc_app;

-- +goose Down
DROP TABLE incomes;
DROP FUNCTION incomes_derive_booked_on();
ALTER TABLE regimes DROP COLUMN income_basis;
