// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'document_acceptance.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_DocumentAcceptance _$DocumentAcceptanceFromJson(Map<String, dynamic> json) =>
    _DocumentAcceptance(
      kind: LegalDocumentKind.fromJson(json['kind'] as String),
      version: json['version'] as String,
    );

Map<String, dynamic> _$DocumentAcceptanceToJson(_DocumentAcceptance instance) =>
    <String, dynamic>{
      'kind': _$LegalDocumentKindEnumMap[instance.kind]!,
      'version': instance.version,
    };

const _$LegalDocumentKindEnumMap = {
  LegalDocumentKind.privacyNoticeIntegral: 'privacy_notice_integral',
  LegalDocumentKind.privacyNoticeSimplified: 'privacy_notice_simplified',
  LegalDocumentKind.terms: 'terms',
  LegalDocumentKind.$unknown: r'$unknown',
};
