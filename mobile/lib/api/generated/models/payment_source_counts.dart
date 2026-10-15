// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'payment_source_counts.freezed.dart';
part 'payment_source_counts.g.dart';

/// How many regime-months of provisional payments came from each source (MODULE-annual §2.5): confirmed here (`filed`), the M5 tax input (`tax_input`), or the app's own monthly estimate (`estimated_by_app`, which raises `payments_estimated`).
@Freezed()
abstract class PaymentSourceCounts with _$PaymentSourceCounts {
  const factory PaymentSourceCounts({
    @JsonKey(name: 'filed_months') required int filedMonths,
    @JsonKey(name: 'tax_input_months') required int taxInputMonths,
    @JsonKey(name: 'estimated_by_app_months') required int estimatedByAppMonths,
  }) = _PaymentSourceCounts;

  factory PaymentSourceCounts.fromJson(Map<String, Object?> json) =>
      _$PaymentSourceCountsFromJson(json);
}
