// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'income_last_used.dart';
import 'iva_rate.dart';
import 'platform_activity.dart';

part 'income_defaults.freezed.dart';
part 'income_defaults.g.dart';

/// Form prefill from the rate catalog (UX only; the server re-validates every write). Withholdings are derived (ADR-0008), so there is no rate to suggest: the form shows `previewIncomeWithholdings`. **0.12.0 (WH-11, ADR-0008 amendment):** the four `*_withholding_*` fields are removed and `last_used` is required.
@Freezed()
abstract class IncomeDefaults with _$IncomeDefaults {
  const factory IncomeDefaults({
    @JsonKey(name: 'iva_rates_allowed') required List<IvaRate> ivaRatesAllowed,
    @JsonKey(name: 'suggested_iva_rate') required IvaRate suggestedIvaRate,
    @JsonKey(name: 'last_used') required IncomeLastUsed lastUsed,

    /// The 625 activity to preselect; `null` outside 625 or when it is ambiguous.
    @JsonKey(includeIfNull: true, name: 'platform_activity')
    required PlatformActivity? platformActivity,
  }) = _IncomeDefaults;

  factory IncomeDefaults.fromJson(Map<String, Object?> json) =>
      _$IncomeDefaultsFromJson(json);
}
