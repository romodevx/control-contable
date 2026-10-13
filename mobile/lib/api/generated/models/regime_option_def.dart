// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'regime_option_def_type.dart';

part 'regime_option_def.freezed.dart';
part 'regime_option_def.g.dart';

/// Per-regime option definition (from the research catalog).
@Freezed()
abstract class RegimeOptionDef with _$RegimeOptionDef {
  const factory RegimeOptionDef({
    required String key,
    required String label,
    required RegimeOptionDefType type,

    /// The name has been replaced because it contains a keyword. Original name: `required`.
    @JsonKey(name: 'required') required bool requiredValue,
    @JsonKey(includeIfNull: false, name: 'allowed_values')
    List<String>? allowedValues,

    /// The name has been replaced because it contains a keyword. Original name: `default`.
    /// The name has been replaced because it contains a keyword. Original name: `default`.
    @JsonKey(includeIfNull: false, name: 'default') dynamic defaultValue,
  }) = _RegimeOptionDef;

  factory RegimeOptionDef.fromJson(Map<String, Object?> json) =>
      _$RegimeOptionDefFromJson(json);
}
