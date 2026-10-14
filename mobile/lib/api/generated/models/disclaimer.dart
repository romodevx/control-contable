// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'disclaimer_code.dart';
import 'disclaimer_severity.dart';

part 'disclaimer.freezed.dart';
part 'disclaimer.g.dart';

/// Text is versioned with the tax parameter set (ADR-0002), so it tracks the law year; es-MX, safe to show, never an amount. Order in a response: `not_tax_advice` first, then by severity (`strong`, `warning`, `info`), then by code.
@Freezed()
abstract class Disclaimer with _$Disclaimer {
  const factory Disclaimer({
    required DisclaimerCode code,
    required DisclaimerSeverity severity,
    required String text,
  }) = _Disclaimer;

  factory Disclaimer.fromJson(Map<String, Object?> json) =>
      _$DisclaimerFromJson(json);
}
