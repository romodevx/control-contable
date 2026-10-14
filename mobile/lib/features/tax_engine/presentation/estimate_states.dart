import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

const String _parametersUnavailable = 'parameters_unavailable';
const String _outOfRange = 'out_of_range';

/// Heights of the skeleton's placeholder blocks: the totals hero, then two
/// regime cards.
const List<double> _skeletonBlocks = [220, 140, 140];

/// The estimate's shape while it loads: the status line (announced) and
/// grey blocks where the totals and regimes will be (not announced).
class EstimateSkeleton extends StatelessWidget {
  /// Creates the skeleton.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return Column(
      key: const ValueKey('estimate-skeleton'),
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        LoadStatus(
          message: AppLocalizations.of(context).estimateLoading,
          busy: true,
        ),
        for (final height in _skeletonBlocks)
          Padding(
            padding: const EdgeInsets.only(top: AppSpacing.md),
            child: ExcludeSemantics(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: tokens.colors.border,
                  borderRadius: BorderRadius.circular(tokens.radii.card),
                ),
                child: SizedBox(height: height),
              ),
            ),
          ),
      ],
    );
  }
}

/// A failed estimate load: why, in es-MX, and a retry.
class EstimateFailure extends StatelessWidget {
  /// The failure of [error], retried by [onRetry].
  const new({required this.error, required this.onRetry, super.key});

  /// What the load threw.
  final Object error;

  /// Loads again.
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => LoadFailed(
    message: _message(AppLocalizations.of(context)),
    onRetry: onRetry,
  );

  String _message(AppLocalizations l10n) => switch (error) {
    ApiUnreachable() => l10n.estimateUnreachable,
    ApiNotFound() => l10n.estimateNotFound,
    ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
        when details.any((d) => d.code == _parametersUnavailable) =>
      l10n.estimateParametersUnavailable,
    ApiValidationFailure(:final details)
        when details.any((d) => d.code == _outOfRange) =>
      l10n.estimateFutureMonth,
    _ => l10n.estimateLoadFailed,
  };
}

/// A plain message where the estimate would be (no taxpayer, no regime).
class EstimateMessage extends StatelessWidget {
  /// Reads [message].
  const new({required this.message, super.key});

  /// What the user reads.
  final String message;

  @override
  Widget build(BuildContext context) =>
      Text(message, style: Theme.of(context).textTheme.bodyLarge);
}
