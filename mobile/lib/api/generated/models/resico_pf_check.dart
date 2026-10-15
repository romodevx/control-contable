// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'money_cents.dart';
import 'non_negative_cents.dart';

part 'resico_pf_check.freezed.dart';
part 'resico_pf_check.g.dart';

/// RESICO PF 113-F informative reconciliation (MODULE-annual §2.7): annual income collected × the annual non-marginal `rate` − the monthly RESICO ISR − the 1.25 % withheld = `difference_cents`, shown as "diferencia informativa", **never** as an amount to pay or to get back (the monthly payments are definitive, rule 3.13.7).
@Freezed()
abstract class ResicoPfCheck with _$ResicoPfCheck {
  const factory ResicoPfCheck({
    @JsonKey(name: 'income_cents') required NonNegativeCents incomeCents,
    required DecimalRate rate,
    @JsonKey(name: 'annual_isr_cents') required NonNegativeCents annualIsrCents,
    @JsonKey(name: 'monthly_payments_cents')
    required NonNegativeCents monthlyPaymentsCents,
    @JsonKey(name: 'withheld_cents') required NonNegativeCents withheldCents,
    @JsonKey(name: 'difference_cents') required MoneyCents differenceCents,
  }) = _ResicoPfCheck;

  factory ResicoPfCheck.fromJson(Map<String, Object?> json) =>
      _$ResicoPfCheckFromJson(json);
}
