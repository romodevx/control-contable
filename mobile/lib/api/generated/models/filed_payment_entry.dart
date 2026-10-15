// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'non_negative_cents.dart';

part 'filed_payment_entry.freezed.dart';
part 'filed_payment_entry.g.dart';

@Freezed()
abstract class FiledPaymentEntry with _$FiledPaymentEntry {
  const factory FiledPaymentEntry({
    required int month,
    @JsonKey(name: 'amount_cents') required NonNegativeCents amountCents,
  }) = _FiledPaymentEntry;

  factory FiledPaymentEntry.fromJson(Map<String, Object?> json) =>
      _$FiledPaymentEntryFromJson(json);
}
