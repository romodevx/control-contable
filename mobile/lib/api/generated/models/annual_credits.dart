// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';
import 'payment_source_counts.dart';

part 'annual_credits.freezed.dart';
part 'annual_credits.g.dart';

/// What is subtracted from the annual ISR: provisional ISR payments of the section's regimes, ISR withheld from the taxpayer (10 % by a persona moral, platforms; PM: interest withholding), ISR withheld on salary (Σ constancias; 0 for PM). `total_cents` is their sum.
@Freezed()
abstract class AnnualCredits with _$AnnualCredits {
  const factory AnnualCredits({
    @JsonKey(name: 'payments_cents') required NonNegativeCents paymentsCents,
    @JsonKey(name: 'withholdings_cents')
    required NonNegativeCents withholdingsCents,
    @JsonKey(name: 'salary_isr_cents') required NonNegativeCents salaryIsrCents,
    @JsonKey(name: 'total_cents') required NonNegativeCents totalCents,
    @JsonKey(name: 'payment_sources')
    required PaymentSourceCounts paymentSources,
  }) = _AnnualCredits;

  factory AnnualCredits.fromJson(Map<String, Object?> json) =>
      _$AnnualCreditsFromJson(json);
}
