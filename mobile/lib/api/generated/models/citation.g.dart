// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'citation.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_Citation _$CitationFromJson(Map<String, dynamic> json) => _Citation(
  law: json['law'] as String,
  article: json['article'] as String,
  note: json['note'] as String?,
);

Map<String, dynamic> _$CitationToJson(_Citation instance) => <String, dynamic>{
  'law': instance.law,
  'article': instance.article,
  'note': ?instance.note,
};
