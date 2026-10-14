// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'iva_components.freezed.dart';
part 'iva_components.g.dart';

/// This regime's contribution to the monthly IVA, which is netted per taxpayer in `totals.iva` (LIVA 5-D, OQ-T2); never an amount to pay on its own. `withheld_from_taxpayer_cents` is only the IVA recorded as withheld on incomes (OQ-T6 a); IVA the taxpayer withheld from suppliers is to remit, not creditable (OQ-T6 b).
@Freezed()
abstract class IvaComponents with _$IvaComponents {
  const factory IvaComponents({
    @JsonKey(name: 'charged_collected_cents')
    required NonNegativeCents chargedCollectedCents,
    @JsonKey(name: 'creditable_paid_cents')
    required NonNegativeCents creditablePaidCents,
    @JsonKey(name: 'withheld_from_taxpayer_cents')
    required NonNegativeCents withheldFromTaxpayerCents,

    /// IVA already definitive via withholding (625 under LIVA 18-M ⚠️); 0 otherwise.
    @JsonKey(name: 'definitive_excluded_cents')
    required NonNegativeCents definitiveExcludedCents,
  }) = _IvaComponents;

  factory IvaComponents.fromJson(Map<String, Object?> json) =>
      _$IvaComponentsFromJson(json);
}
