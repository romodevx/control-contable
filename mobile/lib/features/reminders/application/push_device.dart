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
  /// gone); another failure is thrown, the local marker cleared anyway
  /// (the claim on the next login and the FCM cleanup cover the server).
  Future<void> unregister() async {
    if (!await _store.isRegistered()) return;
    try {
      await _repository.unregisterDevice(await _store.installationId());
    } on ApiNotFound {
      // Already removed (a claim by another user, a second call).
    } finally {
      await _store.setRegistered(registered: false);
    }
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
