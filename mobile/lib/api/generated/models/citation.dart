// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'citation.freezed.dart';
part 'citation.g.dart';

@Freezed()
abstract class Citation with _$Citation {
  const factory Citation({
    required String law,
    required String article,
    @JsonKey(includeIfNull: false) String? note,
  }) = _Citation;

  factory Citation.fromJson(Map<String, Object?> json) =>
      _$CitationFromJson(json);
}
