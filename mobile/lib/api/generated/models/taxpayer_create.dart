// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'rfc_input.dart';
import 'taxpayer_regime_create.dart';
import 'uuid.dart';

part 'taxpayer_create.freezed.dart';
part 'taxpayer_create.g.dart';

@Freezed()
abstract class TaxpayerCreate with _$TaxpayerCreate {
  const factory TaxpayerCreate({
    @JsonKey(name: 'person_type') required PersonType personType,
    required RfcInput rfc,

    /// Trimmed and NFC-normalized; no invisible characters (control, format such as zero-width or bidi controls, line or paragraph separators; a zero-width joiner inside an emoji sequence is allowed): `400 validation_failed`, field code `invisible_characters`.
    @JsonKey(name: 'display_name') required String displayName,

    /// At least one regime; each code at most once (`regimes.N.regime_code` / `duplicate`).
    required List<TaxpayerRegimeCreate> regimes,

    /// Optional client-generated id (idempotent retries).
    @JsonKey(includeIfNull: false) Uuid? id,
  }) = _TaxpayerCreate;

  factory TaxpayerCreate.fromJson(Map<String, Object?> json) =>
      _$TaxpayerCreateFromJson(json);
}
