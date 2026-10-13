// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'register_accepted.freezed.dart';
part 'register_accepted.g.dart';

@Freezed()
abstract class RegisterAccepted with _$RegisterAccepted {
  const factory RegisterAccepted({required String status}) = _RegisterAccepted;

  factory RegisterAccepted.fromJson(Map<String, Object?> json) =>
      _$RegisterAcceptedFromJson(json);
}
