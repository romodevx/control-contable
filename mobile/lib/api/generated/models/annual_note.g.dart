// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'annual_note.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_AnnualNote _$AnnualNoteFromJson(Map<String, dynamic> json) => _AnnualNote(
  code: AnnualNoteCode.fromJson(json['code'] as String),
  text: json['text'] as String,
);

Map<String, dynamic> _$AnnualNoteToJson(_AnnualNote instance) =>
    <String, dynamic>{
      'code': _$AnnualNoteCodeEnumMap[instance.code]!,
      'text': instance.text,
    };

const _$AnnualNoteCodeEnumMap = {
  AnnualNoteCode.refundOnlyPaid: 'refund_only_paid',
  AnnualNoteCode.refundNeedsEfirma: 'refund_needs_efirma',
  AnnualNoteCode.lossCarryforward: 'loss_carryforward',
  AnnualNoteCode.cuSuggestion: 'cu_suggestion',
  AnnualNoteCode.$unknown: r'$unknown',
};
