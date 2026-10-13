// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'device_create_platform.dart';
import 'uuid.dart';

part 'device_create.freezed.dart';
part 'device_create.g.dart';

@Freezed()
abstract class DeviceCreate with _$DeviceCreate {
  const factory DeviceCreate({
    required Uuid id,
    required DeviceCreatePlatform platform,
    @JsonKey(name: 'push_token') required String pushToken,
    @JsonKey(includeIfNull: false, name: 'app_version') String? appVersion,
  }) = _DeviceCreate;

  factory DeviceCreate.fromJson(Map<String, Object?> json) =>
      _$DeviceCreateFromJson(json);
}
