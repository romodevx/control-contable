// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'filed_payment_entry.dart';

part 'filed_payments_update.freezed.dart';
part 'filed_payments_update.g.dart';

/// The complete set of confirmed months (replace, not merge); "Confirmar todos" sends the 12 prefilled values.
@Freezed()
abstract class FiledPaymentsUpdate with _$FiledPaymentsUpdate {
  const factory FiledPaymentsUpdate({required List<FiledPaymentEntry> months}) =
      _FiledPaymentsUpdate;

  factory FiledPaymentsUpdate.fromJson(Map<String, Object?> json) =>
      _$FiledPaymentsUpdateFromJson(json);
}
