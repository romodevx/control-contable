// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'event.dart';

part 'revenue_cat_webhook.freezed.dart';
part 'revenue_cat_webhook.g.dart';

/// Subset we rely on; unknown fields are accepted and stored raw.
@Freezed()
abstract class RevenueCatWebhook with _$RevenueCatWebhook {
  const factory RevenueCatWebhook({
    @JsonKey(name: 'api_version') required String apiVersion,
    required Event event,
  }) = _RevenueCatWebhook;

  factory RevenueCatWebhook.fromJson(Map<String, Object?> json) =>
      _$RevenueCatWebhookFromJson(json);
}
