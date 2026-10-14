import 'package:control_contable/api/generated/models/breakdown_line_tax.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/breakdown_groups.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_line_tile.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// One tax of a breakdown under its heading: the regime's ISR, the month's
/// single IVA (with why it is one payment for every regime, OQ-T2), or the
/// withholdings to remit (paid apart from the taxes).
class BreakdownGroupSection extends StatelessWidget {
  /// The section of [group].
  const new({
    required this.group,
    required this.isIncomplete,
    this.ivaNettingUnverified = false,
    super.key,
  });

  /// The tax and its lines, in `seq` order.
  final BreakdownGroup group;

  /// Whether the regime's figures depend on missing inputs.
  final bool isIncomplete;

  /// Whether the IVA was netted across regimes by a rule no CPA has
  /// validated yet (OQ-T2, OQ-T8).
  final bool ivaNettingUnverified;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final muted = context.tokens.colors.muted;
    final notes = switch (group.tax) {
      BreakdownLineTax.iva => [
        l10n.breakdownIvaHelp,
        if (ivaNettingUnverified) l10n.breakdownIvaNettingUnverified,
      ],
      BreakdownLineTax.withholding => [l10n.breakdownWithholdingsHelp],
      BreakdownLineTax.isr || BreakdownLineTax.$unknown => const <String>[],
    };
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      spacing: AppSpacing.xxs,
      children: [
        Semantics(
          header: true,
          child: Text(_title(l10n), style: text.titleMedium),
        ),
        for (final note in notes)
          Text(note, style: text.bodyMedium?.copyWith(color: muted)),
        for (final line in group.lines)
          BreakdownLineTile(
            key: ValueKey('line-${line.seq}'),
            line: line,
            dependsOnMissingInputs: isIncomplete && !isRecordsLine(line),
          ),
      ],
    );
  }

  String _title(AppLocalizations l10n) => switch (group.tax) {
    BreakdownLineTax.iva => l10n.breakdownGroupIva,
    BreakdownLineTax.withholding => l10n.estimateWithholdingsToRemit,
    BreakdownLineTax.isr || BreakdownLineTax.$unknown => l10n.breakdownGroupIsr,
  };
}
