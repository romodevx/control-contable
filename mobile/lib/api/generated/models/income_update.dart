// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'counterparty_kind.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'operation_kind.dart';
import 'platform_activity.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';

part 'income_update.freezed.dart';
part 'income_update.g.dart';

/// Merge-patch on top-level fields: absent = unchanged, `null` clears a nullable field. `id` is immutable and not accepted. Any change re-derives the withholding (ADR-0008).
@Freezed()
abstract class IncomeUpdate with _$IncomeUpdate {
  const factory IncomeUpdate({
    @JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(includeIfNull: false, name: 'counterparty_kind')
    CounterpartyKind? counterpartyKind,
    @JsonKey(includeIfNull: false, name: 'operation_kind')
    OperationKind? operationKind,
    @JsonKey(includeIfNull: false, name: 'residential_rent')
    bool? residentialRent,
    @JsonKey(includeIfNull: false, name: 'via_platform') bool? viaPlatform,
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    PositiveCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(includeIfNull: false, name: 'iva_cents')
    NonNegativeCents? ivaCents,
    @JsonKey(includeIfNull: false, name: 'invoiced_on')
    BusinessDate? invoicedOn,
    @JsonKey(includeIfNull: false, name: 'collected_on')
    BusinessDate? collectedOn,
    @JsonKey(includeIfNull: false, name: 'platform_activity')
    PlatformActivity? platformActivity,
    @JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,
    @JsonKey(includeIfNull: false) RecordDescription? description,
  }) = _IncomeUpdate;

  factory IncomeUpdate.fromJson(Map<String, Object?> json) =>
      _$IncomeUpdateFromJson(json);
}
