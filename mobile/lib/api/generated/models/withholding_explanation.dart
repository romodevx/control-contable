// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'withholding_rule_id.dart';

part 'withholding_explanation.freezed.dart';
part 'withholding_explanation.g.dart';

@Freezed()
abstract class WithholdingExplanation with _$WithholdingExplanation {
  const factory WithholdingExplanation({
    @JsonKey(name: 'rule_id') required WithholdingRuleId ruleId,

    /// One plain-words line for a matched rule, with its legal basis (es-MX), written by the server from the decision table. Safe to show; never an RFC, a name or a description.
    required String text,
  }) = _WithholdingExplanation;

  factory WithholdingExplanation.fromJson(Map<String, Object?> json) =>
      _$WithholdingExplanationFromJson(json);
}
