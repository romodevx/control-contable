// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_balance.dart';
import 'annual_credits.dart';
import 'non_negative_cents.dart';

part 'annual_tax.freezed.dart';
part 'annual_tax.g.dart';

/// The section's annual ISR. `accumulable_income_cents`: PF, Σ of each chapter's result floored at 0 (a 612 loss stays in 612, LISR 109); PM, ingresos acumulables. `personal_deductions_cents`: applied after caps (PF only, `null` for PM). `base_cents`: PF base (≤ 0 → 0); PM resultado fiscal. `isr_cents`: PF annual tariff (Anexo 8 C.II), PM 30 %.
@Freezed()
abstract class AnnualTax with _$AnnualTax {
  const factory AnnualTax({
    @JsonKey(name: 'accumulable_income_cents')
    required NonNegativeCents accumulableIncomeCents,
    @JsonKey(includeIfNull: true, name: 'personal_deductions_cents')
    required NonNegativeCents? personalDeductionsCents,
    @JsonKey(name: 'base_cents') required NonNegativeCents baseCents,
    @JsonKey(name: 'isr_cents') required NonNegativeCents isrCents,
    required AnnualCredits credits,
    required AnnualBalance balance,
  }) = _AnnualTax;

  factory AnnualTax.fromJson(Map<String, Object?> json) =>
      _$AnnualTaxFromJson(json);
}
