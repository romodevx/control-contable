// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'access_summary.dart';
import 'uuid.dart';

part 'me.freezed.dart';
part 'me.g.dart';

/// The body of `getMe`: every `User` field, plus the caller's `access` (the frozen `AccessSummary`, computed by the server exactly as for `getDashboard`). A superset of `User`, so a client that reads it as a `User` keeps working.
@Freezed()
abstract class Me with _$Me {
  const factory Me({
    required Uuid id,
    required String email,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    required AccessSummary access,
    @JsonKey(includeIfNull: false) String? name,
  }) = _Me;

  factory Me.fromJson(Map<String, Object?> json) => _$MeFromJson(json);
}
