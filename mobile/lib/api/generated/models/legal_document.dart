// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'legal_document_kind.dart';

part 'legal_document.freezed.dart';
part 'legal_document.g.dart';

@Freezed()
abstract class LegalDocument with _$LegalDocument {
  const factory LegalDocument({
    required LegalDocumentKind kind,

    /// Opaque version label (data, not code).
    required String version,
    required String url,
    @JsonKey(name: 'published_at') required DateTime publishedAt,
  }) = _LegalDocument;

  factory LegalDocument.fromJson(Map<String, Object?> json) =>
      _$LegalDocumentFromJson(json);
}
