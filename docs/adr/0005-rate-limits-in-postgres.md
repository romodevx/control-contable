# ADR-0005 — Auth rate limits live in PostgreSQL (GCRA buckets)

| Field | Value |
|---|---|
| Status | PROPOSED (backend, M1-10; tech-lead and security gates M1-16/M1-17 decide) |
| Date | 2026-10-13 |
| Related | [ADR-0001 O9 Redis](0001-stack.md), [MODULE-auth §5](../project-specs/MODULE-auth.md), migration `20261013170000_rate_limits.sql`, PR #43 security gate follow-up |
| Owner | backend-sr / qa-sr |

## Context

ADR-0001 O9 keeps rate-limit counters in process "until we scale beyond one instance". The
auth limits of MODULE-auth §5 are security controls (login brute force, sign-up abuse,
password oracle on `deleteMe`), and the PR #43 gate noted that in-process counters give N
instances N times the attempts. Cloud Run scales out by default, so an in-process login
limit is not a limit. Redis is still not adopted (O9).

## Decision

1. **Shared buckets in PostgreSQL.** One row per `(scope, key_hash)` in
   `rate_limit_buckets`, GCRA (generic cell rate algorithm): the row holds the bucket's
   theoretical arrival time. `rate_limit_take(scope, key, now, interval, burst)` decides with
   one `UPDATE ... WHERE greatest(tat, now) + interval <= now + interval * burst`, which
   READ COMMITTED re-evaluates on the latest row version after waiting on a concurrent take:
   racing instances admit exactly the burst (24-connection test). A denied take writes
   nothing; the wait is returned for `Retry-After`. O(1) per decision, one row per key.
2. **Each decision is its own autocommit statement** on the cc_app pool, never part of the
   request's unit of work: a failed login counts even when the request's transaction rolls
   back.
3. **Keys are HMACs.** The store digests `scope NUL key` with a key derived from the email
   pepper (`HMAC(pepper, "control-contable/rate-limit-key/v1")`), so no client address or
   email reaches the table and digests are unrelated to `email_hash`.
4. **Fail closed.** A limiter error answers 500 (router) or fails the use case; it never
   admits.
5. **Where each limit sits.** Per-client limits per operation run in the router right after
   the flag gate (a disabled module spends nothing and never answers 429) and before
   authentication and validation. Per-email, per-family and per-account limits run in the
   use cases, before any credential work. The in-process edge limiter (O9) stays as flood
   protection in front of everything.
6. **Clock.** The caller passes its injected clock (deterministic tests). Instance clocks
   are NTP-synced and the windows are minutes, so skew is immaterial.

## Consequences

- One extra round trip per limited request (three for login: client, email-network and
  email ceiling; two for register: client + email).
- The table grows with distinct keys. Buckets whose `tat` has passed are full again (the
  same state as a missing row), so **M9 adds a worker job that deletes them** (the worker
  role needs a grant then). There is deliberately no index on `tat`: it would make every
  take a non-HOT update; the purge scans. `fillfactor = 90` keeps room for HOT updates.
- **Targeted lockout (M1-17 security gate).** A per-email failure budget alone lets anyone
  hold a victim's logins at one per 3 minutes by spending it. Login therefore spends two
  budgets, both keyed by the email's blind index and both refilled by a success:
  - `auth_login_email_client`: 5 failures / 15 min per **(email, client network)**, the
    network being the address resolved behind `TRUSTED_PROXY_HOPS` widened to IPv4 /24 or
    IPv6 /64 (`middleware.ClientNetwork`, carried to the use case via `clientnet`). A
    stranger spending it locks out only their own network.
  - `auth_login_email`: a ceiling of 50 failures / h per email across all networks (then
    one per 72 s), so distributed guessing stays bounded (≈ 1 200 guesses / day / account
    at worst, against argon2id and the common-password blocklist).
  A request its network refuses spends nothing of the ceiling. **Residual risk, accepted for
  MVP:** an attacker controlling ≥ 10 distinct /24s (or /64s) can still exhaust the ceiling
  and hold a victim at one try per 72 s; a victim sharing the attacker's /24 (CGNAT, campus
  NAT) shares that network's budget. Revisit with device-bound login tokens or a step-up
  (CAPTCHA / emailed sign-in link) when the ceiling is hit. The 429 is identical for known
  and unknown emails, so neither budget enumerates.
- The per-email sign-up budget (5 / h, register + resend) can likewise be spent by a
  stranger to delay a victim's sign-up; no account exists yet and the window is an hour,
  so it is accepted.
- Keying by (email, network) multiplies bucket rows for spraying attacks; each login also
  costs a client take, so rows stay bounded by the per-client limits until the M9 purge.
- The D3 sign-up notice dedupe stays in process (best effort, not a security limit).

## Alternatives

- **Redis / Memorystore**: rejected for now (O9: one more managed dependency for one use).
- **Fixed-window counters** (`count`, `window_start`): allow 2× bursts at window edges and
  need a reset write; GCRA is exact with a single timestamp.
- **Advisory locks + read-modify-write**: two round trips and lock management for what one
  conditional UPDATE does.
