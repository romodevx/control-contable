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

part 'income_create.freezed.dart';
part 'income_create.g.dart';

/// `user_id` comes from the session and `taxpayer_id` from the path, never from the body. The date the regime's basis needs is required by the server: `collected_on` (cash regimes) or `invoiced_on` (accrual regimes).
@Freezed()
abstract class IncomeCreate with _$IncomeCreate {
  const factory IncomeCreate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'counterparty_type')
    required CounterpartyType counterpartyType,

    /// Subtotal before IVA, MXN centavos.
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,

    /// IVA charged; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,

    /// Client-generated idempotency key (UUIDv7 recommended), once per form session.
    @JsonKey(includeIfNull: false) Uuid? id,

    /// ISR withheld by the payer (default 0); at most `amount_cents`.
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,

    /// IVA withheld by the payer (default 0); at most `iva_cents`.
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
  }) = _IncomeCreate;

  factory IncomeCreate.fromJson(Map<String, Object?> json) =>
      _$IncomeCreateFromJson(json);
}
