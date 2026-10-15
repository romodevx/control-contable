// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'filed_payment_month.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';
import 'regime_code.dart';

part 'filed_payments.freezed.dart';
part 'filed_payments.g.dart';

/// Exactly 12 months, January first. `total_effective_cents` is Σ `effective_cents` (what the annual estimate credits for this regime). `updated_at: null` = nothing confirmed yet.
@Freezed()
abstract class FiledPayments with _$FiledPayments {
  const factory FiledPayments({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<FiledPaymentMonth> months,
    @JsonKey(name: 'total_effective_cents')
    required NonNegativeCents totalEffectiveCents,
    @JsonKey(includeIfNull: true, name: 'updated_at')
    required DateTime? updatedAt,
  }) = _FiledPayments;

  factory FiledPayments.fromJson(Map<String, Object?> json) =>
      _$FiledPaymentsFromJson(json);
}
