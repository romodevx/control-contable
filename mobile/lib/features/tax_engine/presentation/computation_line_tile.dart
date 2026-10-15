import 'package:control_contable/api/generated/models/citation.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/rate_percent.dart';
import 'package:control_contable/features/tax_engine/presentation/citation_disclosure.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

const double _cpaIconSize = 18;

/// How a computation step reads.
enum ComputationLineRole {
  /// A figure that adds, multiplies or is looked up.
  step,

  /// A figure taken away: shown and read with a minus.
  subtraction,

  /// A result, ruled off from the steps above it.
  result,

  /// The rounding adjustment: muted and explained, never cited.
  rounding,

  /// A remark with no figure of its own ("Tope aplicado: …"): its words
  /// carry the amounts.
  note,
}

/// One step of a server computation trace, as the tile shows it. Monthly
/// (M5-26) and annual (ANN-18) breakdowns map their lines to it.
@immutable
class ComputationLine {
  /// A step reading [concept] with [amountCents].
  const new({
    required this.concept,
    required this.amountCents,
    required this.role,
    required this.legalBasis,
    this.rate,
    this.sourceLabel,
    this.regimeLabel,
    this.needsCpaReview = false,
  });

  /// What the step is (server text).
  final String concept;

  /// The figure, in centavos, as the server sends it.
  final int amountCents;

  /// How the step reads.
  final ComputationLineRole role;

  /// The rate of a multiplication (`DecimalRate`), else `null`.
  final String? rate;

  /// The step's legal basis, in the server's order.
  final List<Citation> legalBasis;

  /// Where the figure comes from, in words; `null` when unknown.
  final String? sourceLabel;

  /// The regime the step belongs to, in words; `null` when section-wide.
  final String? regimeLabel;

  /// Whether the step uses a rule no CPA has validated yet.
  final bool needsCpaReview;
}

/// One step of the computation: the concept and its amount (a subtraction
/// shows and reads as one), the rate of a multiplication, where the figure
/// comes from, the ⚠️ of a CPA-unverified rule, and the legal basis.
class ComputationLineTile extends StatelessWidget {
  /// The tile of [line]; [dependsOnMissingInputs] marks its amount.
  const new({
    required this.line,
    required this.dependsOnMissingInputs,
    super.key,
  });

  /// The step.
  final ComputationLine line;

  /// Whether the figure depends on inputs the server reported missing.
  final bool dependsOnMissingInputs;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final text = Theme.of(context).textTheme;
    final muted = colors.muted;
    final role = line.role;
    final rate = line.rate;
    final chips = [?line.regimeLabel, ?line.sourceLabel];
    return Semantics(
      container: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: role == ComputationLineRole.result
              ? Border(top: BorderSide(color: colors.border))
              : null,
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: AppSpacing.xxs,
            children: [
              if (role == ComputationLineRole.note)
                _Note(text: line.concept)
              else
                MoneyRow(
                  label: line.concept,
                  centavos: role == ComputationLineRole.subtraction
                      ? -line.amountCents
                      : line.amountCents,
                  color: role == ComputationLineRole.rounding ? muted : null,
                  labelStyle: role == ComputationLineRole.result
                      ? text.titleSmall
                      : text.bodyLarge,
                  isIncomplete: dependsOnMissingInputs,
                ),
              if (rate != null)
                Text(
                  l10n.breakdownRate(formatRatePercent(rate)),
                  style: text.bodyMedium?.copyWith(color: muted),
                ),
              if (role == ComputationLineRole.rounding)
                Text(
                  l10n.breakdownRoundingHelp,
                  style: text.bodyMedium?.copyWith(color: muted),
                ),
              if (chips.isNotEmpty || line.needsCpaReview)
                Wrap(
                  spacing: AppSpacing.xs,
                  runSpacing: AppSpacing.xxs,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    for (final chip in chips)
                      StatusChip(label: chip, color: muted),
                    if (line.needsCpaReview) const CpaReviewMark(),
                  ],
                ),
              if (line.legalBasis.isNotEmpty)
                CitationDisclosure(citations: line.legalBasis),
            ],
          ),
        ),
      ),
    );
  }
}

/// A remark line: an info icon pulled to the text, then the words.
class _Note extends StatelessWidget {
  const new({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) => Row(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Icon(
        Icons.info_outline,
        size: _cpaIconSize,
        color: context.tokens.colors.muted,
      ),
      const SizedBox(width: AppSpacing.xs),
      Expanded(child: Text(text, style: Theme.of(context).textTheme.bodyLarge)),
    ],
  );
}

/// ⚠️ "Sin validar por un contador": the icon never stands alone.
class CpaReviewMark extends StatelessWidget {
  /// Creates the mark.
  const new({super.key});

  @override
  Widget build(BuildContext context) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Icon(
        Icons.warning_amber_rounded,
        size: _cpaIconSize,
        color: context.tokens.colors.legendText,
      ),
      const SizedBox(width: AppSpacing.xxs),
      Flexible(
        child: Text(
          AppLocalizations.of(context).breakdownCpaLine,
          style: Theme.of(context).textTheme.labelLarge,
        ),
      ),
    ],
  );
}
