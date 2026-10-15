import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

/// The bordered surface of the annual input screens (summaries, totals).
class AnnualInputCard extends StatelessWidget {
  /// A card holding [child].
  const new({required this.child, super.key});

  /// The card's content.
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.surface,
        border: Border.all(color: tokens.colors.border),
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: child,
      ),
    );
  }
}

/// A label with its amount, read as one sentence; wraps under 200 % text.
class AnnualAmountRow extends StatelessWidget {
  /// The row of [label] reading [amount].
  const new({required this.label, required this.amount, super.key});

  /// What the amount is.
  final String label;

  /// The formatted amount.
  final String amount;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    return MergeSemantics(
      child: Wrap(
        alignment: WrapAlignment.spaceBetween,
        spacing: AppSpacing.sm,
        children: [
          Text(label, style: text.bodyLarge),
          Text(amount, style: text.titleMedium),
        ],
      ),
    );
  }
}
