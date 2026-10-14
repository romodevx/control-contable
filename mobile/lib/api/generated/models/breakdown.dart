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

/// Header: `regime.parameter_set` ("Parámetros 2026 v1") and `engine_version` ("Motor 1.0.0"). `legend` and `disclaimers` travel with every breakdown, so the warnings are visible wherever the figures are (MODULE-tax-engine §2.3, §2.9). `disclaimers` are the estimate's disclaimers that bear on these lines, in the same order (`not_tax_advice` first, then `strong`, `warning`, `info`, then code): those raised by this regime, plus those raised by any regime's contribution to the consolidated IVA lines shown here. In particular `cpa_review_pending` (severity `strong`, the banner above the lines) is present whenever `regime.cpa_review_pending` is `true` or a line has `needs_cpa_review`, including the consolidated IVA line when two or more regimes contribute IVA (`cross_regime_iva_netting`, OQ-T2), and `iva_mixed_activity` is present whenever the consolidated IVA took any regime's creditable IVA in full. `incomplete_inputs` is present when `regime.status` is `incomplete`.
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
