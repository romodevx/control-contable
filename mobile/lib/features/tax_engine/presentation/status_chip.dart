import 'package:control_contable/api/generated/models/regime_estimate_status.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A small outlined label: the status is in the word, never in the color.
class StatusChip extends StatelessWidget {
  /// The chip reading [label] in [color] (defaults to the body text).
  const new({required this.label, this.color, super.key});

  /// The chip for a regime's [status]; an unknown one reads as
  /// informational (no amount is shown for it either).
  factory of(BuildContext context, RegimeEstimateStatus status) {
    final l10n = AppLocalizations.of(context);
    return StatusChip(
      label: switch (status) {
        RegimeEstimateStatus.complete => l10n.estimateComplete,
        RegimeEstimateStatus.incomplete => l10n.estimateIncomplete,
        RegimeEstimateStatus.partial => l10n.estimatePartial,
        RegimeEstimateStatus.informational ||
        RegimeEstimateStatus.$unknown => l10n.estimateInformational,
      },
    );
  }

  /// The status word.
  final String label;

  /// Text and border color.
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final ink = color ?? tokens.colors.text;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: ink),
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.xs,
          vertical: AppSpacing.xxs,
        ),
        child: Text(
          label,
          style: Theme.of(context).textTheme.labelLarge?.copyWith(color: ink),
        ),
      ),
    );
  }
}
