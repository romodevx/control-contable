// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'counterparty_kind.dart';
import 'operation_kind.dart';

part 'income_last_used.freezed.dart';
part 'income_last_used.g.dart';

/// The withholding facts of the taxpayer's last income in this regime, so the form preselects them (MODULE-withholdings §2.1). `counterparty_kind` is `null` the first time (no preselection, OQ-W2); `operation_kind` then falls back to the regime default (612 and 626 PF `professional_services`, 606 `rent_real_estate`, 621 `sale_of_goods`, any persona moral `business_services`); `via_platform` starts `true` in 625.
@Freezed()
abstract class IncomeLastUsed with _$IncomeLastUsed {
  const factory IncomeLastUsed({
    @JsonKey(includeIfNull: true, name: 'counterparty_kind')
    required CounterpartyKind? counterpartyKind,
    @JsonKey(name: 'operation_kind') required OperationKind operationKind,
    @JsonKey(name: 'residential_rent') required bool residentialRent,
    @JsonKey(name: 'via_platform') required bool viaPlatform,
  }) = _IncomeLastUsed;

  factory IncomeLastUsed.fromJson(Map<String, Object?> json) =>
      _$IncomeLastUsedFromJson(json);
}
