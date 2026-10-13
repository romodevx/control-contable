// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'taxpayer.dart';

part 'taxpayer_list.freezed.dart';
part 'taxpayer_list.g.dart';

@Freezed()
abstract class TaxpayerList with _$TaxpayerList {
  const factory TaxpayerList({required List<Taxpayer> items}) = _TaxpayerList;

  factory TaxpayerList.fromJson(Map<String, Object?> json) =>
      _$TaxpayerListFromJson(json);
}
