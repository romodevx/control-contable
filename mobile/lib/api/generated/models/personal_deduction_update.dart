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

part 'personal_deduction_update.freezed.dart';
part 'personal_deduction_update.g.dart';

/// Merge-patch; `null` clears a nullable field.
@Freezed()
abstract class PersonalDeductionUpdate with _$PersonalDeductionUpdate {
  const factory PersonalDeductionUpdate({
    @JsonKey(includeIfNull: false) DeductionCategory? category,
    @JsonKey(includeIfNull: false, name: 'amount_cents')
    PositiveCents? amountCents,
    @JsonKey(includeIfNull: false, name: 'paid_on') String? paidOn,
    @JsonKey(includeIfNull: false, name: 'payment_method')
    DeductionPaymentMethod? paymentMethod,
    @JsonKey(includeIfNull: false) DeductionBeneficiary? beneficiary,
    @JsonKey(includeIfNull: false, name: 'donee_type') DoneeType? doneeType,
    @JsonKey(includeIfNull: false, name: 'student_label')
    StudentLabel? studentLabel,
    @JsonKey(includeIfNull: false, name: 'tuition_level')
    TuitionLevel? tuitionLevel,
    @JsonKey(includeIfNull: false, name: 'scholarship_cents')
    NonNegativeCents? scholarshipCents,
    @JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,
  }) = _PersonalDeductionUpdate;

  factory PersonalDeductionUpdate.fromJson(Map<String, Object?> json) =>
      _$PersonalDeductionUpdateFromJson(json);
}
