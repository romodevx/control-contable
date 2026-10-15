// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'device_platform.dart';
import 'installation_id.dart';

part 'device_create.freezed.dart';
part 'device_create.g.dart';

/// Body of `registerDevice`. Closed: an unknown field is `400 additionalProperties`.
@Freezed()
abstract class DeviceCreate with _$DeviceCreate {
  const factory DeviceCreate({
    required InstallationId id,
    required DevicePlatform platform,

    /// The FCM registration token (APNs goes through FCM). Printable ASCII, no spaces. Write-only: encrypted at rest, never logged, never returned.
    @JsonKey(name: 'push_token') required String pushToken,

    /// The app's version string, for support; optional.
    @JsonKey(includeIfNull: false, name: 'app_version') String? appVersion,
  }) = _DeviceCreate;

  factory DeviceCreate.fromJson(Map<String, Object?> json) =>
      _$DeviceCreateFromJson(json);
}
