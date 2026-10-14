# M1-16: Tech-lead gate of module M1 Auth & Account

- **Date:** 2026-10-13
- **Branch:** `feature/auth-qa-fixes` (PR #49 → `dev`), which contains all of M1 (M1-01 … M1-15).
- **Scope:** A) PR #49 (M1-15 QA fixes). B) Holistic review of M1: `backend/internal/auth/**`,
  auth migrations, router/authn, `mobile/lib/features/auth/**`, mobile session core.
  C) QA-deferred D1, "first cold start logged out on a debug build".
- **Canon:** Clean Code, Clean Architecture, The Clean Coder, The Go Programming Language
  (idiomatic Go wins ties), SOLID, `~/.claude/rules/`.
- **Not covered here:** the security gate (M1-17) runs separately.

## Verdict: **PASS**

There are no open blockers. Two defects found in this gate are fixed on this branch, each
with a failing test committed before its fix:

- D1 cold-start logout (High).
- A contract mismatch on the `logout` 401 (High).

Before `module_auth` is enabled in a shared environment, FU1 and FU2 must land. They are
robustness gaps, not merge blockers.

PR #49: **approve once CI is green.**

## Evidence

| Check | Result |
|---|---|
| `flutter analyze` | no issues |
| `flutter test` | 417 passed, 10 skipped |
| `make check` (backend: gofmt, golangci-lint, `go test -race`) | green, 0 lint issues |
| Mutation: drop the generation check in `Session._restoreLate` | `a late answer never revives a session signed in and out meanwhile` fails |
| Revert the D1 fix (`session.dart` from `66d2481`) | `session_test` slow-read test and `session_guard_test` slow-read test both fail (login shown, home never reached) |
| Revert the logout fix | `TestLogout_401sCarryTheBearerChallenge` fails (both subtests) |
| Gate scan (test/CI relaxations in the PR #49 diff) | none detected (3 listed, 0 eligible) |
| Mutation: old `Countdown._refresh` (cancel on `inSeconds == 0`) | PR #49 `a tick just before the deadline…` fails (traced by hand: the last notify carries `isRunning == true`) |

Integration tests (testcontainers) were not re-run locally. No SQL or migration changed;
CI runs them.

## A) PR #49

| Item | Assessment |
|---|---|
| `Countdown._refresh` stops only when `!isRunning`, and notifies on a running→idle change | Correct root cause: the periodic timer fires a few ms early, so `inSeconds == 0` while still running cancelled the ticker before the end reached the listeners. The regression test simulates early ticks deterministically on the fake clock. |
| Login `429`: `RetryWait` countdown; a tap during the wait sends nothing and announces what is left | Correct. The test checks both the request count and the screen-reader announcement (`takeAnnouncements`), per the house a11y rule. |
| `RetryWait` / `announceRetryWait` extracted from the delete dialog; `accountDeleteRetryIn` → `authRetryIn` | Good DRY move; the ARB description was updated. |
| Notice after the wait (Low, FU5) | `describeAuthFailure` notice "Espera 3 min…" stays on screen after the countdown ends (login and delete dialog). Misleading but harmless. |
| Verify-code / register `429` (Low, FU5) | Still a static notice with no countdown, so taps are not blocked. Verify auto-submits only on a complete code, so the risk to the rate budget is small. |

## B) Holistic findings

### Fixed on this branch (failing test first)

| # | Sev | Finding | Fix |
|---|---|---|---|
| F1 | **High** | **D1: a slow secure-store read on cold start looked like a sign-out.** `Session.restore()` treated `read().timeout(5 s)` → `TimeoutException` as "no session" for good. The QA evidence is consistent with this: the token stayed on disk, and the next launch restored it. An error path is ruled out, because Android `resetOnError` defaults to `true` in flutter_secure_storage 11.2.0 and would have wiped the token. The plugin runs its first call (cipher init and Keystore unwrap) on a worker thread and posts the result back to the main looper, and on a starved emulator running a debug build that can exceed 5 s. A low-end phone would show the same thing. *Clean Code ch.7: a timeout is not an absent value.* | `9a2e313` (tests), `488d83a` (fix). The splash stays bounded, but the pending read is kept. When it answers with a token and the generation is unchanged (no sign-in or sign-out in between), the session becomes `signedIn` and the router leaves login for the kept location. Tests: `session_test.dart` (3 restore cases), `session_guard_test.dart` (end to end through the router). The fake store now snapshots the token at read time, as the platform does. |
| F2 | **High** | **`logout` 401s broke the contract.** The no-principal and malformed-refresh-token paths returned `toAppError(...)`, a plain apperr 401 without `WWW-Authenticate`. The contract's logout `401` is `components/responses/Unauthorized`, where the header is required (`openapi.yaml` `Unauthorized.headers`). The old tests only checked the code. *Contract is the source of truth (api-standards).* | `9c6cd71` (table-driven test), `443ae74` (fix). Both paths now return `gen.Logout401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}`, as `DeleteMe` already does. Other failures still go through the mapper (`TestLogout_otherFailuresStayErrors`). |

### What holds up

- **Dependency Rule.** `auth/domain` imports only stdlib and `x/text/unicode/norm` (ADR-0004). `auth/usecase` imports only `domain` and platform `clock`, `ratelimit` and `tx`. archtest has been transitive since P0-29.
- **RLS.** ENABLE + FORCE with per-operation policies on `users`, `user_usage`, `entitlements`, `refresh_tokens` and `consents`. `AssertIsolation` runs in the identity and sessions/consents integration tests.
- **P0-29 FU1 (request deadline)** is done: `middleware.Deadline`, 10 s default. **FU2** is done.
- **Mobile layering.** Presentation never touches Dio. Generated clients are used only behind repositories or `tokenRotationProvider`. `ApiFailure` is sealed, and `describeAuthFailure` is exhaustive.
- **D-decisions honoured.** `403 invalid_password` is mapped by its error code and never triggers a refresh. Refresh is single-flight, and generation guards stop stale work. `Retry-After` is parsed as seconds or as an HTTP-date.
- **Tests.** They use fake clocks and no real sleeps, and assert a11y announcements on the delete and login flows.

### Follow-ups (not fixed here)

| # | Sev | Finding | Owner / target |
|---|---|---|---|
| FU1 | **High** | **`guardApi` lets decode errors escape.** `api_failure.dart`: `} on Exception {`. When a generated decoder casts a null or wrongly typed field (`json['access_token'] as String`), it throws a `TypeError`, which is an `Error`, not an `Exception`. That error gets past every `on ApiFailure`, including `Session._rotateOnce` inside the interceptor. *Clean Code ch.7: wrap at the boundary.* Fix: also map `TypeError` and `CheckedFromJsonException` to `ApiUnexpected`, with a test that decodes `access_token: null`. | frontend-sr, M1 hardening, before `module_auth` is ON in shared envs |
| FU2 | **High** | **Sign-out can leak a store exception.** `SignOut.call` → `Session.readRefreshToken()` does not catch a store `PlatformException`. The `finally` still wipes, but the exception escapes from an `unawaited` call, and `SignOut` has no unit test. Fix: `readRefreshToken` returns `null` on a store `Exception` (as `_rotateOnce` does), and add `sign_out_test.dart` with `readError`. | frontend-sr, M1 hardening, same deadline |
| FU3 | Medium | **Error→HTTP mapping is split**, against MODULE-auth §4 ("one mapper"). `ErrWrongPassword` is mapped inline in `account.go`, `ErrAccountNotFound`→401 is written by hand in 4 places, and the `ErrorMapper` password rules can never fire because `fieldRules` catch those errors first. *SRP/DRY.* | backend-sr, M1 refactor (next touch of `authhttp`) |
| FU4 | Medium | **Rate-limit notice handling is written 4 ways.** Login and delete use `Countdown`, the code step uses `ResendCooldown`, and profile has no countdown. Notice state is `String?` in two forms and a record in two others. Announce calls are copied in 4 places. Extract one `FailureNotice` controller and one `announce` helper into `core/forms`. Use cases `SignIn` / `CompleteSignUp` are missing; presentation calls the repository and `sessionProvider.start` directly. *DRY, Clean Architecture.* | frontend-sr, before M2 forms copy the pattern |
| FU5 | Low | PR #49 leftovers: clear the "Espera N min" notice when the wait ends, and give verify/register `429`s the same countdown and tap block. `formatCountdown` truncates (`0:00` while still running for < 1 s), so round up to match the contract's rounding of `Retry-After`. | frontend-sr, with FU4 |
| FU6 | Medium | **P0-29 FU4 overdue.** The `access` package's ownership is still "pending an architect decision" (`docs/architecture/platform.md`), yet `auth/infra/trials.go` already uses `access.EntitlementTrial`. | architect, ADR now (before M2) |
| FU7 | Medium | **P0-29 FU3 still open.** `platform/httperr` imports `net/http` and has 7 importers, including `authhttp/authn.go`. Move it to `adapter/http/httperr`. The message `"Too many requests. Try again later."` is copied in 3 places, and `msgInvalidInput` duplicates the router's. | backend-sr, M1 refactor |
| FU8 | Low | **Two time sources in the composition root.** `cmd/api/auth.go` builds `clock.NewSystem()` twice and also passes `w.now`, so a test cannot freeze issue time and verify time together. Inject one `clock.Clock`. *DIP.* | backend-sr, M1 refactor |
| FU9 | Low | **Some tables are ENABLE without FORCE:** `trial_registry`, `email_verifications`, `email_verification_failures`, `legal_documents`. They are service tables keyed by email hash, not user-scoped, so this is not an RLS breach. But MODULE-auth §6.3 says "no unforced table", and the `identity.sql` comment ("cc_owner's SECURITY DEFINER functions") is stale since `cc_auth_definer` owns those functions. Either FORCE them with a definer policy, or reword §6.3 and the comment. | db-sr, M1 |
| FU10 | Low | **P0-29 FU5 still open:** archtest uses a denylist of framework prefixes, not an allowlist. | tech-lead, M2 |
| FU11 | Low | Android `resetOnError` is `true` by default, so one transient Keystore read error wipes the session for good. That fails safe, but it would look like D1 again. Measure restore latency on a low-end release build (QA D1's own suggestion), and decide on `AndroidOptions(resetOnError: …)` with security. | frontend-sr + security, M1 hardening |
