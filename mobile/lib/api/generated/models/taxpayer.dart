// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'taxpayer_regime.dart';
import 'uuid.dart';
import 'warning.dart';

part 'taxpayer.freezed.dart';
part 'taxpayer.g.dart';

@Freezed()
abstract class Taxpayer with _$Taxpayer {
  const factory Taxpayer({
    required Uuid id,
    @JsonKey(name: 'person_type') required PersonType personType,
    @JsonKey(name: 'display_name') required String displayName,

    /// Name part, six asterisks for the date, then homoclave and check digit (física `AAAA******XX9`, moral `AAA******XX9`). The full RFC is never returned.
    @JsonKey(name: 'rfc_masked') required String rfcMasked,

    /// Every regime the taxpayer holds or held (ended ones carry `end_period`), ordered by `start_period`, then code.
    required List<TaxpayerRegime> regimes,
    @JsonKey(includeIfNull: true, name: 'archived_at')
    required DateTime? archivedAt,
    @JsonKey(name: 'created_at') required DateTime createdAt,

    /// Non-blocking notices about the input just written. Present only on `createTaxpayer` (201 and 200) and `updateTaxpayer` responses, and only when non-empty; never stored, never on reads.
    @JsonKey(includeIfNull: false) List<Warning>? warnings,
  }) = _Taxpayer;

  factory Taxpayer.fromJson(Map<String, Object?> json) =>
      _$TaxpayerFromJson(json);
}
