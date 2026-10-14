import 'package:control_contable/api/generated/models/disclaimer.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Leading rule of a warning block.
const double _warningRule = 3;

/// Leading rule of the strong (CPA-unverified) banner: heavier than a
/// warning's, so the two never read alike.
const double _strongRule = 6;

const double _iconSize = 22;

/// The strong warning (severity `strong`, `cpa_review_pending`): a
/// non-dismissible banner above the figures, titled, with the server text
/// naming the unverified rules (MODULE-tax-engine §2.9, OQ-T8).
class CpaReviewBanner extends StatelessWidget {
  /// The banner reading [disclaimer]'s text.
  const new({required this.disclaimer, super.key});

  /// The `cpa_review_pending` disclaimer.
  final Disclaimer disclaimer;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    final text = Theme.of(context).textTheme;
    return _RuledBlock(
      rule: _strongRule,
      ruleColor: colors.error,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            Icons.warning_amber_rounded,
            size: _iconSize,
            color: colors.legendText,
          ),
          const SizedBox(width: AppSpacing.xs),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Semantics(
                  header: true,
                  child: Text(
                    AppLocalizations.of(context).estimateCpaBannerTitle,
                    style: text.titleMedium?.copyWith(color: colors.legendText),
                  ),
                ),
                const SizedBox(height: AppSpacing.xxs),
                Text(
                  disclaimer.text,
                  style: text.bodyMedium?.copyWith(color: colors.legendText),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// A `warning` disclaimer (missing inputs, provisional parameters, mixed
/// activity): a tinted block with a thin rule, next to the figures.
class DisclaimerWarning extends StatelessWidget {
  /// The block reading [disclaimer].
  const new({required this.disclaimer, super.key});

  /// The disclaimer.
  final Disclaimer disclaimer;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    return _RuledBlock(
      rule: _warningRule,
      ruleColor: colors.legendText,
      child: Text(
        disclaimer.text,
        style: Theme.of(context).textTheme.bodyMedium
            ?.copyWith(color: colors.legendText),
      ),
    );
  }
}

/// An `info` disclaimer (not tax advice, deductibility…): plain muted text.
class DisclaimerNote extends StatelessWidget {
  /// The note reading [disclaimer].
  const new({required this.disclaimer, super.key});

  /// The disclaimer.
  final Disclaimer disclaimer;

  @override
  Widget build(BuildContext context) => Text(
    disclaimer.text,
    style: Theme.of(context).textTheme.bodyMedium
        ?.copyWith(color: context.tokens.colors.muted),
  );
}

class _RuledBlock extends StatelessWidget {
  const new({required this.rule, required this.ruleColor, required this.child});

  final double rule;
  final Color ruleColor;
  final Widget child;

  @override
  Widget build(BuildContext context) => Semantics(
    container: true,
    child: DecoratedBox(
      decoration: BoxDecoration(
        color: context.tokens.colors.legendBg,
        border: Border(
          left: BorderSide(color: ruleColor, width: rule),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: child,
      ),
    ),
  );
}
