// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

part 'delete_account_request.freezed.dart';
part 'delete_account_request.g.dart';

/// Current password, re-entered to confirm the irreversible deletion (D4).
@Freezed()
abstract class DeleteAccountRequest with _$DeleteAccountRequest {
  const factory DeleteAccountRequest({required String password}) =
      _DeleteAccountRequest;

  factory DeleteAccountRequest.fromJson(Map<String, Object?> json) =>
      _$DeleteAccountRequestFromJson(json);
}
