// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'cfdi_uuid.dart';
import 'deduction_beneficiary.dart';
import 'deduction_category.dart';
import 'deduction_exclusion.dart';
import 'deduction_payment_method.dart';
import 'donee_type.dart';
import 'fiscal_year.dart';
import 'non_negative_cents.dart';
import 'positive_cents.dart';
import 'student_label.dart';
import 'tuition_level.dart';
import 'uuid.dart';
import 'warning.dart';

part 'personal_deduction.freezed.dart';
part 'personal_deduction.g.dart';

/// A live personal deduction. `fiscal_year` is the year of `paid_on` (read-only). `excluded` is set when the row can never be applied (`cash_not_deductible`), `null` otherwise; caps are applied in the estimate, not here. `warnings` is filled on writes only.
@Freezed()
abstract class PersonalDeduction with _$PersonalDeduction {
  const factory PersonalDeduction({
    required Uuid id,
    @JsonKey(name: 'taxpayer_id') required Uuid taxpayerId,
    @JsonKey(name: 'fiscal_year') required FiscalYear fiscalYear,
    required DeductionCategory category,
    @JsonKey(name: 'amount_cents') required PositiveCents amountCents,
    @JsonKey(name: 'paid_on') required String paidOn,
    @JsonKey(name: 'payment_method')
    required DeductionPaymentMethod paymentMethod,
    required DeductionBeneficiary beneficiary,
    @JsonKey(includeIfNull: true, name: 'donee_type')
    required DoneeType? doneeType,
    @JsonKey(includeIfNull: true, name: 'student_label')
    required StudentLabel? studentLabel,
    @JsonKey(includeIfNull: true, name: 'tuition_level')
    required TuitionLevel? tuitionLevel,
    @JsonKey(includeIfNull: true, name: 'scholarship_cents')
    required NonNegativeCents? scholarshipCents,
    @JsonKey(includeIfNull: true, name: 'cfdi_uuid')
    required CfdiUuid? cfdiUuid,
    @JsonKey(includeIfNull: true) required DeductionExclusion? excluded,
    required List<Warning> warnings,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _PersonalDeduction;

  factory PersonalDeduction.fromJson(Map<String, Object?> json) =>
      _$PersonalDeductionFromJson(json);
}
