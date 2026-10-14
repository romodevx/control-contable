import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/features/tax_engine/application/tax_input_entries.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs.dart';
import 'package:control_contable/features/tax_engine/presentation/monthly_payments_section.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Bounds of the coeficiente de utilidad and every other `rate` input.
const String rateMin = '0';

/// See [rateMin].
const String rateMax = '1';

/// The widget that edits [entry]: an amount, a rate with its bounds, a
/// yes/no box or the monthly grid; nothing for a kind this app version
/// does not edit (its values are kept as stored).
Widget? taxInputField(
  BuildContext context,
  TaxInputEntry entry, {
  required TaxInputsQuery query,
  required int? openGroup,
  required ValueChanged<int> onToggleGroup,
}) {
  final l10n = AppLocalizations.of(context);
  final needed = entry.def.requiredValue ? l10n.taxInputsNeeded : null;
  return switch (entry) {
    MoneyEntry(:final field) => AppTextField(
      field: field,
      keyboardType: decimalKeyboard,
      textInputAction: TextInputAction.next,
      autofillHints: const [],
      autocorrect: false,
      prefixText: r'$ ',
      helperText: needed,
    ),
    RateEntry(:final field) => AppTextField(
      field: field,
      keyboardType: decimalKeyboard,
      textInputAction: TextInputAction.next,
      autofillHints: const [],
      autocorrect: false,
      helperText: [
        l10n.onboardingRateHelp(rateMin, rateMax),
        ?needed,
      ].join(' '),
    ),
    FlagEntry(:final field) => AppCheckboxField(field: field),
    MonthlyEntry() => MonthlyPaymentsSection(
      entry: entry,
      query: query,
      openGroup: openGroup,
      onToggleGroup: onToggleGroup,
    ),
    KeptEntry() => null,
  };
}
