# P0-30 · Security gate: backend platform foundation

| Field | Value |
|---|---|
| Scope | Backend platform as built on `dev` @ `a1a189e` (P0-03..P0-27): DB roles and bootstrap, RLS session context, crypto, log redaction, HTTP edge, flags gate, worker, CI and supply chain, repository settings (repo is now **public**) |
| Reviewer | security |
| Date | 2026-10-13 |
| Method | `/cso` (gstack-cso 3.0.0, run `1791919621798-a032a9b282766e87`, daily static: gitleaks, OSV-Scanner 2.6.0, Trivy, zizmor 1.30.1; semgrep returned `INVALID_OUTPUT`, history scan over the helper's limit), followed by a manual review against `security.md` and `database-rls.md`, `govulncheck` with and without test code, and the carried follow-ups of PRs #5, #7, #9, #12, #13, #19 and #21 |
| Verdict | **PASS.** No open blocker. M1/M2 may consume the platform. The deferrals below have owners and target tasks. Shared-environment enablement still needs REL-09 |

## Evidence

| Check | Result |
|---|---|
| `make fmt-check`, `golangci-lint run ./...` | clean, 0 issues |
| `go test -race ./...` | green |
| `make test-integration` + `make cover-check` | green, business-logic coverage 92.0 % (gate 80 %) |
| `govulncheck ./...` (what CI ran) | 0 called vulnerabilities |
| `govulncheck -test -tags=integration ./...` (new CI scope) | **was 1 called (GO-2026-6253, moby/go-archive, via testcontainers)**; 0 after the bump |
| `shellcheck` on `infra/local` scripts, `assert-local_test.sh` | clean / ok |
| Throwaway postgres:18.6 with the bootstrap | `check-runtime-role.sql` passes for cc_app, refuses cc_app once it is made a member of cc_owner, and refuses cc_app's login to `postgres` |
| gitleaks (cso) | 5 candidates, all false positives: base64 test fixtures (`saltsalt…`, `keykey…`), ADR table keys, the spec embedded in `api.gen.go`. GitHub secret scanning (now on) has 0 alerts |
| zizmor (cso, offline) | 0 findings on the 3 workflows |

## Threat model (STRIDE), platform as built

Actors: an anonymous internet caller; an authenticated user (auth lands in M1); a contributor
or fork PR on a **public** repo; a developer running `make` locally; an attacker who
obtains a runtime DB credential.
Assets: taxpayer PII (RFC, CURP, email), password hashes, the AES-GCM keyring and blind-index
peppers, flag state, CI gate configuration, release obfuscation symbols.

| Component / trust boundary | Threat | Control as built | Residual / action |
|---|---|---|---|
| Internet → `cmd/api` (`/healthz`, `/v1/*`) | **S** spoofed client IP to dodge the rate limit | `ClientIP(trustedHops)` takes the n-th XFF entry from the right; IPv6 keyed by /64 | OK. Set `trustedHops` correctly per environment (REL) |
| | **T** malformed or oversized input | Contract validation at the edge (kin-openapi), body cap, uniform 400 envelope | OK |
| | **I** route discovery while a module is off | Uniform 404 before binding, decode and authn; non-canonical paths answer 404 (PR #19 fix); `platform` allow-list now pinned by a test | Small cache-miss timing oracle (accepted, PR #19 FU-1) |
| | **I** error text leaking internals | `apperr` mapper: unknown errors become `internal_error` and the cause never reaches the wire. `httperr.Write(nil)` no longer panics | `Cache-Control: no-store` on errors needs a contract change (deferred D-6) |
| | **D** request floods; hung DB | Rate limit with `Retry-After`; `statement_timeout` per role; flag reads bounded at 2 s (PR #27, D1) | Flag-read bound lands with #27 |
| | **E** unauthenticated access to user data | No user-data route is enabled: every module flag defaults OFF; `module_auth` is not built | M1 gate |
| API/worker → PostgreSQL | **E** runtime role bypasses RLS | cc_app/cc_webhook/cc_worker: NOSUPERUSER, NOBYPASSRLS, NOREPLICATION, NOINHERIT, no memberships, not owner; `CheckRuntimeRole` refuses to boot outside dev | OK. **Fixed now:** maintenance DBs `postgres`/`template1` were open to PUBLIC (F-1) |
| | **T/I** cross-user reads through a pooled connection | `set_config('app.current_user_id', $1, true)` first in every tx; UUID validated; foreign-manager nesting refused (PR #21); FORCE RLS sweep (PR #27) | **Fixed now:** the nil UUID was accepted as a user (F-4) |
| | **E** repository escapes the unit of work (`Commit`, session `SET`) | Review convention | Lint/archtest rule deferred (D-3) |
| | **E** HTTP code reaching `System` (no user context) | Low impact: as cc_app it sees no user rows | Interface split deferred (D-2) |
| Crypto (`platform/crypto`) | **I** key or pepper printed in logs | REDACTED Format/LogValue; **fixed now:** also on dereferenced values (F-3) | — |
| | **D** CPU burn through long passwords or tampered hashes | Stored-hash cost caps (PR #7); **fixed now:** `MaxPasswordBytes` = 1024 in Hash/Verify, on top of the contract's maxLength 128 (F-3) | — |
| | **T** ciphertext swapped between rows | key_id is bound into the AAD (PR #7) | Row-id AAD is an M2-05 obligation (D-4) |
| | **S** account confusion via Unicode case folding | ASCII-only normalisation (PR #7) | NFC for EAI emails deferred (D-5) |
| Logging | **I** PII and secrets in logs | Key- and value-based redaction (PR #9); the "log IDs, not payloads" rule is now in the package doc | Pattern-based by nature; stays a review item |
| Local stack (`make db-*`) | **T** reset/migrate pointed at a shared DB | Docker endpoint must be a unix socket; `CC_DB_HOST` must be loopback; **fixed now:** `DATABASE_URL` host is parsed instead of globbed (F-2) | — |
| CI (public repo) | **T/E** fork PR runs code with secrets | `pull_request` only, `contents: read`, `persist-credentials: false`, no secrets; **set now:** approval required for every external contributor's workflow run | OK |
| | **T** PR weakens a gate (coverage threshold, scanner scope) | **Added:** CODEOWNERS on gate files; aggregate required checks | Required review is not enforceable on a single-account repo (D-7) |
| | **I** obfuscation undone through a public artifact | **Fixed now:** prod split-debug-info is no longer uploaded (F-5) | Artifacts already uploaded expire within 14 days |
| | **E** vulnerable test dependency in CI | **Fixed now:** govulncheck covers test and integration code; moby/go-archive bumped (F-6) | npm dev tooling alerts deferred (D-8) |
| Repository | **R** a report has no channel | **Added:** SECURITY.md + private vulnerability reporting | — |
| | **I** a secret is pushed to the public repo | **Set now:** secret scanning + push protection | Non-provider patterns and validity checks need GitHub Secret Protection (not available on this plan) |

Repudiation (R) at the platform layer: every line carries `request_id`, and the worker logs job,
outcome and duration. Audit logging of sensitive operations is a module concern (M1 auth, M2
taxpayers) and part of their gates.

## Findings fixed in this PR (failing test first in each case)

| ID | Sev. | Finding | Fix | Regression test |
|---|---|---|---|---|
| F-1 | Low | PR #5 FU-2/3: `postgres` and `template1` kept the default ACL, so every cc_* role could CONNECT and create TEMP objects there. `check-runtime-role.sql` ignored REPLICATION and role membership | `roles.sql` revokes ALL on both DBs from PUBLIC. The check script now also fails on REPLICATION, any `pg_auth_members` row, and CONNECT on a maintenance DB | `TestBootstrap_RolesCannotReachMaintenanceDatabasesOrOtherRoles` (was red for all 4 roles) |
| F-2 | Low | PR #5 FU-1: the reset guard globbed `DATABASE_URL`, so `postgres://u:p@prod.example.com/x?a=@localhost/` passed (cso finding) | `url_host` parses the authority: last `@`, brackets for IPv6, port stripped; exact host match | `infra/local/scripts/assert-local_test.sh`: 16 cases, 4 were red; now runs in backend CI |
| F-3 | Low | PR #7 FU-3/4: no password length bound below the HTTP edge; a dereferenced `Keyring`/`BlindIndex` printed key schedules and pepper | `MaxPasswordBytes`/`ErrPasswordTooLong`; value receivers for Format/LogValue | `TestHashAndVerifyRejectOverlongPasswords`, `TestKeyringValueNeverPrintsKeyMaterial`, `TestBlindIndexValueNeverPrintsPepper` |
| F-4 | Info | PR #21 FU-4: the nil UUID was accepted as a user id | `ErrNoUser` | `TestUserUnitOfWorkRejectsBadUserBeforeConnecting` (2 new cases) |
| F-5 | Medium | New, caused by the repo going public: `mobile.yml` uploaded the prod `--split-debug-info` symbols as an artifact, which is now public, so anyone could undo the Dart obfuscation | Upload removed; README updated; REL-04 sends symbols to crash reporting | n/a (CI config) |
| F-6 | Medium | New: CI's govulncheck skipped test code, so GO-2026-6253 (moby/go-archive, path traversal on tar extraction, reached by testcontainers in CI) went unseen; Dependabot alerts #1/#2 | moby/go-archive v0.2.0 → v0.3.3, moby/sys/user v0.4.0 → v0.4.1 (both older than two weeks); CI runs `make vuln PKGS="-test -tags=integration ./..."` | govulncheck: 0 called |
| F-7 | Low | PR #13 FU-2: workflow `paths` filters plus required checks leave PRs stuck on "expected"; `infra/local/**` (the bootstrap the integration harness runs) did not trigger backend CI at all | Each workflow runs on every PR: a `changes` job (diff of the merge commit vs its base) gates the jobs, and an always-running `<wf>-ok` job fails unless every gate succeeded or was skipped. Backend now also covers `infra/local/**` | n/a (CI config); see the CI run on this PR |
| F-8 | Info | PR #9 FU-1: `httperr.Write(w, r, m, nil)` panicked | nil maps to `internal_error` | `TestWrite_nilErrorIsInternalError` |
| F-9 | Info | PR #12 FU-3: nothing stopped production code from importing `pgtest` (superuser logins, fixed passwords) | archtest rule `test-harness` (cmd/ included) | `TestCheckImport` (6 new cases) |
| F-10 | Info | PR #19 FU-4: a new operation mislabelled `x-module: platform` would ship ungated | Pinned allow-list `{getHealth, listEnabledFlags}` | `TestPlatformOperationsAreAllowListed` |

## Carried follow-ups: disposition

| Source | Item | Disposition |
|---|---|---|
| PR #5 | 1 crafted `DATABASE_URL` passes the guard | **Fixed** (F-2) |
| | 2 runtime roles CONNECT/TEMP on `postgres` | **Fixed** (F-1) |
| | 3 check script ignores replication and membership | **Fixed** (F-1) |
| | 4 `PGPASSWORD` on the `docker compose exec` argv | **Deferred D-1.** The fix is in `backend/Makefile`, which PR #27 is changing |
| | 5 random passwords on first `.env` | **Deferred D-1** (same file); local placeholders on a loopback-bound container |
| PR #7 | 3 password length cap | **Fixed**: contract maxLength 128 at the edge plus `MaxPasswordBytes` in the hasher (F-3) |
| | 4 value receivers | **Fixed** (F-3) |
| | 5 NFC for EAI emails | **Deferred D-5** |
| | 6 M2-05 passes the taxpayer row id as AAD | **Deferred D-4**, blocking M2-05's gate |
| PR #9 | 1 `Write(nil)` panics | **Fixed** (F-8) |
| | 2 `Retry-After` on 429 | **Closed**: `middleware.RateLimit` sets it before `Write` (`TestRateLimit_answers429WithRetryAfter`); now documented on `Write` |
| | 3 `Cache-Control: no-store` on errors | **Deferred D-6**: the contract test rejects undeclared headers, so this needs a contract change |
| | 4 "log IDs, not payloads" | **Fixed**: stated in the `logging` package doc |
| PR #12 | 1 `CheckInsertRejected`/`CheckNoBypass` pass on a missing grant | **Fixed**: documented on both functions (`AssertIsolation` already covers it) |
| | 2 `Table.Name` unqualified only | **Closed**: already documented on `Table` |
| | 3 archtest ban on importing `pgtest` | **Fixed** (F-9) |
| PR #13 | 1 CODEOWNERS on gate files | **Fixed**: `.github/CODEOWNERS`. Enforced review is D-7 |
| | 2 `paths` filters vs required checks | **Fixed** (F-7) |
| | 3 loose `/cmd/` in `COVER_EXCLUDE` | **Deferred D-1** (`backend/Makefile`). No business logic sits under a nested `cmd/` today |
| PR #19 | 1 cache-miss timing oracle | **Accepted** (tiny; no data) |
| | 2 DB down serialises gated requests | **Resolved by PR #27** (D1: 2 s bounded reads, shared failure). Closes when #27 merges |
| | 3 API asserts it is not privileged | **Closed** in PR #21 (`CheckRuntimeRole`) |
| | 4 pin the `platform` allow-list | **Fixed** (F-10) |
| PR #21 | 1 split `tx.Manager` / `SystemManager` | **Deferred D-2** |
| | 2 forbid `Commit`/`Rollback`/`SET` in repositories | **Deferred D-3** |
| | 3 "RO is a hint when nested" | **Fixed**: stated on `tx.Manager.ReadOnly` |
| | 4 reject `uuid.Nil` | **Fixed** (F-4) |
| govulncheck | kin-openapi / x/crypto advisories | **kin-openapi v0.149.0: no advisory today.** x/crypto v0.58.0: only GO-2026-5932 (`openpgp` is unmaintained; no fix exists). The package is not imported (module-level only, not called) |

## Deferrals

| ID | Item | Owner | Target |
|---|---|---|---|
| D-1 | `backend/Makefile`: pass `PGPASSWORD` by name (`-e PGPASSWORD`), generate random local passwords when `.env` is first created, anchor `COVER_EXCLUDE` to `^github.com/romodevx/control-contable/backend/(cmd/|adapter/http/gen/)` | devops-sr | First PR after #27 merges |
| D-2 | Split `tx.Manager` (user) from `tx.SystemManager`; archtest forbids `SystemManager` in `adapter/http` except sign-up | backend-sr + tech-lead | M1-05 (before the first module consumes the port) |
| D-3 | Lint or archtest rule: repositories under `internal/*/infra` never call `Commit`/`Rollback`/`SavePoint` or run `SET`/`set_config` | tech-lead | M1 first repository PR |
| D-4 | M2-05 passes the taxpayer row id as GCM AAD (test: a ciphertext moved to another row fails) | backend-sr | M2-05 (blocks its security gate) |
| D-5 | NFC normalisation before `email_hash`, if EAI addresses are accepted | backend-sr | M1-05 decision |
| D-6 | Declare `Cache-Control: no-store` on error responses in `docs/api/openapi.yaml`, then set it in `httperr.Write` | architect | Next contract revision |
| D-7 | Required code-owner review. Not enforceable while agents and owner share the `romodevx` account (GitHub forbids self-approval). Reviews are enforced by process (tech-lead/security PR comments) | owner | When a second account or GitHub App exists |
| D-8 | npm dev-tooling alerts in `tools/contract` (Prism 5.16.0 transitive: @faker-js/faker 5.5.3, lodash 4.17.21, uuid 8.3.2, sprintf-js, braces). Local-only mock and linter, no attacker-controlled input, loopback-bound. Bump Prism or add `overrides` once fixed releases are two weeks old | devops-sr | Next Dependabot cycle |
| D-9 | Re-apply required status checks `backend-ok`, `mobile-ok`, `contract-ok` on `dev` and `main` once this PR is merged (command below) | owner / devops-sr | Right after merge |

## Repository settings applied (gh api, 2026-10-13)

| Setting | Value |
|---|---|
| Secret scanning | enabled |
| Push protection | enabled |
| Non-provider patterns, validity checks | not available (need GitHub Secret Protection); still disabled |
| Dependabot alerts | enabled (9 alerts at first scan: 2 Go, fixed here; 7 npm dev tooling, D-8) |
| Dependabot security updates | enabled |
| Private vulnerability reporting | enabled (SECURITY.md points to it) |
| Fork PR workflow approval | `all_external_contributors` (was `first_time_contributors`) |
| Default `GITHUB_TOKEN` permissions | `read`, and it cannot approve PRs (already set; verified) |
| Branch protection `dev` and `main` | PR required, 0 approving reviews, no code-owner review, force-push blocked, deletion blocked, `enforce_admins` false. **Required status checks: none for now** (D-9) |

Status checks are not required yet because no always-reporting check exists on `dev` until
this PR merges. They briefly were, and that blocked every merge, so they were removed. After
the merge, run:

```sh
cat > /tmp/protection.json <<'EOF'
{"required_status_checks": {"strict": false, "checks": [
   {"context": "backend-ok", "app_id": 15368},
   {"context": "mobile-ok", "app_id": 15368},
   {"context": "contract-ok", "app_id": 15368}]},
 "enforce_admins": false,
 "required_pull_request_reviews": {"required_approving_review_count": 0},
 "restrictions": null, "allow_force_pushes": false, "allow_deletions": false}
EOF
for b in dev main; do
  gh api -X PUT "repos/romodevx/control-contable/branches/$b/protection" --input /tmp/protection.json
done
```

(`app_id` 15368 is GitHub Actions, so only a workflow job can satisfy the check.)

"main only through a release PR" cannot be enforced by branch protection on a user-owned repo
(push restrictions are organisation-only), so it remains a process rule (`release/*` → `main`).

## Scanner coverage gaps (honest limits)

- semgrep (cso helper) returned malformed output, so there was no SAST pass beyond golangci-lint's gosec.
- The cso helper's history scan exceeded its output limit, so history was not scanned for secrets.
  GitGuardian on earlier PRs flagged only the local placeholders and the test fixtures listed above.
- Mobile was assessed only for CI and artifact handling. The app itself was gated in P0-31.
