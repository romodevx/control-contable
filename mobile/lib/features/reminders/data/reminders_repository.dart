import 'package:control_contable/api/generated/clients/reminders_client.dart';
import 'package:control_contable/api/generated/models/device_create.dart';
import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/api/generated/models/tax_calendar.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The reminder operations of M9 (`/v1/devices`, `/v1/reminder-settings`,
/// `/v1/tax-calendar`). Every method throws an `ApiFailure` when the
/// server refuses or cannot be reached.
abstract interface class RemindersRepository {
  /// Registers (or refreshes) this installation's push [token]
  /// (`POST /v1/devices`, an upsert with claim).
  Future<void> registerDevice({
    required String installationId,
    required DevicePlatform platform,
    required PushToken token,
    String? appVersion,
  });

  /// Removes this installation (`DELETE /v1/devices/{id}`).
  Future<void> unregisterDevice(String installationId);

  /// The user's preferences; the defaults if never saved.
  Future<ReminderSettings> getSettings();

  /// Replaces the user's preferences with [settings].
  Future<ReminderSettings> putSettings(ReminderSettings settings);

  /// The monthly due dates of [year]'s periods.
  Future<TaxCalendar> getCalendar(int year);
}

/// [RemindersRepository] over the generated client.
class ApiRemindersRepository implements RemindersRepository {
  /// The repository on [_client].
  new(this._client);

  final RemindersClient _client;

  /// The one place a `DeviceCreate` (whose generated `toString` prints the
  /// token) exists: built inline, handed to the client, never kept.
  @override
  Future<void> registerDevice({
    required String installationId,
    required DevicePlatform platform,
    required PushToken token,
    String? appVersion,
  }) => guardApi(
    () => _client.registerDevice(
      body: DeviceCreate(
        id: installationId,
        platform: platform,
        pushToken: token.value,
        appVersion: appVersion,
      ),
    ),
  );

  @override
  Future<void> unregisterDevice(String installationId) =>
      guardApi(() => _client.unregisterDevice(deviceId: installationId));

  @override
  Future<ReminderSettings> getSettings() =>
      guardApi(_client.getReminderSettings);

  @override
  Future<ReminderSettings> putSettings(ReminderSettings settings) =>
      guardApi(() => _client.putReminderSettings(body: settings));

  @override
  Future<TaxCalendar> getCalendar(int year) =>
      guardApi(() => _client.getTaxCalendar(year: year));
}

/// The reminder operations. Tests override them with a fake.
final remindersRepositoryProvider = Provider<RemindersRepository>(
  (ref) => ApiRemindersRepository(ref.watch(apiClientProvider).reminders),
  name: 'remindersRepositoryProvider',
);
