// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'verification_sent.freezed.dart';
part 'verification_sent.g.dart';

/// Generic acknowledgement of `register` and `resendVerification`; identical whatever the address's state.
@Freezed()
abstract class VerificationSent with _$VerificationSent {
  const factory VerificationSent({required String status}) = _VerificationSent;

  factory VerificationSent.fromJson(Map<String, Object?> json) =>
      _$VerificationSentFromJson(json);
}
