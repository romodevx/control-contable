import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Share of the hero size the centavos are drawn at (DIRECTIONS.md: "cents
/// are shown smaller on hero amounts").
const double _heroCentsScale = 0.55;

/// The widest share of a row the amount may take beside its label; a wider
/// amount (large text, large figure) moves under the label instead of
/// squeezing it or overflowing.
const double _maxAmountShareBeside = 0.5;

/// [label] as shown and read: marked "(incompleto)" when the amount
/// depends on missing inputs, so no figure ever looks final.
String _labelOf(BuildContext context, String label, {required bool marked}) =>
    marked
    ? AppLocalizations.of(context).estimateIncompleteAmount(label)
    : label;

/// A labeled amount: [label] on the left, [centavos] right-aligned in
/// tabular figures, or under the label when it does not fit beside it.
/// Read as one sentence ("IVA a pagar: 2,133 pesos con 33 centavos"); the
/// visual `$2,133.33` is not read.
class MoneyRow extends StatelessWidget {
  /// The row of [label] and [centavos].
  const new({
    required this.label,
    required this.centavos,
    this.color,
    this.labelStyle,
    this.amountStyle,
    this.isIncomplete = false,
    super.key,
  });

  /// What the amount is.
  final String label;

  /// The amount, in centavos (formatted, never rounded).
  final int centavos;

  /// Whether the amount depends on missing inputs (marked in its label).
  final bool isIncomplete;

  /// Text color of both sides (defaults to the theme's).
  final Color? color;

  /// Style of [label] (defaults to body).
  final TextStyle? labelStyle;

  /// Style of the amount (defaults to body, semi-bold).
  final TextStyle? amountStyle;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final shownLabel = _labelOf(context, label, marked: isIncomplete);
    final labelText = Text(
      shownLabel,
      style: (labelStyle ?? text.bodyLarge)?.copyWith(color: color),
    );
    final amount = formatMoney(centavos);
    final shownAmountStyle = (amountStyle ?? text.bodyLarge)?.copyWith(
      color: color,
      fontWeight: FontWeight.w600,
    );
    return Semantics(
      container: true,
      label: '$shownLabel: ${spokenMoney(centavos)}',
      excludeSemantics: true,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xxs),
        child: LayoutBuilder(
          builder: (context, constraints) {
            final amountText = Text(amount, style: shownAmountStyle);
            if (_fitsBeside(context, amount, shownAmountStyle, constraints)) {
              return Row(
                crossAxisAlignment: CrossAxisAlignment.baseline,
                textBaseline: TextBaseline.alphabetic,
                children: [
                  Expanded(child: labelText),
                  const SizedBox(width: AppSpacing.sm),
                  amountText,
                ],
              );
            }
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                labelText,
                Align(
                  alignment: AlignmentDirectional.centerEnd,
                  child: FittedBox(fit: BoxFit.scaleDown, child: amountText),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  /// Whether [amount] in [style], at the user's text scale, takes at most
  /// [_maxAmountShareBeside] of the row (one text layout: O(length)).
  static bool _fitsBeside(
    BuildContext context,
    String amount,
    TextStyle? style,
    BoxConstraints constraints,
  ) {
    final painter = TextPainter(
      text: TextSpan(text: amount, style: style),
      textDirection: Directionality.of(context),
      textScaler: MediaQuery.textScalerOf(context),
      maxLines: 1,
    )..layout();
    final width = painter.width + AppSpacing.sm;
    painter.dispose();
    return width <= constraints.maxWidth * _maxAmountShareBeside;
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
    this.isIncomplete = false,
    super.key,
  });

  /// What the amount is.
  final String label;

  /// The amount, in centavos.
  final int centavos;

  /// Text color on the hero surface.
  final Color color;

  /// Whether the amount depends on missing inputs (marked in its label).
  final bool isIncomplete;

  @override
  Widget build(BuildContext context) {
    final style = Theme.of(context).textTheme.displayLarge
        ?.copyWith(color: color);
    final money = formatMoney(centavos);
    final cut = money.lastIndexOf('.');
    final shownLabel = _labelOf(context, label, marked: isIncomplete);
    return Semantics(
      container: true,
      label: '$shownLabel: ${spokenMoney(centavos)}',
      excludeSemantics: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            shownLabel,
            style: Theme.of(context).textTheme.bodyLarge
                ?.copyWith(color: color),
          ),
          // Shrunk rather than broken across lines: a split figure misreads.
          FittedBox(
            fit: BoxFit.scaleDown,
            alignment: AlignmentDirectional.centerStart,
            child: Text.rich(
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
          ),
        ],
      ),
    );
  }
}
