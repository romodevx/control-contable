// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';
import 'regime_code.dart';
import 'regime_options.dart';
import 'treatment.dart';

part 'taxpayer_regime.freezed.dart';
part 'taxpayer_regime.g.dart';

@Freezed()
abstract class TaxpayerRegime with _$TaxpayerRegime {
  const factory TaxpayerRegime({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    required String name,
    required Treatment treatment,
    required RegimeOptions options,
    @JsonKey(name: 'start_period') required Period startPeriod,
    @JsonKey(includeIfNull: false, name: 'end_period') Period? endPeriod,
  }) = _TaxpayerRegime;

  factory TaxpayerRegime.fromJson(Map<String, Object?> json) =>
      _$TaxpayerRegimeFromJson(json);
}
