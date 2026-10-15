# MODULE-reminders — M9 Reminders (`module_reminders`)

Status: **Spec (stage 1)** · Owner: PM · Date: 2026-10-15 · MoSCoW: **Should** · Depends on: M1 (users,
sessions, account deletion), M2 (taxpayers, `display_name`, archive), M5-22 (`Estimator`, due date),
P0-27 (worker: ticker + `pg_try_advisory_lock` leader election), P0 (flags, access policy, rate limits)
Sources: PRD §6 (US-R1), §7 (read-only), §8 (LFPDPPP, push provider as processor) · MODULES.md M9 · PLAN
M9-01..M9-15, OB7, OB9, SEC FU1 · ADR-0001 O8 (scheduling, `notification_log`), Firebase Admin row ·
ADR-0005 (rate limits) · ADR-0007 (fiscal shared kernel) · [MODULE-tax-engine.md](MODULE-tax-engine.md)
§2.10 (due date, holiday gap) · research
[tax-regimes-mx-2026.md §A.4.1](../research/tax-regimes-mx-2026.md) (CFF art. 12 verbatim, holidays
2026–2027) · contract scaffold `registerDevice` / `unregisterDevice` / `getReminderSettings` /
`putReminderSettings` / `getTaxCalendar` in `docs/api/openapi.yaml`, **frozen by M9-01 (0.13.0)** with the changes of §10.

## 1. Purpose and scope

The user gets **one push notification** before each monthly deadline, **listing all of their taxpayers**,
so nobody misses day 17. Reminders are **opt-in** and **informational**: the copy always says
*estimado* and never claims a filing obligation the app cannot know.

- **Device registration**: the app registers its FCM token after the OS permission is granted; logout,
  opt-out and account deletion remove it.
- **Settings**: on/off, early reminder day (default 14), reminder on the due date (default on), send time
  (default 10:00 America/Mexico_City), **show amounts** (default **OFF**, lock-screen privacy).
- **Tax calendar**: the monthly due dates of a year, day 17 moved to the next business day (§2.1).
- **Reminder job** (worker): at the user's send time on the early day and on the due date, one push per
  user, exactly once per user + period + kind + channel (§2.4).

| Op | Endpoint | Contract |
|---|---|---|
| registerDevice | `POST /v1/devices` | scaffold; changes §10 |
| unregisterDevice | `DELETE /v1/devices/{deviceId}` | scaffold; changes §10 |
| getReminderSettings | `GET /v1/reminder-settings` | scaffold; changes §10 |
| putReminderSettings | `PUT /v1/reminder-settings` | scaffold; changes §10 |
| getTaxCalendar | `GET /v1/tax-calendar?year=YYYY` | scaffold; changes §10 |

**Not in this module, same milestone:** the soft-delete **purge job** (M9-08) and the
`rate_limit_buckets` purge (ADR-0005) share the worker but are **platform jobs**: they run **regardless
of `module_reminders`** and are specified in their own tasks (M9-08 must also purge `personal_deductions`,
SEC FU1). Reminders own no purge logic.

## 2. Domain rules

### 2.1 Tax calendar — one rule, one place

- **Due date** of a monthly period `YYYY-MM` = day **17 of the following month**; if that day is a
  Saturday, Sunday or a `holidays` row with `moves_due_date = true`, move to the next day that is none
  of those (CFF art. 12 ¶1 and ¶5, research §A.4.1). For a taxpayer whose every regime is the bimonthly
  621, the due month follows the bimester (MODULE-tax-engine §2.10).
- **Not applied** (each ⚠️ in research §A.4.1): the RFC-digit grace days; **SAT vacation periods** (CFF 12
  ¶2 makes them business days for payments, so they never move day 17); the **Friday extension** (CFF 12
  ¶5 last sentence, bank payments; the app keeps the earlier Friday until a CPA confirms, OQ-R1).
  Holy Thursday/Friday (RMF 2.1.6) are stored with `moves_due_date = true` ⚠️; no 2026–2027 day 17
  touches them.
- **Single source of truth.** Today the rule lives in **two** copies: `taxengine/usecase/duedate.go`
  (monthly day 17, plus the 621 bimester offset) and `taxengine/usecase/annual/estimate.go`
  `businessDay` (annual deadlines), both asking a `HolidayCalendar` port whose composition-root adapter
  (`noHolidays`, `cmd/api`) answers "never" (the M5 "holiday gap"). M9-04 **moves the rule into the
  fiscal shared kernel** (`internal/kernel/fiscal`, ADR-0007) as pure functions with no I/O, no
  `time.Now`:
  - `NextBusinessDay(d Date, h HolidaySet) (Date, MovedReason, error)` — the CFF 12 move, bounded
    (today's `maxDueShift` 14 → error), used by the monthly **and** the annual deadline;
  - `MonthlyDueDate(p Period, bimonthly bool, h HolidaySet) (nominal, due Date, MovedReason, error)` —
    day 17 of the following month (after the bimester when `bimonthly`), then `NextBusinessDay`;
  - `HolidaySet` — an immutable value built from the `holidays` rows of the seeded years, with
    `Covers(year) bool`; outside a covered year it answers weekends only (the caller decides whether that
    is acceptable, §2.1 next bullet).
  The `holidays` adapter (M9-03) replaces `noHolidays`; the estimate (monthly and annual), History (it
  reads the estimate's `due_date`), the calendar endpoint and the reminder job all call these functions.
  No other copy of the rule survives (an archtest/grep guard in M9-04 fails on `time.Saturday` outside the
  kernel). M9-04 **amends ADR-0007 §3** (the kernel's closed list of contents) to add the due-date
  functions: more than one bounded context reads them and they must agree. A golden table pins it: the
  24 periods **2025-12..2027-11** (every due date in the seeded years 2026–2027, research §A.4.1), plus
  the annual deadlines of fiscal years 2025–2026. No 2026–2027 date changes when the adapter swaps
  (checked in research §A.4.1), so the engine golden figures stay identical and `EngineVersion` is not
  bumped; if any estimate golden changes, M9-04 bumps it per the engine version policy.
- **Unseeded years fail closed.** The calendar covers only years whose holidays are seeded
  (`holiday_years`, §2.6). `getTaxCalendar?year=Y` lists the **periods** `Y-01..Y-12` (so its last entry
  is due in January of `Y+1`); `Y` unseeded → `422 calendar_unavailable`. When `Y` is seeded but `Y+1` is
  not, the December entry is computed weekends-only and emits the metric below (it is already
  `subject_to_confirmation`). The job never
  guesses: for a due date in an unseeded year it falls back to weekends-only **and** emits the
  `reminders_calendar_unseeded` metric/alert (ops seeds the next year each December after the RMF; a
  PLAN follow-up). MODULE-tax-engine §2.10's "beyond 2028" blocker stays.
- Time zone: **America/Mexico_City** (no DST since 2022). Users in Quintana Roo or Baja get the CDMX time;
  per-user zones are out of scope (§11).

### 2.2 Settings

| Field | Default | Rule |
|---|---|---|
| `enabled` | `false` | Opt-in. Turning it on in the app triggers the OS permission prompt first |
| `early_day` | `14` | 1–16 or `null` (no early reminder). Nominal calendar day; sent even on a weekend |
| `on_due_date` | `true` | Reminder on the (moved) due date |
| `send_time` | `10:00` | `HH:MM`, America/Mexico_City; delivered within one job tick (§2.4) |
| `show_amounts` | `false` | Amounts on the notification (§2.5). Privacy opt-in, never defaulted on |

- `GET` returns the defaults when the user never saved (no row is created by a read).
- `PUT` replaces the whole object (`additionalProperties: false`). `enabled: true` with `early_day: null`
  and `on_due_date: false` is a `422 no_reminder_selected` (it would be "on" and send nothing).
- Changing settings never resends: the log (§2.4) already holds what was sent.

### 2.3 Devices and token claim

- A device row = (`id` = the app's **installation id**, client-generated UUIDv7, stable per install),
  `platform`, the FCM token, `app_version`, `last_seen_at`. A user may have several devices (fair use: **10**;
  the 11th evicts the least recently seen).
- **`registerDevice` is an upsert by installation id and by token**: same user + same id → refresh token /
  `last_seen_at`, `200`; new → `201`.
- **Claim semantics (token moves between users on one device).** A push token identifies the *install*, not
  the person. When user B registers a token (or installation id) currently held by user A, the row
  **moves to B**: A loses it atomically, so A's reminders never reach B's phone. RLS forbids B from seeing
  A's row, so the move runs in the SECURITY DEFINER function `claim_push_token(p_user, p_installation_id,
  p_token_hmac, …)` (M9-02): narrow, `search_path` pinned, deletes any row matching the id **or** the token
  hash on **any other row** (another user, or the same user's stale installation after a reinstall), then
  upserts for `p_user`, all in one statement set; executable only by `cc_app`.
  It returns nothing about A (no oracle: B gets the same `201` whether or not the token had an owner).
- **Unregister**: the app calls `DELETE /v1/devices/{id}` on logout (before revoking the session) and on
  opt-out; foreign or unknown id → `404`. Best effort client-side; the claim on the next login and the
  UNREGISTERED cleanup (§2.4) cover a missed call. Account deletion cascades.
- **Token at rest**: stored **encrypted** with the same application-level envelope as the RFC, plus a keyed
  HMAC (`token_hmac`, UNIQUE) for the claim lookup. Never logged, never returned by the API (`Device` has
  no token).

### 2.4 Reminder job — idempotent, exactly-once

Worker job `reminders`, tick **every 5 minutes**, leader only (P0-27).

1. **Flag check per run**: `module_reminders` effectively OFF (or a prerequisite OFF, §3) → the run is a no-op
   (no rows written, nothing sent).
2. **Select due users** through SECURITY DEFINER `due_reminder_users(p_now, p_limit)` (M9-02): users with
   `enabled`, ≥ 1 device, ≥ 1 non-archived taxpayer, and a reminder **kind** (`early` | `due`) whose local
   send moment (date + `send_time`) is `≤ now` and **within today** (catch-up window: until 23:59 local of that
   day; a reminder missed for a whole day is **skipped**, never sent late), and with **no
   `notification_log` row** for (user, period, kind, channel). Returns user ids only, in batches.
3. **Per user, own transaction** with `SET LOCAL app.current_user_id`, bounded `errgroup` (concurrency 8):
   - **Claim**: `INSERT INTO notification_log (user_id, period, kind, channel, status='pending',
     attempts=1, claimed_at=now(), …) ON CONFLICT (user_id, period, kind, channel) DO NOTHING RETURNING id`.
     No row returned → someone else owns it → stop. Commit.
   - **Build** the message (§2.5), **send** via the FCM sender port (`SendEach` to the user's devices,
     ≤ 500) with a **send deadline of 30 s**, far below the 10-minute lease of step 5, so a live sender is
     never reclaimed mid-send. Then **record** in a second short transaction (no token, no PII):
     - at least one device accepted → `status = sent`, `sent_at`, `devices_ok/failed`;
     - only transient errors (FCM unavailable/internal, deadline) and `attempts < 3` → stays `pending`
       (`error_class` set); step 5 retries it;
     - otherwise (`attempts = 3`, or every device permanently rejected) → `status = failed`.
   - **UNREGISTERED / invalid-argument token** (`messaging.IsUnregistered`) → delete that device in the
     record transaction; the user's other devices are unaffected.
4. **Why the dedupe key is (user, period, kind, channel)**: leader election has a split-brain window of
   about **15 s** (`DefaultHeartbeat` 10 s + election timeout 5 s), so two workers can run the same tick;
   the UNIQUE constraint, not the scheduler, decides who sends (ADR-0001 O8). It keys on **period**, not
   on `due_date`, so a holidays correction that moves a date cannot produce a second reminder. `channel`
   is `push` in v1 (room for email without a migration).
5. **Crash or transient failure after the claim**: `due_reminder_users` also returns users with a `pending`
   row whose `claimed_at` is older than **10 minutes** (the lease), `attempts < 3`, and whose reminder is
   still **within today** (same window as step 2; past it the row becomes `failed`, `error_class =
   expired`, never sent late). The retry **re-claims atomically**: `UPDATE notification_log SET
   attempts = attempts + 1, claimed_at = now() WHERE id = $1 AND status = 'pending' AND claimed_at <
   now() - interval '10 minutes' RETURNING id`; no row → another worker re-claimed it → stop. This is
   at-least-once at the transport. Visible duplicates are prevented on the device by a **collapse id** = hex of HMAC(user,
   period, kind), truncated to 64 hex chars (the `apns-collapse-id` limit): `android.collapse_key` /
   `apns-collapse-id` / `tag`, so a resend replaces the earlier
   notification instead of stacking. Net effect for the user: exactly one notification.
6. **Which taxpayers**: the user's non-archived taxpayers whose due date for that period is the reminder's
   date (621-only taxpayers appear only in their bimester's month). None → the job writes `status =
   skipped` (so it is not re-evaluated) and sends nothing.

### 2.5 Notification content (Spanish, always "estimado")

Payload carries display text and a deep link only: **no RFC, no ids, no regime codes**. Taxpayers are
listed by `display_name` (up to 3, then "y N más"), sorted as the switcher.

| Case | Title | Body |
|---|---|---|
| Early, no amounts | `Tu pago mensual estimado vence el {lunes 18 de mayo}` | `Revisa tu estimado de {abril} para: {Ana}, {Taller Sur}.` |
| Due date, no amounts | `Hoy vence tu pago mensual estimado` | `Revisa tu estimado de {abril} para: {Ana}, {Taller Sur}.` |
| With amounts | (same titles) | `{Ana}: {$4,520} estimado · {Taller Sur}: {$12,300} estimado. Cifras aproximadas.` |

- **Amounts appear only when all hold**: `show_amounts = true`, `module_tax_engine` effectively ON, and the
  estimate for that taxpayer/period computes with status `ok` and a total > 0. Otherwise that taxpayer is
  listed by name only (`estimado pendiente` when the estimate is incomplete). An estimator error never
  blocks the reminder: it sends without amounts and logs the error class.
- Amounts are the `Estimator`'s rounded totals in MXN, formatted es-MX; the job never recomputes.
- Deep link: `controlcontable://estimate?period=YYYY-MM` → the dashboard/estimate for that period (the
  active taxpayer stays as is). With the engine OFF it opens Home.
- The OS shows the title/body on the lock screen; that is why amounts are opt-in and the toggle explains it:
  "Mostrar montos en las notificaciones (se verán en la pantalla de bloqueo)".

### 2.6 Data model (migrations M9-02, M9-03)

| Table | Scope | Columns (key ones) | RLS |
|---|---|---|---|
| `devices` | user | `id uuid PK` (installation id), `user_id FK ON DELETE CASCADE`, `platform`, `token_ciphertext`, `token_hmac UNIQUE`, `app_version`, `created_at`, `last_seen_at` | ENABLE + FORCE, per-op policies on `app.current_user_id` |
| `reminder_settings` | user | `user_id PK FK CASCADE`, the five fields of §2.2, `updated_at` | ENABLE + FORCE |
| `notification_log` | user | `id`, `user_id FK CASCADE`, `period`, `kind` (`early`/`due`), `channel` (`push`), `status` (`pending`/`sent`/`failed`/`skipped`), `attempts`, `claimed_at`, `sent_at`, `devices_ok`, `devices_failed`, `error_class`; **UNIQUE (user_id, period, kind, channel)** | ENABLE + FORCE |
| `holidays` | global | `day date PK`, `name`, `basis` (`cff_12` / `rmf_2_1_6`), `moves_due_date bool`, `source_url`, `verified_on` | none (not user data); `cc_app`/`cc_worker` SELECT only |
| `holiday_years` | global | `year PK`, `seeded_on`, `source_note` | none; SELECT only |

- Seed: the 20 rows of research §A.4.1 (2026, 2027) with their sources; `holiday_years` 2026, 2027.
- `notification_log` retention: **13 months**, purged by the platform purge job (M9-08 adds it to its list).
- `cc_worker` grants: EXECUTE on `due_reminder_users`; it never reads `devices`/`settings` outside a
  per-user transaction. Indexes: `notification_log (user_id, period)` is the UNIQUE; `devices (user_id)`.
- Isolation tests (mandatory): user B cannot read/write A's `devices`, `reminder_settings`,
  `notification_log`; `claim_push_token` moves the row and leaves B unable to read A's other devices.

### 2.7 Rate limits (ADR-0005)

`devices_writes` **30/h, burst 10** per user (register + unregister); `reminder_settings_writes`
**60/h, burst 20**; `tax_calendar_reads` **120/h, burst 60**. Reads of settings share the general read
budget. Spent → `429 rate_limited` + `Retry-After`.

## 3. Flags

- `module_reminders`: default **OFF**, registered by migration (M9-11). Preset **`demo_full`** includes it;
  `demo_core` / `demo_monetization` do not.
- **Prerequisite**: `"module_reminders": {"module_taxpayers"}` (the reminder lists taxpayers). Not
  `module_tax_engine`: amounts are optional and degrade (§2.5).
- OFF (or implied OFF) → all five routes `404`; absent from `GET /v1/flags`; the app hides the settings
  entry and never asks for push permission; the job is a no-op. Rows (devices, settings, log) are **kept**
  and cascade on deletion; turning the flag back on resumes with the saved settings (no reminder is
  "caught up" for days that passed, §2.4 step 2).
- Platform purge jobs (M9-08, rate-limit buckets) are **not** behind this flag.
- Suites: (`reminders`, `taxpayers`, `tax_engine`) ∈ {ON/ON/ON, ON/ON/OFF (no amounts), ON/OFF/·
  (implied OFF), OFF/ON/ON}; preset smoke for `demo_full` plus a reminders-only fixture preset.

## 4. Read-only and paywall (coordinator default, OQ-R2)

By analogy with the owner's decisions (read-only after trial, informational app, "history stays
visible", data never deleted for non-payment): **reminders keep working for read-only users.**
Registering/unregistering devices and reading/saving reminder settings are **account preferences, not
records**: no `402`. Amounts follow the same rule as the dashboard (read-only users still see their
estimates). Rationale: a reminder is a safety feature and a re-engagement path to the paywall; cutting it
would punish lapsed users without protecting any paid capability.

## 5. User stories (Given / When / Then)

### US-R1 Opt in
- **Given** reminders were never set, **When** I open Ajustes → Recordatorios, **Then** I see `enabled: off`
  with the defaults (14, due date on, 10:00, amounts off). **When** I turn it on, **Then** the OS
  permission prompt appears; on grant the device registers and settings save; on deny, settings are not
  enabled and the screen explains how to allow notifications in system settings.

### US-R2 Early and due-date reminders
- **Given** reminders on with defaults and 2 active taxpayers, **When** it is 14 May 2026 10:00 CDMX,
  **Then** I receive exactly one push "Tu pago mensual estimado vence el lunes 18 de mayo" listing both
  names. **When** it is 18 May 2026 10:00, **Then** I receive exactly one "Hoy vence…" push.

### US-R3 Weekend and holiday move
- **Given** period 2026-09 (17 Oct 2026 is a Saturday), **Then** the calendar and the reminder say Monday
  19 Oct 2026, `moved_reason: weekend`. **Given** a test holiday row on a day 17, **Then** the date moves
  past it, `moved_reason: holiday`. **Given** period 2026-03 (17 Apr 2026, Friday), **Then** the date
  stays Friday (OQ-R1). **Given** `year=2028` (unseeded), **Then** `422 calendar_unavailable`; **Given**
  `year=2027`, **Then** the December entry (due 2028-01-17) is weekends-only and
  `reminders_calendar_unseeded` fires. **Given** the swap from `noHolidays` to the table, **Then** every
  estimate golden is unchanged.

### US-R4 Amounts only with opt-in
- **Given** `show_amounts: false`, **Then** no figure appears in any payload. **Given** `show_amounts:
  true` and the engine ON, **Then** each taxpayer with a computed estimate shows "$X estimado". **Given**
  the engine OFF, **Then** names only, and the toggle is hidden in the app.

### US-R5 Exactly once
- **Given** two worker instances in the split-brain window, **When** both process the same user/period/kind,
  **Then** one `notification_log` row exists and the fake FCM sender received one send. **Given** a crash
  after the claim, **Then** after the 10-minute lease the retry carries the same collapse id, and two
  workers racing to re-claim it produce one retry (`attempts` = 2). **Given** a `pending` row whose day has
  passed, **Then** it becomes `failed` / `expired` and nothing is sent. **Given** the job runs again later
  that day after a `sent`, **Then** nothing is sent.

### US-R6 Shared phone (token claim)
- **Given** A registered this phone, **When** A logs out without unregistering and B logs in on it,
  **Then** the device row belongs to B, A's next reminder does not target it, and B's response does not
  reveal A existed.

### US-R7 Stale token
- **Given** FCM answers UNREGISTERED for one of my devices, **Then** that device row is deleted and my
  other devices still receive the reminder.

### US-R8 Logout, opt-out, deletion
- **When** I log out, **Then** the app unregisters the device first. **When** I turn reminders off,
  **Then** no further push is sent. **When** I delete my account, **Then** devices, settings and log rows
  are gone (cascade).

### US-R9 Read-only user
- **Given** my trial ended without a plan, **Then** I can still change reminder settings and receive
  reminders (§4).

### US-R10 Flags and isolation
- **Given** `module_reminders` OFF (or `module_taxpayers` OFF), **Then** the five routes `404`, the entry is
  hidden and no push is sent. **Given** user B calls `DELETE /v1/devices/{A's id}`, **Then** `404` and A's
  row is intact (RLS test).

## 6. Errors → ErrorCode (single envelope)

| Case | HTTP | `error.code` | `details[].code` |
|---|---|---|---|
| Malformed body / id / `year`, unknown field, bad `send_time`, `early_day` out of 1–16 | 400 | `validation_failed` | |
| No / invalid / expired token | 401 | `unauthorized` | |
| Flag OFF (or implied), foreign or unknown device id | 404 | `not_found` | |
| `enabled` with no reminder selected | 422 | `business_rule_violation` | `no_reminder_selected` |
| `year` without seeded holidays | 422 | `business_rule_violation` | `calendar_unavailable` |
| Budget spent | 429 | `rate_limited` | |
| Anything else | 500 | `internal_error` | |

## 7. Accessibility and UX

- Settings screen: one column, toggles with visible labels ≥ 44×44 px, the time picker uses the platform
  picker, "Día del recordatorio anticipado" as a stepper/select (1–16 or "Sin recordatorio anticipado").
  Autosave per change with a polite live-region "Guardado"; no primary button needed.
- The amounts toggle carries its consequence in its label (§2.5), not in a tooltip.
- A "Próximos vencimientos" list (from `getTaxCalendar`) shows the next 3 due dates, with "(se recorre por
  fin de semana / día inhábil)" when moved and the existing "sujeto a confirmación" note.
- Permission denied state: text + button "Abrir ajustes del sistema"; never re-prompts on every visit.

## 8. Security and privacy (gate M9-14, LFPDPPP)

- **Lock-screen exposure**: amounts default OFF and require `show_amounts`; no RFC, ids or regimes ever in
  a payload; names only (`display_name`).
- **Token claim**: definer function narrow, `search_path` pinned, no oracle (§2.3); tokens encrypted at rest,
  HMAC-indexed, never logged or returned.
- **Processors**: Google (FCM) and Apple (APNs) receive the notification text; they are already in the
  PRD §8 processor list and the integral privacy notice; the M9 gate checks the notice names push
  notifications and the "show amounts" choice.
- **Consent**: the OS permission + `enabled` is the user's opt-in; `show_amounts` is a separate, explicit
  choice. ARCO: the account export (when built) includes settings and devices (platform, dates; not the
  token); cancellation = account deletion cascade.
- **Retention**: `notification_log` 13 months; devices deleted on UNREGISTERED, claim, logout or
  deletion.
- Server-side flag enforcement; RLS ENABLE + FORCE on the three user tables; `Cache-Control: no-store` on
  device and settings routes.
- Gate checks: flag matrix, `404` indistinguishability, RLS isolation (incl. claim), payload snapshot tests
  with `show_amounts` off/on and engine off, no token in logs.

## 9. Observability

Metrics: `reminders_sent_total{kind}`, `reminders_failed_total{error_class}`, `reminders_skipped_total`,
`devices_unregistered_total`, `reminders_calendar_unseeded` (alert), job duration. Logs carry user id
(UUID) and period only.

## 10. Contract changes for M9-01 (architect)

1. `DeviceCreate.id`: document as the installation id (UUIDv7, client-generated, stable per install);
   `push_token` `writeOnly`. `registerDevice`: keep `201`/`200`; add `429`.
2. `unregisterDevice`: add `429`.
3. `ReminderSettings`: keep the shape but make `early_day` **required** (nullable): with PUT replacing the
   whole object, an omitted `early_day` must be a `400`, never read as `null` (no early reminder) or as
   the default 14. Add the `422 no_reminder_selected` response on `putReminderSettings` and `429`;
   document "defaults if never saved" on GET; `additionalProperties: false`.
4. `getTaxCalendar`: document `year` as the **period** year (December's entry is due in `year + 1`);
   add `422 calendar_unavailable` and `429`; `Cache-Control: private, max-age=3600`
   instead of `no-store` (same data for everyone, behind auth); `TaxCalendarEntry` gains
   `subject_to_confirmation: boolean` (true for every entry while OQ-R1 is open) and `obligation` enum
   documents `monthly_isr_iva` only in v1.
5. Append to the details enum: `no_reminder_selected`, `calendar_unavailable` (append-only).
6. `x-module: module_reminders` on all five; examples for a moved (weekend) entry, a first-time GET,
   and the `422`s.

**As frozen (M9-01, 0.13.0).** All six items landed. Decisions the spec left open:
`FieldError.code` is an open string, so the two new detail codes are listed in the operations'
descriptions and in `docs/api/CONTRACT.md` (there is no detail-code enum to append to);
`registerDevice` is the one create that never answers `409` (the installation id is not a record
id; a claim is not a conflict); `Device` gained `app_version` (nullable) and `last_seen_at`;
`TaxObligation` (append-only) and `TaxCalendarMovedReason` are named schemas; `entries` holds
exactly 12; the request schemas carry no `default` keywords (the defaults are documented, and
answered by GET); `getReminderSettings` has no budget of its own (edge limiter only). The
`module_reminders` → `module_taxpayers` prerequisite (§3) is declared in
`internal/platform/flags/prerequisites.go`.

## 11. Out of scope

- Email/SMS/in-app inbox channels (the `channel` column leaves room), per-user time zones, reminders for the
  annual return or other obligations (DIOT, constancias), RFC-digit grace days, the Friday extension until
  OQ-R1, reminders tied to records not entered ("you have no incomes this month"), notifying estimate
  changes (History owns the marker), marketing pushes.

## 12. Open questions — recommended defaults (coordinator defaults until the owner says otherwise)

| # | Question | Recommended default |
|---|---|---|
| OQ-R1 | CFF 12 ¶5 Friday extension (bank payments): move a Friday day 17 to Monday? | **No** (keep Friday, the earlier safe date) and mark entries "sujeto a confirmación"; **needs owner/CPA confirmation** (affects 2026-04-17, 2026-07-17, 2027-09-17, 2027-12-17) |
| OQ-R2 | Do reminders work for read-only users? | **Yes**, settings and devices are preferences, not records (§4) |
| OQ-R3 | Taxpayer names on the lock screen when `show_amounts` is off | **Yes**, names only (PRD "lists them"); amounts are the sensitive part |
| OQ-R4 | Early reminder on a weekend | **Send anyway** (nominal day) |
| OQ-R5 | Missed reminder (worker down all day) | **Skip**, never send late |
| OQ-R6 | Devices per user | **10**, LRU eviction |
| OQ-R7 | Encrypt push tokens at rest | **Yes**, RFC envelope + HMAC |
| OQ-R8 | Presets | **`demo_full` only** |
