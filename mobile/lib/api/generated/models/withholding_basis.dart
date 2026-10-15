// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// How the IVA withholding was computed: a share of the IVA charged (2/3, 1/2, all) or 4 % of the subtotal (road freight); `none` when no IVA is withheld.
@JsonEnum()
enum WithholdingBasis {
  @JsonValue('none')
  none('none'),
  @JsonValue('two_thirds_of_iva')
  twoThirdsOfIva('two_thirds_of_iva'),
  @JsonValue('half_of_iva')
  halfOfIva('half_of_iva'),
  @JsonValue('all_of_iva')
  allOfIva('all_of_iva'),
  @JsonValue('four_percent_of_price')
  fourPercentOfPrice('four_percent_of_price'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const WithholdingBasis(this.json);

  factory WithholdingBasis.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<WithholdingBasis> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
