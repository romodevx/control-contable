// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'breakdown_line_operation.dart';
import 'breakdown_line_tax.dart';
import 'citation.dart';
import 'decimal_rate.dart';
import 'money_cents.dart';

part 'breakdown_line.freezed.dart';
part 'breakdown_line.g.dart';

@Freezed()
abstract class BreakdownLine with _$BreakdownLine {
  const factory BreakdownLine({
    required int seq,
    required BreakdownLineTax tax,

    /// Human-readable step (Spanish).
    required String concept,
    required BreakdownLineOperation operation,
    @JsonKey(name: 'amount_cents') required MoneyCents amountCents,
    @JsonKey(name: 'legal_basis') required List<Citation> legalBasis,
    @JsonKey(includeIfNull: false) DecimalRate? rate,
  }) = _BreakdownLine;

  factory BreakdownLine.fromJson(Map<String, Object?> json) =>
      _$BreakdownLineFromJson(json);
}
