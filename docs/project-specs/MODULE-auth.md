# MODULE-auth — M1 Auth & Account (`module_auth`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-13 · MoSCoW: **Must** · Depends on: P0
Sources: PRD §6 (US-A1..A4), §7, §8 · MODULES.md M1 · PLAN M1-01..M1-18 · ADR-0001 (O6, O9, OWNER 1 = A) ·
reviews P0-29 (FU1, FU2, FU4, "What M1 must respect") and P0-30 (D-2, D-3, D-5).

## 1. Scope

A person creates an account with email + password after accepting the current legal
documents, proves the email with a 6-digit code, gets a session (access + rotating refresh),
keeps it alive, logs out, reads/edits their profile, and deletes the account in-app.

| Op | Endpoint | Auth | Contract |
|---|---|---|---|
| register | `POST /v1/auth/register` | public | exists |
| verifyEmail | `POST /v1/auth/verify-email` | public | exists |
| resendVerification | `POST /v1/auth/verify-email/resend` | public | **new — add in M1-01** |
| login | `POST /v1/auth/login` | public | exists |
| refreshSession | `POST /v1/auth/refresh` | public | exists |
| logout | `POST /v1/auth/logout` | bearer | exists |
| getMe / updateMe | `GET / PATCH /v1/me` | bearer | exists |
| deleteMe | `DELETE /v1/me` | bearer + password re-auth | exists; **body `{password}` added in M1-01** (D4) |
| listLegalDocuments | `GET /v1/legal/documents` | public | exists |

Owner defaults applied (2026-10-13): verification by **6-digit emailed code**; provider
undecided → `EmailSender` port + **dev log adapter** now (M1-06 blocked on OB2); **generic
response on duplicate email**; **trial starts at successful verification** (server clock);
**in-app deletion**; LFPDPPP consent stored with **document kind + version + timestamp**.

## 2. Token and credential parameters (from ADR-0001 O6)

| Item | Value |
|---|---|
| Access token | JWT EdDSA (Ed25519), `kid` header, 15 min (`access_expires_in: 900`); validate with `WithValidMethods(["EdDSA"])`, issuer, audience, `WithExpirationRequired` |
| Refresh token | `<user_uuid>.<256-bit secret>`, stored as SHA-256, 30-day sliding, 90-day absolute, rotated on every use, grouped by `family_id` |
| Password | 12–128 chars, breached-password check, argon2id in PHC format (OWASP params, raisable; rehash on login if params rose) |
| Verification code | 6 digits from `crypto/rand`; stored as **HMAC-SHA256(server key, email_hash ‖ code)**, never plain; TTL 15 min; max 5 wrong attempts per code, then the code is dead |
| Email | trimmed, lower-cased, **NFC-normalised before hashing** (D-5); `email_hash` = keyed HMAC, used for lookups, `trial_registry` and rate-limit keys |

## 3. User stories and acceptance criteria

### US-A1a Register (start)
- **Given** the current legal documents, **when** I POST email, password, optional name and
  `accepted_documents` covering the current `privacy_notice_integral` (or simplified) and
  `terms` versions, **then** I get `202 {status: verification_sent}` and a code is "sent"
  (dev: written to the log adapter, redacted in every other log line).
- **Given** the email already belongs to a verified account, **when** I register, **then** I
  get the **same 202 body, status and comparable latency** (argon2 still runs); no code is
  created; the existing owner gets a "someone tried to sign up with your email" notice
  through `EmailSender`, at most once per hour (D3).
- **Given** a pending sign-up for the same email, **when** I register again, **then** the
  pending row is replaced (new password hash, new code, attempts reset) — upsert, 202.
- **Given** a weak/breached password or a short one, **then** `400 validation_failed` with
  `FieldError{field:"/password"}` and a generic message (never echo the value).
- **Given** an accepted document version that is not the current one, or a missing kind,
  **then** `400 validation_failed` on `/accepted_documents`.
- Nothing is written to `users`, `consents` or `entitlements` at this step.

### US-A1b Verify email (creates the account)
- **Given** a live pending sign-up and the right code, **when** I POST `{email, code}`,
  **then** in one transaction: the user row is created, consents (kind, version, accepted_at)
  are stored, the pending row is consumed, `trial_registry_check_and_insert` runs, the
  trial entitlement starts (server clock) **only if the email hash was not in the
  registry**, and I get `201 AuthSession`.
- **Given** the email hash is already in `trial_registry` (deleted and came back), **then**
  the account is created with **no trial** (read-only until a plan) — same 201.
- **Given** a wrong, expired, consumed or unknown code (or unknown email), **then**
  `401 unauthorized` with one generic message; each wrong try counts against the code.
- **Given** 5 wrong tries, **then** the code is dead and further tries answer `401` until
  a resend.

### US-A1c Resend code
- **Given** a pending sign-up, **when** I POST `{email}`, **then** a new code replaces the
  old one (attempts reset) and I get `202 {status: verification_sent}`.
- **Given** no pending sign-up or an already verified email, **then** the same 202, nothing sent.
- **Given** a resend less than 60 s after the last one, **then** the same 202, nothing sent
  (cooldown is silent, so it is not an oracle); the client shows a 60 s countdown.

### US-A2a Login
- **Given** valid credentials of a verified account, **then** `200 AuthSession` with a new
  token family.
- **Given** a wrong password, an unknown email, or a pending (unverified) sign-up, **then**
  `401 unauthorized`, identical body and comparable latency (dummy argon2 on unknown).
- **Given** too many failures (see §5), **then** `429 rate_limited` with `Retry-After`.

### US-A2b Refresh
- **Given** a live refresh token, **then** `200 TokenPair`; the old token is marked rotated.
- **Given** an already-rotated token (reuse), **then** the **whole family is revoked** and
  `401`. The user context is set from the token's UUID prefix before lookup (no RLS bypass).
- **Given** an expired (sliding or absolute), revoked, malformed or foreign token, **then** `401`.
- Concurrent refresh with the same token: exactly one wins; the other is treated as reuse
  (test with `testing/synctest`). Mobile serialises refresh in a `QueuedInterceptor`.

### US-A2c Logout
- **Given** a bearer access token and my refresh token, **then** its family is revoked, `204`.
- **Given** a refresh token that is not mine or already revoked, **then** `204` (idempotent;
  never reveal another user's token state). Missing/invalid bearer → `401`.
- Mobile wipes tokens from `flutter_secure_storage` even if the call fails.

### US-A3 Profile
- `GET /v1/me` → `200 User` (id, email, name, created_at). No bearer / bad token → `401`.
- `PATCH /v1/me {name}` → `200 User`. Empty body or name > 120 → `400 validation_failed`.
  Email change is out of scope.

### US-A4 Delete account
- **Given** I confirm in the app (irreversible → confirmation dialog, per house UI rules)
  and re-enter my current password, **when** `DELETE /v1/me {password}`, **then** in one transaction the user and everything it owns
  cascade (taxpayers, records, consents, refresh tokens, usage, entitlements), and `204`.
  Consents are deleted too (data minimization, D6); only the `trial_registry` HMAC survives. Further use of the old access token: `401` on any
  user-scoped call (user no longer exists), so the 15-min JWT lag is harmless.
- **Given** a wrong or missing password, **then** `401 unauthorized` (counts toward the login
  failure limit) and nothing is deleted.
- The app then shows how to cancel the store subscription (it does not cancel it) and
  returns to the welcome screen with storage wiped.

### US-A5 Legal documents
- `GET /v1/legal/documents` → `200` with the current version of each kind (url, version,
  published_at). The register screen links to them and sends back the versions shown.
- Placeholder content until the lawyer delivers (OB10); versions are data, not code.

## 4. Error mapping (`ErrorCode`)

| Situation | HTTP | ErrorCode |
|---|---|---|
| Schema / password policy / stale document version | 400 | `validation_failed` |
| Bad code, bad credentials, bad/reused/expired refresh, missing bearer | 401 | `unauthorized` |
| Duplicate email on register / resend for unknown email | 202 | — (generic success) |
| Flag OFF, unknown route | 404 | `not_found` |
| Rate limit hit | 429 | `rate_limited` (+ `Retry-After`) |
| Email adapter down on register/resend | 202 | — (logged, retried by resend; never leak) |
| DB deadline exceeded / unexpected | 500 | `internal_error` |

Domain sentinels (`ErrInvalidCredentials`, `ErrCodeInvalid`, `ErrTokenReused`, …) map through
one `apperr.Mapper` in the auth adapter. `apperr.Error.Message` never carries email, code or token.

## 5. Rate limits (in-process token buckets, ADR-0001 O9)

| Op | Per IP | Per account key (`email_hash` / user) |
|---|---|---|
| register | 10 / h | 5 / h |
| verify-email | 20 / 15 min | 5 wrong tries per code (hard) |
| resend | 10 / h | 1 / 60 s, 5 / h (silent) |
| login | 20 / 15 min | 5 failures / 15 min → 429 |
| refresh | 60 / min | — |
| legal/documents | 60 / min | — |

Numbers are starting values; security gate (M1-17) may tune them. Limits sit after the
flag check (OFF answers 404 first).

## 6. Security requirements

1. **Per-request deadline** (P0-29 FU1): middleware ≈10 s, below `WriteTimeout` 30 s; test with
   a hanging `tx.Manager` fake. Lands before the first DB-backed handler.
2. **Replace `notBuilt`** (P0-29 FU2): explicit stub answering 404 like a hidden module; no
   nil `StrictServerInterface`, no reliance on `Recover`.
3. **Credential lookup under FORCE RLS** (D1): `auth_lookup_credentials` owned by `cc_owner`
   sees no `users` rows under FORCE RLS. Fix: a **NOLOGIN definer role** owns the lookup
   functions and gets a **narrow RLS policy** granting SELECT on only the needed columns/rows
   (`id`, `password_hash`, by `email_hash`). **No BYPASSRLS role**, no unforced table.
   Final design owned by db-sr in **M1-03**.
4. `trial_registry` and `email_verifications`: service-only, reached only through narrow
   definer functions; `code_hash` = HMAC with a server key from Secret Manager (never in repo).
5. `tx.SystemManager` only for sign-up/login paths; archtest forbids it elsewhere in
   `adapter/http` (D-2). Repositories never commit/rollback or run `SET` (D-3).
6. `refresh_tokens`, `consents`: ENABLE + FORCE RLS, `pgtest.AssertIsolation` each.
7. No PII/secrets in logs: email, code, password, tokens redacted; dev log adapter prints the
   code only when `CC_ENV=local`.
8. Mobile: tokens only in **`flutter_secure_storage`**, never `shared_preferences`; no
   tokens in crash reports or analytics.
9. Audit log events (no PII): register_started, account_created, login_failed, token_reuse,
   logout, account_deleted.
10. Security gate checks: enumeration (body + timing), token reuse, argon2 params, rate limits.

## 7. Flag behaviour

- `module_auth` defaults **OFF**. OFF ⇒ every op in §1 answers **404** before binding,
  decode or authn (uniform with other hidden modules); the Flutter app hides auth screens.
- Flag-aware suite runs each op ON and OFF (M1-10).
- **Do not enable `module_auth` in any shared environment** until `trial_registry`
  (M1-03) exists and both M1 gates (M1-16, M1-17) pass, plus REL-09.
- Presets: member of every demo preset (all other modules need a user).

## 8. Out of scope

Social login, 2FA, SMS reset, email change, password reset by email (Should — separate
feature), session list/device management, ARCO workflow beyond deletion and profile
(documented email channel covers the rest), RevenueCat subscriber deletion (M7), real
email provider (M1-06, blocked on OB2), legal text content (OB10).

## 9. Decisions (DECIDED 2026-10-13, owner defaults via coordinator)

| # | Decision |
|---|---|
| D1 | Credential lookup under FORCE RLS: **no BYPASSRLS role**. A NOLOGIN definer role owns the lookup functions, with a narrow RLS policy granting SELECT on only the needed columns/rows. db-sr owns the final design in M1-03. |
| D2 | Add `POST /v1/auth/verify-email/resend {email}`: rate-limited (§5), always the generic 202. Contract change in M1-01. |
| D3 | Duplicate register: send the existing owner a "someone tried to sign up with your email" notice; the response stays the generic 202. |
| D4 | Account deletion requires **current password re-auth** (`DELETE /v1/me` body `{password}`; contract change in M1-01). |
| D5 | Breached/common password check uses an **offline embedded list**; no third-party calls. |
| D6 | Consents are **deleted with the account** (data minimization); only the `trial_registry` HMAC survives. **Flagged for lawyer review** (OB10). |

No open questions remain for stage 1.
