-- Holidays (M9-03): the global non-business days that move the day-17 due date, seeded
-- for 2026 and 2027, and the years the seed covers.
-- Sources: docs/project-specs/MODULE-reminders.md §2.1 and §2.6, research
-- docs/research/tax-regimes-mx-2026.md §A.4.1 (CFF art. 12 read 2026-10-15) and Part D
-- sources 4 (CFF) and 12 (RMF rule 2.1.6).
--
-- The rule itself (day 17 of the following month, moved past Saturdays, Sundays and these
-- days) lives in the Go fiscal kernel (M9-04), not here: the database only holds the data.
--   * holiday_years lists the years whose days are seeded. A year that is missing here is
--     "unseeded": the calendar answers 422 calendar_unavailable and the job falls back to
--     weekends only and alerts (§2.1). Ops seed the next year each December, after the RMF.
--   * holidays.year is derived from the day (never written) and references holiday_years,
--     so a day can only be added for a year that is declared, and a year cannot be removed
--     while it still has days.
--   * moves_due_date is stated on every row (no default). The CFF 12 ¶1 days move it; Holy
--     Thursday and Friday (RMF 2.1.6) are stored with true per the spec ⚠️: whether they
--     move a payment deadline was not confirmed, and no 2026-2027 day 17 falls on them.
--   * Not stored: SAT vacation periods (CFF 12 ¶2 makes them business days for payments),
--     the RFC-digit grace days and the Friday extension (OQ-R1).
--
-- Not user data: no user_id, the same rows for everyone. RLS is ENABLED (not FORCED, like
-- the other global catalogs) with a SELECT-only policy for the two roles that compute a due
-- date: cc_app (estimate, calendar) and cc_worker (reminder job). Nobody writes at run
-- time; a new year is a migration.

-- +goose Up

CREATE TABLE holiday_years (
  year        smallint    PRIMARY KEY,
  seeded_on   date        NOT NULL,
  source_note text        NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT holiday_years_year_range CHECK (year BETWEEN 2022 AND 2100),
  CONSTRAINT holiday_years_source_note_valid CHECK (
    source_note <> '' AND source_note = btrim(source_note) AND char_length(source_note) <= 500)
);
COMMENT ON TABLE holiday_years IS
  'Years whose non-business days are seeded in holidays (M9-03). Outside them the due-date rule is weekends-only.';

CREATE TABLE holidays (
  day            date        PRIMARY KEY,
  year           smallint    GENERATED ALWAYS AS (extract(year FROM day)::smallint) STORED NOT NULL
                             REFERENCES holiday_years (year),
  name           text        NOT NULL,
  basis          text        NOT NULL,
  moves_due_date boolean     NOT NULL,
  source_url     text        NOT NULL,
  verified_on    date        NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT holidays_name_valid CHECK (
    name <> '' AND name = btrim(name) AND char_length(name) <= 120 AND name IS NFC NORMALIZED),
  -- cff_12: CFF art. 12 ¶1. rmf_2_1_6: RMF rule 2.1.6 (days SAT declares non-business).
  CONSTRAINT holidays_basis_valid CHECK (basis IN ('cff_12', 'rmf_2_1_6')),
  CONSTRAINT holidays_source_url_https CHECK (source_url ~ '^https://[^[:space:]]+$' AND char_length(source_url) <= 500)
);
COMMENT ON TABLE holidays IS
  'Non-business days that move the day-17 due date (CFF art. 12; RMF 2.1.6 ⚠️). Global catalog, read-only at run time.';
COMMENT ON COLUMN holidays.moves_due_date IS
  'true: a due date on this day moves to the next business day. Stated per row, never defaulted.';
COMMENT ON COLUMN holidays.year IS 'Derived from day; references holiday_years (only seeded years hold days).';

CREATE TRIGGER holiday_years_set_updated_at
  BEFORE UPDATE ON holiday_years
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER holidays_set_updated_at
  BEFORE UPDATE ON holidays
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Seed: research §A.4.1, read 2026-10-15. The 2027 Holy Thursday/Friday follow Easter
-- (28 March 2027); RMF 2027 is not yet published, so their source is the 2026 rule.
INSERT INTO holiday_years (year, seeded_on, source_note) VALUES
  (2026, DATE '2026-10-15', 'CFF art. 12 ¶1 (DOF 09-04-2026) + RMF 2026 rule 2.1.6 Holy Thursday/Friday ⚠️; research tax-regimes-mx-2026.md §A.4.1'),
  (2027, DATE '2026-10-15', 'CFF art. 12 ¶1 (DOF 09-04-2026) + Holy Thursday/Friday projected from Easter ⚠️ (RMF 2027 not yet published); research §A.4.1');

INSERT INTO holidays (day, name, basis, moves_due_date, source_url, verified_on) VALUES
  (DATE '2026-01-01', 'Año Nuevo', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-02-02', 'Día de la Constitución (primer lunes de febrero)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-03-16', 'Natalicio de Benito Juárez (tercer lunes de marzo)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-04-02', 'Jueves Santo', 'rmf_2_1_6', true, 'https://siemprealdia.co/mexico/fiscal/dias-inhabiles-del-sat/', DATE '2026-10-15'),
  (DATE '2026-04-03', 'Viernes Santo', 'rmf_2_1_6', true, 'https://siemprealdia.co/mexico/fiscal/dias-inhabiles-del-sat/', DATE '2026-10-15'),
  (DATE '2026-05-01', 'Día del Trabajo', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-05-05', 'Batalla de Puebla', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-09-16', 'Día de la Independencia', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-11-16', 'Revolución Mexicana (tercer lunes de noviembre)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2026-12-25', 'Navidad', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-01-01', 'Año Nuevo', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-02-01', 'Día de la Constitución (primer lunes de febrero)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-03-15', 'Natalicio de Benito Juárez (tercer lunes de marzo)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-03-25', 'Jueves Santo (proyectado; RMF 2027 no publicada)', 'rmf_2_1_6', true, 'https://siemprealdia.co/mexico/fiscal/dias-inhabiles-del-sat/', DATE '2026-10-15'),
  (DATE '2027-03-26', 'Viernes Santo (proyectado; RMF 2027 no publicada)', 'rmf_2_1_6', true, 'https://siemprealdia.co/mexico/fiscal/dias-inhabiles-del-sat/', DATE '2026-10-15'),
  (DATE '2027-05-01', 'Día del Trabajo', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-05-05', 'Batalla de Puebla', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-09-16', 'Día de la Independencia', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-11-15', 'Revolución Mexicana (tercer lunes de noviembre)', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15'),
  (DATE '2027-12-25', 'Navidad', 'cff_12', true, 'https://www.diputados.gob.mx/LeyesBiblio/pdf/CFF.pdf', DATE '2026-10-15');

ALTER TABLE holiday_years ENABLE ROW LEVEL SECURITY;
ALTER TABLE holidays ENABLE ROW LEVEL SECURITY;
-- Intent: the API (estimate, tax calendar) and the worker (reminder job) read every row;
-- no runtime role writes. cc_webhook gets nothing.
CREATE POLICY holiday_years_read ON holiday_years FOR SELECT TO cc_app, cc_worker USING (true);
CREATE POLICY holidays_read ON holidays FOR SELECT TO cc_app, cc_worker USING (true);
GRANT SELECT ON holiday_years, holidays TO cc_app, cc_worker;

-- +goose Down

DROP TABLE holidays;
DROP TABLE holiday_years;
