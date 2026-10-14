import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Whose breakdown this is: the regime's name, its code and month, its
/// status, and the parameter set and engine version used.
class BreakdownHeader extends StatelessWidget {
  /// The header of [breakdown].
  const new({required this.breakdown, super.key});

  /// The breakdown.
  final Breakdown breakdown;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final muted = context.tokens.colors.muted;
    final regime = breakdown.regime;
    final month = MaterialLocalizations.of(context)
        .formatMonthYear(periodMonth(breakdown.period));
    final parameters = regime.parameterSet;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xxs,
      children: [
        Semantics(
          header: true,
          child: Text(regime.regimeName, style: text.titleLarge),
        ),
        Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xxs,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              l10n.breakdownSubtitle(regime.regimeCode, month),
              style: text.bodyMedium,
            ),
            StatusChip.of(context, regime.status),
          ],
        ),
        Text(
          l10n.breakdownVersions(
            '${parameters.fiscalYear}',
            '${parameters.version}',
            breakdown.engineVersion,
          ),
          style: text.bodySmall?.copyWith(color: muted),
        ),
      ],
    );
  }
}
