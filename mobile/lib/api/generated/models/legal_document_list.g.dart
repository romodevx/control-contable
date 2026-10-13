// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'legal_document_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_LegalDocumentList _$LegalDocumentListFromJson(Map<String, dynamic> json) =>
    _LegalDocumentList(
      items: (json['items'] as List<dynamic>)
          .map((e) => LegalDocument.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$LegalDocumentListToJson(_LegalDocumentList instance) =>
    <String, dynamic>{'items': instance.items};
