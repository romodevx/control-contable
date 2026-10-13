// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';
import 'history_month.dart';
import 'legend.dart';
import 'uuid.dart';

part 'history_year.freezed.dart';
part 'history_year.g.dart';

@Freezed()
abstract class HistoryYear with _$HistoryYear {
  const factory HistoryYear({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<HistoryMonth> months,
    required Legend legend,
  }) = _HistoryYear;

  factory HistoryYear.fromJson(Map<String, Object?> json) =>
      _$HistoryYearFromJson(json);
}
