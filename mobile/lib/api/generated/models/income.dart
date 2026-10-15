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
import 'warning.dart';
import 'withholding.dart';

part 'income.freezed.dart';
part 'income.g.dart';

/// `withholding` is the derived withholding (read-only, ADR-0008): sent for every income written since 0.11.3 (WH-06), absent only on a legacy row until the WH-08 backfill derives it. (A `$ref` keeps no sibling description: the strict response validator would drop the referenced properties.)
@Freezed()
abstract class Income with _$Income {
  const factory Income({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'regime_code') required RegimeCode regimeCode,
    @JsonKey(name: 'counterparty_type')
    @Deprecated('This is marked as deprecated')
    required CounterpartyType counterpartyType,
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'iva_rate') required IvaRate ivaRate,
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,

    /// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
    @JsonKey(name: 'isr_withheld_cents')
    required NonNegativeCents isrWithheldCents,

    /// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
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

    /// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes: `duplicate_cfdi_uuid` (`cfdi_uuid`), `provisional_parameters` (`collected_on` / `invoiced_on`), `withholding_cpa_unverified` (`withholding`), `foreign_counterparty_not_supported` (`counterparty_kind`). `withholding_differs_from_rate` is retired: never sent since 0.11.3 (no typed amount is left to compare).
    required List<Warning> warnings,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,

    /// Who was billed. Present on every income written since 0.11.3 (WH-06); a legacy row gets it from the WH-08 backfill.
    @JsonKey(includeIfNull: false, name: 'counterparty_kind')
    CounterpartyKind? counterpartyKind,

    /// As sent, or the default stored when the question was hidden (last used in the regime, else the regime default). Present on every income written since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'operation_kind')
    OperationKind? operationKind,

    /// Always sent since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'residential_rent')
    bool? residentialRent,

    /// Always sent since 0.11.3.
    @JsonKey(includeIfNull: false, name: 'via_platform') bool? viaPlatform,
    @JsonKey(includeIfNull: false) Withholding? withholding,
  }) = _Income;

  factory Income.fromJson(Map<String, Object?> json) => _$IncomeFromJson(json);
}
