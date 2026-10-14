import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The mandatory legend ("Esto es un aproximado. No todos los gastos
/// aplican para tu giro comercial.") in its tinted block
/// (docs/design/DIRECTIONS.md): wherever the app speaks of an estimate.
/// Never hidden, never dismissible.
///
/// Next to a server estimate it reads the response's `legend` ([text]),
/// never a client copy (MODULE-tax-engine §2.3); the app's own copy is the
/// fallback where no estimate is loaded.
class EstimateLegend extends StatelessWidget {
  /// Creates the legend reading [text], or the app's copy.
  const new({this.text, super.key});

  /// The server's legend.
  final String? text;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final colors = tokens.colors;
    final legend = DecoratedBox(
      decoration: BoxDecoration(
        color: colors.legendBg,
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Text(
          text ?? AppLocalizations.of(context).estimateLegend,
          style: Theme.of(context).textTheme.bodyMedium
              ?.copyWith(color: colors.legendText),
        ),
      ),
    );
    // Its own screen-reader stop: never merged into the figures around it.
    return Semantics(container: true, child: legend);
  }
}
