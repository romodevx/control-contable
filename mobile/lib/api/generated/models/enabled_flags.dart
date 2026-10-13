// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'module_key.dart';

part 'enabled_flags.freezed.dart';
part 'enabled_flags.g.dart';

@Freezed()
abstract class EnabledFlags with _$EnabledFlags {
  const factory EnabledFlags({required List<ModuleKey> enabled}) =
      _EnabledFlags;

  factory EnabledFlags.fromJson(Map<String, Object?> json) =>
      _$EnabledFlagsFromJson(json);
}
