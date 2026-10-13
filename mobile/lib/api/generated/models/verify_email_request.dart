// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'verify_email_request.freezed.dart';
part 'verify_email_request.g.dart';

@Freezed()
abstract class VerifyEmailRequest with _$VerifyEmailRequest {
  const factory VerifyEmailRequest({
    required String email,
    required String code,
  }) = _VerifyEmailRequest;

  factory VerifyEmailRequest.fromJson(Map<String, Object?> json) =>
      _$VerifyEmailRequestFromJson(json);
}
