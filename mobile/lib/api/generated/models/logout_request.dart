// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'installation_id.dart';

part 'logout_request.freezed.dart';
part 'logout_request.g.dart';

/// `logout`'s body: `RefreshRequest` plus the optional push `device_id` (0.14.0). Without `device_id` it is exactly a `RefreshRequest`.
@Freezed()
abstract class LogoutRequest with _$LogoutRequest {
  const factory LogoutRequest({
    /// The opaque refresh token from `TokenPair`. A malformed value answers `401`, not `400`.
    @JsonKey(name: 'refresh_token') required String refreshToken,
    @JsonKey(includeIfNull: false, name: 'device_id') InstallationId? deviceId,
  }) = _LogoutRequest;

  factory LogoutRequest.fromJson(Map<String, Object?> json) =>
      _$LogoutRequestFromJson(json);
}
