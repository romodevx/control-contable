# M2-14 · Security gate: M2 Taxpayers & Onboarding

| Field | Value |
|---|---|
| Scope | M2 as built on `origin/dev` @ `abb926a` (2026-10-13): `backend/internal/taxpayers/**`, migrations `20261013180000_regimes_catalog.sql` and `20261013190000_taxpayers.sql`, `cmd/api` wiring (`taxpayers.go`, `platform/config/taxpayers.go`), `platform/crypto` (AES-GCM keyring, RFC blind index), router per-user limits, mobile `lib/features/taxpayers/**` incl. the on-device active-taxpayer id. The flag tests of M2-11 live in PR #68 (not merged yet) |
| Reviewer | security |
| Date | 2026-10-13 |
| Method | `/cso --scope taxpayers` (gstack-cso 3.0.0, run `1791956722959-6a547bce3a14c89b`, daily static, finished **partial**: 0 supported findings; helper-mediated read of the vault, use cases, store, handler, limits, migration and the mobile store; no scanners in the 600 s budget). Then a manual review of the rest of the scope, `go test -race` of the module, crypto, config and `cmd/api`, `go test -tags=integration` of `internal/taxpayers/infra` and `migrations`, `govulncheck`, and an end-to-end probe of the running API against an **isolated local** postgres (compose project `cc-gate-m214`, own container and volume `cc-gate-m214-pgdata`, port 5446, API on 3046; `module_auth` + `module_taxpayers` enabled there only, then disabled, stack and volume destroyed). The shared local stack (`control-contable-postgres`, 5434) was not touched |
| Verdict | **PASS.** No supported finding of severity medium or above. `module_taxpayers` stays **OFF** in every shared environment until the preconditions in §5 hold |

## 1. Evidence

Probe totals: 66 of 67 e2e checks passed on the first full run. The one miss was a probe
error (`DELETE …/regimes/612` sent without the required `end_period`, so both the foreign
and the missing id got the same `400` from contract validation). Re-run with a valid
`end_period`: `404` = `404`, identical bodies.

| Check | Result |
|---|---|
| Unit (`-race`) | `internal/taxpayers/...`, `platform/crypto`, `platform/config`, `cmd/api`: green |
| Integration | `internal/taxpayers/infra`, `migrations` (RLS isolation, composite FKs, grants, cap race, RR refusal, cascades): green |
| `govulncheck ./...` | 0 called vulnerabilities (1 in a required module, not reachable) |
| **S-1 AAD binding** (real rows) | 18 ciphertexts read from the local stack and opened with the stack's own key: each opens under its own `(taxpayer id, user id)`; **96 same-user and 210 cross-user swaps all fail with `crypto.ErrDecrypt`**; the same taxpayer id under a forged user id fails too. AAD = `key_id ‖ taxpayer.id ‖ user_id` (keyring frame + 32 fixed-width bytes, `infra/rfcvault.go`). The id is a uuidv7 generated (or the client id canonicalized) **before** `Seal` (`usecase/taxpayers.go` `rowID`) |
| Ciphertext swap via the DB role | as `cc_app` in user B's context, copying A's `rfc_encrypted` into B's row is impossible: A's row is invisible under RLS (the subquery yields NULL → 23502) |
| **S-2 blind index** | same RFC, two users → different `rfc_hash` and different ciphertexts; the same user typing `" goma-800101 ab6 "` → `409 rfc_already_registered`. `HashScoped(user id, RFC)` under `RFC_INDEX_PEPPER` (`crypto.NewRFCIndex`), separate from `RFC_ENC_KEY` and `AUTH_EMAIL_PEPPER` |
| **Startup secrets** (`APP_ENV=staging`) | missing `RFC_ENC_KEY` → refuses; missing `RFC_INDEX_PEPPER` → refuses; key = pepper → `RFC_INDEX_PEPPER must differ from RFC_ENC_KEY`; pepper = email pepper → refuses; RFC key = access-token key → refuses; 2-byte key → `must be exactly 32 bytes`. Errors name the secret, never the value. `dev` alone may start with ephemeral keys (P-3) |
| **S-3 masking** | create / get / list return only `rfc_masked` (`GOMA******AB6`); the full RFC is absent from every body and from every column (`row_to_json` scan); `Cache-Control: no-store` on taxpayer responses |
| **S-4 RFC never in logs/errors** | 4 invalid RFCs (length, date, generic, charset) and a PATCH with an impossible date → `400` without echoing the value. API log at `LOG_LEVEL=debug` (two runs, ~60 lines): **0** full RFCs (upper or lower case), **0** masked RFCs; audit lines carry `user_id`, `taxpayer_id` (and `regime_code`) only. Mobile: no analytics/crash SDK in `pubspec.yaml`; the one `developer.log` in the module logs the failure type only; RFC fields have suggestions and autocorrect off (no keyboard learning) |
| **IDOR, 8 id ops** | user B on A's taxpayer vs a random uuid: GET, PATCH, archive, unarchive, add regime, update regime, archive regime, delete → `404`, identical body (minus `request_id`) for all 8; A's taxpayer and regime rows byte-identical before/after. Foreign client id on create → generic `409 conflict`, no owner info |
| **Composite FKs, forged ids** (as `cc_app`, B's context) | A's taxpayer id with B's `user_id` → 23503 `taxpayer_regimes_taxpayer_fkey`; A's `user_id` → RLS refusal; `person_type = moral` on a PF taxpayer → 23503; moral-only `601` on PF → 23503 `taxpayer_regimes_regime_fkey`; cross-user `UPDATE taxpayer_regimes` → `UPDATE 0` |
| **Cap race (S-5)** | user at 9, five concurrent creates → exactly `[201, 422, 422, 422, 422]`; 10 rows afterwards |
| **Read-only (OQ-8)** | trial moved to the past (LOCAL, superuser): create, unarchive, rename, RFC correction, add regime, update options → `402 entitlement_required`; get, archive taxpayer, archive regime, delete → allowed |
| **Rate budget 30/h (S-10)** | 1 create + 29 renames → all `200`; 31st write → `429` with `Retry-After: 120`; archive is counted too (`429`); reads still `200`. Budget keyed per user (`taxpayer_writes`), shared by all 8 mutations |
| **person_type immutability** | `PATCH {"person_type": …}` → `400` (`additionalProperties: false`); `UPDATE taxpayers SET person_type` as `cc_app` → `permission denied` (no column grant) |
| **RFC correction (OQ-1)** | `200`, masked only; ciphertext and hash both change (re-sealed for the same row id, re-hashed); audited `taxpayer.rfc_corrected` |
| **Cascade on account delete (S-7)** | `DELETE /v1/me` → `204`; the user's `taxpayers` / `taxpayer_regimes` rows: `0/0`. Integration guard: every FK to `users` cascades, indirect chains included (#57) |
| **Flag OFF** | `GET /v1/regimes`, `GET /v1/taxpayers`, `GET /v1/taxpayers/{id}` without a bearer → `404 not_found` (flag gate runs before authn). All 11 ops covered by the preset smoke in PR #68 |
| Mobile active taxpayer | only the taxpayer id in `SharedPreferences` (`active_taxpayer_id`), cleared on sign-out; `sharedpref` excluded from cloud backup and device transfer; `allowBackup=false`. A stale id is only used if it is in the signed-in user's own list (`resolveActiveTaxpayer`) |

Environment note: the first probe run tripped the **edge** limiter (10 rps, burst 20 per
client IP, in-process) because every probe user shares `127.0.0.1`. The API was restarted
with `RATE_LIMIT_RPS=1000 RATE_LIMIT_BURST=1000` for the probe; the per-user `taxpayer_writes`
budget (Postgres GCRA) is independent of it and was measured as above.

## 2. Threat model (STRIDE), M2 as built

Actors: authenticated user; another authenticated user (curious or hostile); a read-only user
(trial ended); an anonymous caller; an operator or attacker holding the `cc_app` credential
or a DB backup; someone with the phone.
Assets: the RFC (tax identity, PII), its ciphertext / hash / mask, `RFC_ENC_KEY` and
`RFC_INDEX_PEPPER`, the taxpayer ↔ user ownership, regime windows (they gate M3/M4 records),
the per-user cap and write budget.
Entry points: 11 operations under `/v1/regimes` and `/v1/taxpayers/**` (bearer, flag-gated),
the `cc_app` DB role, the device preference.

| | Threat | Control as built | Residual |
|---|---|---|---|
| **S** | Act on another user's taxpayer with a stolen/guessed id | user id only from the verified token; RLS `user_id = app_user_id()` FORCE on both tables; uuidv7 ids carry no authority; 404 parity on all 8 id ops | — |
| **S** | Claim another user's taxpayer through a client-supplied create `id` | PK collision under RLS → generic `409 conflict`; no owner info | Existence of a uuid is observable (409 vs 201); uuidv7 is unguessable, accepted |
| **T** | Insert regime rows on a foreign taxpayer or with a forged person type | composite FKs `(user_id, taxpayer_id, person_type)` and `(regime_code, person_type)`; RLS WITH CHECK; INSERT column grants | — |
| **T** | Change `person_type` / `user_id` / `id` after create | no UPDATE grant on those columns + trigger even for the owner; contract closes the body | — |
| **T** | Swap a ciphertext between rows (DB write access) | GCM AAD binds key id, taxpayer id and user id; any swap fails to decrypt | `rfc_masked`/`rfc_hash` are not authenticated together with the ciphertext: a DB-level attacker can still rewrite the displayed mask. Needs DB write, which already owns the data; accepted |
| **T** | Bypass the cap with concurrent creates | `taxpayer_count_for_create()` locks the user row (`FOR UPDATE`), refuses REPEATABLE READ; archived count | — |
| **T** | Overlap incompatible regimes concurrently | overlap trigger refuses RR (25000), clash lookup after the lock | — |
| **R** | Deny a tax-identity change or a deletion | audit `taxpayer.created/rfc_corrected/archived/unarchived/deleted`, `taxpayer_regime.added/archived`, ids only | Renames and option changes are not audited (not required by S-9); log retention is a deploy concern (P-6) |
| **I** | RFC in responses, logs, errors, analytics | only `rfc_masked` leaves the server; domain errors and field messages never echo it; vault/keyring/index redact in `fmt`/`slog`; no analytics SDK in the app | — |
| **I** | Cross-user linkage of the same RFC (DB/backup reader) | hash scoped by user id under its own pepper; ciphertext randomized (GCM nonce) | — |
| **I** | Offline RFC brute force from a leaked `rfc_hash` | HMAC with a 32-byte secret pepper | Pepper leak ⇒ RFCs are a small space (~10¹² per type) and become enumerable; the pepper must stay in the secret manager (P-2) |
| **I** | Enumerate whether another user registered an RFC | uniqueness is per user; `409 rfc_already_registered` only within the caller's account | — |
| **I** | Device leak of the active taxpayer | only the id (non-secret), excluded from backups, cleared on sign-out | Key is not namespaced per user (spec §2.9 says "per user"); harmless because a foreign id is ignored. Info |
| **D** | Exhaust a user's write budget / burn DB with writes | per-user 30/h on all 8 mutations, fail-closed limiter; per-IP edge limiter; cap of 10 | A client sending invalid bodies spends its own budget (budget runs before validation, accepted #60) |
| **E** | Read-only user keeps writing | `EnsureWritable` in every write use case except archive/delete (table-driven matrix test, e2e above) | — |
| **E** | API compromise rewrites the RFC of any row of the current user | UPDATE grant on `rfc_*` (needed by OQ-1) | The DB cannot tell the RFC-correction use case from other `cc_app` code; enforced by code structure (`ReplaceRFC` has one caller). Accepted |
| **E** | Archive a regime before existing records (orphan records outside the window) | `RegimeRecords` port, `end_period ≥ latest record` | **Wired to `infra.NoRecords{}` today** (no M3/M4 writes exist yet): see P-5 |

## 3. Findings of this gate

No supported finding of severity medium or above (cso run: 0 supported findings; manual
review and e2e: none). Low / informational items, all deferred with an owner (§4):

| ID | Severity | Item |
|---|---|---|
| G-1 | Low (latent) | `cmd/api/taxpayers.go` wires `Records: infra.NoRecords{}`: `archiveTaxpayerRegime` cannot see M3/M4 records, so once incomes/expenses can be written a user could end a regime before its last record (integrity of estimates, not confidentiality). Inert today: no M3/M4 write path exists on `dev`. Becomes a blocker for enabling `module_incomes` / `module_expenses` (P-5) |
| G-2 | Low | No RFC key-rotation path: config takes one `RFC_ENC_KEY` + `RFC_ENC_KEY_ID` (no previous key), and there is no re-seal job. Rotating the key today would make every stored ciphertext unreadable (masks and hashes keep working, nothing user-facing breaks, but the ciphertext becomes dead data). Pepper rotation has no hash-version column and needs plaintext (re-seal job) |
| G-3 | Info | `RFCVault.Seal` stores `ActiveKeyID()` instead of the id framed in the ciphertext (`crypto.KeyIDOf`): correct while the keyring is immutable (#59 nit) |
| G-4 | Info | The device preference `active_taxpayer_id` is one key for all users of the device (spec §2.9 "per user"); cleared on sign-out and ignored when not in the signed-in user's list |

## 4. Carried M2 security follow-ups

Sources: `MODULE-taxpayers.md` §5 (S-1 … S-11), PLAN D-4, reviews on PRs #54, #55, #56, #57,
#58, #59, #60, #61, #68 (M2-11/M2-12), and the M1-17 deferred list items targeted at M2.

### Resolved

| Source | Item | Evidence |
|---|---|---|
| P0-30 D-4, S-1 | AAD = key id ‖ taxpayer id ‖ user id; id generated before encryption | `infra/rfcvault.go`, `usecase/taxpayers.go` `rowID`; e2e swap matrix (§1) |
| S-2 | Blind index scoped by user, own pepper | `crypto.NewRFCIndex`, `HashScoped`; e2e |
| #59 cond. 1, #60 | Startup refuses missing / equal / short RFC secrets, compared against all auth keys | `platform/config/taxpayers.go` `distinctSecrets`, `cmd/api/taxpayers_test.go`; e2e (§1) |
| S-3, S-4 | Masked only; RFC never in bodies, errors, logs | contract + handler tests (#60); e2e log scan |
| S-5, #57 | Cap under the user lock, archived counted, RR refused | `taxpayer_count_for_create()`; integration + e2e race |
| S-6, #57 | RLS FORCE, composite FKs, column grants, person type immutable, overlap trigger refuses RR | migration tests; e2e forged-id matrix |
| S-7, M1-17 (#43) | Deletion-cascade guard covers indirect chains and M2 tables | `migrations/account_deletion_integration_test.go` (#57); e2e |
| S-8, P0-30 D-3 | Repositories never Commit/Rollback/SET | `backend/internal/archtest/repository.go` (+ test) |
| S-9, #59 | `taxpayer.rfc_corrected` audited, ids only | `usecase/taxpayers.go`; e2e |
| S-10, #59 cond. 3, #60 | 30/h per user on all 8 mutations, fail closed | `adapter/http/limits.go`; e2e 429 |
| S-11, #56 | `additionalProperties: false` on every M2 body | contract 0.4.0; e2e `person_type` 400 |
| #60 IDOR | 404 parity on all 8 id ops | `TestContract_taxpayersForeignIsIndistinguishableFromMissing`; e2e |
| #59 matrix | Read-only refused on every write except archive/delete | table-driven use-case test; e2e |
| #54 FU1 | Check-digit mismatch surfaced as a warning, never echoing the RFC | contract `warnings` (D-RFC, ADR-0006) |
| #56 FU | `Regime.review_pending_with` for client-side warning | contract + handler (#60) |
| #57 note | Re-add of an archived regime must not resend an earlier `start_period` | `Store.ReopenRegime` is a plain UPDATE (`end_period = NULL`, options, ack); no start in the statement |
| #60 FU1 | Dependabot alerts in `tools/contract` | #62 |
| #58 | Read-only user no longer fills 4 steps before the 402 (UX, session memory) | `core/access/read_only_access.dart` |
| #61 | Active taxpayer reloads on resume; archived/deleted id falls back | `taxpayer_list.dart`, `resolveActiveTaxpayer` |

### Deferred, with owner

| Source | Item | Owner | Target |
|---|---|---|---|
| G-1 | Replace `infra.NoRecords{}` with the M3/M4 record lookup (latest income/expense period per taxpayer + regime) before any record write path is enabled | backend-sr | M3-04 / M4-05 (blocks P-5) |
| #59 cond. 2, G-2 | **RFC key-rotation re-seal job**: keyring with previous key ids (`RFC_ENC_KEY_PREVIOUS` + id), a worker job that opens with the old key id and re-seals with the active one per row (same AAD), resumable and idempotent; pepper rotation documented as re-hash from the opened plaintext with a hash-version column | backend-sr + db-sr | **M9** (with the other maintenance jobs); until then, do not rotate `RFC_ENC_KEY` (P-4) |
| #59 nit, G-3 | Store the key id parsed from the ciphertext frame (`crypto.KeyIDOf`) | backend-sr | with the M9 re-seal job |
| #59 nit | Idempotent create retry ignores a different body; note it in the contract description | architect | next contract touch |
| G-4 | Namespace `active_taxpayer_id` by user id (or keep and amend spec §2.9) | frontend-sr | M3-08 |
| #61 FU | Undo of archiving the active taxpayer should restore it as active; reload the list on a taxpayer-scoped 404 | frontend-sr | M3-08 |
| #58 non-blocking | `checkRate` names a missing bound as `0`/`999`; "guardado" in the review title | frontend-sr | next mobile touch |
| #68 (M2-12 D1–D3) | Cap shown only at the last onboarding step; read-only users get an Undo they cannot use; generic CPA-review copy | pm + frontend-sr | M2 UX polish backlog |
| #55, #56 | Regime-level `needs_cpa_review` exposure; `first_fiscal_year`; 625 per-option treatment | architect + backend-sr | M5-19 |
| OQ-6, OQ-7 | Compatibility matrix and option changes mid-year reviewed by a CPA | owner + CPA | before launch (`docs/reviews/CPA-REVIEW.md`) |
| M1-17 G-3, P0-30 D-6 | `Cache-Control: no-store` on error responses (error bodies carry no RFC) | backend-sr | next platform touch |
| #60 FU2 | Write budget is spent before contract validation (by design; client validates first) | — | accepted |

## 5. Preconditions to enable `module_taxpayers` in staging or production

All must hold; each is verified by the person flipping the flag and recorded in the deploy PR.

1. **P-1 `module_auth` preconditions** (M1-17 §5) hold in the same environment: M2 is unreachable without it.
2. **P-2 Keys from the secret manager.** `RFC_ENC_KEY` (32 random bytes) and `RFC_INDEX_PEPPER` (≥ 32 random bytes) are Secret Manager secrets, distinct from each other, from every `AUTH_*` key and per environment; never the local values; nothing in the image or the service YAML. Access to the pepper is restricted like the encryption key (a leaked pepper makes `rfc_hash` brute-forceable).
3. **P-3 `APP_ENV` is not `dev`**: only `dev` starts with ephemeral RFC keys; any other value refuses to boot without both secrets (verified in §1).
4. **P-4 No RFC key rotation until the M9 re-seal job ships**: `RFC_ENC_KEY_ID` stays at its first value; a suspected key leak before M9 is handled by a planned migration (re-seal with old + new key), not by swapping the secret.
5. **P-5 `module_incomes` / `module_expenses` stay OFF until `NoRecords` is replaced** (G-1). This does not block `module_taxpayers` itself: no record exists while those modules are off.
6. **P-6 Monitoring**: logs retained with the audit events of §2 (R); `rate limit unavailable` errors alerted (the limiter fails closed, so an outage turns every M2 write into 429).
7. **P-7 Gates**: M2-13 (tech-lead) PASS, this gate PASS, PR #68 (M2-11 flag tests, M2-12 QA) merged with green checks; REL-09 (deploy security gate) PASS.
8. **P-8 Mobile release build** passes `mobile-artifact-check` (no source maps, no internal URLs), HTTPS API base URL.

## 6. Notes

- The cso helper run is recorded as **partial**: its daily 600 s budget covered the snapshot
  inventory and a helper-mediated read of the core files; the rest of the review was manual
  and is listed in the Method row.
- The isolated stack used its own compose project, container name **and** volume name
  (`cc-gate-m214-pgdata`, verified with `docker compose config` before `up`); it was stopped
  and its volume removed afterwards. Only the API process started for this gate was stopped.
  Local flags were left OFF.
- No production code changed in this gate: there was no blocker to fix.
