// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_breakdown_line_operation.dart';
import 'annual_line_source.dart';
import 'citation.dart';
import 'decimal_rate.dart';
import 'money_cents.dart';
import 'regime_code.dart';

part 'annual_breakdown_line.freezed.dart';
part 'annual_breakdown_line.g.dart';

/// One step of an annual computation trace: the monthly `BreakdownLine` shape (same `operation` values and rules) without `tax` (every annual line is ISR) and with `regime_code`, the regime a per-chapter or per-payment line belongs to (`null` for a section-wide line). `rate` only on `multiply` lines.
@Freezed()
abstract class AnnualBreakdownLine with _$AnnualBreakdownLine {
  const factory AnnualBreakdownLine({
    required int seq,
    @JsonKey(includeIfNull: true, name: 'regime_code')
    required RegimeCode? regimeCode,

    /// Human-readable step (Spanish).
    required String concept,
    required AnnualBreakdownLineOperation operation,
    @JsonKey(name: 'amount_cents') required MoneyCents amountCents,
    @JsonKey(includeIfNull: true) required DecimalRate? rate,
    @JsonKey(name: 'legal_basis') required List<Citation> legalBasis,
    required AnnualLineSource source,
    @JsonKey(name: 'needs_cpa_review') required bool needsCpaReview,
  }) = _AnnualBreakdownLine;

  factory AnnualBreakdownLine.fromJson(Map<String, Object?> json) =>
      _$AnnualBreakdownLineFromJson(json);
}
