// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'disclaimer_code.dart';

part 'disclaimer.freezed.dart';
part 'disclaimer.g.dart';

/// Text is versioned with the tax parameter set (ADR-0002).
@Freezed()
abstract class Disclaimer with _$Disclaimer {
  const factory Disclaimer({
    required DisclaimerCode code,
    required String text,
  }) = _Disclaimer;

  factory Disclaimer.fromJson(Map<String, Object?> json) =>
      _$DisclaimerFromJson(json);
}
