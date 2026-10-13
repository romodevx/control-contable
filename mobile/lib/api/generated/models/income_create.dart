// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'counterparty_type.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'regime_code.dart';
import 'uuid.dart';

part 'income_create.freezed.dart';
part 'income_create.g.dart';

@Freezed()
abstract class IncomeCreate with _$IncomeCreate {
  const factory IncomeCreate({
    @JsonKey(name: 'regime_code') RegimeCode? regimeCode,
    String? description,
    @JsonKey(name: 'counterparty_type') CounterpartyType? counterpartyType,

    /// Subtotal before IVA.
    @JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,
    @JsonKey(name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,
    @JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,
    @JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,
    @JsonKey(name: 'invoiced_on') DateTime? invoicedOn,

    /// Required for cash-basis regimes; may be null for PM 601 (pending collection).
    @JsonKey(name: 'collected_on') DateTime? collectedOn,
    Uuid? id,
  }) = _IncomeCreate;

  factory IncomeCreate.fromJson(Map<String, Object?> json) =>
      _$IncomeCreateFromJson(json);
}
