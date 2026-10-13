// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'legal_document.dart';

part 'legal_document_list.freezed.dart';
part 'legal_document_list.g.dart';

@Freezed()
abstract class LegalDocumentList with _$LegalDocumentList {
  const factory LegalDocumentList({
    /// The current version of each kind.
    required List<LegalDocument> items,
  }) = _LegalDocumentList;

  factory LegalDocumentList.fromJson(Map<String, Object?> json) =>
      _$LegalDocumentListFromJson(json);
}
