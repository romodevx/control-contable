// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'user_update.freezed.dart';
part 'user_update.g.dart';

@Freezed()
abstract class UserUpdate with _$UserUpdate {
  const factory UserUpdate({String? name}) = _UserUpdate;

  factory UserUpdate.fromJson(Map<String, Object?> json) =>
      _$UserUpdateFromJson(json);
}
