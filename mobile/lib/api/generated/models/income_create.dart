// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'business_date.dart';
import 'cfdi_uuid.dart';
import 'counterparty_kind.dart';
import 'counterparty_type.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'operation_kind.dart';
import 'platform_activity.dart';
import 'positive_cents.dart';
import 'record_description.dart';
import 'regime_code.dart';
import 'uuid.dart';

part 'income_create.freezed.dart';
part 'income_create.g.dart';

/// `user_id` comes from the session and `taxpayer_id` from the path, never from the body. The date the regime's basis needs is required by the server: `collected_on` (cash regimes) or `invoiced_on` (accrual regimes). Withholdings are **derived** from `counterparty_kind`, `operation_kind`, `residential_rent` and `via_platform` (ADR-0008); no request carries an amount. **0.11.2 (WH-06):** `counterparty_kind` is required (`400 required`); `isr_withheld_cents` / `iva_withheld_cents` with any value answer `400 additionalProperties`; the deprecated `counterparty_type` is still required by the schema and ignored (its stored value mirrors `counterparty_kind` and `via_platform`). The deprecated members leave the schema with WH-11, which regenerates the Dart client with the new form.
@Freezed()
abstract class IncomeCreate with _$IncomeCreate {
  const factory IncomeCreate({
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'counterparty_type')
    @Deprecated('This is marked as deprecated')
    required CounterpartyType counterpartyType,

    /// Subtotal before IVA, MXN centavos.
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,

    /// IVA charged; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,

    /// Client-generated idempotency key (UUIDv7 recommended), once per form session.
    @JsonKey(includeIfNull: false) Uuid? id,

    /// Required since 0.11.2 (`400 required` when missing).
    @JsonKey(includeIfNull: false, name: 'counterparty_kind')
    CounterpartyKind? counterpartyKind,

    /// Asked only when the payer can withhold (`pm`, `gov_*`, not via a platform); omitted → the last value used in this regime, else the regime default (`IncomeLastUsed`).
    @JsonKey(includeIfNull: false, name: 'operation_kind')
    OperationKind? operationKind,

    /// Housing rent (`rent_real_estate` only; default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0` (LIVA 20 fr. II), else `422` `residential_rent_exempt` on `iva_rate`.
    @JsonKey(includeIfNull: false, name: 'residential_rent')
    bool? residentialRent,

    /// Collected through a digital platform (default `true` in 625, `false` otherwise). Allowed for 625 and persona moral regimes; `true` on any other regime answers `400` `not_for_regime`.
    @JsonKey(includeIfNull: false, name: 'via_platform') bool? viaPlatform,

    /// Retired by WH-06 (0.11.2, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,

    /// Retired by WH-06 (0.11.2, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
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
