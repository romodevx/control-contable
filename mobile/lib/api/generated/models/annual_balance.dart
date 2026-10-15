// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_balance_kind.dart';
import 'non_negative_cents.dart';

part 'annual_balance.freezed.dart';
part 'annual_balance.g.dart';

/// `isr_cents − credits.total_cents`, rounded once to the centavo (half away from zero): positive → `a_cargo`, negative → `a_favor` with its absolute value, zero → `cero`.
@Freezed()
abstract class AnnualBalance with _$AnnualBalance {
  const factory AnnualBalance({
    required AnnualBalanceKind kind,
    @JsonKey(name: 'amount_cents') required NonNegativeCents amountCents,
  }) = _AnnualBalance;

  factory AnnualBalance.fromJson(Map<String, Object?> json) =>
      _$AnnualBalanceFromJson(json);
}
