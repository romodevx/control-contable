// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_note_code.dart';

part 'annual_note.freezed.dart';
part 'annual_note.g.dart';

/// An informative note under a section's figure (es-MX, server text, never an amount on its own).
@Freezed()
abstract class AnnualNote with _$AnnualNote {
  const factory AnnualNote({
    required AnnualNoteCode code,
    required String text,
  }) = _AnnualNote;

  factory AnnualNote.fromJson(Map<String, Object?> json) =>
      _$AnnualNoteFromJson(json);
}
