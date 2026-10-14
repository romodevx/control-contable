import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:flutter/material.dart';

/// Share of the hero size the centavos are drawn at (DIRECTIONS.md: "cents
/// are shown smaller on hero amounts").
const double _heroCentsScale = 0.55;

/// A labeled amount: [label] on the left, [centavos] right-aligned in
/// tabular figures. Read as one sentence ("IVA a pagar: 2,133 pesos con
/// 33 centavos"); the visual `$2,133.33` is not read.
class MoneyRow extends StatelessWidget {
  /// The row of [label] and [centavos].
  const new({
    required this.label,
    required this.centavos,
    this.color,
    this.labelStyle,
    this.amountStyle,
    super.key,
  });

  /// What the amount is.
  final String label;

  /// The amount, in centavos (formatted, never rounded).
  final int centavos;

  /// Text color of both sides (defaults to the theme's).
  final Color? color;

  /// Style of [label] (defaults to body).
  final TextStyle? labelStyle;

  /// Style of the amount (defaults to body, semi-bold).
  final TextStyle? amountStyle;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    return Semantics(
      label: '$label: ${spokenMoney(centavos)}',
      excludeSemantics: true,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xxs),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Expanded(
              child: Text(
                label,
                style: (labelStyle ?? text.bodyLarge)?.copyWith(color: color),
              ),
            ),
            const SizedBox(width: AppSpacing.sm),
            Text(
              formatMoney(centavos),
              style: (amountStyle ?? text.bodyLarge)?.copyWith(
                color: color,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// A hero amount in the display face (Source Serif 4) with smaller
/// centavos, read as [label] plus the spoken amount.
class HeroAmount extends StatelessWidget {
  /// The hero of [label] and [centavos] in [color].
  const new({
    required this.label,
    required this.centavos,
    required this.color,
    super.key,
  });

  /// What the amount is.
  final String label;

  /// The amount, in centavos.
  final int centavos;

  /// Text color on the hero surface.
  final Color color;

  @override
  Widget build(BuildContext context) {
    final style = Theme.of(context).textTheme.displayLarge
        ?.copyWith(color: color);
    final money = formatMoney(centavos);
    final cut = money.lastIndexOf('.');
    return Semantics(
      label: '$label: ${spokenMoney(centavos)}',
      excludeSemantics: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: Theme.of(context).textTheme.bodyLarge
                ?.copyWith(color: color),
          ),
          Text.rich(
            TextSpan(
              text: money.substring(0, cut),
              children: [
                TextSpan(
                  text: money.substring(cut),
                  style: TextStyle(
                    fontSize: (style?.fontSize ?? 0) * _heroCentsScale,
                  ),
                ),
              ],
            ),
            style: style,
          ),
        ],
      ),
    );
  }
}
