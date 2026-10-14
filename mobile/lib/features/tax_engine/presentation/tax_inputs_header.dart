import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Which inputs the form edits (regime and fiscal year) and why they
/// matter: what is left empty is computed with zero, so the estimate
/// stays incomplete.
class TaxInputsHeader extends StatelessWidget {
  /// The header of [inputs].
  const new({required this.inputs, super.key});

  /// The regime and year edited.
  final TaxInputs inputs;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xs,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.taxInputsSubtitle(inputs.regimeCode, '${inputs.fiscalYear}'),
            style: text.titleMedium,
          ),
        ),
        Text(l10n.taxInputsIntro, style: text.bodyMedium),
      ],
    );
  }
}
