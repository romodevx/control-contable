// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'iva_components.freezed.dart';
part 'iva_components.g.dart';

/// This regime's contribution; monthly IVA is netted per taxpayer in totals.iva.
@Freezed()
abstract class IvaComponents with _$IvaComponents {
  const factory IvaComponents({
    @JsonKey(name: 'charged_collected_cents')
    required NonNegativeCents chargedCollectedCents,
    @JsonKey(name: 'creditable_paid_cents')
    required NonNegativeCents creditablePaidCents,
    @JsonKey(name: 'withheld_from_taxpayer_cents')
    required NonNegativeCents withheldFromTaxpayerCents,

    /// IVA already definitive via withholding (e.g. 625 under LIVA 18-M).
    @JsonKey(includeIfNull: false, name: 'definitive_excluded_cents')
    NonNegativeCents? definitiveExcludedCents,
  }) = _IvaComponents;

  factory IvaComponents.fromJson(Map<String, Object?> json) =>
      _$IvaComponentsFromJson(json);
}
