# Runbook: records purge (worker job `records_purge`)

The worker deletes, every hour, what the product promises not to keep (PLAN M9-08, data-model
"Deletion"). It runs **whatever the module flags say**: retention is a legal promise
(LFPDPPP, PRD §8), not a feature, so there is no flag to turn it off.

| | |
|---|---|
| Owner | db-sr (function, policies), backend-sr (job) |
| Code | `backend/internal/retention` (job + adapter), registered in `backend/cmd/worker` |
| Database | Migration `20261020120000_records_purge`: `purge_soft_deleted(batch_size)`, SECURITY DEFINER, owned by the NOLOGIN `cc_worker_definer`; EXECUTE for `cc_worker` only |
| Schedule | Every hour on the leader worker; at most 50 passes of 1000 rows per target per run, each pass its own short transaction; a run is cut at 10 minutes and the rest waits for the next tick |
| Logs | One line per run: `"msg":"records purge finished"`, `passes`, `more` (true = rows left for the next run) and `purged` = row counts per target. Counts only: never an id, a hash or a value |

## 1. What goes, and when

The cutoffs are computed by the database clock inside the function. The caller only chooses a
batch size (1..5000); it cannot move a cutoff.

| Target | Deleted when | Why that long |
|---|---|---|
| `incomes`, `expenses`, `personal_deductions` | soft-deleted (`deleted_at`) more than 30 days (720 h) ago | Undo / restore window |
| `refresh_tokens` | the family's absolute expiry, or its revocation, is more than 24 h old | Rotated tokens of a **live** family stay (reuse detection needs them). A family that only slid out of its 30-day window goes at its absolute expiry (≤ 90 days after login) |
| `email_verifications` | pending sign-up expired more than 24 h ago | Until then a late code still answers "expired" |
| `email_verification_failures` | the 24 h failure window ended | The consume function already treats it as a fresh window |
| `notification_log` | created more than 13 months ago (M9-06, migration `20261022101000`) | MODULE-reminders §2.6 retention: the reminder log answers "was this sent?" for a year plus a margin, and holds no token or PII |
| `auth_email_outbox` | expired (`expires_at`, the code's 15 min; migration `20261023100000`) | Its sealed payload holds an address and maybe a code: an expired one is never sent, so it goes at once. The delivery job also deletes expired rows on every run and delivered ones right away; this is the backstop |

Not purged here: `rate_limit_buckets`. A row that vanished mid-call would make `rate_limit_take`
return NULL (PLAN #47 FU2); that fix and the bucket purge ship together.

Live records are safe by construction, not only by the function's `WHERE`: the policies that let
`cc_worker_definer` see and delete rows (`*_purge_select`, `*_purge_delete`) match only rows past
the cutoff. `cc_worker` itself has no DELETE on any table and still sees no user row without a
user context.

## 2. Once per environment: the definer roles

Both definer roles are bootstrap roles (cc_owner has NOCREATEROLE), created **before**
`migrate up`; the migrations that hand them functions refuse to run, with a hint, until they
exist with the right posture. Locally `make -C backend db-reset` creates them from
`infra/local/postgres/bootstrap/roles.sql`. In a shared environment (staging, production), run
this once as the administrative role that created `cc_owner` (on Cloud SQL, the built-in
`postgres` user), PostgreSQL ≥ 16:

```sql
-- Auth definer functions (M1-03 onwards). Skip if it already exists (docs/runbooks/auth.md §5).
CREATE ROLE cc_auth_definer NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT;
GRANT cc_auth_definer TO cc_owner WITH INHERIT FALSE, SET TRUE;

-- Worker definer functions (WH-08 withholding_backfill_users, M9-08 purge_soft_deleted).
-- Skip if it already exists (docs/runbooks/withholdings-backfill.md §0).
CREATE ROLE cc_worker_definer NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT;
GRANT cc_worker_definer TO cc_owner WITH INHERIT FALSE, SET TRUE;
```

Check before migrating. You should get exactly two rows, each `f | f | f | f | t`:

```sql
SELECT r.rolname, r.rolcanlogin, r.rolsuper, r.rolbypassrls, m.inherit_option, m.set_option
FROM pg_catalog.pg_roles r
JOIN pg_catalog.pg_auth_members m ON m.roleid = r.oid AND m.member = 'cc_owner'::regrole
WHERE r.rolname IN ('cc_auth_definer', 'cc_worker_definer')
ORDER BY r.rolname;
```

No runtime role (`cc_app`, `cc_webhook`, `cc_worker`) may ever be a member of either.

## 3. Check it is working

On the worker's logs, filter on `"job":"records_purge"`:

- `"msg":"job finished"` every hour on exactly one instance (the leader);
- the `records purge finished` line before it: `more` should be `false` on most runs. `true`
  several runs in a row means a backlog larger than 50 000 rows per target per hour: check that
  nothing is mass-deleting, then raise the per-run limit in a reviewed change.
- `"msg":"job failed"`: read the error. A `40P01` (deadlock) or a statement timeout during a
  worker failover is harmless (the next run repeats the work); anything else is a bug.

By hand, as `cc_worker` (it deletes what is due, exactly as the job would):

```sql
SELECT target, purged FROM purge_soft_deleted(1000);
```

## 4. Safety properties (what the tests prove)

`backend/migrations/records_purge_integration_test.go` and
`backend/internal/retention/infra/purger_integration_test.go`:

- only rows past the cutoff go, for every user; live and recent rows stay, and `cc_app` keeps its
  own-rows isolation;
- the function is definer-owned, pins `search_path`, and only `cc_worker` may execute it;
  `cc_app` and `cc_worker` cannot `DELETE` a record directly;
- batches are bounded and a re-run purges nothing (idempotent);
- two overlapping runs (a worker split brain during leader failover) delete each row once: the
  later one waits for the earlier one's row locks and then skips the rows it finds gone. A
  restore that commits first wins the same way;
- each scan uses an index (`*_deleted_at_idx`, `refresh_tokens_absolute_expires_at_idx`,
  `refresh_tokens_revoked_at_idx`, `email_verifications_expires_at_idx`).
