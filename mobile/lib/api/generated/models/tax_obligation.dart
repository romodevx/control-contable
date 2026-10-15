// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only (like `ErrorCode`, not frozen). v1 lists only `monthly_isr_iva`, the monthly provisional ISR and definitive IVA payment due on day 17. Clients skip an entry whose obligation they do not know.
@JsonEnum()
enum TaxObligation {
  @JsonValue('monthly_isr_iva')
  monthlyIsrIva('monthly_isr_iva'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const TaxObligation(this.json);

  factory TaxObligation.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<TaxObligation> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
