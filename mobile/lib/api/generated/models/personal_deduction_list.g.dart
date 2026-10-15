// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'personal_deduction_list.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_PersonalDeductionList _$PersonalDeductionListFromJson(
  Map<String, dynamic> json,
) => _PersonalDeductionList(
  fiscalYear: (json['fiscal_year'] as num).toInt(),
  items: (json['items'] as List<dynamic>)
      .map((e) => PersonalDeduction.fromJson(e as Map<String, dynamic>))
      .toList(),
  totals: (json['totals'] as List<dynamic>)
      .map((e) => DeductionCategoryTotal.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$PersonalDeductionListToJson(
  _PersonalDeductionList instance,
) => <String, dynamic>{
  'fiscal_year': instance.fiscalYear,
  'items': instance.items,
  'totals': instance.totals,
};
