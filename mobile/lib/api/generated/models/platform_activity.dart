// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// 625 activity of an income (same values as the 625 `activities` option). Selects the ISR withholding rate (LIF 2026 art. 25-VI).
@JsonEnum()
enum PlatformActivity {
  @JsonValue('transport_delivery')
  transportDelivery('transport_delivery'),
  @JsonValue('lodging')
  lodging('lodging'),
  @JsonValue('goods_services')
  goodsServices('goods_services'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const PlatformActivity(this.json);

  factory PlatformActivity.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<PlatformActivity> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
