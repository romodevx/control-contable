// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'legal_document.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_LegalDocument _$LegalDocumentFromJson(Map<String, dynamic> json) =>
    _LegalDocument(
      kind: LegalDocumentKind.fromJson(json['kind'] as String),
      version: json['version'] as String,
      url: json['url'] as String,
      publishedAt: DateTime.parse(json['published_at'] as String),
    );

Map<String, dynamic> _$LegalDocumentToJson(_LegalDocument instance) =>
    <String, dynamic>{
      'kind': _$LegalDocumentKindEnumMap[instance.kind]!,
      'version': instance.version,
      'url': instance.url,
      'published_at': instance.publishedAt.toIso8601String(),
    };

const _$LegalDocumentKindEnumMap = {
  LegalDocumentKind.privacyNoticeIntegral: 'privacy_notice_integral',
  LegalDocumentKind.privacyNoticeSimplified: 'privacy_notice_simplified',
  LegalDocumentKind.terms: 'terms',
  LegalDocumentKind.$unknown: r'$unknown',
};
