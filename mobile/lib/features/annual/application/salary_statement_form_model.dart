import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_create.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// Longest employer alias the contract takes.
const int _employerLabelMax = 60;

/// The labels and messages of the constancia form, already localized.
typedef SalaryStatementFormCopy = ({
  String employer,
  String employerInvalid,
  String accumulable,
  String exempt,
  String isr,
  String simultaneous,
  String left,
  String payerNotWithholding,
});

/// One employer's constancia de sueldos (ANN-20): its optional alias (no
/// employer RFC in v1), its three totals and the special situations. New,
/// or an edit of [editing] that sends only what changed.
class SalaryStatementFormModel {
  /// A form prefilled from [editing] when given.
  new({required SalaryStatementFormCopy copy, this.editing}) {
    final stored = editing;
    employer = form.textField(
      label: copy.employer,
      isRequired: false,
      initialText: stored?.employerLabel ?? '',
      validator: (value) => value.length > _employerLabelMax
          ? CustomFieldError(copy.employerInvalid)
          : null,
    );
    accumulable = form.moneyField(
      label: copy.accumulable,
      initialText: _money(stored?.accumulableCents),
    );
    exempt = form.moneyField(
      label: copy.exempt,
      initialText: _money(stored?.exemptCents),
    );
    isr = form.moneyField(
      label: copy.isr,
      initialText: _money(stored?.isrWithheldCents),
    );
    simultaneous = _box(copy.simultaneous, stored?.simultaneous);
    leftBeforeYearEnd = _box(copy.left, stored?.leftBeforeYearEnd);
    payerNotWithholding = _box(
      copy.payerNotWithholding,
      stored?.payerNotWithholding,
    );
  }

  /// The constancia being edited; `null` for a new one.
  final SalaryStatement? editing;

  /// The fields and submit state.
  final AppFormController form = AppFormController();

  /// The employer's alias (optional).
  late final AppFieldController employer;

  /// Gravados (acumulables).
  late final MoneyFieldController accumulable;

  /// Exentos.
  late final MoneyFieldController exempt;

  /// ISR withheld.
  late final MoneyFieldController isr;

  /// Two or more employers at the same time.
  late final AppCheckboxController simultaneous;

  /// Left before 31 December.
  late final AppCheckboxController leftBeforeYearEnd;

  /// The employer was not obliged to withhold.
  late final AppCheckboxController payerNotWithholding;

  /// Whether a special situation is marked (its disclosure starts open).
  bool get hasSpecialSituation =>
      simultaneous.checked ||
      leftBeforeYearEnd.checked ||
      payerNotWithholding.checked;

  /// The server's field paths mapped to this form's fields.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'employer_label': employer,
    'accumulable_cents': accumulable,
    'exempt_cents': exempt,
    'isr_withheld_cents': isr,
    'simultaneous': simultaneous,
    'left_before_year_end': leftBeforeYearEnd,
    'payer_not_withholding': payerNotWithholding,
  };

  String? get _label => employer.value.isEmpty ? null : employer.value;

  /// The create body for [fiscalYear] with the client [id].
  SalaryStatementCreate toCreate({
    required int fiscalYear,
    required String id,
  }) => SalaryStatementCreate(
    id: id,
    fiscalYear: fiscalYear,
    employerLabel: _label,
    accumulableCents: accumulable.centavos!,
    exemptCents: exempt.centavos!,
    isrWithheldCents: isr.centavos!,
    simultaneous: simultaneous.checked,
    leftBeforeYearEnd: leftBeforeYearEnd.checked,
    payerNotWithholding: payerNotWithholding.checked,
  );

  /// The merge-patch of what differs from [stored] (`null` clears the
  /// alias); empty when nothing changed.
  Map<String, Object?> toPatch(SalaryStatement stored) => {
    if (_label != stored.employerLabel) 'employer_label': _label,
    if (accumulable.centavos != stored.accumulableCents)
      'accumulable_cents': accumulable.centavos,
    if (exempt.centavos != stored.exemptCents) 'exempt_cents': exempt.centavos,
    if (isr.centavos != stored.isrWithheldCents)
      'isr_withheld_cents': isr.centavos,
    if (simultaneous.checked != stored.simultaneous)
      'simultaneous': simultaneous.checked,
    if (leftBeforeYearEnd.checked != stored.leftBeforeYearEnd)
      'left_before_year_end': leftBeforeYearEnd.checked,
    if (payerNotWithholding.checked != stored.payerNotWithholding)
      'payer_not_withholding': payerNotWithholding.checked,
  };

  AppCheckboxController _box(String label, bool? initial) => form.add(
    AppCheckboxController(
      label: label,
      isRequired: false,
      initialValue: initial ?? false,
    ),
  );

  static String _money(int? cents) =>
      cents == null ? '' : formatCentavos(cents);

  /// Releases the fields.
  void dispose() => form.dispose();
}
