# Runbook: Taxpayers & Onboarding (`module_taxpayers`)

Operations for module M2 ([module page](../modules/taxpayers.md),
[spec](../project-specs/MODULE-taxpayers.md)). Every command in a `sh` block was run on a
local stack on 2026-10-14. Procedures for shared environments (staging, production) say so
explicitly. Read the whole procedure before you start one.

Ground rules:

- **The RFC never leaves the vault.** Never ask a user for their RFC, never paste one (full or
  masked) into a ticket, chat or shell history, and never decrypt `rfc_encrypted` to look at
  it. Work with the user id, the taxpayer id (both UUIDs) and the `request_id` of a response.
  Nothing in this runbook needs the plaintext.
- **RLS applies to operators too.** `taxpayers` and `taxpayer_regimes` are
  `FORCE ROW LEVEL SECURITY` with policies only for `cc_app`, and no role has BYPASSRLS.
  Per-user SQL runs as `cc_app` inside a transaction that sets `app.current_user_id`, exactly
  as the API does. A wrong user id returns no rows, it does not fail.
- Locally, `make -C backend db-psql-app` runs SQL from stdin as `cc_app`, and only against the
  local stack. Shared environments run the same SQL as `cc_app` through their own connection
  (Cloud SQL Auth Proxy, credentials from Secret Manager).
- The catalog tables `regimes` and `regime_compatibility` are global, not per user: `cc_app`
  can read them without a user context, and only migrations write them.

## 1. Investigate a user's taxpayer issue, without exposing the RFC

You need the user id (from the account, as in the
[auth runbook §4](auth.md#4-investigate-a-user-cant-log-in-without-pii)) and, if the user
has it, the `request_id` of the failed call.

1. **What did the API answer?** Most M2 problems are a rule working as designed.

   | API answer | Meaning | Next step |
   |---|---|---|
   | `400 validation_failed`, field `rfc` | The RFC was refused: `rfc_length_mismatch` (física 13, moral 12), `rfc_format`, `rfc_date`, `rfc_generic` | The user retypes it. A wrong **check digit** is never a 400: it is a warning and the taxpayer was saved |
   | `409 rfc_already_registered` | The same RFC is already in **this** account, possibly archived | Step 4: look for an archived taxpayer; the user unarchives it instead of creating a new one |
   | `422 taxpayer_limit_reached` | 10 taxpayers, archived included | The user deletes one (archiving does not free a slot). The cap is not raised in v1 (OQ-4) |
   | `422 invalid_regime_combination` | Two overlapping regimes the catalog refuses | `message` is the reason; check the pair with §4. If the user's accountant disagrees, it is a catalog change (§3) |
   | `422 business_rule_violation` | `taxpayer_archived`, `last_active_regime` or `records_after_end_period` | Unarchive first; archive the taxpayer instead of its last regime; choose a later `end_period` |
   | `409 regime_has_records` | Ending a regime that has records, without `confirm=true` | The app asks for confirmation; records are kept |
   | `402 entitlement_required` | Read-only user (trial over, no plan) | Expected: only read, archive and delete are allowed. M7 subscriptions |
   | `404 not_found` | Another user's id, a deleted taxpayer, or `module_taxpayers` OFF | Step 2 |
   | `429 rate_limited` | More than 30 M2 writes in an hour for this user | Wait for `Retry-After` |

2. **Is the flag ON?**

   ```sh
   curl -s http://localhost:3000/v1/flags   # {"enabled":["module_auth","module_taxpayers", …]}
   make -C backend flags                    # LOCAL; lists every flag and its state
   ```

3. **What do the audit events say?** Audit lines carry ids only: `user_id`, `taxpayer_id`,
   `regime_code` for regime events, and `request_id`. The events are `taxpayer.created`,
   `taxpayer.rfc_corrected`, `taxpayer.archived`, `taxpayer.unarchived`, `taxpayer.deleted`,
   `taxpayer_regime.added` and `taxpayer_regime.archived`. Renames and option changes are not
   audited. In a shared environment, filter the log store on the same JSON fields. Locally,
   with the API output saved to `api.log`:

   ```sh
   grep '^{' api.log | jq -c 'select((.msg | startswith("taxpayer")) and .user_id == "<user-uuid>") | {time, msg, taxpayer_id, regime_code, request_id}'
   ```

4. **Taxpayer state (needs the user id).** The query reads no RFC column
   (no `rfc_encrypted`, `rfc_hash` or `rfc_masked`, and no `display_name`: it is whatever the
   user typed, often a person's name).

   ```sh
   make -C backend db-psql-app PSQL_ARGS="-v user_id=<user-uuid>" <<'SQL'
   BEGIN;
   SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
   SELECT t.id, t.person_type, t.created_at, t.archived_at, t.rfc_check_digit_valid, t.rfc_key_id,
          string_agg(r.regime_code || ' ' || r.start_period || '..' || coalesce(r.end_period, 'open'),
                     ', ' ORDER BY r.regime_code) AS regimes
   FROM taxpayers t LEFT JOIN taxpayer_regimes r ON r.taxpayer_id = t.id
   GROUP BY t.id ORDER BY t.created_at;
   SELECT count(*) AS toward_cap, count(*) FILTER (WHERE archived_at IS NOT NULL) AS archived
   FROM taxpayers;
   COMMIT;
   SQL
   ```

   Expected output: `context_set = t`, one row per taxpayer (id, person type, dates, the
   check-digit flag, the key id and the regime windows such as `612 2026-10..open`), then
   `toward_cap` and `archived`.

   - No rows: the user id is wrong, the account was deleted, or the user has not onboarded.
   - `archived_at` set on a taxpayer the user "lost": it is hidden from the switcher. A `409
     rfc_already_registered` with an archived taxpayer in the list almost always means this
     one; the user turns on "Mostrar archivados" in the switcher and unarchives it from its settings.
   - `rfc_check_digit_valid = f`: the RFC was saved with the D-RFC warning. That is not an
     error; if the user says the RFC is wrong, they correct it in the app (`PATCH rfc`).
   - `rfc_key_id` is the encryption key id of the row. It is 1 everywhere until the M9
     re-seal job exists (§2); another value is an incident.
   - `regimes`: windows explain `records_after_end_period` and `invalid_regime_combination`
     (two windows that overlap).

You cannot tell from the database whether an RFC belongs to a user: `rfc_hash` is keyed by a
secret pepper and the user id, so there is no lookup from an RFC to an account, by design.

## 2. Do not rotate `RFC_ENC_KEY` until the M9 re-seal job exists

The API holds exactly one RFC encryption key (`RFC_ENC_KEY`, id `RFC_ENC_KEY_ID`) and no
previous key. Each row records the key id it was sealed with (`rfc_key_id`). Swapping the
secret today would leave every stored ciphertext under a key the API no longer has: the
masks and the duplicate check would keep working, so nothing breaks visibly, but every
stored RFC would become unreadable dead data, with no way back once the old secret version
is destroyed. That is security gate finding G-2 and enable precondition P-4.

`RFC_INDEX_PEPPER` is the same: `rfc_hash` has no version column, so a new pepper makes
every existing hash miss. The duplicate check would let the same RFC in twice per account.

Until the M9 job ships:

- `RFC_ENC_KEY_ID` stays at its first value, and the secret versions of both keys are never
  disabled or destroyed.
- A suspected key leak is handled by a **planned** maintenance, not by swapping the secret:
  build the re-seal job early (below), then rotate.
- In dev, the API may start without the keys and use ephemeral ones: every restart then
  makes earlier rows unreadable and resets duplicate detection. Keep `CC_RFC_ENC_KEY` and
  `CC_RFC_INDEX_PEPPER` in `infra/local/.env` (`make -C backend init-env` writes them).

What the M9 job must do (PLAN M2 follow-ups, owners backend-sr and db-sr): load the old key as
a previous keyring entry (`RFC_ENC_KEY_PREVIOUS` + id); for each row, open `rfc_encrypted`
under its `rfc_key_id` and re-seal it with the active key and the same binding (key id,
taxpayer id, user id), updating `rfc_encrypted` and `rfc_key_id` together, per user under
RLS, resumable and idempotent; retire the old key once no row names it. Pepper rotation is a
re-hash from the opened RFC with a new hash-version column, in one maintenance window,
before the new pepper serves writes.

## 3. Update the regimes catalog

`regimes` and `regime_compatibility` change **only through a new, forward migration**. Never
edit `20261013180000_regimes_catalog.sql` (it is applied everywhere and goose will not rerun
it), never `UPDATE` the tables by hand in a shared environment (`cc_app` has no write grant
anyway), and never AutoMigrate. Typical changes: a CPA verdict from
[CPA-REVIEW.md](../reviews/CPA-REVIEW.md) (a pair becomes `allowed = false`, or loses
`needs_cpa_review`), the M5-19 `explanation` texts, or a new SAT key.

The rules a change must keep are in
[spec §2.5](../project-specs/MODULE-taxpayers.md#25-compatibility-data-not-code) (read the
current version: it is being reconciled with the seed) and in the table CHECKs: a pair is
stored once with `regime_a < regime_b`; `needs_cpa_review` rows must be `allowed = true`
and carry a `review_note`; a verified row carries no note.

1. On a branch, add `backend/migrations/<YYYYMMDDHHMMSS>_<snake_case>.sql`, with a timestamp
   later than every existing file, a `-- +goose Up` and a non-empty `-- +goose Down` that
   restores the previous values. For example, settling a pair:

   ```sql
   -- +goose Up
   UPDATE regime_compatibility
   SET allowed = false, needs_cpa_review = false, review_note = NULL,
       reason = '<reason shown to the user, es-MX>', legal_basis = '<article>', source_url = '<url>'
   WHERE person_type = 'fisica' AND regime_a = '612' AND regime_b = '621';

   -- +goose Down
   UPDATE regime_compatibility
   SET allowed = true, needs_cpa_review = true, review_note = '<previous note>',
       reason = '<previous reason>', legal_basis = '<previous>', source_url = '<previous>'
   WHERE person_type = 'fisica' AND regime_a = '612' AND regime_b = '621';
   ```

   Turning a pair to `allowed = false` does not touch taxpayers that already hold both
   regimes; it refuses new overlaps only. Say so in the PR and decide with the owner whether
   affected users are told.

2. Update the catalog tests first (TDD): `backend/migrations/regimes_catalog_integration_test.go`
   asserts the research rules (`TestRegimeCompatibility_ResearchRules`) and the
   allowed-with-warning pairs. Then apply and test locally:

   ```sh
   make -C backend migrate-up
   make -C backend migrate-status
   make -C backend migrate-down     # proves Down works
   make -C backend migrate-up
   cd backend && go test -count=1 -tags=integration -run 'TestRegime' ./migrations/
   ```

3. Tick the item in `CPA-REVIEW.md` and fix the spec table in the same PR.
4. Shared environments get the migration through the normal deploy (`cmd/migrate` as
   `cc_owner`). The catalog is read on every request, so no restart is needed.

## 4. Check what the catalog says about a regime pair

When a user (or their accountant) asks why two regimes are refused or flagged:

```sh
make -C backend db-psql-app PSQL_ARGS="-v code=626 -v person_type=fisica" <<'SQL'
SELECT regime_a, regime_b, allowed, needs_cpa_review, reason
FROM regime_compatibility
WHERE person_type = :'person_type' AND :'code' IN (regime_a, regime_b)
ORDER BY regime_a, regime_b;
SQL
```

`allowed = f` refuses an overlap (`422 invalid_regime_combination`). `needs_cpa_review = t`
allows it with the `regime_combination_needs_cpa_review` warning. The `reason` column is the
message the user saw.

## 5. Enable `module_taxpayers` in a shared environment

`module_taxpayers` stays **OFF** in staging and production until every item holds. Each item
is verified by whoever flips the flag and recorded in the deploy PR. Source:
[M2-14 §5](../reviews/M2-14-security-gate.md#5-preconditions-to-enable-module_taxpayers-in-staging-or-production).

- [ ] **P-1 `module_auth` preconditions** ([auth runbook §6](auth.md#6-enable-module_auth-in-a-shared-environment))
      hold in the same environment: M2 is unreachable without M1.
- [ ] **P-2 Keys from the secret manager.** `RFC_ENC_KEY` (32 random bytes) and
      `RFC_INDEX_PEPPER` (at least 32 random bytes) are Secret Manager secrets, distinct
      from each other, from every `AUTH_*` key and per environment; never the local values;
      nothing in the image or the service YAML. Access to the pepper is restricted like the
      encryption key: a leaked pepper makes `rfc_hash` brute-forceable.
- [ ] **P-3 `APP_ENV` is not `dev`**: only `dev` starts with ephemeral RFC keys.
- [ ] **P-4 No RFC key rotation** until the M9 re-seal job ships (§2): `RFC_ENC_KEY_ID` stays
      at its first value.
- [ ] **P-5 `module_incomes` and `module_expenses` stay OFF** until `NoRecords` is replaced by
      the M3/M4 record lookup (G-1). This does not block `module_taxpayers` itself.
- [ ] **P-6 Monitoring**: logs retained with the audit events of §1; alert on
      `rate limit unavailable` errors (the limiter fails closed, so an outage turns every M2
      write into `429`).
- [ ] **P-7 Gates**: M2-13 PASS, M2-14 PASS, PR #68 merged with green checks, REL-09 (deploy
      security gate) PASS.
- [ ] **P-8 Mobile release build** passes `mobile-artifact-check` (no source maps, no internal
      URLs), with an HTTPS API base URL.
- [ ] **Compatibility matrix settled** (M2-13 FU1): spec §2.5 and the seed agree, either by
      amending the spec or by a forward migration (§3).

Then flip it with the environment's flag tooling (`cmd/flags` as `cc_owner`, credentials from
Secret Manager). Check that `GET /v1/flags` lists `module_taxpayers` within 5 s.
