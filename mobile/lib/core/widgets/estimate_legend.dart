import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The mandatory legend ("Esto es un aproximado. No todos los gastos
/// aplican para tu giro comercial.") in its tinted block
/// (docs/design/DIRECTIONS.md): wherever the app speaks of an estimate.
/// Never hidden, never dismissible.
class EstimateLegend extends StatelessWidget {
  /// Creates the legend.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final colors = tokens.colors;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.legendBg,
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Text(
          AppLocalizations.of(context).estimateLegend,
          style: Theme.of(context).textTheme.bodyMedium
              ?.copyWith(color: colors.legendText),
        ),
      ),
    );
  }
}
