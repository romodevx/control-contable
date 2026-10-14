import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/treatment.dart';
import 'package:control_contable/core/fiscal/iva_labels.dart';
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

/// Which collapsed sections are open.
typedef IncomeSections = ({bool withholdings, bool more});

/// The open sections: as the user left them ([withholdingsOpen] `null`
/// until toggled: opened when a withholding is stored or suggested),
/// and always while a field inside shows an error (collapsed fields are
/// not built, so the error summary could not reach them).
IncomeSections incomeSections(
  IncomeFormModel model, {
  required bool? withholdingsOpen,
  required bool moreOpen,
}) {
  final suggested =
      model.startsWithWithholding ||
      model.defaults?.isrWithholdingRate != null ||
      model.defaults?.ivaWithholdingRate != null;
  bool hasError(List<AppFormFieldModel> fields) =>
      fields.any((field) => field.error != null);
  return (
    withholdings:
        (withholdingsOpen ?? suggested) ||
        hasError([model.isrWithheld, model.ivaWithheld]),
    more:
        moreOpen ||
        hasError([model.ivaAmount, model.otherDate, model.cfdiUuid]),
  );
}

/// The fields of the new-income page in the list's column order (date,
/// description, amount, then how it is taxed and who paid), the regime and
/// activity only when there is a choice, then the collapsed sections.
List<Widget> incomeFormFields(
  BuildContext context, {
  required IncomeFormModel model,
  required IncomeSections open,
  required VoidCallback onToggleWithholdings,
  required VoidCallback onToggleMore,
}) {
  final l10n = AppLocalizations.of(context);
  final regimes = model.regimes;
  final activities = model.activities;
  final held = regimes.where((r) => r.regimeCode == model.regimeCode).first;
  return [
    AppDateField(
      field: model.bookingDate,
      textInputAction: TextInputAction.next,
    ),
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
    Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        AppMoneyField(
          field: model.amount,
          textInputAction: TextInputAction.done,
        ),
        const SizedBox(height: AppSpacing.xs),
        _AmountSummary(model: model),
      ],
    ),
    if (held.treatment == Treatment.informational)
      Text(l10n.incomeInformationalRegime),
    AppChoiceField(
      field: model.ivaRate,
      options: [
        for (final rate in model.ivaRatesOffered)
          (value: rate.json!, label: ivaRateLabel(rate, l10n)),
      ],
    ),
    AppChoiceField(
      field: model.counterparty,
      options: [
        for (final payer in CounterpartyType.$valuesDefined)
          (value: payer.json!, label: counterpartyLabel(payer, l10n)),
      ],
    ),
    if (regimes.length > 1)
      AppChoiceField(
        field: model.regime,
        options: [
          for (final regime in regimes)
            (value: regime.regimeCode, label: regime.name),
        ],
      ),
    if (activities.length > 1)
      AppChoiceField(
        field: model.activity,
        options: [
          for (final activity in activities)
            (value: activity.json!, label: activityLabel(activity, l10n)),
        ],
      ),
    if (model.withholdingsApply)
      FormDisclosure(
        title: l10n.incomeSectionWithholdings,
        help: l10n.incomeWithholdingsHelp,
        isOpen: open.withholdings,
        onToggle: onToggleWithholdings,
        children: [
          AppMoneyField(
            field: model.isrWithheld,
            textInputAction: TextInputAction.next,
          ),
          AppMoneyField(
            field: model.ivaWithheld,
            textInputAction: TextInputAction.done,
          ),
        ],
      ),
    FormDisclosure(
      title: l10n.incomeSectionMore,
      isOpen: open.more,
      onToggle: onToggleMore,
      children: [
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

/// The live total under the amount; with withholdings, the breakdown down
/// to the net. Integer centavos only.
class _AmountSummary extends StatelessWidget {
  const new({required this.model});

  final IncomeFormModel model;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final amounts = model.amounts;
    final style = Theme.of(context).textTheme.bodyMedium;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.incomeSummaryTotal(formatPesos(amounts.totalCents)),
          style: style,
        ),
        if (amounts.withheldCents > 0)
          Text(
            l10n.incomeSummaryNet(
              formatPesos(amounts.ivaCents),
              formatPesos(-amounts.withheldCents),
              formatPesos(amounts.netCents),
            ),
            style: style,
          ),
      ],
    );
  }
}
