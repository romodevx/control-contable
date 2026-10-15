# ADR-0011: Access policy ownership and request authorization

| Field | Value |
|---|---|
| Status | **Proposed** (2026-10-15, WP4: PR #176 `refactor/platform-m7-readiness` and `refactor/platform-hygiene`). It records decisions that are already built and tested. The architect and tech-lead accept or amend it in review |
| Date | 2026-10-15 |
| Related | P0-29 tech-lead gate FU4 / FU5 / FU3, M1-16 gate FU6 / FU7 / FU10, [ADR-0001](0001-stack.md) (Dependency Rule, the flag gate before authentication), [ADR-0007](0007-fiscal-shared-kernel.md), [platform.md](../architecture/platform.md) `access`, [MODULES.md](../project-specs/MODULES.md) M7 (PR #58 follow-up), [CONTRACT.md](../api/CONTRACT.md) 0.17.0, `backend/internal/archtest` |
| Owner | architect (drafted by backend-sr) |

## Context

Two questions stayed open after P0, and the reviews kept carrying them:

1. **Where does the access policy live?** `internal/platform/access` (with `accessgorm`) holds
   business rules: the 7-day trial, the 5 + 5 record caps, entitlements, the 10-taxpayer cap.
   P0-29 FU4 asked for a decision: keep it as a documented shared-kernel port, or promote it to
   a module `internal/access/{domain,usecase,infra}`. M1-16 FU6 flagged it as overdue, because
   `auth/infra` already writes `access.EntitlementTrial`.
2. **What is the authorization model?** Nothing written down covered how a request is
   authenticated, which rows a caller may touch, or how "may this user write" is decided and
   shown. The pieces existed in code but were never stated as one decision. M7 (RevenueCat)
   brings the first caller that is not a user (the webhook), and the app needs to know a user's
   access before a write (MODULES.md M7, PR #58 follow-up).

## Decision

### 1. `access` stays a documented shared platform port

`internal/platform/access` stays where it is. It is **the only package under `platform/` that
holds business rules**. The rule "no new business logic in `platform/`" (P0-29 FU4) still holds
for everything else.

Why not a module:

- Every record module consumes it **inside its own transaction**: `ReserveRecord` bumps the
  counter in the caller's unit of work, so a rolled-back create gives its slot back. A module
  `internal/access` could not be imported by any of them (archtest `cross-module`). Each
  consumer would then need its own port and adapter for the same five rules: seven copies of
  one policy, which is exactly what drifts.
- It has no HTTP surface of its own. Its state reaches the client through the operations that
  need it (`getDashboard`, `getMe`, and M7's `getSubscription`), all through one shape (§3).
- M7 writes `pro` entitlements. That is the subscriptions module's job, done through its own
  infra on the `entitlements` table. The policy only reads them.

Constraints that come with the exception:

- `access` depends only on `apperr`, `clock` and `tx`. Its storage is the `accessgorm` adapter,
  running as `cc_app` inside the caller's unit of work, so RLS hides other users'
  `entitlements` and `user_usage`.
- If a second bounded context ever needs business rules shared this way, it goes to
  `internal/kernel` (ADR-0007) or becomes a module. `platform/` gets no second exception.

### 2. Authentication is chosen per operation from the contract

- The router runs, for each secured operation, **only the authenticator of the one security
  scheme that operation declares** (or the contract default `bearerAuth`).
  `router.Options.Authenticators` is keyed by scheme name. A user's access token never satisfies
  `revenueCatWebhookSecret`, and the webhook secret never stands in for a user (contract tests
  in `adapter/http/router/authn_test.go`).
- The router refuses to start, failing closed, when any of these holds:
  - a built operation's scheme has no authenticator;
  - an authenticator is keyed by a scheme the contract does not declare;
  - an operation declares alternative or combined schemes.
  A not-built operation answers the hidden-module `404` before authentication, so it needs no
  authenticator yet.
- Order (ADR-0001): flag gate → client limit → **authentication** → user limit → validation →
  handler. A caller cannot tell a disabled module (`404`) from an enabled one (`401`).
- Machine callers (the RevenueCat webhook) get no user principal. They run on their own database
  role (`cc_webhook`) and never reach a user operation.

### 3. Authorization: the principal's own rows, `404` otherwise; `402` only for writes

- **Scope.** Every user operation acts on the authenticated principal's user id, taken from the
  verified token and never from the request. Other users' resources answer `404` (RLS plus the
  ownership read), never `403`, so their existence is never revealed.
- **Defense in depth.** RLS (`ENABLE` + `FORCE`, policies on `app_user_id()`) is the floor. A
  query that looks a row up by a client-supplied id also names the caller in SQL
  (`user_id = (SELECT app_user_id())`, e.g. the reminders device read and delete), so a
  loosened policy alone never exposes another user's row. Integration tests prove it with FORCE
  RLS lifted.
- **Entitlement.** `access.Policy` decides whether a user may create or edit (`402
  entitlement_required` when read-only, or when a trial cap is spent). Reads, deletes and
  restores (Undo) stay open to a read-only user.
- **One view for the client.** `access.Summary.View` is the only computation of the client's
  `AccessSummary` (level, trial end, whole days left on the server clock, records left), and
  `adapter/http/accesshttp.Summary` is its only presenter. `getDashboard` and `getMe` (contract
  0.17.0) both use them, whatever `module_subscriptions` says. M7's `getSubscription` must use
  them too.

### 4. Delivery details stay at the edge

The HTTP error writer moved from `internal/platform/httperr` to `adapter/http/httperr` (P0-29
FU3). archtest now enforces the direction:

- `inward-adapter`: only adapters and `cmd/` import the top-level `adapter/` tree.
- `usecase-imports`: a use case imports from outside the module only an allowlist (P0-29 FU5).

## Consequences

- P0-29 FU3, FU4 and FU5, and M1-16 FU6, FU7 (the move; message dedupe stays open) and FU10,
  are closed by this ADR and its PRs.
- A new security scheme, such as an admin API key, is one entry in the contract plus one
  authenticator in `cmd/api`. Forgetting the authenticator is a boot failure, not an open route.
- A delegated-access feature (an accountant acting for a user) is a new ADR: it breaks the "own
  rows only" scope. PLAN M5 G-3 and M6 G-4 list the queries that rely on RLS alone.
