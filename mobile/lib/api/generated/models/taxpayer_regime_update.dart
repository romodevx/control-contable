// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'regime_options.dart';

part 'taxpayer_regime_update.freezed.dart';
part 'taxpayer_regime_update.g.dart';

@Freezed()
abstract class TaxpayerRegimeUpdate with _$TaxpayerRegimeUpdate {
  const factory TaxpayerRegimeUpdate({required RegimeOptions options}) =
      _TaxpayerRegimeUpdate;

  factory TaxpayerRegimeUpdate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerRegimeUpdateFromJson(json);
}
