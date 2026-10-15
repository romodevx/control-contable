# Module M9: Reminders

| | |
|---|---|
| Flag | `module_reminders` (OFF by default, seeded by the baseline migration; requires `module_taxpayers`, **not** `module_tax_engine`). Preset: `demo_full` only (OQ-R8). See [Flags, prerequisites and presets](#flags-prerequisites-and-presets) |
| Spec | [MODULE-reminders.md](../project-specs/MODULE-reminders.md); legal research: [CFF art. 12 and holidays, §A.4.1](../research/tax-regimes-mx-2026.md) |
| Design | [ADR-0007 Amendment 1](../adr/0007-fiscal-shared-kernel.md#amendment-1-m9-01-2026-10-15-the-due-date-rule) (the due-date rule moves into the fiscal shared kernel), ADR-0001 O8 (worker, `notification_log`), ADR-0005 (rate limits) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml) **0.14.0**, tag `Reminders`, `x-module: module_reminders` (5 operations, frozen by M9-01 in 0.13.0; 0.14.0 added the optional `device_id` to `logout`). Reference generated from it: [API reference](#api-reference) |
| Status | **Built**: M9-01..M9-14 done. QA M9-12 pass, tech-lead M9-13 **PASS**, security M9-14 **PASS** ([Gates](#gates)). **Live push not verified** (no Firebase project, OB9; iOS, OB6) |
| Operations | [Enable reminders in a shared environment](../runbooks/reminders.md) (Firebase, APNs, secrets, deploy order, yearly holidays) · [records purge](../runbooks/records-purge.md) (13-month log retention) |
| Code | `backend/internal/kernel/fiscal/duedate.go` (the due-date rule), `backend/internal/reminders/` (`domain`, `usecase`, `adapter/http`, `infra/{fcm,holidays,jobstore,pushtoken,store}`), `backend/cmd/api/reminders.go`, `backend/cmd/worker/{reminders,push}.go`, migrations `20261021100000_holidays.sql`, `20261021101000_reminders.sql`, `20261022100000_reminders_job_reads.sql`, `20261022101000_notification_log_purge.sql`, `mobile/lib/features/reminders/` |

## What the user sees

Reminders are **opt-in** and **informational**: one push per user before each monthly due
date, listing every active taxpayer. The copy always says *estimado* and never claims a filing
obligation the app cannot know.

The entry is the bell in Home's app bar, tooltip **"Recordatorios"** (`/recordatorios`), shown
only while `module_reminders` is effectively ON.

| Control | Default | Behaviour |
|---|---|---|
| **Recibir recordatorios de pago** | off | Turning it on asks the OS permission **first** (never at launch). Granted → the device registers and the setting saves. Denied → "Las notificaciones están bloqueadas para Control Contable…" with **"Abrir ajustes del sistema"**; the app never re-prompts. A build without Firebase shows "Notificaciones no disponibles en esta versión de la app." and the switch does not move |
| **Recordatorio anticipado** | "El día 14 del mes de pago" | Days 1..16, or "Sin recordatorio anticipado". A nominal calendar day: sent even on a weekend (OQ-R4) |
| **Avisarme el día del vencimiento** | on | On the due date **after** any move (below) |
| **Hora del aviso** | "10:00 (hora del centro de México)" | Platform time picker ("Cambiar"). America/Mexico_City for everyone; per-user zones are out of scope |
| **Mostrar montos en las notificaciones (se verán en la pantalla de bloqueo)** | **off** | Hidden while `module_tax_engine` is OFF. See [Privacy](#privacy) |

- **Autosave per change**, no save button: the control moves at once and a polite
  **"Guardado."** confirms it. A failed save puts the control back with "No se pudo guardar el
  cambio; se restauró el valor anterior. Inténtelo de nuevo." A change that would leave nothing to
  send (no early reminder and the due-date one off) is refused on the device: "Elija al menos un
  recordatorio: el anticipado o el del día del vencimiento." (the server answers the same case
  with `422 no_reminder_selected`).
- **Próximos vencimientos**: the next three due dates from `getTaxCalendar`, each "Pago mensual de
  {mes}", with "Se recorre por fin de semana" or "Se recorre por día inhábil" when moved, and
  "Fechas sujetas a confirmación." under the list (OQ-R1, below).
- The privacy note under the controls: "Las notificaciones se ven en la pantalla de bloqueo.
  Muestran los nombres de sus contribuyentes, nunca su RFC; los montos sólo si usted lo elige."
- Read-only users (trial over, no plan) keep everything: settings, devices and reminders are
  account preferences, not records, so they never answer `402` (OQ-R2).

### The notification

| Case | Title | Body |
|---|---|---|
| Early reminder | "Tu pago mensual estimado vence el lunes 19 de octubre" | "Revisa tu estimado de septiembre para: Ana Honorarios, Taller Sur." |
| Due date | "Hoy vence tu pago mensual estimado" | (same body) |
| With amounts (opted in) | (same titles) | "Ana Honorarios: estimado pendiente · Taller Sur: $14,000 estimado. Cifras aproximadas." |

- Taxpayers by **`display_name`** only, up to 3, then "y N más", in the switcher's order; only the
  non-archived ones whose due date is that day (a 621-only taxpayer appears in its bimester's
  month). Nobody due → nothing is sent (`skipped`).
- Tapping it opens the estimate of that period (`controlcontable://estimate?period=YYYY-MM`,
  carried in the push payload); Home when `module_tax_engine` is OFF at tap time.
- Never in a payload: an RFC, an id, a regime code, a push token.

## The due date: one rule, one place

A monthly period `YYYY-MM` is due on **day 17 of the following month**; a Saturday, a Sunday or a
seeded holiday moves it to the next day that is none of those (CFF art. 12 ¶1 and ¶5). A
taxpayer whose every regime is the bimonthly 621 follows its bimester.

The rule lives **once**, in the fiscal shared kernel (`fiscal.MonthlyDueDate`,
`fiscal.NextBusinessDay`, `fiscal.PaysBimonthly`, ADR-0007 Amendment 1). The estimate's
`due_date` (M5), the annual deadlines, History, the dashboard, `getTaxCalendar` and the reminder
job all call it, so the date on a notification and the date on the estimate screen cannot drift.
`internal/archtest/weekend_test.go` fails on a `time.Saturday` outside the kernel;
`TestMonthlyDueDateGolden` pins the 24 periods 2025-12..2027-11 and
`TestNextBusinessDayAnnualDeadlinesGolden` the annual deadlines.

- **Holidays** come from the `holidays` / `holiday_years` tables (migration `20261021100000`),
  seeded for **2026 and 2027** with their sources. The API reads them **once at boot**; the worker
  reads them on every run. A new or corrected year is a migration, never a runtime write.
- **Unseeded year, fail closed**: `getTaxCalendar` answers `422 calendar_unavailable`. When the
  year is seeded but the next is not, December's entry (due in January) is computed
  **weekends-only** and the `reminders_calendar_unseeded` alert fires; the job does the same.
  Operations seed each next year in December
  ([runbook §6](../runbooks/reminders.md#6-every-december-seed-next-years-holidays)).
- **Sujeto a confirmación (OQ-R1).** CFF 12 ¶5 lets a bank payment due on a Friday be made the
  next business day. The app does **not** apply it: it keeps the earlier Friday (2026-04-17,
  2026-07-17, 2027-09-17, 2027-12-17) until the owner or a CPA confirms, and every calendar entry
  carries `subject_to_confirmation: true` ("Fechas sujetas a confirmación."). Also not applied: the
  RFC-digit grace days and SAT vacation periods (business days for payments). Holy Thursday and
  Friday (RMF 2.1.6) are stored as moving the date ⚠️; no 2026–2027 day 17 falls on them.

## Exactly once

The worker job `reminders` runs every **5 minutes** on the leader instance (each run cut at 4
minutes). For each user due now, in that user's own transaction:

1. **Claim**: insert a `notification_log` row keyed **UNIQUE (user, period, kind, channel)**
   (`kind` `early` | `due`, `channel` `push`). No row back → another worker owns it → stop. The
   key is the **period**, not the date, so a holiday correction cannot cause a second reminder.
2. **Send** to the user's devices (FCM `SendEach`, 30 s deadline), then **record** the outcome in a
   second short transaction: `sent`, `pending` (transient error, retried) or `failed`.
3. **Retry after a crash**: a `pending` row older than the **10-minute lease**, under 3 attempts and
   still within the day, is re-claimed atomically (`attempts + 1`). The retry carries the same
   **collapse id** (HMAC of user, period and kind), so the phone replaces rather than stacks it.
4. **Never late**: a reminder is sent from its send time until 23:59 that day; a day the worker
   missed entirely is skipped (OQ-R5), and its stale `pending` row becomes `failed / expired`.

Leader election has a split-brain window of about 15 s; the UNIQUE key, not the scheduler,
decides who sends. QA M9-12 ran 2 instances × 2 runs and 8 instances × 3 runs on separate pools:
one send and one row each time. Pinned by `TestReminderJob_EndToEnd`,
`TestClaim_ExactlyOnceAndLease` and `TestReminderJob_PushDisabledClaimsNothingAndLaterSends`.

**A worker without push consumes nothing** (M9-13). With no `FCM_PROJECT_ID` the job returns
before claiming: no row is written, one `event=reminders_push_disabled` warning is logged, and a
worker started with FCM later the same day still sends that day's reminders.

A device that FCM answers `UNREGISTERED` for, or `INVALID_ARGUMENT` naming `message.token`, is
deleted in the record transaction; the user's other devices still get the reminder. Any other
`INVALID_ARGUMENT` keeps the device (M9-14 F2).

## Privacy

- **Amounts are opt-in.** A figure appears only when **all** hold: "Mostrar montos…" is on,
  `module_tax_engine` is effectively ON, and that taxpayer's estimate for the period is complete
  with a total above zero. Otherwise the name only, or "estimado pendiente" when the estimate is
  incomplete. An estimator error never blocks the reminder: it goes without amounts.
- **Lock screen.** The OS shows title and body on the lock screen; that is why the toggle's label
  states the consequence and the default is off. Names are shown with amounts off (OQ-R3): the
  amounts are the sensitive part.
- **Push tokens** are sealed at rest (AES-256-GCM under `PUSH_TOKEN_ENC_KEY`, AAD bound to the
  installation and the user) with an HMAC blind index (`PUSH_TOKEN_INDEX_PEPPER`) for the claim.
  They are never returned by the API, never logged (redacted under every print path on both
  sides) and never granted to `cc_app` as columns.
- **Shared phone (token claim).** A token identifies the install, not the person. When user B
  registers an installation id or token that user A holds, `claim_push_token` moves the row to B
  atomically and answers B exactly as for any new device: no oracle.
- **Logout** removes the device. The app sends `DELETE /v1/devices/{id}` before revoking the
  session; if that fails (offline), it keeps its "registered" marker and sends the id as
  `logout`'s `device_id`, and the server deletes the caller's row in the same transaction as the
  revocation (0.14.0, M9-14 F1). If even that is lost, the next sign-in on the phone claims the
  leftover registration and deletes it unless the new user registers it (reminders off, or no OS
  permission). Accepted residual: a phone nobody signs in to after an offline sign-out keeps A's
  row until the next sign-in.
- **Opt-out** ("Recibir recordatorios de pago" off) unregisters the device and stops every send.
  **Account deletion** cascades: devices, settings and log rows go with the user.
- **Retention**: `notification_log` rows **13 months** (they hold no token or PII beyond the user
  id); soft-deleted incomes, expenses and personal deductions **30 days**. Both by the platform
  [records purge](../runbooks/records-purge.md), which runs whatever the flags say.
- **Processors**: Google (FCM) and Apple (APNs) receive the notification text; both are in the
  PRD §8 processor list.

## Flags, prerequisites and presets

| Flag OFF | Effect |
|---|---|
| `module_reminders` | The five routes answer `404 not_found` before any input check, a malformed request included; absent from `GET /v1/flags`; the bell, the screen and the permission prompt are gone; the job is a no-op. Devices, settings and log rows are **kept**: turning it back on resumes with the saved settings and catches up no past day |
| `module_taxpayers` | Reminders implied OFF (prerequisite, `flags.Prerequisites`): same `404`, entry hidden |
| `module_tax_engine` | Reminders still work, **names only**; the amounts toggle is hidden (the saved choice is kept) |

- `logout`'s `device_id` is **not** gated by `module_reminders` (a sign-out must always clean up).
- The records purge is a platform job, never behind this flag.
- Presets: **`demo_full` only**; `demo_core` and `demo_monetization` leave it out
  (`reminders_preset_smoke_integration_test.go`, `reminders_preset_flag_test.dart`).
- Budgets (ADR-0005), per user: `devices_writes` 30/h (burst 10) for register and unregister,
  `reminder_settings_writes` 60/h (burst 20), `tax_calendar_reads` 120/h (burst 60); spent →
  `429 rate_limited` with `Retry-After`. Settings and device answers are `Cache-Control:
  no-store`; the calendar's `200` is `private, max-age=3600`.
- **Enable in a shared environment** only when every item of the
  [runbook's enable checklist](../runbooks/reminders.md#7-enable-module_reminders) holds: both gates
  passed (done, 2026-10-15), the Firebase project and APNs key exist (OB9, OB7), the secrets are in
  the secret manager, the worker runs with push, the year's holidays are seeded, and live push was
  checked on a real Android and iOS device.

```sh
make -C backend flags ARGS="enable module_reminders"      # also needs module_taxpayers ON
make -C backend flags ARGS="apply-preset demo_full"       # local demo with every module
```

## API reference

The five operations (`registerDevice`, `unregisterDevice`, `getReminderSettings`,
`putReminderSettings`, `getTaxCalendar`) and the `logout` `device_id`, their schemas, parameters
and examples live only in the contract. Render them instead of copying them here:

```sh
make api-docs             # build/api-docs/index.html, rendered from docs/api/openapi.yaml (tag Reminders)
make api-error-tables     # regenerate the error table below after a contract change
make mock                 # Prism mock of the contract, for the app without a backend
```

Reading notes the reference states but that are easy to miss:

- `registerDevice`'s `id` is the app's **installation id** (UUIDv7, once per install), not a
  record id, so it never answers `409`: same caller and id → `200`, anything else → `201`. At most
  10 devices per user; the 11th evicts the least recently seen.
- `getReminderSettings` answers the defaults until the first save, without writing a row.
  `putReminderSettings` replaces the whole object; `early_day` is required and nullable (`null` =
  no early reminder), never read as 14 when missing.
- `getTaxCalendar?year=Y` lists the **periods** `Y-01..Y-12`, so its last entry is due in January
  of `Y+1`.

When this page and `openapi.yaml` disagree, the contract wins and this page is the bug.

## Errors

Every error uses the single envelope `ErrorResponse`. `details[].code` carries the rule
(`no_reminder_selected` on `body`, `calendar_unavailable` on `year`). A push token is never
echoed in an error.

The table below is **generated from the contract** by `make api-error-tables`; do not edit
it by hand. CI (`make api-error-tables-check`) fails when it differs from `openapi.yaml`.

<!-- BEGIN GENERATED error-table module_reminders -->
| Operation | Endpoint | HTTP | `error.code` | When |
|---|---|---|---|---|
| `registerDevice` | `POST /v1/devices` | 400 | `validation_failed` | Input does not match the contract (the operation's body, `deviceId` or `year`): a malformed body, an unknown or missing field (`early_day` included), a bad `platform` or `push_token`, `early_day` outside 1..16, `send_time` not `HH:MM`, a malformed `deviceId`, or a missing or out-of-range `year`. |
| `registerDevice` | `POST /v1/devices` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `registerDevice` | `POST /v1/devices` | 404 | `not_found` | `module_reminders` is effectively OFF (`module_reminders` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose); or, for `unregisterDevice`, the device does not exist, was already removed or belongs to another user. With the module ON a malformed input is `400`. |
| `registerDevice` | `POST /v1/devices` | 429 | `rate_limited` | Rate limit exceeded. |
| `registerDevice` | `POST /v1/devices` | 500 | `internal_error` | Unexpected server error. |
| `unregisterDevice` | `DELETE /v1/devices/{deviceId}` | 400 | `validation_failed` | Input does not match the contract (the operation's body, `deviceId` or `year`): a malformed body, an unknown or missing field (`early_day` included), a bad `platform` or `push_token`, `early_day` outside 1..16, `send_time` not `HH:MM`, a malformed `deviceId`, or a missing or out-of-range `year`. |
| `unregisterDevice` | `DELETE /v1/devices/{deviceId}` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `unregisterDevice` | `DELETE /v1/devices/{deviceId}` | 404 | `not_found` | `module_reminders` is effectively OFF (`module_reminders` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose); or, for `unregisterDevice`, the device does not exist, was already removed or belongs to another user. With the module ON a malformed input is `400`. |
| `unregisterDevice` | `DELETE /v1/devices/{deviceId}` | 429 | `rate_limited` | Rate limit exceeded. |
| `unregisterDevice` | `DELETE /v1/devices/{deviceId}` | 500 | `internal_error` | Unexpected server error. |
| `getReminderSettings` | `GET /v1/reminder-settings` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getReminderSettings` | `GET /v1/reminder-settings` | 404 | `not_found` | `module_reminders` is effectively OFF (`module_reminders` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose); or, for `unregisterDevice`, the device does not exist, was already removed or belongs to another user. With the module ON a malformed input is `400`. |
| `getReminderSettings` | `GET /v1/reminder-settings` | 429 | `rate_limited` | Rate limit exceeded. |
| `getReminderSettings` | `GET /v1/reminder-settings` | 500 | `internal_error` | Unexpected server error. |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 400 | `validation_failed` | Input does not match the contract (the operation's body, `deviceId` or `year`): a malformed body, an unknown or missing field (`early_day` included), a bad `platform` or `push_token`, `early_day` outside 1..16, `send_time` not `HH:MM`, a malformed `deviceId`, or a missing or out-of-range `year`. |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 404 | `not_found` | `module_reminders` is effectively OFF (`module_reminders` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose); or, for `unregisterDevice`, the device does not exist, was already removed or belongs to another user. With the module ON a malformed input is `400`. |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 422 | `business_rule_violation` | Well-formed but selects no reminder: `enabled: true` with `early_day: null` and `on_due_date: false` (`no_reminder_selected` on `body`). |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 429 | `rate_limited` | Rate limit exceeded. |
| `putReminderSettings` | `PUT /v1/reminder-settings` | 500 | `internal_error` | Unexpected server error. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 400 | `validation_failed` | Input does not match the contract (the operation's body, `deviceId` or `year`): a malformed body, an unknown or missing field (`early_day` included), a bad `platform` or `push_token`, `early_day` outside 1..16, `send_time` not `HH:MM`, a malformed `deviceId`, or a missing or out-of-range `year`. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 401 | `unauthorized` | Missing, invalid or expired credentials. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 404 | `not_found` | `module_reminders` is effectively OFF (`module_reminders` or `module_taxpayers` disabled), which answers `404` before any input check, a malformed request included (indistinguishable on purpose); or, for `unregisterDevice`, the device does not exist, was already removed or belongs to another user. With the module ON a malformed input is `400`. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 422 | `business_rule_violation` | The year's holidays are not seeded, so its due dates cannot be computed (`calendar_unavailable` on `year`). Operations seed each next year in December, after the RMF. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 429 | `rate_limited` | Rate limit exceeded. |
| `getTaxCalendar` | `GET /v1/tax-calendar` | 500 | `internal_error` | Unexpected server error. |
<!-- END GENERATED error-table module_reminders -->

## Gates

| Gate | Verdict | Notes |
|---|---|---|
| M9-12 QA ([report](../../.gstack/qa-reports/qa-report-reminders-2026-10-15.md), [PR #170](https://github.com/romodevx/control-contable/pull/170)) | **Pass**, no app defect | Android emulator against its own empty stack; settings screen, autosave and revert, "Próximos vencimientos" (19 Oct moved by the weekend), engine OFF hides amounts, flag OFF hides the entry. Live API: calendar, holiday move on a copy, claim with no oracle, validation. Worker with a fake sender: exactly-once at 2 × 2 and 8 × 3 instances, `UNREGISTERED` cleanup, never late, amounts only with opt-in. Purge (30 days, 13 months) and account deletion on a copy. **Live push blocked** (OB9), iOS (OB6) |
| M9-13 tech lead ([review on PR #170](https://github.com/romodevx/control-contable/pull/170)) | **PASS** after two fixes | QA O3: a worker without FCM consumed the day's reminders; it now skips before claiming (`PushSwitch`). One bimonthly rule: `fiscal.PaysBimonthly` / `fiscal.RegimeRIF` replace two copies. Canon checklist clean (TDD, ports at the consumer, no `time.Now`, bounded `errgroup`, goleak) |
| M9-14 security ([verdict on PR #170](https://github.com/romodevx/control-contable/pull/170), follow-ups [PR #172](https://github.com/romodevx/control-contable/pull/172)) | **PASS**, no blocking finding | Amounts gate, payload copy, token lifecycle and claim, no token in logs, FCM credentials from the secret source, RLS ENABLE + FORCE with column grants, retention, flag `404`, budgets, Firebase files git-ignored. F1 (logout deletes the device), F2 (FCM attribution, payload cap) and F3 (OSV scan of `pubspec.lock`) fixed in #172, plus one more blocker found there (a leftover registration was kept for a new user without OS permission) |

## Decisions

Full text in [MODULE-reminders §12](../project-specs/MODULE-reminders.md#12-open-questions--recommended-defaults-coordinator-defaults-until-the-owner-says-otherwise).
These are coordinator defaults until the owner says otherwise.

| # | Decision | Effect |
|---|---|---|
| OQ-R1 | Friday extension (CFF 12 ¶5) **not** applied | A Friday day 17 stays Friday; every date "sujeta a confirmación". **Owner/CPA to confirm** |
| OQ-R2 | Reminders work for read-only users | No `402` on any reminders route |
| OQ-R3 | Names on the lock screen with amounts off | Amounts are the opt-in part |
| OQ-R4 | Early reminder on a weekend | Sent anyway (nominal day) |
| OQ-R5 | Worker down all day | Skipped, never sent late |
| OQ-R6 | Devices per user | 10, least recently seen evicted |
| OQ-R7 | Push tokens at rest | Encrypted (RFC envelope) + HMAC index |
| OQ-R8 | Presets | `demo_full` only |
| M9-14 | `claim_push_token` owned by the shared `cc_auth_definer` | Accepted; split into a dedicated `cc_device_definer` when a second definer touches `devices` or a `cc_auth_definer` function uses dynamic SQL |

## Known limits and follow-ups

Tracked in the [PLAN M9 follow-ups](../project-specs/PLAN-control-contable.md#m9-follow-ups).
The ones a user or an operator can notice:

- **Live push never verified (OB9, OB6, OB7).** The FCM sender and the Flutter push code are
  built and tested against fakes. Owed on real devices: the OS permission prompt and the denied
  notice, device registration from the app, a notification on the lock screen, the tap-to-open
  deep link, Android and iOS. The [runbook §8](../runbooks/reminders.md#8-first-live-push-check)
  lists the walk.
- **OQ-R1 open**: the Friday extension needs the owner's or a CPA's answer; until then four dates
  a year may be one business day earlier than the law allows.
- **Holidays are seeded through 2027.** 2028 must be seeded in December 2027 (the December 2027
  period is due in January 2028 and is computed weekends-only until then); Holy Thursday/Friday
  2027 are projected until the RMF 2027 is read.
- **QA O1 (copy, PM)**: with amounts on, when no listed taxpayer has a figure the body falls back
  to the names-only text, without "estimado pendiente". Deliberate and pinned
  (`TestBuildNotification_AmountsModeWithoutAnyAmountFallsBackToNames`); PM to confirm and amend
  spec §2.5.
- **QA O2**: with push unavailable the opt-in switch gives no feedback of its own (the notice
  above it explains why). Disappears with OB9.
- **No key rotation path for `PUSH_TOKEN_ENC_KEY`** (M9-05 N2): one keyring entry. Rotating
  `PUSH_TOKEN_INDEX_PEPPER` also changes collapse ids (one possible duplicate on that day).
- **Per-user time zones** are out of scope: Quintana Roo and Baja get the CDMX send time.
- Security nits (#172, non-blocking): `deleteDeviceSQL` relies on RLS alone (add
  `AND user_id = app_user_id()`); the OSV scan runs only on mobile changes (add a schedule).
- Out of scope: email/SMS channels (the `channel` column leaves room), reminders for the annual
  return or other obligations, the RFC-digit grace days, marketing pushes.
