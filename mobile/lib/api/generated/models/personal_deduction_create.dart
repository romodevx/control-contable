// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'cfdi_uuid.dart';
import 'deduction_beneficiary.dart';
import 'deduction_category.dart';
import 'deduction_payment_method.dart';
import 'donee_type.dart';
import 'non_negative_cents.dart';
import 'positive_cents.dart';
import 'student_label.dart';
import 'tuition_level.dart';
import 'uuid.dart';

part 'personal_deduction_create.freezed.dart';
part 'personal_deduction_create.g.dart';

/// Category first, then only that category's fields (`donee_type` for `donation`; `student_label`, `tuition_level`, `scholarship_cents` for `tuition`). `paid_on` of an annual constancia (mortgage interest, retirement) defaults in the form to December 31.
@Freezed()
abstract class PersonalDeductionCreate with _$PersonalDeductionCreate {
  const factory PersonalDeductionCreate({
    required DeductionCategory category,
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'paid_on') required String paidOn,
    @JsonKey(name: 'payment_method')
    required DeductionPaymentMethod paymentMethod,
    required DeductionBeneficiary beneficiary,
    @JsonKey(includeIfNull: false) Uuid? id,
    @JsonKey(includeIfNull: false, name: 'donee_type') DoneeType? doneeType,
    @JsonKey(includeIfNull: false, name: 'student_label')
    StudentLabel? studentLabel,
    @JsonKey(includeIfNull: false, name: 'tuition_level')
    TuitionLevel? tuitionLevel,
    @JsonKey(includeIfNull: false, name: 'scholarship_cents')
    NonNegativeCents? scholarshipCents,
    @JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,
  }) = _PersonalDeductionCreate;

  factory PersonalDeductionCreate.fromJson(Map<String, Object?> json) =>
      _$PersonalDeductionCreateFromJson(json);
}
