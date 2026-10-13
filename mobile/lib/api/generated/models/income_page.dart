// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'income.dart';
import 'period_totals.dart';

part 'income_page.freezed.dart';
part 'income_page.g.dart';

@Freezed()
abstract class IncomePage with _$IncomePage {
  const factory IncomePage({
    required List<Income> items,
    required PeriodTotals totals,
    @JsonKey(includeIfNull: false, name: 'next_cursor') String? nextCursor,
  }) = _IncomePage;

  factory IncomePage.fromJson(Map<String, Object?> json) =>
      _$IncomePageFromJson(json);
}
