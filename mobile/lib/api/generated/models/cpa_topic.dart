// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only. A rule a CPA has not validated yet (MODULE-tax-engine §2.9, each mapped to a `docs/reviews/CPA-REVIEW.md` item): `rif_reduction` (any 621 figure), `rif_bimonthly_iva` (621 IVA by bimester, OQ-T7), `platform_definitive_iva` (625 definitive option with IVA), `cross_regime_iva_netting` (two or more regimes contribute IVA, OQ-T2), `resico_pm_interest_withholding` (626 PM with interest withholding), `withholding_rule_unverified` (a line fed by records whose derived withholding matched a ⚠️ rule, MODULE-withholdings §6), `withheld_iva_credit_timing` (IVA withheld from suppliers credited the month after, assuming timely remittance, LIVA 5 fr. IV; 0.10.0).
@JsonEnum()
enum CpaTopic {
  @JsonValue('rif_reduction')
  rifReduction('rif_reduction'),
  @JsonValue('rif_bimonthly_iva')
  rifBimonthlyIva('rif_bimonthly_iva'),
  @JsonValue('platform_definitive_iva')
  platformDefinitiveIva('platform_definitive_iva'),
  @JsonValue('cross_regime_iva_netting')
  crossRegimeIvaNetting('cross_regime_iva_netting'),
  @JsonValue('resico_pm_interest_withholding')
  resicoPmInterestWithholding('resico_pm_interest_withholding'),
  @JsonValue('withholding_rule_unverified')
  withholdingRuleUnverified('withholding_rule_unverified'),
  @JsonValue('withheld_iva_credit_timing')
  withheldIvaCreditTiming('withheld_iva_credit_timing'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const CpaTopic(this.json);

  factory CpaTopic.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<CpaTopic> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
