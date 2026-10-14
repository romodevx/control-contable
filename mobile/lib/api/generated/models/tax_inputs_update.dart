// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'tax_input_value.dart';

part 'tax_inputs_update.freezed.dart';
part 'tax_inputs_update.g.dart';

/// The complete set of values for the regime and fiscal year (replace, not merge).
@Freezed()
abstract class TaxInputsUpdate with _$TaxInputsUpdate {
  const factory TaxInputsUpdate({required List<TaxInputValue> values}) =
      _TaxInputsUpdate;

  factory TaxInputsUpdate.fromJson(Map<String, Object?> json) =>
      _$TaxInputsUpdateFromJson(json);
}
