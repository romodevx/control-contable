import 'package:control_contable/api/generated/models/deduction_beneficiary.dart';
import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/donee_type.dart';
import 'package:control_contable/api/generated/models/tuition_level.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_choice_fields.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_disclosure.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/annual/application/personal_deduction_form_model.dart';
import 'package:control_contable/features/annual/domain/deduction_rules.dart';
import 'package:control_contable/features/annual/presentation/deduction_labels.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The fields of [model] that its category asks for, in reading order:
/// category, amount, date, how it was paid, for whom, the category's own
/// fields, then the optional CFDI folio under "Más datos".
List<Widget> personalDeductionFields(
  BuildContext context, {
  required PersonalDeductionFormModel model,
  required bool moreOpen,
  required VoidCallback onToggleMore,
}) {
  final l10n = AppLocalizations.of(context);
  final chosen = model.chosenCategory;
  return [
    AppChoiceField(
      field: model.category,
      options: [
        for (final category in DeductionCategory.$valuesDefined)
          (
            value: category.json!,
            label: deductionCategoryLabel(category, l10n),
          ),
      ],
    ),
    if (chosen != null) ...[
      AppMoneyField(
        field: model.amount,
        textInputAction: TextInputAction.next,
        helperText: chosen == DeductionCategory.tuition
            ? l10n.annualDeductionTuitionAmountHelp
            : l10n.annualDeductionAmountHelp,
      ),
      AppDateField(
        field: model.paidOn,
        textInputAction: TextInputAction.next,
        helperText: yearEndCategories.contains(chosen)
            ? l10n.annualDeductionYearEndHelp
            : null,
      ),
      if (model.paidElectronically.isActive)
        AppCheckboxField(
          field: model.paidElectronically,
          below: model.isCashExcluded ? const _CashWarning() : null,
        ),
      if (model.beneficiary.isActive)
        AppChoiceField(
          field: model.beneficiary,
          options: [
            for (final who in DeductionBeneficiary.$valuesDefined)
              (value: who.json!, label: beneficiaryLabel(who, l10n)),
          ],
        ),
      if (model.attestation.isActive)
        AppCheckboxField(field: model.attestation),
      if (model.donee.isActive)
        AppChoiceField(
          field: model.donee,
          options: [
            for (final donee in DoneeType.$valuesDefined)
              (value: donee.json!, label: doneeLabel(donee, l10n)),
          ],
        ),
      if (chosen == DeductionCategory.tuition) ..._tuitionFields(model, l10n),
      FormDisclosure(
        title: l10n.annualDeductionMore,
        // Kept open while the folio shows an error, so it is never hidden.
        isOpen: moreOpen || model.cfdi.error != null,
        onToggle: onToggleMore,
        children: [
          AppTextField(
            field: model.cfdi,
            keyboardType: TextInputType.text,
            textInputAction: TextInputAction.done,
            autofillHints: const [],
            autocorrect: false,
          ),
        ],
      ),
    ],
  ];
}

List<Widget> _tuitionFields(
  PersonalDeductionFormModel model,
  AppLocalizations l10n,
) => [
  AppTextField(
    field: model.student,
    keyboardType: TextInputType.name,
    textInputAction: TextInputAction.next,
    autofillHints: const [],
    textCapitalization: TextCapitalization.words,
    helperText: l10n.annualDeductionStudentHelp,
  ),
  AppChoiceField(
    field: model.level,
    options: [
      for (final level in TuitionLevel.$valuesDefined)
        (value: level.json!, label: tuitionLevelLabel(level, l10n)),
    ],
  ),
  AppMoneyField(
    field: model.scholarship,
    textInputAction: TextInputAction.next,
  ),
];

/// The cash warning under the payment box: an icon and words, never
/// colour alone.
class _CashWarning extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(top: AppSpacing.xs),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ExcludeSemantics(
            child: Icon(
              Icons.warning_amber_rounded,
              color: theme.colorScheme.error,
            ),
          ),
          const SizedBox(width: AppSpacing.xs),
          Expanded(
            child: Text(
              AppLocalizations.of(context).annualDeductionCashWarning,
              style: theme.textTheme.bodyMedium,
            ),
          ),
        ],
      ),
    );
  }
}
