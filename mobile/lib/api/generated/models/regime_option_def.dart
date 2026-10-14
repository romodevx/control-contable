// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'regime_option_def_type.dart';

part 'regime_option_def.freezed.dart';
part 'regime_option_def.g.dart';

/// One per-regime option (`regimes.option_defs`, data not code). `enum` takes one of `allowed_values`; `multi_enum` takes a non-empty array of distinct `allowed_values`; `boolean` takes `true`/`false`; `rate` takes a decimal string (`DecimalRate`, 0.16 = 16%) within `[min, max]`. An omitted option with a `default` is filled server-side; a `required` option with `default: null` must be sent.
@Freezed()
abstract class RegimeOptionDef with _$RegimeOptionDef {
  const factory RegimeOptionDef({
    required String key,

    /// Spanish label shown in the options form.
    required String label,
    required RegimeOptionDefType type,

    /// The name has been replaced because it contains a keyword. Original name: `required`.
    @JsonKey(name: 'required') required bool requiredValue,

    /// Present for `enum` and `multi_enum`.
    @JsonKey(includeIfNull: false, name: 'allowed_values')
    List<String>? allowedValues,

    /// Inclusive lower bound for `rate`; `null` or absent otherwise.
    @JsonKey(includeIfNull: false) DecimalRate? min,

    /// Inclusive upper bound for `rate`; `null` or absent otherwise.
    @JsonKey(includeIfNull: false) DecimalRate? max,

    /// Default value, or `null` when the option has none.
    /// The name has been replaced because it contains a keyword. Original name: `default`.
    /// The name has been replaced because it contains a keyword. Original name: `default`.
    @JsonKey(includeIfNull: false, name: 'default') dynamic defaultValue,
  }) = _RegimeOptionDef;

  factory RegimeOptionDef.fromJson(Map<String, Object?> json) =>
      _$RegimeOptionDefFromJson(json);
}
