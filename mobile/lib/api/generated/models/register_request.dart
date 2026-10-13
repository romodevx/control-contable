// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'document_acceptance.dart';

part 'register_request.freezed.dart';
part 'register_request.g.dart';

@Freezed()
abstract class RegisterRequest with _$RegisterRequest {
  const factory RegisterRequest({
    /// Trimmed, lower-cased and NFC-normalised server-side before lookup or hashing.
    required String email,

    /// 12-128 characters; the breached/common-password check (offline list, D5) runs server-side.
    required String password,

    /// The current version of `terms` and of one privacy notice (integral or simplified), as returned by `listLegalDocuments`.
    @JsonKey(name: 'accepted_documents')
    required List<DocumentAcceptance> acceptedDocuments,
    String? name,
  }) = _RegisterRequest;

  factory RegisterRequest.fromJson(Map<String, Object?> json) =>
      _$RegisterRequestFromJson(json);
}
