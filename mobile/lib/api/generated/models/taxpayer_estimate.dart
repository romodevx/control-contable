// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'disclaimer.dart';
import 'estimate_totals.dart';
import 'figures_fingerprint.dart';
import 'legend.dart';
import 'parameter_set_ref.dart';
import 'period.dart';
import 'records_fingerprint.dart';
import 'regime_estimate.dart';
import 'uuid.dart';

part 'taxpayer_estimate.freezed.dart';
part 'taxpayer_estimate.g.dart';

/// Informational estimate of one month (derived on every read; same records, inputs, parameters and engine version → byte-identical JSON). `due_date` is day 17 of the next month (after the bimester for 621), moved to the next business day (CFF 12 ⚠️, "sujeto a confirmación"); RFC-digit grace days are not applied. `parameter_set` is the set of the period (each regime repeats it).
@Freezed()
abstract class TaxpayerEstimate with _$TaxpayerEstimate {
  const factory TaxpayerEstimate({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    required Period period,
    @JsonKey(name: 'due_date') required String dueDate,
    required Legend legend,
    required List<Disclaimer> disclaimers,
    required List<RegimeEstimate> regimes,
    required EstimateTotals totals,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
    @JsonKey(name: 'records_fingerprint')
    required RecordsFingerprint recordsFingerprint,
    @JsonKey(name: 'engine_version') required String engineVersion,

    /// Hash of the figures shown (M8-01, additive). Sent on every estimate once M8-03 ships (the app posts it, with `records_fingerprint`, `engine_version` and `parameter_set`, to `markHistoryViewed`); a client that receives no value skips marking the month viewed.
    @JsonKey(includeIfNull: false, name: 'figures_fingerprint')
    FiguresFingerprint? figuresFingerprint,
  }) = _TaxpayerEstimate;

  factory TaxpayerEstimate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerEstimateFromJson(json);
}
