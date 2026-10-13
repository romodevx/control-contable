// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'deductibility_hint_deductibility.dart';
import 'regime_code.dart';

part 'deductibility_hint.freezed.dart';
part 'deductibility_hint.g.dart';

@Freezed()
abstract class DeductibilityHint with _$DeductibilityHint {
  const factory DeductibilityHint({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    required DeductibilityHintDeductibility deductibility,
    String? note,
    @JsonKey(name: 'legal_basis') String? legalBasis,
  }) = _DeductibilityHint;

  factory DeductibilityHint.fromJson(Map<String, Object?> json) =>
      _$DeductibilityHintFromJson(json);
}
