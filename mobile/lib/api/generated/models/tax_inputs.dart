// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';
import 'regime_code.dart';
import 'tax_input_def.dart';
import 'tax_input_value.dart';

part 'tax_inputs.freezed.dart';
part 'tax_inputs.g.dart';

@Freezed()
abstract class TaxInputs with _$TaxInputs {
  const factory TaxInputs({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required List<TaxInputDef> definitions,
    required List<TaxInputValue> values,
  }) = _TaxInputs;

  factory TaxInputs.fromJson(Map<String, Object?> json) =>
      _$TaxInputsFromJson(json);
}
