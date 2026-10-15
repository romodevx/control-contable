// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `electronic`: card, transfer, check or e-wallet. `cash` is excluded for the four cash-restricted categories.
@JsonEnum()
enum DeductionPaymentMethod {
  @JsonValue('electronic')
  electronic('electronic'),
  @JsonValue('cash')
  cash('cash'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DeductionPaymentMethod(this.json);

  factory DeductionPaymentMethod.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DeductionPaymentMethod> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
