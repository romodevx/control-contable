# Module M1: Auth & Account

| | |
|---|---|
| Flag | `module_auth` (OFF by default) |
| Spec | [MODULE-auth.md](../project-specs/MODULE-auth.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Auth`, `x-module: module_auth` |
| Gates | [M1-16 tech lead](../reviews/M1-16-tech-lead-gate.md) PASS, [M1-17 security](../reviews/M1-17-security-gate.md) PASS |
| Operations | [Auth runbook](../runbooks/auth.md) · [encryption key rotation](../runbooks/key-rotation.md) |
| Code | `backend/internal/auth/` (domain, usecase, infra, adapter/http), `mobile/lib/features/auth/`, `mobile/lib/core/session/` |

## What it does

A person signs up with an email address and accepts the current legal documents. The
server emails a 6-digit code. Whoever enters that code chooses the password, and the
account is created at that moment. The person then gets a session: a 15-minute access
token and a refresh token that rotates on every use. From there they can refresh, log out,
read and edit their profile, and delete the account from inside the app. Every other
module needs a signed-in user, so `module_auth` belongs to every demo preset.

The API reference for each operation is generated from the contract (`make api-docs`, see
[API reference](#api-reference)). This page covers how the operations fit together.

| Operation | Endpoint | Auth |
|---|---|---|
| `listLegalDocuments` | `GET /v1/legal/documents` | public |
| `register` | `POST /v1/auth/register` | public |
| `resendVerification` | `POST /v1/auth/verify-email/resend` | public |
| `verifyEmail` | `POST /v1/auth/verify-email` | public |
| `login` | `POST /v1/auth/login` | public |
| `refreshSession` | `POST /v1/auth/refresh` | public (refresh token in the body) |
| `logout` | `POST /v1/auth/logout` | bearer |
| `getMe` / `updateMe` | `GET` / `PATCH /v1/me` | bearer |
| `deleteMe` | `DELETE /v1/me` | bearer + current password |

## Flow

```mermaid
sequenceDiagram
    autonumber
    actor U as Person (app)
    participant API as API (module_auth)
    participant DB as PostgreSQL (RLS)
    participant W as Worker (auth_email_delivery)
    participant M as EmailSender

    U->>API: GET /v1/legal/documents
    API-->>U: 200 current versions (terms, privacy notice)
    U->>API: POST /v1/auth/register {email, accepted_documents}
    API->>DB: one tx: pending sign-up (no password, no name) + sealed email in auth_email_outbox
    API-->>U: 202 verification_sent (same body and same work for new, pending or verified email)
    W->>DB: claim due emails (SKIP LOCKED, leased)
    W->>M: 6-digit code, or the D3 notice (dev: worker log)
    W->>DB: delete the row (or schedule its retry)
    U->>API: POST /v1/auth/verify-email {email, code, password, name?}
    API->>DB: one tx: user + consents + trial (if email never had one), pending row consumed
    API-->>U: 201 AuthSession {user, tokens}
    U->>API: POST /v1/auth/login {email, password}
    API-->>U: 200 AuthSession (new token family)
    U->>API: POST /v1/auth/refresh {refresh_token R1}
    API->>DB: rotate R1 to R2 in the same family
    API-->>U: 200 TokenPair {access, R2}
    Note over U,API: Presenting R1 again revokes the whole family (D9)
    U->>API: POST /v1/auth/logout (Bearer) {refresh_token R2, device_id?}
    API->>DB: one tx: revoke the family, delete the caller's push device device_id (0.14.0)
    API-->>U: 204
    U->>API: DELETE /v1/me (Bearer) {password}
    API->>DB: one tx: user and everything it owns cascade; trial_registry HMAC stays
    API-->>U: 204 (wrong password: 403 invalid_password, nothing deleted)
```

### Try it locally

These requests were run against `make -C backend api` and `make -C backend worker` with
`module_auth` ON (`make -C backend flags ARGS="enable module_auth"`). The API only queues the
email; the worker delivers it within ~5 s, and in dev its email adapter writes the code to the
**worker** log, with the address redacted.

```sh
API=http://localhost:3000
EMAIL=ana.gomez@example.mx
PASSWORD=$(openssl rand -base64 18)   # a random 24-character password for this session
CREDENTIALS=$(jq -nc --arg email "$EMAIL" --arg password "$PASSWORD" '{email: $email, password: $password}')

# 1. Current legal documents: register must echo these versions.
DOCS=$(curl -s "$API/v1/legal/documents" \
  | jq -c '[.items[] | select(.kind == "privacy_notice_integral" or .kind == "terms") | {kind, version}]')

# 2. Register: 202 {"status":"verification_sent"}
curl -s -X POST "$API/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"accepted_documents\":$DOCS}"

# 3. Read the code from the worker log line "dev email: verification code" (field verification_code).
CODE=123456

# 4. Verify: 201 {"user":{...},"tokens":{...}}. The password is chosen here (D8).
curl -s -X POST "$API/v1/auth/verify-email" -H 'Content-Type: application/json' \
  -d "$(echo "$CREDENTIALS" | jq -c --arg code "$CODE" '. + {code: $code, name: "Ana Gómez"}')"

# 5. Login, keep both tokens.
SESSION=$(curl -s -X POST "$API/v1/auth/login" -H 'Content-Type: application/json' -d "$CREDENTIALS")
ACCESS=$(echo "$SESSION" | jq -r .tokens.access_token)
REFRESH=$(echo "$SESSION" | jq -r .tokens.refresh_token)

# 6. Profile: 200 {"id":...,"email":...,"name":...,"created_at":...}
curl -s "$API/v1/me" -H "Authorization: Bearer $ACCESS"

# 7. Refresh: 200 TokenPair. The old refresh token is now spent.
PAIR=$(curl -s -X POST "$API/v1/auth/refresh" -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
ACCESS=$(echo "$PAIR" | jq -r .access_token)
REFRESH=$(echo "$PAIR" | jq -r .refresh_token)

# 8. Logout: 204. Refreshing with that token afterwards answers 401.
curl -s -o /dev/null -w '%{http_code}\n' -X POST "$API/v1/auth/logout" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"

# 9. Delete the account: 204. A wrong password answers 403 invalid_password; nothing is deleted.
ACCESS=$(curl -s -X POST "$API/v1/auth/login" -H 'Content-Type: application/json' -d "$CREDENTIALS" \
  | jq -r .tokens.access_token)
curl -s -o /dev/null -w '%{http_code}\n' -X DELETE "$API/v1/me" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d "$(echo "$CREDENTIALS" | jq -c '{password}')"
```

A common or leaked password (a well-known passphrase, for example) fails step 4 with
`400 validation_failed`, `details[0].code = password_policy`. It does not spend a code try.

## Token and credential parameters

| Item | Value |
|---|---|
| Access token | JWT, EdDSA (Ed25519), `kid` header, 15 min (`access_expires_in: 900`). Verified against the current key and at most one previous, verify-only key ([rotation](../runbooks/auth.md#1-rotate-the-access-token-signing-key)) |
| Refresh token | `<user_uuid>.<256-bit secret>`, stored as SHA-256 only. 30-day sliding, 90-day absolute. Rotated on every use and grouped by family |
| Password | 12 to 128 characters, offline common/breached list (D5), argon2id (m = 19 MiB, t = 2, p = 1), rehashed on login when the parameters rise |
| Verification code | 6 digits from `crypto/rand`, stored as an HMAC, 15 min TTL. 5 wrong tries per code and 15 per email per 24 h across resends |
| Email | Trimmed, lower-cased and NFC-normalised. Looked up by a keyed blind index (`email_hash`) and stored encrypted |

## Flag behaviour (`module_auth`)

- OFF (the default): every operation in the table above answers `404 not_found` with the
  standard error envelope, before the body is parsed, before authentication and before any
  rate limit is spent. The module is indistinguishable from a route that does not exist. The
  app reads `GET /v1/flags` and hides the auth screens.
- ON: the operations behave as described here. A running API picks up a change within 5 s.
- The flag suite runs every operation with the flag ON and OFF
  (`backend/cmd/api/auth_flag_contract_integration_test.go`).
- `module_auth` is a member of `demo_core`, `demo_monetization` and `demo_full`.
- **Do not enable it in staging or production** until every precondition in
  [runbook §6](../runbooks/auth.md#6-enable-module_auth-in-a-shared-environment) holds.

```sh
make -C backend flags ARGS="enable module_auth"
make -C backend flags ARGS="disable module_auth"
```

## Rate limits

Source: [ADR-0005](../adr/0005-rate-limits-in-postgres.md) and [MODULE-auth §5](../project-specs/MODULE-auth.md#5-rate-limits-shared-postgresql-gcra-buckets-adr-0005-m1-10).
The values live in code: `authhttp.ClientLimits` (per client,
`backend/internal/auth/adapter/http/limits.go`) and `usecase.*Policy` (per account,
`backend/internal/auth/usecase/limits.go`). `TestClientLimits_matchTheSpec` keeps the
per-client values equal to the spec.

| Operation | Per client network (IPv4 address, IPv6 /64) | Per account key |
|---|---|---|
| `register` | 10 / h | 5 / h per email, shared with `resendVerification` |
| `verifyEmail` | 20 / 15 min | 5 wrong tries per code; 15 wrong per email per 24 h |
| `resendVerification` | 10 / h | 1 per 60 s (silent `202`); 5 / h per email, shared with `register` |
| `login` | 20 / 15 min | 5 failures / 15 min per (email, client network: IPv4 /24, IPv6 /64), and 50 failures / h per email from all networks. A success refills both |
| `refreshSession` | 60 / min | 20 rotations / h per token family (the token stays live) |
| `deleteMe` | none | 5 password attempts, then 1 per 3 min per user id |
| `listLegalDocuments` | 60 / min | none |

How the buckets work:

- A budget of N per window W admits N at once, then one more every W/N. For example, after 5
  login failures from one network, that network gets one more try every 3 minutes.
- Every `429` is `rate_limited` and carries `Retry-After` in whole seconds.
- Buckets are GCRA rows in `rate_limit_buckets`, shared by every API instance. Their keys are
  HMACs, so the table holds no email or IP address.
- A limiter error fails closed.
- The 429 is identical for known and unknown emails, so a limit never reveals whether an
  account exists.

## Errors

Every error uses the single envelope `ErrorResponse`:

```json
{"error":{"code":"validation_failed","message":"One or more fields are invalid.","request_id":"req_…",
  "details":[{"field":"password","code":"password_policy","message":"Use at least 12 characters that are not a common or leaked password."}]}}
```

`error.code` is the stable value to branch on. `message` is for people and never carries an
email address, code or token. `details[].field` is the dotted path inside the body (see
[CONTRACT.md](../api/CONTRACT.md)). Quote `request_id` when reporting a problem.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_auth -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `register` | `POST /v1/auth/register` | 400 | `validation_failed` | Input does not match the contract or the current legal document versions. |
| `register` | `POST /v1/auth/register` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `register` | `POST /v1/auth/register` | 429 | `rate_limited` | Rate limit exceeded. |
| `register` | `POST /v1/auth/register` | 500 | `internal_error` | Unexpected server error. |
| `verifyEmail` | `POST /v1/auth/verify-email` | 400 | `validation_failed` | Input does not match the contract or the password policy (checked before the code; not a try). |
| `verifyEmail` | `POST /v1/auth/verify-email` | 401 | `unauthorized` | The code is wrong, expired, consumed or unknown (one generic answer). |
| `verifyEmail` | `POST /v1/auth/verify-email` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `verifyEmail` | `POST /v1/auth/verify-email` | 429 | `rate_limited` | Rate limit exceeded. |
| `verifyEmail` | `POST /v1/auth/verify-email` | 500 | `internal_error` | Unexpected server error. |
| `resendVerification` | `POST /v1/auth/verify-email/resend` | 400 | `validation_failed` | Input does not match the contract. |
| `resendVerification` | `POST /v1/auth/verify-email/resend` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `resendVerification` | `POST /v1/auth/verify-email/resend` | 429 | `rate_limited` | Rate limit exceeded. |
| `resendVerification` | `POST /v1/auth/verify-email/resend` | 500 | `internal_error` | Unexpected server error. |
| `login` | `POST /v1/auth/login` | 400 | `validation_failed` | Input does not match the contract. |
| `login` | `POST /v1/auth/login` | 401 | `unauthorized` | Wrong credentials, unknown email or unverified sign-up (one generic answer). |
| `login` | `POST /v1/auth/login` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `login` | `POST /v1/auth/login` | 429 | `rate_limited` | Rate limit exceeded. |
| `login` | `POST /v1/auth/login` | 500 | `internal_error` | Unexpected server error. |
| `refreshSession` | `POST /v1/auth/refresh` | 400 | `validation_failed` | Input does not match the contract. |
| `refreshSession` | `POST /v1/auth/refresh` | 401 | `unauthorized` | Refresh token expired, revoked, reused, malformed or unknown (one generic answer). |
| `refreshSession` | `POST /v1/auth/refresh` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `refreshSession` | `POST /v1/auth/refresh` | 429 | `rate_limited` | Rate limit exceeded. |
| `refreshSession` | `POST /v1/auth/refresh` | 500 | `internal_error` | Unexpected server error. |
| `logout` | `POST /v1/auth/logout` | 400 | `validation_failed` | Input does not match the contract. |
| `logout` | `POST /v1/auth/logout` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `logout` | `POST /v1/auth/logout` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `logout` | `POST /v1/auth/logout` | 429 | `rate_limited` | Rate limit exceeded. |
| `logout` | `POST /v1/auth/logout` | 500 | `internal_error` | Unexpected server error. |
| `getMe` | `GET /v1/me` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getMe` | `GET /v1/me` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `getMe` | `GET /v1/me` | 429 | `rate_limited` | Rate limit exceeded. |
| `getMe` | `GET /v1/me` | 500 | `internal_error` | Unexpected server error. |
| `updateMe` | `PATCH /v1/me` | 400 | `validation_failed` | Input does not match the contract. |
| `updateMe` | `PATCH /v1/me` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `updateMe` | `PATCH /v1/me` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `updateMe` | `PATCH /v1/me` | 429 | `rate_limited` | Rate limit exceeded. |
| `updateMe` | `PATCH /v1/me` | 500 | `internal_error` | Unexpected server error. |
| `deleteMe` | `DELETE /v1/me` | 400 | `validation_failed` | Input does not match the contract. |
| `deleteMe` | `DELETE /v1/me` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `deleteMe` | `DELETE /v1/me` | 403 | `invalid_password` | Wrong current password; nothing was deleted. |
| `deleteMe` | `DELETE /v1/me` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `deleteMe` | `DELETE /v1/me` | 429 | `rate_limited` | Rate limit exceeded. |
| `deleteMe` | `DELETE /v1/me` | 500 | `internal_error` | Unexpected server error. |
| `listLegalDocuments` | `GET /v1/legal/documents` | 404 | `not_found` | Resource does not exist, belongs to another user, or its module is disabled. |
| `listLegalDocuments` | `GET /v1/legal/documents` | 429 | `rate_limited` | Rate limit exceeded. |
| `listLegalDocuments` | `GET /v1/legal/documents` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_auth -->

Some answers are deliberately the same for different causes, so they reveal nothing:

- `register` and `resendVerification` answer `202` for a new, pending or verified email.
- `verifyEmail` gives one `401` for a wrong, expired, used or unknown code.
- `login` gives one `401` for a wrong password, an unknown email or an unverified sign-up.

## Decisions

Full text in [MODULE-auth §9](../project-specs/MODULE-auth.md#9-decisions-decided-2026-10-13-owner-defaults-via-coordinator).

| # | Decision | Effect |
|---|---|---|
| D1 | Credential lookup under FORCE RLS without BYPASSRLS | Definer functions owned by the NOLOGIN role `cc_auth_definer`, each narrowed by policy to the one email hash it is handling |
| D2 | `POST /v1/auth/verify-email/resend` | Rate-limited, always the generic `202` |
| D3 | Duplicate register | The existing owner gets a "someone tried to sign up with your email" notice (at most hourly); the response stays `202` |
| D4 | Deletion needs the current password | `DELETE /v1/me` body `{password}` |
| D5 | Breached-password check | Offline embedded list; no third-party calls |
| D6 | Consents are deleted with the account | Only the `trial_registry` HMAC survives. Pending lawyer review (OB10) |
| D7 | A live pending sign-up is never overwritten | A second register behaves like a resend |
| D8 | The password is set at verification | Whoever proves inbox ownership picks it; register takes no password. Closes pre-registration takeover |
| D9 | Strict refresh reuse, no grace window | Any replay of a rotated token revokes the family. Cost: a lost refresh response signs the user out |
| D10 | Wrong password on `deleteMe` is `403 invalid_password` | A `401` would make the client refresh and resend the same password |
| D11 | Rate limits live in PostgreSQL | Limits hold across instances; see [ADR-0005](../adr/0005-rate-limits-in-postgres.md) |

## API reference

```sh
make api-docs                 # build/api-docs/index.html, rendered from docs/api/openapi.yaml
make api-error-tables         # regenerate the error table on this page after a contract change
```

## Known limits

Accepted risks R-1 to R-4 and the deferred items are listed in
[M1-17 §4](../reviews/M1-17-security-gate.md#4-carried-m1-security-follow-ups) and in the
[PLAN M1 follow-ups](../project-specs/PLAN-control-contable.md#m1-follow-ups). The ones a
user can notice:

- An access token stays valid for up to 15 minutes after logout. A deleted account gets
  `401` at once.
- A refresh whose response is lost on a flaky network signs the user out (D9).
- Verification emails go only to the API log until M1-06 (email provider, blocked on OB2).
