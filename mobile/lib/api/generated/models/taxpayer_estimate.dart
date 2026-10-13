// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'disclaimer.dart';
import 'estimate_totals.dart';
import 'legend.dart';
import 'period.dart';
import 'regime_estimate.dart';
import 'uuid.dart';

part 'taxpayer_estimate.freezed.dart';
part 'taxpayer_estimate.g.dart';

@Freezed()
abstract class TaxpayerEstimate with _$TaxpayerEstimate {
  const factory TaxpayerEstimate({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    required Period period,
    @JsonKey(name: 'due_date') required DateTime dueDate,
    required Legend legend,
    required List<Disclaimer> disclaimers,
    required List<RegimeEstimate> regimes,
    required EstimateTotals totals,
    @JsonKey(name: 'engine_version') required String engineVersion,
  }) = _TaxpayerEstimate;

  factory TaxpayerEstimate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerEstimateFromJson(json);
}
