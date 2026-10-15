import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/reminders/domain/uuid_v7.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// This installation's push identity on the device: the installation id
/// (`DeviceCreate.id`, a UUIDv7 made once per install) and whether it is
/// registered. Neither is secret: the token itself is never stored here.
abstract interface class InstallationStore {
  /// The installation id; made and kept on first use.
  Future<String> installationId();

  /// Whether this installation is registered for the signed-in user.
  Future<bool> isRegistered();

  /// Keeps whether this installation is registered.
  Future<void> setRegistered({required bool registered});
}

/// Preference key of the installation id.
const String installationIdKey = 'push_installation_id';

/// Preference key of the "registered" marker.
const String deviceRegisteredKey = 'push_device_registered';

/// [InstallationStore] over the platform preferences.
class PreferencesInstallationStore implements InstallationStore {
  /// The store over [_preferences]; new ids are stamped with [_clock].
  new(this._preferences, this._clock);

  final SharedPreferencesAsync _preferences;
  final Clock _clock;

  @override
  Future<String> installationId() async {
    final kept = await _preferences.getString(installationIdKey);
    if (kept != null) return kept;
    final made = newUuidV7(_clock());
    await _preferences.setString(installationIdKey, made);
    return made;
  }

  @override
  Future<bool> isRegistered() async =>
      await _preferences.getBool(deviceRegisteredKey) ?? false;

  @override
  Future<void> setRegistered({required bool registered}) =>
      _preferences.setBool(deviceRegisteredKey, registered);
}

/// An [InstallationStore] held in memory (tests).
class InMemoryInstallationStore implements InstallationStore {
  /// A store whose installation id is [id].
  new(this.id);

  /// The installation id.
  final String id;

  /// Whether registered.
  bool registered = false;

  @override
  Future<String> installationId() async => id;

  @override
  Future<bool> isRegistered() async => registered;

  @override
  Future<void> setRegistered({required bool registered}) async =>
      this.registered = registered;
}

/// The device's installation store. Tests override it.
final installationStoreProvider = Provider<InstallationStore>(
  (ref) => PreferencesInstallationStore(
    SharedPreferencesAsync(),
    ref.watch(clockProvider),
  ),
  name: 'installationStoreProvider',
);
