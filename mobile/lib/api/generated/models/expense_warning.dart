// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'expense_warning_code.dart';

part 'expense_warning.freezed.dart';
part 'expense_warning.g.dart';

@Freezed()
abstract class ExpenseWarning with _$ExpenseWarning {
  const factory ExpenseWarning({
    required ExpenseWarningCode code,
    required String message,
  }) = _ExpenseWarning;

  factory ExpenseWarning.fromJson(Map<String, Object?> json) =>
      _$ExpenseWarningFromJson(json);
}
