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
| Grants | Migration `20261018100000_withholdings_backfill_worker`: see [data-model.md](../architecture/data-model.md#database-roles) |
| Gate | WH-15 (security) before running against a shared environment |

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

The contract step that drops `incomes.counterparty_type` (PLAN WH-08) needs this backfill
applied in every environment first: the column is the only source of a legacy income's
counterparty. It ships as its own migration after the run, together with the API's deprecated
`counterparty_type` mirror (WH-11).
