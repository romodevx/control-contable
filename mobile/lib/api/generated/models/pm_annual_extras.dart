// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'money_cents.dart';
import 'non_negative_cents.dart';

part 'pm_annual_extras.freezed.dart';
part 'pm_annual_extras.g.dart';

/// Informative PM figures. `ptu_base_cents`: 601, ingresos acumulables − deductions (without PTU paid or losses; the art. 28 fr. XXX adjustment is not modelled ⚠️); RESICO PM, utilidad fiscal per art. 212 (⚠️ `resico_pm_ptu_base`). `ptu_to_distribute_cents`: 10 % of the base (601 only, ⚠️ `ptu_rate`; `null` for RESICO PM). `suggested_cu_next_year`: utilidad fiscal ÷ ingresos nominales, 4 decimals half-up (601 only, ⚠️ `cu_rounding`; `"0.0000"` when the utilidad is ≤ 0; `null` for RESICO PM), "Sugerido para tus pagos de {fy+1}".
@Freezed()
abstract class PmAnnualExtras with _$PmAnnualExtras {
  const factory PmAnnualExtras({
    @JsonKey(name: 'ptu_base_cents') required MoneyCents ptuBaseCents,
    @JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents')
    required NonNegativeCents? ptuToDistributeCents,
    @JsonKey(includeIfNull: true, name: 'suggested_cu_next_year')
    required DecimalRate? suggestedCuNextYear,
  }) = _PmAnnualExtras;

  factory PmAnnualExtras.fromJson(Map<String, Object?> json) =>
      _$PmAnnualExtrasFromJson(json);
}
