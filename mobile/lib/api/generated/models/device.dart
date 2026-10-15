// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'device_platform.dart';
import 'installation_id.dart';

part 'device.freezed.dart';
part 'device.g.dart';

/// A registered installation. Never carries the push token.
@Freezed()
abstract class Device with _$Device {
  const factory Device({
    required InstallationId id,
    required DevicePlatform platform,
    @JsonKey(includeIfNull: true, name: 'app_version')
    required String? appVersion,
    @JsonKey(name: 'created_at') required DateTime createdAt,

    /// Last `registerDevice` of this installation; the 11th device evicts the oldest.
    @JsonKey(name: 'last_seen_at') required DateTime lastSeenAt,
  }) = _Device;

  factory Device.fromJson(Map<String, Object?> json) => _$DeviceFromJson(json);
}
