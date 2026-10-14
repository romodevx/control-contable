// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'warning_code.dart';

part 'warning.freezed.dart';
part 'warning.g.dart';

/// A non-blocking notice: the write succeeded, the client should ask the user to double-check `field`. `code` is a `WarningCode`, the one list of every warning code and the field it names (append-only; unknown codes: show `message`).
@Freezed()
abstract class Warning with _$Warning {
  const factory Warning({
    required WarningCode code,

    /// Same dotted naming as `FieldError.field`.
    required String field,

    /// Safe to show (es-MX); never contains the RFC.
    required String message,
  }) = _Warning;

  factory Warning.fromJson(Map<String, Object?> json) =>
      _$WarningFromJson(json);
}
