import 'package:app_settings/app_settings.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Opens this app's notification settings in the OS (the way back after
/// the user denied the permission: the app never re-prompts, §7).
typedef SystemSettingsOpener = Future<void> Function();

/// The OS opener. Tests override it.
final systemSettingsOpenerProvider = Provider<SystemSettingsOpener>(
  (ref) =>
      () => AppSettings.openAppSettings(type: AppSettingsType.notification),
  name: 'systemSettingsOpenerProvider',
);
