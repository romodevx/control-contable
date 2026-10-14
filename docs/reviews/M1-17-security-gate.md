# M1-17 · Security gate: M1 Auth & Account

| Field | Value |
|---|---|
| Scope | M1 as built on `dev` @ `origin/dev` (2026-10-13): `backend/internal/auth/**`, the auth migrations (`20261013130000_identity` … `20261013170000_rate_limits`, definer functions owned by `cc_auth_definer`), `router` authn and per-client limits, `ratelimitpg` (ADR-0005), `platform/crypto` (argon2id), `platform/config` (auth keys), mobile session storage (`mobile/lib/core/session`, Android backup rules) |
| Reviewer | security |
| Date | 2026-10-13 |
| Method | `/cso --scope auth` (gstack-cso 3.0.0, run `1791943587728-2e03de89fcef04bf`, daily static, finished **partial**: the helper's 600 s budget covers inventory and OSV only; gitleaks could not run under the worktree isolation guard), then a manual line-by-line review of the scope above, `govulncheck`, `go test ./internal/auth/...`, and an end-to-end probe of the running API against an **isolated local** postgres (compose project `cc-secgate`, own volume, port 5444; `module_auth` enabled there only, then disabled, stack and volume destroyed) |
| Verdict | **PASS.** No blocking finding in the code as built. `module_auth` stays **OFF** in every shared environment until the preconditions in §5 are met; they are deployment and legal conditions, not code defects |

## 1. Evidence

| Check | Result |
|---|---|
| `go test ./internal/auth/...` | green |
| `govulncheck ./...` | 0 called vulnerabilities (1 in a required module, not reachable) |
| OSV (cso) | `golang.org/x/crypto/openpgp` GO-2026-5932: package not imported (govulncheck agrees). `tools/contract` npm dev tool (faker, braces, lodash): dev-only Prism mock, never shipped; already tracked as P0-31 D-8 |
| Register enumeration (e2e) | new / pending / verified email: identical `202 {"status":"verification_sent"}` |
| Register timing (e2e, 15 samples, log sender) | new median 11.6 ms, verified 9.4 ms (p10–p90 ranges overlap). Exploitable only at ≤ 5 samples per email per hour and 10 per network per hour; see R-1 |
| Resend enumeration (e2e) | unknown / pending / verified: identical `202` |
| Verify enumeration (e2e) | unknown email / pending + wrong code / verified email: identical `401 "The code is invalid or has expired."` |
| Login enumeration (e2e, 20 samples) | one body for unknown email and wrong password: `401 "Email or password is incorrect."`; timing unknown median 29.3 ms vs known 28.1 ms (one argon2id run on both paths, dummy hash under current params) |
| Login limits (e2e) | 5 failures per (email, network) then `429` with `Retry-After` (179 s); the correct password is also refused while limited (no oracle); rotating `X-Forwarded-For` does not mint new networks (`TRUSTED_PROXY_HOPS=0` locally; rightmost-trusted-hop parsing otherwise) |
| Refresh reuse ⇒ family revoke (D9, e2e) | rotate R1→R2 `200` + `Cache-Control: no-store`; replay R1 `401`; R2 afterwards `401` (family revoked). `RevokeFamily` loops until 0 rows, so a successor committed concurrently is revoked too |
| Account deletion (e2e) | wrong password `403 invalid_password` (not 401: no client sign-out loop); right password `204`; `users`, `refresh_tokens`, `consents` rows for the id: 0/0/0; login, refresh and `GET /v1/me` with the old access token all `401`. Integration tests assert every FK to `users` cascades and every `user_id` column has one |
| Logs (e2e, 178 lines) | no email, password, refresh or access token. `to_email` is `[REDACTED]` even in dev; the verification code is logged **only** when `APP_ENV=dev` (`LogSender`), dropped with a warning otherwise |
| argon2id | m = 19 MiB, t = 2, p = 1, 16-byte salt, 32-byte key: exactly the OWASP Password Storage Cheat Sheet minimum. Rehash-on-login exists (`NeedsRehash`), PHC parsing bounds memory × time (≤ 256 MiB·iter) so a planted hash cannot DoS verify; 1024-byte password cap before NFKC and after |
| JWT | EdDSA only (`WithValidMethods`), `iss`/`aud`/`exp`/`nbf`/`iat` required, 30 s leeway, lifetime ≤ 15 min enforced on verify, canonical-UUID `sub`, `jti` required, `kid` allow-list (current + one previous verify-only key), 4 KiB cap, single `Authorization` header only |
| Keys | `AUTH_EMAIL_PEPPER`, `AUTH_EMAIL_KEY`, `AUTH_CODE_KEY`, `AUTH_ACCESS_TOKEN_KEY` loaded through the `SecretSource` port, size-checked, boot fails closed; signer and hashers redact in `fmt`/`slog`. Only `EnvSecrets` exists today (see P-4) |
| Android secure storage | `flutter_secure_storage` 11.2.0 (Keystore-wrapped); access token in memory only; `allowBackup=false` plus `dataExtractionRules`/`fullBackupContent` excluding shared prefs (cloud and device transfer); iOS `first_unlock_this_device` |
| Verification code | 6 digits, unbiased draw from `crypto/rand`, stored as HMAC(code key, email index ‖ code); 5 wrong per code and **15 wrong per email per 24 h across resends** (`email_verification_failures`), single use, 15 min TTL |

## 2. Threat model (STRIDE), M1 as built

Actors: anonymous internet caller (single network or a botnet); holder of a stolen refresh
token; authenticated user; operator with a runtime DB credential (`cc_app`); someone with
the phone (unlocked, or a backup of it).
Assets: account takeover, refresh-token families, password hashes, email addresses
(blind-indexed + encrypted), the HMAC/peppers and the Ed25519 signing key, the trial
entitlement.

| | Threat | Control as built | Residual |
|---|---|---|---|
| **S** | Credential stuffing / online guessing | argon2id; 5 failures per (email, /24 or /64) per 15 min; 50 per email per hour from all networks; 20 logins per network per 15 min; edge limiter | R-2 (distributed ceiling hold, accepted ADR-0005) |
| **S** | Pre-registration takeover (claim a victim's email first) | No password at register; password chosen at verify by whoever holds the code (D8) | — |
| **S** | Brute-force the 6-digit code | 5 per code, 15 per email per 24 h across resends, 20 verifies per network per 15 min | 15/10⁶ per day per email |
| **S** | Forged access token / alg confusion | EdDSA only, `kid` allow-list, claims strictly validated | — |
| **T** | Tamper with another user's rows via the API | RLS FORCE on user tables, session `app.current_user_id` per transaction; definer functions narrowed to one bound email hash by policy and column grants; `cc_auth_definer` NOLOGIN NOBYPASSRLS, `search_path` pinned | — |
| **R** | Deny a security event | audit events `login_failed`, `login_rate_limited`, `token_reuse` (warn), `account_created`, `account_deleted`, `delete_attempts_exhausted`, `logout`, user id only | Logs are stdout only; retention/alerting on `token_reuse` is a deploy concern (P-7) |
| **I** | Account enumeration (body) | register/resend `202` always; verify one `401`; login one `401`; delete wrong password `403` only for the authenticated owner | — |
| **I** | Account enumeration (timing) | login: one argon2id on every path; verify hashes the password only after a matching code; register no password hashing | R-1: register/resend differ by DB work; **becomes large once a real email provider is called synchronously** (P-1) |
| **I** | PII/secrets in logs | slog redaction of `Secret`/`LogValuer`, `to_email` redacted, codes only in dev | P-3 (`APP_ENV` must not be `dev` in shared envs) |
| **I** | Token theft from device | refresh in Keystore/Keychain, never backed up; access token in memory | Rooted device: out of scope |
| **D** | Lock a victim out | login budget keyed per network; register/resend 5 per email per hour shared | R-3 (sign-up griefing, accepted #34/#40/#47) |
| **D** | CPU exhaustion via argon2 | per-network login limit before lookup; malformed input still spends the dummy hash but is per-network limited; argon2 params bounded at parse | — |
| **E** | Stolen refresh token keeps access | rotation, reuse revokes the whole family (strict, no grace), absolute expiry inherited, account status re-checked on refresh | R-4 (access token valid ≤ 15 min after logout/suspension/deletion; deleted users get 401 because the account is gone) |
| **E** | API compromise escalates to auth tables | `cc_app` cannot `UPDATE email_verifications` directly or call `email_verification_upsert`; only definer functions with argument checks | — |

## 3. Findings of this gate

No supported finding of severity medium or above. Low/informational items, all deferred
with an owner (§4):

| ID | Severity | Item |
|---|---|---|
| G-1 | Low | `rate_limit_take` returns a wait ≤ 0 when a concurrent `rate_limit_reset` lands between its `UPDATE` and `SELECT`; `ratelimitpg` treats ≤ 0 as admitted, so one event is free. A reset only follows a successful login (the attacker already has the password), so the gain is one extra guess per victim sign-in. Fix: `greatest(..., interval '1 microsecond')` (carried PR #47 FU1) |
| G-2 | Low | Register latency differs ~2 ms between a new and a verified email with the log sender (R-1). With ≤ 5 samples per email per hour it is not a practical oracle today; with a synchronous provider it would be (P-1) |
| G-3 | Info | Error responses do not carry `Cache-Control: no-store` (token and account responses do). Error bodies hold no secret (carried P0-30 D-6) |

## 4. Carried M1 security follow-ups

Sources: `docs/project-specs/MODULE-auth.md` §5–§9, PLAN, `docs/reviews/`, ADR-0005, gate
reviews on PRs #33, #34, #40, #41, #43, #44, #47.

### Resolved

| Source | Item | Evidence |
|---|---|---|
| #34 H2, #40 D8 | Pre-registration takeover | password set at verify; `TestAuth_preRegistrationTakeover` |
| #34 H1 | Code brute force across resends | `email_verification_failures`, 15 / email / 24 h (`20261013150000_auth_service.sql`) |
| #33 C3, #41 #2 | Refresh re-checks account status | `usecase/refresh.go` `CanSignIn` |
| #33 C2/C4/C5, #44 | Reuse revokes the whole family incl. a concurrent successor; logout revokes, never deletes (evidence kept); successors inherit absolute expiry | `infra/refreshtokens.go` loop, `TestRefresh_reuseRevokesASuccessorCommittedWhileItWaited`; e2e above |
| #41 | 401 carries `WWW-Authenticate` | `authhttp.Authenticator`, contract header |
| #40, #41 #1, #43 | Per-account limits on login and deleteMe; per-operation per-client limits | `usecase/limits.go`, `authhttp.ClientLimits`; e2e above |
| #43 | In-process limiter gave N instances N× attempts | Postgres GCRA buckets (`20261013170000_rate_limits.sql`, ADR-0005) |
| ADR-0005 | Targeted lockout via email-only budget | budget keyed by (email, network) + per-email ceiling; e2e above |
| #43 | Mobile maps `403 invalid_password` | `mobile/lib/core/network/auth_interceptor.dart`, `delete_account.dart` |
| P0-31 F-03 | Android backups could carry tokens | `AndroidManifest.xml` + `res/xml` rules, `android_backup_test.dart` |
| Spec §6.8/§6.9 | Tokens in secure storage; audit events | `refresh_token_store.dart`; events listed in §2 |
| ADR-0005 | Security side of the decision | **Accepted by security** at this gate (status flips to ACCEPTED once M1-16 also accepts) |

### Deferred, with owner

| Source | Item | Owner | Target |
|---|---|---|---|
| #40, G-2 | **Async email (outbox)** so provider latency and failures never reach the response | backend-sr | M1-06 (blocks P-1) |
| #34 | **`cc_auth_definer` role** in every shared environment (NOLOGIN NOBYPASSRLS; `cc_owner` member WITH INHERIT FALSE, SET TRUE; PG ≥ 16) before the M1 migrations run | devops-sr | REL-09 (blocks P-2) |
| #34, OB10 | Legal documents are placeholders (`pending_legal_review`, `.invalid` URLs) | owner + legal | before P-5 |
| #40 low, OB10 | Consents are those of the first registrant, not re-affirmed at verify | backend-sr + legal | with OB10 |
| #47 FU1, G-1 | `rate_limit_take` wait ≤ 0 under reset race | db-sr | M9 (with the purge job) |
| #47 FU2 | `rate_limit_take` returns NULL if the purge deletes the row mid-call (fails closed, 500) | db-sr | M9 |
| #33, #44 | Purge of dead refresh-token families + `(absolute_expires_at)` index | db-sr | M9 |
| #34 | Purge of `email_verification_failures` and `rate_limit_buckets` past their windows | db-sr | M9 |
| #41 | Suspension does not revoke refresh families; no operator suspend path (definer function) | db-sr + backend-sr | before any suspend feature ships |
| #41 #3 | Router authenticates every secured operation with the user bearer scheme regardless of declared scheme (webhook fails closed today) | backend-sr | before M7-05 |
| P0-30 D-6, G-3 | `Cache-Control: no-store` on error responses | backend-sr | next platform touch |
| P0-30 D-3 | Lint/archtest forbidding repositories from Commit/Rollback/SET | tech-lead | M2 |
| P0-31 F-08/D-8 | pub dependency scan (osv-scanner) in CI; Prism dev-tool advisories | devops-sr | REL-09 |
| #43 | Deletion-cascade guard misses indirect chains | db-sr | when M2 tables land |
| — | argon2id at the OWASP minimum: raise memory (e.g. 46 MiB, t = 1) once the Cloud Run instance size is fixed; rehash-on-login migrates hashes | backend-sr | REL-09 sizing |

### Accepted risks (confirmed at this gate)

- **R-1** register/resend latency differs by a few ms of DB work (log sender). Re-measure with the real provider behind the outbox.
- **R-2** an attacker on ≥ 10 networks can hold a victim at the 50/h per-email ceiling; users behind one CGNAT /24 share a budget (ADR-0005, D11).
- **R-3** 15 wrong codes lock a victim's sign-up for 24 h; re-registering every 60 s rotates a victim's pending code; a stranger can spend the 5/h register/resend budget (#34, #40, #47). Griefing, not takeover.
- **R-4** strict reuse policy: a lost refresh response followed by a retry signs the user out (D9); the mobile client serialises refreshes. Access tokens live ≤ 15 min after logout/suspension.
- Plus-addressing allows trial farming (product decision, owner). Corrupt/legacy hashes answer faster at login (#41).

## 5. Preconditions to enable `module_auth` in staging or production

All must hold; each is verified by the person flipping the flag and recorded in the deploy PR.

1. **P-1 Real email provider, asynchronous.** M1-06 adapter behind an outbox: the request path never waits on the provider and never changes its answer on provider failure. Re-run the register/resend timing probe of §1 against staging; medians within noise.
2. **P-2 `cc_auth_definer` exists** in the target database (NOLOGIN, NOSUPERUSER, NOBYPASSRLS, `cc_owner` member WITH INHERIT FALSE, SET TRUE, PG ≥ 16) **before** the M1 migrations run; the migration's own guard refuses otherwise. `check-runtime-role.sql` passes for `cc_app`.
3. **P-3 `APP_ENV` is not `dev`** (otherwise verification codes are logged), and `TRUSTED_PROXY_HOPS` equals the real number of proxies appending to `X-Forwarded-For` (otherwise every caller shares the load balancer's budget, or callers mint networks).
4. **P-4 Keys from the secret manager.** `AUTH_EMAIL_PEPPER`, `AUTH_EMAIL_KEY`, `AUTH_CODE_KEY`, `AUTH_ACCESS_TOKEN_KEY` (and `DATABASE_URL`) are Secret Manager secrets injected at runtime, distinct per environment, never the local values; nothing in the image or the service YAML in plain text. Rotation runbook for the access-token key (previous key verify-only) exercised once in staging.
5. **P-5 Legal documents reviewed by a lawyer**: real URLs and versions published in `legal_documents` (aviso de privacidad integral/simplificado, términos) replacing the `.invalid` placeholders.
6. **P-6 Gates**: M1-16 (tech-lead) PASS, this gate PASS, REL-09 (deploy security gate) PASS; ADR-0005 ACCEPTED.
7. **P-7 Monitoring**: logs retained in the platform's log store with an alert on `event=token_reuse` and on a spike of `login_rate_limited`.
8. **P-8 Mobile release build**: release APK/AAB/IPA pass `mobile-artifact-check` (no source maps, no internal URLs), API base URL is HTTPS.

## 6. Notes

- The cso helper run is recorded as **partial**: its daily budget is 600 s, which covers the
  snapshot inventory and the OSV scan; the substantive review above was manual.
- Local verification used an isolated compose project with its own volume. A first attempt
  reused the compose file's fixed volume name (`control-contable-pgdata`), briefly starting a
  second postgres on the shared local dev volume before it was stopped; no migration or write
  succeeded through it. If the shared local DB misbehaves, `make db-reset` restores it. For
  future isolated runs, override both `container_name` and `volumes.pgdata.name`.
