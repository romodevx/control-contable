// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'tax_inputs_update.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_TaxInputsUpdate _$TaxInputsUpdateFromJson(Map<String, dynamic> json) =>
    _TaxInputsUpdate(
      values: (json['values'] as List<dynamic>)
          .map((e) => TaxInputValue.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$TaxInputsUpdateToJson(_TaxInputsUpdate instance) =>
    <String, dynamic>{'values': instance.values};
