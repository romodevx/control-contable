// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'filed_payment_source.dart';
import 'non_negative_cents.dart';
import 'period.dart';

part 'filed_payment_month.freezed.dart';
part 'filed_payment_month.g.dart';

/// One month. `filed_cents`: confirmed here. `tax_input_cents`: the M5 `prior_provisional_payments` value of the month. `app_estimate_cents`: the app's monthly ISR estimate for the regime, `max(0, isr)` (`null` outside the window or for a month not closed yet); `app_estimate_incomplete` when that estimate was `incomplete`, which makes the annual section `incomplete` unless the month is filed. `effective_cents`/`source`: the first present by precedence (`null`/`none` when none is).
@Freezed()
abstract class FiledPaymentMonth with _$FiledPaymentMonth {
  const factory FiledPaymentMonth({
    required int month,
    required Period period,
    @JsonKey(name: 'in_window') required bool inWindow,
    @JsonKey(includeIfNull: true, name: 'filed_cents')
    required NonNegativeCents? filedCents,
    @JsonKey(includeIfNull: true, name: 'tax_input_cents')
    required NonNegativeCents? taxInputCents,
    @JsonKey(includeIfNull: true, name: 'app_estimate_cents')
    required NonNegativeCents? appEstimateCents,
    @JsonKey(name: 'app_estimate_incomplete')
    required bool appEstimateIncomplete,
    @JsonKey(includeIfNull: true, name: 'effective_cents')
    required NonNegativeCents? effectiveCents,
    required FiledPaymentSource source,
  }) = _FiledPaymentMonth;

  factory FiledPaymentMonth.fromJson(Map<String, Object?> json) =>
      _$FiledPaymentMonthFromJson(json);
}
