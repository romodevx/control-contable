// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'taxpayer_update.freezed.dart';
part 'taxpayer_update.g.dart';

@Freezed()
abstract class TaxpayerUpdate with _$TaxpayerUpdate {
  const factory TaxpayerUpdate({
    @JsonKey(name: 'display_name') required String displayName,
  }) = _TaxpayerUpdate;

  factory TaxpayerUpdate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerUpdateFromJson(json);
}
