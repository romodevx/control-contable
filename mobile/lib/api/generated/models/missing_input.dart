// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';

part 'missing_input.freezed.dart';
part 'missing_input.g.dart';

/// A tax input the period needs and the taxpayer has not entered (`TaxInputDef.key` of the regime). For a `money_monthly` key, `months` lists the months with no value, in order. The CTA opens `getTaxInputs` for the regime and the period's fiscal year, focused on `key`.
@Freezed()
abstract class MissingInput with _$MissingInput {
  const factory MissingInput({
    required String key,

    /// Spanish label (`TaxInputDef.label`).
    required String label,

    /// Only for `money_monthly` keys.
    @JsonKey(includeIfNull: false) List<Period>? months,
  }) = _MissingInput;

  factory MissingInput.fromJson(Map<String, Object?> json) =>
      _$MissingInputFromJson(json);
}
