// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'counterparty_type.dart';
import 'iva_rate.dart';
import 'non_negative_cents.dart';
import 'regime_code.dart';
import 'uuid.dart';

part 'income.freezed.dart';
part 'income.g.dart';

@Freezed()
abstract class Income with _$Income {
  const factory Income({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
    @JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,
    @JsonKey(includeIfNull: false) String? description,
    @JsonKey(includeIfNull: false, name: 'counterparty_type')
    CounterpartyType? counterpartyType,

    /// Subtotal before IVA.
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    NonNegativeCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,
    @JsonKey(includeIfNull: false, name: 'iva_cents')
    NonNegativeCents? ivaCents,
    @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')
    NonNegativeCents? isrWithheldCents,
    @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')
    NonNegativeCents? ivaWithheldCents,
    @JsonKey(includeIfNull: false, name: 'invoiced_on') DateTime? invoicedOn,

    /// Required for cash-basis regimes; may be null for PM 601 (pending collection).
    @JsonKey(includeIfNull: false, name: 'collected_on') DateTime? collectedOn,
  }) = _Income;

  factory Income.fromJson(Map<String, Object?> json) => _$IncomeFromJson(json);
}
