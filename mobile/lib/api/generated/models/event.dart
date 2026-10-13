// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'environment.dart';

part 'event.freezed.dart';
part 'event.g.dart';

@Freezed()
abstract class Event with _$Event {
  const factory Event({
    required String id,
    required String type,
    @JsonKey(name: 'app_user_id') required String appUserId,
    @JsonKey(name: 'event_timestamp_ms') required int eventTimestampMs,
    @JsonKey(name: 'original_app_user_id') String? originalAppUserId,
    @JsonKey(name: 'product_id') String? productId,
    @JsonKey(name: 'entitlement_ids') List<String>? entitlementIds,
    Environment? environment,
    String? store,
    @JsonKey(name: 'expiration_at_ms') int? expirationAtMs,
  }) = _Event;

  factory Event.fromJson(Map<String, Object?> json) => _$EventFromJson(json);
}
