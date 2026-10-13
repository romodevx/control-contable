// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'fiscal_year.dart';

part 'parameter_set_ref.freezed.dart';
part 'parameter_set_ref.g.dart';

@Freezed()
abstract class ParameterSetRef with _$ParameterSetRef {
  const factory ParameterSetRef({
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required int version,
  }) = _ParameterSetRef;

  factory ParameterSetRef.fromJson(Map<String, Object?> json) =>
      _$ParameterSetRefFromJson(json);
}
