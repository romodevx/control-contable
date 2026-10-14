-- Annual module inputs (ANN-03): the four annual-only input tables of
-- docs/project-specs/MODULE-annual.md §2.6 (fields), §5 (validation backstops), §6 (rows
-- kept with the flag OFF, cascade on deletion) and §9 (security): personal_deductions,
-- salary_statements, annual_filed_payments and annual_inputs. The estimate itself is
-- never stored (derived on every read, ADR-0002 §5).
--
-- Ownership integrity, enforced by the database rather than trusted to the app (§9):
--   * every table carries user_id with ENABLE + FORCE ROW LEVEL SECURITY and one policy
--     per operation the app runs, keyed on (SELECT app_user_id()) with WITH CHECK on every
--     write; an unset context matches no row;
--   * composite FKs bypass RLS, so they (not the policies) prove a row belongs to a
--     taxpayer of its own user, even with forged ids:
--       personal_deductions, salary_statements -> taxpayers (user_id, id, person_type)
--         with person_type pinned to 'fisica' (column default + CHECK, no grant): a row
--         for a persona moral is impossible (§5 person_type_not_allowed backstop);
--       annual_filed_payments -> taxpayer_regimes (user_id, taxpayer_id, regime_code): a
--         regime the taxpayer holds (whether it is held in that month is the use case's
--         regime_not_active rule);
--       annual_inputs -> taxpayers (user_id, id): any person type (the spec does not limit
--         it; its fields only matter to the PF section);
--   * user_id, taxpayer_id and the key columns are immutable: cc_app has no UPDATE grant on
--     them.
-- Deletion: account deletion (US-A4) cascades through each table's own users FK (the
-- users-FK guard sees them directly), taxpayer deletion through the composite FKs
-- (taxpayers -> [taxpayer_regimes ->] these tables). Both hold with module_annual OFF (§6).
-- Cascades bypass RLS and need no DELETE grant.
--
-- Shared field rules:
--   * money is bigint centavos in [0, NonNegativeCents.maximum = 99999999999999]
--     (docs/api/openapi.yaml); a personal deduction amount is >= 1 (§2.6 "> 0");
--   * fiscal_year in [2026, 2100]: 2026 is the first year with parameters (§5 fiscal_year
--     < 2026 = 422 parameters_unavailable), 2100 is FiscalYear.maximum. "Not in the
--     future" (out_of_range) is a clock rule of the use case;
--   * free-text labels (student_label, employer_label) are optional aliases, never required
--     PII (§9): trimmed, NFC (ADR-0004), bounded, never empty (empty is stored as NULL).
--     No employer RFC is stored (§2.6 "no employer RFC in v1", §9 data minimization).
-- Amounts, constancia figures and deduction details are never logged (§5); nothing here
-- writes them anywhere but these rows.

-- +goose Up

-- ---------------------------------------------------------------------------------------
-- personal_deductions: one row per deductible personal expense of a persona física
-- (AR §B.2, LISR 151 fr. I-VIII + decree DOF 26-12-2013 for tuition). Soft delete + restore
-- (Undo, rules/ui-forms-and-modals.md); rows leave by cascade or the purge job (M9-08).
--   * id: the client-generated idempotency key (uuid, any version); default uuidv7;
--   * fiscal_year: GENERATED from paid_on (§2.6 "date in the fiscal year"), so the year a
--     deduction counts in can never disagree with its date; edits of paid_on move it;
--   * payment_method cash is STORED: it is excluded from the four cash-restricted
--     categories by the calculator with a visible line (§2.7 step 1), not refused;
--   * beneficiary not self: the user attested the dependant's income < 1 UMA anual before
--     the use case accepts it (checkbox, not verified); nothing to store beyond the row;
--   * donee_type: set exactly for donation; student_label, tuition_level and
--     scholarship_cents: set exactly for tuition (the cap is per student, §2.7);
--   * cfdi_uuid: optional folio fiscal, format-validated only, upper case like expenses;
--     never sent to the SAT and not unique.
-- ---------------------------------------------------------------------------------------
CREATE TABLE personal_deductions (
  id                uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id           uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id       uuid        NOT NULL,
  person_type       text        NOT NULL DEFAULT 'fisica',
  fiscal_year       smallint    NOT NULL GENERATED ALWAYS AS (EXTRACT(YEAR FROM paid_on)::smallint) STORED,
  category          text        NOT NULL,
  amount_cents      bigint      NOT NULL,
  paid_on           date        NOT NULL,
  payment_method    text        NOT NULL,
  beneficiary       text        NOT NULL,
  donee_type        text,
  student_label     text,
  tuition_level     text,
  scholarship_cents bigint,
  cfdi_uuid         text,
  deleted_at        timestamptz,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  -- A persona física of this user (see the header).
  CONSTRAINT personal_deductions_taxpayer_fkey FOREIGN KEY (user_id, taxpayer_id, person_type)
    REFERENCES taxpayers (user_id, id, person_type) ON DELETE CASCADE,
  CONSTRAINT personal_deductions_person_type_fisica CHECK (person_type = 'fisica'),
  -- AR §B.2 fr. I-VIII + decree (tuition); openapi PersonalDeduction.category (ANN-02).
  CONSTRAINT personal_deductions_category_valid CHECK (category IN (
    'medical', 'medical_disability', 'funeral', 'donation', 'mortgage_real_interest',
    'retirement_contribution', 'health_insurance', 'school_transport', 'local_payroll_tax', 'tuition')),
  CONSTRAINT personal_deductions_amount_cents_range CHECK (amount_cents BETWEEN 1 AND 99999999999999),
  CONSTRAINT personal_deductions_paid_on_range CHECK (paid_on BETWEEN DATE '2026-01-01' AND DATE '2100-12-31'),
  CONSTRAINT personal_deductions_payment_method_valid CHECK (payment_method IN ('electronic', 'cash')),
  CONSTRAINT personal_deductions_beneficiary_valid CHECK (
    beneficiary IN ('self', 'spouse', 'ascendant', 'descendant')),
  -- donee_type exactly for donations. Built from IS [NOT] NULL, so never NULL itself.
  CONSTRAINT personal_deductions_donee_type_iff_donation CHECK (
    (category = 'donation') = (donee_type IS NOT NULL)),
  CONSTRAINT personal_deductions_donee_type_valid CHECK (donee_type IN ('authorized', 'government')),
  -- The tuition fields exactly for tuition, all three together.
  CONSTRAINT personal_deductions_tuition_fields_iff_tuition CHECK (
    CASE WHEN category = 'tuition'
      THEN student_label IS NOT NULL AND tuition_level IS NOT NULL AND scholarship_cents IS NOT NULL
      ELSE student_label IS NULL AND tuition_level IS NULL AND scholarship_cents IS NULL
    END),
  CONSTRAINT personal_deductions_tuition_level_valid CHECK (tuition_level IN (
    'preescolar', 'primaria', 'secundaria', 'profesional_tecnico', 'bachillerato')),
  CONSTRAINT personal_deductions_scholarship_cents_range CHECK (scholarship_cents BETWEEN 0 AND 99999999999999),
  -- Up to 40 characters, trimmed, NFC; empty is stored as NULL (tuition then fails above).
  CONSTRAINT personal_deductions_student_label_valid CHECK (
    student_label <> '' AND student_label = btrim(student_label)
    AND char_length(student_label) <= 40 AND student_label IS NFC NORMALIZED),
  -- 8-4-4-4-12 hex, upper case.
  CONSTRAINT personal_deductions_cfdi_uuid_format CHECK (
    cfdi_uuid ~ '^[0-9A-F]{8}-([0-9A-F]{4}-){3}[0-9A-F]{12}$')
);
COMMENT ON TABLE personal_deductions IS
  'Personal deductions of a persona física per fiscal year (RLS on user_id). Soft delete via deleted_at.';
COMMENT ON COLUMN personal_deductions.fiscal_year IS
  'Generated: the year of paid_on. Never written by the app.';
COMMENT ON COLUMN personal_deductions.person_type IS
  'Always fisica; part of the composite FK that refuses a persona moral. Never written by the app.';
COMMENT ON COLUMN personal_deductions.payment_method IS
  'electronic | cash. cash is stored and excluded by the calculator for medical, medical_disability, school_transport and tuition.';

-- Indexes. Every one leads with user_id: the RLS predicate is an index condition.
--   * live by year: the annual estimate (all live rows of a taxpayer and year), the list
--     (fiscal_year, optional category filter, grouped by category, newest first). Partial:
--     soft-deleted rows leave every read at once;
--   * the composite FK: cascades from taxpayers and users must reach soft-deleted rows too,
--     which the partial index skips;
--   * deleted_at: the purge job's range scan (M9-08), only over deleted rows.
CREATE INDEX personal_deductions_live_year_idx
  ON personal_deductions (user_id, taxpayer_id, fiscal_year, category, paid_on, id)
  WHERE deleted_at IS NULL;
CREATE INDEX personal_deductions_taxpayer_fkey_idx ON personal_deductions (user_id, taxpayer_id);
CREATE INDEX personal_deductions_deleted_at_idx ON personal_deductions (deleted_at)
  WHERE deleted_at IS NOT NULL;

CREATE TRIGGER personal_deductions_set_updated_at
  BEFORE UPDATE ON personal_deductions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE personal_deductions ENABLE ROW LEVEL SECURITY;
ALTER TABLE personal_deductions FORCE ROW LEVEL SECURITY;
-- Intent: a user lists, records, edits, soft-deletes and restores only the deductions of
-- their own taxpayers; a write can never hand a row to another user (WITH CHECK). No
-- DELETE policy or grant: DELETE on the API is a soft delete (UPDATE deleted_at).
CREATE POLICY personal_deductions_select_own ON personal_deductions FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY personal_deductions_insert_own ON personal_deductions FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY personal_deductions_update_own ON personal_deductions FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- INSERT names the id (idempotency) and never a deleted row; UPDATE covers the edit, the
-- soft delete and the restore. Never id, user_id, taxpayer_id, person_type, the generated
-- fiscal_year or the audit columns. cc_webhook and cc_worker get nothing.
GRANT SELECT ON personal_deductions TO cc_app;
GRANT INSERT (id, user_id, taxpayer_id, category, amount_cents, paid_on, payment_method, beneficiary, donee_type,
              student_label, tuition_level, scholarship_cents, cfdi_uuid)
  ON personal_deductions TO cc_app;
GRANT UPDATE (category, amount_cents, paid_on, payment_method, beneficiary, donee_type, student_label,
              tuition_level, scholarship_cents, cfdi_uuid, deleted_at)
  ON personal_deductions TO cc_app;

-- ---------------------------------------------------------------------------------------
-- salary_statements: one row per employer constancia (regime 605 income) of a persona
-- física, used only by the annual estimate (owner decision 3). Allowed for any PF, 605
-- registered or not (OQ-A2). Figures are the constancia's totals; the app never
-- recomputes art. 93 exemptions or the employment subsidy. Hard delete (no restore in §1).
-- ---------------------------------------------------------------------------------------
CREATE TABLE salary_statements (
  id                    uuid        PRIMARY KEY DEFAULT uuidv7(),
  user_id               uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id           uuid        NOT NULL,
  person_type           text        NOT NULL DEFAULT 'fisica',
  fiscal_year           smallint    NOT NULL,
  employer_label        text,
  accumulable_cents     bigint      NOT NULL,
  exempt_cents          bigint      NOT NULL,
  isr_withheld_cents    bigint      NOT NULL,
  simultaneous          boolean     NOT NULL DEFAULT false,
  left_before_year_end  boolean     NOT NULL DEFAULT false,
  payer_not_withholding boolean     NOT NULL DEFAULT false,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),
  -- A persona física of this user (see the header).
  CONSTRAINT salary_statements_taxpayer_fkey FOREIGN KEY (user_id, taxpayer_id, person_type)
    REFERENCES taxpayers (user_id, id, person_type) ON DELETE CASCADE,
  CONSTRAINT salary_statements_person_type_fisica CHECK (person_type = 'fisica'),
  CONSTRAINT salary_statements_fiscal_year_range CHECK (fiscal_year BETWEEN 2026 AND 2100),
  CONSTRAINT salary_statements_accumulable_cents_range CHECK (accumulable_cents BETWEEN 0 AND 99999999999999),
  CONSTRAINT salary_statements_exempt_cents_range CHECK (exempt_cents BETWEEN 0 AND 99999999999999),
  CONSTRAINT salary_statements_isr_withheld_cents_range CHECK (isr_withheld_cents BETWEEN 0 AND 99999999999999),
  -- Up to 60 characters, trimmed, NFC; empty is stored as NULL. An alias, never an RFC.
  CONSTRAINT salary_statements_employer_label_valid CHECK (
    employer_label <> '' AND employer_label = btrim(employer_label)
    AND char_length(employer_label) <= 60 AND employer_label IS NFC NORMALIZED)
);
COMMENT ON TABLE salary_statements IS
  'Salary constancias (605) of a persona física per fiscal year (RLS on user_id). No employer RFC stored.';
COMMENT ON COLUMN salary_statements.simultaneous IS
  'Worked for two or more employers at the same time (LISR 98 fr. III, obligation test).';

-- The list and the annual estimate read one taxpayer and year; the btree leads with
-- user_id (RLS) and its (user_id, taxpayer_id) prefix serves both cascades.
CREATE INDEX salary_statements_year_idx ON salary_statements (user_id, taxpayer_id, fiscal_year, id);

CREATE TRIGGER salary_statements_set_updated_at
  BEFORE UPDATE ON salary_statements
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE salary_statements ENABLE ROW LEVEL SECURITY;
ALTER TABLE salary_statements FORCE ROW LEVEL SECURITY;
-- Intent: a user lists, records, edits and deletes only the constancias of their own
-- taxpayers; a write can never hand a row to another user (WITH CHECK).
CREATE POLICY salary_statements_select_own ON salary_statements FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY salary_statements_insert_own ON salary_statements FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY salary_statements_update_own ON salary_statements FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY salary_statements_delete_own ON salary_statements FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
-- The fiscal year is fixed at creation (the list is per year). cc_webhook and cc_worker
-- get nothing.
GRANT SELECT, DELETE ON salary_statements TO cc_app;
GRANT INSERT (id, user_id, taxpayer_id, fiscal_year, employer_label, accumulable_cents, exempt_cents,
              isr_withheld_cents, simultaneous, left_before_year_end, payer_not_withholding)
  ON salary_statements TO cc_app;
GRANT UPDATE (employer_label, accumulable_cents, exempt_cents, isr_withheld_cents, simultaneous,
              left_before_year_end, payer_not_withholding)
  ON salary_statements TO cc_app;

-- ---------------------------------------------------------------------------------------
-- annual_filed_payments: the provisional ISR payment the user actually filed, per regime
-- and month (§2.5 source 1, "filed"). An absent month is not confirmed and falls back to
-- the M5 tax input, then to the app's estimate. PUT replaces the year's set for one
-- regime: DELETE then INSERT (or INSERT ... ON CONFLICT on the key) in one transaction.
-- Two concurrent PUTs of the same regime and year can race to a 23505; the use case
-- serializes them by locking the taxpayer_regimes row first (SELECT ... FOR NO KEY
-- UPDATE, as PUT tax inputs does).
-- ---------------------------------------------------------------------------------------
CREATE TABLE annual_filed_payments (
  user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id  uuid        NOT NULL,
  regime_code  text        NOT NULL,
  fiscal_year  smallint    NOT NULL,
  month        smallint    NOT NULL,
  amount_cents bigint      NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  -- The key (§2.6). Its btree leads with user_id: it serves the RLS predicate, the read of
  -- one regime and year, the upsert, and the prefix (user_id, taxpayer_id, regime_code) of
  -- both cascades. No other index is needed.
  CONSTRAINT annual_filed_payments_pkey PRIMARY KEY (user_id, taxpayer_id, regime_code, fiscal_year, month),
  -- A regime this user's taxpayer holds; deleting the taxpayer (and its regimes) deletes it.
  CONSTRAINT annual_filed_payments_taxpayer_regime_fkey FOREIGN KEY (user_id, taxpayer_id, regime_code)
    REFERENCES taxpayer_regimes (user_id, taxpayer_id, regime_code) ON DELETE CASCADE,
  CONSTRAINT annual_filed_payments_fiscal_year_range CHECK (fiscal_year BETWEEN 2026 AND 2100),
  CONSTRAINT annual_filed_payments_month_range CHECK (month BETWEEN 1 AND 12),
  CONSTRAINT annual_filed_payments_amount_cents_range CHECK (amount_cents BETWEEN 0 AND 99999999999999)
);
COMMENT ON TABLE annual_filed_payments IS
  'Provisional ISR payments actually filed, per taxpayer + regime + fiscal year + month (RLS on user_id).';

CREATE TRIGGER annual_filed_payments_set_updated_at
  BEFORE UPDATE ON annual_filed_payments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE annual_filed_payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE annual_filed_payments FORCE ROW LEVEL SECURITY;
-- Intent: a user reads, confirms, changes and clears only the filed payments of their own
-- taxpayers; a write can never hand a row to another user (WITH CHECK). DELETE is the
-- replace of PUT.
CREATE POLICY annual_filed_payments_select_own ON annual_filed_payments FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY annual_filed_payments_insert_own ON annual_filed_payments FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY annual_filed_payments_update_own ON annual_filed_payments FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY annual_filed_payments_delete_own ON annual_filed_payments FOR DELETE TO cc_app
  USING (user_id = (SELECT app_user_id()));
-- The key columns never change; only the amount does (upsert). cc_webhook and cc_worker
-- get nothing.
GRANT SELECT, DELETE ON annual_filed_payments TO cc_app;
GRANT INSERT (user_id, taxpayer_id, regime_code, fiscal_year, month, amount_cents)
  ON annual_filed_payments TO cc_app;
GRANT UPDATE (amount_cents) ON annual_filed_payments TO cc_app;

-- ---------------------------------------------------------------------------------------
-- annual_inputs: one row per taxpayer and fiscal year of the per-year facts no record
-- carries (§2.6). No row = the defaults below.
--   * prior_year_accumulable_income_cents: the previous year's acumulables before personal
--     deductions, for the 7 % donation cap. NULL = unknown: donations are then excluded
--     and the section is incomplete (OQ-A5); 0 is a real value;
--   * other_exempt_income_cents: for the 15 % global cap's "total income";
--   * told_employer_will_file: an obligation trigger of §2.4.
-- PUT is an upsert (INSERT ... ON CONFLICT on the key); no DELETE.
-- ---------------------------------------------------------------------------------------
CREATE TABLE annual_inputs (
  user_id                             uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id                         uuid        NOT NULL,
  fiscal_year                         smallint    NOT NULL,
  prior_year_accumulable_income_cents bigint,
  other_exempt_income_cents           bigint      NOT NULL DEFAULT 0,
  told_employer_will_file             boolean     NOT NULL DEFAULT false,
  created_at                          timestamptz NOT NULL DEFAULT now(),
  updated_at                          timestamptz NOT NULL DEFAULT now(),
  -- The key. Its btree leads with user_id: RLS, the read of one year, the upsert and the
  -- prefix (user_id, taxpayer_id) of both cascades. No other index is needed.
  CONSTRAINT annual_inputs_pkey PRIMARY KEY (user_id, taxpayer_id, fiscal_year),
  CONSTRAINT annual_inputs_taxpayer_fkey FOREIGN KEY (user_id, taxpayer_id)
    REFERENCES taxpayers (user_id, id) ON DELETE CASCADE,
  CONSTRAINT annual_inputs_fiscal_year_range CHECK (fiscal_year BETWEEN 2026 AND 2100),
  CONSTRAINT annual_inputs_prior_year_income_range CHECK (
    prior_year_accumulable_income_cents BETWEEN 0 AND 99999999999999),
  CONSTRAINT annual_inputs_other_exempt_income_range CHECK (other_exempt_income_cents BETWEEN 0 AND 99999999999999)
);
COMMENT ON TABLE annual_inputs IS
  'Per-fiscal-year annual facts of a taxpayer (RLS on user_id): prior-year income, other exempt income, told employer.';
COMMENT ON COLUMN annual_inputs.prior_year_accumulable_income_cents IS
  'Prior-year acumulables before personal deductions (7 % donation cap). NULL = unknown (OQ-A5), not zero.';

CREATE TRIGGER annual_inputs_set_updated_at
  BEFORE UPDATE ON annual_inputs
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE annual_inputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE annual_inputs FORCE ROW LEVEL SECURITY;
-- Intent: a user reads and saves only the annual facts of their own taxpayers; a write
-- can never hand a row to another user (WITH CHECK). No DELETE: PUT overwrites the row,
-- and rows leave with their taxpayer or user (cascade).
CREATE POLICY annual_inputs_select_own ON annual_inputs FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY annual_inputs_insert_own ON annual_inputs FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY annual_inputs_update_own ON annual_inputs FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- cc_webhook and cc_worker get nothing.
GRANT SELECT ON annual_inputs TO cc_app;
GRANT INSERT (user_id, taxpayer_id, fiscal_year, prior_year_accumulable_income_cents, other_exempt_income_cents,
              told_employer_will_file)
  ON annual_inputs TO cc_app;
GRANT UPDATE (prior_year_accumulable_income_cents, other_exempt_income_cents, told_employer_will_file)
  ON annual_inputs TO cc_app;

-- +goose Down
DROP TABLE annual_inputs;
DROP TABLE annual_filed_payments;
DROP TABLE salary_statements;
DROP TABLE personal_deductions;
