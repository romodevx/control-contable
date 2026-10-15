// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/device.dart';
import '../models/device_create.dart';
import '../models/fiscal_year.dart';
import '../models/installation_id.dart';
import '../models/reminder_settings.dart';
import '../models/tax_calendar.dart';

part 'reminders_client.g.dart';

@RestApi()
abstract class RemindersClient {
  factory RemindersClient(Dio dio, {String? baseUrl}) = _RemindersClient;

  /// Register or refresh this installation's push token.
  ///
  /// The app calls it after the OS notification permission is granted, after every login and whenever FCM rotates the token (MODULE-reminders §2.3). `id` is the **installation id**: a UUIDv7 the app generates once per install and keeps. It is not a record id, so the API-wide "same `id`, different body → `409`" rule does **not** apply: this operation never answers `409`.
  ///
  /// **Upsert with claim.** Same caller and same `id` → the token, `app_version` and `last_seen_at` are refreshed, `200`. Anything else → `201`: the row is written for the caller and, in the same transaction, every other row holding this `id` or this token is removed, whoever owns it (another user after a shared phone changed hands, or the caller's own stale installation after a reinstall). A push token therefore reaches exactly one user. The answer is the same whether or not the token had another owner (no oracle). A user keeps at most 10 devices: an 11th evicts the least recently seen.
  ///
  /// **Token.** `push_token` is write-only: stored encrypted, looked up by a keyed HMAC, never logged and never returned (`Device` has no token).
  ///
  /// **Access.** Allowed for a read-only user (an account preference, not a record: no `402`, no trial slot). Per-user budget `devices_writes`, 30 per hour (burst 10, ADR-0005), shared with `unregisterDevice`; spent → `429 rate_limited` with `Retry-After`.
  @POST('/v1/devices')
  Future<Device> registerDevice({@Body() required DeviceCreate body});

  /// Remove this installation's push token (logout, opt-out).
  ///
  /// The app calls it on logout (before revoking the session) and when the user turns reminders off. Removes the caller's row; the token stops receiving reminders at once. Another user's or an unknown installation id answers `404` (RLS), so does a second call: the app treats `404` as done. Best effort: a missed call is covered by the claim on the next `registerDevice` and by the FCM `UNREGISTERED` cleanup of the reminder job. Account deletion removes every device (cascade).
  ///
  /// **Access.** Allowed for a read-only user. Budget `devices_writes` (shared with `registerDevice`); spent → `429` with `Retry-After`.
  ///
  /// [deviceId] - The installation id the device was registered with.
  @DELETE('/v1/devices/{deviceId}')
  Future<void> unregisterDevice({
    @Path('deviceId') required InstallationId deviceId,
  });

  /// The caller's reminder preferences (defaults if never saved).
  ///
  /// A user who never saved answers the defaults (`enabled: false`, `early_day: 14`, `on_due_date: true`, `send_time: 10:00`, `show_amounts: false`) and no row is created by the read (MODULE-reminders §2.2).
  ///
  /// **Access.** Allowed for a read-only user. No per-user budget of its own (a one-row read); the edge limiter bounds it like every route, so `429` with `Retry-After` is still possible.
  @GET('/v1/reminder-settings')
  Future<ReminderSettings> getReminderSettings();

  /// Replace the caller's reminder preferences.
  ///
  /// Replaces the whole object (every field required, `additionalProperties: false`): an omitted `early_day` is `400 required`, never read as `null` or as the default 14. Last write wins; the same body again answers `200` unchanged. Saving never resends a reminder: the reminder log already holds what was sent (MODULE-reminders §2.2, §2.4). `enabled: true` needs a registered device to deliver anything, but saving does not check for one (the app registers the device first).
  ///
  /// **Access.** Allowed for a read-only user (no `402`). Per-user budget `reminder_settings_writes`, 60 per hour (burst 20, ADR-0005); spent → `429 rate_limited` with `Retry-After`.
  ///
  /// `FieldError.code` values are the JSON Schema keywords (`required`, `minimum`, `maximum`, `pattern`, `type`, `additionalProperties`). `422 business_rule_violation` details: `body` → `no_reminder_selected` (`enabled: true` with `early_day: null` and `on_due_date: false`: it would be on and send nothing).
  @PUT('/v1/reminder-settings')
  Future<ReminderSettings> putReminderSettings({
    @Body() required ReminderSettings body,
  });

  /// Monthly due dates of a year's periods (day 17 moved past weekends and holidays).
  ///
  /// One entry per **period** `year-01` .. `year-12`, in order, so December's entry is due in January of `year + 1` (MODULE-reminders §2.1). Due date = day 17 of the month after the period; on a Saturday, Sunday or a seeded holiday that moves due dates it moves to the next day that is none of those (CFF art. 12 ¶1, ¶5). Not applied: RFC-digit grace days, SAT vacation periods and the Friday extension (a Friday day 17 is kept, OQ-R1), hence `subject_to_confirmation: true` on every entry while OQ-R1 is open. The calendar is the same for every user and does not depend on the caller's taxpayers: a bimonthly (621) period is due on the entry of the bimester's second month. It is the same rule the estimates' `due_date` uses (fiscal shared kernel, ADR-0007 Amendment 1).
  ///
  /// **Coverage.** Only years whose holidays are seeded are answered; `year` unseeded → `422 calendar_unavailable`. When `year` is seeded and `year + 1` is not, December's entry is computed with weekends only (already `subject_to_confirmation`).
  ///
  /// **Caching.** The `200` is `Cache-Control: private, max-age=3600`, the second exception to `no-store`: the same public data for every user, behind authentication, so only the app's own cache may keep it, for an hour (a holiday correction shows within that window). Errors are `no-store`.
  ///
  /// **Access.** Allowed for a read-only user. Per-user budget `tax_calendar_reads`, 120 per hour (burst 60, ADR-0005); spent → `429 rate_limited` with `Retry-After`.
  ///
  /// `FieldError.code` values are the JSON Schema keywords on `year` (`required`, `type`, `minimum`, `maximum`). `422 business_rule_violation` details: `year` → `calendar_unavailable`.
  ///
  /// [year] - The year of the **periods** listed (December's entry is due in `year + 1`).
  @GET('/v1/tax-calendar')
  Future<TaxCalendar> getTaxCalendar({@Query('year') required FiscalYear year});
}
