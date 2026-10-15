import 'package:control_contable/api/generated/models/treatment.dart';
import 'package:control_contable/core/fiscal/iva_labels.dart';
import 'package:control_contable/core/fiscal/withholding_labels.dart';
import 'package:control_contable/core/forms/app_choice_fields.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/presentation/income_labels.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Longest description the contract accepts (`RecordDescription`).
const int _descriptionMaxLength = 200;

/// Whether "Más detalles" is open: as the user left it, and always while
/// a field inside shows an error (collapsed fields are not built, so the
/// error summary could not reach them).
bool incomeMoreOpen(IncomeFormModel model, {required bool moreOpen}) =>
    moreOpen ||
    <AppFormFieldModel>[
      model.description,
      model.ivaAmount,
      model.otherDate,
      model.cfdiUuid,
      if (model.facts.viaPlatformIsSecondary) model.facts.viaPlatform,
    ].any((field) => field.error != null);

/// The fields of the income page (WH-11): the date and the amount, how it
/// is taxed, then who was billed and only the follow-up questions that can
/// change a withholding, the derived card ([card]), the regime and
/// activity only when there is a choice, then "Más detalles". At most six
/// fields show at once (ui-forms-and-modals.md).
List<Widget> incomeFormFields(
  BuildContext context, {
  required IncomeFormModel model,
  required Widget card,
  required bool moreOpen,
  required VoidCallback onToggleMore,
}) {
  final l10n = AppLocalizations.of(context);
  final facts = model.facts;
  final held = model.regimes
      .where((r) => r.regimeCode == model.regimeCode)
      .first;
  final viaPlatform = AppChoiceField(
    field: facts.viaPlatform,
    options: yesNoOptions(l10n),
  );
  return [
    AppDateField(
      field: model.bookingDate,
      textInputAction: TextInputAction.next,
    ),
    Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        AppMoneyField(
          field: model.amount,
          textInputAction: TextInputAction.done,
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(
          l10n.incomeSummaryTotal(formatPesos(model.amounts.totalCents)),
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      ],
    ),
    if (held.treatment == Treatment.informational)
      Text(l10n.incomeInformationalRegime),
    if (model.asksIvaRate)
      AppChoiceField(
        field: model.ivaRate,
        options: [
          for (final rate in model.ivaRatesOffered)
            (value: rate.json!, label: ivaRateLabel(rate, l10n)),
        ],
      ),
    AppChoiceField(
      field: facts.counterpartyKind,
      options: counterpartyKindOptions(l10n),
    ),
    if (facts.asksViaPlatform && !facts.viaPlatformIsSecondary) viaPlatform,
    if (facts.asksOperation)
      AppChoiceField(
        field: facts.operationKind,
        options: operationKindOptions(l10n),
      ),
    if (facts.asksResidentialRent)
      AppChoiceField(
        field: facts.residentialRent,
        helperText: l10n.withholdingResidentialRentHelp,
        options: yesNoOptions(l10n),
      ),
    card,
    ..._choices(model, l10n),
    FormDisclosure(
      title: l10n.incomeSectionMore,
      isOpen: incomeMoreOpen(model, moreOpen: moreOpen),
      onToggle: onToggleMore,
      children: [
        AppTextField(
          field: model.description,
          keyboardType: TextInputType.text,
          textInputAction: TextInputAction.next,
          autofillHints: const [],
          textCapitalization: TextCapitalization.sentences,
          inputFormatters: [
            LengthLimitingTextInputFormatter(_descriptionMaxLength),
          ],
        ),
        if (facts.asksViaPlatform && facts.viaPlatformIsSecondary) viaPlatform,
        AppMoneyField(
          field: model.ivaAmount,
          textInputAction: TextInputAction.next,
        ),
        AppDateField(
          field: model.otherDate,
          textInputAction: TextInputAction.next,
        ),
        AppTextField(
          field: model.cfdiUuid,
          keyboardType: TextInputType.visiblePassword,
          textInputAction: TextInputAction.done,
          autofillHints: const [],
          autocorrect: false,
          textCapitalization: TextCapitalization.characters,
        ),
      ],
    ),
  ];
}

/// The regime and the 625 activity, each only when there is a choice.
List<Widget> _choices(IncomeFormModel model, AppLocalizations l10n) => [
  if (model.regimes.length > 1)
    AppChoiceField(
      field: model.regime,
      options: [
        for (final regime in model.regimes)
          (value: regime.regimeCode, label: regime.name),
      ],
    ),
  if (model.activities.length > 1)
    AppChoiceField(
      field: model.activity,
      options: [
        for (final activity in model.activities)
          (value: activity.json!, label: activityLabel(activity, l10n)),
      ],
    ),
];
