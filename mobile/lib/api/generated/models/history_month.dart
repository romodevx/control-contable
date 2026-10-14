// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'engine_version.dart';
import 'figures_fingerprint.dart';
import 'history_change_reason.dart';
import 'history_tax.dart';
import 'parameter_set_ref.dart';
import 'period.dart';
import 'records_fingerprint.dart';
import 'taxpayer_estimate_status.dart';

part 'history_month.freezed.dart';
part 'history_month.g.dart';

/// One month of `HistoryYear`. `status` uses the M5 precedence (`TaxpayerEstimateStatus`; `none` when no regime is held that month). `tax` is `null` if and only if `status` is `none` or every regime is informational (no amount shown). `has_records`: any live income or expense dated in the month, counting only enabled modules ("Sin registros en este mes" when `false`; a YTD regime may still show a figure). `is_current`: the current month (America/Mexico_City), labelled "Mes en curso", never marked. `records_fingerprint`, `figures_fingerprint`, `engine_version` and `parameter_set` are the current estimate's (opaque to the client; the app posts the values of the estimate it renders to `markHistoryViewed`). `viewed_at`: the caller's last view (server clock), `null` if never viewed. `changed_since_viewed` and `change_reason` follow `getHistory`: `change_reason` is non-null if and only if `changed_since_viewed` is `true`, which needs a view and a non-current month.
@Freezed()
abstract class HistoryMonth with _$HistoryMonth {
  const factory HistoryMonth({
    required Period period,
    required TaxpayerEstimateStatus status,
    @JsonKey(name: 'has_records') required bool hasRecords,
    @JsonKey(includeIfNull: true) required HistoryTax? tax,
    @JsonKey(name: 'due_date') required String dueDate,

    /// Any regime of the month is pending CPA review; the app shows the strong, non-dismissible banner.
    @JsonKey(name: 'cpa_review_pending') required bool cpaReviewPending,
    @JsonKey(name: 'is_current') required bool isCurrent,
    @JsonKey(name: 'records_fingerprint')
    required RecordsFingerprint recordsFingerprint,
    @JsonKey(name: 'figures_fingerprint')
    required FiguresFingerprint figuresFingerprint,
    @JsonKey(name: 'engine_version') required EngineVersion engineVersion,
    @JsonKey(name: 'parameter_set') required ParameterSetRef parameterSet,
    @JsonKey(includeIfNull: true, name: 'viewed_at')
    required DateTime? viewedAt,
    @JsonKey(name: 'changed_since_viewed') required bool changedSinceViewed,
    @JsonKey(includeIfNull: true, name: 'change_reason')
    required HistoryChangeReason? changeReason,
  }) = _HistoryMonth;

  factory HistoryMonth.fromJson(Map<String, Object?> json) =>
      _$HistoryMonthFromJson(json);
}
