// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'token_pair.freezed.dart';
part 'token_pair.g.dart';

@Freezed()
abstract class TokenPair with _$TokenPair {
  const factory TokenPair({
    @JsonKey(name: 'access_token') required String accessToken,

    /// Seconds.
    @JsonKey(name: 'access_expires_in') required int accessExpiresIn,

    /// Opaque, format <user_uuid>.<secret>.
    @JsonKey(name: 'refresh_token') required String refreshToken,

    /// Seconds.
    @JsonKey(name: 'refresh_expires_in') required int refreshExpiresIn,
    @JsonKey(name: 'token_type') required String tokenType,
  }) = _TokenPair;

  factory TokenPair.fromJson(Map<String, Object?> json) =>
      _$TokenPairFromJson(json);
}
