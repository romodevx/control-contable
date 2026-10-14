# Runbook: Auth & Account (`module_auth`)

Operations for module M1 ([module page](../modules/auth.md),
[spec](../project-specs/MODULE-auth.md)). Every command in a `sh` block was run on the local
stack on 2026-10-13. Procedures for shared environments (staging, production) say so
explicitly. Read the whole procedure before you start one.

Ground rules:

- **No PII in tickets, chats or shell history.** Work with the user id (a UUID) and the
  `request_id` of a response. Never paste an email address, password, code or token anywhere.
- **RLS applies to operators too.** The user tables are `FORCE ROW LEVEL SECURITY` with
  policies only for `cc_app`. Even `cc_owner` sees no rows, and no role has BYPASSRLS. Per-user
  SQL runs as `cc_app` inside a transaction that sets `app.current_user_id`, exactly as the
  API does.
- Locally, `make -C backend db-psql-app` runs SQL from stdin as `cc_app`, and only against the
  local stack. Shared environments run the same SQL as `cc_app` through their own
  connection (Cloud SQL Auth Proxy, credentials from Secret Manager).

## 1. Rotate the access-token signing key

Access tokens are EdDSA JWTs whose `kid` header names the signing key. The API signs with the
current key and verifies with the current key plus at most one **previous, verify-only** key.
Access tokens live 15 minutes, so a rotation never signs anyone out.

### Local

```sh
make -C backend access-key-rotate   # rotated the access-token key: active kid k20261014…, previous kid k1
# restart make -C backend api
```

After the restart, new tokens carry the new `kid`, and tokens signed before the rotation
still answer `200` until they expire. A second rotation drops the oldest key.

### Shared environment

The four variables come from Secret Manager (keys) and the service config (key ids):
`AUTH_ACCESS_TOKEN_KEY`, `AUTH_ACCESS_TOKEN_KEY_ID`, `AUTH_ACCESS_TOKEN_PREVIOUS_KEY` and
`AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID`. A previous key and its id are set together or not at all;
boot fails closed otherwise.

1. Generate the new seed (32 random bytes, hex) straight into a new secret version. Never
   print it or put it in a file:
   `openssl rand -hex 32 | <add as a new version of the environment's access-key secret>`.
2. Choose the new kid. The convention is `k` + UTC timestamp, for example `k20261014022645`.
3. Deploy one revision that has:
   - `AUTH_ACCESS_TOKEN_PREVIOUS_KEY` = the old key's secret version, and
     `AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID` = the old kid;
   - `AUTH_ACCESS_TOKEN_KEY` = the new version, and `AUTH_ACCESS_TOKEN_KEY_ID` = the new kid.
4. Check it. A fresh login returns a token whose header carries the new kid, and a token
   issued before the deploy still gets `200` on `GET /v1/me`.
5. Wait at least 16 minutes: the 15-minute lifetime plus the 30 s leeway, plus the time for
   the rollout to finish. Then deploy again without the two `PREVIOUS` variables, and disable
   the old secret version.

**Emergency, the key leaked:** skip the overlap. Deploy the new key with **no** previous key,
so every outstanding access token fails with `401` and clients refresh. Refresh tokens are
not JWTs, so a leaked signing key cannot forge them. Record the incident in
`lessons-learned/`.

## 2. Revoke a user's sessions

Revoking a user's sessions means revoking every refresh-token family they hold. Their
access tokens keep working for up to 15 minutes (accepted risk R-4). After that the app must
log in again. You need the user id.

```sh
make -C backend db-psql-app PSQL_ARGS="-v user_id=01a1383a-5c0e-7b42-b311-22e2277eee8f" <<'SQL'
BEGIN;
SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = 'admin' WHERE revoked_at IS NULL;
COMMIT;
SQL
```

Expected output: `context_set = t`, then `UPDATE <n>` (n ≥ 0), then `COMMIT`. If you get
`UPDATE 0` for a user you know is signed in, the id is wrong: the policy hides other users'
rows, so a typo fails quietly. Check the id with the status query in §4.

- A refresh with any revoked token answers `401`, and the app signs out.
- Revocations are history. A trigger refuses to change or clear `revoked_at` or
  `revoke_reason` once set, so this cannot be undone; the user logs in again.
- There is no suspend path yet (`users.status` is not writable by `cc_app`; M1-17
  follow-up "Suspension"). To lock someone out for good, revoke now and escalate.

## 3. Unlock a rate-limited account

Limits are GCRA buckets in `rate_limit_buckets` ([ADR-0005](../adr/0005-rate-limits-in-postgres.md),
[limits table](../modules/auth.md#rate-limits)). Bucket keys are HMACs derived from a server
secret, so you **cannot find one person's bucket** from the database, and that is
deliberate. Try the options in this order:

1. **Wait.** Every `429` carries `Retry-After`, and the app shows a countdown. The login
   failure budget refills one try every 3 minutes per network. The per-email ceiling refills
   one every 72 s.
2. **Another network.** The login failure budget is per (email, network), so mobile data
   instead of the office Wi-Fi gets a fresh budget, unless the per-email ceiling (50 / h) is
   spent. A successful login refills both budgets.
3. **Reset a whole scope (emergency).** Only with the security owner's approval, and noted in
   the incident. This refills the budget of **every** key in that scope, attackers included,
   so it only makes sense during a support incident that affects many users (for example a
   CGNAT range locked out):

   ```sh
   make -C backend db-psql-app <<'SQL'
   UPDATE rate_limit_buckets SET tat = '-infinity' WHERE scope = 'auth_login_email_client' AND tat > now();
   SQL
   ```

   Expected output: `UPDATE <n>`. The user's next login is admitted.

| Scope | Limit |
|---|---|
| `auth_login_email_client` | login failures per (email, network) |
| `auth_login_email` | login failure ceiling per email |
| `auth_signup_email` | register + resend per email |
| `auth_refresh_family` | rotations per refresh family |
| `auth_delete_password` | `deleteMe` password attempts per user |
| `client_<operation>` | per-client limits (`client_login`, `client_register`, …) |

The 15 wrong codes per email per 24 h on `verifyEmail` are not a bucket; they live in
`email_verification_failures`. They clear 24 h after the first failure, and no operator
procedure shortens that.

## 4. Investigate "a user can't log in", without PII

Never ask the user for their password or verification code. Never search logs or tables by
email address: the API stores only a keyed blind index, and the logs never contain the
address.

1. **What does the app show?** Each answer means something different.

   | App message (es-MX) | API answer | Meaning | Next step |
   |---|---|---|---|
   | "Correo o contraseña incorrectos." | `401` | Wrong password, unknown email, or a sign-up whose code was never confirmed. All three get the same answer on purpose | Re-register with the same email: a pending sign-up gets a new code; a verified account gets the "someone tried to sign up" notice instead. There is no password reset yet |
   | "Demasiados intentos. Espera N min…" | `429` | Rate limited | §3 |
   | "Este servicio no está disponible por ahora." | `404` | `module_auth` is OFF in that environment | Step 2 |
   | "Algo salió mal…" | `500` | Server error | Step 3, look for `level=ERROR` |
   | "No hay conexión con el servidor…" | none | Network or DNS | Not an auth problem |

2. **Is the flag ON?**

   ```sh
   curl -s http://localhost:3000/v1/flags   # {"enabled":["module_auth", …]}
   make -C backend flags                    # LOCAL; lists every flag and its state
   ```

3. **What do the audit events say?** The API logs JSON to stdout. Audit lines have
   `"msg":"auth audit"`, an `event` and a `request_id`, and `user_id` only when an account
   exists. The events are `account_created`, `login_failed`, `login_rate_limited`,
   `token_reuse`, `logout`, `account_deleted` and `delete_attempts_exhausted`. In a shared
   environment, filter the log store on the same JSON fields. Locally, with the API output
   saved to `api.log`:

   ```sh
   grep '^{' api.log | jq -c 'select(.msg == "auth audit" and .user_id == "<user-uuid>") | {time, event, request_id}'
   grep '^{' api.log | jq -c 'select(.msg == "auth audit" and .event == "login_rate_limited") | {time, request_id}'
   ```

   - Repeated `login_failed` for the id: wrong password.
   - `token_reuse`: a rotated refresh token was replayed and the family was revoked (D9). This
     is a sign-out, and on a flaky network it can be innocent.
   - Nothing at all for the id: the request never reached a known account. Either the
     address differs from the one registered, or the sign-up was never verified.

4. **Account state (needs the user id).** Run as `cc_app`, like §2:

   ```sh
   make -C backend db-psql-app PSQL_ARGS="-v user_id=<user-uuid>" <<'SQL'
   BEGIN;
   SELECT set_config('app.current_user_id', :'user_id', true) IS NOT NULL AS context_set;
   SELECT u.status, u.created_at,
          count(t.id) FILTER (WHERE t.revoked_at IS NULL AND t.rotated_at IS NULL
                                AND t.expires_at > now()) AS live_refresh_tokens,
          max(t.revoked_at) AS last_revocation,
          (array_agg(t.revoke_reason ORDER BY t.revoked_at DESC NULLS LAST))[1] AS last_revoke_reason
   FROM users u LEFT JOIN refresh_tokens t ON t.user_id = u.id
   GROUP BY u.status, u.created_at;
   COMMIT;
   SQL
   ```

   - No row: the id does not exist, or the account was deleted.
   - `status = suspended`: login is refused by design.
   - `last_revoke_reason = reuse_detected`: D9 signed the user out; they log in again.
   - `last_revoke_reason = admin`: an operator ran §2.

Known gap: the app does not show the `request_id`, and there is no operator lookup from an
email to a user id. Both are follow-ups (PLAN M1 follow-ups). Until they land, steps 1–3 are
the only way in for a user who cannot sign in.

## 5. Create `cc_auth_definer` in a new environment

The auth migrations (`20261013150000_auth_service.sql` onwards) hand their `SECURITY DEFINER`
functions to `cc_auth_definer`. The migration refuses to run, with a hint, if that role is
missing or privileged. `cc_owner` has `NOCREATEROLE`, so the role has to exist **before**
`migrate up`. That makes it a bootstrap step, not a migration (D1, M1-17 P-2; PostgreSQL ≥ 16
for `WITH INHERIT FALSE, SET TRUE`).

Locally, `infra/local/postgres/bootstrap/roles.sql` creates it when the volume is first
initialised (`make -C backend db-reset`). In a shared environment, run this once as the
administrative role that created `cc_owner` (on Cloud SQL, the built-in `postgres` user):

```sql
CREATE ROLE cc_auth_definer NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS NOINHERIT;
GRANT cc_auth_definer TO cc_owner WITH INHERIT FALSE, SET TRUE;
```

Check it before migrating. You should get exactly one row, `f | f | f | f | t`:

```sql
SELECT r.rolcanlogin, r.rolsuper, r.rolbypassrls, m.inherit_option, m.set_option
FROM pg_catalog.pg_roles r
JOIN pg_catalog.pg_auth_members m ON m.roleid = r.oid AND m.member = 'cc_owner'::regrole
WHERE r.rolname = 'cc_auth_definer';
```

Then run the migrations as `cc_owner`, and check the runtime role with
`infra/local/postgres/bootstrap/check-runtime-role.sql` connected as `cc_app` (locally:
`make -C backend db-roles-check`). No runtime role may ever be a member of `cc_auth_definer`.

## 6. Enable `module_auth` in a shared environment

`module_auth` stays **OFF** in staging and production until every item holds. Each item is
verified by whoever flips the flag and recorded in the deploy PR. Source:
[M1-17 §5](../reviews/M1-17-security-gate.md#5-preconditions-to-enable-module_auth-in-staging-or-production).

- [ ] **P-1 Real email provider, asynchronous.** The M1-06 adapter sits behind an outbox: the
      request path never waits on the provider and its answer never changes when the provider
      fails. Re-run the register/resend timing probe against staging: medians within noise.
- [ ] **P-2 `cc_auth_definer` exists** (§5) before the M1 migrations run, and
      `check-runtime-role.sql` passes for `cc_app`.
- [ ] **P-3 `APP_ENV` is not `dev`** (in `dev` the verification codes are logged), and
      `TRUSTED_PROXY_HOPS` equals the real number of proxies appending to `X-Forwarded-For`.
- [ ] **P-4 Keys from Secret Manager**: `AUTH_EMAIL_PEPPER`, `AUTH_EMAIL_KEY`,
      `AUTH_CODE_KEY`, `AUTH_ACCESS_TOKEN_KEY` and `DATABASE_URL` are injected at runtime,
      distinct per environment, never the local values, and never plain text in the image
      or service YAML. Run the key rotation (§1) once in staging.
- [ ] **P-5 Legal documents reviewed by a lawyer** (OB10): real URLs and versions in
      `legal_documents` replace the `.invalid` placeholders.
- [ ] **P-6 Gates**: M1-16 PASS, M1-17 PASS, REL-09 (deploy security gate) PASS, and
      ADR-0005 ACCEPTED.
- [ ] **P-7 Monitoring**: logs retained in the platform log store, with alerts on
      `event=token_reuse` and on a spike of `login_rate_limited`.
- [ ] **P-8 Mobile release build**: the release APK/AAB/IPA passes `mobile-artifact-check`,
      and the API base URL is HTTPS.
- [ ] Tech-lead hardening items FU1 and FU2 from [M1-16](../reviews/M1-16-tech-lead-gate.md)
      (mobile decode errors and sign-out store exceptions) are merged.

Then flip it with the environment's flag tooling (`cmd/flags` as `cc_owner`, credentials from
Secret Manager). Check that `GET /v1/flags` lists `module_auth` within 5 s.
