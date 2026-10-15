import 'package:control_contable/api/generated/models/deduction_beneficiary.dart';
import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/deduction_payment_method.dart';
import 'package:control_contable/api/generated/models/donee_type.dart';
import 'package:control_contable/api/generated/models/personal_deduction_create.dart';
import 'package:control_contable/api/generated/models/tuition_level.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/annual/domain/deduction_rules.dart';
import 'package:flutter/foundation.dart';

/// Longest student label the contract takes.
const int _studentLabelMax = 40;

/// A folio fiscal: 8-4-4-4-12 hexadecimal digits (contract `CfdiUuid`).
final RegExp _cfdiUuid = RegExp(
  r'^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$',
);

/// The labels and messages of the deduction form, already localized.
typedef PersonalDeductionFormCopy = ({
  String category,
  String amount,
  String paidOn,
  String electronic,
  String beneficiary,
  String attest,
  String donee,
  String student,
  String studentInvalid,
  String level,
  String scholarship,
  String cfdi,
  String cfdiInvalid,
});

/// The new personal deduction form (ANN-19): the category first, then only
/// that category's fields (the others are inactive, never validated).
class PersonalDeductionFormModel extends ChangeNotifier {
  /// A form for [fiscalYear] whose dates start at `today` (or 31 December).
  /// [newId] makes the client id of this form session.
  new({
    required this.fiscalYear,
    required this._today,
    required PersonalDeductionFormCopy copy,
    required String Function() newId,
  }) : _id = newId() {
    category = form.add(ChoiceFieldController(label: copy.category));
    amount = form.moneyField(label: copy.amount);
    paidOn = form.dateField(label: copy.paidOn, initialText: _defaultDate);
    paidElectronically = form.add(
      AppCheckboxController(
        label: copy.electronic,
        isRequired: false,
        initialValue: true,
      ),
    );
    beneficiary = form.add(
      ChoiceFieldController(
        label: copy.beneficiary,
        initialValue: DeductionBeneficiary.self.json,
      ),
    );
    attestation = form.checkboxField(label: copy.attest);
    donee = form.add(ChoiceFieldController(label: copy.donee));
    student = form.textField(
      label: copy.student,
      validator: (value) => value.length > _studentLabelMax
          ? CustomFieldError(copy.studentInvalid)
          : null,
    );
    level = form.add(ChoiceFieldController(label: copy.level));
    scholarship = form.moneyField(label: copy.scholarship, initialText: '0.00');
    cfdi = form.textField(
      label: copy.cfdi,
      isRequired: false,
      validator: (value) =>
          _cfdiUuid.hasMatch(value) ? null : CustomFieldError(copy.cfdiInvalid),
    );
    category.addListener(_followCategory);
    beneficiary.addListener(_followBeneficiary);
    _followCategory();
  }

  /// The year the deduction belongs to.
  final int fiscalYear;

  final DateTime _today;
  final String _id;

  /// The fields and submit state.
  final AppFormController form = AppFormController();

  /// What was paid for (asked first).
  late final ChoiceFieldController category;

  /// The amount on the CFDI or constancia.
  late final MoneyFieldController amount;

  /// When it was paid.
  late final DateFieldController paidOn;

  /// Paid by card, transfer or cheque (cash-restricted categories only).
  late final AppCheckboxController paidElectronically;

  /// For whom (categories spent on a dependant).
  late final ChoiceFieldController beneficiary;

  /// The dependant's income is below one UMA (a beneficiary not `self`).
  late final AppCheckboxController attestation;

  /// Who received a donation.
  late final ChoiceFieldController donee;

  /// Which student (tuition; the cap is per student).
  late final AppFieldController student;

  /// The school level (tuition).
  late final ChoiceFieldController level;

  /// Scholarship received (tuition).
  late final MoneyFieldController scholarship;

  /// The optional CFDI folio.
  late final AppFieldController cfdi;

  /// The chosen category, or `null` while none is.
  DeductionCategory? get chosenCategory => switch (category.value) {
    final value? => DeductionCategory.fromJson(value),
    null => null,
  };

  /// Whether the payment was cash in a category where cash is not
  /// deductible (the form warns; the server stores it excluded).
  bool get isCashExcluded =>
      paidElectronically.isActive && !paidElectronically.checked;

  /// Fields that the chosen category asks for, in reading order.
  List<AppFormFieldModel> get activeFields => [
    for (final field in form.fields)
      if (field.isActive) field,
  ];

  /// The server's field paths mapped to this form's fields.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'category': category,
    'amount_cents': amount,
    'paid_on': paidOn,
    'payment_method': paidElectronically,
    'beneficiary': beneficiary,
    'donee_type': donee,
    'student_label': student,
    'tuition_level': level,
    'scholarship_cents': scholarship,
    'cfdi_uuid': cfdi,
  };

  /// The create body; call only after the form validated.
  PersonalDeductionCreate toCreate() {
    final chosen = chosenCategory!;
    final isTuition = chosen == DeductionCategory.tuition;
    final cfdiValue = cfdi.value;
    return PersonalDeductionCreate(
      id: _id,
      category: chosen,
      amountCents: amount.centavos!,
      paidOn: businessDate(paidOn.date!),
      paymentMethod: isCashExcluded
          ? DeductionPaymentMethod.cash
          : DeductionPaymentMethod.electronic,
      beneficiary: beneficiary.isActive
          ? DeductionBeneficiary.fromJson(beneficiary.value!)
          : DeductionBeneficiary.self,
      doneeType: donee.isActive ? DoneeType.fromJson(donee.value!) : null,
      studentLabel: isTuition ? student.value : null,
      tuitionLevel: isTuition ? TuitionLevel.fromJson(level.value!) : null,
      scholarshipCents: isTuition ? scholarship.centavos : null,
      cfdiUuid: cfdiValue.isEmpty ? null : cfdiValue,
    );
  }

  String get _defaultDate =>
      formatDate(defaultPaidOn(chosenCategory, fiscalYear, _today));

  void _followCategory() {
    final chosen = chosenCategory;
    final isTuition = chosen == DeductionCategory.tuition;
    final known = chosen != null;
    for (final field in [amount, paidOn, cfdi]) {
      field.isActive = known;
    }
    paidElectronically.isActive = cashRestrictedCategories.contains(chosen);
    beneficiary.isActive = beneficiaryCategories.contains(chosen);
    donee.isActive = chosen == DeductionCategory.donation;
    for (final field in [student, level, scholarship]) {
      field.isActive = isTuition;
    }
    // A date the user has not touched follows the category's default.
    if (!paidOn.isDirty) {
      paidOn.text.text = _defaultDate;
      paidOn.markPristine();
    }
    _followBeneficiary();
  }

  void _followBeneficiary() {
    attestation.isActive =
        beneficiary.isActive &&
        beneficiary.value != DeductionBeneficiary.self.json;
    notifyListeners();
  }

  @override
  void dispose() {
    form.dispose();
    super.dispose();
  }
}
