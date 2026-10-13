// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'device_platform.dart';
import 'uuid.dart';

part 'device.freezed.dart';
part 'device.g.dart';

@Freezed()
abstract class Device with _$Device {
  const factory Device({
    required Uuid id,
    required DevicePlatform platform,
    @JsonKey(name: 'created_at') required DateTime createdAt,
  }) = _Device;

  factory Device.fromJson(Map<String, Object?> json) => _$DeviceFromJson(json);
}
