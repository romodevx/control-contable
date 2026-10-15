# Module M9: Reminders

| | |
|---|---|
| Flag | `module_reminders` (OFF by default; implied OFF when `module_taxpayers` is OFF) |
| Spec | [MODULE-reminders.md](../project-specs/MODULE-reminders.md) |
| Contract | [`docs/api/openapi.yaml`](../api/openapi.yaml), tag `Reminders`, `x-module: module_reminders` (frozen by M9-01, 0.13.0) |
| Status | Contract frozen; implementation in progress (PLAN M9-02..M9-14). The full page lands with the module docs task (M9-15) |

## Contract highlights

- `registerDevice` (`POST /v1/devices`): `id` is the app's **installation id** (UUIDv7,
  generated once per install), not a record id, so it never answers `409`. Same caller and
  same id → `200` (token refreshed); anything else → `201`, and any other row with this id
  or this token is removed in the same transaction (shared phone, reinstall). The answer is
  the same whether or not the token had another owner. `push_token` is write-only and never
  returned; at most 10 devices per user, the 11th evicts the least recently seen.
- `unregisterDevice` (`DELETE /v1/devices/{deviceId}`): on logout and opt-out; another
  user's, unknown or already removed id → `404`, which the app treats as done.
- `getReminderSettings` / `putReminderSettings` (`/v1/reminder-settings`): GET answers the
  defaults until the first save (`enabled: false`, `early_day: 14`, `on_due_date: true`,
  `send_time: 10:00`, `show_amounts: false`) without writing a row. PUT replaces the whole
  closed object; `early_day` is required and nullable (`null` = no early reminder).
  `enabled: true` with no reminder selected → `422 no_reminder_selected`.
- `getTaxCalendar` (`GET /v1/tax-calendar?year=YYYY`): the 12 **periods** of `year`
  (December is due in January of `year + 1`), day 17 moved past weekends and seeded
  holidays, `moved_reason` `weekend` | `holiday` | `null`, and `subject_to_confirmation:
  true` while OQ-R1 (the Friday extension) is open. A year without seeded holidays →
  `422 calendar_unavailable`. Its `200` is `Cache-Control: private, max-age=3600`, the
  second exception to `no-store`; every other response is `no-store`.
- Flags: `module_reminders` or `module_taxpayers` OFF → all five routes `404`, before any
  input check. Open to read-only users (no `402`: preferences, not records). Budgets
  `devices_writes` 30/h (burst 10), `reminder_settings_writes` 60/h (burst 20),
  `tax_calendar_reads` 120/h (burst 60); spent → `429` with `Retry-After`.

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
