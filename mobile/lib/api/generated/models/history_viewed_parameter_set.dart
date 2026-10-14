// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';

part 'history_viewed_parameter_set.freezed.dart';
part 'history_viewed_parameter_set.g.dart';

/// The `parameter_set` (`ParameterSetRef`) of the rendered estimate, copied as is; closed for input.
@Freezed()
abstract class HistoryViewedParameterSet with _$HistoryViewedParameterSet {
  const factory HistoryViewedParameterSet({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required int version,
  }) = _HistoryViewedParameterSet;

  factory HistoryViewedParameterSet.fromJson(Map<String, Object?> json) =>
      _$HistoryViewedParameterSetFromJson(json);
}
