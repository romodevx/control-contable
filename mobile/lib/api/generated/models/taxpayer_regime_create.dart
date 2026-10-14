// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'period.dart';
import 'regime_code.dart';
import 'regime_options.dart';

part 'taxpayer_regime_create.freezed.dart';
part 'taxpayer_regime_create.g.dart';

@Freezed()
abstract class TaxpayerRegimeCreate with _$TaxpayerRegimeCreate {
  const factory TaxpayerRegimeCreate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(includeIfNull: false) RegimeOptions? options,

    /// Defaults to the current month (America/Mexico_City); `2022-01` … current month.
    @JsonKey(includeIfNull: false, name: 'start_period') Period? startPeriod,
  }) = _TaxpayerRegimeCreate;

  factory TaxpayerRegimeCreate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerRegimeCreateFromJson(json);
}
