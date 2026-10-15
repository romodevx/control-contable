// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_breakdown_line.dart';
import 'annual_legend.dart';
import 'annual_section.dart';
import 'disclaimer.dart';
import 'engine_version.dart';
import 'fiscal_year.dart';
import 'uuid.dart';

part 'annual_breakdown.freezed.dart';
part 'annual_breakdown.g.dart';

/// Header: `section.parameter_set` and `engine_version`. `legend` and `disclaimers` travel with every breakdown (the estimate's disclaimers that bear on this section, same order), so the warnings are visible wherever the figures are.
@Freezed()
abstract class AnnualBreakdown with _$AnnualBreakdown {
  const factory AnnualBreakdown({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required AnnualSection section,
    required List<AnnualBreakdownLine> lines,
    required AnnualLegend legend,
    required List<Disclaimer> disclaimers,
    @JsonKey(name: 'engine_version') required EngineVersion engineVersion,
  }) = _AnnualBreakdown;

  factory AnnualBreakdown.fromJson(Map<String, Object?> json) =>
      _$AnnualBreakdownFromJson(json);
}
