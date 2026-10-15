// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

/// LISR 151 fr. I–VIII and the colegiaturas decree (MODULE-annual §2.6). Capped group (min(5 UMA anuales, 15 % of total income)): medical, funeral, donation, mortgage_real_interest, health_insurance, school_transport, local_payroll_tax; own caps: funeral (1 UMA anual), donation (7 %; government 4 %), retirement_contribution (min(10 %, 5 UMA)), tuition (per student and level); uncapped: medical_disability. Not deductible in cash: medical, medical_disability, school_transport, tuition.
@JsonEnum()
enum DeductionCategory {
  @JsonValue('medical')
  medical('medical'),
  @JsonValue('medical_disability')
  medicalDisability('medical_disability'),
  @JsonValue('funeral')
  funeral('funeral'),
  @JsonValue('donation')
  donation('donation'),
  @JsonValue('mortgage_real_interest')
  mortgageRealInterest('mortgage_real_interest'),
  @JsonValue('retirement_contribution')
  retirementContribution('retirement_contribution'),
  @JsonValue('health_insurance')
  healthInsurance('health_insurance'),
  @JsonValue('school_transport')
  schoolTransport('school_transport'),
  @JsonValue('local_payroll_tax')
  localPayrollTax('local_payroll_tax'),
  @JsonValue('tuition')
  tuition('tuition'),

  /// Default value for all unparsed values, allows backward compatibility when adding new values on the backend.
  $unknown(null);

  const DeductionCategory(this.json);

  factory DeductionCategory.fromJson(String json) =>
      values.firstWhere((e) => e.json == json, orElse: () => $unknown);

  final String? json;

  @override
  String toString() => json?.toString() ?? super.toString();

  /// Returns all defined enum values excluding the $unknown value.
  static List<DeductionCategory> get $valuesDefined =>
      values.where((value) => value != $unknown).toList();
}
