# QA report: Reminders (M9-12), 2026-10-15

| | |
|---|---|
| Scope | `module_reminders` as merged on `dev` @ `a91dda3`: purge #161, migrations #163, calendar/DueDate #164, devices/settings #165, FCM adapter #166, worker job #167, mobile #168, flag #169. Spec [MODULE-reminders.md](../../docs/project-specs/MODULE-reminders.md) |
| Branch | `feature/reminders-qa-fixes` (from `origin/dev` @ `a91dda3`) |
| Device walk | **Done** on the Android emulator (`emulator-5554`, visible window), against this run's own API, worker and database (§1) |
| Blocked (owner prerequisites, not faked) | **Real push delivery on Android and iOS** (no Firebase project, OB9): the OS permission prompt, the "denied" notice, device registration from the app, a notification on the lock screen, the tap-to-open deep link. **iOS** entirely (Xcode, OB6) |
| Verdict | **Pass. No app defect found**; nothing to fix, so this branch carries only this report. Three observations for the owners (§5) |

## 1. Environment

Docker Desktop is down, so the run used its **own** stack, separate from every other agent's (same method as [the WH-13 report](qa-report-withholdings-2026-10-14.md) §1):

- Homebrew PostgreSQL 18.6, a fresh cluster in the session scratchpad on `127.0.0.1:5502` (TCP only), database `cc_remqa`. Roles from `infra/local/postgres/bootstrap/roles.sql`; every password and key (auth, RFC, `PUSH_TOKEN_ENC_KEY`, `PUSH_TOKEN_INDEX_PEPPER`) generated at run time into a 0600 file in the scratchpad, never in the repo.
- `cmd/migrate up` as `cc_owner`, `cmd/flags apply-preset demo_full` (10 of 10 ON), `cmd/api` as `cc_app` on `:3102`, `cmd/worker` as `cc_worker` (no `FCM_PROJECT_ID`: "push disabled", both jobs registered, leadership acquired).
- The dev APK (profile, arm64) built with `API_BASE_URL=http://10.0.2.2:3102`, reinstalled clean on the emulator (the coordinator confirmed the emulator was free; the app left in the foreground was from finished walks).
- **Worker with the fake sender.** A throwaway `qaharness` build-tagged test in `cmd/worker` (deleted after the run, not committed) built the reminder job exactly as `newReminderJob` / `newJobs` wire it (cc_worker pool, the API's push keys, the flags, the holidays, the monthly Estimator), with `clock.NewFixed(<instant>)` and `pushtest.Sender`, against `cc_remqa`. N instances, each with **its own connection pool**, start on a barrier and run R times.
- A copy `cc_remqa_copy` (`pg_dump` / `pg_restore`, owners kept) for the destructive checks, and a second API on `:3103` against it.

**Wipe**: the database was created empty for this run, so nothing had to be wiped. No other database, container or process was touched.

## 2. Created through the app (no SQL, no seed)

One account, `qa-rem@…`, signed up in *Crear cuenta* (code from the dev email log). In the app: **Ana Honorarios** (PF 612 from January 2026) with a September income of 50,000.00 + IVA to the public; **Taller Sur** (PF 626 from January 2026, RFC check-digit warning accepted) with a September income of 80,000.00 + IVA; an October expense (stationery) for Taller Sur, then deleted from the month list (*Gasto eliminado / Deshacer*).

Workarounds, the smallest available (the app cannot do these without Firebase, OB9):

- **Turning reminders on**: with push unavailable the opt-in switch does nothing by design (`TurnOnOutcome.unavailable`). Settings were enabled once with `PUT /v1/reminder-settings` as the same user; every later change was made on the screen.
- **Devices**: two installations (android + ios) registered with `POST /v1/devices` as that user, with push tokens generated at run time.
- **Read-only**: the trial was ended with one `UPDATE entitlements` (operator SQL, as in M5-29).
- Separate API-only probe accounts for the device-claim and validation checks.

## 3. Results

### Settings screen on the device (screenshots `rem-*.png` in `.gstack/qa-reports/screenshots/`, local only)

| Check | Result |
|---|---|
| Never saved: `GET` answers the defaults (off, 14, due date on, 10:00, **amounts off**) and writes no row | Pass (API); the screen shows the switch off (`rem-03`) |
| Push unavailable: "Notificaciones no disponibles en esta versión de la app." above the switch; the switch does not move, no OS prompt, no request | Pass (`rem-04`); no permission prompt at launch either |
| Autosave per change: due-date toggle off ⇒ polite "Guardado.", server holds `on_due_date: false` | Pass (`rem-06`) |
| A change that leaves nothing to send ("Sin recordatorio anticipado" with due date off) ⇒ refused locally, "Elija al menos un recordatorio…", nothing sent, server unchanged | Pass (`rem-07`) |
| **Revert**: a save that fails (this run's API stopped) ⇒ the control goes back and "No se pudo guardar el cambio; se restauró el valor anterior…" | Pass (`rem-08`); API restarted after |
| Send time with the platform picker (text-entry mode, 07:30) ⇒ "07:30 (hora del centro de México)", saved | Pass (`rem-09`, `rem-10`) |
| Amounts toggle: label carries the lock-screen consequence; saved `show_amounts: true` | Pass (`rem-10`) |
| `module_tax_engine` OFF ⇒ amounts toggle **hidden**, privacy note kept | Pass (`rem-11`) |
| "Próximos vencimientos" on 15 Oct 2026: **lunes, 19 de octubre** (period 2026-09, "Se recorre por fin de semana"), 17 Nov, 17 Dec, and "Fechas sujetas a confirmación." | Pass (`rem-03`) |
| Opt-out from the switch ⇒ saved `enabled: false`; the next due reminder sends nothing | Pass (`rem-15`) |
| Logout with no registered installation ⇒ back to *Iniciar sesión*, no device call, no error | Pass |

### Calendar (API, live)

| Check | Result |
|---|---|
| `year=2026`: 12 entries, period 2026-09 due **2026-10-19** `moved_reason: weekend` (nominal 10-17); 2026-04 → 05-18, 2026-12 → 2027-01-18 weekend; 2026-03 stays **Friday 04-17** (OQ-R1); every entry `subject_to_confirmation: true`; `Cache-Control: private, max-age=3600` | Pass |
| `year=2027`: December due 2028-01-17 (weekends-only) and the API logs `event=reminders_calendar_unseeded, year=2028` | Pass |
| `year=2028` / `2025` ⇒ `422 calendar_unavailable`; `abc`, `1999`, missing ⇒ `400`; no bearer ⇒ `401` | Pass |
| **Holiday move** (copy only): a test row on 2026-11-17 ⇒ calendar `2026-11-18 moved_reason: holiday`, both taxpayers' estimate `due_date` 2026-11-18 (main DB: 11-17), and the job sends the due reminder on 18 Nov and nothing on 17 Nov: one rule, one place | Pass |

### Settings and devices (API, live)

| Check | Result |
|---|---|
| `PUT` without `early_day` ⇒ `400 required` (never read as null or 14); 0 / 17 ⇒ `400`; `24:00`, `9:00` ⇒ `400 pattern`; unknown field ⇒ `400`; on with nothing selected ⇒ `422 no_reminder_selected`; off with nothing selected ⇒ `200`. `Cache-Control: no-store` on every answer | Pass |
| `registerDevice`: new ⇒ `201`, same id ⇒ `200` (`last_seen_at` refreshed), bad platform / no token ⇒ `400` | Pass |
| **Claim (US-R6)**: B registers A's token under a new installation id ⇒ `201`, same body as any create (no oracle); the row now belongs to B, A has none; A's `DELETE` of its old id ⇒ `404` | Pass |
| B deletes A's device ⇒ `404`, A's row intact; unknown id ⇒ `404`; malformed ⇒ `400` | Pass |
| The push token never appears in the API log (full and partial search) | Pass |
| 11th device evicts the least recently seen | Not re-run live (the `devices_writes` burst is 10); pinned by `migrations/reminders_integration_test.go` |

### Worker: exactly-once with the fake sender (US-R2, US-R5, US-R7)

| Run (fixed clock, CDMX) | Result |
|---|---|
| **2 instances × 2 runs**, 14 Oct 2026 07:31 (early, period 2026-09) | **1 send**, 1 `notification_log` row (`sent`, attempts 1, devices 2/0); the second instance's first run reports `not_claimed: 1` (the race was real); second runs send nothing |
| Body with amounts | "Ana Honorarios: estimado pendiente · Taller Sur: $14,000 estimado. Cifras aproximadas." Taller Sur's 14,000 equals the app (IVA 12,800 + ISR 1,200, RESICO 1.5 %); title "…vence el lunes 19 de octubre"; deep link `controlcontable://estimate?period=2026-09`; collapse id 64 hex |
| **8 instances × 3 runs**, 19 Oct 2026 07:30 (due), one device answered `PushTokenInvalid` | **1 send** ("Hoy vence tu pago mensual estimado"), 1 row `sent` with devices 1 ok / 1 failed, `error_class: unregistered`; **that device row deleted**, the other kept, `devices_unregistered: 1` |
| 17 Jan 2027 (nominal day, Sunday) / 18 Jan 07:29 (before send time) / 18 Jan 23:50 (catch-up) | nothing / nothing / one due reminder for 2026-12 |
| 15 Feb 2027 (the early day 14 missed) | nothing (skipped, never late) |
| Engine OFF, `show_amounts: true`, 14 Nov 2026 | names only: "Revisa tu estimado de octubre para: Ana Honorarios, Taller Sur." |
| `module_reminders` OFF (and implied OFF by `module_taxpayers` OFF) at a due instant, 2 instances | no send, no row written; devices, settings and log rows kept |
| Final row count after all runs | one row per user + period + kind + channel; no duplicates |

### Flags (live)

| Check | Result |
|---|---|
| `module_reminders` OFF: GET/PUT settings, calendar (seeded and unseeded year), POST device, DELETE own and unknown device, malformed PUT ⇒ all `404 not_found`; absent from `GET /v1/flags` | Pass (8/8) |
| `module_taxpayers` OFF ⇒ reminders implied OFF: routes `404`, flags list only `module_auth`, `module_subscriptions` | Pass |
| App after relaunch with the flag OFF: no *Recordatorios* entry in the app bar | Pass (`rem-12`) |
| Records purge runs with `module_reminders` OFF (copy): purge `expenses 1`, `notification_log 1`; the reminder job logged "skipped: module off" | Pass |

### Purge on the copy (M9-08 + the 13-month log retention)

Rows made through the app, aged on `cc_remqa_copy` only: the deleted expense to 31 days, Ana's income soft-deleted 29 days ago, one log row created 14 months ago and one 12 months 27 days ago. A content-hash snapshot of incomes, expenses, log, devices, settings and taxpayers before the run, compared after:

- **Gone**: exactly the 31-day expense and the 14-month log row. Everything else, including the 29-day income and the 12-month row, hash-identical.
- Second run: every target `0` (idempotent).

### Read-only user (US-R9, §4)

Trial ended (*Modo lectura…* on Home, `rem-13`): settings GET/PUT `200`, device register `201` / unregister `204`, calendar `200`, no `402`. On the screen a change autosaves with "Guardado." (`rem-14`). The job still sends, **with amounts** (copy, 19 Oct: "…Taller Sur: $14,000 estimado…").

### Account deletion (US-R8, copy)

`DELETE /v1/me` with the password ⇒ `204`; that user's devices 1 → 0, settings 1 → 0, log rows 5 → 0, user gone (cascade).

## 4. Fixes

None: no defect found. Every behaviour above that is not already pinned by a CI suite was checked live; the ones the walk relied on most are already pinned (`TestReminderJob_EndToEnd` two-instance split brain + UNREGISTERED, `TestClaim_ExactlyOnceAndLease`, `TestWorkerJobs_remindersFollowTheFlagThePurgeAlwaysRuns`, `TestPurgeJob_RunIsIdempotentAgainstTheDatabase`, the read-only contract test in `reminders_preferences_contract_integration_test.go`, the device cap in `migrations/reminders_integration_test.go`).

## 5. Observations (not fixed; for the owners)

- **O1, "estimado pendiente" is dropped when no taxpayer has a figure.** With `show_amounts` on, if every listed taxpayer is incomplete or zero, the body falls back to "Revisa tu estimado de octubre para: Ana Honorarios, Taller Sur." although Ana's October estimate is `incomplete` (live, 17 Nov). It is deliberate and pinned (`TestBuildNotification_AmountsModeWithoutAnyAmountFallsBackToNames`: never "Cifras aproximadas" alone), but §2.5 reads as if the pending mark should still show. Spec wording or copy decision for PM.
- **O2, with push unavailable the opt-in switch gives no feedback of its own.** A tap does nothing (the notice above explains why). Fine for builds without Firebase; disappears with OB9.
- **O3, a deployment without FCM consumes reminders.** The real worker without `FCM_PROJECT_ID` records a due reminder as `failed / push_disabled` (permanent, by design in M9-06; read from `deliveryTally.wholeCall`, not exercised live: no reminder fell due on the real clock during the run), so enabling FCM later never re-sends that day's reminder. Worth one line in the deploy runbook: configure FCM before enabling `module_reminders`.
- The reminder deep link is a field of the push payload routed on tap, not an OS intent: `adb am start -d controlcontable://…` opens the app without routing. As designed (MODULE-reminders §2.5, mobile README); the tap path waits for OB9.

## 6. Suites

| Suite | Where | Result |
|---|---|---|
| `go test ./...` (backend unit) | local | Pass (every package) |
| `flutter test test/features/reminders` | local | Pass: 90 tests |
| Integration (testcontainers) | CI | see the PR checks (Docker is down locally) |

## 7. Teardown

This run's API processes (`:3102`, `:3103`), its worker and its PostgreSQL cluster (`:5502`, data directory in the session scratchpad) were stopped by this run; the `qaharness` test file was deleted. Nothing else was stopped. The dev app stays installed on the emulator, signed in with the QA account (local data only).
