// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'regime.dart';

part 'regime_list.freezed.dart';
part 'regime_list.g.dart';

@Freezed()
abstract class RegimeList with _$RegimeList {
  const factory RegimeList({required List<Regime> items}) = _RegimeList;

  factory RegimeList.fromJson(Map<String, Object?> json) =>
      _$RegimeListFromJson(json);
}
