// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'verify_email_request.freezed.dart';
part 'verify_email_request.g.dart';

@Freezed()
abstract class VerifyEmailRequest with _$VerifyEmailRequest {
  const factory VerifyEmailRequest({
    /// Trimmed, lower-cased and NFC-normalised server-side before lookup or hashing.
    required String email,

    /// 6-digit emailed code; 15 min TTL, 5 wrong tries.
    required String code,

    /// The new account's password (D8). 12-128 characters; the breached/common-password check (offline list, D5) runs server-side.
    required String password,
    @JsonKey(includeIfNull: false) String? name,
  }) = _VerifyEmailRequest;

  factory VerifyEmailRequest.fromJson(Map<String, Object?> json) =>
      _$VerifyEmailRequestFromJson(json);
}
