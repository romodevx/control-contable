// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'disclaimer.dart';
import 'fiscal_year.dart';
import 'history_month.dart';
import 'history_year_totals.dart';
import 'legend.dart';
import 'uuid.dart';

part 'history_year.freezed.dart';
part 'history_year.g.dart';

/// Body of `getHistory`. `months` newest first (see the operation). `legend` is mandatory under the list. `disclaimers` is the union of the months' disclaimers, in the M5 order (`not_tax_advice` first, then by severity, then by code); a `strong` `cpa_review_pending` is a non-dismissible banner. The example is a 612 taxpayer whose regime starts in 2026-03, read on 2026-10-14: October is the current month, June changed by records, July by an engine correction (1.0.0 → 1.0.1), August and April were never viewed, January and February have no regime.
@Freezed()
abstract class HistoryYear with _$HistoryYear {
  const factory HistoryYear({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<HistoryMonth> months,
    @JsonKey(name: 'year_totals') required HistoryYearTotals yearTotals,
    required Legend legend,
    required List<Disclaimer> disclaimers,
  }) = _HistoryYear;

  factory HistoryYear.fromJson(Map<String, Object?> json) =>
      _$HistoryYearFromJson(json);
}
