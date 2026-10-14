// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// What was sold or which service was given (MODULE-withholdings §2.1 Q2, §2.2 E3). `rent_real_estate` is the one that asks `residential_rent`; `ground_freight` is road freight of goods (IVA withheld 4 % of the price); `scrap` is desperdicios y reciclables (IVA withheld 100 %); `other` never withholds IVA. There is no subcontracting value: LIVA 1-A fr. IV (6 %) was repealed in 2021 (owner decision 4). A platform sale is not an operation kind on the wire: it is `via_platform: true`.
@JsonEnum()
enum OperationKind {
  @JsonValue('professional_services')
  professionalServices('professional_services'),
  @JsonValue('business_services')
  businessServices('business_services'),
  @JsonValue('rent_real_estate')
  rentRealEstate('rent_real_estate'),
  @JsonValue('rent_movable')
  rentMovable('rent_movable'),
  @JsonValue('sale_of_goods')
  saleOfGoods('sale_of_goods'),
  @JsonValue('ground_freight')
  groundFreight('ground_freight'),
  @JsonValue('commission')
  commission('commission'),
  @JsonValue('scrap')
  scrap('scrap'),
  @JsonValue('other')
  other('other'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const OperationKind(this.json);

  factory OperationKind.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<OperationKind> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
