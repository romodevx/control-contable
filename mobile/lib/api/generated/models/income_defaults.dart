// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'iva_rate.dart';
import 'platform_activity.dart';

part 'income_defaults.freezed.dart';
part 'income_defaults.g.dart';

/// Form prefill from the rate catalog (UX only; the server re-validates every write).
@Freezed()
abstract class IncomeDefaults with _$IncomeDefaults {
  const factory IncomeDefaults({
    @JsonKey(name: 'iva_rates_allowed') required List<IvaRate> ivaRatesAllowed,
    @JsonKey(name: 'suggested_iva_rate') required IvaRate suggestedIvaRate,
    @JsonKey(name: 'isr_withholding_allowed')
    required bool isrWithholdingAllowed,

    /// Fraction of `amount_cents`; `null` when not allowed or when there is no suggestion.
    @JsonKey(includeIfNull: true, name: 'isr_withholding_rate')
    required DecimalRate? isrWithholdingRate,
    @JsonKey(name: 'iva_withholding_allowed')
    required bool ivaWithholdingAllowed,

    /// Fraction of `iva_cents` (2/3 = `0.666667`); `null` when not allowed or when there is no suggestion.
    @JsonKey(includeIfNull: true, name: 'iva_withholding_rate')
    required DecimalRate? ivaWithholdingRate,

    /// The 625 activity the ISR rate is for; `null` outside 625 or when it is ambiguous.
    @JsonKey(includeIfNull: true, name: 'platform_activity')
    required PlatformActivity? platformActivity,
  }) = _IncomeDefaults;

  factory IncomeDefaults.fromJson(Map<String, Object?> json) =>
      _$IncomeDefaultsFromJson(json);
}
