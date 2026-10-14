# Runbook: Dashboard (`module_dashboard`)

Operations for module M6 ([dashboard](../modules/dashboard.md); spec
[MODULE-dashboard](../project-specs/MODULE-dashboard.md)). Every command in a `sh` block was run
on an isolated local stack on 2026-10-14. The ground rules of the
[incomes and expenses runbook](incomes-expenses.md) apply unchanged: **amounts never leave the
database** (never ask a user for one, never select one, never paste one into a ticket), the API
never logs them, and per-user SQL runs as `cc_app` under RLS with `app.current_user_id` set.

The dashboard owns no table and no tax math. Every figure on Home is either a sum of the user's
own records or the M5 estimate, so "Home shows wrong numbers" is almost always one of: the
records are not where the user thinks (date, taxpayer, deleted), Home measures something other
than what the user expects, or a module is OFF.

## 1. Investigate "Home shows wrong numbers", without exposing amounts

You need the user id (from the account, as in the
[auth runbook §4](auth.md#4-investigate-a-user-cant-log-in-without-pii)), which section looks
wrong (banner, hero, month card, a month of the chart), the month, and if possible the
`request_id` of the failed call. Ask for **which** figure, never **what** figure.

1. **Is it a rule working as designed?** Check the user's expectation against what Home measures:

   | The user says | What Home does | Next step |
   |---|---|---|
   | "My income this month is missing an invoice" | "Ingresos cobrados" counts incomes by **`collected_on`**, before IVA. An income invoiced but not collected ("Por cobrar") is not on Home | Step 4: a row in the `income` list with an empty `home_month` |
   | "The figure does not include IVA" | Income and expenses are **before IVA**; the hero's IVA is separate | Not a bug |
   | "Expenses are higher than what I can deduct" | "Gastos pagados" counts every live expense by `paid_on`, deductible or not | Not a bug; deductibility is in the estimate's breakdown |
   | "A month I never used has a tax figure" | ISR of 612, 626 PM and 601 is **year to date**: earlier months carry into later ones | Not a bug |
   | "It says Incompleto" / the total looks too high | A tax input the period needs is missing (often the prior provisional payments) and is computed as **zero** | Step 4 lists which inputs exist; the user completes them from the estimate screen |
   | "Home and the estimate screen differ" | They are the same `TaxpayerEstimate` to the centavo (contract-tested). The estimate screen may be on another **period**, or Home on another **taxpayer** | Ask which taxpayer and month each screen showed |
   | "Saldo a favor shows as zero" | A favor balance never makes a bar; the table says "IVA: saldo a favor" without the amount (QA D-3) | Known limit |
   | "The month just changed and Home still shows last month" | The current month is America/Mexico_City's, from the server clock | Compare the time in Mexico City, not UTC |
   | "I deleted a record and the trial counter did not go up" | Trial counters never rise after a delete (PRD §7), across all the user's taxpayers | Not a bug |
   | "The income (or expense) figure disappeared" | That module's flag is OFF: the server returns `null` and the app hides the figure | Step 2 |
   | "There is no estimate on Home" | `module_tax_engine` is OFF (Home shows income/expense only), or no regime is held this month (`NoRegimePrompt`) | Steps 2 and 4 |
   | "Home showed old figures, then an error" | A failed refresh shows the error state, never stale figures (since PR #116) | An older app build; ask for the version |

2. **Which modules are ON?** Home's sections follow the flags, not a client setting:

   ```sh
   curl -s http://localhost:3000/v1/flags   # {"enabled":["module_auth","module_dashboard","module_expenses","module_incomes","module_tax_engine","module_taxpayers"]}
   make -C backend flags                    # LOCAL; "hidden" + REQUIRES module_taxpayers = the dashboard 404s
   ```

3. **What did the API answer?** The dashboard logs nothing on success or on a `4xx`. A `500` is
   logged with its `request_id` and an error chain of periods, counts and sentinels, never an
   amount. `404` is a foreign or unknown taxpayer, or the flag (or `module_taxpayers`) OFF; `422`
   is a `period` before 2026-01; `429` is the `dashboard_reads` budget (600/h). The full table is
   in the [module page](../modules/dashboard.md#errors). Locally, with the API output in `api.log`:

   ```sh
   grep '^{' api.log | jq -c 'select(.request_id == "<request-id>") | {time, level, msg, code, error}'
   ```

4. **Where the records are, by month (needs the user id).** The queries read no amount, no
   description and no folio. They count live and deleted records per taxpayer and per **Home
   month** (the date Home uses), and list which tax inputs exist, not their values:

   ```sh
   make -C backend db-psql-app PSQL_ARGS="-v user_id=<user-uuid>" <<'SQL'
   BEGIN;
   SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
   SELECT id, archived_at IS NOT NULL AS archived FROM taxpayers ORDER BY created_at;
   SELECT taxpayer_id, regime_code, start_period, coalesce(end_period, 'open') AS end_period
   FROM taxpayer_regimes ORDER BY taxpayer_id, regime_code;
   SELECT 'income' AS kind, taxpayer_id, to_char(collected_on, 'YYYY-MM') AS home_month,
          count(*) FILTER (WHERE deleted_at IS NULL) AS live, count(*) FILTER (WHERE deleted_at IS NOT NULL) AS deleted
   FROM incomes GROUP BY 1, 2, 3
   UNION ALL
   SELECT 'expense', taxpayer_id, to_char(paid_on, 'YYYY-MM'),
          count(*) FILTER (WHERE deleted_at IS NULL), count(*) FILTER (WHERE deleted_at IS NOT NULL)
   FROM expenses GROUP BY 1, 2, 3
   ORDER BY 2, 3 NULLS FIRST, 1;
   SELECT taxpayer_id, regime_code, fiscal_year, input_key, array_agg(month ORDER BY month) AS months
   FROM taxpayer_tax_inputs GROUP BY 1, 2, 3, 4 ORDER BY 1, 2, 3, 4;
   SELECT kind, source, starts_at, ends_at FROM entitlements ORDER BY starts_at;
   SELECT incomes_created, expenses_created FROM user_usage;
   COMMIT;
   SQL
   ```

   Expected output (the [module walkthrough](../modules/dashboard.md#try-it-locally) user):
   `context_set = t`; one taxpayer, not archived; `612 | 2026-01 | open`; `income 2026-01 1 0`,
   `expense 2026-03 1 0`, `income 2026-03 1 0`; no tax inputs; one `trial` entitlement; usage
   `2 | 1`.

   - **A record is in another month** than the user thinks: Home groups by `collected_on` /
     `paid_on`, while the incomes list groups an accrual income by `invoiced_on`.
   - **An `income` row with an empty `home_month`**: invoiced, not collected; Home leaves it out
     until the user enters *Fecha de cobro*.
   - **Records under another `taxpayer_id`**: the user is looking at the other taxpayer's Home.
     There is no combined view (PRD Q14).
   - **Only `deleted` counts in a month**: the records were deleted; Home and the estimate count
     live records only. Undo restores them.
   - **No regime covers the month** (`start_period` after it, or `end_period` before it): no
     estimate that month (`status: none`).
   - **`prior_provisional_payments` without the months before the period**: the estimate is
     `incomplete` and counts them as zero, which raises the ISR.
   - **No rows at all**: the user id is wrong, or nothing was recorded (`is_empty`, the empty
     state).

If the records are where the user expects and the figure still looks wrong, the question is the
estimate's, not Home's: continue with the [tax engine](../modules/tax-engine.md) (its breakdown
shows every line and its legal basis). Operators do not read or correct amounts.

## 2. Enable `module_dashboard` in a shared environment

Only when every precondition of the
[security gate §5](../reviews/M6-08-security-gate.md#5-preconditions-to-enable-module_dashboard-in-staging-or-production)
holds, recorded in the deploy PR. The ones that are still open on 2026-10-14:

- **P-5**: the owner's manual TalkBack (real Android device) and VoiceOver (iOS) pass
  ([M6-07 FU6](../reviews/M6-07-tech-lead-gate.md#follow-ups) has the checklist).
- **P-8**: dashboard p95 < 1 s on staging, cold start included (REL-06 k6).
- **P-2**: with `module_tax_engine` ON in that environment, its own preconditions, the CPA
  sign-off (M5-33) first: the hero and the tax view are the engine's figures.

`module_taxpayers` must be ON for the flag to take effect. Shared environments flip flags through
their own owner connection (`cmd/flags` with `DATABASE_URL_OWNER` from the secret manager),
never through the local make target.
