// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'legal_document_kind.dart';

part 'document_acceptance.freezed.dart';
part 'document_acceptance.g.dart';

@Freezed()
abstract class DocumentAcceptance with _$DocumentAcceptance {
  const factory DocumentAcceptance({
    required LegalDocumentKind kind,

    /// Opaque version label (data, not code).
    required String version,
  }) = _DocumentAcceptance;

  factory DocumentAcceptance.fromJson(Map<String, Object?> json) =>
      _$DocumentAcceptanceFromJson(json);
}
