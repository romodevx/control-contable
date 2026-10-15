// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Which date books an income of the regime for ISR (MODULE-incomes §2.2): `cash` books on `collected_on`, `accrual` on `invoiced_on` (which is then required). Catalog data (`regimes.income_basis`), never a switch on the regime code: `accrual` is the personas morales of Título II (601, 620, 622, 623, 624); `cash` is every persona física regime, RESICO PM (626), 603 and 610. IVA is always cash (LIVA 1-B) whatever the basis. Added in 0.16.0 so the income form no longer mirrors the catalog in code.
@JsonEnum()
enum IncomeBasis {
  @JsonValue('cash')
  cash('cash'),
  @JsonValue('accrual')
  accrual('accrual'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const IncomeBasis(this.json);

  factory IncomeBasis.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<IncomeBasis> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
