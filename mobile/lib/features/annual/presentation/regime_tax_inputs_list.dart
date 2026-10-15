import 'dart:async';

import 'package:control_contable/api/generated/models/annual_regime_tax_input.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// The regimes' annual tax inputs (losses, PTU paid, depreciation), shown
/// read-only: each is owned by the regime's tax-inputs form (one source per
/// fact, ADR-0009), which its button opens at that key.
class RegimeTaxInputsList extends StatelessWidget {
  /// The [inputs] of [query].
  const new({required this.query, required this.inputs, super.key});

  /// Whose inputs and which year.
  final AnnualQuery query;

  /// As the server ordered them (regime, then catalog order).
  final List<AnnualRegimeTaxInput> inputs;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(l10n.annualFactsRegimeTitle, style: text.titleMedium),
        ),
        const SizedBox(height: AppSpacing.xs),
        Text(l10n.annualFactsRegimeHelp, style: text.bodyMedium),
        for (final (index, input) in inputs.indexed) ...[
          if (index == 0 || inputs[index - 1].regimeCode != input.regimeCode)
            Padding(
              padding: const EdgeInsets.only(top: AppSpacing.sm),
              child: Text(
                l10n.estimateRegimeCode(input.regimeCode),
                style: text.titleSmall,
              ),
            ),
          ListTile(
            contentPadding: EdgeInsets.zero,
            title: Text(input.label),
            subtitle: Text(_value(input, l10n)),
            trailing: IconButton(
              tooltip: l10n.annualFactsRegimeEdit(
                input.label,
                input.regimeCode,
              ),
              icon: const Icon(Icons.edit_outlined),
              onPressed: () => unawaited(
                context.push<void>(
                  TaxEngineRoutes.taxInputs(
                    query.taxpayerId,
                    input.regimeCode,
                    fiscalYear: query.fiscalYear,
                    field: input.key,
                  ),
                ),
              ),
            ),
          ),
        ],
      ],
    );
  }

  static String _value(AnnualRegimeTaxInput input, AppLocalizations l10n) =>
      switch (input.valueCents) {
        final cents? => formatPesos(cents),
        null => l10n.annualFactsRegimeEmpty,
      };
}
