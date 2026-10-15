# Runbook: Incomes & Expenses (`module_incomes`, `module_expenses`)

Operations for modules M3 and M4 ([incomes](../modules/incomes.md),
[expenses](../modules/expenses.md); specs [MODULE-incomes](../project-specs/MODULE-incomes.md),
[MODULE-expenses](../project-specs/MODULE-expenses.md)). Every command in a `sh` block was run
on an isolated local stack on 2026-10-14. Procedures for shared environments (staging,
production) say so explicitly. Read the whole procedure before you start one.

Ground rules:

- **Amounts, descriptions and CFDI folios never leave the database.** They are financial PII:
  a description may name the user's client. Never ask a user for an amount or a folio, never
  paste one into a ticket, chat or shell history, and never select those columns. Work with
  the user id, the taxpayer id, the record id (all UUIDs), dates, enums and the `request_id`
  of a response. Nothing in this runbook needs a value.
- **The API never logs them either.** Audit events carry ids only, error bodies carry codes and
  fixed messages, and no log level (`LOG_LEVEL=debug` included) prints a record field
  (verified by the [security gate](../reviews/M3-M4-security-gate.md) log scan). Do not add a
  log line, a trace attribute or a debug print that does; that is a security defect.
- **RLS applies to operators too.** `incomes`, `expenses` and `user_usage` are
  `FORCE ROW LEVEL SECURITY` with policies only for `cc_app`, and no role has BYPASSRLS.
  Per-user SQL runs as `cc_app` inside a transaction that sets `app.current_user_id`, exactly as
  the API does. A wrong user id returns no rows, it does not fail.
- Locally, `make -C backend db-psql-app` runs SQL from stdin as `cc_app`, and only against the
  local stack. Shared environments run the same SQL as `cc_app` through their own connection
  (Cloud SQL Auth Proxy, credentials from Secret Manager).
- `expense_categories`, `category_regime_rules` and `regimes` are global catalogs: `cc_app`
  reads them without a user context, and only migrations write them.

## 1. Investigate a user's record issue, without exposing amounts

You need the user id (from the account, as in the
[auth runbook §4](auth.md#4-investigate-a-user-cant-log-in-without-pii)), the taxpayer id if the
user has several, and, if possible, the `request_id` of the failed call.

1. **What did the API answer?** Most record problems are a rule working as designed.

   | API answer | Meaning | Next step |
   |---|---|---|
   | `400`, `out_of_range` on a date | The date is in the future or **before 2026-01-01** (no tax parameters earlier) | The app cannot record that year. Not a bug |
   | `400`, `required` on `collected_on` / `invoiced_on` | The regime's basis needs that date (cash: collected, accrual: invoiced) | Step 4 shows the basis |
   | `400`, `unknown_category` / `has_cfdi_mismatch` | Expense category not in the catalog; a folio with "sin factura" | The user fixes the form |
   | `402 entitlement_required`, `incomes_remaining` / `expenses_remaining` | Trial cap spent (5 + 5 per user, all taxpayers). Deleting does not free a slot | Step 4 shows the counters. Expected; M7 subscriptions |
   | `402` without details | Read-only user (trial over, no plan): create and edit are refused | Expected; read, delete and Undo still work |
   | `404 not_found` | Another user's id, a deleted or purged record, the module flag OFF, or `module_taxpayers` OFF | Step 2, then step 4 |
   | `409 conflict`, `id` / `id_reused` | The app resent a form id with different content, or the record was deleted | Step 4: look the id up. A client bug if the app reuses ids across forms |
   | `409 conflict` without details | The id belongs to another user | Nothing to investigate: client uuids should never collide |
   | `422`, `regime_not_active` | A date outside the regime's `[start_period, end_period]`. Often a regime started this month by onboarding (QA D-9) | Step 4 shows the windows. The user edits the regime's start, or picks the right regime |
   | `422`, `taxpayer_archived` | Archived taxpayers are frozen | The user unarchives it |
   | `422`, `iva_amount_mismatch` / `rate_not_valid_on_date` / `withholding_*` / `activity_not_held` | A domain rule ([incomes](../modules/incomes.md#iva), [expenses](../modules/expenses.md#iva)) | The `details[].field` names the field; the user compares with the CFDI. Do not ask for the amount |
   | `429 rate_limited` | More than 300 writes per module in an hour for this user | Wait for `Retry-After` |

   "Not deductible" and "IVA not creditable" are never errors: they are `warnings[]` on a `2xx`.

2. **Are the flags ON, prerequisite included?**

   ```sh
   curl -s http://localhost:3000/v1/flags   # {"enabled":["module_auth","module_expenses","module_incomes","module_taxpayers"]}
   make -C backend flags                    # LOCAL; a module shown as "hidden" lacks module_taxpayers
   ```

3. **What do the audit events say?** The events are `income.created`, `income.updated`,
   `income.deleted`, `income.restored` and the same four for `expense.*`. Each line carries
   `user_id`, `taxpayer_id`, `income_id` or `expense_id`, and `request_id`; nothing else. In a
   shared environment, filter the log store on the same JSON fields. Locally, with the API
   output saved to `api.log`:

   ```sh
   grep '^{' api.log | jq -c 'select((.msg | test("^(income|expense)\\.")) and .user_id == "<user-uuid>") | {time, msg, taxpayer_id, income_id, expense_id, request_id}'
   ```

   A record that "disappeared" usually shows a `*.deleted` without a later `*.restored`.

4. **Record state (needs the user id).** The queries read no amount, no description and no
   folio (only whether there is one):

   ```sh
   make -C backend db-psql-app PSQL_ARGS="-v user_id=<user-uuid>" <<'SQL'
   BEGIN;
   SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
   SELECT incomes_created, expenses_created FROM user_usage;
   SELECT i.id, i.taxpayer_id, i.regime_code, r.income_basis, i.counterparty_type, i.iva_rate,
          i.invoiced_on, i.collected_on, i.booked_on, i.cfdi_uuid IS NOT NULL AS has_folio,
          i.deleted_at, i.updated_at
   FROM incomes i
   JOIN taxpayer_regimes tr USING (user_id, taxpayer_id, regime_code)
   JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
   ORDER BY i.booked_on DESC, i.id DESC;
   SELECT id, taxpayer_id, regime_code, category_code, payment_method, paid_on, has_cfdi,
          deductible, deductible_overridden, cash_over_limit, creditable_iva_cents > 0 AS iva_creditable,
          deleted_at, updated_at
   FROM expenses ORDER BY paid_on DESC, id DESC;
   SELECT taxpayer_id, regime_code, start_period, coalesce(end_period, 'open') AS end_period
   FROM taxpayer_regimes ORDER BY taxpayer_id, regime_code;
   COMMIT;
   SQL
   ```

   Expected output: `context_set = t`; one `user_usage` row (`incomes_created`,
   `expenses_created`, each at most 5 in the trial); one row per income and per expense,
   soft-deleted ones included; the regime windows (`612 | 2025-01 | open`).

   - No rows: the user id is wrong, the account was deleted, or nothing was recorded.
   - `deleted_at` set: the record was deleted and is restorable (Undo) until the M9-08 purge.
     A record deleted before its regime was ended cannot be restored (`422 regime_not_active`).
   - **Wrong month in the list.** The list month is `booked_on` (incomes) or `paid_on`
     (expenses). An accrual income (`income_basis = accrual`) books on `invoiced_on` even when
     collected later; with `collected_on` empty it is "Por cobrar".
   - **"Why is it not deductible?"** `deductible_overridden = f` with `deductible = f`: the
     default said so, because of the category for that regime and options, `has_cfdi = f` or
     `cash_over_limit = t`. `deductible_overridden = t`: the user chose it.
     [Section 3](#3-derived-columns-clients-cannot-set) explains `cash_over_limit` and
     `iva_creditable`.
   - **"I deleted one and still cannot create"**: `incomes_created` / `expenses_created` never go
     down; that is the trial rule (PRD §7), not a bug.

If an amount really looks wrong (a total that does not match the user's CFDI), the user
compares the record in the app with the CFDI. Operators do not read or correct amounts.

## 2. Why amounts never reach the logs

- Request and response bodies are never logged (the API has no access log yet), and the
  audit logger takes ids only (S-5, S-6).
- Errors are written by `adapter/http/httperr` with a fixed message per code; a PostgreSQL error
  never reaches the wire, and the domain's violations name the field, never the value.
- The mobile app keeps no record data on the device (no storage, providers dispose on leave),
  and what the forms remember in a session (last payer, payment method, Undo stacks) is
  cleared on sign-out.
- To prove it after a change, repeat the gate's scan: record with a marker description, folio
  and distinctive amount at `LOG_LEVEL=debug`, then `grep` the API output for each marker.
  The expected count is 0 for every marker.

## 3. Derived columns clients cannot set

Some columns are **tax facts the estimate sums**. If a client could write them, a user (or an
attacker with a token) could inflate deductions or move income to another month. So they are
not in the contract and the API refuses them, and the database is the backstop.

| Column | Set by | Why a client cannot set it |
|---|---|---|
| `incomes.booked_on` | The trigger `incomes_derive_booked_on`, from the regime's `income_basis` | It decides the ISR month. `cc_app` has no INSERT or UPDATE grant on it; the trigger refuses (instead of guessing) when the regime row is not the caller's |
| `expenses.cash_over_limit` | The expense domain (`Assess`), from the dated cash limit | It removes the ISR deduction and the IVA credit |
| `expenses.creditable_iva_cents` | The expense domain | It is the creditable IVA M5 sums. DB CHECKs bind it: `0 ≤ creditable ≤ iva`, and `> 0` only with a CFDI and not cash over the limit |
| `user_id`, `taxpayer_id`, `id` | The session and the path | No UPDATE grant: a record cannot move to another user or taxpayer |
| `deleted_at` | `DELETE` and `restore` only | There is no DELETE grant at all: deletes are soft |

Over HTTP, `additionalProperties: false` on create and update answers `400 validation_failed`,
code `additionalProperties`, for any of them (for example `PATCH {"booked_on": "2026-09-30"}`).
In the database, the grants say the same:

```sh
make -C backend db-psql-app <<'SQL'
SELECT has_column_privilege('incomes', 'booked_on', 'INSERT, UPDATE')   AS app_writes_booked_on,
       has_column_privilege('incomes', 'user_id', 'UPDATE')             AS app_moves_income,
       has_column_privilege('expenses', 'taxpayer_id', 'UPDATE')        AS app_moves_expense,
       has_table_privilege('incomes', 'DELETE') OR has_table_privilege('expenses', 'DELETE') AS app_hard_deletes;
SQL
```

Expected: `f | f | f | f`. A `t` is an incident: a migration widened a grant.

`cc_app` **can** write `cash_over_limit`, `creditable_iva_cents` and `deductible` (the API needs
to); the category's creditability is enforced by the domain only. Correcting a wrong tax fact
is a code fix plus a forward migration that recomputes it, never a hand `UPDATE`. If a
regime's `income_basis` ever changes, the same migration must re-derive `incomes.booked_on`.
The app follows on its next taxpayer read: it takes the basis from `TaxpayerRegime.income_basis`
(ADR-0010), so no app release is needed.

## 4. Regime end and records

Ending a regime (`DELETE /v1/taxpayers/{id}/regimes/{code}?end_period=YYYY-MM`, M2) must
never leave live records after the end: the estimate would read records in a regime the
taxpayer no longer holds.

| Ending request | Answer |
|---|---|
| A live income or expense dated after `end_period` | `422 business_rule_violation`, `end_period` / `records_after_end_period`. An income counts on the **later** of `invoiced_on` and `collected_on`; an expense on `paid_on` |
| Records in the regime, all within the window, no `confirm=true` | `409 regime_has_records`: the app asks the user to confirm |
| Same, with `confirm=true` | `200`; the records are kept |
| The last active regime | `422 last_active_regime`: archive the taxpayer instead |

Once ended, every create, edit and restore dated after `end_period` answers
`422 regime_not_active`. Soft-deleted records do not block the end, so a record deleted before
the end may become impossible to restore.

**Locks.** The end path locks the `taxpayers` row, then the `taxpayer_regimes` row
(`FOR NO KEY UPDATE`), then looks up the latest record. Every record create, update and restore
checks its window through `CheckRegime`, which takes `FOR SHARE` on the same two rows in the
same order, inside the record's own transaction. A write that checked first commits before the
lookup runs; one that checks later waits and then reads the ended window. Proven by
`backend/cmd/api/records_race_integration_test.go` (`TestRegimeEndRace_recordWrites`). Known gap:
`DELETE` of a record reads the archive flag without a lock (G-2, soft delete only).

When a user cannot end a regime, find the latest period that blocks it (ids and dates only):

```sh
make -C backend db-psql-app PSQL_ARGS="-v user_id=<user-uuid> -v taxpayer_id=<taxpayer-uuid>" <<'SQL'
BEGIN;
SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
SELECT tr.regime_code, tr.start_period, coalesce(tr.end_period, 'open') AS end_period,
       to_char(greatest(
         (SELECT max(greatest(i.invoiced_on, i.collected_on)) FROM incomes i
          WHERE i.taxpayer_id = tr.taxpayer_id AND i.regime_code = tr.regime_code AND i.deleted_at IS NULL),
         (SELECT max(e.paid_on) FROM expenses e
          WHERE e.taxpayer_id = tr.taxpayer_id AND e.regime_code = tr.regime_code AND e.deleted_at IS NULL)),
         'YYYY-MM') AS latest_record_period,
       (SELECT count(*) FROM incomes i WHERE i.taxpayer_id = tr.taxpayer_id
          AND i.regime_code = tr.regime_code AND i.deleted_at IS NOT NULL) AS deleted_incomes,
       (SELECT count(*) FROM expenses e WHERE e.taxpayer_id = tr.taxpayer_id
          AND e.regime_code = tr.regime_code AND e.deleted_at IS NOT NULL) AS deleted_expenses
FROM taxpayer_regimes tr
WHERE tr.taxpayer_id = :'taxpayer_id'
ORDER BY tr.regime_code;
COMMIT;
SQL
```

`end_period` must be at or after `latest_record_period`. The user either picks a later end, or
moves or deletes the later records first; operators never edit them.

## 5. Enable `module_incomes` and `module_expenses` in a shared environment

Both flags stay **OFF** in staging and production until every item holds. Each item is
verified by whoever flips the flag and recorded in the deploy PR. Source:
[security gate §5](../reviews/M3-M4-security-gate.md#5-preconditions-to-enable-module_incomes--module_expenses-in-staging-or-production).

- [ ] **P-1 PR #103 merged** with green checks: it declares the `module_taxpayers` prerequisite
      (without it both modules answer, and write, with taxpayers OFF) and clears session
      memory on sign-out. *Merged into `dev` as `8bd3194`*; check the target environment's
      build contains it and re-run its preset smoke
      (`backend/cmd/api/records_preset_smoke_integration_test.go`) there.
- [ ] **P-2 Parent modules**: the `module_auth` ([auth runbook §6](auth.md#6-enable-module_auth-in-a-shared-environment))
      and `module_taxpayers` ([taxpayers runbook §5](taxpayers.md#5-enable-module_taxpayers-in-a-shared-environment))
      preconditions hold in the same environment. Their item P-5 (records lookup wired) is
      satisfied by M3/M4.
- [ ] **P-3 Gates**: M3-12 / M4-13 tech lead PASS, M3-13 / M4-14 security PASS, REL-09 (deploy
      security gate) PASS.
- [ ] **P-4 `APP_ENV` is not `dev`**: the dev email adapter logs one-time codes.
- [ ] **P-5 Limiter healthy**: the per-user write budgets fail closed, so a limiter outage turns
      every record write into `429`. Alert on `rate limit unavailable`.
- [ ] **P-6 Logs retained** with the `income.*` / `expense.*` audit events (ids only).
- [ ] **P-7 Mobile release build** passes `make mobile-artifact-check` (no source maps, no
      internal URLs) and uses an HTTPS API base URL.
- [ ] **P-8 CPA review** of [CPA-REVIEW.md](../reviews/CPA-REVIEW.md) (income basis of 620/622/624,
      booking dates, category catalog, cash rule, creditable IVA) before the estimate built on
      these records is shown as anything but provisional. A product precondition, not a
      security one.

Then flip both with the environment's flag tooling (`cmd/flags` as `cc_owner`, credentials from
Secret Manager): `enable module_incomes`, `enable module_expenses`. Check that `GET /v1/flags`
lists both within 5 s; if either is missing, `module_taxpayers` is OFF. To hide them again,
`disable` each: the data stays, and the estimate keeps reading it (OQ-8).
