# QA report: M8 History (M8-07), 2026-10-14

| | |
|---|---|
| Scope | Module M8 History (`module_history`, MODULE-history.md), as merged on `dev` through #147 (M8-06). Checks: the month list per RFC and fiscal year, the "Cambió desde que lo viste" marker only when the shown figures change, flag OFF ⇒ `404` and no entry point, read-only users and archived taxpayers, persona física and persona moral |
| Branch | `feature/history-qa-fixes` (from `origin/dev` @ `2d760b4`), PR #149 |
| Device walk | **Done 2026-10-15** (§7) on the Android emulator against the run's own PostgreSQL and API, branch `feature/history-device-walk`. It was blocked on 2026-10-14 (§1) |
| Automated evidence | Backend unit suite (local), backend integration + contract suites on real PostgreSQL (CI, testcontainers), the full mobile suite (local), the mobile client against the contract mock (Prism, local) |
| Verdict | **Pass after two fixes.** The automated level passed on 2026-10-14 (§2, §3). The device walk found two stale-screen defects, both fixed with a regression test first (§7.4 F1, F2), and a third (F3) fixed in the tech-lead review through one tax-data revision signal that also replaced F2's per-screen reload. The TalkBack focus check could not be completed with adb (§7.5) |

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

## 5. Owed (blocked by the environment), done on 2026-10-15 except as §7.6 says

The visible walk, as M6-06 ran it, once Docker or a native PostgreSQL is available: wipe the run's own
database, `cmd/flags apply-preset demo_full`, sign up and create a PF and a PM through the app, enter
records across months, open months from History and from the dashboard's "Ver detalle", add a late record
and watch the marker appear and clear, switch taxpayer, a read-only account, flags OFF (History, engine)
and a TalkBack pass on the month rows (§6 of the spec: row label order, focus return, polite announcement).
iOS not run (OB6, as in every module QA).

## 6. Teardown

Prism mock on `:3041` stopped by its PID. No database, container or app install was created.

## 7. Device walk, 2026-10-15

### 7.1 Environment

Docker Desktop was still down. The run used a stack of its own, as the withholdings QA did, and shared none of it with the other agents:

- A new PostgreSQL 18.6 cluster from Homebrew `postgresql@18`, with its data directory in the session scratchpad. It listened on `127.0.0.1:5495` over TCP only and held one database, `cc_histqa`. The roles came from the repo's `infra/local/postgres/bootstrap/roles.sql`, with random passwords and keys generated at run time and never written to the repo.
- `cmd/migrate up` as `cc_owner`, then `cmd/flags apply-preset demo_full`, then `cmd/api` as `cc_app` on `:3095` (`APP_ENV=dev`; verification codes came from the dev email log).
- The dev APK was a profile build for arm64 with `API_BASE_URL=http://10.0.2.2:3095`. It was installed on `emulator-5554` after the ANN-23 and WH-07 agents confirmed they were done with the emulator. The old install was removed first, so the app started with no session.
- **Wipe:** the database was created empty for this run, so nothing needed wiping. Before the walk it held only the migrations' catalogs and flags.

### 7.2 Created through the app (no seed, no SQL for the data under test)

| Account | Taxpayer (onboarding) | Records entered in the forms |
|---|---|---|
| `qa-hist@…` (trial) | Ana Fisica, PF **612** from January 2026 | Incomes 30,000.00 (15 Jul), 45,000.00 (12 Aug), 20,000.00 (10 Sep) to the general public, 16 %, entered with *Guardar y crear otro* from History's empty state. Expenses: Software 5,000.00 (20 Aug) from Home and Papelería 1,000.00 (15 Oct) from History → *Ver gastos* |
| same | Empresa QA, PM **601** from **June** 2026, coeficiente de utilidad 0.2 | Incomes 100,000.00 (invoiced 10 Jul) and 60,000.00 (invoiced 5 Sep) to a company. Prior provisional payment for June = 0 under *Datos fiscales* |
| `qa-hist2@…` (trial) | Beto Resico, PF **626** from January 2026 | Income 40,000.00 (20 Sep), entered from History's empty state, to re-check F1 |

The only SQL was operator-side: the first account's trial was moved into the past so the account became read-only, because the app cannot age a trial.

### 7.3 Results (screenshots `hist-*.png` in `.gstack/qa-reports/screenshots/`, local only)

| Check | Result |
|---|---|
| Year list, PF 612: October .. January, newest first, one row per month. The year totals (35,994.59) equal the sum of the rows. Each opened month equals the estimate screen and the dashboard trend to the centavo (July 6,455.16, August 11,144.13, September 9,761.14) | Pass (`hist-03`, `hist-06`) |
| PM 601 from June: January–May read "Sin régimen en este mes" with no amount, June–October have figures, totals add up | Pass (`hist-14`). PM IVA 0 is correct: the incomes were invoiced and not yet collected |
| Opening July, August and September (History rows) → "Visto el 15 oct". The current month is never marked | Pass (`hist-07`) |
| Edit with no figure change (August income moved from the 12th to the 14th) → nothing marked | Pass (`hist-08`) |
| Edit that moves the figures (August 45,000 → 50,000) → August **and** September marked "Cambió desde que lo viste: agregaste o cambiaste registros o datos fiscales.", because 612 is year-to-date. July and October are not marked | Pass (`hist-09`) |
| Editing the amount back to 45,000 → both marks clear without opening either month | Pass (`hist-10`) |
| Delete the July income → July, August and September marked. *Deshacer* restores it, and the marks clear once History is read again | Pass after a pull-to-refresh (F3) (`hist-11`..`hist-13`) |
| A change in fiscal data (PM June prior payment = 0): July's status goes Incompleto → Completo with the same amount, so July is marked; August and September are not | Pass (`hist-16`) |
| Opening from the estimate screen's month picker (OQ-H7) clears July | Pass |
| Taxpayer switch PM ↔ PF: each taxpayer keeps its own rows and views, and no previous figures show | Pass |
| Read-only account (trial ended): Home shows "Modo lectura", History is readable, and opening June marks it "Visto" | Pass (`hist-19`, `hist-20`) |
| Archived taxpayer | Not reachable in the UI. Choosing an archived taxpayer opens its settings and never makes it active (M2 design). The API side is `TestContract_history` |
| `module_history` OFF (real bearer, own ids): `getHistory` and `markHistoryViewed` answer `404`; so do an unknown taxpayer, a bad `fiscal_year` and a request without auth. `/v1/flags` omits it. After a relaunch Home has no "Historial", and the estimate screen still works | Pass (`hist-22`) |
| `module_tax_engine` OFF with History ON: History is hidden (`flags list` shows `hidden`), both routes answer `404`, and there is no entry on Home | Pass (`hist-23`) |
| `module_incomes` and `module_expenses` OFF: the figures and views are unchanged, every row says "Sin registros en este mes", the "Ver …" links are hidden, and History does **not** fall into the empty state (M8-08 F1 on device) | Pass (`hist-24`) |
| Row semantics (uiautomator): each month is one button read as month, total, status, then marker | Pass |

### 7.4 Defects

| # | What the walk showed | Status |
|---|---|---|
| **F1** | History's empty state "Registrar ingreso" pushed the income form without reading the year again. After three incomes were saved, History still said "Aún no hay registros…" until it was pulled down (`hist-02`) | **Fixed** in `fix(mobile): re-read History after "Registrar ingreso" from the empty year`. The empty state now opens the form through the year view's own open-then-reload path. Regression test `history_screen_test.dart` "returning from \"Registrar ingreso\" reads the year again": red before the fix (one `getHistory` call), green after. Checked on the device with a new account (`hist-18`) |
| **F2** | Home pushed History without reloading when it returned. Records saved from inside History (its empty state, or *Ver ingresos/gastos* → form) left the dashboard saying "Aún no hay ingresos…" with the trial count unspent (`hist-04`) | **Fixed** first in `fix(mobile): read Home's dashboard again on return from History` (an `onReturn` reload on Home's History button, checked on the device, `hist-17`). The tech-lead review replaced that per-screen reload with the single path of F3 below: History's links are writes, so the writes move the signal and Home reads again only when something changed. Regression test `history_flag_test.dart` "a write made over History reads the dashboard again", plus "History opened and closed without a write reads nothing" |
| **F3** | The delete toast's *Deshacer* lives in the app-wide messenger and stays up across routes. Tapping it while History (or Home) is on top restores the record, but that screen is not read again, so the marks stay until a pull-to-refresh. The data is right | **Fixed** in the tech-lead review (`fix(mobile): one tax-data revision re-reads every derived figure`). `taxDataRevisionProvider` (`core/fiscal`) is bumped by every income and expense write (create, edit, delete, Undo) and by the fiscal-data save; `historyYearProvider`, `dashboardProvider`, `estimateProvider` and `breakdownProvider` watch it. Regression tests first: `tax_data_revision_test.dart` (red: revision 0, 4 reads instead of 8) and `history_screen_test.dart` "an Undo run while History is open reads the year again" (red: one `getHistory` call). Not yet re-checked on the device |

### 7.5 TalkBack

TalkBack (installed on the emulator) was turned on through `settings` and turned off again afterwards (`enabled_accessibility_services` back to `null`). adb's injected taps bypass touch exploration: a tap activates the row instead of focusing it, and TalkBack's speech is not logged. So the reading order was checked through the semantics tree (§7.3, last row), but **focus return to the row after a pop could not be shown**. After the back key, TalkBack focus sat on the app bar's back button (`hist-26`). Because the row never held TalkBack focus before it was opened, this is not evidence of a defect. A person still needs to do this pass with real gestures, together with the polite announcement.

### 7.6 Still owed

- The TalkBack pass with real gestures (§7.5), and iOS/VoiceOver (OB6).
- A device re-check of F2/F3 on the revision signal (the widget tests cover it).

### 7.7 Suites (this branch)

| Suite | Where | Result |
|---|---|---|
| `dart format`, `flutter analyze`, `flutter test` | local | Pass: 0 issues, 1384 tests, 39 skipped (mock- and golden-gated) |
| CI required checks | PR | see the PR |

### 7.8 Teardown

This run stopped only what it started: the API on `:3095` (by its PID) and the PostgreSQL cluster on `:5495` (`pg_ctl stop`; its data directory was in the session scratchpad). The flags were restored with `apply-preset demo_full` before the stop. The dev app stays installed on the emulator with the two QA accounts (local data only). TalkBack is off.
