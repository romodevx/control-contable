// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// `cash` drives the cash-payment deductibility rule (`cash_over_threshold`).
@JsonEnum()
enum PaymentMethod {
  @JsonValue('cash')
  cash('cash'),
  @JsonValue('transfer')
  transfer('transfer'),
  @JsonValue('card_debit')
  cardDebit('card_debit'),
  @JsonValue('card_credit')
  cardCredit('card_credit'),
  @JsonValue('check')
  check('check'),
  @JsonValue('e_wallet')
  eWallet('e_wallet'),
  @JsonValue('other')
  other('other'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const PaymentMethod(this.json);

  factory PaymentMethod.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<PaymentMethod> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
