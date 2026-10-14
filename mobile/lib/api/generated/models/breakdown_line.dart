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

/// One step of the computation trace. `amount_cents` is the line's value rounded to the centavo for display. `rate` is set only on `multiply` lines. `legal_basis` has at least one citation on every line except `info` and `rounding_adjustment`. `operation`: `input` (a sum of records or a tax input), `add`, `subtract`, `multiply`, `lookup` (a tariff row; the concept names the row and month), `result` (a subtotal or the tax result), `rounding_adjustment` ("Ajuste por redondeo", ±0.01…, so the displayed lines reconcile with the rounded result), `info` (a fact with no arithmetic, e.g. "Retención definitiva"). `tax: withholding` lines are the withholdings to remit.
@Freezed()
abstract class BreakdownLine with _$BreakdownLine {
  const factory BreakdownLine({
    /// 1-based, fixed order per strategy.
    required int seq,
    required BreakdownLineTax tax,

    /// Human-readable step (Spanish).
    required String concept,
    required BreakdownLineOperation operation,
    @JsonKey(name: 'amount_cents') required MoneyCents amountCents,
    @JsonKey(includeIfNull: true) required DecimalRate? rate,
    @JsonKey(name: 'legal_basis') required List<Citation> legalBasis,

    /// Where the figure comes from: `records` (incomes/expenses of the period or year), `tax_input:<key>`, `params:<key>` (the parameter set) or `computed`.
    required String source,

    /// The line applies a rule a CPA has not validated yet (⚠️ next to the line).
    @JsonKey(name: 'needs_cpa_review') required bool needsCpaReview,
  }) = _BreakdownLine;

  factory BreakdownLine.fromJson(Map<String, Object?> json) =>
      _$BreakdownLineFromJson(json);
}
