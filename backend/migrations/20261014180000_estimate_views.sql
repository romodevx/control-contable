-- Estimate views (M8-02): the last time each user looked at one month's estimate of one
-- taxpayer, and what they saw, so the history can say "changed since you viewed it".
-- Sources: docs/project-specs/MODULE-history.md §2.3 (DECIDED), §2.4, §2.6 (read path)
-- and §3; docs/architecture/data-model.md (estimate_views).
--
-- One row per (user, taxpayer, month):
--   * period is the first day of the month, in [2022-01, 2100-12] (the catalog floor and
--     FiscalYear.maximum). "Not before 2026" (parameters_unavailable), "not in the future"
--     and "not the current month" are clock or parameter rules of the use case (§2.1);
--   * records_fingerprint / figures_fingerprint are lower-case SHA-256 hex digests as the
--     user saw them (ADR-0002 §5 and §2.4); only figures_fingerprint drives the marker, the
--     rest explain its reason. char(64) refuses a longer value (22001) and the pattern a
--     shorter one (a short char value is space-padded, and the pad is not hex);
--   * engine_version and parameter_set are opaque labels of at most 32 characters;
--   * viewed_at is the server clock of the view (the use case's Clock, not the device).
-- No tax amount is ever stored here (§2.3).
--
-- Ownership integrity, enforced by the database rather than trusted to the app:
--   * ENABLE + FORCE ROW LEVEL SECURITY and one policy per operation the app runs, keyed
--     on (SELECT app_user_id()); an unset context matches no row. Per user, not per
--     taxpayer: with shared access (G-3) each person keeps their own "seen" state;
--   * the composite FK (user_id, taxpayer_id) -> taxpayers (user_id, id) makes a view of
--     another user's taxpayer impossible even with forged ids (FK checks bypass RLS, so
--     the FK, not the policy, proves the pair belongs together);
--   * user_id, taxpayer_id and period are the key and immutable: cc_app has no UPDATE
--     grant on them. The write is one INSERT ... ON CONFLICT (user_id, taxpayer_id, period)
--     DO UPDATE (idempotent, the last view wins); it needs the SELECT policy to see the
--     conflicting row.
-- Deletion: account deletion (US-A4) cascades through the users FK (the users-FK guard
-- sees this table directly); taxpayer deletion (MODULE-taxpayers §2.7) cascades through
-- the composite FK. Both hold with module_history OFF (§3: rows are kept, never purged by
-- the flag). Cascades bypass RLS and need no DELETE grant, so cc_app has none.

-- +goose Up

CREATE TABLE estimate_views (
  user_id             uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  taxpayer_id         uuid        NOT NULL,
  period              date        NOT NULL,
  records_fingerprint char(64)    NOT NULL,
  figures_fingerprint char(64)    NOT NULL,
  engine_version      text        NOT NULL,
  parameter_set       text        NOT NULL,
  viewed_at           timestamptz NOT NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  -- Its btree leads with user_id: it serves the RLS predicate, the fiscal-year read
  -- (user + taxpayer + period range, §2.6 step 6: one range scan), the upsert's conflict
  -- target, the composite FK cascade (prefix user_id, taxpayer_id) and the users FK
  -- cascade (prefix user_id). No other index is needed.
  CONSTRAINT estimate_views_pkey PRIMARY KEY (user_id, taxpayer_id, period),
  -- The taxpayer is one of this user's; deleting it deletes its views.
  CONSTRAINT estimate_views_taxpayer_fkey FOREIGN KEY (user_id, taxpayer_id)
    REFERENCES taxpayers (user_id, id) ON DELETE CASCADE,
  CONSTRAINT estimate_views_period_first_of_month CHECK (extract(day FROM period) = 1),
  CONSTRAINT estimate_views_period_range CHECK (period BETWEEN DATE '2022-01-01' AND DATE '2100-12-01'),
  CONSTRAINT estimate_views_records_fingerprint_format CHECK (records_fingerprint ~ '^[0-9a-f]{64}$'),
  CONSTRAINT estimate_views_figures_fingerprint_format CHECK (figures_fingerprint ~ '^[0-9a-f]{64}$'),
  CONSTRAINT estimate_views_engine_version_length CHECK (char_length(engine_version) BETWEEN 1 AND 32),
  CONSTRAINT estimate_views_parameter_set_length CHECK (char_length(parameter_set) BETWEEN 1 AND 32)
);
COMMENT ON TABLE estimate_views IS
  'Last view of one month''s estimate per user + taxpayer (RLS on user_id). Fingerprints only, never amounts.';
COMMENT ON COLUMN estimate_views.period IS
  'First day of the viewed month.';
COMMENT ON COLUMN estimate_views.figures_fingerprint IS
  'SHA-256 hex of the figures the user saw (MODULE-history 2.4); the only input of changed_since_viewed.';
COMMENT ON COLUMN estimate_views.records_fingerprint IS
  'Input fingerprint as seen (ADR-0002 5); with engine_version and parameter_set it explains change_reason.';

CREATE TRIGGER estimate_views_set_updated_at
  BEFORE UPDATE ON estimate_views
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE estimate_views ENABLE ROW LEVEL SECURITY;
ALTER TABLE estimate_views FORCE ROW LEVEL SECURITY;
-- Intent: a user reads and marks only their own views; a write can never hand a row to
-- another user (WITH CHECK on INSERT and on the upsert's UPDATE). No DELETE policy: no
-- route deletes a view; rows leave with their taxpayer or account (cascade).
CREATE POLICY estimate_views_select_own ON estimate_views FOR SELECT TO cc_app
  USING (user_id = (SELECT app_user_id()));
CREATE POLICY estimate_views_insert_own ON estimate_views FOR INSERT TO cc_app
  WITH CHECK (user_id = (SELECT app_user_id()));
CREATE POLICY estimate_views_update_own ON estimate_views FOR UPDATE TO cc_app
  USING (user_id = (SELECT app_user_id()))
  WITH CHECK (user_id = (SELECT app_user_id()));
-- INSERT names the key and what was seen, never the audit columns. UPDATE (the upsert)
-- changes what was seen; the key never changes. cc_webhook and cc_worker get nothing.
GRANT SELECT ON estimate_views TO cc_app;
GRANT INSERT (user_id, taxpayer_id, period, records_fingerprint, figures_fingerprint, engine_version, parameter_set, viewed_at)
  ON estimate_views TO cc_app;
GRANT UPDATE (records_fingerprint, figures_fingerprint, engine_version, parameter_set, viewed_at)
  ON estimate_views TO cc_app;

-- +goose Down
DROP TABLE estimate_views;
