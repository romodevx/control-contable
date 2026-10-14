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
import 'uuid.dart';
import 'warning.dart';

part 'income.freezed.dart';
part 'income.g.dart';

@Freezed()
abstract class Income with _$Income {
  const factory Income({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'counterparty_type')
    required CounterpartyType counterpartyType,
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,
    @JsonKey(name: 'iva_withheld_cents')
    required NonNegativeCents ivaWithheldCents,
    @JsonKey(includeIfNull: true, name: 'invoiced_on')
    required BusinessDate? invoicedOn,
    @JsonKey(includeIfNull: true, name: 'collected_on')
    required BusinessDate? collectedOn,
    @JsonKey(includeIfNull: true, name: 'platform_activity')
    required PlatformActivity? platformActivity,
    @JsonKey(includeIfNull: true, name: 'cfdi_uuid')
    required CfdiUuid? cfdiUuid,
    required RecordDescription description,

    /// Accrual income with `collected_on` still `null` ("Por cobrar"); counts for IVA only once collected.
    @JsonKey(name: 'pending_collection') required bool pendingCollection,

    /// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes: `duplicate_cfdi_uuid` (`cfdi_uuid`), `withholding_differs_from_rate` (`isr_withheld_cents`, `iva_withheld_cents`).
    required List<Warning> warnings,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _Income;

  factory Income.fromJson(Map<String, Object?> json) => _$IncomeFromJson(json);
}
