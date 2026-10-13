# P0-28 — QA of the backend platform foundation

- **Date:** 2026-10-13
- **Scope:** `dev` at `78a1b0c` (P0-03 … P0-27 merged). API-level QA; there is no UI yet.
- **Branch:** `feature/p0-qa-fixes`
- **Method:** the real binaries (`cmd/api`, `cmd/worker`, `cmd/flags`) against the LOCAL
  docker compose stack, plus the full Go suite (`make check`, `make test-integration`,
  `cover-check`, `codegen-check`, `vuln`).

## Summary

| Area | Result |
|---|---|
| Fresh-clone path (`db-reset` → `migrate-up` → api/worker/flags) | Works; 3 frictions, 1 fixed (`make api` / `make worker`) |
| RLS fail-closed (users, user_usage, entitlements × cc_app/cc_webhook/cc_worker; email_verifications) | Pass. Two guard tests added |
| Flag ON/OFF via real binary + `apply-preset demo_core` / `demo_full` | Pass: OFF answers match the unknown-route answer exactly |
| Exploratory (paths, bodies, rate limit, SIGTERM, DB down) | **1 defect found and fixed** (hung database held gated requests 60 s) |
| Full suite | Green on the branch; business-logic coverage 92.0% |
| CI on `dev` | Green on both pipelines after re-running a run that billing had blocked (environment, not code) |

## 1. Fresh-clone experience

Ran from `backend/`: `make db-reset` (6 s, healthy), `make migrate-up` (baseline + identity),
`make migrate-status`, `make db-roles-check` (3 roles: NOSUPERUSER, NOBYPASSRLS, not owner,
no DDL), `make flags`, `make flags ARGS=presets`, then `cmd/api` and `cmd/worker`.

| # | Friction | Status |
|---|---|---|
| F1 | No target or documented recipe to run `cmd/api` / `cmd/worker` locally: the `cc_app` / `cc_worker` URL had to be hand-built from `infra/local/.env`. | **Fixed**: `make api` (PORT=`API_PORT`, default 3000) and `make worker`, both LOCAL-only through `assert-local.sh`; README updated with the from-zero sequence. |
| F2 | Config default `PORT=8080` is taken on this machine by another stack (`bind: address already in use`). The registry slot is 3000. | **Fixed for local** via `make api`. The config default stays 8080 (what Cloud Run injects). |
| F3 | `compose.yaml` pins `container_name` and the volume name, so every worktree of this repo shares one database. Two parallel agents running `db-reset` destroy each other's data. | Deferred (DevOps): consider `COMPOSE_PROJECT_NAME`-scoped names, or document "one local stack per machine". |
| F4 | `cmd/flags` run as the wrong role says `migrations: must run as cc_owner` (the sentinel lives in `migrations`). Correct, but the prefix misleads. | Deferred (cosmetic). |

## 2. RLS fail-closed

Two layers:

1. **Live local stack, over TCP as each role** (two users seeded as `cc_app` under their own
   context). Every result was fail-closed:
   - `cc_app`, no context: 0 rows on users, user_usage, entitlements.
   - `cc_app`, malformed context: error (`invalid input syntax for type uuid`), no rows.
   - `cc_app` as A: 0 rows of B; own rows visible; UPDATE/DELETE of B match nothing; INSERT
     of a user_usage row for B rejected by policy.
   - `cc_app`: `SET ROLE cc_owner` denied; `SET row_security = off` refused by FORCE.
   - `cc_webhook`: permission denied on users and user_usage; entitlements only through its policy.
   - `cc_worker`: permission denied on all three.
   - `email_verifications`: SELECT and INSERT denied to all three roles.
2. **Suite** (`migrations/*_integration_test.go`, testcontainers): `AssertIsolation` per table
   already covered isolation and unset context. Two guard tests added. Both pass on the current
   schema; they exist so a later migration cannot regress:
   - `TestSchemaPublic_EveryTableHasRowLevelSecurity`: a catalog sweep. Any table in `public`
     without ENABLE, or without FORCE outside an explicit allow-list (flag tables,
     `email_verifications`), fails. A subtest creates two bad tables to prove the sweep is not
     vacuous.
   - `TestEmailVerifications_DenyAllForEveryRuntimeRoleDespiteAGrant`: the existing deny-all
     check covered only `cc_app`. It now also covers `cc_webhook` and `cc_worker`.

## 3. Flag ON/OFF through the real binary

The probe hit 30 requests: platform routes, unknown routes, a sample of each module's routes,
malformed ids and odd methods.

- **All OFF (default):** every module route answers `404` with `{"error":{"code":"not_found",...}}`
  and the same header set (`Content-Type`, `X-Content-Type-Options`). That is byte-identical to
  `GET /nope`, `POST /v1/nope` and `/v1/taxpayers/x/nope` once `request_id` is masked. `PUT /v1/me`,
  `OPTIONS`, `TRACE` and `POST /healthz` also get a 404, never a 405 + `Allow`.
- **`apply-preset demo_core`:** `/v1/flags` lists the 6 modules within the 5 s TTL. History,
  reminders and subscriptions stay identical to an unknown route. Core routes reach contract
  validation (400 `validation_failed`, values never echoed). Routes that pass validation reach
  the not-yet-built handler.
- **`apply-preset demo_full`:** all 9 listed, and the M7/M8/M9 routes leave 404.
- **Back to OFF** (`cmd/flags disable` ×9): `/v1/flags` returns `[]` and routes return 404 within the TTL.
- CLI input validation: an unknown preset, an unknown key and an injection-shaped key are all
  rejected (exit 1). The wrong role is refused.

## 4. Exploratory

| Probe | Result |
|---|---|
| Unclean paths (`//healthz`, `/healthz/`, `/v1//flags`, `/V1/FLAGS`, `../`) | 404 envelope, no redirect |
| `%2e%2e` and `%00` in a path id | 400 `validation_failed` (format) |
| Malformed query on a parameterless route (`/v1/flags?x=%zz`) | 200, ignored (acceptable) |
| 20 MB body to an ON route, an OFF route, `/healthz` and an unknown route | ON: 400 (body capped by `MaxBytesReader`); OFF/unknown: 404 in under 1 ms; RSS stable (~31 MB) |
| Rate limit, defaults (10 rps, burst 20) | 429 + `Retry-After: 1`, also on unknown routes |
| Spoofed `X-Forwarded-For` with `TRUSTED_PROXY_HOPS=0` | Ignored: one bucket, 429 after the burst |
| SIGTERM during an in-flight request (DB paused) | Request answered (200 `degraded` after the 500 ms bounded ping); exit 0; new connections refused |
| DB stopped mid-run | `/healthz` 200 `unreachable`; gated routes 404 (fail closed); `/v1/flags` 500 envelope (contract-declared); full recovery after restart |
| DB down at boot | dev: starts, degraded, fail closed. staging: **refuses to start** (`runtime database role` check) as designed |
| Worker: DB down / hung mid-run, then SIGTERM | Leader election retries every 10 s, leadership lost and re-acquired, `worker stopped`, advisory lock released (0 in `pg_locks`) |
| **DB hung (paused) mid-run** | **DEFECT D1** (below) |

### D1 — a hung database held every gated request (fixed)

- **Symptom:** with the database paused (it accepts TCP but never answers), every gated request
  and `GET /v1/flags` hung until the client gave up: 60 s in the probe, for each of 10
  concurrent clients. The flag read had no deadline. Callers queued behind a failed read each
  retried it in turn, so the N-th caller waited N reads. The "fails closed" promise held for a
  database that refuses connections, but not for one that hangs.
- **Fix** (`internal/platform/flags/service.go`):
  - each read is bounded by `flags.DefaultLoadTimeout` (2 s; `WithLoadTimeout` option);
  - the read runs detached from the cancellation of the caller that started it, so one client
    disconnecting no longer fails the callers waiting on the same read;
  - callers that waited on a failed read share its error instead of queueing another read.
- **Regression tests** (written first and seen failing: unbounded hang, 16 loads instead of
  shared): `TestLoadIsBoundedByTheLoadTimeout`, `TestWaitersShareAFailedLoad`,
  `TestLoadSurvivesTheCancellationOfTheCallerThatStartedIt`, `TestDefaultLoadTimeoutIsShort`,
  `TestNewServiceRejectsANonPositiveLoadTimeout`.
- **Verified on the real binary:** same scenario, gated routes now answer 404 in 2.0 s
  (sequential and 10 concurrent); `/v1/flags` answers the 500 envelope in 2.0 s.

### Observations (not defects, routed)

| # | Observation | Owner |
|---|---|---|
| O1 | A flag turned ON for a module that is not built yet makes the embedded nil `notBuilt` interface panic. `middleware.Recover` turns that into a 500, and every request logs an ERROR with a full stack trace. This is documented in `cmd/api/api.go`, but it means **`demo_core` / `demo_full` cannot be demoed today**: every core route answers 500. Option: a `notBuilt` that answers 501 `not_implemented` (contract change) or the 404 of a hidden module. Either way, presets need their E2E smoke once M1 lands. | Architect + Tech Lead |
| O2 | `X-Content-Type-Options: nosniff` is set only on error envelopes; 200 JSON responses carry no `nosniff` or `Cache-Control: no-store`. This matters once authenticated M1 payloads exist. | Security |
| O3 | Contract validation reports only `period` (required) for `/v1/taxpayers/not-a-uuid/incomes`, so the malformed id shows up only on the next attempt (kin-openapi stops after a missing required query param). | Backend (M3) |
| O4 | `GET /v1/me` etc. reach handlers without authentication (`router.NoAuthentication` until M1-07). Expected; listed so M1 QA checks 401 before 500. | QA (M1) |

## 5. CI on `dev`

`gh run list --branch dev`: the latest backend run (P0-27 merge, `38450807531`) **failed in
4 s**. None of its three jobs started: *"recent account payments have failed or your spending
limit needs to be increased"*. This is an **environment issue, not a code defect**. It was
transient: this PR's run went green, and the blocked `dev` run was re-run (`gh run rerun`) and
went **green** (lint/codegen/vuln, unit, integration + coverage). The latest mobile run on `dev`
(P0-18 merge) is green, and no mobile path changed after it. Locally, the same pipeline steps
all pass on `dev` and on this branch:

| Step | `dev` (`78a1b0c`) | branch |
|---|---|---|
| `fmt-check`, golangci-lint v2.14.0 | 0 issues | 0 issues |
| `test` (-race) | ok | ok |
| `test-integration` (testcontainers, contract tests included) | ok | ok |
| `cover-check` | 92.0% (1700/1846) | 92.0% (1713/1860) |
| `codegen-check` | no drift | no drift |
| `vuln` (govulncheck v1.8.0) | no vulnerabilities | not rerun (no dependency change) |

**Owner action:** check the GitHub billing/spending limit. One run was blocked by it, and it
can block the next one.

## Environment issues (kept apart from app bugs)

- GitHub Actions jobs were blocked by billing once (section 5); the re-run is green.
- Port 8080 is taken by an unrelated local stack (F2).
- The local compose stack is shared by every worktree (F3); nothing collided during this run.
- The local database was wiped (`make db-reset`) and migrated at the start. The test users
  seeded for the RLS probes were deleted afterwards, and all flags were returned to OFF.
