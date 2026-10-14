// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'regime_code.dart';

part 'regime_incompatibility.freezed.dart';
part 'regime_incompatibility.g.dart';

/// A regime that cannot be held in overlapping periods (`regime_compatibility`).
@Freezed()
abstract class RegimeIncompatibility with _$RegimeIncompatibility {
  const factory RegimeIncompatibility({
    required RegimeCode code,

    /// Plain-language reason shown to the user.
    required String reason,
  }) = _RegimeIncompatibility;

  factory RegimeIncompatibility.fromJson(Map<String, Object?> json) =>
      _$RegimeIncompatibilityFromJson(json);
}
