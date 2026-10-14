// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'cpa_topic.dart';
import 'iva_components.dart';
import 'missing_input.dart';
import 'non_negative_cents.dart';
import 'parameter_set_ref.dart';
import 'regime_code.dart';
import 'regime_estimate_status.dart';
import 'tax_result.dart';
import 'treatment.dart';

part 'regime_estimate.freezed.dart';
part 'regime_estimate.g.dart';

/// One regime held in the period. `status` precedence: `informational` > `incomplete` > `partial` > `complete`. `complete`: every input this period needs is present. `incomplete`: an input is missing; computed with zero for it, still shown with an "Incompleto" chip and the CTA. `partial`: only some taxes are estimated (621, 622, 603, 625 without the definitive option); `explanation` says what is not. `informational`: no strategy or an informational catalog regime; `explanation` only, `isr` and `iva_components` are `null`. `missing_inputs` is filled whenever inputs are missing, whatever the status. `isr` is `null` when the regime has no ISR figure in the period (informational, partial, an odd month for 621).
@Freezed()
abstract class RegimeEstimate with _$RegimeEstimate {
  const factory RegimeEstimate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'regime_name') required String regimeName,
    required Treatment treatment,
    required RegimeEstimateStatus status,
    @JsonKey(name: 'missing_inputs') required List<MissingInput> missingInputs,

    /// Plain-language text for a partial or informational regime (or an odd 621 month); `null` otherwise.
    @JsonKey(includeIfNull: true) required String? explanation,
    @JsonKey(includeIfNull: true) required TaxResult? isr,
    @JsonKey(includeIfNull: true, name: 'iva_components')
    required IvaComponents? ivaComponents,

    /// ISR/IVA the taxpayer withheld from its suppliers in this regime (PM only).
    @JsonKey(name: 'withheld_to_remit_cents')
    required NonNegativeCents withheldToRemitCents,

    /// A rule of this regime's computation is not validated by a CPA yet (⚠️ on its lines).
    @JsonKey(name: 'cpa_review_pending') required bool cpaReviewPending,
    @JsonKey(name: 'cpa_topics') required List<CpaTopic> cpaTopics,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
  }) = _RegimeEstimate;

  factory RegimeEstimate.fromJson(Map<String, Object?> json) =>
      _$RegimeEstimateFromJson(json);
}
