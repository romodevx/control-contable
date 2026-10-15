# ADR-0012: Encryption key rotation and re-sealing

| Field | Value |
|---|---|
| Status | **Proposed** (2026-10-15, WP3 branch `feature/key-reseal-job`). It records decisions that are already built and tested. The architect and tech-lead accept or amend it in review |
| Date | 2026-10-15 |
| Related | WP3 encryption key rotation / re-seal job, [ADR-0011](0011-access-and-authorization.md), migration `20261024100000_key_reseal.sql`, [key rotation runbook](../runbooks/key-rotation.md) |
| Owner | architect (drafted by backend-sr) |

## Context

The application seals RFCs, account email addresses, push tokens, and short-lived auth email
outbox payloads with AES-256-GCM. These keys need rotation without interrupting reads or
requiring a one-time table rewrite. The worker must only see the rows and columns needed for a
rewrite, remain safe if two instances run concurrently, and report when stored values no longer
depend on a retired key.

Before rotation support, ciphertexts already had a key id in their header. Rows written before
any rotation use the original key id (for example, `1`); they are legacy-key rows, not rows that
need a data migration to add an id.

## Decision

### Keyring and ciphertext format

- A keyring has one active key and at most one previous key during a rotation. `Seal` always
  uses the active key. `Open` reads the key id from each ciphertext and selects that key, using
  the previous key while it remains configured. `KeyIDOf`, `NeedsReseal`, and `Reseal` support
  inspection and rewriting without exposing plaintext outside the keyring.
- The key id is stored in each ciphertext header and authenticated with the ciphertext. Rows
  sealed before the first rotation already carry their original id (normally `1`); there is no
  data migration to label them. A rotation changes configuration and then re-seals rows to the
  new active id.
- Active key settings use `X_KEY` and `X_KEY_ID`; optional retired settings are derived as
  `X_PREVIOUS_KEY` and `X_PREVIOUS_KEY_ID`. The configured key ids must be distinct.

The sealing keys and exact settings are:

| Sealed value | Active key and id | Previous key and id |
|---|---|---|
| `taxpayers.rfc_encrypted` | `RFC_ENC_KEY`, `RFC_ENC_KEY_ID` | `RFC_ENC_PREVIOUS_KEY`, `RFC_ENC_PREVIOUS_KEY_ID` |
| `users.email_encrypted` | `AUTH_EMAIL_KEY`, `AUTH_EMAIL_KEY_ID` | `AUTH_EMAIL_PREVIOUS_KEY`, `AUTH_EMAIL_PREVIOUS_KEY_ID` |
| `devices.token_ciphertext` | `PUSH_TOKEN_ENC_KEY`, `PUSH_TOKEN_ENC_KEY_ID` | `PUSH_TOKEN_ENC_PREVIOUS_KEY`, `PUSH_TOKEN_ENC_PREVIOUS_KEY_ID` |
| `auth_email_outbox.payload` | `AUTH_OUTBOX_KEY`, `AUTH_OUTBOX_KEY_ID` | `AUTH_OUTBOX_PREVIOUS_KEY`, `AUTH_OUTBOX_PREVIOUS_KEY_ID` |

### Resumable worker job

- The worker registers `key_reseal`, which runs every 15 minutes with a 10-minute run timeout.
  It uses pages of 200 rows and processes at most 50 pages per column per run. The migration
  bounds a page to 1–1000 rows.
- Re-sealable columns are `users.email_encrypted`, `taxpayers.rfc_encrypted` (including
  archived taxpayers), and `devices.token_ciphertext`. The worker always has the push token
  keys for its reminders job; it loads `RFC_ENC_KEY` and `AUTH_EMAIL_KEY` (and their ids and
  previous-key settings) for re-sealing only when supplied. It still reports `key_usage()` for
  every target when those optional keys are absent.
- Pending rows are found with keyset pagination over row ids. Re-sealing is resumable and
  idempotent: a row's key id is its checkpoint, so completed rows leave the pending set and a
  later run picks up the rest. Each row is processed in its owner's unit of work under RLS;
  the worker locks the row, re-checks its key id, opens it, seals with the active key, and updates
  the ciphertext and key id together. A row it cannot open is counted and left untouched.
- Two workers, or old and new worker instances running together, are safe: the row lock
  serializes the rewrite and the second worker re-checks the key id, so each row is re-sealed
  once. API instances must have both active and previous keys before any worker re-seals. An API
  instance still on old-only configuration cannot open new-key rows.
- Logs contain counts only, never plaintext, ciphertext, row ids, or key material. The messages
  `key reseal column`, `key usage`, and `key reseal finished` report per-column progress and
  usage. `rows_on_old_keys` is the readiness metric: it counts rows whose key id differs from
  the active id; unreadable rows remain on an old key and are also reported in `unreadable`.

### Least privilege and lookup hashes

- `cc_worker` gets only the column grants and per-user RLS policies needed to lock and update
  each sealed value and its key id. Cross-user enumeration is done by SECURITY DEFINER
  `reseal_pending_users`, `reseal_pending_taxpayers`, and `reseal_pending_devices`, owned by
  `cc_worker_definer`; they return ids and owner ids only. SECURITY DEFINER `key_usage()` returns
  target, version, and counts only. `cc_worker` alone may execute these functions.
- The pending functions and `key_usage()` use the migration's narrowly scoped
  `taxpayers_key_reseal` policy to include archived taxpayers. RLS still scopes each actual
  rewrite to the owning user's context. The definer functions never return ciphertext or hashes.
- `email_hash`, `rfc_hash`, and `token_hmac` are peppered lookup hashes, not ciphertext, so they
  cannot be re-sealed. Migration `20261024100000_key_reseal.sql` adds `email_hash_version`,
  `rfc_hash_version`, and `token_hmac_version`, each defaulting to `1`, to identify the pepper
  used. `key_usage()` reports counts by hash version as well as encryption key id. Runtime roles
  cannot change these version columns yet. Pepper rotation is future work and out of scope for
  this re-seal job.
- `auth_email_outbox.payload` is sealed but short-lived (15 minutes). Its previous key stays
  configured until queued rows drain; this job does not re-seal the outbox.

## Consequences

- Rotations deploy in two phases: first every API and worker instance gets the new key in the
  previous-key slot while the old key stays active (the slot means "also opens"), then the two
  are swapped. A single deploy that adds and activates the new key at once leaves a rolling
  window in which old API instances cannot open rows sealed by new instances or re-sealed by the
  worker, which starts on its own. The worker then runs until `rows_on_old_keys` is zero for the
  rotating columns, and the old key is removed with another redeploy. The full procedure and rollback limits are in the
  [key rotation runbook](../runbooks/key-rotation.md).
- Legacy ciphertexts remain readable because their key id already identifies the original key;
  there is no bulk migration before a rotation.
- A missing or incorrect previous key, or an unreadable ciphertext, prevents that row from
  becoming ready. Operators must retain the old key until verification and the rollback window
  have passed.
- Lookup-hash pepper rotation needs a separate design and implementation; version columns only
  prepare the schema and readiness report.

## Alternatives considered

- **Rewrite every row in one migration:** rejected because it would create an unbounded
  deployment operation and require plaintext handling in migration code.
- **Store a separate database checkpoint:** rejected because the ciphertext key id already
  identifies pending rows; a completed rewrite naturally removes a row from the pending query.
- **Grant `cc_worker` cross-user table access:** rejected because it would expose unrelated user
  data. Definer functions return identifiers only, and rewrites run under per-user RLS.
- **Re-seal peppered lookup hashes:** not possible; an HMAC is one-way. Hash versions prepare
  for a separate pepper-rotation process instead.
- **Re-seal outbox payloads:** unnecessary for the short 15-minute lifetime; keeping the old
  outbox key until the queue drains is simpler.
