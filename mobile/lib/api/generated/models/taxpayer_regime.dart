// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'income_basis.dart';
import 'period.dart';
import 'regime_code.dart';
import 'regime_options.dart';
import 'treatment.dart';
import 'warning.dart';

part 'taxpayer_regime.freezed.dart';
part 'taxpayer_regime.g.dart';

/// A regime the taxpayer holds over `[start_period, end_period]` (open end while `end_period` is `null`). `options` always carries every option of the regime, defaults included. `name`, `treatment` and `income_basis` are the catalog's (`Regime`) for the taxpayer's person type, embedded so the record forms need no catalog round-trip.
@Freezed()
abstract class TaxpayerRegime with _$TaxpayerRegime {
  const factory TaxpayerRegime({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    required String name,
    required Treatment treatment,
    @JsonKey(name: 'income_basis') required IncomeBasis incomeBasis,
    required RegimeOptions options,
    @JsonKey(name: 'start_period') required Period startPeriod,
    @JsonKey(includeIfNull: true, name: 'end_period')
    required Period? endPeriod,

    /// Non-blocking notices about the regime just added. Present only on the `addTaxpayerRegime` response, and only when non-empty; never stored, never on reads or inside `Taxpayer.regimes`.
    @JsonKey(includeIfNull: false) List<Warning>? warnings,
  }) = _TaxpayerRegime;

  factory TaxpayerRegime.fromJson(Map<String, Object?> json) =>
      _$TaxpayerRegimeFromJson(json);
}
