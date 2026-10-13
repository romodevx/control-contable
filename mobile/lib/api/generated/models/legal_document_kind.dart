// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

@JsonEnum()
enum LegalDocumentKind {
  @JsonValue('privacy_notice_integral')
  privacyNoticeIntegral('privacy_notice_integral'),
  @JsonValue('privacy_notice_simplified')
  privacyNoticeSimplified('privacy_notice_simplified'),
  @JsonValue('terms')
  terms('terms'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const LegalDocumentKind(this.json);

  factory LegalDocumentKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<LegalDocumentKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
