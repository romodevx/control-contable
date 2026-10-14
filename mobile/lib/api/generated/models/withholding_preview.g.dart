// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'withholding_preview.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_WithholdingPreview _$WithholdingPreviewFromJson(Map<String, dynamic> json) =>
    _WithholdingPreview(
      withholding: Withholding.fromJson(
        json['withholding'] as Map<String, dynamic>,
      ),
      warnings: (json['warnings'] as List<dynamic>)
          .map((e) => Warning.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WithholdingPreviewToJson(_WithholdingPreview instance) =>
    <String, dynamic>{
      'withholding': instance.withholding,
      'warnings': instance.warnings,
    };
