# Runbook: encryption key rotation (`key_reseal`)

This runbook rotates one or more AES-256-GCM sealing keys and runs the worker's resumable
re-seal job. See [ADR-0012](../adr/0012-encryption-key-rotation.md) for the decision and
[ADR-0011](../adr/0011-access-and-authorization.md) for the authorization model.

| | |
|---|---|
| Owner | backend-sr (run), owner (approves the window and readiness) |
| Role | `cc_worker` for `key_reseal`; API and worker runtime roles for application traffic. Never `cc_owner`, a superuser, or a `BYPASSRLS` login for the job |
| Grants | Migration `20261024100000_key_reseal.sql`: `cc_worker` executes `reseal_pending_users(smallint, uuid, integer)`, `reseal_pending_taxpayers(smallint, uuid, integer)`, `reseal_pending_devices(smallint, uuid, integer)`, and `key_usage()`. The pending functions return ids only and are owned by `cc_worker_definer`; row rewrites use per-user RLS and column grants |
| Gate | WP3 security and tech-lead gates for this version of the worker and migration must pass before a shared-environment rotation |
| Deploy order | Deploy active and previous keys to every API and worker instance first; then run `key_reseal`; remove the previous key only after readiness is zero and the rollback retention period has passed |

## 0. Prerequisites

1. Confirm the migration `20261024100000_key_reseal.sql` is applied and the `cc_worker_definer`
   bootstrap role exists. The worker connects as `cc_worker`; do not run the readiness query as
   an owner or superuser.
2. Confirm the key rotation and the maintenance window with the owner. Identify the exact
   ciphertext target or targets to rotate and record the currently deployed key ids. Decide the
   retention period for the old key before starting.
3. Confirm that API and worker deployments can receive the same active and previous settings.
   For `RFC_ENC_KEY` and `AUTH_EMAIL_KEY`, the worker's optional re-seal configuration is loaded
   through `LoadResealKeys`; push token keys are already loaded for reminders. Outbox keys are
   used for email delivery and drain naturally rather than being re-sealed.

## 1. Generate and store the new key

1. Generate 32 random bytes and encode as hex, for example:

   ```sh
   openssl rand -hex 32
   ```

2. Put the value directly into the secret manager as `<new-key-hex>`. Never put key material in
   the repository, chat, a ticket, shell history, or logs.
3. Choose a key id not used by the current key, for example the next unused positive id within
   the configured range. Record the id, not the key value, in the change record.

## 2. Deploy both keys everywhere

For each key being rotated, set the new key and id as active and the existing key and id as
previous. Deploy these settings to **all API and worker instances**, and wait until all instances
are healthy before starting re-sealing. This ensures every live API can open either ciphertext
version while the worker writes new-key ciphertexts.

| Sealed value | Active settings | Previous settings |
|---|---|---|
| RFC | `RFC_ENC_KEY`, `RFC_ENC_KEY_ID` | `RFC_ENC_PREVIOUS_KEY`, `RFC_ENC_PREVIOUS_KEY_ID` |
| Account email | `AUTH_EMAIL_KEY`, `AUTH_EMAIL_KEY_ID` | `AUTH_EMAIL_PREVIOUS_KEY`, `AUTH_EMAIL_PREVIOUS_KEY_ID` |
| Push token | `PUSH_TOKEN_ENC_KEY`, `PUSH_TOKEN_ENC_KEY_ID` | `PUSH_TOKEN_ENC_PREVIOUS_KEY`, `PUSH_TOKEN_ENC_PREVIOUS_KEY_ID` |
| Auth email outbox | `AUTH_OUTBOX_KEY`, `AUTH_OUTBOX_KEY_ID` | `AUTH_OUTBOX_PREVIOUS_KEY`, `AUTH_OUTBOX_PREVIOUS_KEY_ID` |

Keep all unchanged keys and ids as they are. The worker needs `RFC_ENC_KEY` and/or
`AUTH_EMAIL_KEY` with their corresponding `_ID` and `_PREVIOUS_KEY` settings to re-seal those
columns. It already loads `PUSH_TOKEN_ENC_KEY` and `PUSH_TOKEN_ENC_KEY_ID` for the reminders
job; provide the previous push key settings during rotation. The `auth_email_outbox.payload`
uses `AUTH_OUTBOX_KEY`; retain its previous key until queued payloads drain (15-minute lifetime).
The `key_reseal` job does not rewrite outbox rows.

## 3. Run or wait for `key_reseal`

The worker schedules `key_reseal` every 15 minutes. A run has a 10-minute timeout, processes
pages of 200 rows, and handles at most 50 pages per column per run. It resumes automatically on
the next run; no checkpoint or manual cursor is required. A controlled restart or interrupted
run is safe. Do not start a separate ad hoc query that reads ciphertext.

For each column, the worker scans pending row ids using keyset pagination, then locks and
re-checks each row in the owner's unit of work before re-sealing it. It can process these
columns: `users.email_encrypted`, `taxpayers.rfc_encrypted` (including archived taxpayers), and
`devices.token_ciphertext`. The worker reports all targets even when it was not given optional
RFC or email keys.

## 4. Verify readiness and unreadable rows

Review logs for `key reseal column`, `key usage`, and `key reseal finished`. The column log's
`rows_on_old_keys` field is the readiness metric. It must be zero for each rotated column, and
`unreadable` must be zero. An unreadable row is left unchanged and remains in the old-key count;
stop key removal, investigate the missing key or authentication failure, and rerun after the
cause is resolved.

You can also query the exact readiness function as `cc_worker`:

```sql
SELECT target, version, row_count FROM key_usage();
```

For each rotated target, sum the `row_count` values for versions other than the active key id;
that total must be zero. The function returns counts only. It also reports hash-version counts
for `users.email_hash`, `taxpayers.rfc_hash`, and `devices.token_hmac`; these are peppered lookup
hashes and are not re-sealed by this job. Pepper rotation is out of scope.

## 5. Remove the previous key

After all rotated targets report zero `rows_on_old_keys`, `unreadable` is zero, and the owner
approves readiness, remove the corresponding `X_PREVIOUS_KEY` and `X_PREVIOUS_KEY_ID` settings
from API and worker configuration and redeploy all instances. Keep the old key in the secret
manager for the agreed rollback retention period after removal (at least 30 days unless the owner sets a longer one in the change record). Do not
delete the secret before that period ends.

For the outbox key, first confirm queued payloads have drained before removing
`AUTH_OUTBOX_PREVIOUS_KEY` and `AUTH_OUTBOX_PREVIOUS_KEY_ID`.

## 6. Rollback

- **Before step 5:** restore the old key and id as active and the new key and id as previous on
  every API and worker instance, then redeploy. The job will re-seal rows back to the restored
  active key. Keep both versions available until the reverse re-seal reaches zero old-key rows.
- **After step 5:** rollback is impossible without the old key. Restore it from the secret
  manager, configure it as previous alongside the current active key, and redeploy all API and
  worker instances before any further action. This is why the old key remains in the secret
  manager through the retention period.

For outbox payloads, retain both key versions until the 15-minute queue lifetime has passed;
the outbox is not handled by `key_reseal`.
