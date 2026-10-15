import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/installation_store.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// This installation as a push target of the signed-in user
/// (MODULE-reminders §2.3): registers it, keeps its token current, and
/// removes it on opt-out and sign-out.
class PushDevice {
  /// The device over its collaborators.
  new({
    required this._messaging,
    required this._repository,
    required this._store,
    required this._platform,
  });

  final PushMessaging _messaging;
  final RemindersRepository _repository;
  final InstallationStore _store;
  final DevicePlatform _platform;

  /// Registers the current token. Throws [PushUnavailable] when the
  /// provider has none, an `ApiFailure` when the server refuses.
  Future<void> register() async {
    final token = await _messaging.token();
    if (token == null) throw const PushUnavailable();
    await _registerToken(token);
  }

  /// FCM rotated the token: re-register, only if this install is a target.
  Future<void> onTokenRefreshed(PushToken token) async {
    if (await _store.isRegistered()) await _registerToken(token);
  }

  /// Removes this installation, if registered. A `404` is done (already
  /// gone). Another failure is thrown and the local marker **kept** (M9
  /// security gate F1): the delete may not have happened, so the sign-out
  /// sends the id with `logout` ([registeredId]) and the next sign-in
  /// removes what is left ([evict]).
  Future<void> unregister() async {
    if (!await _store.isRegistered()) return;
    try {
      await _repository.unregisterDevice(await _store.installationId());
    } on ApiNotFound {
      // Already removed (a claim by another user, a second call).
    }
    await _store.setRegistered(registered: false);
  }

  /// The installation id while this installation may still be registered
  /// on the server, else null (what `logout` removes with the session).
  Future<String?> registeredId() async =>
      await _store.isRegistered() ? await _store.installationId() : null;

  /// Removes a registration the signed-in user did not ask for (one a
  /// previous user's failed sign-out left): claims the installation for
  /// the caller, which removes the other user's row, then deletes it.
  /// Throws [PushUnavailable] without a token, an `ApiFailure` when the
  /// server refuses; the marker then stays for the next try.
  Future<void> evict() async {
    await register();
    await unregister();
  }

  Future<void> _registerToken(PushToken token) async {
    await _repository.registerDevice(
      installationId: await _store.installationId(),
      platform: _platform,
      token: token,
    );
    await _store.setRegistered(registered: true);
  }
}

/// The platform this build reports to `registerDevice`.
final devicePlatformProvider = Provider<DevicePlatform>(
  (ref) => defaultTargetPlatform == TargetPlatform.iOS
      ? DevicePlatform.ios
      : DevicePlatform.android,
  name: 'devicePlatformProvider',
);

/// This installation as a push target.
final pushDeviceProvider = Provider<PushDevice>(
  (ref) => PushDevice(
    messaging: ref.watch(pushMessagingProvider),
    repository: ref.watch(remindersRepositoryProvider),
    store: ref.watch(installationStoreProvider),
    platform: ref.watch(devicePlatformProvider),
  ),
  name: 'pushDeviceProvider',
);
