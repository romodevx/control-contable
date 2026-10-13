// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'breakdown_line.dart';
import 'disclaimer.dart';
import 'legend.dart';
import 'period.dart';
import 'regime_estimate.dart';
import 'uuid.dart';

part 'breakdown.freezed.dart';
part 'breakdown.g.dart';

@Freezed()
abstract class Breakdown with _$Breakdown {
  const factory Breakdown({
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    required Period period,
    required RegimeEstimate regime,
    required List<BreakdownLine> lines,
    required Legend legend,
    required List<Disclaimer> disclaimers,
    @JsonKey(name: 'engine_version') required String engineVersion,
  }) = _Breakdown;

  factory Breakdown.fromJson(Map<String, Object?> json) =>
      _$BreakdownFromJson(json);
}
