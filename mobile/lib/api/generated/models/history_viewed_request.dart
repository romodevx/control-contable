// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'engine_version.dart';
import 'figures_fingerprint.dart';
import 'history_viewed_parameter_set.dart';
import 'records_fingerprint.dart';

part 'history_viewed_request.freezed.dart';
part 'history_viewed_request.g.dart';

/// What the user saw: the values of the `TaxpayerEstimate` the app rendered for the month, copied verbatim. Format-validated and only ever compared (never used for a figure, never logged).
@Freezed()
abstract class HistoryViewedRequest with _$HistoryViewedRequest {
  const factory HistoryViewedRequest({
    @JsonKey(name: 'records_fingerprint')
    required RecordsFingerprint recordsFingerprint,
    @JsonKey(name: 'figures_fingerprint')
    required FiguresFingerprint figuresFingerprint,
    @JsonKey(name: 'engine_version') required EngineVersion engineVersion,
    @JsonKey(name: 'parameter_set')
    required HistoryViewedParameterSet parameterSet,
  }) = _HistoryViewedRequest;

  factory HistoryViewedRequest.fromJson(Map<String, Object?> json) =>
      _$HistoryViewedRequestFromJson(json);
}
