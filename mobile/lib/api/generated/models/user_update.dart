// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'user_update.freezed.dart';
part 'user_update.g.dart';

@Freezed()
abstract class UserUpdate with _$UserUpdate {
  const factory UserUpdate({
    /// Trimmed and NFC-normalized; no invisible characters (control, format such as zero-width or bidi controls, line or paragraph separators): `400 validation_failed`, field code `invisible_characters`.
    @JsonKey(includeIfNull: false) String? name,
  }) = _UserUpdate;

  factory UserUpdate.fromJson(Map<String, Object?> json) =>
      _$UserUpdateFromJson(json);
}
