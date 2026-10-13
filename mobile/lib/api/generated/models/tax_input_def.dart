// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'tax_input_def_kind.dart';

part 'tax_input_def.freezed.dart';
part 'tax_input_def.g.dart';

@Freezed()
abstract class TaxInputDef with _$TaxInputDef {
  const factory TaxInputDef({
    required String key,
    required String label,
    required TaxInputDefKind kind,

    /// The name has been replaced because it contains a keyword. Original name: `required`.
    @JsonKey(name: 'required') required bool requiredValue,
  }) = _TaxInputDef;

  factory TaxInputDef.fromJson(Map<String, Object?> json) =>
      _$TaxInputDefFromJson(json);
}
