-- Expenses (M4-03): the gastos of a taxpayer, one row per payment.
-- Sources: docs/project-specs/MODULE-expenses.md §2.1, §2.4, §2.6, §5 (SE-1, SE-2), §8
-- (data-model deltas), MODULE-incomes.md §2.1, §2.3, §2.5, §2.7 (shared rules) and
-- docs/architecture/data-model.md (expenses, RLS policy outline, indexes).
--
-- Ownership integrity, enforced by the database rather than trusted to the app:
--   * ENABLE + FORCE ROW LEVEL SECURITY with one policy per operation keyed on
--     (SELECT app_user_id()); an unset context matches no row;
--   * the composite FK (user_id, taxpayer_id, regime_code) -> taxpayer_regimes proves the
--     expense belongs to a regime its own taxpayer holds: FK checks bypass RLS, so even
--     forged ids can never point at another user's taxpayer or at a regime the taxpayer
--     does not hold. Whether the regime is held ON paid_on (its period window) is a use
--     case rule (MODULE-incomes §2.4);
--   * category_code -> expense_categories (SE-1): an unknown category never lands.
-- Account deletion (US-A4) and taxpayer deletion cascade here: user_id references users
-- ON DELETE CASCADE (so the users-FK guard sees it) and the composite FK cascades from
-- taxpayer_regimes, which cascades from taxpayers. Cascades bypass RLS and need no DELETE
-- grant.
--
-- Field rules (§2.1):
--   * id: the client-generated idempotency key (uuid, any version); the server default is
--     uuidv7 when the client sends none. A repeated id is a primary-key conflict;
--   * money in MXN centavos, bounded by docs/api/openapi.yaml NonNegativeCents
--     (0..99,999,999,999,999); amount_cents is the subtotal before IVA and >= 1;
--   * rate_0 / exempt carry no IVA. The rate's validity on paid_on and the iva_cents
--     tolerance against the rate are use-case rules (dated RateCatalog data, ADR-0002 §3);
--   * withholdings the taxpayer made to its supplier (§2.6): never above their base. Only
--     a persona moral may make them: a person-type rule of the use case (422);
--   * paid_on is the only date, and the booking date of every regime in v1's monthly
--     scope: PF regimes (LISR 103/105, 115) and RESICO PM (208, 211) deduct when paid,
--     creditable IVA is paid basis (LIVA 5 fr. III), and the Título II PM provisional ISR
--     (601, 622; art. 14) uses the profit coefficient, so their accrual-basis deductions
--     never enter a monthly figure. An annual PM computation would need an accrual date
--     (CPA-REVIEW, EXP). 2022-01-01 is the catalog floor; "not after today" is a clock
--     rule of the use case;
--   * cfdi_uuid: the CFDI folio fiscal, normalized to upper case, only with has_cfdi
--     (SE-2). NOT unique: a duplicate is the duplicate_cfdi_uuid warning;
--   * deductible: the effective verdict the use case computed (§2.4: category default AND
--     has_cfdi AND NOT the cash rule, unless overridden); deductible_overridden: the user
--     set it (deductible_source = user). No default for deductible: every write states it;
--   * deleted_at: soft delete (Undo / restore); the purge job (M9-08) removes rows after
--     30 days.

-- +goose Up

CREATE TABLE expenses (
  id                    uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id               uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id           uuid        NOT NULL,
  regime_code           text        NOT NULL,
  category_code         text        NOT NULL REFERENCES expense_categories (code),
  description           text,
  amount_cents          bigint      NOT NULL,
  iva_rate              text        NOT NULL,
  iva_cents             bigint      NOT NULL,
  payment_method        text        NOT NULL,
  has_cfdi              boolean     NOT NULL,
  cfdi_uuid             text,
  deductible            boolean     NOT NULL,
  deductible_overridden boolean     NOT NULL DEFAULT false,
  isr_withheld_cents    bigint      NOT NULL DEFAULT 0,
  iva_withheld_cents    bigint      NOT NULL DEFAULT 0,
  paid_on               date        NOT NULL,
  deleted_at            timestamptz,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  -- The regime is one the user's own taxpayer holds (see the header).
  CONSTRAINT expenses_taxpayer_regime_fkey FOREIGN KEY (user_id, taxpayer_id, regime_code)
    REFERENCES taxpayer_regimes (user_id, taxpayer_id, regime_code) ON DELETE CASCADE,
  CONSTRAINT expenses_amount_cents_range CHECK (amount_cents BETWEEN 1 AND 99999999999999),
  CONSTRAINT expenses_iva_cents_range CHECK (iva_cents BETWEEN 0 AND 99999999999999),
  -- docs/api/openapi.yaml IvaRate.
  CONSTRAINT expenses_iva_rate_valid CHECK (iva_rate IN ('rate_16', 'rate_8', 'rate_0', 'exempt')),
  CONSTRAINT expenses_untaxed_has_no_iva CHECK (iva_rate IN ('rate_16', 'rate_8') OR iva_cents = 0),
  -- docs/api/openapi.yaml PaymentMethod; cash drives the LISR 27 fr. III rule (§2.5).
  CONSTRAINT expenses_payment_method_valid CHECK (payment_method IN (
    'cash', 'transfer', 'card_debit', 'card_credit', 'check', 'e_wallet', 'other')),
  CONSTRAINT expenses_isr_withheld_within_base CHECK (isr_withheld_cents BETWEEN 0 AND amount_cents),
  CONSTRAINT expenses_iva_withheld_within_base CHECK (iva_withheld_cents BETWEEN 0 AND iva_cents),
  -- SE-2: a folio fiscal only on an expense with a CFDI; 8-4-4-4-12 hex, upper case.
  CONSTRAINT expenses_cfdi_uuid_needs_cfdi CHECK (cfdi_uuid IS NULL OR has_cfdi),
  CONSTRAINT expenses_cfdi_uuid_format CHECK (
    cfdi_uuid IS NULL OR cfdi_uuid ~ '^[0-9A-F]{8}-([0-9A-F]{4}-){3}[0-9A-F]{12}$'),
  CONSTRAINT expenses_paid_on_floor CHECK (paid_on >= DATE '2022-01-01'),
  -- Contract: up to 200 characters, trimmed, NFC; empty is stored as NULL.
  CONSTRAINT expenses_description_valid CHECK (
    description IS NULL OR (
      description <> '' AND description = btrim(description)
      AND char_length(description) <= 200 AND description IS NFC NORMALIZED))
);
COMMENT ON TABLE expenses IS
  'Expenses per taxpayer and regime (RLS on user_id). Composite FK to taxpayer_regimes; soft delete via deleted_at.';
COMMENT ON COLUMN expenses.id IS
  'Client-generated idempotency key; server default uuidv7.';
COMMENT ON COLUMN expenses.deductible IS
  'Effective verdict computed by the use case (category default AND has_cfdi AND NOT cash rule, or the user override).';
COMMENT ON COLUMN expenses.deductible_overridden IS
  'true = the user set deductible (deductible_source = user); false = category default, recomputed on edit.';
COMMENT ON COLUMN expenses.cfdi_uuid IS
  'CFDI folio fiscal, upper case. Not unique: a duplicate is the duplicate_cfdi_uuid warning.';
COMMENT ON COLUMN expenses.isr_withheld_cents IS
  'ISR the taxpayer (persona moral) withheld from the supplier; remitted, never netted (ADR-0002 §1).';

-- Indexes. Every one leads with user_id: the RLS predicate (user_id = app_user_id()) is
-- an index condition, never a filter over other users' rows.
--   * live by paid_on: the list (period filter, keyset ORDER BY paid_on DESC, id DESC),
--     the month totals and the M5 estimate aggregate (paid_on range per taxpayer). Partial:
--     soft-deleted rows leave every read at once (MODULE-incomes §2.7);
--   * live by folio: the duplicate_cfdi_uuid warning lookup;
--   * deleted_at: the purge job's range scan (M9-08), only over deleted rows;
--   * the composite FK: cascades from taxpayer_regimes (taxpayer deletion) and users
--     (account deletion) must reach soft-deleted rows too, which the partial indexes skip.
CREATE INDEX expenses_live_paid_on_idx ON expenses (user_id, taxpayer_id, paid_on, id)
  WHERE deleted_at IS NULL;
CREATE INDEX expenses_live_cfdi_uuid_idx ON expenses (user_id, taxpayer_id, cfdi_uuid)
  WHERE cfdi_uuid IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX expenses_deleted_at_idx ON expenses (deleted_at)
  WHERE deleted_at IS NOT NULL;
CREATE INDEX expenses_taxpayer_regime_fkey_idx ON expenses (user_id, taxpayer_id, regime_code);

CREATE TRIGGER expenses_set_updated_at
  BEFORE UPDATE ON expenses
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE expenses ENABLE ROW LEVEL SECURITY;
ALTER TABLE expenses FORCE ROW LEVEL SECURITY;
-- Intent: a user lists, reads, records, edits, soft-deletes and restores only their own
-- expenses; a write can never hand a row to another user (WITH CHECK). No DELETE policy
-- or grant: DELETE on the API is a soft delete (UPDATE deleted_at); rows go away only with
-- their taxpayer or user (cascade) or through the purge job (M9-08).
CREATE POLICY expenses_select_own ON expenses FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY expenses_insert_own ON expenses FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY expenses_update_own ON expenses FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- INSERT names the id (idempotency) and never a deleted row; UPDATE covers the edit
-- (regime, category, amounts, payment, CFDI, override, withholdings, date), the soft delete
-- and the restore. Never id, user_id, taxpayer_id or the audit columns: an expense never
-- moves to another user or taxpayer. cc_webhook and cc_worker get nothing.
GRANT SELECT ON expenses TO cc_app;
GRANT INSERT (id, user_id, taxpayer_id, regime_code, category_code, description, amount_cents, iva_rate,
              iva_cents, payment_method, has_cfdi, cfdi_uuid, deductible, deductible_overridden,
              isr_withheld_cents, iva_withheld_cents, paid_on)
  ON expenses TO cc_app;
GRANT UPDATE (regime_code, category_code, description, amount_cents, iva_rate, iva_cents, payment_method,
              has_cfdi, cfdi_uuid, deductible, deductible_overridden, isr_withheld_cents, iva_withheld_cents,
              paid_on, deleted_at)
  ON expenses TO cc_app;

-- +goose Down
DROP TABLE expenses;
