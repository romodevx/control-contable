// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'rfc_input.dart';

part 'taxpayer_update.freezed.dart';
part 'taxpayer_update.g.dart';

/// Person type is immutable. `rfc` corrects a mistyped RFC of the same person type (OQ-1).
@Freezed()
abstract class TaxpayerUpdate with _$TaxpayerUpdate {
  const factory TaxpayerUpdate({
    @JsonKey(includeIfNull: false, name: 'display_name') String? displayName,
    @JsonKey(includeIfNull: false) RfcInput? rfc,
  }) = _TaxpayerUpdate;

  factory TaxpayerUpdate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerUpdateFromJson(json);
}
