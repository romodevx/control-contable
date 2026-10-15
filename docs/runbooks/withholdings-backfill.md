# Runbook: withholdings backfill (`cmd/backfill-withholdings`)

The server derives every ISR and IVA withholding from the facts of an income or an expense and
stores the result with the decision table's `table_version`
([ADR-0008](../adr/0008-withholdings-are-derived.md), [MODULE-withholdings §5.2, §5.3](../project-specs/MODULE-withholdings.md)).
Two kinds of records need this command:

- records written **before automatic derivation** (incomes before WH-06, expenses before WH-07),
  which still have typed amounts and no facts, soft-deleted ones included. Until a persona moral's
  deleted expense is backfilled, restoring it answers `400 supplier_kind required`;
- records derived by an **older `table_version`**. Every bump of `fiscal.WithholdingTableVersion`
  (a CPA answer, a new year's rates; see [tax-params-update.md](tax-params-update.md)) is followed
  by a run of this command.

| | |
|---|---|
| Owner | backend-sr (run), owner (reads the dry run before `-apply`) |
| Role | `cc_worker` (`DATABASE_URL`, the worker's secret). Never `cc_owner`: under FORCE RLS it sees no row. The command refuses a superuser, owner or BYPASSRLS login in every environment |
| Grants | Migration `20261018100000_withholdings_backfill_worker`: see [data-model.md](../architecture/data-model.md#database-roles). The user list comes from the SECURITY DEFINER `withholding_backfill_users()` (ids only), owned by the NOLOGIN `cc_worker_definer`; `cc_worker` itself reads no row outside a user's context |
| Gate | WH-15 (security) before running against a shared environment |

## 0. Once per environment: the `cc_worker_definer` role

The migration refuses to run, with a hint, until the bootstrap role exists (cc_owner has
NOCREATEROLE). Locally `make -C backend db-reset` recreates the roles from
`infra/local/postgres/bootstrap/roles.sql`. Elsewhere, as the bootstrap superuser:

```sql
CREATE ROLE cc_worker_definer NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT;
GRANT cc_worker_definer TO cc_owner WITH INHERIT FALSE, SET TRUE;
```

## Shared environments (staging, production): the safe procedure

Locally, `make -C backend backfill-withholdings` is enough. Anywhere shared, every step below is
mandatory and is done in this order; skipping one is a stop.

1. **Gate.** WH-15 (security) has passed for this version of the command, and the owner has
   agreed the window. Pick a low-traffic window: rows are locked `FOR NO KEY UPDATE` per user
   while derived, so an edit of the same record waits for that user's transaction.
2. **Credentials.** `DATABASE_URL` is the `cc_worker` secret, read from the secret manager into
   the environment of the process that runs the command (the worker's runtime, not a laptop
   shell history). Never the owner's or a superuser's URL: the command refuses them anyway.
   Never paste the URL into a ticket, a PR or the report.
3. **Backup.** Take a restorable point first: a managed snapshot / PITR mark, or a `pg_dump`
   of `incomes` and `expenses` by the bootstrap superuser (`cc_owner` cannot take it: under FORCE
   RLS it sees no row). Note the snapshot id or LSN and confirm it completed before going on.
   The dump holds every user's records: it stays encrypted, in the environment's backup store,
   and is deleted with that store's retention, never copied elsewhere.
4. **Dry run first.** Run without `-apply` (section 1). Keep the report with the task (it holds
   record ids and counts only). The owner reviews "different amounts" and "assumed facts"
   before anything is written.
5. **Metrics, before.** Record in the task: the dry-run counts per table, the start time, and
   the command's duration (`time ...`). A count much larger than the number of legacy records
   (or than the records of the bumped table version) is a stop: find out why first.
6. **Apply** (section 2), under `time`, with the same `DATABASE_URL`. A failure stops the run
   with the users before it committed; fix the cause and re-run, it resumes (idempotent).
7. **Metrics, after.** Run the dry run again (section 3) and record: `0 to re-derive` per table
   except the "need user input" ids, which must equal the first dry run's; the apply's duration;
   API error rate and latency over the window (no 5xx or lock-wait spike while it ran).
8. **Rollback**, only if the derivations are wrong: restore the derived columns of the listed
   ids from the step 3 backup (the command writes no other column), then fix the kernel and run
   again. Never drop the backup until step 7 is signed off.

## 1. Dry run (default): read the report

```sh
make -C backend backfill-withholdings     # local stack only
# elsewhere, from backend/: APP_ENV=<env> DATABASE_URL=<cc_worker URL from the secret manager> go run ./cmd/backfill-withholdings
```

Nothing is written (each user is read in a READ ONLY transaction). The report lists, per table:

```
withholdings backfill: DRY RUN (nothing written; re-run with -apply to write)
table_version: 2026.1
users visited: 2

incomes: 3 to re-derive, 2 with different amounts, 2 with assumed facts, 0 need user input
  different amounts:
    0199d4a3-…
  assumed facts (ask the user to review):
    0199d4a3-…
```

- **to re-derive**: records without facts or a complete derivation, or with another `table_version`.
- **different amounts**: the derived ISR or IVA withheld differs from the stored one. The estimates
  of those months change after `-apply` (M8 history shows them as "changed").
- **assumed facts**: legacy records whose facts follow §5.2. For an income, `counterparty_type`
  maps `persona_moral` → `pm`, `persona_fisica` → `pf_with_activity`, `publico_general` →
  `pf_no_activity_or_public` and `plataforma` → `pm` via a platform; the operation is the regime's
  default, and a 606 exempt rent is housing. A persona moral's expense with a typed withholding
  gets a 612 persona física supplier, without one a persona moral, and the operation comes from the
  category. The user should confirm them.
- **need user input**: records the rules refuse with the facts they hold (a regime the taxpayer
  no longer holds, an unknown category, a date out of range). They are left as stored; the user
  fixes them by editing the record.

The report holds record ids and counts only: no amount, description or user id.

## 2. Apply

```sh
make -C backend backfill-withholdings ARGS="-apply"
```

Each user runs in a transaction of their own with `app.current_user_id` set, so RLS scopes every
read and write; rows are locked `FOR NO KEY UPDATE` while derived. Only the facts, the derived
withholding columns and an expense's creditable IVA split are written: never an amount, a date,
the deduction verdict or a description. The first error stops the run: the failing user's
transaction rolls back and the users before it stay done. Fix the cause and run again.

## 3. Verify

Run the dry run again: every table must report `0 to re-derive`, except the records listed under
"need user input", which stay listed until the user edits them. The command is idempotent: a
derived record is not visited twice.

## 4. After a `table_version` bump

1. Merge and deploy the kernel change that bumps `fiscal.WithholdingTableVersion` (new writes
   derive with the new table at once).
2. Run the dry run, review "different amounts" with the owner, then `-apply`.
3. Record the run (date, counts) in the PR or task that bumped the version.

## 5. Not done here

API contract 0.12.0 (WH-11 / WH-12, ADR-0008 Amendment 2) does **not** depend on this
backfill having run: a record it has not derived is served with its facts (an income's from
`counterparty_type`, the same §5.2 mapping as here) and `withholding: null`, and an edit
derives it. Run the backfill so those records show their derivation and the estimate uses the
derived amounts, not as a deploy precondition.

The step that drops the `incomes.counterparty_type` column (PLAN WH-08) needs this backfill
applied in every environment first: the column is the only source of a legacy income's
counterparty. It ships as its own migration after the run. The API stopped accepting and
serving the column in 0.12.0 (WH-11); writes still fill it as a mirror until the drop.
