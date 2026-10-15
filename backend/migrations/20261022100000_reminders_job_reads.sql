-- Reminder job reads (M9-06): the amounts of a reminder (MODULE-reminders §2.5) are the tax
-- engine's monthly estimate, which the worker computes with the same Estimator as the API,
-- inside the user's own read-only transaction as cc_worker. Everything that estimate reads
-- is already readable by cc_worker under the user's RLS (taxpayers, taxpayer_regimes,
-- regimes, incomes, expenses: WH-08's backfill; holidays: M9-03) except the taxpayer's tax
-- inputs, which taxpayer_tax_inputs (M5-20) left for "M9 adds its own grant and policy".
--
-- Intent: cc_worker reads, inside one user's transaction, that user's tax inputs only, and
-- writes none of them. No definer, no cross-user reach.

-- +goose Up
CREATE POLICY taxpayer_tax_inputs_select_worker ON taxpayer_tax_inputs FOR SELECT TO cc_worker
  USING (user_id = (SELECT app_user_id()));
GRANT SELECT ON taxpayer_tax_inputs TO cc_worker;

-- +goose Down
REVOKE SELECT ON taxpayer_tax_inputs FROM cc_worker;
DROP POLICY taxpayer_tax_inputs_select_worker ON taxpayer_tax_inputs;
