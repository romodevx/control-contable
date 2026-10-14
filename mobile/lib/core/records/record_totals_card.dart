import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

/// One line of a [RecordTotalsCard]: a label and an amount in centavos
/// (negative for what is subtracted, shown with a real minus).
typedef RecordTotalLine = ({String label, int cents});

/// A record list's month totals: the headline amount, how many records,
/// and the lines that make it up. Amounts use tabular figures.
class RecordTotalsCard extends StatelessWidget {
  /// The card titled [title] with [totalCents] and its [lines].
  const new({
    required this.title,
    required this.totalCents,
    required this.countLabel,
    required this.lines,
    super.key,
  });

  /// What the headline amount is.
  final String title;

  /// The headline amount.
  final int totalCents;

  /// How many records the totals cover ("3 ingresos").
  final String countLabel;

  /// The breakdown, in reading order.
  final List<RecordTotalLine> lines;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final theme = Theme.of(context);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.surface,
        border: Border.all(color: tokens.colors.border),
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(title, style: theme.textTheme.bodySmall),
            const SizedBox(height: AppSpacing.xxs),
            Text(formatPesos(totalCents), style: theme.textTheme.titleLarge),
            Text(countLabel, style: theme.textTheme.bodySmall),
            const SizedBox(height: AppSpacing.xs),
            for (final line in lines)
              Padding(
                padding: const EdgeInsets.only(top: AppSpacing.xxs),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        line.label,
                        style: theme.textTheme.bodyMedium,
                      ),
                    ),
                    Text(
                      formatPesos(line.cents),
                      style: theme.textTheme.bodyMedium,
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }
}
