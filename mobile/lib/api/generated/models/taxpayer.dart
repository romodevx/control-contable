// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'person_type.dart';
import 'taxpayer_regime.dart';
import 'uuid.dart';

part 'taxpayer.freezed.dart';
part 'taxpayer.g.dart';

@Freezed()
abstract class Taxpayer with _$Taxpayer {
  const factory Taxpayer({
    required Uuid id,
    @JsonKey(name: 'person_type') required PersonType personType,
    @JsonKey(name: 'display_name') required String displayName,

    /// The full RFC is never returned.
    @JsonKey(name: 'rfc_masked') required String rfcMasked,
    required List<TaxpayerRegime> regimes,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(includeIfNull: false, name: 'archived_at') DateTime? archivedAt,
  }) = _Taxpayer;

  factory Taxpayer.fromJson(Map<String, Object?> json) =>
      _$TaxpayerFromJson(json);
}
