// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_computed_section_kind.dart';
import 'annual_legend.dart';
import 'annual_section.dart';
import 'annual_year_status.dart';
import 'disclaimer.dart';
import 'engine_version.dart';
import 'fiscal_year.dart';
import 'parameter_set_ref.dart';
import 'period.dart';
import 'person_type.dart';
import 'uuid.dart';

part 'annual_estimate.freezed.dart';
part 'annual_estimate.g.dart';

/// Informational annual estimate (derived on every read; byte-identical for the same inputs, parameters and engine version). `disclaimers` order: `annual_estimate` first, then by severity (`strong`, `warning`, `info`), then by code; `no_annual_iva` and, for fiscal year 2026, `rmf_2027_pending` are always present. `deadline` is the earliest section deadline (`null` when no section has one). `data_through` is the last month with records counted (`YYYY-12` for a closed year, the last closed month in the current year, `null` in its January). `informative_income_over_500k`: a PF whose total income, exempt included, exceeds 500,000.00 (LISR 150).
@Freezed()
abstract class AnnualEstimate with _$AnnualEstimate {
  const factory AnnualEstimate({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    @JsonKey(name: 'person_type') required PersonType personType,
    required AnnualLegend legend,
    required List<Disclaimer> disclaimers,
    @JsonKey(name: 'year_status') required AnnualYearStatus yearStatus,
    @JsonKey(includeIfNull: true, name: 'data_through')
    required Period? dataThrough,
    @JsonKey(includeIfNull: true) required String? deadline,

    /// The sections whose `tax.balance` is the headline, each shown on its own (never netted).
    @JsonKey(name: 'headline_sections')
    required List<AnnualComputedSectionKind> headlineSections,
    required List<AnnualSection> sections,
    @JsonKey(name: 'informative_income_over_500k')
    required bool informativeIncomeOver500k,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
    @JsonKey(name: 'engine_version') required EngineVersion engineVersion,
  }) = _AnnualEstimate;

  factory AnnualEstimate.fromJson(Map<String, Object?> json) =>
      _$AnnualEstimateFromJson(json);
}
