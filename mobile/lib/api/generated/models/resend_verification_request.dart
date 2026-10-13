// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'resend_verification_request.freezed.dart';
part 'resend_verification_request.g.dart';

@Freezed()
abstract class ResendVerificationRequest with _$ResendVerificationRequest {
  const factory ResendVerificationRequest({
    /// Trimmed, lower-cased and NFC-normalised server-side before lookup or hashing.
    required String email,
  }) = _ResendVerificationRequest;

  factory ResendVerificationRequest.fromJson(Map<String, Object?> json) =>
      _$ResendVerificationRequestFromJson(json);
}
