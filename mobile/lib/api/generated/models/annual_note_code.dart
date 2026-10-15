// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// Append-only (like `ErrorCode`): clients show the `text` of a code they do not know. `refund_only_paid` (a saldo a favor that depends on a payment estimated by the app: only what was actually paid or withheld is refunded, LISR 152); `refund_needs_efirma` (saldo a favor ≥ 10,000.00); `loss_carryforward` (a chapter or PM loss, amortizable in later years); `cu_suggestion` (601 suggested coeficiente for next year).
@JsonEnum()
enum AnnualNoteCode {
  @JsonValue('refund_only_paid')
  refundOnlyPaid('refund_only_paid'),
  @JsonValue('refund_needs_efirma')
  refundNeedsEfirma('refund_needs_efirma'),
  @JsonValue('loss_carryforward')
  lossCarryforward('loss_carryforward'),
  @JsonValue('cu_suggestion')
  cuSuggestion('cu_suggestion'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const AnnualNoteCode(this.json);

  factory AnnualNoteCode.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<AnnualNoteCode> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
