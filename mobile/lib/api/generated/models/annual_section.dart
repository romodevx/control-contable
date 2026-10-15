// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'annual_missing_input.dart';
import 'annual_note.dart';
import 'annual_obligation.dart';
import 'annual_section_kind.dart';
import 'annual_section_status.dart';
import 'annual_tax.dart';
import 'cpa_topic.dart';
import 'parameter_set_ref.dart';
import 'pm_annual_extras.dart';
import 'regime_code.dart';
import 'resico_pf_check.dart';

part 'annual_section.freezed.dart';
part 'annual_section.g.dart';

/// One section of the annual estimate (MODULE-annual §2.1). `tax` is set for `pf_art152`, `pm_601` and `pm_resico` (also when `incomplete`), `null` for `pf_resico_check` and `explanation`. `pm_extras` only on PM sections, `resico_check` only on `pf_resico_check`. `explanation` is the text of an `explanation` or `pf_resico_check` section (or an informative remark), `null` otherwise. `obligation` is `null` for an `explanation` section; `obligation_reasons` are the plain-words reasons that decided it (e.g. "Tuviste dos patrones al mismo tiempo"). `deadline` is the section's own due date from the parameters (PF 2027-04-30, PM 2027-03-31, 603 informative 2027-02-15), moved to the next business day (CFF 12 ⚠️); `null` when nothing is filed for it.
@Freezed()
abstract class AnnualSection with _$AnnualSection {
  const factory AnnualSection({
    required AnnualSectionKind kind,
    required List<RegimeCode> regimes,
    required AnnualSectionStatus status,
    @JsonKey(name: 'missing_inputs')
    required List<AnnualMissingInput> missingInputs,
    @JsonKey(includeIfNull: true) required AnnualObligation? obligation,
    @JsonKey(name: 'obligation_reasons')
    required List<String> obligationReasons,
    @JsonKey(includeIfNull: true) required String? deadline,
    @JsonKey(includeIfNull: true) required AnnualTax? tax,
    @JsonKey(includeIfNull: true, name: 'pm_extras')
    required PmAnnualExtras? pmExtras,
    @JsonKey(includeIfNull: true, name: 'resico_check')
    required ResicoPfCheck? resicoCheck,
    @JsonKey(includeIfNull: true) required String? explanation,
    required List<AnnualNote> notes,

    /// A rule of this section is not validated by a CPA yet (⚠️, strong banner).
    @JsonKey(name: 'cpa_review_pending') required bool cpaReviewPending,
    @JsonKey(name: 'cpa_topics') required List<CpaTopic> cpaTopics,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
  }) = _AnnualSection;

  factory AnnualSection.fromJson(Map<String, Object?> json) =>
      _$AnnualSectionFromJson(json);
}
