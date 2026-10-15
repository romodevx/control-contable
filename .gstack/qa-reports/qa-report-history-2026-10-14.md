# QA report: M8 History (M8-07), 2026-10-14

| | |
|---|---|
| Scope | Module M8 History (`module_history`, MODULE-history.md), as merged on `dev` through #147 (M8-06). Checks: the month list per RFC and fiscal year, the "Cambió desde que lo viste" marker only when the shown figures change, flag OFF ⇒ `404` and no entry point, read-only users and archived taxpayers, persona física and persona moral |
| Branch | `feature/history-qa-fixes` (from `origin/dev` @ `2d760b4`), PR #149 |
| Device walk | **BLOCKED by the environment** (§1). Nothing was wiped and nothing was created through the UI |
| Automated evidence | Backend unit suite (local), backend integration + contract suites on real PostgreSQL (CI, testcontainers), the full mobile suite (local), the mobile client against the contract mock (Prism, local) |
| Verdict | **Pass at the automated level. No app defect found.** Two coverage gaps closed with new tests (§3). The device walk is still owed (§5) |

## 1. Environment: why the device walk is blocked

Earlier module QAs (M2-12, M3-M4, M5-29, M6-06 in `docs/reviews/`) ran the app on the Android emulator
against the local API on this run's own PostgreSQL in Docker (compose project per run).

- Docker Desktop is **down** and cannot be started from the terminal (`docker info` fails).
- No native PostgreSQL server: Homebrew has only `libpq` 18.6 (client tools, `initdb`/`pg_ctl`, but
  no `postgres` binary); no Postgres.app, Colima, Podman or OrbStack.
- Without a database the API cannot start, so there is no local stack to point the app at. The
  emulator (`emulator-5554`) was running but was **not used**: a walk against the stateless mock
  cannot show a marker appearing or clearing, and the emulator may be in use by another agent.
- Nothing was installed, no shared database was touched, no process was stopped except this run's own
  Prism mock (by PID).

Environment problems only; none of this is an app defect.

## 2. What ran, and the result

| Suite | Where | Result |
|---|---|---|
| `go test ./...` (backend unit, every package incl. `internal/history/*`, `cmd/api`) | local | Pass |
| `make test-integration` (unit + integration + contract vs `docs/api/openapi.yaml`, real PostgreSQL) | CI, PR #149 | Pass; `cmd/api` 82.2 %, `history/usecase` 100 %, `history/domain` 98.4 %, `history/adapter/http` 98.0 %, `history/infra` 81.0 % |
| `flutter analyze`, `flutter test` (1368 tests, 39 skipped: mock-gated and golden-gated) | local + CI | Pass |
| `test/integration/history_mock_test.dart` against Prism on `127.0.0.1:3041` (own port) | local | Pass (2/2): the contract example decodes newest first; `markHistoryViewed` accepts the closed body |
| Release build (prod, obfuscated) + artifact gate, codegen drift, govulncheck, gitleaks, GitGuardian | CI | Pass |

### What the suites prove, per check asked

| Check | Evidence | Result |
|---|---|---|
| List per RFC / month: newest first, October .. January for 2026, each month's totals equal the estimate screen's to the centavo, year totals = sum | `TestContract_history` "the current year…", `TestHistoryQA_noFalseMarks` (new), `history_screen_test.dart` | Pass |
| Months before the first regime: status `none`, no amount, "Sin régimen en este mes" | `TestHistoryQA_noFalseMarks` PM (new, real engine); mobile `history_screen_test.dart` | Pass |
| Marker only when the shown figures change: late income marks March and April (`records`); opening clears only that month; deleting the income clears April without opening it | `TestContract_history` "viewed, then a late income…" | Pass |
| Engine/parameter change that moved the figures → `engine`; an engine or input change with identical figures → no mark (OQ-H4) | `TestContract_history` "an engine or parameter change…" | Pass |
| Never-viewed and current month never marked | `TestContract_history`, `history_viewed_test.dart` (current month never posted) | Pass |
| **No false mark**: opening every past month, then re-reading, marks nothing (History's range estimate hashes each month as the single-month estimate the app renders) | `TestHistoryQA_noFalseMarks` (new), PF and PM | Pass |
| Flag OFF ⇒ both routes `404` before any input check, indistinguishable from an unknown taxpayer; engine OFF or taxpayers OFF ⇒ implied OFF; record flags change only `has_records`, never a figure | `TestContract_history` OFF, `TestContract_historyFlagMatrix`, `history_preset_smoke_integration_test.go` (only `demo_full` exposes it) | Pass |
| Flag OFF ⇒ no "Historial" entry, `/historial` unreachable, no `markHistoryViewed` post | `history_flag_test.dart` (dashboard Home **and**, new, placeholder Home), `history_viewed_test.dart` | Pass |
| Read-only user and archived taxpayer: History readable, marking works | `TestContract_history` "an archived taxpayer and a read-only user…" | Pass |
| RLS: user B cannot read/insert/update/delete A's `estimate_views`; B's history never shows A's views; foreign id → `404` | `history/infra` `TestViewStore_isolation`, `TestContract_history` "another user's…" | Pass |
| No amounts or fingerprints in logs; `Cache-Control: no-store`; own rate-limit budgets | `TestContract_history`, `TestContract_historyRateLimits` | Pass |
| Persona física | TX-NET-01 (612 + 606) in every contract suite | Pass |
| Persona moral | Before this QA: none for History. Now `TestHistoryQA_noFalseMarks` PM | Pass |

## 3. Tests added on this branch (coverage gaps, not defects)

| Test | Gap it closes |
|---|---|
| `backend/cmd/api/history_qa_integration_test.go` `TestHistoryQA_noFalseMarks` | No History test used a persona moral, a regime starting mid-year (months with status `none` from the real engine), or a cross-month figure (June's withheld IVA credited in July, WH-09). None checked, for every month, that History's range estimate and the single-month estimate the app posts agree on `figures_fingerprint`; if they ever diverge, opening a month marks it at once. PF (TX-NET-01) and PM (601 from April, PM-601-IVA-01's June records entered through the API): every month equals the estimate screen; opening every past month marks none; a late June expense marks June (`records`) and not the earlier months |
| `mobile/test/features/history/history_flag_test.dart` (2 cases) | The entry was only tested on the dashboard Home. With the dashboard OFF (placeholder Home) it must still follow `module_history` alone: hidden when OFF, shown and opening History when ON |

Both pinned existing behaviour, so they passed first. The mobile cases were shown able to fail (gating
the entry on `module_tax_engine` instead turns both OFF cases red). The backend test could not be
mutation-checked locally (it needs Docker); it ran green in CI.

## 4. Observations (not defects, for the owners)

- Every month row is tappable, including `none` months; the estimate screen for a month before the
  first regime loads (the new test reads it for January–March), so this is consistent, just low value.
- `.gstack/` is git-ignored; this report was added with `git add -f` because the task asks for it in the PR.

## 5. Owed (blocked by the environment)

The visible walk, as M6-06 ran it, once Docker or a native PostgreSQL is available: wipe the run's own
database, `cmd/flags apply-preset demo_full`, sign up and create a PF and a PM through the app, enter
records across months, open months from History and from the dashboard's "Ver detalle", add a late record
and watch the marker appear and clear, switch taxpayer, a read-only account, flags OFF (History, engine)
and a TalkBack pass on the month rows (§6 of the spec: row label order, focus return, polite announcement).
iOS not run (OB6, as in every module QA).

## 6. Teardown

Prism mock on `:3041` stopped by its PID. No database, container or app install was created.
