// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'taxpayer_regime_create.dart';
import 'uuid.dart';

part 'taxpayer_create.freezed.dart';
part 'taxpayer_create.g.dart';

@Freezed()
abstract class TaxpayerCreate with _$TaxpayerCreate {
  const factory TaxpayerCreate({
    @JsonKey(name: 'person_type') required PersonType personType,

    /// 13 chars for fisica, 12 for moral; check digit validated server-side.
    required String rfc,
    @JsonKey(name: 'display_name') required String displayName,
    required List<TaxpayerRegimeCreate> regimes,

    /// Optional client-generated id (idempotent retries).
    Uuid? id,
  }) = _TaxpayerCreate;

  factory TaxpayerCreate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerCreateFromJson(json);
}
