import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/annual_inputs_update.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// The labels of the annual facts form, already localized.
typedef AnnualFactsFormCopy = ({
  String priorYear,
  String otherExempt,
  String toldEmployer,
});

/// The per-year facts of a persona física (ANN-21), prefilled from
/// [facts]: an empty prior-year income means "unknown" (`null`), never
/// zero (OQ-A5).
class AnnualFactsFormModel {
  /// The form over [facts].
  new({required AnnualInputs facts, required AnnualFactsFormCopy copy}) {
    priorYear = form.moneyField(
      label: copy.priorYear,
      isRequired: false,
      initialText: switch (facts.priorYearAccumulableIncomeCents) {
        final cents? => formatCentavos(cents),
        null => '',
      },
    );
    otherExempt = form.moneyField(
      label: copy.otherExempt,
      initialText: formatCentavos(facts.otherExemptIncomeCents),
    );
    toldEmployer = form.add(
      AppCheckboxController(
        label: copy.toldEmployer,
        isRequired: false,
        initialValue: facts.toldEmployerWillFile,
      ),
    );
  }

  /// The fields and submit state.
  final AppFormController form = AppFormController();

  /// The previous year's acumulables (for the 7 % donation cap).
  late final MoneyFieldController priorYear;

  /// Exempt income outside the records (for the 15 % cap).
  late final MoneyFieldController otherExempt;

  /// Told the employer they will file.
  late final AppCheckboxController toldEmployer;

  /// The server's field paths mapped to this form's fields.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'prior_year_accumulable_income_cents': priorYear,
    'other_exempt_income_cents': otherExempt,
    'told_employer_will_file': toldEmployer,
  };

  /// The PUT body; call only after the form validated.
  AnnualInputsUpdate toUpdate() => AnnualInputsUpdate(
    priorYearAccumulableIncomeCents: priorYear.centavos,
    otherExemptIncomeCents: otherExempt.centavos!,
    toldEmployerWillFile: toldEmployer.checked,
  );

  /// Releases the fields.
  void dispose() => form.dispose();
}
