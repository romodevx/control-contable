import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/features/annual/application/salary_statement_form_model.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The constancia fields in the order of its list: the employer's alias,
/// the three totals, then the special situations under a disclosure.
List<Widget> salaryStatementFields(
  BuildContext context, {
  required SalaryStatementFormModel model,
  required bool moreOpen,
  required VoidCallback onToggleMore,
}) {
  final l10n = AppLocalizations.of(context);
  return [
    Text(l10n.annualSalaryHelp, style: Theme.of(context).textTheme.bodyMedium),
    AppTextField(
      field: model.employer,
      keyboardType: TextInputType.text,
      textInputAction: TextInputAction.next,
      autofillHints: const [AutofillHints.organizationName],
      textCapitalization: TextCapitalization.words,
      helperText: l10n.annualSalaryEmployerHelp,
    ),
    AppMoneyField(
      field: model.accumulable,
      textInputAction: TextInputAction.next,
    ),
    AppMoneyField(field: model.exempt, textInputAction: TextInputAction.next),
    AppMoneyField(field: model.isr, textInputAction: TextInputAction.done),
    FormDisclosure(
      title: l10n.annualSalaryMore,
      isOpen: moreOpen,
      onToggle: onToggleMore,
      children: [
        AppCheckboxField(field: model.simultaneous),
        AppCheckboxField(field: model.leftBeforeYearEnd),
        AppCheckboxField(field: model.payerNotWithholding),
      ],
    ),
  ];
}

/// The constancia form's copy.
SalaryStatementFormCopy salaryStatementCopy(AppLocalizations l10n) => (
  employer: l10n.annualSalaryEmployer,
  employerInvalid: l10n.annualWriteInvalidText,
  accumulable: l10n.annualSalaryAccumulable,
  exempt: l10n.annualSalaryExempt,
  isr: l10n.annualSalaryIsrWithheld,
  simultaneous: l10n.annualSalarySimultaneous,
  left: l10n.annualSalaryLeft,
  payerNotWithholding: l10n.annualSalaryPayerNotWithholding,
);

/// Confirms a saved [statement] on [messenger]; with warnings (605 not
/// held) the toast stays until closed so they are not missed.
void showSalarySavedToast(
  ScaffoldMessengerState messenger,
  AppLocalizations l10n,
  SalaryStatement statement, {
  required bool updated,
}) {
  final heading = updated ? l10n.annualSalaryUpdated : l10n.annualSalarySaved;
  final notes = [for (final w in statement.warnings) _warningText(w, l10n)];
  messenger
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text([heading, ...notes].join(' ')),
        persist: notes.isNotEmpty,
        showCloseIcon: notes.isNotEmpty,
      ),
    );
}

String _warningText(Warning warning, AppLocalizations l10n) =>
    warning.code == WarningCode.salaryRegimeNotHeld
    ? l10n.annualSalaryRegimeNotHeld
    : warning.message;
