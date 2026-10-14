// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'warning.dart';
import 'withholding.dart';

part 'withholding_preview.freezed.dart';
part 'withholding_preview.g.dart';

/// What the matching create would derive for the same facts; nothing is stored. `warnings` holds the withholding warnings only: `withholding_cpa_unverified`, `foreign_counterparty_not_supported`, `supplier_regime_unknown`, `deduction_requires_remittance`.
@Freezed()
abstract class WithholdingPreview with _$WithholdingPreview {
  const factory WithholdingPreview({
    required Withholding withholding,
    required List<Warning> warnings,
  }) = _WithholdingPreview;

  factory WithholdingPreview.fromJson(Map<String, Object?> json) =>
      _$WithholdingPreviewFromJson(json);
}
