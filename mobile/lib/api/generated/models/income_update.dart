// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'counterparty_type.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'platform_activity.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';

part 'income_update.freezed.dart';
part 'income_update.g.dart';

/// Merge-patch on top-level fields: absent = unchanged, `null` clears a nullable field. `id` is immutable and not accepted.
@Freezed()
abstract class IncomeUpdate with _$IncomeUpdate {
  const factory IncomeUpdate({
    @JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(includeIfNull: false, name: 'counterparty_type')
    CounterpartyType? counterpartyType,
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    PositiveCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(includeIfNull: false, name: 'iva_cents')
    NonNegativeCents? ivaCents,
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
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
